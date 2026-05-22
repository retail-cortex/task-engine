package model

import (
	"time"
)

// Asset represents equipment, machinery, or other physical inventory at a Location.
type Asset struct {
	ID         string    `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	LocationID string    `gorm:"type:uuid;not null;index"`
	Name       string    `gorm:"type:varchar(255);not null"`
	AssetTag   string    `gorm:"type:varchar(100);uniqueIndex"`
	Status     string    `gorm:"type:varchar(50);not null;default:'AVAILABLE'"`
	Metadata   JSONB     `gorm:"type:jsonb;not null;default:'{}'"`
	CreatedAt  time.Time `gorm:"not null;default:now()"`
	UpdatedAt  time.Time `gorm:"not null;default:now()"`
	Version    int       `gorm:"not null;default:1"`
}

// Certification represents the base definition of a certificate, license, or course definition.
type Certification struct {
	ID          string    `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	Name        string    `gorm:"type:varchar(255);not null"`
	Issuer      string    `gorm:"type:varchar(255)"`
	Description string    `gorm:"type:text"`
	CreatedAt   time.Time `gorm:"not null;default:now()"`
	Version     int       `gorm:"not null;default:1"`
}
