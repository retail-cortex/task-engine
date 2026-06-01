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

// ShiftAgentSession holds long-context conversation history and system context for the Gemini ADK agent.
type ShiftAgentSession struct {
	ID              string    `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	AssigneeID      string    `gorm:"column:assignee_id;type:uuid;not null;uniqueIndex:idx_assignee_shift,priority:1"`
	ShiftInstanceID string    `gorm:"column:shift_instance_id;type:uuid;not null;uniqueIndex:idx_assignee_shift,priority:2"`
	MessageHistory  JSONB     `gorm:"type:jsonb;not null;default:'[]'"`
	SessionContext  JSONB     `gorm:"type:jsonb;not null;default:'{}'"`
	Status          string    `gorm:"type:varchar(50);not null;default:'ACTIVE'"`
	CreatedAt       time.Time `gorm:"not null;default:now()"`
	UpdatedAt       time.Time `gorm:"not null;default:now()"`
	Version         int       `gorm:"not null;default:1"`
}
