import re
import json
import uuid
from typing import Optional, List, Dict, Any
from google.genai import types as genai_types
from a2a import types as a2a_types
import google.adk.a2a.converters.part_converter as pc
import google.adk.a2a.converters.event_converter as ec

def normalize_card_to_a2ui_messages(flat_card: Dict[str, Any], surface_id: str = "@default") -> List[Dict[str, Any]]:
    components = {}
    comp_counter = 0

    def next_id(prefix: str) -> str:
        nonlocal comp_counter
        comp_counter += 1
        return f"{prefix}_{comp_counter}"

    def process_children(children_list: List[Dict[str, Any]]) -> List[str]:
        processed_ids = []
        switcher_buttons = [c for c in children_list if c.get("type", "").lower() == "button" and c.get("action") == "SET_STORE"]
        
        if len(switcher_buttons) >= 2:
            choice_id = next_id("multiplechoice")
            options_list = []
            for btn in switcher_buttons:
                label = btn.get("label", "")
                action_data = btn.get("actionData", {})
                site_id = action_data.get("siteID", action_data.get("siteId", ""))
                options_list.append({
                    "label": {"literalString": str(label)},
                    "value": str(site_id)
                })
            
            selections_path = "/store_switcher/selected"
            components[choice_id] = {
                "id": choice_id,
                "component": {
                    "MultipleChoice": {
                        "options": options_list,
                        "selections": {
                            "path": selections_path
                        },
                        "maxAllowedSelections": 1
                    }
                }
            }
            
            btn_id = next_id("button")
            btn_label_id = next_id("text")
            components[btn_label_id] = {
                "id": btn_label_id,
                "component": {
                    "Text": {
                        "text": {"literalString": "Switch Store"}
                    }
                }
            }
            components[btn_id] = {
                "id": btn_id,
                "component": {
                    "Button": {
                        "child": btn_label_id,
                        "action": {
                            "name": "SET_STORE",
                            "context": [
                                {
                                    "key": "siteID",
                                    "value": {
                                        "path": selections_path
                                    }
                                }
                            ]
                        }
                    }
                }
            }
            
            switcher_set = set(id(b) for b in switcher_buttons)
            for child in children_list:
                if id(child) in switcher_set:
                    continue
                child_id = process_node(child)
                if child_id:
                    processed_ids.append(child_id)
            
            processed_ids.append(choice_id)
            processed_ids.append(btn_id)
        else:
            for child in children_list:
                child_id = process_node(child)
                if child_id:
                    processed_ids.append(child_id)
                    
        return processed_ids

    def process_node(node: Dict[str, Any]) -> Optional[str]:
        node_type = node.get("type", "").lower()
        
        if node_type == "card":
            comp_id = next_id("card")
            properties = {}
            children_ids = []
            
            title = node.get("title")
            if title:
                properties["title"] = str(title)
                title_id = next_id("text")
                components[title_id] = {
                    "id": title_id,
                    "component": {
                        "Text": {
                            "text": {"literalString": str(title)},
                            "usageHint": "h2"
                        }
                    }
                }
                children_ids.append(title_id)
                
            children_ids.extend(process_children(node.get("children", [])))
            
            if len(children_ids) == 1:
                properties["child"] = children_ids[0]
            elif len(children_ids) > 1:
                wrapper_col_id = next_id("column")
                components[wrapper_col_id] = {
                    "id": wrapper_col_id,
                    "component": {
                        "Column": {
                            "alignment": "stretch",
                            "distribution": "start",
                            "children": {
                                "explicitList": children_ids
                            }
                        }
                    }
                }
                properties["child"] = wrapper_col_id
                    
            components[comp_id] = {
                "id": comp_id,
                "component": {
                    "Card": properties
                }
            }
            return comp_id

        elif node_type == "column":
            comp_id = next_id("column")
            properties = {
                "alignment": node.get("alignment", "stretch"),
                "distribution": node.get("distribution", "start")
            }
            properties["children"] = {
                "explicitList": process_children(node.get("children", []))
            }
            components[comp_id] = {
                "id": comp_id,
                "component": {
                    "Column": properties
                }
            }
            return comp_id

        elif node_type == "row":
            comp_id = next_id("row")
            properties = {
                "alignment": node.get("alignment", "stretch"),
                "distribution": node.get("distribution", "start")
            }
            properties["children"] = {
                "explicitList": process_children(node.get("children", []))
            }
            components[comp_id] = {
                "id": comp_id,
                "component": {
                    "Row": properties
                }
            }
            return comp_id

        elif node_type == "button":
            comp_id = next_id("button")
            
            action_name = node.get("action")
            action_data = node.get("actionData", {})
            context_list = []
            for k, v in action_data.items():
                val_obj = {}
                if isinstance(v, bool):
                    val_obj["literalBoolean"] = v
                elif isinstance(v, (int, float)):
                    val_obj["literalNumber"] = v
                else:
                    val_obj["literalString"] = str(v)
                context_list.append({
                    "key": k,
                    "value": val_obj
                })
                
            action_properties = {
                "name": action_name,
                "context": context_list
            }
            
            label_text = node.get("label", "")
            label_id = next_id("text")
            components[label_id] = {
                "id": label_id,
                "component": {
                    "Text": {
                        "text": {"literalString": str(label_text)}
                    }
                }
            }
            
            components[comp_id] = {
                "id": comp_id,
                "component": {
                    "Button": {
                        "action": action_properties,
                        "child": label_id,
                        "label": str(label_text)
                    }
                }
            }
            return comp_id

        elif node_type == "text":
            comp_id = next_id("text")
            text_val = node.get("text", node.get("content", ""))
            properties = {
                "text": {"literalString": str(text_val)}
            }
            if "style" in node or "usageHint" in node:
                properties["usageHint"] = node.get("usageHint", node.get("style", "body"))
            components[comp_id] = {
                "id": comp_id,
                "component": {
                    "Text": properties
                }
            }
            return comp_id

        elif node_type == "table":
            comp_id = next_id("column")
            row_ids = []
            for row in node.get("rows", []):
                row_id = next_id("row")
                lbl = row.get("label", "")
                val = row.get("value", "")
                
                lbl_id = next_id("text")
                components[lbl_id] = {
                    "id": lbl_id,
                    "component": {
                        "Text": {
                            "text": {"literalString": str(lbl)}
                        }
                    }
                }
                
                val_id = next_id("text")
                components[val_id] = {
                    "id": val_id,
                    "component": {
                        "Text": {
                            "text": {"literalString": str(val)},
                            "usageHint": "secondary"
                        }
                    }
                }
                
                components[row_id] = {
                    "id": row_id,
                    "component": {
                        "Row": {
                            "distribution": "spaceBetween",
                            "children": {
                                "explicitList": [lbl_id, val_id]
                            }
                        }
                    }
                }
                row_ids.append(row_id)
                
            components[comp_id] = {
                "id": comp_id,
                "component": {
                    "Column": {
                        "gap": 8,
                        "children": {
                            "explicitList": row_ids
                        }
                    }
                }
            }
            return comp_id

        elif node_type == "divider":
            comp_id = next_id("divider")
            components[comp_id] = {
                "id": comp_id,
                "component": {
                    "Divider": {}
                }
            }
            return comp_id

        elif node_type == "checkbox":
            comp_id = next_id("checkbox")
            label_text = node.get("label", "")
            val = node.get("value", False)
            
            label_bound = {"literalString": str(label_text)}
            
            value_bound = {}
            if isinstance(val, bool):
                value_bound["literalBoolean"] = val
            elif isinstance(val, str) and val.startswith("/"):
                value_bound["path"] = val
            elif isinstance(val, str):
                value_bound["literalString"] = val
            else:
                value_bound["literalBoolean"] = bool(val)
                
            components[comp_id] = {
                "id": comp_id,
                "component": {
                    "CheckBox": {
                        "label": label_bound,
                        "value": value_bound
                    }
                }
            }
            return comp_id

        elif node_type == "canvas":
            comp_id = next_id("image")
            
            layout = node.get("layout", "linear")
            beacon = node.get("beacon", {})
            x = beacon.get("x")
            y = beacon.get("y")
            
            import os
            agent_host = os.environ.get("AGENT_HOST_URL")
            if not agent_host:
                agent_host = "https://gemini-task-agent-dev-10781708810.us-central1.run.app"
                
            img_url = f"{agent_host}/api/v1/blueprint?layout={layout}"
            if x is not None and y is not None:
                img_url += f"&x={x}&y={y}"
                
            components[comp_id] = {
                "id": comp_id,
                "component": {
                    "Image": {
                        "url": {"literalString": img_url}
                    }
                }
            }
            return comp_id

        return None

    root_id = process_node(flat_card)
    if not root_id:
        return []

    return [
        {
            "beginRendering": {
                "root": root_id,
                "surfaceId": surface_id,
                "styles": {}
            }
        },
        {
            "surfaceUpdate": {
                "surfaceId": surface_id,
                "components": list(components.values())
            }
        }
    ]

original_event_converter = ec.convert_event_to_a2a_message
original_part_converter = pc.convert_genai_part_to_a2a_part
original_convert_a2a_part_to_genai_part = pc.convert_a2a_part_to_genai_part

def custom_convert_genai_part_to_a2a_part(part: genai_types.Part) -> Optional[a2a_types.Part]:
    # Pass through standard conversion unless intercepted by event converter
    return original_part_converter(part)

def custom_convert_a2a_part_to_genai_part(a2a_part: a2a_types.Part) -> Optional[genai_types.Part]:
    part = a2a_part.root
    if isinstance(part, a2a_types.DataPart):
        # Check if this is an incoming A2UI button click / userInteraction event
        data = part.data
        if isinstance(data, dict):
            interaction = data.get("userInteraction", data)
            action = interaction.get("action", {})
            action_name = action.get("name")
            if action_name:
                # Extract the key-value context parameters of the button click
                context_params = {}
                for item in action.get("context", []):
                    key = item.get("key")
                    val_obj = item.get("value", {})
                    # Resolve string, number, or boolean literal values
                    val = val_obj.get("literalString", val_obj.get("literalNumber", val_obj.get("literalBoolean")))
                    if key:
                        context_params[key] = val
                
                # Format a crystal-clear, structured command that the Gemini model can reason about
                param_str = ", ".join(f'{k}="{v}"' for k, v in context_params.items())
                command = f'User clicked A2UI button action: "{action_name}" with parameters: {param_str}'
                print(f"[A2UI Monkeypatch] Transpiled incoming A2UI action event: {command}")
                return genai_types.Part(text=command)

    # Fallback to the original ADK part converter
    return original_convert_a2a_part_to_genai_part(a2a_part)
import google.adk.a2a.converters.event_converter as ec

# Global memory cache for pre-synthesized Go-native A2UI Cards to prevent LLM text corruption
CACHED_A2UI_CARD = None


def custom_convert_event_to_a2a_message(
    event,
    invocation_context,
    role = a2a_types.Role.agent,
    part_converter = None,
) -> Optional[a2a_types.Message]:
    if part_converter is None:
        part_converter = pc.convert_genai_part_to_a2a_part
        
    if not event or not event.content or not event.content.parts:
        return original_event_converter(event, invocation_context, role, part_converter)
        
    try:
        a2a_parts = []
        for part in event.content.parts:
            is_card = False
            if part.text:
                text = part.text.strip()
                json_match = re.search(r'```(?:json)?\s*(.*?)\s*```', text, re.DOTALL | re.IGNORECASE)
                json_str = None
                text_prefix = ""
                
                # Check for cached card token to bypass LLM parsing
                cache_tokens = [
                    "[A2UI_CARD_TASK_LIST_CACHED]",
                    "[A2UI_CARD_LOCATIONS_CACHED]",
                    "[A2UI_CARD_TASK_DETAILS_CACHED]",
                    "[A2UI_CARD_WEATHER_CACHED]",
                    "[A2UI_CARD_STORE_SELECTOR_CACHED]"
                ]
                has_token = any(t in text for t in cache_tokens)
                if has_token:
                    import builtins
                    cached_card = getattr(builtins, "CACHED_A2UI_CARD", None)
                    if cached_card:
                        print("[A2UI Monkeypatch] Found cached card token. Retrieving cached JSON card from builtins...")
                        json_str = cached_card
                        
                        # Strip markdown code fences if present in cached_card to prevent json.loads failure
                        json_match_cached = re.search(r'```(?:json)?\s*(.*?)\s*```', json_str, re.DOTALL | re.IGNORECASE)
                        if json_match_cached:
                            json_str = json_match_cached.group(1).strip()
                        else:
                            json_str = json_str.strip()

                        token = next(t for t in cache_tokens if t in text)
                        idx = text.find(token)
                        text_prefix = text[:idx].strip()
                    else:
                        print("[A2UI Monkeypatch] Error: Cached token found but builtins.CACHED_A2UI_CARD is None!")

                elif json_match:
                    json_str = json_match.group(1).strip()
                    text_prefix = text[:json_match.start()].strip()
                else:
                    if text.startswith("{") and text.endswith("}"):
                        json_str = text
                        text_prefix = ""
                        
                if json_str:
                    try:
                        card_data = json.loads(json_str)
                        if isinstance(card_data, dict):
                            # Check if this is a pre-synthesized Go-native flat A2UI transaction
                            if "surfaceUpdate" in card_data and "beginRendering" in card_data:
                                is_card = True
                                print("[A2UI Monkeypatch] Intercepted pre-synthesized Go-native flat A2UI transaction!")
                                a2ui_messages = []
                                if "surfaceUpdate" in card_data:
                                    a2ui_messages.append({"surfaceUpdate": card_data["surfaceUpdate"]})
                                if "dataModelUpdate" in card_data:
                                    a2ui_messages.append({"dataModelUpdate": card_data["dataModelUpdate"]})
                                if "beginRendering" in card_data:
                                    a2ui_messages.append({"beginRendering": card_data["beginRendering"]})
                                    
                                for msg in a2ui_messages:
                                    data_part = a2a_types.DataPart(
                                        data=msg,
                                        metadata={"mimeType": "application/json+a2ui"}
                                    )
                                    a2a_parts.append(a2a_types.Part(root=data_part))

                            # Fallback to legacy nested card transpilation
                            elif card_data.get("type") == "card":
                                is_card = True
                                print(f"[A2UI Monkeypatch] Intercepted A2UI Card in custom event converter!")
                                if text_prefix:
                                    clean_prefix = re.sub(r'[\r\n]+', ' ', text_prefix).strip()
                                    if clean_prefix:
                                        children = card_data.get("children", [])
                                        children.insert(0, {
                                            "type": "text",
                                            "content": clean_prefix,
                                            "style": "secondary"
                                        })
                                        card_data["children"] = children
                                        
                                a2ui_messages = normalize_card_to_a2ui_messages(card_data)
                                for msg in a2ui_messages:
                                    data_part = a2a_types.DataPart(
                                        data=msg,
                                        metadata={"mimeType": "application/json+a2ui"}
                                    )
                                    a2a_parts.append(a2a_types.Part(root=data_part))
                    except Exception as e:
                        print(f"[A2UI Monkeypatch] Failed parsing JSON card structure: {e}")
                        
            if not is_card:
                a2a_part = part_converter(part)
                if a2a_part:
                    a2a_parts.append(a2a_part)
                    
        if a2a_parts:
            return a2a_types.Message(message_id=str(uuid.uuid4()), role=role, parts=a2a_parts)
            
    except Exception as e:
        print(f"[A2UI Monkeypatch] Error in custom event converter: {e}")
        raise
        
    return None

def register_monkeypatch():
    pc.convert_genai_part_to_a2a_part = custom_convert_genai_part_to_a2a_part
    pc.convert_a2a_part_to_genai_part = custom_convert_a2a_part_to_genai_part
    ec.convert_event_to_a2a_message = custom_convert_event_to_a2a_message
