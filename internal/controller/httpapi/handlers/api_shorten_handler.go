package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"

	"go.uber.org/zap"

	"github.com/AGENT3128/shortener-url/internal/controller/httpapi/middleware"
	"github.com/AGENT3128/shortener-url/internal/dto"
	"github.com/AGENT3128/shortener-url/internal/entity"
)

type APIShortenHandler struct {
	usecase URLSaver
	logger  *zap.Logger
	baseURL string
}

type apiShortenOptions struct {
	usecase URLSaver
	logger  *zap.Logger
	baseURL string
}

type apiShortenOption func(options *apiShortenOptions) error

func WithAPIShortenUsecase(usecase URLSaver) apiShortenOption {
	return func(options *apiShortenOptions) error {
		options.usecase = usecase
		return nil
	}
}

func WithAPIShortenLogger(logger *zap.Logger) apiShortenOption {
	return func(options *apiShortenOptions) error {
		options.logger = logger.With(zap.String("handler", "APIShortenHandler"))
		return nil
	}
}

func WithAPIShortenBaseURL(baseURL string) apiShortenOption {
	return func(options *apiShortenOptions) error {
		options.baseURL = baseURL
		return nil
	}
}
func NewAPIShortenHandler(opts ...apiShortenOption) (*APIShortenHandler, error) {
	options := &apiShortenOptions{}
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
	return &APIShortenHandler{
		usecase: options.usecase,
		logger:  options.logger,
		baseURL: options.baseURL,
	}, nil
}

func (h *APIShortenHandler) Pattern() string {
	return "/api/shorten"
}

func (h *APIShortenHandler) Method() string {
	return http.MethodPost
}

func (h *APIShortenHandler) HandlerFunc() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, ok := r.Context().Value(middleware.UserIDKey).(string)
		if !ok {
			h.logger.Error("userID not found in context")
			JSONResponse(w, http.StatusUnauthorized, "Unauthorized")
			return
		}

		var request dto.ShortenRequest
		body, err := io.ReadAll(r.Body)
		if err != nil {
			JSONResponse(w, http.StatusBadRequest, "Failed to read request body")
			return
		}
		defer r.Body.Close()

		if err := json.Unmarshal(body, &request); err != nil {
			JSONResponse(w, http.StatusBadRequest, "Failed to unmarshal request body")
			return
		}

		if request.URL == "" {
			JSONResponse(w, http.StatusBadRequest, "URL is required")
			return
		}

		shortURL, err := h.usecase.Add(r.Context(), userID, request.URL)
		h.logger.Info("short URL", zap.String("short_url", shortURL))
		if err != nil {
			h.handleError(w, err)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		err = json.NewEncoder(w).Encode(dto.ShortenResponse{Result: fmt.Sprintf("%s/%s", h.baseURL, shortURL)})
		if err != nil {
			h.logger.Error("failed to encode response", zap.Error(err))
			JSONResponse(w, http.StatusInternalServerError, "Failed to encode response")
			return
		}
	}
}

func (h *APIShortenHandler) handleError(w http.ResponseWriter, err error) {
	if errors.Is(err, entity.ErrURLExists) {
		JSONResponse(w, http.StatusConflict, "URL already exists")
		return
	}
	h.logger.Error("failed to shorten URL", zap.Error(err))
	JSONResponse(w, http.StatusInternalServerError, "Failed to shorten URL")
}
