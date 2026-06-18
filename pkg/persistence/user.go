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

// UserRepository manages profile and role operations.
type UserRepository interface {
	Create(ctx context.Context, u *model.User) error
	FindByID(ctx context.Context, id string) (*model.User, error)
	FindByOAuth(ctx context.Context, provider, oauthID string) (*model.User, error)
	Update(ctx context.Context, u *model.User) error
	AddRole(ctx context.Context, userID, roleID string) error
	List(ctx context.Context) ([]*model.User, error)
	ListRange(ctx context.Context, offset, limit int) ([]*model.User, error)
	Delete(ctx context.Context, id string) error
	ListActiveOnShiftUsers(ctx context.Context, siteID string) ([]*model.User, error)
	CreateRole(ctx context.Context, r *model.Role) error
	FindRoleByID(ctx context.Context, id string) (*model.Role, error)
	UpdateRole(ctx context.Context, r *model.Role) error
	DeleteRole(ctx context.Context, id string) error
	ListRoles(ctx context.Context) ([]*model.Role, error)
	ListRolesRange(ctx context.Context, offset, limit int) ([]*model.Role, error)
}

type userRepository struct {
	db *gorm.DB
}

// NewUserRepository creates a new UserRepository instance.
func NewUserRepository(db *gorm.DB) UserRepository {
	return &userRepository{db: db}
}

func (r *userRepository) Create(ctx context.Context, u *model.User) error {
	return r.db.WithContext(ctx).Create(u).Error
}

func (r *userRepository) FindByID(ctx context.Context, id string) (*model.User, error) {
	var u model.User
	err := r.db.WithContext(ctx).Preload("Roles").Preload("Organizations").Preload("Sites").First(&u, "id = ?", id).Error
	return &u, err
}

func (r *userRepository) Update(ctx context.Context, u *model.User) error {
	return r.db.WithContext(ctx).Save(u).Error
}

func (r *userRepository) AddRole(ctx context.Context, userID, roleID string) error {
	ur := model.UserRole{UserID: userID, RoleID: roleID}
	return r.db.WithContext(ctx).Create(&ur).Error
}

func (r *userRepository) List(ctx context.Context) ([]*model.User, error) {
	var users []*model.User
	err := r.db.WithContext(ctx).Preload("Roles").Preload("Organizations").Preload("Sites").Find(&users).Error
	return users, err
}

func (r *userRepository) ListRange(ctx context.Context, offset, limit int) ([]*model.User, error) {
	var users []*model.User
	err := r.db.WithContext(ctx).Preload("Roles").Preload("Organizations").Preload("Sites").Offset(offset).Limit(limit).Find(&users).Error
	return users, err
}

func (r *userRepository) Delete(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Delete(&model.User{}, "id = ?", id).Error
}

func (r *userRepository) ListActiveOnShiftUsers(ctx context.Context, siteID string) ([]*model.User, error) {
	var users []*model.User
	err := r.db.WithContext(ctx).
		Preload("Roles").
		Preload("Organizations").
		Preload("Sites").
		Joins("JOIN user_event_schedules ON user_event_schedules.user_id = users.id").
		Joins("JOIN user_event_instances ON user_event_instances.schedule_id = user_event_schedules.id").
		Joins("JOIN events ON events.id = user_event_schedules.event_id").
		Where("events.site_id = ? AND events.event_type = 'RetailShift' AND user_event_instances.event_status = 'EventActive'", siteID).
		Find(&users).Error
	return users, err
}

func (r *userRepository) CreateRole(ctx context.Context, role *model.Role) error {
	return r.db.WithContext(ctx).Create(role).Error
}

func (r *userRepository) FindRoleByID(ctx context.Context, id string) (*model.Role, error) {
	var role model.Role
	err := r.db.WithContext(ctx).First(&role, "id = ?", id).Error
	return &role, err
}

func (r *userRepository) UpdateRole(ctx context.Context, role *model.Role) error {
	return r.db.WithContext(ctx).Save(role).Error
}

func (r *userRepository) DeleteRole(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Delete(&model.Role{}, "id = ?", id).Error
}

func (r *userRepository) ListRoles(ctx context.Context) ([]*model.Role, error) {
	var roles []*model.Role
	err := r.db.WithContext(ctx).Find(&roles).Error
	return roles, err
}

func (r *userRepository) ListRolesRange(ctx context.Context, offset, limit int) ([]*model.Role, error) {
	var roles []*model.Role
	err := r.db.WithContext(ctx).Offset(offset).Limit(limit).Find(&roles).Error
	return roles, err
}

func (r *userRepository) FindByOAuth(ctx context.Context, provider, oauthID string) (*model.User, error) {
	var u model.User
	err := r.db.WithContext(ctx).Preload("Roles").Preload("Organizations").Preload("Sites").First(&u, "o_auth_provider = ? AND o_auth_id = ?", provider, oauthID).Error
	return &u, err
}
