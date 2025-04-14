package controllers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/the-digital-watchdog-initiative/models"
	"github.com/the-digital-watchdog-initiative/repository"
	"github.com/the-digital-watchdog-initiative/utils"
)

// AuthController handles authentication-related operations
type AuthController struct {
	userRepo    repository.UserRepository
	sessionRepo repository.SessionRepository
	auditRepo   repository.AuditLogRepository
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
func NewAuthController() *AuthController {
	return &AuthController{
		userRepo:    repository.NewUserRepository(),
		sessionRepo: repository.NewSessionRepository(),
		auditRepo:   repository.NewAuditLogRepository(),
	}
}

// Register registers a new user
func (ac *AuthController) Register(c *gin.Context) {
	var req struct {
		Name     string `json:"name" binding:"required"`
		Email    string `json:"email" binding:"required,email"`
		Password string `json:"password" binding:"required,min=8"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Check if email already exists
	existingUser, err := ac.userRepo.FindByEmail(req.Email)
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
		Role:         models.RolePublic,
		IsActive:     true,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	if err := ac.userRepo.Create(user); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user"})
		return
	}

	// Create audit log
	auditLog := &models.AuditLog{
		Action:     "USER_REGISTERED",
		UserID:     user.ID,
		EntityType: "User",
		EntityID:   user.ID,
		Detail:     "User registered: " + user.Email,
		IP:         c.ClientIP(),
		Timestamp:  time.Now(),
	}

	if err := ac.auditRepo.Create(auditLog); err != nil {
		utils.Logger.Warn("Failed to create audit log for user registration")
	}

	c.JSON(http.StatusCreated, gin.H{
		"id":    user.ID,
		"name":  user.Name,
		"email": user.Email,
		"role":  user.Role,
	})
}

// Login authenticates a user
func (ac *AuthController) Login(c *gin.Context) {
	var req struct {
		Email    string `json:"email" binding:"required,email"`
		Password string `json:"password" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Find user by email
	user, err := ac.userRepo.FindByEmail(req.Email)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}

	// Check if user is active
	if !user.IsActive {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Account is inactive"})
		return
	}

	// Check password
	if !utils.CheckPasswordHash(req.Password, user.PasswordHash) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}

	// Generate tokens
	tokenDetails, err := utils.GenerateTokens(user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
		return
	}

	// Create session
	session := &models.UserSession{
		UserID:       user.ID,
		Token:        tokenDetails.AccessToken,
		RefreshToken: tokenDetails.RefreshToken,
		ExpiresAt:    time.Unix(tokenDetails.AtExpires, 0),
		IP:           c.ClientIP(),
		UserAgent:    c.Request.UserAgent(),
		CreatedAt:    time.Now(),
	}

	if err := ac.sessionRepo.Create(session); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create session"})
		return
	}

	// Update last login time
	if err := ac.userRepo.UpdateLastLogin(user.ID, time.Now()); err != nil {
		utils.Logger.Warn("Failed to update last login time")
	}

	// Create audit log
	auditLog := &models.AuditLog{
		Action:     "USER_LOGGED_IN",
		UserID:     user.ID,
		EntityType: "User",
		EntityID:   user.ID,
		Detail:     "User logged in: " + user.Email,
		IP:         c.ClientIP(),
		Timestamp:  time.Now(),
	}

	if err := ac.auditRepo.Create(auditLog); err != nil {
		utils.Logger.Warn("Failed to create audit log for user login")
	}

	c.JSON(http.StatusOK, gin.H{
		"accessToken":  tokenDetails.AccessToken,
		"refreshToken": tokenDetails.RefreshToken,
		"expiresAt":    tokenDetails.AtExpires,
		"user": gin.H{
			"id":    user.ID,
			"name":  user.Name,
			"email": user.Email,
			"role":  user.Role,
		},
	})
}

// RefreshToken refreshes an access token
func (ac *AuthController) RefreshToken(c *gin.Context) {
	var req struct {
		RefreshToken string `json:"refreshToken" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Find session by refresh token
	session, err := ac.sessionRepo.FindByRefreshToken(req.RefreshToken)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid refresh token"})
		return
	}

	// Find user
	user, err := ac.userRepo.FindByID(session.UserID)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found"})
		return
	}

	// Check if user is active
	if !user.IsActive {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Account is inactive"})
		return
	}

	// Generate new tokens
	tokenDetails, err := utils.GenerateTokens(user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
		return
	}

	// Revoke the old session
	if err := ac.sessionRepo.RevokeByToken(session.Token); err != nil {
		utils.Logger.Warn("Failed to revoke old session")
	}

	// Create new session
	newSession := &models.UserSession{
		UserID:       user.ID,
		Token:        tokenDetails.AccessToken,
		RefreshToken: tokenDetails.RefreshToken,
		ExpiresAt:    time.Unix(tokenDetails.AtExpires, 0),
		IP:           c.ClientIP(),
		UserAgent:    c.Request.UserAgent(),
		CreatedAt:    time.Now(),
	}

	if err := ac.sessionRepo.Create(newSession); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create session"})
		return
	}

	// Create audit log
	auditLog := &models.AuditLog{
		Action:     "TOKEN_REFRESHED",
		UserID:     user.ID,
		EntityType: "User",
		EntityID:   user.ID,
		Detail:     "User refreshed token: " + user.Email,
		IP:         c.ClientIP(),
		Timestamp:  time.Now(),
	}

	if err := ac.auditRepo.Create(auditLog); err != nil {
		utils.Logger.Warn("Failed to create audit log for token refresh")
	}

	c.JSON(http.StatusOK, gin.H{
		"accessToken":  tokenDetails.AccessToken,
		"refreshToken": tokenDetails.RefreshToken,
		"expiresAt":    tokenDetails.AtExpires,
	})
}

// Logout logs out a user
func (ac *AuthController) Logout(c *gin.Context) {
	token := c.GetHeader("Authorization")
	if token == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Authorization header required"})
		return
	}

	// Remove "Bearer " prefix if present
	if len(token) > 7 && token[:7] == "Bearer " {
		token = token[7:]
	}

	// Revoke the session
	if err := ac.sessionRepo.RevokeByToken(token); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to logout"})
		return
	}

	// Get user ID from context (set by auth middleware)
	userID, exists := c.Get("userID")
	if exists {
		// Create audit log
		auditLog := &models.AuditLog{
			Action:     "USER_LOGGED_OUT",
			UserID:     userID.(string),
			EntityType: "User",
			EntityID:   userID.(string),
			Detail:     "User logged out",
			IP:         c.ClientIP(),
			Timestamp:  time.Now(),
		}

		if err := ac.auditRepo.Create(auditLog); err != nil {
			utils.Logger.Warn("Failed to create audit log for user logout")
		}
	}

	c.JSON(http.StatusOK, gin.H{"message": "Successfully logged out"})
}
