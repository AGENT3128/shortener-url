package handlers

import (
	"errors"
	"fmt"
	"io"
	"net/http"

	"go.uber.org/zap"

	"github.com/AGENT3128/shortener-url/internal/controller/httpapi/middleware"
	"github.com/AGENT3128/shortener-url/internal/entity"
)

type shortenOptions struct {
	baseURL string
	logger  *zap.Logger
	usecase URLSaver
}

type ShortenOption func(options *shortenOptions) error

func WithShortenBaseURL(baseURL string) ShortenOption {
	return func(options *shortenOptions) error {
		options.baseURL = baseURL
		return nil
	}
}

func WithShortenLogger(logger *zap.Logger) ShortenOption {
	return func(options *shortenOptions) error {
		options.logger = logger
		return nil
	}
}

func WithShortenUsecase(usecase URLSaver) ShortenOption {
	return func(options *shortenOptions) error {
		options.usecase = usecase
		return nil
	}
}

type ShortenHandler struct {
	baseURL string
	logger  *zap.Logger
	usecase URLSaver
}

// NewShortenHandler creates a new instance of ShortenHandler.
func NewShortenHandler(opts ...ShortenOption) (*ShortenHandler, error) {
	options := &shortenOptions{}
	for _, opt := range opts {
		if err := opt(options); err != nil {
			return nil, err
		}
	}

	if options.baseURL == "" {
		return nil, errors.New("baseURL is required")
	}
	if options.logger == nil {
		return nil, errors.New("logger is required")
	}
	if options.usecase == nil {
		return nil, errors.New("repository is required")
	}

	return &ShortenHandler{
		baseURL: options.baseURL,
		logger:  options.logger,
		usecase: options.usecase,
	}, nil
}

// Pattern returns the URL pattern for the handler.
func (h *ShortenHandler) Pattern() string {
	return "/"
}

// Method returns the HTTP method for the handler.
func (h *ShortenHandler) Method() string {
	return http.MethodPost
}

func (h *ShortenHandler) HandlerFunc() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, ok := r.Context().Value(middleware.UserIDKey).(string)
		if !ok {
			h.logger.Error("userID not found in context")
			JSONResponse(w, http.StatusUnauthorized, "Unauthorized")
			return
		}

		body, err := io.ReadAll(r.Body)
		if err != nil {
			h.logger.Error("failed to read request body", zap.Error(err))
			JSONResponse(w, http.StatusBadRequest, "failed to read request body")
			return
		}
		defer r.Body.Close()

		originalURL := string(body)
		if originalURL == "" {
			h.logger.Error("original URL is empty")
			JSONResponse(w, http.StatusBadRequest, "original URL is empty")
			return
		}

		shortURL, err := h.usecase.Add(r.Context(), userID, originalURL)
		if err != nil {
			h.handleError(w, err, shortURL)
			return
		}
		TextResponse(w, http.StatusCreated, fmt.Sprintf("%s/%s", h.baseURL, shortURL))
	}
}

func (h *ShortenHandler) handleError(w http.ResponseWriter, err error, shortURL string) {
	if errors.Is(err, entity.ErrURLExists) {
		TextResponse(w, http.StatusConflict, fmt.Sprintf("%s/%s", h.baseURL, shortURL))
		return
	}
	JSONResponse(w, http.StatusInternalServerError, "failed to add URL")
}
