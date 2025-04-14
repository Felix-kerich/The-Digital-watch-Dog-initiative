package repository

import (
	"time"

	"github.com/google/uuid"
	"github.com/the-digital-watchdog-initiative/models"
	"github.com/the-digital-watchdog-initiative/utils"
	"gorm.io/gorm"
)

// GormAdminRepository implements AdminRepository using GORM
type GormAdminRepository struct {
	DB *gorm.DB
}

// NewAdminRepository creates a new AdminRepository
func NewAdminRepository() AdminRepository {
	return &GormAdminRepository{
		DB: utils.DB,
	}
}

// GetUsers retrieves all users with filtering and pagination
func (r *GormAdminRepository) GetUsers(page, limit int, filter map[string]interface{}) ([]models.User, int64, error) {
	var users []models.User
	var total int64

	query := r.DB.Model(&models.User{})

	// Apply filters
	for key, value := range filter {
		switch key {
		case "role":
			query = query.Where("role = ?", value)
		case "entityId":
			query = query.Where("entity_id = ?", value)
		case "isActive":
			query = query.Where("is_active = ?", value)
		case "search":
			searchValue := "%" + value.(string) + "%"
			query = query.Where("name LIKE ? OR email LIKE ?", searchValue, searchValue)
		}
	}

	// Get total count
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Get users with pagination
	offset := (page - 1) * limit
	if err := query.Offset(offset).Limit(limit).Find(&users).Error; err != nil {
		return nil, 0, err
	}

	// Don't return password hashes
	for i := range users {
		users[i].PasswordHash = ""
	}

	return users, total, nil
}

// GetUserByID retrieves a user by ID
func (r *GormAdminRepository) GetUserByID(id string) (*models.User, error) {
	var user models.User
	if err := r.DB.First(&user, "id = ?", id).Error; err != nil {
		return nil, err
	}

	// Don't return password hash
	user.PasswordHash = ""
	return &user, nil
}

// CreateUser creates a new user
func (r *GormAdminRepository) CreateUser(user *models.User) error {
	// Check if email already exists
	var existingUser models.User
	if err := r.DB.Where("email = ?", user.Email).First(&existingUser).Error; err == nil {
		return &utils.ConflictError{Message: "Email already registered"}
	} else if err != gorm.ErrRecordNotFound {
		return err
	}

	// Set ID if not already set
	if user.ID == "" {
		user.ID = uuid.New().String()
	}

	// Set timestamps if not already set
	now := time.Now()
	if user.CreatedAt.IsZero() {
		user.CreatedAt = now
	}
	if user.UpdatedAt.IsZero() {
		user.UpdatedAt = now
	}

	return r.DB.Create(user).Error
}

// UpdateUser updates a user
func (r *GormAdminRepository) UpdateUser(user *models.User) error {
	// Check if user exists
	var existingUser models.User
	if err := r.DB.First(&existingUser, "id = ?", user.ID).Error; err != nil {
		return err
	}

	// Check if email is being changed and if it's already in use
	if user.Email != "" && user.Email != existingUser.Email {
		var duplicateUser models.User
		if err := r.DB.Where("email = ?", user.Email).First(&duplicateUser).Error; err == nil {
			return &utils.ConflictError{Message: "Email already in use"}
		} else if err != gorm.ErrRecordNotFound {
			return err
		}
	}

	// Update the user
	user.UpdatedAt = time.Now()
	return r.DB.Save(user).Error
}

// ResetUserPassword resets a user's password
func (r *GormAdminRepository) ResetUserPassword(id string, newPassword string) error {
	return r.DB.Model(&models.User{}).Where("id = ?", id).Update("password_hash", newPassword).Error
}

// GetSystemStats returns general system information and statistics
func (r *GormAdminRepository) GetSystemStats() (map[string]interface{}, error) {
	var stats = make(map[string]interface{})

	// Get counts
	var userCount, transactionCount, entityCount, fundCount, flaggedTransactionCount int64

	if err := r.DB.Model(&models.User{}).Count(&userCount).Error; err != nil {
		return nil, err
	}
	if err := r.DB.Model(&models.Transaction{}).Count(&transactionCount).Error; err != nil {
		return nil, err
	}
	if err := r.DB.Model(&models.Entity{}).Count(&entityCount).Error; err != nil {
		return nil, err
	}
	if err := r.DB.Model(&models.Fund{}).Count(&fundCount).Error; err != nil {
		return nil, err
	}
	if err := r.DB.Model(&models.Transaction{}).Where("ai_flagged = ?", true).Count(&flaggedTransactionCount).Error; err != nil {
		return nil, err
	}

	stats["counts"] = map[string]int64{
		"users":               userCount,
		"transactions":        transactionCount,
		"entities":            entityCount,
		"funds":               fundCount,
		"flaggedTransactions": flaggedTransactionCount,
	}

	return stats, nil
}

// GetRecentActivity returns recent activity from the audit log
func (r *GormAdminRepository) GetRecentActivity(limit int) ([]models.AuditLog, error) {
	var logs []models.AuditLog
	if err := r.DB.Order("timestamp DESC").Limit(limit).Find(&logs).Error; err != nil {
		return nil, err
	}
	return logs, nil
}

// GetUserRegistrationTrends returns user registration trends over time
func (r *GormAdminRepository) GetUserRegistrationTrends(months int) ([]map[string]interface{}, error) {
	var trends []map[string]interface{}

	query := `
		SELECT
			DATE_FORMAT(created_at, '%Y-%m-01') as month,
			COUNT(*) as count
		FROM
			users
		WHERE
			created_at >= DATE_SUB(NOW(), INTERVAL ? MONTH)
		GROUP BY
			DATE_FORMAT(created_at, '%Y-%m')
		ORDER BY
			month
	`

	if err := r.DB.Raw(query, months).Scan(&trends).Error; err != nil {
		return nil, err
	}

	return trends, nil
}
