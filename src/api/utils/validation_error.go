package utils

// ValidationError represents validation errors
type ValidationError struct {
	Message string
	Field   string
}

func (e *ValidationError) Error() string {
	if e.Field != "" {
		return e.Field + ": " + e.Message
	}
	return e.Message
}

// NewValidationError creates a new ValidationError
func NewValidationError(message string, field string) *ValidationError {
	return &ValidationError{
		Message: message,
		Field:   field,
	}
}

// ErrInvalidState represents an error when an operation is attempted on a resource in an invalid state
var ErrInvalidState = NewValidationError("Invalid state for this operation", "")
