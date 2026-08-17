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

package service

import (
	"context"
	"database/sql"
	"testing"
	"time"

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

func TestSchedulerService_HelpersAndConfig(t *testing.T) {
	d := parseDurationWithDefault("100ms", 50*time.Millisecond)
	assert.Equal(t, 100*time.Millisecond, d)

	dDef := parseDurationWithDefault("invalid-dur", 50*time.Millisecond)
	assert.Equal(t, 50*time.Millisecond, dDef)

	i := parseIntWithDefault(10, 5)
	assert.Equal(t, 10, i)

	iDef := parseIntWithDefault(-1, 5)
	assert.Equal(t, 5, iDef)

	cfg := SchedulerConfig{
		ElectionInterval:        "10ms",
		TaskClaimInterval:       "10ms",
		RAGProcessClaimInterval: "10ms",
		SOPCheckInterval:        "10ms",
		WatchdogInterval:        "10ms",
		LockTimeout:             "5m",
		TaskClaimLimit:          2,
		RAGProcessClaimLimit:    2,
	}

	db, err := gorm.Open(dummyDialector{}, &gorm.Config{
		DryRun: true,
	})
	assert.NoError(t, err)

	svc := NewSchedulerServiceWithConfig(db, nil, nil, nil, cfg)
	assert.NotNil(t, svc)

	sImpl, ok := svc.(*schedulerService)
	assert.True(t, ok)

	sImpl.SetNodeID("test-node-1")
	assert.Equal(t, "test-node-1", sImpl.NodeID())

	sImpl.updateLeaderStatus(true)
	assert.True(t, sImpl.IsLeader())

	status := sImpl.GetStatus()
	assert.Equal(t, "test-node-1", status.NodeID)
	assert.True(t, status.IsLeader)

	sImpl.ReleaseLeaderLock()
	assert.False(t, sImpl.IsLeader())

	// Test executeTaskPayload on dummy steps
	exec := &model.TaskExecution{
		ID: "exec-1",
		ChecklistState: model.JSONB(`[
			{"step_id": 1, "action": "SET_STATUS", "target": "COMPLETED"}
		]`),
	}
	sImpl.executeTaskPayload(context.Background(), exec)
	assert.Equal(t, "COMPLETED", exec.Status)
}

type mockLockClient struct{}

func (m mockLockClient) TryAdvisoryLock(ctx context.Context, key int64, conn **sql.Conn, db *gorm.DB, nodeID string) (bool, error) {
	if conn != nil {
		*conn = new(sql.Conn)
	}
	return true, nil
}
func (m mockLockClient) ReleaseAdvisoryLock(ctx context.Context, key int64, conn **sql.Conn, nodeID string) error {
	return nil
}
func (m mockLockClient) CheckAdvisoryLock(ctx context.Context, key int64, conn *sql.Conn) (bool, error) {
	return true, nil
}
func (m mockLockClient) ClaimPendingTasks(ctx context.Context, limit int, cutoff time.Time, nodeID string) ([]*model.TaskExecution, error) {
	return nil, nil
}
func (m mockLockClient) ClaimPendingProcesses(ctx context.Context, limit int, cutoff time.Time, nodeID string) ([]*model.SOPProcess, error) {
	return nil, nil
}

func TestSchedulerService_StartStopCycle(t *testing.T) {
	db, err := gorm.Open(dummyDialector{}, &gorm.Config{
		DryRun: true,
	})
	assert.NoError(t, err)

	svc := NewSchedulerServiceWithClient(db, nil, nil, nil, mockLockClient{})
	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()

	err = svc.Start(ctx)
	assert.NoError(t, err)

	time.Sleep(40 * time.Millisecond)
	svc.ForceTriggerBatchSweep(ctx)

	err = svc.Stop()
	assert.NoError(t, err)

	svcDefault := NewSchedulerService(db, nil, nil, nil)
	assert.NotNil(t, svcDefault)
}

func TestSchedulerService_LoopsAndSweeps(t *testing.T) {
	db, err := gorm.Open(dummyDialector{}, &gorm.Config{
		DryRun: true,
	})
	assert.NoError(t, err)

	svc := NewSchedulerServiceWithClient(db, nil, nil, nil, mockLockClient{})
	sImpl := svc.(*schedulerService)

	ctx := context.Background()
	sImpl.AttemptLeaderElection(ctx)
	sImpl.AttemptLeaderElection(ctx)

	sImpl.claimAndProcessRAGIndexes(ctx)
	sImpl.executeSOPUpdatesAudit(ctx)
	sImpl.executeScheduledBatchSweep(ctx)
	sImpl.logError("test msg", assert.AnError)

	pgClient := &postgresLockClient{db: db}
	_, _ = pgClient.ClaimPendingTasks(ctx, 10, time.Now(), "node-1")
	_, _ = pgClient.ClaimPendingProcesses(ctx, 10, time.Now(), "node-1")
	var conn *sql.Conn
	_, _ = pgClient.TryAdvisoryLock(ctx, 123, &conn, db, "node-1")
	_ = pgClient.ReleaseAdvisoryLock(ctx, 123, &conn, "node-1")
	_, _ = pgClient.CheckAdvisoryLock(ctx, 123, conn)
}
