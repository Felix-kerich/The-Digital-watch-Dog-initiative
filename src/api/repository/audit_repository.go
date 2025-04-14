package repository

import (
	"errors"

	"github.com/the-digital-watchdog-initiative/models"
	"github.com/the-digital-watchdog-initiative/utils"
	"gorm.io/gorm"
)

// GormAuditLogRepository implements AuditLogRepository with GORM
type GormAuditLogRepository struct {
	DB *gorm.DB
}

// NewAuditLogRepository creates a new AuditLogRepository
func NewAuditLogRepository() AuditLogRepository {
	return &GormAuditLogRepository{
		DB: utils.DB,
	}
}

// Create adds a new audit log to the database
func (r *GormAuditLogRepository) Create(auditLog *models.AuditLog) error {
	return r.DB.Create(auditLog).Error
}

// FindByID retrieves an audit log by ID
func (r *GormAuditLogRepository) FindByID(id string) (*models.AuditLog, error) {
	var auditLog models.AuditLog
	if err := r.DB.First(&auditLog, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("audit log not found")
		}
		return nil, err
	}
	return &auditLog, nil
}

// List retrieves a paginated list of audit logs with optional filters
func (r *GormAuditLogRepository) List(page, limit int, filter map[string]interface{}) ([]models.AuditLog, int64, error) {
	var auditLogs []models.AuditLog
	var total int64

	query := r.DB.Model(&models.AuditLog{})

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
	if err := query.Order("timestamp DESC").Offset(offset).Limit(limit).Find(&auditLogs).Error; err != nil {
		return nil, 0, err
	}

	return auditLogs, total, nil
}

// FindByUserID retrieves audit logs for a specific user
func (r *GormAuditLogRepository) FindByUserID(userID string, page, limit int) ([]models.AuditLog, int64, error) {
	return r.List(page, limit, map[string]interface{}{"user_id": userID})
}

// FindByEntityID retrieves audit logs for a specific entity
func (r *GormAuditLogRepository) FindByEntityID(entityID string, entityType string, page, limit int) ([]models.AuditLog, int64, error) {
	filter := map[string]interface{}{"entity_id": entityID}
	if entityType != "" {
		filter["entity_type"] = entityType
	}
	return r.List(page, limit, filter)
}
