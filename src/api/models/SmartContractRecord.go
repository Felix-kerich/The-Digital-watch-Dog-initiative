package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// EventType represents the type of blockchain event
type EventType string

const (
	// Transaction Events
	EventCreated   EventType = "CREATED"
	EventApproved  EventType = "APPROVED"
	EventRejected  EventType = "REJECTED"
	EventCompleted EventType = "COMPLETED"
	EventFlagged   EventType = "FLAGGED"

	// AI Events
	EventAIFlagged EventType = "AI_FLAGGED"
	EventAICleared EventType = "AI_CLEARED"
	EventAIWarning EventType = "AI_WARNING"

	// Budget Events
	EventBudgetOverrun   EventType = "BUDGET_OVERRUN"
	EventHighUtilization EventType = "HIGH_UTILIZATION"
)

// SmartContractRecord stores events or logs generated from blockchain interactions.
type SmartContractRecord struct {
	ID             string    `json:"id" gorm:"primaryKey;type:char(36)"`
	TransactionID  string    `json:"transactionId" gorm:"type:char(36);index"`
	EventType      EventType `json:"eventType" gorm:"index"`
	TxHash         string    `json:"txHash" gorm:"index"`
	EventTimestamp time.Time `json:"eventTimestamp"`
	Details        string    `json:"details"`
	AnomalyScore   float64   `json:"anomalyScore,omitempty"` // AI confidence score
	BlockNumber    uint64    `json:"blockNumber"`
	GasUsed        uint64    `json:"gasUsed"`
	CreatedAt      time.Time `json:"createdAt"`
}

func (s *SmartContractRecord) BeforeCreate(tx *gorm.DB) (err error) {
	s.ID = uuid.NewString()
	return
}
