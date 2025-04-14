package repository

import (
	"errors"
	"time"

	"github.com/the-digital-watchdog-initiative/models"
	"github.com/the-digital-watchdog-initiative/utils"
	"gorm.io/gorm"
)

// GormTransactionRepository implements TransactionRepository with GORM
type GormTransactionRepository struct {
	DB *gorm.DB
}

// NewTransactionRepository creates a new TransactionRepository
func NewTransactionRepository() TransactionRepository {
	return &GormTransactionRepository{
		DB: utils.DB,
	}
}

// Create adds a new transaction to the database
func (r *GormTransactionRepository) Create(transaction *models.Transaction) error {
	// Verify that the fund exists
	var fund models.Fund
	if err := r.DB.First(&fund, "id = ?", transaction.FundID).Error; err != nil {
		return errors.New("invalid fund ID")
	}

	// Verify that the source entity exists
	var sourceEntity models.Entity
	if err := r.DB.First(&sourceEntity, "id = ?", transaction.SourceID).Error; err != nil {
		return errors.New("invalid source entity ID")
	}

	// Verify that the destination entity exists
	var destEntity models.Entity
	if err := r.DB.First(&destEntity, "id = ?", transaction.DestinationID).Error; err != nil {
		return errors.New("invalid destination entity ID")
	}

	// Verify budget line item if provided
	if transaction.BudgetLineItemID != "" {
		var budgetLineItem models.BudgetLineItem
		if err := r.DB.First(&budgetLineItem, "id = ?", transaction.BudgetLineItemID).Error; err != nil {
			return errors.New("invalid budget line item ID")
		}
	}

	// Create the transaction
	return r.DB.Create(transaction).Error
}

// FindByID retrieves a transaction by ID
func (r *GormTransactionRepository) FindByID(id string) (*models.Transaction, error) {
	var transaction models.Transaction
	err := r.DB.
		Preload("Fund").
		Preload("BudgetLineItem").
		Preload("Source").
		Preload("Destination").
		Preload("ApprovedBy").
		Preload("RejectedBy").
		Preload("ReviewedBy").
		Preload("CreatedBy").
		First(&transaction, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &transaction, nil
}

// Update updates a transaction in the database
func (r *GormTransactionRepository) Update(transaction *models.Transaction) error {
	return r.DB.Save(transaction).Error
}

// Delete removes a transaction from the database
func (r *GormTransactionRepository) Delete(id string) error {
	return r.DB.Delete(&models.Transaction{}, "id = ?", id).Error
}

// List retrieves a paginated list of transactions with optional filters
func (r *GormTransactionRepository) List(page, limit int, filter map[string]interface{}) ([]models.Transaction, int64, error) {
	var transactions []models.Transaction
	var total int64

	query := r.DB.Model(&models.Transaction{})

	// Apply filters
	for key, value := range filter {
		query = query.Where(key, value)
	}

	// Get total count
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Apply pagination and ordering
	offset := (page - 1) * limit
	err := query.Preload("Fund").
		Preload("Source").
		Preload("Destination").
		Preload("BudgetLineItem").
		Preload("CreatedBy").
		Preload("ApprovedBy").
		Order("created_at DESC").
		Offset(offset).
		Limit(limit).
		Find(&transactions).Error

	return transactions, total, err
}

// FindByFundID retrieves transactions associated with a fund
func (r *GormTransactionRepository) FindByFundID(fundID string, page, limit int) ([]models.Transaction, int64, error) {
	return r.List(page, limit, map[string]interface{}{"fund_id": fundID})
}

// FindByEntityID retrieves transactions associated with an entity as source or destination
func (r *GormTransactionRepository) FindByEntityID(entityID string, page, limit int) ([]models.Transaction, int64, error) {
	var transactions []models.Transaction
	var total int64

	query := r.DB.Model(&models.Transaction{}).
		Where("source_id = ? OR destination_id = ?", entityID, entityID)

	// Get total count
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Apply pagination and ordering
	offset := (page - 1) * limit
	err := query.Preload("Fund").
		Preload("Source").
		Preload("Destination").
		Preload("BudgetLineItem").
		Preload("CreatedBy").
		Preload("ApprovedBy").
		Order("created_at DESC").
		Offset(offset).
		Limit(limit).
		Find(&transactions).Error

	return transactions, total, err
}

// GetAll retrieves all transactions with filters and pagination
func (r *GormTransactionRepository) GetAll(page, limit int, filter map[string]interface{}) ([]models.Transaction, int64, error) {
	var transactions []models.Transaction
	var total int64

	query := r.DB.Model(&models.Transaction{})

	// Apply filters
	if transactionType, ok := filter["type"].(string); ok && transactionType != "" {
		query = query.Where("transaction_type = ?", transactionType)
	}
	if status, ok := filter["status"].(string); ok && status != "" {
		query = query.Where("status = ?", status)
	}
	if fundID, ok := filter["fundId"].(string); ok && fundID != "" {
		query = query.Where("fund_id = ?", fundID)
	}
	if sourceID, ok := filter["sourceId"].(string); ok && sourceID != "" {
		query = query.Where("source_id = ?", sourceID)
	}
	if destinationID, ok := filter["destinationId"].(string); ok && destinationID != "" {
		query = query.Where("destination_id = ?", destinationID)
	}
	if startDate, ok := filter["startDate"].(string); ok && startDate != "" {
		query = query.Where("created_at >= ?", startDate)
	}
	if endDate, ok := filter["endDate"].(string); ok && endDate != "" {
		query = query.Where("created_at <= ?", endDate)
	}
	if aiFlagged, ok := filter["aiFlagged"].(string); ok && aiFlagged != "" {
		query = query.Where("ai_flagged = ?", aiFlagged == "true")
	}

	// Count total items for pagination
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Get transactions with pagination and preload relationships
	offset := (page - 1) * limit
	err := query.Preload("Fund").
		Preload("Source").
		Preload("Destination").
		Preload("BudgetLineItem").
		Preload("CreatedBy").
		Preload("ApprovedBy").
		Offset(offset).
		Limit(limit).
		Order("created_at DESC").
		Find(&transactions).Error

	return transactions, total, err
}

// GetByID retrieves a transaction by ID with all relationships
func (r *GormTransactionRepository) GetByID(id string) (*models.Transaction, error) {
	var transaction models.Transaction
	result := r.DB.
		Preload("Fund").
		Preload("Source").
		Preload("Destination").
		Preload("BudgetLineItem").
		Preload("CreatedBy").
		Preload("ApprovedBy").
		First(&transaction, "id = ?", id)

	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return nil, utils.ErrNotFound
		}
		return nil, result.Error
	}

	return &transaction, nil
}

// Approve approves a transaction
func (r *GormTransactionRepository) Approve(id string, approverID string) error {
	transaction := &models.Transaction{}
	if err := r.DB.First(transaction, "id = ?", id).Error; err != nil {
		return err
	}

	// Check if transaction can be approved
	if transaction.Status != models.TransactionPending {
		return errors.New("only pending transactions can be approved")
	}

	// Update transaction
	transaction.Status = models.TransactionApproved
	transaction.ApprovedByID = &approverID
	transaction.UpdatedAt = time.Now()

	return r.DB.Save(transaction).Error
}

// Reject rejects a transaction
func (r *GormTransactionRepository) Reject(id string, rejectedByID string, reason string) error {
	transaction := &models.Transaction{}
	if err := r.DB.First(transaction, "id = ?", id).Error; err != nil {
		return err
	}

	// Check if transaction can be rejected
	if transaction.Status != models.TransactionPending && transaction.Status != models.TransactionFlagged {
		return errors.New("only pending or flagged transactions can be rejected")
	}

	// Update transaction
	transaction.Status = models.TransactionRejected
	transaction.RejectionReason = reason
	transaction.RejectedByID = &rejectedByID
	transaction.UpdatedAt = time.Now()

	return r.DB.Save(transaction).Error
}

// Complete marks a transaction as completed
func (r *GormTransactionRepository) Complete(id string) error {
	return r.DB.Transaction(func(tx *gorm.DB) error {
		transaction := &models.Transaction{}
		if err := tx.First(transaction, "id = ?", id).Error; err != nil {
			return err
		}

		// Get associated fund
		fund := &models.Fund{}
		if err := tx.First(fund, "id = ?", transaction.FundID).Error; err != nil {
			return err
		}

		// Update fund values based on transaction type
		switch transaction.TransactionType {
		case models.TransactionAllocation:
			fund.Allocated = fund.Allocated.Add(transaction.Amount)
		case models.TransactionDisbursement:
			fund.Disbursed = fund.Disbursed.Add(transaction.Amount)
		case models.TransactionExpenditure:
			fund.Utilized = fund.Utilized.Add(transaction.Amount)
		case models.TransactionReturns:
			// Return funds to the fund
			if transaction.Amount.LessThanOrEqual(fund.Allocated) {
				fund.Allocated = fund.Allocated.Sub(transaction.Amount)
			}
		}

		// Save updated fund
		if err := tx.Save(fund).Error; err != nil {
			return err
		}

		// Update transaction
		transaction.Status = models.TransactionCompleted
		transaction.UpdatedAt = time.Now()

		return tx.Save(transaction).Error
	})
}

// Flag marks a transaction as flagged
func (r *GormTransactionRepository) Flag(id string, reason string) error {
	transaction := &models.Transaction{}
	if err := r.DB.First(transaction, "id = ?", id).Error; err != nil {
		return err
	}

	// Update transaction
	transaction.Status = models.TransactionFlagged
	transaction.AiFlagged = true
	transaction.AiReasonDetails = reason
	transaction.UpdatedAt = time.Now()

	return r.DB.Save(transaction).Error
}

// CompleteTransaction handles the completion of a transaction with fund updates
func (r *GormTransactionRepository) CompleteTransaction(transaction *models.Transaction, fund *models.Fund) error {
	return r.DB.Transaction(func(tx *gorm.DB) error {
		// Update transaction
		if err := tx.Save(transaction).Error; err != nil {
			return err
		}

		// Update fund
		if err := tx.Save(fund).Error; err != nil {
			return err
		}

		// Update budget line item if specified
		if transaction.BudgetLineItemID != "" {
			var lineItem models.BudgetLineItem
			if err := tx.First(&lineItem, "id = ?", transaction.BudgetLineItemID).Error; err != nil {
				return err
			}

			// Update line item amounts based on transaction type
			switch transaction.TransactionType {
			case models.TransactionAllocation:
				lineItem.Utilized = lineItem.Utilized.Add(transaction.Amount)
			case models.TransactionExpenditure:
				lineItem.Utilized = lineItem.Utilized.Add(transaction.Amount)
			}

			if err := tx.Save(&lineItem).Error; err != nil {
				return err
			}
		}

		return nil
	})
}
