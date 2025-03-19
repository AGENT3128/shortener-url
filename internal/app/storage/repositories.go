package storage

import "sync"

type Repository interface {
	Add(shortID, originalURL string)
	GetByShortID(shortID string) (string, bool)
	GetByOriginalURL(originalURL string) (string, bool)
}

type MemStotage struct {
	mu   sync.RWMutex
	urls map[string]string
}

func NewMemStorage() *MemStotage {
	return &MemStotage{
		urls: make(map[string]string),
	}
}

func (m *MemStotage) Add(shortID, originalURL string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.urls[shortID] = originalURL
}

func (m *MemStotage) GetByShortID(shortID string) (string, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	originalURL, ok := m.urls[shortID]
	return originalURL, ok
}

func (m *MemStotage) GetByOriginalURL(originalURL string) (string, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	for shortID, url := range m.urls {
		if url == originalURL {
			return shortID, true
		}
	}
	return "", false
}
