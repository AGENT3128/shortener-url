package storage

import "sync"

type MemStorage struct {
	mu   sync.RWMutex
	urls map[string]string
}

func NewMemStorage() *MemStorage {
	return &MemStorage{
		urls: make(map[string]string),
	}
}

func (m *MemStorage) Add(shortID, originalURL string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.urls[shortID] = originalURL
}

func (m *MemStorage) GetByShortID(shortID string) (string, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	originalURL, ok := m.urls[shortID]
	return originalURL, ok
}

func (m *MemStorage) GetByOriginalURL(originalURL string) (string, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	for shortID, url := range m.urls {
		if url == originalURL {
			return shortID, true
		}
	}
	return "", false
}
