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
	"errors"
	"testing"

	"github.com/rmcguinness/gemini_task_engine/pkg/model"
	"github.com/rmcguinness/gemini_task_engine/pkg/persistence"
	"github.com/stretchr/testify/assert"
)

type MockTaskExecutionRepository struct {
	persistence.TaskExecutionRepository
	FindByIDFunc                 func(ctx context.Context, id string) (*model.TaskExecution, error)
	UpdateFunc                   func(ctx context.Context, e *model.TaskExecution) error
	CreateTradeFunc              func(ctx context.Context, t *model.TaskTrade) error
	FindTradeByIDFunc            func(ctx context.Context, id string) (*model.TaskTrade, error)
	UpdateTradeFunc              func(ctx context.Context, t *model.TaskTrade) error
	FindPendingTradesForUserFunc func(ctx context.Context, userID string) ([]*model.TaskTrade, error)
}

func (m *MockTaskExecutionRepository) FindByID(ctx context.Context, id string) (*model.TaskExecution, error) {
	if m.FindByIDFunc != nil {
		return m.FindByIDFunc(ctx, id)
	}
	return nil, errors.New("not implemented")
}

func (m *MockTaskExecutionRepository) Update(ctx context.Context, e *model.TaskExecution) error {
	if m.UpdateFunc != nil {
		return m.UpdateFunc(ctx, e)
	}
	return nil
}

func (m *MockTaskExecutionRepository) CreateTrade(ctx context.Context, t *model.TaskTrade) error {
	if m.CreateTradeFunc != nil {
		return m.CreateTradeFunc(ctx, t)
	}
	return nil
}

func (m *MockTaskExecutionRepository) FindTradeByID(ctx context.Context, id string) (*model.TaskTrade, error) {
	if m.FindTradeByIDFunc != nil {
		return m.FindTradeByIDFunc(ctx, id)
	}
	return nil, errors.New("not implemented")
}

func (m *MockTaskExecutionRepository) UpdateTrade(ctx context.Context, t *model.TaskTrade) error {
	if m.UpdateTradeFunc != nil {
		return m.UpdateTradeFunc(ctx, t)
	}
	return nil
}

func (m *MockTaskExecutionRepository) FindPendingTradesForUser(ctx context.Context, userID string) ([]*model.TaskTrade, error) {
	if m.FindPendingTradesForUserFunc != nil {
		return m.FindPendingTradesForUserFunc(ctx, userID)
	}
	return nil, nil
}

func TestTaskService_ProposeTrade(t *testing.T) {
	t.Run("task execution not found", func(t *testing.T) {
		mockRepo := &MockTaskExecutionRepository{
			FindByIDFunc: func(ctx context.Context, id string) (*model.TaskExecution, error) {
				return nil, errors.New("not found")
			},
		}
		svc := NewTaskService(mockRepo, nil)
		err := svc.ProposeTrade(context.Background(), "exec-1", "user-jenna", "user-ryan")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to find task execution")
	})

	t.Run("cannot trade completed tasks", func(t *testing.T) {
		mockRepo := &MockTaskExecutionRepository{
			FindByIDFunc: func(ctx context.Context, id string) (*model.TaskExecution, error) {
				return &model.TaskExecution{ID: "exec-1", Status: "COMPLETED"}, nil
			},
		}
		svc := NewTaskService(mockRepo, nil)
		err := svc.ProposeTrade(context.Background(), "exec-1", "user-jenna", "user-ryan")
		assert.Error(t, err)
		assert.Equal(t, "cannot trade completed tasks", err.Error())
	})

	t.Run("task is already pending a trade", func(t *testing.T) {
		mockRepo := &MockTaskExecutionRepository{
			FindByIDFunc: func(ctx context.Context, id string) (*model.TaskExecution, error) {
				return &model.TaskExecution{ID: "exec-1", Status: "TRADE_PENDING"}, nil
			},
		}
		svc := NewTaskService(mockRepo, nil)
		err := svc.ProposeTrade(context.Background(), "exec-1", "user-jenna", "user-ryan")
		assert.Error(t, err)
		assert.Equal(t, "task is already pending a trade proposal", err.Error())
	})

	t.Run("propose trade success", func(t *testing.T) {
		exec := &model.TaskExecution{ID: "exec-1", Status: "PENDING"}
		var createdTrade *model.TaskTrade
		var updatedExec *model.TaskExecution

		mockRepo := &MockTaskExecutionRepository{
			FindByIDFunc: func(ctx context.Context, id string) (*model.TaskExecution, error) {
				return exec, nil
			},
			CreateTradeFunc: func(ctx context.Context, t *model.TaskTrade) error {
				createdTrade = t
				return nil
			},
			UpdateFunc: func(ctx context.Context, e *model.TaskExecution) error {
				updatedExec = e
				return nil
			},
		}
		svc := NewTaskService(mockRepo, nil)
		err := svc.ProposeTrade(context.Background(), "exec-1", "user-jenna", "user-ryan")
		assert.NoError(t, err)

		assert.NotNil(t, createdTrade)
		assert.Equal(t, "exec-1", createdTrade.TaskExecutionID)
		assert.Equal(t, "user-ryan", createdTrade.InitiatorID)
		assert.Equal(t, "user-jenna", createdTrade.ProposedAssigneeID)
		assert.Equal(t, "PENDING", createdTrade.Status)

		assert.NotNil(t, updatedExec)
		assert.Equal(t, "TRADE_PENDING", updatedExec.Status)
	})
}

func TestTaskService_AcceptTrade(t *testing.T) {
	t.Run("trade not found", func(t *testing.T) {
		mockRepo := &MockTaskExecutionRepository{
			FindTradeByIDFunc: func(ctx context.Context, id string) (*model.TaskTrade, error) {
				return nil, errors.New("trade not found")
			},
		}
		svc := NewTaskService(mockRepo, nil)
		err := svc.AcceptTrade(context.Background(), "trade-1", "user-jenna")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to find trade")
	})

	t.Run("trade not pending", func(t *testing.T) {
		mockRepo := &MockTaskExecutionRepository{
			FindTradeByIDFunc: func(ctx context.Context, id string) (*model.TaskTrade, error) {
				return &model.TaskTrade{ID: "trade-1", Status: "APPROVED"}, nil
			},
		}
		svc := NewTaskService(mockRepo, nil)
		err := svc.AcceptTrade(context.Background(), "trade-1", "user-jenna")
		assert.Error(t, err)
		assert.Equal(t, "trade request is not pending", err.Error())
	})

	t.Run("unauthorized proposed assignee cannot accept", func(t *testing.T) {
		mockRepo := &MockTaskExecutionRepository{
			FindTradeByIDFunc: func(ctx context.Context, id string) (*model.TaskTrade, error) {
				return &model.TaskTrade{
					ID:                 "trade-1",
					ProposedAssigneeID: "user-jenna",
					Status:             "PENDING",
				}, nil
			},
		}
		svc := NewTaskService(mockRepo, nil)
		// Ryan tries to accept a trade targeted at Jenna! (Violates Principle of Least Privilege!)
		err := svc.AcceptTrade(context.Background(), "trade-1", "user-ryan")
		assert.Error(t, err)
		assert.Equal(t, "only the proposed colleague can accept this task trade", err.Error())
	})

	t.Run("accept trade success - checklist empty", func(t *testing.T) {
		trade := &model.TaskTrade{
			ID:                 "trade-1",
			TaskExecutionID:    "exec-1",
			ProposedAssigneeID: "user-jenna",
			Status:             "PENDING",
		}
		exec := &model.TaskExecution{
			ID:             "exec-1",
			Status:         "TRADE_PENDING",
			ChecklistState: model.JSONB("[]"),
		}
		var updatedExec *model.TaskExecution
		var updatedTrade *model.TaskTrade

		mockRepo := &MockTaskExecutionRepository{
			FindTradeByIDFunc: func(ctx context.Context, id string) (*model.TaskTrade, error) {
				return trade, nil
			},
			FindByIDFunc: func(ctx context.Context, id string) (*model.TaskExecution, error) {
				return exec, nil
			},
			UpdateFunc: func(ctx context.Context, e *model.TaskExecution) error {
				updatedExec = e
				return nil
			},
			UpdateTradeFunc: func(ctx context.Context, t *model.TaskTrade) error {
				updatedTrade = t
				return nil
			},
		}

		svc := NewTaskService(mockRepo, nil)
		err := svc.AcceptTrade(context.Background(), "trade-1", "user-jenna")
		assert.NoError(t, err)

		assert.NotNil(t, updatedExec)
		assert.Equal(t, "user-jenna", *updatedExec.AssigneeID)
		assert.Equal(t, "PENDING", updatedExec.Status) // empty checklist resets status back to PENDING!

		assert.NotNil(t, updatedTrade)
		assert.Equal(t, "APPROVED", updatedTrade.Status)
	})

	t.Run("accept trade success - checklist partially completed", func(t *testing.T) {
		trade := &model.TaskTrade{
			ID:                 "trade-1",
			TaskExecutionID:    "exec-1",
			ProposedAssigneeID: "user-jenna",
			Status:             "PENDING",
		}
		exec := &model.TaskExecution{
			ID:             "exec-1",
			Status:         "TRADE_PENDING",
			ChecklistState: model.JSONB(`[{"step": 1, "action": "Step 1 Action", "completed": true}, {"step": 2, "action": "Step 2 Action", "completed": false}]`),
		}
		var updatedExec *model.TaskExecution

		mockRepo := &MockTaskExecutionRepository{
			FindTradeByIDFunc: func(ctx context.Context, id string) (*model.TaskTrade, error) {
				return trade, nil
			},
			FindByIDFunc: func(ctx context.Context, id string) (*model.TaskExecution, error) {
				return exec, nil
			},
			UpdateFunc: func(ctx context.Context, e *model.TaskExecution) error {
				updatedExec = e
				return nil
			},
			UpdateTradeFunc: func(ctx context.Context, t *model.TaskTrade) error {
				return nil
			},
		}

		svc := NewTaskService(mockRepo, nil)
		err := svc.AcceptTrade(context.Background(), "trade-1", "user-jenna")
		assert.NoError(t, err)

		assert.NotNil(t, updatedExec)
		assert.Equal(t, "user-jenna", *updatedExec.AssigneeID)
		assert.Equal(t, "IN_PROGRESS", updatedExec.Status) // has a completed step, so resets back to IN_PROGRESS!
	})
}

func TestTaskService_RejectTrade(t *testing.T) {
	t.Run("unauthorized proposed assignee cannot reject", func(t *testing.T) {
		mockRepo := &MockTaskExecutionRepository{
			FindTradeByIDFunc: func(ctx context.Context, id string) (*model.TaskTrade, error) {
				return &model.TaskTrade{
					ID:                 "trade-1",
					ProposedAssigneeID: "user-jenna",
					Status:             "PENDING",
				}, nil
			},
		}
		svc := NewTaskService(mockRepo, nil)
		err := svc.RejectTrade(context.Background(), "trade-1", "user-ryan")
		assert.Error(t, err)
		assert.Equal(t, "only the proposed colleague can reject this task trade", err.Error())
	})

	t.Run("reject trade success", func(t *testing.T) {
		trade := &model.TaskTrade{
			ID:                 "trade-1",
			TaskExecutionID:    "exec-1",
			InitiatorID:        "user-ryan",
			ProposedAssigneeID: "user-jenna",
			Status:             "PENDING",
		}
		// Initiated by Ryan, who already had it assigned
		exec := &model.TaskExecution{
			ID:             "exec-1",
			AssigneeID:     nil, // let's mock empty or matching ryan
			Status:         "TRADE_PENDING",
			ChecklistState: model.JSONB("[]"),
		}
		var updatedExec *model.TaskExecution
		var updatedTrade *model.TaskTrade

		mockRepo := &MockTaskExecutionRepository{
			FindTradeByIDFunc: func(ctx context.Context, id string) (*model.TaskTrade, error) {
				return trade, nil
			},
			FindByIDFunc: func(ctx context.Context, id string) (*model.TaskExecution, error) {
				return exec, nil
			},
			UpdateFunc: func(ctx context.Context, e *model.TaskExecution) error {
				updatedExec = e
				return nil
			},
			UpdateTradeFunc: func(ctx context.Context, t *model.TaskTrade) error {
				updatedTrade = t
				return nil
			},
		}

		svc := NewTaskService(mockRepo, nil)
		err := svc.RejectTrade(context.Background(), "trade-1", "user-jenna")
		assert.NoError(t, err)

		assert.NotNil(t, updatedExec)
		assert.Equal(t, "PENDING", updatedExec.Status) // task goes back to PENDING and assignee remains unchanged!

		assert.NotNil(t, updatedTrade)
		assert.Equal(t, "REJECTED", updatedTrade.Status)
	})
}
