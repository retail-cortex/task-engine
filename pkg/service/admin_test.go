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
	"testing"

	"github.com/rmcguinness/gemini_task_engine/pkg/model"
	"github.com/rmcguinness/gemini_task_engine/pkg/persistence"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

func setupAdminServiceWithDryRunDB(t *testing.T) AdminService {
	db, err := gorm.Open(dummyDialector{}, &gorm.Config{
		DryRun: true,
	})
	assert.NoError(t, err)

	return NewAdminService(
		persistence.NewUserRepository(db),
		persistence.NewOrganizationRepository(db),
		persistence.NewSiteRepository(db),
		persistence.NewTaskRepository(db),
	)
}

func TestAdminService_AllMethods(t *testing.T) {
	svc := setupAdminServiceWithDryRunDB(t)
	ctx := context.Background()

	// 1. Users
	u := &model.User{ID: "u1", Email: "admin@google.com"}
	assert.NoError(t, svc.RegisterUser(ctx, u))
	assert.NoError(t, svc.UpdateUser(ctx, u))
	assert.NoError(t, svc.AssignRole(ctx, "u1", "r1"))
	_, err := svc.GetUserByID(ctx, "u1")
	assert.NoError(t, err)
	_, err = svc.FindUserByOAuth(ctx, "google", "100")
	assert.NoError(t, err)
	_, err = svc.ListUsers(ctx)
	assert.NoError(t, err)
	_, err = svc.ListUsersRange(ctx, 0, 10)
	assert.NoError(t, err)
	assert.NoError(t, svc.DeleteUser(ctx, "u1"))

	// 2. Roles
	r := &model.Role{ID: "r1", Name: "Admin"}
	assert.NoError(t, svc.CreateRole(ctx, r))
	assert.NoError(t, svc.UpdateRole(ctx, r))
	_, err = svc.GetRoleByID(ctx, "r1")
	assert.NoError(t, err)
	_, err = svc.ListRoles(ctx)
	assert.NoError(t, err)
	_, err = svc.ListRolesRange(ctx, 0, 10)
	assert.NoError(t, err)
	assert.NoError(t, svc.DeleteRole(ctx, "r1"))

	// 3. Organizations
	org := &model.Organization{ID: "org1", Name: "Retail Org"}
	assert.NoError(t, svc.RegisterOrganization(ctx, org))
	assert.NoError(t, svc.UpdateOrganization(ctx, org))
	assert.NoError(t, svc.AssignUserToOrganization(ctx, "org1", "u1"))
	_, err = svc.GetOrganizationByID(ctx, "org1")
	assert.NoError(t, err)
	_, err = svc.ListOrganizations(ctx)
	assert.NoError(t, err)
	_, err = svc.ListOrganizationsRange(ctx, 0, 10)
	assert.NoError(t, err)
	assert.NoError(t, svc.DeleteOrganization(ctx, "org1"))

	// 4. Sites
	site := &model.Site{ID: "site1", Name: "Store 1", OrganizationID: "org1"}
	assert.NoError(t, svc.RegisterSite(ctx, site))
	assert.NoError(t, svc.UpdateSite(ctx, site))
	_, err = svc.GetSiteByID(ctx, "site1")
	assert.NoError(t, err)
	_, err = svc.ListSites(ctx)
	assert.NoError(t, err)
	_, err = svc.ListSitesRange(ctx, 0, 10)
	assert.NoError(t, err)
	assert.NoError(t, svc.DeleteSite(ctx, "site1"))

	// 5. Locations
	loc := &model.Location{ID: "loc1", SiteID: "site1", Name: "Aisle A"}
	assert.NoError(t, svc.RegisterLocation(ctx, loc))
	assert.NoError(t, svc.UpdateLocation(ctx, loc))
	_, err = svc.GetLocationByID(ctx, "loc1")
	assert.NoError(t, err)
	_, err = svc.ListLocations(ctx)
	assert.NoError(t, err)
	_, err = svc.ListLocationsRange(ctx, 0, 10)
	assert.NoError(t, err)
	assert.NoError(t, svc.DeleteLocation(ctx, "loc1"))

	// 6. Assets
	asset := &model.Asset{ID: "a1", LocationID: "loc1", Name: "Scanner"}
	assert.NoError(t, svc.RegisterAsset(ctx, asset))
	assert.NoError(t, svc.UpdateAsset(ctx, asset))
	_, err = svc.GetAssetByID(ctx, "a1")
	assert.NoError(t, err)
	_, err = svc.ListAssets(ctx)
	assert.NoError(t, err)
	_, err = svc.ListAssetsRange(ctx, 0, 10)
	assert.NoError(t, err)
	assert.NoError(t, svc.DeleteAsset(ctx, "a1"))

	// 7. Task Templates
	task := &model.Task{ID: "t1", Name: "CHECK_TEMP"}
	assert.NoError(t, svc.CreateTaskTemplate(ctx, task))
	assert.NoError(t, svc.UpdateTaskTemplate(ctx, task))
	_, err = svc.GetTaskTemplateByID(ctx, "t1")
	assert.NoError(t, err)
	_, err = svc.ListTaskTemplates(ctx)
	assert.NoError(t, err)
	_, err = svc.ListTaskTemplatesRange(ctx, 0, 10)
	assert.NoError(t, err)
	assert.NoError(t, svc.DeleteTaskTemplate(ctx, "t1"))
}
