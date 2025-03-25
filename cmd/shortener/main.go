package main

import (
	"github.com/AGENT3128/shortener-url/internal/app/config"
	"github.com/AGENT3128/shortener-url/internal/app/handlers"
	"github.com/AGENT3128/shortener-url/internal/app/logger"
	"github.com/AGENT3128/shortener-url/internal/app/server"
	"github.com/AGENT3128/shortener-url/internal/app/storage"
	"go.uber.org/zap"
)

func main() {
	if err := run(); err != nil {
		panic(err)
	}
}

func run() error {
	cfg := config.NewConfig()
	repository := storage.NewMemStorage()

	if err := logger.InitLogger(cfg.LogLevel); err != nil {
		return err
	}

	server, err := server.NewServer(
		server.WithMode(cfg.ReleaseMode),
		server.WithServerAddress(cfg.ServerAddress),
		server.WithBaseURL(cfg.BaseURLAddress),
		server.WithHandler(handlers.NewShortenHandler(repository, cfg.BaseURLAddress)),
		server.WithHandler(handlers.NewRedirectHandler(repository)),
	)
	if err != nil {
		return err
	}

	logger.Log.Info("Server address", zap.String("address", cfg.ServerAddress))
	logger.Log.Info("Base URL address", zap.String("address", cfg.BaseURLAddress))
	logger.Log.Info("Release mode", zap.String("mode", cfg.ReleaseMode))
	logger.Log.Info("Log level", zap.String("level", cfg.LogLevel))

	return server.Run()
}
