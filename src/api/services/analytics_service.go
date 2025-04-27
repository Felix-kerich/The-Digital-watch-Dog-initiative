package services

import (
	"github.com/the-digital-watchdog-initiative/models"
	"github.com/the-digital-watchdog-initiative/repository"
	"github.com/the-digital-watchdog-initiative/utils"
)

// AnalyticsServiceImpl implements AnalyticsService interface
type AnalyticsServiceImpl struct {
	analyticsRepo repository.AnalyticsRepository
	adminRepo     repository.AdminRepository
	auditRepo     repository.AuditLogRepository
	logger        *utils.NamedLogger
}

// NewAnalyticsService creates a new analytics service
func NewAnalyticsService(analyticsRepo repository.AnalyticsRepository, auditRepo repository.AuditLogRepository) AnalyticsService {
	return &AnalyticsServiceImpl{
		analyticsRepo: analyticsRepo,
		auditRepo:     auditRepo,
		logger:        utils.NewLogger("analytics-service"),
	}
}

// GetTransactionSummary returns a summary of transactions
func (s *AnalyticsServiceImpl) GetTransactionSummary(filter map[string]interface{}) (map[string]interface{}, error) {
	s.logger.Info("Getting transaction summary", map[string]interface{}{
		"filter": filter,
	})

	summary, err := s.analyticsRepo.GetTransactionSummary(filter)
	if err != nil {
		s.logger.Error("Failed to get transaction summary", map[string]interface{}{
			"filter": filter,
			"error":  err.Error(),
		})
		return nil, err
	}

	return summary, nil
}

// GetUserActivitySummary returns a summary of user activity
func (s *AnalyticsServiceImpl) GetUserActivitySummary(filter map[string]interface{}) (map[string]interface{}, error) {
	s.logger.Info("Getting user activity summary", map[string]interface{}{
		"filter": filter,
	})

	summary, err := s.analyticsRepo.GetUserActivitySummary(filter)
	if err != nil {
		s.logger.Error("Failed to get user activity summary", map[string]interface{}{
			"filter": filter,
			"error":  err.Error(),
		})
		return nil, err
	}

	return summary, nil
}

// GetFundUtilizationReport returns a report on fund utilization
func (s *AnalyticsServiceImpl) GetFundUtilizationReport(fundID, entityID, fiscalYear string) (map[string]interface{}, error) {
	s.logger.Info("Getting fund utilization report", map[string]interface{}{
		"fundID":     fundID,
		"entityID":   entityID,
		"fiscalYear": fiscalYear,
	})

	report, err := s.analyticsRepo.GetFundUtilizationReport(fundID, entityID, fiscalYear)
	if err != nil {
		s.logger.Error("Failed to get fund utilization report", map[string]interface{}{
			"fundID":     fundID,
			"entityID":   entityID,
			"fiscalYear": fiscalYear,
			"error":      err.Error(),
		})
		return nil, err
	}

	return report, nil
}

// GetSystemStats returns system-wide statistics
func (s *AnalyticsServiceImpl) GetSystemStats() (map[string]interface{}, error) {
	s.logger.Info("Getting system stats", nil)

	// This method might be implemented in the admin repository
	// For now, we'll assume it's part of the analytics repository
	stats, err := s.adminRepo.GetSystemStats()
	if err != nil {
		s.logger.Error("Failed to get system stats", map[string]interface{}{
			"error": err.Error(),
		})
		return nil, err
	}

	return stats, nil
}

// GetRecentActivity returns recent system activity
func (s *AnalyticsServiceImpl) GetRecentActivity(limit int) ([]models.AuditLog, error) {
	s.logger.Info("Getting recent activity", map[string]interface{}{
		"limit": limit,
	})

	// This method might be implemented in the admin repository
	// For now, we'll assume it's part of the analytics repository
	activity, err := s.adminRepo.GetRecentActivity(limit)
	if err != nil {
		s.logger.Error("Failed to get recent activity", map[string]interface{}{
			"limit": limit,
			"error": err.Error(),
		})
		return nil, err
	}

	return activity, nil
}

// GetUserRegistrationTrends returns trends in user registration
func (s *AnalyticsServiceImpl) GetUserRegistrationTrends(months int) ([]map[string]interface{}, error) {
	s.logger.Info("Getting user registration trends", map[string]interface{}{
		"months": months,
	})

	// This method might be implemented in the admin repository
	// For now, we'll assume it's part of the analytics repository
	trends, err := s.adminRepo.GetUserRegistrationTrends(months)
	if err != nil {
		s.logger.Error("Failed to get user registration trends", map[string]interface{}{
			"months": months,
			"error":  err.Error(),
		})
		return nil, err
	}

	return trends, nil
}
