package main

import (
	"github.com/AGENT3128/shortener-url/internal/app/config"
	"github.com/AGENT3128/shortener-url/internal/app/logger"
	"github.com/AGENT3128/shortener-url/internal/app/server"
	"go.uber.org/zap"
)

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

	server, err := server.NewServer(
		server.WithConfig(cfg),
		server.WithLogger(logger.Log),
	)
	if err != nil {
		return err
	}
	defer server.Close()

	logger.Log.Info("Server address", zap.String("address", cfg.ServerAddress))
	logger.Log.Info("Base URL address", zap.String("address", cfg.BaseURLAddress))
	logger.Log.Info("Release mode", zap.String("mode", cfg.ReleaseMode))
	logger.Log.Info("Log level", zap.String("level", cfg.LogLevel))
	logger.Log.Info("File storage path", zap.String("path", cfg.FileStoragePath))

	return server.Run()
}
