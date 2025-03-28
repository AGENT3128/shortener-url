package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"

	"github.com/AGENT3128/shortener-url/internal/app/config"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

var testConfig = &config.Config{
	ServerAddress:  "localhost:8080",
	BaseURLAddress: "http://localhost:8080",
	ReleaseMode:    "debug",
	LogLevel:       "info",
}

type MockRepository struct {
	mu   sync.RWMutex
	urls map[string]string
}

func NewMockRepository() *MockRepository {
	return &MockRepository{
		urls: make(map[string]string),
	}
}

func (m *MockRepository) Add(shortID, originalURL string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.urls[shortID] = originalURL
}

func (m *MockRepository) GetByShortID(shortID string) (string, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	originalURL, ok := m.urls[shortID]
	return originalURL, ok
}

func (m *MockRepository) GetByOriginalURL(originalURL string) (string, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	for shortID, url := range m.urls {
		if url == originalURL {
			return shortID, true
		}
	}
	return "", false
}

func TestShortenHandler(t *testing.T) {
	// cfg := config.NewConfig()
	repo := NewMockRepository()
	logger, err := zap.NewDevelopment()
	if err != nil {
		t.Fatalf("failed to create logger: %v", err)
	}
	shortenHandler := NewShortenHandler(repo, testConfig.BaseURLAddress, logger)
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

func TestAPIShortenHandler(t *testing.T) {
	// cfg := config.NewConfig()
	repo := NewMockRepository()
	logger, err := zap.NewDevelopment()
	if err != nil {
		t.Fatalf("failed to create logger: %v", err)
	}
	shortenHandler := NewAPIShortenHandler(repo, testConfig.BaseURLAddress, logger)
	gin.SetMode(gin.TestMode)
	router := gin.Default()
	router.POST("/api/shorten", shortenHandler.Handler())

	type want struct {
		contentType string
		statusCode  int
		response    ShortenResponse
	}
	type request struct {
		method string
		path   string
		body   ShortenRequest
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
				path:   "/api/shorten",
				body:   ShortenRequest{URL: "ya.ru"},
			},
			want: want{
				statusCode:  http.StatusCreated,
				contentType: "application/json; charset=utf-8",
				response: ShortenResponse{
					Result: testConfig.BaseURLAddress + "/",
				},
			},
		},
		{
			name: "create second short URL",
			request: request{
				method: http.MethodPost,
				path:   "/api/shorten",
				body:   ShortenRequest{URL: "yandex.ru"},
			},
			want: want{
				statusCode:  http.StatusCreated,
				contentType: "application/json; charset=utf-8",
				response: ShortenResponse{
					Result: testConfig.BaseURLAddress + "/",
				},
			},
		},
		{
			name: "create empty short URL",
			request: request{
				method: http.MethodPost,
				path:   "/api/shorten",
				body:   ShortenRequest{URL: ""},
			},
			want: want{
				statusCode:  http.StatusBadRequest,
				contentType: "application/json; charset=utf-8",
				response:    ShortenResponse{},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			jsonBody := fmt.Sprintf(`{"url": "%s"}`, tt.request.body.URL)
			w := httptest.NewRecorder()
			request := httptest.NewRequest(tt.request.method, tt.request.path, strings.NewReader(jsonBody))
			//request.Header.Set("Content-Type", "application/json")
			router.ServeHTTP(w, request)
			result := w.Result()
			defer result.Body.Close()

			assert.Equal(t, tt.want.statusCode, result.StatusCode)
			assert.Equal(t, tt.want.contentType, result.Header.Get("Content-Type"))
			// check body response
			if tt.want.statusCode == http.StatusCreated {
				var response ShortenResponse
				err := json.NewDecoder(result.Body).Decode(&response)
				assert.NoError(t, err)

				// check that response contains correct base URL
				assert.Contains(t, response.Result, testConfig.BaseURLAddress+"/")

				// check that after base URL there is a non-empty identifier
				shortID := strings.TrimPrefix(response.Result, testConfig.BaseURLAddress+"/")
				assert.NotEmpty(t, shortID)
			}
		})
	}
}
