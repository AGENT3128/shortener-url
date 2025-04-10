package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/AGENT3128/shortener-url/internal/app/storage/mocks"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestUserURLsHandler(t *testing.T) {
	ctx := context.Background()

	repo := mocks.NewMockMemoryRepository()
	logger, err := zap.NewDevelopment()
	if err != nil {
		t.Fatalf("failed to create logger: %v", err)
	}
	shortenHandler := NewUserURLsHandler(repo, mocks.TestConfig.BaseURLAddress, logger)
	gin.SetMode(gin.TestMode)

	// Add test data
	repo.Add(ctx, "test-user", "short1", "ya.ru")
	repo.Add(ctx, "test-user", "short2", "yandex.ru")
	repo.Add(ctx, "test-user1", "short3", "google.com")
	repo.Add(ctx, "test-user1", "short4", "github.com")

	type want struct {
		contentType string
		statusCode  int
		response    int // count of urls
	}
	type request struct {
		method string
		path   string
		userID string
	}

	tests := []struct {
		name    string
		request request
		want    want
	}{
		{
			name: "first sample for the user",
			request: request{
				method: http.MethodGet,
				path:   "/api/user/urls",
				userID: "test-user",
			},
			want: want{
				statusCode:  http.StatusOK,
				contentType: "application/json; charset=utf-8",
				response:    2,
			},
		},
		{
			name: "second sample for the user",
			request: request{
				method: http.MethodGet,
				path:   "/api/user/urls",
				userID: "test-user1",
			},
			want: want{
				statusCode:  http.StatusOK,
				contentType: "application/json; charset=utf-8",
				response:    2,
			},
		},
		{
			name: "third sample for the user",
			request: request{
				method: http.MethodGet,
				path:   "/api/user/urls",
				userID: "test-user2",
			},
			want: want{
				statusCode:  http.StatusNoContent,
				contentType: "application/json; charset=utf-8",
				response:    0,
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
			router.GET("/api/user/urls", shortenHandler.Handler())

			// context for request
			requestCtx, requestCancel := context.WithTimeout(ctx, 2*time.Second)
			defer requestCancel()
			w := httptest.NewRecorder()
			request := httptest.NewRequestWithContext(requestCtx, tt.request.method, tt.request.path, nil)

			router.ServeHTTP(w, request)
			result := w.Result()
			defer result.Body.Close()

			assert.Equal(t, tt.want.statusCode, result.StatusCode)

			// Проверяем Content-Type только если он ожидается
			if tt.want.contentType != "" {
				assert.Equal(t, tt.want.contentType, result.Header.Get("Content-Type"))
			}

			// check body response
			if tt.want.statusCode == http.StatusOK {
				var response []UserURLResponse
				err := json.NewDecoder(result.Body).Decode(&response)
				assert.NoError(t, err)

				assert.Equal(t, tt.want.response, len(response))
			}
		})
	}
}
