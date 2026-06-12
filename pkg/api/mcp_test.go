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

package api_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/rmcguinness/gemini_task_engine/pkg/agents"
	"github.com/rmcguinness/gemini_task_engine/pkg/api"
	"github.com/rmcguinness/gemini_task_engine/pkg/model"
	"github.com/stretchr/testify/assert"
)

func TestMCPHandler_TriggerAlertTool(t *testing.T) {
	taskSvc := &mockTaskService{}
	shiftSvc := &mockShiftService{}
	ragSvc := &mockRAGService{}
	automationSvc := &mockAutomationService{}

	handler, err := api.NewMCPHandler(taskSvc, ragSvc, shiftSvc, automationSvc)
	assert.NoError(t, err)

	t.Run("successfully triggers adhoc alert via HandleTriggerAlert", func(t *testing.T) {
		mockExec := &model.TaskExecution{
			ID:              "exec-adhoc-ticket-777",
			TaskTemplateID:  "template-till-out",
			ExecutionType:   "ADHOC",
			EventInstanceID: "00000000-0000-0000-0000-ffffffffffff",
			Status:          "PENDING",
			Priority:        1,
		}

		// Configure mock callbacks to capture and verify parameters
		var capturedSiteID, capturedOrganizerID string
		var capturedEventType model.EventType
		var capturedDescription string

		automationSvc.TriggerStreamingEventFunc = func(
			ctx context.Context,
			siteID string,
			organizerID string,
			eventType model.EventType,
			description string,
		) (*model.TaskExecution, error) {
			capturedSiteID = siteID
			capturedOrganizerID = organizerID
			capturedEventType = eventType
			capturedDescription = description
			return mockExec, nil
		}

		// Execute tool trigger directly bypassing blocking transport
		args := api.TriggerAlertArgs{
			SiteID:      "site-dallas-1000",
			OrganizerID: "register-lane-5",
			EventType:   "TillDrawerDropEvent",
			Description:  "Till Drawer exceeds cash ceiling. Supervisor transfer required.",
		}

		resp, err := handler.HandleTriggerAlert(context.Background(), args)

		// Assert parameters passed to backing services match
		assert.NoError(t, err)
		assert.NotNil(t, resp)
		assert.Equal(t, "site-dallas-1000", capturedSiteID)
		assert.Equal(t, "register-lane-5", capturedOrganizerID)
		assert.Equal(t, model.EventTillDrawerDrop, capturedEventType)
		assert.Equal(t, "Till Drawer exceeds cash ceiling. Supervisor transfer required.", capturedDescription)

		// Validate text output content structures
		assert.Len(t, resp.Content, 1)
		textMap := resp.Content[0]
		assert.Contains(t, textMap.TextContent.Text, "Streaming alert successfully ingested")
		assert.Contains(t, textMap.TextContent.Text, "exec-adhoc-ticket-777")
		assert.Contains(t, textMap.TextContent.Text, "Priority: 1")
		assert.Contains(t, textMap.TextContent.Text, "Status: PENDING")
	})

	t.Run("fails to execute HandleTriggerAlert under service layer errors", func(t *testing.T) {
		automationSvc.TriggerStreamingEventFunc = func(
			ctx context.Context,
			siteID string,
			organizerID string,
			eventType model.EventType,
			description string,
		) (*model.TaskExecution, error) {
			return nil, errors.New("alloydb write bottleneck failure")
		}

		args := api.TriggerAlertArgs{
			SiteID:      "site-1",
			OrganizerID: "sensor-1",
			EventType:   "StockoutCorrectEvent",
			Description:  "desc",
		}

		resp, err := handler.HandleTriggerAlert(context.Background(), args)

		assert.Error(t, err)
		assert.Nil(t, resp)
		assert.Contains(t, err.Error(), "alloydb write bottleneck failure")
	})
}

func TestMCPHandler_GetTasksTool(t *testing.T) {
	taskSvc := &mockTaskService{}
	shiftSvc := &mockShiftService{}
	ragSvc := &mockRAGService{}
	automationSvc := &mockAutomationService{}

	handler, err := api.NewMCPHandler(taskSvc, ragSvc, shiftSvc, automationSvc)
	assert.NoError(t, err)

	t.Run("successfully queries task queues", func(t *testing.T) {
		taskSvc.GetQueueFunc = func(ctx context.Context, siteID string) ([]*model.TaskExecution, error) {
			assert.Equal(t, "site-dallas-1000", siteID)
			return []*model.TaskExecution{
				{
					ID:             "exec-1",
					TaskTemplateID: "template-opening",
					Description:    "Grounded cash audit SOP: secure drops and verify vault.",
					Priority:       1,
					Status:         "PENDING",
				},
			}, nil
		}

		args := api.GetTasksArgs{
			SiteID: "site-dallas-1000",
		}

		resp, err := handler.HandleGetTasks(context.Background(), args)

		assert.NoError(t, err)
		assert.NotNil(t, resp)
		assert.Len(t, resp.Content, 1)

		textMap := resp.Content[0]
		assert.Contains(t, textMap.TextContent.Text, "Template ID: template-opening")
		assert.Contains(t, textMap.TextContent.Text, "Task ID: exec-1")
		assert.Contains(t, textMap.TextContent.Text, "Priority: 1")
		assert.Contains(t, textMap.TextContent.Text, "Status: PENDING")
		assert.Contains(t, textMap.TextContent.Text, "Description: Grounded cash audit SOP")
	})
}

func TestMCPHandler_PromptsList(t *testing.T) {
	t.Run("successfully asserts shift_agent hanna prompt registrations", func(t *testing.T) {
		agent, exists := agents.Get("shift_agent")
		assert.True(t, exists)
		assert.Equal(t, "shift_agent", agent.ID)
		assert.Contains(t, agent.SystemInstruction, "You are Hanna, a highly capable, professional, and direct retail shift assistant")
		assert.Len(t, agent.AllowedTools, 6)
		assert.Contains(t, agent.AllowedTools, "get_tasks")
		assert.Contains(t, agent.AllowedTools, "override_asset")
		assert.Contains(t, agent.AllowedTools, "propose_trade")
		assert.Contains(t, agent.AllowedTools, "accept_trade")
		assert.Contains(t, agent.AllowedTools, "reject_trade")
		assert.Contains(t, agent.AllowedTools, "query_sop")
	})
}

func TestMCPHandler_VersionNegotiationAndErrorTranslation(t *testing.T) {
	gin.SetMode(gin.TestMode)

	taskSvc := &mockTaskService{}
	shiftSvc := &mockShiftService{}
	ragSvc := &mockRAGService{}
	automationSvc := &mockAutomationService{}

	handler, err := api.NewMCPHandler(taskSvc, ragSvc, shiftSvc, automationSvc)
	assert.NoError(t, err)

	router := gin.New()
	router.POST("/mcp", handler.Handler())

	t.Run("overrides initialize protocol version to 2025-11-25", func(t *testing.T) {
		initReq := map[string]interface{}{
			"jsonrpc": "2.0",
			"id":      1,
			"method":  "initialize",
			"params": map[string]interface{}{
				"protocolVersion": "2025-11-25",
				"capabilities":    map[string]interface{}{},
				"clientInfo": map[string]interface{}{
					"name":    "test-client",
					"version": "1.0.0",
				},
			},
		}

		reqBody, err := json.Marshal(initReq)
		assert.NoError(t, err)

		w := httptest.NewRecorder()
		req, err := http.NewRequest("POST", "/mcp", bytes.NewBuffer(reqBody))
		assert.NoError(t, err)
		req.Header.Set("Content-Type", "application/json")

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var respBody map[string]interface{}
		err = json.Unmarshal(w.Body.Bytes(), &respBody)
		assert.NoError(t, err)

		result, ok := respBody["result"].(map[string]interface{})
		assert.True(t, ok)
		assert.Equal(t, "2025-11-25", result["protocolVersion"])
	})

	t.Run("translates tool call parameter validation error to isError true result", func(t *testing.T) {
		toolReq := map[string]interface{}{
			"jsonrpc": "2.0",
			"id":      2,
			"method":  "tools/call",
			"params": map[string]interface{}{
				"name":      "get_tasks",
				"arguments": "invalid-arguments-string-instead-of-object",
			},
		}

		reqBody, err := json.Marshal(toolReq)
		assert.NoError(t, err)

		w := httptest.NewRecorder()
		req, err := http.NewRequest("POST", "/mcp", bytes.NewBuffer(reqBody))
		assert.NoError(t, err)
		req.Header.Set("Content-Type", "application/json")

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var respBody map[string]interface{}
		err = json.Unmarshal(w.Body.Bytes(), &respBody)
		assert.NoError(t, err)

		assert.Nil(t, respBody["error"])
		result, ok := respBody["result"].(map[string]interface{})
		assert.True(t, ok)
		assert.Equal(t, true, result["isError"])

		content, ok := result["content"].([]interface{})
		assert.True(t, ok)
		assert.NotEmpty(t, content)

		contentMap, ok := content[0].(map[string]interface{})
		assert.True(t, ok)
		assert.Equal(t, "text", contentMap["type"])
		assert.Contains(t, contentMap["text"], "unmarshal")
	})
}

func TestMCPHandler_GetUserContextTool(t *testing.T) {
	taskSvc := &mockTaskService{}
	shiftSvc := &mockShiftService{}
	ragSvc := &mockRAGService{}
	automationSvc := &mockAutomationService{}

	handler, err := api.NewMCPHandler(taskSvc, ragSvc, shiftSvc, automationSvc)
	assert.NoError(t, err)

	t.Run("successfully retrieves user profile and mappings from context", func(t *testing.T) {
		mockUser := &model.User{
			ID:    "11111111-1111-1111-1111-111111111111",
			Email: "alex@omnimart.com",
			Name:  "Alex Associate",
			Roles: []model.Role{
				{ID: "role-1", Name: "Retail Associate"},
			},
			Organizations: []model.Organization{
				{ID: "org-1", Name: "OmniMart"},
			},
			Sites: []model.Site{
				{ID: "site-1", Name: "OmniMart - Dallas #14", OrganizationID: "org-1"},
			},
		}

		shiftSvc.GetUserProfileFunc = func(ctx context.Context, userID string) (*model.User, error) {
			assert.Equal(t, "11111111-1111-1111-1111-111111111111", userID)
			return mockUser, nil
		}

		ctx := context.WithValue(context.Background(), "userID", "11111111-1111-1111-1111-111111111111")
		args := api.GetUserContextArgs{}

		resp, err := handler.HandleGetUserContext(ctx, args)

		assert.NoError(t, err)
		assert.NotNil(t, resp)
		assert.Len(t, resp.Content, 1)

		text := resp.Content[0].TextContent.Text
		assert.Contains(t, text, "11111111-1111-1111-1111-111111111111")
		assert.Contains(t, text, "alex@omnimart.com")
		assert.Contains(t, text, "Alex Associate")
		assert.Contains(t, text, "Retail Associate")
		assert.Contains(t, text, "OmniMart")
		assert.Contains(t, text, "OmniMart - Dallas #14")
	})

	t.Run("fails when userID context key is missing", func(t *testing.T) {
		args := api.GetUserContextArgs{}
		resp, err := handler.HandleGetUserContext(context.Background(), args)

		assert.Error(t, err)
		assert.Nil(t, resp)
		assert.Contains(t, err.Error(), "unauthorized: user ID not found in context")
	})
}
