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
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/rmcguinness/gemini_task_engine/pkg/api"
	"github.com/rmcguinness/gemini_task_engine/pkg/model"
	"github.com/stretchr/testify/assert"
)

func TestChatHandler_SendMessage(t *testing.T) {
	gin.SetMode(gin.TestMode)

	taskSvc := &mockTaskService{}
	shiftSvc := &mockShiftService{}
	ragSvc := &mockRAGService{}
	automationSvc := &mockAutomationService{}

	handler := api.NewChatHandler(taskSvc, shiftSvc, ragSvc, automationSvc)

	t.Run("successfully routes weather query intents", func(t *testing.T) {
		router := gin.New()
		router.Use(func(c *gin.Context) {
			c.Set("userID", "00000000-0000-0000-0000-000000000000")
			c.Next()
		})
		router.POST("/api/v1/organizations/:orgId/sites/:siteId/users/:userId/sessions/shift/:shiftId/message", handler.SendMessage)

		mockSession := &model.ShiftAgentSession{
			ID:              "session-uuid-1",
			AssigneeID:      "00000000-0000-0000-0000-000000000000",
			ShiftInstanceID: "shift-instance-uuid",
			MessageHistory:  model.JSONB("[]"),
			SessionContext:  model.JSONB(`{"user_name":"Hanna (Mock)"}`),
		}

		shiftSvc.InitializeShiftFunc = func(ctx context.Context, userID, shiftInstanceID string) (*model.ShiftAgentSession, error) {
			return mockSession, nil
		}

		var updateCaptured *model.ShiftAgentSession
		shiftSvc.UpdateSessionFunc = func(ctx context.Context, s *model.ShiftAgentSession) error {
			updateCaptured = s
			return nil
		}

		reqBody := `{"message": "check the DFW Airport live weather forecasts please"}`
		req, _ := http.NewRequest(http.MethodPost, "/api/v1/organizations/33333333-3333-3333-3333-333333333333/sites/55555555-5555-5555-5555-555555550000/users/00000000-0000-0000-0000-000000000000/sessions/shift/shift-instance-uuid/message", bytes.NewBufferString(reqBody))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var resp api.MessageResponse
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		assert.NoError(t, err)

		// Assert response contains dynamic weather blocks A2UI triggers
		assert.Equal(t, "WEATHER", resp.A2UIType)
		assert.Contains(t, resp.Content, "Live barometrics wind patterns resolved")
		
		a2uiMap, ok := resp.A2UIData.(map[string]interface{})
		assert.True(t, ok)

		beginRendering := a2uiMap["beginRendering"].(map[string]interface{})
		assert.Equal(t, "surface_weather", beginRendering["surfaceId"])
		rootID := beginRendering["root"].(string)

		surfaceUpdate := a2uiMap["surfaceUpdate"].(map[string]interface{})
		assert.Equal(t, "surface_weather", surfaceUpdate["surfaceId"])

		components := surfaceUpdate["components"].([]interface{})
		assert.NotEmpty(t, components)

		// Find the root card component in the registry
		var rootComponent map[string]interface{}
		for _, c := range components {
			comp := c.(map[string]interface{})
			if comp["id"] == rootID {
				rootComponent = comp
				break
			}
		}
		assert.NotNil(t, rootComponent)

		cardProps := rootComponent["component"].(map[string]interface{})["Card"].(map[string]interface{})
		assert.Equal(t, "METAR AIRPORT WIND AUDIT (KDFW)", cardProps["title"])

		// Verify database storage hooks updated and logged conversation history
		assert.NotNil(t, updateCaptured)
		assert.Contains(t, string(updateCaptured.MessageHistory), "WEATHER")
		assert.Contains(t, string(updateCaptured.MessageHistory), "DFW Airport")
	})

	t.Run("successfully routes map query intents", func(t *testing.T) {
		router := gin.New()
		router.Use(func(c *gin.Context) {
			c.Set("userID", "00000000-0000-0000-0000-000000000000")
			c.Next()
		})
		router.POST("/api/v1/organizations/:orgId/sites/:siteId/users/:userId/sessions/shift/:shiftId/message", handler.SendMessage)

		mockSession := &model.ShiftAgentSession{
			ID:              "session-uuid-1",
			AssigneeID:      "00000000-0000-0000-0000-000000000000",
			ShiftInstanceID: "shift-instance-uuid",
			MessageHistory:  model.JSONB("[]"),
			SessionContext:  model.JSONB(`{"user_name":"Hanna (Mock)"}`),
		}

		shiftSvc.InitializeShiftFunc = func(ctx context.Context, userID, shiftInstanceID string) (*model.ShiftAgentSession, error) {
			return mockSession, nil
		}

		var updateCaptured *model.ShiftAgentSession
		shiftSvc.UpdateSessionFunc = func(ctx context.Context, s *model.ShiftAgentSession) error {
			updateCaptured = s
			return nil
		}

		// SF site has boutique layout (44444444-4444-4444-4444-444444440001)
		reqBody := `{"message": "show me the store blueprint map and find the cash vault location"}`
		req, _ := http.NewRequest(http.MethodPost, "/api/v1/organizations/33333333-3333-3333-3333-333333333333/sites/44444444-4444-4444-4444-444444440001/users/00000000-0000-0000-0000-000000000000/sessions/shift/shift-instance-uuid/message", bytes.NewBufferString(reqBody))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var resp api.MessageResponse
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		assert.NoError(t, err)

		// Assert response contains dynamic weather blocks A2UI triggers
		assert.Equal(t, "STORE_LAYOUT", resp.A2UIType)
		assert.Contains(t, resp.Content, "Here is the interactive store layout blueprint map")
		
		a2uiMap, ok := resp.A2UIData.(map[string]interface{})
		assert.True(t, ok)

		beginRendering := a2uiMap["beginRendering"].(map[string]interface{})
		assert.Equal(t, "surface_store_layout", beginRendering["surfaceId"])
		rootID := beginRendering["root"].(string)
		assert.NotEmpty(t, rootID)

		surfaceUpdate := a2uiMap["surfaceUpdate"].(map[string]interface{})
		assert.Equal(t, "surface_store_layout", surfaceUpdate["surfaceId"])

		components := surfaceUpdate["components"].([]interface{})
		assert.NotEmpty(t, components)

		// Verify there is an Image canvas component with layout boutique and coordinates
		var canvasComponent map[string]interface{}
		for _, c := range components {
			comp := c.(map[string]interface{})
			wrapper := comp["component"].(map[string]interface{})
			if img, ok := wrapper["Image"].(map[string]interface{}); ok {
				canvasComponent = img
				break
			}
		}
		assert.NotNil(t, canvasComponent)
		urlVal := canvasComponent["url"].(map[string]interface{})["literalString"].(string)
		assert.Contains(t, urlVal, "/api/v1/blueprint?layout=boutique")
		assert.Contains(t, urlVal, "&x=175&y=25")

		// Verify database storage hooks updated and logged conversation history
		assert.NotNil(t, updateCaptured)
		assert.Contains(t, string(updateCaptured.MessageHistory), "STORE_LAYOUT")
	})
}

