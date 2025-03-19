package handlers

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/AGENT3128/shortener-url/internal/app/storage"
	"github.com/stretchr/testify/assert"
)

func TestURLHandler(t *testing.T) {
	storage.InitStorage()

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

	var firstShortID, secondShortID string

	w1 := httptest.NewRecorder()
	req1 := httptest.NewRequest(http.MethodPost, "/", strings.NewReader("ya.ru"))
	URLHandler(w1, req1)
	firstShortID = strings.TrimPrefix(w1.Body.String(), "http://localhost:8080/")

	w2 := httptest.NewRecorder()
	req2 := httptest.NewRequest(http.MethodPost, "/", strings.NewReader("yandex.ru"))
	URLHandler(w2, req2)
	secondShortID = strings.TrimPrefix(w2.Body.String(), "http://localhost:8080/")

	assert.NotEqual(t, firstShortID, secondShortID)

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
			name: "get first original URL",
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
			name: "get second original URL",
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

			URLHandler(w, request)
			result := w.Result()

			assert.Equal(t, tt.want.statusCode, result.StatusCode)
			assert.Equal(t, tt.want.contentType, result.Header.Get("Content-Type"))
			assert.Equal(t, tt.want.location, result.Header.Get("Location"))
		})
	}
}
