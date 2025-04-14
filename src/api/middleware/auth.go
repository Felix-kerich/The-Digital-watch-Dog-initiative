package middleware

import (
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/the-digital-watchdog-initiative/models"
	"github.com/the-digital-watchdog-initiative/utils"
)

// RequireAuth middleware verifies JWT token and authorizes the request
func RequireAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get the Authorization header
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			utils.Logger.Warn("Missing Authorization header")
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header is required"})
			c.Abort()
			return
		}

		// Check if the header has the Bearer format
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			utils.Logger.Warn("Invalid Authorization header format")
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header format must be Bearer {token}"})
			c.Abort()
			return
		}

		// Validate the token
		claims, err := utils.ValidateToken(parts[1])
		if err != nil {
			switch err {
			case utils.ErrTokenExpired:
				utils.Logger.Warnf("Token expired for user: %s", claims.UserID)
				c.JSON(http.StatusUnauthorized, gin.H{"error": "Token has expired", "code": "token_expired"})
			case utils.ErrJWTSecretNotSet:
				utils.Logger.Error("JWT secret not set in environment")
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
			default:
				utils.Logger.Warnf("Invalid token: %v", err)
				c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired token"})
			}
			c.Abort()
			return
		}

		// Check if user exists and is active
		var user models.User
		if result := utils.DB.First(&user, "id = ?", claims.UserID); result.Error != nil {
			utils.Logger.Warnf("User not found: %s", claims.UserID)
			c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found or inactive"})
			c.Abort()
			return
		}

		if !user.IsActive {
			utils.Logger.Warnf("Inactive user attempting access: %s", claims.UserID)
			c.JSON(http.StatusForbidden, gin.H{"error": "Account is inactive"})
			c.Abort()
			return
		}

		// Update last login time if needed (not on every request to avoid excessive DB writes)
		// Only update if last login was more than 1 hour ago
		if user.LastLogin == nil || time.Since(*user.LastLogin) > time.Hour {
			now := time.Now()
			user.LastLogin = &now
			utils.DB.Save(&user)
		}

		// Set user ID and role in context for use in handlers
		c.Set("userID", claims.UserID)
		c.Set("userEmail", claims.Email)
		c.Set("userRole", claims.Role)
		c.Set("user", user) // Set the entire user object for convenience

		c.Next()
	}
}

// RequireRole middleware ensures the user has the required role
func RequireRole(roles ...models.UserRole) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get user role from context (set by RequireAuth middleware)
		userRole, exists := c.Get("userRole")
		if !exists {
			utils.Logger.Warn("User not authenticated in RequireRole middleware")
			c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
			c.Abort()
			return
		}

		// Check if user role is in the allowed roles
		roleMatched := false
		for _, role := range roles {
			if userRole == role {
				roleMatched = true
				break
			}
		}

		if !roleMatched {
			userID, _ := c.Get("userID")
			utils.Logger.Warnf("Insufficient permissions for user %s with role %s", userID, userRole)
			c.JSON(http.StatusForbidden, gin.H{"error": "Insufficient permissions"})
			c.Abort()
			return
		}

		c.Next()
	}
}

// AuditLog middleware logs user actions to the audit log
func AuditLog(action string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Execute the request handler first
		c.Next()

		// Only log successful requests
		if c.Writer.Status() < 400 {
			// Get user ID from context
			userID, exists := c.Get("userID")
			if !exists {
				userID = "anonymous" // Set to anonymous for unauthenticated requests
			}

			// Extract entity type and ID from request parameters
			entityType := c.Param("entityType")
			if entityType == "" {
				// Try to determine entity type from path
				path := c.Request.URL.Path
				parts := strings.Split(path, "/")
				if len(parts) > 2 {
					entityType = parts[2] // Assuming path format like /api/transactions
				}
			}

			entityID := c.Param("id")
			if entityID == "" {
				// For list operations or create operations, entityID might not be in the path
				entityID = "multiple"
			}

			// Create audit log entry
			auditLog := models.AuditLog{
				ID:         uuid.New().String(),
				UserID:     userID.(string),
				Action:     action,
				EntityType: entityType,
				EntityID:   entityID,
				Detail:     c.Request.URL.Path,
				IP:         c.ClientIP(),
				Timestamp:  time.Now(),
			}

			// Log audit entry
			if err := utils.DB.Create(&auditLog).Error; err != nil {
				utils.Logger.Errorf("Failed to create audit log: %v", err)
			}
		}
	}
}
