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
from unittest import TestCase, mock

# 1. Create distinct mocks for each module/class to prevent mock contamination
mock_llm_agent = mock.MagicMock()
mock_base_tool = mock.MagicMock()
mock_base_toolset = mock.MagicMock()

# Create individual module mocks
mock_genai_types = mock.MagicMock()
mock_readonly_context = mock.MagicMock()
mock_tool_context = mock.MagicMock()

# Setup sys.modules
sys.modules['google'] = mock.MagicMock()
sys.modules['google.genai'] = mock_genai_types
sys.modules['google.genai.types'] = mock_genai_types

adk_agents_module = mock.MagicMock()
adk_agents_module.LlmAgent = mock_llm_agent
sys.modules['google.adk.agents'] = adk_agents_module
sys.modules['google.adk.agents.readonly_context'] = mock_readonly_context

class DummyBaseTool:
    def __init__(self, name=None, description=None, *args, **kwargs):
        self.name = name
        self.description = description

class DummyBaseToolset:
    def __init__(self, *args, **kwargs):
        pass

adk_tools_base_tool = mock.MagicMock()
adk_tools_base_tool.BaseTool = DummyBaseTool
sys.modules['google.adk.tools.base_tool'] = adk_tools_base_tool

adk_tools_base_toolset = mock.MagicMock()
adk_tools_base_toolset.BaseToolset = DummyBaseToolset
sys.modules['google.adk.tools.base_toolset'] = adk_tools_base_toolset

sys.modules['google.adk.tools.tool_context'] = mock_tool_context

mock_httpx = mock.MagicMock()
sys.modules['httpx'] = mock_httpx

# Import the agent under test
from agents.task_agent import root_agent


class TestTaskAgentStructure(TestCase):
    def test_agent_initialization(self):
        # Verify LlmAgent constructor was called exactly once
        mock_llm_agent.assert_called_once()
        
        # Extract constructor arguments
        _, kwargs = mock_llm_agent.call_args
        
        # Verify core metadata
        self.assertEqual(kwargs["name"], "Gemini_Task_Agent")
        self.assertEqual(kwargs["model"], "gemini-2.5-flash")
        self.assertIn("retail operations coordinator", kwargs["description"])
        
        # Verify key parts of the system instruction are present
        instruction = kwargs["instruction"]
        self.assertIn("OPERATIONAL PROTOCOL", instruction)
        self.assertIn("get_user_context", instruction)
        self.assertIn("query_sop", instruction)
        self.assertIn("get_store_selector", instruction)
        self.assertIn("A2UI CARD OUTPUT PROTOCOL", instruction)
        
        # Verify tools configuration
        tools = kwargs["tools"]
        self.assertEqual(len(tools), 1)
        
        # Verify that the toolset is an instance of StatelessMcpToolset
        from agents.task_agent.task_agent import StatelessMcpToolset
        self.assertIsInstance(tools[0], StatelessMcpToolset)
        
        # Verify callback registration
        from agents.task_agent.task_agent import strip_tool_namespaces_callback
        self.assertEqual(kwargs["after_model_callback"], strip_tool_namespaces_callback)
