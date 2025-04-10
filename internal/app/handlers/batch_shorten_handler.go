package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/AGENT3128/shortener-url/internal/app/helpers"
	"github.com/AGENT3128/shortener-url/internal/app/models"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// BatchShortenI interface for batch operations
type BatchShortenI interface {
	AddBatchSetter
	URLOriginalGetter
}

// AddBatchSetter interface for batch operations
type AddBatchSetter interface {
	AddBatch(ctx context.Context, userID string, urls []models.URL) error
}

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
	repository BatchShortenI
	baseURL    string
	logger     *zap.Logger
}

// NewShortenBatchHandler creates a new instance of ShortenBatchHandler
func NewShortenBatchHandler(repo BatchShortenI, baseURL string, logger *zap.Logger) *ShortenBatchHandler {
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
		userID, ok := c.Get("userID")
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
			return
		}

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
			c.JSON(http.StatusBadRequest, gin.H{"error": "Request body is empty"})
			return
		}

		// prepare data
		var urls []models.URL
		response := make([]ShortenBatchResponse, 0, len(requests))

		// process requests
		for _, request := range requests {
			if request.OriginalURL == "" || request.CorrelationID == "" {
				continue
			}

			shortID, exists := h.repository.GetByOriginalURL(c.Request.Context(), request.OriginalURL)
			if !exists {
				shortID = helpers.GenerateShortID()
				urls = append(urls, models.URL{
					ShortID:     shortID,
					OriginalURL: request.OriginalURL,
				})
			}

			response = append(response, ShortenBatchResponse{
				CorrelationID: request.CorrelationID,
				ShortURL:      fmt.Sprintf("%s/%s", h.baseURL, shortID),
			})
		}

		// batch adding for DB
		if len(urls) > 0 {
			if err := h.repository.AddBatch(c.Request.Context(), userID.(string), urls); err != nil {
				h.logger.Error("failed to add batch", zap.Error(err))
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
				return
			}
		}

		c.JSON(http.StatusCreated, response)
	}
}
