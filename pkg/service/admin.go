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

	"github.com/rmcguinness/gemini_task_engine/pkg/model"
	"github.com/rmcguinness/gemini_task_engine/pkg/persistence"
)

// AdminService handles master data management (MDM) operations for organizational config.
type AdminService interface {
	// Users
	RegisterUser(ctx context.Context, user *model.User) error
	GetUserByID(ctx context.Context, id string) (*model.User, error)
	UpdateUser(ctx context.Context, user *model.User) error
	DeleteUser(ctx context.Context, id string) error
	ListUsers(ctx context.Context) ([]*model.User, error)
	ListUsersRange(ctx context.Context, offset, limit int) ([]*model.User, error)
	AssignRole(ctx context.Context, userID, roleID string) error
	FindUserByOAuth(ctx context.Context, provider, oauthID string) (*model.User, error)

	// Roles
	CreateRole(ctx context.Context, role *model.Role) error
	GetRoleByID(ctx context.Context, id string) (*model.Role, error)
	UpdateRole(ctx context.Context, role *model.Role) error
	DeleteRole(ctx context.Context, id string) error
	ListRoles(ctx context.Context) ([]*model.Role, error)
	ListRolesRange(ctx context.Context, offset, limit int) ([]*model.Role, error)

	// Organizations
	RegisterOrganization(ctx context.Context, org *model.Organization) error
	GetOrganizationByID(ctx context.Context, id string) (*model.Organization, error)
	UpdateOrganization(ctx context.Context, org *model.Organization) error
	DeleteOrganization(ctx context.Context, id string) error
	AssignUserToOrganization(ctx context.Context, orgID, userID string) error
	ListOrganizations(ctx context.Context) ([]*model.Organization, error)
	ListOrganizationsRange(ctx context.Context, offset, limit int) ([]*model.Organization, error)

	// Sites
	RegisterSite(ctx context.Context, site *model.Site) error
	GetSiteByID(ctx context.Context, id string) (*model.Site, error)
	UpdateSite(ctx context.Context, site *model.Site) error
	DeleteSite(ctx context.Context, id string) error
	ListSites(ctx context.Context) ([]*model.Site, error)
	ListSitesRange(ctx context.Context, offset, limit int) ([]*model.Site, error)

	// Locations
	RegisterLocation(ctx context.Context, loc *model.Location) error
	GetLocationByID(ctx context.Context, id string) (*model.Location, error)
	UpdateLocation(ctx context.Context, loc *model.Location) error
	DeleteLocation(ctx context.Context, id string) error
	ListLocations(ctx context.Context) ([]*model.Location, error)
	ListLocationsRange(ctx context.Context, offset, limit int) ([]*model.Location, error)

	// Assets
	RegisterAsset(ctx context.Context, asset *model.Asset) error
	GetAssetByID(ctx context.Context, id string) (*model.Asset, error)
	UpdateAsset(ctx context.Context, asset *model.Asset) error
	DeleteAsset(ctx context.Context, id string) error
	ListAssets(ctx context.Context) ([]*model.Asset, error)
	ListAssetsRange(ctx context.Context, offset, limit int) ([]*model.Asset, error)

	// Tasks
	CreateTaskTemplate(ctx context.Context, task *model.Task) error
	GetTaskTemplateByID(ctx context.Context, id string) (*model.Task, error)
	UpdateTaskTemplate(ctx context.Context, task *model.Task) error
	DeleteTaskTemplate(ctx context.Context, id string) error
	ListTaskTemplates(ctx context.Context) ([]*model.Task, error)
	ListTaskTemplatesRange(ctx context.Context, offset, limit int) ([]*model.Task, error)
}

type adminService struct {
	userRepo persistence.UserRepository
	orgRepo  persistence.OrganizationRepository
	siteRepo persistence.SiteRepository
	taskRepo persistence.TaskRepository
}

// NewAdminService instantiates a new AdminService.
func NewAdminService(
	userRepo persistence.UserRepository,
	orgRepo persistence.OrganizationRepository,
	siteRepo persistence.SiteRepository,
	taskRepo persistence.TaskRepository,
) AdminService {
	return &adminService{
		userRepo: userRepo,
		orgRepo:  orgRepo,
		siteRepo: siteRepo,
		taskRepo: taskRepo,
	}
}

func (s *adminService) RegisterUser(ctx context.Context, user *model.User) error {
	return s.userRepo.Create(ctx, user)
}

func (s *adminService) GetUserByID(ctx context.Context, id string) (*model.User, error) {
	return s.userRepo.FindByID(ctx, id)
}

func (s *adminService) UpdateUser(ctx context.Context, user *model.User) error {
	return s.userRepo.Update(ctx, user)
}

func (s *adminService) DeleteUser(ctx context.Context, id string) error {
	return s.userRepo.Delete(ctx, id)
}

func (s *adminService) ListUsers(ctx context.Context) ([]*model.User, error) {
	return s.userRepo.List(ctx)
}

func (s *adminService) ListUsersRange(ctx context.Context, offset, limit int) ([]*model.User, error) {
	return s.userRepo.ListRange(ctx, offset, limit)
}

func (s *adminService) AssignRole(ctx context.Context, userID, roleID string) error {
	return s.userRepo.AddRole(ctx, userID, roleID)
}

func (s *adminService) FindUserByOAuth(ctx context.Context, provider, oauthID string) (*model.User, error) {
	return s.userRepo.FindByOAuth(ctx, provider, oauthID)
}

func (s *adminService) CreateRole(ctx context.Context, role *model.Role) error {
	return s.userRepo.CreateRole(ctx, role)
}

func (s *adminService) GetRoleByID(ctx context.Context, id string) (*model.Role, error) {
	return s.userRepo.FindRoleByID(ctx, id)
}

func (s *adminService) UpdateRole(ctx context.Context, role *model.Role) error {
	return s.userRepo.UpdateRole(ctx, role)
}

func (s *adminService) DeleteRole(ctx context.Context, id string) error {
	return s.userRepo.DeleteRole(ctx, id)
}

func (s *adminService) ListRoles(ctx context.Context) ([]*model.Role, error) {
	return s.userRepo.ListRoles(ctx)
}

func (s *adminService) ListRolesRange(ctx context.Context, offset, limit int) ([]*model.Role, error) {
	return s.userRepo.ListRolesRange(ctx, offset, limit)
}

func (s *adminService) RegisterOrganization(ctx context.Context, org *model.Organization) error {
	return s.orgRepo.Create(ctx, org)
}

func (s *adminService) GetOrganizationByID(ctx context.Context, id string) (*model.Organization, error) {
	return s.orgRepo.FindByID(ctx, id)
}

func (s *adminService) UpdateOrganization(ctx context.Context, org *model.Organization) error {
	return s.orgRepo.Update(ctx, org)
}

func (s *adminService) DeleteOrganization(ctx context.Context, id string) error {
	return s.orgRepo.Delete(ctx, id)
}

func (s *adminService) AssignUserToOrganization(ctx context.Context, orgID, userID string) error {
	return s.orgRepo.AddUser(ctx, orgID, userID)
}

func (s *adminService) ListOrganizations(ctx context.Context) ([]*model.Organization, error) {
	return s.orgRepo.List(ctx)
}

func (s *adminService) ListOrganizationsRange(ctx context.Context, offset, limit int) ([]*model.Organization, error) {
	return s.orgRepo.ListRange(ctx, offset, limit)
}

func (s *adminService) RegisterSite(ctx context.Context, site *model.Site) error {
	return s.siteRepo.Create(ctx, site)
}

func (s *adminService) GetSiteByID(ctx context.Context, id string) (*model.Site, error) {
	return s.siteRepo.FindByID(ctx, id)
}

func (s *adminService) UpdateSite(ctx context.Context, site *model.Site) error {
	return s.siteRepo.Update(ctx, site)
}

func (s *adminService) DeleteSite(ctx context.Context, id string) error {
	return s.siteRepo.Delete(ctx, id)
}

func (s *adminService) ListSites(ctx context.Context) ([]*model.Site, error) {
	return s.siteRepo.List(ctx)
}

func (s *adminService) ListSitesRange(ctx context.Context, offset, limit int) ([]*model.Site, error) {
	return s.siteRepo.ListRange(ctx, offset, limit)
}

func (s *adminService) RegisterLocation(ctx context.Context, loc *model.Location) error {
	return s.siteRepo.CreateLocation(ctx, loc)
}

func (s *adminService) GetLocationByID(ctx context.Context, id string) (*model.Location, error) {
	return s.siteRepo.FindLocationByID(ctx, id)
}

func (s *adminService) UpdateLocation(ctx context.Context, loc *model.Location) error {
	return s.siteRepo.UpdateLocation(ctx, loc)
}

func (s *adminService) DeleteLocation(ctx context.Context, id string) error {
	return s.siteRepo.DeleteLocation(ctx, id)
}

func (s *adminService) ListLocations(ctx context.Context) ([]*model.Location, error) {
	return s.siteRepo.ListLocations(ctx)
}

func (s *adminService) ListLocationsRange(ctx context.Context, offset, limit int) ([]*model.Location, error) {
	return s.siteRepo.ListLocationsRange(ctx, offset, limit)
}

func (s *adminService) RegisterAsset(ctx context.Context, asset *model.Asset) error {
	return s.siteRepo.CreateAsset(ctx, asset)
}

func (s *adminService) GetAssetByID(ctx context.Context, id string) (*model.Asset, error) {
	return s.siteRepo.FindAssetByID(ctx, id)
}

func (s *adminService) UpdateAsset(ctx context.Context, asset *model.Asset) error {
	return s.siteRepo.UpdateAsset(ctx, asset)
}

func (s *adminService) DeleteAsset(ctx context.Context, id string) error {
	return s.siteRepo.DeleteAsset(ctx, id)
}

func (s *adminService) ListAssets(ctx context.Context) ([]*model.Asset, error) {
	return s.siteRepo.ListAssets(ctx)
}

func (s *adminService) ListAssetsRange(ctx context.Context, offset, limit int) ([]*model.Asset, error) {
	return s.siteRepo.ListAssetsRange(ctx, offset, limit)
}

func (s *adminService) CreateTaskTemplate(ctx context.Context, task *model.Task) error {
	return s.taskRepo.Create(ctx, task)
}

func (s *adminService) GetTaskTemplateByID(ctx context.Context, id string) (*model.Task, error) {
	return s.taskRepo.FindByID(ctx, id)
}

func (s *adminService) UpdateTaskTemplate(ctx context.Context, task *model.Task) error {
	return s.taskRepo.Update(ctx, task)
}

func (s *adminService) DeleteTaskTemplate(ctx context.Context, id string) error {
	return s.taskRepo.Delete(ctx, id)
}

func (s *adminService) ListTaskTemplates(ctx context.Context) ([]*model.Task, error) {
	return s.taskRepo.List(ctx)
}

func (s *adminService) ListTaskTemplatesRange(ctx context.Context, offset, limit int) ([]*model.Task, error) {
	return s.taskRepo.ListRange(ctx, offset, limit)
}
