package model

import (
	"time"
)

// Organization represents a corporate entity or tenant.
type Organization struct {
	ID        string         `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	ParentID  *string        `gorm:"type:uuid;index;default:null"` // Support parent-child hierarchies
	Name      string         `gorm:"type:varchar(255);not null"`
	Metadata  JSONB          `gorm:"type:jsonb;not null;default:'{}'"`
	CreatedAt time.Time      `gorm:"not null;default:now()"`
	UpdatedAt time.Time      `gorm:"not null;default:now()"`
	Sites     []Site         `gorm:"foreignKey:OrganizationID;constraint:OnDelete:CASCADE"`
	Users     []User         `gorm:"many2many:user_organizations;constraint:OnDelete:CASCADE"`
	Parent    *Organization  `gorm:"foreignKey:ParentID;references:ID"`
	Children  []Organization `gorm:"foreignKey:ParentID;references:ID;constraint:OnDelete:CASCADE"`
}

// UserOrganization is the explicit join model mapping Users to Organizations.
type UserOrganization struct {
	UserID         string `gorm:"type:uuid;primaryKey;column:user_id"`
	OrganizationID string `gorm:"type:uuid;primaryKey;column:organization_id"`
}
