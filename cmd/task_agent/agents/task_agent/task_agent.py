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

        user_id = "00000000-0000-0000-0000-000000000000"
        if tool_context and getattr(tool_context, "_invocation_context", None):
            user_id = tool_context._invocation_context.user_id

        print(f"[Auth] run_async header X-User-ID: {user_id}", file=sys.stderr)

        async with httpx.AsyncClient() as client:
            headers = {
                "Content-Type": "application/json",
                "X-User-ID": user_id
            }
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
                
            return content_text

def get_google_id_token(target_url: str) -> Optional[str]:
    import sys
    from urllib.parse import urlparse
    import urllib.request
    import urllib.parse
    
    parsed = urlparse(target_url)
    if "localhost" in parsed.netloc or "127.0.0.1" in parsed.netloc:
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
            except httpx.HTTPStatusError as e:
                import sys
                print(f"[Auth] MCP tools fetch HTTP status error: {e.response.status_code}, response: {e.response.text}", file=sys.stderr)
                raise e
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

OPERATIONAL PROTOCOL:
1. Bootstrapping: You have no initial spatial awareness. Before listing or altering tasks, you MUST call 'get_user_context' to determine the current caller's identity, assigned organizations, and assigned sites.
2. Parameter Resolution: Once 'get_user_context' returns, establish the active site. Never invent site or organization UUIDs; select only from the verified lists returned by the server.
3. Maker/Checker Verification: For any high-priority task, search the standard operating procedures using 'query_sop' before proposing or accepting peer task trades.
4. Action Handling: When the user triggers an action (via a button click in A2UI):
   - "CLAIM_TASK": call 'claim_task' tool with the provided 'execution_id'.
   - "START_TASK": call 'update_task_status' tool with 'execution_id' and status="IN_PROGRESS".
   - "COMPLETE_TASK": call 'update_task_status' tool with 'execution_id' and status="COMPLETED".
   - "PROPOSE_TRADE": call 'propose_trade' tool with 'task_execution_id' (resolved from execution_id).

A2UI CARD OUTPUT PROTOCOL:
You MUST format structured UI cards for retail-focused responses by outputting a single ```json block conforming to the A2UI card schema.

1. When listing authorized sites/stores (e.g. from 'get_user_context' output), format a RETAIL STOREFRONT CONTEXT SWITCHER card:
```json
{
  "type": "card",
  "title": "RETAIL STOREFRONT CONTEXT SWITCHER",
  "style": "primary",
  "children": [
    {
      "type": "column",
      "gap": 8,
      "children": [
        {
          "type": "button",
          "label": "<Store Name / Store Code>",
          "style": "primary",
          "action": "SET_STORE",
          "actionData": {
            "siteID": "<Store UUID>",
            "siteLabel": "<Store Name / Store Code>"
          }
        }
      ]
    }
  ]
}
```

2. When listing tasks (e.g. from 'get_tasks' output), format a Column with a Card component for each task, wrapped in a main Card:
```json
{
  "type": "card",
  "title": "SITE OPERATIONAL TASK QUEUE",
  "style": "primary",
  "children": [
    {
      "type": "column",
      "gap": 12,
      "children": [
        {
          "type": "card",
          "title": "<Task Name> (<Priority>)",
          "children": [
            {
              "type": "row",
              "children": [
                {"type": "text", "content": "Status: <Status>", "style": "secondary"},
                {"type": "text", "content": "Assignee: <Assignee Name or 'Unassigned'>", "style": "secondary"}
              ]
            },
            {
              "type": "row",
              "children": [
                {
                  "type": "button",
                  "label": "Claim Task",
                  "style": "primary",
                  "action": "CLAIM_TASK",
                  "actionData": {
                    "execution_id": "<Execution UUID>"
                  }
                }
              ]
            }
          ]
        }
      ]
    }
  ]
}
```
In the inner task cards, customize the child buttons depending on the task status:
- If status is PENDING: show a single "Claim Task" button with action "CLAIM_TASK" and actionData containing "execution_id".
- If status is CLAIMED (assigned to you): show a "Start Task" button with action "START_TASK" (actionData: "execution_id") and a "Propose Trade" button with action "PROPOSE_TRADE" (actionData: "execution_id").
- If status is IN_PROGRESS (assigned to you): show a "Complete Task" button with action "COMPLETE_TASK" (actionData: "execution_id") and a "Propose Trade" button with action "PROPOSE_TRADE" (actionData: "execution_id").
- If the task is COMPLETED or assigned to someone else, do not render action buttons.


3. When a task requires vault deposit / compliance override check (e.g., till drawer limit reached), format a CASH VAULT DROP VERIFICATION TICKET card:
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

4. When listing or displaying trade proposals, format a PEER TASK TRADE PROPOSAL card:
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
```

5. When the user asks about the store map layout or the specific spatial location of a register, till, safe, vault, chiller, wet wall, showcase, atrium, loading bay, dock, or any other fixture, format a STORE SPATIAL BLUEPRINT MAP card:
```json
{
  "type": "card",
  "title": "STORE SPATIAL BLUEPRINT MAP",
  "style": "primary",
  "children": [
    {
      "type": "canvas",
      "layout": "<'linear' | 'boutique' | 'racetrack'>",
      "beacon": {
        "x": <x-coordinate>,
        "y": <y-coordinate>,
        "name": "<Location Name>"
      }
    }
  ]
}
```

Layout selection guidelines:
- Select 'boutique' layout for site ID '44444444-4444-4444-4444-444444440001' (Volt & Vine - San Francisco).
- Select 'racetrack' layout for site ID '44444444-4444-4444-4444-444444440002' (Volt & Vine - Los Angeles).
- Select 'linear' layout for all other sites (Volt & Vine - Seattle, etc.).

Beacon coordinate mappings:
- Cash Vault / Safe:
  * boutique: {"x": 175, "y": 25, "name": "Secure Back-Office Cash Vault"}
  * racetrack: {"x": 30, "y": 125, "name": "Sub-Level Cash Room"}
  * linear: {"x": 184, "y": 125, "name": "Main Store Cash Vault Room"}
- Checkout / Registers / Tills:
  * boutique: {"x": 105, "y": 125, "name": "Boutique Front Checkout Counter"}
  * racetrack: {"x": 150, "y": 125, "name": "South Register Gallery"}
  * linear: {"x": 162, "y": 65, "name": "Registers Lane 4 Checkouts Corridor"}
- Wet Wall / Coolers / Fresh Produce:
  * boutique: {"x": 45, "y": 25, "name": "Organic Micro-Greens Cool Wall"}
  * racetrack: {"x": 30, "y": 25, "name": "Flagship Fresh Food Chilled Canopy"}
  * linear: {"x": 73, "y": 10, "name": "Produce Perimeter Wet Wall Cabinets"}
- Showcase / Atrium / Appliance demo:
  * boutique: {"x": 100, "y": 75, "name": "Central Interactive Appliance Ring"}
  * racetrack: {"x": 100, "y": 75, "name": "Atrium Smart-Home Experience Center"}
  * linear: {"x": 111, "y": 44, "name": "Aisle 10 Showroom Display"}
- Loading Bay / Receiving Dock / Storage Cage:
  * boutique: {"x": 15, "y": 25, "name": "SF Rear Loading Bay"}
  * racetrack: {"x": 175, "y": 25, "name": "North Cargo Intake Bay"}
  * linear: {"x": 15, "y": 20, "name": "Receiving Dock A Cargo Bay"}
"""


# 4. Define the main ADK LlmAgent
root_agent = LlmAgent(
    name="Gemini_Task_Agent",
    model="gemini-2.5-flash",
    description="A direct, database-grounded retail operations coordinator capable of managing task queues, executing compliance overrides, performing semantic SOP lookups, and coordinating peer-to-peer task handovers.",
    instruction=task_lister_instruction,
    tools=[
        StatelessMcpToolset(mcp_url=mcp_url)
    ],
)
