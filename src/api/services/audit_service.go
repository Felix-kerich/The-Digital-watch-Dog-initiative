package services

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/the-digital-watchdog-initiative/models"
	"github.com/the-digital-watchdog-initiative/repository"
	"github.com/the-digital-watchdog-initiative/utils"
)

// AuditServiceImpl implements AuditService interface
type AuditServiceImpl struct {
	auditRepo repository.AuditLogRepository
	logger    *utils.NamedLogger
}

// NewAuditService creates a new audit service
func NewAuditService(auditRepo repository.AuditLogRepository) AuditService {
	return &AuditServiceImpl{
		auditRepo: auditRepo,
		logger:    utils.NewLogger("audit-service"),
	}
}

// LogActivity logs a user activity
func (s *AuditServiceImpl) LogActivity(userID, action, entityType, entityID string, metadata map[string]interface{}) error {
	s.logger.Info("Logging activity", map[string]interface{}{
		"userID":     userID,
		"action":     action,
		"entityType": entityType,
		"entityID":   entityID,
	})

	// Convert metadata to string for storage in Detail field
	detailStr := ""
	if metadata != nil {
		detailBytes, err := json.Marshal(metadata)
		if err != nil {
			s.logger.Error("Failed to marshal metadata", map[string]interface{}{
				"error": err.Error(),
			})
		} else {
			detailStr = string(detailBytes)
		}
	}

	auditLog := &models.AuditLog{
		ID:         uuid.New().String(),
		UserID:     userID,
		Action:     action,
		EntityType: entityType,
		EntityID:   entityID,
		Timestamp:  time.Now(),
		Detail:     detailStr,
	}

	if err := s.auditRepo.Create(auditLog); err != nil {
		s.logger.Error("Failed to create audit log", map[string]interface{}{
			"userID":     userID,
			"action":     action,
			"entityType": entityType,
			"entityID":   entityID,
			"error":      err.Error(),
		})
		return err
	}

	return nil
}

// GetAuditLogs retrieves audit logs with pagination and filtering
func (s *AuditServiceImpl) GetAuditLogs(page, limit int, filter map[string]interface{}) ([]models.AuditLog, int64, error) {
	s.logger.Info("Getting audit logs", map[string]interface{}{
		"page":   page,
		"limit":  limit,
		"filter": filter,
	})

	logs, total, err := s.auditRepo.List(page, limit, filter)
	if err != nil {
		s.logger.Error("Failed to get audit logs", map[string]interface{}{
			"page":   page,
			"limit":  limit,
			"filter": filter,
			"error":  err.Error(),
		})
		return nil, 0, err
	}

	return logs, total, nil
}

// GetAuditLogsByUserID retrieves audit logs for a specific user
func (s *AuditServiceImpl) GetAuditLogsByUserID(userID string, page, limit int) ([]models.AuditLog, int64, error) {
	s.logger.Info("Getting audit logs by user ID", map[string]interface{}{
		"userID": userID,
		"page":   page,
		"limit":  limit,
	})

	logs, total, err := s.auditRepo.FindByUserID(userID, page, limit)
	if err != nil {
		s.logger.Error("Failed to get audit logs by user ID", map[string]interface{}{
			"userID": userID,
			"page":   page,
			"limit":  limit,
			"error":  err.Error(),
		})
		return nil, 0, err
	}

	return logs, total, nil
}

// GetAuditLogsByEntityID retrieves audit logs for a specific entity
func (s *AuditServiceImpl) GetAuditLogsByEntityID(entityID string, entityType string, page, limit int) ([]models.AuditLog, int64, error) {
	s.logger.Info("Getting audit logs by entity ID", map[string]interface{}{
		"entityID":   entityID,
		"entityType": entityType,
		"page":       page,
		"limit":      limit,
	})

	logs, total, err := s.auditRepo.FindByEntityID(entityID, entityType, page, limit)
	if err != nil {
		s.logger.Error("Failed to get audit logs by entity ID", map[string]interface{}{
			"entityID":   entityID,
			"entityType": entityType,
			"page":       page,
			"limit":      limit,
			"error":      err.Error(),
		})
		return nil, 0, err
	}

	return logs, total, nil
}
