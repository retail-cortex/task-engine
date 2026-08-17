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
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/rmcguinness/gemini_task_engine/pkg/model"
	"github.com/rmcguinness/gemini_task_engine/pkg/service"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"gorm.io/gorm/schema"
)

type mockOpTaskForAPI struct {
	service.TaskService
}

func (m *mockOpTaskForAPI) ListActiveSites(ctx context.Context) ([]*model.Site, error) {
	if ctx.Value("err") != nil {
		return nil, errors.New("err")
	}
	return []*model.Site{{ID: "s1"}}, nil
}
func (m *mockOpTaskForAPI) GetQueue(ctx context.Context, siteID string) ([]*model.TaskExecution, error) {
	if siteID == "err" {
		return nil, errors.New("err")
	}
	return []*model.TaskExecution{{ID: "e1"}}, nil
}
func (m *mockOpTaskForAPI) GetOrgTasks(ctx context.Context, orgID string) ([]*model.TaskExecution, error) {
	if orgID == "err" {
		return nil, errors.New("err")
	}
	return []*model.TaskExecution{{ID: "e1"}}, nil
}
func (m *mockOpTaskForAPI) GetSiteTasks(ctx context.Context, siteID string) ([]*model.TaskExecution, error) {
	if siteID == "err" {
		return nil, errors.New("err")
	}
	return []*model.TaskExecution{{ID: "e1"}}, nil
}
func (m *mockOpTaskForAPI) GetUserSiteTasks(ctx context.Context, siteID, userID string) ([]*model.TaskExecution, error) {
	if siteID == "err" {
		return nil, errors.New("err")
	}
	return []*model.TaskExecution{{ID: "e1"}}, nil
}
func (m *mockOpTaskForAPI) UpdateStatus(ctx context.Context, execID, status, state, userID string) error {
	if execID == "err" {
		return errors.New("err")
	}
	return nil
}
func (m *mockOpTaskForAPI) OverrideAssetConstraint(ctx context.Context, execID, assetID, just, userID string) error {
	if execID == "err" {
		return errors.New("err")
	}
	return nil
}
func (m *mockOpTaskForAPI) ProposeTrade(ctx context.Context, execID, propID, userID string) error {
	if execID == "err" {
		return errors.New("err")
	}
	return nil
}
func (m *mockOpTaskForAPI) AcceptTrade(ctx context.Context, tradeID, userID string) error {
	if tradeID == "err" {
		return errors.New("err")
	}
	return nil
}
func (m *mockOpTaskForAPI) RejectTrade(ctx context.Context, tradeID, userID string) error {
	if tradeID == "err" {
		return errors.New("err")
	}
	return nil
}
func (m *mockOpTaskForAPI) ListPendingTrades(ctx context.Context, userID string) ([]*model.TaskTrade, error) {
	if userID == "err" {
		return nil, errors.New("err")
	}
	return []*model.TaskTrade{{ID: "tr1"}}, nil
}
func (m *mockOpTaskForAPI) ClaimTask(ctx context.Context, execID, userID string, roles []string) error {
	if execID == "err" {
		return errors.New("err")
	}
	return nil
}

type mockOpShiftForAPI struct {
	service.ShiftService
}

func (m *mockOpShiftForAPI) GetUserProfile(ctx context.Context, userID string) (*model.User, error) {
	if userID == "err" {
		return nil, errors.New("err")
	}
	return &model.User{ID: userID, Roles: []model.Role{{ID: "r1"}}}, nil
}
func (m *mockOpShiftForAPI) ListActiveOnShiftUsers(ctx context.Context, siteID string) ([]*model.User, error) {
	if siteID == "err" {
		return nil, errors.New("err")
	}
	return []*model.User{{ID: "u1"}}, nil
}

type mockOpAutomationForAPI struct {
	service.AutomationService
}

func (m *mockOpAutomationForAPI) TriggerStreamingEvent(ctx context.Context, siteID, orgID string, et model.EventType, desc string) (*model.TaskExecution, error) {
	if siteID == "err" {
		return nil, errors.New("err")
	}
	return &model.TaskExecution{ID: "e1"}, nil
}
func (m *mockOpAutomationForAPI) ListTemplates(ctx context.Context) ([]*model.Task, error) {
	return []*model.Task{
		{ID: "d000fa44-0000-0000-0000-000000000001", Name: "Drop Alert", TaskType: "ADHOC"},
		{ID: "d000fa44-0000-0000-0000-000000000002", Name: "Stockout Alert", TaskType: "ADHOC"},
		{ID: "d000fa44-0000-0000-0000-000000000003", Name: "Customer Alert", TaskType: "ADHOC"},
		{ID: "d000fa44-0000-0000-0000-000000000004", Name: "DSD Alert", TaskType: "ADHOC"},
	}, nil
}

func TestOperationalHandler_AllEndpoints(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := NewOperationalHandler(
		&mockOpTaskForAPI{},
		&mockOpShiftForAPI{},
		&mockOpAutomationForAPI{},
		Config{},
		nil,
	)

	// Liveness & Startup
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	handler.Liveness(c)
	assert.Equal(t, http.StatusOK, w.Code)

	w = httptest.NewRecorder()
	c, _ = gin.CreateTestContext(w)
	handler.Startup(c)
	assert.Equal(t, http.StatusOK, w.Code)

	// Readiness without db
	w = httptest.NewRecorder()
	c, _ = gin.CreateTestContext(w)
	handler.Readiness(c)
	assert.Equal(t, http.StatusServiceUnavailable, w.Code)

	// GetMe
	w = httptest.NewRecorder()
	c, _ = gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest("GET", "/me", nil)
	c.Set("userID", "u1")
	handler.GetMe(c)
	assert.Equal(t, http.StatusOK, w.Code)

	// ListSites
	w = httptest.NewRecorder()
	c, _ = gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest("GET", "/sites", nil)
	handler.ListSites(c)
	assert.Equal(t, http.StatusOK, w.Code)

	// GetOrgTasks
	w = httptest.NewRecorder()
	c, _ = gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest("GET", "/orgs/org1/tasks", nil)
	c.Params = []gin.Param{{Key: "orgId", Value: "org1"}}
	handler.GetOrgTasks(c)
	assert.Equal(t, http.StatusOK, w.Code)

	// GetSiteTasks
	w = httptest.NewRecorder()
	c, _ = gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest("GET", "/sites/s1/tasks", nil)
	c.Params = []gin.Param{{Key: "siteId", Value: "s1"}}
	handler.GetSiteTasks(c)
	assert.Equal(t, http.StatusOK, w.Code)

	// GetSiteAssociates
	w = httptest.NewRecorder()
	c, _ = gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest("GET", "/sites/s1/associates", nil)
	c.Params = []gin.Param{{Key: "siteId", Value: "s1"}}
	handler.GetSiteAssociates(c)
	assert.Equal(t, http.StatusOK, w.Code)

	// GetUserSiteTasks
	w = httptest.NewRecorder()
	c, _ = gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest("GET", "/sites/s1/users/u1/tasks", nil)
	c.Params = []gin.Param{{Key: "siteId", Value: "s1"}, {Key: "userId", Value: "u1"}}
	handler.GetUserSiteTasks(c)
	assert.Equal(t, http.StatusOK, w.Code)

	// UpdateTaskStatus
	w = httptest.NewRecorder()
	c, _ = gin.CreateTestContext(w)
	c.Set("userID", "u1")
	c.Params = []gin.Param{{Key: "id", Value: "e1"}}
	c.Request, _ = http.NewRequest("PUT", "/tasks/e1/status", bytes.NewBuffer([]byte(`{"status":"COMPLETED"}`)))
	handler.UpdateTaskStatus(c)
	assert.Equal(t, http.StatusOK, w.Code)

	// OverrideAsset
	w = httptest.NewRecorder()
	c, _ = gin.CreateTestContext(w)
	c.Set("userID", "u1")
	c.Params = []gin.Param{{Key: "id", Value: "e1"}}
	c.Request, _ = http.NewRequest("POST", "/tasks/e1/override-asset", bytes.NewBuffer([]byte(`{"asset_id":"a1","justification":"broken"}`)))
	handler.OverrideAsset(c)
	assert.Equal(t, http.StatusOK, w.Code)

	// ProposeTrade
	w = httptest.NewRecorder()
	c, _ = gin.CreateTestContext(w)
	c.Set("userID", "u1")
	c.Request, _ = http.NewRequest("POST", "/trades/propose", bytes.NewBuffer([]byte(`{"task_execution_id":"e1","proposed_assignee_id":"u2"}`)))
	handler.ProposeTrade(c)
	assert.Equal(t, http.StatusCreated, w.Code)

	// AcceptTrade
	w = httptest.NewRecorder()
	c, _ = gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest("POST", "/trades/tr1/accept", nil)
	c.Set("userID", "u1")
	c.Params = []gin.Param{{Key: "tradeId", Value: "tr1"}}
	handler.AcceptTrade(c)
	assert.Equal(t, http.StatusOK, w.Code)

	// RejectTrade
	w = httptest.NewRecorder()
	c, _ = gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest("POST", "/trades/tr1/reject", nil)
	c.Set("userID", "u1")
	c.Params = []gin.Param{{Key: "tradeId", Value: "tr1"}}
	handler.RejectTrade(c)
	assert.Equal(t, http.StatusOK, w.Code)

	// ListPendingTrades
	w = httptest.NewRecorder()
	c, _ = gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest("GET", "/trades/pending", nil)
	c.Set("userID", "u1")
	handler.ListPendingTrades(c)
	assert.Equal(t, http.StatusOK, w.Code)

	// TriggerAlert
	w = httptest.NewRecorder()
	c, _ = gin.CreateTestContext(w)
	c.Params = []gin.Param{{Key: "siteId", Value: "s1"}}
	c.Request, _ = http.NewRequest("POST", "/sites/s1/alerts", bytes.NewBuffer([]byte(`{"organizer_id":"org1","event_type":"EventStockoutCorrect","description":"alert"}`)))
	handler.TriggerAlert(c)
	assert.Equal(t, http.StatusCreated, w.Code)

	// ClaimTask
	w = httptest.NewRecorder()
	c, _ = gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest("POST", "/tasks/e1/claim", nil)
	c.Set("userID", "u1")
	c.Params = []gin.Param{{Key: "id", Value: "e1"}}
	handler.ClaimTask(c)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestOperationalHandler_ErrorBranches(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := NewOperationalHandler(&mockOpTaskForAPI{}, &mockOpShiftForAPI{}, &mockOpAutomationForAPI{}, Config{}, nil)

	// GetMe unauthorized
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest("GET", "/me", nil)
	handler.GetMe(c)
	assert.Equal(t, http.StatusUnauthorized, w.Code)

	// ListSites error
	w = httptest.NewRecorder()
	c, _ = gin.CreateTestContext(w)
	reqErr, _ := http.NewRequestWithContext(context.WithValue(context.Background(), "err", true), "GET", "/sites", nil)
	c.Request = reqErr
	handler.ListSites(c)
	assert.Equal(t, http.StatusInternalServerError, w.Code)

	// Missing path params
	missingParams := []gin.HandlerFunc{
		handler.GetOrgTasks, handler.GetSiteTasks, handler.GetSiteAssociates, handler.GetUserSiteTasks,
	}
	for _, fn := range missingParams {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request, _ = http.NewRequest("GET", "/", nil)
		fn(c)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	}

	// GET 500 errors with param "err"
	w = httptest.NewRecorder()
	c, _ = gin.CreateTestContext(w)
	c.Params = []gin.Param{{Key: "orgId", Value: "err"}}
	c.Request, _ = http.NewRequest("GET", "/", nil)
	handler.GetOrgTasks(c)
	assert.Equal(t, http.StatusInternalServerError, w.Code)

	w = httptest.NewRecorder()
	c, _ = gin.CreateTestContext(w)
	c.Params = []gin.Param{{Key: "siteId", Value: "err"}}
	c.Request, _ = http.NewRequest("GET", "/", nil)
	handler.GetSiteTasks(c)
	assert.Equal(t, http.StatusInternalServerError, w.Code)

	w = httptest.NewRecorder()
	c, _ = gin.CreateTestContext(w)
	c.Params = []gin.Param{{Key: "siteId", Value: "err"}}
	c.Request, _ = http.NewRequest("GET", "/", nil)
	handler.GetSiteAssociates(c)
	assert.Equal(t, http.StatusInternalServerError, w.Code)

	// Invalid JSON POST/PUT
	badJSONs := []gin.HandlerFunc{
		handler.UpdateTaskStatus, handler.OverrideAsset, handler.ProposeTrade, handler.TriggerAlert,
	}
	for _, fn := range badJSONs {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Set("userID", "u1")
		c.Request, _ = http.NewRequest("POST", "/", bytes.NewBuffer([]byte(`{bad`)))
		fn(c)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	}

	// POST 500 errors with id "err"
	w = httptest.NewRecorder()
	c, _ = gin.CreateTestContext(w)
	c.Set("userID", "u1")
	c.Params = []gin.Param{{Key: "id", Value: "err"}}
	c.Request, _ = http.NewRequest("PUT", "/", bytes.NewBuffer([]byte(`{"status":"COMPLETED"}`)))
	handler.UpdateTaskStatus(c)
	assert.Equal(t, http.StatusInternalServerError, w.Code)

	w = httptest.NewRecorder()
	c, _ = gin.CreateTestContext(w)
	c.Set("userID", "u1")
	c.Params = []gin.Param{{Key: "id", Value: "err"}}
	c.Request, _ = http.NewRequest("POST", "/", bytes.NewBuffer([]byte(`{"asset_id":"a1","justification":"x"}`)))
	handler.OverrideAsset(c)
	assert.Equal(t, http.StatusInternalServerError, w.Code)

	w = httptest.NewRecorder()
	c, _ = gin.CreateTestContext(w)
	c.Set("userID", "u1")
	c.Request, _ = http.NewRequest("POST", "/", bytes.NewBuffer([]byte(`{"task_execution_id":"err","proposed_assignee_id":"u2"}`)))
	handler.ProposeTrade(c)
	assert.Equal(t, http.StatusInternalServerError, w.Code)

	w = httptest.NewRecorder()
	c, _ = gin.CreateTestContext(w)
	c.Set("userID", "u1")
	c.Params = []gin.Param{{Key: "tradeId", Value: "err"}}
	c.Request, _ = http.NewRequest("POST", "/", nil)
	handler.AcceptTrade(c)
	assert.Equal(t, http.StatusInternalServerError, w.Code)

	w = httptest.NewRecorder()
	c, _ = gin.CreateTestContext(w)
	c.Set("userID", "u1")
	c.Params = []gin.Param{{Key: "tradeId", Value: "err"}}
	c.Request, _ = http.NewRequest("POST", "/", nil)
	handler.RejectTrade(c)
	assert.Equal(t, http.StatusInternalServerError, w.Code)

	w = httptest.NewRecorder()
	c, _ = gin.CreateTestContext(w)
	c.Set("userID", "u1")
	c.Params = []gin.Param{{Key: "siteId", Value: "err"}}
	c.Request, _ = http.NewRequest("POST", "/", bytes.NewBuffer([]byte(`{"organizer_id":"org1","event_type":"EventStockoutCorrect","description":"d"}`)))
	handler.TriggerAlert(c)
	assert.Equal(t, http.StatusInternalServerError, w.Code)

	w = httptest.NewRecorder()
	c, _ = gin.CreateTestContext(w)
	c.Set("userID", "err")
	c.Request, _ = http.NewRequest("GET", "/trades/pending", nil)
	handler.ListPendingTrades(c)
	assert.Equal(t, http.StatusInternalServerError, w.Code)

	// Readiness with db
	db, _ := gorm.Open(dummyDialector{}, &gorm.Config{DryRun: true})
	handlerWithDB := NewOperationalHandler(&mockOpTaskForAPI{}, &mockOpShiftForAPI{}, &mockOpAutomationForAPI{}, Config{}, db)
	w = httptest.NewRecorder()
	c, _ = gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest("GET", "/readiness", nil)
	handlerWithDB.Readiness(c)
	assert.Equal(t, http.StatusServiceUnavailable, w.Code) // ping returns err on dummy sqlDB or nil

	// ClaimTask 401
	w = httptest.NewRecorder()
	c, _ = gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest("POST", "/tasks/e1/claim", nil)
	handler.ClaimTask(c)
	assert.Equal(t, http.StatusUnauthorized, w.Code)

	// ClaimTask 500 GetUserProfile failed
	w = httptest.NewRecorder()
	c, _ = gin.CreateTestContext(w)
	c.Set("userID", "err")
	c.Request, _ = http.NewRequest("POST", "/tasks/e1/claim", nil)
	handler.ClaimTask(c)
	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

type dummyDialector struct{}

func (d dummyDialector) Name() string                                         { return "dummy" }
func (d dummyDialector) Initialize(db *gorm.DB) error                         { return nil }
func (d dummyDialector) Migrator(db *gorm.DB) gorm.Migrator                   { return nil }
func (d dummyDialector) DataTypeOf(field *schema.Field) string                { return "text" }
func (d dummyDialector) DefaultValueOf(field *schema.Field) clause.Expression { return nil }
func (d dummyDialector) BindVarTo(writer clause.Writer, stmt *gorm.Statement, v interface{}) {}
func (d dummyDialector) QuoteTo(writer clause.Writer, str string)             { writer.WriteString(str) }
func (d dummyDialector) Explain(sql string, vars ...interface{}) string       { return sql }
