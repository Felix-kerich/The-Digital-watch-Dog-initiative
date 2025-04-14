package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Entity represents an organization or department that can be source or destination of funds
type Entity struct {
	ID           string         `json:"id" gorm:"primaryKey;type:char(36)"`
	Name         string         `json:"name"`
	Type         string         `json:"type" gorm:"index"` // Ministry, Department, Agency, Contractor, etc.
	Code         string         `json:"code" gorm:"unique"`
	Description  string         `json:"description"`
	IsGovernment bool           `json:"isGovernment" gorm:"index"`
	Location     string         `json:"location"`
	ContactInfo  string         `json:"contactInfo"`
	IsActive     bool           `json:"isActive" gorm:"default:true;index"`
	CreatedByID  string         `json:"createdById" gorm:"type:char(36)"`
	CreatedAt    time.Time      `json:"createdAt"`
	UpdatedAt    time.Time      `json:"updatedAt"`
	DeletedAt    gorm.DeletedAt `json:"deletedAt" gorm:"index"`
}

// BeforeCreate will set a UUID rather than numeric ID
func (e *Entity) BeforeCreate(tx *gorm.DB) (err error) {
	e.ID = uuid.NewString()
	return
}
