package memory_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"github.com/AGENT3128/shortener-url/internal/entity"
	"github.com/AGENT3128/shortener-url/internal/repository/memory"
)

func TestMemStorage_Add(t *testing.T) {
	logger, err := zap.NewDevelopment()
	require.NoError(t, err)
	repo := memory.NewMemStorage(logger)

	tests := []struct {
		name        string
		shortURL    string
		originalURL string
		userID      string
		wantErr     bool
	}{
		{
			name:        "success",
			shortURL:    "shortURL1",
			originalURL: "https://example.com",
			userID:      "userID1",
			wantErr:     false,
		},
		{
			name:        "success_another_url",
			shortURL:    "shortURL2",
			originalURL: "https://another-example.com",
			userID:      "userID1",
			wantErr:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			shortURL, errAdd := repo.Add(t.Context(), tt.userID, tt.shortURL, tt.originalURL)
			if tt.wantErr {
				require.Error(t, errAdd)
				return
			}
			require.NoError(t, errAdd)
			require.Equal(t, tt.shortURL, shortURL)
		})
	}
}

func TestMemStorage_GetByShortURL(t *testing.T) {
	logger, err := zap.NewDevelopment()
	require.NoError(t, err)
	repo := memory.NewMemStorage(logger)

	// Prepare test data
	shortURL := "test-short"
	originalURL := "https://test.com"
	userID := "user1"
	_, err = repo.Add(t.Context(), userID, shortURL, originalURL)
	require.NoError(t, err)

	tests := []struct {
		wantError error
		name      string
		shortURL  string
		want      string
	}{
		{
			name:      "existing_url",
			shortURL:  shortURL,
			want:      originalURL,
			wantError: nil,
		},
		{
			name:      "non_existing_url",
			shortURL:  "non-existing",
			want:      "",
			wantError: entity.ErrURLNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, errGet := repo.GetByShortURL(t.Context(), tt.shortURL)
			assert.Equal(t, tt.wantError, errGet)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestMemStorage_GetByOriginalURL(t *testing.T) {
	logger, err := zap.NewDevelopment()
	require.NoError(t, err)
	repo := memory.NewMemStorage(logger)

	// Prepare test data
	shortURL := "test-short"
	originalURL := "https://test.com"
	userID := "user1"
	_, err = repo.Add(t.Context(), userID, shortURL, originalURL)
	require.NoError(t, err)

	tests := []struct {
		wantError   error
		name        string
		originalURL string
		want        string
	}{
		{
			name:        "existing_url",
			originalURL: originalURL,
			want:        shortURL,
			wantError:   nil,
		},
		{
			name:        "non_existing_url",
			originalURL: "https://non-existing.com",
			want:        "",
			wantError:   entity.ErrURLNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, errGet := repo.GetByOriginalURL(t.Context(), tt.originalURL)
			assert.Equal(t, tt.wantError, errGet)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestMemStorage_AddBatch(t *testing.T) {
	logger, err := zap.NewDevelopment()
	require.NoError(t, err)
	repo := memory.NewMemStorage(logger)

	urls := []entity.URL{
		{
			ShortURL:    "short1",
			OriginalURL: "https://test1.com",
		},
		{
			ShortURL:    "short2",
			OriginalURL: "https://test2.com",
		},
	}

	userID := "user1"
	err = repo.AddBatch(t.Context(), userID, urls)
	require.NoError(t, err)

	// Verify each URL was added correctly
	for _, url := range urls {
		got, errGet := repo.GetByShortURL(t.Context(), url.ShortURL)
		require.NoError(t, errGet)
		assert.Equal(t, url.OriginalURL, got)
	}
}

func TestMemStorage_GetUserURLs(t *testing.T) {
	logger, err := zap.NewDevelopment()
	require.NoError(t, err)
	repo := memory.NewMemStorage(logger)

	userID1 := "user1"
	userID2 := "user2"

	// Add URLs for different users
	urls1 := []entity.URL{
		{ShortURL: "short1", OriginalURL: "https://test1.com"},
		{ShortURL: "short2", OriginalURL: "https://test2.com"},
	}
	urls2 := []entity.URL{
		{ShortURL: "short3", OriginalURL: "https://test3.com"},
	}

	err = repo.AddBatch(t.Context(), userID1, urls1)
	require.NoError(t, err)
	err = repo.AddBatch(t.Context(), userID2, urls2)
	require.NoError(t, err)

	tests := []struct {
		name    string
		userID  string
		want    int
		wantErr bool
	}{
		{
			name:    "user_with_multiple_urls",
			userID:  userID1,
			want:    2,
			wantErr: false,
		},
		{
			name:    "user_with_single_url",
			userID:  userID2,
			want:    1,
			wantErr: false,
		},
		{
			name:    "user_with_no_urls",
			userID:  "non-existing",
			want:    0,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, errGet := repo.GetUserURLs(t.Context(), tt.userID)
			if tt.wantErr {
				require.Error(t, errGet)
				return
			}
			require.NoError(t, errGet)
			assert.Len(t, got, tt.want)
		})
	}
}

func TestMemStorage_MarkDeletedBatch(t *testing.T) {
	logger, err := zap.NewDevelopment()
	require.NoError(t, err)
	repo := memory.NewMemStorage(logger)

	userID := "user1"
	urls := []entity.URL{
		{ShortURL: "short1", OriginalURL: "https://test1.com"},
		{ShortURL: "short2", OriginalURL: "https://test2.com"},
	}

	// Add URLs
	err = repo.AddBatch(t.Context(), userID, urls)
	require.NoError(t, err)

	// Mark URLs as deleted
	shortURLsToDelete := []string{"short1", "short2"}
	err = repo.MarkDeletedBatch(t.Context(), userID, shortURLsToDelete)
	require.NoError(t, err)

	deletedURLs, err := repo.GetUserURLs(t.Context(), userID)
	require.NoError(t, err)
	assert.Len(t, deletedURLs, 2)
	for _, url := range deletedURLs {
		assert.True(t, url.DeletedFlag)
	}
}

func TestMemStorage_Ping(t *testing.T) {
	logger, err := zap.NewDevelopment()
	require.NoError(t, err)
	repo := memory.NewMemStorage(logger)

	err = repo.Ping(t.Context())
	require.NoError(t, err)
}

func TestMemStorage_GetStats(t *testing.T) {
	logger, err := zap.NewDevelopment()
	require.NoError(t, err)
	repo := memory.NewMemStorage(logger)

	urlsCount, usersCount, err := repo.GetStats(t.Context())
	require.NoError(t, err)
	require.Equal(t, 0, urlsCount)
	require.Equal(t, 0, usersCount)

	urls := []entity.URL{
		{ShortURL: "short1", OriginalURL: "https://test1.com"},
		{ShortURL: "short2", OriginalURL: "https://test2.com"},
	}

	err = repo.AddBatch(t.Context(), "user1", urls)
	require.NoError(t, err)

	urlsCount, usersCount, err = repo.GetStats(t.Context())
	require.NoError(t, err)
	require.Equal(t, 2, urlsCount)
	require.Equal(t, 1, usersCount)
}
