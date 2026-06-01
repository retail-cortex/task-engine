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
	CreateRole(ctx context.Context, r *model.Role) error
	ListRoles(ctx context.Context) ([]*model.Role, error)
}

// OrganizationRepository manages corporate units/tenants.
type OrganizationRepository interface {
	Create(ctx context.Context, o *model.Organization) error
	FindByID(ctx context.Context, id string) (*model.Organization, error)
	AddUser(ctx context.Context, organizationID, userID string) error
	List(ctx context.Context) ([]*model.Organization, error)
}

// SiteRepository manages physical storefronts/facilities and sub-locations (fixtures/shelves).
type SiteRepository interface {
	Create(ctx context.Context, s *model.Site) error
	FindByID(ctx context.Context, id string) (*model.Site, error)
	List(ctx context.Context) ([]*model.Site, error)
	CreateLocation(ctx context.Context, l *model.Location) error
	FindLocationByID(ctx context.Context, id string) (*model.Location, error)
	CreateAsset(ctx context.Context, a *model.Asset) error
	FindAssetByID(ctx context.Context, id string) (*model.Asset, error)
	UpdateAsset(ctx context.Context, a *model.Asset) error
}

// TaskRepository manages task definitions, rules, checklist templates, and certifications.
type TaskRepository interface {
	Create(ctx context.Context, t *model.Task) error
	FindByID(ctx context.Context, id string) (*model.Task, error)
	List(ctx context.Context) ([]*model.Task, error)
	AddApprovalRule(ctx context.Context, r *model.TaskApprovalRule) error
}

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
	CreateAudit(ctx context.Context, a *model.TaskExecutionAudit) error
}

// ShiftAgentSessionRepository manages context windows for Gemini ADK model sessions.
type ShiftAgentSessionRepository interface {
	Create(ctx context.Context, s *model.ShiftAgentSession) error
	FindByID(ctx context.Context, id string) (*model.ShiftAgentSession, error)
	FindByShift(ctx context.Context, assigneeID, shiftInstanceID string) (*model.ShiftAgentSession, error)
	Update(ctx context.Context, s *model.ShiftAgentSession) error
}

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
}

// GORM Implementations

type userRepository struct {
	db *gorm.DB
}

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

func (r *userRepository) CreateRole(ctx context.Context, role *model.Role) error {
	return r.db.WithContext(ctx).Create(role).Error
}

func (r *userRepository) ListRoles(ctx context.Context) ([]*model.Role, error) {
	var roles []*model.Role
	err := r.db.WithContext(ctx).Find(&roles).Error
	return roles, err
}

func (r *userRepository) FindByOAuth(ctx context.Context, provider, oauthID string) (*model.User, error) {
	var u model.User
	err := r.db.WithContext(ctx).Preload("Roles").Preload("Organizations").Preload("Sites").First(&u, "o_auth_provider = ? AND o_auth_id = ?", provider, oauthID).Error
	return &u, err
}

type organizationRepository struct {
	db *gorm.DB
}

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

func (r *organizationRepository) AddUser(ctx context.Context, organizationID, userID string) error {
	uo := model.UserOrganization{OrganizationID: organizationID, UserID: userID}
	return r.db.WithContext(ctx).Create(&uo).Error
}

func (r *organizationRepository) List(ctx context.Context) ([]*model.Organization, error) {
	var orgs []*model.Organization
	err := r.db.WithContext(ctx).Find(&orgs).Error
	return orgs, err
}

type siteRepository struct {
	db *gorm.DB
}

func NewSiteRepository(db *gorm.DB) SiteRepository {
	return &siteRepository{db: db}
}

func (r *siteRepository) Create(ctx context.Context, s *model.Site) error {
	return r.db.WithContext(ctx).Create(s).Error
}

func (r *siteRepository) FindByID(ctx context.Context, id string) (*model.Site, error) {
	var s model.Site
	err := r.db.WithContext(ctx).Preload("Locations").First(&s, "id = ?", id).Error
	return &s, err
}

func (r *siteRepository) List(ctx context.Context) ([]*model.Site, error) {
	var list []*model.Site
	if err := r.db.WithContext(ctx).Find(&list).Error; err != nil {
		return nil, err
	}
	return list, nil
}

func (r *siteRepository) CreateLocation(ctx context.Context, l *model.Location) error {
	return r.db.WithContext(ctx).Create(l).Error
}

func (r *siteRepository) FindLocationByID(ctx context.Context, id string) (*model.Location, error) {
	var l model.Location
	err := r.db.WithContext(ctx).First(&l, "id = ?", id).Error
	return &l, err
}

func (r *siteRepository) CreateAsset(ctx context.Context, a *model.Asset) error {
	return r.db.WithContext(ctx).Create(a).Error
}

func (r *siteRepository) FindAssetByID(ctx context.Context, id string) (*model.Asset, error) {
	var a model.Asset
	err := r.db.WithContext(ctx).First(&a, "id = ?", id).Error
	return &a, err
}

func (r *siteRepository) UpdateAsset(ctx context.Context, a *model.Asset) error {
	return r.db.WithContext(ctx).Save(a).Error
}

type taskRepository struct {
	db *gorm.DB
}

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

func (r *taskRepository) AddApprovalRule(ctx context.Context, rule *model.TaskApprovalRule) error {
	return r.db.WithContext(ctx).Create(rule).Error
}

func (r *taskRepository) List(ctx context.Context) ([]*model.Task, error) {
	var tasks []*model.Task
	err := r.db.WithContext(ctx).Preload("Assets").Preload("ApprovalRules").Preload("SOPs").Find(&tasks).Error
	return tasks, err
}

type taskExecutionRepository struct {
	db *gorm.DB
}

func NewTaskExecutionRepository(db *gorm.DB) TaskExecutionRepository {
	return &taskExecutionRepository{db: db}
}

func (r *taskExecutionRepository) Create(ctx context.Context, e *model.TaskExecution) error {
	return r.db.WithContext(ctx).Create(e).Error
}

func (r *taskExecutionRepository) FindByID(ctx context.Context, id string) (*model.TaskExecution, error) {
	var e model.TaskExecution
	err := r.db.WithContext(ctx).Preload("Task").First(&e, "id = ?", id).Error
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
	err := r.db.WithContext(ctx).
		Preload("Task").
		Joins("JOIN user_event_instances ON user_event_instances.id = task_executions.event_instance_id").
		Joins("JOIN user_event_schedules ON user_event_schedules.id = user_event_instances.schedule_id").
		Joins("JOIN events ON events.id = user_event_schedules.event_id").
		Where("events.site_id = ? AND task_executions.assignee_id = ?", siteID, userID).
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

func (r *taskExecutionRepository) CreateAudit(ctx context.Context, a *model.TaskExecutionAudit) error {
	return r.db.WithContext(ctx).Create(a).Error
}

type shiftAgentSessionRepository struct {
	db *gorm.DB
}

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

type sopRepository struct {
	db *gorm.DB
}

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
