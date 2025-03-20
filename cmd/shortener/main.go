package main

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/AGENT3128/shortener-url/internal/app/config"
	"github.com/AGENT3128/shortener-url/internal/app/handlers"
	"github.com/AGENT3128/shortener-url/internal/app/storage"
)

func main() {
	cfg := config.NewConfig()
	if err := runServer(cfg); err != nil {
		panic(err)
	}
}

func runServer(cfg *config.Config) error {
	gin.SetMode(gin.ReleaseMode)
	router := gin.Default()
	router.Use(gin.Logger())
	router.Use(gin.Recovery())

	repository := storage.NewMemStorage()
	handler := handlers.NewURLHandler(repository, cfg)
	handler.SetupRoutes(router)

	server := &http.Server{
		Addr:              cfg.ServerAddress,
		Handler:           router,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      10 * time.Second,
		IdleTimeout:       60 * time.Second,
		ReadHeaderTimeout: 10 * time.Second,
	}
	return server.ListenAndServe()
}
