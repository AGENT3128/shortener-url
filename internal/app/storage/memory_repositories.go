package storage

import (
	"sync"

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

func (m *MemStorage) Add(shortID, originalURL string) {
	const method = "Add"
	m.mu.Lock()
	defer m.mu.Unlock()

	m.urls[shortID] = originalURL
	m.logger.Info(method, zap.String("shortID", shortID), zap.String("originalURL", originalURL))
}

func (m *MemStorage) GetByShortID(shortID string) (string, bool) {
	const method = "GetByShortID"
	m.mu.RLock()
	defer m.mu.RUnlock()

	originalURL, ok := m.urls[shortID]
	m.logger.Info(method, zap.String("shortID", shortID), zap.String("originalURL", originalURL), zap.Bool("ok", ok))
	return originalURL, ok
}

func (m *MemStorage) GetByOriginalURL(originalURL string) (string, bool) {
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
