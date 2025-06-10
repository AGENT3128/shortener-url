package middleware

import (
	"errors"
	"net/http"
	"time"

	"go.uber.org/zap"
)

type optionsMiddlewareLogger struct {
	logger *zap.Logger
}

// OptionMiddlewareLogger is the option for the middleware logger.
type OptionMiddlewareLogger func(options *optionsMiddlewareLogger) error

// Logger is the logger for the middleware.
type Logger struct {
	logger *zap.Logger
}

// WithMiddlewareLogger is the option for the middleware logger.
func WithMiddlewareLogger(logger *zap.Logger) OptionMiddlewareLogger {
	return func(options *optionsMiddlewareLogger) error {
		options.logger = logger.With(zap.String("middleware", "logger"))
		return nil
	}
}

// NewHandlerLogger creates a new handler logger.
func NewHandlerLogger(opts ...OptionMiddlewareLogger) (*Logger, error) {
	options := &optionsMiddlewareLogger{}
	for _, opt := range opts {
		if err := opt(options); err != nil {
			return nil, err
		}
	}
	if options.logger == nil {
		return nil, errors.New("logger is required")
	}
	return &Logger{logger: options.logger}, nil
}

// Handler is the handler for the logger.
func (h *Logger) Handler() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			// Create a response writer wrapper to capture status and size
			ww := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}

			next.ServeHTTP(ww, r)

			duration := time.Since(start)
			h.logger.Info("statistics",
				zap.Dict("request",
					zap.String("uri", r.URL.Path),
					zap.String("method", r.Method),
					zap.Duration("duration", duration),
				),
				zap.Dict("response",
					zap.Int("status", ww.statusCode),
					zap.Int("size", ww.size),
				),
			)
		})
	}
}

type responseWriter struct {
	http.ResponseWriter
	statusCode int
	size       int
}

// WriteHeader writes the header to the client.
func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

// Write writes the response to the client.
func (rw *responseWriter) Write(b []byte) (int, error) {
	size, err := rw.ResponseWriter.Write(b)
	rw.size += size
	return size, err
}
