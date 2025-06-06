package memory

import (
	"context"
	"sync"

	"go.uber.org/zap"

	"github.com/AGENT3128/shortener-url/internal/entity"
)

type MemStorage struct {
	mu     sync.RWMutex
	urls   map[string]entity.URL
	logger *zap.Logger
}

func NewMemStorage(logger *zap.Logger) *MemStorage {
	logger = logger.With(zap.String("storage", "memory"))
	return &MemStorage{
		urls:   make(map[string]entity.URL),
		logger: logger,
	}
}

func (m *MemStorage) Add(_ context.Context, userID, shortURL, originalURL string) (string, error) {
	const method = "Add"
	m.mu.Lock()
	defer m.mu.Unlock()

	m.urls[shortURL] = entity.URL{
		ShortURL:    shortURL,
		OriginalURL: originalURL,
		UserID:      userID,
	}
	m.logger.Info(method, zap.String("shortURL", shortURL), zap.String("originalURL", originalURL))
	return shortURL, nil
}

func (m *MemStorage) GetByShortURL(_ context.Context, shortURL string) (string, error) {
	const method = "GetByShortURL"
	m.mu.RLock()
	defer m.mu.RUnlock()

	url, ok := m.urls[shortURL]
	m.logger.Info(
		method,
		zap.String("shortURL", shortURL),
		zap.String("originalURL", url.OriginalURL),
		zap.Bool("ok", ok),
	)
	if !ok {
		return "", entity.ErrURLNotFound
	}
	if url.DeletedFlag {
		return "", entity.ErrURLDeleted
	}
	return url.OriginalURL, nil
}

func (m *MemStorage) GetByOriginalURL(_ context.Context, originalURL string) (string, error) {
	const method = "GetByOriginalURL"
	m.mu.RLock()
	defer m.mu.RUnlock()

	for shortURL, url := range m.urls {
		if url.OriginalURL == originalURL {
			m.logger.Info(method, zap.String("shortURL", shortURL), zap.String("url", url.OriginalURL))
			return shortURL, nil
		}
	}
	return "", entity.ErrURLNotFound
}

func (m *MemStorage) AddBatch(_ context.Context, userID string, urls []entity.URL) error {
	const method = "AddBatch"
	m.mu.Lock()
	defer m.mu.Unlock()

	for _, url := range urls {
		m.logger.Info(
			method,
			zap.String("shortURL", url.ShortURL),
			zap.String("originalURL", url.OriginalURL),
			zap.String("userID", userID),
		)
		m.urls[url.ShortURL] = entity.URL{
			ShortURL:    url.ShortURL,
			OriginalURL: url.OriginalURL,
			UserID:      userID,
		}
	}

	return nil
}

func (m *MemStorage) Ping(_ context.Context) error {
	// not needed for memory storage
	return nil
}

func (m *MemStorage) GetUserURLs(_ context.Context, userID string) ([]entity.URL, error) {
	const method = "GetUserURLs"
	m.mu.RLock()
	defer m.mu.RUnlock()

	urls := make([]entity.URL, 0)
	for _, url := range m.urls {
		if url.UserID == userID {
			urls = append(urls, url)
		}
	}
	m.logger.Info(method, zap.String("userID", userID), zap.Int("count", len(urls)))
	return urls, nil
}

// MarkDeletedBatch marks URLs as deleted in batch.
func (m *MemStorage) MarkDeletedBatch(_ context.Context, userID string, shortURLs []string) error {
	const method = "MarkDeletedBatch"
	m.mu.Lock()
	defer m.mu.Unlock()

	for _, shortURL := range shortURLs {
		url, exists := m.urls[shortURL]
		if exists && url.UserID == userID {
			url.DeletedFlag = true
			m.urls[shortURL] = url
			m.logger.Info(method, zap.String("shortURL", shortURL), zap.String("userID", userID))
		}
	}

	return nil
}
