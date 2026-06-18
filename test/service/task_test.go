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

package service_test

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/rmcguinness/gemini_task_engine/pkg/model"
	"github.com/rmcguinness/gemini_task_engine/pkg/persistence"
	"github.com/rmcguinness/gemini_task_engine/pkg/service"
	"github.com/stretchr/testify/assert"
)

type MockTaskExecutionRepository struct {
	persistence.TaskExecutionRepository
	CreateFunc                   func(ctx context.Context, e *model.TaskExecution) error
	FindByIDFunc                 func(ctx context.Context, id string) (*model.TaskExecution, error)
	UpdateFunc                   func(ctx context.Context, e *model.TaskExecution) error
	CreateTradeFunc              func(ctx context.Context, t *model.TaskTrade) error
	FindTradeByIDFunc            func(ctx context.Context, id string) (*model.TaskTrade, error)
	UpdateTradeFunc              func(ctx context.Context, t *model.TaskTrade) error
	FindPendingTradesForUserFunc func(ctx context.Context, userID string) ([]*model.TaskTrade, error)
	ListFunc                     func(ctx context.Context) ([]*model.TaskExecution, error)
	ListRangeFunc                func(ctx context.Context, offset, limit int) ([]*model.TaskExecution, error)
	DeleteFunc                   func(ctx context.Context, id string) error
	GetQueueFunc                 func(ctx context.Context, siteID string) ([]*model.TaskExecution, error)
	GetOrgTasksFunc              func(ctx context.Context, orgID string) ([]*model.TaskExecution, error)
	GetUserSiteTasksFunc         func(ctx context.Context, siteID, userID string) ([]*model.TaskExecution, error)
	FindPendingTradeByExecutionFunc func(ctx context.Context, executionID string) (*model.TaskTrade, error)
	GetSiteIDForExecutionFunc    func(ctx context.Context, execID string) (string, error)
}

func (m *MockTaskExecutionRepository) Create(ctx context.Context, e *model.TaskExecution) error {
	if m.CreateFunc != nil {
		return m.CreateFunc(ctx, e)
	}
	return nil
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

func (m *MockTaskExecutionRepository) List(ctx context.Context) ([]*model.TaskExecution, error) {
	if m.ListFunc != nil {
		return m.ListFunc(ctx)
	}
	return nil, nil
}

func (m *MockTaskExecutionRepository) ListRange(ctx context.Context, offset, limit int) ([]*model.TaskExecution, error) {
	if m.ListRangeFunc != nil {
		return m.ListRangeFunc(ctx, offset, limit)
	}
	return nil, nil
}

func (m *MockTaskExecutionRepository) Delete(ctx context.Context, id string) error {
	if m.DeleteFunc != nil {
		return m.DeleteFunc(ctx, id)
	}
	return nil
}

func (m *MockTaskExecutionRepository) GetQueue(ctx context.Context, siteID string) ([]*model.TaskExecution, error) {
	if m.GetQueueFunc != nil {
		return m.GetQueueFunc(ctx, siteID)
	}
	return nil, nil
}

func (m *MockTaskExecutionRepository) GetOrgTasks(ctx context.Context, orgID string) ([]*model.TaskExecution, error) {
	if m.GetOrgTasksFunc != nil {
		return m.GetOrgTasksFunc(ctx, orgID)
	}
	return nil, nil
}

func (m *MockTaskExecutionRepository) GetUserSiteTasks(ctx context.Context, siteID, userID string) ([]*model.TaskExecution, error) {
	if m.GetUserSiteTasksFunc != nil {
		return m.GetUserSiteTasksFunc(ctx, siteID, userID)
	}
	return nil, nil
}

func (m *MockTaskExecutionRepository) FindPendingTradeByExecution(ctx context.Context, executionID string) (*model.TaskTrade, error) {
	if m.FindPendingTradeByExecutionFunc != nil {
		return m.FindPendingTradeByExecutionFunc(ctx, executionID)
	}
	return nil, nil
}

func (m *MockTaskExecutionRepository) GetSiteIDForExecution(ctx context.Context, execID string) (string, error) {
	if m.GetSiteIDForExecutionFunc != nil {
		return m.GetSiteIDForExecutionFunc(ctx, execID)
	}
	return "", nil
}

type MockSiteRepository struct {
	persistence.SiteRepository
	FindByIDFunc         func(ctx context.Context, id string) (*model.Site, error)
	ListFunc             func(ctx context.Context) ([]*model.Site, error)
	FindLocationByIDFunc func(ctx context.Context, id string) (*model.Location, error)
	FindAssetByIDFunc    func(ctx context.Context, id string) (*model.Asset, error)
}

func (m *MockSiteRepository) FindByID(ctx context.Context, id string) (*model.Site, error) {
	if m.FindByIDFunc != nil {
		return m.FindByIDFunc(ctx, id)
	}
	return nil, nil
}

func (m *MockSiteRepository) List(ctx context.Context) ([]*model.Site, error) {
	if m.ListFunc != nil {
		return m.ListFunc(ctx)
	}
	return nil, nil
}

func (m *MockSiteRepository) FindLocationByID(ctx context.Context, id string) (*model.Location, error) {
	if m.FindLocationByIDFunc != nil {
		return m.FindLocationByIDFunc(ctx, id)
	}
	return nil, nil
}

func (m *MockSiteRepository) FindAssetByID(ctx context.Context, id string) (*model.Asset, error) {
	if m.FindAssetByIDFunc != nil {
		return m.FindAssetByIDFunc(ctx, id)
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
		svc := service.NewTaskService(mockRepo, nil)
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
		svc := service.NewTaskService(mockRepo, nil)
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
		svc := service.NewTaskService(mockRepo, nil)
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
		svc := service.NewTaskService(mockRepo, nil)
		err := svc.ProposeTrade(context.Background(), "exec-1", "user-jenna", "user-ryan")
		assert.NoError(t, err)

		assert.NotNil(t, createdTrade)
		assert.Equal(t, "exec-1", createdTrade.TaskExecutionID)
		assert.Equal(t, "user-ryan", createdTrade.InitiatorID)
		assert.Equal(t, stringPtr("user-jenna"), createdTrade.ProposedAssigneeID)
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
		svc := service.NewTaskService(mockRepo, nil)
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
		svc := service.NewTaskService(mockRepo, nil)
		err := svc.AcceptTrade(context.Background(), "trade-1", "user-jenna")
		assert.Error(t, err)
		assert.Equal(t, "trade request is not pending", err.Error())
	})

	t.Run("unauthorized proposed assignee cannot accept", func(t *testing.T) {
		mockRepo := &MockTaskExecutionRepository{
			FindTradeByIDFunc: func(ctx context.Context, id string) (*model.TaskTrade, error) {
				return &model.TaskTrade{
					ID:                 "trade-1",
					ProposedAssigneeID: stringPtr("user-jenna"),
					Status:             "PENDING",
				}, nil
			},
		}
		svc := service.NewTaskService(mockRepo, nil)
		// Ryan tries to accept a trade targeted at Jenna! (Violates Principle of Least Privilege!)
		err := svc.AcceptTrade(context.Background(), "trade-1", "user-ryan")
		assert.Error(t, err)
		assert.Equal(t, "only the proposed colleague can accept this task trade", err.Error())
	})

	t.Run("accept trade success - checklist empty", func(t *testing.T) {
		trade := &model.TaskTrade{
			ID:                 "trade-1",
			TaskExecutionID:    "exec-1",
			ProposedAssigneeID: stringPtr("user-jenna"),
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

		svc := service.NewTaskService(mockRepo, nil)
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
			ProposedAssigneeID: stringPtr("user-jenna"),
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

		svc := service.NewTaskService(mockRepo, nil)
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
					ProposedAssigneeID: stringPtr("user-jenna"),
					Status:             "PENDING",
				}, nil
			},
		}
		svc := service.NewTaskService(mockRepo, nil)
		err := svc.RejectTrade(context.Background(), "trade-1", "user-ryan")
		assert.Error(t, err)
		assert.Equal(t, "only the proposed colleague can reject this task trade", err.Error())
	})

	t.Run("reject trade success", func(t *testing.T) {
		trade := &model.TaskTrade{
			ID:                 "trade-1",
			TaskExecutionID:    "exec-1",
			InitiatorID:        "user-ryan",
			ProposedAssigneeID: stringPtr("user-jenna"),
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

		svc := service.NewTaskService(mockRepo, nil)
		err := svc.RejectTrade(context.Background(), "trade-1", "user-jenna")
		assert.NoError(t, err)

		assert.NotNil(t, updatedExec)
		assert.Equal(t, "PENDING", updatedExec.Status) // task goes back to PENDING and assignee remains unchanged!

		assert.NotNil(t, updatedTrade)
		assert.Equal(t, "REJECTED", updatedTrade.Status)
	})

	t.Run("reject trade trade not found", func(t *testing.T) {
		mockRepo := &MockTaskExecutionRepository{
			FindTradeByIDFunc: func(ctx context.Context, id string) (*model.TaskTrade, error) {
				return nil, errors.New("trade not found")
			},
		}
		svc := service.NewTaskService(mockRepo, nil)
		err := svc.RejectTrade(context.Background(), "trade-1", "user-jenna")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to find trade")
	})

	t.Run("reject trade execution not found", func(t *testing.T) {
		trade := &model.TaskTrade{
			ID:                 "trade-1",
			TaskExecutionID:    "exec-1",
			ProposedAssigneeID: stringPtr("user-jenna"),
			Status:             "PENDING",
		}
		mockRepo := &MockTaskExecutionRepository{
			FindTradeByIDFunc: func(ctx context.Context, id string) (*model.TaskTrade, error) {
				return trade, nil
			},
			FindByIDFunc: func(ctx context.Context, id string) (*model.TaskExecution, error) {
				return nil, errors.New("exec not found")
			},
		}
		svc := service.NewTaskService(mockRepo, nil)
		err := svc.RejectTrade(context.Background(), "trade-1", "user-jenna")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to find execution for trade")
	})

	t.Run("reject trade not pending fails", func(t *testing.T) {
		trade := &model.TaskTrade{
			ID:                 "trade-1",
			ProposedAssigneeID: stringPtr("user-jenna"),
			Status:             "APPROVED",
		}
		mockRepo := &MockTaskExecutionRepository{
			FindTradeByIDFunc: func(ctx context.Context, id string) (*model.TaskTrade, error) {
				return trade, nil
			},
		}
		svc := service.NewTaskService(mockRepo, nil)
		err := svc.RejectTrade(context.Background(), "trade-1", "user-jenna")
		assert.Error(t, err)
		assert.Equal(t, "trade request is not pending", err.Error())
	})

	t.Run("reject trade with completed steps resets to IN_PROGRESS", func(t *testing.T) {
		trade := &model.TaskTrade{
			ID:                 "trade-1",
			TaskExecutionID:    "exec-1",
			InitiatorID:        "user-ryan",
			ProposedAssigneeID: stringPtr("user-jenna"),
			Status:             "PENDING",
		}
		exec := &model.TaskExecution{
			ID:             "exec-1",
			Status:         "TRADE_PENDING",
			ChecklistState: model.JSONB(`[{"step_id":"step-1","completed":true}]`),
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
		svc := service.NewTaskService(mockRepo, nil)
		err := svc.RejectTrade(context.Background(), "trade-1", "user-jenna")
		assert.NoError(t, err)
		assert.Equal(t, "IN_PROGRESS", updatedExec.Status)
	})
}

func TestTaskService_CRUD(t *testing.T) {
	t.Run("GetTaskExecutionByID success", func(t *testing.T) {
		expected := &model.TaskExecution{ID: "exec-1", Status: "PENDING"}
		mockRepo := &MockTaskExecutionRepository{
			FindByIDFunc: func(ctx context.Context, id string) (*model.TaskExecution, error) {
				assert.Equal(t, "exec-1", id)
				return expected, nil
			},
		}
		svc := service.NewTaskService(mockRepo, nil)
		res, err := svc.GetTaskExecutionByID(context.Background(), "exec-1")
		assert.NoError(t, err)
		assert.Equal(t, expected, res)
	})

	t.Run("ListTaskExecutions success", func(t *testing.T) {
		expected := []*model.TaskExecution{
			{ID: "exec-1", Status: "PENDING"},
			{ID: "exec-2", Status: "COMPLETED"},
		}
		mockRepo := &MockTaskExecutionRepository{
			ListFunc: func(ctx context.Context) ([]*model.TaskExecution, error) {
				return expected, nil
			},
		}
		svc := service.NewTaskService(mockRepo, nil)
		res, err := svc.ListTaskExecutions(context.Background())
		assert.NoError(t, err)
		assert.Equal(t, expected, res)
	})

	t.Run("ListTaskExecutionsRange success", func(t *testing.T) {
		expected := []*model.TaskExecution{
			{ID: "exec-2", Status: "COMPLETED"},
		}
		mockRepo := &MockTaskExecutionRepository{
			ListRangeFunc: func(ctx context.Context, offset, limit int) ([]*model.TaskExecution, error) {
				assert.Equal(t, 1, offset)
				assert.Equal(t, 10, limit)
				return expected, nil
			},
		}
		svc := service.NewTaskService(mockRepo, nil)
		res, err := svc.ListTaskExecutionsRange(context.Background(), 1, 10)
		assert.NoError(t, err)
		assert.Equal(t, expected, res)
	})

	t.Run("DeleteTaskExecution success", func(t *testing.T) {
		called := false
		mockRepo := &MockTaskExecutionRepository{
			DeleteFunc: func(ctx context.Context, id string) error {
				assert.Equal(t, "exec-1", id)
				called = true
				return nil
			},
		}
		svc := service.NewTaskService(mockRepo, nil)
		err := svc.DeleteTaskExecution(context.Background(), "exec-1")
		assert.NoError(t, err)
		assert.True(t, called)
	})
}

func TestTaskService_GettersAndDelegations(t *testing.T) {
	t.Run("GetQueue success", func(t *testing.T) {
		expected := []*model.TaskExecution{{ID: "exec-1"}}
		mockRepo := &MockTaskExecutionRepository{
			GetQueueFunc: func(ctx context.Context, siteID string) ([]*model.TaskExecution, error) {
				assert.Equal(t, "site-123", siteID)
				return expected, nil
			},
		}
		svc := service.NewTaskService(mockRepo, nil)
		res, err := svc.GetQueue(context.Background(), "site-123")
		assert.NoError(t, err)
		assert.Equal(t, expected, res)
	})

	t.Run("ListActiveSites success", func(t *testing.T) {
		expected := []*model.Site{{ID: "site-123"}}
		mockSite := &MockSiteRepository{
			ListFunc: func(ctx context.Context) ([]*model.Site, error) {
				return expected, nil
			},
		}
		svc := service.NewTaskService(nil, mockSite)
		res, err := svc.ListActiveSites(context.Background())
		assert.NoError(t, err)
		assert.Equal(t, expected, res)
	})

	t.Run("GetOrgTasks success", func(t *testing.T) {
		expected := []*model.TaskExecution{{ID: "exec-1"}}
		mockRepo := &MockTaskExecutionRepository{
			GetOrgTasksFunc: func(ctx context.Context, orgID string) ([]*model.TaskExecution, error) {
				assert.Equal(t, "org-123", orgID)
				return expected, nil
			},
		}
		svc := service.NewTaskService(mockRepo, nil)
		res, err := svc.GetOrgTasks(context.Background(), "org-123")
		assert.NoError(t, err)
		assert.Equal(t, expected, res)
	})

	t.Run("GetUserSiteTasks success", func(t *testing.T) {
		expected := []*model.TaskExecution{{ID: "exec-1"}}
		mockRepo := &MockTaskExecutionRepository{
			GetUserSiteTasksFunc: func(ctx context.Context, siteID, userID string) ([]*model.TaskExecution, error) {
				assert.Equal(t, "site-123", siteID)
				assert.Equal(t, "user-123", userID)
				return expected, nil
			},
		}
		svc := service.NewTaskService(mockRepo, nil)
		res, err := svc.GetUserSiteTasks(context.Background(), "site-123", "user-123")
		assert.NoError(t, err)
		assert.Equal(t, expected, res)
	})

	t.Run("GetSiteIDForExecution success", func(t *testing.T) {
		mockRepo := &MockTaskExecutionRepository{
			GetSiteIDForExecutionFunc: func(ctx context.Context, execID string) (string, error) {
				assert.Equal(t, "exec-123", execID)
				return "site-123", nil
			},
		}
		svc := service.NewTaskService(mockRepo, nil)
		res, err := svc.GetSiteIDForExecution(context.Background(), "exec-123")
		assert.NoError(t, err)
		assert.Equal(t, "site-123", res)
	})

	t.Run("GetLocationByID success", func(t *testing.T) {
		expected := &model.Location{ID: "loc-123"}
		mockSite := &MockSiteRepository{
			FindLocationByIDFunc: func(ctx context.Context, id string) (*model.Location, error) {
				assert.Equal(t, "loc-123", id)
				return expected, nil
			},
		}
		svc := service.NewTaskService(nil, mockSite)
		res, err := svc.GetLocationByID(context.Background(), "loc-123")
		assert.NoError(t, err)
		assert.Equal(t, expected, res)
	})

	t.Run("GetSiteLocations success", func(t *testing.T) {
		site := &model.Site{
			ID: "site-123",
			Locations: []model.Location{
				{ID: "loc-1"},
				{ID: "loc-2"},
			},
		}
		mockSite := &MockSiteRepository{
			FindByIDFunc: func(ctx context.Context, id string) (*model.Site, error) {
				assert.Equal(t, "site-123", id)
				return site, nil
			},
		}
		svc := service.NewTaskService(nil, mockSite)
		res, err := svc.GetSiteLocations(context.Background(), "site-123")
		assert.NoError(t, err)
		assert.Len(t, res, 2)
		assert.Equal(t, "loc-1", res[0].ID)
		assert.Equal(t, "loc-2", res[1].ID)
	})

	t.Run("ListPendingTrades success", func(t *testing.T) {
		expected := []*model.TaskTrade{{ID: "trade-1"}}
		mockRepo := &MockTaskExecutionRepository{
			FindPendingTradesForUserFunc: func(ctx context.Context, userID string) ([]*model.TaskTrade, error) {
				assert.Equal(t, "user-123", userID)
				return expected, nil
			},
		}
		svc := service.NewTaskService(mockRepo, nil)
		res, err := svc.ListPendingTrades(context.Background(), "user-123")
		assert.NoError(t, err)
		assert.Equal(t, expected, res)
	})
}

func TestTaskService_UpdateStatus(t *testing.T) {
	t.Run("transition PENDING to IN_PROGRESS", func(t *testing.T) {
		exec := &model.TaskExecution{ID: "exec-1", Status: "PENDING"}
		var updated *model.TaskExecution
		mockRepo := &MockTaskExecutionRepository{
			FindByIDFunc: func(ctx context.Context, id string) (*model.TaskExecution, error) {
				return exec, nil
			},
			UpdateFunc: func(ctx context.Context, e *model.TaskExecution) error {
				updated = e
				return nil
			},
		}
		svc := service.NewTaskService(mockRepo, nil)
		err := svc.UpdateStatus(context.Background(), "exec-1", "IN_PROGRESS", "", "user-123")
		assert.NoError(t, err)
		assert.Equal(t, "IN_PROGRESS", updated.Status)
		assert.NotNil(t, updated.StartedAt)
	})

	t.Run("transition PAUSED to IN_PROGRESS", func(t *testing.T) {
		now := time.Now()
		pausedAt := now.Add(-10 * time.Second)
		exec := &model.TaskExecution{
			ID:                 "exec-1",
			Status:             "PAUSED",
			PausedAt:           &pausedAt,
			TotalPausedSeconds: 5,
		}
		var updated *model.TaskExecution
		mockRepo := &MockTaskExecutionRepository{
			FindByIDFunc: func(ctx context.Context, id string) (*model.TaskExecution, error) {
				return exec, nil
			},
			UpdateFunc: func(ctx context.Context, e *model.TaskExecution) error {
				updated = e
				return nil
			},
		}
		svc := service.NewTaskService(mockRepo, nil)
		err := svc.UpdateStatus(context.Background(), "exec-1", "IN_PROGRESS", "", "user-123")
		assert.NoError(t, err)
		assert.Equal(t, "IN_PROGRESS", updated.Status)
		assert.Nil(t, updated.PausedAt)
		assert.GreaterOrEqual(t, updated.TotalPausedSeconds, 15)
	})

	t.Run("transition IN_PROGRESS to PAUSED", func(t *testing.T) {
		exec := &model.TaskExecution{ID: "exec-1", Status: "IN_PROGRESS"}
		var updated *model.TaskExecution
		mockRepo := &MockTaskExecutionRepository{
			FindByIDFunc: func(ctx context.Context, id string) (*model.TaskExecution, error) {
				return exec, nil
			},
			UpdateFunc: func(ctx context.Context, e *model.TaskExecution) error {
				updated = e
				return nil
			},
		}
		svc := service.NewTaskService(mockRepo, nil)
		err := svc.UpdateStatus(context.Background(), "exec-1", "PAUSED", "", "user-123")
		assert.NoError(t, err)
		assert.Equal(t, "PAUSED", updated.Status)
		assert.NotNil(t, updated.PausedAt)
	})

	t.Run("transition to COMPLETED", func(t *testing.T) {
		exec := &model.TaskExecution{ID: "exec-1", Status: "IN_PROGRESS"}
		var updated *model.TaskExecution
		mockRepo := &MockTaskExecutionRepository{
			FindByIDFunc: func(ctx context.Context, id string) (*model.TaskExecution, error) {
				return exec, nil
			},
			UpdateFunc: func(ctx context.Context, e *model.TaskExecution) error {
				updated = e
				return nil
			},
		}
		svc := service.NewTaskService(mockRepo, nil)
		err := svc.UpdateStatus(context.Background(), "exec-1", "COMPLETED", "", "user-123")
		assert.NoError(t, err)
		assert.Equal(t, "COMPLETED", updated.Status)
		assert.NotNil(t, updated.CompletedAt)
	})

	t.Run("checklist update with valid JSON", func(t *testing.T) {
		exec := &model.TaskExecution{ID: "exec-1", Status: "PENDING"}
		var updated *model.TaskExecution
		mockRepo := &MockTaskExecutionRepository{
			FindByIDFunc: func(ctx context.Context, id string) (*model.TaskExecution, error) {
				return exec, nil
			},
			UpdateFunc: func(ctx context.Context, e *model.TaskExecution) error {
				updated = e
				return nil
			},
		}
		svc := service.NewTaskService(mockRepo, nil)
		err := svc.UpdateStatus(context.Background(), "exec-1", "PENDING", `[{"step_order":1,"completed":true}]`, "user-123")
		assert.NoError(t, err)
		assert.JSONEq(t, `[{"step_order":1,"completed":true}]`, string(updated.ChecklistState))
	})

	t.Run("checklist delta update START", func(t *testing.T) {
		exec := &model.TaskExecution{
			ID:             "exec-1",
			Status:         "PENDING",
			ChecklistState: model.JSONB(`[{"step":1,"status":"PENDING","completed":false}]`),
		}
		var updated *model.TaskExecution
		mockRepo := &MockTaskExecutionRepository{
			FindByIDFunc: func(ctx context.Context, id string) (*model.TaskExecution, error) {
				return exec, nil
			},
			UpdateFunc: func(ctx context.Context, e *model.TaskExecution) error {
				updated = e
				return nil
			},
		}
		svc := service.NewTaskService(mockRepo, nil)
		err := svc.UpdateStatus(context.Background(), "exec-1", "PENDING", `{"step":1,"action":"START"}`, "user-123")
		assert.NoError(t, err)

		var list []map[string]interface{}
		err = json.Unmarshal(updated.ChecklistState, &list)
		assert.NoError(t, err)
		assert.Equal(t, "IN_PROGRESS", list[0]["status"])
		assert.NotEmpty(t, list[0]["started_at"])
	})

	t.Run("checklist delta update PAUSE and RESUME", func(t *testing.T) {
		exec := &model.TaskExecution{
			ID:             "exec-1",
			Status:         "PENDING",
			ChecklistState: model.JSONB(`[{"step":1,"status":"IN_PROGRESS","completed":false}]`),
		}
		var updated *model.TaskExecution
		mockRepo := &MockTaskExecutionRepository{
			FindByIDFunc: func(ctx context.Context, id string) (*model.TaskExecution, error) {
				return exec, nil
			},
			UpdateFunc: func(ctx context.Context, e *model.TaskExecution) error {
				updated = e
				return nil
			},
		}
		svc := service.NewTaskService(mockRepo, nil)
		
		err := svc.UpdateStatus(context.Background(), "exec-1", "PENDING", `{"step":1,"action":"PAUSE"}`, "user-123")
		assert.NoError(t, err)

		var list []map[string]interface{}
		err = json.Unmarshal(updated.ChecklistState, &list)
		assert.NoError(t, err)
		assert.Equal(t, "PAUSED", list[0]["status"])
		assert.NotEmpty(t, list[0]["paused_at"])

		pausedAtStr := time.Now().Add(-10 * time.Second).Format(time.RFC3339)
		exec.ChecklistState = model.JSONB(`[{"step":1,"status":"PAUSED","paused_at":"` + pausedAtStr + `","total_paused_seconds":5}]`)
		err = svc.UpdateStatus(context.Background(), "exec-1", "PENDING", `{"step":1,"action":"RESUME"}`, "user-123")
		assert.NoError(t, err)

		err = json.Unmarshal(updated.ChecklistState, &list)
		assert.NoError(t, err)
		assert.Equal(t, "IN_PROGRESS", list[0]["status"])
		assert.Nil(t, list[0]["paused_at"])
		assert.GreaterOrEqual(t, int(list[0]["total_paused_seconds"].(float64)), 14)
	})

	t.Run("checklist delta update COMPLETE", func(t *testing.T) {
		startedAtStr := time.Now().Add(-70 * time.Second).Format(time.RFC3339)
		exec := &model.TaskExecution{
			ID:             "exec-1",
			Status:         "PENDING",
			ChecklistState: model.JSONB(`[{"step":1,"status":"IN_PROGRESS","started_at":"` + startedAtStr + `","total_paused_seconds":5,"slo_seconds":60}]`),
		}
		var updated *model.TaskExecution
		mockRepo := &MockTaskExecutionRepository{
			FindByIDFunc: func(ctx context.Context, id string) (*model.TaskExecution, error) {
				return exec, nil
			},
			UpdateFunc: func(ctx context.Context, e *model.TaskExecution) error {
				updated = e
				return nil
			},
		}
		svc := service.NewTaskService(mockRepo, nil)
		err := svc.UpdateStatus(context.Background(), "exec-1", "PENDING", `{"step":1,"action":"COMPLETE"}`, "user-123")
		assert.NoError(t, err)

		var list []map[string]interface{}
		err = json.Unmarshal(updated.ChecklistState, &list)
		assert.NoError(t, err)
		assert.Equal(t, "COMPLETED", list[0]["status"])
		assert.True(t, list[0]["completed"].(bool))
		assert.Equal(t, "user-123", list[0]["completed_by_id"])
		assert.GreaterOrEqual(t, int(list[0]["slo_delta_seconds"].(float64)), 4)
	})
}

func TestTaskService_OverrideAssetConstraint(t *testing.T) {
	t.Run("override asset constraint success", func(t *testing.T) {
		exec := &model.TaskExecution{ID: "exec-1", OverrideFlags: model.JSONB(`{"existing-asset":{"justification":"old"}}`)}
		asset := &model.Asset{ID: "asset-123", Name: "Cool Cooler"}
		var updated *model.TaskExecution

		mockRepo := &MockTaskExecutionRepository{
			FindByIDFunc: func(ctx context.Context, id string) (*model.TaskExecution, error) {
				return exec, nil
			},
			UpdateFunc: func(ctx context.Context, e *model.TaskExecution) error {
				updated = e
				return nil
			},
		}
		mockSite := &MockSiteRepository{
			FindAssetByIDFunc: func(ctx context.Context, id string) (*model.Asset, error) {
				assert.Equal(t, "asset-123", id)
				return asset, nil
			},
		}
		svc := service.NewTaskService(mockRepo, mockSite)
		err := svc.OverrideAssetConstraint(context.Background(), "exec-1", "asset-123", "Too hot today", "user-123")
		assert.NoError(t, err)

		var flags map[string]interface{}
		err = json.Unmarshal(updated.OverrideFlags, &flags)
		assert.NoError(t, err)

		assert.Contains(t, flags, "existing-asset")
		assert.Contains(t, flags, "asset-123")
		newOverride := flags["asset-123"].(map[string]interface{})
		assert.Equal(t, "Cool Cooler", newOverride["asset_name"])
		assert.Equal(t, "Too hot today", newOverride["justification"])
		assert.Equal(t, "user-123", newOverride["overridden_by"])
		assert.NotEmpty(t, newOverride["timestamp"])
	})

	t.Run("override asset constraint execution not found", func(t *testing.T) {
		mockRepo := &MockTaskExecutionRepository{
			FindByIDFunc: func(ctx context.Context, id string) (*model.TaskExecution, error) {
				return nil, errors.New("exec not found")
			},
		}
		svc := service.NewTaskService(mockRepo, nil)
		err := svc.OverrideAssetConstraint(context.Background(), "exec-1", "asset-123", "Too hot today", "user-123")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to find task execution")
	})

	t.Run("override asset constraint asset not found", func(t *testing.T) {
		exec := &model.TaskExecution{ID: "exec-1"}
		mockRepo := &MockTaskExecutionRepository{
			FindByIDFunc: func(ctx context.Context, id string) (*model.TaskExecution, error) {
				return exec, nil
			},
		}
		mockSite := &MockSiteRepository{
			FindAssetByIDFunc: func(ctx context.Context, id string) (*model.Asset, error) {
				return nil, errors.New("asset not found")
			},
		}
		svc := service.NewTaskService(mockRepo, mockSite)
		err := svc.OverrideAssetConstraint(context.Background(), "exec-1", "asset-123", "Too hot today", "user-123")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to find asset")
	})
}

func TestTaskService_ApproveTrade(t *testing.T) {
	t.Run("approve trade success", func(t *testing.T) {
		trade := &model.TaskTrade{
			ID:                 "trade-1",
			TaskExecutionID:    "exec-1",
			ProposedAssigneeID: stringPtr("user-jenna"),
			Status:             "PENDING",
		}
		exec := &model.TaskExecution{
			ID:         "exec-1",
			AssigneeID: stringPtr("user-ryan"),
			Status:     "TRADE_PENDING",
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

		svc := service.NewTaskService(mockRepo, nil)
		err := svc.ApproveTrade(context.Background(), "trade-1", "supervisor-123")
		assert.NoError(t, err)

		assert.Equal(t, stringPtr("user-jenna"), updatedExec.AssigneeID)
		assert.Equal(t, "PENDING", updatedExec.Status)
		assert.Equal(t, "APPROVED", updatedTrade.Status)
	})

	t.Run("approve trade trade not found", func(t *testing.T) {
		mockRepo := &MockTaskExecutionRepository{
			FindTradeByIDFunc: func(ctx context.Context, id string) (*model.TaskTrade, error) {
				return nil, errors.New("trade not found")
			},
		}
		svc := service.NewTaskService(mockRepo, nil)
		err := svc.ApproveTrade(context.Background(), "trade-1", "supervisor-123")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to find trade")
	})

	t.Run("approve trade execution not found", func(t *testing.T) {
		trade := &model.TaskTrade{
			ID:                 "trade-1",
			TaskExecutionID:    "exec-1",
			Status:             "PENDING",
		}
		mockRepo := &MockTaskExecutionRepository{
			FindTradeByIDFunc: func(ctx context.Context, id string) (*model.TaskTrade, error) {
				return trade, nil
			},
			FindByIDFunc: func(ctx context.Context, id string) (*model.TaskExecution, error) {
				return nil, errors.New("exec not found")
			},
		}
		svc := service.NewTaskService(mockRepo, nil)
		err := svc.ApproveTrade(context.Background(), "trade-1", "supervisor-123")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to find execution for trade")
	})
}

func TestTaskService_ClaimTask(t *testing.T) {
	t.Run("claim standard open task success", func(t *testing.T) {
		exec := &model.TaskExecution{
			ID:     "exec-1",
			Status: "PENDING",
			Task:   model.Task{TargetRoleID: nil},
		}
		var updated *model.TaskExecution

		mockRepo := &MockTaskExecutionRepository{
			FindByIDFunc: func(ctx context.Context, id string) (*model.TaskExecution, error) {
				return exec, nil
			},
			UpdateFunc: func(ctx context.Context, e *model.TaskExecution) error {
				updated = e
				return nil
			},
		}
		svc := service.NewTaskService(mockRepo, nil)
		err := svc.ClaimTask(context.Background(), "exec-1", "user-123", nil)
		assert.NoError(t, err)
		assert.Equal(t, stringPtr("user-123"), updated.AssigneeID)
		assert.Equal(t, "PENDING", updated.Status)
	})

	t.Run("claim standard open task with role check success", func(t *testing.T) {
		roleID := "role-associate"
		exec := &model.TaskExecution{
			ID:     "exec-1",
			Status: "PENDING",
			Task:   model.Task{TargetRoleID: &roleID},
		}
		var updated *model.TaskExecution

		mockRepo := &MockTaskExecutionRepository{
			FindByIDFunc: func(ctx context.Context, id string) (*model.TaskExecution, error) {
				return exec, nil
			},
			UpdateFunc: func(ctx context.Context, e *model.TaskExecution) error {
				updated = e
				return nil
			},
		}
		svc := service.NewTaskService(mockRepo, nil)
		err := svc.ClaimTask(context.Background(), "exec-1", "user-123", []string{"some-other-role", "role-associate"})
		assert.NoError(t, err)
		assert.Equal(t, stringPtr("user-123"), updated.AssigneeID)
	})

	t.Run("claim standard open task role mismatch fails", func(t *testing.T) {
		roleID := "role-admin"
		exec := &model.TaskExecution{
			ID:     "exec-1",
			Status: "PENDING",
			Task:   model.Task{TargetRoleID: &roleID},
		}

		mockRepo := &MockTaskExecutionRepository{
			FindByIDFunc: func(ctx context.Context, id string) (*model.TaskExecution, error) {
				return exec, nil
			},
		}
		svc := service.NewTaskService(mockRepo, nil)
		err := svc.ClaimTask(context.Background(), "exec-1", "user-123", []string{"role-associate"})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "unauthorized")
	})

	t.Run("claim task trade success", func(t *testing.T) {
		trade := &model.TaskTrade{
			ID:                 "trade-1",
			TaskExecutionID:    "exec-1",
			InitiatorID:        "user-ryan",
			ProposedAssigneeID: stringPtr("user-jenna"),
			Status:             "PENDING",
		}
		exec := &model.TaskExecution{
			ID:     "exec-1",
			Status: "TRADE_PENDING",
			Task:   model.Task{TargetRoleID: nil},
		}
		var updatedExec *model.TaskExecution
		var updatedTrade *model.TaskTrade

		mockRepo := &MockTaskExecutionRepository{
			FindByIDFunc: func(ctx context.Context, id string) (*model.TaskExecution, error) {
				return exec, nil
			},
			FindPendingTradeByExecutionFunc: func(ctx context.Context, executionID string) (*model.TaskTrade, error) {
				assert.Equal(t, "exec-1", executionID)
				return trade, nil
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
		svc := service.NewTaskService(mockRepo, nil)
		err := svc.ClaimTask(context.Background(), "exec-1", "user-jenna", nil)
		assert.NoError(t, err)
		assert.Equal(t, stringPtr("user-jenna"), updatedExec.AssigneeID)
		assert.Equal(t, "APPROVED", updatedTrade.Status)
	})

	t.Run("claim task execution not found", func(t *testing.T) {
		mockRepo := &MockTaskExecutionRepository{
			FindByIDFunc: func(ctx context.Context, id string) (*model.TaskExecution, error) {
				return nil, errors.New("exec not found")
			},
		}
		svc := service.NewTaskService(mockRepo, nil)
		err := svc.ClaimTask(context.Background(), "exec-1", "user-123", nil)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to find task execution")
	})

	t.Run("claim task trade record not found", func(t *testing.T) {
		exec := &model.TaskExecution{
			ID:     "exec-1",
			Status: "TRADE_PENDING",
		}
		mockRepo := &MockTaskExecutionRepository{
			FindByIDFunc: func(ctx context.Context, id string) (*model.TaskExecution, error) {
				return exec, nil
			},
			FindPendingTradeByExecutionFunc: func(ctx context.Context, executionID string) (*model.TaskTrade, error) {
				return nil, errors.New("trade not found")
			},
		}
		svc := service.NewTaskService(mockRepo, nil)
		err := svc.ClaimTask(context.Background(), "exec-1", "user-jenna", nil)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to find pending trade record for this execution")
	})

	t.Run("claim completed task fails", func(t *testing.T) {
		exec := &model.TaskExecution{
			ID:     "exec-1",
			Status: "COMPLETED",
		}
		mockRepo := &MockTaskExecutionRepository{
			FindByIDFunc: func(ctx context.Context, id string) (*model.TaskExecution, error) {
				return exec, nil
			},
		}
		svc := service.NewTaskService(mockRepo, nil)
		err := svc.ClaimTask(context.Background(), "exec-1", "user-123", nil)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "task is not in a claimable state")
	})

	t.Run("claim task trade unauthorized fails", func(t *testing.T) {
		trade := &model.TaskTrade{
			ID:                 "trade-1",
			TaskExecutionID:    "exec-1",
			InitiatorID:        "user-ryan",
			ProposedAssigneeID: stringPtr("user-jenna"),
			Status:             "PENDING",
		}
		exec := &model.TaskExecution{
			ID:     "exec-1",
			Status: "TRADE_PENDING",
			Task:   model.Task{TargetRoleID: stringPtr("role-associate")},
		}
		mockRepo := &MockTaskExecutionRepository{
			FindByIDFunc: func(ctx context.Context, id string) (*model.TaskExecution, error) {
				return exec, nil
			},
			FindPendingTradeByExecutionFunc: func(ctx context.Context, executionID string) (*model.TaskTrade, error) {
				return trade, nil
			},
		}
		svc := service.NewTaskService(mockRepo, nil)
		err := svc.ClaimTask(context.Background(), "exec-1", "user-marcus", []string{"some-other-role"})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "unauthorized: you are not eligible to claim this task trade")
	})

	t.Run("claim task trade initiator claims back success", func(t *testing.T) {
		trade := &model.TaskTrade{
			ID:                 "trade-1",
			TaskExecutionID:    "exec-1",
			InitiatorID:        "user-ryan",
			ProposedAssigneeID: stringPtr("user-jenna"),
			Status:             "PENDING",
		}
		exec := &model.TaskExecution{
			ID:     "exec-1",
			Status: "TRADE_PENDING",
			Task:   model.Task{TargetRoleID: nil},
		}
		var updatedExec *model.TaskExecution
		var updatedTrade *model.TaskTrade

		mockRepo := &MockTaskExecutionRepository{
			FindByIDFunc: func(ctx context.Context, id string) (*model.TaskExecution, error) {
				return exec, nil
			},
			FindPendingTradeByExecutionFunc: func(ctx context.Context, executionID string) (*model.TaskTrade, error) {
				return trade, nil
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
		svc := service.NewTaskService(mockRepo, nil)
		err := svc.ClaimTask(context.Background(), "exec-1", "user-ryan", nil)
		assert.NoError(t, err)
		assert.Equal(t, stringPtr("user-ryan"), updatedExec.AssigneeID)
		assert.Equal(t, "APPROVED", updatedTrade.Status)
	})

	t.Run("claim task trade with completed steps resets to IN_PROGRESS", func(t *testing.T) {
		trade := &model.TaskTrade{
			ID:                 "trade-1",
			TaskExecutionID:    "exec-1",
			InitiatorID:        "user-ryan",
			ProposedAssigneeID: stringPtr("user-jenna"),
			Status:             "PENDING",
		}
		exec := &model.TaskExecution{
			ID:             "exec-1",
			Status:         "TRADE_PENDING",
			ChecklistState: model.JSONB(`[{"step_id":"step-1","completed":true}]`),
			Task:           model.Task{TargetRoleID: nil},
		}
		var updatedExec *model.TaskExecution

		mockRepo := &MockTaskExecutionRepository{
			FindByIDFunc: func(ctx context.Context, id string) (*model.TaskExecution, error) {
				return exec, nil
			},
			FindPendingTradeByExecutionFunc: func(ctx context.Context, executionID string) (*model.TaskTrade, error) {
				return trade, nil
			},
			UpdateFunc: func(ctx context.Context, e *model.TaskExecution) error {
				updatedExec = e
				return nil
			},
			UpdateTradeFunc: func(ctx context.Context, t *model.TaskTrade) error {
				return nil
			},
		}
		svc := service.NewTaskService(mockRepo, nil)
		err := svc.ClaimTask(context.Background(), "exec-1", "user-jenna", nil)
		assert.NoError(t, err)
		assert.Equal(t, "IN_PROGRESS", updatedExec.Status)
	})
}


