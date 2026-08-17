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

package persistence

import (
	"context"
	"testing"

	"github.com/rmcguinness/gemini_task_engine/pkg/model"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"gorm.io/gorm/schema"
)

type dummyDialector struct{}

func (d dummyDialector) Name() string                                                { return "dummy" }
func (d dummyDialector) Initialize(db *gorm.DB) error                                { return nil }
func (d dummyDialector) Migrator(db *gorm.DB) gorm.Migrator                          { return nil }
func (d dummyDialector) DataTypeOf(field *schema.Field) string                       { return "text" }
func (d dummyDialector) DefaultValueOf(field *schema.Field) clause.Expression        { return nil }
func (d dummyDialector) BindVarTo(writer clause.Writer, stmt *gorm.Statement, v interface{}) {}
func (d dummyDialector) QuoteTo(writer clause.Writer, str string)                    { writer.WriteString(str) }
func (d dummyDialector) Explain(sql string, vars ...interface{}) string              { return sql }

func setupDryRunDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(dummyDialector{}, &gorm.Config{
		DryRun: true,
	})
	assert.NoError(t, err)
	return db
}

func TestUserRepository_CRUD(t *testing.T) {
	db := setupDryRunDB(t)
	repo := NewUserRepository(db)
	ctx := context.Background()

	u := &model.User{
		ID:    "u1",
		Email: "test@google.com",
	}

	assert.NoError(t, repo.Create(ctx, u))
	assert.NoError(t, repo.Update(ctx, u))
	assert.NoError(t, repo.AddRole(ctx, "u1", "r1"))

	_, err := repo.FindByID(ctx, "u1")
	assert.NoError(t, err)

	_, err = repo.FindByOAuth(ctx, "google", "1000")
	assert.NoError(t, err)

	_, err = repo.List(ctx)
	assert.NoError(t, err)

	_, err = repo.ListRange(ctx, 0, 10)
	assert.NoError(t, err)

	_, err = repo.ListActiveOnShiftUsers(ctx, "site1")
	assert.NoError(t, err)

	assert.NoError(t, repo.Delete(ctx, "u1"))

	role := &model.Role{ID: "r1", Name: "Admin"}
	assert.NoError(t, repo.CreateRole(ctx, role))
	assert.NoError(t, repo.UpdateRole(ctx, role))

	_, err = repo.FindRoleByID(ctx, "r1")
	assert.NoError(t, err)

	_, err = repo.ListRoles(ctx)
	assert.NoError(t, err)

	_, err = repo.ListRolesRange(ctx, 0, 10)
	assert.NoError(t, err)

	assert.NoError(t, repo.DeleteRole(ctx, "r1"))
}

func TestOrganizationRepository_CRUD(t *testing.T) {
	db := setupDryRunDB(t)
	repo := NewOrganizationRepository(db)
	ctx := context.Background()

	org := &model.Organization{ID: "org1", Name: "Retail Org"}
	assert.NoError(t, repo.Create(ctx, org))
	assert.NoError(t, repo.Update(ctx, org))
	_, err := repo.FindByID(ctx, "org1")
	assert.NoError(t, err)

	_, err = repo.List(ctx)
	assert.NoError(t, err)

	_, err = repo.ListRange(ctx, 0, 10)
	assert.NoError(t, err)

	assert.NoError(t, repo.Delete(ctx, "org1"))
}

func TestSiteRepository_CRUD(t *testing.T) {
	db := setupDryRunDB(t)
	repo := NewSiteRepository(db)
	ctx := context.Background()

	site := &model.Site{ID: "site1", Name: "Store 101", OrganizationID: "org1"}
	assert.NoError(t, repo.Create(ctx, site))
	assert.NoError(t, repo.Update(ctx, site))
	_, err := repo.FindByID(ctx, "site1")
	assert.NoError(t, err)

	_, err = repo.List(ctx)
	assert.NoError(t, err)

	_, err = repo.ListRange(ctx, 0, 10)
	assert.NoError(t, err)

	loc := &model.Location{ID: "loc1", SiteID: "site1", Name: "Aisle A"}
	assert.NoError(t, repo.CreateLocation(ctx, loc))
	assert.NoError(t, repo.UpdateLocation(ctx, loc))

	_, err = repo.FindLocationByID(ctx, "loc1")
	assert.NoError(t, err)

	_, err = repo.ListLocations(ctx)
	assert.NoError(t, err)

	_, err = repo.ListLocationsRange(ctx, 0, 10)
	assert.NoError(t, err)

	assert.NoError(t, repo.DeleteLocation(ctx, "loc1"))

	asset := &model.Asset{ID: "asset1", LocationID: "loc1", Name: "Forklift"}
	assert.NoError(t, repo.CreateAsset(ctx, asset))
	assert.NoError(t, repo.UpdateAsset(ctx, asset))

	_, err = repo.FindAssetByID(ctx, "asset1")
	assert.NoError(t, err)

	_, err = repo.ListAssets(ctx)
	assert.NoError(t, err)

	_, err = repo.ListAssetsRange(ctx, 0, 10)
	assert.NoError(t, err)

	assert.NoError(t, repo.DeleteAsset(ctx, "asset1"))
	assert.NoError(t, repo.Delete(ctx, "site1"))
}

func TestSOPRepository_CRUD(t *testing.T) {
	db := setupDryRunDB(t)
	repo := NewSOPRepository(db)
	ctx := context.Background()

	sop := &model.SOP{ID: "sop1", Title: "Produce Freshness"}
	assert.NoError(t, repo.Create(ctx, sop))
	assert.NoError(t, repo.Update(ctx, sop))
	_, err := repo.FindByID(ctx, "sop1")
	assert.NoError(t, err)

	_, err = repo.List(ctx)
	assert.NoError(t, err)

	_, err = repo.ListRange(ctx, 0, 10)
	assert.NoError(t, err)

	proc := &model.SOPProcess{ID: "proc1", SOPID: "sop1", Status: "PENDING"}
	assert.NoError(t, repo.CreateProcess(ctx, proc))
	assert.NoError(t, repo.UpdateProcess(ctx, proc))

	_, err = repo.FindProcessByID(ctx, "proc1")
	assert.NoError(t, err)

	_, err = repo.ListProcesses(ctx)
	assert.NoError(t, err)

	_, err = repo.ListProcessesRange(ctx, 0, 10)
	assert.NoError(t, err)

	chunks := []*model.SOPChunk{{ID: "c1", SOPID: "sop1"}}
	assert.NoError(t, repo.CreateChunks(ctx, chunks))

	_, err = repo.QuerySimilarity(ctx, model.Float32Vector{0.1, 0.2}, 5)
	if err != nil {
		assert.Contains(t, err.Error(), "dry run mode unsupported")
	}

	assert.NoError(t, repo.DeleteProcess(ctx, "proc1"))
	assert.NoError(t, repo.Delete(ctx, "sop1"))
}

func TestTaskRepository_CRUD(t *testing.T) {
	db := setupDryRunDB(t)
	repo := NewTaskRepository(db)
	ctx := context.Background()

	task := &model.Task{ID: "task1", Name: "STOCK_MILK"}
	assert.NoError(t, repo.Create(ctx, task))
	assert.NoError(t, repo.Update(ctx, task))
	_, err := repo.FindByID(ctx, "task1")
	assert.NoError(t, err)

	_, err = repo.List(ctx)
	assert.NoError(t, err)

	_, err = repo.ListRange(ctx, 0, 10)
	assert.NoError(t, err)

	rule := &model.TaskApprovalRule{ID: "r1", TaskID: "task1"}
	assert.NoError(t, repo.AddApprovalRule(ctx, rule))

	assert.NoError(t, repo.Delete(ctx, "task1"))
}

func TestTaskExecutionRepository_CRUD(t *testing.T) {
	db := setupDryRunDB(t)
	repo := NewTaskExecutionRepository(db)
	ctx := context.Background()

	exec := &model.TaskExecution{ID: "exec1", Status: "ASSIGNED"}
	assert.NoError(t, repo.Create(ctx, exec))
	assert.NoError(t, repo.Update(ctx, exec))
	_, err := repo.FindByID(ctx, "exec1")
	assert.NoError(t, err)

	_, err = repo.List(ctx)
	assert.NoError(t, err)

	_, err = repo.ListRange(ctx, 0, 10)
	assert.NoError(t, err)

	_, err = repo.GetQueue(ctx, "site1")
	assert.NoError(t, err)

	_, err = repo.GetOrgTasks(ctx, "org1")
	assert.NoError(t, err)

	_, err = repo.GetUserSiteTasks(ctx, "site1", "user1")
	assert.NoError(t, err)

	_, err = repo.GetSiteIDForExecution(ctx, "exec1")
	if err != nil {
		assert.Contains(t, err.Error(), "dry run mode unsupported")
	}

	trade := &model.TaskTrade{ID: "trade1", InitiatorID: "u1", TaskExecutionID: "exec1"}
	assert.NoError(t, repo.CreateTrade(ctx, trade))
	assert.NoError(t, repo.UpdateTrade(ctx, trade))

	_, err = repo.FindTradeByID(ctx, "trade1")
	assert.NoError(t, err)

	_, err = repo.FindPendingTradesForUser(ctx, "u1")
	assert.NoError(t, err)

	_, err = repo.FindPendingTradeByExecution(ctx, "exec1")
	assert.NoError(t, err)

	audit := &model.TaskExecutionAudit{ID: "a1", TaskExecutionID: "exec1"}
	assert.NoError(t, repo.CreateAudit(ctx, audit))

	assert.NoError(t, repo.Delete(ctx, "exec1"))
}

func TestShiftAgentSessionRepository_CRUD(t *testing.T) {
	db := setupDryRunDB(t)
	repo := NewShiftAgentSessionRepository(db)
	ctx := context.Background()

	session := &model.ShiftAgentSession{ID: "session1", AssigneeID: "u1"}
	assert.NoError(t, repo.Create(ctx, session))
	assert.NoError(t, repo.Update(ctx, session))
	_, err := repo.FindByID(ctx, "session1")
	assert.NoError(t, err)

	_, err = repo.FindByShift(ctx, "u1", "shift1")
	assert.NoError(t, err)

	_, err = repo.List(ctx)
	assert.NoError(t, err)

	_, err = repo.ListRange(ctx, 0, 10)
	assert.NoError(t, err)

	assert.NoError(t, repo.Delete(ctx, "session1"))
}

func TestDBConfigAndInit(t *testing.T) {
	cfg := DBConfig{
		Host:     "localhost",
		Port:     "5432",
		User:     "user",
		Password: "password",
		DBName:   "dbname",
		SSLMode:  "disable",
	}
	assert.Equal(t, "localhost", cfg.Host)
	assert.Equal(t, "5432", cfg.Port)

	// Test InitDB with invalid connection string project format
	cfgAlloy := DBConfig{
		ConnectionString: "projects/invalid-uri",
	}
	_, err := InitDB(cfgAlloy)
	assert.Error(t, err)
}
