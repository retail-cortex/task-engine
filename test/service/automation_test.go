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
	"errors"
	"testing"

	"github.com/rmcguinness/gemini_task_engine/pkg/model"
	"github.com/rmcguinness/gemini_task_engine/pkg/persistence"
	"github.com/rmcguinness/gemini_task_engine/pkg/service"
	"github.com/stretchr/testify/assert"
)



// MockTaskRepository maps master template structures.
type MockTaskRepository struct {
	persistence.TaskRepository
	FindByIDFunc func(ctx context.Context, id string) (*model.Task, error)
	CreateFunc   func(ctx context.Context, task *model.Task) error
	ListFunc     func(ctx context.Context) ([]*model.Task, error)
}

func (m *MockTaskRepository) FindByID(ctx context.Context, id string) (*model.Task, error) {
	if m.FindByIDFunc != nil {
		return m.FindByIDFunc(ctx, id)
	}
	return nil, errors.New("not found")
}

func (m *MockTaskRepository) Create(ctx context.Context, task *model.Task) error {
	if m.CreateFunc != nil {
		return m.CreateFunc(ctx, task)
	}
	return nil
}

func (m *MockTaskRepository) List(ctx context.Context) ([]*model.Task, error) {
	if m.ListFunc != nil {
		return m.ListFunc(ctx)
	}
	return nil, nil
}



type MockSOPRepository struct {
	persistence.SOPRepository
	QuerySimilarityFunc func(ctx context.Context, embedding model.Float32Vector, limit int) ([]*model.SOPChunk, error)
}

func (m *MockSOPRepository) QuerySimilarity(ctx context.Context, embedding model.Float32Vector, limit int) ([]*model.SOPChunk, error) {
	if m.QuerySimilarityFunc != nil {
		return m.QuerySimilarityFunc(ctx, embedding, limit)
	}
	return nil, nil
}

type MockEmbeddingGenerator struct {
	service.EmbeddingGenerator
	GenerateEmbeddingsFunc func(ctx context.Context, text string) (model.Float32Vector, error)
}

func (m *MockEmbeddingGenerator) GenerateEmbeddings(ctx context.Context, text string) (model.Float32Vector, error) {
	if m.GenerateEmbeddingsFunc != nil {
		return m.GenerateEmbeddingsFunc(ctx, text)
	}
	return make(model.Float32Vector, 768), nil
}

func TestAutomationService_ProcessBatchEvent(t *testing.T) {
	t.Run("successfully materializes a batch opening task", func(t *testing.T) {
		eventInstanceID := "instance-batch-0000"

		// Task template find returns error first time, forcing auto-seed creation
		var templatesCreated []*model.Task
		mockTaskRepo := &MockTaskRepository{
			FindByIDFunc: func(ctx context.Context, id string) (*model.Task, error) {
				return nil, errors.New("template not found database gap")
			},
			CreateFunc: func(ctx context.Context, task *model.Task) error {
				assert.Equal(t, "d000fa44-0000-0000-0000-000000000000", task.ID)
				assert.Equal(t, "Register Terminal & Cash Opening Checkout Suite", task.Name)
				templatesCreated = append(templatesCreated, task)
				return nil
			},
		}

		var executionsCreated []*model.TaskExecution
		mockExecRepo := &MockTaskExecutionRepository{
			CreateFunc: func(ctx context.Context, e *model.TaskExecution) error {
				assert.Equal(t, "d000fa44-0000-0000-0000-000000000000", e.TaskTemplateID)
				assert.Equal(t, "STANDARD", e.ExecutionType)
				assert.Equal(t, "PENDING", e.Status)
				assert.Equal(t, 1, e.Priority)
				assert.Equal(t, eventInstanceID, e.EventInstanceID)
				executionsCreated = append(executionsCreated, e)
				return nil
			},
		}

		mockSiteRepo := &MockSiteRepository{}
		mockUserRepo := &MockUserRepository{}
		mockSopRepo := &MockSOPRepository{}
		mockEmbed := &MockEmbeddingGenerator{}

		svc := service.NewAutomationService(mockExecRepo, mockTaskRepo, mockSiteRepo, mockUserRepo, mockSopRepo, mockEmbed)
		execs, err := svc.ProcessBatchEvent(context.Background(), eventInstanceID)

		assert.NoError(t, err)
		assert.Len(t, execs, 1)
		assert.Len(t, templatesCreated, 1)
		assert.Len(t, executionsCreated, 1)
		assert.Equal(t, "d000fa44-0000-0000-0000-000000000000", execs[0].TaskTemplateID)
	})

	t.Run("empty eventInstanceID parameters returns error", func(t *testing.T) {
		mockTaskRepo := &MockTaskRepository{}
		mockExecRepo := &MockTaskExecutionRepository{}
		mockSiteRepo := &MockSiteRepository{}
		mockUserRepo := &MockUserRepository{}
		mockSopRepo := &MockSOPRepository{}
		mockEmbed := &MockEmbeddingGenerator{}

		svc := service.NewAutomationService(mockExecRepo, mockTaskRepo, mockSiteRepo, mockUserRepo, mockSopRepo, mockEmbed)
		execs, err := svc.ProcessBatchEvent(context.Background(), "")

		assert.Error(t, err)
		assert.Nil(t, execs)
		assert.Contains(t, err.Error(), "eventInstanceID parameter is required")
	})
}

func TestAutomationService_TriggerStreamingEvent(t *testing.T) {
	t.Run("successfully triggers a critical cash drop streaming event", func(t *testing.T) {
		siteID := "site-omnimart-1"
		organizerID := "register-terminal-4"
		eventType := model.EventTillDrawerDrop
		desc := "Drawer limits exceeded $1500 trigger drop request"

		mockTaskRepo := &MockTaskRepository{
			FindByIDFunc: func(ctx context.Context, id string) (*model.Task, error) {
				return nil, errors.New("not found")
			},
			CreateFunc: func(ctx context.Context, task *model.Task) error {
				assert.Equal(t, "d000fa44-0000-0000-0000-000000000001", task.ID)
				assert.Equal(t, "Urgent Till Drawer Cash Drop Request", task.Name)
				assert.Equal(t, "ADHOC", task.TaskType)
				return nil
			},
		}

		var capturedExecution *model.TaskExecution
		mockExecRepo := &MockTaskExecutionRepository{
			CreateFunc: func(ctx context.Context, e *model.TaskExecution) error {
				assert.Equal(t, "d000fa44-0000-0000-0000-000000000001", e.TaskTemplateID)
				assert.Equal(t, "ADHOC", e.ExecutionType)
				assert.Equal(t, 1, e.Priority) // Critical alert priority
				assert.Equal(t, "00000000-0000-0000-0000-ffffffffffff", e.EventInstanceID)
				capturedExecution = e
				return nil
			},
		}

		mockSiteRepo := &MockSiteRepository{}
		mockUserRepo := &MockUserRepository{}

		mockSopRepo := &MockSOPRepository{
			QuerySimilarityFunc: func(ctx context.Context, embedding model.Float32Vector, limit int) ([]*model.SOPChunk, error) {
				return []*model.SOPChunk{
					{
						SOPID:      "sop-dallas-vault",
						ChunkIndex: 0,
						Content:    "SOP Vault drop pouch check checklist: Verify cash drawer value, print dynamic drop receipt, and deposit physically in vault.",
					},
				}, nil
			},
		}
		mockEmbed := &MockEmbeddingGenerator{
			GenerateEmbeddingsFunc: func(ctx context.Context, text string) (model.Float32Vector, error) {
				assert.Contains(t, text, "Drawer limits exceeded $1500")
				return make(model.Float32Vector, 768), nil
			},
		}

		svc := service.NewAutomationService(mockExecRepo, mockTaskRepo, mockSiteRepo, mockUserRepo, mockSopRepo, mockEmbed)
		exec, err := svc.TriggerStreamingEvent(context.Background(), siteID, organizerID, eventType, desc)

		assert.NoError(t, err)
		assert.NotNil(t, exec)
		assert.Equal(t, capturedExecution, exec)
		assert.Contains(t, exec.Description, "Drawer limits exceeded $1500")
		assert.Contains(t, exec.Description, "[Grounded SOP Compliance Context]")
		assert.Contains(t, exec.Description, "SOP Vault drop pouch check checklist")
	})

	t.Run("successfully triggers a shelf empty restock alert", func(t *testing.T) {
		siteID := "site-omnimart-2"
		organizerID := "aisle-4-shelf-camera"
		eventType := model.EventStockoutCorrect
		desc := "Produce section empty. Request immediate backroom transfer."

		mockTaskRepo := &MockTaskRepository{
			FindByIDFunc: func(ctx context.Context, id string) (*model.Task, error) {
				return nil, errors.New("not found")
			},
			CreateFunc: func(ctx context.Context, task *model.Task) error {
				assert.Equal(t, "d000fa44-0000-0000-0000-000000000002", task.ID)
				assert.Equal(t, "Critical Shelf Empty Restock Directive", task.Name)
				assert.Equal(t, "ADHOC", task.TaskType)
				return nil
			},
		}

		var capturedExecution *model.TaskExecution
		mockExecRepo := &MockTaskExecutionRepository{
			CreateFunc: func(ctx context.Context, e *model.TaskExecution) error {
				assert.Equal(t, "d000fa44-0000-0000-0000-000000000002", e.TaskTemplateID)
				assert.Equal(t, "ADHOC", e.ExecutionType)
				assert.Equal(t, 2, e.Priority) // High alert priority
				capturedExecution = e
				return nil
			},
		}

		mockSiteRepo := &MockSiteRepository{}
		mockUserRepo := &MockUserRepository{}
		mockSopRepo := &MockSOPRepository{}
		mockEmbed := &MockEmbeddingGenerator{}

		svc := service.NewAutomationService(mockExecRepo, mockTaskRepo, mockSiteRepo, mockUserRepo, mockSopRepo, mockEmbed)
		exec, err := svc.TriggerStreamingEvent(context.Background(), siteID, organizerID, eventType, desc)

		assert.NoError(t, err)
		assert.NotNil(t, exec)
		assert.Equal(t, capturedExecution, exec)
	})

	t.Run("missing parameters trigger failure", func(t *testing.T) {
		mockTaskRepo := &MockTaskRepository{}
		mockExecRepo := &MockTaskExecutionRepository{}
		mockSiteRepo := &MockSiteRepository{}
		mockUserRepo := &MockUserRepository{}
		mockSopRepo := &MockSOPRepository{}
		mockEmbed := &MockEmbeddingGenerator{}

		svc := service.NewAutomationService(mockExecRepo, mockTaskRepo, mockSiteRepo, mockUserRepo, mockSopRepo, mockEmbed)

		exec, err := svc.TriggerStreamingEvent(context.Background(), "", "organizer-1", model.EventTillDrawerDrop, "desc")
		assert.Error(t, err)
		assert.Nil(t, exec)
		assert.Contains(t, err.Error(), "siteID, organizerID, and eventType are mandatory parameters")
	})

	t.Run("ListTemplates returns tasks", func(t *testing.T) {
		expected := []*model.Task{{ID: "t1"}}
		mockTaskRepo := &MockTaskRepository{
			ListFunc: func(ctx context.Context) ([]*model.Task, error) {
				return expected, nil
			},
		}
		svc := service.NewAutomationService(nil, mockTaskRepo, nil, nil, nil, nil)
		res, err := svc.ListTemplates(context.Background())
		assert.NoError(t, err)
		assert.Equal(t, expected, res)
	})
}
