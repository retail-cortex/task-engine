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

// OrganizationRepository manages corporate units/tenants.
type OrganizationRepository interface {
	Create(ctx context.Context, o *model.Organization) error
	FindByID(ctx context.Context, id string) (*model.Organization, error)
	Update(ctx context.Context, o *model.Organization) error
	Delete(ctx context.Context, id string) error
	AddUser(ctx context.Context, organizationID, userID string) error
	List(ctx context.Context) ([]*model.Organization, error)
	ListRange(ctx context.Context, offset, limit int) ([]*model.Organization, error)
}

type organizationRepository struct {
	db *gorm.DB
}

// NewOrganizationRepository creates a new OrganizationRepository instance.
func NewOrganizationRepository(db *gorm.DB) OrganizationRepository {
	return &organizationRepository{db: db}
}

func (r *organizationRepository) Create(ctx context.Context, o *model.Organization) error {
	return r.db.WithContext(ctx).Create(o).Error
}

func (r *organizationRepository) FindByID(ctx context.Context, id string) (*model.Organization, error) {
	var o model.Organization
	err := r.db.WithContext(ctx).Preload("Sites").Preload("Users").First(&o, "id = ?", id).Error
	return &o, err
}

func (r *organizationRepository) Update(ctx context.Context, o *model.Organization) error {
	return r.db.WithContext(ctx).Save(o).Error
}

func (r *organizationRepository) Delete(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Delete(&model.Organization{}, "id = ?", id).Error
}

func (r *organizationRepository) AddUser(ctx context.Context, organizationID, userID string) error {
	uo := model.UserOrganization{OrganizationID: organizationID, UserID: userID}
	return r.db.WithContext(ctx).Create(&uo).Error
}

func (r *organizationRepository) List(ctx context.Context) ([]*model.Organization, error) {
	var orgs []*model.Organization
	err := r.db.WithContext(ctx).Find(&orgs).Error
	return orgs, err
}

func (r *organizationRepository) ListRange(ctx context.Context, offset, limit int) ([]*model.Organization, error) {
	var orgs []*model.Organization
	err := r.db.WithContext(ctx).Offset(offset).Limit(limit).Find(&orgs).Error
	return orgs, err
}
