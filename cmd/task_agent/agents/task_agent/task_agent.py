# Copyright 2026 Google LLC
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

import os
import re
import json
import httpx
from typing import List, Optional, Any

def call_mcp_tool_sync(tool_name: str, arguments: dict, user_id: str, mcp_url: str) -> dict:
    import os
    import httpx
    from urllib.parse import urlparse
    parsed = urlparse(mcp_url)
    mcp_endpoint = f"{parsed.scheme}://{parsed.netloc}{parsed.path}"
    
    headers = {
        "Content-Type": "application/json",
        "X-User-ID": user_id
    }
    
    token = get_google_id_token(mcp_endpoint)
    if token:
        headers["Authorization"] = f"Bearer {token}"
        
    try:
        with httpx.Client() as client:
            resp = client.post(
                mcp_endpoint,
                json={
                    "jsonrpc": "2.0",
                    "method": "tools/call",
                    "params": {
                        "name": tool_name,
                        "arguments": arguments
                    },
                    "id": 1
                },
                headers=headers,
                timeout=10.0
            )
            resp.raise_for_status()
            return resp.json().get("result", {})
    except Exception as e:
        import sys
        print(f"[State Assembly MCP Call Error] Failed calling tool {tool_name}: {e}", file=sys.stderr)
        return {}

def assemble_unified_portal_card(
    user_id: str,
    mcp_url: str,
    active_site_id: Optional[str] = None,
    active_execution_id: Optional[str] = None,
    selected_site_id: Optional[str] = None,
    raw_tasks_json: Optional[str] = None,
    raw_task_details_json: Optional[str] = None,
    force_selector: bool = False
) -> str:
    import json
    import os
    import sys
    
    import builtins
    if not hasattr(builtins, "LAST_ACTIVE_SITE_ID"):
        builtins.LAST_ACTIVE_SITE_ID = None

    if force_selector or active_site_id == "":
        # Explicitly cleared (e.g. Switch Store clicked or get_store_selector called)
        builtins.LAST_ACTIVE_SITE_ID = None
        active_site_id = None
    elif active_site_id is None:
        # Implicitly dropped; restore from memory if available
        if builtins.LAST_ACTIVE_SITE_ID:
            active_site_id = builtins.LAST_ACTIVE_SITE_ID
    else:
        # Active site ID is set; save it to memory
        builtins.LAST_ACTIVE_SITE_ID = active_site_id
    
    user_context = {}
    sites_list = []
    
    context_result = call_mcp_tool_sync("get_user_context", {}, user_id, mcp_url)
    content = context_result.get("content", [])
    if content:
        try:
            user_context = json.loads(content[0].get("text", "{}"))
            sites_list = user_context.get("sites", [])
            if not active_site_id:
                active_site_id = user_context.get("active_site_id")
        except Exception as e:
            print(f"[State Assembly] Error parsing user context: {e}", file=sys.stderr)

    tasks_list = []
    if active_site_id:
        if raw_tasks_json:
            try:
                tasks_list = json.loads(raw_tasks_json)
            except Exception as e:
                print(f"[State Assembly] Error parsing provided raw tasks list: {e}", file=sys.stderr)
        else:
            tasks_result = call_mcp_tool_sync("get_tasks", {"site_id": active_site_id, "format": "raw"}, user_id, mcp_url)
            content = tasks_result.get("content", [])
            if content:
                try:
                    tasks_list = json.loads(content[0].get("text", "[]"))
                except Exception as e:
                    print(f"[State Assembly] Error parsing fetched tasks list: {e}", file=sys.stderr)
            
            # If there are no active tasks in the queue for this store, automatically
            # clear the active storefront context and return to the Store Selector!
            if not tasks_list:
                print(f"[State Assembly] Store {active_site_id} has 0 tasks. Automatically returning to Store Selector.")
                active_site_id = None
                import builtins
                builtins.LAST_ACTIVE_SITE_ID = None

    active_task_data = None
    if active_execution_id:
        if raw_task_details_json:
            try:
                active_task_data = json.loads(raw_task_details_json)
                if not active_site_id and active_task_data:
                    active_site_id = active_task_data.get("site_id")
            except Exception as e:
                print(f"[State Assembly] Error parsing provided raw task details: {e}", file=sys.stderr)
        else:
            task_result = call_mcp_tool_sync("get_task_details", {"execution_id": active_execution_id, "format": "raw"}, user_id, mcp_url)
            content = task_result.get("content", [])
            if content:
                try:
                    active_task_data = json.loads(content[0].get("text", "{}"))
                    if not active_site_id and active_task_data:
                        active_site_id = active_task_data.get("site_id")
                except Exception as e:
                    print(f"[State Assembly] Error parsing fetched task details: {e}", file=sys.stderr)

    unified_state = {
      "active_site_id": active_site_id,
      "sites": sites_list,
      "tasks": tasks_list,
      "active_task": active_task_data,
      "user_name": user_context.get("name", "Associate"),
      "selected_site_id": selected_site_id
    }

    state_json_str = json.dumps(unified_state)
    
    try:
        template_path = "/app/cmd/task_agent/templates/task_portal.html"
        if not os.path.exists(template_path):
            base_dir = os.path.dirname(os.path.abspath(__file__))
            template_path = os.path.abspath(os.path.join(base_dir, "../../templates/task_portal.html"))
            
        with open(template_path, "r", encoding="utf-8") as f:
            html_content = f.read()
            
        pattern = r"const state\s*=\s*window\.__TASK_STATE__\s*\|\|.*?\};"
        safe_replacement = f"window.__TASK_STATE__ = {state_json_str};\n    const state = window.__TASK_STATE__;".replace("\\", "\\\\")
        hydrated_html = re.sub(pattern, safe_replacement, html_content, flags=re.DOTALL)
    except Exception as e:
        print(f"[State Assembly] Error loading/hydrating template: {e}", file=sys.stderr)
        hydrated_html = f"<html><body><h3>Error loading interactive task portal</h3><p>{e}</p></body></html>"

    children = [
        {
            "type": "webframe",
            "htmlContent": hydrated_html,
            "height": 460
        }
    ]

    from a2ui_transpiler import set_pending_suggestions
    set_pending_suggestions([])

    card_buttons = []
    
    # If the user has highlighted a store context but has not committed it yet,
    # render a native "Next" button in the card footer (agent action)!
    if selected_site_id and not active_site_id:
        card_buttons.append({
            "type": "button",
            "label": "Next",
            "style": "primary",
            "action": "SET_STORE",
            "actionData": {
                "site_id": selected_site_id
            }
        })

    if active_task_data:
        status = active_task_data.get("status")
        steps = active_task_data.get("steps", [])
        execution_id = active_task_data.get("id")
        
        if status == "IN_PROGRESS":
            sorted_steps = sorted(steps, key=lambda s: s.get("step", 999))
            active_step = None
            for s in sorted_steps:
                if s.get("status") in ["IN_PROGRESS", "PAUSED", "PENDING"]:
                    active_step = s
                    break
            
            if active_step:
                step_num = active_step.get("step")
                step_title = active_step.get("title")
                step_status = active_step.get("status")
                
                if step_status == "PENDING":
                    card_buttons.append({
                        "type": "button",
                        "label": f"Start: {step_title}",
                        "style": "primary",
                        "action": "UPDATE_CHECKLIST",
                        "actionData": {
                            "execution_id": execution_id,
                            "status": "IN_PROGRESS",
                            "checklist_state": json.dumps({"step": step_num, "action": "START"})
                        }
                    })
                elif step_status == "IN_PROGRESS":
                    card_buttons.append({
                        "type": "button",
                        "label": f"Complete: {step_title}",
                        "style": "primary",
                        "action": "UPDATE_CHECKLIST",
                        "actionData": {
                            "execution_id": execution_id,
                            "status": "IN_PROGRESS",
                            "checklist_state": json.dumps({"step": step_num, "action": "COMPLETE"})
                        }
                    })
                    card_buttons.append({
                        "type": "button",
                        "label": "Pause Step",
                        "style": "secondary",
                        "action": "UPDATE_CHECKLIST",
                        "actionData": {
                            "execution_id": execution_id,
                            "status": "IN_PROGRESS",
                            "checklist_state": json.dumps({"step": step_num, "action": "PAUSE"})
                        }
                    })
                elif step_status == "PAUSED":
                    card_buttons.append({
                        "type": "button",
                        "label": "Resume Step",
                        "style": "primary",
                        "action": "UPDATE_CHECKLIST",
                        "actionData": {
                            "execution_id": execution_id,
                            "status": "IN_PROGRESS",
                            "checklist_state": json.dumps({"step": step_num, "action": "RESUME"})
                        }
                    })
            if not active_step:
                card_buttons.append({
                    "type": "button",
                    "label": "Complete Task",
                    "style": "primary",
                    "action": "UPDATE_TASK_STATUS",
                    "actionData": {
                        "execution_id": execution_id,
                        "status": "COMPLETED"
                    }
                })
            
            card_buttons.append({
                "type": "button",
                "label": "Pause Task",
                "style": "secondary",
                "action": "UPDATE_TASK_STATUS",
                "actionData": {
                    "execution_id": execution_id,
                    "status": "PAUSED"
                }
            })
        elif status == "PAUSED":
            card_buttons.append({
                "type": "button",
                "label": "Resume Task",
                "style": "primary",
                "action": "UPDATE_TASK_STATUS",
                "actionData": {
                    "execution_id": execution_id,
                    "status": "IN_PROGRESS"
                }
            })
        elif status == "PENDING":
            card_buttons.append({
                "type": "button",
                "label": "Start Task",
                "style": "primary",
                "action": "UPDATE_TASK_STATUS",
                "actionData": {
                    "execution_id": execution_id,
                    "status": "IN_PROGRESS"
                }
            })

    if card_buttons:
        button_row = {
            "type": "row",
            "children": card_buttons
        }
        children.append(button_row)

    card = {
        "type": "card",
        "title": "Interactive Task Portal",
        "children": children
    }
    
    return json.dumps(card)

def hydrate_task_portal_card(task_state_json_str: str) -> str:
    try:
        task_data = json.loads(task_state_json_str)
        execution_id = task_data.get("id")
        site_id = task_data.get("site_id")
    except Exception:
        execution_id = None
        site_id = None
    return assemble_unified_portal_card("00000000-0000-0000-0000-000000000000", "", site_id, execution_id, None, task_state_json_str)

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
            
        if self.name == "get_store_selector":
            properties["selected_site_id"] = types.Schema(
                type=types.Type.STRING,
                description="Optional site ID to highlight in the store selector portal card."
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

        # Automatically request raw JSON output format for task lists, details and updates
        if self.name in ["get_tasks", "get_task_details", "update_task_status", "claim_task"]:
            args["format"] = "raw"

        async with httpx.AsyncClient() as client:
            headers = {
                "Content-Type": "application/json",
                "X-User-ID": user_id
            }
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
                
            import builtins
            if not hasattr(builtins, "CACHED_A2UI_CARDS"):
                builtins.CACHED_A2UI_CARDS = {}

            # Unify and cache all storefront operations under the golden [A2UI_CARD_TASK_PORTAL_CACHED] token
            if self.name in ["get_store_selector", "get_tasks", "get_task_details", "update_task_status", "claim_task"]:
                active_site_id = args.get("site_id")
                active_execution_id = args.get("execution_id")
                selected_site_id = args.get("selected_site_id")
                
                raw_tasks = None
                raw_details = None
                
                if self.name == "get_tasks":
                    raw_tasks = content_text
                elif self.name in ["get_task_details", "update_task_status", "claim_task"]:
                    raw_details = content_text
                    try:
                        task_data = json.loads(content_text)
                        active_site_id = task_data.get("site_id")
                    except Exception:
                        pass

                portal_card_json = assemble_unified_portal_card(
                    user_id=user_id,
                    mcp_url=self.mcp_url,
                    active_site_id=active_site_id,
                    active_execution_id=active_execution_id,
                    selected_site_id=selected_site_id,
                    raw_tasks_json=raw_tasks,
                    raw_task_details_json=raw_details,
                    force_selector=(self.name == "get_store_selector")
                )
                
                builtins.CACHED_A2UI_CARDS["[A2UI_CARD_TASK_PORTAL_CACHED]"] = portal_card_json
                print(f"[A2UI Cache] Compiled and cached unified task portal card under [A2UI_CARD_TASK_PORTAL_CACHED] from tool {self.name}.")
                return "[A2UI_CARD_TASK_PORTAL_CACHED]"

            if self.name == "get_site_locations" and args.get("format") == "a2ui":
                builtins.CACHED_A2UI_CARDS["[A2UI_CARD_LOCATIONS_CACHED]"] = content_text
                print(f"[A2UI Cache] Intercepted get_site_locations a2ui card and cached it in builtins.CACHED_A2UI_CARDS. Length: {len(content_text)}")
                return "[A2UI_CARD_LOCATIONS_CACHED]"

            if self.name == "get_weather":
                builtins.CACHED_A2UI_CARDS["[A2UI_CARD_WEATHER_CACHED]"] = content_text
                print(f"[A2UI Cache] Intercepted get_weather a2ui card and cached it in builtins.CACHED_A2UI_CARDS. Length: {len(content_text)}")
                return "[A2UI_CARD_WEATHER_CACHED]"

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
1. Bootstream: You have no initial spatial awareness. Before listing or altering tasks, you MUST call 'get_user_context' to determine the current caller's identity, assigned organizations, and assigned sites. If the user's message is a greeting (e.g. "hello", "hi", "hey", "good morning"), or if no active storefront site context has been established yet, you MUST immediately invoke the 'get_store_selector' tool as your very first action to present the storefront selector, allowing the user to select their active store.
2. Parameter Resolution: Once 'get_user_context' returns, establish the active site. Never invent site or organization UUIDs; select only from the verified lists returned by the server. If no site ID is actively selected, default to invoking the 'get_store_selector' tool.
3. Maker/Checker Verification: For any high-priority task, search the standard operating procedures using 'query_sop' before proposing or accepting peer task trades.
4. Action Handling: When the user triggers an action (via a button click in A2UI):
   - "SELECT_STORE": immediately call 'get_store_selector' tool with 'selected_site_id' set to the provided site ID. E.g., 'get_store_selector(selected_site_id="00000000-0000-0000-0000-000000000002")'. This will return the token '[A2UI_CARD_TASK_PORTAL_CACHED]'. You MUST output this token exactly as-is.
   - "SET_STORE": immediately call 'get_tasks' tool with 'site_id' (resolved from siteID) to display the active task list for the newly selected store. This will return the token '[A2UI_CARD_TASK_PORTAL_CACHED]'. You MUST output this token exactly as-is.
   - "VIEW_TASK": immediately call 'get_task_details' tool with the provided 'execution_id'. This will return the token '[A2UI_CARD_TASK_PORTAL_CACHED]'. You MUST output this token exactly as-is.
   - "BACK_TO_TASKS": immediately call 'get_tasks' tool with 'site_id' (resolved from siteID) to return to the active store task list. This will return the token '[A2UI_CARD_TASK_PORTAL_CACHED]'. You MUST output this token exactly as-is.
   - "CLAIM_TASK": call 'claim_task' tool with the provided 'execution_id'.
   - "START_TASK": call 'update_task_status' tool with 'execution_id' and status="IN_PROGRESS".
   - "PAUSE_TASK": call 'update_task_status' tool with 'execution_id' and status="PAUSED".
   - "RESUME_TASK": call 'update_task_status' tool with 'execution_id' and status="IN_PROGRESS".
   - "COMPLETE_TASK": call 'update_task_status' tool with 'execution_id' and status="COMPLETED".
   - "UPDATE_CHECKLIST": call 'update_task_status' tool with 'execution_id', status="IN_PROGRESS", and the provided 'checklist_state'.
   - "PROPOSE_TRADE": call 'propose_trade' tool with 'task_execution_id' (resolved from execution_id).

5. Auto-Refresh Details: The tools 'claim_task' and 'update_task_status' will directly return the updated A2UI portal card in their tool responses (returning the token '[A2UI_CARD_TASK_PORTAL_CACHED]'). You do NOT need to make a separate 'get_task_details' call if you have just successfully executed one of these tools; simply return the updated card token returned by the tool directly to the user. This optimizes token usage, decreases latency, and prevents API rate limiting.
6. Conversational Commands: The user may send direct conversational commands (e.g., via suggestion chips or text inputs):
   - "Start Step: <step_title>": call 'update_task_status' tool for the active task, setting status="IN_PROGRESS" and checklist_state containing the step number and action "START". E.g. '{"step": 2, "action": "START"}'.
   - "Complete Step: <step_title>": call 'update_task_status' tool for the active task, setting status="IN_PROGRESS" and checklist_state containing the step number and action "COMPLETE". E.g. '{"step": 2, "action": "COMPLETE"}'.
   - "Pause Step: <step_title>": call 'update_task_status' tool for the active task, setting status="IN_PROGRESS" and checklist_state containing the step number and action "PAUSE". E.g. '{"step": 2, "action": "PAUSE"}'.
   - "Resume Step: <step_title>": call 'update_task_status' tool for the active task, setting status="IN_PROGRESS" and checklist_state containing the step number and action "RESUME". E.g. '{"step": 2, "action": "RESUME"}'.
   - "Start Task": call 'update_task_status' tool with status="IN_PROGRESS".
   - "Pause Task": call 'update_task_status' tool with status="PAUSED".
   - "Resume Task": call 'update_task_status' tool with status="IN_PROGRESS".
   - "Complete Task": call 'update_task_status' tool with status="COMPLETED".
You MUST resolve the task's execution ID and the target step's number from the conversation history before invoking the tool.

A2UI CARD OUTPUT PROTOCOL:
You MUST format structured UI cards for retail-focused responses by outputting a single cached token exactly as-is in your response. Do not attempt to write or synthesize any JSON yourself.
CRITICAL REQUIREMENT: You can ONLY output the golden token '[A2UI_CARD_TASK_PORTAL_CACHED]' if you have actively called one of the operational tools ('get_store_selector', 'get_tasks', 'get_task_details', 'claim_task', or 'update_task_status') in the current conversational turn. Never output this token directly from memory without invoking its tool, as doing so will result in a blank cache error and render failure.

1. When the user greets you (e.g. "hello", "hi", "hey"), asks to switch stores, or change storefront context, you MUST invoke the 'get_store_selector' tool to present the portal storefront switcher. You MUST output the token '[A2UI_CARD_TASK_PORTAL_CACHED]' exactly as-is in your response.
2. If the user specifies or selects a store (e.g. "Volt & Vine - Seattle"), or asks to list tasks for a store:
   - Call 'get_user_context' if you don't already have the site list to resolve the store name to its site ID.
   - Once the site ID is resolved, immediately call 'get_tasks' with that 'site_id' to display their tasks. You MUST output the token '[A2UI_CARD_TASK_PORTAL_CACHED]' exactly as-is in your response.
3. When displaying task details (such as when 'get_task_details' is called), the tool will return the token '[A2UI_CARD_TASK_PORTAL_CACHED]'. You MUST output this token exactly as-is in your response.
4. When claiming a task or updating task/step status, the tools 'claim_task' and 'update_task_status' will return the token '[A2UI_CARD_TASK_PORTAL_CACHED]'. You MUST output this token exactly as-is in your response.
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
