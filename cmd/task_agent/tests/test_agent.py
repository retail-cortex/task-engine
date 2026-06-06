import sys
from unittest import mock

# Mock the entire google.adk package structure to ensure the test runs 100% hermetically
# without requiring connectivity or credentials to the private google-adk Artifact Registry.
mock_adk = mock.MagicMock()
# Register dummy sys.modules for all third-party dependencies to keep test hermetic
sys.modules['google'] = mock_adk
sys.modules['google.genai'] = mock_adk
sys.modules['google.adk'] = mock_adk
sys.modules['google.adk.agents'] = mock_adk
sys.modules['google.adk.agents.readonly_context'] = mock_adk
sys.modules['google.adk.tools'] = mock_adk
sys.modules['google.adk.tools.base_tool'] = mock_adk
sys.modules['google.adk.tools.base_toolset'] = mock_adk
sys.modules['google.adk.tools.tool_context'] = mock_adk
sys.modules['google.adk.tools.mcp_tool'] = mock_adk
sys.modules['google.adk.tools.mcp_tool.mcp_toolset'] = mock_adk
sys.modules['google.adk.tools.mcp_tool.mcp_session_manager'] = mock_adk

mock_httpx = mock.MagicMock()
sys.modules['httpx'] = mock_httpx

# Now import the target agent under test
from agents.task_agent import root_agent

def test_agent_structure():
    # Since LlmAgent is a mock, root_agent is the return value of LlmAgent(...) call.
    from google.adk.agents import LlmAgent
    from google.adk.tools.mcp_tool.mcp_toolset import McpToolset
    from google.adk.tools.mcp_tool.mcp_session_manager import StreamableHTTPConnectionParams

    # Verify LlmAgent was instantiated
    LlmAgent.assert_called_once()
    
    # Extract constructor arguments
    _, kwargs = LlmAgent.call_args
    
    assert kwargs["name"] == "Gemini_Task_Agent"
    assert "OPERATIONAL PROTOCOL" in kwargs["instruction"]
    assert "get_user_context" in kwargs["instruction"]
    assert "query_sop" in kwargs["instruction"]
    assert "STORE SPATIAL BLUEPRINT MAP" in kwargs["instruction"]
    assert "canvas" in kwargs["instruction"]
    assert "boutique" in kwargs["instruction"]
    assert "racetrack" in kwargs["instruction"]
    assert "Active store layout context" in kwargs["instruction"]
    assert "Store Layout Style" in kwargs["instruction"]
    assert "LOCATION_ACKNOWLEDGE" in kwargs["instruction"]
    
    # Verify tools configuration
    tools = kwargs["tools"]
    assert len(tools) == 1
    
    # Verify McpToolset was instantiated inside tools list
    McpToolset.assert_called_once()
    _, tool_kwargs = McpToolset.call_args
    
    # Verify connection params passed to McpToolset
    params = tool_kwargs["connection_params"]
    assert params is not None
    
    # Verify StreamableHTTPConnectionParams was instantiated with correct url
    StreamableHTTPConnectionParams.assert_called_once()
    _, params_kwargs = StreamableHTTPConnectionParams.call_args
    assert "api/v1/mcp" in params_kwargs["url"]
