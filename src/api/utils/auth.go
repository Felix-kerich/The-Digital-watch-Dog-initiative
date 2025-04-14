package utils

import (
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/the-digital-watchdog-initiative/models"

	"golang.org/x/crypto/bcrypt"
)

// Common errors
var (
	ErrJWTSecretNotSet = errors.New("JWT_SECRET is not set in environment")
	ErrInvalidToken    = errors.New("invalid token")
	ErrTokenExpired    = errors.New("token has expired")
	ErrInvalidClaims   = errors.New("invalid token claims")
)

// CustomClaims extends standard JWT claims with user information
type CustomClaims struct {
	UserID string          `json:"userId"`
	Email  string          `json:"email"`
	Role   models.UserRole `json:"role"`
	jwt.RegisteredClaims
}

// TokenDetails contains token information
type TokenDetails struct {
	AccessToken  string
	RefreshToken string
	AccessUUID   string
	RefreshUUID  string
	AtExpires    int64
	RtExpires    int64
}

// GenerateTokens creates new access and refresh tokens for a user
func GenerateTokens(user *models.User) (*TokenDetails, error) {
	// Get the JWT secret from environment
	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		return nil, ErrJWTSecretNotSet
	}

	// Create token details
	td := &TokenDetails{
		AccessUUID:  uuid.New().String(),
		RefreshUUID: uuid.New().String(),
	}

	// Get token expiry times from environment or use defaults
	accessExpiry := os.Getenv("JWT_ACCESS_EXPIRY")
	if accessExpiry == "" {
		accessExpiry = "24h" // Default to 24 hours
	}

	refreshExpiry := os.Getenv("JWT_REFRESH_EXPIRY")
	if refreshExpiry == "" {
		refreshExpiry = "168h" // Default to 7 days
	}

	// Parse access token expiry duration
	atDuration, err := time.ParseDuration(accessExpiry)
	if err != nil {
		return nil, fmt.Errorf("invalid access token expiry duration: %w", err)
	}
	td.AtExpires = time.Now().Add(atDuration).Unix()

	// Parse refresh token expiry duration
	rtDuration, err := time.ParseDuration(refreshExpiry)
	if err != nil {
		return nil, fmt.Errorf("invalid refresh token expiry duration: %w", err)
	}
	td.RtExpires = time.Now().Add(rtDuration).Unix()

	// Create access token claims
	atClaims := CustomClaims{
		UserID: user.ID,
		Email:  user.Email,
		Role:   user.Role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Unix(td.AtExpires, 0)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    "api-service",
			Subject:   user.ID,
			ID:        td.AccessUUID,
		},
	}

	// Create refresh token claims with longer expiry
	rtClaims := jwt.RegisteredClaims{
		ExpiresAt: jwt.NewNumericDate(time.Unix(td.RtExpires, 0)),
		IssuedAt:  jwt.NewNumericDate(time.Now()),
		NotBefore: jwt.NewNumericDate(time.Now()),
		Issuer:    "api-service",
		Subject:   user.ID,
		ID:        td.RefreshUUID,
	}

	// Create the tokens
	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, atClaims)
	refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, rtClaims)

	// Sign the tokens
	td.AccessToken, err = accessToken.SignedString([]byte(jwtSecret))
	if err != nil {
		return nil, fmt.Errorf("failed to sign access token: %w", err)
	}

	td.RefreshToken, err = refreshToken.SignedString([]byte(jwtSecret))
	if err != nil {
		return nil, fmt.Errorf("failed to sign refresh token: %w", err)
	}

	return td, nil
}

// ValidateToken validates an access token
func ValidateToken(tokenString string) (*CustomClaims, error) {
	// Get the JWT secret from environment
	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		return nil, ErrJWTSecretNotSet
	}

	// Parse and validate the token
	token, err := jwt.ParseWithClaims(tokenString, &CustomClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return []byte(jwtSecret), nil
	})

	if err != nil {
		// In jwt/v5, validation errors are wrapped in the error, so we need to check via errors.Is
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, ErrTokenExpired
		}
		return nil, err
	}

	// Extract claims
	if claims, ok := token.Claims.(*CustomClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, ErrInvalidClaims
}

// RefreshAccessToken creates a new access token given a valid refresh token
func RefreshAccessToken(refreshToken string) (*TokenDetails, error) {
	// Get the JWT secret from environment
	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		return nil, ErrJWTSecretNotSet
	}

	// Parse the refresh token
	token, err := jwt.ParseWithClaims(refreshToken, &jwt.RegisteredClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return []byte(jwtSecret), nil
	})

	if err != nil {
		return nil, err
	}

	// Get claims from token
	if claims, ok := token.Claims.(*jwt.RegisteredClaims); ok && token.Valid {
		// Get user by ID
		var user models.User
		if err := DB.First(&user, "id = ?", claims.Subject).Error; err != nil {
			return nil, errors.New("user not found")
		}

		// Generate new tokens
		return GenerateTokens(&user)
	}

	return nil, ErrInvalidToken
}

// GeneratePasswordHash creates a bcrypt hash of a password
func GeneratePasswordHash(password string) (string, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hashedPassword), nil
}

// CheckPasswordHash compares a password with a hash
func CheckPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}
