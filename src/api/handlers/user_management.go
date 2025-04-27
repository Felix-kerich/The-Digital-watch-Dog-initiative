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

// UserManagementController handles user management operations
type UserManagementController struct {
	userService  services.UserService
	auditService services.AuditService
	logger       *utils.NamedLogger
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
func NewUserManagementController(userService services.UserService, auditService services.AuditService) *UserManagementController {
	return &UserManagementController{
		userService:  userService,
		auditService: auditService,
		logger:       utils.NewLogger("user-management-controller"),
	}
}

// Create creates a new user (admin only)
func (uc *UserManagementController) Create(c *gin.Context) {
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

	// Extract admin ID from context
	adminID, exists := c.Get("userID")
	if !exists {
		uc.logger.Error("Admin ID not found in context", nil)
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}
	adminIDStr := adminID.(string)

	// Create user
	user := &models.User{
		Name:      req.Name,
		Email:     req.Email,
		Role:      req.Role,
		EntityID:  req.EntityID,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// Hash password and create user
	hashedPassword, err := utils.GeneratePasswordHash(req.Password)
	if err != nil {
		uc.logger.Error("Failed to hash password", map[string]interface{}{"error": err.Error()})
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to hash password"})
		return
	}
	user.PasswordHash = hashedPassword

	// Create user using service
	if err := uc.userService.CreateUser(user); err != nil {
		uc.logger.Error("Failed to create user", map[string]interface{}{"error": err.Error()})
		
		// Check for specific error types
		if err.Error() == "user with this email already exists" {
			c.JSON(http.StatusConflict, gin.H{"error": "User with this email already exists"})
			return
		}
		
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user"})
		return
	}

	// Log the activity
	if err := uc.auditService.LogActivity(adminIDStr, "USER_CREATED", "User", user.ID, map[string]interface{}{
		"name": user.Name,
		"email": user.Email,
		"role": string(user.Role),
	}); err != nil {
		// Log error but continue
		uc.logger.Error("Failed to log user creation activity", map[string]interface{}{"error": err.Error()})
	}

	// Return user without password hash
	user.PasswordHash = ""
	uc.logger.Info("User created successfully", map[string]interface{}{"id": user.ID})
	c.JSON(http.StatusCreated, gin.H{"data": user, "message": "User created successfully"})
}

// GetAll retrieves all users with pagination and filtering
func (uc *UserManagementController) GetAll(c *gin.Context) {
	uc.logger.Info("Handling get all users request", nil)
	
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))

	// Apply filters
	filter := map[string]interface{}{}
	if role := c.Query("role"); role != "" {
		filter["role"] = role
	}
	if entityID := c.Query("entityId"); entityID != "" {
		filter["entity_id"] = entityID
	}

	uc.logger.Info("Retrieving users", map[string]interface{}{
		"page": page,
		"limit": limit,
		"filter": filter,
	})

	users, total, err := uc.userService.GetUsers(page, limit, filter)
	if err != nil {
		uc.logger.Error("Failed to retrieve users", map[string]interface{}{"error": err.Error()})
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve users"})
		return
	}

	// Remove password hashes
	for i := range users {
		users[i].PasswordHash = ""
	}

	uc.logger.Info("Users retrieved successfully", map[string]interface{}{"count": len(users)})
	c.JSON(http.StatusOK, gin.H{
		"data":  users,
		"total": total,
		"page":  page,
		"limit": limit,
	})
}
