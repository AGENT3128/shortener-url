package handlers

import (
	"net/http"

	"github.com/AGENT3128/shortener-url/internal/app/storage"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// PingHandler handles the ping database connection request
type PingHandler struct {
	db     *storage.Database
	logger *zap.Logger
}

func NewPingHandler(db *storage.Database, logger *zap.Logger) *PingHandler {
	logger = logger.With(zap.String("handler", "PingHandler"))
	return &PingHandler{
		db:     db,
		logger: logger,
	}
}

func (h *PingHandler) Pattern() string {
	return "/ping"
}

func (h *PingHandler) Method() string {
	return http.MethodGet
}

func (h *PingHandler) Handler() gin.HandlerFunc {
	return func(c *gin.Context) {
		if err := h.db.Ping(c.Request.Context()); err != nil {
			h.logger.Error("Failed to ping database", zap.Error(err))
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to ping database"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "Database is alive"})
	}
}
