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
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/google/uuid"
	"github.com/gin-gonic/gin"
	mcp "github.com/metoro-io/mcp-golang"
	mcphttp "github.com/metoro-io/mcp-golang/transport/http"
	"github.com/rmcguinness/gemini_task_engine/pkg/agents"
	"github.com/rmcguinness/gemini_task_engine/pkg/model"
	"github.com/rmcguinness/gemini_task_engine/pkg/service"
)

type MCPHandler struct {
	mcpServer         *mcp.Server
	ginTransport      *mcphttp.GinTransport
	taskService        service.TaskService
	ragService         service.RAGService
	shiftService       service.ShiftService
	automationService service.AutomationService
}

func NewMCPHandler(taskService service.TaskService, ragService service.RAGService, shiftService service.ShiftService, automationService service.AutomationService) (*MCPHandler, error) {
	transport := mcphttp.NewGinTransport()
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

	if err := server.Serve(); err != nil {
		return nil, fmt.Errorf("failed to serve MCP: %w", err)
	}

	return handler, nil
}

type jsonrpcRequest struct {
	Method string          `json:"method"`
	ID     json.RawMessage `json:"id"`
}

type jsonrpcResponse struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      json.RawMessage `json:"id"`
	Result  json.RawMessage `json:"result,omitempty"`
	Error   *jsonrpcError   `json:"error,omitempty"`
}

type jsonrpcError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

type toolCallSuccessResult struct {
	Content []toolContent `json:"content"`
	IsError bool          `json:"isError"`
}

type toolContent struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

type interceptingWriter struct {
	gin.ResponseWriter
	body   *bytes.Buffer
	status int
}

func (w *interceptingWriter) WriteHeader(statusCode int) {
	w.status = statusCode
}

func (w *interceptingWriter) Write(b []byte) (int, error) {
	return w.body.Write(b)
}

func (w *interceptingWriter) WriteString(s string) (int, error) {
	return w.body.WriteString(s)
}

func (w *interceptingWriter) Status() int {
	if w.status == 0 {
		return http.StatusOK
	}
	return w.status
}

// Handler returns the Gin handler function to mount on the chat route.
func (h *MCPHandler) Handler() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Retrieve parameters from request path and headers
		shiftID := c.Param("shiftId")
		if shiftID == "" {
			shiftID = c.GetHeader("X-Shift-ID")
		}
		userID, _ := c.Get("userID")

		// Inject shift context and user context into standard request context
		ctx := c.Request.Context()
		ctx = context.WithValue(ctx, "shiftID", shiftID)
		ctx = context.WithValue(ctx, "userID", userID)
		c.Request = c.Request.WithContext(ctx)

		// 1. Read request body to inspect the JSON-RPC method
		var reqBody []byte
		if c.Request.Body != nil {
			var err error
			reqBody, err = io.ReadAll(c.Request.Body)
			if err == nil {
				// Restore body so SDK can read it
				c.Request.Body = io.NopCloser(bytes.NewBuffer(reqBody))
			}
		}

		// Parse method from request
		var method string
		var rpcReq jsonrpcRequest
		if len(reqBody) > 0 {
			if err := json.Unmarshal(reqBody, &rpcReq); err == nil {
				method = rpcReq.Method
			}
		}

		// 2. Intercept response using a custom writer
		buffer := &bytes.Buffer{}
		originalWriter := c.Writer
		interceptWriter := &interceptingWriter{
			ResponseWriter: originalWriter,
			body:           buffer,
		}
		c.Writer = interceptWriter

		// 3. Call the SDK handler
		h.ginTransport.Handler()(c)

		// Restore original writer
		c.Writer = originalWriter

		// 4. Process and potentially modify response
		respBytes := buffer.Bytes()
		if len(respBytes) > 0 {
			var rpcResp jsonrpcResponse
			if err := json.Unmarshal(respBytes, &rpcResp); err == nil {
				modified := false

				if method == "initialize" && rpcResp.Result != nil {
					var initResp struct {
						Capabilities    map[string]interface{} `json:"capabilities"`
						Instructions    *string                `json:"instructions,omitempty"`
						ProtocolVersion string                 `json:"protocolVersion"`
						ServerInfo      map[string]interface{} `json:"serverInfo"`
					}
					if err := json.Unmarshal(rpcResp.Result, &initResp); err == nil {
						initResp.ProtocolVersion = "2025-11-25"
						newResultBytes, err := json.Marshal(initResp)
						if err == nil {
							rpcResp.Result = newResultBytes
							modified = true
						}
					}
				} else if method == "tools/call" && rpcResp.Error != nil {
					// Translate input validation/unmarshaling errors to isError: true result
					resultVal := toolCallSuccessResult{
						Content: []toolContent{
							{
								Type: "text",
								Text: rpcResp.Error.Message,
							},
						},
						IsError: true,
					}
					if resultBytes, err := json.Marshal(resultVal); err == nil {
						rpcResp.Result = resultBytes
						rpcResp.Error = nil
						modified = true
					}
				}

				if modified {
					if newRespBytes, err := json.Marshal(rpcResp); err == nil {
						respBytes = newRespBytes
					}
				}
			}
		}

		// 5. Write final headers and data to the real response writer
		// GinTransport sets Content-Type to application/json and status code to 200 OK
		originalWriter.Header().Set("Content-Type", "application/json")
		originalWriter.WriteHeader(http.StatusOK)
		_, _ = originalWriter.Write(respBytes)
	}
}

type TriggerAlertArgs struct {
	SiteID      string `json:"site_id" jsonschema:"description=The ID of the store site target where the alert fires"`
	OrganizerID string `json:"organizer_id" jsonschema:"description=The sensor, camera, or device ID that generated the alert"`
	EventType   string `json:"event_type" jsonschema:"description=The standard category raw string event descriptor (e.g. TillDrawerDropEvent, CustomerAssistanceEvent, StockoutCorrectEvent)"`
	Description string `json:"description" jsonschema:"description=Detailed context explaining the alert state"`
}

type GetTasksArgs struct {
	SiteID     string  `json:"site_id" jsonschema:"description=The ID of the site to fetch tasks for"`
	RoleID     *string `json:"role_id,omitempty" jsonschema:"description=Optional role ID to filter tasks by required role"`
	AssigneeID *string `json:"assignee_id,omitempty" jsonschema:"description=Optional user ID to filter tasks by assignee"`
	LocationID *string `json:"location_id,omitempty" jsonschema:"description=Optional location ID to filter tasks by location"`
	Format     *string `json:"format,omitempty" jsonschema:"description=Optional format for the output: 'a2ui' for pre-formatted A2UI cards or 'text' for plain text"`
}

type GetTaskDetailsArgs struct {
	ExecutionID string `json:"execution_id" jsonschema:"description=The GORM UUID of the task execution to fetch details for"`
}

type GetSiteLocationsArgs struct {
	SiteID     string  `json:"site_id" jsonschema:"description=The GORM UUID site ID to fetch locations for"`
	Format     *string `json:"format,omitempty" jsonschema:"description=Optional format for the output: 'a2ui' for pre-formatted A2UI store spatial map cards"`
	LocationID *string `json:"location_id,omitempty" jsonschema:"description=Optional location ID to place a beacon/marker at"`
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

type GetUserContextArgs struct{}

type OrgContextInfo struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type SiteContextInfo struct {
	ID             string `json:"id"`
	Name           string `json:"name"`
	OrganizationID string `json:"organization_id"`
}

type UserContextResponse struct {
	UserID        string            `json:"user_id"`
	Email         string            `json:"email"`
	Name          string            `json:"name"`
	Roles         []string          `json:"roles"`
	Organizations []OrgContextInfo  `json:"organizations"`
	Sites         []SiteContextInfo `json:"sites"`
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
	if err != nil {
		return err
	}

	// Tool 8: get_user_context
	err = h.mcpServer.RegisterTool("get_user_context", "Retrieves the current authenticated user's profile details, roles, assigned organizations, and assigned sites.", h.HandleGetUserContext)
	if err != nil {
		return err
	}

	// Tool 9: claim_task
	err = h.mcpServer.RegisterTool("claim_task", "Claims/assigns a prioritized task execution from the queue to the authenticated user.", h.HandleClaimTask)
	if err != nil {
		return err
	}

	// Tool 10: update_task_status
	err = h.mcpServer.RegisterTool("update_task_status", "Updates the execution status (e.g. IN_PROGRESS, COMPLETED) and optional checklist state of a task.", h.HandleUpdateTaskStatus)
	if err != nil {
		return err
	}

	// Tool 11: list_pending_trades
	err = h.mcpServer.RegisterTool("list_pending_trades", "Lists all active, pending coworker task trade proposals relevant to the authenticated user.", h.HandleListPendingTrades)
	if err != nil {
		return err
	}

	// Tool 12: get_site_locations
	err = h.mcpServer.RegisterTool("get_site_locations", "Lists all locations (fixtures, registers, aisles) under a physical site.", h.HandleGetSiteLocations)
	if err != nil {
		return err
	}

	// Tool 13: get_task_details
	err = h.mcpServer.RegisterTool("get_task_details", "Retrieves the detailed checklist steps and operational state of a specific task execution.", h.HandleGetTaskDetails)
	if err != nil {
		return err
	}

	// Tool 14: get_weather
	err = h.mcpServer.RegisterTool("get_weather", "Query real-time meteorological observations and wind patterns for a regional airport station.", h.HandleGetWeather)
	if err != nil {
		return err
	}

	// Tool 15: get_store_selector
	err = h.mcpServer.RegisterTool("get_store_selector", "Retrieves a RETAIL STOREFRONT CONTEXT SWITCHER card containing buttons to change storefront site context.", h.HandleGetStoreSelector)
	return err
}

func getUserID(ctx context.Context) string {
	var uid string
	if ginCtx, ok := ctx.Value("ginContext").(*gin.Context); ok {
		if val, exists := ginCtx.Get("userID"); exists {
			if s, ok := val.(string); ok {
				uid = s
			}
		}
	}
	if uid == "" {
		if s, ok := ctx.Value("userID").(string); ok {
			uid = s
		}
	}
	if uid != "" {
		uid = strings.TrimPrefix(uid, "A2A_USER_")
		if _, err := uuid.Parse(uid); err == nil {
			return uid
		}
	}
	return ""
}


func (h *MCPHandler) HandleGetTasks(ctx context.Context, args GetTasksArgs) (*mcp.ToolResponse, error) {
	queue, err := h.taskService.GetQueue(ctx, args.SiteID)
	if err != nil {
		return nil, err
	}

	var filteredQueue []*model.TaskExecution
	for _, item := range queue {
		// Filter by RoleID
		if args.RoleID != nil && *args.RoleID != "" {
			if item.Task.TargetRoleID == nil || *item.Task.TargetRoleID != *args.RoleID {
				continue
			}
		}

		// Filter by AssigneeID
		if args.AssigneeID != nil && *args.AssigneeID != "" {
			if item.AssigneeID == nil || *item.AssigneeID != *args.AssigneeID {
				continue
			}
		}

		// Filter by LocationID
		if args.LocationID != nil && *args.LocationID != "" {
			loc, err := h.taskService.GetLocationByID(ctx, *args.LocationID)
			if err != nil || loc == nil {
				continue
			}
			locName := strings.ToLower(loc.Name)
			taskName := strings.ToLower(item.Task.Name)
			taskDesc := strings.ToLower(item.Description)
			matched := false
			if strings.Contains(taskName, locName) || strings.Contains(taskDesc, locName) {
				matched = true
			} else {
				terms := strings.Fields(locName)
				for _, term := range terms {
					if len(term) <= 2 || term == "volt" || term == "vine" || term == "omnimart" {
						continue
					}
					if strings.Contains(taskName, term) || strings.Contains(taskDesc, term) {
						matched = true
						break
					}
				}
			}
			if !matched {
				continue
			}
		}

		filteredQueue = append(filteredQueue, item)
	}

	// If the client explicitly requests the pre-formatted A2UI card, bypass plain text!
	if args.Format != nil && *args.Format == "a2ui" {
		a2uiOutput, err := h.formatTasksA2UI(ctx, filteredQueue, args.SiteID, getUserID(ctx))
		if err != nil {
			return nil, err
		}
		return mcp.NewToolResponse(mcp.NewTextContent(a2uiOutput)), nil
	}

	var output string
	for _, item := range filteredQueue {
		output += fmt.Sprintf("Task Name: %s | Template ID: %s | Task ID: %s | Priority: %d | Status: %s", item.Task.Name, item.TaskTemplateID, item.ID, item.Priority, item.Status)
		if item.Task.TargetRoleID != nil {
			output += fmt.Sprintf(" | Target Role ID: %s", *item.Task.TargetRoleID)
		}
		if item.AssigneeID != nil {
			output += fmt.Sprintf(" | Assignee ID: %s", *item.AssigneeID)
		}
		if item.Description != "" {
			output += fmt.Sprintf(" | Description: %s", item.Description)
		}
		output += "\n"
	}
	if output == "" {
		return mcp.NewToolResponse(mcp.NewTextContent("No active tasks found matching the criteria.")), nil
	}

	return mcp.NewToolResponse(mcp.NewTextContent(output)), nil
}

func (h *MCPHandler) isUserEligibleForTask(ctx context.Context, targetRoleID *string, userID string) bool {
	if targetRoleID == nil || *targetRoleID == "" {
		return true
	}
	user, err := h.shiftService.GetUserProfile(ctx, userID)
	if err != nil {
		return false
	}
	for _, r := range user.Roles {
		if r.ID == *targetRoleID {
			return true
		}
	}
	return false
}

func (h *MCPHandler) formatTasksA2UI(ctx context.Context, queue []*model.TaskExecution, siteID string, userID string) (string, error) {
	var taskCards []interface{}

	for _, item := range queue {
		assigneeStr := "Unassigned"
		if item.Assignee != nil {
			assigneeStr = item.Assignee.Name
			if assigneeStr == "" {
				assigneeStr = item.Assignee.Email
			}
		}

		// Create a beautifully nested card conforming to A2UI v0.8 specifications
		taskCard := map[string]interface{}{
			"type":  "card",
			"title": fmt.Sprintf("%s (Priority: %d)", item.Task.Name, item.Priority),
			"style": "standard",
			"children": []interface{}{
				map[string]interface{}{
					"type": "row",
					"children": []interface{}{
						map[string]interface{}{
							"type":    "text",
							"content": fmt.Sprintf("Status: %s", item.Status),
							"style":   "secondary",
						},
						map[string]interface{}{
							"type":    "text",
							"content": fmt.Sprintf("Assignee: %s", assigneeStr),
							"style":   "secondary",
						},
					},
				},
			},
		}

		// Dynamically generate the appropriate active supervisor buttons
		var buttons []interface{}
		buttons = append(buttons, map[string]interface{}{
			"type":  "button",
			"label": "View Details",
			"style": "secondary",
			"action": "VIEW_TASK",
			"actionData": map[string]interface{}{
				"execution_id": item.ID,
			},
		})
		if item.Status != "COMPLETED" {
			if item.AssigneeID != nil && *item.AssigneeID == userID {
				// Belongs to the current user
				if item.Status == "PENDING" {
					buttons = append(buttons, map[string]interface{}{
						"type":  "button",
						"label": "Start Task",
						"style": "primary",
						"action": "START_TASK",
						"actionData": map[string]interface{}{
							"execution_id": item.ID,
						},
					}, map[string]interface{}{
						"type":  "button",
						"label": "Propose Trade",
						"style": "secondary",
						"action": "PROPOSE_TRADE",
						"actionData": map[string]interface{}{
							"execution_id": item.ID,
						},
					})
				} else if item.Status == "IN_PROGRESS" {
					buttons = append(buttons, map[string]interface{}{
						"type":  "button",
						"label": "Continue Task",
						"style": "primary",
						"action": "CONTINUE_TASK",
						"actionData": map[string]interface{}{
							"execution_id": item.ID,
						},
					}, map[string]interface{}{
						"type":  "button",
						"label": "Complete Task",
						"style": "secondary",
						"action": "COMPLETE_TASK",
						"actionData": map[string]interface{}{
							"execution_id": item.ID,
						},
					})
				}
			} else {
				// Does not belong to the current user: allow claiming/taking if eligible
				if h.isUserEligibleForTask(ctx, item.Task.TargetRoleID, userID) {
					var label string
					if item.AssigneeID == nil || *item.AssigneeID == "" {
						label = "Claim Task"
					} else {
						label = "Take Task"
					}
					buttons = append(buttons, map[string]interface{}{
						"type":  "button",
						"label": label,
						"style": "primary",
						"action": "CLAIM_TASK",
						"actionData": map[string]interface{}{
							"execution_id": item.ID,
						},
					})
				}
			}
		}

		if len(buttons) > 0 {
			taskCard["children"] = append(taskCard["children"].([]interface{}), map[string]interface{}{
				"type":     "row",
				"children": buttons,
			})
		}

		taskCards = append(taskCards, taskCard)
	}

	// Wrap all task cards inside a primary column under the site queue
	mainCard := map[string]interface{}{
		"type":  "card",
		"title": "SITE OPERATIONAL TASK QUEUE",
		"style": "primary",
		"children": []interface{}{
			map[string]interface{}{
				"type":     "column",
				"gap":      12,
				"children": taskCards,
			},
		},
	}

	flatTransaction := NormalizeCardToA2UITransaction(mainCard, "surface_site_tasks")
	jsonBytes, err := json.Marshal(flatTransaction)
	if err != nil {
		return "", err
	}

	return string(jsonBytes), nil
}

func (h *MCPHandler) HandleOverrideAsset(ctx context.Context, args OverrideAssetArgs) (*mcp.ToolResponse, error) {
	userID := getUserID(ctx)
	if userID == "" {
		return nil, fmt.Errorf("unauthorized: user ID not found in context")
	}
	err := h.taskService.OverrideAssetConstraint(ctx, args.ExecutionID, args.AssetID, args.Justification, userID)
	if err != nil {
		return nil, err
	}
	return mcp.NewToolResponse(mcp.NewTextContent("Asset constraint override successfully applied and logged to ledger.")), nil
}

func (h *MCPHandler) HandleProposeTrade(ctx context.Context, args ProposeTradeArgs) (*mcp.ToolResponse, error) {
	userID := getUserID(ctx)
	if userID == "" {
		return nil, fmt.Errorf("unauthorized: user ID not found in context")
	}
	err := h.taskService.ProposeTrade(ctx, args.TaskExecutionID, args.ProposedAssigneeID, userID)
	if err != nil {
		return nil, err
	}
	return mcp.NewToolResponse(mcp.NewTextContent("Task trade request successfully proposed and created.")), nil
}

func (h *MCPHandler) HandleAcceptTrade(ctx context.Context, args AcceptTradeArgs) (*mcp.ToolResponse, error) {
	userID := getUserID(ctx)
	if userID == "" {
		return nil, fmt.Errorf("unauthorized: user ID not found in context")
	}
	err := h.taskService.AcceptTrade(ctx, args.TradeID, userID)
	if err != nil {
		return nil, err
	}
	return mcp.NewToolResponse(mcp.NewTextContent("Task trade request successfully accepted. Task reassigned under GORM database guidelines.")), nil
}

func (h *MCPHandler) HandleRejectTrade(ctx context.Context, args RejectTradeArgs) (*mcp.ToolResponse, error) {
	userID := getUserID(ctx)
	if userID == "" {
		return nil, fmt.Errorf("unauthorized: user ID not found in context")
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

func (h *MCPHandler) HandleGetUserContext(ctx context.Context, args GetUserContextArgs) (*mcp.ToolResponse, error) {
	userID := getUserID(ctx)
	if userID == "" {
		return nil, fmt.Errorf("unauthorized: user ID not found in context")
	}

	user, err := h.shiftService.GetUserProfile(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch user profile: %w", err)
	}

	var roles []string
	isAdmin := false
	for _, r := range user.Roles {
		roles = append(roles, r.Name)
		if r.Name == "ADMIN" {
			isAdmin = true
		}
	}

	var orgs []OrgContextInfo
	for _, o := range user.Organizations {
		orgs = append(orgs, OrgContextInfo{
			ID:   o.ID,
			Name: o.Name,
		})
	}

	var sites []SiteContextInfo
	if isAdmin {
		activeSites, err := h.taskService.ListActiveSites(ctx)
		if err == nil {
			for _, s := range activeSites {
				sites = append(sites, SiteContextInfo{
					ID:             s.ID,
					Name:           s.Name,
					OrganizationID: s.OrganizationID,
				})
			}
		}
	}
	if len(sites) == 0 {
		for _, s := range user.Sites {
			sites = append(sites, SiteContextInfo{
				ID:             s.ID,
				Name:           s.Name,
				OrganizationID: s.OrganizationID,
			})
		}
	}

	resp := UserContextResponse{
		UserID:        user.ID,
		Email:         user.Email,
		Name:          user.Name,
		Roles:         roles,
		Organizations: orgs,
		Sites:         sites,
	}

	respBytes, err := json.MarshalIndent(resp, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("failed to marshal user context: %w", err)
	}

	return mcp.NewToolResponse(mcp.NewTextContent(string(respBytes))), nil
}

type ClaimTaskArgs struct {
	ExecutionID string `json:"execution_id" jsonschema:"description=The GORM UUID of the task execution being claimed"`
}

func (h *MCPHandler) HandleClaimTask(ctx context.Context, args ClaimTaskArgs) (*mcp.ToolResponse, error) {
	userID := getUserID(ctx)
	if userID == "" {
		return nil, fmt.Errorf("user context not initialized")
	}

	user, err := h.shiftService.GetUserProfile(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve user profile: %w", err)
	}

	var roleIDs []string
	for _, r := range user.Roles {
		roleIDs = append(roleIDs, r.ID)
	}

	err = h.taskService.ClaimTask(ctx, args.ExecutionID, userID, roleIDs)
	if err != nil {
		return nil, err
	}

	return mcp.NewToolResponse(mcp.NewTextContent("Task successfully claimed and assigned.")), nil
}

type UpdateTaskStatusArgs struct {
	ExecutionID    string `json:"execution_id" jsonschema:"description=The GORM UUID of the task execution being updated"`
	Status         string `json:"status" jsonschema:"description=The new status of the task execution (e.g. IN_PROGRESS, COMPLETED)"`
	ChecklistState string `json:"checklist_state,omitempty" jsonschema:"description=Optional JSON string representing the updated checklist items state"`
}

func (h *MCPHandler) HandleUpdateTaskStatus(ctx context.Context, args UpdateTaskStatusArgs) (*mcp.ToolResponse, error) {
	userID := getUserID(ctx)
	if userID == "" {
		return nil, fmt.Errorf("user context not initialized")
	}

	err := h.taskService.UpdateStatus(ctx, args.ExecutionID, args.Status, args.ChecklistState, userID)
	if err != nil {
		return nil, err
	}

	return mcp.NewToolResponse(mcp.NewTextContent("Task status successfully updated.")), nil
}

type ListPendingTradesArgs struct{}

func (h *MCPHandler) HandleListPendingTrades(ctx context.Context, args ListPendingTradesArgs) (*mcp.ToolResponse, error) {
	userID := getUserID(ctx)
	if userID == "" {
		return nil, fmt.Errorf("user context not initialized")
	}

	trades, err := h.taskService.ListPendingTrades(ctx, userID)
	if err != nil {
		return nil, err
	}

	respBytes, err := json.MarshalIndent(trades, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("failed to marshal pending trades: %w", err)
	}

	return mcp.NewToolResponse(mcp.NewTextContent(string(respBytes))), nil
}

func (h *MCPHandler) formatLocationsA2UI(ctx context.Context, siteID string, locationID *string) (string, error) {
	// Determine layout based on siteID
	layout := "linear"
	if siteID == "44444444-4444-4444-4444-444444440001" {
		layout = "boutique"
	} else if siteID == "44444444-4444-4444-4444-444444440002" {
		layout = "racetrack"
	}

	type A2UIBeacon struct {
		X    *int   `json:"x,omitempty"`
		Y    *int   `json:"y,omitempty"`
		Name string `json:"name,omitempty"`
	}

	type A2UICanvas struct {
		Type   string      `json:"type"`
		Layout string      `json:"layout"`
		Beacon *A2UIBeacon `json:"beacon,omitempty"`
	}

	type A2UIStoreMapCard struct {
		Type     string       `json:"type"`
		Title    string       `json:"title"`
		Style    string       `json:"style"`
		Children []A2UICanvas `json:"children"`
	}

	var beacon *A2UIBeacon
	if locationID != nil && *locationID != "" {
		// Find location by ID
		loc, err := h.taskService.GetLocationByID(ctx, *locationID)
		if err == nil && loc != nil {
			funcType := strings.ToLower(loc.LocationFunctionType)
			name := strings.ToLower(loc.Name)

			var x, y int
			var beaconName string

			// Determine coordinate based on layout and type/name
			if strings.Contains(funcType, "vault") || strings.Contains(funcType, "safe") || strings.Contains(name, "vault") || strings.Contains(name, "safe") {
				beaconName = "Secure Back-Office Cash Vault"
				if layout == "boutique" {
					x, y = 175, 25
				} else if layout == "racetrack" {
					x, y = 30, 125
					beaconName = "Sub-Level Cash Room"
				} else {
					x, y = 184, 125
					beaconName = "Main Store Cash Vault Room"
				}
			} else if strings.Contains(funcType, "register") || strings.Contains(funcType, "checkout") || strings.Contains(name, "register") || strings.Contains(name, "checkout") {
				beaconName = "Boutique Front Checkout Counter"
				if layout == "boutique" {
					x, y = 105, 125
				} else if layout == "racetrack" {
					x, y = 150, 125
					beaconName = "South Register Gallery"
				} else {
					x, y = 162, 65
					beaconName = "Registers Lane 4 Checkouts Corridor"
				}
			} else if strings.Contains(name, "cooler") || strings.Contains(name, "wet") || strings.Contains(name, "produce") || strings.Contains(name, "greens") {
				beaconName = "Organic Micro-Greens Cool Wall"
				if layout == "boutique" {
					x, y = 45, 25
				} else if layout == "racetrack" {
					x, y = 30, 25
					beaconName = "Flagship Fresh Food Chilled Canopy"
				} else {
					x, y = 73, 10
					beaconName = "Produce Perimeter Wet Wall Cabinets"
				}
			} else if strings.Contains(name, "showcase") || strings.Contains(name, "atrium") || strings.Contains(name, "smart") || strings.Contains(name, "display") {
				beaconName = "Central Interactive Appliance Ring"
				if layout == "boutique" {
					x, y = 100, 75
				} else if layout == "racetrack" {
					x, y = 100, 75
					beaconName = "Atrium Smart-Home Experience Center"
				} else {
					x, y = 111, 44
					beaconName = "Aisle 10 Showroom Display"
				}
			} else if strings.Contains(name, "loading") || strings.Contains(name, "dock") || strings.Contains(name, "storage") || strings.Contains(name, "bay") {
				beaconName = "SF Rear Loading Bay"
				if layout == "boutique" {
					x, y = 15, 25
				} else if layout == "racetrack" {
					x, y = 175, 25
					beaconName = "North Cargo Intake Bay"
				} else {
					x, y = 15, 20
					beaconName = "Receiving Dock A Cargo Bay"
				}
			} else {
				// Fallback generic beacon if type not matched but ID given
				beaconName = loc.Name
				x, y = 100, 75
			}

			beacon = &A2UIBeacon{
				X:    &x,
				Y:    &y,
				Name: beaconName,
			}
		}
	}

	canvas := A2UICanvas{
		Type:   "canvas",
		Layout: layout,
		Beacon: beacon,
	}

	card := A2UIStoreMapCard{
		Type:     "card",
		Title:    "STORE SPATIAL BLUEPRINT MAP",
		Style:    "primary",
		Children: []A2UICanvas{canvas},
	}

	respBytes, err := json.Marshal(card)
	if err != nil {
		return "", err
	}

	// Convert nested struct to flat A2UI transaction on the fly
	var cardMap map[string]interface{}
	if err := json.Unmarshal(respBytes, &cardMap); err == nil {
		flatTransaction := NormalizeCardToA2UITransaction(cardMap, "surface_store_locations")
		if flatBytes, err := json.Marshal(flatTransaction); err == nil {
			return string(flatBytes), nil
		}
	}

	return string(respBytes), nil
}

func (h *MCPHandler) HandleGetSiteLocations(ctx context.Context, args GetSiteLocationsArgs) (*mcp.ToolResponse, error) {
	userID := getUserID(ctx)
	if userID == "" {
		return nil, fmt.Errorf("user context not initialized")
	}

	if args.Format != nil && *args.Format == "a2ui" {
		cardJSON, err := h.formatLocationsA2UI(ctx, args.SiteID, args.LocationID)
		if err != nil {
			return nil, fmt.Errorf("failed to format locations to A2UI: %w", err)
		}
		return mcp.NewToolResponse(mcp.NewTextContent(cardJSON)), nil
	}

	locations, err := h.taskService.GetSiteLocations(ctx, args.SiteID)
	if err != nil {
		return nil, err
	}

	var output string
	for _, l := range locations {
		output += fmt.Sprintf("Location Name: %s | ID: %s | Type: %s | Function: %s\n", l.Name, l.ID, l.LocationType, l.LocationFunctionType)
	}
	if output == "" {
		output = "No locations configured for site."
	}
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

func (h *MCPHandler) getTaskBeaconAndLayout(ctx context.Context, taskName string, activeStepAction string, siteID string) (string, interface{}) {
	layout := "linear"
	if siteID == "44444444-4444-4444-4444-444444440001" {
		layout = "boutique"
	} else if siteID == "44444444-4444-4444-4444-444444440002" {
		layout = "racetrack"
	}

	// 1. Try matching the active step action first to get a highly dynamic location
	if activeStepAction != "" {
		beacon := h.resolveBeaconByString(ctx, activeStepAction, layout)
		if beacon != nil {
			return layout, beacon
		}
	}

	// 2. Fallback to matching the overall task name
	beacon := h.resolveBeaconByString(ctx, taskName, layout)
	return layout, beacon
}

func (h *MCPHandler) resolveBeaconByString(ctx context.Context, str string, layout string) interface{} {
	lower := strings.ToLower(str)
	var x, y int
	var beaconName string
	matched := false

	if strings.Contains(lower, "vault") || strings.Contains(lower, "safe") {
		matched = true
		beaconName = "Secure Back-Office Cash Vault"
		if layout == "boutique" {
			x, y = 175, 25
		} else if layout == "racetrack" {
			x, y = 30, 125
			beaconName = "Sub-Level Cash Room"
		} else {
			x, y = 184, 125
			beaconName = "Main Store Cash Vault Room"
		}
	} else if strings.Contains(lower, "register") || strings.Contains(lower, "till") || strings.Contains(lower, "checkout") {
		matched = true
		beaconName = "Boutique Front Checkout Counter"
		if layout == "boutique" {
			x, y = 105, 125
		} else if layout == "racetrack" {
			x, y = 150, 125
			beaconName = "South Register Gallery"
		} else {
			x, y = 162, 65
			beaconName = "Registers Lane 4 Checkouts Corridor"
		}
	} else if strings.Contains(lower, "greens") || strings.Contains(lower, "produce") || strings.Contains(lower, "wall") || strings.Contains(lower, "wet") || strings.Contains(lower, "shelf") || strings.Contains(lower, "replenishment") || strings.Contains(lower, "aisle 7") {
		matched = true
		beaconName = "Organic Micro-Greens Cool Wall"
		if layout == "boutique" {
			x, y = 45, 25
		} else if layout == "racetrack" {
			x, y = 30, 25
			beaconName = "Flagship Fresh Food Chilled Canopy"
		} else {
			x, y = 73, 10
			beaconName = "Produce Perimeter Wet Wall Cabinets"
		}
	} else if strings.Contains(lower, "showcase") || strings.Contains(lower, "atrium") || strings.Contains(lower, "experience") || strings.Contains(lower, "display") || strings.Contains(lower, "calibration") || strings.Contains(lower, "volt") {
		matched = true
		beaconName = "Central Interactive Appliance Ring"
		if layout == "boutique" {
			x, y = 100, 75
		} else if layout == "racetrack" {
			x, y = 100, 75
			beaconName = "Atrium Smart-Home Experience Center"
		} else {
			x, y = 111, 44
			beaconName = "Aisle 10 Showroom Display"
		}
	} else if strings.Contains(lower, "dock") || strings.Contains(lower, "loading") || strings.Contains(lower, "receiving") || strings.Contains(lower, "stock") || strings.Contains(lower, "cage") {
		matched = true
		beaconName = "SF Rear Loading Bay"
		if layout == "boutique" {
			x, y = 15, 25
		} else if layout == "racetrack" {
			x, y = 175, 25
			beaconName = "North Cargo Intake Bay"
		} else {
			x, y = 15, 20
			beaconName = "Receiving Dock A Cargo Bay"
		}
	}

	if matched {
		return map[string]interface{}{
			"x":    x,
			"y":    y,
			"name": beaconName,
		}
	}
	return nil
}

func (h *MCPHandler) HandleGetTaskDetails(ctx context.Context, args GetTaskDetailsArgs) (*mcp.ToolResponse, error) {
	exec, err := h.taskService.GetTaskExecutionByID(ctx, args.ExecutionID)
	if err != nil {
		return nil, err
	}

	userID := getUserID(ctx)

	// Unmarshal the checklist state
	var steps []struct {
		Step      int    `json:"step"`
		Action    string `json:"action"`
		Required  bool   `json:"required"`
		Completed bool   `json:"completed"`
	}
	if len(exec.ChecklistState) > 0 {
		_ = json.Unmarshal(exec.ChecklistState, &steps)
	}

	assigneeStr := "Unassigned"
	if exec.Assignee != nil {
		assigneeStr = exec.Assignee.Name
		if assigneeStr == "" {
			assigneeStr = exec.Assignee.Email
		}
	}

	// Build the A2UI Card
	var children []interface{}
	
	// Description
	desc := exec.Task.Description
	if desc == "" {
		desc = "No description provided."
	}
	children = append(children, map[string]interface{}{
		"type":    "text",
		"content": desc,
		"style":   "primary",
	})

	// Metadata Row
	children = append(children, map[string]interface{}{
		"type": "row",
		"children": []interface{}{
			map[string]interface{}{
				"type":    "text",
				"content": fmt.Sprintf("Status: %s", exec.Status),
				"style":   "secondary",
			},
			map[string]interface{}{
				"type":    "text",
				"content": fmt.Sprintf("Assignee: %s", assigneeStr),
				"style":   "secondary",
			},
		},
	})

	// Steps Header
	children = append(children, map[string]interface{}{
		"type":    "text",
		"content": "Checklist Steps:",
		"style":   "primary",
	})

	// Steps list
	isAssignedToCaller := exec.AssigneeID != nil && *exec.AssigneeID == userID
	for _, step := range steps {
		statusEmoji := "🔲"
		style := "secondary"
		if step.Completed {
			statusEmoji = "✅"
			style = "primary"
		}
		reqStr := ""
		if step.Required {
			reqStr = " (Required)"
		}
		textLabel := fmt.Sprintf("Step %d: %s%s", step.Step, step.Action, reqStr)

		if isAssignedToCaller && exec.Status == "IN_PROGRESS" {
			// Construct a single-step delta payload rather than a full state replacement
			deltaPayload := map[string]interface{}{
				"step":      step.Step,
				"completed": !step.Completed,
			}
			deltaBytes, _ := json.Marshal(deltaPayload)

			// Render as a clean row containing an independent checkbox button on the left
			// and the static step description text on the right!
			children = append(children, map[string]interface{}{
				"type": "row",
				"children": []interface{}{
					map[string]interface{}{
						"type":  "button",
						"label": statusEmoji,
						"style": style,
						"action": "UPDATE_CHECKLIST",
						"actionData": map[string]interface{}{
							"execution_id":    exec.ID,
							"status":          "IN_PROGRESS",
							"checklist_state": string(deltaBytes),
						},
					},
					map[string]interface{}{
						"type":    "text",
						"content": textLabel,
						"style":   "primary",
					},
				},
			})
		} else {
			// Render as static row for unassigned / non-interactive states
			children = append(children, map[string]interface{}{
				"type": "row",
				"children": []interface{}{
					map[string]interface{}{
						"type":    "text",
						"content": statusEmoji,
						"style":   "secondary",
					},
					map[string]interface{}{
						"type":    "text",
						"content": textLabel,
						"style":   "secondary",
					},
				},
			})
		}
	}


	// Actions Row
	var buttons []interface{}
	
	if exec.Status != "COMPLETED" {
		if isAssignedToCaller {
			// Belongs to the current user
			if exec.Status == "PENDING" {
				buttons = append(buttons, map[string]interface{}{
					"type":  "button",
					"label": "Start Task",
					"style": "primary",
					"action": "START_TASK",
					"actionData": map[string]interface{}{
						"execution_id": exec.ID,
					},
				}, map[string]interface{}{
					"type":  "button",
					"label": "Propose Trade",
					"style": "secondary",
					"action": "PROPOSE_TRADE",
					"actionData": map[string]interface{}{
						"execution_id": exec.ID,
					},
				})
			} else if exec.Status == "IN_PROGRESS" {
				buttons = append(buttons, map[string]interface{}{
					"type":  "button",
					"label": "Complete Task",
					"style": "primary",
					"action": "COMPLETE_TASK",
					"actionData": map[string]interface{}{
						"execution_id": exec.ID,
					},
				})
			}
		} else {
			// Does not belong to the current user: allow claiming/taking if eligible
			if h.isUserEligibleForTask(ctx, exec.Task.TargetRoleID, userID) {
				var label string
				if exec.AssigneeID == nil || *exec.AssigneeID == "" {
					label = "Claim Task"
				} else {
					label = "Take Task"
				}
				buttons = append(buttons, map[string]interface{}{
					"type":  "button",
					"label": label,
					"style": "primary",
					"action": "CLAIM_TASK",
					"actionData": map[string]interface{}{
						"execution_id": exec.ID,
					},
				})
			}
		}
	}

	if len(buttons) > 0 {
		children = append(children, map[string]interface{}{
			"type":     "row",
			"children": buttons,
		})
	}

	// Find the first incomplete step to dynamically focus the map on the active location
	var activeStepAction string
	for _, step := range steps {
		if !step.Completed {
			activeStepAction = step.Action
			break
		}
	}

	// Resolve the site ID and embed the Spatial Canvas
	siteID, err := h.taskService.GetSiteIDForExecution(ctx, exec.ID)
	if err == nil && siteID != "" {
		layout, beacon := h.getTaskBeaconAndLayout(ctx, exec.Task.Name, activeStepAction, siteID)
		if beacon != nil {
			children = append(children, map[string]interface{}{
				"type":   "canvas",
				"layout": layout,
				"beacon": beacon,
			})
		}
	}

	cardPayload := map[string]interface{}{
		"type":  "card",
		"title": fmt.Sprintf("TASK DETAILS: %s (Priority: %d)", exec.Task.Name, exec.Priority),
		"style": "standard",
		"children": children,
	}

	flatTransaction := NormalizeCardToA2UITransaction(cardPayload, "surface_task_details")
	cardBytes, err := json.Marshal(flatTransaction)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal A2UI card: %w", err)
	}

	a2uiOutput := fmt.Sprintf("```json\n%s\n```", string(cardBytes))
	return mcp.NewToolResponse(mcp.NewTextContent(a2uiOutput)), nil
}

type GetWeatherArgs struct {
	Station string `json:"station" jsonschema:"description=The ICAO airport station code, e.g. KDFW"`
}

func (h *MCPHandler) HandleGetWeather(ctx context.Context, args GetWeatherArgs) (*mcp.ToolResponse, error) {
	station := strings.ToUpper(args.Station)
	if station == "" {
		station = "KDFW"
	}

	tempVal, windVal, pressureVal, visibilityVal := GenerateDeterministicWeather(station)

	card := map[string]interface{}{
		"type":  "card",
		"title": fmt.Sprintf("METAR AIRPORT WIND AUDIT (%s)", station),
		"style": "standard",
		"children": []interface{}{
			map[string]interface{}{
				"type": "table",
				"rows": []interface{}{
					map[string]interface{}{"label": "Station", "value": station},
					map[string]interface{}{"label": "Temperature", "value": tempVal},
					map[string]interface{}{"label": "Wind", "value": windVal},
					map[string]interface{}{"label": "Barometric Pressure", "value": pressureVal},
					map[string]interface{}{"label": "Visibility", "value": visibilityVal},
				},
			},
		},
	}

	flatTransaction := NormalizeCardToA2UITransaction(card, "surface_weather")
	respBytes, err := json.Marshal(flatTransaction)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal weather card: %w", err)
	}

	return mcp.NewToolResponse(mcp.NewTextContent(string(respBytes))), nil
}

type GetStoreSelectorArgs struct{}

func (h *MCPHandler) HandleGetStoreSelector(ctx context.Context, args GetStoreSelectorArgs) (*mcp.ToolResponse, error) {
	userID := getUserID(ctx)
	if userID == "" {
		return nil, fmt.Errorf("unauthorized: user ID not found in context")
	}

	user, err := h.shiftService.GetUserProfile(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch user profile: %w", err)
	}

	var sites []SiteContextInfo
	isAdmin := false
	for _, r := range user.Roles {
		if r.Name == "ADMIN" {
			isAdmin = true
			break
		}
	}

	if isAdmin {
		activeSites, err := h.taskService.ListActiveSites(ctx)
		if err == nil {
			for _, s := range activeSites {
				sites = append(sites, SiteContextInfo{
					ID:             s.ID,
					Name:           s.Name,
					OrganizationID: s.OrganizationID,
				})
			}
		}
	}
	if len(sites) == 0 {
		for _, s := range user.Sites {
			sites = append(sites, SiteContextInfo{
				ID:             s.ID,
				Name:           s.Name,
				OrganizationID: s.OrganizationID,
			})
		}
	}

	buttons := []interface{}{}
	for _, s := range sites {
		buttons = append(buttons, map[string]interface{}{
			"type":  "button",
			"label": s.Name,
			"style": "primary",
			"action": "SET_STORE",
			"actionData": map[string]interface{}{
				"siteID":    s.ID,
				"siteLabel": s.Name,
			},
		})
	}

	card := map[string]interface{}{
		"type":  "card",
		"title": "RETAIL STOREFRONT CONTEXT SWITCHER",
		"style": "primary",
		"children": []interface{}{
			map[string]interface{}{
				"type":     "column",
				"gap":      8,
				"children": buttons,
			},
		},
	}

	flatTransaction := NormalizeCardToA2UITransaction(card, "surface_store_selector")
	respBytes, err := json.Marshal(flatTransaction)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal store selector card: %w", err)
	}

	return mcp.NewToolResponse(mcp.NewTextContent(string(respBytes))), nil
}

