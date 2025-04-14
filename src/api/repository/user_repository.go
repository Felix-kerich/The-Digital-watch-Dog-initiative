package repository

import (
	"errors"
	"time"

	"github.com/the-digital-watchdog-initiative/models"
	"github.com/the-digital-watchdog-initiative/utils"
	"gorm.io/gorm"
)

// GormUserRepository implements UserRepository with GORM
type GormUserRepository struct {
	DB *gorm.DB
}

// NewUserRepository creates a new UserRepository
func NewUserRepository() UserRepository {
	return &GormUserRepository{
		DB: utils.DB,
	}
}

// Create adds a new user to the database
func (r *GormUserRepository) Create(user *models.User) error {
	return r.DB.Create(user).Error
}

// FindByID retrieves a user by ID
func (r *GormUserRepository) FindByID(id string) (*models.User, error) {
	var user models.User
	if err := r.DB.First(&user, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("user not found")
		}
		return nil, err
	}
	return &user, nil
}

// FindByEmail retrieves a user by email
func (r *GormUserRepository) FindByEmail(email string) (*models.User, error) {
	var user models.User
	if err := r.DB.First(&user, "email = ?", email).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("user not found")
		}
		return nil, err
	}
	return &user, nil
}

// Update updates a user in the database
func (r *GormUserRepository) Update(user *models.User) error {
	return r.DB.Save(user).Error
}

// Delete performs a soft delete on a user
func (r *GormUserRepository) Delete(id string) error {
	return r.DB.Delete(&models.User{}, "id = ?", id).Error
}

// List retrieves a paginated list of users with optional filters
func (r *GormUserRepository) List(page, limit int, filter map[string]interface{}) ([]models.User, int64, error) {
	var users []models.User
	var total int64

	query := r.DB

	// Apply filters
	for key, value := range filter {
		query = query.Where(key, value)
	}

	// Get total count
	if err := query.Model(&models.User{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Apply pagination
	offset := (page - 1) * limit
	if err := query.Offset(offset).Limit(limit).Find(&users).Error; err != nil {
		return nil, 0, err
	}

	return users, total, nil
}

// UpdateLastLogin updates the last login time of a user
func (r *GormUserRepository) UpdateLastLogin(id string, loginTime time.Time) error {
	return r.DB.Model(&models.User{}).Where("id = ?", id).Update("last_login", loginTime).Error
}

// FindAll retrieves all users from the database
func (r *GormUserRepository) FindAll() ([]models.User, error) {
	var users []models.User
	result := r.DB.Find(&users)
	return users, result.Error
}
