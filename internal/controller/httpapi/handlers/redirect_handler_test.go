package handlers

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"go.uber.org/zap"

	"github.com/AGENT3128/shortener-url/internal/controller/httpapi/handlers/mocks"
	customMiddleware "github.com/AGENT3128/shortener-url/internal/controller/httpapi/middleware"
	"github.com/AGENT3128/shortener-url/internal/entity"
)

func TestRedirectHandler(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	usecase := mocks.NewMockURLGetter(ctrl)
	logger, err := zap.NewDevelopment()
	require.NoError(t, err)

	handler, err := NewRedirectHandler(
		WithRedirectUsecase(usecase),
		WithRedirectLogger(logger),
	)
	require.NoError(t, err)

	require.Equal(t, "/{id}", handler.Pattern())
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
		location    string
	}
	tests := []struct {
		name    string
		request request
		want    want
		setup   func()
	}{
		{
			name: "success redirect",
			request: request{
				path:   "/shortURL123",
				method: http.MethodGet,
			},
			want: want{
				statusCode:  http.StatusTemporaryRedirect,
				contentType: "text/plain",
				location:    "https://example.com",
			},
			setup: func() {
				usecase.EXPECT().GetByShortURL(gomock.Any(), "shortURL123").Return("https://example.com", nil)
			},
		},
		{
			name: "url deleted",
			request: request{
				path:   "/shortURL1234",
				method: http.MethodGet,
			},
			want: want{
				statusCode:  http.StatusGone,
				contentType: "application/json",
				location:    "",
			},
			setup: func() {
				usecase.EXPECT().GetByShortURL(gomock.Any(), "shortURL1234").Return("", entity.ErrURLDeleted)
			},
		},
		{
			name: "url not found",
			request: request{
				path:   "/shortURL12345",
				method: http.MethodGet,
			},
			want: want{
				statusCode:  http.StatusNotFound,
				contentType: "application/json",
				location:    "",
			},
			setup: func() {
				usecase.EXPECT().GetByShortURL(gomock.Any(), "shortURL12345").Return("", entity.ErrURLNotFound)
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
			require.Equal(t, test.want.location, recorder.Header().Get("Location"))
		})
	}
}
