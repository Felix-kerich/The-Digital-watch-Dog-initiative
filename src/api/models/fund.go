package models

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

// FundStatus defines the status of a fund
type FundStatus string

const (
	// FundStatusActive represents an active fund
	FundStatusActive FundStatus = "ACTIVE"
	// FundStatusInactive represents an inactive fund
	FundStatusInactive FundStatus = "INACTIVE"
	// FundStatusDeleted represents a deleted fund
	FundStatusDeleted FundStatus = "DELETED"
)

// Fund represents a financial fund in the system
type Fund struct {
	ID          string          `json:"id" gorm:"primaryKey;type:char(36)"`
	Name        string          `json:"name"`
	Description string          `json:"description"`
	Code        string          `json:"code" gorm:"unique"`
	Category    FundCategory    `json:"category" gorm:"index"`
	FiscalYear  string          `json:"fiscalYear" gorm:"index"`
	Amount      decimal.Decimal `json:"amount" gorm:"type:decimal(20,2)"`
	TotalAmount decimal.Decimal `json:"totalAmount" gorm:"type:decimal(20,2)"`
	Allocated   decimal.Decimal `json:"allocated" gorm:"type:decimal(20,2);default:0"`
	Disbursed   decimal.Decimal `json:"disbursed" gorm:"type:decimal(20,2);default:0"`
	Utilized    decimal.Decimal `json:"utilized" gorm:"type:decimal(20,2);default:0"`
	Currency    string          `json:"currency" gorm:"default:'USD'"`
	Status      FundStatus      `json:"status" gorm:"default:'ACTIVE';index"`
	EntityID    string          `json:"entityId" gorm:"type:char(36);index"`
	CreatedByID string          `json:"createdById" gorm:"type:char(36)"`
	CreatedAt   time.Time       `json:"createdAt"`
	UpdatedAt   time.Time       `json:"updatedAt"`
	DeletedAt   gorm.DeletedAt  `json:"deletedAt" gorm:"index"`
}

// BeforeCreate will set a UUID rather than numeric ID
func (f *Fund) BeforeCreate(tx *gorm.DB) (err error) {
	f.ID = uuid.NewString()
	return
}
