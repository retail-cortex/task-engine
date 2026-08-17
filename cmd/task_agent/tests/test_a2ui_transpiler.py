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

import sys
from unittest import mock

# Mock google-adk converter dependency
mock_pc = mock.MagicMock()
mock_ec = mock.MagicMock()
sys.modules['google.adk.a2a.converters.part_converter'] = mock_pc
sys.modules['google.adk.a2a.converters.event_converter'] = mock_ec

# Mock a2a dependencies
mock_a2a = mock.MagicMock()

class MockTextPart:
    def __init__(self, text):
        self.text = text

class MockDataPart:
    def __init__(self, data, metadata=None):
        self.data = data
        self.metadata = metadata

class MockPart:
    def __init__(self, root):
        self.root = root

mock_a2a.TextPart = MockTextPart
mock_a2a.DataPart = MockDataPart
mock_a2a.Part = MockPart

sys.modules['a2a'] = mock_a2a
sys.modules['a2a.types'] = mock_a2a

# Ensure google.genai is mocked if not present in the runtime to keep tests hermetic
try:
    from google.genai import types as genai_types
except ImportError:
    mock_genai = mock.MagicMock()
    sys.modules['google'] = mock_genai
    sys.modules['google.genai'] = mock_genai
    from google.genai import types as genai_types

from a2a import types as a2a_types
from a2ui_transpiler import normalize_card_to_a2ui_messages, custom_convert_genai_part_to_a2a_part

def test_normalize_card_to_a2ui_messages_switcher():
    card = {
        "type": "card",
        "title": "RETAIL STOREFRONT CONTEXT SWITCHER",
        "style": "primary",
        "children": [
            {
                "type": "column",
                "alignment": "stretch",
                "children": [
                    {
                        "type": "button",
                        "label": "Volt & Vine - Seattle",
                        "action": "SET_STORE",
                        "actionData": {
                            "siteID": "44444444-4444-4444-4444-444444440000",
                            "siteLabel": "Volt & Vine - Seattle"
                        }
                    }
                ]
            }
        ]
    }
    
    messages = normalize_card_to_a2ui_messages(card)
    assert len(messages) == 2
    assert "beginRendering" in messages[0]
    assert "surfaceUpdate" in messages[1]
    
    components = messages[1]["surfaceUpdate"]["components"]
    # We expect text_2 (title), text_5 (button label), button_4, column_3, card_1
    component_types = [list(c["component"].keys())[0] for c in components]
    assert "Card" in component_types
    assert "Column" in component_types
    assert "Button" in component_types
    assert "Text" in component_types

def test_normalize_card_to_a2ui_messages_table():
    card = {
        "type": "card",
        "title": "SITE OPERATIONAL TASK QUEUE",
        "children": [
            {
                "type": "table",
                "rows": [
                    {"label": "Register Drop", "value": "PENDING"}
                ]
            }
        ]
    }
    
    messages = normalize_card_to_a2ui_messages(card)
    assert len(messages) == 2
    components = messages[1]["surfaceUpdate"]["components"]
    
    # Check that Column and Row were generated for the table transpilation fallback
    component_types = [list(c["component"].keys())[0] for c in components]
    assert "Row" in component_types
    assert "Column" in component_types

def test_normalize_card_to_a2ui_messages_canvas():
    card = {
        "type": "card",
        "children": [
            {
                "type": "canvas",
                "layout": "racetrack",
                "beacon": {
                    "x": 10,
                    "y": 20,
                    "name": "Beacon A"
                }
            }
        ]
    }
    
    messages = normalize_card_to_a2ui_messages(card)
    assert len(messages) == 2
    components = messages[1]["surfaceUpdate"]["components"]
    
    # Check that Image was generated for the canvas transpilation
    component_types = [list(c["component"].keys())[0] for c in components]
    assert "Image" in component_types

def test_custom_convert_genai_part_to_a2a_part_non_json():
    part = genai_types.Part(text="Hello, this is standard text.")
    res = custom_convert_genai_part_to_a2a_part(part)
    assert res is not None
    assert isinstance(res.root, a2a_types.TextPart)
    assert res.root.text == "Hello, this is standard text."

def test_custom_convert_event_to_a2a_message_with_card():
    from unittest.mock import MagicMock
    from a2ui_transpiler import custom_convert_event_to_a2a_message
    
    card_json = """
    ```json
    {
      "type": "card",
      "title": "TEST CARD",
      "children": []
    }
    ```
    """
    
    mock_part = MagicMock()
    mock_part.text = card_json
    
    mock_event = MagicMock()
    mock_event.content.parts = [mock_part]
    
    mock_context = MagicMock()
    
    res = custom_convert_event_to_a2a_message(mock_event, mock_context)
    
    assert res is not None
    assert len(res.parts) == 2
    assert isinstance(res.parts[0].root, a2a_types.DataPart)
    assert res.parts[0].root.data == {
        "beginRendering": {
            "root": "card_1",
            "surfaceId": "@default",
            "styles": {}
        }
    }
    assert res.parts[0].root.metadata == {"mimeType": "application/json+a2ui"}
    
    assert isinstance(res.parts[1].root, a2a_types.DataPart)
    assert "surfaceUpdate" in res.parts[1].root.data
    assert res.parts[1].root.metadata == {"mimeType": "application/json+a2ui"}

def test_custom_convert_event_to_a2a_message_with_card_uppercase_json():
    from unittest.mock import MagicMock
    from a2ui_transpiler import custom_convert_event_to_a2a_message
    
    card_json = """
    ```JSON
    {
      "type": "card",
      "title": "TEST CARD",
      "children": []
    }
    ```
    """
    
    mock_part = MagicMock()
    mock_part.text = card_json
    
    mock_event = MagicMock()
    mock_event.content.parts = [mock_part]
    
    mock_context = MagicMock()
    
    res = custom_convert_event_to_a2a_message(mock_event, mock_context)
    
    assert res is not None
    assert len(res.parts) == 2
    assert isinstance(res.parts[0].root, a2a_types.DataPart)
    assert res.parts[0].root.data["beginRendering"]["root"] == "card_1"

def test_normalize_card_to_a2ui_messages_dropdown():
    from a2ui_transpiler import normalize_card_to_a2ui_messages
    card = {
        "type": "card",
        "title": "RETAIL STOREFRONT CONTEXT SWITCHER",
        "style": "primary",
        "children": [
            {
                "type": "button",
                "label": "Volt & Vine - Seattle",
                "action": "SET_STORE",
                "actionData": {
                    "siteID": "44444444-4444-4444-4444-444444440000"
                }
            },
            {
                "type": "button",
                "label": "Volt & Vine - San Francisco",
                "action": "SET_STORE",
                "actionData": {
                    "siteID": "55555555-5555-5555-5555-555555550000"
                }
            }
        ]
    }
    
    messages = normalize_card_to_a2ui_messages(card)
    assert len(messages) == 2
    
    components = messages[1]["surfaceUpdate"]["components"]
    
    # We expect: MultipleChoice (dropdown) and Button (Switch Store)
    choice = next(c for c in components if "MultipleChoice" in c["component"])
    mc = choice["component"]["MultipleChoice"]
    assert len(mc["options"]) == 2
    assert mc["options"][0]["label"]["literalString"] == "Volt & Vine - Seattle"
    assert mc["options"][0]["value"] == "44444444-4444-4444-4444-444444440000"
    assert mc["selections"]["path"] == "/store_switcher/selected"
    
    button = next(c for c in components if "Button" in c["component"])
    btn = button["component"]["Button"]
    assert btn["action"]["name"] == "SET_STORE"
    assert btn["action"]["context"][0]["key"] == "siteID"
    assert btn["action"]["context"][0]["value"]["path"] == "/store_switcher/selected"

def test_normalize_card_to_a2ui_messages_v08_layout():
    from a2ui_transpiler import normalize_card_to_a2ui_messages
    card = {
        "type": "card",
        "title": "TEST V0.8 CARD",
        "children": [
            {
                "type": "column",
                "children": [
                    {
                        "type": "text",
                        "content": "Line 1"
                    },
                    {
                        "type": "text",
                        "content": "Line 2"
                    }
                ]
            }
        ]
    }
    
    messages = normalize_card_to_a2ui_messages(card)
    components = messages[1]["surfaceUpdate"]["components"]
    
    # Verify Card uses singular 'child' string property in v0.8
    card_comp = next(c for c in components if "Card" in c["component"])["component"]["Card"]
    assert "child" in card_comp
    assert "children" not in card_comp
    assert isinstance(card_comp["child"], str)
    
    # Verify Column uses 'children.explicitList' layout structure in v0.8
    col_comp = next(c for c in components if "Column" in c["component"])["component"]["Column"]
    assert "children" in col_comp
    assert isinstance(col_comp["children"], dict)
    assert "explicitList" in col_comp["children"]
    assert len(col_comp["children"]["explicitList"]) == 2

def test_normalize_card_to_a2ui_messages_dropdown_nested():
    from a2ui_transpiler import normalize_card_to_a2ui_messages
    card = {
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
                        "label": "Volt & Vine - Seattle",
                        "action": "SET_STORE",
                        "actionData": {
                            "siteID": "44444444-4444-4444-4444-444444440000"
                        }
                    },
                    {
                        "type": "button",
                        "label": "Volt & Vine - San Francisco",
                        "action": "SET_STORE",
                        "actionData": {
                            "siteID": "55555555-5555-5555-5555-555555550000"
                        }
                    }
                ]
            }
        ]
    }
    
    messages = normalize_card_to_a2ui_messages(card)
    assert len(messages) == 2
    
    components = messages[1]["surfaceUpdate"]["components"]
    
    # Check that Column has the collapsed MultipleChoice and Button as children
    column = next(c for c in components if "Column" in c["component"])
    children = column["component"]["Column"]["children"]["explicitList"]
    assert len(children) == 2
    
    choice_id = children[0]
    button_id = children[1]
    
    choice = next(c for c in components if c["id"] == choice_id)
    assert "MultipleChoice" in choice["component"]
    
    btn = next(c for c in components if c["id"] == button_id)
    assert "Button" in btn["component"]


def test_custom_convert_event_to_a2a_message_with_cached_card_markdown():
    from unittest.mock import MagicMock
    from a2ui_transpiler import custom_convert_event_to_a2a_message
    import builtins

    # Seed the cache with markdown-fenced card payload, mimicking Go MCP return format
    builtins.CACHED_A2UI_CARD = """```json
    {
      "type": "card",
      "title": "TEST CACHED CARD",
      "children": []
    }
    ```"""

    mock_part = MagicMock()
    mock_part.text = "Here are the details: [A2UI_CARD_TASK_DETAILS_CACHED]"

    mock_event = MagicMock()
    mock_event.content.parts = [mock_part]

    mock_context = MagicMock()

    res = custom_convert_event_to_a2a_message(mock_event, mock_context)

    # Clean up cache
    if hasattr(builtins, "CACHED_A2UI_CARD"):
        delattr(builtins, "CACHED_A2UI_CARD")

    assert res is not None
    assert len(res.parts) == 2
    assert isinstance(res.parts[0].root, a2a_types.DataPart)
    
    # Verify the cached card was parsed and normalized successfully
    assert res.parts[0].root.data["beginRendering"]["root"] == "card_1"
    assert res.parts[0].root.metadata == {"mimeType": "application/json+a2ui"}
    
    assert isinstance(res.parts[1].root, a2a_types.DataPart)
    assert "surfaceUpdate" in res.parts[1].root.data
    assert res.parts[1].root.metadata == {"mimeType": "application/json+a2ui"}


def test_normalize_card_to_a2ui_messages_checkbox():
    card = {
        "type": "card",
        "title": "CHECKBOX TEST",
        "children": [
            {
                "type": "checkbox",
                "label": "I agree to the terms",
                "value": "/form/agreedToTerms"
            }
        ]
    }
    
    messages = normalize_card_to_a2ui_messages(card)
    assert len(messages) == 2
    components = messages[1]["surfaceUpdate"]["components"]
    
    checkbox = next(c for c in components if "CheckBox" in c["component"])
    assert checkbox["component"]["CheckBox"]["label"] == {"literalString": "I agree to the terms"}
    assert checkbox["component"]["CheckBox"]["value"] == {"path": "/form/agreedToTerms"}


def test_normalize_card_to_a2ui_messages_webframe():
    card = {
        "type": "card",
        "title": "WEBFRAME TEST",
        "children": [
            {
                "type": "webframe",
                "htmlContent": "<html><body><h1>Hello World</h1></body></html>",
                "height": 450
            }
        ]
    }
    
    messages = normalize_card_to_a2ui_messages(card)
    assert len(messages) == 2
    components = messages[1]["surfaceUpdate"]["components"]
    
    webframe = next(c for c in components if "WebFrameSrcdoc" in c["component"])
    assert webframe["component"]["WebFrameSrcdoc"]["htmlContent"] == {
        "literalString": "<html><body><h1>Hello World</h1></body></html>"
    }
    assert webframe["component"]["WebFrameSrcdoc"]["height"] == 450


def test_normalize_card_to_a2ui_messages_webframe_hybrid():
    card = {
        "type": "card",
        "title": "HYBRID TEST",
        "children": [
            {
                "type": "webframe",
                "htmlContent": "<html><body><h1>Hello World</h1></body></html>",
                "height": 450
            },
            {
                "type": "button",
                "label": "Start Step",
                "action": "UPDATE_CHECKLIST",
                "actionData": {
                    "execution_id": "123",
                    "status": "IN_PROGRESS"
                }
            }
        ]
    }
    
    messages = normalize_card_to_a2ui_messages(card)
    assert len(messages) == 2
    components = messages[1]["surfaceUpdate"]["components"]
    
    card_comp = next(c for c in components if "Card" in c["component"])
    root_id = card_comp["component"]["Card"]["child"]
    
    column_comp = next(c for c in components if c["id"] == root_id)
    assert "Column" in column_comp["component"]
    children_list = column_comp["component"]["Column"]["children"]["explicitList"]
    
    # The children list should contain the title text, the WebFrame, and the button
    assert len(children_list) == 3
    
    webframe = next(c for c in components if "WebFrameSrcdoc" in c["component"])
    button = next(c for c in components if "Button" in c["component"])
    
    assert webframe["id"] in children_list
    assert button["id"] in children_list
    assert button["component"]["Button"]["label"] == "Start Step"
