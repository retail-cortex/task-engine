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

package api

import (
	"github.com/rmcguinness/gemini_task_engine/pkg/model"
)

type UserDTO struct {
	ID            string            `json:"id"`
	Name          string            `json:"name"`
	Email         string            `json:"email"`
	Roles         []RoleDTO         `json:"roles,omitempty"`
	Organizations []OrganizationDTO `json:"organizations,omitempty"`
	Sites         []SiteDTO         `json:"sites,omitempty"`
}

type RoleDTO struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type OrganizationDTO struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type SiteDTO struct {
	ID             string `json:"id"`
	Name           string `json:"name"`
	OrganizationID string `json:"organization_id"`
}

func toUserDTO(u *model.User) UserDTO {
	var roles []RoleDTO
	for _, r := range u.Roles {
		roles = append(roles, RoleDTO{ID: r.ID, Name: r.Name})
	}
	var orgs []OrganizationDTO
	for _, o := range u.Organizations {
		orgs = append(orgs, OrganizationDTO{ID: o.ID, Name: o.Name})
	}
	var sites []SiteDTO
	for _, s := range u.Sites {
		sites = append(sites, SiteDTO{ID: s.ID, Name: s.Name, OrganizationID: s.OrganizationID})
	}
	return UserDTO{
		ID:            u.ID,
		Name:          u.Name,
		Email:         u.Email,
		Roles:         roles,
		Organizations: orgs,
		Sites:         sites,
	}
}

func toSiteDTO(s *model.Site) SiteDTO {
	return SiteDTO{
		ID:             s.ID,
		Name:           s.Name,
		OrganizationID: s.OrganizationID,
	}
}
