package handlers_test

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
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

func TestUserURLsHandler(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	usecase := mocks.NewMockUserURLGetter(ctrl)
	logger, err := zap.NewDevelopment()
	require.NoError(t, err)

	handler, err := handlers.NewUserURLsHandler(
		handlers.WithUserURLsUsecase(usecase),
		handlers.WithUserURLsLogger(logger),
		handlers.WithUserURLsBaseURL("http://localhost:8080"),
	)
	require.NoError(t, err)

	require.Equal(t, "/api/user/urls", handler.Pattern())
	require.Equal(t, http.MethodGet, handler.Method())

	authMiddleware, err := customMiddleware.NewAuthMiddleware(
		customMiddleware.WithAuthMiddlewareLogger(logger),
	)
	require.NoError(t, err)

	router := chi.NewRouter()
	router.Use(authMiddleware.Handler())
	router.Method(handler.Method(), handler.Pattern(), handler.HandlerFunc())

	type request struct {
		path   string
		method string
	}
	type want struct {
		response    any
		contentType string
		statusCode  int
	}
	type test struct {
		setup   func()
		request request
		name    string
		want    want
	}
	tests := []test{
		{
			name: "success get user urls",
			request: request{
				path:   "/api/user/urls",
				method: http.MethodGet,
			},
			want: want{
				statusCode:  http.StatusOK,
				contentType: "application/json",
				response: []dto.UserURLsResponse{
					{
						ShortURL:    "http://localhost:8080/shortURL1",
						OriginalURL: "https://example.com/1",
					},
					{
						ShortURL:    "http://localhost:8080/shortURL2",
						OriginalURL: "https://example.com/2",
					},
				},
			},
			setup: func() {
				usecase.EXPECT().GetUserURLs(gomock.Any(), gomock.Any()).Return([]entity.URL{
					{
						ShortURL:    "shortURL1",
						OriginalURL: "https://example.com/1",
					},
					{
						ShortURL:    "shortURL2",
						OriginalURL: "https://example.com/2",
					},
				}, nil)
			},
		},
		{
			name: "no urls found",
			request: request{
				path:   "/api/user/urls",
				method: http.MethodGet,
			},
			want: want{
				statusCode:  http.StatusNoContent,
				contentType: "application/json",
				response: handlers.Response{
					Status:  http.StatusNoContent,
					Message: "No Content",
					Data:    "No URLs found",
				},
			},
			setup: func() {
				usecase.EXPECT().GetUserURLs(gomock.Any(), gomock.Any()).Return(nil, nil)
			},
		},
		{
			name: "internal server error",
			request: request{
				path:   "/api/user/urls",
				method: http.MethodGet,
			},
			want: want{
				statusCode:  http.StatusInternalServerError,
				contentType: "application/json",
				response: handlers.Response{
					Status:  http.StatusInternalServerError,
					Message: "Internal Server Error",
					Data:    "Failed to get user URLs",
				},
			},
			setup: func() {
				usecase.EXPECT().
					GetUserURLs(gomock.Any(), gomock.Any()).
					Return(nil, errors.New("failed to get user URLs"))
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			test.setup()
			req, errRequest := http.NewRequest(test.request.method, test.request.path, nil)
			require.NoError(t, errRequest)
			recorder := httptest.NewRecorder()
			router.ServeHTTP(recorder, req)

			require.Equal(t, test.want.statusCode, recorder.Code)
			require.Equal(t, test.want.contentType, recorder.Header().Get("Content-Type"))
			switch test.want.response.(type) {
			case []dto.UserURLsResponse:
				var response []dto.UserURLsResponse
				err = json.NewDecoder(recorder.Body).Decode(&response)
				require.NoError(t, err)
				require.Equal(t, test.want.response, response)
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
