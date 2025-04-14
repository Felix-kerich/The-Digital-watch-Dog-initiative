package models

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

// TransactionType defines the possible transaction types
type TransactionType string

const (
	// TransactionAllocation represents funds being allocated from treasury to departments
	TransactionAllocation TransactionType = "ALLOCATION"
	// TransactionDisbursement represents funds being disbursed to projects/contractors
	TransactionDisbursement TransactionType = "DISBURSEMENT"
	// TransactionExpenditure represents spending of funds by responsible entities
	TransactionExpenditure TransactionType = "EXPENDITURE"
	// TransactionReturns represents unused funds being returned
	TransactionReturns TransactionType = "RETURNS"
)

// TransactionStatus defines the possible statuses of a transaction
type TransactionStatus string

const (
	// TransactionPending indicates a transaction that is proposed but not confirmed
	TransactionPending TransactionStatus = "PENDING"
	// TransactionApproved indicates a transaction that has been approved by required authorities
	TransactionApproved TransactionStatus = "APPROVED"
	// TransactionRejected indicates a transaction that was rejected
	TransactionRejected TransactionStatus = "REJECTED"
	// TransactionCompleted indicates a transaction that has been executed
	TransactionCompleted TransactionStatus = "COMPLETED"
	// TransactionFlagged indicates a transaction that has been flagged by the AI for potential irregularities
	TransactionFlagged TransactionStatus = "FLAGGED"
)

// Transaction represents a financial transaction between entities
type Transaction struct {
	ID                string            `json:"id" gorm:"primaryKey;type:char(36)"`
	TransactionType   TransactionType   `json:"transactionType" gorm:"index"`
	Amount            decimal.Decimal   `json:"amount" gorm:"type:decimal(20,2)"`
	Currency          string            `json:"currency" gorm:"default:'KES'"`
	Description       string            `json:"description"`
	Status            TransactionStatus `json:"status" gorm:"index"`
	AiFlagged         bool              `json:"aiFlagged" gorm:"default:false;index"`
	AiConfidence      decimal.Decimal   `json:"aiConfidence" gorm:"type:decimal(5,2)"`
	AiReasonCategory  string            `json:"aiReasonCategory"`
	AiReasonDetails   string            `json:"aiReasonDetails"`
	FundID            string            `json:"fundId" gorm:"type:char(36);index"`
	Fund              *Fund             `json:"fund" gorm:"foreignKey:FundID"`
	BudgetLineItemID  string            `json:"budgetLineItemId" gorm:"type:char(36);index"`
	BudgetLineItem    *BudgetLineItem   `json:"budgetLineItem" gorm:"foreignKey:BudgetLineItemID"`
	SourceID          string            `json:"sourceId" gorm:"type:char(36);index"`
	Source            *Entity           `json:"source" gorm:"foreignKey:SourceID"`
	DestinationID     string            `json:"destinationId" gorm:"type:char(36);index"`
	Destination       *Entity           `json:"destination" gorm:"foreignKey:DestinationID"`
	BlockchainTxHash  string            `json:"blockchainTxHash"`
	BlockchainStatus  string            `json:"blockchainStatus"`
	ApprovedByID      *string           `json:"approvedById" gorm:"type:char(36);null"`
	ApprovedBy        *User             `json:"approvedBy" gorm:"foreignKey:ApprovedByID"`
	RejectedByID      *string           `json:"rejectedById" gorm:"type:char(36);null"`
	RejectedBy        *User             `json:"rejectedBy" gorm:"foreignKey:RejectedByID"`
	RejectionReason   string            `json:"rejectionReason"`
	ReviewedByAuditor bool              `json:"reviewedByAuditor" gorm:"default:false"`
	ReviewedByID      *string           `json:"reviewedById" gorm:"type:char(36);null"`
	ReviewedBy        *User             `json:"reviewedBy" gorm:"foreignKey:ReviewedByID"`
	CreatedByID       *string           `json:"createdById" gorm:"type:char(36);null"`
	CreatedBy         *User             `json:"createdBy" gorm:"foreignKey:CreatedByID"`
	CreatedAt         time.Time         `json:"createdAt" gorm:"index"`
	UpdatedAt         time.Time         `json:"updatedAt"`
	DeletedAt         gorm.DeletedAt    `json:"deletedAt" gorm:"index"`
}

// BeforeCreate will set a UUID rather than numeric ID
func (t *Transaction) BeforeCreate(tx *gorm.DB) (err error) {
	t.ID = uuid.NewString()
	return
}

// FundCategory defines the type of fund
type FundCategory string

const (
	// FundDevelopment is for development projects
	FundDevelopment FundCategory = "DEVELOPMENT"
	// FundRecurrent is for recurrent expenditure
	FundRecurrent FundCategory = "RECURRENT"
	// FundEmergency is for emergency funds
	FundEmergency FundCategory = "EMERGENCY"
	// FundSpecial is for special purpose funds
	FundSpecial FundCategory = "SPECIAL"
)

// BudgetLineItem represents a specific budget allocation within a fund
type BudgetLineItem struct {
	ID          string          `json:"id" gorm:"primaryKey;type:char(36)"`
	Name        string          `json:"name"`
	Description string          `json:"description"`
	Code        string          `json:"code" gorm:"unique"`
	Amount      decimal.Decimal `json:"amount" gorm:"type:decimal(20,2)"`
	Utilized    decimal.Decimal `json:"utilized" gorm:"type:decimal(20,2);default:0"`
	FundID      string          `json:"fundId" gorm:"type:char(36);index"`
	CreatedByID string          `json:"createdById" gorm:"type:char(36)"`
	CreatedAt   time.Time       `json:"createdAt"`
	UpdatedAt   time.Time       `json:"updatedAt"`
	DeletedAt   gorm.DeletedAt  `json:"deletedAt" gorm:"index"`
}

// BeforeCreate will set a UUID rather than numeric ID
func (b *BudgetLineItem) BeforeCreate(tx *gorm.DB) (err error) {
	b.ID = uuid.NewString()
	return
}
