package storage

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestFileStorage(t *testing.T) {
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
	}{
		{
			name:        "add new original URL1",
			shortID:     "abc123",
			originalURL: "https://ya.ru",
		},
		{
			name:        "add new original URL2",
			shortID:     "abc456",
			originalURL: "https://yandex.ru",
		},
		{
			name:        "add duplicate original URL2",
			shortID:     "abc456",
			originalURL: "https://yandex.ru",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo.Add(tt.shortID, tt.originalURL)
		})
		t.Run("check urls by shortID:"+tt.shortID, func(t *testing.T) {
			originalURL, ok := repo.GetByShortID(tt.shortID)
			assert.Equal(t, tt.originalURL, originalURL)
			assert.True(t, ok)
		})
		t.Run("check urls by originalURL:"+tt.originalURL, func(t *testing.T) {
			shortID, ok := repo.GetByOriginalURL(tt.originalURL)
			assert.Equal(t, tt.shortID, shortID)
			assert.True(t, ok)
		})
	}
	assert.Equal(t, 2, len(repo.urls))
	assert.Empty(t, repo.urls["abc1234"])
}
