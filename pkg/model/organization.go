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

package model

import (
	"time"
)

// Organization represents a corporate entity or tenant.
type Organization struct {
	ID        string         `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	ParentID  *string        `gorm:"type:uuid;index;default:null"` // Support parent-child hierarchies
	Name      string         `gorm:"type:varchar(255);not null"`
	Metadata  JSONB          `gorm:"type:jsonb;not null;default:'{}'"`
	CreatedAt time.Time      `gorm:"not null;default:now()"`
	UpdatedAt time.Time      `gorm:"not null;default:now()"`
	Sites     []Site         `gorm:"foreignKey:OrganizationID;constraint:OnDelete:CASCADE"`
	Users     []User         `gorm:"many2many:user_organizations;constraint:OnDelete:CASCADE"`
	Parent    *Organization  `gorm:"foreignKey:ParentID;references:ID"`
	Children  []Organization `gorm:"foreignKey:ParentID;references:ID;constraint:OnDelete:CASCADE"`
}

// UserOrganization is the explicit join model mapping Users to Organizations.
type UserOrganization struct {
	UserID         string `gorm:"type:uuid;primaryKey;column:user_id"`
	OrganizationID string `gorm:"type:uuid;primaryKey;column:organization_id"`
}
