package main

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/AGENT3128/shortener-url/internal/app/config"
	"github.com/AGENT3128/shortener-url/internal/app/handlers"
	"github.com/AGENT3128/shortener-url/internal/app/storage"
)

func main() {
	config.InitConfig()
	// TODO: remove this, it's for debugging
	fmt.Println("Server address:", config.Config.ServerAddress)
	fmt.Println("Base URL address:", config.Config.BaseURLAddress)
	//

	if err := runServer(); err != nil {
		panic(err)
	}
}

func runServer() error {
	gin.SetMode(gin.ReleaseMode)
	router := gin.Default()
	router.Use(gin.Logger())
	router.Use(gin.Recovery())

	repository := storage.NewMemStorage()
	handler := handlers.NewURLHandler(repository)
	handler.SetupRoutes(router)

	server := &http.Server{
		Addr:              config.Config.ServerAddress,
		Handler:           router,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      10 * time.Second,
		IdleTimeout:       60 * time.Second,
		ReadHeaderTimeout: 10 * time.Second,
	}
	return server.ListenAndServe()
}
