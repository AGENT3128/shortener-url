package file_test

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"github.com/AGENT3128/shortener-url/internal/entity"
	"github.com/AGENT3128/shortener-url/internal/repository/file"
)

type testStorage struct {
	storage *file.Storage
	file    *os.File
}

func setupTestStorage(t *testing.T) *testStorage {
	t.Helper()
	testLogger, err := zap.NewDevelopment()
	require.NoError(t, err, "failed to create test logger")

	tempFile, err := os.CreateTemp(t.TempDir(), "test_storage.json")
	require.NoError(t, err, "failed to create temp file")

	// Test with custom save ticker for faster testing
	storage, err := file.NewFileStorage(
		tempFile.Name(),
		testLogger,
		file.WithSaveTicker(100*time.Millisecond),
	)
	require.NoError(t, err, "failed to create file storage")

	return &testStorage{
		storage: storage,
		file:    tempFile,
	}
}

func (ts *testStorage) cleanup(t *testing.T) {
	t.Helper()
	require.NoError(t, ts.storage.Close())
	require.NoError(t, os.Remove(ts.file.Name()))
}

func TestBasicOperations(t *testing.T) {
	ts := setupTestStorage(t)
	defer ts.cleanup(t)

	tests := []struct {
		name        string
		shortID     string
		originalURL string
		userID      string
		wantError   bool
	}{
		{
			name:        "add new original URL1",
			shortID:     "abc123",
			originalURL: "https://ya.ru",
			userID:      "user1",
			wantError:   false,
		},
		{
			name:        "add new original URL2",
			shortID:     "abc456",
			originalURL: "https://yandex.ru",
			userID:      "user1",
			wantError:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(t.Context(), 2*time.Second)
			defer cancel()

			shortURL, err := ts.storage.Add(ctx, tt.userID, tt.shortID, tt.originalURL)
			if tt.wantError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.shortID, shortURL)

				gotURL, errGet := ts.storage.GetByShortURL(ctx, tt.shortID)
				require.NoError(t, errGet)
				assert.Equal(t, tt.originalURL, gotURL)

				gotShortURL, errGet := ts.storage.GetByOriginalURL(ctx, tt.originalURL)
				require.NoError(t, errGet)
				assert.Equal(t, tt.shortID, gotShortURL)
			}
		})
	}
}

func TestBatchOperations(t *testing.T) {
	ts := setupTestStorage(t)
	defer ts.cleanup(t)

	ctx, cancel := context.WithTimeout(t.Context(), 2*time.Second)
	defer cancel()

	urls := []entity.URL{
		{ShortURL: "batch1", OriginalURL: "https://example1.com"},
		{ShortURL: "batch2", OriginalURL: "https://example2.com"},
	}

	err := ts.storage.AddBatch(ctx, "user2", urls)
	require.NoError(t, err)

	for _, url := range urls {
		t.Run("check batch URL "+url.ShortURL, func(t *testing.T) {
			originalURL, errGet := ts.storage.GetByShortURL(ctx, url.ShortURL)
			require.NoError(t, errGet)
			assert.Equal(t, url.OriginalURL, originalURL)
		})
	}
}

func TestGetUserURLs(t *testing.T) {
	ts := setupTestStorage(t)
	defer ts.cleanup(t)

	ctx, cancel := context.WithTimeout(t.Context(), 2*time.Second)
	defer cancel()

	_, err := ts.storage.Add(ctx, "user3", "test1", "https://test1.com")
	require.NoError(t, err)
	_, err = ts.storage.Add(ctx, "user3", "test2", "https://test2.com")
	require.NoError(t, err)

	userURLs, err := ts.storage.GetUserURLs(ctx, "user3")
	require.NoError(t, err)
	assert.Len(t, userURLs, 2)
}

func TestURLDeletion(t *testing.T) {
	ts := setupTestStorage(t)
	defer ts.cleanup(t)

	ctx, cancel := context.WithTimeout(t.Context(), 2*time.Second)
	defer cancel()

	shortURL := "delete-test"
	_, err := ts.storage.Add(ctx, "user4", shortURL, "https://delete-test.com")
	require.NoError(t, err)

	err = ts.storage.MarkDeletedBatch(ctx, "user4", []string{shortURL})
	require.NoError(t, err)

	_, err = ts.storage.GetByShortURL(ctx, shortURL)
	assert.ErrorIs(t, err, entity.ErrURLDeleted)
}

func TestNonExistentURLs(t *testing.T) {
	ts := setupTestStorage(t)
	defer ts.cleanup(t)

	ctx, cancel := context.WithTimeout(t.Context(), 2*time.Second)
	defer cancel()

	t.Run("get non-existent short URL", func(t *testing.T) {
		_, err := ts.storage.GetByShortURL(ctx, "non-existent")
		assert.ErrorIs(t, err, entity.ErrURLNotFound)
	})

	t.Run("get non-existent original URL", func(t *testing.T) {
		_, err := ts.storage.GetByOriginalURL(ctx, "https://non-existent.com")
		assert.ErrorIs(t, err, entity.ErrURLNotFound)
	})
}

func TestPing(t *testing.T) {
	ts := setupTestStorage(t)
	defer ts.cleanup(t)

	ctx, cancel := context.WithTimeout(t.Context(), 2*time.Second)
	defer cancel()

	err := ts.storage.Ping(ctx)
	assert.NoError(t, err)
}

func TestStatePreservation(t *testing.T) {
	tempDir := t.TempDir()
	filePath := tempDir + "/storage.json"
	logger, err := zap.NewDevelopment()
	require.NoError(t, err)

	// Create initial storage and add data
	storage1, err := file.NewFileStorage(filePath, logger)
	require.NoError(t, err)

	ctx := t.Context()
	_, err = storage1.Add(ctx, "user1", "test1", "https://example1.com")
	require.NoError(t, err)
	_, err = storage1.Add(ctx, "user1", "test2", "https://example2.com")
	require.NoError(t, err)

	// Close storage to ensure state is saved
	err = storage1.Close()
	require.NoError(t, err)

	// Create new storage instance and verify state is restored
	storage2, err := file.NewFileStorage(filePath, logger)
	require.NoError(t, err)
	defer storage2.Close()

	// Verify URLs were restored
	url1, err := storage2.GetByShortURL(ctx, "test1")
	require.NoError(t, err)
	assert.Equal(t, "https://example1.com", url1)

	url2, err := storage2.GetByShortURL(ctx, "test2")
	require.NoError(t, err)
	assert.Equal(t, "https://example2.com", url2)
}

func TestPeriodicSaving(t *testing.T) {
	tempDir := t.TempDir()
	filePath := tempDir + "/periodic_storage.json"
	logger, err := zap.NewDevelopment()
	require.NoError(t, err)

	// Create storage with short save interval
	storage, err := file.NewFileStorage(
		filePath,
		logger,
		file.WithSaveTicker(100*time.Millisecond),
	)
	require.NoError(t, err)

	ctx := t.Context()
	_, err = storage.Add(ctx, "user1", "periodic1", "https://periodic1.com")
	require.NoError(t, err)

	// Wait for periodic save
	time.Sleep(200 * time.Millisecond)

	// Close storage
	err = storage.Close()
	require.NoError(t, err)

	// Create new storage and verify state
	newStorage, err := file.NewFileStorage(filePath, logger)
	require.NoError(t, err)
	defer newStorage.Close()

	url, err := newStorage.GetByShortURL(ctx, "periodic1")
	require.NoError(t, err)
	assert.Equal(t, "https://periodic1.com", url)
}
