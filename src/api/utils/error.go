package utils

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
	"github.com/gin-gonic/gin"
	"github.com/go-sql-driver/mysql"
	"gorm.io/gorm"
)

// APIError represents a standardized API error
type APIError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Details string `json:"details,omitempty"`
}

// Common error codes
const (
	ErrorCodeValidation     = "validation_error"
	ErrorCodeDatabase       = "database_error"
	ErrorCodeNotFound       = "not_found"
	ErrorCodeUnauthorized   = "unauthorized"
	ErrorCodeForbidden      = "forbidden"
	ErrorCodeBadRequest     = "bad_request"
	ErrorCodeInternalServer = "internal_server_error"
	ErrorCodeDuplicate      = "duplicate_entry"
)

// ErrorResponse sends a standardized error response
func ErrorResponse(c *gin.Context, status int, code, message string, details string) {
	c.JSON(status, gin.H{
		"error": APIError{
			Code:    code,
			Message: message,
			Details: details,
		},
	})
}

// HandleError is a helper function that handles common errors and sends appropriate responses
func HandleError(c *gin.Context, err error) {
	// Log the error
	Logger.Errorf("Error: %v", err)

	switch {
	case errors.Is(err, gorm.ErrRecordNotFound):
		// Handle not found errors
		ErrorResponse(c, http.StatusNotFound, ErrorCodeNotFound, "Resource not found", "")
	case errors.Is(err, ErrInvalidToken), errors.Is(err, ErrTokenExpired):
		// Handle authentication errors
		ErrorResponse(c, http.StatusUnauthorized, ErrorCodeUnauthorized, "Authentication required", "")
	case errors.Is(err, gorm.ErrInvalidTransaction), errors.Is(err, gorm.ErrNotImplemented):
		// Handle database operation errors
		ErrorResponse(c, http.StatusInternalServerError, ErrorCodeDatabase, "Database operation failed", "")
	default:
		// Check for MySQL specific errors
		var mysqlErr *mysql.MySQLError
		if errors.As(err, &mysqlErr) {
			switch mysqlErr.Number {
			case 1062: // Duplicate entry error
				field := extractFieldFromDuplicateError(mysqlErr.Message)
				ErrorResponse(c, http.StatusConflict, ErrorCodeDuplicate, fmt.Sprintf("%s already exists", field), "")
				return
			}
		}

		// Generic error handling as fallback
		ErrorResponse(c, http.StatusInternalServerError, ErrorCodeInternalServer, "An unexpected error occurred", "")
	}
}

// HandleValidationError handles validation errors and sends a standardized response
func HandleValidationError(c *gin.Context, fieldErrors map[string]string) {
	details := ""
	for field, msg := range fieldErrors {
		if details != "" {
			details += ", "
		}
		details += fmt.Sprintf("%s: %s", field, msg)
	}

	ErrorResponse(c, http.StatusBadRequest, ErrorCodeValidation, "Validation failed", details)
}

// extractFieldFromDuplicateError attempts to extract field name from MySQL duplicate entry error
func extractFieldFromDuplicateError(errMsg string) string {
	// Example error message: "Duplicate entry 'value' for key 'table.field'"
	if !strings.Contains(errMsg, "Duplicate entry") {
		return "Record"
	}

	parts := strings.Split(errMsg, "key ")
	if len(parts) < 2 {
		return "Record"
	}

	keyPart := parts[1]
	keyPart = strings.Trim(keyPart, "'")
	keyParts := strings.Split(keyPart, ".")
	if len(keyParts) < 2 {
		return keyPart
	}

	// Return the field name
	field := keyParts[1]
	// Convert snake_case to Title Case
	field = strings.ReplaceAll(field, "_", " ")
	words := strings.Fields(field)
	for i := range words {
		words[i] = strings.Title(words[i])
	}
	return strings.Join(words, " ")
}
