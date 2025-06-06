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

type FileStorage struct {
	file     *os.File
	mu       sync.RWMutex
	urls     map[string]URLData
	lastUUID int
	logger   *zap.Logger
}

type URLData struct {
	OriginalURL string
	UUID        string
	UserID      string
	IsDeleted   bool
}

type URLRecord struct {
	UUID        string `json:"uuid"`
	ShortURL    string `json:"short_url"`
	OriginalURL string `json:"original_url"`
}

func NewFileStorage(path string, logger *zap.Logger) (*FileStorage, error) {
	logger = logger.With(zap.String("storage", "file"))
	file, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE, 0666)
	if err != nil {
		logger.Error("failed to open file storage", zap.Error(err))
		return nil, err
	}

	fs := &FileStorage{
		file:     file,
		urls:     make(map[string]URLData),
		lastUUID: 0,
		logger:   logger,
	}
	if err := fs.loadFromFile(); err != nil {
		fs.logger.Error("failed to load file storage", zap.Error(err))
		return nil, err
	}
	return fs, nil
}

func (f *FileStorage) Add(ctx context.Context, userID, shortID, originalURL string) (string, error) {
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

func (f *FileStorage) GetByShortURL(ctx context.Context, shortURL string) (string, error) {
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

func (f *FileStorage) GetByOriginalURL(ctx context.Context, originalURL string) (string, error) {
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

func (f *FileStorage) AddBatch(ctx context.Context, userID string, urls []entity.URL) error {
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

func (f *FileStorage) loadFromFile() error {
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

func (f *FileStorage) saveToFile() error {
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

func (f *FileStorage) Close() error {
	return f.file.Close()
}

func (f *FileStorage) Ping(ctx context.Context) error {
	// not needed for file storage
	return nil
}

func (f *FileStorage) GetUserURLs(ctx context.Context, userID string) ([]entity.URL, error) {
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
func (f *FileStorage) MarkDeletedBatch(ctx context.Context, userID string, shortURLs []string) error {
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
