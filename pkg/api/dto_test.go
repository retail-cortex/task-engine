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
	"testing"

	"github.com/rmcguinness/gemini_task_engine/pkg/model"
	"github.com/stretchr/testify/assert"
)

func TestDTOConversion(t *testing.T) {
	site := &model.Site{ID: "s1", Name: "Store 1", OrganizationID: "org1"}
	sDTO := toSiteDTO(site)
	assert.Equal(t, "s1", sDTO.ID)
	assert.Equal(t, "Store 1", sDTO.Name)

	u := &model.User{
		ID:    "u1",
		Name:  "Ryan",
		Email: "rmcguinness@google.com",
		Roles: []model.Role{{ID: "r1", Name: "Admin"}},
		Organizations: []model.Organization{{ID: "org1", Name: "Retail Org"}},
		Sites: []model.Site{*site},
	}
	uDTO := toUserDTO(u)
	assert.Equal(t, "u1", uDTO.ID)
	assert.Len(t, uDTO.Roles, 1)
	assert.Len(t, uDTO.Organizations, 1)
	assert.Len(t, uDTO.Sites, 1)
}
