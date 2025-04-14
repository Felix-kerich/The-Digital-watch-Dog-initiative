package repository

import (
	"github.com/shopspring/decimal"
	"github.com/the-digital-watchdog-initiative/models"
	"github.com/the-digital-watchdog-initiative/utils"
	"gorm.io/gorm"
)

// BudgetLineItemRepository interface defines the methods for budget line item operations
type BudgetLineItemRepository interface {
	Create(item *models.BudgetLineItem) error
	FindByID(id string) (*models.BudgetLineItem, error)
	FindAll(page, limit int, filter map[string]interface{}) ([]models.BudgetLineItem, int64, error)
	Update(item *models.BudgetLineItem) error
	Delete(id string) error
	FindByFundID(fundID string) ([]models.BudgetLineItem, error)
	UpdateUtilized(id string, amount decimal.Decimal) error
}

// GormBudgetLineItemRepository implements BudgetLineItemRepository using GORM
type GormBudgetLineItemRepository struct {
	DB *gorm.DB
}

// NewBudgetLineItemRepository creates a new budget line item repository
func NewBudgetLineItemRepository() BudgetLineItemRepository {
	return &GormBudgetLineItemRepository{
		DB: utils.DB,
	}
}

// Create creates a new budget line item
func (r *GormBudgetLineItemRepository) Create(item *models.BudgetLineItem) error {
	return r.DB.Create(item).Error
}

// FindByID finds a budget line item by ID
func (r *GormBudgetLineItemRepository) FindByID(id string) (*models.BudgetLineItem, error) {
	var item models.BudgetLineItem
	err := r.DB.First(&item, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &item, nil
}

// FindAll retrieves all budget line items with pagination and filtering
func (r *GormBudgetLineItemRepository) FindAll(page, limit int, filter map[string]interface{}) ([]models.BudgetLineItem, int64, error) {
	var items []models.BudgetLineItem
	var total int64

	query := r.DB.Model(&models.BudgetLineItem{})

	// Apply filters
	for key, value := range filter {
		query = query.Where(key+" = ?", value)
	}

	// Get total count
	err := query.Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	// Get paginated results
	err = query.Offset((page - 1) * limit).Limit(limit).Find(&items).Error
	if err != nil {
		return nil, 0, err
	}

	return items, total, nil
}

// Update updates a budget line item
func (r *GormBudgetLineItemRepository) Update(item *models.BudgetLineItem) error {
	return r.DB.Save(item).Error
}

// Delete deletes a budget line item
func (r *GormBudgetLineItemRepository) Delete(id string) error {
	return r.DB.Delete(&models.BudgetLineItem{}, "id = ?", id).Error
}

// FindByFundID finds all budget line items for a specific fund
func (r *GormBudgetLineItemRepository) FindByFundID(fundID string) ([]models.BudgetLineItem, error) {
	var items []models.BudgetLineItem
	err := r.DB.Where("fund_id = ?", fundID).Find(&items).Error
	return items, err
}

// UpdateUtilized updates the utilized amount of a budget line item
func (r *GormBudgetLineItemRepository) UpdateUtilized(id string, amount decimal.Decimal) error {
	// First retrieve the current budget line item
	var item models.BudgetLineItem
	if err := r.DB.First(&item, "id = ?", id).Error; err != nil {
		return err
	}

	// Update the utilized amount using Add method
	item.Utilized = item.Utilized.Add(amount)

	// Save the updated item
	return r.DB.Save(&item).Error
}
