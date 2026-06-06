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
		assert.Equal(t, "card", a2uiMap["type"])
		assert.Equal(t, "METAR AIRPORT WIND AUDIT (KDFW)", a2uiMap["title"])
		assert.Equal(t, "standard", a2uiMap["style"])

		children, ok := a2uiMap["children"].([]interface{})
		assert.True(t, ok)
		assert.Len(t, children, 1)

		tableChild, ok := children[0].(map[string]interface{})
		assert.True(t, ok)
		assert.Equal(t, "table", tableChild["type"])

		rows, ok := tableChild["rows"].([]interface{})
		assert.True(t, ok)
		assert.Len(t, rows, 5)

		row0, ok := rows[0].(map[string]interface{})
		assert.True(t, ok)
		assert.Equal(t, "Station", row0["label"])
		assert.Equal(t, "KDFW", row0["value"])

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
		assert.Equal(t, "card", a2uiMap["type"])
		assert.Equal(t, "STORE SPATIAL BLUEPRINT MAP", a2uiMap["title"])

		children, ok := a2uiMap["children"].([]interface{})
		assert.True(t, ok)
		assert.Len(t, children, 4)

		textChild, ok := children[0].(map[string]interface{})
		assert.True(t, ok)
		assert.Equal(t, "text", textChild["type"])

		canvasChild, ok := children[1].(map[string]interface{})
		assert.True(t, ok)
		assert.Equal(t, "canvas", canvasChild["type"])
		assert.Equal(t, "boutique", canvasChild["layout"])

		beacon, ok := canvasChild["beacon"].(map[string]interface{})
		assert.True(t, ok)
		assert.Equal(t, float64(175), beacon["x"])
		assert.Equal(t, float64(25), beacon["y"])
		assert.Equal(t, "Secure Back-Office Cash Vault", beacon["name"])

		tableChild, ok := children[2].(map[string]interface{})
		assert.True(t, ok)
		assert.Equal(t, "table", tableChild["type"])

		rows, ok := tableChild["rows"].([]interface{})
		assert.True(t, ok)
		assert.Len(t, rows, 3)

		rowVal, ok := rows[1].(map[string]interface{})
		assert.True(t, ok)
		assert.Equal(t, "Focal Highlight Beacon", rowVal["label"])
		assert.Equal(t, "Secure Back-Office Cash Vault", rowVal["value"])

		buttonRow, ok := children[3].(map[string]interface{})
		assert.True(t, ok)
		assert.Equal(t, "row", buttonRow["type"])

		// Verify database storage hooks updated and logged conversation history
		assert.NotNil(t, updateCaptured)
		assert.Contains(t, string(updateCaptured.MessageHistory), "STORE_LAYOUT")
	})
}

