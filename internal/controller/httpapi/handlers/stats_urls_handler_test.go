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
)

func TestStatsURLsHandler(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	usecase := mocks.NewMockStatsGetter(ctrl)
	logger, err := zap.NewDevelopment()
	require.NoError(t, err)

	handler, err := handlers.NewStatsURLsHandler(
		handlers.WithStatsURLsUsecase(usecase),
		handlers.WithStatsURLsLogger(logger),
		handlers.WithStatsURLsTrustedSubnet("127.0.0.0/24"),
	)
	require.NoError(t, err)

	require.Equal(t, "/api/internal/stats", handler.Pattern())
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
	tests := []struct {
		setup             func()
		request           request
		name              string
		clientRealIP      string
		clientForwardedIP string
		want              want
	}{
		{
			name: "success get stats with real IP",
			request: request{
				path:   "/api/internal/stats",
				method: http.MethodGet,
			},
			want: want{
				statusCode:  http.StatusOK,
				contentType: "application/json",
				response: dto.StatsResponse{
					URLs:  1,
					Users: 1,
				},
			},
			clientRealIP: "127.0.0.2",
			setup: func() {
				usecase.EXPECT().GetStats(gomock.Any()).Return(1, 1, nil)
			},
		},
		{
			name: "success get stats with forwarded IP",
			request: request{
				path:   "/api/internal/stats",
				method: http.MethodGet,
			},
			want: want{
				statusCode:  http.StatusOK,
				contentType: "application/json",
				response: dto.StatsResponse{
					URLs:  1,
					Users: 1,
				},
			},
			clientForwardedIP: "127.0.0.3",
			setup: func() {
				usecase.EXPECT().GetStats(gomock.Any()).Return(1, 1, nil)
			},
		},
		{
			name: "forbidden access",
			request: request{
				path:   "/api/internal/stats",
				method: http.MethodGet,
			},
			want: want{
				statusCode:  http.StatusForbidden,
				contentType: "application/json",
				response: handlers.Response{
					Status:  http.StatusForbidden,
					Message: "Forbidden",
					Data:    "Forbidden",
				},
			},
			setup: func() {
			},
			clientRealIP: "192.168.1.2",
		},
		{
			name: "internal server error",
			request: request{
				path:   "/api/internal/stats",
				method: http.MethodGet,
			},
			want: want{
				statusCode:  http.StatusInternalServerError,
				contentType: "application/json",
				response: handlers.Response{
					Status:  http.StatusInternalServerError,
					Message: "Internal Server Error",
					Data:    "Internal Server Error",
				},
			},
			setup: func() {
				usecase.EXPECT().GetStats(gomock.Any()).Return(0, 0, errors.New("internal server error"))
			},
			clientRealIP: "127.0.0.1",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			test.setup()
			req, errRequest := http.NewRequest(
				test.request.method,
				test.request.path,
				nil,
			)
			if test.clientRealIP != "" {
				req.Header.Set("X-Real-IP", test.clientRealIP)
			}
			if test.clientForwardedIP != "" {
				req.Header.Set("X-Forwarded-For", test.clientForwardedIP)
			}
			require.NoError(t, errRequest)
			recorder := httptest.NewRecorder()
			router.ServeHTTP(recorder, req)

			require.Equal(t, test.want.statusCode, recorder.Code)
			require.Equal(t, test.want.contentType, recorder.Header().Get("Content-Type"))
			switch test.want.response.(type) {
			case dto.StatsResponse:
				var response dto.StatsResponse
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
