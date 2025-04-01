package handlers

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/AGENT3128/shortener-url/internal/app/storage/mocks"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestShortenHandler(t *testing.T) {
	repo := mocks.NewMockMemoryRepository()
	logger, err := zap.NewDevelopment()
	if err != nil {
		t.Fatalf("failed to create logger: %v", err)
	}
	shortenHandler := NewShortenHandler(repo, mocks.TestConfig.BaseURLAddress, logger)
	gin.SetMode(gin.TestMode)
	router := gin.Default()
	router.POST("/", shortenHandler.Handler())

	type want struct {
		contentType string
		statusCode  int
		location    string
	}
	type request struct {
		method string
		path   string
		body   string
	}

	tests := []struct {
		name    string
		request request
		want    want
	}{
		{
			name: "create first short URL",
			request: request{
				method: http.MethodPost,
				path:   "/",
				body:   "ya.ru",
			},
			want: want{
				statusCode:  http.StatusCreated,
				contentType: "text/plain",
			},
		},
		{
			name: "create second short URL",
			request: request{
				method: http.MethodPost,
				path:   "/",
				body:   "yandex.ru",
			},
			want: want{
				statusCode:  http.StatusCreated,
				contentType: "text/plain",
			},
		},
		{
			name: "create empty short URL",
			request: request{
				method: http.MethodPost,
				path:   "/",
				body:   "",
			},
			want: want{
				statusCode:  http.StatusBadRequest,
				contentType: "application/json; charset=utf-8",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			request := httptest.NewRequest(tt.request.method, tt.request.path, strings.NewReader(tt.request.body))
			router.ServeHTTP(w, request)
			result := w.Result()
			defer result.Body.Close()

			assert.Equal(t, tt.want.statusCode, result.StatusCode)
			assert.Equal(t, tt.want.contentType, result.Header.Get("Content-Type"))
			assert.Equal(t, tt.want.location, result.Header.Get("Location"))
		})
	}
}
