package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

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

	// base context for server
	serverCtx, serverStopCtx := context.WithCancel(context.Background())

	// channel for OS signals
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)

	// start goroutine for handling OS signals
	go func() {
		<-sig

		// base context for graceful shutdown
		shutdownCtx, cancel := context.WithTimeout(serverCtx, 30*time.Second)
		defer cancel()

		// start goroutine for monitoring timeout
		go func() {
			<-shutdownCtx.Done()
			if shutdownCtx.Err() == context.DeadlineExceeded {
				logger.Log.Fatal("Graceful shutdown timed out.. forcing exit.")
			}
		}()

		// call graceful shutdown
		logger.Log.Info("Shutting down server...")
		if err := server.Shutdown(shutdownCtx); err != nil {
			logger.Log.Fatal("Server forced to shutdown", zap.Error(err))
		}

		serverStopCtx()
	}()

	// start server
	logger.Log.Info("Starting server...")
	if err := server.Run(); err != nil && err != http.ErrServerClosed {
		logger.Log.Error("Error starting server", zap.Error(err))
		return err
	}

	// wait for server to exit
	<-serverCtx.Done()
	logger.Log.Info("Server exited properly")

	return nil
}
