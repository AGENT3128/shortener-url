package storage

import (
	"context"
	"sync"

	"github.com/AGENT3128/shortener-url/internal/app/models"
	"go.uber.org/zap"
)

type MemStorage struct {
	mu     sync.RWMutex
	urls   map[string]string
	logger *zap.Logger
}

func NewMemStorage(logger *zap.Logger) *MemStorage {
	logger = logger.With(zap.String("storage", "memory"))
	return &MemStorage{
		urls:   make(map[string]string),
		logger: logger,
	}
}

func (m *MemStorage) Add(ctx context.Context, shortID, originalURL string) (string, error) {
	const method = "Add"
	// before adding, check if the URL already exists (check by original URL)
	if _, ok := m.GetByOriginalURL(ctx, originalURL); ok {
		m.logger.Info(method, zap.String("originalURL", originalURL), zap.String("shortID", shortID), zap.Bool("exists", ok))
		return shortID, models.ErrURLExists
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	m.urls[shortID] = originalURL
	m.logger.Info(method, zap.String("shortID", shortID), zap.String("originalURL", originalURL))
	return shortID, nil
}

func (m *MemStorage) GetByShortID(ctx context.Context, shortID string) (string, bool) {
	const method = "GetByShortID"
	m.mu.RLock()
	defer m.mu.RUnlock()

	originalURL, ok := m.urls[shortID]
	m.logger.Info(method, zap.String("shortID", shortID), zap.String("originalURL", originalURL), zap.Bool("ok", ok))
	return originalURL, ok
}

func (m *MemStorage) GetByOriginalURL(ctx context.Context, originalURL string) (string, bool) {
	const method = "GetByOriginalURL"
	m.mu.RLock()
	defer m.mu.RUnlock()

	for shortID, url := range m.urls {
		if url == originalURL {
			m.logger.Info(method, zap.String("shortID", shortID), zap.String("url", url))
			return shortID, true
		}
	}
	return "", false
}

func (m *MemStorage) AddBatch(ctx context.Context, urls []models.URL) error {
	const method = "AddBatch"
	m.mu.Lock()
	defer m.mu.Unlock()

	for _, url := range urls {
		m.logger.Info(method, zap.String("shortID", url.ShortID), zap.String("originalURL", url.OriginalURL))
		m.urls[url.ShortID] = url.OriginalURL
	}

	return nil
}

func (m *MemStorage) Ping(ctx context.Context) error {
	// not needed for memory storage
	return nil
}
