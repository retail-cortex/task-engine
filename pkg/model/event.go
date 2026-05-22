package model

import (
	"time"
)

// Event defines a scheduled operational target or workload boundaries.
type Event struct {
	ID          string    `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	OrganizerID string    `gorm:"type:uuid;not null;index"`
	SiteID      *string   `gorm:"type:uuid;index;default:null"`
	TaskID      *string   `gorm:"type:uuid;index;default:null"`
	Name        string    `gorm:"type:varchar(255);not null"`
	CreatedAt   time.Time `gorm:"not null;default:now()"`
}

// UserEventSchedule holds individual user recurring shift or workload schedules.
type UserEventSchedule struct {
	ID        string    `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	UserID    string    `gorm:"type:uuid;not null;uniqueIndex:idx_user_event,priority:1"`
	EventID   string    `gorm:"type:uuid;not null;uniqueIndex:idx_user_event,priority:2"`
	StartDate time.Time `gorm:"not null"`
	EndDate   time.Time `gorm:"not null"`
	Timezone  string    `gorm:"type:varchar(50);not null"`
	Rrule     *string   `gorm:"type:text;default:null"`
	CreatedAt time.Time `gorm:"not null;default:now()"`
}

// UserEventInstance represents materialized instances of user schedules.
type UserEventInstance struct {
	ID                string    `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	ScheduleID        string    `gorm:"type:uuid;not null;uniqueIndex:idx_schedule_instance,priority:1"`
	InstanceStartDate time.Time `gorm:"not null;uniqueIndex:idx_schedule_instance,priority:2"`
	InstanceEndDate   time.Time `gorm:"not null"`
	EventStatus       string    `gorm:"type:varchar(50);not null;default:'EventScheduled'"`
	CreatedAt         time.Time `gorm:"not null;default:now()"`
}
