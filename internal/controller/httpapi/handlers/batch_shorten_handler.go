package handlers

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"

	"go.uber.org/zap"

	"github.com/AGENT3128/shortener-url/internal/controller/httpapi/middleware"
	"github.com/AGENT3128/shortener-url/internal/dto"
	"github.com/AGENT3128/shortener-url/internal/entity"
)

type batchShortenOptions struct {
	usecase BatchURLSaver
	logger  *zap.Logger
	baseURL string
}

type batchShortenOption func(options *batchShortenOptions) error

type BatchShortenHandler struct {
	usecase BatchURLSaver
	logger  *zap.Logger
	baseURL string
}

func WithBatchShortenBaseURL(baseURL string) batchShortenOption {
	return func(options *batchShortenOptions) error {
		options.baseURL = baseURL
		return nil
	}
}

func WithBatchShortenUsecase(usecase BatchURLSaver) batchShortenOption {
	return func(options *batchShortenOptions) error {
		options.usecase = usecase
		return nil
	}
}

func WithBatchShortenLogger(logger *zap.Logger) batchShortenOption {
	return func(options *batchShortenOptions) error {
		options.logger = logger.With(zap.String("handler", "BatchShortenHandler"))
		return nil
	}
}

func NewBatchShortenHandler(opts ...batchShortenOption) (*BatchShortenHandler, error) {
	options := &batchShortenOptions{}
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
	return &BatchShortenHandler{
		usecase: options.usecase,
		logger:  options.logger,
		baseURL: options.baseURL,
	}, nil
}

func (h *BatchShortenHandler) Pattern() string {
	return "/api/shorten/batch"
}

func (h *BatchShortenHandler) Method() string {
	return http.MethodPost
}

func (h *BatchShortenHandler) HandlerFunc() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, ok := r.Context().Value(middleware.UserIDKey).(string)
		if !ok {
			h.logger.Error("userID not found in context")
			JSONResponse(w, http.StatusUnauthorized, "Unauthorized")
			return
		}

		body, err := io.ReadAll(r.Body)
		if err != nil {
			h.logger.Error("Failed to read request body", zap.Error(err))
			JSONResponse(w, http.StatusBadRequest, "Failed to read request body")
			return
		}
		defer r.Body.Close()

		var requests []dto.ShortenBatchRequest
		if err := json.Unmarshal(body, &requests); err != nil {
			h.logger.Error("Failed to unmarshal request body", zap.Error(err))
			JSONResponse(w, http.StatusBadRequest, "Failed to unmarshal request body")
			return
		}

		if len(requests) == 0 {
			h.logger.Error("Request body is empty")
			JSONResponse(w, http.StatusBadRequest, "Request body is empty")
			return
		}

		urls := make([]entity.URL, 0, len(requests))
		correlationMap := make(map[string]string)

		for _, req := range requests {
			if req.OriginalURL == "" || req.CorrelationID == "" {
				continue
			}
			url := entity.URL{
				OriginalURL: req.OriginalURL,
			}
			urls = append(urls, url)
			correlationMap[req.OriginalURL] = req.CorrelationID
		}

		h.logger.Info("urls to send to usecase", zap.Any("urls", urls))
		shortenedURLs, err := h.usecase.AddBatch(r.Context(), userID, urls)
		h.logger.Info("shortenedURLs", zap.Any("shortenedURLs", shortenedURLs), zap.Error(err))
		if err != nil {
			h.logger.Error("Failed to shorten URLs", zap.Error(err))
			JSONResponse(w, http.StatusInternalServerError, "Failed to shorten URLs")
			return
		}

		responses := h.toResponse(shortenedURLs, correlationMap)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(responses)
	}
}

func (h *BatchShortenHandler) toResponse(
	urls []entity.URL,
	correlationMap map[string]string,
) []dto.ShortenBatchResponse {
	responses := make([]dto.ShortenBatchResponse, 0, len(urls))
	for _, url := range urls {
		response := dto.ShortenBatchResponse{
			CorrelationID: correlationMap[url.OriginalURL],
			ShortURL:      h.baseURL + "/" + url.ShortURL,
		}
		responses = append(responses, response)
	}
	return responses
}
