package repository

import (
	"github.com/the-digital-watchdog-initiative/models"
	"github.com/the-digital-watchdog-initiative/utils"
	"gorm.io/gorm"
)

// SmartContractRepository interface defines methods for smart contract record operations
type SmartContractRepository interface {
	Create(record *models.SmartContractRecord) error
	FindByTransactionID(transactionID string) ([]models.SmartContractRecord, error)
	FindByEventType(eventType models.EventType) ([]models.SmartContractRecord, error)
	FindByTxHash(txHash string) (*models.SmartContractRecord, error)
	ListAIFlags(page, limit int) ([]models.SmartContractRecord, int64, error)
}

// GormSmartContractRepository implements SmartContractRepository
type GormSmartContractRepository struct {
	DB *gorm.DB
}

// NewSmartContractRepository creates a new SmartContractRepository
func NewSmartContractRepository() SmartContractRepository {
	return &GormSmartContractRepository{
		DB: utils.DB,
	}
}

// Create adds a new smart contract record
func (r *GormSmartContractRepository) Create(record *models.SmartContractRecord) error {
	return r.DB.Create(record).Error
}

// FindByTransactionID retrieves all records for a transaction ID
func (r *GormSmartContractRepository) FindByTransactionID(transactionID string) ([]models.SmartContractRecord, error) {
	var records []models.SmartContractRecord
	err := r.DB.Where("transaction_id = ?", transactionID).
		Order("event_timestamp DESC").
		Find(&records).Error
	return records, err
}

// FindByEventType retrieves records by event type
func (r *GormSmartContractRepository) FindByEventType(eventType models.EventType) ([]models.SmartContractRecord, error) {
	var records []models.SmartContractRecord
	err := r.DB.Where("event_type = ?", eventType).
		Order("event_timestamp DESC").
		Find(&records).Error
	return records, err
}

// FindByTxHash retrieves a record by transaction hash
func (r *GormSmartContractRepository) FindByTxHash(txHash string) (*models.SmartContractRecord, error) {
	var record models.SmartContractRecord
	err := r.DB.Where("tx_hash = ?", txHash).First(&record).Error
	if err != nil {
		return nil, err
	}
	return &record, nil
}

// ListAIFlags retrieves AI-flagged records with pagination
func (r *GormSmartContractRepository) ListAIFlags(page, limit int) ([]models.SmartContractRecord, int64, error) {
	var records []models.SmartContractRecord
	var total int64

	query := r.DB.Model(&models.SmartContractRecord{}).
		Where("event_type IN ?", []models.EventType{
			models.EventAIFlagged,
			models.EventAIWarning,
		})

	// Get total count
	err := query.Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	// Get paginated records
	err = query.Offset((page - 1) * limit).
		Limit(limit).
		Order("event_timestamp DESC").
		Find(&records).Error

	return records, total, err
}
