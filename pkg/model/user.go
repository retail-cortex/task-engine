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

// User represents the system identity mapped to oauth providers and holding profile metadata.
type User struct {
	ID             string              `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	OAuthProvider  string              `gorm:"type:varchar(50);not null;uniqueIndex:idx_users_oauth"`
	OAuthID        string              `gorm:"type:varchar(255);not null;uniqueIndex:idx_users_oauth"`
	Email          string              `gorm:"type:varchar(255);not null"`
	Name           string              `gorm:"type:varchar(255)"`
	Metadata       JSONB               `gorm:"type:jsonb;not null;default:'{}'"`
	CreatedAt      time.Time           `gorm:"not null;default:now()"`
	UpdatedAt      time.Time           `gorm:"not null;default:now()"`
	Version        int                 `gorm:"not null;default:1"`
	Roles          []Role              `gorm:"many2many:user_roles;constraint:OnDelete:CASCADE"`
	Sites          []Site              `gorm:"many2many:user_sites;constraint:OnDelete:CASCADE"`
	Organizations  []Organization      `gorm:"many2many:user_organizations;constraint:OnDelete:CASCADE"`
	Certifications []UserCertification `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE"`
}

// Role represents standard organizational roles (e.g., Shift Supervisor, Regional Manager).
type Role struct {
	ID          string    `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	Name        string    `gorm:"type:varchar(255);not null;unique"`
	Description string    `gorm:"type:text"`
	CreatedAt   time.Time `gorm:"not null;default:now()"`
}

// UserRole is the explicit join model mapping Users to Roles.
type UserRole struct {
	UserID string `gorm:"type:uuid;primaryKey"`
	RoleID string `gorm:"type:uuid;primaryKey"`
}

// UserSite defines a user's assignment to a physical retail or corporate site.
type UserSite struct {
	ID        string    `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	UserID    string    `gorm:"type:uuid;not null;uniqueIndex:idx_user_site,priority:1"`
	SiteID    string    `gorm:"type:uuid;not null;uniqueIndex:idx_user_site,priority:2"`
	IsPrimary bool      `gorm:"not null;default:false"`
	Metadata  JSONB     `gorm:"type:jsonb;not null;default:'{}'"`
	CreatedAt time.Time `gorm:"not null;default:now()"`
}

// UserCertification links users to completed and active credentials or certifications.
type UserCertification struct {
	ID              string     `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	UserID          string     `gorm:"type:uuid;not null;uniqueIndex:idx_user_certification,priority:1"`
	CertificationID string     `gorm:"type:uuid;not null;uniqueIndex:idx_user_certification,priority:2"`
	IssuedDate      time.Time  `gorm:"not null"`
	ExpirationDate  *time.Time `gorm:"default:null"`
	Status          string     `gorm:"type:varchar(50);not null;default:'ACTIVE'"`
	CreatedAt       time.Time  `gorm:"not null;default:now()"`
	Version         int        `gorm:"not null;default:1"`
}
