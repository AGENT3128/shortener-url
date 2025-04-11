package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/AGENT3128/shortener-url/internal/app/storage/mocks"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestUserURLsDeleteHandler(t *testing.T) {
	ctx := context.Background()

	repo := mocks.NewMockMemoryRepository()
	logger, err := zap.NewDevelopment()
	if err != nil {
		t.Fatalf("failed to create logger: %v", err)
	}
	shortenHandler := NewUserURLsDeleteHandler(repo, logger)
	gin.SetMode(gin.TestMode)

	// Add test data
	repo.Add(ctx, "test-user", "short1", "ya.ru")
	repo.Add(ctx, "test-user", "short2", "yandex.ru")
	repo.Add(ctx, "test-user", "short3", "google.com")
	repo.Add(ctx, "test-user", "short4", "github.com")

	type want struct {
		contentType string
		statusCode  int
	}
	type request struct {
		method   string
		path     string
		userID   string
		shortIDs []string
	}

	tests := []struct {
		name    string
		request request
		want    want
	}{
		{
			name: "delete empty list",
			request: request{
				method:   http.MethodDelete,
				path:     "/api/user/urls",
				userID:   "test-user",
				shortIDs: []string{},
			},
			want: want{
				statusCode:  http.StatusBadRequest,
				contentType: "application/json; charset=utf-8",
			},
		},
		{
			name: "delete one url",
			request: request{
				method:   http.MethodDelete,
				path:     "/api/user/urls",
				userID:   "test-user",
				shortIDs: []string{"short1"},
			},
			want: want{
				statusCode:  http.StatusAccepted,
				contentType: "application/json; charset=utf-8",
			},
		},
		{
			name: "delete three urls",
			request: request{
				method:   http.MethodDelete,
				path:     "/api/user/urls",
				userID:   "test-user",
				shortIDs: []string{"short2", "short3", "short4"},
			},
			want: want{
				statusCode:  http.StatusAccepted,
				contentType: "application/json; charset=utf-8",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a new router for each test
			router := gin.New()

			// First register middleware
			router.Use(func(c *gin.Context) {
				c.Set("userID", tt.request.userID)
				c.Next()
			})

			// Then register the route
			router.DELETE("/api/user/urls", shortenHandler.Handler())

			// context for request
			requestCtx, requestCancel := context.WithTimeout(ctx, 2*time.Second)
			defer requestCancel()
			w := httptest.NewRecorder()

			// Use json.Marshal to properly encode the shortIDs array
			jsonData, err := json.Marshal(tt.request.shortIDs)
			assert.NoError(t, err)

			request := httptest.NewRequestWithContext(requestCtx, tt.request.method, tt.request.path, strings.NewReader(string(jsonData)))
			request.Header.Set("Content-Type", "application/json")

			router.ServeHTTP(w, request)
			result := w.Result()
			defer result.Body.Close()

			assert.Equal(t, tt.want.statusCode, result.StatusCode)
			assert.Equal(t, tt.want.contentType, result.Header.Get("Content-Type"))

		})
	}

	// check that urls are deleted
	time.Sleep(1 * time.Second)
	for _, tt := range tests {
		t.Run("check deleted "+tt.name, func(t *testing.T) {
			for _, shortID := range tt.request.shortIDs {
				deleted, err := repo.IsURLDeleted(ctx, shortID)
				assert.NoError(t, err)
				assert.True(t, deleted)
			}
		})
	}

}
