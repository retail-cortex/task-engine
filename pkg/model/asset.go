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

// Asset represents equipment, machinery, or other physical inventory at a Location.
type Asset struct {
	ID         string    `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	LocationID string    `gorm:"type:uuid;not null;index"`
	Name       string    `gorm:"type:varchar(255);not null"`
	AssetTag   string    `gorm:"type:varchar(100);uniqueIndex"`
	Status     string    `gorm:"type:varchar(50);not null;default:'AVAILABLE'"`
	Metadata   JSONB     `gorm:"type:jsonb;not null;default:'{}'"`
	CreatedAt  time.Time `gorm:"not null;default:now()"`
	UpdatedAt  time.Time `gorm:"not null;default:now()"`
	Version    int       `gorm:"not null;default:1"`
}

// Certification represents the base definition of a certificate, license, or course definition.
type Certification struct {
	ID          string    `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	Name        string    `gorm:"type:varchar(255);not null"`
	Issuer      string    `gorm:"type:varchar(255)"`
	Description string    `gorm:"type:text"`
	CreatedAt   time.Time `gorm:"not null;default:now()"`
	Version     int       `gorm:"not null;default:1"`
}
