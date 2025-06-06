package handlers

import (
	"errors"
	"net/http"

	"go.uber.org/zap"

	"github.com/AGENT3128/shortener-url/internal/controller/httpapi/middleware"
)

type pingOptions struct {
	usecase Pinger
	logger  *zap.Logger
}

// PingOption is the option for the ping handler.
type PingOption func(options *pingOptions) error

// PingHandler is the handler for the ping.
type PingHandler struct {
	usecase Pinger
	logger  *zap.Logger
}

// WithPingUsecase is the option for the ping handler to set the usecase.
func WithPingUsecase(usecase Pinger) PingOption {
	return func(options *pingOptions) error {
		options.usecase = usecase
		return nil
	}
}

// WithPingLogger is the option for the ping handler to set the logger.
func WithPingLogger(logger *zap.Logger) PingOption {
	return func(options *pingOptions) error {
		options.logger = logger.With(zap.String("handler", "PingHandler"))
		return nil
	}
}

// NewPingHandler creates a new ping handler.
func NewPingHandler(opts ...PingOption) (*PingHandler, error) {
	options := &pingOptions{}
	for _, opt := range opts {
		if err := opt(options); err != nil {
			return nil, err
		}
	}
	if options.usecase == nil {
		return nil, errors.New("usecase is required")
	}
	if options.logger == nil {
		return nil, errors.New("logger is required")
	}
	return &PingHandler{
		usecase: options.usecase,
		logger:  options.logger,
	}, nil
}

// Pattern is the pattern for the ping.
func (h *PingHandler) Pattern() string {
	return "/ping"
}

// Method is the method for the ping.
func (h *PingHandler) Method() string {
	return http.MethodGet
}

// HandlerFunc is the handler func for the ping.
func (h *PingHandler) HandlerFunc() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		_, ok := r.Context().Value(middleware.UserIDKey).(string)
		if !ok {
			h.logger.Error("userID not found in context")
			JSONResponse(w, http.StatusUnauthorized, "Unauthorized")
			return
		}
		if err := h.usecase.Ping(r.Context()); err != nil {
			h.logger.Error("Failed to ping database", zap.Error(err))
			JSONResponse(w, http.StatusInternalServerError, "Failed to ping database")
			return
		}
		JSONResponse(w, http.StatusOK, "Database is alive")
	}
}
