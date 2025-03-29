package storage

import (
	"context"
	"testing"

	"github.com/AGENT3128/shortener-url/internal/app/db"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestDBStorage(t *testing.T) {
	testLogger, err := zap.NewDevelopment()
	if err != nil {
		t.Fatalf("failed to create test logger: %v", err)
	}

	ctx := context.Background()

	database, err := db.NewDatabase("postgres://postgres:postgres@localhost:5432/postgres?sslmode=disable")
	require.NoError(t, err)
	defer database.Conn.Close()

	// Check if table exists and create if not exists
	sql := `
		CREATE TABLE IF NOT EXISTS urls_tests (
			id SERIAL PRIMARY KEY,
			short_id TEXT NOT NULL UNIQUE,
			original_url TEXT NOT NULL,
			created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP
		);
	`
	_, err = database.Conn.Exec(ctx, sql)
	require.NoError(t, err)

	repo := NewURLRepository(database, testLogger).WithTableName("urls_tests")

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
		t.Run(tt.name, func(_ *testing.T) {
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

	// check count of records in db
	var count int
	err = database.Conn.QueryRow(ctx, "SELECT COUNT(*) FROM urls_tests").Scan(&count)
	require.NoError(t, err)
	assert.Equal(t, 2, count)

	// truncate table after ensuring it exists
	_, err = database.Conn.Exec(ctx, "TRUNCATE TABLE urls_tests")
	require.NoError(t, err)
}
