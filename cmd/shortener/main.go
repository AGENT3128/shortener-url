package main

import (
	"net/http"
	"time"

	"github.com/AGENT3128/shortener-url/internal/app/handlers"
	"github.com/AGENT3128/shortener-url/internal/app/storage"
)

func main() {
	if err := runServer(); err != nil {
		panic(err)
	}
}

func runServer() error {
	storage.InitStorage()
	mux := http.NewServeMux()
	mux.HandleFunc(`/`, handlers.URLHandler)
	server := &http.Server{
		Addr:              "localhost:8080",
		Handler:           mux,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      10 * time.Second,
		IdleTimeout:       60 * time.Second,
		ReadHeaderTimeout: 10 * time.Second,
	}
	return server.ListenAndServe()
}
