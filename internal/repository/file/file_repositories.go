package file

import (
	"bufio"
	"context"
	"encoding/json"
	"os"
	"strconv"
	"sync"

	"go.uber.org/zap"

	"github.com/AGENT3128/shortener-url/internal/entity"
)

// Storage is the file storage for the URL.
type Storage struct {
	file     *os.File
	mu       sync.RWMutex
	urls     map[string]URLData
	lastUUID int
	logger   *zap.Logger
}

// URLData is the data for the URL.
type URLData struct {
	OriginalURL string
	UUID        string
	UserID      string
	IsDeleted   bool
}

// URLRecord is the record for the URL.
type URLRecord struct {
	UUID        string `json:"uuid"`
	ShortURL    string `json:"short_url"`
	OriginalURL string `json:"original_url"`
}

// NewFileStorage creates a new FileStorage.
func NewFileStorage(path string, logger *zap.Logger) (*Storage, error) {
	logger = logger.With(zap.String("storage", "file"))
	file, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE, 0600)
	if err != nil {
		logger.Error("failed to open file storage", zap.Error(err))
		return nil, err
	}

	fs := &Storage{
		file:     file,
		urls:     make(map[string]URLData),
		lastUUID: 0,
		logger:   logger,
	}
	if errLoad := fs.loadFromFile(); errLoad != nil {
		fs.logger.Error("failed to load file storage", zap.Error(errLoad))
		return nil, errLoad
	}
	return fs, nil
}

// Add adds a URL.
func (f *Storage) Add(_ context.Context, userID, shortID, originalURL string) (string, error) {
	const method = "Add"
	f.mu.Lock()
	defer f.mu.Unlock()

	f.lastUUID++
	uuid := strconv.Itoa(f.lastUUID)

	f.urls[shortID] = URLData{
		OriginalURL: originalURL,
		UUID:        uuid,
		UserID:      userID,
	}

	if err := f.saveToFile(); err != nil {
		f.logger.Error(method, zap.Error(err))
	}
	return shortID, nil
}

// GetByShortURL gets the original URL by the short URL.
func (f *Storage) GetByShortURL(_ context.Context, shortURL string) (string, error) {
	const method = "GetByShortURL"
	f.mu.RLock()
	defer f.mu.RUnlock()

	url, ok := f.urls[shortURL]
	f.logger.Info(method, zap.String("shortURL", shortURL), zap.Any("urlData", url))
	if !ok {
		return "", entity.ErrURLNotFound
	}
	if url.IsDeleted {
		return "", entity.ErrURLDeleted
	}
	return url.OriginalURL, nil
}

// GetByOriginalURL gets the short URL by the original URL.
func (f *Storage) GetByOriginalURL(_ context.Context, originalURL string) (string, error) {
	const method = "GetByOriginalURL"
	f.mu.RLock()
	defer f.mu.RUnlock()

	for shortID, urlData := range f.urls {
		if urlData.OriginalURL == originalURL {
			f.logger.Info(method, zap.String("shortID", shortID), zap.String("originalURL", originalURL))
			return shortID, nil
		}
	}
	return "", entity.ErrURLNotFound
}

// AddBatch adds a batch of URLs.
func (f *Storage) AddBatch(_ context.Context, userID string, urls []entity.URL) error {
	const method = "AddBatch"
	f.mu.Lock()
	defer f.mu.Unlock()

	for _, url := range urls {
		f.lastUUID++
		uuid := strconv.Itoa(f.lastUUID)

		f.urls[url.ShortURL] = URLData{
			OriginalURL: url.OriginalURL,
			UUID:        uuid,
			UserID:      userID,
		}
	}

	if err := f.saveToFile(); err != nil {
		f.logger.Error(method, zap.Error(err))
		return err
	}

	return nil
}

// loadFromFile loads the URLs from the file.
func (f *Storage) loadFromFile() error {
	const method = "loadFromFile"
	if _, err := f.file.Seek(0, 0); err != nil {
		f.logger.Error(method, zap.Error(err))
		return err
	}

	scanner := bufio.NewScanner(f.file)
	for scanner.Scan() {
		var record URLRecord
		if err := json.Unmarshal(scanner.Bytes(), &record); err != nil {
			f.logger.Error(method, zap.Error(err))
			continue
		}

		f.urls[record.ShortURL] = URLData{
			OriginalURL: record.OriginalURL,
			UUID:        record.UUID,
		}

		if uuid, err := strconv.Atoi(record.UUID); err == nil && uuid > f.lastUUID {
			f.lastUUID = uuid
		}
	}
	return scanner.Err()
}

// saveToFile saves the URLs to the file.
func (f *Storage) saveToFile() error {
	const method = "saveToFile"
	if err := f.file.Truncate(0); err != nil {
		f.logger.Error(method, zap.Error(err))
		return err
	}
	if _, err := f.file.Seek(0, 0); err != nil {
		f.logger.Error(method, zap.Error(err))
		return err
	}

	writer := bufio.NewWriter(f.file)
	for shortURL, urlData := range f.urls {
		record := URLRecord{
			UUID:        urlData.UUID,
			ShortURL:    shortURL,
			OriginalURL: urlData.OriginalURL,
		}
		data, err := json.Marshal(record)
		if err != nil {
			continue
		}
		_, err = writer.Write(data)
		if err != nil {
			f.logger.Error(method, zap.Error(err))
			continue
		}
		err = writer.WriteByte('\n')
		if err != nil {
			f.logger.Error(method, zap.Error(err))
			continue
		}
		f.logger.Info(method, zap.Any("record", record))
	}
	return writer.Flush()
}

// Close closes the file storage.
func (f *Storage) Close() error {
	return f.file.Close()
}

// Ping pings the file storage.
func (f *Storage) Ping(_ context.Context) error {
	// not needed for file storage
	return nil
}

// GetUserURLs gets user URLs.
func (f *Storage) GetUserURLs(_ context.Context, userID string) ([]entity.URL, error) {
	const method = "GetUserURLs"
	f.mu.RLock()
	defer f.mu.RUnlock()

	urls := make([]entity.URL, 0)
	for shortURL, urlData := range f.urls {
		if urlData.UserID == userID {
			urls = append(urls, entity.URL{ShortURL: shortURL, OriginalURL: urlData.OriginalURL})
		}
	}
	f.logger.Info(method, zap.String("userID", userID), zap.Int("count", len(urls)))
	return urls, nil
}

// MarkDeletedBatch marks URLs as deleted in batch.
func (f *Storage) MarkDeletedBatch(_ context.Context, userID string, shortURLs []string) error {
	const method = "MarkDeletedBatch"
	f.mu.Lock()
	defer f.mu.Unlock()

	modified := false
	for _, shortURL := range shortURLs {
		urlData, exists := f.urls[shortURL]
		if exists && urlData.UserID == userID {
			urlData.IsDeleted = true
			f.urls[shortURL] = urlData
			modified = true
			f.logger.Info(method, zap.String("shortURL", shortURL), zap.String("userID", userID))
		}
	}

	if modified {
		if err := f.saveToFile(); err != nil {
			f.logger.Error(method, zap.Error(err))
			return err
		}
	}

	return nil
}
