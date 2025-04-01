package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/AGENT3128/shortener-url/internal/app/storage/mocks"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestAPIShortenHandler(t *testing.T) {
	repo := mocks.NewMockMemoryRepository()
	logger, err := zap.NewDevelopment()
	if err != nil {
		t.Fatalf("failed to create logger: %v", err)
	}
	shortenHandler := NewAPIShortenHandler(repo, mocks.TestConfig.BaseURLAddress, logger)
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
					Result: mocks.TestConfig.BaseURLAddress + "/",
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
					Result: mocks.TestConfig.BaseURLAddress + "/",
				},
			},
		},
		{
			name: "create short URL that already exists",
			request: request{
				method: http.MethodPost,
				path:   "/api/shorten",
				body:   ShortenRequest{URL: "ya.ru"},
			},
			want: want{
				statusCode:  http.StatusConflict,
				contentType: "application/json; charset=utf-8",
				response: ShortenResponse{
					Result: mocks.TestConfig.BaseURLAddress + "/",
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
				assert.Contains(t, response.Result, mocks.TestConfig.BaseURLAddress+"/")

				// check that after base URL there is a non-empty identifier
				shortID := strings.TrimPrefix(response.Result, mocks.TestConfig.BaseURLAddress+"/")
				assert.NotEmpty(t, shortID)
			}
		})
	}
}
