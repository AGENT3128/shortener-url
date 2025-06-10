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
)

func TestPingHandler(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	usecase := mocks.NewMockPinger(ctrl)
	logger, err := zap.NewDevelopment()
	require.NoError(t, err)

	handler, err := handlers.NewPingHandler(
		handlers.WithPingUsecase(usecase),
		handlers.WithPingLogger(logger),
	)
	require.NoError(t, err)

	require.Equal(t, "/ping", handler.Pattern())
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
			name: "success ping",
			request: request{
				path:   "/ping",
				method: http.MethodGet,
			},
			want: want{
				statusCode:  http.StatusOK,
				contentType: "application/json",
				response:    handlers.Response{Status: http.StatusOK, Message: "OK", Data: "Database is alive"},
			},
			setup: func() {
				usecase.EXPECT().Ping(gomock.Any()).Return(nil)
			},
		},
		{
			name: "internal server error",
			request: request{
				path:   "/ping",
				method: http.MethodGet,
			},
			want: want{
				statusCode:  http.StatusInternalServerError,
				contentType: "application/json",
			},
			setup: func() {
				usecase.EXPECT().Ping(gomock.Any()).Return(errors.New("internal server error"))
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

			if test.want.response != nil {
				switch test.want.response.(type) {
				case handlers.Response:
					var response handlers.Response
					err = json.NewDecoder(recorder.Body).Decode(&response)
					require.NoError(t, err)
					require.Equal(t, test.want.response, response)
				default:
					t.Fatalf("unexpected response type: %T", test.want.response)
				}
			}
		})
	}
}
