package model

import (
	"time"
)

// TaskExecution represents a running instance of a task template.
type TaskExecution struct {
	ID                      string     `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	TaskTemplateID          string     `gorm:"column:task_template_id;type:uuid;not null;index"`
	ParentExecutionID       *string    `gorm:"column:parent_execution_id;type:uuid;index;default:null"`
	ExecutionType           string     `gorm:"type:varchar(50);not null;default:'STANDARD'"`
	SubjectExecutionID      *string    `gorm:"column:subject_execution_id;type:uuid;index;default:null"`
	InitiatorID             *string    `gorm:"column:initiator_id;type:uuid;index;default:null"`
	AssigneeID              *string    `gorm:"column:assignee_id;type:uuid;index;default:null"`
	EventInstanceID         string     `gorm:"column:event_instance_id;type:uuid;not null;index"`
	Status                  string     `gorm:"type:varchar(50);not null;default:'PENDING'"`
	Priority                int        `gorm:"not null;default:3"`
	DueAt                   *time.Time `gorm:"default:null"`
	PrerequisiteExecutionID *string    `gorm:"column:prerequisite_execution_id;type:uuid;index;default:null"`
	Decision                *string    `gorm:"type:varchar(50);default:null"`
	CompletedAt             *time.Time `gorm:"default:null"`
	ChecklistState          JSONB      `gorm:"type:jsonb;default:'{}'"`
	OverrideFlags           JSONB      `gorm:"type:jsonb;not null;default:'{}'"`
	CreatedAt               time.Time  `gorm:"not null;default:now()"`
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
	ProposedAssigneeID string    `gorm:"column:proposed_assignee_id;type:uuid;not null;index"`
	Status             string    `gorm:"type:varchar(50);not null;default:'PENDING'"`
	CreatedAt          time.Time `gorm:"not null;default:now()"`
	UpdatedAt          time.Time `gorm:"not null;default:now()"`
	Version            int       `gorm:"not null;default:1"`
}
