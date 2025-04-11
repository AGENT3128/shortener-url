package mocks

import (
	"context"
	"sync"

	"github.com/AGENT3128/shortener-url/internal/app/config"
	"github.com/AGENT3128/shortener-url/internal/app/models"
)

// TestConfig provides a common configuration for tests
var TestConfig = &config.Config{
	ServerAddress:  "localhost:8080",
	BaseURLAddress: "http://localhost:8080",
	ReleaseMode:    "debug",
	LogLevel:       "info",
}

// ErrURLExists is re-exported from models
var ErrURLExists = models.ErrURLExists

// MockMemoryRepository is a mock implementation of URLRepository for testing
type MockMemoryRepository struct {
	mu   sync.RWMutex
	urls map[string]models.URL
}

// NewMockMemoryRepository creates a new instance of MockMemoryRepository
func NewMockMemoryRepository() *MockMemoryRepository {
	return &MockMemoryRepository{
		urls: make(map[string]models.URL),
	}
}

// Add adds a new URL to the repository
func (m *MockMemoryRepository) Add(ctx context.Context, userID, shortID, originalURL string) (string, error) {
	if _, ok := m.GetByOriginalURL(ctx, originalURL); ok {
		return shortID, ErrURLExists
	}
	m.mu.Lock()
	defer m.mu.Unlock()

	m.urls[shortID] = models.URL{
		ShortID:     shortID,
		OriginalURL: originalURL,
		UserID:      userID,
		DeletedFlag: false,
	}
	return shortID, nil
}

// GetByShortID retrieves the original URL by short ID
func (m *MockMemoryRepository) GetByShortID(ctx context.Context, shortID string) (string, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	url, ok := m.urls[shortID]
	return url.OriginalURL, ok
}

// GetByOriginalURL retrieves the short ID by original URL
func (m *MockMemoryRepository) GetByOriginalURL(ctx context.Context, originalURL string) (string, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	for shortID, url := range m.urls {
		if url.OriginalURL == originalURL {
			return shortID, true
		}
	}
	return "", false
}

func (m *MockMemoryRepository) AddBatch(ctx context.Context, userID string, urls []models.URL) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	for _, url := range urls {
		m.urls[url.ShortID] = models.URL{
			ShortID:     url.ShortID,
			OriginalURL: url.OriginalURL,
			UserID:      userID,
			DeletedFlag: false,
		}
	}

	return nil
}

func (m *MockMemoryRepository) GetUserURLs(ctx context.Context, userID string) ([]models.URL, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	urls := make([]models.URL, 0)
	for _, url := range m.urls {
		if url.UserID == userID {
			urls = append(urls, url)
		}
	}
	return urls, nil
}

// IsURLDeleted checks if a URL is marked as deleted
func (m *MockMemoryRepository) IsURLDeleted(ctx context.Context, shortID string) (bool, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	url, exists := m.urls[shortID]
	if !exists {
		return false, nil
	}

	return url.DeletedFlag, nil
}

// MarkDeletedBatch marks URLs as deleted in batch
func (m *MockMemoryRepository) MarkDeletedBatch(ctx context.Context, userID string, shortIDs []string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	for _, shortID := range shortIDs {
		url, exists := m.urls[shortID]
		if exists && url.UserID == userID {
			url.DeletedFlag = true
			m.urls[shortID] = url
		}
	}

	return nil
}
