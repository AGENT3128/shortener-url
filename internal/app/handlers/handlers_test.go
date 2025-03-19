package handlers

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
)

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

func TestURLHandler(t *testing.T) {
	repo := NewMockRepository()
	handler := NewURLHandler(repo)

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
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			request := httptest.NewRequest(tt.request.method, tt.request.path, strings.NewReader(tt.request.body))
			w := httptest.NewRecorder()

			handler.ServeHTTP(w, request)
			result := w.Result()
			defer result.Body.Close()

			assert.Equal(t, tt.want.statusCode, result.StatusCode)
			assert.Equal(t, tt.want.contentType, result.Header.Get("Content-Type"))
			assert.Equal(t, tt.want.location, result.Header.Get("Location"))
		})
	}

	firstShortID, ok := repo.GetByOriginalURL("ya.ru")
	assert.True(t, ok)

	secondShortID, ok := repo.GetByOriginalURL("yandex.ru")
	assert.True(t, ok)

	tests = []struct {
		name    string
		request request
		want    want
	}{
		{
			name: "get first original URL from short URL",
			request: request{
				method: http.MethodGet,
				path:   "/" + firstShortID,
			},
			want: want{
				statusCode:  http.StatusTemporaryRedirect,
				contentType: "text/plain",
				location:    "ya.ru",
			},
		},
		{
			name: "get second original URL from short URL",
			request: request{
				method: http.MethodGet,
				path:   "/" + secondShortID,
			},
			want: want{
				statusCode:  http.StatusTemporaryRedirect,
				contentType: "text/plain",
				location:    "yandex.ru",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			request := httptest.NewRequest(tt.request.method, tt.request.path, strings.NewReader(tt.request.body))
			w := httptest.NewRecorder()

			handler.ServeHTTP(w, request)
			result := w.Result()
			defer result.Body.Close()

			assert.Equal(t, tt.want.statusCode, result.StatusCode)
			assert.Equal(t, tt.want.contentType, result.Header.Get("Content-Type"))
			assert.Equal(t, tt.want.location, result.Header.Get("Location"))
		})
	}
}
