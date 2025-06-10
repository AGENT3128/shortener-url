package handlers_test

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"go.uber.org/zap"

	"github.com/AGENT3128/shortener-url/internal/controller/httpapi/handlers"
	"github.com/AGENT3128/shortener-url/internal/controller/httpapi/handlers/mocks"
	customMiddleware "github.com/AGENT3128/shortener-url/internal/controller/httpapi/middleware"
	"github.com/AGENT3128/shortener-url/internal/dto"
	"github.com/AGENT3128/shortener-url/internal/entity"
)

func TestBatchShortenHandler(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	batchURLSaverMock := mocks.NewMockBatchURLSaver(ctrl)
	logger, err := zap.NewDevelopment()
	require.NoError(t, err)

	handler, err := handlers.NewBatchShortenHandler(
		handlers.WithBatchShortenUsecase(batchURLSaverMock),
		handlers.WithBatchShortenLogger(logger),
		handlers.WithBatchShortenBaseURL("http://localhost:8080"),
	)
	require.NoError(t, err)
	require.Equal(t, "/api/shorten/batch", handler.Pattern())
	require.Equal(t, http.MethodPost, handler.Method())

	authMiddleware, err := customMiddleware.NewAuthMiddleware(
		customMiddleware.WithAuthMiddlewareLogger(logger),
	)
	require.NoError(t, err)

	router := chi.NewRouter()
	router.Use(authMiddleware.Handler())
	router.Method(handler.Method(), handler.Pattern(), handler.HandlerFunc())

	type request struct {
		body   any
		path   string
		method string
	}
	type want struct {
		statusCode  int
		contentType string
		response    any
	}
	tests := []struct {
		name    string
		request request
		want    want
		setup   func()
	}{
		{
			name: "success save url",
			request: request{
				body: []dto.ShortenBatchRequest{
					{CorrelationID: "1", OriginalURL: "https://example1.com"},
					{CorrelationID: "2", OriginalURL: "https://example2.com"},
				},
				path:   "/api/shorten/batch",
				method: http.MethodPost,
			},
			want: want{
				statusCode:  http.StatusCreated,
				contentType: "application/json",
				response: []dto.ShortenBatchResponse{
					{CorrelationID: "1", ShortURL: "http://localhost:8080/exampleShortURL1"},
					{CorrelationID: "2", ShortURL: "http://localhost:8080/exampleShortURL2"},
				},
			},
			setup: func() {
				batchURLSaverMock.EXPECT().
					AddBatch(gomock.Any(), gomock.Any(), gomock.Any()).
					Return([]entity.URL{
						{OriginalURL: "https://example1.com", ShortURL: "http://localhost:8080/exampleShortURL1"},
						{OriginalURL: "https://example2.com", ShortURL: "http://localhost:8080/exampleShortURL2"},
					}, nil)
			},
		},
		{
			name: "incorrect body",
			request: request{
				body:   `{"correlationID": "1", "originalURL": "https://example1.com"}`,
				path:   "/api/shorten/batch",
				method: http.MethodPost,
			},
			want: want{
				statusCode:  http.StatusBadRequest,
				contentType: "application/json",
				response: handlers.Response{
					Status:  http.StatusBadRequest,
					Message: "Failed to unmarshal request body",
					Data:    nil,
				},
			},
			setup: func() {},
		},
		{
			name: "empty body",
			request: request{
				body:   []dto.ShortenBatchRequest{},
				path:   "/api/shorten/batch",
				method: http.MethodPost,
			},
			want: want{
				statusCode:  http.StatusBadRequest,
				contentType: "application/json",
				response: handlers.Response{
					Status:  http.StatusBadRequest,
					Message: "Request body is empty",
					Data:    nil,
				},
			},
			setup: func() {},
		},
		{
			name: "internal server error",
			request: request{
				body: []dto.ShortenBatchRequest{
					{CorrelationID: "1", OriginalURL: "https://example1.com"},
					{CorrelationID: "2", OriginalURL: "https://example2.com"},
				},
				path:   "/api/shorten/batch",
				method: http.MethodPost,
			},
			want: want{
				statusCode:  http.StatusInternalServerError,
				contentType: "application/json",
				response: handlers.Response{
					Status:  http.StatusInternalServerError,
					Message: "Failed to shorten URL",
					Data:    nil,
				},
			},
			setup: func() {
				batchURLSaverMock.EXPECT().
					AddBatch(gomock.Any(), gomock.Any(), gomock.Any()).
					Return(nil, errors.New("internal server error"))
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			test.setup()
			requestBody, errMarshal := json.Marshal(test.request.body)
			require.NoError(t, errMarshal)
			req, errRequest := http.NewRequest(
				test.request.method,
				test.request.path,
				strings.NewReader(string(requestBody)),
			)

			require.NoError(t, errRequest)
			req.Header.Set("Content-Type", "application/json")

			rr := httptest.NewRecorder()
			router.ServeHTTP(rr, req)
			require.Equal(t, test.want.statusCode, rr.Code)
			require.Equal(t, test.want.contentType, rr.Header().Get("Content-Type"))

			var response any
			err = json.NewDecoder(rr.Body).Decode(&response)
			require.NoError(t, err)

			switch response := response.(type) {
			case []dto.ShortenBatchResponse:
				require.Equal(t, test.want.response, response)
			default:
			}
		})
	}
}
