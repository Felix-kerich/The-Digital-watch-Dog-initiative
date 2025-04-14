package repository

import (
	"errors"
	"os"

	"github.com/the-digital-watchdog-initiative/models"
	"github.com/the-digital-watchdog-initiative/utils"
	"gorm.io/gorm"
)

// GormFileRepository implements FileRepository with GORM
type GormFileRepository struct {
	DB *gorm.DB
}

// NewFileRepository creates a new FileRepository
func NewFileRepository() FileRepository {
	return &GormFileRepository{
		DB: utils.DB,
	}
}

// Create adds a new file to the database
func (r *GormFileRepository) Create(file *models.File) error {
	return r.DB.Create(file).Error
}

// FindByID retrieves a file by ID
func (r *GormFileRepository) FindByID(id string) (*models.File, error) {
	var file models.File
	if err := r.DB.First(&file, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("file not found")
		}
		return nil, err
	}
	return &file, nil
}

// Update updates a file in the database
func (r *GormFileRepository) Update(file *models.File) error {
	return r.DB.Save(file).Error
}

// Delete removes a file from the database and disk
func (r *GormFileRepository) Delete(id string) error {
	var file models.File
	if err := r.DB.First(&file, "id = ?", id).Error; err != nil {
		return err
	}

	// Delete the file from disk
	if file.FilePath != "" {
		if err := os.Remove(file.FilePath); err != nil && !os.IsNotExist(err) {
			utils.Logger.Warnf("Failed to delete file from disk: %v", err)
			// Continue with deletion from database even if disk removal fails
		}
	}

	// Delete from database
	return r.DB.Delete(&file).Error
}

// List retrieves a paginated list of files with optional filters
func (r *GormFileRepository) List(page, limit int, filter map[string]interface{}) ([]models.File, int64, error) {
	var files []models.File
	var total int64

	query := r.DB.Model(&models.File{})

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
	if err := query.Order("created_at DESC").Offset(offset).Limit(limit).Find(&files).Error; err != nil {
		return nil, 0, err
	}

	return files, total, nil
}

// FindByEntityID retrieves files associated with an entity
func (r *GormFileRepository) FindByEntityID(entityID string, page, limit int) ([]models.File, int64, error) {
	return r.List(page, limit, map[string]interface{}{"entity_id": entityID})
}

// FindByTransactionID retrieves files associated with a transaction
func (r *GormFileRepository) FindByTransactionID(transactionID string, page, limit int) ([]models.File, int64, error) {
	return r.List(page, limit, map[string]interface{}{"transaction_id": transactionID})
}

// FindByFundID retrieves files associated with a fund
func (r *GormFileRepository) FindByFundID(fundID string, page, limit int) ([]models.File, int64, error) {
	return r.List(page, limit, map[string]interface{}{"fund_id": fundID})
}
