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

// ShiftAgentSessionRepository manages context windows for Gemini ADK model sessions.
type ShiftAgentSessionRepository interface {
	Create(ctx context.Context, s *model.ShiftAgentSession) error
	FindByID(ctx context.Context, id string) (*model.ShiftAgentSession, error)
	FindByShift(ctx context.Context, assigneeID, shiftInstanceID string) (*model.ShiftAgentSession, error)
	Update(ctx context.Context, s *model.ShiftAgentSession) error
	List(ctx context.Context) ([]*model.ShiftAgentSession, error)
	ListRange(ctx context.Context, offset, limit int) ([]*model.ShiftAgentSession, error)
	Delete(ctx context.Context, id string) error
}

type shiftAgentSessionRepository struct {
	db *gorm.DB
}

// NewShiftAgentSessionRepository creates a new ShiftAgentSessionRepository instance.
func NewShiftAgentSessionRepository(db *gorm.DB) ShiftAgentSessionRepository {
	return &shiftAgentSessionRepository{db: db}
}

func (r *shiftAgentSessionRepository) Create(ctx context.Context, s *model.ShiftAgentSession) error {
	return r.db.WithContext(ctx).Create(s).Error
}

func (r *shiftAgentSessionRepository) FindByID(ctx context.Context, id string) (*model.ShiftAgentSession, error) {
	var s model.ShiftAgentSession
	err := r.db.WithContext(ctx).First(&s, "id = ?", id).Error
	return &s, err
}

func (r *shiftAgentSessionRepository) FindByShift(ctx context.Context, assigneeID, shiftInstanceID string) (*model.ShiftAgentSession, error) {
	var s model.ShiftAgentSession
	err := r.db.WithContext(ctx).First(&s, "assignee_id = ? AND shift_instance_id = ?", assigneeID, shiftInstanceID).Error
	return &s, err
}

func (r *shiftAgentSessionRepository) Update(ctx context.Context, s *model.ShiftAgentSession) error {
	return r.db.WithContext(ctx).Save(s).Error
}

func (r *shiftAgentSessionRepository) List(ctx context.Context) ([]*model.ShiftAgentSession, error) {
	var list []*model.ShiftAgentSession
	err := r.db.WithContext(ctx).Find(&list).Error
	return list, err
}

func (r *shiftAgentSessionRepository) ListRange(ctx context.Context, offset, limit int) ([]*model.ShiftAgentSession, error) {
	var list []*model.ShiftAgentSession
	err := r.db.WithContext(ctx).Offset(offset).Limit(limit).Find(&list).Error
	return list, err
}

func (r *shiftAgentSessionRepository) Delete(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Delete(&model.ShiftAgentSession{}, "id = ?", id).Error
}
