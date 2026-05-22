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
	"github.com/rmcguinness/gemini_task_engine/pkg/api"
	"github.com/rmcguinness/gemini_task_engine/pkg/model"
	"github.com/stretchr/testify/assert"
)

// Mock implementations for service layers
type mockAdminService struct {
	ListUsersFunc                func(ctx context.Context) ([]*model.User, error)
	CreateRoleFunc               func(ctx context.Context, role *model.Role) error
	AssignRoleFunc               func(ctx context.Context, userID, roleID string) error
	RegisterUserFunc             func(ctx context.Context, user *model.User) error
	RegisterLocationFunc         func(ctx context.Context, loc *model.Location) error
	RegisterAssetFunc            func(ctx context.Context, asset *model.Asset) error
	CreateTaskTemplateFunc       func(ctx context.Context, task *model.Task) error
	ListRolesFunc                func(ctx context.Context) ([]*model.Role, error)
	RegisterSiteFunc             func(ctx context.Context, site *model.Site) error
	RegisterOrganizationFunc     func(ctx context.Context, org *model.Organization) error
	AssignUserToOrganizationFunc func(ctx context.Context, orgID, userID string) error
	ListOrganizationsFunc        func(ctx context.Context) ([]*model.Organization, error)
}

func (m *mockAdminService) ListUsers(ctx context.Context) ([]*model.User, error) {
	if m.ListUsersFunc != nil {
		return m.ListUsersFunc(ctx)
	}
	return nil, nil
}

func (m *mockAdminService) CreateRole(ctx context.Context, role *model.Role) error {
	if m.CreateRoleFunc != nil {
		return m.CreateRoleFunc(ctx, role)
	}
	return nil
}

func (m *mockAdminService) AssignRole(ctx context.Context, userID, roleID string) error {
	if m.AssignRoleFunc != nil {
		return m.AssignRoleFunc(ctx, userID, roleID)
	}
	return nil
}

func (m *mockAdminService) RegisterUser(ctx context.Context, user *model.User) error {
	if m.RegisterUserFunc != nil {
		return m.RegisterUserFunc(ctx, user)
	}
	return nil
}

func (m *mockAdminService) RegisterLocation(ctx context.Context, loc *model.Location) error {
	if m.RegisterLocationFunc != nil {
		return m.RegisterLocationFunc(ctx, loc)
	}
	return nil
}

func (m *mockAdminService) RegisterAsset(ctx context.Context, asset *model.Asset) error {
	if m.RegisterAssetFunc != nil {
		return m.RegisterAssetFunc(ctx, asset)
	}
	return nil
}

func (m *mockAdminService) CreateTaskTemplate(ctx context.Context, task *model.Task) error {
	if m.CreateTaskTemplateFunc != nil {
		return m.CreateTaskTemplateFunc(ctx, task)
	}
	return nil
}

func (m *mockAdminService) ListRoles(ctx context.Context) ([]*model.Role, error) {
	if m.ListRolesFunc != nil {
		return m.ListRolesFunc(ctx)
	}
	return nil, nil
}

func (m *mockAdminService) RegisterSite(ctx context.Context, site *model.Site) error {
	if m.RegisterSiteFunc != nil {
		return m.RegisterSiteFunc(ctx, site)
	}
	return nil
}

func (m *mockAdminService) RegisterOrganization(ctx context.Context, org *model.Organization) error {
	if m.RegisterOrganizationFunc != nil {
		return m.RegisterOrganizationFunc(ctx, org)
	}
	return nil
}

func (m *mockAdminService) AssignUserToOrganization(ctx context.Context, orgID, userID string) error {
	if m.AssignUserToOrganizationFunc != nil {
		return m.AssignUserToOrganizationFunc(ctx, orgID, userID)
	}
	return nil
}

func (m *mockAdminService) ListOrganizations(ctx context.Context) ([]*model.Organization, error) {
	if m.ListOrganizationsFunc != nil {
		return m.ListOrganizationsFunc(ctx)
	}
	return nil, nil
}

type mockTaskService struct {
	GetQueueFunc                func(ctx context.Context, siteID string) ([]*model.TaskExecution, error)
	GetOrgTasksFunc             func(ctx context.Context, orgID string) ([]*model.TaskExecution, error)
	GetUserSiteTasksFunc        func(ctx context.Context, siteID, userID string) ([]*model.TaskExecution, error)
	UpdateStatusFunc            func(ctx context.Context, executionID, status, userID string) error
	OverrideAssetConstraintFunc func(ctx context.Context, executionID, assetID, justification, userID string) error
	ProposeTradeFunc            func(ctx context.Context, executionID, proposedAssigneeID, initiatorID string) error
	ApproveTradeFunc            func(ctx context.Context, tradeID, supervisorID string) error
}

func (m *mockTaskService) GetQueue(ctx context.Context, siteID string) ([]*model.TaskExecution, error) {
	if m.GetQueueFunc != nil {
		return m.GetQueueFunc(ctx, siteID)
	}
	return nil, nil
}

func (m *mockTaskService) GetOrgTasks(ctx context.Context, orgID string) ([]*model.TaskExecution, error) {
	if m.GetOrgTasksFunc != nil {
		return m.GetOrgTasksFunc(ctx, orgID)
	}
	return nil, nil
}

func (m *mockTaskService) GetUserSiteTasks(ctx context.Context, siteID, userID string) ([]*model.TaskExecution, error) {
	if m.GetUserSiteTasksFunc != nil {
		return m.GetUserSiteTasksFunc(ctx, siteID, userID)
	}
	return nil, nil
}

func (m *mockTaskService) UpdateStatus(ctx context.Context, executionID, status, userID string) error {
	if m.UpdateStatusFunc != nil {
		return m.UpdateStatusFunc(ctx, executionID, status, userID)
	}
	return nil
}

func (m *mockTaskService) OverrideAssetConstraint(ctx context.Context, executionID, assetID, justification, userID string) error {
	if m.OverrideAssetConstraintFunc != nil {
		return m.OverrideAssetConstraintFunc(ctx, executionID, assetID, justification, userID)
	}
	return nil
}

func (m *mockTaskService) ProposeTrade(ctx context.Context, executionID, proposedAssigneeID, initiatorID string) error {
	if m.ProposeTradeFunc != nil {
		return m.ProposeTradeFunc(ctx, executionID, proposedAssigneeID, initiatorID)
	}
	return nil
}

func (m *mockTaskService) ApproveTrade(ctx context.Context, tradeID, supervisorID string) error {
	if m.ApproveTradeFunc != nil {
		return m.ApproveTradeFunc(ctx, tradeID, supervisorID)
	}
	return nil
}

type mockShiftService struct {
	InitializeShiftFunc func(ctx context.Context, userID, shiftInstanceID string) (*model.ShiftAgentSession, error)
}

func (m *mockShiftService) InitializeShift(ctx context.Context, userID, shiftInstanceID string) (*model.ShiftAgentSession, error) {
	if m.InitializeShiftFunc != nil {
		return m.InitializeShiftFunc(ctx, userID, shiftInstanceID)
	}
	return nil, nil
}

type mockRAGService struct {
	RegisterSOPFunc     func(ctx context.Context, sop *model.SOP) error
	SaveChunksFunc      func(ctx context.Context, chunks []*model.SOPChunk) error
	QuerySimilarityFunc func(ctx context.Context, query model.Float32Vector, limit int) ([]*model.SOPChunk, error)
}

func (m *mockRAGService) RegisterSOP(ctx context.Context, sop *model.SOP) error {
	if m.RegisterSOPFunc != nil {
		return m.RegisterSOPFunc(ctx, sop)
	}
	return nil
}

func (m *mockRAGService) SaveChunks(ctx context.Context, chunks []*model.SOPChunk) error {
	if m.SaveChunksFunc != nil {
		return m.SaveChunksFunc(ctx, chunks)
	}
	return nil
}

func (m *mockRAGService) QuerySimilarity(ctx context.Context, query model.Float32Vector, limit int) ([]*model.SOPChunk, error) {
	if m.QuerySimilarityFunc != nil {
		return m.QuerySimilarityFunc(ctx, query, limit)
	}
	return nil, nil
}

func init() {
	gin.SetMode(gin.TestMode)
}

func TestReadiness(t *testing.T) {
	adminSvc := &mockAdminService{}
	taskSvc := &mockTaskService{}
	shiftSvc := &mockShiftService{}
	ragSvc := &mockRAGService{}

	srv, err := api.NewServer(api.Config{Address: "127.0.0.1", Port: "8080"}, adminSvc, taskSvc, shiftSvc, ragSvc)
	assert.NoError(t, err)

	req := httptest.NewRequest("GET", "/health/readiness", nil)
	w := httptest.NewRecorder()
	srv.Engine().ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var body map[string]string
	err = json.Unmarshal(w.Body.Bytes(), &body)
	assert.NoError(t, err)
	assert.Equal(t, "healthy", body["status"])
	assert.Equal(t, "connected", body["database"])
}

func TestCORSHeaders(t *testing.T) {
	adminSvc := &mockAdminService{}
	taskSvc := &mockTaskService{}
	shiftSvc := &mockShiftService{}
	ragSvc := &mockRAGService{}

	srv, err := api.NewServer(api.Config{Address: "127.0.0.1", Port: "8080"}, adminSvc, taskSvc, shiftSvc, ragSvc)
	assert.NoError(t, err)

	req := httptest.NewRequest("OPTIONS", "/health/readiness", nil)
	w := httptest.NewRecorder()
	srv.Engine().ServeHTTP(w, req)

	assert.Equal(t, http.StatusNoContent, w.Code)
	assert.Equal(t, "*", w.Header().Get("Access-Control-Allow-Origin"))
	assert.Equal(t, "POST, OPTIONS, GET, PUT, PATCH, DELETE", w.Header().Get("Access-Control-Allow-Methods"))
}

func TestUserContextMiddleware(t *testing.T) {
	adminSvc := &mockAdminService{}
	taskSvc := &mockTaskService{}
	shiftSvc := &mockShiftService{}
	ragSvc := &mockRAGService{}

	srv, err := api.NewServer(api.Config{Address: "127.0.0.1", Port: "8080"}, adminSvc, taskSvc, shiftSvc, ragSvc)
	assert.NoError(t, err)

	// Target endpoint calls task service, let's capture the user ID passed from operational.go
	var capturedUserID string
	taskSvc.UpdateStatusFunc = func(ctx context.Context, executionID, status, userID string) error {
		capturedUserID = userID
		return nil
	}

	// Scenario A: X-User-ID Header set
	reqBody := `{"status":"COMPLETED"}`
	req := httptest.NewRequest("PATCH", "/api/v1/organizations/org-123/sites/site-123/tasks/task-123/status", bytes.NewBufferString(reqBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-User-ID", "custom-user-id-1")
	w := httptest.NewRecorder()
	srv.Engine().ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "custom-user-id-1", capturedUserID)

	// Scenario B: Bearer Authorization Header set
	req = httptest.NewRequest("PATCH", "/api/v1/organizations/org-123/sites/site-123/tasks/task-123/status", bytes.NewBufferString(reqBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer jwt-user-id-2")
	w = httptest.NewRecorder()
	srv.Engine().ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "jwt-user-id-2", capturedUserID)

	// Scenario C: No Header fallback
	req = httptest.NewRequest("PATCH", "/api/v1/organizations/org-123/sites/site-123/tasks/task-123/status", bytes.NewBufferString(reqBody))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	srv.Engine().ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "00000000-0000-0000-0000-000000000000", capturedUserID)
}

func TestAdminEndpoints(t *testing.T) {
	adminSvc := &mockAdminService{}
	taskSvc := &mockTaskService{}
	shiftSvc := &mockShiftService{}
	ragSvc := &mockRAGService{}

	srv, err := api.NewServer(api.Config{Address: "127.0.0.1", Port: "8080"}, adminSvc, taskSvc, shiftSvc, ragSvc)
	assert.NoError(t, err)

	t.Run("ListUsers", func(t *testing.T) {
		adminSvc.ListUsersFunc = func(ctx context.Context) ([]*model.User, error) {
			return []*model.User{
				{ID: "user-1", Name: "Hanna", Email: "hanna@walmart.com"},
			}, nil
		}

		req := httptest.NewRequest("GET", "/api/v1/admin/users", nil)
		w := httptest.NewRecorder()
		srv.Engine().ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		var users []*model.User
		err := json.Unmarshal(w.Body.Bytes(), &users)
		assert.NoError(t, err)
		assert.Len(t, users, 1)
		assert.Equal(t, "Hanna", users[0].Name)
	})

	t.Run("CreateRole", func(t *testing.T) {
		var capturedRole *model.Role
		adminSvc.CreateRoleFunc = func(ctx context.Context, role *model.Role) error {
			capturedRole = role
			role.ID = "role-created-id"
			return nil
		}

		reqBody := `{"name":"Supervisor","description":"Shift Lead"}`
		req := httptest.NewRequest("POST", "/api/v1/admin/roles", bytes.NewBufferString(reqBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		srv.Engine().ServeHTTP(w, req)

		assert.Equal(t, http.StatusCreated, w.Code)
		assert.NotNil(t, capturedRole)
		assert.Equal(t, "Supervisor", capturedRole.Name)
		assert.Equal(t, "Shift Lead", capturedRole.Description)
	})

	t.Run("AssignRole", func(t *testing.T) {
		var capturedUserID, capturedRoleID string
		adminSvc.AssignRoleFunc = func(ctx context.Context, userID, roleID string) error {
			capturedUserID = userID
			capturedRoleID = roleID
			return nil
		}

		reqBody := `{"role_id":"role-123"}`
		req := httptest.NewRequest("PUT", "/api/v1/admin/users/user-456/roles", bytes.NewBufferString(reqBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		srv.Engine().ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, "user-456", capturedUserID)
		assert.Equal(t, "role-123", capturedRoleID)
	})
}

func TestOperationalEndpoints(t *testing.T) {
	adminSvc := &mockAdminService{}
	taskSvc := &mockTaskService{}
	shiftSvc := &mockShiftService{}
	ragSvc := &mockRAGService{}

	srv, err := api.NewServer(api.Config{Address: "127.0.0.1", Port: "8080"}, adminSvc, taskSvc, shiftSvc, ragSvc)
	assert.NoError(t, err)

	t.Run("GetSiteTasks - Success", func(t *testing.T) {
		taskSvc.GetQueueFunc = func(ctx context.Context, siteID string) ([]*model.TaskExecution, error) {
			assert.Equal(t, "site-abc", siteID)
			return []*model.TaskExecution{
				{ID: "exec-1", TaskTemplateID: "temp-1", Priority: 1, Status: "PENDING"},
			}, nil
		}

		req := httptest.NewRequest("GET", "/api/v1/organizations/org-abc/sites/site-abc/tasks", nil)
		w := httptest.NewRecorder()
		srv.Engine().ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		var execs []*model.TaskExecution
		err := json.Unmarshal(w.Body.Bytes(), &execs)
		assert.NoError(t, err)
		assert.Len(t, execs, 1)
		assert.Equal(t, "exec-1", execs[0].ID)
	})

	t.Run("GetOrgTasks - Success", func(t *testing.T) {
		taskSvc.GetOrgTasksFunc = func(ctx context.Context, orgID string) ([]*model.TaskExecution, error) {
			assert.Equal(t, "org-abc", orgID)
			return []*model.TaskExecution{
				{ID: "exec-org", TaskTemplateID: "temp-org", Priority: 1, Status: "PENDING"},
			}, nil
		}

		req := httptest.NewRequest("GET", "/api/v1/organizations/org-abc/tasks", nil)
		w := httptest.NewRecorder()
		srv.Engine().ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		var execs []*model.TaskExecution
		err := json.Unmarshal(w.Body.Bytes(), &execs)
		assert.NoError(t, err)
		assert.Len(t, execs, 1)
		assert.Equal(t, "exec-org", execs[0].ID)
	})

	t.Run("GetUserSiteTasks - Success", func(t *testing.T) {
		taskSvc.GetUserSiteTasksFunc = func(ctx context.Context, siteID, userID string) ([]*model.TaskExecution, error) {
			assert.Equal(t, "site-abc", siteID)
			assert.Equal(t, "user-123", userID)
			return []*model.TaskExecution{
				{ID: "exec-user", TaskTemplateID: "temp-user", Priority: 1, Status: "PENDING"},
			}, nil
		}

		req := httptest.NewRequest("GET", "/api/v1/organizations/org-abc/sites/site-abc/users/user-123/tasks", nil)
		w := httptest.NewRecorder()
		srv.Engine().ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		var execs []*model.TaskExecution
		err := json.Unmarshal(w.Body.Bytes(), &execs)
		assert.NoError(t, err)
		assert.Len(t, execs, 1)
		assert.Equal(t, "exec-user", execs[0].ID)
	})

	t.Run("OverrideAssetConstraint", func(t *testing.T) {
		var capturedExecID, capturedAssetID, capturedJustification, capturedUserID string
		taskSvc.OverrideAssetConstraintFunc = func(ctx context.Context, executionID, assetID, justification, userID string) error {
			capturedExecID = executionID
			capturedAssetID = assetID
			capturedJustification = justification
			capturedUserID = userID
			return nil
		}

		reqBody := `{"asset_id":"asset-789","justification":"Safety clearance verified"}`
		req := httptest.NewRequest("POST", "/api/v1/organizations/org-abc/sites/site-abc/tasks/exec-111/override", bytes.NewBufferString(reqBody))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-User-ID", "user-admin")
		w := httptest.NewRecorder()
		srv.Engine().ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, "exec-111", capturedExecID)
		assert.Equal(t, "asset-789", capturedAssetID)
		assert.Equal(t, "Safety clearance verified", capturedJustification)
		assert.Equal(t, "user-admin", capturedUserID)
	})

	t.Run("ProposeTrade", func(t *testing.T) {
		var capturedExecID, capturedProposedAssignee, capturedUserID string
		taskSvc.ProposeTradeFunc = func(ctx context.Context, executionID, proposedAssigneeID, initiatorID string) error {
			capturedExecID = executionID
			capturedProposedAssignee = proposedAssigneeID
			capturedUserID = initiatorID
			return nil
		}

		reqBody := `{"task_execution_id":"exec-222","proposed_assignee_id":"user-333"}`
		req := httptest.NewRequest("POST", "/api/v1/organizations/org-abc/sites/site-abc/trades", bytes.NewBufferString(reqBody))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-User-ID", "user-initiator")
		w := httptest.NewRecorder()
		srv.Engine().ServeHTTP(w, req)

		assert.Equal(t, http.StatusCreated, w.Code)
		assert.Equal(t, "exec-222", capturedExecID)
		assert.Equal(t, "user-333", capturedProposedAssignee)
		assert.Equal(t, "user-initiator", capturedUserID)
	})
}

func TestOperationalEndpointsFailure(t *testing.T) {
	adminSvc := &mockAdminService{}
	taskSvc := &mockTaskService{}
	shiftSvc := &mockShiftService{}
	ragSvc := &mockRAGService{}

	srv, err := api.NewServer(api.Config{Address: "127.0.0.1", Port: "8080"}, adminSvc, taskSvc, shiftSvc, ragSvc)
	assert.NoError(t, err)

	t.Run("UpdateTaskStatus - Error", func(t *testing.T) {
		taskSvc.UpdateStatusFunc = func(ctx context.Context, executionID, status, userID string) error {
			return errors.New("database failure")
		}

		reqBody := `{"status":"COMPLETED"}`
		req := httptest.NewRequest("PATCH", "/api/v1/organizations/org-abc/sites/site-abc/tasks/exec-fail/status", bytes.NewBufferString(reqBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		srv.Engine().ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		assert.Contains(t, w.Body.String(), "database failure")
	})
}
