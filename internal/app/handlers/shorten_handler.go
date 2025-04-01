package handlers

import (
	"errors"
	"io"
	"net/http"

	"github.com/AGENT3128/shortener-url/internal/app/helpers"
	"github.com/AGENT3128/shortener-url/internal/app/storage"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// ShortenHandler handles the creation of short URLs through plain text endpoint
type ShortenHandler struct {
	repository URLRepository
	baseURL    string
	logger     *zap.Logger
}

// NewShortenHandler creates a new instance of ShortenHandler
func NewShortenHandler(repo URLRepository, baseURL string, logger *zap.Logger) *ShortenHandler {
	logger = logger.With(zap.String("handler", "ShortenHandler"))
	return &ShortenHandler{
		repository: repo,
		baseURL:    baseURL,
		logger:     logger,
	}
}

// Pattern returns the URL pattern for the handler
func (h *ShortenHandler) Pattern() string {
	return "/"
}

// Method returns the HTTP method for the handler
func (h *ShortenHandler) Method() string {
	return http.MethodPost
}

// Handler returns the gin.HandlerFunc for the handler
func (h *ShortenHandler) Handler() gin.HandlerFunc {
	return func(c *gin.Context) {
		body, err := io.ReadAll(c.Request.Body)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to read request body"})
			return
		}
		defer c.Request.Body.Close()

		originalURL := string(body)
		if originalURL == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Original URL is empty"})
			return
		}

		shortID, err := h.repository.Add(helpers.GenerateShortID(), originalURL)
		if err != nil {
			if errors.Is(err, storage.ErrURLExists) {
				h.logger.Info("URL already exists", zap.String("originalURL", originalURL))
				c.String(http.StatusConflict, "%s/%s", h.baseURL, shortID)
				return
			}
			h.logger.Error("failed to add to repository", zap.Error(err))
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
			return
		}
		h.logger.Info("add to repository", zap.String("shortID", shortID), zap.String("originalURL", originalURL))

		c.Header("Content-Type", "text/plain")
		c.String(http.StatusCreated, "%s/%s", h.baseURL, shortID)
	}
}
