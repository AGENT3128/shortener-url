package models

import (
	"time"
)

// URL represents a URL entity in the storage
type URL struct {
	ShortID     string    `json:"short_id"`
	OriginalURL string    `json:"original_url"`
	CreatedAt   time.Time `json:"created_at"`
	UserID      string    `json:"user_id"`
	DeletedFlag bool      `json:"is_deleted"`
}

// URLExistsError represents an error when a URL already exists
type URLExistsError struct {
	Message string
}

func (e *URLExistsError) Error() string {
	return e.Message
}

// Is implements the errors.Is interface for compatibility with other errors
func (e *URLExistsError) Is(target error) bool {
	// Check by error message for backward compatibility
	if targetErr, ok := target.(*URLExistsError); ok {
		return e.Message == targetErr.Message
	}
	// For compatibility with simple errors, also check the message
	return target.Error() == e.Message
}

// ErrURLExists is a singleton error for URL already exists
var ErrURLExists = &URLExistsError{Message: "url already exists"}
