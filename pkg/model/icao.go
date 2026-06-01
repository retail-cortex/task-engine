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

import "time"

// ICAOCode represents International Civil Aviation Organization codes used for weather station lookup.
type ICAOCode struct {
	ID                   string    `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	CountryID            string    `gorm:"type:varchar(10);not null"`
	CountrySubdivisionID string    `gorm:"type:varchar(10)"`
	LocationName         string    `gorm:"type:varchar(255);not null"`
	StationName          string    `gorm:"type:varchar(255);not null"`
	Type                 string    `gorm:"type:varchar(50)"`
	StationKey           string    `gorm:"type:varchar(50)"`
	Status               string    `gorm:"type:varchar(50)"`
	ICAO                 string    `gorm:"type:varchar(10);not null;uniqueIndex"` // e.g. KSFO
	NationalID           string    `gorm:"type:varchar(50);not null"`
	Wmo                  string    `gorm:"type:varchar(50)"`
	Ghcn                 string    `gorm:"type:varchar(50)"`
	Special              string    `gorm:"type:varchar(50)"`
	Latitude             string    `gorm:"type:varchar(50);not null"`
	Longitude            string    `gorm:"type:varchar(50);not null"`
	ElevationInMeters    string    `gorm:"type:varchar(50)"`
	TimeZone             string    `gorm:"type:varchar(50)"`
	CreatedAt            time.Time `gorm:"not null;default:now()"`
	UpdatedAt            time.Time `gorm:"not null;default:now()"`
}
