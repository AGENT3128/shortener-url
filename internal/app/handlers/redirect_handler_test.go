package handlers

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/AGENT3128/shortener-url/internal/app/storage/mocks"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestRedirectHandler(t *testing.T) {
	// base context for setup
	ctx := context.Background()

	repo := mocks.NewMockMemoryRepository()
	logger, err := zap.NewDevelopment()
	if err != nil {
		t.Fatalf("failed to create logger: %v", err)
	}
	handler := NewRedirectHandler(repo, logger)

	gin.SetMode(gin.TestMode)
	router := gin.Default()
	router.Use(func(c *gin.Context) {
		c.Set("userID", "test-user")
		c.Next()
	})
	router.GET("/:id", handler.Handler())

	// setup test cases
	testCases := map[string]string{
		"abc123": "ya.ru",
		"def456": "yandex.ru",
	}

	for shortID, originalURL := range testCases {
		// base context for setup
		setupCtx, setupCancel := context.WithTimeout(ctx, 5*time.Second)
		defer setupCancel()
		shortID, err := repo.Add(setupCtx, "test-user", shortID, originalURL)
		if err != nil {
			t.Fatalf("failed to add url: %v", err)
		}
		t.Logf("added url: %s -> %s", shortID, originalURL)
	}

	tests := []struct {
		name    string
		shortID string
		want    struct {
			statusCode  int
			contentType string
			location    string
		}
	}{
		{
			name:    "successful redirect",
			shortID: "abc123",
			want: struct {
				statusCode  int
				contentType string
				location    string
			}{
				statusCode:  http.StatusTemporaryRedirect,
				contentType: "text/plain",
				location:    "ya.ru",
			},
		},
		{
			name:    "not found",
			shortID: "nonexistent",
			want: struct {
				statusCode  int
				contentType string
				location    string
			}{
				statusCode:  http.StatusNotFound,
				contentType: "application/json; charset=utf-8",
				location:    "",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// context for request
			requestCtx, requestCancel := context.WithTimeout(ctx, 2*time.Second)
			defer requestCancel()

			w := httptest.NewRecorder()
			request := httptest.NewRequestWithContext(requestCtx, http.MethodGet, "/"+tt.shortID, nil)
			router.ServeHTTP(w, request)
			result := w.Result()
			defer result.Body.Close()

			assert.Equal(t, tt.want.statusCode, result.StatusCode)
			assert.Equal(t, tt.want.contentType, result.Header.Get("Content-Type"))
			assert.Equal(t, tt.want.location, result.Header.Get("Location"))
		})
	}
}
