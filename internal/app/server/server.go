package server

import (
	"fmt"
	"net/http"
	"time"

	"github.com/AGENT3128/shortener-url/internal/app/config"
	"github.com/AGENT3128/shortener-url/internal/app/handlers"
	"github.com/AGENT3128/shortener-url/internal/app/logger"
	"github.com/AGENT3128/shortener-url/internal/app/middleware"
	"github.com/AGENT3128/shortener-url/internal/app/storage"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type IHandler interface {
	IPattern
	IMethod
	IHandlerFunc
}

type IPattern interface {
	Pattern() string
}

type IMethod interface {
	Method() string
}

type IHandlerFunc interface {
	Handler() gin.HandlerFunc
}

type Repository interface {
	ShortenerSet
	ShortenerGet
}

type ShortenerSet interface {
	Add(shortID, originalURL string)
}
type ShortenerGet interface {
	GetByShortID(shortID string) (string, bool)
	GetByOriginalURL(originalURL string) (string, bool)
}
type Server struct {
	httpServer *http.Server
	router     *gin.Engine
	handlers   []IHandler
}
type options struct {
	config   *config.Config
	logger   *zap.Logger
	repo     Repository
	handlers []IHandler
}
type Option func(options *options) error

func WithHandler(handler IHandler) Option {
	return func(o *options) error {
		o.handlers = append(o.handlers, handler)
		return nil
	}
}

func WithConfig(config *config.Config) Option {
	return func(o *options) error {
		o.config = config
		return nil
	}
}

func WithLogger(logger *zap.Logger) Option {
	return func(o *options) error {
		o.logger = logger
		return nil
	}
}

func NewServer(opts ...Option) (*Server, error) {
	options := &options{}
	for _, opt := range opts {
		if err := opt(options); err != nil {
			return nil, err
		}
	}

	if options.config == nil {
		return nil, fmt.Errorf("config is nil")
	}
	if options.logger == nil {
		return nil, fmt.Errorf("logger is nil")
	}

	// Set Gin mode
	switch options.config.ReleaseMode {
	case "debug":
		gin.SetMode(gin.DebugMode)
	case "release":
		gin.SetMode(gin.ReleaseMode)
	case "test":
		gin.SetMode(gin.TestMode)
	default:
		return nil, fmt.Errorf("invalid release mode: %s", options.config.ReleaseMode)
	}

	if options.config.FileStoragePath != "" {
		repo, err := storage.NewFileStorage(options.config.FileStoragePath, options.logger)
		if err != nil {
			return nil, err
		}
		defer repo.Close()
		options.repo = repo
	} else {
		options.repo = storage.NewMemStorage(logger.Log)
	}

	// Create router and setup middleware
	router := gin.Default()
	// router.Use(gin.Logger())
	router.Use(middleware.HandlerLogger())
	router.Use(gin.Recovery())
	router.Use(middleware.GzipMiddleware())

	// Setup handlers
	options.handlers = append(options.handlers, handlers.NewShortenHandler(options.repo, options.config.BaseURLAddress, options.logger))
	options.handlers = append(options.handlers, handlers.NewRedirectHandler(options.repo, options.logger))
	options.handlers = append(options.handlers, handlers.NewAPIShortenHandler(options.repo, options.config.BaseURLAddress, options.logger))

	for _, handler := range options.handlers {
		router.Handle(handler.Method(), handler.Pattern(), handler.Handler())
	}

	return &Server{
		httpServer: &http.Server{
			Addr:              options.config.ServerAddress,
			Handler:           router,
			ReadTimeout:       10 * time.Second,
			WriteTimeout:      10 * time.Second,
			IdleTimeout:       60 * time.Second,
			ReadHeaderTimeout: 10 * time.Second,
		},
		router:   router,
		handlers: options.handlers,
	}, nil
}

func (s *Server) Run() error {
	return s.httpServer.ListenAndServe()
}
