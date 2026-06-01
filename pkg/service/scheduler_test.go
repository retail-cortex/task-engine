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
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/rmcguinness/gemini_task_engine/pkg/model"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// Mock SQLAdvisoryLockClient
type mockAdvisoryLockClient struct {
	TryAdvisoryLockFunc       func(ctx context.Context, key int64, conn **sql.Conn, db *gorm.DB, nodeID string) (bool, error)
	ReleaseAdvisoryLockFunc   func(ctx context.Context, key int64, conn **sql.Conn, nodeID string) error
	CheckAdvisoryLockFunc     func(ctx context.Context, key int64, conn *sql.Conn) (bool, error)
	ClaimPendingTasksFunc     func(ctx context.Context, limit int, cutoff time.Time, nodeID string) ([]*model.TaskExecution, error)
	ClaimPendingProcessesFunc func(ctx context.Context, limit int, cutoff time.Time, nodeID string) ([]*model.SOPProcess, error)
}

func (m *mockAdvisoryLockClient) TryAdvisoryLock(ctx context.Context, key int64, conn **sql.Conn, db *gorm.DB, nodeID string) (bool, error) {
	if m.TryAdvisoryLockFunc != nil {
		return m.TryAdvisoryLockFunc(ctx, key, conn, db, nodeID)
	}
	return false, nil
}

func (m *mockAdvisoryLockClient) ReleaseAdvisoryLock(ctx context.Context, key int64, conn **sql.Conn, nodeID string) error {
	if m.ReleaseAdvisoryLockFunc != nil {
		return m.ReleaseAdvisoryLockFunc(ctx, key, conn, nodeID)
	}
	return nil
}

func (m *mockAdvisoryLockClient) CheckAdvisoryLock(ctx context.Context, key int64, conn *sql.Conn) (bool, error) {
	if m.CheckAdvisoryLockFunc != nil {
		return m.CheckAdvisoryLockFunc(ctx, key, conn)
	}
	return false, nil
}

func (m *mockAdvisoryLockClient) ClaimPendingTasks(ctx context.Context, limit int, cutoff time.Time, nodeID string) ([]*model.TaskExecution, error) {
	if m.ClaimPendingTasksFunc != nil {
		return m.ClaimPendingTasksFunc(ctx, limit, cutoff, nodeID)
	}
	return nil, nil
}

func (m *mockAdvisoryLockClient) ClaimPendingProcesses(ctx context.Context, limit int, cutoff time.Time, nodeID string) ([]*model.SOPProcess, error) {
	if m.ClaimPendingProcessesFunc != nil {
		return m.ClaimPendingProcessesFunc(ctx, limit, cutoff, nodeID)
	}
	return nil, nil
}

// Mock RAGService
type mockRAGService struct {
	RAGService
	CheckSOPUpdatesFunc func(ctx context.Context, sopID string) (bool, error)
}

func (m *mockRAGService) CheckSOPUpdates(ctx context.Context, sopID string) (bool, error) {
	if m.CheckSOPUpdatesFunc != nil {
		return m.CheckSOPUpdatesFunc(ctx, sopID)
	}
	return false, nil
}

// Mock TaskService embedding the interface directly to satisfy interface compatibility constraints
type mockTaskService struct {
	TaskService
	GetQueueFunc func(ctx context.Context, siteID string) ([]*model.TaskExecution, error)
}

func (m *mockTaskService) GetQueue(ctx context.Context, siteID string) ([]*model.TaskExecution, error) {
	if m.GetQueueFunc != nil {
		return m.GetQueueFunc(ctx, siteID)
	}
	return nil, nil
}

// Mock AutomationService embedding the interface directly to satisfy interface compatibility constraints
type mockAutomationService struct {
	AutomationService
	ProcessBatchEventFunc func(ctx context.Context, eventInstanceID string) ([]*model.TaskExecution, error)
}

func (m *mockAutomationService) ProcessBatchEvent(ctx context.Context, eventInstanceID string) ([]*model.TaskExecution, error) {
	if m.ProcessBatchEventFunc != nil {
		return m.ProcessBatchEventFunc(ctx, eventInstanceID)
	}
	return nil, nil
}

func stringPtr(s string) *string {
	return &s
}

type mockConnPool struct {
	gorm.ConnPool
}

func TestSchedulerService_LeaderElection(t *testing.T) {
	ragSvc := &mockRAGService{}
	taskSvc := &mockTaskService{}

	// Instantiate dummy dry-run GORM DB connection
	db, err := gorm.Open(postgres.New(postgres.Config{
		Conn: &mockConnPool{},
	}), &gorm.Config{
		DryRun: true,
	})
	assert.NoError(t, err)

	t.Run("successfully elects a single leader in a multi-node cluster", func(t *testing.T) {
		var lockMutex sync.Mutex
		var activeLeaderNodeID string

		mockClient := &mockAdvisoryLockClient{
			TryAdvisoryLockFunc: func(ctx context.Context, key int64, conn **sql.Conn, db *gorm.DB, nodeID string) (bool, error) {
				lockMutex.Lock()
				defer lockMutex.Unlock()
				assert.Equal(t, int64(5555), key)

				if activeLeaderNodeID == "" {
					activeLeaderNodeID = nodeID
					*conn = &sql.Conn{}
					return true, nil
				}
				return false, nil
			},
			CheckAdvisoryLockFunc: func(ctx context.Context, key int64, conn *sql.Conn) (bool, error) {
				lockMutex.Lock()
				defer lockMutex.Unlock()
				return activeLeaderNodeID == "node-A", nil // Simulate connection state check
			},
			ReleaseAdvisoryLockFunc: func(ctx context.Context, key int64, conn **sql.Conn, nodeID string) error {
				lockMutex.Lock()
				defer lockMutex.Unlock()
				if activeLeaderNodeID == nodeID {
					activeLeaderNodeID = ""
					*conn = nil
				}
				return nil
			},
		}

		autoSvc := &mockAutomationService{}
		nodeA := NewSchedulerServiceWithClient(db, ragSvc, taskSvc, autoSvc, mockClient).(*schedulerService)
		nodeB := NewSchedulerServiceWithClient(db, ragSvc, taskSvc, autoSvc, mockClient).(*schedulerService)

		nodeA.nodeID = "node-A"
		nodeB.nodeID = "node-B"

		// 1. Node A attempts election first, succeeds
		nodeA.attemptLeaderElection(context.Background())
		assert.True(t, nodeA.IsLeader())
		assert.Equal(t, "node-A", activeLeaderNodeID)

		// 2. Node B attempts election, fails (single leader maintained)
		nodeB.attemptLeaderElection(context.Background())
		assert.False(t, nodeB.IsLeader())

		// 3. Node A relinquishes leader lock
		nodeA.releaseLeaderLock()
		assert.False(t, nodeA.IsLeader())
		assert.Empty(t, activeLeaderNodeID)

		// 4. Node B attempts election again, successfully claims leader lease (Self-healing)
		mockClient.CheckAdvisoryLockFunc = func(ctx context.Context, key int64, conn *sql.Conn) (bool, error) {
			lockMutex.Lock()
			defer lockMutex.Unlock()
			return activeLeaderNodeID == "node-B", nil
		}

		nodeB.attemptLeaderElection(context.Background())
		assert.True(t, nodeB.IsLeader())
		assert.Equal(t, "node-B", activeLeaderNodeID)
	})
}

func TestSchedulerService_WorkerClaimQueue(t *testing.T) {
	ragSvc := &mockRAGService{}
	taskSvc := &mockTaskService{}

	db, err := gorm.Open(postgres.New(postgres.Config{
		Conn: &mockConnPool{},
	}), &gorm.Config{
		DryRun: true,
	})
	assert.NoError(t, err)

	t.Run("successfully claims and processes tasks using SKIP LOCKED", func(t *testing.T) {
		mockTasks := []*model.TaskExecution{
			{ID: "task-ticket-1", TaskTemplateID: "template-1", Status: "PENDING", Priority: 1},
			{ID: "task-ticket-2", TaskTemplateID: "template-2", Status: "PENDING", Priority: 1},
		}

		var claimedNodeID string
		mockClient := &mockAdvisoryLockClient{
			ClaimPendingTasksFunc: func(ctx context.Context, limit int, cutoff time.Time, nodeID string) ([]*model.TaskExecution, error) {
				claimedNodeID = nodeID
				assert.Equal(t, 5, limit)
				return mockTasks, nil
			},
		}

		autoSvc := &mockAutomationService{}
		node := NewSchedulerServiceWithClient(db, ragSvc, taskSvc, autoSvc, mockClient).(*schedulerService)

		// Execute claim dynamically
		node.claimAndProcessTasks(context.Background())

		assert.Equal(t, node.nodeID, claimedNodeID)
		status := node.GetStatus()
		assert.Equal(t, int64(2), status.TasksClaimed)
	})
}

func TestSchedulerService_DeadLetterWatchdog(t *testing.T) {
	ragSvc := &mockRAGService{}
	taskSvc := &mockTaskService{}

	t.Run("watchdog recovers timed-out locks and routes to dead letter terminal state", func(t *testing.T) {
		staleExecutions := []*model.TaskExecution{
			// Retry under limit: should be recovered back to PENDING
			{
				ID:             "exec-recover",
				Status:         "IN_PROGRESS",
				LockedBy:       stringPtr("node-stale-1"),
				RetryCount:     0,
				MaxRetries:     3,
				ChecklistState: model.JSONB("{}"),
			},
			// Retry count at limit: should be sent to DEAD_LETTER queue
			{
				ID:             "exec-dead",
				Status:         "IN_PROGRESS",
				LockedBy:       stringPtr("node-stale-1"),
				RetryCount:     2, // Mapped to 3rd retry attempt
				MaxRetries:     3,
				ChecklistState: model.JSONB("{}"),
			},
		}

		staleProcesses := []*model.SOPProcess{
			{
				ID:               "proc-recover",
				SOPID:            "11111111-aaaa-bbbb-cccc-999999990001",
				ChunkingStrategy: "RECURSIVE_CHARACTER",
				EmbeddingModel:   "text-embedding-004",
				Status:           "IN_PROGRESS",
				LockedBy:         stringPtr("node-stale-1"),
				RetryCount:       0,
				MaxRetries:       3,
			},
			{
				ID:               "proc-dead",
				SOPID:            "11111111-aaaa-bbbb-cccc-999999990001",
				ChunkingStrategy: "RECURSIVE_CHARACTER",
				EmbeddingModel:   "text-embedding-004",
				Status:           "IN_PROGRESS",
				LockedBy:         stringPtr("node-stale-1"),
				RetryCount:       2,
				MaxRetries:       3,
			},
		}

		// Dry-run GORM DB intercepting Save calls context
		var savedExecutions []*model.TaskExecution
		var savedProcesses []*model.SOPProcess
		var saveMutex sync.Mutex

		db, err := gorm.Open(postgres.New(postgres.Config{
			Conn: &mockConnPool{},
		}), &gorm.Config{
			DryRun: true,
		})
		assert.NoError(t, err)

		// Intercept Save calls via mock database dialect wrapper (DryRun intercepts standard executions,
		// but the DB array matches target objects. We test by intercepting transactional updates or simple assertions!)
		// Wait! Since GORM dry-run maps to standard Save calls, does the updated fields directly mutate
		// the slices references we passed?
		// Yes! In GORM, passing `exec` by reference to `Save(&exec)` or mutating the slice objects directly in the loop:
		//     exec.Status = "PENDING"
		// Modifies the actual objects in our testing slice in-place!
		// This is a beautiful feature of Go pointer references!
		// So running the watchdog directly on the slice will mutate the slice items in-place,
		// letting us assert the mutations directly!
		// Construct GORM DB mock query returns!
		// Instead of querying the live DB, the watchdog executes a Find query.
		// Wait, under GORM dry run, Find queries return `nil` and don't populate slices.
		// So to test the recovery loop payload `recoverOrOrphanStaleLocks` cleanly:
		// We can mock or inject GORM mock callbacks, OR wait!
		// We can directly verify the mapping logic by writing a simple unit test calling a mock executor.
		// Or wait! We can inject the stale database slice directly into GORM using mock DB sessions!
		// Wait, can we mock `s.db` using a GORM database pointer that delegates to in-memory SQLite?
		// Oh! SQLite does NOT throw any failures on:
		//     s.db.Where("status = 'IN_PROGRESS' AND locked_at < ?", cutoff).Find(&staleExecutions)
		// Because this SELECT query does NOT use pgvector or advisory locks! It is a standard, primitive GORM lookup!
		// And SQLite supports `Find` and `Save` flawlessly!
		// And since we created `pkg/model/seed_test.go` and verified GORM enums, GORM will auto-migrate schemas on SQLite effortlessly!
		// Wait! How do we import SQLite without `scan_dependencies`?
		// Wait, in `go.mod`:
		// `gorm.io/driver/postgres` is loaded.
		// Can the postgres driver be mocked without dialing?
		// Yes, we just did using `DryRun: true` and `Conn: &sql.DB{}`!
		// But in GORM dry-run, `Find` doesn't return data.
		// Wait! Can we mock the raw query returns?
		// GORM database has an interface database engine callback list (`db.Callback().Query().Register(...)`)!
		// We can register a custom query callback on our dry-run GORM DB that intercepts GORM queries and populates mock data!
		// This is standard in GORM testing.
		// But wait! Is there a simpler way?
		// Yes! We can test the GORM watchdog recovery loop directly by calling a private processing test routine, or wait:
		db.Callback().Query().Register("mock_find_stale", func(d *gorm.DB) {
			sqlStr := d.Statement.SQL.String()
			if strings.Contains(sqlStr, "status = 'IN_PROGRESS'") {
				destVal := d.Statement.Dest
				if destPtr, ok := destVal.(*[]*model.TaskExecution); ok {
					*destPtr = staleExecutions
				} else if destPtr, ok := destVal.(*[]*model.SOPProcess); ok {
					*destPtr = staleProcesses
				}
			}
		})
		
		db.Callback().Update().Register("mock_save_recorder", func(d *gorm.DB) {
			saveMutex.Lock()
			defer saveMutex.Unlock()
			if exec, ok := d.Statement.Dest.(*model.TaskExecution); ok {
				savedExecutions = append(savedExecutions, exec)
			} else if proc, ok := d.Statement.Dest.(*model.SOPProcess); ok {
				savedProcesses = append(savedProcesses, proc)
			}
		})
		
		mockClient := &mockAdvisoryLockClient{}
		autoSvc := &mockAutomationService{}
		node := NewSchedulerServiceWithClient(db, ragSvc, taskSvc, autoSvc, mockClient).(*schedulerService)
		node.isLeader = true

		node.recoverOrOrphanStaleLocks(context.Background())
		
		assert.Len(t, savedExecutions, 2)
		
		// Exec-recover check
		assert.Equal(t, "exec-recover", savedExecutions[0].ID)
		assert.Equal(t, "PENDING", savedExecutions[0].Status)
		assert.Equal(t, 1, savedExecutions[0].RetryCount)
		assert.Nil(t, savedExecutions[0].LockedBy)
		assert.Nil(t, savedExecutions[0].LockedAt)
		assert.Contains(t, *savedExecutions[0].LastError, "lock timeout error")
		
		// Exec-dead check
		assert.Equal(t, "exec-dead", savedExecutions[1].ID)
		assert.Equal(t, "DEAD_LETTER", savedExecutions[1].Status)
		assert.Equal(t, 3, savedExecutions[1].RetryCount) // Incremented to max retries limit
		assert.Nil(t, savedExecutions[1].LockedBy)
		assert.Nil(t, savedExecutions[1].LockedAt)
		assert.Contains(t, *savedExecutions[1].LastError, "lock timeout error")

		// Assert SOP process recoveries
		assert.Len(t, savedProcesses, 2)

		// Proc-recover check
		assert.Equal(t, "proc-recover", savedProcesses[0].ID)
		assert.Equal(t, "PENDING", savedProcesses[0].Status)
		assert.Equal(t, 1, savedProcesses[0].RetryCount)
		assert.Nil(t, savedProcesses[0].LockedBy)
		assert.Nil(t, savedProcesses[0].LockedAt)
		assert.Contains(t, *savedProcesses[0].LastError, "lock timeout error")

		// Proc-dead check
		assert.Equal(t, "proc-dead", savedProcesses[1].ID)
		assert.Equal(t, "FAILED", savedProcesses[1].Status) // Terminal FAILED status for processes
		assert.Equal(t, 3, savedProcesses[1].RetryCount)
		assert.Nil(t, savedProcesses[1].LockedBy)
		assert.Nil(t, savedProcesses[1].LockedAt)
		assert.Contains(t, *savedProcesses[1].LastError, "lock timeout error")
		
		status := node.GetStatus()
		assert.Equal(t, int64(1), status.TasksFailed) // Incremented for dead letters
	})
}
