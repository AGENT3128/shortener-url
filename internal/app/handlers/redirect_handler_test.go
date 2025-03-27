package handlers

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestRedirectHandler(t *testing.T) {
	repo := NewMockRepository()
	logger, err := zap.NewDevelopment()
	if err != nil {
		t.Fatalf("failed to create logger: %v", err)
	}
	handler := NewRedirectHandler(repo, logger)

	gin.SetMode(gin.TestMode)
	router := gin.Default()
	router.GET("/:id", handler.Handler())

	// setup test cases
	testCases := map[string]string{
		"abc123": "ya.ru",
		"def456": "yandex.ru",
	}

	for shortID, originalURL := range testCases {
		repo.Add(shortID, originalURL)
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
			w := httptest.NewRecorder()
			request := httptest.NewRequest(http.MethodGet, "/"+tt.shortID, nil)
			router.ServeHTTP(w, request)
			result := w.Result()
			defer result.Body.Close()

			assert.Equal(t, tt.want.statusCode, result.StatusCode)
			assert.Equal(t, tt.want.contentType, result.Header.Get("Content-Type"))
			assert.Equal(t, tt.want.location, result.Header.Get("Location"))
		})
	}
}
