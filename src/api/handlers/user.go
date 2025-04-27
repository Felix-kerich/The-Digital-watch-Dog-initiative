package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/the-digital-watchdog-initiative/services"
	"github.com/the-digital-watchdog-initiative/utils"
)

// UserController handles user-related operations
type UserController struct {
	userService services.UserService
	logger      *utils.NamedLogger
}

// NewUserController creates a new user controller
func NewUserController(userService services.UserService) *UserController {
	return &UserController{
		userService: userService,
		logger:      utils.NewLogger("user-handler"),
	}
}

// GetProfile retrieves the authenticated user's profile
func (uc *UserController) GetProfile(c *gin.Context) {
	userID, _ := c.Get("userID")

	uc.logger.Info("Getting user profile", map[string]interface{}{"userID": userID})

	user, err := uc.userService.GetUserProfile(userID.(string))
	if err != nil {
		uc.logger.Error("Failed to get user profile", map[string]interface{}{
			"userID": userID,
			"error":  err.Error(),
		})
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	c.JSON(http.StatusOK, user)
}

// UpdateProfile updates the authenticated user's profile
func (uc *UserController) UpdateProfile(c *gin.Context) {
	userID, _ := c.Get("userID")

	var req struct {
		Name  string `json:"name"`
		Email string `json:"email"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		uc.logger.Warn("Invalid update profile request", map[string]interface{}{
			"userID": userID,
			"error":  err.Error(),
		})
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	uc.logger.Info("Updating user profile", map[string]interface{}{
		"userID": userID,
		"name":   req.Name,
		"email":  req.Email,
	})

	// Create update data map
	updateData := map[string]interface{}{}
	if req.Name != "" {
		updateData["name"] = req.Name
	}
	if req.Email != "" {
		updateData["email"] = req.Email
	}

	updatedUser, err := uc.userService.UpdateUserProfile(userID.(string), updateData)
	if err != nil {
		uc.logger.Error("Failed to update user profile", map[string]interface{}{
			"userID": userID,
			"error":  err.Error(),
		})

		// Check for specific error types
		if conflictErr, ok := err.(*utils.ConflictError); ok {
			c.JSON(http.StatusConflict, gin.H{"error": conflictErr.Error()})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update profile"})
		return
	}

	c.JSON(http.StatusOK, updatedUser)
}

// ChangePassword changes the authenticated user's password
func (uc *UserController) ChangePassword(c *gin.Context) {
	userID, _ := c.Get("userID")

	var req struct {
		CurrentPassword string `json:"currentPassword" binding:"required"`
		NewPassword     string `json:"newPassword" binding:"required,min=8"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		uc.logger.Warn("Invalid change password request", map[string]interface{}{
			"userID": userID,
			"error":  err.Error(),
		})
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	uc.logger.Info("Changing user password", map[string]interface{}{"userID": userID})

	err := uc.userService.ChangeUserPassword(
		userID.(string),
		req.CurrentPassword,
		req.NewPassword,
	)

	if err != nil {
		uc.logger.Error("Failed to change password", map[string]interface{}{
			"userID": userID,
			"error":  err.Error(),
		})
		
		// Check for specific error types
		if err == utils.ErrInvalidCredentials {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Current password is incorrect"})
			return
		}
		
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update password"})
		return
	}


	c.JSON(http.StatusOK, gin.H{"message": "Password updated successfully"})
}
