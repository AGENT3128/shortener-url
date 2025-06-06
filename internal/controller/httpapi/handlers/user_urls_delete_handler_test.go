package handlers

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"go.uber.org/zap"

	"github.com/AGENT3128/shortener-url/internal/controller/httpapi/handlers/mocks"
	customMiddleware "github.com/AGENT3128/shortener-url/internal/controller/httpapi/middleware"
)

func TestUserURLsDeleteHandler(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	usecase := mocks.NewMockUserURLDeleter(ctrl)
	logger, err := zap.NewDevelopment()
	require.NoError(t, err)

	handler, err := NewUserURLsDeleteHandler(
		WithUserURLsDeleteUsecase(usecase),
		WithUserURLsDeleteLogger(logger),
	)
	require.NoError(t, err)

	require.Equal(t, "/api/user/urls", handler.Pattern())
	require.Equal(t, http.MethodDelete, handler.Method())

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
			name: "success delete urls",
			request: request{
				path:   "/api/user/urls",
				method: http.MethodDelete,
				body:   []string{"shortURL1", "shortURL2"},
			},
			want: want{
				statusCode:  http.StatusAccepted,
				contentType: "application/json",
				response:    Response{Status: http.StatusAccepted, Message: "Accepted", Data: "success"},
			},
			setup: func() {
				usecase.EXPECT().
					DeleteUserURLs(gomock.Any(), gomock.Any(), []string{"shortURL1", "shortURL2"}).
					Return(nil)
			},
		},
		{
			name: "invalid request body",
			request: request{
				path:   "/api/user/urls",
				method: http.MethodDelete,
				body:   "invalid",
			},
			want: want{
				statusCode:  http.StatusBadRequest,
				contentType: "application/json",
				response: Response{
					Status:  http.StatusBadRequest,
					Message: "Bad Request",
					Data:    "invalid request format",
				},
			},
			setup: func() {

			},
		},
		{
			name: "empty body",
			request: request{
				path:   "/api/user/urls",
				method: http.MethodDelete,
				body:   nil,
			},
			want: want{
				statusCode:  http.StatusBadRequest,
				contentType: "application/json",
				response: Response{
					Status:  http.StatusBadRequest,
					Message: "Bad Request",
					Data:    "no URLs provided for deletion",
				},
			},
			setup: func() {
			},
		},
		{
			name: "internal server error",
			want: want{
				statusCode:  http.StatusInternalServerError,
				contentType: "application/json",
				response: Response{
					Status:  http.StatusInternalServerError,
					Message: "Internal Server Error",
					Data:    "failed to delete URLs",
				},
			},
			request: request{
				path:   "/api/user/urls",
				method: http.MethodDelete,
				body:   []string{"shortURL1", "shortURL2"},
			},
			setup: func() {
				usecase.EXPECT().
					DeleteUserURLs(gomock.Any(), gomock.Any(), []string{"shortURL1", "shortURL2"}).
					Return(errors.New("failed to delete URLs"))
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			test.setup()
			body, err := json.Marshal(test.request.body)
			require.NoError(t, err)

			req, err := http.NewRequest(test.request.method, test.request.path, bytes.NewReader(body))
			require.NoError(t, err)

			recorder := httptest.NewRecorder()
			router.ServeHTTP(recorder, req)

			require.Equal(t, test.want.statusCode, recorder.Code)
			require.Equal(t, test.want.contentType, recorder.Header().Get("Content-Type"))
			switch test.want.response.(type) {
			case Response:
				var response Response
				err = json.NewDecoder(recorder.Body).Decode(&response)
				require.NoError(t, err)
				require.Equal(t, test.want.response, response)
			default:
				require.Equal(t, test.want.response, recorder.Body.String())
			}
		})
	}
}
