package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/rmcguinness/gemini_task_engine/pkg/model"
	"github.com/rmcguinness/gemini_task_engine/pkg/persistence"
)

// TaskService orchestrates prioritized queues, state transitions, constraint overrides, and trades.
type TaskService interface {
	GetQueue(ctx context.Context, siteID string) ([]*model.TaskExecution, error)
	GetOrgTasks(ctx context.Context, orgID string) ([]*model.TaskExecution, error)
	GetUserSiteTasks(ctx context.Context, siteID, userID string) ([]*model.TaskExecution, error)
	UpdateStatus(ctx context.Context, executionID, status, userID string) error
	OverrideAssetConstraint(ctx context.Context, executionID, assetID, justification, userID string) error
	ProposeTrade(ctx context.Context, executionID, proposedAssigneeID, initiatorID string) error
	ApproveTrade(ctx context.Context, tradeID, supervisorID string) error
}

type taskService struct {
	execRepo persistence.TaskExecutionRepository
	siteRepo persistence.SiteRepository
}

// NewTaskService instantiates a new TaskService.
func NewTaskService(execRepo persistence.TaskExecutionRepository, siteRepo persistence.SiteRepository) TaskService {
	return &taskService{
		execRepo: execRepo,
		siteRepo:  siteRepo,
	}
}

func (s *taskService) GetQueue(ctx context.Context, siteID string) ([]*model.TaskExecution, error) {
	return s.execRepo.GetQueue(ctx, siteID)
}

func (s *taskService) GetOrgTasks(ctx context.Context, orgID string) ([]*model.TaskExecution, error) {
	return s.execRepo.GetOrgTasks(ctx, orgID)
}

func (s *taskService) GetUserSiteTasks(ctx context.Context, siteID, userID string) ([]*model.TaskExecution, error) {
	return s.execRepo.GetUserSiteTasks(ctx, siteID, userID)
}

func (s *taskService) UpdateStatus(ctx context.Context, executionID, status, userID string) error {
	// Bind the user to context so GORM hooks can retrieve it for audit trails
	ctxWithUser := context.WithValue(ctx, "userID", userID)

	exec, err := s.execRepo.FindByID(ctxWithUser, executionID)
	if err != nil {
		return fmt.Errorf("failed to find task execution: %w", err)
	}

	exec.Status = status
	if status == "COMPLETED" {
		now := time.Now()
		exec.CompletedAt = &now
	}

	return s.execRepo.Update(ctxWithUser, exec)
}

func (s *taskService) OverrideAssetConstraint(ctx context.Context, executionID, assetID, justification, userID string) error {
	ctxWithUser := context.WithValue(ctx, "userID", userID)

	exec, err := s.execRepo.FindByID(ctxWithUser, executionID)
	if err != nil {
		return fmt.Errorf("failed to find task execution: %w", err)
	}

	asset, err := s.siteRepo.FindAssetByID(ctxWithUser, assetID)
	if err != nil {
		return fmt.Errorf("failed to find asset: %w", err)
	}

	// Decodes existing override flags JSONB
	var flags map[string]interface{}
	if len(exec.OverrideFlags) > 0 {
		if err := json.Unmarshal(exec.OverrideFlags, &flags); err != nil {
			flags = make(map[string]interface{})
		}
	} else {
		flags = make(map[string]interface{})
	}

	// Inject the new override justification log
	flags[assetID] = map[string]interface{}{
		"asset_name":    asset.Name,
		"justification": justification,
		"overridden_by": userID,
		"timestamp":     time.Now().Format(time.RFC3339),
	}

	flagsBytes, err := json.Marshal(flags)
	if err != nil {
		return err
	}
	exec.OverrideFlags = model.JSONB(flagsBytes)

	// Perform update. The GORM AfterUpdate hook automatically writes to task_execution_audits.
	return s.execRepo.Update(ctxWithUser, exec)
}

func (s *taskService) ProposeTrade(ctx context.Context, executionID, proposedAssigneeID, initiatorID string) error {
	trade := &model.TaskTrade{
		TaskExecutionID:    executionID,
		InitiatorID:        initiatorID,
		ProposedAssigneeID: proposedAssigneeID,
		Status:             "PENDING",
	}
	return s.execRepo.CreateTrade(ctx, trade)
}

func (s *taskService) ApproveTrade(ctx context.Context, tradeID, supervisorID string) error {
	ctxWithUser := context.WithValue(ctx, "userID", supervisorID)

	trade, err := s.execRepo.FindTradeByID(ctxWithUser, tradeID)
	if err != nil {
		return fmt.Errorf("failed to find trade: %w", err)
	}

	if trade.Status != "PENDING" {
		return errors.New("trade request is not pending")
	}

	exec, err := s.execRepo.FindByID(ctxWithUser, trade.TaskExecutionID)
	if err != nil {
		return fmt.Errorf("failed to find execution for trade: %w", err)
	}

	// Performs physical handover
	exec.AssigneeID = &trade.ProposedAssigneeID

	if err := s.execRepo.Update(ctxWithUser, exec); err != nil {
		return fmt.Errorf("failed to reassign task during trade approval: %w", err)
	}

	trade.Status = "APPROVED"
	return s.execRepo.UpdateTrade(ctxWithUser, trade)
}
