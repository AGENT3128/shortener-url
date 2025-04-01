package storage

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestFileStorage(t *testing.T) {
	ctx := context.Background()
	testLogger, err := zap.NewDevelopment()
	if err != nil {
		t.Fatalf("failed to create test logger: %v", err)
	}
	tempFile, err := os.CreateTemp(os.TempDir(), "test_storage.json")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	defer os.Remove(tempFile.Name())

	repo, err := NewFileStorage(tempFile.Name(), testLogger)
	if err != nil {
		t.Fatalf("failed to create file storage: %v", err)
	}
	defer repo.Close()

	tests := []struct {
		name        string
		shortID     string
		originalURL string
		wantError   bool
	}{
		{
			name:        "add new original URL1",
			shortID:     "abc123",
			originalURL: "https://ya.ru",
			wantError:   false,
		},
		{
			name:        "add new original URL2",
			shortID:     "abc456",
			originalURL: "https://yandex.ru",
			wantError:   false,
		},
		{
			name:        "add duplicate original URL2",
			shortID:     "abc456",
			originalURL: "https://yandex.ru",
			wantError:   true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(_ *testing.T) {
			// context for request
			requestCtx, requestCancel := context.WithTimeout(ctx, 2*time.Second)
			defer requestCancel()

			_, err := repo.Add(requestCtx, tt.shortID, tt.originalURL)
			if tt.wantError {
				assert.Error(t, err)
				assert.Equal(t, ErrURLExists, err)
			} else {
				assert.NoError(t, err)
			}
		})
		t.Run("check urls by shortID:"+tt.shortID, func(t *testing.T) {
			// context for request
			requestCtx, requestCancel := context.WithTimeout(ctx, 2*time.Second)
			defer requestCancel()

			originalURL, ok := repo.GetByShortID(requestCtx, tt.shortID)
			assert.Equal(t, tt.originalURL, originalURL)
			assert.True(t, ok)
		})
		t.Run("check urls by originalURL:"+tt.originalURL, func(t *testing.T) {
			// context for request
			requestCtx, requestCancel := context.WithTimeout(ctx, 2*time.Second)
			defer requestCancel()

			shortID, ok := repo.GetByOriginalURL(requestCtx, tt.originalURL)
			assert.Equal(t, tt.shortID, shortID)
			assert.True(t, ok)
		})
	}
	assert.Equal(t, 2, len(repo.urls))
	assert.Empty(t, repo.urls["abc1234"])
}
