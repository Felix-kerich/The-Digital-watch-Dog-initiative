package handlers

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/the-digital-watchdog-initiative/models"
	"github.com/the-digital-watchdog-initiative/services"
	"github.com/the-digital-watchdog-initiative/utils"
)

// AdminController handles administrative operations
type AdminController struct {
	adminService services.AdminService
	userService  services.UserService
	auditService services.AuditService
	logger       *utils.NamedLogger
}

// NewAdminController creates a new admin controller
func NewAdminController(adminService services.AdminService, userService services.UserService, auditService services.AuditService) *AdminController {
	return &AdminController{
		adminService: adminService,
		userService:  userService,
		auditService: auditService,
		logger:       utils.NewLogger("admin-controller"),
	}
}

// GetUsers retrieves all users with filtering and pagination
func (ac *AdminController) GetUsers(c *gin.Context) {
	ac.logger.Info("Getting users with pagination", nil)

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

	// Get users with pagination using the admin service
	users, total, err := ac.adminService.GetUsers(page, limit, filter)
	if err != nil {
		ac.logger.Error("Failed to retrieve users", map[string]interface{}{"error": err.Error()})
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve users"})
		return
	}

	// Log the activity
	userID, exists := c.Get("userID")
	if exists {
		metadata := map[string]interface{}{
			"page":   page,
			"limit":  limit,
			"filter": filter,
		}

		if err := ac.auditService.LogActivity(
			userID.(string),
			"ADMIN_USERS_VIEWED",
			"User",
			"",
			metadata,
		); err != nil {
			ac.logger.Warn("Failed to log admin activity", map[string]interface{}{"error": err.Error()})
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"users":      users,
		"total":      total,
		"page":       page,
		"limit":      limit,
		"totalPages": (total + int64(limit) - 1) / int64(limit),
	})
}

// GetUserByID retrieves a user by ID
func (ac *AdminController) GetUserByID(c *gin.Context) {
	ac.logger.Info("Getting user by ID", nil)

	userID := c.Param("id")
	if userID == "" {
		ac.logger.Warn("User ID is required", nil)
		c.JSON(http.StatusBadRequest, gin.H{"error": "User ID is required"})
		return
	}

	user, err := ac.adminService.GetUserByID(userID)
	if err != nil {
		ac.logger.Error("User not found", map[string]interface{}{"userID": userID, "error": err.Error()})
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	// Log the activity
	requesterID, exists := c.Get("userID")
	if exists {
		metadata := map[string]interface{}{
			"targetUserID": userID,
		}

		if err := ac.auditService.LogActivity(
			requesterID.(string),
			"ADMIN_USER_VIEWED",
			"User",
			userID,
			metadata,
		); err != nil {
			ac.logger.Warn("Failed to log admin activity", map[string]interface{}{"error": err.Error()})
		}
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

	if err := ac.adminService.CreateUser(user); err != nil {
		if _, ok := err.(*utils.ConflictError); ok {
			c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user"})
		}
		return
	}

	// Get current admin's user ID from context
	adminID, _ := c.Get("userID")

	// Log the activity
	metadata := map[string]interface{}{
		"userName": user.Name,
		"userEmail": user.Email,
		"userRole": user.Role,
	}

	if err := ac.auditService.LogActivity(
		adminID.(string),
		"USER_CREATED",
		"User",
		user.ID,
		metadata,
	); err != nil {
		ac.logger.Warn("Failed to log user creation activity", map[string]interface{}{"error": err.Error()})
	}

	// Don't return password hash
	user.PasswordHash = ""
	c.JSON(http.StatusCreated, user)
}

// UpdateUser updates a user
func (ac *AdminController) UpdateUser(c *gin.Context) {
	id := c.Param("id")

	user, err := ac.adminService.GetUserByID(id)
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

	if err := ac.adminService.UpdateUser(user); err != nil {
		if _, ok := err.(*utils.ConflictError); ok {
			c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update user"})
		}
		return
	}

	// Get current user ID from context
	currentUserID, _ := c.Get("userID")

	// Log the activity
	metadata := map[string]interface{}{
		"userName": user.Name,
		"userEmail": user.Email,
		"userRole": user.Role,
	}

	if err := ac.auditService.LogActivity(
		currentUserID.(string),
		"USER_UPDATED",
		"User",
		user.ID,
		metadata,
	); err != nil {
		ac.logger.Warn("Failed to log user update activity", map[string]interface{}{"error": err.Error()})
	}

	// Don't return password hash in the response
	c.JSON(http.StatusOK, user)
}

// ResetUserPassword resets a user's password
func (ac *AdminController) ResetUserPassword(c *gin.Context) {
	id := c.Param("id")

	user, err := ac.adminService.GetUserByID(id)
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

	if err := ac.adminService.ResetUserPassword(id, hashedPassword); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to reset password"})
		return
	}

	// Get current admin's user ID from context
	currentUserID, _ := c.Get("userID")

	// Log the activity
	metadata := map[string]interface{}{
		"userEmail": user.Email,
	}

	if err := ac.auditService.LogActivity(
		currentUserID.(string),
		"USER_PASSWORD_RESET",
		"User",
		id,
		metadata,
	); err != nil {
		ac.logger.Warn("Failed to log password reset activity", map[string]interface{}{"error": err.Error()})
	}

	c.JSON(http.StatusOK, gin.H{"message": "Password reset successfully"})
}

// GetSystemInfo returns general system information and statistics
func (ac *AdminController) GetSystemInfo(c *gin.Context) {
	ac.logger.Info("Getting system information", nil)

	// Get system info from service
	info, err := ac.adminService.GetSystemInfo()
	if err != nil {
		ac.logger.Error("Failed to retrieve system information", map[string]interface{}{"error": err.Error()})
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve system information"})
		return
	}

	// Add additional information if needed
	info["serverTime"] = time.Now()

	// Log the activity
	userID, exists := c.Get("userID")
	if exists {
		if err := ac.auditService.LogActivity(
			userID.(string),
			"ADMIN_SYSTEM_INFO_VIEWED",
			"System",
			"",
			nil,
		); err != nil {
			ac.logger.Warn("Failed to log admin activity", map[string]interface{}{"error": err.Error()})
		}
	}

	c.JSON(http.StatusOK, info)
}
