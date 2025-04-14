package controllers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/the-digital-watchdog-initiative/models"
	"github.com/the-digital-watchdog-initiative/repository"
	"github.com/the-digital-watchdog-initiative/utils"
)

// UserController handles user-related operations
type UserController struct {
	userRepo  repository.UserRepository
	auditRepo repository.AuditLogRepository
}

// NewUserController creates a new user controller
func NewUserController() *UserController {
	return &UserController{
		userRepo:  repository.NewUserRepository(),
		auditRepo: repository.NewAuditLogRepository(),
	}
}

// GetProfile retrieves the authenticated user's profile
func (uc *UserController) GetProfile(c *gin.Context) {
	userID, _ := c.Get("userID")

	user, err := uc.userRepo.FindByID(userID.(string))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	// Don't send password hash
	user.PasswordHash = ""

	c.JSON(http.StatusOK, user)
}

// UpdateProfile updates the authenticated user's profile
func (uc *UserController) UpdateProfile(c *gin.Context) {
	userID, _ := c.Get("userID")

	user, err := uc.userRepo.FindByID(userID.(string))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	var req struct {
		Name  string `json:"name"`
		Email string `json:"email"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Check if email is being changed and if it's already in use
	if req.Email != "" && req.Email != user.Email {
		existingUser, err := uc.userRepo.FindByEmail(req.Email)
		if err == nil && existingUser != nil {
			c.JSON(http.StatusConflict, gin.H{"error": "Email already in use"})
			return
		}
		user.Email = req.Email
	}

	// Update name if provided
	if req.Name != "" {
		user.Name = req.Name
	}

	user.UpdatedAt = time.Now()

	if err := uc.userRepo.Update(user); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update profile"})
		return
	}

	// Create audit log
	auditLog := &models.AuditLog{
		Action:     "USER_PROFILE_UPDATED",
		UserID:     user.ID,
		EntityType: "User",
		EntityID:   user.ID,
		Detail:     "User profile updated: " + user.Email,
		IP:         c.ClientIP(),
		Timestamp:  time.Now(),
	}

	if err := uc.auditRepo.Create(auditLog); err != nil {
		utils.Logger.Warn("Failed to create audit log for profile update")
	}

	// Don't send password hash
	user.PasswordHash = ""

	c.JSON(http.StatusOK, user)
}

// ChangePassword changes the authenticated user's password
func (uc *UserController) ChangePassword(c *gin.Context) {
	userID, _ := c.Get("userID")

	user, err := uc.userRepo.FindByID(userID.(string))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	var req struct {
		CurrentPassword string `json:"currentPassword" binding:"required"`
		NewPassword     string `json:"newPassword" binding:"required,min=8"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Verify current password
	if !utils.CheckPasswordHash(req.CurrentPassword, user.PasswordHash) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Current password is incorrect"})
		return
	}

	// Hash new password
	hashedPassword, err := utils.GeneratePasswordHash(req.NewPassword)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to hash password"})
		return
	}

	// Update password
	user.PasswordHash = hashedPassword
	user.UpdatedAt = time.Now()

	if err := uc.userRepo.Update(user); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update password"})
		return
	}

	// Create audit log
	auditLog := &models.AuditLog{
		Action:     "USER_PASSWORD_CHANGED",
		UserID:     user.ID,
		EntityType: "User",
		EntityID:   user.ID,
		Detail:     "User changed password",
		IP:         c.ClientIP(),
		Timestamp:  time.Now(),
	}

	if err := uc.auditRepo.Create(auditLog); err != nil {
		utils.Logger.Warn("Failed to create audit log for password change")
	}

	c.JSON(http.StatusOK, gin.H{"message": "Password updated successfully"})
}
