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

// Task represents a master task template definition in the system.
type Task struct {
	ID                       string             `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	ParentTaskID             *string            `gorm:"type:uuid;index;default:null"`
	Name                     string             `gorm:"type:varchar(255);not null"`
	Description              string             `gorm:"type:text"`
	TaskType                 string             `gorm:"type:varchar(50);not null;default:'STANDARD'"`
	TargetRoleID             *string            `gorm:"type:uuid;index;default:null"`
	Priority                 int                `gorm:"not null;default:3"`
	StepOrder                int                `gorm:"default:0"`
	EstimatedDurationMinutes int                `gorm:"default:0"`
	ChecklistTemplate        JSONB              `gorm:"type:jsonb;default:'[]'"`
	Metadata                 JSONB              `gorm:"type:jsonb;not null;default:'{}'"`
	CreatedAt                time.Time          `gorm:"not null;default:now()"`
	UpdatedAt                time.Time          `gorm:"not null;default:now()"`
	Version                  int                `gorm:"not null;default:1"`
	Assets                   []TaskAsset        `gorm:"foreignKey:TaskID;constraint:OnDelete:CASCADE"`
	ApprovalRules            []TaskApprovalRule `gorm:"foreignKey:TaskID;constraint:OnDelete:CASCADE"`
	SOPs                     []SOP              `gorm:"many2many:task_sops;constraint:OnDelete:CASCADE"`
}

// TaskAsset defines equipment/material requirements for executing a task.
type TaskAsset struct {
	ID               string `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	TaskID           string `gorm:"type:uuid;not null;uniqueIndex:idx_task_asset,priority:1"`
	AssetID          string `gorm:"type:uuid;not null;uniqueIndex:idx_task_asset,priority:2"`
	IsConsumable     bool   `gorm:"not null;default:false"`
	QuantityRequired int    `gorm:"default:1"`
	IsHardBlocker    bool   `gorm:"not null;default:false"`
}

// TaskApprovalRule configures supervisor/role checking rules for task templates.
type TaskApprovalRule struct {
	ID             string    `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	TaskID         string    `gorm:"type:uuid;not null;index"`
	RequiredRoleID string    `gorm:"type:uuid;not null;index"`
	Timing         string    `gorm:"type:varchar(50);not null;default:'POST_EXECUTION'"`
	IsStrict       bool      `gorm:"not null;default:true"`
	CreatedAt      time.Time `gorm:"not null;default:now()"`
}

// TaskSOP is the explicit join model mapping Tasks to SOPs.
type TaskSOP struct {
	TaskID string `gorm:"type:uuid;primaryKey"`
	SOPID  string `gorm:"type:uuid;primaryKey"`
}
