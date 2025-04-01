package mocks

import (
	"context"
	"errors"
	"sync"

	"github.com/AGENT3128/shortener-url/internal/app/config"
)

// TestConfig provides a common configuration for tests
var TestConfig = &config.Config{
	ServerAddress:  "localhost:8080",
	BaseURLAddress: "http://localhost:8080",
	ReleaseMode:    "debug",
	LogLevel:       "info",
}

// ErrURLExists is the error returned when trying to add a URL that already exists
type urlExistsError struct {
	msg string
}

func (e *urlExistsError) Error() string {
	return e.msg
}

// Is implements errors.Is interface for compatibility with storage.ErrURLExists
func (e *urlExistsError) Is(target error) bool {
	// Check if target error message is "url already exists"
	return target.Error() == "url already exists" || errors.Is(target, &urlExistsError{})
}

// ErrURLExists is the singleton instance
var ErrURLExists = &urlExistsError{msg: "url already exists"}

// MockMemoryRepository is a mock implementation of URLRepository for testing
type MockMemoryRepository struct {
	mu   sync.RWMutex
	urls map[string]string
}

// NewMockMemoryRepository creates a new instance of MockMemoryRepository
func NewMockMemoryRepository() *MockMemoryRepository {
	return &MockMemoryRepository{
		urls: make(map[string]string),
	}
}

// Add adds a new URL to the repository
func (m *MockMemoryRepository) Add(ctx context.Context, shortID, originalURL string) (string, error) {
	if _, ok := m.GetByOriginalURL(ctx, originalURL); ok {
		return shortID, ErrURLExists
	}
	m.mu.Lock()
	defer m.mu.Unlock()

	m.urls[shortID] = originalURL
	return shortID, nil
}

// GetByShortID retrieves the original URL by short ID
func (m *MockMemoryRepository) GetByShortID(ctx context.Context, shortID string) (string, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	originalURL, ok := m.urls[shortID]
	return originalURL, ok
}

// GetByOriginalURL retrieves the short ID by original URL
func (m *MockMemoryRepository) GetByOriginalURL(ctx context.Context, originalURL string) (string, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	for shortID, url := range m.urls {
		if url == originalURL {
			return shortID, true
		}
	}
	return "", false
}
