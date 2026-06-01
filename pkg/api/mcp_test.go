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
	"context"
	"errors"
	"testing"

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
		assert.Contains(t, textMap.TextContent.Text, "Task: template-opening")
		assert.Contains(t, textMap.TextContent.Text, "ID: exec-1")
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
