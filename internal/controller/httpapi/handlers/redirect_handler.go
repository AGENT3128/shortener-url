package handlers

import (
	"errors"
	"net/http"
	"strings"

	"go.uber.org/zap"

	"github.com/AGENT3128/shortener-url/internal/controller/httpapi/middleware"
	"github.com/AGENT3128/shortener-url/internal/entity"
)

// RedirectHandler is the handler for the redirect.
type RedirectHandler struct {
	usecase URLGetter
	logger  *zap.Logger
}

type redirectOptions struct {
	usecase URLGetter
	logger  *zap.Logger
}

// RedirectOption is the option for the redirect handler.
type RedirectOption func(options *redirectOptions) error

// WithRedirectUsecase is the option for the redirect handler to set the usecase.
func WithRedirectUsecase(usecase URLGetter) RedirectOption {
	return func(options *redirectOptions) error {
		options.usecase = usecase
		return nil
	}
}

// WithRedirectLogger is the option for the redirect handler to set the logger.
func WithRedirectLogger(logger *zap.Logger) RedirectOption {
	return func(options *redirectOptions) error {
		options.logger = logger.With(zap.String("handler", "RedirectHandler"))
		return nil
	}
}

// NewRedirectHandler creates a new redirect handler.
func NewRedirectHandler(opts ...RedirectOption) (*RedirectHandler, error) {
	options := &redirectOptions{}
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
	return &RedirectHandler{usecase: options.usecase, logger: options.logger}, nil
}

// Pattern is the pattern for the redirect.
func (h *RedirectHandler) Pattern() string {
	return "/{id}"
}

// Method is the method for the redirect.
func (h *RedirectHandler) Method() string {
	return http.MethodGet
}

// HandlerFunc is the handler func for the redirect.
func (h *RedirectHandler) HandlerFunc() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		_, ok := r.Context().Value(middleware.UserIDKey).(string)
		if !ok {
			h.logger.Error("userID not found in context")
			JSONResponse(w, http.StatusUnauthorized, "Unauthorized")
			return
		}

		shortURL := strings.TrimPrefix(r.URL.Path, "/")
		if shortURL == "" {
			h.logger.Error("shortURL is empty")
			JSONResponse(w, http.StatusBadRequest, "shortURL is empty")
			return
		}

		h.logger.Info("searching for short URL", zap.String("short_url", shortURL))
		originalURL, err := h.usecase.GetByShortURL(r.Context(), shortURL)
		if err != nil {
			h.handleError(w, err)
			return
		}

		w.Header().Set("Location", originalURL)
		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusTemporaryRedirect)
	}
}

func (h *RedirectHandler) handleError(w http.ResponseWriter, err error) {
	h.logger.Error("failed to get original URL", zap.Error(err))
	if errors.Is(err, entity.ErrURLDeleted) {
		JSONResponse(w, http.StatusGone, "URL has been deleted")
		return
	}
	JSONResponse(w, http.StatusNotFound, "URL not found")
}
