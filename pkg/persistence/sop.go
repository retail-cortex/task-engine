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

// SOPRepository coordinates the storage, indexing processes, and cosine distance vector searches.
type SOPRepository interface {
	Create(ctx context.Context, s *model.SOP) error
	FindByID(ctx context.Context, id string) (*model.SOP, error)
	Update(ctx context.Context, s *model.SOP) error
	CreateProcess(ctx context.Context, p *model.SOPProcess) error
	FindProcessByID(ctx context.Context, id string) (*model.SOPProcess, error)
	UpdateProcess(ctx context.Context, p *model.SOPProcess) error
	CreateChunks(ctx context.Context, chunks []*model.SOPChunk) error
	QuerySimilarity(ctx context.Context, embedding model.Float32Vector, limit int) ([]*model.SOPChunk, error)
	List(ctx context.Context) ([]*model.SOP, error)
	ListRange(ctx context.Context, offset, limit int) ([]*model.SOP, error)
	Delete(ctx context.Context, id string) error
	ListProcesses(ctx context.Context) ([]*model.SOPProcess, error)
	ListProcessesRange(ctx context.Context, offset, limit int) ([]*model.SOPProcess, error)
	DeleteProcess(ctx context.Context, id string) error
}

type sopRepository struct {
	db *gorm.DB
}

// NewSOPRepository creates a new SOPRepository instance.
func NewSOPRepository(db *gorm.DB) SOPRepository {
	return &sopRepository{db: db}
}

func (r *sopRepository) Create(ctx context.Context, s *model.SOP) error {
	return r.db.WithContext(ctx).Create(s).Error
}

func (r *sopRepository) FindByID(ctx context.Context, id string) (*model.SOP, error) {
	var s model.SOP
	err := r.db.WithContext(ctx).First(&s, "id = ?", id).Error
	return &s, err
}

func (r *sopRepository) Update(ctx context.Context, s *model.SOP) error {
	return r.db.WithContext(ctx).Save(s).Error
}

func (r *sopRepository) CreateProcess(ctx context.Context, p *model.SOPProcess) error {
	return r.db.WithContext(ctx).Create(p).Error
}

func (r *sopRepository) FindProcessByID(ctx context.Context, id string) (*model.SOPProcess, error) {
	var p model.SOPProcess
	err := r.db.WithContext(ctx).First(&p, "id = ?", id).Error
	return &p, err
}

func (r *sopRepository) UpdateProcess(ctx context.Context, p *model.SOPProcess) error {
	return r.db.WithContext(ctx).Save(p).Error
}

func (r *sopRepository) CreateChunks(ctx context.Context, chunks []*model.SOPChunk) error {
	return r.db.WithContext(ctx).Create(&chunks).Error
}

func (r *sopRepository) QuerySimilarity(ctx context.Context, embedding model.Float32Vector, limit int) ([]*model.SOPChunk, error) {
	var chunks []*model.SOPChunk
	// GORM pgvector cosine distance lookup utilizing pgvector order Operator '<=>'
	err := r.db.WithContext(ctx).
		Order(gorm.Expr("embedding <=> ?", embedding)).
		Limit(limit).
		Find(&chunks).Error
	return chunks, err
}

func (r *sopRepository) List(ctx context.Context) ([]*model.SOP, error) {
	var list []*model.SOP
	err := r.db.WithContext(ctx).Find(&list).Error
	return list, err
}

func (r *sopRepository) ListRange(ctx context.Context, offset, limit int) ([]*model.SOP, error) {
	var list []*model.SOP
	err := r.db.WithContext(ctx).Offset(offset).Limit(limit).Find(&list).Error
	return list, err
}

func (r *sopRepository) Delete(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Delete(&model.SOP{}, "id = ?", id).Error
}

func (r *sopRepository) ListProcesses(ctx context.Context) ([]*model.SOPProcess, error) {
	var list []*model.SOPProcess
	err := r.db.WithContext(ctx).Find(&list).Error
	return list, err
}

func (r *sopRepository) ListProcessesRange(ctx context.Context, offset, limit int) ([]*model.SOPProcess, error) {
	var list []*model.SOPProcess
	err := r.db.WithContext(ctx).Offset(offset).Limit(limit).Find(&list).Error
	return list, err
}

func (r *sopRepository) DeleteProcess(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Delete(&model.SOPProcess{}, "id = ?", id).Error
}
