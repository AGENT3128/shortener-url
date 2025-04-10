package handlers

import (
	"context"
	"net/http"

	"github.com/AGENT3128/shortener-url/internal/app/models"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// userURLsGetter describes the behavior for getting user's URLs
type userURLsGetter interface {
	GetUserURLs(ctx context.Context, userID string) ([]models.URL, error)
}

// UserURLsResponse represents individual URL in the response
type UserURLResponse struct {
	ShortURL    string `json:"short_url"`
	OriginalURL string `json:"original_url"`
}

// UserURLsHandler handles retrieving all URLs created by a user
type UserURLsHandler struct {
	repository userURLsGetter
	baseURL    string
	logger     *zap.Logger
}

// NewUserURLsHandler creates a new instance of UserURLsHandler
func NewUserURLsHandler(repo userURLsGetter, baseURL string, logger *zap.Logger) *UserURLsHandler {
	logger = logger.With(zap.String("handler", "UserURLsHandler"))
	return &UserURLsHandler{
		repository: repo,
		baseURL:    baseURL,
		logger:     logger,
	}
}

// Pattern returns the URL pattern for the handler
func (h *UserURLsHandler) Pattern() string {
	return "/api/user/urls"
}

// Method returns the HTTP method for the handler
func (h *UserURLsHandler) Method() string {
	return http.MethodGet
}

// Handler returns the gin.HandlerFunc for the handler
func (h *UserURLsHandler) Handler() gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, ok := c.Get("userID")
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
			return
		}

		urls, err := h.repository.GetUserURLs(c.Request.Context(), userID.(string))
		if err != nil {
			h.logger.Error("error getting user URLs", zap.Error(err))
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal Server Error"})
			return
		}

		if len(urls) == 0 {
			c.JSON(http.StatusNoContent, gin.H{"message": "No content"})
			return
		}

		var response []UserURLResponse
		for _, url := range urls {
			response = append(response, UserURLResponse{
				ShortURL:    h.baseURL + "/" + url.ShortID,
				OriginalURL: url.OriginalURL,
			})
		}

		c.JSON(http.StatusOK, response)
	}
}
