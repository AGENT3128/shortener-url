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

type pingOption func(options *pingOptions) error

type PingHandler struct {
	usecase Pinger
	logger  *zap.Logger
}

func WithPingUsecase(usecase Pinger) pingOption {
	return func(options *pingOptions) error {
		options.usecase = usecase
		return nil
	}
}

func WithPingLogger(logger *zap.Logger) pingOption {
	return func(options *pingOptions) error {
		options.logger = logger.With(zap.String("handler", "PingHandler"))
		return nil
	}
}

func NewPingHandler(opts ...pingOption) (*PingHandler, error) {
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

func (h *PingHandler) Pattern() string {
	return "/ping"
}

func (h *PingHandler) Method() string {
	return http.MethodGet
}

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
