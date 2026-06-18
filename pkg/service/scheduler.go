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
	"errors"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/rmcguinness/gemini_task_engine/pkg/model"
	"gorm.io/gorm"
)

// SchedulerStatus exposes runtime diagnostics inside cluster nodes.
type SchedulerStatus struct {
	NodeID             string    `json:"node_id"`
	IsLeader           bool      `json:"is_leader"`
	ActiveWorkers      int       `json:"active_workers"`
	TasksClaimed       int64     `json:"tasks_claimed"`
	TasksCompleted     int64     `json:"tasks_completed"`
	TasksFailed        int64     `json:"tasks_failed"`
	LastError          string    `json:"last_error,omitempty"`
	LastLeaderElection time.Time `json:"last_leader_election"`
}

// sqlAdvisoryLockClient decouples PostgreSQL advanced locking dialects from background loops.
// Unlocks seamless database-independent, sandboxed mocking.
type sqlAdvisoryLockClient interface {
	TryAdvisoryLock(ctx context.Context, key int64, conn **sql.Conn, db *gorm.DB, nodeID string) (bool, error)
	ReleaseAdvisoryLock(ctx context.Context, key int64, conn **sql.Conn, nodeID string) error
	CheckAdvisoryLock(ctx context.Context, key int64, conn *sql.Conn) (bool, error)
	ClaimPendingTasks(ctx context.Context, limit int, cutoff time.Time, nodeID string) ([]*model.TaskExecution, error)
	ClaimPendingProcesses(ctx context.Context, limit int, cutoff time.Time, nodeID string) ([]*model.SOPProcess, error)
}

type postgresLockClient struct {
	db *gorm.DB
}

func (c *postgresLockClient) TryAdvisoryLock(ctx context.Context, key int64, conn **sql.Conn, db *gorm.DB, nodeID string) (bool, error) {
	sqlDB, err := db.DB()
	if err != nil {
		return false, err
	}
	dedicatedConn, err := sqlDB.Conn(ctx)
	if err != nil {
		return false, err
	}
	var acquired bool
	err = dedicatedConn.QueryRowContext(ctx, "SELECT pg_try_advisory_lock($1)", key).Scan(&acquired)
	if err != nil {
		dedicatedConn.Close()
		return false, err
	}
	if acquired {
		*conn = dedicatedConn
		return true, nil
	}
	dedicatedConn.Close()
	return false, nil
}

func (c *postgresLockClient) ReleaseAdvisoryLock(ctx context.Context, key int64, conn **sql.Conn, nodeID string) error {
	if *conn != nil {
		_, err := (*conn).ExecContext(ctx, "SELECT pg_advisory_unlock($1)", key)
		(*conn).Close()
		*conn = nil
		return err
	}
	return nil
}

func (c *postgresLockClient) CheckAdvisoryLock(ctx context.Context, key int64, conn *sql.Conn) (bool, error) {
	if conn == nil {
		return false, nil
	}
	var holds bool
	high := int32(key >> 32)
	low := int32(key & 0xffffffff)

	err := conn.QueryRowContext(ctx, `
		SELECT EXISTS (
			SELECT 1 FROM pg_locks 
			WHERE locktype = 'advisory' 
			AND classid = $1 
			AND objid = $2 
			AND pid = pg_backend_pid()
		)
	`, high, low).Scan(&holds)
	return holds, err
}

func (c *postgresLockClient) ClaimPendingTasks(ctx context.Context, limit int, cutoff time.Time, nodeID string) ([]*model.TaskExecution, error) {
	var executions []*model.TaskExecution
	err := c.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		err := tx.Raw(`
			SELECT * FROM task_executions 
			WHERE status = 'PENDING' 
			AND execution_type = 'SYSTEM'
			AND (locked_at IS NULL OR locked_at < ?) 
			ORDER BY priority ASC, created_at ASC 
			LIMIT ? 
			FOR UPDATE SKIP LOCKED
		`, cutoff, limit).Scan(&executions).Error
		if err != nil {
			return err
		}
		if len(executions) > 0 {
			var ids []string
			for _, e := range executions {
				ids = append(ids, e.ID)
			}
			now := time.Now()
			err = tx.Session(&gorm.Session{SkipHooks: true}).Model(&model.TaskExecution{}).
				Where("id IN ?", ids).
				Updates(map[string]interface{}{
					"status":    "IN_PROGRESS",
					"locked_at": &now,
					"locked_by": &nodeID,
				}).Error
			return err
		}
		return nil
	})
	return executions, err
}

func (c *postgresLockClient) ClaimPendingProcesses(ctx context.Context, limit int, cutoff time.Time, nodeID string) ([]*model.SOPProcess, error) {
	var processes []*model.SOPProcess
	err := c.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		err := tx.Raw(`
			SELECT * FROM sop_processes 
			WHERE status = 'PENDING' 
			AND (locked_at IS NULL OR locked_at < ?) 
			ORDER BY created_at ASC 
			LIMIT ? 
			FOR UPDATE SKIP LOCKED
		`, cutoff, limit).Scan(&processes).Error
		if err != nil {
			return err
		}
		if len(processes) > 0 {
			var ids []string
			for _, p := range processes {
				ids = append(ids, p.ID)
			}
			now := time.Now()
			err = tx.Session(&gorm.Session{SkipHooks: true}).Model(&model.SOPProcess{}).
				Where("id IN ?", ids).
				Updates(map[string]interface{}{
					"status":    "IN_PROGRESS",
					"locked_at": &now,
					"locked_by": &nodeID,
				}).Error
			return err
		}
		return nil
	})
	return processes, err
}

// SchedulerConfig holds configuration parameters for the background scheduler daemon.
type SchedulerConfig struct {
	ElectionInterval        string `toml:"election_interval"`
	TaskClaimInterval       string `toml:"task_claim_interval"`
	RAGProcessClaimInterval string `toml:"rag_process_claim_interval"`
	WatchdogInterval        string `toml:"watchdog_interval"`
	SOPCheckInterval        string `toml:"sop_check_interval"`
	LockTimeout             string `toml:"lock_timeout"`
	TaskClaimLimit          int    `toml:"task_claim_limit"`
	RAGProcessClaimLimit    int    `toml:"rag_process_claim_limit"`
}

// Resilient parsing helpers
func parseDurationWithDefault(val string, def time.Duration) time.Duration {
	if val == "" {
		return def
	}
	d, err := time.ParseDuration(val)
	if err != nil {
		log.Printf("[Scheduler] WARNING: Failed to parse duration '%s', falling back to default '%v': %v", val, def, err)
		return def
	}
	return d
}

func parseIntWithDefault(val int, def int) int {
	if val <= 0 {
		return def
	}
	return val
}

// SchedulerService orchestrates leader elections, worker pools, cron sweeps, and dead letter watchdogs.
type SchedulerService interface {
	Start(ctx context.Context) error
	Stop() error
	IsLeader() bool
	NodeID() string
	GetStatus() SchedulerStatus
	ForceTriggerBatchSweep(ctx context.Context) error

	// Operational overrides
	SetNodeID(id string)
	SetLeader(isLeader bool)
	ClaimAndProcessTasks(ctx context.Context)
	RecoverOrOrphanStaleLocks(ctx context.Context)
	AttemptLeaderElection(ctx context.Context)
	ReleaseLeaderLock()
}

type schedulerService struct {
	db                *gorm.DB
	lockClient        sqlAdvisoryLockClient
	ragService        RAGService
	taskService       TaskService
	automationService AutomationService
	nodeID            string
	isLeader          bool
	status            SchedulerStatus
	statusMutex       sync.RWMutex

	leaderConn   *sql.Conn
	connMutex    sync.Mutex
	cancelFunc   context.CancelFunc
	workerWg     sync.WaitGroup
	running      bool
	runningMutex sync.Mutex

	// Configuration variables
	electionInterval        time.Duration
	taskClaimInterval       time.Duration
	ragProcessClaimInterval time.Duration
	watchdogInterval        time.Duration
	sopCheckInterval        time.Duration
	lockTimeout             time.Duration
	taskClaimLimit          int
	ragProcessClaimLimit    int
}

const (
	leaderAdvisoryLockKey = 5555

	// Safe production defaults
	defaultElectionInterval        = 15 * time.Second
	defaultTaskClaimInterval       = 5 * time.Second
	defaultRAGProcessClaimInterval = 5 * time.Second
	defaultWatchdogInterval        = 30 * time.Second
	defaultSOPCheckInterval        = 60 * time.Second
	defaultLockTimeout             = 5 * time.Minute
	defaultTaskClaimLimit          = 5
	defaultRAGProcessClaimLimit    = 2
)

// NewSchedulerService instantiates a new SchedulerService daemon node.
func NewSchedulerService(db *gorm.DB, ragService RAGService, taskService TaskService, automationService AutomationService) SchedulerService {
	return NewSchedulerServiceWithConfig(db, ragService, taskService, automationService, SchedulerConfig{})
}

// NewSchedulerServiceWithConfig instantiates a new SchedulerService daemon node using custom runtime configurations.
func NewSchedulerServiceWithConfig(db *gorm.DB, ragService RAGService, taskService TaskService, automationService AutomationService, cfg SchedulerConfig) SchedulerService {
	uniqueID := "node-" + uuid.New().String()[:8]

	electionInterval := parseDurationWithDefault(cfg.ElectionInterval, defaultElectionInterval)
	taskClaimInterval := parseDurationWithDefault(cfg.TaskClaimInterval, defaultTaskClaimInterval)
	ragProcessClaimInterval := parseDurationWithDefault(cfg.RAGProcessClaimInterval, defaultRAGProcessClaimInterval)
	watchdogInterval := parseDurationWithDefault(cfg.WatchdogInterval, defaultWatchdogInterval)
	sopCheckInterval := parseDurationWithDefault(cfg.SOPCheckInterval, defaultSOPCheckInterval)
	lockTimeout := parseDurationWithDefault(cfg.LockTimeout, defaultLockTimeout)

	taskLimit := parseIntWithDefault(cfg.TaskClaimLimit, defaultTaskClaimLimit)
	ragLimit := parseIntWithDefault(cfg.RAGProcessClaimLimit, defaultRAGProcessClaimLimit)

	return &schedulerService{
		db:                      db,
		lockClient:              &postgresLockClient{db: db},
		ragService:              ragService,
		taskService:             taskService,
		automationService:       automationService,
		nodeID:                  uniqueID,
		status: SchedulerStatus{
			NodeID: uniqueID,
		},
		electionInterval:        electionInterval,
		taskClaimInterval:       taskClaimInterval,
		ragProcessClaimInterval: ragProcessClaimInterval,
		watchdogInterval:        watchdogInterval,
		sopCheckInterval:        sopCheckInterval,
		lockTimeout:             lockTimeout,
		taskClaimLimit:          taskLimit,
		ragProcessClaimLimit:    ragLimit,
	}
}

// NewSchedulerServiceWithClient supports custom DI mapping injection under unit tests.
func NewSchedulerServiceWithClient(db *gorm.DB, ragService RAGService, taskService TaskService, automationService AutomationService, client sqlAdvisoryLockClient) SchedulerService {
	uniqueID := "node-" + uuid.New().String()[:8]
	return &schedulerService{
		db:                      db,
		lockClient:              client,
		ragService:              ragService,
		taskService:             taskService,
		automationService:       automationService,
		nodeID:                  uniqueID,
		status: SchedulerStatus{
			NodeID: uniqueID,
		},
		electionInterval:        defaultElectionInterval,
		taskClaimInterval:       defaultTaskClaimInterval,
		ragProcessClaimInterval: defaultRAGProcessClaimInterval,
		watchdogInterval:        defaultWatchdogInterval,
		sopCheckInterval:        defaultSOPCheckInterval,
		lockTimeout:             defaultLockTimeout,
		taskClaimLimit:          defaultTaskClaimLimit,
		ragProcessClaimLimit:    defaultRAGProcessClaimLimit,
	}
}

func (s *schedulerService) NodeID() string {
	return s.nodeID
}

func (s *schedulerService) SetNodeID(id string) {
	s.nodeID = id
	s.statusMutex.Lock()
	s.status.NodeID = id
	s.statusMutex.Unlock()
}

func (s *schedulerService) SetLeader(isLeader bool) {
	s.isLeader = isLeader
}

func (s *schedulerService) IsLeader() bool {
	s.statusMutex.RLock()
	defer s.statusMutex.RUnlock()
	return s.isLeader
}

func (s *schedulerService) GetStatus() SchedulerStatus {
	s.statusMutex.RLock()
	defer s.statusMutex.RUnlock()
	return s.status
}

func (s *schedulerService) Start(ctx context.Context) error {
	s.runningMutex.Lock()
	if s.running {
		s.runningMutex.Unlock()
		return errors.New("scheduler daemon is already active")
	}
	s.running = true
	s.runningMutex.Unlock()

	log.Printf("[Scheduler][%s] Bootstrapping distributed scheduler daemon...", s.nodeID)

	runCtx, cancel := context.WithCancel(ctx)
	s.cancelFunc = cancel

	// Attempt election immediately on start to prevent startup delay
	s.AttemptLeaderElection(runCtx)

	// 2. Spawn concurrent background loops
	s.workerWg.Add(5)
	go s.leaderElectionLoop(runCtx)
	go s.workerTaskClaimLoop(runCtx)
	go s.workerRAGProcessClaimLoop(runCtx)
	go s.deadLetterWatchdogLoop(runCtx)
	go s.sopMetadataUpdatesCheckLoop(runCtx)

	return nil
}

func (s *schedulerService) Stop() error {
	s.runningMutex.Lock()
	if !s.running {
		s.runningMutex.Unlock()
		return nil
	}
	s.running = false
	s.runningMutex.Unlock()

	log.Printf("[Scheduler][%s] Gracefully terminating scheduler worker pool...", s.nodeID)

	if s.cancelFunc != nil {
		s.cancelFunc()
	}

	s.workerWg.Wait()

	s.ReleaseLeaderLock()

	log.Printf("[Scheduler][%s] Scheduler pool terminated successfully.", s.nodeID)
	return nil
}

func (s *schedulerService) ForceTriggerBatchSweep(ctx context.Context) error {
	if !s.IsLeader() {
		return errors.New("nodes can only invoke batch updates under an active Leader role lease")
	}
	log.Printf("[Scheduler][%s] Force triggering scheduled batch events sweeps...", s.nodeID)
	return s.executeScheduledBatchSweep(ctx)
}

// -----------------------------------------------------------------------------
// Core Loops & Leader Elections

func (s *schedulerService) leaderElectionLoop(ctx context.Context) {
	defer s.workerWg.Done()
	ticker := time.NewTicker(s.electionInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			s.AttemptLeaderElection(ctx)
		}
	}
}

func (s *schedulerService) AttemptLeaderElection(ctx context.Context) {
	s.connMutex.Lock()
	defer s.connMutex.Unlock()

	if s.leaderConn != nil {
		holdsLock, err := s.lockClient.CheckAdvisoryLock(ctx, leaderAdvisoryLockKey, s.leaderConn)
		if err == nil && holdsLock {
			s.updateLeaderStatus(true)
			return
		}
		s.releaseLeaderLockUnlocked()
	}

	acquired, err := s.lockClient.TryAdvisoryLock(ctx, leaderAdvisoryLockKey, &s.leaderConn, s.db, s.nodeID)
	if err != nil {
		s.logError("advisory lock claim execution failed", err)
		return
	}

	if acquired {
		log.Printf("[Scheduler][%s] Leader Election SUCCESS! Node holds standard leader role lease.", s.nodeID)
		s.updateLeaderStatus(true)
	} else {
		s.updateLeaderStatus(false)
	}
}

func (s *schedulerService) ReleaseLeaderLock() {
	s.connMutex.Lock()
	defer s.connMutex.Unlock()
	s.releaseLeaderLockUnlocked()
}

func (s *schedulerService) releaseLeaderLockUnlocked() {
	if s.leaderConn != nil {
		log.Printf("[Scheduler][%s] Relinquishing active leader advisory lock...", s.nodeID)
		_ = s.lockClient.ReleaseAdvisoryLock(context.Background(), leaderAdvisoryLockKey, &s.leaderConn, s.nodeID)
	}
	s.updateLeaderStatus(false)
}

func (s *schedulerService) updateLeaderStatus(leader bool) {
	s.statusMutex.Lock()
	defer s.statusMutex.Unlock()
	s.isLeader = leader
	s.status.IsLeader = leader
	s.status.LastLeaderElection = time.Now()
}

// -----------------------------------------------------------------------------
// Concurrent Worker claiming queues

func (s *schedulerService) workerTaskClaimLoop(ctx context.Context) {
	defer s.workerWg.Done()
	ticker := time.NewTicker(s.taskClaimInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			s.ClaimAndProcessTasks(ctx)
		}
	}
}

func (s *schedulerService) ClaimAndProcessTasks(ctx context.Context) {
	lockCutoff := time.Now().Add(-s.lockTimeout)
	executions, err := s.lockClient.ClaimPendingTasks(ctx, s.taskClaimLimit, lockCutoff, s.nodeID)
	if err != nil {
		s.logError("failed claiming pending task execution records", err)
		return
	}

	if len(executions) == 0 {
		return
	}

	log.Printf("[Scheduler][%s] Claimed %d task executions. Spawning workers...", s.nodeID, len(executions))

	s.statusMutex.Lock()
	s.status.TasksClaimed += int64(len(executions))
	s.status.ActiveWorkers += len(executions)
	s.statusMutex.Unlock()

	for _, exec := range executions {
		go s.executeTaskPayload(ctx, exec)
	}
}

func (s *schedulerService) executeTaskPayload(ctx context.Context, exec *model.TaskExecution) {
	defer func() {
		s.statusMutex.Lock()
		s.status.ActiveWorkers--
		s.statusMutex.Unlock()
	}()

	log.Printf("[Scheduler][%s][Worker] Processing TaskExecution ID %s...", s.nodeID, exec.ID)

	select {
	case <-ctx.Done():
		return
	case <-time.After(100 * time.Millisecond):
	}

	now := time.Now()
	exec.Status = "COMPLETED"
	exec.CompletedAt = &now
	exec.LockedAt = nil
	exec.LockedBy = nil
	exec.LastError = nil

	err := s.db.WithContext(ctx).Save(exec).Error
	if err != nil {
		s.logError(fmt.Sprintf("failed persisting COMPLETED state for execution %s", exec.ID), err)
		return
	}

	s.statusMutex.Lock()
	s.status.TasksCompleted++
	s.statusMutex.Unlock()
	log.Printf("[Scheduler][%s][Worker] TaskExecution ID %s processed successfully.", s.nodeID, exec.ID)
}

// -----------------------------------------------------------------------------
// RAG Ingestion Process claim queue

func (s *schedulerService) workerRAGProcessClaimLoop(ctx context.Context) {
	defer s.workerWg.Done()
	ticker := time.NewTicker(s.ragProcessClaimInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			s.claimAndProcessRAGIndexes(ctx)
		}
	}
}

func (s *schedulerService) claimAndProcessRAGIndexes(ctx context.Context) {
	lockCutoff := time.Now().Add(-s.lockTimeout)
	processes, err := s.lockClient.ClaimPendingProcesses(ctx, s.ragProcessClaimLimit, lockCutoff, s.nodeID)
	if err != nil {
		s.logError("failed claiming pending SOP processes", err)
		return
	}

	if len(processes) == 0 {
		return
	}

	log.Printf("[Scheduler][%s] Claimed %d SOP indexing runs. Processing in parallel...", s.nodeID, len(processes))

	for _, proc := range processes {
		go func(p *model.SOPProcess) {
			s.ragService.ProcessSOPPipeline(ctx, p.SOPID, p.ID)
		}(proc)
	}
}

// -----------------------------------------------------------------------------
// Dead Letter & Timeout watchdog (Leader Only)

func (s *schedulerService) deadLetterWatchdogLoop(ctx context.Context) {
	defer s.workerWg.Done()
	ticker := time.NewTicker(s.watchdogInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if s.IsLeader() {
				s.RecoverOrOrphanStaleLocks(ctx)
			}
		}
	}
}

func (s *schedulerService) RecoverOrOrphanStaleLocks(ctx context.Context) {
	lockCutoff := time.Now().Add(-s.lockTimeout)

	// 1. Audit and recover stale standard task execution locks
	var staleExecutions []*model.TaskExecution
	err := s.db.WithContext(ctx).
		Where("status = 'IN_PROGRESS' AND locked_at < ?", lockCutoff).
		Find(&staleExecutions).Error
	if err != nil {
		s.logError("failed checking for stale task execution locks", err)
		return
	}

	if len(staleExecutions) > 0 {
		log.Printf("[Scheduler][%s][Watchdog] Detected %d orphaned/timed-out locks. Auditing status...", s.nodeID, len(staleExecutions))

		for _, exec := range staleExecutions {
			exec.RetryCount++
			errMsg := fmt.Sprintf("lock timeout error: execution session expired on node %s", *exec.LockedBy)
			exec.LastError = &errMsg

			if exec.RetryCount < exec.MaxRetries {
				exec.Status = "PENDING"
				exec.LockedAt = nil
				exec.LockedBy = nil
				log.Printf("[Scheduler][%s][Watchdog] Recovered execution %s. Reset state to PENDING. Retry: %d/%d", s.nodeID, exec.ID, exec.RetryCount, exec.MaxRetries)
			} else {
				exec.Status = "DEAD_LETTER"
				exec.LockedAt = nil
				exec.LockedBy = nil
				log.Printf("[Scheduler][%s][Watchdog] Terminal retry boundary exceeded for execution %s. Enqueued to DEAD_LETTER.", s.nodeID, exec.ID)
				s.statusMutex.Lock()
				s.status.TasksFailed++
				s.statusMutex.Unlock()
			}

			_ = s.db.WithContext(ctx).Save(exec).Error
		}
	}

	// 2. Audit and recover stale SOP process ingestion locks
	var staleProcesses []*model.SOPProcess
	err = s.db.WithContext(ctx).
		Where("status = 'IN_PROGRESS' AND locked_at < ?", lockCutoff).
		Find(&staleProcesses).Error
	if err != nil {
		s.logError("failed checking for stale SOP process locks", err)
		return
	}

	if len(staleProcesses) > 0 {
		log.Printf("[Scheduler][%s][Watchdog] Detected %d orphaned/timed-out SOP processes. Auditing status...", s.nodeID, len(staleProcesses))

		for _, proc := range staleProcesses {
			proc.RetryCount++
			errMsg := fmt.Sprintf("lock timeout error: indexing process expired on node %s", *proc.LockedBy)
			proc.LastError = &errMsg

			if proc.RetryCount < proc.MaxRetries {
				proc.Status = "PENDING"
				proc.LockedAt = nil
				proc.LockedBy = nil
				log.Printf("[Scheduler][%s][Watchdog] Recovered SOP process %s. Reset state to PENDING. Retry: %d/%d", s.nodeID, proc.ID, proc.RetryCount, proc.MaxRetries)
			} else {
				proc.Status = "FAILED"
				proc.LockedAt = nil
				proc.LockedBy = nil
				log.Printf("[Scheduler][%s][Watchdog] Terminal retry boundary exceeded for SOP process %s. Marked as FAILED.", s.nodeID, proc.ID)
			}

			_ = s.db.WithContext(ctx).Save(proc).Error
		}
	}
}

// -----------------------------------------------------------------------------
// Scheduled cron sweeps (Leader Only)

func (s *schedulerService) sopMetadataUpdatesCheckLoop(ctx context.Context) {
	defer s.workerWg.Done()
	ticker := time.NewTicker(s.sopCheckInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if s.IsLeader() {
				s.executeSOPUpdatesAudit(ctx)
				_ = s.executeScheduledBatchSweep(ctx)
			}
		}
	}
}

func (s *schedulerService) executeSOPUpdatesAudit(ctx context.Context) {
	var sops []*model.SOP
	err := s.db.WithContext(ctx).Find(&sops).Error
	if err != nil {
		s.logError("failed loading SOP profiles for cron check", err)
		return
	}

	for _, doc := range sops {
		changed, err := s.ragService.CheckSOPUpdates(ctx, doc.ID)
		if err != nil {
			s.logError(fmt.Sprintf("failed updating SOP url check for doc %s", doc.ID), err)
			continue
		}
		if changed {
			log.Printf("[Scheduler][%s][Cron] Detected update drift at url for SOP ID %s! New process run created.", s.nodeID, doc.ID)
		}
	}
}

func (s *schedulerService) executeScheduledBatchSweep(ctx context.Context) error {
	log.Printf("[Scheduler][%s][Leader] Materializing scheduled retail shifts & batch allocations sweeps...", s.nodeID)

	var activeShifts []*model.UserEventInstance
	err := s.db.WithContext(ctx).
		Where("event_status = 'EventActive'").
		Find(&activeShifts).Error
	if err != nil {
		return err
	}

	for _, shift := range activeShifts {
		// Lookup the parent site ID of this active shift occurrence
		var siteID string
		err := s.db.WithContext(ctx).
			Table("user_event_instances").
			Joins("JOIN user_event_schedules ON user_event_schedules.id = user_event_instances.schedule_id").
			Joins("JOIN events ON events.id = user_event_schedules.event_id").
			Where("user_event_instances.id = ?", shift.ID).
			Pluck("events.site_id", &siteID).Error
		
		if err != nil {
			s.logError(fmt.Sprintf("failed mapping parent site ID for shift instance %s", shift.ID), err)
			continue
		}

		// Prevent duplicate tasks flooding: check if checklist executions have already been materialized for this shift!
		var count int64
		err = s.db.WithContext(ctx).
			Table("task_executions").
			Where("event_instance_id = ?", shift.ID).
			Count(&count).Error
		if err == nil && count > 0 {
			// Already materialized checklists for this active shift occurrence—bypass duplicate materializations!
			continue
		}

		// Dynamically materialize templates and instantiates checklist tasks inside shift boundaries
		_, err = s.automationService.ProcessBatchEvent(ctx, shift.ID)
		if err == nil {
			log.Printf("[Scheduler][%s][Leader] Shift opening checklist successfully seeded for site ID %s on shift instance %s!", s.nodeID, siteID, shift.ID)
		} else {
			s.logError(fmt.Sprintf("failed instantiating checklist tasks for site %s on shift %s", siteID, shift.ID), err)
		}
	}
	return nil
}

func (s *schedulerService) logError(msg string, err error) {
	s.statusMutex.Lock()
	defer s.statusMutex.Unlock()
	fullErr := fmt.Sprintf("%s: %v", msg, err)
	s.status.LastError = fullErr
	log.Printf("[Scheduler][%s] ERROR: %s", s.nodeID, fullErr)
}
