package handlers

import (
	"encoding/json"
	"errors"
	"net/http"

	"go.uber.org/zap"

	"github.com/AGENT3128/shortener-url/internal/controller/httpapi/middleware"
	"github.com/AGENT3128/shortener-url/internal/dto"
	"github.com/AGENT3128/shortener-url/internal/entity"
)

type userURLsOptions struct {
	usecase UserURLGetter
	logger  *zap.Logger
	baseURL string
}

type UserURLsOption func(options *userURLsOptions) error

type UserURLsHandler struct {
	usecase UserURLGetter
	logger  *zap.Logger
	baseURL string
}

func WithUserURLsBaseURL(baseURL string) UserURLsOption {
	return func(options *userURLsOptions) error {
		options.baseURL = baseURL
		return nil
	}
}

func WithUserURLsUsecase(usecase UserURLGetter) UserURLsOption {
	return func(options *userURLsOptions) error {
		options.usecase = usecase
		return nil
	}
}

func WithUserURLsLogger(logger *zap.Logger) UserURLsOption {
	return func(options *userURLsOptions) error {
		options.logger = logger.With(zap.String("handler", "UserURLsHandler"))
		return nil
	}
}

func NewUserURLsHandler(opts ...UserURLsOption) (*UserURLsHandler, error) {
	options := &userURLsOptions{}
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
	return &UserURLsHandler{
		usecase: options.usecase,
		logger:  options.logger,
		baseURL: options.baseURL,
	}, nil
}

func (h *UserURLsHandler) Pattern() string {
	return "/api/user/urls"
}

func (h *UserURLsHandler) Method() string {
	return http.MethodGet
}

func (h *UserURLsHandler) HandlerFunc() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, ok := r.Context().Value(middleware.UserIDKey).(string)
		if !ok {
			h.logger.Error("userID not found in context")
			JSONResponse(w, http.StatusUnauthorized, "Unauthorized")
			return
		}

		urls, err := h.usecase.GetUserURLs(r.Context(), userID)
		h.logger.Info("urls", zap.Any("urls", urls), zap.String("userID", userID))
		if err != nil {
			h.logger.Error("failed to get user URLs", zap.Error(err))
			JSONResponse(w, http.StatusInternalServerError, "Failed to get user URLs")
			return
		}
		if len(urls) == 0 {
			JSONResponse(w, http.StatusNoContent, "No URLs found")
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		if errEncode := json.NewEncoder(w).Encode(h.toResponse(urls)); errEncode != nil {
			http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		}
	}
}

func (h *UserURLsHandler) toResponse(urls []entity.URL) []dto.UserURLsResponse {
	response := make([]dto.UserURLsResponse, 0, len(urls))
	for _, url := range urls {
		response = append(response, dto.UserURLsResponse{
			ShortURL:    h.baseURL + "/" + url.ShortURL,
			OriginalURL: url.OriginalURL,
		})
	}
	return response
}
