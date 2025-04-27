package handlers

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/the-digital-watchdog-initiative/models"
	"github.com/the-digital-watchdog-initiative/services"
	"github.com/the-digital-watchdog-initiative/utils"
)

// AuthController handles authentication-related operations
type AuthController struct {
	authService services.AuthService
	auditService services.AuditService
	logger      *utils.NamedLogger
}

// RegisterRequest represents user registration data
type RegisterRequest struct {
	Name     string `json:"name" binding:"required"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=8"`
	EntityID string `json:"entityId"`
}

// LoginRequest represents user login data
type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

// TokenResponse represents the response with JWT tokens
type TokenResponse struct {
	AccessToken  string `json:"accessToken"`
	RefreshToken string `json:"refreshToken"`
	ExpiresAt    int64  `json:"expiresAt"`
	User         struct {
		ID    string          `json:"id"`
		Name  string          `json:"name"`
		Email string          `json:"email"`
		Role  models.UserRole `json:"role"`
	} `json:"user"`
}

// NewAuthController creates a new auth controller
func NewAuthController(authService services.AuthService, auditService services.AuditService) *AuthController {
	return &AuthController{
		authService: authService,
		auditService: auditService,
		logger:      utils.NewLogger("auth-handler"),
	}
}

// Register handles user registration
func (ac *AuthController) Register(c *gin.Context) {
	var req RegisterRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		ac.logger.Warn("Invalid registration request", map[string]interface{}{
			"error": err.Error(),
		})
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ac.logger.Info("Processing registration request", map[string]interface{}{
		"email": req.Email,
		"name":  req.Name,
	})

	// Register the user using the auth service
	user, err := ac.authService.Register(req.Name, req.Email, req.Password, req.EntityID)
	if err != nil {
		ac.logger.Error("Registration failed", map[string]interface{}{
			"email": req.Email,
			"error": err.Error(),
		})

		// Check for specific error types
		if conflictErr, ok := err.(*utils.ConflictError); ok {
			c.JSON(http.StatusConflict, gin.H{"error": conflictErr.Error()})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to register user"})
		return
	}

	// Login the user to get a session
	session, err := ac.authService.Login(req.Email, req.Password)
	if err != nil {
		ac.logger.Error("Auto-login after registration failed", map[string]interface{}{
			"email": req.Email,
			"error": err.Error(),
		})
		c.JSON(http.StatusCreated, gin.H{
			"user":    user,
			"message": "Registration successful but login failed. Please login manually.",
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"user":         user,
		"accessToken":  session.Token,
		"refreshToken": session.RefreshToken,
		"expiresAt":    session.ExpiresAt,
	})
}

// Login handles user login
func (ac *AuthController) Login(c *gin.Context) {
	var req LoginRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		ac.logger.Warn("Invalid login request", map[string]interface{}{
			"error": err.Error(),
		})
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ac.logger.Info("Processing login request", map[string]interface{}{
		"email": req.Email,
	})

	// Login the user using the auth service
	session, err := ac.authService.Login(req.Email, req.Password)
	if err != nil {
		ac.logger.Error("Login failed", map[string]interface{}{
			"email": req.Email,
			"error": err.Error(),
		})

		// Check for specific error types
		if err == utils.ErrInvalidCredentials {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
			return
		} else if err == utils.ErrUnauthorized {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Account is not active"})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to login"})
		return
	}

	// Get user details
	user, err := ac.authService.ValidateToken(session.Token)
	if err != nil {
		ac.logger.Error("Failed to get user details after login", map[string]interface{}{
			"email": req.Email,
			"error": err.Error(),
		})
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get user details"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"user":         user,
		"accessToken":  session.Token,
		"refreshToken": session.RefreshToken,
		"expiresAt":    session.ExpiresAt,
	})
}

// RefreshToken refreshes an access token using a refresh token
func (ac *AuthController) RefreshToken(c *gin.Context) {
	var req struct {
		RefreshToken string `json:"refreshToken" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		ac.logger.Warn("Invalid refresh token request", map[string]interface{}{
			"error": err.Error(),
		})
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ac.logger.Info("Processing refresh token request", nil)

	// Refresh the token using the auth service
	session, err := ac.authService.RefreshToken(req.RefreshToken)
	if err != nil {
		ac.logger.Error("Token refresh failed", map[string]interface{}{
			"error": err.Error(),
		})

		// Check for specific error types
		if err == utils.ErrUnauthorized {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired refresh token"})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to refresh token"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"accessToken":  session.Token,
		"refreshToken": session.RefreshToken,
		"expiresAt":    session.ExpiresAt,
	})
}

// Logout handles user logout
func (ac *AuthController) Logout(c *gin.Context) {
	// Get token from Authorization header
	token := c.GetHeader("Authorization")
	if token == "" {
		ac.logger.Warn("Logout attempt with no token", nil)
		c.JSON(http.StatusBadRequest, gin.H{"error": "No token provided"})
		return
	}

	// Remove Bearer prefix if present
	token = strings.TrimPrefix(token, "Bearer ")

	ac.logger.Info("Processing logout request", nil)

	// Logout using the auth service
	err := ac.authService.Logout(token)
	if err != nil {
		ac.logger.Error("Logout failed", map[string]interface{}{
			"error": err.Error(),
		})
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to logout"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Logged out successfully"})
}
