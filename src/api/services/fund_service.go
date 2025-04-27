package services

import (
	"fmt"
	"time"

	"github.com/the-digital-watchdog-initiative/models"
	"github.com/the-digital-watchdog-initiative/repository"
	"github.com/the-digital-watchdog-initiative/utils"
)

// FundServiceImpl implements FundService interface
type FundServiceImpl struct {
	fundRepo  repository.FundRepository
	auditRepo repository.AuditLogRepository
	logger    *utils.NamedLogger
}

// NewFundService creates a new fund service
func NewFundService(fundRepo repository.FundRepository, auditRepo repository.AuditLogRepository) FundService {
	return &FundServiceImpl{
		fundRepo:  fundRepo,
		auditRepo: auditRepo,
		logger:    utils.NewLogger("fund-service"),
	}
}

// CreateFund creates a new fund
func (s *FundServiceImpl) CreateFund(fund *models.Fund) error {
	s.logger.Info("Creating new fund", map[string]interface{}{
		"name":     fund.Name,
		"entityID": fund.EntityID,
	})

	// Set creation timestamp if not set
	if fund.CreatedAt.IsZero() {
		fund.CreatedAt = time.Now()
	}
	fund.UpdatedAt = time.Now()

	// Create the fund
	if err := s.fundRepo.Create(fund); err != nil {
		s.logger.Error("Failed to create fund", map[string]interface{}{
			"name":     fund.Name,
			"entityID": fund.EntityID,
			"error":    err.Error(),
		})
		return err
	}

	// Log the activity
	s.auditRepo.Create(&models.AuditLog{
		UserID:     fund.CreatedByID,
		Action:     "CREATE_FUND",
		EntityType: "FUND",
		EntityID:   fund.ID,
		Timestamp:  time.Now(),
		Detail:     fmt.Sprintf("Created fund: %s (%s)", fund.Name, fund.EntityID),
	})

	return nil
}

// GetFundByID retrieves a fund by ID
func (s *FundServiceImpl) GetFundByID(id string) (*models.Fund, error) {
	s.logger.Info("Getting fund by ID", map[string]interface{}{"fundID": id})

	fund, err := s.fundRepo.FindByID(id)
	if err != nil {
		s.logger.Error("Failed to get fund", map[string]interface{}{
			"fundID": id,
			"error":  err.Error(),
		})
		return nil, err
	}

	return fund, nil
}

// GetFunds retrieves funds with pagination and filtering
func (s *FundServiceImpl) GetFunds(page, limit int, filter map[string]interface{}) ([]models.Fund, int64, error) {
	s.logger.Info("Getting funds list", map[string]interface{}{
		"page":   page,
		"limit":  limit,
		"filter": filter,
	})

	funds, total, err := s.fundRepo.List(page, limit, filter)
	if err != nil {
		s.logger.Error("Failed to get funds list", map[string]interface{}{
			"page":   page,
			"limit":  limit,
			"filter": filter,
			"error":  err.Error(),
		})
		return nil, 0, err
	}

	return funds, total, nil
}

// GetFundsByEntityID retrieves funds for a specific entity
func (s *FundServiceImpl) GetFundsByEntityID(entityID string, page, limit int) ([]models.Fund, int64, error) {
	s.logger.Info("Getting funds by entity ID", map[string]interface{}{
		"entityID": entityID,
		"page":     page,
		"limit":    limit,
	})

	funds, total, err := s.fundRepo.FindByEntityID(entityID, page, limit)
	if err != nil {
		s.logger.Error("Failed to get funds by entity ID", map[string]interface{}{
			"entityID": entityID,
			"page":     page,
			"limit":    limit,
			"error":    err.Error(),
		})
		return nil, 0, err
	}

	return funds, total, nil
}

// UpdateFund updates a fund
func (s *FundServiceImpl) UpdateFund(fund *models.Fund) error {
	s.logger.Info("Updating fund", map[string]interface{}{"fundID": fund.ID})

	// Get the current fund to verify it exists
	existingFund, err := s.fundRepo.FindByID(fund.ID)
	if err != nil {
		s.logger.Error("Failed to find fund for update", map[string]interface{}{
			"fundID": fund.ID,
			"error":  err.Error(),
		})
		return err
	}

	// Update timestamp
	fund.UpdatedAt = time.Now()
	fund.CreatedAt = existingFund.CreatedAt // Preserve creation time

	// Update the fund
	if err := s.fundRepo.Update(fund); err != nil {
		s.logger.Error("Failed to update fund", map[string]interface{}{
			"fundID": fund.ID,
			"error":  err.Error(),
		})
		return err
	}

	// Log the activity
	s.auditRepo.Create(&models.AuditLog{
		UserID:     fund.CreatedByID,
		Action:     "UPDATE_FUND",
		EntityType: "FUND",
		EntityID:   fund.ID,
		Timestamp:  time.Now(),
	})

	return nil
}

// DeleteFund deletes a fund
func (s *FundServiceImpl) DeleteFund(id string) error {
	s.logger.Info("Deleting fund", map[string]interface{}{"fundID": id})

	// Check if fund exists
	_, err := s.fundRepo.FindByID(id)
	if err != nil {
		s.logger.Error("Failed to find fund for deletion", map[string]interface{}{
			"fundID": id,
			"error":  err.Error(),
		})
		return err
	}

	// Delete the fund
	if err := s.fundRepo.Delete(id); err != nil {
		s.logger.Error("Failed to delete fund", map[string]interface{}{
			"fundID": id,
			"error":  err.Error(),
		})
		return err
	}

	// Log the activity
	s.auditRepo.Create(&models.AuditLog{
		Action:     "DELETE_FUND",
		EntityType: "FUND",
		EntityID:   id,
		Timestamp:  time.Now(),
	})

	return nil
}
