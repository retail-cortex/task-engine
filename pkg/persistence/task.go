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

	"github.com/rmcguinness/gemini_task_engine/pkg/model"
	"gorm.io/gorm"
)

// TaskRepository manages task definitions, rules, checklist templates, and certifications.
type TaskRepository interface {
	Create(ctx context.Context, t *model.Task) error
	FindByID(ctx context.Context, id string) (*model.Task, error)
	Update(ctx context.Context, t *model.Task) error
	Delete(ctx context.Context, id string) error
	List(ctx context.Context) ([]*model.Task, error)
	ListRange(ctx context.Context, offset, limit int) ([]*model.Task, error)
	AddApprovalRule(ctx context.Context, r *model.TaskApprovalRule) error
}

type taskRepository struct {
	db *gorm.DB
}

// NewTaskRepository creates a new TaskRepository instance.
func NewTaskRepository(db *gorm.DB) TaskRepository {
	return &taskRepository{db: db}
}

func (r *taskRepository) Create(ctx context.Context, t *model.Task) error {
	return r.db.WithContext(ctx).Create(t).Error
}

func (r *taskRepository) FindByID(ctx context.Context, id string) (*model.Task, error) {
	var t model.Task
	err := r.db.WithContext(ctx).Preload("Assets").Preload("ApprovalRules").Preload("SOPs").First(&t, "id = ?", id).Error
	return &t, err
}

func (r *taskRepository) Update(ctx context.Context, t *model.Task) error {
	return r.db.WithContext(ctx).Save(t).Error
}

func (r *taskRepository) Delete(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Delete(&model.Task{}, "id = ?", id).Error
}

func (r *taskRepository) AddApprovalRule(ctx context.Context, rule *model.TaskApprovalRule) error {
	return r.db.WithContext(ctx).Create(rule).Error
}

func (r *taskRepository) List(ctx context.Context) ([]*model.Task, error) {
	var tasks []*model.Task
	err := r.db.WithContext(ctx).Preload("Assets").Preload("ApprovalRules").Preload("SOPs").Find(&tasks).Error
	return tasks, err
}

func (r *taskRepository) ListRange(ctx context.Context, offset, limit int) ([]*model.Task, error) {
	var tasks []*model.Task
	err := r.db.WithContext(ctx).Preload("Assets").Preload("ApprovalRules").Preload("SOPs").Offset(offset).Limit(limit).Find(&tasks).Error
	return tasks, err
}
