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

// TaskExecutionRepository handles queues, state mutations, trades, and explicit audits.
type TaskExecutionRepository interface {
	Create(ctx context.Context, e *model.TaskExecution) error
	FindByID(ctx context.Context, id string) (*model.TaskExecution, error)
	Update(ctx context.Context, e *model.TaskExecution) error
	GetQueue(ctx context.Context, siteID string) ([]*model.TaskExecution, error)
	GetOrgTasks(ctx context.Context, orgID string) ([]*model.TaskExecution, error)
	GetUserSiteTasks(ctx context.Context, siteID, userID string) ([]*model.TaskExecution, error)
	CreateTrade(ctx context.Context, t *model.TaskTrade) error
	FindTradeByID(ctx context.Context, id string) (*model.TaskTrade, error)
	UpdateTrade(ctx context.Context, t *model.TaskTrade) error
	FindPendingTradesForUser(ctx context.Context, userID string) ([]*model.TaskTrade, error)
	FindPendingTradeByExecution(ctx context.Context, executionID string) (*model.TaskTrade, error)
	CreateAudit(ctx context.Context, a *model.TaskExecutionAudit) error
	List(ctx context.Context) ([]*model.TaskExecution, error)
	ListRange(ctx context.Context, offset, limit int) ([]*model.TaskExecution, error)
	Delete(ctx context.Context, id string) error
	GetSiteIDForExecution(ctx context.Context, execID string) (string, error)
}

type taskExecutionRepository struct {
	db *gorm.DB
}

// NewTaskExecutionRepository creates a new TaskExecutionRepository instance.
func NewTaskExecutionRepository(db *gorm.DB) TaskExecutionRepository {
	return &taskExecutionRepository{db: db}
}

func (r *taskExecutionRepository) Create(ctx context.Context, e *model.TaskExecution) error {
	return r.db.WithContext(ctx).Create(e).Error
}

func (r *taskExecutionRepository) FindByID(ctx context.Context, id string) (*model.TaskExecution, error) {
	var e model.TaskExecution
	err := r.db.WithContext(ctx).Preload("Task").Preload("Task.SOPs").Preload("Assignee").First(&e, "id = ?", id).Error
	return &e, err
}

func (r *taskExecutionRepository) Update(ctx context.Context, e *model.TaskExecution) error {
	return r.db.WithContext(ctx).Save(e).Error
}

func (r *taskExecutionRepository) GetQueue(ctx context.Context, siteID string) ([]*model.TaskExecution, error) {
	var list []*model.TaskExecution
	// Join with event instances and schedules to filter by siteID
	err := r.db.WithContext(ctx).
		Preload("Task").
		Preload("Task.SOPs").
		Preload("Assignee").
		Joins("JOIN user_event_instances ON user_event_instances.id = task_executions.event_instance_id").
		Joins("JOIN user_event_schedules ON user_event_schedules.id = user_event_instances.schedule_id").
		Joins("JOIN events ON events.id = user_event_schedules.event_id").
		Where("events.site_id = ?", siteID).
		Order("task_executions.priority ASC, task_executions.created_at ASC").
		Find(&list).Error
	return list, err
}

func (r *taskExecutionRepository) GetOrgTasks(ctx context.Context, orgID string) ([]*model.TaskExecution, error) {
	var list []*model.TaskExecution
	err := r.db.WithContext(ctx).
		Preload("Task").
		Preload("Task.SOPs").
		Preload("Assignee").
		Joins("JOIN user_event_instances ON user_event_instances.id = task_executions.event_instance_id").
		Joins("JOIN user_event_schedules ON user_event_schedules.id = user_event_instances.schedule_id").
		Joins("JOIN events ON events.id = user_event_schedules.event_id").
		Joins("JOIN sites ON sites.id = events.site_id").
		Where("sites.organization_id = ?", orgID).
		Order("task_executions.priority ASC, task_executions.created_at ASC").
		Find(&list).Error
	return list, err
}

func (r *taskExecutionRepository) GetUserSiteTasks(ctx context.Context, siteID, userID string) ([]*model.TaskExecution, error) {
	var list []*model.TaskExecution

	// 1. Resolve the requesting user's roles to check operational privileges
	var user model.User
	isManager := false
	if err := r.db.WithContext(ctx).Preload("Roles").First(&user, "id = ?", userID).Error; err == nil {
		for _, role := range user.Roles {
			if role.Name == "ADMIN" || role.Name == "REGION_MANAGER" || role.Name == "SITE_MANAGER" {
				isManager = true
				break
			}
		}
	}

	// 2. Query tasks with elevated manager/admin visibility or standard associate restriction
	query := r.db.WithContext(ctx).
		Preload("Task").
		Preload("Task.SOPs").
		Preload("Assignee").
		Joins("JOIN user_event_instances ON user_event_instances.id = task_executions.event_instance_id").
		Joins("JOIN user_event_schedules ON user_event_schedules.id = user_event_instances.schedule_id").
		Joins("JOIN events ON events.id = user_event_schedules.event_id")

	if isManager {
		// Managers & Admins see ALL active store tasks
		err := query.Where("events.site_id = ?", siteID).
			Order("task_executions.priority ASC, task_executions.created_at ASC").
			Find(&list).Error
		return list, err
	}

	// Standard associates only see their personally assigned tasks
	err := query.Where("events.site_id = ? AND task_executions.assignee_id = ?", siteID, userID).
		Order("task_executions.priority ASC, task_executions.created_at ASC").
		Find(&list).Error
	return list, err
}

func (r *taskExecutionRepository) CreateTrade(ctx context.Context, t *model.TaskTrade) error {
	return r.db.WithContext(ctx).Create(t).Error
}

func (r *taskExecutionRepository) FindTradeByID(ctx context.Context, id string) (*model.TaskTrade, error) {
	var t model.TaskTrade
	err := r.db.WithContext(ctx).First(&t, "id = ?", id).Error
	return &t, err
}

func (r *taskExecutionRepository) UpdateTrade(ctx context.Context, t *model.TaskTrade) error {
	return r.db.WithContext(ctx).Save(t).Error
}

func (r *taskExecutionRepository) FindPendingTradesForUser(ctx context.Context, userID string) ([]*model.TaskTrade, error) {
	var trades []*model.TaskTrade
	err := r.db.WithContext(ctx).Where("proposed_assignee_id = ? AND status = 'PENDING'", userID).Find(&trades).Error
	return trades, err
}

func (r *taskExecutionRepository) FindPendingTradeByExecution(ctx context.Context, executionID string) (*model.TaskTrade, error) {
	var t model.TaskTrade
	err := r.db.WithContext(ctx).Where("task_execution_id = ? AND status = 'PENDING'", executionID).First(&t).Error
	return &t, err
}

func (r *taskExecutionRepository) CreateAudit(ctx context.Context, a *model.TaskExecutionAudit) error {
	return r.db.WithContext(ctx).Create(a).Error
}

func (r *taskExecutionRepository) List(ctx context.Context) ([]*model.TaskExecution, error) {
	var list []*model.TaskExecution
	err := r.db.WithContext(ctx).Preload("Task").Preload("Assignee").Find(&list).Error
	return list, err
}

func (r *taskExecutionRepository) ListRange(ctx context.Context, offset, limit int) ([]*model.TaskExecution, error) {
	var list []*model.TaskExecution
	err := r.db.WithContext(ctx).Preload("Task").Preload("Assignee").Offset(offset).Limit(limit).Find(&list).Error
	return list, err
}

func (r *taskExecutionRepository) Delete(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Delete(&model.TaskExecution{}, "id = ?", id).Error
}

func (r *taskExecutionRepository) GetSiteIDForExecution(ctx context.Context, execID string) (string, error) {
	var siteID string
	err := r.db.WithContext(ctx).Table("task_executions").
		Select("events.site_id").
		Joins("JOIN user_event_instances ON user_event_instances.id = task_executions.event_instance_id").
		Joins("JOIN user_event_schedules ON user_event_schedules.id = user_event_instances.schedule_id").
		Joins("JOIN events ON events.id = user_event_schedules.event_id").
		Where("task_executions.id = ?", execID).
		Scan(&siteID).Error
	return siteID, err
}
