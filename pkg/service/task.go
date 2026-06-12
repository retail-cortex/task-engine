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
	UpdateStatus(ctx context.Context, executionID, status, checklistState, userID string) error
	OverrideAssetConstraint(ctx context.Context, executionID, assetID, justification, userID string) error
	ProposeTrade(ctx context.Context, executionID, proposedAssigneeID, initiatorID string) error
	ApproveTrade(ctx context.Context, tradeID, supervisorID string) error
	ListPendingTrades(ctx context.Context, userID string) ([]*model.TaskTrade, error)
	AcceptTrade(ctx context.Context, tradeID, targetUserID string) error
	RejectTrade(ctx context.Context, tradeID, targetUserID string) error
	ClaimTask(ctx context.Context, executionID, userID string, userRoleIDs []string) error
	ListActiveSites(ctx context.Context) ([]*model.Site, error)
	GetSiteLocations(ctx context.Context, siteID string) ([]*model.Location, error)
	GetLocationByID(ctx context.Context, id string) (*model.Location, error)
	GetTaskExecutionByID(ctx context.Context, id string) (*model.TaskExecution, error)
	GetSiteIDForExecution(ctx context.Context, execID string) (string, error)
	ListTaskExecutions(ctx context.Context) ([]*model.TaskExecution, error)
	ListTaskExecutionsRange(ctx context.Context, offset, limit int) ([]*model.TaskExecution, error)
	DeleteTaskExecution(ctx context.Context, id string) error
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

func (s *taskService) ListActiveSites(ctx context.Context) ([]*model.Site, error) {
	return s.siteRepo.List(ctx)
}

func (s *taskService) GetOrgTasks(ctx context.Context, orgID string) ([]*model.TaskExecution, error) {
	return s.execRepo.GetOrgTasks(ctx, orgID)
}

func (s *taskService) GetUserSiteTasks(ctx context.Context, siteID, userID string) ([]*model.TaskExecution, error) {
	return s.execRepo.GetUserSiteTasks(ctx, siteID, userID)
}

func (s *taskService) UpdateStatus(ctx context.Context, executionID, status, checklistState, userID string) error {
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
	if checklistState != "" {
		// Try parsing as a single-step delta first to support out-of-order execution safely
		var delta struct {
			Step      int  `json:"step"`
			Completed bool `json:"completed"`
		}
		if err := json.Unmarshal([]byte(checklistState), &delta); err == nil && delta.Step > 0 {
			// Load and parse existing checklist from DB
			var currentChecklist []struct {
				Step      int    `json:"step"`
				Action    string `json:"action"`
				Required  bool   `json:"required"`
				Completed bool   `json:"completed"`
			}
			if len(exec.ChecklistState) > 0 {
				if err := json.Unmarshal(exec.ChecklistState, &currentChecklist); err == nil {
					// Apply the delta to the matched step only
					updated := false
					for idx, step := range currentChecklist {
						if step.Step == delta.Step {
							currentChecklist[idx].Completed = delta.Completed
							updated = true
							break
						}
					}
					if updated {
						updatedBytes, _ := json.Marshal(currentChecklist)
						exec.ChecklistState = model.JSONB(updatedBytes)
					}
				}
			}
		} else {
			// Fallback: overwrite the entire checklist state (for full array updates)
			exec.ChecklistState = model.JSONB([]byte(checklistState))
		}
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
	ctxWithUser := context.WithValue(ctx, "userID", initiatorID)

	exec, err := s.execRepo.FindByID(ctxWithUser, executionID)
	if err != nil {
		return fmt.Errorf("failed to find task execution for trade: %w", err)
	}

	if exec.Status == "COMPLETED" {
		return errors.New("cannot trade completed tasks")
	}
	if exec.Status == "TRADE_PENDING" {
		return errors.New("task is already pending a trade proposal")
	}

	trade := &model.TaskTrade{
		TaskExecutionID:    executionID,
		InitiatorID:        initiatorID,
		ProposedAssigneeID: proposedAssigneeID,
		Status:             "PENDING",
	}

	if err := s.execRepo.CreateTrade(ctxWithUser, trade); err != nil {
		return fmt.Errorf("failed to create trade record: %w", err)
	}

	exec.Status = "TRADE_PENDING"
	if err := s.execRepo.Update(ctxWithUser, exec); err != nil {
		return fmt.Errorf("failed to update task execution status to TRADE_PENDING: %w", err)
	}

	return nil
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
	exec.Status = "PENDING"

	if err := s.execRepo.Update(ctxWithUser, exec); err != nil {
		return fmt.Errorf("failed to reassign task during trade approval: %w", err)
	}

	trade.Status = "APPROVED"
	return s.execRepo.UpdateTrade(ctxWithUser, trade)
}

func (s *taskService) ListPendingTrades(ctx context.Context, userID string) ([]*model.TaskTrade, error) {
	return s.execRepo.FindPendingTradesForUser(ctx, userID)
}

func (s *taskService) AcceptTrade(ctx context.Context, tradeID, targetUserID string) error {
	ctxWithUser := context.WithValue(ctx, "userID", targetUserID)

	trade, err := s.execRepo.FindTradeByID(ctxWithUser, tradeID)
	if err != nil {
		return fmt.Errorf("failed to find trade: %w", err)
	}

	if trade.Status != "PENDING" {
		return errors.New("trade request is not pending")
	}

	if trade.ProposedAssigneeID != targetUserID {
		return errors.New("only the proposed colleague can accept this task trade")
	}

	exec, err := s.execRepo.FindByID(ctxWithUser, trade.TaskExecutionID)
	if err != nil {
		return fmt.Errorf("failed to find execution for trade: %w", err)
	}

	// Performs physical handover
	exec.AssigneeID = &trade.ProposedAssigneeID

	// Reset status back to PENDING (or IN_PROGRESS if steps were completed)
	hasCompletedSteps := false
	if len(exec.ChecklistState) > 0 {
		var checklist []map[string]interface{}
		if err := json.Unmarshal(exec.ChecklistState, &checklist); err == nil {
			for _, step := range checklist {
				if comp, ok := step["completed"].(bool); ok && comp {
					hasCompletedSteps = true
					break
				}
			}
		}
	}
	if hasCompletedSteps {
		exec.Status = "IN_PROGRESS"
	} else {
		exec.Status = "PENDING"
	}

	if err := s.execRepo.Update(ctxWithUser, exec); err != nil {
		return fmt.Errorf("failed to reassign task during trade acceptance: %w", err)
	}

	trade.Status = "APPROVED"
	return s.execRepo.UpdateTrade(ctxWithUser, trade)
}

func (s *taskService) RejectTrade(ctx context.Context, tradeID, targetUserID string) error {
	ctxWithUser := context.WithValue(ctx, "userID", targetUserID)

	trade, err := s.execRepo.FindTradeByID(ctxWithUser, tradeID)
	if err != nil {
		return fmt.Errorf("failed to find trade: %w", err)
	}

	if trade.Status != "PENDING" {
		return errors.New("trade request is not pending")
	}

	if trade.ProposedAssigneeID != targetUserID {
		return errors.New("only the proposed colleague can reject this task trade")
	}

	exec, err := s.execRepo.FindByID(ctxWithUser, trade.TaskExecutionID)
	if err != nil {
		return fmt.Errorf("failed to find execution for trade: %w", err)
	}

	// Reset status back to PENDING (or IN_PROGRESS if steps were completed)
	hasCompletedSteps := false
	if len(exec.ChecklistState) > 0 {
		var checklist []map[string]interface{}
		if err := json.Unmarshal(exec.ChecklistState, &checklist); err == nil {
			for _, step := range checklist {
				if comp, ok := step["completed"].(bool); ok && comp {
					hasCompletedSteps = true
					break
				}
			}
		}
	}
	if hasCompletedSteps {
		exec.Status = "IN_PROGRESS"
	} else {
		exec.Status = "PENDING"
	}

	if err := s.execRepo.Update(ctxWithUser, exec); err != nil {
		return fmt.Errorf("failed to restore task status during trade rejection: %w", err)
	}

	trade.Status = "REJECTED"
	return s.execRepo.UpdateTrade(ctxWithUser, trade)
}

func (s *taskService) ClaimTask(ctx context.Context, executionID, userID string, userRoleIDs []string) error {
	ctxWithUser := context.WithValue(ctx, "userID", userID)

	exec, err := s.execRepo.FindByID(ctxWithUser, executionID)
	if err != nil {
		return fmt.Errorf("failed to find task execution: %w", err)
	}

	// Support claiming any open task (not COMPLETED)
	if exec.Status == "COMPLETED" {
		return fmt.Errorf("task is not in a claimable state (current status: %s)", exec.Status)
	}

	// If it is a trade, we validate and approve the trade record
	var trade *model.TaskTrade
	if exec.Status == "TRADE_PENDING" {
		// 1. Find the active pending trade record for this execution
		var err error
		trade, err = s.execRepo.FindPendingTradeByExecution(ctxWithUser, executionID)
		if err != nil {
			return fmt.Errorf("failed to find pending trade record for this execution: %w", err)
		}

		// 2. Validate if user is authorized to take this task:
		// - If the user is the original initiator of the trade swap (trade.InitiatorID == userID), they are allowed to take it back.
		// - OR, if the task does not enforce target roles (TargetRoleID is nil or empty).
		// - OR, if the user possesses the TargetRoleID.
		isAuthorized := false
		if trade.InitiatorID == userID {
			isAuthorized = true
		} else if exec.Task.TargetRoleID == nil || *exec.Task.TargetRoleID == "" {
			isAuthorized = true
		} else {
			for _, rID := range userRoleIDs {
				if rID == *exec.Task.TargetRoleID {
					isAuthorized = true
					break
				}
			}
		}

		if !isAuthorized {
			return errors.New("unauthorized: you do not possess the eligible operational role required to take this task")
		}
	} else {
		// If it is a standard open task, we validate:

		// - And the user must possess the required TargetRoleID if set.
		isAuthorized := false
		if exec.Task.TargetRoleID == nil || *exec.Task.TargetRoleID == "" {
			isAuthorized = true
		} else {
			for _, rID := range userRoleIDs {
				if rID == *exec.Task.TargetRoleID {
					isAuthorized = true
					break
				}
			}
		}

		if !isAuthorized {
			return errors.New("unauthorized: you do not possess the eligible operational role required to claim this task")
		}
	}

	// 3. Reassign assignee to the claiming user
	exec.AssigneeID = &userID

	// 4. Reset status based on completed steps
	hasCompletedSteps := false
	if len(exec.ChecklistState) > 0 {
		var checklist []map[string]interface{}
		if err := json.Unmarshal(exec.ChecklistState, &checklist); err == nil {
			for _, step := range checklist {
				if comp, ok := step["completed"].(bool); ok && comp {
					hasCompletedSteps = true
					break
				}
			}
		}
	}
	if hasCompletedSteps {
		exec.Status = "IN_PROGRESS"
	} else {
		exec.Status = "PENDING"
	}

	// 5. Save task execution
	if err := s.execRepo.Update(ctxWithUser, exec); err != nil {
		return fmt.Errorf("failed to update task execution assignee: %w", err)
	}

	// 6. If it was a trade, close the trade record
	if trade != nil {
		trade.Status = "APPROVED"
		trade.ProposedAssigneeID = userID
		return s.execRepo.UpdateTrade(ctxWithUser, trade)
	}

	return nil
}

func (s *taskService) GetSiteLocations(ctx context.Context, siteID string) ([]*model.Location, error) {
	site, err := s.siteRepo.FindByID(ctx, siteID)
	if err != nil {
		return nil, err
	}
	var locs []*model.Location
	for i := range site.Locations {
		locs = append(locs, &site.Locations[i])
	}
	return locs, nil
}

func (s *taskService) GetLocationByID(ctx context.Context, id string) (*model.Location, error) {
	return s.siteRepo.FindLocationByID(ctx, id)
}

func (s *taskService) GetTaskExecutionByID(ctx context.Context, id string) (*model.TaskExecution, error) {
	return s.execRepo.FindByID(ctx, id)
}

func (s *taskService) GetSiteIDForExecution(ctx context.Context, execID string) (string, error) {
	return s.execRepo.GetSiteIDForExecution(ctx, execID)
}

func (s *taskService) ListTaskExecutions(ctx context.Context) ([]*model.TaskExecution, error) {
	return s.execRepo.List(ctx)
}

func (s *taskService) ListTaskExecutionsRange(ctx context.Context, offset, limit int) ([]*model.TaskExecution, error) {
	return s.execRepo.ListRange(ctx, offset, limit)
}

func (s *taskService) DeleteTaskExecution(ctx context.Context, id string) error {
	return s.execRepo.Delete(ctx, id)
}
