package handlers

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/AGENT3128/shortener-url/internal/app/helpers"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// URLOriginalGetter describes the behavior for retrieving short URL by original URL
type URLOriginalGetter interface {
	GetByOriginalURL(originalURL string) (string, bool)
}

// URLSaver describes the behavior for saving short URL by original URL
type URLSaver interface {
	Add(shortID, originalURL string)
}

// URLRepository combines URL getting and saving capabilities
type URLRepository interface {
	URLOriginalGetter
	URLSaver
}

// ShortenHandler handles the creation of short URLs
type ShortenHandler struct {
	repository URLRepository
	baseURL    string
	logger     *zap.Logger
}

func NewShortenHandler(repo URLRepository, baseURL string, logger *zap.Logger) *ShortenHandler {
	logger = logger.With(zap.String("handler", "ShortenHandler"))
	return &ShortenHandler{
		repository: repo,
		baseURL:    baseURL,
		logger:     logger,
	}
}

func (h *ShortenHandler) Pattern() string {
	return "/"
}

func (h *ShortenHandler) Method() string {
	return http.MethodPost
}

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

		// before adding to repository, check if the original URL already exists
		shortID, ok := h.repository.GetByOriginalURL(originalURL)
		h.logger.Info("get original URL", zap.String("originalURL", originalURL), zap.Bool("exists", ok))
		if !ok {
			shortID = helpers.GenerateShortID()
			h.repository.Add(shortID, originalURL)
			h.logger.Info("add to repository", zap.String("shortID", shortID), zap.String("originalURL", originalURL))
		}

		c.Header("Content-Type", "text/plain")
		c.String(http.StatusCreated, "%s/%s", h.baseURL, shortID)
	}
}

type ShortenRequest struct {
	URL string `json:"url"`
}

type ShortenResponse struct {
	Result string `json:"result"`
}

// APIShortenHandler handles the creation of short URLs
type APIShortenHandler struct {
	repository URLRepository
	baseURL    string
	logger     *zap.Logger
}

func NewAPIShortenHandler(repo URLRepository, baseURL string, logger *zap.Logger) *APIShortenHandler {
	logger = logger.With(zap.String("handler", "APIShortenHandler"))
	return &APIShortenHandler{
		repository: repo,
		baseURL:    baseURL,
		logger:     logger,
	}
}

func (h *APIShortenHandler) Pattern() string {
	return "/api/shorten"
}

func (h *APIShortenHandler) Method() string {
	return http.MethodPost
}

func (h *APIShortenHandler) Handler() gin.HandlerFunc {
	return func(c *gin.Context) {
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

		// before adding to repository, check if the original URL already exists
		shortID, ok := h.repository.GetByOriginalURL(request.URL)
		h.logger.Info("get already existing original URL", zap.String("originalURL", request.URL), zap.Bool("exists", ok))
		if !ok {
			shortID = helpers.GenerateShortID()
			h.repository.Add(shortID, request.URL)
			h.logger.Info("add to repository", zap.String("shortID", shortID), zap.String("originalURL", request.URL))
		}
		c.JSON(http.StatusCreated, ShortenResponse{Result: fmt.Sprintf("%s/%s", h.baseURL, shortID)})
	}
}
