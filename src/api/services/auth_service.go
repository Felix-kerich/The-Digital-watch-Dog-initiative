package services

import (
	"time"

	"github.com/google/uuid"
	"github.com/the-digital-watchdog-initiative/models"
	"github.com/the-digital-watchdog-initiative/repository"
	"github.com/the-digital-watchdog-initiative/utils"
	"golang.org/x/crypto/bcrypt"
)

// AuthServiceImpl implements AuthService interface
type AuthServiceImpl struct {
	userRepo    repository.UserRepository
	sessionRepo repository.SessionRepository
	auditRepo   repository.AuditLogRepository
	logger      *utils.NamedLogger
}

// NewAuthService creates a new auth service
func NewAuthService(
	userRepo repository.UserRepository,
	sessionRepo repository.SessionRepository,
	auditRepo repository.AuditLogRepository,
) AuthService {
	return &AuthServiceImpl{
		userRepo:    userRepo,
		sessionRepo: sessionRepo,
		auditRepo:   auditRepo,
		logger:      utils.NewLogger("auth-service"),
	}
}

// Register registers a new user
func (s *AuthServiceImpl) Register(name, email, password, entityID string) (*models.User, error) {
	s.logger.Info("Registering new user", map[string]interface{}{"email": email})

	// Check if email is already in use
	existingUser, _ := s.userRepo.FindByEmail(email)
	if existingUser != nil {
		return nil, utils.NewConflictError("Email already in use", "email")
	}

	// Hash the password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		s.logger.Error("Failed to hash password", map[string]interface{}{
			"email": email,
			"error": err.Error(),
		})
		return nil, err
	}

	// Create the user
	user := &models.User{
		ID:           uuid.New().String(),
		Name:         name,
		Email:        email,
		PasswordHash: string(hashedPassword),
		EntityID:     entityID,
		Role:         models.RolePublic, // Default role
		IsActive:     true,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	if err := s.userRepo.Create(user); err != nil {
		s.logger.Error("Failed to create user", map[string]interface{}{
			"email": email,
			"error": err.Error(),
		})
		return nil, err
	}

	// Log the registration
	s.auditRepo.Create(&models.AuditLog{
		UserID:     user.ID,
		Action:     "REGISTER",
		EntityType: "USER",
		EntityID:   user.ID,
		Timestamp:  time.Now(),
	})

	// Don't expose password hash
	user.PasswordHash = ""

	return user, nil
}

// Login authenticates a user and creates a session
func (s *AuthServiceImpl) Login(email, password string) (*models.UserSession, error) {
	s.logger.Info("User login attempt", map[string]interface{}{"email": email})

	// Find the user by email
	user, err := s.userRepo.FindByEmail(email)
	if err != nil {
		s.logger.Warn("Login failed: user not found", map[string]interface{}{"email": email})
		return nil, utils.ErrInvalidCredentials
	}

	// Check if user is active
	if !user.IsActive {
		s.logger.Warn("Login failed: user account not active", map[string]interface{}{
			"email":  email,
			"status": user.IsActive,
		})
		return nil, utils.ErrUnauthorized
	}

	// Verify password
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		s.logger.Warn("Login failed: invalid password", map[string]interface{}{"email": email})
		return nil, utils.ErrInvalidCredentials
	}

	// Generate tokens
	accessToken, err := utils.GenerateTokens(user)
	if err != nil {
		s.logger.Error("Failed to generate access token", map[string]interface{}{
			"userID": user.ID,
			"error":  err.Error(),
		})
		return nil, err
	}

	// Create session
	session := &models.UserSession{
		ID:           uuid.New().String(),
		UserID:       user.ID,
		Token:        accessToken.AccessToken,
		RefreshToken: accessToken.RefreshToken,
		ExpiresAt:    time.Unix(accessToken.AtExpires, 0),
		CreatedAt:    time.Now(),
	}

	if err := s.sessionRepo.Create(session); err != nil {
		s.logger.Error("Failed to create session", map[string]interface{}{
			"userID": user.ID,
			"error":  err.Error(),
		})
		return nil, err
	}

	// Update last login time
	s.userRepo.UpdateLastLogin(user.ID, time.Now())

	// Log the login
	s.auditRepo.Create(&models.AuditLog{
		UserID:     user.ID,
		Action:     "LOGIN",
		EntityType: "USER",
		EntityID:   user.ID,
		Timestamp:  time.Now(),
	})

	return session, nil
}

// RefreshToken refreshes an access token using a refresh token
func (s *AuthServiceImpl) RefreshToken(refreshToken string) (*models.UserSession, error) {
	s.logger.Info("Refreshing token", map[string]interface{}{"refreshToken": refreshToken})

	// Find session by refresh token
	session, err := s.sessionRepo.FindByRefreshToken(refreshToken)
	if err != nil {
		s.logger.Warn("Token refresh failed: invalid refresh token", map[string]interface{}{
			"refreshToken": refreshToken,
		})
		return nil, utils.ErrUnauthorized
	}

	// Check if session is expired
	if session.ExpiresAt.Before(time.Now()) {
		s.logger.Warn("Token refresh failed: session expired", map[string]interface{}{
			"sessionID": session.ID,
			"userID":    session.UserID,
		})
		return nil, utils.ErrUnauthorized
	}

	// Get the user
	user, err := s.userRepo.FindByID(session.UserID)
	if err != nil {
		s.logger.Error("Token refresh failed: user not found", map[string]interface{}{
			"userID": session.UserID,
			"error":  err.Error(),
		})
		return nil, utils.ErrUnauthorized
	}

	// Check if user is active
	if !user.IsActive {
		s.logger.Warn("Token refresh failed: user account not active", map[string]interface{}{
			"userID": user.ID,
			"status": user.IsActive,
		})
		return nil, utils.ErrUnauthorized
	}

	// Generate new access token
	newAccessToken, err := utils.GenerateTokens(user)
	if err != nil {
		s.logger.Error("Failed to generate new access token", map[string]interface{}{
			"userID": user.ID,
			"error":  err.Error(),
		})
		return nil, err
	}

	// Generate new refresh token
	newRefreshToken := uuid.New().String()

	// Update session
	session.Token = newAccessToken.AccessToken
	session.RefreshToken = newRefreshToken
	session.ExpiresAt = time.Unix(newAccessToken.AtExpires, 0)

	// Create a new session instead of updating the old one
	newSession := &models.UserSession{
		ID:           uuid.New().String(),
		UserID:       user.ID,
		Token:        newAccessToken.AccessToken,
		RefreshToken: newRefreshToken,
		ExpiresAt:    time.Unix(newAccessToken.AtExpires, 0),
		CreatedAt:    time.Now(),
	}

	if err := s.sessionRepo.Create(newSession); err != nil {
		s.logger.Error("Failed to create new session", map[string]interface{}{
			"userID": user.ID,
			"error":  err.Error(),
		})
		return nil, err
	}

	// Revoke old session
	s.sessionRepo.RevokeByToken(session.Token)

	return newSession, nil
}

// Logout invalidates a user's session
func (s *AuthServiceImpl) Logout(token string) error {
	s.logger.Info("User logout", map[string]interface{}{"token": token})

	// Find session by token
	session, err := s.sessionRepo.FindByToken(token)
	if err != nil {
		// If session not found, consider logout successful
		return nil
	}

	// Revoke the session
	if err := s.sessionRepo.RevokeByToken(token); err != nil {
		s.logger.Error("Failed to revoke session", map[string]interface{}{
			"sessionID": session.ID,
			"userID":    session.UserID,
			"error":     err.Error(),
		})
		return err
	}

	// Log the logout
	s.auditRepo.Create(&models.AuditLog{
		UserID:     session.UserID,
		Action:     "LOGOUT",
		EntityType: "USER",
		EntityID:   session.UserID,
		Timestamp:  time.Now(),
	})

	return nil
}

// ValidateToken validates a JWT token and returns the associated user
func (s *AuthServiceImpl) ValidateToken(token string) (*models.User, error) {
	s.logger.Info("Validating token", map[string]interface{}{"token": token})

	// Find session by token
	session, err := s.sessionRepo.FindByToken(token)
	if err != nil {
		s.logger.Warn("Token validation failed: session not found", map[string]interface{}{
			"token": token,
		})
		return nil, utils.ErrUnauthorized
	}

	// Check if session is expired
	if session.ExpiresAt.Before(time.Now()) {
		s.logger.Warn("Token validation failed: session expired", map[string]interface{}{
			"sessionID": session.ID,
			"userID":    session.UserID,
		})
		return nil, utils.ErrUnauthorized
	}

	// Parse and validate the JWT
	claims, err := utils.ValidateToken(token)
	if err != nil {
		s.logger.Warn("Token validation failed: invalid JWT", map[string]interface{}{
			"token": token,
			"error": err.Error(),
		})
		return nil, utils.ErrUnauthorized
	}

	// Get the user
	user, err := s.userRepo.FindByID(claims.UserID)
	if err != nil {
		s.logger.Error("Token validation failed: user not found", map[string]interface{}{
			"userID": claims.UserID,
			"error":  err.Error(),
		})
		return nil, utils.ErrUnauthorized
	}

	// Check if user is active
	if !user.IsActive {
		s.logger.Warn("Token validation failed: user account not active", map[string]interface{}{
			"userID": user.ID,
			"status": user.IsActive,
		})
		return nil, utils.ErrUnauthorized
	}

	// Don't expose password hash
	user.PasswordHash = ""

	return user, nil
}
