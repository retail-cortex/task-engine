import re
import json
import uuid
from typing import Optional, List, Dict, Any
from google.genai import types as genai_types
from a2a import types as a2a_types
import google.adk.a2a.converters.part_converter as pc
import google.adk.a2a.converters.event_converter as ec

def generate_blueprint_svg(layout: str = "linear", x: Optional[float] = None, y: Optional[float] = None) -> str:
    svg_styles = """
    rect.fixture {
        fill: rgba(99, 102, 241, 0.04);
        stroke: rgba(99, 102, 241, 0.18);
        stroke-width: 1;
    }
    .text-primary {
        fill: rgba(255, 255, 255, 0.65);
        font-family: Outfit, sans-serif;
        font-weight: 600;
    }
    .text-secondary {
        fill: rgba(255, 255, 255, 0.45);
        font-family: Outfit, sans-serif;
        font-weight: 500;
    }
    .text-muted {
        fill: rgba(255, 255, 255, 0.25);
        font-family: Outfit, sans-serif;
    }
    .beacon {
        fill: #f87171;
    }
    """
    
    svg_content = f"""<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 200 150" style="background: #0b0d19; width: 100%; height: 100%;">
      <style>{svg_styles}</style>
      <rect x="2" y="2" width="196" height="146" rx="3" fill="none" stroke="rgba(255, 255, 255, 0.03)" stroke-width="1" stroke-dasharray="4,2" />
    """
    
    if layout == "boutique":
        svg_content += """
        <rect class="fixture" x="5" y="5" width="22" height="40" rx="1" />
        <text x="16" y="27" class="text-secondary" font-size="3.2" text-anchor="middle">LOADING BAY</text>

        <rect class="fixture" x="5" y="65" width="22" height="78" rx="1" />
        <text x="16" y="105" class="text-secondary" font-size="3.2" text-anchor="middle">STOCK CAGE</text>

        <rect class="fixture" x="35" y="5" width="60" height="10" rx="1" />
        <text x="65" y="11" class="text-secondary" font-size="3" text-anchor="middle">ORGANIC MICRO-GREENS WALL</text>

        <rect class="fixture" x="155" y="5" width="40" height="35" rx="1" />
        <text x="175" y="23" class="text-secondary" font-size="3.2" text-anchor="middle">SECURE VAULT</text>

        <circle cx="100" cy="75" r="20" class="fixture" fill="none" stroke-width="3" />
        <text x="100" y="76" class="text-secondary" font-size="3.2" text-anchor="middle">SHOWCASE A</text>

        <circle cx="60" cy="75" r="12" class="fixture" fill="none" stroke-width="2" />
        <text x="60" y="76" class="text-muted" font-size="2.5" text-anchor="middle">SHOWCASE B</text>

        <circle cx="140" cy="75" r="12" class="fixture" fill="none" stroke-width="2" />
        <text x="140" y="76" class="text-muted" font-size="2.5" text-anchor="middle">SHOWCASE C</text>

        <rect class="fixture" x="145" y="110" width="50" height="33" rx="2" />
        <text x="170" y="128" class="text-secondary" font-size="3" text-anchor="middle">COFFEE BAR & LOUNGE</text>

        <rect class="fixture" x="90" y="120" width="30" height="12" rx="1" />
        <text x="105" y="127" class="text-primary" font-size="3" text-anchor="middle">CHECKOUT COUNTER</text>
        """
    elif layout == "racetrack":
        svg_content += """
        <rect class="fixture" x="50" y="40" width="100" height="65" rx="2" fill="none" stroke-width="1.5" stroke-dasharray="3,3" />
        <text x="100" y="74" class="text-secondary" font-size="3.8" text-anchor="middle">ATRIUM EXPERIENCE CENTER</text>

        <rect class="fixture" x="165" y="5" width="30" height="25" rx="1" />
        <text x="180" y="19" class="text-secondary" font-size="3" text-anchor="middle">INTAKE BAY</text>

        <rect class="fixture" x="165" y="35" width="30" height="55" rx="1" />
        <text x="180" y="65" class="text-secondary" font-size="3" text-anchor="middle" transform="rotate(-90 180 65)">STAGING AREA C</text>

        <rect class="fixture" x="5" y="110" width="40" height="35" rx="1" />
        <text x="25" y="129" class="text-secondary" font-size="3.2" text-anchor="middle">SUB-LEVEL VAULT</text>

        <rect class="fixture" x="120" y="115" width="40" height="28" rx="1" />
        <text x="140" y="131" class="text-primary" font-size="3" text-anchor="middle">REGISTER GALLERY</text>

        <rect class="fixture" x="5" y="5" width="55" height="15" rx="1" />
        <text x="32.5" y="14" class="text-secondary" font-size="3" text-anchor="middle">FRESH CANOPY</text>

        <rect class="fixture" x="65" y="5" width="90" height="15" rx="1" />
        <text x="110" y="14" class="text-secondary" font-size="3" text-anchor="middle">PERISHABLES MARKET</text>

        <rect class="fixture" x="20" y="30" width="10" height="70" rx="1" />
        <text x="25" y="65" class="text-muted" font-size="3" text-anchor="middle" transform="rotate(-90 25 65)">AISLE A</text>

        <rect class="fixture" x="140" y="30" width="10" height="70" rx="1" />
        <text x="145" y="65" class="text-muted" font-size="3" text-anchor="middle" transform="rotate(-90 145 65)">AISLE B</text>

        <rect x="10" y="24" width="150" height="2" fill="rgba(255, 255, 255, 0.03)" opacity="0.4" />
        <rect x="40" y="107" width="115" height="2" fill="rgba(255, 255, 255, 0.03)" opacity="0.4" />
        """
    else: # linear
        svg_content += """
        <rect class="fixture" x="5" y="5" width="22" height="15" rx="1" />
        <text x="16" y="14" class="text-secondary" font-size="3.2" text-anchor="middle">DOCK A</text>

        <rect class="fixture" x="5" y="23" width="22" height="15" rx="1" />
        <text x="16" y="32" class="text-secondary" font-size="3.2" text-anchor="middle">DOCK B</text>

        <rect class="fixture" x="5" y="41" width="22" height="24" rx="1" />
        <text x="16" y="54" class="text-secondary" font-size="3.2" text-anchor="middle">STORAGE CAGE B</text>

        <rect class="fixture" x="5" y="68" width="22" height="28" rx="1" />
        <text x="16" y="83" class="text-secondary" font-size="3.2" text-anchor="middle">WALK-IN COOLER</text>

        <rect class="fixture" x="5" y="99" width="22" height="46" rx="1" />
        <text x="16" y="123" class="text-secondary" font-size="3.2" text-anchor="middle">PHARMACY WING</text>

        <rect class="fixture" x="30" y="5" width="87" height="8" rx="1" />
        <text x="73.5" y="10.5" class="text-secondary" font-size="3.2" text-anchor="middle">PRODUCE PERIMETER WET WALL</text>

        <rect class="fixture" x="120" y="5" width="50" height="8" rx="1" />
        <text x="145" y="10.5" class="text-secondary" font-size="3.2" text-anchor="middle">HOT FOOD DELI DEPOT</text>

        <rect class="fixture" x="173" y="5" width="22" height="15" rx="1" />
        <text x="184" y="14" class="text-secondary" font-size="3.2" text-anchor="middle">BAKERY OVENS</text>

        <rect class="fixture" x="30" y="22" width="12" height="40" rx="1" />
        <text x="36" y="44" class="text-muted" font-size="3.5" text-anchor="middle" transform="rotate(-90 36 44)">AISLE 7A</text>
        
        <rect class="fixture" x="30" y="82" width="12" height="40" rx="1" />
        <text x="36" y="104" class="text-muted" font-size="3.5" text-anchor="middle" transform="rotate(-90 36 104)">AISLE 7B</text>

        <rect class="fixture" x="30" y="16" width="12" height="4" rx="0.5" fill="rgba(99, 102, 241, 0.25)" />
        <text x="36" y="19.2" class="text-primary" font-size="2.5" text-anchor="middle">E1</text>
        <rect class="fixture" x="30" y="124" width="12" height="4" rx="0.5" fill="rgba(99, 102, 241, 0.25)" />
        <text x="36" y="127.2" class="text-primary" font-size="2.5" text-anchor="middle">E2</text>

        <rect class="fixture" x="55" y="22" width="12" height="40" rx="1" />
        <text x="61" y="44" class="text-muted" font-size="3.5" text-anchor="middle" transform="rotate(-90 61 44)">AISLE 8A</text>

        <rect class="fixture" x="55" y="82" width="12" height="40" rx="1" />
        <text x="61" y="104" class="text-muted" font-size="3.5" text-anchor="middle" transform="rotate(-90 61 104)">AISLE 8B</text>

        <rect class="fixture" x="55" y="16" width="12" height="4" rx="0.5" fill="rgba(99, 102, 241, 0.25)" />
        <text x="61" y="19.2" class="text-primary" font-size="2.5" text-anchor="middle">E3</text>
        <rect class="fixture" x="55" y="124" width="12" height="4" rx="0.5" fill="rgba(99, 102, 241, 0.25)" />
        <text x="61" y="127.2" class="text-primary" font-size="2.5" text-anchor="middle">E4</text>

        <rect class="fixture" x="80" y="22" width="12" height="40" rx="1" />
        <text x="86" y="44" class="text-muted" font-size="3.5" text-anchor="middle" transform="rotate(-90 86 44)">AISLE 9A</text>

        <rect class="fixture" x="80" y="82" width="12" height="40" rx="1" />
        <text x="86" y="104" class="text-muted" font-size="3.5" text-anchor="middle" transform="rotate(-90 86 104)">AISLE 9B</text>

        <rect class="fixture" x="80" y="16" width="12" height="4" rx="0.5" fill="rgba(99, 102, 241, 0.25)" />
        <text x="86" y="19.2" class="text-primary" font-size="2.5" text-anchor="middle">E5</text>
        <rect class="fixture" x="80" y="124" width="12" height="4" rx="0.5" fill="rgba(99, 102, 241, 0.25)" />
        <text x="86" y="127.2" class="text-primary" font-size="2.5" text-anchor="middle">E6</text>

        <rect class="fixture" x="105" y="22" width="12" height="40" rx="1" />
        <text x="111" y="44" class="text-muted" font-size="3.5" text-anchor="middle" transform="rotate(-90 111 44)">AISLE 10A</text>

        <rect class="fixture" x="105" y="82" width="12" height="40" rx="1" />
        <text x="111" y="104" class="text-muted" font-size="3.5" text-anchor="middle" transform="rotate(-90 111 104)">AISLE 10B</text>

        <rect class="fixture" x="105" y="16" width="12" height="4" rx="0.5" fill="rgba(99, 102, 241, 0.25)" />
        <text x="111" y="19.2" class="text-primary" font-size="2.5" text-anchor="middle">E7</text>
        <rect class="fixture" x="105" y="124" width="12" height="4" rx="0.5" fill="rgba(99, 102, 241, 0.25)" />
        <text x="111" y="127.2" class="text-primary" font-size="2.5" text-anchor="middle">E8</text>

        <rect class="fixture" x="148" y="25" width="4" height="12" rx="0.5" />
        <text x="150" y="32.5" class="text-primary" font-size="2.8" text-anchor="middle">R1</text>

        <rect class="fixture" x="148" y="45" width="4" height="12" rx="0.5" />
        <text x="150" y="52.5" class="text-primary" font-size="2.8" text-anchor="middle">R2</text>

        <rect class="fixture" x="148" y="65" width="4" height="12" rx="0.5" />
        <text x="150" y="72.5" class="text-primary" font-size="2.8" text-anchor="middle">R3</text>

        <rect class="fixture" x="148" y="85" width="4" height="12" rx="0.5" />
        <text x="150" y="92.5" class="text-primary" font-size="2.8" text-anchor="middle">R4</text>

        <rect class="fixture" x="135" y="112" width="30" height="10" rx="1" />
        <text x="150" y="118.5" class="text-secondary" font-size="3" text-anchor="middle">HELP DESK</text>

        <rect class="fixture" x="173" y="112" width="22" height="33" rx="1" />
        <text x="184" y="130" class="text-secondary" font-size="3.2" text-anchor="middle">CASH VAULT</text>
        """
        
    if x is not None and y is not None:
        svg_content += f"""
        <circle class="beacon" cx="{x}" cy="{y}" r="2" />
        <circle cx="{x}" cy="{y}" r="4" fill="none" stroke="#f87171" stroke-width="0.5" opacity="0.8">
          <animate attributeName="r" values="2;7;2" dur="2s" repeatCount="indefinite" />
          <animate attributeName="opacity" values="0.8;0;0.8" dur="2s" repeatCount="indefinite" />
        </circle>
        """
        
    svg_content += "</svg>"
    return svg_content

def normalize_usage_hint(hint: str) -> str:
    hint = str(hint).lower()
    if hint in ["h1", "h2", "h3", "body", "caption"]:
        return hint
    if hint == "primary":
        return "body"
    if hint == "secondary":
        return "caption"
    return "body"

def process_and_inline_blueprints(card_data: Any):
    """Recursively scans and replaces any blueprint canvas Image URLs with self-contained Base64 Data URIs."""
    if isinstance(card_data, list):
        for item in card_data:
            process_and_inline_blueprints(item)
        return

    if not isinstance(card_data, dict):
        return

    # If this is a surfaceUpdate containing components
    if "surfaceUpdate" in card_data and isinstance(card_data["surfaceUpdate"], dict):
        components = card_data["surfaceUpdate"].get("components", [])
        if isinstance(components, list):
            for comp in components:
                if not isinstance(comp, dict):
                    continue
                raw_comp = comp.get("component", {})
                if not isinstance(raw_comp, dict):
                    continue
                # If this component is an Image
                if "Image" in raw_comp and isinstance(raw_comp["Image"], dict):
                    url_obj = raw_comp["Image"].get("url", {})
                    if isinstance(url_obj, dict):
                        url_str = url_obj.get("literalString", "")
                        if "/api/v1/blueprint" in url_str:
                            try:
                                # Parse query params to extract layout, x, y
                                from urllib.parse import urlparse, parse_qs
                                parsed = urlparse(url_str)
                                params = parse_qs(parsed.query)
                                
                                layout = params.get("layout", ["linear"])[0]
                                x_val = params.get("x")
                                y_val = params.get("y")
                                
                                x = float(x_val[0]) if x_val else None
                                y = float(y_val[0]) if y_val else None
                                
                                # Generate SVG directly in python
                                svg_str = generate_blueprint_svg(layout, x, y)
                                import base64
                                encoded = base64.b64encode(svg_str.encode('utf-8')).decode('utf-8')
                                inline_url = f"data:image/svg+xml;base64,{encoded}"
                                
                                # Replace with inline Data URI
                                url_obj["literalString"] = inline_url
                                print(f"[A2UI Monkeypatch] Inlined blueprint SVG into flat transaction component {comp.get('id')}! Layout: {layout}, x={x}, y={y}")
                            except Exception as ex:
                                print(f"[A2UI Monkeypatch] Failed to inline blueprint SVG: {ex}")

    # Also recursively scan any other sub-dicts or lists just in case
    for k, v in card_data.items():
        if isinstance(v, (dict, list)):
            process_and_inline_blueprints(v)

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
                raw_hint = node.get("usageHint", node.get("style", "body"))
                properties["usageHint"] = normalize_usage_hint(raw_hint)
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
                            "usageHint": "caption"
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
            
            svg_data = generate_blueprint_svg(layout, x, y)
            import base64
            encoded_svg = base64.b64encode(svg_data.encode('utf-8')).decode('utf-8')
            img_url = f"data:image/svg+xml;base64,{encoded_svg}"
                
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
CACHED_A2UI_CARDS = {}


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
                    token = next(t for t in cache_tokens if t in text)
                    cached_cards = getattr(builtins, "CACHED_A2UI_CARDS", {})
                    json_str = cached_cards.get(token)
                    if json_str:
                        print(f"[A2UI Monkeypatch] Found cached card token {token}. Retrieving cached JSON card from builtins.CACHED_A2UI_CARDS...")
                        
                        # Strip markdown code fences if present in cached_card to prevent json.loads failure
                        json_match_cached = re.search(r'```(?:json)?\s*(.*?)\s*```', json_str, re.DOTALL | re.IGNORECASE)
                        if json_match_cached:
                            json_str = json_match_cached.group(1).strip()
                        else:
                            json_str = json_str.strip()

                        idx = text.find(token)
                        text_prefix = text[:idx].strip()
                    else:
                        print(f"[A2UI Monkeypatch] Error: Cached token {token} found but not present in builtins.CACHED_A2UI_CARDS!")

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
                            # Recursively scan and inline all blueprint canvas URLs as self-contained Base64 Data URIs
                            process_and_inline_blueprints(card_data)
                            
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
