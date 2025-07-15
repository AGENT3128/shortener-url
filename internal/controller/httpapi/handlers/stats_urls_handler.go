package handlers

import (
	"encoding/json"
	"errors"
	"net/http"

	"go.uber.org/zap"

	"github.com/AGENT3128/shortener-url/internal/dto"
)

type statsURLsOptions struct {
	usecase       StatsGetter
	logger        *zap.Logger
	trustedSubnet string
}

// StatsURLsOption is the option for the stats URLs handler.
type StatsURLsOption func(options *statsURLsOptions) error

// StatsURLsHandler is the handler for the stats URLs.
type StatsURLsHandler struct {
	usecase       StatsGetter
	logger        *zap.Logger
	trustedSubnet string
}

// WithStatsURLsTrustedSubnet is the option for the stats URLs handler to set the trusted subnet.
func WithStatsURLsTrustedSubnet(trustedSubnet string) StatsURLsOption {
	return func(options *statsURLsOptions) error {
		options.trustedSubnet = trustedSubnet
		return nil
	}
}

// WithStatsURLsUsecase is the option for the stats URLs handler to set the usecase.
func WithStatsURLsUsecase(usecase StatsGetter) StatsURLsOption {
	return func(options *statsURLsOptions) error {
		options.usecase = usecase
		return nil
	}
}

// WithStatsURLsLogger is the option for the stats URLs handler to set the logger.
func WithStatsURLsLogger(logger *zap.Logger) StatsURLsOption {
	return func(options *statsURLsOptions) error {
		options.logger = logger.With(zap.String("handler", "StatsURLsHandler"))
		return nil
	}
}

// NewStatsURLsHandler creates a new stats URLs handler.
func NewStatsURLsHandler(opts ...StatsURLsOption) (*StatsURLsHandler, error) {
	options := &statsURLsOptions{}
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
	return &StatsURLsHandler{
		usecase:       options.usecase,
		logger:        options.logger,
		trustedSubnet: options.trustedSubnet,
	}, nil
}

// Pattern is the pattern for the stats URLs.
func (h *StatsURLsHandler) Pattern() string {
	return "/api/internal/stats"
}

// Method is the method for the stats URLs.
func (h *StatsURLsHandler) Method() string {
	return http.MethodGet
}

// HandlerFunc is the handler func for the stats URLs.
func (h *StatsURLsHandler) HandlerFunc() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		clientIP := getClientIP(r)
		// Check trusted subnet
		isTrusted, err := isIPInTrustedSubnet(r, h.trustedSubnet)
		if err != nil {
			h.logger.Error("failed to check trusted subnet", zap.Error(err))
			JSONResponse(w, http.StatusInternalServerError, "Internal Server Error")
			return
		}
		if !isTrusted {
			h.logger.Warn("access denied: IP not in trusted subnet",
				zap.String("ip", clientIP),
				zap.String("trusted_subnet", h.trustedSubnet))
			JSONResponse(w, http.StatusForbidden, "Forbidden")
			return
		}

		// Get stats from usecase
		urlsCount, usersCount, err := h.usecase.GetStats(ctx)
		if err != nil {
			h.logger.Error("failed to get stats", zap.Error(err))
			JSONResponse(w, http.StatusInternalServerError, "Internal Server Error")
			return
		}

		response := dto.StatsResponse{
			URLs:  urlsCount,
			Users: usersCount,
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		if errEncode := json.NewEncoder(w).Encode(response); errEncode != nil {
			http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		}

		h.logger.Info("stats retrieved successfully",
			zap.Int("urls_count", urlsCount),
			zap.Int("users_count", usersCount),
			zap.String("client_ip", clientIP))
	}
}
