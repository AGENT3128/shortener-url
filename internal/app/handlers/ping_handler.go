package handlers

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// pingI describes the behavior for pinging a database
type pingI interface {
	Ping(ctx context.Context) error
}

// PingHandler handles the ping database connection request
type PingHandler struct {
	repository pingI
	logger     *zap.Logger
}

func NewPingHandler(repository pingI, logger *zap.Logger) *PingHandler {
	logger = logger.With(zap.String("handler", "PingHandler"))
	return &PingHandler{
		repository: repository,
		logger:     logger,
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
		if err := h.repository.Ping(c.Request.Context()); err != nil {
			h.logger.Error("Failed to ping database", zap.Error(err))
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to ping database"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "Database is alive"})
	}
}
