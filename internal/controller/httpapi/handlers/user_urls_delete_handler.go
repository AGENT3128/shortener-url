package handlers

import (
	"encoding/json"
	"errors"
	"net/http"

	"go.uber.org/zap"

	"github.com/AGENT3128/shortener-url/internal/controller/httpapi/middleware"
)

type userURLsDeleteOptions struct {
	usecase UserURLDeleter
	logger  *zap.Logger
}

// UserURLsDeleteOption is the option for the user URLs delete handler.
type UserURLsDeleteOption func(options *userURLsDeleteOptions) error

// UserURLsDeleteHandler is the handler for the user URLs delete.
type UserURLsDeleteHandler struct {
	usecase UserURLDeleter
	logger  *zap.Logger
}

// WithUserURLsDeleteUsecase is the option for the user URLs delete handler to set the usecase.
func WithUserURLsDeleteUsecase(usecase UserURLDeleter) UserURLsDeleteOption {
	return func(options *userURLsDeleteOptions) error {
		options.usecase = usecase
		return nil
	}
}

// WithUserURLsDeleteLogger is the option for the user URLs delete handler to set the logger.
func WithUserURLsDeleteLogger(logger *zap.Logger) UserURLsDeleteOption {
	return func(options *userURLsDeleteOptions) error {
		options.logger = logger.With(zap.String("handler", "UserURLsDeleteHandler"))
		return nil
	}
}

// NewUserURLsDeleteHandler creates a new user URLs delete handler.
func NewUserURLsDeleteHandler(opts ...UserURLsDeleteOption) (*UserURLsDeleteHandler, error) {
	options := &userURLsDeleteOptions{}
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
	return &UserURLsDeleteHandler{
		usecase: options.usecase,
		logger:  options.logger,
	}, nil
}

// Pattern is the pattern for the user URLs delete.
func (h *UserURLsDeleteHandler) Pattern() string {
	return "/api/user/urls"
}

// Method is the method for the user URLs delete.
func (h *UserURLsDeleteHandler) Method() string {
	return http.MethodDelete
}

// HandlerFunc is the handler func for the user URLs delete.
func (h *UserURLsDeleteHandler) HandlerFunc() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, ok := r.Context().Value(middleware.UserIDKey).(string)
		if !ok {
			h.logger.Error("userID not found in context")
			JSONResponse(w, http.StatusUnauthorized, "Unauthorized")
			return
		}
		var shortURLs []string
		if err := json.NewDecoder(r.Body).Decode(&shortURLs); err != nil {
			h.logger.Error("failed to bind request body", zap.Error(err))
			JSONResponse(w, http.StatusBadRequest, "invalid request format")
			return
		}

		if len(shortURLs) == 0 {
			JSONResponse(w, http.StatusBadRequest, "no URLs provided for deletion")
			return
		}

		if err := h.usecase.DeleteUserURLs(r.Context(), userID, shortURLs); err != nil {
			h.logger.Error("failed to delete user URLs", zap.Error(err))
			JSONResponse(w, http.StatusInternalServerError, "failed to delete URLs")
			return
		}

		JSONResponse(w, http.StatusAccepted, "success")
	}
}
