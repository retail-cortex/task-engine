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
