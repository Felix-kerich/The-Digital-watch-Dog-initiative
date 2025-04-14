package controllers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/the-digital-watchdog-initiative/models"
	"github.com/the-digital-watchdog-initiative/repository"
	"github.com/the-digital-watchdog-initiative/utils"
)

// UserManagementController handles user management operations
type UserManagementController struct {
	userRepo  repository.UserRepository
	auditRepo repository.AuditLogRepository
}

// CreateUserRequest represents the request to create a new user
type CreateUserRequest struct {
	Name     string          `json:"name" binding:"required"`
	Email    string          `json:"email" binding:"required,email"`
	Password string          `json:"password" binding:"required,min=8"`
	Role     models.UserRole `json:"role" binding:"required"`
	EntityID string          `json:"entityId"`
}

// NewUserManagementController creates a new user management controller
func NewUserManagementController() *UserManagementController {
	return &UserManagementController{
		userRepo:  repository.NewUserRepository(),
		auditRepo: repository.NewAuditLogRepository(),
	}
}

// CreateUser creates a new user (admin only)
func (uc *UserManagementController) CreateUser(c *gin.Context) {
	// Check if the current user is an admin
	currentUserRole, exists := c.Get("userRole")
	if !exists || currentUserRole != models.RoleAdmin {
		c.JSON(http.StatusForbidden, gin.H{"error": "Only administrators can create users"})
		return
	}

	var req CreateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Validate role
	switch req.Role {
	case models.RoleAdmin, models.RoleAuditor, models.RoleFinanceOfficer, models.RoleManager, models.RolePublic:
		// Valid role
	default:
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid role specified"})
		return
	}

	// Check if email already exists
	existingUser, err := uc.userRepo.FindByEmail(req.Email)
	if err == nil && existingUser != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Email already registered"})
		return
	}

	// Hash password
	hashedPassword, err := utils.GeneratePasswordHash(req.Password)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to hash password"})
		return
	}

	// Create user
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

	if err := uc.userRepo.Create(user); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user"})
		return
	}

	// Create audit log
	currentUserID, _ := c.Get("userID")
	auditLog := &models.AuditLog{
		Action:     "USER_CREATED",
		UserID:     currentUserID.(string),
		EntityType: "User",
		EntityID:   user.ID,
		Detail:     "Admin created user: " + user.Email + " with role: " + string(user.Role),
		IP:         c.ClientIP(),
		Timestamp:  time.Now(),
	}

	if err := uc.auditRepo.Create(auditLog); err != nil {
		utils.Logger.Warn("Failed to create audit log for user creation")
	}

	c.JSON(http.StatusCreated, gin.H{
		"id":    user.ID,
		"name":  user.Name,
		"email": user.Email,
		"role":  user.Role,
	})
}

// GetUsers retrieves all users (admin only)
func (uc *UserManagementController) GetUsers(c *gin.Context) {
	// Check if the current user is an admin
	currentUserRole, exists := c.Get("userRole")
	if !exists || currentUserRole != models.RoleAdmin {
		c.JSON(http.StatusForbidden, gin.H{"error": "Only administrators can view all users"})
		return
	}

	users, err := uc.userRepo.FindAll()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve users"})
		return
	}

	c.JSON(http.StatusOK, users)
}
