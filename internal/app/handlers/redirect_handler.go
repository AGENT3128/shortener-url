package handlers

import (
	"context"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// urlShortGetter describes the behavior for retrieving original URL by short ID
type urlShortGetter interface {
	GetByShortID(ctx context.Context, shortID string) (string, bool)
	IsURLDeleted(ctx context.Context, shortID string) (bool, error)
}

// RedirectHandler handles redirects by short URL
type RedirectHandler struct {
	repository urlShortGetter
	logger     *zap.Logger
}

func NewRedirectHandler(repo urlShortGetter, logger *zap.Logger) *RedirectHandler {
	logger = logger.With(zap.String("handler", "RedirectHandler"))
	return &RedirectHandler{
		repository: repo,
		logger:     logger,
	}
}

func (h *RedirectHandler) Pattern() string {
	return "/:id"
}

func (h *RedirectHandler) Method() string {
	return http.MethodGet
}

func (h *RedirectHandler) Handler() gin.HandlerFunc {
	return func(c *gin.Context) {
		_, ok := c.Get("userID")
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
			return
		}
		shortID := strings.TrimPrefix(c.Request.URL.Path, "/")

		// check if URL is marked as deleted
		isDeleted, err := h.repository.IsURLDeleted(c.Request.Context(), shortID)
		if err != nil {
			h.logger.Error("error checking URL deleted status", zap.Error(err))
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
			return
		}

		if isDeleted {
			h.logger.Info("URL is marked as deleted", zap.String("shortID", shortID))
			c.JSON(http.StatusGone, gin.H{"error": "URL has been deleted"})
			return
		}

		originalURL, ok := h.repository.GetByShortID(c.Request.Context(), shortID)
		h.logger.Info("get original URL", zap.String("shortID", shortID), zap.String("originalURL", originalURL), zap.Bool("exists", ok))
		if !ok {
			c.JSON(http.StatusNotFound, gin.H{"error": "Short URL not found"})
			return
		}

		c.Header("Location", originalURL)
		c.Header("Content-Type", "text/plain")
		c.Status(http.StatusTemporaryRedirect)
	}
}
