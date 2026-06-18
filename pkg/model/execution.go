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

// TaskExecution represents a running instance of a task template.
type TaskExecution struct {
	ID                      string     `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	TaskTemplateID          string     `gorm:"column:task_template_id;type:uuid;not null;index"`
	Task                    Task       `gorm:"foreignKey:TaskTemplateID;constraint:OnUpdate:CASCADE,OnDelete:RESTRICT"`
	ParentExecutionID       *string    `gorm:"column:parent_execution_id;type:uuid;index;default:null"`
	ExecutionType           string     `gorm:"type:varchar(50);not null;default:'STANDARD'"`
	SubjectExecutionID      *string    `gorm:"column:subject_execution_id;type:uuid;index;default:null"`
	InitiatorID             *string    `gorm:"column:initiator_id;type:uuid;index;default:null"`
	AssigneeID              *string    `gorm:"column:assignee_id;type:uuid;index;default:null"`
	Assignee                *User      `gorm:"foreignKey:AssigneeID;constraint:OnUpdate:CASCADE,OnDelete:RESTRICT"`
	EventInstanceID         string     `gorm:"column:event_instance_id;type:uuid;not null;index;index:idx_task_executions_queue,priority:1"`
	Description             string     `gorm:"type:text;default:null"`
	Status                  string     `gorm:"type:varchar(50);not null;default:'PENDING';index:idx_task_executions_status_locked_at,priority:1"`
	Priority                int        `gorm:"not null;default:3;index:idx_task_executions_queue,priority:2"`
	DueAt                   *time.Time `gorm:"default:null"`
	PrerequisiteExecutionID *string    `gorm:"column:prerequisite_execution_id;type:uuid;index;default:null"`
	Decision                *string    `gorm:"type:varchar(50);default:null"`
	StartedAt               *time.Time `gorm:"default:null"`
	PausedAt                *time.Time `gorm:"default:null"`
	TotalPausedSeconds      int        `gorm:"not null;default:0"`
	CompletedAt             *time.Time `gorm:"default:null"`
	ChecklistState          JSONB      `gorm:"type:jsonb;default:'{}'"`
	OverrideFlags           JSONB      `gorm:"type:jsonb;not null;default:'{}'"`
	LockedAt                *time.Time `gorm:"default:null;index:idx_task_executions_status_locked_at,priority:2"`
	LockedBy                *string    `gorm:"type:varchar(255);default:null"`
	RetryCount              int        `gorm:"not null;default:0"`
	MaxRetries              int        `gorm:"not null;default:3"`
	LastError               *string    `gorm:"type:text;default:null"`
	CreatedAt               time.Time  `gorm:"not null;default:now();index:idx_task_executions_queue,priority:3"`
	UpdatedAt               time.Time  `gorm:"not null;default:now()"`
	Version                 int        `gorm:"not null;default:1"`
}

// TaskExecutionAudit stores historical logs of all execution state updates.
type TaskExecutionAudit struct {
	ID              string    `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	TaskExecutionID string    `gorm:"column:task_execution_id;type:uuid;not null;index"`
	ChangedByID     *string   `gorm:"column:changed_by_id;type:uuid;index;default:null"`
	ActionType      string    `gorm:"type:varchar(50);not null"`
	PreviousState   JSONB     `gorm:"type:jsonb"`
	NewState        JSONB     `gorm:"type:jsonb"`
	CreatedAt       time.Time `gorm:"not null;default:now()"`
}

// TaskTrade represents peer-to-peer shift/task handovers.
type TaskTrade struct {
	ID                 string    `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	TaskExecutionID    string    `gorm:"column:task_execution_id;type:uuid;not null;index"`
	InitiatorID        string    `gorm:"column:initiator_id;type:uuid;not null;index"`
	ProposedAssigneeID *string   `gorm:"column:proposed_assignee_id;type:uuid;index;default:null"`
	Status             string    `gorm:"type:varchar(50);not null;default:'PENDING'"`
	CreatedAt          time.Time `gorm:"not null;default:now()"`
	UpdatedAt          time.Time `gorm:"not null;default:now()"`
	Version            int       `gorm:"not null;default:1"`
}
