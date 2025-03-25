package handlers

import (
	"io"
	"net/http"

	"github.com/AGENT3128/shortener-url/internal/app/helpers"
	"github.com/gin-gonic/gin"
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
}

func NewShortenHandler(repo URLRepository, baseURL string) *ShortenHandler {
	return &ShortenHandler{
		repository: repo,
		baseURL:    baseURL,
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
		if !ok {
			shortID = helpers.GenerateShortID()
			h.repository.Add(shortID, originalURL)
		}

		c.Header("Content-Type", "text/plain")
		c.String(http.StatusCreated, "%s/%s", h.baseURL, shortID)
	}
}
