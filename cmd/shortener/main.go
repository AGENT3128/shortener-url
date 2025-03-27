package main

import (
	"github.com/AGENT3128/shortener-url/internal/app/config"
	"github.com/AGENT3128/shortener-url/internal/app/handlers"
	"github.com/AGENT3128/shortener-url/internal/app/logger"
	"github.com/AGENT3128/shortener-url/internal/app/server"
	"github.com/AGENT3128/shortener-url/internal/app/storage"
	"go.uber.org/zap"
)

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

func main() {
	if err := run(); err != nil {
		panic(err)
	}
}

func run() error {
	cfg := config.NewConfig()
	if err := logger.InitLogger(cfg.LogLevel); err != nil {
		return err
	}

	var repository Repository
	if cfg.FileStoragePath != "" {
		repo, err := storage.NewFileStorage(cfg.FileStoragePath, logger.Log)
		if err != nil {
			return err
		}
		defer repo.Close()
		repository = repo
	} else {
		repository = storage.NewMemStorage(logger.Log)
	}

	server, err := server.NewServer(
		server.WithMode(cfg.ReleaseMode),
		server.WithServerAddress(cfg.ServerAddress),
		server.WithBaseURL(cfg.BaseURLAddress),
		server.WithHandler(handlers.NewShortenHandler(repository, cfg.BaseURLAddress, logger.Log)),
		server.WithHandler(handlers.NewRedirectHandler(repository, logger.Log)),
		server.WithHandler(handlers.NewAPIShortenHandler(repository, cfg.BaseURLAddress, logger.Log)),
	)
	if err != nil {
		return err
	}

	logger.Log.Info("Server address", zap.String("address", cfg.ServerAddress))
	logger.Log.Info("Base URL address", zap.String("address", cfg.BaseURLAddress))
	logger.Log.Info("Release mode", zap.String("mode", cfg.ReleaseMode))
	logger.Log.Info("Log level", zap.String("level", cfg.LogLevel))
	logger.Log.Info("File storage path", zap.String("path", cfg.FileStoragePath))

	return server.Run()
}
