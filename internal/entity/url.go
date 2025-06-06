package entity

import (
	"errors"
	"time"
)

// URL represents a URL entity in the storage.
type URL struct {
	ShortURL    string    `json:"short_url"`
	OriginalURL string    `json:"original_url"`
	UserID      string    `json:"user_id"`
	CreatedAt   time.Time `json:"created_at"`
	DeletedFlag bool      `json:"is_deleted"`
}

// Errors for the URL.
var (
	ErrURLExists   = errors.New("url already exists") // error when url already exists
	ErrURLDeleted  = errors.New("url deleted")        // error when url is deleted
	ErrURLNotFound = errors.New("url not found")      // error when url is not found
)
