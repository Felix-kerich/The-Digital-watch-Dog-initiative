package repository

import (
	"errors"

	"github.com/the-digital-watchdog-initiative/models"
	"github.com/the-digital-watchdog-initiative/utils"
	"gorm.io/gorm"
)

// GormFundRepository implements FundRepository with GORM
type GormFundRepository struct {
	DB *gorm.DB
}

// NewFundRepository creates a new FundRepository
func NewFundRepository() FundRepository {
	return &GormFundRepository{
		DB: utils.DB,
	}
}

// Create adds a new fund to the database
func (r *GormFundRepository) Create(fund *models.Fund) error {
	return r.DB.Create(fund).Error
}

// FindByID retrieves a fund by ID
func (r *GormFundRepository) FindByID(id string) (*models.Fund, error) {
	var fund models.Fund
	if err := r.DB.First(&fund, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("fund not found")
		}
		return nil, err
	}
	return &fund, nil
}

// Update updates a fund in the database
func (r *GormFundRepository) Update(fund *models.Fund) error {
	return r.DB.Save(fund).Error
}

// Delete performs a soft delete on a fund
func (r *GormFundRepository) Delete(id string) error {
	return r.DB.Delete(&models.Fund{}, "id = ?", id).Error
}

// List retrieves a paginated list of funds with optional filters
func (r *GormFundRepository) List(page, limit int, filter map[string]interface{}) ([]models.Fund, int64, error) {
	var funds []models.Fund
	var total int64

	query := r.DB.Model(&models.Fund{})

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
	if err := query.Order("created_at DESC").Offset(offset).Limit(limit).Find(&funds).Error; err != nil {
		return nil, 0, err
	}

	return funds, total, nil
}

// FindByEntityID retrieves funds associated with an entity
func (r *GormFundRepository) FindByEntityID(entityID string, page, limit int) ([]models.Fund, int64, error) {
	return r.List(page, limit, map[string]interface{}{"entity_id": entityID})
}
