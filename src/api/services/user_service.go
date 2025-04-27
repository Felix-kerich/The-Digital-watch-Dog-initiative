package services

import (
	"fmt"
	"time"

	"github.com/the-digital-watchdog-initiative/models"
	"github.com/the-digital-watchdog-initiative/repository"
	"github.com/the-digital-watchdog-initiative/utils"
	"golang.org/x/crypto/bcrypt"
)

// UserServiceImpl implements UserService interface
type UserServiceImpl struct {
	userRepo  repository.UserRepository
	auditRepo repository.AuditLogRepository
	logger    *utils.NamedLogger
}

// NewUserService creates a new user service
func NewUserService(userRepo repository.UserRepository, auditRepo repository.AuditLogRepository) UserService {
	return &UserServiceImpl{
		userRepo:  userRepo,
		auditRepo: auditRepo,
		logger:    utils.NewLogger("user-service"),
	}
}

// GetUserProfile retrieves a user's profile
func (s *UserServiceImpl) GetUserProfile(userID string) (*models.User, error) {
	s.logger.Info("Getting user profile", map[string]interface{}{"userID": userID})
	
	user, err := s.userRepo.FindByID(userID)
	if err != nil {
		s.logger.Error("Failed to get user profile", map[string]interface{}{
			"userID": userID,
			"error":  err.Error(),
		})
		return nil, err
	}

	// Don't expose password hash
	user.PasswordHash = ""
	
	return user, nil
}

// UpdateUserProfile updates a user's profile
func (s *UserServiceImpl) UpdateUserProfile(userID string, userData map[string]interface{}) (*models.User, error) {
	s.logger.Info("Updating user profile", map[string]interface{}{"userID": userID})
	
	// Get the current user
	user, err := s.userRepo.FindByID(userID)
	if err != nil {
		s.logger.Error("Failed to find user for update", map[string]interface{}{
			"userID": userID,
			"error":  err.Error(),
		})
		return nil, err
	}

	// Update fields that are allowed to be updated
	if name, ok := userData["name"].(string); ok && name != "" {
		user.Name = name
	}
	
	if email, ok := userData["email"].(string); ok && email != "" && email != user.Email {
		// Check if email is already in use
		existingUser, _ := s.userRepo.FindByEmail(email)
		if existingUser != nil {
			return nil, utils.NewConflictError("Email already in use", "email")
		}
		user.Email = email
	}

	// Update the user
	if err := s.userRepo.Update(user); err != nil {
		s.logger.Error("Failed to update user", map[string]interface{}{
			"userID": userID,
			"error":  err.Error(),
		})
		return nil, err
	}

	// Log the activity
	s.auditRepo.Create(&models.AuditLog{
		UserID:     userID,
		Action:     "UPDATE_PROFILE",
		EntityType: "USER",
		EntityID:   userID,
		Timestamp:  time.Now(),
		Detail:     fmt.Sprintf("Updated user profile: %s", userData),
	})

	// Don't expose password hash
	user.PasswordHash = ""
	
	return user, nil
}

// ChangeUserPassword changes a user's password
func (s *UserServiceImpl) ChangeUserPassword(userID, currentPassword, newPassword string) error {
	s.logger.Info("Changing user password", map[string]interface{}{"userID": userID})
	
	// Get the current user
	user, err := s.userRepo.FindByID(userID)
	if err != nil {
		s.logger.Error("Failed to find user for password change", map[string]interface{}{
			"userID": userID,
			"error":  err.Error(),
		})
		return err
	}

	// Verify current password
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(currentPassword)); err != nil {
		s.logger.Warn("Invalid current password during password change", map[string]interface{}{"userID": userID})
		return utils.ErrInvalidCredentials
	}

	// Hash the new password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		s.logger.Error("Failed to hash new password", map[string]interface{}{
			"userID": userID,
			"error":  err.Error(),
		})
		return err
	}

	// Update the password
	user.PasswordHash = string(hashedPassword)
	if err := s.userRepo.Update(user); err != nil {
		s.logger.Error("Failed to update user password", map[string]interface{}{
			"userID": userID,
			"error":  err.Error(),
		})
		return err
	}

	// Log the activity
	s.auditRepo.Create(&models.AuditLog{
		UserID:     userID,
		Action:     "CHANGE_PASSWORD",
		EntityType: "USER",
		EntityID:   userID,
		Timestamp:  time.Now(),
	})
	
	return nil
}

// CreateUser creates a new user (admin function)
func (s *UserServiceImpl) CreateUser(user *models.User) error {
	s.logger.Info("Creating new user", map[string]interface{}{"email": user.Email})
	
	// Check if email is already in use
	existingUser, _ := s.userRepo.FindByEmail(user.Email)
	if existingUser != nil {
		return utils.NewConflictError("Email already in use", "email")
	}

	// Hash the password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.PasswordHash), bcrypt.DefaultCost)
	if err != nil {
		s.logger.Error("Failed to hash password", map[string]interface{}{
			"email": user.Email,
			"error": err.Error(),
		})
		return err
	}
	user.PasswordHash = string(hashedPassword)

	// Set creation timestamp
	user.CreatedAt = time.Now()
	
	// Create the user
	if err := s.userRepo.Create(user); err != nil {
		s.logger.Error("Failed to create user", map[string]interface{}{
			"email": user.Email,
			"error": err.Error(),
		})
		return err
	}

	return nil
}

// GetUserByID retrieves a user by ID
func (s *UserServiceImpl) GetUserByID(id string) (*models.User, error) {
	s.logger.Info("Getting user by ID", map[string]interface{}{"userID": id})
	
	user, err := s.userRepo.FindByID(id)
	if err != nil {
		s.logger.Error("Failed to get user by ID", map[string]interface{}{
			"userID": id,
			"error":  err.Error(),
		})
		return nil, err
	}

	// Don't expose password hash
	user.PasswordHash = ""
	
	return user, nil
}

// GetUsers retrieves users with pagination and filtering
func (s *UserServiceImpl) GetUsers(page, limit int, filter map[string]interface{}) ([]models.User, int64, error) {
	s.logger.Info("Getting users list", map[string]interface{}{
		"page":   page,
		"limit":  limit,
		"filter": filter,
	})
	
	users, total, err := s.userRepo.List(page, limit, filter)
	if err != nil {
		s.logger.Error("Failed to get users list", map[string]interface{}{
			"page":   page,
			"limit":  limit,
			"filter": filter,
			"error":  err.Error(),
		})
		return nil, 0, err
	}

	// Don't expose password hashes
	for i := range users {
		users[i].PasswordHash = ""
	}
	
	return users, total, nil
}

// UpdateUser updates a user (admin function)
func (s *UserServiceImpl) UpdateUser(user *models.User) error {
	s.logger.Info("Updating user", map[string]interface{}{"userID": user.ID})
	
	// Get the current user to verify it exists
	existingUser, err := s.userRepo.FindByID(user.ID)
	if err != nil {
		s.logger.Error("Failed to find user for update", map[string]interface{}{
			"userID": user.ID,
			"error":  err.Error(),
		})
		return err
	}

	// If email is being changed, check if it's already in use
	if user.Email != existingUser.Email {
		emailUser, _ := s.userRepo.FindByEmail(user.Email)
		if emailUser != nil && emailUser.ID != user.ID {
			return utils.NewConflictError("Email already in use", "email")
		}
	}

	// Preserve the password hash if not being updated
	if user.PasswordHash == "" {
		user.PasswordHash = existingUser.PasswordHash
	}

	// Update the user
	if err := s.userRepo.Update(user); err != nil {
		s.logger.Error("Failed to update user", map[string]interface{}{
			"userID": user.ID,
			"error":  err.Error(),
		})
		return err
	}
	
	return nil
}

// DeleteUser deletes a user
func (s *UserServiceImpl) DeleteUser(id string) error {
	s.logger.Info("Deleting user", map[string]interface{}{"userID": id})
	
	// Check if user exists
	_, err := s.userRepo.FindByID(id)
	if err != nil {
		s.logger.Error("Failed to find user for deletion", map[string]interface{}{
			"userID": id,
			"error":  err.Error(),
		})
		return err
	}

	// Delete the user (soft delete)
	if err := s.userRepo.Delete(id); err != nil {
		s.logger.Error("Failed to delete user", map[string]interface{}{
			"userID": id,
			"error":  err.Error(),
		})
		return err
	}
	
	return nil
}

// ResetUserPassword resets a user's password (admin function)
func (s *UserServiceImpl) ResetUserPassword(id string, newPassword string) error {
	s.logger.Info("Resetting user password", map[string]interface{}{"userID": id})
	
	// Get the current user
	user, err := s.userRepo.FindByID(id)
	if err != nil {
		s.logger.Error("Failed to find user for password reset", map[string]interface{}{
			"userID": id,
			"error":  err.Error(),
		})
		return err
	}

	// Hash the new password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		s.logger.Error("Failed to hash new password", map[string]interface{}{
			"userID": id,
			"error":  err.Error(),
		})
		return err
	}

	// Update the password
	user.PasswordHash = string(hashedPassword)
	if err := s.userRepo.Update(user); err != nil {
		s.logger.Error("Failed to update user password", map[string]interface{}{
			"userID": id,
			"error":  err.Error(),
		})
		return err
	}

	// Log the activity
	s.auditRepo.Create(&models.AuditLog{
		Action:     "RESET_PASSWORD",
		EntityType: "USER",
		EntityID:   id,
		Timestamp:  time.Now(),
	})
	
	return nil
}
