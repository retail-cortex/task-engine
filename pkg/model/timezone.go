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
