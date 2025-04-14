package utils

import "errors"

// Common errors
var (
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrUserNotFound       = errors.New("user not found")
	ErrUnauthorized       = errors.New("unauthorized")
	ErrForbidden          = errors.New("forbidden")
	ErrNotFound           = errors.New("resource not found")
	ErrInvalidInput       = errors.New("invalid input")
	ErrInternalServer     = errors.New("internal server error")
)

// ConflictError represents a conflict in database operations
type ConflictError struct {
	Message string
	Field   string
}

func (e *ConflictError) Error() string {
	if e.Field != "" {
		return e.Field + ": " + e.Message
	}
	return e.Message
}

// NewConflictError creates a new ConflictError
func NewConflictError(message string, field string) *ConflictError {
	return &ConflictError{
		Message: message,
		Field:   field,
	}
}

// DatabaseError represents database operation errors
type DatabaseError struct {
	Message string
	Field   string
}

func (e *DatabaseError) Error() string {
	if e.Field != "" {
		return e.Field + ": " + e.Message
	}
	return e.Message
}

// NewDatabaseError creates a new DatabaseError
func NewDatabaseError(message string, field string) *DatabaseError {
	return &DatabaseError{
		Message: message,
		Field:   field,
	}
}

// NotFoundError represents resource not found errors
type NotFoundError struct {
	Message string
	Field   string
}

func (e *NotFoundError) Error() string {
	if e.Field != "" {
		return e.Field + ": " + e.Message
	}
	return e.Message
}

// NewNotFoundError creates a new NotFoundError
func NewNotFoundError(message string, field string) *NotFoundError {
	return &NotFoundError{
		Message: message,
		Field:   field,
	}
}
