package handlers_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"

	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"

	"github.com/AGENT3128/shortener-url/internal/controller/httpapi/handlers"
	"github.com/AGENT3128/shortener-url/internal/dto"
)

// This example demonstrates how to shorten a URL using the plain text endpoint.
func Example_shortenURL() {
	// Create a new router and register the handler
	r := chi.NewRouter()

	// Initialize your dependencies here
	var urlService handlers.URLSaver
	logger := zap.NewExample()

	h, err := handlers.NewShortenHandler(
		handlers.WithShortenBaseURL("http://localhost:8080"),
		handlers.WithShortenLogger(logger),
		handlers.WithShortenUsecase(urlService),
	)
	if err != nil {
		fmt.Println("Error creating handler:", err)
		return
	}
	r.Method(h.Method(), h.Pattern(), h.HandlerFunc())

	// Create a test server
	ts := httptest.NewServer(r)
	defer ts.Close()

	// Create the request
	longURL := "https://practicum.yandex.ru"
	req, _ := http.NewRequest(http.MethodPost, ts.URL+"/", bytes.NewBufferString(longURL))

	// Send the request
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		fmt.Println("Error sending request:", err)
		return
	}
	defer resp.Body.Close()

	// Read and print the response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("Error reading response body:", err)
		return
	}
	fmt.Printf("Status: %d\nShortened URL: %s\n", resp.StatusCode, string(body))

	// Output:
	// Status: 201
	// Shortened URL: http://localhost:8080/abc123
}

// This example demonstrates how to shorten a URL using the JSON API endpoint.
func Example_shortenURLJSON() {
	r := chi.NewRouter()

	var urlService handlers.URLSaver
	logger := zap.NewExample()

	h, err := handlers.NewAPIShortenHandler(
		handlers.WithAPIShortenBaseURL("http://localhost:8080"),
		handlers.WithAPIShortenLogger(logger),
		handlers.WithAPIShortenUsecase(urlService),
	)
	if err != nil {
		fmt.Println("Error creating handler:", err)
		return
	}
	r.Method(h.Method(), h.Pattern(), h.HandlerFunc())

	ts := httptest.NewServer(r)
	defer ts.Close()

	// Create request payload
	payload := dto.ShortenRequest{
		URL: "https://practicum.yandex.ru",
	}
	jsonData, err := json.Marshal(payload)
	if err != nil {
		fmt.Println("Error marshalling payload:", err)
		return
	}

	// Send request
	req, _ := http.NewRequest(http.MethodPost, ts.URL+"/api/shorten", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		fmt.Println("Error sending request:", err)
		return
	}
	defer resp.Body.Close()

	// Parse and print response
	var result dto.ShortenResponse
	json.NewDecoder(resp.Body).Decode(&result)
	fmt.Printf("Status: %d\nResult: %+v\n", resp.StatusCode, result)

	// Output:
	// Status: 201
	// Result: {Result:http://localhost:8080/abc123}
}

// This example demonstrates how to batch shorten multiple URLs.
func Example_batchShorten() {
	r := chi.NewRouter()

	var urlService handlers.BatchURLSaver
	logger := zap.NewExample()

	h, err := handlers.NewBatchShortenHandler(
		handlers.WithBatchShortenBaseURL("http://localhost:8080"),
		handlers.WithBatchShortenLogger(logger),
		handlers.WithBatchShortenUsecase(urlService),
	)
	if err != nil {
		fmt.Println("Error creating handler:", err)
		return
	}
	r.Method(h.Method(), h.Pattern(), h.HandlerFunc())

	ts := httptest.NewServer(r)
	defer ts.Close()

	// Create batch request
	batch := []dto.ShortenBatchRequest{
		{
			CorrelationID: "1",
			OriginalURL:   "https://practicum.yandex.ru",
		},
		{
			CorrelationID: "2",
			OriginalURL:   "https://ya.ru",
		},
	}
	jsonData, err := json.Marshal(batch)
	if err != nil {
		fmt.Println("Error marshalling batch:", err)
		return
	}

	// Send request
	req, _ := http.NewRequest(http.MethodPost, ts.URL+"/api/shorten/batch", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		fmt.Println("Error sending request:", err)
		return
	}
	defer resp.Body.Close()

	// Parse and print response
	var result []dto.ShortenBatchResponse
	json.NewDecoder(resp.Body).Decode(&result)
	fmt.Printf("Status: %d\nBatch Result: %+v\n", resp.StatusCode, result)

	// Output:
	// Status: 201
	// Batch Result: [{CorrelationID:1 ShortURL:http://localhost:8080/abc123} {CorrelationID:2 ShortURL:http://localhost:8080/def456}]
}

// This example demonstrates how to get user's URLs.
func Example_getUserURLs() {
	r := chi.NewRouter()

	var urlService handlers.UserURLGetter
	logger := zap.NewExample()

	h, err := handlers.NewUserURLsHandler(
		handlers.WithUserURLsBaseURL("http://localhost:8080"),
		handlers.WithUserURLsLogger(logger),
		handlers.WithUserURLsUsecase(urlService),
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
	// Status: 200
	// User URLs: [{ShortURL:http://localhost:8080/abc123 OriginalURL:https://practicum.yandex.ru} {ShortURL:http://localhost:8080/def456 OriginalURL:https://ya.ru}]
}

// This example demonstrates how to delete user's URLs.
func Example_deleteUserURLs() {
	r := chi.NewRouter()

	var urlService handlers.UserURLDeleter
	logger := zap.NewExample()

	h, err := handlers.NewUserURLsDeleteHandler(
		handlers.WithUserURLsDeleteLogger(logger),
		handlers.WithUserURLsDeleteUsecase(urlService),
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
	r := chi.NewRouter()

	var storage handlers.Pinger
	logger := zap.NewExample()

	h, err := handlers.NewPingHandler(
		handlers.WithPingLogger(logger),
		handlers.WithPingUsecase(storage),
	)
	if err != nil {
		fmt.Println("Error creating handler:", err)
		return
	}
	r.Method(h.Method(), h.Pattern(), h.HandlerFunc())

	ts := httptest.NewServer(r)
	defer ts.Close()

	// Create request with authorization header
	req, err := http.NewRequest(http.MethodGet, ts.URL+"/ping", nil)
	if err != nil {
		fmt.Println("Error creating request:", err)
		return
	}
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
	// Response: Database is alive
}
