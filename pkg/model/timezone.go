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

// TimeZone represents a geographic time zone metadata.
type TimeZone struct {
	ID             string    `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	Name           string    `gorm:"type:varchar(100);not null;uniqueIndex"` // e.g. America/New_York
	TimezoneOffset string    `gorm:"type:varchar(50);not null"`             // e.g. UTC-05:00
	CreatedAt      time.Time `gorm:"not null;default:now()"`
	UpdatedAt      time.Time `gorm:"not null;default:now()"`
}
