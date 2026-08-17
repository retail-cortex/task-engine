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
)

type mockAdminServiceForAPI struct{}

func (m *mockAdminServiceForAPI) RegisterUser(ctx context.Context, user *model.User) error {
	if user.Name == "err" || user.Email == "reg-err@google.com" {
		return errors.New("err")
	}
	user.ID = "u-new"
	return nil
}
func (m *mockAdminServiceForAPI) GetUserByID(ctx context.Context, id string) (*model.User, error) {
	if id == "err" {
		return nil, errors.New("err")
	}
	return &model.User{ID: id}, nil
}
func (m *mockAdminServiceForAPI) UpdateUser(ctx context.Context, user *model.User) error {
	if user.Name == "err" {
		return errors.New("err")
	}
	return nil
}
func (m *mockAdminServiceForAPI) DeleteUser(ctx context.Context, id string) error {
	if id == "err" {
		return errors.New("err")
	}
	return nil
}
func (m *mockAdminServiceForAPI) ListUsers(ctx context.Context) ([]*model.User, error) {
	if ctx.Value("match-email") != nil {
		return []*model.User{{ID: "u2", Email: "new@google.com"}}, nil
	}
	return []*model.User{{ID: "u1"}}, nil
}
func (m *mockAdminServiceForAPI) ListUsersRange(ctx context.Context, o, l int) ([]*model.User, error) {
	if o == 999 {
		return nil, errors.New("err")
	}
	return []*model.User{{ID: "u1"}}, nil
}
func (m *mockAdminServiceForAPI) AssignRole(ctx context.Context, u, r string) error {
	if u == "err" {
		return errors.New("err")
	}
	return nil
}
func (m *mockAdminServiceForAPI) FindUserByOAuth(ctx context.Context, p, o string) (*model.User, error) {
	if o == "sub-notfound" || o == "sub-bind" {
		return nil, gorm.ErrRecordNotFound
	}
	if o == "sub-err" {
		return nil, errors.New("err")
	}
	return &model.User{ID: "u1", Email: "user@google.com"}, nil
}

func (m *mockAdminServiceForAPI) CreateRole(ctx context.Context, r *model.Role) error {
	if r.Name == "err" {
		return errors.New("err")
	}
	return nil
}
func (m *mockAdminServiceForAPI) GetRoleByID(ctx context.Context, id string) (*model.Role, error) {
	if id == "err" {
		return nil, errors.New("err")
	}
	return &model.Role{ID: id}, nil
}
func (m *mockAdminServiceForAPI) UpdateRole(ctx context.Context, r *model.Role) error {
	if r.Name == "err" {
		return errors.New("err")
	}
	return nil
}
func (m *mockAdminServiceForAPI) DeleteRole(ctx context.Context, id string) error {
	if id == "err" {
		return errors.New("err")
	}
	return nil
}
func (m *mockAdminServiceForAPI) ListRoles(ctx context.Context) ([]*model.Role, error) {
	return []*model.Role{{ID: "r1"}}, nil
}
func (m *mockAdminServiceForAPI) ListRolesRange(ctx context.Context, o, l int) ([]*model.Role, error) {
	if o == 999 {
		return nil, errors.New("err")
	}
	return []*model.Role{{ID: "r1"}}, nil
}

func (m *mockAdminServiceForAPI) RegisterOrganization(ctx context.Context, org *model.Organization) error {
	if org.Name == "err" {
		return errors.New("err")
	}
	return nil
}
func (m *mockAdminServiceForAPI) GetOrganizationByID(ctx context.Context, id string) (*model.Organization, error) {
	if id == "err" {
		return nil, errors.New("err")
	}
	return &model.Organization{ID: id}, nil
}
func (m *mockAdminServiceForAPI) UpdateOrganization(ctx context.Context, org *model.Organization) error {
	if org.Name == "err" {
		return errors.New("err")
	}
	return nil
}
func (m *mockAdminServiceForAPI) DeleteOrganization(ctx context.Context, id string) error {
	if id == "err" {
		return errors.New("err")
	}
	return nil
}
func (m *mockAdminServiceForAPI) AssignUserToOrganization(ctx context.Context, o, u string) error {
	if o == "err" {
		return errors.New("err")
	}
	return nil
}
func (m *mockAdminServiceForAPI) ListOrganizations(ctx context.Context) ([]*model.Organization, error) { return []*model.Organization{{ID: "org1"}}, nil }
func (m *mockAdminServiceForAPI) ListOrganizationsRange(ctx context.Context, o, l int) ([]*model.Organization, error) {
	if o == 999 {
		return nil, errors.New("err")
	}
	return []*model.Organization{{ID: "org1"}}, nil
}

func (m *mockAdminServiceForAPI) RegisterSite(ctx context.Context, s *model.Site) error {
	if s.Name == "err" {
		return errors.New("err")
	}
	return nil
}
func (m *mockAdminServiceForAPI) GetSiteByID(ctx context.Context, id string) (*model.Site, error) {
	if id == "err" {
		return nil, errors.New("err")
	}
	return &model.Site{ID: id}, nil
}
func (m *mockAdminServiceForAPI) UpdateSite(ctx context.Context, s *model.Site) error {
	if s.Name == "err" {
		return errors.New("err")
	}
	return nil
}
func (m *mockAdminServiceForAPI) DeleteSite(ctx context.Context, id string) error {
	if id == "err" {
		return errors.New("err")
	}
	return nil
}
func (m *mockAdminServiceForAPI) ListSites(ctx context.Context) ([]*model.Site, error) { return []*model.Site{{ID: "s1"}}, nil }
func (m *mockAdminServiceForAPI) ListSitesRange(ctx context.Context, o, l int) ([]*model.Site, error) {
	if o == 999 {
		return nil, errors.New("err")
	}
	return []*model.Site{{ID: "s1"}}, nil
}

func (m *mockAdminServiceForAPI) RegisterLocation(ctx context.Context, loc *model.Location) error {
	if loc.Name == "err" {
		return errors.New("err")
	}
	return nil
}
func (m *mockAdminServiceForAPI) GetLocationByID(ctx context.Context, id string) (*model.Location, error) {
	if id == "err" {
		return nil, errors.New("err")
	}
	return &model.Location{ID: id}, nil
}
func (m *mockAdminServiceForAPI) UpdateLocation(ctx context.Context, loc *model.Location) error {
	if loc.Name == "err" {
		return errors.New("err")
	}
	return nil
}
func (m *mockAdminServiceForAPI) DeleteLocation(ctx context.Context, id string) error {
	if id == "err" {
		return errors.New("err")
	}
	return nil
}
func (m *mockAdminServiceForAPI) ListLocations(ctx context.Context) ([]*model.Location, error) { return []*model.Location{{ID: "l1"}}, nil }
func (m *mockAdminServiceForAPI) ListLocationsRange(ctx context.Context, o, l int) ([]*model.Location, error) {
	if o == 999 {
		return nil, errors.New("err")
	}
	return []*model.Location{{ID: "l1"}}, nil
}

func (m *mockAdminServiceForAPI) RegisterAsset(ctx context.Context, a *model.Asset) error {
	if a.Name == "err" {
		return errors.New("err")
	}
	return nil
}
func (m *mockAdminServiceForAPI) GetAssetByID(ctx context.Context, id string) (*model.Asset, error) {
	if id == "err" {
		return nil, errors.New("err")
	}
	return &model.Asset{ID: id}, nil
}
func (m *mockAdminServiceForAPI) UpdateAsset(ctx context.Context, a *model.Asset) error {
	if a.Name == "err" {
		return errors.New("err")
	}
	return nil
}
func (m *mockAdminServiceForAPI) DeleteAsset(ctx context.Context, id string) error {
	if id == "err" {
		return errors.New("err")
	}
	return nil
}
func (m *mockAdminServiceForAPI) ListAssets(ctx context.Context) ([]*model.Asset, error) { return []*model.Asset{{ID: "a1"}}, nil }
func (m *mockAdminServiceForAPI) ListAssetsRange(ctx context.Context, o, l int) ([]*model.Asset, error) {
	if o == 999 {
		return nil, errors.New("err")
	}
	return []*model.Asset{{ID: "a1"}}, nil
}

func (m *mockAdminServiceForAPI) CreateTaskTemplate(ctx context.Context, t *model.Task) error {
	if t.Name == "err" {
		return errors.New("err")
	}
	return nil
}
func (m *mockAdminServiceForAPI) GetTaskTemplateByID(ctx context.Context, id string) (*model.Task, error) {
	if id == "err" {
		return nil, errors.New("err")
	}
	return &model.Task{ID: id}, nil
}
func (m *mockAdminServiceForAPI) UpdateTaskTemplate(ctx context.Context, t *model.Task) error {
	if t.Name == "err" {
		return errors.New("err")
	}
	return nil
}
func (m *mockAdminServiceForAPI) DeleteTaskTemplate(ctx context.Context, id string) error {
	if id == "err" {
		return errors.New("err")
	}
	return nil
}
func (m *mockAdminServiceForAPI) ListTaskTemplates(ctx context.Context) ([]*model.Task, error) { return []*model.Task{{ID: "t1"}}, nil }
func (m *mockAdminServiceForAPI) ListTaskTemplatesRange(ctx context.Context, o, l int) ([]*model.Task, error) {
	if o == 999 {
		return nil, errors.New("err")
	}
	return []*model.Task{{ID: "t1"}}, nil
}

type mockSchedulerForAPI struct {
	service.SchedulerService
}

func (m *mockSchedulerForAPI) ForceTriggerBatchSweep(ctx context.Context) error {
	if ctx.Value("err") != nil {
		return errors.New("err")
	}
	return nil
}
func (m *mockSchedulerForAPI) GetStatus() service.SchedulerStatus {
	return service.SchedulerStatus{NodeID: "node-1"}
}

type mockTaskForAPI struct {
	service.TaskService
}

func (m *mockTaskForAPI) ListTaskExecutions(ctx context.Context) ([]*model.TaskExecution, error) {
	return []*model.TaskExecution{{ID: "e1"}}, nil
}
func (m *mockTaskForAPI) ListTaskExecutionsRange(ctx context.Context, o, l int) ([]*model.TaskExecution, error) {
	if o == 999 {
		return nil, errors.New("err")
	}
	return []*model.TaskExecution{{ID: "e1"}}, nil
}
func (m *mockTaskForAPI) GetTaskExecutionByID(ctx context.Context, id string) (*model.TaskExecution, error) {
	if id == "err" {
		return nil, errors.New("err")
	}
	return &model.TaskExecution{ID: id}, nil
}
func (m *mockTaskForAPI) DeleteTaskExecution(ctx context.Context, id string) error {
	if id == "err" {
		return errors.New("err")
	}
	return nil
}

type mockShiftForAPI struct {
	service.ShiftService
}

func (m *mockShiftForAPI) ListSessions(ctx context.Context) ([]*model.ShiftAgentSession, error) {
	return []*model.ShiftAgentSession{{ID: "sess1"}}, nil
}
func (m *mockShiftForAPI) ListSessionsRange(ctx context.Context, o, l int) ([]*model.ShiftAgentSession, error) {
	if o == 999 {
		return nil, errors.New("err")
	}
	return []*model.ShiftAgentSession{{ID: "sess1"}}, nil
}
func (m *mockShiftForAPI) GetSessionByID(ctx context.Context, id string) (*model.ShiftAgentSession, error) {
	if id == "err" {
		return nil, errors.New("err")
	}
	return &model.ShiftAgentSession{ID: id}, nil
}
func (m *mockShiftForAPI) DeleteSession(ctx context.Context, id string) error {
	if id == "err" {
		return errors.New("err")
	}
	return nil
}

type mockRAGForAPI struct {
	service.RAGService
}

func (m *mockRAGForAPI) QuerySimilarity(ctx context.Context, v model.Float32Vector, topK int) ([]*model.SOPChunk, error) {
	return []*model.SOPChunk{
		{
			ID:           "c1",
			SOPProcessID: "proc1",
			Content:      "Verify refrigeration chiller temp is under 38F every 4 hours.",
		},
	}, nil
}
func (m *mockRAGForAPI) ListSOPs(ctx context.Context) ([]*model.SOP, error) {
	return []*model.SOP{{ID: "sop1"}}, nil
}
func (m *mockRAGForAPI) ListSOPsRange(ctx context.Context, o, l int) ([]*model.SOP, error) {
	if o == 999 {
		return nil, errors.New("err")
	}
	return []*model.SOP{{ID: "sop1"}}, nil
}
func (m *mockRAGForAPI) GetSOPByID(ctx context.Context, id string) (*model.SOP, error) {
	if id == "err" {
		return nil, errors.New("err")
	}
	return &model.SOP{ID: id}, nil
}
func (m *mockRAGForAPI) DeleteSOP(ctx context.Context, id string) error {
	if id == "err" {
		return errors.New("err")
	}
	return nil
}
func (m *mockRAGForAPI) ListProcesses(ctx context.Context) ([]*model.SOPProcess, error) {
	return []*model.SOPProcess{{ID: "p1"}}, nil
}
func (m *mockRAGForAPI) ListProcessesRange(ctx context.Context, o, l int) ([]*model.SOPProcess, error) {
	if o == 999 {
		return nil, errors.New("err")
	}
	return []*model.SOPProcess{{ID: "p1"}}, nil
}
func (m *mockRAGForAPI) GetProcessByID(ctx context.Context, id string) (*model.SOPProcess, error) {
	if id == "err" {
		return nil, errors.New("err")
	}
	return &model.SOPProcess{ID: id}, nil
}
func (m *mockRAGForAPI) DeleteProcess(ctx context.Context, id string) error {
	if id == "err" {
		return errors.New("err")
	}
	return nil
}

func TestAdminHandler_AllEndpoints(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := NewAdminHandler(
		&mockAdminServiceForAPI{},
		&mockSchedulerForAPI{},
		&mockTaskForAPI{},
		&mockShiftForAPI{},
		&mockRAGForAPI{},
	)

	reqJSON := []byte(`{"name":"test"}`)

	// Users
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest("GET", "/users", nil)
	handler.ListUsers(c)
	assert.Equal(t, http.StatusOK, w.Code)

	w = httptest.NewRecorder()
	c, _ = gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest("GET", "/users?offset=0&limit=10", nil)
	handler.ListUsers(c)
	assert.Equal(t, http.StatusOK, w.Code)

	w = httptest.NewRecorder()
	c, _ = gin.CreateTestContext(w)
	c.Params = []gin.Param{{Key: "id", Value: "u1"}}
	c.Request, _ = http.NewRequest("GET", "/users/u1", nil)
	handler.GetUser(c)
	assert.Equal(t, http.StatusOK, w.Code)

	w = httptest.NewRecorder()
	c, _ = gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest("POST", "/users", bytes.NewBuffer([]byte(`{"name":"test"}`)))
	handler.CreateUser(c)
	assert.Equal(t, http.StatusCreated, w.Code)

	w = httptest.NewRecorder()
	c, _ = gin.CreateTestContext(w)
	c.Params = []gin.Param{{Key: "id", Value: "u1"}}
	c.Request, _ = http.NewRequest("PUT", "/users/u1", bytes.NewBuffer([]byte(`{"name":"test"}`)))
	handler.UpdateUser(c)
	assert.Equal(t, http.StatusOK, w.Code)

	w = httptest.NewRecorder()
	c, _ = gin.CreateTestContext(w)
	c.Params = []gin.Param{{Key: "id", Value: "u1"}}
	c.Request, _ = http.NewRequest("DELETE", "/users/u1", nil)
	handler.DeleteUser(c)
	assert.Equal(t, http.StatusOK, w.Code)

	w = httptest.NewRecorder()
	c, _ = gin.CreateTestContext(w)
	c.Params = []gin.Param{{Key: "id", Value: "u1"}}
	c.Request, _ = http.NewRequest("PUT", "/users/u1/roles", bytes.NewBuffer([]byte(`{"role_id":"r1"}`)))
	handler.AssignRole(c)
	assert.Equal(t, http.StatusOK, w.Code)

	// Roles
	w = httptest.NewRecorder()
	c, _ = gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest("GET", "/roles", nil)
	handler.ListRoles(c)
	assert.Equal(t, http.StatusOK, w.Code)

	w = httptest.NewRecorder()
	c, _ = gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest("GET", "/roles?offset=0&limit=10", nil)
	handler.ListRoles(c)
	assert.Equal(t, http.StatusOK, w.Code)

	w = httptest.NewRecorder()
	c, _ = gin.CreateTestContext(w)
	c.Params = []gin.Param{{Key: "id", Value: "r1"}}
	c.Request, _ = http.NewRequest("GET", "/roles/r1", nil)
	handler.GetRole(c)
	assert.Equal(t, http.StatusOK, w.Code)

	w = httptest.NewRecorder()
	c, _ = gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest("POST", "/roles", bytes.NewBuffer([]byte(`{"name":"test"}`)))
	handler.CreateRole(c)
	assert.Equal(t, http.StatusCreated, w.Code)

	w = httptest.NewRecorder()
	c, _ = gin.CreateTestContext(w)
	c.Params = []gin.Param{{Key: "id", Value: "r1"}}
	c.Request, _ = http.NewRequest("PUT", "/roles/r1", bytes.NewBuffer([]byte(`{"name":"test"}`)))
	handler.UpdateRole(c)
	assert.Equal(t, http.StatusOK, w.Code)

	w = httptest.NewRecorder()
	c, _ = gin.CreateTestContext(w)
	c.Params = []gin.Param{{Key: "id", Value: "r1"}}
	c.Request, _ = http.NewRequest("DELETE", "/roles/r1", nil)
	handler.DeleteRole(c)
	assert.Equal(t, http.StatusOK, w.Code)

	// Organizations
	w = httptest.NewRecorder()
	c, _ = gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest("GET", "/organizations", nil)
	handler.ListOrganizations(c)
	assert.Equal(t, http.StatusOK, w.Code)

	w = httptest.NewRecorder()
	c, _ = gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest("GET", "/organizations?offset=0&limit=10", nil)
	handler.ListOrganizations(c)
	w = httptest.NewRecorder()
	c, _ = gin.CreateTestContext(w)
	c.Params = []gin.Param{{Key: "id", Value: "o1"}, {Key: "userId", Value: "u1"}}
	c.Request, _ = http.NewRequest("POST", "/organizations/o1/users/u1", nil)
	handler.AssignUserToOrganization(c)
	assert.Equal(t, http.StatusOK, w.Code)

	w = httptest.NewRecorder()
	c, _ = gin.CreateTestContext(w)
	c.Params = []gin.Param{{Key: "id", Value: "org1"}}
	c.Request, _ = http.NewRequest("GET", "/organizations/org1", nil)
	handler.GetOrganization(c)
	assert.Equal(t, http.StatusOK, w.Code)

	w = httptest.NewRecorder()
	c, _ = gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest("POST", "/organizations", bytes.NewBuffer([]byte(`{"name":"test"}`)))
	handler.CreateOrganization(c)
	assert.Equal(t, http.StatusCreated, w.Code)

	w = httptest.NewRecorder()
	c, _ = gin.CreateTestContext(w)
	c.Params = []gin.Param{{Key: "id", Value: "org1"}}
	c.Request, _ = http.NewRequest("PUT", "/organizations/org1", bytes.NewBuffer([]byte(`{"name":"test"}`)))
	handler.UpdateOrganization(c)
	assert.Equal(t, http.StatusOK, w.Code)

	w = httptest.NewRecorder()
	c, _ = gin.CreateTestContext(w)
	c.Params = []gin.Param{{Key: "id", Value: "org1"}}
	c.Request, _ = http.NewRequest("DELETE", "/organizations/org1", nil)
	handler.DeleteOrganization(c)
	assert.Equal(t, http.StatusOK, w.Code)

	// Sites
	w = httptest.NewRecorder()
	c, _ = gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest("GET", "/sites", nil)
	handler.ListSites(c)
	assert.Equal(t, http.StatusOK, w.Code)

	w = httptest.NewRecorder()
	c, _ = gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest("GET", "/sites?offset=0&limit=10", nil)
	handler.ListSites(c)
	assert.Equal(t, http.StatusOK, w.Code)

	w = httptest.NewRecorder()
	c, _ = gin.CreateTestContext(w)
	c.Params = []gin.Param{{Key: "id", Value: "s1"}}
	c.Request, _ = http.NewRequest("GET", "/sites/s1", nil)
	handler.GetSite(c)
	assert.Equal(t, http.StatusOK, w.Code)

	w = httptest.NewRecorder()
	c, _ = gin.CreateTestContext(w)
	c.Params = []gin.Param{{Key: "id", Value: "org1"}}
	c.Request, _ = http.NewRequest("POST", "/sites", bytes.NewBuffer([]byte(`{"name":"test"}`)))
	handler.CreateSite(c)
	assert.Equal(t, http.StatusCreated, w.Code)

	w = httptest.NewRecorder()
	c, _ = gin.CreateTestContext(w)
	c.Params = []gin.Param{{Key: "id", Value: "s1"}}
	c.Request, _ = http.NewRequest("PUT", "/sites/s1", bytes.NewBuffer([]byte(`{"name":"test"}`)))
	handler.UpdateSite(c)
	assert.Equal(t, http.StatusOK, w.Code)

	w = httptest.NewRecorder()
	c, _ = gin.CreateTestContext(w)
	c.Params = []gin.Param{{Key: "id", Value: "s1"}}
	c.Request, _ = http.NewRequest("DELETE", "/sites/s1", nil)
	handler.DeleteSite(c)
	assert.Equal(t, http.StatusOK, w.Code)

	// Locations
	w = httptest.NewRecorder()
	c, _ = gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest("GET", "/locations", nil)
	handler.ListLocations(c)
	assert.Equal(t, http.StatusOK, w.Code)

	w = httptest.NewRecorder()
	c, _ = gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest("GET", "/locations?offset=0&limit=10", nil)
	handler.ListLocations(c)
	assert.Equal(t, http.StatusOK, w.Code)

	w = httptest.NewRecorder()
	c, _ = gin.CreateTestContext(w)
	c.Params = []gin.Param{{Key: "id", Value: "l1"}}
	c.Request, _ = http.NewRequest("GET", "/locations/l1", nil)
	handler.GetLocation(c)
	assert.Equal(t, http.StatusOK, w.Code)

	w = httptest.NewRecorder()
	c, _ = gin.CreateTestContext(w)
	c.Params = []gin.Param{{Key: "siteId", Value: "s1"}}
	c.Request, _ = http.NewRequest("POST", "/locations", bytes.NewBuffer([]byte(`{"name":"test","metadata":{}}`)))
	handler.CreateLocation(c)
	assert.Equal(t, http.StatusCreated, w.Code, w.Body.String())

	w = httptest.NewRecorder()
	c, _ = gin.CreateTestContext(w)
	c.Params = []gin.Param{{Key: "id", Value: "l1"}}
	c.Request, _ = http.NewRequest("PUT", "/locations/l1", bytes.NewBuffer([]byte(`{"name":"test"}`)))
	handler.UpdateLocation(c)
	assert.Equal(t, http.StatusOK, w.Code)

	w = httptest.NewRecorder()
	c, _ = gin.CreateTestContext(w)
	c.Params = []gin.Param{{Key: "id", Value: "l1"}}
	c.Request, _ = http.NewRequest("DELETE", "/locations/l1", nil)
	handler.DeleteLocation(c)
	assert.Equal(t, http.StatusOK, w.Code)

	// Assets
	w = httptest.NewRecorder()
	c, _ = gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest("GET", "/assets", nil)
	handler.ListAssets(c)
	assert.Equal(t, http.StatusOK, w.Code)

	w = httptest.NewRecorder()
	c, _ = gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest("GET", "/assets?offset=0&limit=10", nil)
	handler.ListAssets(c)
	assert.Equal(t, http.StatusOK, w.Code)

	w = httptest.NewRecorder()
	c, _ = gin.CreateTestContext(w)
	c.Params = []gin.Param{{Key: "id", Value: "a1"}}
	c.Request, _ = http.NewRequest("GET", "/assets/a1", nil)
	handler.GetAsset(c)
	assert.Equal(t, http.StatusOK, w.Code)

	w = httptest.NewRecorder()
	c, _ = gin.CreateTestContext(w)
	c.Params = []gin.Param{{Key: "locationId", Value: "l1"}}
	c.Request, _ = http.NewRequest("POST", "/assets", bytes.NewBuffer([]byte(`{"name":"test"}`)))
	handler.CreateAsset(c)
	assert.Equal(t, http.StatusCreated, w.Code)

	w = httptest.NewRecorder()
	c, _ = gin.CreateTestContext(w)
	c.Params = []gin.Param{{Key: "id", Value: "a1"}}
	c.Request, _ = http.NewRequest("PUT", "/assets/a1", bytes.NewBuffer([]byte(`{"name":"test"}`)))
	handler.UpdateAsset(c)
	assert.Equal(t, http.StatusOK, w.Code)

	w = httptest.NewRecorder()
	c, _ = gin.CreateTestContext(w)
	c.Params = []gin.Param{{Key: "id", Value: "a1"}}
	c.Request, _ = http.NewRequest("DELETE", "/assets/a1", nil)
	handler.DeleteAsset(c)
	assert.Equal(t, http.StatusOK, w.Code)

	// Task Templates
	w = httptest.NewRecorder()
	c, _ = gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest("GET", "/task-templates", nil)
	handler.ListTaskTemplates(c)
	assert.Equal(t, http.StatusOK, w.Code)

	w = httptest.NewRecorder()
	c, _ = gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest("GET", "/task-templates?offset=0&limit=10", nil)
	handler.ListTaskTemplates(c)
	assert.Equal(t, http.StatusOK, w.Code)

	w = httptest.NewRecorder()
	c, _ = gin.CreateTestContext(w)
	c.Params = []gin.Param{{Key: "id", Value: "t1"}}
	c.Request, _ = http.NewRequest("GET", "/task-templates/t1", nil)
	handler.GetTaskTemplate(c)
	assert.Equal(t, http.StatusOK, w.Code)

	w = httptest.NewRecorder()
	c, _ = gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest("POST", "/task-templates", bytes.NewBuffer([]byte(`{"name":"test","checklist_template":[],"metadata":{}}`)))
	handler.CreateTaskTemplate(c)
	assert.Equal(t, http.StatusCreated, w.Code)

	w = httptest.NewRecorder()
	c, _ = gin.CreateTestContext(w)
	c.Params = []gin.Param{{Key: "id", Value: "t1"}}
	c.Request, _ = http.NewRequest("PUT", "/task-templates/t1", bytes.NewBuffer(reqJSON))
	handler.UpdateTaskTemplate(c)
	assert.Equal(t, http.StatusOK, w.Code)

	w = httptest.NewRecorder()
	c, _ = gin.CreateTestContext(w)
	c.Params = []gin.Param{{Key: "id", Value: "t1"}}
	c.Request, _ = http.NewRequest("DELETE", "/task-templates/t1", nil)
	handler.DeleteTaskTemplate(c)
	assert.Equal(t, http.StatusOK, w.Code)

	// Scheduler Status / Sweep
	w = httptest.NewRecorder()
	c, _ = gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest("POST", "/scheduler/sweep", nil)
	handler.TriggerSchedulerSweep(c)
	assert.Equal(t, http.StatusOK, w.Code)

	w = httptest.NewRecorder()
	c, _ = gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest("GET", "/scheduler/status", nil)
	handler.GetSchedulerStatus(c)
	assert.Equal(t, http.StatusOK, w.Code)

	// Task Executions
	w = httptest.NewRecorder()
	c, _ = gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest("GET", "/task-executions", nil)
	handler.ListTaskExecutions(c)
	assert.Equal(t, http.StatusOK, w.Code)

	w = httptest.NewRecorder()
	c, _ = gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest("GET", "/task-executions?offset=0&limit=10", nil)
	handler.ListTaskExecutions(c)
	assert.Equal(t, http.StatusOK, w.Code)

	w = httptest.NewRecorder()
	c, _ = gin.CreateTestContext(w)
	c.Params = []gin.Param{{Key: "id", Value: "e1"}}
	c.Request, _ = http.NewRequest("GET", "/task-executions/e1", nil)
	handler.GetTaskExecution(c)
	assert.Equal(t, http.StatusOK, w.Code)

	w = httptest.NewRecorder()
	c, _ = gin.CreateTestContext(w)
	c.Params = []gin.Param{{Key: "id", Value: "e1"}}
	c.Request, _ = http.NewRequest("DELETE", "/task-executions/e1", nil)
	handler.DeleteTaskExecution(c)
	assert.Equal(t, http.StatusOK, w.Code)

	// Shift Sessions
	w = httptest.NewRecorder()
	c, _ = gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest("GET", "/shift-sessions", nil)
	handler.ListShiftSessions(c)
	assert.Equal(t, http.StatusOK, w.Code)

	w = httptest.NewRecorder()
	c, _ = gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest("GET", "/shift-sessions?offset=0&limit=10", nil)
	handler.ListShiftSessions(c)
	assert.Equal(t, http.StatusOK, w.Code)

	w = httptest.NewRecorder()
	c, _ = gin.CreateTestContext(w)
	c.Params = []gin.Param{{Key: "id", Value: "sess1"}}
	c.Request, _ = http.NewRequest("GET", "/shift-sessions/sess1", nil)
	handler.GetShiftSession(c)
	assert.Equal(t, http.StatusOK, w.Code)

	w = httptest.NewRecorder()
	c, _ = gin.CreateTestContext(w)
	c.Params = []gin.Param{{Key: "id", Value: "sess1"}}
	c.Request, _ = http.NewRequest("DELETE", "/shift-sessions/sess1", nil)
	handler.DeleteShiftSession(c)
	assert.Equal(t, http.StatusOK, w.Code)

	// SOPs
	w = httptest.NewRecorder()
	c, _ = gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest("GET", "/sops", nil)
	handler.ListSOPs(c)
	assert.Equal(t, http.StatusOK, w.Code)

	w = httptest.NewRecorder()
	c, _ = gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest("GET", "/sops?offset=0&limit=10", nil)
	handler.ListSOPs(c)
	assert.Equal(t, http.StatusOK, w.Code)

	w = httptest.NewRecorder()
	c, _ = gin.CreateTestContext(w)
	c.Params = []gin.Param{{Key: "id", Value: "sop1"}}
	c.Request, _ = http.NewRequest("GET", "/sops/sop1", nil)
	handler.GetSOP(c)
	assert.Equal(t, http.StatusOK, w.Code)

	w = httptest.NewRecorder()
	c, _ = gin.CreateTestContext(w)
	c.Params = []gin.Param{{Key: "id", Value: "sop1"}}
	c.Request, _ = http.NewRequest("DELETE", "/sops/sop1", nil)
	handler.DeleteSOP(c)
	assert.Equal(t, http.StatusOK, w.Code)

	// Processes
	w = httptest.NewRecorder()
	c, _ = gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest("GET", "/sop-processes", nil)
	handler.ListProcesses(c)
	assert.Equal(t, http.StatusOK, w.Code)

	w = httptest.NewRecorder()
	c, _ = gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest("GET", "/sop-processes?offset=0&limit=10", nil)
	handler.ListProcesses(c)
	assert.Equal(t, http.StatusOK, w.Code)

	w = httptest.NewRecorder()
	c, _ = gin.CreateTestContext(w)
	c.Params = []gin.Param{{Key: "id", Value: "p1"}}
	c.Request, _ = http.NewRequest("GET", "/sop-processes/p1", nil)
	handler.GetProcess(c)
	assert.Equal(t, http.StatusOK, w.Code)

	w = httptest.NewRecorder()
	c, _ = gin.CreateTestContext(w)
	c.Params = []gin.Param{{Key: "id", Value: "p1"}}
	c.Request, _ = http.NewRequest("DELETE", "/sop-processes/p1", nil)
	handler.DeleteProcess(c)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestAdminHandler_ErrorBranches(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := NewAdminHandler(&mockAdminServiceForAPI{}, &mockSchedulerForAPI{}, &mockTaskForAPI{}, &mockShiftForAPI{}, &mockRAGForAPI{})

	endpoints := []struct {
		method  string
		handler gin.HandlerFunc
	}{
		{"POST", handler.CreateUser},
		{"PUT", handler.UpdateUser},
		{"POST", handler.CreateRole},
		{"PUT", handler.UpdateRole},
		{"POST", handler.AssignRole},
		{"POST", handler.CreateOrganization},
		{"PUT", handler.UpdateOrganization},
		{"POST", handler.CreateSite},
		{"PUT", handler.UpdateSite},
		{"POST", handler.CreateLocation},
		{"PUT", handler.UpdateLocation},
		{"POST", handler.CreateAsset},
		{"PUT", handler.UpdateAsset},
		{"POST", handler.CreateTaskTemplate},
		{"PUT", handler.UpdateTaskTemplate},
	}

	for _, ep := range endpoints {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request, _ = http.NewRequest(ep.method, "/", bytes.NewBuffer([]byte(`{invalid-json`)))
		ep.handler(c)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	}

	// AssignUserToOrganization missing params & service error
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest("POST", "/organizations//users/", nil)
	handler.AssignUserToOrganization(c)
	assert.Equal(t, http.StatusBadRequest, w.Code)

	w = httptest.NewRecorder()
	c, _ = gin.CreateTestContext(w)
	c.Params = []gin.Param{{Key: "id", Value: "err"}, {Key: "userId", Value: "u1"}}
	c.Request, _ = http.NewRequest("POST", "/organizations/err/users/u1", nil)
	handler.AssignUserToOrganization(c)
	assert.Equal(t, http.StatusInternalServerError, w.Code)

	// GET and DELETE with id = "err"
	errGetDelete := []gin.HandlerFunc{
		handler.GetUser, handler.DeleteUser,
		handler.GetRole, handler.DeleteRole,
		handler.GetOrganization, handler.DeleteOrganization,
		handler.GetSite, handler.DeleteSite,
		handler.GetLocation, handler.DeleteLocation,
		handler.GetAsset, handler.DeleteAsset,
		handler.GetTaskTemplate, handler.DeleteTaskTemplate,
		handler.GetTaskExecution, handler.DeleteTaskExecution,
		handler.GetShiftSession, handler.DeleteShiftSession,
		handler.GetSOP, handler.DeleteSOP,
		handler.GetProcess, handler.DeleteProcess,
	}

	for _, fn := range errGetDelete {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Params = []gin.Param{{Key: "id", Value: "err"}}
		c.Request, _ = http.NewRequest("GET", "/err", nil)
		fn(c)
		assert.Equal(t, http.StatusInternalServerError, w.Code)
	}

	// POST/PUT with {"name":"err"}
	errPostPut := []struct {
		method string
		fn     gin.HandlerFunc
	}{
		{"POST", handler.CreateUser},
		{"PUT", handler.UpdateUser},
		{"POST", handler.CreateRole},
		{"PUT", handler.UpdateRole},
		{"POST", handler.CreateOrganization},
		{"PUT", handler.UpdateOrganization},
		{"POST", handler.CreateSite},
		{"PUT", handler.UpdateSite},
		{"POST", handler.CreateLocation},
		{"PUT", handler.UpdateLocation},
		{"POST", handler.CreateAsset},
		{"PUT", handler.UpdateAsset},
		{"POST", handler.CreateTaskTemplate},
		{"PUT", handler.UpdateTaskTemplate},
	}
	for _, ep := range errPostPut {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Params = []gin.Param{{Key: "id", Value: "org1"}, {Key: "siteId", Value: "site1"}, {Key: "locationId", Value: "loc1"}}
		c.Request, _ = http.NewRequest(ep.method, "/", bytes.NewBuffer([]byte(`{"name":"err"}`)))
		ep.fn(c)
		assert.Equal(t, http.StatusInternalServerError, w.Code)
	}

	// Missing parent path parameters for CreateSite, CreateLocation, CreateAsset
	missingParentMethods := []gin.HandlerFunc{
		handler.CreateSite, handler.CreateLocation, handler.CreateAsset,
	}
	for _, fn := range missingParentMethods {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request, _ = http.NewRequest("POST", "/", bytes.NewBuffer([]byte(`{"name":"test"}`)))
		fn(c)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	}

	// Range error checks
	r := gin.New()
	r.GET("/users", handler.ListUsers)
	r.GET("/roles", handler.ListRoles)
	r.GET("/orgs", handler.ListOrganizations)
	r.GET("/sites", handler.ListSites)
	r.GET("/locs", handler.ListLocations)
	r.GET("/assets", handler.ListAssets)
	r.GET("/tpls", handler.ListTaskTemplates)
	r.GET("/execs", handler.ListTaskExecutions)
	r.GET("/sessions", handler.ListShiftSessions)
	r.GET("/sops", handler.ListSOPs)
	r.GET("/procs", handler.ListProcesses)

	paths := []string{
		"/users", "/roles", "/orgs", "/sites", "/locs", "/assets",
		"/tpls", "/execs", "/sessions", "/sops", "/procs",
	}
	for _, p := range paths {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", p+"?offset=999&limit=10", nil)
		r.ServeHTTP(w, req)
		assert.Equal(t, http.StatusInternalServerError, w.Code, "failed on path "+p)
	}

	// TriggerSchedulerSweep error check
	w = httptest.NewRecorder()
	c, _ = gin.CreateTestContext(w)
	reqErr, _ := http.NewRequestWithContext(context.WithValue(context.Background(), "err", true), "POST", "/scheduler/sweep", nil)
	c.Request = reqErr
	handler.TriggerSchedulerSweep(c)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}
