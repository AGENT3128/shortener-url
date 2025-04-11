package storage

import (
	"context"
	"sync"

	"github.com/AGENT3128/shortener-url/internal/app/models"
	"go.uber.org/zap"
)

type MemStorage struct {
	mu     sync.RWMutex
	urls   map[string]models.URL
	logger *zap.Logger
}

func NewMemStorage(logger *zap.Logger) *MemStorage {
	logger = logger.With(zap.String("storage", "memory"))
	return &MemStorage{
		urls:   make(map[string]models.URL),
		logger: logger,
	}
}

func (m *MemStorage) Add(ctx context.Context, userID, shortID, originalURL string) (string, error) {
	const method = "Add"
	// before adding, check if the URL already exists (check by original URL)
	if _, ok := m.GetByOriginalURL(ctx, originalURL); ok {
		m.logger.Info(method, zap.String("originalURL", originalURL), zap.String("shortID", shortID), zap.Bool("exists", ok))
		return shortID, models.ErrURLExists
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	m.urls[shortID] = models.URL{
		ShortID:     shortID,
		OriginalURL: originalURL,
		UserID:      userID,
	}
	m.logger.Info(method, zap.String("shortID", shortID), zap.String("originalURL", originalURL))
	return shortID, nil
}

func (m *MemStorage) GetByShortID(ctx context.Context, shortID string) (string, bool) {
	const method = "GetByShortID"
	m.mu.RLock()
	defer m.mu.RUnlock()

	url, ok := m.urls[shortID]
	m.logger.Info(method, zap.String("shortID", shortID), zap.String("originalURL", url.OriginalURL), zap.Bool("ok", ok))
	if url.DeletedFlag {
		return url.OriginalURL, false
	}
	return url.OriginalURL, ok
}

func (m *MemStorage) GetByOriginalURL(ctx context.Context, originalURL string) (string, bool) {
	const method = "GetByOriginalURL"
	m.mu.RLock()
	defer m.mu.RUnlock()

	for shortID, url := range m.urls {
		if url.OriginalURL == originalURL {
			m.logger.Info(method, zap.String("shortID", shortID), zap.String("url", url.OriginalURL))
			return shortID, true
		}
	}
	return "", false
}

func (m *MemStorage) AddBatch(ctx context.Context, userID string, urls []models.URL) error {
	const method = "AddBatch"
	m.mu.Lock()
	defer m.mu.Unlock()

	for _, url := range urls {
		m.logger.Info(method, zap.String("shortID", url.ShortID), zap.String("originalURL", url.OriginalURL), zap.String("userID", userID))
		m.urls[url.ShortID] = models.URL{
			ShortID:     url.ShortID,
			OriginalURL: url.OriginalURL,
			UserID:      userID,
		}
	}

	return nil
}

func (m *MemStorage) Ping(ctx context.Context) error {
	// not needed for memory storage
	return nil
}

func (m *MemStorage) GetUserURLs(ctx context.Context, userID string) ([]models.URL, error) {
	const method = "GetUserURLs"
	m.mu.RLock()
	defer m.mu.RUnlock()

	urls := make([]models.URL, 0)
	for _, url := range m.urls {
		if url.UserID == userID {
			urls = append(urls, url)
		}
	}
	m.logger.Info(method, zap.String("userID", userID), zap.Int("count", len(urls)))
	return urls, nil
}

// MarkDeletedBatch marks URLs as deleted in batch
func (m *MemStorage) MarkDeletedBatch(ctx context.Context, userID string, shortIDs []string) error {
	const method = "MarkDeletedBatch"
	m.mu.Lock()
	defer m.mu.Unlock()

	for _, shortID := range shortIDs {
		url, exists := m.urls[shortID]
		if exists && url.UserID == userID {
			url.DeletedFlag = true
			m.urls[shortID] = url
			m.logger.Info(method, zap.String("shortID", shortID), zap.String("userID", userID))
		}
	}

	return nil
}

// IsURLDeleted checks if URL is marked as deleted
func (m *MemStorage) IsURLDeleted(ctx context.Context, shortID string) (bool, error) {
	const method = "IsURLDeleted"
	m.mu.RLock()
	defer m.mu.RUnlock()

	url, exists := m.urls[shortID]
	m.logger.Info(method, zap.String("shortID", shortID), zap.Bool("exists", exists))
	if !exists {
		return false, nil
	}

	return url.DeletedFlag, nil
}
