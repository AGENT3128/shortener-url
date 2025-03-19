package handlers

import (
	"fmt"
	"io"
	"net/http"

	"github.com/AGENT3128/shortener-url/internal/app/helpers"
	"github.com/AGENT3128/shortener-url/internal/app/storage"
)

func PostShortURLHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Failed to read request body", http.StatusBadRequest)
		return
	}
	originalURL := string(body)
	if originalURL == "" {
		http.Error(w, "Original URL is empty", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	w.Header().Set("Content-Type", "text/plain")

	shortID, ok := storage.StorageURLs.GetByOriginalURL(originalURL)
	if ok {
		shortURL := fmt.Sprintf("http://localhost:8080/%s", shortID)
		w.WriteHeader(http.StatusCreated)
		_, err = w.Write([]byte(shortURL))
		if err != nil {
			http.Error(w, "Failed to write response", http.StatusInternalServerError)
			return
		}
		return
	}
	shortID = helpers.GenerateShortID()
	shortURL := fmt.Sprintf("http://localhost:8080/%s", shortID)
	storage.StorageURLs.Add(shortID, originalURL)
	w.WriteHeader(http.StatusCreated)
	_, err = w.Write([]byte(shortURL))
	if err != nil {
		http.Error(w, "Failed to write response", http.StatusInternalServerError)
		return
	}
}

func GetOriginalURLHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	shortID := r.URL.Path[1:]
	originalURL, ok := storage.StorageURLs.GetByShortID(shortID)
	if !ok {
		http.Error(w, "Short URL not found", http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "text/plain")
	w.Header().Set("Location", originalURL)
	w.WriteHeader(http.StatusTemporaryRedirect)
}

func URLHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		PostShortURLHandler(w, r)
	case http.MethodGet:
		if r.URL.Path != "/" {
			GetOriginalURLHandler(w, r)
		} else {
			http.Error(w, "Not found", http.StatusBadRequest)
		}
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}
