import os
import httpx
from typing import List, Optional, Any
from google.genai import types
from google.adk.agents import LlmAgent
from google.adk.tools.base_tool import BaseTool
from google.adk.tools.base_toolset import BaseToolset
from google.adk.agents.readonly_context import ReadonlyContext
from google.adk.tools.tool_context import ToolContext
from config import settings

# 1. Resolve MCP Server URL from environment or fallback default
mcp_url = os.environ.get(
    "MCP_SERVER_URL",
    settings.get("server", {}).get("mcp_url", "http://127.0.0.1:8080/api/v1/mcp")
)

# 2. Custom Stateless MCP Tool and Toolset Implementation
class StatelessMcpTool(BaseTool):
    def __init__(self, name: str, description: str, input_schema: dict, mcp_url: str):
        super().__init__(name=name, description=description)
        self.input_schema = input_schema
        self.mcp_url = mcp_url

    def _get_declaration(self) -> Optional[types.FunctionDeclaration]:
        properties = {}
        required = self.input_schema.get("required", [])
        
        for prop_name, prop_val in self.input_schema.get("properties", {}).items():
            t = prop_val.get("type", "string")
            genai_type = types.Type.STRING
            if t == "integer":
                genai_type = types.Type.INTEGER
            elif t == "number":
                genai_type = types.Type.NUMBER
            elif t == "boolean":
                genai_type = types.Type.BOOLEAN
            elif t == "object":
                genai_type = types.Type.OBJECT
            elif t == "array":
                genai_type = types.Type.ARRAY
                
            properties[prop_name] = types.Schema(
                type=genai_type,
                description=prop_val.get("description", ""),
            )
            
        parameters = types.Schema(
            type=types.Type.OBJECT,
            properties=properties,
            required=required,
        )
        
        return types.FunctionDeclaration(
            name=self.name,
            description=self.description,
            parameters=parameters,
        )

    async def run_async(self, *, args: dict[str, Any], tool_context: ToolContext) -> Any:
        import sys
        print(f"[Auth] run_async tool_context: {tool_context}", file=sys.stderr)
        if tool_context:
            print(f"[Auth] run_async invocation_context: {getattr(tool_context, '_invocation_context', None)}", file=sys.stderr)
            if hasattr(tool_context, '_invocation_context') and tool_context._invocation_context:
                try:
                    print(f"[Auth] run_async invocation_context dict: {tool_context._invocation_context.__dict__}", file=sys.stderr)
                except Exception as ex:
                    print(f"[Auth] run_async failed to dump invocation_context: {ex}", file=sys.stderr)

        # Import and resolve GORM User ID from thread-local context variable
        from shared_context import active_user_id_var, session_user_map
        user_id = active_user_id_var.get()

        # Fallback to ADK session context if contextvar is empty
        if not user_id and tool_context and getattr(tool_context, "_invocation_context", None):
            # Try to read the verified OIDC user email directly from the ADK context
            user_email = getattr(tool_context._invocation_context, "user_email", "")
            if user_email:
                user_id = user_email
            else:
                session_id = tool_context._invocation_context.user_id
                # Strip the ADK "A2A_USER_" prefix if present to match the JSON-RPC context_id
                clean_session_id = session_id.removeprefix("A2A_USER_")
                # Resolve the true user email address from the global session mapping if present
                user_id = session_user_map.get(clean_session_id, session_id)

        # Strictly reject any missing or bypass user contexts
        if not user_id or user_id == "00000000-0000-0000-0000-000000000000":
            raise ValueError("Unauthorized: A valid, authenticated user context is required to execute MCP tools")

        print(f"[Auth] run_async header X-User-ID: {user_id}", file=sys.stderr)

        async with httpx.AsyncClient() as client:
            headers = {
                "Content-Type": "application/json",
                "X-User-ID": user_id
            }
            # Inject the active OpenTelemetry tracing context into the headers
            # to propagate the trace to the Go MCP server and nest the spans!
            from opentelemetry.trace.propagation.tracecontext import TraceContextTextMapPropagator
            TraceContextTextMapPropagator().inject(headers)

            token = get_google_id_token(self.mcp_url)
            if token:
                headers["Authorization"] = f"Bearer {token}"

            try:
                resp = await client.post(
                    self.mcp_url,
                    json={
                        "jsonrpc": "2.0",
                        "method": "tools/call",
                        "params": {
                            "name": self.name,
                            "arguments": args
                        },
                        "id": 1
                    },
                    headers=headers,
                    timeout=30.0
                )
                resp.raise_for_status()
            except httpx.HTTPStatusError as e:
                import sys
                print(f"[Auth] MCP tool run HTTP status error: {e.response.status_code}, response: {e.response.text}", file=sys.stderr)
                raise e
            data = resp.json()
            if "error" in data:
                raise Exception(f"MCP error: {data['error']}")
            
            result = data.get("result", {})
            content_text = ""
            for content in result.get("content", []):
                if content.get("type") == "text":
                    content_text += content.get("text", "")
                    
            if result.get("isError"):
                raise Exception(f"Tool execution failed: {content_text}")
                
            # Intercept and cache high-volume A2UI JSON cards to prevent LLM text corruption
            import builtins
            if not hasattr(builtins, "CACHED_A2UI_CARDS"):
                builtins.CACHED_A2UI_CARDS = {}

            if self.name == "get_tasks" and args.get("format") == "a2ui":
                builtins.CACHED_A2UI_CARDS["[A2UI_CARD_TASK_LIST_CACHED]"] = content_text
                print(f"[A2UI Cache] Intercepted get_tasks a2ui card and cached it in builtins.CACHED_A2UI_CARDS. Length: {len(content_text)}")
                return "[A2UI_CARD_TASK_LIST_CACHED]"
                
            if self.name == "get_site_locations" and args.get("format") == "a2ui":
                builtins.CACHED_A2UI_CARDS["[A2UI_CARD_LOCATIONS_CACHED]"] = content_text
                print(f"[A2UI Cache] Intercepted get_site_locations a2ui card and cached it in builtins.CACHED_A2UI_CARDS. Length: {len(content_text)}")
                return "[A2UI_CARD_LOCATIONS_CACHED]"
                
            if self.name == "get_task_details":
                builtins.CACHED_A2UI_CARDS["[A2UI_CARD_TASK_DETAILS_CACHED]"] = content_text
                print(f"[A2UI Cache] Intercepted get_task_details a2ui card and cached it in builtins.CACHED_A2UI_CARDS. Length: {len(content_text)}")
                return "[A2UI_CARD_TASK_DETAILS_CACHED]"

            if self.name == "get_weather":
                builtins.CACHED_A2UI_CARDS["[A2UI_CARD_WEATHER_CACHED]"] = content_text
                print(f"[A2UI Cache] Intercepted get_weather a2ui card and cached it in builtins.CACHED_A2UI_CARDS. Length: {len(content_text)}")
                return "[A2UI_CARD_WEATHER_CACHED]"

            if self.name == "get_store_selector":
                builtins.CACHED_A2UI_CARDS["[A2UI_CARD_STORE_SELECTOR_CACHED]"] = content_text
                print(f"[A2UI Cache] Intercepted get_store_selector a2ui card and cached it in builtins.CACHED_A2UI_CARDS. Length: {len(content_text)}")
                return "[A2UI_CARD_STORE_SELECTOR_CACHED]"
                
            return content_text

def get_google_id_token(target_url: str) -> Optional[str]:
    import sys
    from urllib.parse import urlparse
    import urllib.request
    import urllib.parse
    
    parsed = urlparse(target_url)
    if "localhost" in parsed.netloc or "127.0.0.1" in parsed.netloc:
        # Local development: retrieve developer identity token via gcloud command
        import subprocess
        try:
            result = subprocess.run(
                ["gcloud", "auth", "print-identity-token"],
                capture_output=True,
                text=True,
                check=True
            )
            token = result.stdout.strip()
            print(f"[Auth] Local Google ID Token retrieved via gcloud, length: {len(token)}", file=sys.stderr)
            return token
        except Exception as e:
            print(f"[Auth] Local gcloud token fetch failed: {e}", file=sys.stderr)
            return None
        
    audience = f"{parsed.scheme}://{parsed.netloc}"
    print(f"[Auth] Requesting Google ID Token for audience: {audience}", file=sys.stderr)
    
    try:
        url = f"http://metadata.google.internal/computeMetadata/v1/instance/service-accounts/default/identity?audience={urllib.parse.quote_plus(audience)}"
        req = urllib.request.Request(url)
        req.add_header("Metadata-Flavor", "Google")
        with urllib.request.urlopen(req, timeout=2) as response:
            token = response.read().decode("utf-8").strip()
            print(f"[Auth] Google ID Token fetched successfully, length: {len(token)}", file=sys.stderr)
            return token
    except Exception as e:
        print(f"[Auth] Metadata token fetch failed: {e}", file=sys.stderr)
        return None

class StatelessMcpToolset(BaseToolset):
    def __init__(self, mcp_url: str):
        super().__init__()
        self.mcp_url = mcp_url

    async def get_tools(self, readonly_context: Optional[ReadonlyContext] = None) -> List[BaseTool]:
        async with httpx.AsyncClient() as client:
            headers = {"Content-Type": "application/json"}
            token = get_google_id_token(self.mcp_url)
            if token:
                headers["Authorization"] = f"Bearer {token}"

            try:
                resp = await client.post(
                    self.mcp_url,
                    json={
                        "jsonrpc": "2.0",
                        "method": "tools/list",
                        "id": 1
                    },
                    headers=headers,
                    timeout=10.0
                )
                resp.raise_for_status()
                data = resp.json()
                tools_list = data.get("result", {}).get("tools", [])
                
                tools = []
                for t in tools_list:
                    tools.append(
                        StatelessMcpTool(
                            name=t["name"],
                            description=t["description"],
                            input_schema=t["inputSchema"],
                            mcp_url=self.mcp_url
                        )
                    )
                return tools
            except Exception as e:
                import sys
                print(f"[Warning] Failed to fetch tools from Go MCP server at {self.mcp_url}: {e}.", file=sys.stderr)
                print("[Warning] ADK Agent will boot with an empty toolset and dynamically retry on the next conversational turn.", file=sys.stderr)
                return []

task_lister_instruction = """
You are a direct, capability-focused operations coordinator integrated into the Gemini Task Engine. 
You interact with retail storefront databases, compliance logs, and vector SOP bases via Model Context Protocol (MCP) tools.

You support the following task types and capabilities:
1. View your tasks: Retrieve the prioritized active task queue for your site. You can optionally filter the returned list by required role (`role_id`), assigned associate (`assignee_id`), or location (`location_id`).
2. Claim a task: Assign a task from the site queue to yourself to begin execution.
3. Update task status: Mark your claimed tasks as IN_PROGRESS, COMPLETED, or update their checklist items.
4. Propose a task trade: Propose trading a task execution with a colleague.
5. List pending trades: List any active coworker task trade proposals waiting for checker approval.
6. Accept or Reject a task trade: Action pending coworker task trades.
7. Override an asset constraint: File an administrative compliance override justification for a blocked asset.
8. Search SOPs: Perform semantic lookup queries on standard operating procedures.
9. Trigger an alert: Send a site alert triggering immediate ad-hoc task generation.
10. View locations: List the locations (aisles, registers, fixtures) configured for a site using 'get_site_locations'.
11. View weather: Query real-time meteorological observations and wind patterns using 'get_weather'.
12. Switch storefront: Scope, select, or switch your active retail storefront site context using 'get_store_selector'.

OPERATIONAL PROTOCOL:
1. Bootstream: You have no initial spatial awareness. Before listing or altering tasks, you MUST call 'get_user_context' to determine the current caller's identity, assigned organizations, and assigned sites.
2. Parameter Resolution: Once 'get_user_context' returns, establish the active site. Never invent site or organization UUIDs; select only from the verified lists returned by the server.
3. Maker/Checker Verification: For any high-priority task, search the standard operating procedures using 'query_sop' before proposing or accepting peer task trades.
4. Action Handling: When the user triggers an action (via a button click in A2UI):
   - "SET_STORE": immediately call 'get_tasks' tool with 'site_id' (resolved from siteID) and 'format': 'a2ui' to display the active task list for the newly selected store. This will return the token '[A2UI_CARD_TASK_LIST_CACHED]'. You MUST output this token exactly as-is.
   - "CLAIM_TASK": call 'claim_task' tool with the provided 'execution_id'.
   - "START_TASK": call 'update_task_status' tool with 'execution_id' and status="IN_PROGRESS".
   - "COMPLETE_TASK": call 'update_task_status' tool with 'execution_id' and status="COMPLETED".
   - "UPDATE_CHECKLIST": call 'update_task_status' tool with 'execution_id', status="IN_PROGRESS", and the provided 'checklist_state'.
   - "PROPOSE_TRADE": call 'propose_trade' tool with 'task_execution_id' (resolved from execution_id).

5. Auto-Refresh Details: Upon successfully executing a 'CLAIM_TASK', 'START_TASK', or 'UPDATE_CHECKLIST' action, you MUST immediately invoke the 'get_task_details' tool for that specific 'execution_id' and return its updated A2UI details card in your response. This guarantees that the user and your own context are always looking at the freshest, most up-to-date task details and checklist state.

A2UI CARD OUTPUT PROTOCOL:
You MUST format structured UI cards for retail-focused responses by outputting a single cached token exactly as-is in your response. Do not attempt to write or synthesize any JSON yourself.

1. When the user asks to switch stores, change storefront context, or list available stores WITHOUT specifying a store name, you MUST invoke the 'get_store_selector' tool to present the RETAIL STOREFRONT CONTEXT SWITCHER card. This will return the token '[A2UI_CARD_STORE_SELECTOR_CACHED]'. You MUST output this token exactly as-is in your response.
2. If the user SPECIFIES a store name (e.g. "Volt & Vine - Seattle") and asks to select it or show tasks for it:
   - Call 'get_user_context' if you don't already have the site list to resolve the store name to its site ID.
   - Once the site ID is resolved, immediately call 'get_tasks' with that 'site_id' and 'format': 'a2ui' to display their tasks for that specific store. This will return the token '[A2UI_CARD_TASK_LIST_CACHED]'. You MUST output this token exactly as-is in your response.
3. When listing tasks, you MUST call the 'get_tasks' tool with 'format': 'a2ui'. This will return the token '[A2UI_CARD_TASK_LIST_CACHED]'. You MUST output this token exactly as-is in your response.
4. When displaying task details (such as when 'get_task_details' is called), the tool will return the token '[A2UI_CARD_TASK_DETAILS_CACHED]'. You MUST output this token exactly as-is in your response.
5. When a task requires vault deposit / compliance override check (e.g., till drawer limit reached), format a CASH VAULT DROP VERIFICATION TICKET card:
```json
{
  "type": "card",
  "title": "CASH VAULT DROP VERIFICATION TICKET",
  "style": "critical",
  "children": [
    {
      "type": "table",
      "rows": [
        {"label": "Register Channel", "value": "Register Terminal <ID>"},
        {"label": "Audit Ceiling", "value": "$<Limit> (EXCEEDED LIMIT)"},
        {"label": "Target Secure Pouch", "value": "<Pouch ID>"},
        {"label": "Deposit Safe Vault", "value": "<Vault Location>"}
      ]
    },
    {
      "type": "row",
      "align": "end",
      "children": [
        {
          "type": "button",
          "label": "Force Vault Compliance Verify & Override",
          "style": "primary",
          "action": "OVERRIDE",
          "actionData": {
            "taskExecutionID": "<Task Execution ID>",
            "pouchID": "<Pouch ID>",
            "assetID": "<Asset ID>"
          }
        }
      ]
    }
  ]
}
```

6. When listing or displaying trade proposals, format a PEER TASK TRADE PROPOSAL card:
```json
{
  "type": "card",
  "title": "PEER TASK TRADE PROPOSAL",
  "style": "primary",
  "children": [
    {
      "type": "table",
      "rows": [
        {"label": "Proposing Coworker", "value": "<Proposer Name>"},
        {"label": "Task Description", "value": "<Task Name>"}
      ]
    },
    {
      "type": "table",
      "rows": [
        {"label": "Proposed Assignee", "value": "<Assignee Name>"},
        {"label": "Assignee Role", "value": "<Assignee Role>"}
      ]
    },
    {
      "type": "row",
      "align": "end",
      "children": [
        {
          "type": "button",
          "label": "Accept Trade Swap",
          "style": "primary",
          "action": "TRADE_ACCEPT",
          "actionData": {"tradeID": "<Trade ID>"}
        },
        {
          "type": "button",
          "label": "Deny",
          "style": "secondary",
          "action": "TRADE_DENY",
          "actionData": {"tradeID": "<Trade ID>"}
        }
      ]
    }
  ]
}
```

7. When the user asks about the store map layout or the specific spatial location of a fixture, register, safe, or vault, you MUST call the 'get_site_locations' tool with 'format': 'a2ui'. This will return the token '[A2UI_CARD_LOCATIONS_CACHED]'. You MUST output this token exactly as-is in your response.
8. When the user asks about the weather, regional meteorological conditions, or METAR observations for a site or station, you MUST call the 'get_weather' tool with the regional station code (e.g. 'KDFW'). This will return the token '[A2UI_CARD_WEATHER_CACHED]'. You MUST output this token exactly as-is in your response.
"""

# 4. Define the after_model_callback to strip tool namespaces (e.g. 'default_api.')
def strip_tool_namespaces_callback(callback_context, llm_response):
    """Surgically strips any namespace prefixes (e.g. 'default_api.') from model tool calls."""
    if llm_response and llm_response.content and llm_response.content.parts:
        for part in llm_response.content.parts:
            if part.function_call and part.function_call.name:
                name = part.function_call.name
                if "." in name:
                    clean_name = name.split(".")[-1]
                    import sys
                    print(f"[Callback] Stripping tool namespace: {name} -> {clean_name}", file=sys.stderr)
                    part.function_call.name = clean_name
    return None

# 5. Define the main ADK LlmAgent
root_agent = LlmAgent(
    name="Gemini_Task_Agent",
    model="gemini-2.5-flash",
    description="A direct, database-grounded retail operations coordinator capable of managing task queues, executing compliance overrides, performing semantic SOP lookups, and coordinating peer-to-peer task handovers.",
    instruction=task_lister_instruction,
    tools=[
        StatelessMcpToolset(mcp_url=mcp_url)
    ],
    after_model_callback=strip_tool_namespaces_callback,
)
