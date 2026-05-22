package model

import (
	"time"
)

// UserAvailability holds workforce shift scheduling patterns, constraints, and availability types using RRULEs.
type UserAvailability struct {
	ID                 string     `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	UserID             string     `gorm:"type:uuid;not null;index"`
	SiteID             *string    `gorm:"type:uuid;index;default:null"`
	Name               string     `gorm:"type:varchar(255)"`
	EffectiveStartDate time.Time  `gorm:"not null"`
	EffectiveEndDate   *time.Time `gorm:"default:null"`
	Timezone           string     `gorm:"type:varchar(50);not null"`
	ShiftStartTime     string     `gorm:"type:time;not null"`
	ShiftEndTime       string     `gorm:"type:time;not null"`
	Rrule              string     `gorm:"type:text;not null"`
	AvailabilityType   string     `gorm:"type:varchar(50);not null;default:'AVAILABLE'"`
	Metadata           JSONB      `gorm:"type:jsonb;not null;default:'{}'"`
	CreatedAt          time.Time  `gorm:"not null;default:now()"`
	UpdatedAt          time.Time  `gorm:"not null;default:now()"`
	Version            int        `gorm:"not null;default:1"`
}
