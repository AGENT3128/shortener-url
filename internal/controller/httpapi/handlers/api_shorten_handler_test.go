package handlers

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

	"github.com/AGENT3128/shortener-url/internal/controller/httpapi/handlers/mocks"
	customMiddleware "github.com/AGENT3128/shortener-url/internal/controller/httpapi/middleware"
	"github.com/AGENT3128/shortener-url/internal/dto"
	"github.com/AGENT3128/shortener-url/internal/entity"
)

func TestAPIShortenHandler(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	urlUsecaseMock := mocks.NewMockURLSaver(ctrl)
	logger, err := zap.NewDevelopment()
	require.NoError(t, err)

	handler, err := NewAPIShortenHandler(
		WithAPIShortenUsecase(urlUsecaseMock),
		WithAPIShortenLogger(logger),
		WithAPIShortenBaseURL("http://localhost:8080"),
	)
	require.NoError(t, err)
	require.Equal(t, "/api/shorten", handler.Pattern())
	require.Equal(t, http.MethodPost, handler.Method())

	authMiddleware, err := customMiddleware.NewAuthMiddleware(
		customMiddleware.WithAuthMiddlewareLogger(logger),
	)
	require.NoError(t, err)

	router := chi.NewRouter()
	router.Use(authMiddleware.Handler())
	router.Method(handler.Method(), handler.Pattern(), handler.HandlerFunc())

	type request struct {
		body   string
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
				body:   `{"url": "https://example.com"}`,
				path:   "/api/shorten",
				method: http.MethodPost,
			},
			want: want{
				statusCode:  http.StatusCreated,
				contentType: "application/json",
				response:    dto.ShortenResponse{Result: "http://localhost:8080/exampleShortURL"},
			},
			setup: func() {
				urlUsecaseMock.EXPECT().
					Add(gomock.Any(), gomock.Any(), gomock.Any()).
					Return("exampleShortURL", nil)
			},
		},
		{
			name: "empty url in request",
			request: request{
				body:   `{"url": ""}`,
				path:   "/api/shorten",
				method: http.MethodPost,
			},
			want: want{
				statusCode:  http.StatusBadRequest,
				contentType: "application/json",
				response:    Response{Status: http.StatusBadRequest, Message: "URL is required", Data: nil},
			},
			setup: func() {},
		},
		{
			name: "invalid json in request",
			request: request{
				body:   `{"url": invalid_json}`,
				path:   "/api/shorten",
				method: http.MethodPost,
			},
			want: want{
				statusCode:  http.StatusBadRequest,
				contentType: "application/json",
				response: Response{
					Status:  http.StatusBadRequest,
					Message: "Failed to unmarshal request body",
					Data:    nil,
				},
			},
			setup: func() {},
		},
		{
			name: "url already exists",
			request: request{
				body:   `{"url": "https://example.com"}`,
				path:   "/api/shorten",
				method: http.MethodPost,
			},
			want: want{
				statusCode:  http.StatusConflict,
				contentType: "application/json",
				response:    Response{Status: http.StatusConflict, Message: "URL already exists", Data: nil},
			},
			setup: func() {
				urlUsecaseMock.EXPECT().
					Add(gomock.Any(), gomock.Any(), gomock.Any()).
					Return("", entity.ErrURLExists)
			},
		},
		{
			name: "internal server error",
			request: request{
				body:   `{"url": "https://example.com"}`,
				path:   "/api/shorten",
				method: http.MethodPost,
			},
			want: want{
				statusCode:  http.StatusInternalServerError,
				contentType: "application/json",
				response: Response{
					Status:  http.StatusInternalServerError,
					Message: "Failed to shorten URL",
					Data:    nil,
				},
			},
			setup: func() {
				urlUsecaseMock.EXPECT().
					Add(gomock.Any(), gomock.Any(), gomock.Any()).
					Return("", errors.New("internal error"))
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			test.setup()
			req, errRequest := http.NewRequest(
				test.request.method,
				test.request.path,
				strings.NewReader(test.request.body),
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
			case dto.ShortenResponse:
				require.Equal(t, test.want.response, response)
			case Response:
				require.Equal(t, test.want.response, response)
			default:
			}
		})
	}
}
