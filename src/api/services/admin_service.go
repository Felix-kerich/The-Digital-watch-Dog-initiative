package services

import (
	"time"

	"github.com/the-digital-watchdog-initiative/models"
	"github.com/the-digital-watchdog-initiative/repository"
	"github.com/the-digital-watchdog-initiative/utils"
)

// AdminServiceImpl implements AdminService interface
type AdminServiceImpl struct {
	adminRepo repository.AdminRepository
	userRepo  repository.UserRepository
	logger    *utils.NamedLogger
}

// NewAdminService creates a new admin service
func NewAdminService(adminRepo repository.AdminRepository, userRepo repository.UserRepository) AdminService {
	return &AdminServiceImpl{
		adminRepo: adminRepo,
		userRepo:  userRepo,
		logger:    utils.NewLogger("admin-service"),
	}
}

// GetUsers retrieves all users with filtering and pagination
func (s *AdminServiceImpl) GetUsers(page, limit int, filter map[string]interface{}) ([]models.User, int64, error) {
	s.logger.Info("Getting users with pagination", map[string]interface{}{
		"page":   page,
		"limit":  limit,
		"filter": filter,
	})

	return s.adminRepo.GetUsers(page, limit, filter)
}

// GetUserByID retrieves a user by ID
func (s *AdminServiceImpl) GetUserByID(id string) (*models.User, error) {
	s.logger.Info("Getting user by ID", map[string]interface{}{
		"userID": id,
	})

	return s.userRepo.FindByID(id)
}

// CreateUser creates a new user
func (s *AdminServiceImpl) CreateUser(user *models.User) error {
	s.logger.Info("Creating new user", map[string]interface{}{
		"email": user.Email,
		"role":  user.Role,
	})

	// Set creation timestamps
	if user.CreatedAt.IsZero() {
		user.CreatedAt = time.Now()
	}
	if user.UpdatedAt.IsZero() {
		user.UpdatedAt = time.Now()
	}

	return s.userRepo.Create(user)
}

// UpdateUser updates a user
func (s *AdminServiceImpl) UpdateUser(user *models.User) error {
	s.logger.Info("Updating user", map[string]interface{}{
		"userID": user.ID,
	})

	// Update timestamp
	user.UpdatedAt = time.Now()

	return s.userRepo.Update(user)
}

// ResetUserPassword resets a user's password
func (s *AdminServiceImpl) ResetUserPassword(id string, newPassword string) error {
	s.logger.Info("Resetting user password", map[string]interface{}{
		"userID": id,
	})

	// Get the user
	user, err := s.userRepo.FindByID(id)
	if err != nil {
		s.logger.Error("Failed to get user for password reset", map[string]interface{}{
			"userID": id,
			"error":  err.Error(),
		})
		return err
	}

	// Hash the new password
	hashedPassword, err := utils.GeneratePasswordHash(newPassword)
	if err != nil {
		s.logger.Error("Failed to hash new password", map[string]interface{}{
			"error": err.Error(),
		})
		return err
	}

	// Update the user's password
	user.PasswordHash = hashedPassword
	user.UpdatedAt = time.Now()

	return s.userRepo.Update(user)
}

// GetSystemInfo returns general system information and statistics
func (s *AdminServiceImpl) GetSystemInfo() (map[string]interface{}, error) {
	s.logger.Info("Getting system information", nil)

	return s.adminRepo.GetSystemStats()
}
