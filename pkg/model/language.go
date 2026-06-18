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

// Language represents an extensible database model for languages and locales, aligning BCP 47 with ISO standards.
type Language struct {
	ID           string    `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	Code         LanguageCode `gorm:"type:varchar(15);not null;uniqueIndex"` // BCP 47 code, e.g. "en-US"
	Name         string    `gorm:"type:varchar(100);not null"`            // Human-readable locale name, e.g. "English (United States)"
	LanguageCode string    `gorm:"type:varchar(3);not null"`              // ISO 639-1 code, e.g. "en"
	LanguageName string    `gorm:"type:varchar(100);not null"`            // Human-readable language name, e.g. "English"
	CountryCode  string    `gorm:"type:varchar(3);not null"`              // ISO 3166-1 Alpha-2 country code, e.g. "US"
	CountryName  string    `gorm:"type:varchar(100);not null"`            // Human-readable country name, e.g. "United States"
	TTSSupported bool      `gorm:"not null;default:true"`                 // True if supported by Google Cloud TTS
	CreatedAt    time.Time `gorm:"not null;default:now()"`
	UpdatedAt    time.Time `gorm:"not null;default:now()"`
}
