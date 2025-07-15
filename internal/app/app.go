package app

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"go.uber.org/zap"

	"github.com/AGENT3128/shortener-url/internal/config"
	"github.com/AGENT3128/shortener-url/internal/controller/httpapi"
	"github.com/AGENT3128/shortener-url/internal/entity"
	"github.com/AGENT3128/shortener-url/internal/infrastructure/httpserver"
	"github.com/AGENT3128/shortener-url/internal/logger"
	"github.com/AGENT3128/shortener-url/internal/repository/file"
	"github.com/AGENT3128/shortener-url/internal/repository/memory"
	"github.com/AGENT3128/shortener-url/internal/repository/postgres"
	"github.com/AGENT3128/shortener-url/internal/usecase"
	"github.com/AGENT3128/shortener-url/internal/worker"
	"github.com/AGENT3128/shortener-url/pkg/database"
)

// URLSaver is an interface that defines the methods for saving a URL.
type URLSaver interface {
	Add(ctx context.Context, userID, shortURL, originalURL string) (string, error)
}

// URLGetter is an interface that defines the methods for getting a URL.
type URLGetter interface {
	GetByOriginalURL(ctx context.Context, originalURL string) (string, error)
	GetByShortURL(ctx context.Context, shortURL string) (string, error)
}

// Pinger is an interface that defines the method for pinging the database.
type Pinger interface {
	Ping(ctx context.Context) error
}

// BatchURLSaver is an interface that defines the method for saving a batch of URLs.
type BatchURLSaver interface {
	AddBatch(ctx context.Context, userID string, urls []entity.URL) error
}

// UserURLGetter is an interface that defines the method for getting a user's URLs.
type UserURLGetter interface {
	GetUserURLs(ctx context.Context, userID string) ([]entity.URL, error)
}

// URLDeleter is an interface that defines the method for deleting a URL.
type URLDeleter interface {
	MarkDeletedBatch(ctx context.Context, userID string, shortURLs []string) error
}

// Closer is an interface that defines the method for closing the repository.
type Closer interface {
	Close() error
}

// StatsGetter is an interface that defines the method for getting stats.
type StatsGetter interface {
	GetStats(ctx context.Context) (urlsCount int, usersCount int, err error)
}

// Repository is an interface that defines the methods for the repository.
type Repository interface {
	URLSaver
	URLGetter
	Pinger
	BatchURLSaver
	UserURLGetter
	URLDeleter
	Closer
	StatsGetter
}

// Run is the main function for running the application.
func Run(cfg *config.Config) error {
	logger, err := logger.NewLogger(cfg.LogLevel)
	if err != nil {
		return fmt.Errorf("failed to create logger: %w", err)
	}
	defer func() {
		if errSync := logger.Sync(); errSync != nil {
			// https://github.com/uber-go/zap/issues/991#issuecomment-962098428
			if !errors.Is(errSync, syscall.ENOTTY) {
				logger.Error("Failed to sync logger", zap.Error(errSync))
			}
		}
	}()

	ctx := context.Background()

	var db *database.Database
	if cfg.DatabaseDSN != "" {
		var errDB error
		db, errDB = database.New(
			ctx,
			cfg.DatabaseDSN,
			database.WithConnMaxIdleTime(cfg.DatabaseConnMaxIdleTime),
			database.WithConnMaxLifetime(cfg.DatabaseConnMaxLifetime),
			database.WithMaxConns(cfg.DatabaseMaxConns),
			database.WithMinConns(cfg.DatabaseMinConns),
			database.WithHealthCheckPeriod(cfg.DatabaseHealthCheckPeriod),
		)
		if errDB != nil {
			return fmt.Errorf("failed to create database: %w", errDB)
		}
	}

	// repositories
	var urlRepository Repository

	switch {
	case cfg.DatabaseDSN != "":
		urlRepository = postgres.NewURLRepository(db, logger)
	case cfg.FileStoragePath != "":
		urlRepository, err = file.NewFileStorage(cfg.FileStoragePath, logger)
		if err != nil {
			return fmt.Errorf("failed to create file storage: %w", err)
		}
	default:
		urlRepository = memory.NewMemStorage(logger)
	}

	// worker
	deleteWorker := worker.NewDeleteWorker(
		urlRepository,
		logger,
	)

	// usecases
	urlUsecase, err := usecase.NewURLUsecase(
		usecase.WithURLUsecaseLogger(logger),
		usecase.WithURLUsecaseRepository(urlRepository),
		usecase.WithDeleteWorker(deleteWorker),
	)
	if err != nil {
		return fmt.Errorf("failed to create url usecase: %w", err)
	}

	router, err := httpapi.NewRouter(
		httpapi.WithLogger(logger),
		httpapi.WithBaseURL(cfg.BaseURLAddress),
		httpapi.WithURLUsecase(urlUsecase),
		httpapi.WithTrustedSubnet(cfg.TrustedSubnet),
	)
	if err != nil {
		return fmt.Errorf("failed to create router: %w", err)
	}

	httpserver, err := httpserver.New(
		httpserver.WithAddress(cfg.HTTPServerAddress),
		httpserver.WithReadHeaderTimeout(cfg.HTTPServerReadHeaderTimeout),
		httpserver.WithReadTimeout(cfg.HTTPServerReadTimeout),
		httpserver.WithWriteTimeout(cfg.HTTPServerWriteTimeout),
		httpserver.WithIdleTimeout(cfg.HTTPServerIdleTimeout),
		httpserver.WithHandler(router),
		httpserver.WithHTTPS(cfg.EnableHTTPS),
		httpserver.WithTLSCertPath(cfg.TLSCertPath),
		httpserver.WithTLSKeyPath(cfg.TLSKeyPath),
	)
	if err != nil {
		return fmt.Errorf("failed to create server: %w", err)
	}

	serverCtx, serverCancel := context.WithCancel(context.Background())
	gracefulShutdown(serverCtx, serverCancel, logger, httpserver, cfg.GracefulShutdownTimeout, urlUsecase)

	// TODO: may be rework to receive from a channel and blocking select to run multiple servers
	logger.Info("Starting server", zap.String("address", httpserver.Address()))
	if errRun := httpserver.Start(); errRun != nil && !errors.Is(errRun, http.ErrServerClosed) {
		logger.Fatal("Error starting server", zap.Error(errRun))
	}

	<-serverCtx.Done()
	logger.Info("Server exited properly")

	return nil
}

func gracefulShutdown(
	ctx context.Context,
	serverCancel context.CancelFunc,
	logger *zap.Logger,
	server *httpserver.Server,
	gracefulShutdownTimeout time.Duration,
	urlUsecase *usecase.URLUsecase,
) {
	// https://github.com/go-chi/chi/blob/master/_examples/graceful/main.go

	// channel for OS signals
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)

	// start goroutine for handling OS signals
	go func() {
		<-sig

		// base context for graceful shutdown
		shutdownCtx, cancel := context.WithTimeout(ctx, gracefulShutdownTimeout)
		defer cancel()

		// start goroutine for monitoring timeout
		go func() {
			select {
			case <-shutdownCtx.Done():
				if shutdownCtx.Err() == context.DeadlineExceeded {
					logger.Fatal("Graceful shutdown timed out.. forcing exit.")
				}
			default:
			}
		}()

		logger.Info("Shutting down worker...")
		urlUsecase.Shutdown()

		// call graceful shutdown server
		logger.Info("Shutting down server...")
		if err := server.Shutdown(shutdownCtx); err != nil {
			logger.Fatal("Server forced to shutdown", zap.Error(err))
		}

		serverCancel()
	}()
}
