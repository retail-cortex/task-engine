package model

import (
	"time"
)

// Site represents a physical facility or storefront (e.g. store, warehouse).
type Site struct {
	ID             string     `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	OrganizationID string     `gorm:"type:uuid;not null;index"`
	Name           string     `gorm:"type:varchar(255);not null"`
	SiteType       string     `gorm:"type:varchar(50);not null;default:'STORE'"` // e.g. STORE, WAREHOUSE
	Address        string     `gorm:"type:text"`
	Latitude       float64    `gorm:"type:decimal(10,8)"`
	Longitude      float64    `gorm:"type:decimal(11,8)"`
	AltitudeMeters float64    `gorm:"type:decimal(8,2);not null;default:0.0"` // Elevation above sea level
	ICAOCode       string     `gorm:"type:varchar(10);not null;default:''"`   // Weather/Airport reporting station (e.g. KSFO)
	TimeZone       string     `gorm:"type:varchar(50);not null;default:'UTC'"`
	Metadata       JSONB      `gorm:"type:jsonb;not null;default:'{}'"`
	CreatedAt      time.Time  `gorm:"not null;default:now()"`
	UpdatedAt      time.Time  `gorm:"not null;default:now()"`
	Version        int        `gorm:"not null;default:1"`
	Locations      []Location `gorm:"foreignKey:SiteID;constraint:OnDelete:CASCADE"`
}
