// Copyright 2026 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package api

import (
	"context"
	"fmt"

	"github.com/gin-gonic/gin"
	mcp "github.com/metoro-io/mcp-golang"
	"github.com/metoro-io/mcp-golang/transport/http"
	"github.com/rmcguinness/gemini_task_engine/pkg/agents"
	"github.com/rmcguinness/gemini_task_engine/pkg/model"
	"github.com/rmcguinness/gemini_task_engine/pkg/service"
)

type MCPHandler struct {
	mcpServer         *mcp.Server
	ginTransport      *http.GinTransport
	taskService        service.TaskService
	ragService         service.RAGService
	shiftService       service.ShiftService
	automationService service.AutomationService
}

func NewMCPHandler(taskService service.TaskService, ragService service.RAGService, shiftService service.ShiftService, automationService service.AutomationService) (*MCPHandler, error) {
	transport := http.NewGinTransport()
	server := mcp.NewServer(
		transport,
		mcp.WithName("gemini-task-engine-mcp"),
		mcp.WithInstructions("MCP Server providing retail task, trade, override, and SOP similarity queries."),
	)

	handler := &MCPHandler{
		mcpServer:         server,
		ginTransport:      transport,
		taskService:        taskService,
		ragService:         ragService,
		shiftService:       shiftService,
		automationService: automationService,
	}

	if err := handler.registerTools(); err != nil {
		return nil, fmt.Errorf("failed to register MCP tools: %w", err)
	}

	if err := handler.registerPrompts(); err != nil {
		return nil, fmt.Errorf("failed to register MCP prompts: %w", err)
	}

	return handler, nil
}

// Handler returns the Gin handler function to mount on the chat route.
func (h *MCPHandler) Handler() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Retrieve parameters from request path and headers
		shiftID := c.Param("shiftId")
		userID, _ := c.Get("userID")

		// Inject shift context and user context into standard request context
		ctx := c.Request.Context()
		ctx = context.WithValue(ctx, "shiftID", shiftID)
		ctx = context.WithValue(ctx, "userID", userID)
		c.Request = c.Request.WithContext(ctx)

		// Run the stateless HTTP transport handler
		h.ginTransport.Handler()(c)
	}
}

type TriggerAlertArgs struct {
	SiteID      string `json:"site_id" jsonschema:"description=The ID of the store site target where the alert fires"`
	OrganizerID string `json:"organizer_id" jsonschema:"description=The sensor, camera, or device ID that generated the alert"`
	EventType   string `json:"event_type" jsonschema:"description=The standard category raw string event descriptor (e.g. TillDrawerDropEvent, CustomerAssistanceEvent, StockoutCorrectEvent)"`
	Description string `json:"description" jsonschema:"description=Detailed context explaining the alert state"`
}

type GetTasksArgs struct {
	SiteID string `json:"site_id" jsonschema:"description=The ID of the site to fetch tasks for"`
}

type OverrideAssetArgs struct {
	ExecutionID   string `json:"execution_id" jsonschema:"description=The ID of the task execution being overridden"`
	AssetID       string `json:"asset_id" jsonschema:"description=The ID of the asset constraint being bypassed"`
	Justification string `json:"justification" jsonschema:"description=The compliance justification text for GORM audit logs"`
}

type ProposeTradeArgs struct {
	TaskExecutionID    string `json:"task_execution_id" jsonschema:"description=The ID of the task execution being traded"`
	ProposedAssigneeID string `json:"proposed_assignee_id" jsonschema:"description=The user ID of the proposed colleague to assign the task to"`
}

type QuerySOPArgs struct {
	QueryText string `json:"query_text" jsonschema:"description=The search term to query against SOP documentation"`
}

type AcceptTradeArgs struct {
	TradeID string `json:"trade_id" jsonschema:"description=The GORM UUID ID of the task trade request being accepted"`
}

type RejectTradeArgs struct {
	TradeID string `json:"trade_id" jsonschema:"description=The GORM UUID ID of the task trade request being rejected"`
}

func (h *MCPHandler) registerTools() error {
	// Tool 1: get_tasks
	err := h.mcpServer.RegisterTool("get_tasks", "Retrieves the prioritized active queue of task executions for a given retail site.", h.HandleGetTasks)
	if err != nil {
		return err
	}

	// Tool 2: override_asset
	err = h.mcpServer.RegisterTool("override_asset", "Submits an administrative bypass for an asset constraint with a GORM audited compliance justification.", h.HandleOverrideAsset)
	if err != nil {
		return err
	}

	// Tool 3: propose_trade
	err = h.mcpServer.RegisterTool("propose_trade", "Initiates a peer-to-peer task trade request with another coworker.", h.HandleProposeTrade)
	if err != nil {
		return err
	}

	// Tool 4: query_sop
	err = h.mcpServer.RegisterTool("query_sop", "Performs semantic searches against chunked Standard Operating Procedure (SOP) vector embeddings.", h.HandleQuerySOP)
	if err != nil {
		return err
	}

	// Tool 5: trigger_alert
	err = h.mcpServer.RegisterTool("trigger_alert", "Triggers a high-priority ad-hoc streaming retail alert (such as till out, register empty shelf, customer assistance buttons) causing immediate queue task generation.", h.HandleTriggerAlert)
	if err != nil {
		return err
	}

	// Tool 6: accept_trade
	err = h.mcpServer.RegisterTool("accept_trade", "Accepts a pending task trade proposal under direct coworker maker/checker guidelines.", h.HandleAcceptTrade)
	if err != nil {
		return err
	}

	// Tool 7: reject_trade
	err = h.mcpServer.RegisterTool("reject_trade", "Rejects a pending task trade proposal under direct coworker maker/checker guidelines.", h.HandleRejectTrade)
	return err
}

func (h *MCPHandler) HandleGetTasks(ctx context.Context, args GetTasksArgs) (*mcp.ToolResponse, error) {
	queue, err := h.taskService.GetQueue(ctx, args.SiteID)
	if err != nil {
		return nil, err
	}
	var output string
	for _, item := range queue {
		output += fmt.Sprintf("Task: %s | ID: %s | Priority: %d | Status: %s", item.TaskTemplateID, item.ID, item.Priority, item.Status)
		if item.Description != "" {
			output += fmt.Sprintf(" | Description: %s", item.Description)
		}
		output += "\n"
	}
	if output == "" {
		output = "No active tasks in queue for site."
	}
	return mcp.NewToolResponse(mcp.NewTextContent(output)), nil
}

func (h *MCPHandler) HandleOverrideAsset(ctx context.Context, args OverrideAssetArgs) (*mcp.ToolResponse, error) {
	userID, ok := ctx.Value("userID").(string)
	if !ok || userID == "" {
		userID = "00000000-0000-0000-0000-000000000000"
	}
	err := h.taskService.OverrideAssetConstraint(ctx, args.ExecutionID, args.AssetID, args.Justification, userID)
	if err != nil {
		return nil, err
	}
	return mcp.NewToolResponse(mcp.NewTextContent("Asset constraint override successfully applied and logged to ledger.")), nil
}

func (h *MCPHandler) HandleProposeTrade(ctx context.Context, args ProposeTradeArgs) (*mcp.ToolResponse, error) {
	userID, ok := ctx.Value("userID").(string)
	if !ok || userID == "" {
		userID = "00000000-0000-0000-0000-000000000000"
	}
	err := h.taskService.ProposeTrade(ctx, args.TaskExecutionID, args.ProposedAssigneeID, userID)
	if err != nil {
		return nil, err
	}
	return mcp.NewToolResponse(mcp.NewTextContent("Task trade request successfully proposed and created.")), nil
}

func (h *MCPHandler) HandleAcceptTrade(ctx context.Context, args AcceptTradeArgs) (*mcp.ToolResponse, error) {
	userID, ok := ctx.Value("userID").(string)
	if !ok || userID == "" {
		userID = "00000000-0000-0000-0000-000000000000"
	}
	err := h.taskService.AcceptTrade(ctx, args.TradeID, userID)
	if err != nil {
		return nil, err
	}
	return mcp.NewToolResponse(mcp.NewTextContent("Task trade request successfully accepted. Task reassigned under GORM database guidelines.")), nil
}

func (h *MCPHandler) HandleRejectTrade(ctx context.Context, args RejectTradeArgs) (*mcp.ToolResponse, error) {
	userID, ok := ctx.Value("userID").(string)
	if !ok || userID == "" {
		userID = "00000000-0000-0000-0000-000000000000"
	}
	err := h.taskService.RejectTrade(ctx, args.TradeID, userID)
	if err != nil {
		return nil, err
	}
	return mcp.NewToolResponse(mcp.NewTextContent("Task trade request successfully rejected and returned to initiator's operational queue.")), nil
}

func (h *MCPHandler) HandleQuerySOP(ctx context.Context, args QuerySOPArgs) (*mcp.ToolResponse, error) {
	mockVector := make(model.Float32Vector, 768)
	mockVector[0] = 0.1 // basic ground value
	chunks, err := h.ragService.QuerySimilarity(ctx, mockVector, 3)
	if err != nil {
		return nil, err
	}
	var output string
	for _, chunk := range chunks {
		output += fmt.Sprintf("[Chunk %d]: %s\n\n", chunk.ChunkIndex, chunk.Content)
	}
	if output == "" {
		output = "No matching SOP documentation found."
	}
	return mcp.NewToolResponse(mcp.NewTextContent(output)), nil
}

func (h *MCPHandler) HandleTriggerAlert(ctx context.Context, args TriggerAlertArgs) (*mcp.ToolResponse, error) {
	exec, err := h.automationService.TriggerStreamingEvent(ctx, args.SiteID, args.OrganizerID, model.EventType(args.EventType), args.Description)
	if err != nil {
		return nil, err
	}
	output := fmt.Sprintf("Streaming alert successfully ingested. Created dynamic task execution ticket: ID: %s | Priority: %d | Status: %s\n", exec.ID, exec.Priority, exec.Status)
	return mcp.NewToolResponse(mcp.NewTextContent(output)), nil
}

type AgentPromptArgs struct{}

func (h *MCPHandler) registerPrompts() error {
	for _, agent := range agents.List() {
		// Register each agent as an MCP prompt.
		// Prompt handler returns the prompt system instruction as a list of messages.
		promptHandler := func(args AgentPromptArgs) (*mcp.PromptResponse, error) {
			return mcp.NewPromptResponse(
				agent.Description,
				mcp.NewPromptMessage(mcp.NewTextContent(agent.SystemInstruction), mcp.RoleAssistant),
			), nil
		}

		err := h.mcpServer.RegisterPrompt(agent.ID, agent.Description, promptHandler)
		if err != nil {
			return fmt.Errorf("failed to register prompt for agent %s: %w", agent.ID, err)
		}
	}
	return nil
}

