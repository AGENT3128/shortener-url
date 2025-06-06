package httpapi

import (
	//nolint:gosec // pprof is used for debugging
	_ "net/http/pprof"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"go.uber.org/zap"

	"github.com/AGENT3128/shortener-url/internal/controller/httpapi/handlers"
	customMiddleware "github.com/AGENT3128/shortener-url/internal/controller/httpapi/middleware"
)

type options struct {
	logger     *zap.Logger
	baseURL    string
	URLusecase URLusecase
}
type Option func(options *options) error

func WithLogger(logger *zap.Logger) Option {
	return func(options *options) error {
		options.logger = logger
		return nil
	}
}

func WithBaseURL(baseURL string) Option {
	return func(options *options) error {
		options.baseURL = baseURL
		return nil
	}
}

func WithURLUsecase(usecase URLusecase) Option {
	return func(options *options) error {
		options.URLusecase = usecase
		return nil
	}
}

func NewRouter(opts ...Option) (*chi.Mux, error) {
	options := &options{}
	for _, opt := range opts {
		if err := opt(options); err != nil {
			return nil, err
		}
	}

	// custom middlewares
	customLogger, err := customMiddleware.NewHandlerLogger(
		customMiddleware.WithMiddlewareLogger(options.logger),
	)
	if err != nil {
		return nil, err
	}
	authMiddleware, err := customMiddleware.NewAuthMiddleware(
		customMiddleware.WithAuthMiddlewareLogger(options.logger),
	)
	if err != nil {
		return nil, err
	}

	router := chi.NewRouter()
	router.Use(middleware.Logger)
	router.Use(middleware.Recoverer)
	router.Use(customLogger.Handler())
	router.Use(authMiddleware.Handler())
	router.Use(customMiddleware.GzipMiddleware())
	// pprof
	router.Mount("/debug", middleware.Profiler())

	err = initializeHandlers(router, options)
	if err != nil {
		return nil, err
	}
	return router, nil
}

func initializeHandlers(router *chi.Mux, options *options) error {
	// handlers
	shortenHandler, err := handlers.NewShortenHandler(
		handlers.WithShortenBaseURL(options.baseURL),
		handlers.WithShortenUsecase(options.URLusecase),
		handlers.WithShortenLogger(options.logger),
	)
	if err != nil {
		return err
	}

	redirectHandler, err := handlers.NewRedirectHandler(
		handlers.WithRedirectUsecase(options.URLusecase),
		handlers.WithRedirectLogger(options.logger),
	)
	if err != nil {
		return err
	}

	apiShortenHandler, err := handlers.NewAPIShortenHandler(
		handlers.WithAPIShortenUsecase(options.URLusecase),
		handlers.WithAPIShortenLogger(options.logger),
		handlers.WithAPIShortenBaseURL(options.baseURL),
	)
	if err != nil {
		return err
	}

	pingHandler, err := handlers.NewPingHandler(
		handlers.WithPingUsecase(options.URLusecase),
		handlers.WithPingLogger(options.logger),
	)
	if err != nil {
		return err
	}

	batchShortenHandler, err := handlers.NewBatchShortenHandler(
		handlers.WithBatchShortenUsecase(options.URLusecase),
		handlers.WithBatchShortenLogger(options.logger),
		handlers.WithBatchShortenBaseURL(options.baseURL),
	)
	if err != nil {
		return err
	}

	userURLsHandler, err := handlers.NewUserURLsHandler(
		handlers.WithUserURLsBaseURL(options.baseURL),
		handlers.WithUserURLsUsecase(options.URLusecase),
		handlers.WithUserURLsLogger(options.logger),
	)
	if err != nil {
		return err
	}

	userURLsDeleteHandler, err := handlers.NewUserURLsDeleteHandler(
		handlers.WithUserURLsDeleteUsecase(options.URLusecase),
		handlers.WithUserURLsDeleteLogger(options.logger),
	)
	if err != nil {
		return err
	}
	h := []handler{
		shortenHandler,
		redirectHandler,
		apiShortenHandler,
		pingHandler,
		batchShortenHandler,
		userURLsHandler,
		userURLsDeleteHandler,
	}

	for _, h := range h {
		router.Method(h.Method(), h.Pattern(), h.HandlerFunc())
	}
	return nil
}
