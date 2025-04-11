package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/AGENT3128/shortener-url/internal/app/helpers"
	"github.com/AGENT3128/shortener-url/internal/app/models"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// URLOriginalGetter describes the behavior for retrieving short URL by original URL
type URLOriginalGetter interface {
	GetByOriginalURL(ctx context.Context, originalURL string) (string, bool)
}

// URLSaver describes the behavior for saving short URL by original URL
type URLSaver interface {
	Add(ctx context.Context, userID, shortID, originalURL string) (string, error)
}

// URLRepository combines URL getting and saving capabilities
type URLRepository interface {
	URLOriginalGetter
	URLSaver
}

// ShortenRequest represents the request for shortening a URL
type ShortenRequest struct {
	URL string `json:"url"`
}

// ShortenResponse represents the response for a shortened URL
type ShortenResponse struct {
	Result string `json:"result"`
}

// APIShortenHandler handles the creation of short URLs through JSON API
type APIShortenHandler struct {
	repository URLRepository
	baseURL    string
	logger     *zap.Logger
}

// NewAPIShortenHandler creates a new instance of APIShortenHandler
func NewAPIShortenHandler(repo URLRepository, baseURL string, logger *zap.Logger) *APIShortenHandler {
	logger = logger.With(zap.String("handler", "APIShortenHandler"))
	return &APIShortenHandler{
		repository: repo,
		baseURL:    baseURL,
		logger:     logger,
	}
}

// Pattern returns the URL pattern for the handler
func (h *APIShortenHandler) Pattern() string {
	return "/api/shorten"
}

// Method returns the HTTP method for the handler
func (h *APIShortenHandler) Method() string {
	return http.MethodPost
}

// Handler returns the gin.HandlerFunc for the handler
func (h *APIShortenHandler) Handler() gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, ok := c.Get("userID")
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
			return
		}

		var request ShortenRequest

		body, err := io.ReadAll(c.Request.Body)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to read request body"})
			return
		}
		defer c.Request.Body.Close()

		if err := json.Unmarshal(body, &request); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON format in request body"})
			return
		}

		if request.URL == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "URL is empty"})
			return
		}

		shortID, err := h.repository.Add(c.Request.Context(), userID.(string), helpers.GenerateShortID(), request.URL)
		if err != nil {
			if errors.Is(err, models.ErrURLExists) {
				h.logger.Info("URL already exists", zap.String("originalURL", request.URL))
				c.JSON(http.StatusConflict, ShortenResponse{Result: fmt.Sprintf("%s/%s", h.baseURL, shortID)})
				return
			}
			h.logger.Error("failed to add to repository", zap.Error(err))
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
			return
		}
		h.logger.Info("add to repository", zap.String("shortID", shortID), zap.String("originalURL", request.URL))
		c.JSON(http.StatusCreated, ShortenResponse{Result: fmt.Sprintf("%s/%s", h.baseURL, shortID)})
	}
}
