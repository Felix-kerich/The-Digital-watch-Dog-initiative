package models

import (
    "time"

    "github.com/google/uuid"
    "gorm.io/gorm"
)

// SmartContractRecord stores events or logs generated from blockchain interactions.
type SmartContractRecord struct {
    ID             string    `json:"id" gorm:"primaryKey;type:char(36)"`
    TransactionID  string    `json:"transactionId" gorm:"type:char(36)"`
    EventType      string    `json:"eventType"` // e.g., "CREATED", "UPDATED"
    TxHash         string    `json:"txHash"`
    EventTimestamp time.Time `json:"eventTimestamp"`
    Details        string    `json:"details"`
    CreatedAt      time.Time `json:"createdAt"`
}

func (s *SmartContractRecord) BeforeCreate(tx *gorm.DB) (err error) {
    s.ID = uuid.NewString()
    return
}
