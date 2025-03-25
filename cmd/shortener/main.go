package main

import (
	"fmt"

	"github.com/AGENT3128/shortener-url/internal/app/config"
	"github.com/AGENT3128/shortener-url/internal/app/handlers"
	"github.com/AGENT3128/shortener-url/internal/app/server"
	"github.com/AGENT3128/shortener-url/internal/app/storage"
)

func main() {
	if err := run(); err != nil {
		panic(err)
	}
}

func run() error {
	cfg := config.NewConfig()
	repository := storage.NewMemStorage()

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

	// TODO: remove this, it's for debugging
	fmt.Println("Server address:", cfg.ServerAddress)
	fmt.Println("Base URL address:", cfg.BaseURLAddress)

	return server.Run()

}
