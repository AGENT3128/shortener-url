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
	"github.com/AGENT3128/shortener-url/internal/entity"
)

func TestShortenHandler(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	usecase := mocks.NewMockURLSaver(ctrl)
	logger, err := zap.NewDevelopment()
	require.NoError(t, err)

	handler, err := handlers.NewShortenHandler(
		handlers.WithShortenUsecase(usecase),
		handlers.WithShortenLogger(logger),
		handlers.WithShortenBaseURL("http://localhost:8080"),
	)
	require.NoError(t, err)

	require.Equal(t, "/", handler.Pattern())
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
		response    any
		contentType string
		statusCode  int
	}
	tests := []struct {
		setup   func()
		request request
		name    string
		want    want
	}{
		{
			name: "success save url",
			request: request{
				path:   "/",
				method: http.MethodPost,
				body:   "https://example.com",
			},
			want: want{
				statusCode:  http.StatusCreated,
				contentType: "text/plain",
				response:    "http://localhost:8080/shortURL123",
			},
			setup: func() {
				usecase.EXPECT().Add(gomock.Any(), gomock.Any(), "https://example.com").Return("shortURL123", nil)
			},
		},
		{
			name: "url already exists",
			request: request{
				path:   "/",
				method: http.MethodPost,
				body:   "https://example.com",
			},
			want: want{
				statusCode:  http.StatusConflict,
				contentType: "text/plain",
				response:    "http://localhost:8080/shortURL123",
			},
			setup: func() {
				usecase.EXPECT().
					Add(gomock.Any(), gomock.Any(), "https://example.com").
					Return("shortURL123", entity.ErrURLExists)
			},
		},
		{
			name: "internal server error",
			request: request{
				path:   "/",
				method: http.MethodPost,
				body:   "https://example.com",
			},
			want: want{
				statusCode:  http.StatusInternalServerError,
				contentType: "application/json",
				response: handlers.Response{
					Status:  http.StatusInternalServerError,
					Message: "Internal Server Error",
					Data:    "failed to add URL",
				},
			},
			setup: func() {
				usecase.EXPECT().
					Add(gomock.Any(), gomock.Any(), "https://example.com").
					Return("", errors.New("failed to add URL"))
			},
		},
		{
			name: "empty body",
			request: request{
				path:   "/",
				method: http.MethodPost,
				body:   "",
			},
			want: want{
				statusCode:  http.StatusBadRequest,
				contentType: "application/json",
				response: handlers.Response{
					Status:  http.StatusBadRequest,
					Message: "Bad Request",
					Data:    "original URL is empty",
				},
			},
			setup: func() {

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
			recorder := httptest.NewRecorder()
			router.ServeHTTP(recorder, req)

			require.Equal(t, test.want.statusCode, recorder.Code)
			require.Equal(t, test.want.contentType, recorder.Header().Get("Content-Type"))
			switch test.want.response.(type) {
			case handlers.Response:
				var response handlers.Response
				err = json.NewDecoder(recorder.Body).Decode(&response)
				require.NoError(t, err)
				require.Equal(t, test.want.response, response)
			default:
				require.Equal(t, test.want.response, recorder.Body.String())
			}
		})
	}
}
