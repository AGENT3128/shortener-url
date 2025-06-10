package examples_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"

	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"

	"github.com/AGENT3128/shortener-url/internal/controller/httpapi/handlers"
	customMiddleware "github.com/AGENT3128/shortener-url/internal/controller/httpapi/middleware"
	"github.com/AGENT3128/shortener-url/internal/dto"
	"github.com/AGENT3128/shortener-url/internal/repository/memory"
	"github.com/AGENT3128/shortener-url/internal/usecase"
	"github.com/AGENT3128/shortener-url/internal/worker"
)

// This example demonstrates how to get user's URLs.
func Example_getUserURLs() {
	// create logger
	logger := zap.NewNop()
	// create auth middleware
	authMiddleware, _ := customMiddleware.NewAuthMiddleware(
		customMiddleware.WithAuthMiddlewareLogger(logger),
	)
	// create router
	r := chi.NewRouter()
	// use auth middleware
	r.Use(authMiddleware.Handler())
	// create url repository
	urlRepository := memory.NewMemStorage(logger)
	// create url usecase
	usecase, _ := usecase.NewURLUsecase(
		usecase.WithURLUsecaseLogger(logger),
		usecase.WithURLUsecaseRepository(urlRepository),
	)
	// create user urls handler
	h, err := handlers.NewUserURLsHandler(
		handlers.WithUserURLsBaseURL("http://localhost:8080"),
		handlers.WithUserURLsLogger(logger),
		handlers.WithUserURLsUsecase(usecase),
	)
	if err != nil {
		fmt.Println("Error creating handler:", err)
		return
	}
	r.Method(h.Method(), h.Pattern(), h.HandlerFunc())

	ts := httptest.NewServer(r)
	defer ts.Close()

	// Send request with user token
	req, _ := http.NewRequest(http.MethodGet, ts.URL+"/api/user/urls", nil)
	req.Header.Set("Authorization", "Bearer user-token")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		fmt.Println("Error sending request:", err)
		return
	}
	defer resp.Body.Close()

	// Parse and print response
	var urls []dto.UserURLsResponse
	json.NewDecoder(resp.Body).Decode(&urls)
	fmt.Printf("Status: %d\nUser URLs: %+v\n", resp.StatusCode, urls)

	// Output:
	// Status: 204
	// User URLs: []
}

// This example demonstrates how to delete user's URLs.
func Example_deleteUserURLs() {
	// create logger
	logger := zap.NewNop()
	// create auth middleware
	authMiddleware, _ := customMiddleware.NewAuthMiddleware(
		customMiddleware.WithAuthMiddlewareLogger(logger),
	)
	// create router
	r := chi.NewRouter()
	// use auth middleware
	r.Use(authMiddleware.Handler())
	// create url repository
	urlRepository := memory.NewMemStorage(logger)
	// create delete worker
	deleteWorker := worker.NewDeleteWorker(
		urlRepository,
		logger,
	)
	// create url usecase
	usecase, _ := usecase.NewURLUsecase(
		usecase.WithURLUsecaseLogger(logger),
		usecase.WithURLUsecaseRepository(urlRepository),
		usecase.WithDeleteWorker(deleteWorker),
	)

	h, err := handlers.NewUserURLsDeleteHandler(
		handlers.WithUserURLsDeleteLogger(logger),
		handlers.WithUserURLsDeleteUsecase(usecase),
	)
	if err != nil {
		fmt.Println("Error creating handler:", err)
		return
	}
	r.Method(h.Method(), h.Pattern(), h.HandlerFunc())

	ts := httptest.NewServer(r)
	defer ts.Close()

	// Create delete request
	urls := []string{"short-id-1", "short-id-2"}
	jsonData, err := json.Marshal(urls)
	if err != nil {
		fmt.Println("Error marshalling urls:", err)
		return
	}

	// Send request
	req, _ := http.NewRequest(http.MethodDelete, ts.URL+"/api/user/urls", bytes.NewBuffer(jsonData))
	req.Header.Set("Authorization", "Bearer user-token")
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		fmt.Println("Error sending request:", err)
		return
	}
	defer resp.Body.Close()

	fmt.Printf("Status: %d\n", resp.StatusCode)

	// Output:
	// Status: 202
}

// This example demonstrates how to check service health.
func Example_pingDB() {
	// create logger
	logger := zap.NewNop()
	// create auth middleware
	authMiddleware, _ := customMiddleware.NewAuthMiddleware(
		customMiddleware.WithAuthMiddlewareLogger(logger),
	)
	// create router
	r := chi.NewRouter()
	// use auth middleware
	r.Use(authMiddleware.Handler())
	// create url repository
	urlRepository := memory.NewMemStorage(logger)
	// create url usecase
	usecase, _ := usecase.NewURLUsecase(
		usecase.WithURLUsecaseLogger(logger),
		usecase.WithURLUsecaseRepository(urlRepository),
	)
	// create ping handler
	h, _ := handlers.NewPingHandler(
		handlers.WithPingLogger(logger),
		handlers.WithPingUsecase(usecase),
	)

	r.Method(h.Method(), h.Pattern(), h.HandlerFunc())

	ts := httptest.NewServer(r)
	defer ts.Close()

	// Create request with authorization header
	req, _ := http.NewRequest(http.MethodGet, ts.URL+"/ping", nil)
	req.Header.Set("Authorization", "Bearer user-token")

	// Send request
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		fmt.Println("Error sending request:", err)
		return
	}
	defer resp.Body.Close()

	// Parse and print response
	var response map[string]string
	json.NewDecoder(resp.Body).Decode(&response)
	fmt.Printf("Status: %d\nResponse: %s\n", resp.StatusCode, response["message"])

	// Output:
	// Status: 200
	// Response: OK
}
