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
	"encoding/json"
	"time"
)

// User represents the system identity mapped to oauth providers and holding profile metadata.
type User struct {
	ID             string              `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	OAuthProvider  string              `gorm:"type:varchar(50);not null;uniqueIndex:idx_users_oauth"`
	OAuthID        string              `gorm:"type:varchar(255);not null;uniqueIndex:idx_users_oauth"`
	Email          string              `gorm:"type:varchar(255);not null"`
	Name           string              `gorm:"type:varchar(255)"`
	Metadata         JSONB               `gorm:"type:jsonb;not null;default:'{}'"`
	ConsentRecording []byte              `gorm:"type:bytea;default:null"`
	PreferredLanguageID *string             `gorm:"type:uuid;index;default:null"`
	PreferredLanguage   *Language           `gorm:"foreignKey:PreferredLanguageID;constraint:OnDelete:SET NULL"`
	VoiceGenderPreference string            `gorm:"type:varchar(50);not null;default:'FEMALE'"`
	VoiceNamePreference   string            `gorm:"type:varchar(100);not null;default:'en-US-Journey-F'"`
	ClonedVoiceKey        string            `gorm:"type:varchar(255);not null;default:''"`
	CreatedAt        time.Time           `gorm:"not null;default:now()"`
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

// Gender represents the user's gender for voice and translation styling.
type Gender string

const (
	GenderMale   Gender = "Male"
	GenderFemale Gender = "Female"
)

func (u *User) getMetadataMap() (map[string]interface{}, error) {
	var meta map[string]interface{}
	if len(u.Metadata) == 0 || string(u.Metadata) == "{}" || string(u.Metadata) == "null" {
		return make(map[string]interface{}), nil
	}
	if err := json.Unmarshal(u.Metadata, &meta); err != nil {
		return nil, err
	}
	return meta, nil
}

func (u *User) setMetadataMap(meta map[string]interface{}) error {
	bytes, err := json.Marshal(meta)
	if err != nil {
		return err
	}
	u.Metadata = JSONB(bytes)
	return nil
}

// GetGender retrieves the Gender enum from the user's metadata.
func (u *User) GetGender() (Gender, error) {
	meta, err := u.getMetadataMap()
	if err != nil {
		return "", err
	}
	if val, ok := meta["gender"]; ok {
		if str, ok := val.(string); ok {
			return Gender(str), nil
		}
	}
	return "", nil
}

// SetGender stores the Gender enum inside the user's metadata.
func (u *User) SetGender(g Gender) error {
	meta, err := u.getMetadataMap()
	if err != nil {
		return err
	}
	meta["gender"] = string(g)
	return u.setMetadataMap(meta)
}

// GetAssistantVoice retrieves the AssistantVoice string from the user's metadata.
func (u *User) GetAssistantVoice() (string, error) {
	meta, err := u.getMetadataMap()
	if err != nil {
		return "", err
	}
	if val, ok := meta["assistant_voice"]; ok {
		if str, ok := val.(string); ok {
			return str, nil
		}
	}
	return "", nil
}

// SetAssistantVoice stores the AssistantVoice string inside the user's metadata.
func (u *User) SetAssistantVoice(voice string) error {
	meta, err := u.getMetadataMap()
	if err != nil {
		return err
	}
	meta["assistant_voice"] = voice
	return u.setMetadataMap(meta)
}

// GetMyVoice retrieves the MyVoice string from the user's metadata.
func (u *User) GetMyVoice() (string, error) {
	meta, err := u.getMetadataMap()
	if err != nil {
		return "", err
	}
	if val, ok := meta["my_voice"]; ok {
		if str, ok := val.(string); ok {
			return str, nil
		}
	}
	return "", nil
}

// SetMyVoice stores the MyVoice string inside the user's metadata.
func (u *User) SetMyVoice(voice string) error {
	meta, err := u.getMetadataMap()
	if err != nil {
		return err
	}
	meta["my_voice"] = voice
	return u.setMetadataMap(meta)
}

// GetCustomVoice retrieves the CustomVoice boolean from the user's metadata.
func (u *User) GetCustomVoice() (bool, error) {
	meta, err := u.getMetadataMap()
	if err != nil {
		return false, err
	}
	if val, ok := meta["custom_voice"]; ok {
		if b, ok := val.(bool); ok {
			return b, nil
		}
	}
	return false, nil
}

// SetCustomVoice stores the CustomVoice boolean inside the user's metadata.
func (u *User) SetCustomVoice(custom bool) error {
	meta, err := u.getMetadataMap()
	if err != nil {
		return err
	}
	meta["custom_voice"] = custom
	return u.setMetadataMap(meta)
}

// LanguageCode represents a BCP 47 language and locale code.
type LanguageCode string

const (
	LanguageEnglish   LanguageCode = "en-US"
	LanguageSpanish   LanguageCode = "es-ES"
	LanguageSpanishMX LanguageCode = "es-MX"
	LanguageFrench    LanguageCode = "fr-FR"
	LanguageGerman    LanguageCode = "de-DE"
	LanguageChinese   LanguageCode = "zh-CN"
)

// SetPreferredLanguage associates a Language with the User.
func (u *User) SetPreferredLanguage(lang *Language) {
	u.PreferredLanguage = lang
	if lang != nil {
		u.PreferredLanguageID = &lang.ID
	} else {
		u.PreferredLanguageID = nil
	}
}

// GetPreferredLanguage returns the associated Language relation, if loaded.
func (u *User) GetPreferredLanguage() *Language {
	return u.PreferredLanguage
}

