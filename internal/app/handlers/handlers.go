package handlers

import (
	"fmt"
	"io"
	"net/http"

	"github.com/AGENT3128/shortener-url/internal/app/helpers"
	"github.com/AGENT3128/shortener-url/internal/app/storage"
)

type URLHandler struct {
	repository storage.Repository
}

func NewURLHandler(repo storage.Repository) *URLHandler {
	return &URLHandler{
		repository: repo,
	}
}

func (h *URLHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		h.handlePost(w, r)
	case http.MethodGet:
		if r.URL.Path == "/" {
			http.Error(w, "Not found", http.StatusNotFound)
			return
		}
		h.handleGet(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (h *URLHandler) handlePost(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Failed to read request body", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	originalURL := string(body)
	if originalURL == "" {
		http.Error(w, "Original URL is empty", http.StatusBadRequest)
		return
	}

	shortID, ok := h.repository.GetByOriginalURL(originalURL)
	if !ok {
		shortID = helpers.GenerateShortID()
		h.repository.Add(shortID, originalURL)
	}

	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusCreated)
	fmt.Fprintf(w, "http://localhost:8080/%s", shortID)
}

func (h *URLHandler) handleGet(w http.ResponseWriter, r *http.Request) {
	shortID := r.URL.Path[1:]
	originalURL, ok := h.repository.GetByShortID(shortID)
	if !ok {
		http.Error(w, "Short URL not found", http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "text/plain")
	w.Header().Set("Location", originalURL)
	w.WriteHeader(http.StatusTemporaryRedirect)
}
