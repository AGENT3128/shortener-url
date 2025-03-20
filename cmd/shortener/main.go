package main

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/AGENT3128/shortener-url/internal/app/handlers"
	"github.com/AGENT3128/shortener-url/internal/app/storage"
)

func main() {
	if err := runServer(); err != nil {
		panic(err)
	}
}

func runServer() error {
	router := gin.Default()
	router.Use(gin.Logger())
	router.Use(gin.Recovery())

	repository := storage.NewMemStorage()
	handler := handlers.NewURLHandler(repository)
	handler.SetupRoutes(router)

	server := &http.Server{
		Addr:              "localhost:8080",
		Handler:           router,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      10 * time.Second,
		IdleTimeout:       60 * time.Second,
		ReadHeaderTimeout: 10 * time.Second,
	}
	return server.ListenAndServe()
}
