package memory

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"github.com/AGENT3128/shortener-url/internal/entity"
)

func TestMemStorage_Add(t *testing.T) {
	logger, err := zap.NewDevelopment()
	require.NoError(t, err)
	repo := NewMemStorage(logger)

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
			shortURL, err := repo.Add(t.Context(), tt.userID, tt.shortURL, tt.originalURL)
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			require.Equal(t, tt.shortURL, shortURL)
		})
	}
}

func TestMemStorage_GetByShortURL(t *testing.T) {
	logger, err := zap.NewDevelopment()
	require.NoError(t, err)
	repo := NewMemStorage(logger)

	// Prepare test data
	shortURL := "test-short"
	originalURL := "https://test.com"
	userID := "user1"
	_, err = repo.Add(t.Context(), userID, shortURL, originalURL)
	require.NoError(t, err)

	tests := []struct {
		name      string
		shortURL  string
		want      string
		wantError error
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
			got, err := repo.GetByShortURL(t.Context(), tt.shortURL)
			assert.Equal(t, tt.wantError, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestMemStorage_GetByOriginalURL(t *testing.T) {
	logger, err := zap.NewDevelopment()
	require.NoError(t, err)
	repo := NewMemStorage(logger)

	// Prepare test data
	shortURL := "test-short"
	originalURL := "https://test.com"
	userID := "user1"
	_, err = repo.Add(t.Context(), userID, shortURL, originalURL)
	require.NoError(t, err)

	tests := []struct {
		name        string
		originalURL string
		want        string
		wantError   error
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
			got, err := repo.GetByOriginalURL(t.Context(), tt.originalURL)
			assert.Equal(t, tt.wantError, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestMemStorage_AddBatch(t *testing.T) {
	logger, err := zap.NewDevelopment()
	require.NoError(t, err)
	repo := NewMemStorage(logger)

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
		got, err := repo.GetByShortURL(t.Context(), url.ShortURL)
		require.NoError(t, err)
		assert.Equal(t, url.OriginalURL, got)
	}
}

func TestMemStorage_GetUserURLs(t *testing.T) {
	logger, err := zap.NewDevelopment()
	require.NoError(t, err)
	repo := NewMemStorage(logger)

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
			got, err := repo.GetUserURLs(t.Context(), tt.userID)
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Len(t, got, tt.want)
		})
	}
}

func TestMemStorage_MarkDeletedBatch(t *testing.T) {
	logger, err := zap.NewDevelopment()
	require.NoError(t, err)
	repo := NewMemStorage(logger)

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
	repo := NewMemStorage(logger)

	err = repo.Ping(t.Context())
	require.NoError(t, err)
}
