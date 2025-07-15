package file

import (
	"bufio"
	"context"
	"encoding/json"
	"os"
	"strconv"
	"sync"
	"time"

	"go.uber.org/zap"

	"maps"

	"github.com/AGENT3128/shortener-url/internal/entity"
)

const (
	defaultSaveTicker = 10 * time.Second
)

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

// Memento represents a snapshot of the storage state.
type Memento struct {
	URLs     map[string]URLData
	LastUUID int
}

// Caretaker handles saving and restoring storage state.
type Caretaker struct {
	logger      *zap.Logger
	filePath    string
	saveTimeout time.Duration
}

// Storage is the file storage for the URL.
type Storage struct {
	urls       map[string]URLData
	logger     *zap.Logger
	caretaker  *Caretaker
	stopSaving chan struct{}
	lastUUID   int
	mu         sync.RWMutex
	isDirty    bool
}

// Option is the option for the FileStorage.
type Option func(*Caretaker)

// WithSaveTicker is the option for the FileStorage.
func WithSaveTicker(ticker time.Duration) Option {
	return func(c *Caretaker) {
		c.saveTimeout = ticker
	}
}

// NewFileStorage creates a new FileStorage.
func NewFileStorage(path string, logger *zap.Logger, opts ...Option) (*Storage, error) {
	logger = logger.With(zap.String("storage", "file"))

	caretaker := &Caretaker{
		filePath:    path,
		logger:      logger,
		saveTimeout: defaultSaveTicker,
	}

	for _, opt := range opts {
		opt(caretaker)
	}

	storage := &Storage{
		urls:       make(map[string]URLData),
		lastUUID:   0,
		logger:     logger,
		caretaker:  caretaker,
		isDirty:    false,
		stopSaving: make(chan struct{}),
	}

	if err := storage.restore(); err != nil {
		return nil, err
	}

	// Start periodic saving
	go storage.periodicSave()

	return storage, nil
}

// createMemento creates a snapshot of the current state.
func (f *Storage) createMemento() *Memento {
	f.mu.RLock()
	defer f.mu.RUnlock()

	urlsCopy := make(map[string]URLData)
	maps.Copy(urlsCopy, f.urls)

	return &Memento{
		URLs:     urlsCopy,
		LastUUID: f.lastUUID,
	}
}

// restoreFromMemento restores state from a memento.
func (f *Storage) restoreFromMemento(m *Memento) {
	f.mu.Lock()
	defer f.mu.Unlock()

	f.urls = m.URLs
	f.lastUUID = m.LastUUID
}

// periodicSave periodically saves state if it's dirty.
func (f *Storage) periodicSave() {
	ticker := time.NewTicker(f.caretaker.saveTimeout)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			if f.isDirty {
				if err := f.save(); err != nil {
					f.logger.Error("periodic save failed", zap.Error(err))
				}
			}
		case <-f.stopSaving:
			return
		}
	}
}

// save saves the current state to file.
func (f *Storage) save() error {
	memento := f.createMemento()

	file, err := os.OpenFile(f.caretaker.filePath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return err
	}
	defer file.Close()

	writer := bufio.NewWriter(file)

	for shortURL, urlData := range memento.URLs {
		record := URLRecord{
			UUID:        urlData.UUID,
			ShortURL:    shortURL,
			OriginalURL: urlData.OriginalURL,
		}

		data, errMarshal := json.Marshal(record)
		if errMarshal != nil {
			continue
		}

		if _, errWrite := writer.Write(data); errWrite != nil {
			return errWrite
		}
		if errWrite := writer.WriteByte('\n'); errWrite != nil {
			return errWrite
		}
	}

	if errFlush := writer.Flush(); errFlush != nil {
		return errFlush
	}

	f.mu.Lock()
	f.isDirty = false
	f.mu.Unlock()

	return nil
}

// restore loads the state from file.
func (f *Storage) restore() error {
	file, err := os.OpenFile(f.caretaker.filePath, os.O_RDONLY|os.O_CREATE, 0600)
	if err != nil {
		return err
	}
	defer file.Close()

	urls := make(map[string]URLData)
	lastUUID := 0

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		var record URLRecord
		if errUnmarshal := json.Unmarshal(scanner.Bytes(), &record); errUnmarshal != nil {
			continue
		}

		urls[record.ShortURL] = URLData{
			OriginalURL: record.OriginalURL,
			UUID:        record.UUID,
		}

		if uuid, errAtoi := strconv.Atoi(record.UUID); errAtoi == nil && uuid > lastUUID {
			lastUUID = uuid
		}
	}

	if errScan := scanner.Err(); errScan != nil {
		return errScan
	}

	memento := &Memento{
		URLs:     urls,
		LastUUID: lastUUID,
	}

	f.restoreFromMemento(memento)
	return nil
}

// Add adds a URL.
func (f *Storage) Add(_ context.Context, userID, shortID, originalURL string) (string, error) {
	f.mu.Lock()
	defer f.mu.Unlock()

	f.lastUUID++
	uuid := strconv.Itoa(f.lastUUID)

	f.urls[shortID] = URLData{
		OriginalURL: originalURL,
		UUID:        uuid,
		UserID:      userID,
	}

	f.isDirty = true
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
		f.logger.Info(
			method,
			zap.String("shortURL", url.ShortURL),
			zap.String("originalURL", url.OriginalURL),
			zap.String("userID", userID),
		)
	}

	f.isDirty = true
	return nil
}

// Close closes the file storage.
func (f *Storage) Close() error {
	close(f.stopSaving)
	if f.isDirty {
		return f.save()
	}
	return nil
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

	for _, shortURL := range shortURLs {
		urlData, exists := f.urls[shortURL]
		if exists && urlData.UserID == userID {
			urlData.IsDeleted = true
			f.urls[shortURL] = urlData
			f.isDirty = true
			f.logger.Info(method, zap.String("shortURL", shortURL), zap.String("userID", userID))
		}
	}

	return nil
}

// GetStats gets stats.
func (f *Storage) GetStats(_ context.Context) (int, int, error) {
	f.mu.RLock()
	defer f.mu.RUnlock()

	countUsers := make(map[string]int)
	countURLs := 0

	for _, url := range f.urls {
		if !url.IsDeleted {
			countURLs++
		}
		countUsers[url.UserID]++
	}

	usersCount := len(countUsers)
	urlsCount := countURLs

	return urlsCount, usersCount, nil
}
