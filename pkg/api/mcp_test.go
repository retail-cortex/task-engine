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
	"strings"
	"testing"

	"github.com/rmcguinness/gemini_task_engine/pkg/model"
	"github.com/rmcguinness/gemini_task_engine/pkg/service"
	"github.com/stretchr/testify/assert"
)

type mockMCPTaskForAPI struct {
	service.TaskService
}

func (m *mockMCPTaskForAPI) GetQueue(ctx context.Context, siteID string) ([]*model.TaskExecution, error) {
	r1 := "r1"
	u1 := "00000000-0000-0000-0000-000000000001"
	u2 := "00000000-0000-0000-0000-000000000002"
	return []*model.TaskExecution{
		{
			ID: "00000000-0000-0000-0000-000000000010",
			Task: model.Task{
				ID:           "t1",
				Name:         "STOCK_MILK",
				TargetRoleID: &r1,
			},
			Status:      "PENDING",
			Description: "aisle 1 dairy cooler",
		},
		{
			ID: "00000000-0000-0000-0000-000000000011",
			Task: model.Task{
				ID:   "t2",
				Name: "CHECK_REGISTER",
			},
			Status:     "IN_PROGRESS",
			AssigneeID: &u1,
			Assignee:   &model.User{Name: "Ryan", Email: "ryan@google.com"},
		},
		{
			ID: "00000000-0000-0000-0000-000000000012",
			Task: model.Task{
				ID:   "t3",
				Name: "CLEAN_SPILL",
			},
			Status:     "PENDING",
			AssigneeID: &u2,
		},
	}, nil
}
func (m *mockMCPTaskForAPI) GetLocationByID(ctx context.Context, id string) (*model.Location, error) {
	name := "Aisle 1"
	fnType := "STANDARD"
	if strings.Contains(id, "vault") {
		name = "Vault Room"
		fnType = "VAULT"
	} else if strings.Contains(id, "register") {
		name = "Register 1"
		fnType = "REGISTER"
	} else if strings.Contains(id, "produce") {
		name = "Produce Wall"
		fnType = "PRODUCE"
	} else if strings.Contains(id, "showcase") {
		name = "Showcase Atrium"
		fnType = "SHOWCASE"
	} else if strings.Contains(id, "dock") {
		name = "Loading Dock Bay"
		fnType = "LOADING_DOCK"
	}
	return &model.Location{ID: id, Name: name, LocationFunctionType: fnType}, nil
}
func (m *mockMCPTaskForAPI) GetSiteLocations(ctx context.Context, siteID string) ([]*model.Location, error) {
	return []*model.Location{{ID: "l1", Name: "Aisle 1"}}, nil
}
func (m *mockMCPTaskForAPI) GetTaskExecutionByID(ctx context.Context, id string) (*model.TaskExecution, error) {
	if id == "err" {
		return nil, errors.New("err")
	}
	tplID := ""
	chk := []byte{}
	var assignee *model.User
	if id == "reg-open" {
		tplID = "d000fa44-0000-0000-0000-000000000000"
		chk = []byte(`[{"step":1,"action":"Unlock Register Drawers"},{"step":2,"action":"Verify Cash Vault"},{"step":3,"action":"Receipt Thermal Roll"}]`)
		assignee = &model.User{Name: "Bob"}
	} else if id == "prod-fresh" {
		tplID = "d000fa55-0000-0000-0000-000000000000"
		chk = []byte(`[{"step":1,"action":"Cull Spoiled Greens"},{"step":2,"action":"Log Chiller Temp"},{"step":3,"action":"Rotate Crates"}]`)
		assignee = &model.User{Email: "bob@google.com"}
	} else if id == "shelf-rep" {
		tplID = "d000fa66-0000-0000-0000-000000000000"
		chk = []byte(`[{"step":1,"action":"Locate Pallet Box"},{"step":2,"action":"Transport Carton Pallet"}]`)
	} else if id == "e1" {
		chk = []byte(`[{"step":1,"action":"Generic Step"}]`)
	}
	return &model.TaskExecution{
		ID:             id,
		TaskTemplateID: tplID,
		ChecklistState: chk,
		Assignee:       assignee,
		Task: model.Task{
			ID:   "t1",
			Name: "STOCK_MILK",
		},
	}, nil
}
func (m *mockMCPTaskForAPI) GetSiteIDForExecution(ctx context.Context, execID string) (string, error) {
	return "s1", nil
}
func (m *mockMCPTaskForAPI) ListActiveSites(ctx context.Context) ([]*model.Site, error) {
	return []*model.Site{{ID: "s1", Name: "Store 1"}}, nil
}
func (m *mockMCPTaskForAPI) OverrideAssetConstraint(ctx context.Context, execID, assetID, just, userID string) error {
	return nil
}
func (m *mockMCPTaskForAPI) ProposeTrade(ctx context.Context, execID, propID, userID string) error {
	return nil
}
func (m *mockMCPTaskForAPI) AcceptTrade(ctx context.Context, tradeID, userID string) error {
	return nil
}
func (m *mockMCPTaskForAPI) RejectTrade(ctx context.Context, tradeID, userID string) error {
	return nil
}
func (m *mockMCPTaskForAPI) ListPendingTrades(ctx context.Context, userID string) ([]*model.TaskTrade, error) {
	return []*model.TaskTrade{{ID: "tr1"}}, nil
}
func (m *mockMCPTaskForAPI) ClaimTask(ctx context.Context, execID, userID string, roles []string) error {
	return nil
}
func (m *mockMCPTaskForAPI) UpdateStatus(ctx context.Context, execID, status, state, userID string) error {
	return nil
}

type mockMCPShiftForAPI struct {
	service.ShiftService
}

func (m *mockMCPShiftForAPI) GetUserProfile(ctx context.Context, userID string) (*model.User, error) {
	return &model.User{
		ID:    userID,
		Name:  "Test User",
		Roles: []model.Role{{Name: "ADMIN"}},
		Sites: []model.Site{{ID: "s1", Name: "Store 1"}},
	}, nil
}

type mockMCPRAGForAPI struct {
	service.RAGService
}

func (m *mockMCPRAGForAPI) QuerySimilarity(ctx context.Context, vec model.Float32Vector, limit int) ([]*model.SOPChunk, error) {
	return []*model.SOPChunk{{ChunkIndex: 1, Content: "SOP doc"}}, nil
}

func TestMCPHandler_Tools(t *testing.T) {
	mcpHandler, err := NewMCPHandler(
		&mockMCPTaskForAPI{},
		&mockMCPRAGForAPI{},
		&mockMCPShiftForAPI{},
		&mockOpAutomationForAPI{},
	)
	assert.NoError(t, err)
	assert.NotNil(t, mcpHandler)

	userID := "00000000-0000-0000-0000-000000000001"
	ctx := context.WithValue(context.Background(), "userID", userID)

	// 1. HandleGetTasks
	res, err := mcpHandler.HandleGetTasks(ctx, GetTasksArgs{SiteID: "s1"})
	assert.NoError(t, err)
	assert.NotNil(t, res)

	a2ui := "a2ui"
	resA2UI, err := mcpHandler.HandleGetTasks(ctx, GetTasksArgs{SiteID: "s1", Format: &a2ui})
	assert.NoError(t, err)
	assert.NotNil(t, resA2UI)

	r1 := "r1"
	u1 := "00000000-0000-0000-0000-000000000001"
	locID := "l1"
	rawFormat := "raw"
	_, _ = mcpHandler.HandleGetTasks(ctx, GetTasksArgs{SiteID: "s1", RoleID: &r1, Format: &rawFormat})
	_, _ = mcpHandler.HandleGetTasks(ctx, GetTasksArgs{SiteID: "s1", AssigneeID: &u1, Format: &rawFormat})
	_, _ = mcpHandler.HandleGetTasks(ctx, GetTasksArgs{SiteID: "s1", LocationID: &locID, Format: &rawFormat})

	// 2. HandleGetTaskDetails
	resDet, err := mcpHandler.HandleGetTaskDetails(ctx, GetTaskDetailsArgs{ExecutionID: "exec-1"})
	assert.NoError(t, err)
	assert.NotNil(t, resDet)

	resDetA2UI, err := mcpHandler.HandleGetTaskDetails(ctx, GetTaskDetailsArgs{ExecutionID: "exec-1", Format: &a2ui})
	assert.NoError(t, err)
	assert.NotNil(t, resDetA2UI)

	for _, id := range []string{"reg-open", "prod-fresh", "shelf-rep"} {
		_, _ = mcpHandler.HandleGetTaskDetails(ctx, GetTaskDetailsArgs{ExecutionID: id, Format: &rawFormat})
		_, _ = mcpHandler.HandleGetTaskDetails(ctx, GetTaskDetailsArgs{ExecutionID: id, Format: &a2ui})
	}

	// 3. HandleGetSiteLocations & Beacon/Layout resolution
	resLoc, err := mcpHandler.HandleGetSiteLocations(ctx, GetSiteLocationsArgs{SiteID: "s1"})
	assert.NoError(t, err)
	assert.NotNil(t, resLoc)

	sites := []string{"s1", "44444444-4444-4444-4444-444444440001", "44444444-4444-4444-4444-444444440002"}
	locIDs := []string{"loc-vault", "loc-register", "loc-produce", "loc-showcase", "loc-dock", "loc-other"}
	for _, sid := range sites {
		for _, lid := range locIDs {
			lID := lid
			_, _ = mcpHandler.HandleGetSiteLocations(ctx, GetSiteLocationsArgs{SiteID: sid, Format: &a2ui, LocationID: &lID})
			_, _ = mcpHandler.getTaskBeaconAndLayout(ctx, "task "+lid, lid, sid)
		}
	}

	// 4. HandleOverrideAsset
	resOvr, err := mcpHandler.HandleOverrideAsset(ctx, OverrideAssetArgs{ExecutionID: "e1", AssetID: "a1", Justification: "broken"})
	assert.NoError(t, err)
	assert.NotNil(t, resOvr)

	// 5. HandleProposeTrade
	resProp, err := mcpHandler.HandleProposeTrade(ctx, ProposeTradeArgs{TaskExecutionID: "e1", ProposedAssigneeID: userID})
	assert.NoError(t, err)
	assert.NotNil(t, resProp)

	// 6. HandleAcceptTrade
	resAcc, err := mcpHandler.HandleAcceptTrade(ctx, AcceptTradeArgs{TradeID: "tr1"})
	assert.NoError(t, err)
	assert.NotNil(t, resAcc)

	// 7. HandleRejectTrade
	resRej, err := mcpHandler.HandleRejectTrade(ctx, RejectTradeArgs{TradeID: "tr1"})
	assert.NoError(t, err)
	assert.NotNil(t, resRej)

	// 8. HandleQuerySOP
	resSOP, err := mcpHandler.HandleQuerySOP(ctx, QuerySOPArgs{QueryText: "produce milk"})
	assert.NoError(t, err)
	assert.NotNil(t, resSOP)

	// 9. HandleTriggerAlert
	resAlert, err := mcpHandler.HandleTriggerAlert(ctx, TriggerAlertArgs{SiteID: "s1", OrganizerID: "org1", EventType: "EventStockoutCorrect", Description: "out of milk"})
	assert.NoError(t, err)
	assert.NotNil(t, resAlert)

	// 10. HandleGetUserContext
	resCtx, err := mcpHandler.HandleGetUserContext(ctx, GetUserContextArgs{})
	assert.NoError(t, err)
	assert.NotNil(t, resCtx)

	// 11. HandleClaimTask
	resClaim, err := mcpHandler.HandleClaimTask(ctx, ClaimTaskArgs{ExecutionID: "e1"})
	assert.NoError(t, err)
	assert.NotNil(t, resClaim)

	// 12. HandleUpdateTaskStatus
	resUpd, err := mcpHandler.HandleUpdateTaskStatus(ctx, UpdateTaskStatusArgs{ExecutionID: "e1", Status: "COMPLETED"})
	assert.NoError(t, err)
	assert.NotNil(t, resUpd)

	// 13. HandleListPendingTrades
	resPend, err := mcpHandler.HandleListPendingTrades(ctx, ListPendingTradesArgs{})
	assert.NoError(t, err)
	assert.NotNil(t, resPend)

	// 14. HandleGetWeather
	resWth, err := mcpHandler.HandleGetWeather(ctx, GetWeatherArgs{Station: "KDFW"})
	assert.NoError(t, err)
	assert.NotNil(t, resWth)

	// 15. HandleGetStoreSelector
	resSel, err := mcpHandler.HandleGetStoreSelector(ctx, GetStoreSelectorArgs{})
	assert.NoError(t, err)
	assert.NotNil(t, resSel)

	// Raw checklist formats
	rawfmt := "raw"
	for _, id := range []string{"reg-open", "prod-fresh", "shelf-rep", "e1"} {
		rRaw, e := mcpHandler.HandleGetTaskDetails(ctx, GetTaskDetailsArgs{ExecutionID: id, Format: &rawfmt})
		assert.NoError(t, e)
		assert.NotNil(t, rRaw)
	}
	_, eErr := mcpHandler.HandleGetTaskDetails(ctx, GetTaskDetailsArgs{ExecutionID: "err"})
	assert.Error(t, eErr)

	// interceptingWriter methods
	iw := &interceptingWriter{body: new(bytes.Buffer)}
	assert.Equal(t, http.StatusOK, iw.Status())
	iw.WriteHeader(http.StatusCreated)
	assert.Equal(t, http.StatusCreated, iw.Status())
	n, errW := iw.WriteString("hello")
	assert.NoError(t, errW)
	assert.Equal(t, 5, n)
}
