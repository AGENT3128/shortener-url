package storage

import (
	"bufio"
	"context"
	"encoding/json"
	"os"
	"strconv"
	"sync"

	"go.uber.org/zap"
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

func (f *FileStorage) Add(ctx context.Context, shortID, originalURL string) (string, error) {
	const method = "Add"
	// before adding, check if the URL already exists (check by original URL)
	if _, ok := f.GetByOriginalURL(ctx, originalURL); ok {
		f.logger.Info(method, zap.String("originalURL", originalURL), zap.String("shortID", shortID), zap.Bool("exists", ok))
		return shortID, ErrURLExists
	}

	f.mu.Lock()
	defer f.mu.Unlock()

	f.lastUUID++
	uuid := strconv.Itoa(f.lastUUID)

	f.urls[shortID] = URLData{
		OriginalURL: originalURL,
		UUID:        uuid,
	}

	if err := f.saveToFile(); err != nil {
		f.logger.Error(method, zap.Error(err))
	}
	return shortID, nil
}

func (f *FileStorage) GetByShortID(ctx context.Context, shortID string) (string, bool) {
	const method = "GetByShortID"
	f.mu.RLock()
	defer f.mu.RUnlock()

	urlData, ok := f.urls[shortID]
	f.logger.Info(method, zap.String("shortID", shortID), zap.Any("urlData", urlData), zap.Bool("ok", ok))
	if !ok {
		return "", false
	}
	return urlData.OriginalURL, true
}

func (f *FileStorage) GetByOriginalURL(ctx context.Context, originalURL string) (string, bool) {
	const method = "GetByOriginalURL"
	f.mu.RLock()
	defer f.mu.RUnlock()

	for shortID, urlData := range f.urls {
		if urlData.OriginalURL == originalURL {
			f.logger.Info(method, zap.String("shortID", shortID), zap.String("originalURL", originalURL))
			return shortID, true
		}
	}
	return "", false
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
