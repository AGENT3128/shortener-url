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

func TestShortenBatchHandler(t *testing.T) {
	// context for test
	ctx := context.Background()


	repo := mocks.NewMockMemoryRepository()
	logger, err := zap.NewDevelopment()
	if err != nil {
		t.Fatalf("failed to create logger: %v", err)
	}

	batchHandler := NewShortenBatchHandler(repo, mocks.TestConfig.BaseURLAddress, logger)
	gin.SetMode(gin.TestMode)
	router := gin.Default()
	router.POST("/api/shorten/batch", batchHandler.Handler())

	tests := []struct {
		name       string
		requests   []ShortenBatchRequest
		wantStatus int
	}{
		{
			name: "successful batch creation",
			requests: []ShortenBatchRequest{
				{
					CorrelationID: "1",
					OriginalURL:   "https://ya.ru",
				},
				{
					CorrelationID: "2",
					OriginalURL:   "https://yandex.ru",
				},
			},
			wantStatus: http.StatusCreated,
		},
		{
			name: "duplicate URLs in batch",
			requests: []ShortenBatchRequest{
				{
					CorrelationID: "3",
					OriginalURL:   "https://ya.ru", // already exists
				},
				{
					CorrelationID: "4",
					OriginalURL:   "https://yandex-test.ru", // new URL
				},
			},
			wantStatus: http.StatusCreated,
		},
		{
			name:       "empty batch",
			requests:   []ShortenBatchRequest{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name: "batch with empty URL",
			requests: []ShortenBatchRequest{
				{
					CorrelationID: "5",
					OriginalURL:   "",
				},
			},
			wantStatus: http.StatusCreated,
		},
		{
			name: "batch with empty item",
			requests: []ShortenBatchRequest{
				{
					CorrelationID: "6",
					OriginalURL:   "https://ya.ru",
				},
				{
					CorrelationID: "",
					OriginalURL:   "",
				},
			},
			wantStatus: http.StatusCreated,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// context for request
			requestCtx, requestCancel := context.WithTimeout(ctx, 2*time.Second)
			defer requestCancel()

			// prepare JSON request
			jsonData, err := json.Marshal(tt.requests)
			assert.NoError(t, err)

			// create request
			w := httptest.NewRecorder()
			req := httptest.NewRequestWithContext(requestCtx, http.MethodPost, "/api/shorten/batch", strings.NewReader(string(jsonData)))
			req.Header.Set("Content-Type", "application/json")

			// execute request
			router.ServeHTTP(w, req)
			result := w.Result()
			defer result.Body.Close()

			// check status
			assert.Equal(t, tt.wantStatus, result.StatusCode)

			// if expected successful response, check content
			if tt.wantStatus == http.StatusCreated {
				var response []ShortenBatchResponse
				err := json.NewDecoder(result.Body).Decode(&response)
				assert.NoError(t, err)

				// check format of each response
				for i, resp := range response {
					// check correlation_id
					assert.Equal(t, tt.requests[i].CorrelationID, resp.CorrelationID)

					// check format of short_url
					assert.Contains(t, resp.ShortURL, mocks.TestConfig.BaseURLAddress+"/")

					// check that short_url is not empty
					shortID := strings.TrimPrefix(resp.ShortURL, mocks.TestConfig.BaseURLAddress+"/")
					assert.NotEmpty(t, shortID)
					t.Logf("shortID: %s", shortID)
				}
			}
		})
	}
}
