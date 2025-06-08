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
	ctx     context.Context
}

func setupTestStorage(t *testing.T) *testStorage {
	t.Helper()
	ctx := t.Context()
	testLogger, err := zap.NewDevelopment()
	require.NoError(t, err, "failed to create test logger")

	tempFile, err := os.CreateTemp(t.TempDir(), "test_storage.json")
	require.NoError(t, err, "failed to create temp file")

	storage, err := file.NewFileStorage(tempFile.Name(), testLogger)
	require.NoError(t, err, "failed to create file storage")

	return &testStorage{
		storage: storage,
		file:    tempFile,
		ctx:     ctx,
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
			ctx, cancel := context.WithTimeout(ts.ctx, 2*time.Second)
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

	ctx, cancel := context.WithTimeout(ts.ctx, 2*time.Second)
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

	ctx, cancel := context.WithTimeout(ts.ctx, 2*time.Second)
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

	ctx, cancel := context.WithTimeout(ts.ctx, 2*time.Second)
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

	ctx, cancel := context.WithTimeout(ts.ctx, 2*time.Second)
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

	ctx, cancel := context.WithTimeout(ts.ctx, 2*time.Second)
	defer cancel()

	err := ts.storage.Ping(ctx)
	assert.NoError(t, err)
}
