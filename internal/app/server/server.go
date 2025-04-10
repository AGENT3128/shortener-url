package server

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/AGENT3128/shortener-url/internal/app/config"
	"github.com/AGENT3128/shortener-url/internal/app/db"
	"github.com/AGENT3128/shortener-url/internal/app/handlers"
	"github.com/AGENT3128/shortener-url/internal/app/logger"
	"github.com/AGENT3128/shortener-url/internal/app/middleware"
	"github.com/AGENT3128/shortener-url/internal/app/models"
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
	PingDB
	GetUserURLs
}

type ShortenerSet interface {
	Add(ctx context.Context, userID, shortID, originalURL string) (string, error)
	AddBatch(ctx context.Context, userID string, urls []models.URL) error
}

type ShortenerGet interface {
	GetByShortID(ctx context.Context, shortID string) (string, bool)
	GetByOriginalURL(ctx context.Context, originalURL string) (string, bool)
}

type PingDB interface {
	Ping(ctx context.Context) error
}

type GetUserURLs interface {
	GetUserURLs(ctx context.Context, userID string) ([]models.URL, error)
}

type Server struct {
	httpServer *http.Server
	router     *gin.Engine
	handlers   []IHandler
	repo       Repository
}
type options struct {
	config *config.Config
	logger *zap.Logger
}
type Option func(options *options) error

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

	ctx := context.Background()

	var database *db.Database
	if options.config.DatabaseDSN != "" {
		var err error

		ctxDB, cancel := context.WithTimeout(ctx, 10*time.Second)
		defer cancel()

		database, err = db.NewDatabase(ctxDB, options.config.DatabaseDSN)
		if err != nil {
			return nil, err
		}

		ctxMigration, cancelMigration := context.WithTimeout(ctx, 1*time.Minute)
		defer cancelMigration()

		err = database.MigrateWithContext(ctxMigration)
		if err != nil {
			return nil, err
		}
	}

	var repo Repository
	if database != nil {
		repo = storage.NewURLRepository(database, options.logger)
	} else if options.config.FileStoragePath != "" {
		var err error
		repo, err = storage.NewFileStorage(options.config.FileStoragePath, options.logger)
		if err != nil {
			return nil, err
		}
	} else {
		repo = storage.NewMemStorage(logger.Log)
	}

	// Create router and setup middleware
	router := gin.Default()
	// router.Use(gin.Logger())
	router.Use(gin.Recovery())
	router.Use(middleware.HandlerLogger())
	router.Use(middleware.GzipMiddleware())
	router.Use(middleware.AuthMiddleware())

	// Setup handlers
	handlers := []IHandler{
		handlers.NewShortenHandler(repo, options.config.BaseURLAddress, options.logger),
		handlers.NewRedirectHandler(repo, options.logger),
		handlers.NewAPIShortenHandler(repo, options.config.BaseURLAddress, options.logger),
		handlers.NewPingHandler(repo, options.logger),
		handlers.NewShortenBatchHandler(repo, options.config.BaseURLAddress, options.logger),
		handlers.NewUserURLsHandler(repo, options.config.BaseURLAddress, options.logger),
	}

	for _, handler := range handlers {
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
		handlers: handlers,
		repo:     repo,
	}, nil
}

func (s *Server) Run() error {
	return s.httpServer.ListenAndServe()
}

func (s *Server) Shutdown(ctx context.Context) error {
	return s.httpServer.Shutdown(ctx)
}

func (s *Server) Close() error {
	if closer, ok := s.repo.(interface{ Close() error }); ok {
		return closer.Close()
	}
	return nil
}
