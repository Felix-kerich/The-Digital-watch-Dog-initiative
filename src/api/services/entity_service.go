package services

import (
	"fmt"
	"time"

	"github.com/the-digital-watchdog-initiative/models"
	"github.com/the-digital-watchdog-initiative/repository"
	"github.com/the-digital-watchdog-initiative/utils"
)

// EntityServiceImpl implements EntityService interface
type EntityServiceImpl struct {
	entityRepo repository.EntityRepository
	auditRepo  repository.AuditLogRepository
	logger     *utils.NamedLogger
}

// NewEntityService creates a new entity service
func NewEntityService(entityRepo repository.EntityRepository, auditRepo repository.AuditLogRepository) EntityService {
	return &EntityServiceImpl{
		entityRepo: entityRepo,
		auditRepo:  auditRepo,
		logger:     utils.NewLogger("entity-service"),
	}
}

// CreateEntity creates a new entity
func (s *EntityServiceImpl) CreateEntity(entity *models.Entity) error {
	s.logger.Info("Creating new entity", map[string]interface{}{
		"name": entity.Name,
		"type": entity.Type,
	})

	// Set creation timestamp if not set
	if entity.CreatedAt.IsZero() {
		entity.CreatedAt = time.Now()
	}
	entity.UpdatedAt = time.Now()

	// Create the entity
	if err := s.entityRepo.Create(entity); err != nil {
		s.logger.Error("Failed to create entity", map[string]interface{}{
			"name":  entity.Name,
			"type":  entity.Type,
			"error": err.Error(),
		})
		return err
	}

	// Log the activity
	s.auditRepo.Create(&models.AuditLog{
		UserID:     entity.CreatedByID,
		Action:     "CREATE_ENTITY",
		EntityType: "ENTITY",
		EntityID:   entity.ID,
		Timestamp:  time.Now(),
		Detail:     fmt.Sprintf("Created entity: %s (%s)", entity.Name, entity.Type),
	})

	return nil
}

// GetEntityByID retrieves an entity by ID
func (s *EntityServiceImpl) GetEntityByID(id string) (*models.Entity, error) {
	s.logger.Info("Getting entity by ID", map[string]interface{}{"entityID": id})

	entity, err := s.entityRepo.FindByID(id)
	if err != nil {
		s.logger.Error("Failed to get entity", map[string]interface{}{
			"entityID": id,
			"error":    err.Error(),
		})
		return nil, err
	}

	return entity, nil
}

// GetEntities retrieves entities with pagination and filtering
func (s *EntityServiceImpl) GetEntities(page, limit int, filter map[string]interface{}) ([]models.Entity, int64, error) {
	s.logger.Info("Getting entities list", map[string]interface{}{
		"page":   page,
		"limit":  limit,
		"filter": filter,
	})

	entities, total, err := s.entityRepo.List(page, limit, filter)
	if err != nil {
		s.logger.Error("Failed to get entities list", map[string]interface{}{
			"page":   page,
			"limit":  limit,
			"filter": filter,
			"error":  err.Error(),
		})
		return nil, 0, err
	}

	return entities, total, nil
}

// UpdateEntity updates an entity
func (s *EntityServiceImpl) UpdateEntity(id string, updateData map[string]interface{}) (*models.Entity, error) {
	s.logger.Info("Updating entity", map[string]interface{}{
		"entityID":   id,
		"updateData": updateData,
	})

	// Check if entity exists
	existingEntity, err := s.entityRepo.FindByID(id)
	if err != nil {
		s.logger.Error("Entity not found", map[string]interface{}{
			"entityID": id,
			"error":    err.Error(),
		})
		return nil, utils.ErrNotFound
	}

	// Apply updates to the entity
	for key, value := range updateData {
		switch key {
		case "name":
			existingEntity.Name = value.(string)
		case "description":
			existingEntity.Description = value.(string)
		case "type":
			existingEntity.Type = value.(string)
		case "location":
			existingEntity.Location = value.(string)
		case "contactInfo":
			existingEntity.ContactInfo = value.(string)
		case "isActive":
			existingEntity.IsActive = value.(bool)
		}
	}

	// Update timestamp
	existingEntity.UpdatedAt = time.Now()

	// Update the entity
	if err := s.entityRepo.Update(existingEntity); err != nil {
		s.logger.Error("Failed to update entity", map[string]interface{}{
			"entityID": id,
			"error":    err.Error(),
		})
		return nil, err
	}

	// Log the activity if UpdatedByID is available
	if updatedByID, ok := updateData["updatedByID"].(string); ok && updatedByID != "" {
		s.auditRepo.Create(&models.AuditLog{
			UserID:     updatedByID,
			Action:     "UPDATE_ENTITY",
			EntityType: "ENTITY",
			EntityID:   id,
			Timestamp:  time.Now(),
			Detail:     "Updated entity: " + existingEntity.Name,
		})
	}

	return existingEntity, nil
}

// DeleteEntity deletes an entity
func (s *EntityServiceImpl) DeleteEntity(id string) error {
	s.logger.Info("Deleting entity", map[string]interface{}{"entityID": id})

	// Check if entity exists
	existingEntity, err := s.entityRepo.FindByID(id)
	if err != nil {
		s.logger.Error("Entity not found", map[string]interface{}{
			"entityID": id,
			"error":    err.Error(),
		})
		return utils.ErrNotFound
	}

	// Soft delete by marking as inactive
	existingEntity.IsActive = false
	existingEntity.UpdatedAt = time.Now()

	// Update the entity
	if err := s.entityRepo.Update(existingEntity); err != nil {
		s.logger.Error("Failed to delete entity", map[string]interface{}{
			"entityID": id,
			"error":    err.Error(),
		})
		return err
	}

	s.auditRepo.Create(&models.AuditLog{
		Action:     "DELETE_ENTITY",
		EntityType: "ENTITY",
		EntityID:   id,
		Timestamp:  time.Now(),
		Detail:     "Deleted entity: " + existingEntity.Name,
	})

	return nil
}
