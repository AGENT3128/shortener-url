package handlers

import (
	"encoding/json"
	"net/http"
)

// Response is the response for the JSON.
type Response struct {
	Status  int    `json:"status"`
	Message string `json:"message"`
	Data    any    `json:"data"`
}

// JSONResponse is the response for the JSON.
func JSONResponse(w http.ResponseWriter, status int, data any) {
	response := Response{
		Status:  status,
		Message: http.StatusText(status),
		Data:    data,
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
	}
}

// TextResponse is the response for the text.
func TextResponse(w http.ResponseWriter, status int, data string) {
	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(status)
	_, _ = w.Write([]byte(data))
}
