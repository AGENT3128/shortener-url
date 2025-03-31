package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/AGENT3128/shortener-url/internal/app/helpers"
	"github.com/AGENT3128/shortener-url/internal/app/storage"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// URLOriginalGetter describes the behavior for retrieving short URL by original URL
type URLOriginalGetter interface {
	GetByOriginalURL(originalURL string) (string, bool)
}

// URLSaver describes the behavior for saving short URL by original URL
type URLSaver interface {
	Add(shortID, originalURL string) (string, error)
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
		// shortID, ok := h.repository.GetByOriginalURL(originalURL)
		// h.logger.Info("get original URL", zap.String("originalURL", originalURL), zap.Bool("exists", ok))
		// if !ok {
		// 	shortID = helpers.GenerateShortID()
		// 	h.repository.Add(shortID, originalURL)
		// 	h.logger.Info("add to repository", zap.String("shortID", shortID), zap.String("originalURL", originalURL))
		// }

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
		// shortID, ok := h.repository.GetByOriginalURL(request.URL)
		// h.logger.Info("get already existing original URL", zap.String("originalURL", request.URL), zap.Bool("exists", ok))
		// if !ok {
		// 	shortID = helpers.GenerateShortID()
		// 	h.repository.Add(shortID, request.URL)
		// 	h.logger.Info("add to repository", zap.String("shortID", shortID), zap.String("originalURL", request.URL))
		// }
		shortID, err := h.repository.Add(helpers.GenerateShortID(), request.URL)
		if err != nil {
			if errors.Is(err, storage.ErrURLExists) {
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

type TransactionSupport interface {
	// Begin(ctx context.Context) (pgx.Tx, error)
	// Commit(ctx context.Context) error
	// Rollback(ctx context.Context) error
	AddBatch(urls []storage.URL) error
}

type ShortenBatchRequest struct {
	CorrelationID string `json:"correlation_id"`
	OriginalURL   string `json:"original_url"`
}

type ShortenBatchResponse struct {
	CorrelationID string `json:"correlation_id"`
	ShortURL      string `json:"short_url"`
}

// ShortenBatchHandler handles the creation of short URLs
type ShortenBatchHandler struct {
	repository URLRepository
	baseURL    string
	logger     *zap.Logger
}

func NewShortenBatchHandler(repo URLRepository, baseURL string, logger *zap.Logger) *ShortenBatchHandler {
	logger = logger.With(zap.String("handler", "ShortenBatchHandler"))
	return &ShortenBatchHandler{
		repository: repo,
		baseURL:    baseURL,
		logger:     logger,
	}
}

func (h *ShortenBatchHandler) Pattern() string {
	return "/api/shorten/batch"
}

func (h *ShortenBatchHandler) Method() string {
	return http.MethodPost
}

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
