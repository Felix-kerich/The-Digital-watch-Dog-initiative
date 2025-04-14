package repository

import (
	"errors"

	"github.com/the-digital-watchdog-initiative/models"
	"github.com/the-digital-watchdog-initiative/utils"
	"gorm.io/gorm"
)

// GormEntityRepository implements EntityRepository with GORM
type GormEntityRepository struct {
	DB *gorm.DB
}

// NewEntityRepository creates a new EntityRepository
func NewEntityRepository() EntityRepository {
	return &GormEntityRepository{
		DB: utils.DB,
	}
}

// Create adds a new entity to the database
func (r *GormEntityRepository) Create(entity *models.Entity) error {
	return r.DB.Create(entity).Error
}

// FindByID retrieves an entity by ID
func (r *GormEntityRepository) FindByID(id string) (*models.Entity, error) {
	var entity models.Entity
	if err := r.DB.First(&entity, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("entity not found")
		}
		return nil, err
	}
	return &entity, nil
}

// Update updates an entity in the database
func (r *GormEntityRepository) Update(entity *models.Entity) error {
	return r.DB.Save(entity).Error
}

// Delete performs a soft delete on an entity
func (r *GormEntityRepository) Delete(id string) error {
	return r.DB.Delete(&models.Entity{}, "id = ?", id).Error
}

// List retrieves a paginated list of entities with optional filters
func (r *GormEntityRepository) List(page, limit int, filter map[string]interface{}) ([]models.Entity, int64, error) {
	var entities []models.Entity
	var total int64

	query := r.DB.Model(&models.Entity{})

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
	if err := query.Order("created_at DESC").Offset(offset).Limit(limit).Find(&entities).Error; err != nil {
		return nil, 0, err
	}

	return entities, total, nil
}
