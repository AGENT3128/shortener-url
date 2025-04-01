package middleware

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type ShortenRequest struct {
	URL string `json:"url"`
}

type ShortenResponse struct {
	Result string `json:"result"`
}

func TestGzipMiddleware(t *testing.T) {
	// context for test
	ctx := context.Background()

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(GzipMiddleware())

	// test handler without business logic
	router.POST("/test", func(c *gin.Context) {
		var req ShortenRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
			return
		}

		c.JSON(http.StatusCreated, ShortenResponse{
			Result: "http://localhost:8080/" + "testID",
		})
	})

	tests := []struct {
		name           string
		url            string
		useGzipRequest bool
		acceptGzip     bool
		expectedStatus int
	}{
		{
			name:           "Compressed request and response",
			useGzipRequest: true,
			acceptGzip:     true,
			expectedStatus: http.StatusCreated,
		},
		{
			name:           "Uncompressed request, compressed response",
			useGzipRequest: false,
			acceptGzip:     true,
			expectedStatus: http.StatusCreated,
		},
		{
			name:           "Compressed request, uncompressed response",
			useGzipRequest: true,
			acceptGzip:     false,
			expectedStatus: http.StatusCreated,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			requestCtx, requestCancel := context.WithTimeout(ctx, 2*time.Second)
			defer requestCancel()

			if tt.url == "" {
				tt.url = fmt.Sprintf("https://%s.ru", strings.Repeat("yandex", 500))
			}
			jsonBody := fmt.Sprintf(`{"url": "%s"}`, tt.url)
			t.Log("size of jsonBody", len(jsonBody))

			var reqBody io.Reader
			if tt.useGzipRequest {
				compressedData := compressInternal(t, jsonBody)
				t.Log("size of compressedData", len(compressedData))
				reqBody = bytes.NewReader(compressedData)
			} else {
				reqBody = strings.NewReader(jsonBody)
			}

			req := httptest.NewRequestWithContext(requestCtx, http.MethodPost, "/test", reqBody)
			req.Header.Set("Content-Type", "application/json")

			if tt.useGzipRequest {
				req.Header.Set("Content-Encoding", "gzip")
			}

			if tt.acceptGzip {
				req.Header.Set("Accept-Encoding", "gzip")
			}

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)
			resp := w.Result()
			defer resp.Body.Close()

			assert.Equal(t, tt.expectedStatus, resp.StatusCode)

			if tt.expectedStatus == http.StatusCreated {
				var responseBody []byte
				var err error

				if tt.acceptGzip {
					assert.Equal(t, "gzip", resp.Header.Get("Content-Encoding"))

					reader, err := gzip.NewReader(resp.Body)
					require.NoError(t, err)

					responseBody, err = io.ReadAll(reader)
					require.NoError(t, err)

					reader.Close()
				} else {
					responseBody, err = io.ReadAll(resp.Body)
					require.NoError(t, err)
				}
				t.Log("size of responseBody", len(responseBody))
				var response ShortenResponse
				err = json.Unmarshal(responseBody, &response)
				require.NoError(t, err)

				assert.Contains(t, response.Result, "http://localhost:8080/")
				assert.Contains(t, response.Result, "testID")
			}
		})
	}
}

func compressInternal(t *testing.T, data string) []byte {
	var buf bytes.Buffer
	gzipWriter := gzip.NewWriter(&buf)

	_, err := gzipWriter.Write([]byte(data))
	require.NoError(t, err)

	err = gzipWriter.Close()
	require.NoError(t, err)

	return buf.Bytes()
}
