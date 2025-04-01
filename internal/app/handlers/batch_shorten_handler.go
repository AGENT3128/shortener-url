package handlers

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/AGENT3128/shortener-url/internal/app/helpers"
	"github.com/AGENT3128/shortener-url/internal/app/storage"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// ShortenBatchRequest represents an item in the batch shortening request
type ShortenBatchRequest struct {
	CorrelationID string `json:"correlation_id"`
	OriginalURL   string `json:"original_url"`
}

// ShortenBatchResponse represents an item in the batch shortening response
type ShortenBatchResponse struct {
	CorrelationID string `json:"correlation_id"`
	ShortURL      string `json:"short_url"`
}

// ShortenBatchHandler handles batch creation of short URLs
type ShortenBatchHandler struct {
	repository URLRepository
	baseURL    string
	logger     *zap.Logger
}

// NewShortenBatchHandler creates a new instance of ShortenBatchHandler
func NewShortenBatchHandler(repo URLRepository, baseURL string, logger *zap.Logger) *ShortenBatchHandler {
	logger = logger.With(zap.String("handler", "ShortenBatchHandler"))
	return &ShortenBatchHandler{
		repository: repo,
		baseURL:    baseURL,
		logger:     logger,
	}
}

// Pattern returns the URL pattern for the handler
func (h *ShortenBatchHandler) Pattern() string {
	return "/api/shorten/batch"
}

// Method returns the HTTP method for the handler
func (h *ShortenBatchHandler) Method() string {
	return http.MethodPost
}

// Handler returns the gin.HandlerFunc for the handler
func (h *ShortenBatchHandler) Handler() gin.HandlerFunc {
	return func(c *gin.Context) {
		body, err := io.ReadAll(c.Request.Body)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to read request body"})
			return
		}
		defer c.Request.Body.Close()

		var requests []ShortenBatchRequest
		if err := json.Unmarshal(body, &requests); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON format in request body"})
			return
		}

		if len(requests) == 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Original URL is empty"})
			return
		}

		// prepare data
		var urls []storage.URL
		response := make([]ShortenBatchResponse, 0, len(requests))
		txRepo, supportsTx := h.repository.(TransactionSupport)

		// process requests
		for _, request := range requests {
			if request.OriginalURL == "" || request.CorrelationID == "" {
				continue
			}

			shortID, exists := h.repository.GetByOriginalURL(request.OriginalURL)
			if !exists {
				shortID = helpers.GenerateShortID()
				// for DB, collect batch, for other storages add immediately
				if supportsTx {
					urls = append(urls, storage.URL{
						ShortID:     shortID,
						OriginalURL: request.OriginalURL,
					})
				} else {
					h.repository.Add(shortID, request.OriginalURL)
				}
			}

			response = append(response, ShortenBatchResponse{
				CorrelationID: request.CorrelationID,
				ShortURL:      fmt.Sprintf("%s/%s", h.baseURL, shortID),
			})
		}

		// batch adding for DB
		if supportsTx && len(urls) > 0 {
			if err := txRepo.AddBatch(urls); err != nil {
				h.logger.Error("failed to add batch", zap.Error(err))
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
				return
			}
		}

		c.JSON(http.StatusCreated, response)
	}
}
