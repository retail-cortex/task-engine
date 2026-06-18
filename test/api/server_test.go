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
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/rmcguinness/gemini_task_engine/pkg/api"
	"github.com/rmcguinness/gemini_task_engine/pkg/model"
	"github.com/rmcguinness/gemini_task_engine/pkg/service"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

// Helper to retrieve a real Google identity token from local gcloud session
func getGcloudToken() (string, error) {
	paths := []string{
		"/Users/rmcguinness/Applications/google-cloud-sdk/bin/gcloud",
		"gcloud",
		"/opt/homebrew/bin/gcloud",
		"/usr/local/bin/gcloud",
		"/usr/bin/gcloud",
	}
	var errs []string
	var token string

	// Construct a robust PATH for the child process to find python3, homebrew, etc.
	robustPath := "PATH=/usr/bin:/bin:/usr/sbin:/sbin:/usr/local/bin:/opt/homebrew/bin:/Users/rmcguinness/Applications/google-cloud-sdk/bin"

	for _, path := range paths {
		cmd := exec.Command(path, "auth", "print-identity-token")
		
		// Inject the robust PATH into the command's environment
		cmd.Env = append(cmd.Env, robustPath)
		// Propagate other env vars like HOME (required by gcloud to find config)
		for _, e := range os.Environ() {
			if !strings.HasPrefix(e, "PATH=") {
				cmd.Env = append(cmd.Env, e)
			}
		}

		var stdout, stderr bytes.Buffer
		cmd.Stdout = &stdout
		cmd.Stderr = &stderr
		if err := cmd.Run(); err == nil {
			token = strings.TrimSpace(stdout.String())
			break
		} else {
			errs = append(errs, fmt.Sprintf("path %s failed: %v (stderr: %s)", path, err, strings.TrimSpace(stderr.String())))
		}
	}

	if token == "" {
		return "", fmt.Errorf("failed to retrieve gcloud identity token: %s", strings.Join(errs, "\n"))
	}
	return token, nil
}

// Helper to create authenticated requests with a valid mock user UUID for protected endpoints
func newAuthRequest(method, target string, body io.Reader) *http.Request {
	req := httptest.NewRequest(method, target, body)
	req.Header.Set("X-User-ID", "11111111-2222-3333-4444-555555555555")
	return req
}

// Mock implementations for service layers
type mockAdminService struct {
	ListUsersFunc                func(ctx context.Context) ([]*model.User, error)
	CreateRoleFunc               func(ctx context.Context, role *model.Role) error
	AssignRoleFunc               func(ctx context.Context, userID, roleID string) error
	RegisterUserFunc             func(ctx context.Context, user *model.User) error
	UpdateUserFunc               func(ctx context.Context, user *model.User) error
	RegisterLocationFunc         func(ctx context.Context, loc *model.Location) error
	RegisterAssetFunc            func(ctx context.Context, asset *model.Asset) error
	CreateTaskTemplateFunc       func(ctx context.Context, task *model.Task) error
	ListRolesFunc                func(ctx context.Context) ([]*model.Role, error)
	RegisterSiteFunc             func(ctx context.Context, site *model.Site) error
	RegisterOrganizationFunc     func(ctx context.Context, org *model.Organization) error
	AssignUserToOrganizationFunc func(ctx context.Context, orgID, userID string) error
	ListOrganizationsFunc        func(ctx context.Context) ([]*model.Organization, error)
	FindUserByOAuthFunc          func(ctx context.Context, provider, oauthID string) (*model.User, error)
	GetUserByIDFunc              func(ctx context.Context, id string) (*model.User, error)
	DeleteUserFunc               func(ctx context.Context, id string) error
	ListUsersRangeFunc           func(ctx context.Context, offset, limit int) ([]*model.User, error)
}

func (m *mockAdminService) ListUsers(ctx context.Context) ([]*model.User, error) {
	if m.ListUsersFunc != nil {
		return m.ListUsersFunc(ctx)
	}
	return nil, nil
}

func (m *mockAdminService) FindUserByOAuth(ctx context.Context, provider, oauthID string) (*model.User, error) {
	if m.FindUserByOAuthFunc != nil {
		return m.FindUserByOAuthFunc(ctx, provider, oauthID)
	}
	return nil, gorm.ErrRecordNotFound
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
	if user.ID == "" {
		user.ID = "11111111-2222-3333-4444-555555555555"
	}
	return nil
}

func (m *mockAdminService) UpdateUser(ctx context.Context, user *model.User) error {
	if m.UpdateUserFunc != nil {
		return m.UpdateUserFunc(ctx, user)
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

// Stubs for new AdminService methods
func (m *mockAdminService) GetUserByID(ctx context.Context, id string) (*model.User, error) {
	if m.GetUserByIDFunc != nil {
		return m.GetUserByIDFunc(ctx, id)
	}
	return nil, nil
}
func (m *mockAdminService) DeleteUser(ctx context.Context, id string) error {
	if m.DeleteUserFunc != nil {
		return m.DeleteUserFunc(ctx, id)
	}
	return nil
}
func (m *mockAdminService) ListUsersRange(ctx context.Context, offset, limit int) ([]*model.User, error) {
	if m.ListUsersRangeFunc != nil {
		return m.ListUsersRangeFunc(ctx, offset, limit)
	}
	return nil, nil
}
func (m *mockAdminService) GetRoleByID(ctx context.Context, id string) (*model.Role, error) { return nil, nil }
func (m *mockAdminService) UpdateRole(ctx context.Context, role *model.Role) error { return nil }
func (m *mockAdminService) DeleteRole(ctx context.Context, id string) error { return nil }
func (m *mockAdminService) ListRolesRange(ctx context.Context, offset, limit int) ([]*model.Role, error) { return nil, nil }
func (m *mockAdminService) GetOrganizationByID(ctx context.Context, id string) (*model.Organization, error) { return nil, nil }
func (m *mockAdminService) UpdateOrganization(ctx context.Context, org *model.Organization) error { return nil }
func (m *mockAdminService) DeleteOrganization(ctx context.Context, id string) error { return nil }
func (m *mockAdminService) ListOrganizationsRange(ctx context.Context, offset, limit int) ([]*model.Organization, error) { return nil, nil }
func (m *mockAdminService) GetSiteByID(ctx context.Context, id string) (*model.Site, error) { return nil, nil }
func (m *mockAdminService) UpdateSite(ctx context.Context, site *model.Site) error { return nil }
func (m *mockAdminService) DeleteSite(ctx context.Context, id string) error { return nil }
func (m *mockAdminService) ListSites(ctx context.Context) ([]*model.Site, error) { return nil, nil }
func (m *mockAdminService) ListSitesRange(ctx context.Context, offset, limit int) ([]*model.Site, error) { return nil, nil }
func (m *mockAdminService) GetLocationByID(ctx context.Context, id string) (*model.Location, error) { return nil, nil }
func (m *mockAdminService) UpdateLocation(ctx context.Context, loc *model.Location) error { return nil }
func (m *mockAdminService) DeleteLocation(ctx context.Context, id string) error { return nil }
func (m *mockAdminService) ListLocations(ctx context.Context) ([]*model.Location, error) { return nil, nil }
func (m *mockAdminService) ListLocationsRange(ctx context.Context, offset, limit int) ([]*model.Location, error) { return nil, nil }
func (m *mockAdminService) GetAssetByID(ctx context.Context, id string) (*model.Asset, error) { return nil, nil }
func (m *mockAdminService) UpdateAsset(ctx context.Context, asset *model.Asset) error { return nil }
func (m *mockAdminService) DeleteAsset(ctx context.Context, id string) error { return nil }
func (m *mockAdminService) ListAssets(ctx context.Context) ([]*model.Asset, error) { return nil, nil }
func (m *mockAdminService) ListAssetsRange(ctx context.Context, offset, limit int) ([]*model.Asset, error) { return nil, nil }
func (m *mockAdminService) GetTaskTemplateByID(ctx context.Context, id string) (*model.Task, error) { return nil, nil }
func (m *mockAdminService) UpdateTaskTemplate(ctx context.Context, task *model.Task) error { return nil }
func (m *mockAdminService) DeleteTaskTemplate(ctx context.Context, id string) error { return nil }
func (m *mockAdminService) ListTaskTemplates(ctx context.Context) ([]*model.Task, error) { return nil, nil }
func (m *mockAdminService) ListTaskTemplatesRange(ctx context.Context, offset, limit int) ([]*model.Task, error) { return nil, nil }

type mockTaskService struct {
	GetQueueFunc                func(ctx context.Context, siteID string) ([]*model.TaskExecution, error)
	GetOrgTasksFunc             func(ctx context.Context, orgID string) ([]*model.TaskExecution, error)
	GetUserSiteTasksFunc        func(ctx context.Context, siteID, userID string) ([]*model.TaskExecution, error)
	UpdateStatusFunc            func(ctx context.Context, executionID, status, checklistState, userID string) error
	OverrideAssetConstraintFunc func(ctx context.Context, executionID, assetID, justification, userID string) error
	ProposeTradeFunc            func(ctx context.Context, executionID, proposedAssigneeID, initiatorID string) error
	ApproveTradeFunc            func(ctx context.Context, tradeID, supervisorID string) error
	ListPendingTradesFunc       func(ctx context.Context, userID string) ([]*model.TaskTrade, error)
	AcceptTradeFunc            func(ctx context.Context, tradeID, targetUserID string) error
	RejectTradeFunc            func(ctx context.Context, tradeID, targetUserID string) error
	ClaimTaskFunc              func(ctx context.Context, executionID, userID string, userRoleIDs []string) error
	ListActiveSitesFunc         func(ctx context.Context) ([]*model.Site, error)
	GetSiteLocationsFunc        func(ctx context.Context, siteID string) ([]*model.Location, error)
	GetLocationByIDFunc         func(ctx context.Context, id string) (*model.Location, error)
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

func (m *mockTaskService) UpdateStatus(ctx context.Context, executionID, status, checklistState, userID string) error {
	if m.UpdateStatusFunc != nil {
		return m.UpdateStatusFunc(ctx, executionID, status, checklistState, userID)
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

func (m *mockTaskService) ListActiveSites(ctx context.Context) ([]*model.Site, error) {
	if m.ListActiveSitesFunc != nil {
		return m.ListActiveSitesFunc(ctx)
	}
	return nil, nil
}

func (m *mockTaskService) ListPendingTrades(ctx context.Context, userID string) ([]*model.TaskTrade, error) {
	if m.ListPendingTradesFunc != nil {
		return m.ListPendingTradesFunc(ctx, userID)
	}
	return nil, nil
}

func (m *mockTaskService) AcceptTrade(ctx context.Context, tradeID, targetUserID string) error {
	if m.AcceptTradeFunc != nil {
		return m.AcceptTradeFunc(ctx, tradeID, targetUserID)
	}
	return nil
}

func (m *mockTaskService) RejectTrade(ctx context.Context, tradeID, targetUserID string) error {
	if m.RejectTradeFunc != nil {
		return m.RejectTradeFunc(ctx, tradeID, targetUserID)
	}
	return nil
}

func (m *mockTaskService) ClaimTask(ctx context.Context, executionID, userID string, userRoleIDs []string) error {
	if m.ClaimTaskFunc != nil {
		return m.ClaimTaskFunc(ctx, executionID, userID, userRoleIDs)
	}
	return nil
}

func (m *mockTaskService) GetSiteLocations(ctx context.Context, siteID string) ([]*model.Location, error) {
	if m.GetSiteLocationsFunc != nil {
		return m.GetSiteLocationsFunc(ctx, siteID)
	}
	return nil, nil
}

func (m *mockTaskService) GetLocationByID(ctx context.Context, id string) (*model.Location, error) {
	if m.GetLocationByIDFunc != nil {
		return m.GetLocationByIDFunc(ctx, id)
	}
	return nil, nil
}

func (m *mockTaskService) GetTaskExecutionByID(ctx context.Context, id string) (*model.TaskExecution, error) { return nil, nil }
func (m *mockTaskService) GetSiteIDForExecution(ctx context.Context, execID string) (string, error) { return "", nil }
func (m *mockTaskService) ListTaskExecutions(ctx context.Context) ([]*model.TaskExecution, error) { return nil, nil }
func (m *mockTaskService) ListTaskExecutionsRange(ctx context.Context, offset, limit int) ([]*model.TaskExecution, error) { return nil, nil }
func (m *mockTaskService) DeleteTaskExecution(ctx context.Context, id string) error { return nil }

type mockShiftService struct {
	InitializeShiftFunc    func(ctx context.Context, userID, shiftInstanceID string) (*model.ShiftAgentSession, error)
	UpdateSessionFunc      func(ctx context.Context, session *model.ShiftAgentSession) error
	ListActiveUsersFunc    func(ctx context.Context) ([]*model.User, error)
	ListActiveOnShiftUsersFunc func(ctx context.Context, siteID string) ([]*model.User, error)
	GetUserProfileFunc     func(ctx context.Context, userID string) (*model.User, error)
}

func (m *mockShiftService) InitializeShift(ctx context.Context, userID, shiftInstanceID string) (*model.ShiftAgentSession, error) {
	if m.InitializeShiftFunc != nil {
		return m.InitializeShiftFunc(ctx, userID, shiftInstanceID)
	}
	return nil, nil
}

func (m *mockShiftService) UpdateSession(ctx context.Context, session *model.ShiftAgentSession) error {
	if m.UpdateSessionFunc != nil {
		return m.UpdateSessionFunc(ctx, session)
	}
	return nil
}

func (m *mockShiftService) ListActiveUsers(ctx context.Context) ([]*model.User, error) {
	if m.ListActiveUsersFunc != nil {
		return m.ListActiveUsersFunc(ctx)
	}
	return nil, nil
}

func (m *mockShiftService) ListActiveOnShiftUsers(ctx context.Context, siteID string) ([]*model.User, error) {
	if m.ListActiveOnShiftUsersFunc != nil {
		return m.ListActiveOnShiftUsersFunc(ctx, siteID)
	}
	return nil, nil
}

func (m *mockShiftService) GetUserProfile(ctx context.Context, userID string) (*model.User, error) {
	if m.GetUserProfileFunc != nil {
		return m.GetUserProfileFunc(ctx, userID)
	}
	return nil, nil
}

func (m *mockShiftService) GetSessionByID(ctx context.Context, id string) (*model.ShiftAgentSession, error) { return nil, nil }
func (m *mockShiftService) ListSessions(ctx context.Context) ([]*model.ShiftAgentSession, error) { return nil, nil }
func (m *mockShiftService) ListSessionsRange(ctx context.Context, offset, limit int) ([]*model.ShiftAgentSession, error) { return nil, nil }
func (m *mockShiftService) DeleteSession(ctx context.Context, id string) error { return nil }

type mockRAGService struct {
	RegisterSOPFunc     func(ctx context.Context, sop *model.SOP) error
	SaveChunksFunc      func(ctx context.Context, chunks []*model.SOPChunk) error
	QuerySimilarityFunc func(ctx context.Context, query model.Float32Vector, limit int) ([]*model.SOPChunk, error)
	IngestSOPAsyncFunc  func(ctx context.Context, title string, canonicalURL string) (*model.SOP, *model.SOPProcess, error)
	CheckSOPUpdatesFunc func(ctx context.Context, sopID string) (bool, error)
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

func (m *mockRAGService) IngestSOPAsync(ctx context.Context, title string, canonicalURL string) (*model.SOP, *model.SOPProcess, error) {
	if m.IngestSOPAsyncFunc != nil {
		return m.IngestSOPAsyncFunc(ctx, title, canonicalURL)
	}
	return nil, nil, nil
}

func (m *mockRAGService) CheckSOPUpdates(ctx context.Context, sopID string) (bool, error) {
	if m.CheckSOPUpdatesFunc != nil {
		return m.CheckSOPUpdatesFunc(ctx, sopID)
	}
	return false, nil
}

func (m *mockRAGService) GetSOPByID(ctx context.Context, id string) (*model.SOP, error) { return nil, nil }
func (m *mockRAGService) ListSOPs(ctx context.Context) ([]*model.SOP, error) { return nil, nil }
func (m *mockRAGService) ListSOPsRange(ctx context.Context, offset, limit int) ([]*model.SOP, error) { return nil, nil }
func (m *mockRAGService) DeleteSOP(ctx context.Context, id string) error { return nil }
func (m *mockRAGService) GetProcessByID(ctx context.Context, id string) (*model.SOPProcess, error) { return nil, nil }
func (m *mockRAGService) ListProcesses(ctx context.Context) ([]*model.SOPProcess, error) { return nil, nil }
func (m *mockRAGService) ListProcessesRange(ctx context.Context, offset, limit int) ([]*model.SOPProcess, error) { return nil, nil }
func (m *mockRAGService) DeleteProcess(ctx context.Context, id string) error { return nil }
func (m *mockRAGService) ProcessSOPPipeline(ctx context.Context, sopID, processID string) {}

type mockAutomationService struct {
	ProcessBatchEventFunc     func(ctx context.Context, eventInstanceID string) ([]*model.TaskExecution, error)
	TriggerStreamingEventFunc func(ctx context.Context, siteID string, organizerID string, eventType model.EventType, description string) (*model.TaskExecution, error)
	ListTemplatesFunc         func(ctx context.Context) ([]*model.Task, error)
}

func (m *mockAutomationService) ProcessBatchEvent(ctx context.Context, eventInstanceID string) ([]*model.TaskExecution, error) {
	if m.ProcessBatchEventFunc != nil {
		return m.ProcessBatchEventFunc(ctx, eventInstanceID)
	}
	return nil, nil
}

func (m *mockAutomationService) TriggerStreamingEvent(ctx context.Context, siteID string, organizerID string, eventType model.EventType, description string) (*model.TaskExecution, error) {
	if m.TriggerStreamingEventFunc != nil {
		return m.TriggerStreamingEventFunc(ctx, siteID, organizerID, eventType, description)
	}
	return nil, nil
}

func (m *mockAutomationService) ListTemplates(ctx context.Context) ([]*model.Task, error) {
	if m.ListTemplatesFunc != nil {
		return m.ListTemplatesFunc(ctx)
	}
	return nil, nil
}

type mockSchedulerService struct {
	service.SchedulerService
	GetStatusFunc              func() service.SchedulerStatus
	ForceTriggerBatchSweepFunc func(ctx context.Context) error
}

func (m *mockSchedulerService) Start(ctx context.Context) error {
	return nil
}

func (m *mockSchedulerService) Stop() error {
	return nil
}

func (m *mockSchedulerService) IsLeader() bool {
	return true
}

func (m *mockSchedulerService) NodeID() string {
	return "mock-node-ID-777"
}

func (m *mockSchedulerService) GetStatus() service.SchedulerStatus {
	if m.GetStatusFunc != nil {
		return m.GetStatusFunc()
	}
	return service.SchedulerStatus{NodeID: "mock-node-ID-777", IsLeader: true}
}

func (m *mockSchedulerService) ForceTriggerBatchSweep(ctx context.Context) error {
	if m.ForceTriggerBatchSweepFunc != nil {
		return m.ForceTriggerBatchSweepFunc(ctx)
	}
	return nil
}

func init() {
	gin.SetMode(gin.TestMode)
}

func TestReadiness(t *testing.T) {
	adminSvc := &mockAdminService{}
	taskSvc := &mockTaskService{}
	shiftSvc := &mockShiftService{}
	ragSvc := &mockRAGService{}
	automationSvc := &mockAutomationService{}
	schedulerSvc := &mockSchedulerService{}

	srv, err := api.NewServer(api.Config{Address: "127.0.0.1", Port: "8080"}, adminSvc, taskSvc, shiftSvc, ragSvc, automationSvc, schedulerSvc)
	assert.NoError(t, err)

	req := httptest.NewRequest("GET", "/health/readiness", nil)
	w := httptest.NewRecorder()
	srv.Engine().ServeHTTP(w, req)

	assert.Equal(t, http.StatusServiceUnavailable, w.Code)
	var body map[string]string
	err = json.Unmarshal(w.Body.Bytes(), &body)
	assert.NoError(t, err)
	assert.Equal(t, "healthy", body["status"])
	assert.Equal(t, "disconnected: no database initialized", body["database"])
}

func TestLiveness(t *testing.T) {
	adminSvc := &mockAdminService{}
	taskSvc := &mockTaskService{}
	shiftSvc := &mockShiftService{}
	ragSvc := &mockRAGService{}
	automationSvc := &mockAutomationService{}
	schedulerSvc := &mockSchedulerService{}

	srv, err := api.NewServer(api.Config{Address: "127.0.0.1", Port: "8080"}, adminSvc, taskSvc, shiftSvc, ragSvc, automationSvc, schedulerSvc)
	assert.NoError(t, err)

	req := httptest.NewRequest("GET", "/liveness", nil)
	w := httptest.NewRecorder()
	srv.Engine().ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var body map[string]string
	err = json.Unmarshal(w.Body.Bytes(), &body)
	assert.NoError(t, err)
	assert.Equal(t, "healthy", body["status"])
}

func TestStartup(t *testing.T) {
	adminSvc := &mockAdminService{}
	taskSvc := &mockTaskService{}
	shiftSvc := &mockShiftService{}
	ragSvc := &mockRAGService{}
	automationSvc := &mockAutomationService{}
	schedulerSvc := &mockSchedulerService{}

	srv, err := api.NewServer(api.Config{Address: "127.0.0.1", Port: "8080"}, adminSvc, taskSvc, shiftSvc, ragSvc, automationSvc, schedulerSvc)
	assert.NoError(t, err)

	req := httptest.NewRequest("GET", "/startup", nil)
	w := httptest.NewRecorder()
	srv.Engine().ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var body map[string]string
	err = json.Unmarshal(w.Body.Bytes(), &body)
	assert.NoError(t, err)
	assert.Equal(t, "healthy", body["status"])
}

func TestSwaggerEndpoints(t *testing.T) {
	adminSvc := &mockAdminService{}
	taskSvc := &mockTaskService{}
	shiftSvc := &mockShiftService{}
	ragSvc := &mockRAGService{}
	automationSvc := &mockAutomationService{}
	schedulerSvc := &mockSchedulerService{}

	srv, err := api.NewServer(api.Config{Address: "127.0.0.1", Port: "8080"}, adminSvc, taskSvc, shiftSvc, ragSvc, automationSvc, schedulerSvc)
	assert.NoError(t, err)

	// 1. Test GET /swagger
	req := httptest.NewRequest("GET", "/swagger", nil)
	w := httptest.NewRecorder()
	srv.Engine().ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "Gemini Task Engine - API Explorer")

	// 2. Test GET /swagger/openapi.json
	req = httptest.NewRequest("GET", "/swagger/openapi.json", nil)
	w = httptest.NewRecorder()
	srv.Engine().ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var body map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &body)
	assert.NoError(t, err)
	assert.Equal(t, "Gemini Task Engine API", body["info"].(map[string]interface{})["title"])
}

func TestStaticMCPEndpoint(t *testing.T) {
	adminSvc := &mockAdminService{}
	taskSvc := &mockTaskService{}
	shiftSvc := &mockShiftService{}
	ragSvc := &mockRAGService{}
	automationSvc := &mockAutomationService{}
	schedulerSvc := &mockSchedulerService{}

	srv, err := api.NewServer(api.Config{Address: "127.0.0.1", Port: "8080"}, adminSvc, taskSvc, shiftSvc, ragSvc, automationSvc, schedulerSvc)
	assert.NoError(t, err)

	// Verify route registration in Gin Engine
	found := false
	for _, route := range srv.Engine().Routes() {
		if route.Path == "/api/v1/mcp" && route.Method == "POST" {
			found = true
			break
		}
	}
	assert.True(t, found, "Expected to find POST /api/v1/mcp route registered")
}

func TestCORSHeaders(t *testing.T) {
	adminSvc := &mockAdminService{}
	taskSvc := &mockTaskService{}
	shiftSvc := &mockShiftService{}
	ragSvc := &mockRAGService{}
	automationSvc := &mockAutomationService{}
	schedulerSvc := &mockSchedulerService{}

	srv, err := api.NewServer(api.Config{Address: "127.0.0.1", Port: "8080"}, adminSvc, taskSvc, shiftSvc, ragSvc, automationSvc, schedulerSvc)
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
	automationSvc := &mockAutomationService{}
	schedulerSvc := &mockSchedulerService{}

	// Configure server with real Client ID to arm cryptographic verification
	srv, err := api.NewServer(api.Config{
		Address: "127.0.0.1",
		Port:    "8080",
		OAuth: api.OAuthConfig{
			ClientID: "32555940559.apps.googleusercontent.com",
		},
	}, adminSvc, taskSvc, shiftSvc, ragSvc, automationSvc, schedulerSvc)
	assert.NoError(t, err)

	// Target endpoint calls task service, let's capture the user ID passed from operational.go
	var capturedUserID string
	taskSvc.UpdateStatusFunc = func(ctx context.Context, executionID, status, checklistState, userID string) error {
		capturedUserID = userID
		return nil
	}

	// Scenario A: X-User-ID Header set
	reqBody := `{"status":"COMPLETED"}`
	req := newAuthRequest("PATCH", "/api/v1/organizations/org-123/sites/site-123/tasks/task-123/status", bytes.NewBufferString(reqBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	srv.Engine().ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "11111111-2222-3333-4444-555555555555", capturedUserID)

	// Scenario B: Bearer Authorization Header set (fallback mock when client ID is not configured, but here it is configured so it will fail cryptographic check because it's a mock token)
	req = httptest.NewRequest("PATCH", "/api/v1/organizations/org-123/sites/site-123/tasks/task-123/status", bytes.NewBufferString(reqBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer jwt-user-id-2")
	w = httptest.NewRecorder()
	srv.Engine().ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code) // Fails because it's a mock token and ClientID is configured

	// Scenario C: No Header fails (strictly rejects unauthenticated requests)
	req = httptest.NewRequest("PATCH", "/api/v1/organizations/org-123/sites/site-123/tasks/task-123/status", bytes.NewBufferString(reqBody))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	srv.Engine().ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)

	// Scenario D: Real Google ID Token (retrieved via gcloud) - Cryptographic Validation Pass
	token, err := getGcloudToken()
	if err != nil {
		t.Skipf("Skipping Scenario D: gcloud credentials not available or expired: %v. Run 'gcloud auth login' to run this test.", err)
		return
	}
	req = httptest.NewRequest("PATCH", "/api/v1/organizations/org-123/sites/site-123/tasks/task-123/status", bytes.NewBufferString(reqBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer " + token)
	w = httptest.NewRecorder()
	srv.Engine().ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Logf("Scenario D Response Body: %s", w.Body.String())
	}
	assert.Equal(t, http.StatusOK, w.Code)
	assert.NotEmpty(t, capturedUserID)
	// Must be a valid dynamically registered GORM UUID
	_, parseErr := uuid.Parse(capturedUserID)
	assert.NoError(t, parseErr)
}

func TestAdminEndpoints(t *testing.T) {
	adminSvc := &mockAdminService{}
	taskSvc := &mockTaskService{}
	shiftSvc := &mockShiftService{}
	ragSvc := &mockRAGService{}
	automationSvc := &mockAutomationService{}
	schedulerSvc := &mockSchedulerService{}

	srv, err := api.NewServer(api.Config{Address: "127.0.0.1", Port: "8080"}, adminSvc, taskSvc, shiftSvc, ragSvc, automationSvc, schedulerSvc)
	assert.NoError(t, err)

	t.Run("ListUsers", func(t *testing.T) {
		adminSvc.ListUsersFunc = func(ctx context.Context) ([]*model.User, error) {
			return []*model.User{
				{ID: "user-1", Name: "Hanna", Email: "hanna@walmart.com"},
			}, nil
		}

		req := newAuthRequest("GET", "/api/v1/admin/users", nil)
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
		req := newAuthRequest("POST", "/api/v1/admin/roles", bytes.NewBufferString(reqBody))
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
		req := newAuthRequest("PUT", "/api/v1/admin/users/user-456/roles", bytes.NewBufferString(reqBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		srv.Engine().ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, "user-456", capturedUserID)
		assert.Equal(t, "role-123", capturedRoleID)
	})

	t.Run("GetSchedulerStatus", func(t *testing.T) {
		req := newAuthRequest("GET", "/api/v1/admin/scheduler/status", nil)
		w := httptest.NewRecorder()
		srv.Engine().ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		var status service.SchedulerStatus
		err := json.Unmarshal(w.Body.Bytes(), &status)
		assert.NoError(t, err)
		assert.Equal(t, "mock-node-ID-777", status.NodeID)
		assert.True(t, status.IsLeader)
	})

	t.Run("TriggerSchedulerSweep", func(t *testing.T) {
		req := newAuthRequest("POST", "/api/v1/admin/scheduler/trigger", nil)
		w := httptest.NewRecorder()
		srv.Engine().ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), "Scheduled batch sweeps materialized successfully")
	})
}

func TestOperationalEndpoints(t *testing.T) {
	adminSvc := &mockAdminService{}
	taskSvc := &mockTaskService{}
	shiftSvc := &mockShiftService{}
	ragSvc := &mockRAGService{}
	automationSvc := &mockAutomationService{}
	schedulerSvc := &mockSchedulerService{}

	srv, err := api.NewServer(api.Config{Address: "127.0.0.1", Port: "8080"}, adminSvc, taskSvc, shiftSvc, ragSvc, automationSvc, schedulerSvc)
	assert.NoError(t, err)

	t.Run("GetSiteTasks - Success", func(t *testing.T) {
		taskSvc.GetQueueFunc = func(ctx context.Context, siteID string) ([]*model.TaskExecution, error) {
			assert.Equal(t, "site-abc", siteID)
			return []*model.TaskExecution{
				{ID: "exec-1", TaskTemplateID: "temp-1", Priority: 1, Status: "PENDING"},
			}, nil
		}

		req := newAuthRequest("GET", "/api/v1/organizations/org-abc/sites/site-abc/tasks", nil)
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

		req := newAuthRequest("GET", "/api/v1/organizations/org-abc/tasks", nil)
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

		req := newAuthRequest("GET", "/api/v1/organizations/org-abc/sites/site-abc/users/user-123/tasks", nil)
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
		req := newAuthRequest("POST", "/api/v1/organizations/org-abc/sites/site-abc/tasks/exec-111/override", bytes.NewBufferString(reqBody))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-User-ID", "user-admin") // Keep custom mock to test role privileges
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
		req := newAuthRequest("POST", "/api/v1/organizations/org-abc/sites/site-abc/trades", bytes.NewBufferString(reqBody))
		req.Header.Set("Content-Type", "application/json")
		// Use a valid GORM mock UUID instead of blacklisted "user-initiator"
		req.Header.Set("X-User-ID", "11111111-2222-3333-4444-555555555555")
		w := httptest.NewRecorder()
		srv.Engine().ServeHTTP(w, req)

		assert.Equal(t, http.StatusCreated, w.Code)
		assert.Equal(t, "exec-222", capturedExecID)
		assert.Equal(t, "user-333", capturedProposedAssignee)
		assert.Equal(t, "11111111-2222-3333-4444-555555555555", capturedUserID)
	})
}

func TestOperationalEndpointsFailure(t *testing.T) {
	adminSvc := &mockAdminService{}
	taskSvc := &mockTaskService{}
	shiftSvc := &mockShiftService{}
	ragSvc := &mockRAGService{}
	automationSvc := &mockAutomationService{}
	schedulerSvc := &mockSchedulerService{}

	srv, err := api.NewServer(api.Config{Address: "127.0.0.1", Port: "8080"}, adminSvc, taskSvc, shiftSvc, ragSvc, automationSvc, schedulerSvc)
	assert.NoError(t, err)

	t.Run("UpdateTaskStatus - Error", func(t *testing.T) {
		taskSvc.UpdateStatusFunc = func(ctx context.Context, executionID, status, checklistState, userID string) error {
			return errors.New("database failure")
		}

		reqBody := `{"status":"COMPLETED"}`
		req := newAuthRequest("PATCH", "/api/v1/organizations/org-abc/sites/site-abc/tasks/exec-fail/status", bytes.NewBufferString(reqBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		srv.Engine().ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		assert.Contains(t, w.Body.String(), "database failure")
	})
}

func TestAdminCRUDEndpoints(t *testing.T) {
	adminSvc := &mockAdminService{}
	taskSvc := &mockTaskService{}
	shiftSvc := &mockShiftService{}
	ragSvc := &mockRAGService{}
	automationSvc := &mockAutomationService{}
	schedulerSvc := &mockSchedulerService{}

	srv, err := api.NewServer(api.Config{Address: "127.0.0.1", Port: "8080"}, adminSvc, taskSvc, shiftSvc, ragSvc, automationSvc, schedulerSvc)
	assert.NoError(t, err)

	t.Run("GetUserByID - Success", func(t *testing.T) {
		var capturedID string
		adminSvc.GetUserByIDFunc = func(ctx context.Context, id string) (*model.User, error) {
			capturedID = id
			return &model.User{ID: id, Email: "test@example.com"}, nil
		}

		req := newAuthRequest("GET", "/api/v1/admin/users/user-123", nil)
		w := httptest.NewRecorder()
		srv.Engine().ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, "user-123", capturedID)
		assert.Contains(t, w.Body.String(), "test@example.com")
	})

	t.Run("ListUsersRange - Success", func(t *testing.T) {
		var capturedOffset, capturedLimit int
		adminSvc.ListUsersRangeFunc = func(ctx context.Context, offset, limit int) ([]*model.User, error) {
			capturedOffset = offset
			capturedLimit = limit
			return []*model.User{{ID: "user-range", Email: "range@example.com"}}, nil
		}

		req := newAuthRequest("GET", "/api/v1/admin/users?offset=5&limit=15", nil)
		w := httptest.NewRecorder()
		srv.Engine().ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, 5, capturedOffset)
		assert.Equal(t, 15, capturedLimit)
		assert.Contains(t, w.Body.String(), "range@example.com")
	})

	t.Run("DeleteUser - Success", func(t *testing.T) {
		var capturedID string
		adminSvc.DeleteUserFunc = func(ctx context.Context, id string) error {
			capturedID = id
			return nil
		}

		req := newAuthRequest("DELETE", "/api/v1/admin/users/user-delete", nil)
		w := httptest.NewRecorder()
		srv.Engine().ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, "user-delete", capturedID)
		assert.Contains(t, w.Body.String(), "success")
	})
}
