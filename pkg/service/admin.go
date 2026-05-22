package service

import (
	"context"

	"github.com/rmcguinness/gemini_task_engine/pkg/model"
	"github.com/rmcguinness/gemini_task_engine/pkg/persistence"
)

// AdminService handles master data management (MDM) operations for organizational config.
type AdminService interface {
	RegisterUser(ctx context.Context, user *model.User) error
	AssignRole(ctx context.Context, userID, roleID string) error
	RegisterOrganization(ctx context.Context, org *model.Organization) error
	AssignUserToOrganization(ctx context.Context, orgID, userID string) error
	ListOrganizations(ctx context.Context) ([]*model.Organization, error)
	RegisterSite(ctx context.Context, site *model.Site) error
	RegisterLocation(ctx context.Context, loc *model.Location) error
	RegisterAsset(ctx context.Context, asset *model.Asset) error
	CreateTaskTemplate(ctx context.Context, task *model.Task) error
	ListUsers(ctx context.Context) ([]*model.User, error)
	CreateRole(ctx context.Context, role *model.Role) error
	ListRoles(ctx context.Context) ([]*model.Role, error)
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

func (s *adminService) AssignRole(ctx context.Context, userID, roleID string) error {
	return s.userRepo.AddRole(ctx, userID, roleID)
}

func (s *adminService) RegisterOrganization(ctx context.Context, org *model.Organization) error {
	return s.orgRepo.Create(ctx, org)
}

func (s *adminService) AssignUserToOrganization(ctx context.Context, orgID, userID string) error {
	return s.orgRepo.AddUser(ctx, orgID, userID)
}

func (s *adminService) ListOrganizations(ctx context.Context) ([]*model.Organization, error) {
	return s.orgRepo.List(ctx)
}

func (s *adminService) RegisterSite(ctx context.Context, site *model.Site) error {
	return s.siteRepo.Create(ctx, site)
}

func (s *adminService) RegisterLocation(ctx context.Context, loc *model.Location) error {
	return s.siteRepo.CreateLocation(ctx, loc)
}

func (s *adminService) RegisterAsset(ctx context.Context, asset *model.Asset) error {
	return s.siteRepo.CreateAsset(ctx, asset)
}

func (s *adminService) CreateTaskTemplate(ctx context.Context, task *model.Task) error {
	return s.taskRepo.Create(ctx, task)
}

func (s *adminService) ListUsers(ctx context.Context) ([]*model.User, error) {
	return s.userRepo.List(ctx)
}

func (s *adminService) CreateRole(ctx context.Context, role *model.Role) error {
	return s.userRepo.CreateRole(ctx, role)
}

func (s *adminService) ListRoles(ctx context.Context) ([]*model.Role, error) {
	return s.userRepo.ListRoles(ctx)
}
