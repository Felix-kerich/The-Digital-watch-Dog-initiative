package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// File represents an uploaded file in the system
type File struct {
	ID            string         `json:"id" gorm:"primaryKey;type:char(36)"`
	FileName      string         `json:"fileName"`
	FilePath      string         `json:"filePath"`
	FileSize      int64          `json:"fileSize"`
	FileType      string         `json:"fileType" gorm:"index"`
	ContentType   string         `json:"contentType"`
	TransactionID string         `json:"transactionId" gorm:"type:char(36);index"`
	EntityID      string         `json:"entityId" gorm:"type:char(36);index"`
	FundID        string         `json:"fundId" gorm:"type:char(36);index"`
	UploadedByID  string         `json:"uploadedById" gorm:"type:char(36);index"`
	IsPublic      bool           `json:"isPublic" gorm:"default:false;index"`
	Description   string         `json:"description"`
	CreatedAt     time.Time      `json:"createdAt"`
	UpdatedAt     time.Time      `json:"updatedAt"`
	DeletedAt     gorm.DeletedAt `json:"deletedAt" gorm:"index"`
}

// BeforeCreate will set a UUID rather than numeric ID
func (f *File) BeforeCreate(tx *gorm.DB) (err error) {
	f.ID = uuid.NewString()
	return
}
