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

// Location represents a non-movable sub-location within a Site (e.g. fixture, shelf, aisle, register).
type Location struct {
	ID                   string     `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	SiteID               string     `gorm:"type:uuid;not null;index"`
	ParentID             *string    `gorm:"type:uuid;index;default:null"` // Supports hierarchical nesting
	Name                 string     `gorm:"type:varchar(255);not null"`
	LocationType         string     `gorm:"type:varchar(50);not null;default:'FIXTURE'"` // e.g. FIXTURE, SHELF, AISLE, REGISTER
	LocationFunctionType string     `gorm:"type:varchar(50);not null;default:'DISPLAY'"` // e.g. DISPLAY, STOCK_POINT, RECEIVING
	X                    float64    `gorm:"type:decimal(10,4);not null;default:0.0"`      // X offset relative to site origin
	Y                    float64    `gorm:"type:decimal(10,4);not null;default:0.0"`      // Y offset relative to site origin
	Z                    float64    `gorm:"type:decimal(10,4);not null;default:0.0"`      // Z offset (local elevation/height from floor)
	Metadata             JSONB      `gorm:"type:jsonb;not null;default:'{}'"`
	CreatedAt            time.Time  `gorm:"not null;default:now()"`
	UpdatedAt            time.Time  `gorm:"not null;default:now()"`
	Parent               *Location  `gorm:"foreignKey:ParentID;references:ID"`
	Children             []Location `gorm:"foreignKey:ParentID;references:ID;constraint:OnDelete:CASCADE"`
}
