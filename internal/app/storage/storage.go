package storage

var StorageURLs *Storage

type Storage struct {
	URLs map[string]string
}

func InitStorage() {
	StorageURLs = &Storage{
		URLs: make(map[string]string),
	}
}

func (s *Storage) Add(shortID, originalURL string) {
	s.URLs[shortID] = originalURL
}

func (s *Storage) GetByShortID(shortID string) (string, bool) {
	originalURL, ok := s.URLs[shortID]
	return originalURL, ok
}

func (s *Storage) GetByOriginalURL(originalURL string) (string, bool) {
	for shortID, url := range s.URLs {
		if url == originalURL {
			return shortID, true
		}
	}
	return "", false
}
