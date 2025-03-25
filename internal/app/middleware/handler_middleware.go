package middleware

import (
	"time"

	"github.com/AGENT3128/shortener-url/internal/app/logger"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func HandlerLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()
		duration := time.Since(start)
		logger.Log.Info("Handler",
			zap.Dict("request",
				zap.String("uri", c.Request.URL.Path),
				zap.String("method", c.Request.Method),
				zap.Duration("duration", duration),
			),
			zap.Dict("response",
				zap.Int("status", c.Writer.Status()),
				zap.Int("size", c.Writer.Size()),
			),
		)
	}
}
