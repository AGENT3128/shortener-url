package handlers

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// URLShortGetter describes the behavior for retrieving original URL by short ID
type URLShortGetter interface {
	GetByShortID(shortID string) (string, bool)
}

// RedirectHandler handles redirects by short URL
type RedirectHandler struct {
	repository URLShortGetter
	logger     *zap.Logger
}

func NewRedirectHandler(repo URLShortGetter, logger *zap.Logger) *RedirectHandler {
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
		shortID := strings.TrimPrefix(c.Request.URL.Path, "/")
		originalURL, ok := h.repository.GetByShortID(shortID)
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
