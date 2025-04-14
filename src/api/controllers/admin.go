package controllers

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/the-digital-watchdog-initiative/models"
	"github.com/the-digital-watchdog-initiative/repository"
	"github.com/the-digital-watchdog-initiative/utils"
)

// AdminController handles administrative operations
type AdminController struct {
	adminRepo repository.AdminRepository
	userRepo  repository.UserRepository
	auditRepo repository.AuditLogRepository
}

// NewAdminController creates a new admin controller
func NewAdminController() *AdminController {
	return &AdminController{
		adminRepo: repository.NewAdminRepository(),
		userRepo:  repository.NewUserRepository(),
		auditRepo: repository.NewAuditLogRepository(),
	}
}

// GetUsers retrieves all users with filtering and pagination
func (ac *AdminController) GetUsers(c *gin.Context) {
	// Parse pagination parameters
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 10
	}

	// Get filter parameters
	filter := map[string]interface{}{
		"role":     c.Query("role"),
		"entityId": c.Query("entityId"),
		"isActive": c.Query("isActive"),
		"search":   c.Query("search"),
	}

	// Get users with pagination
	users, total, err := ac.adminRepo.GetUsers(page, limit, filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve users"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"users":      users,
		"total":      total,
		"page":       page,
		"limit":      limit,
		"totalPages": (int(total) + limit - 1) / limit,
	})
}

// GetUserByID retrieves a user by ID
func (ac *AdminController) GetUserByID(c *gin.Context) {
	id := c.Param("id")

	user, err := ac.adminRepo.GetUserByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	c.JSON(http.StatusOK, user)
}

// CreateUser creates a new user
func (ac *AdminController) CreateUser(c *gin.Context) {
	var req struct {
		Name     string          `json:"name" binding:"required"`
		Email    string          `json:"email" binding:"required,email"`
		Password string          `json:"password" binding:"required,min=8"`
		Role     models.UserRole `json:"role" binding:"required"`
		EntityID string          `json:"entityId"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Hash the password
	hashedPassword, err := utils.GeneratePasswordHash(req.Password)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to hash password"})
		return
	}

	// Create new user
	user := &models.User{
		Name:         req.Name,
		Email:        req.Email,
		PasswordHash: hashedPassword,
		Role:         req.Role,
		EntityID:     req.EntityID,
		IsActive:     true,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	if err := ac.adminRepo.CreateUser(user); err != nil {
		if _, ok := err.(*utils.ConflictError); ok {
			c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user"})
		}
		return
	}

	// Get current admin's user ID from context
	adminID, _ := c.Get("userID")

	// Create audit log
	auditLog := &models.AuditLog{
		Action:     "USER_CREATED",
		UserID:     adminID.(string),
		EntityType: "User",
		EntityID:   user.ID,
		Detail:     "Created user: " + user.Email + " with role: " + string(user.Role),
		IP:         c.ClientIP(),
		Timestamp:  time.Now(),
	}

	if err := ac.auditRepo.Create(auditLog); err != nil {
		utils.Logger.Warn("Failed to create audit log for user creation")
	}

	// Don't return password hash
	user.PasswordHash = ""
	c.JSON(http.StatusCreated, user)
}

// UpdateUser updates a user
func (ac *AdminController) UpdateUser(c *gin.Context) {
	id := c.Param("id")

	user, err := ac.adminRepo.GetUserByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	var req struct {
		Name     string          `json:"name"`
		Email    string          `json:"email"`
		Role     models.UserRole `json:"role"`
		EntityID string          `json:"entityId"`
		IsActive *bool           `json:"isActive"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Update fields if provided
	if req.Name != "" {
		user.Name = req.Name
	}
	if req.Email != "" {
		user.Email = req.Email
	}
	if req.Role != "" {
		user.Role = req.Role
	}
	if req.EntityID != "" {
		user.EntityID = req.EntityID
	}
	if req.IsActive != nil {
		user.IsActive = *req.IsActive
	}

	user.UpdatedAt = time.Now()

	if err := ac.adminRepo.UpdateUser(user); err != nil {
		if _, ok := err.(*utils.ConflictError); ok {
			c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update user"})
		}
		return
	}

	// Get current admin's user ID from context
	adminID, _ := c.Get("userID")

	// Create audit log
	auditLog := &models.AuditLog{
		Action:     "USER_UPDATED",
		UserID:     adminID.(string),
		EntityType: "User",
		EntityID:   user.ID,
		Detail:     "Updated user: " + user.Email,
		IP:         c.ClientIP(),
		Timestamp:  time.Now(),
	}

	if err := ac.auditRepo.Create(auditLog); err != nil {
		utils.Logger.Warn("Failed to create audit log for user update")
	}

	// Don't return password hash
	user.PasswordHash = ""
	c.JSON(http.StatusOK, user)
}

// ResetUserPassword resets a user's password
func (ac *AdminController) ResetUserPassword(c *gin.Context) {
	id := c.Param("id")

	user, err := ac.adminRepo.GetUserByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	var req struct {
		NewPassword string `json:"newPassword" binding:"required,min=8"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Hash the new password
	hashedPassword, err := utils.GeneratePasswordHash(req.NewPassword)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to hash password"})
		return
	}

	if err := ac.adminRepo.ResetUserPassword(id, hashedPassword); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to reset password"})
		return
	}

	// Get current admin's user ID from context
	adminID, _ := c.Get("userID")

	// Create audit log
	auditLog := &models.AuditLog{
		Action:     "PASSWORD_RESET",
		UserID:     adminID.(string),
		EntityType: "User",
		EntityID:   user.ID,
		Detail:     "Reset password for user: " + user.Email,
		IP:         c.ClientIP(),
		Timestamp:  time.Now(),
	}

	if err := ac.auditRepo.Create(auditLog); err != nil {
		utils.Logger.Warn("Failed to create audit log for password reset")
	}

	c.JSON(http.StatusOK, gin.H{"message": "Password reset successfully"})
}

// GetSystemInfo returns general system information and statistics
func (ac *AdminController) GetSystemInfo(c *gin.Context) {
	stats, err := ac.adminRepo.GetSystemStats()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve system stats"})
		return
	}

	// Get recent activity
	recentActivity, err := ac.adminRepo.GetRecentActivity(10)
	if err != nil {
		utils.Logger.Warn("Failed to retrieve recent activity")
	}

	// Get user registration trends
	trends, err := ac.adminRepo.GetUserRegistrationTrends(12)
	if err != nil {
		utils.Logger.Warn("Failed to retrieve user registration trends")
	}

	response := gin.H{
		"counts":                 stats["counts"],
		"recentActivity":         recentActivity,
		"userRegistrationTrends": trends,
	}

	c.JSON(http.StatusOK, response)
}
