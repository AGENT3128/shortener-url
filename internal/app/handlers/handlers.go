package handlers

import (
	"io"
	"net/http"

	"github.com/AGENT3128/shortener-url/internal/app/helpers"
	"github.com/AGENT3128/shortener-url/internal/app/storage"
	"github.com/gin-gonic/gin"
)

type URLHandler struct {
	repository storage.Repository
}

func NewURLHandler(repo storage.Repository) *URLHandler {
	return &URLHandler{
		repository: repo,
	}
}

func (h *URLHandler) SetupRoutes(r *gin.Engine) {
	r.POST("/", h.handlePost)
	r.GET("/:id", h.handleGet)
}

func (h *URLHandler) handlePost(c *gin.Context) {
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

	shortID, ok := h.repository.GetByOriginalURL(originalURL)
	if !ok {
		shortID = helpers.GenerateShortID()
		h.repository.Add(shortID, originalURL)
	}

	c.Header("Content-Type", "text/plain")
	c.Status(http.StatusCreated)
	c.String(http.StatusCreated, "http://localhost:8080/%s", shortID)
}

func (h *URLHandler) handleGet(c *gin.Context) {
	shortID := c.Param("id")
	originalURL, ok := h.repository.GetByShortID(shortID)
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"error": "Short URL not found"})
		return
	}
	c.Header("Content-Type", "text/plain")
	c.Header("Location", originalURL)
	c.Status(http.StatusTemporaryRedirect)
}
