package services

import (
	"github.com/sirupsen/logrus"
	"github.com/the-digital-watchdog-initiative/models"
	"github.com/the-digital-watchdog-initiative/repository"
	"github.com/the-digital-watchdog-initiative/utils"
)

// BudgetLineItemServiceImpl implements BudgetLineItemService
type BudgetLineItemServiceImpl struct {
	budgetRepo repository.BudgetLineItemRepository
	fundRepo   repository.FundRepository
	logger     *logrus.Logger
}

// NewBudgetLineItemService creates a new budget line item service
func NewBudgetLineItemService(budgetRepo repository.BudgetLineItemRepository, fundRepo repository.FundRepository) BudgetLineItemService {
	return &BudgetLineItemServiceImpl{
		budgetRepo: budgetRepo,
		fundRepo:   fundRepo,
		logger:     utils.Logger,
	}
}

// CreateBudgetLineItem creates a new budget line item
func (s *BudgetLineItemServiceImpl) CreateBudgetLineItem(item *models.BudgetLineItem) error {
	s.logger.WithFields(logrus.Fields{
		"name":   item.Name,
		"code":   item.Code,
		"fundID": item.FundID,
	}).Info("Creating new budget line item")

	// Verify fund exists
	_, err := s.fundRepo.FindByID(item.FundID)
	if err != nil {
		s.logger.WithError(err).WithField("fundID", item.FundID).Error("Invalid fund ID")
		return err
	}

	// Create budget line item
	err = s.budgetRepo.Create(item)
	if err != nil {
		s.logger.WithError(err).Error("Failed to create budget line item")
		return err
	}

	s.logger.WithField("id", item.ID).Info("Budget line item created successfully")
	return nil
}

// GetBudgetLineItemByID retrieves a budget line item by ID
func (s *BudgetLineItemServiceImpl) GetBudgetLineItemByID(id string) (*models.BudgetLineItem, error) {
	s.logger.WithField("id", id).Info("Retrieving budget line item")
	
	item, err := s.budgetRepo.FindByID(id)
	if err != nil {
		s.logger.WithError(err).WithField("id", id).Error("Failed to retrieve budget line item")
		return nil, err
	}
	
	return item, nil
}

// GetBudgetLineItems retrieves all budget line items with pagination and filtering
func (s *BudgetLineItemServiceImpl) GetBudgetLineItems(page, limit int, filter map[string]interface{}) ([]models.BudgetLineItem, int64, error) {
	s.logger.WithFields(logrus.Fields{
		"page":   page,
		"limit":  limit,
		"filter": filter,
	}).Info("Retrieving budget line items")
	
	items, total, err := s.budgetRepo.FindAll(page, limit, filter)
	if err != nil {
		s.logger.WithError(err).Error("Failed to retrieve budget line items")
		return nil, 0, err
	}
	
	return items, total, nil
}

// GetBudgetLineItemsByFundID retrieves budget line items by fund ID
func (s *BudgetLineItemServiceImpl) GetBudgetLineItemsByFundID(fundID string, page, limit int) ([]models.BudgetLineItem, int64, error) {
	s.logger.WithFields(logrus.Fields{
		"fundID": fundID,
		"page":   page,
		"limit":  limit,
	}).Info("Retrieving budget line items by fund ID")
	
	filter := map[string]interface{}{"fund_id": fundID}
	items, total, err := s.budgetRepo.FindAll(page, limit, filter)
	if err != nil {
		s.logger.WithError(err).WithField("fundID", fundID).Error("Failed to retrieve budget line items")
		return nil, 0, err
	}
	
	return items, total, nil
}

// UpdateBudgetLineItem updates a budget line item
func (s *BudgetLineItemServiceImpl) UpdateBudgetLineItem(item *models.BudgetLineItem) error {
	s.logger.WithField("id", item.ID).Info("Updating budget line item")
	
	// Verify budget line item exists
	existingItem, err := s.budgetRepo.FindByID(item.ID)
	if err != nil {
		s.logger.WithError(err).WithField("id", item.ID).Error("Budget line item not found")
		return err
	}
	
	// Verify fund exists if fund ID is changed
	if existingItem.FundID != item.FundID {
		_, err := s.fundRepo.FindByID(item.FundID)
		if err != nil {
			s.logger.WithError(err).WithField("fundID", item.FundID).Error("Invalid fund ID")
			return err
		}
	}
	
	// Update budget line item
	err = s.budgetRepo.Update(item)
	if err != nil {
		s.logger.WithError(err).Error("Failed to update budget line item")
		return err
	}
	
	s.logger.WithField("id", item.ID).Info("Budget line item updated successfully")
	return nil
}

// DeleteBudgetLineItem deletes a budget line item
func (s *BudgetLineItemServiceImpl) DeleteBudgetLineItem(id string) error {
	s.logger.WithField("id", id).Info("Deleting budget line item")
	
	// Verify budget line item exists
	_, err := s.budgetRepo.FindByID(id)
	if err != nil {
		s.logger.WithError(err).WithField("id", id).Error("Budget line item not found")
		return err
	}
	
	// Delete budget line item
	err = s.budgetRepo.Delete(id)
	if err != nil {
		s.logger.WithError(err).Error("Failed to delete budget line item")
		return err
	}
	
	s.logger.WithField("id", id).Info("Budget line item deleted successfully")
	return nil
}
