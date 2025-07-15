package handlers

import (
	"encoding/json"
	"net"
	"net/http"
)

// Response is the response for the JSON.
type Response struct {
	Data    any    `json:"data"`
	Message string `json:"message"`
	Status  int    `json:"status"`
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

// isIPInTrustedSubnet checks if the client IP is in the trusted subnet.
func isIPInTrustedSubnet(r *http.Request, trustedSubnet string) (bool, error) {
	if trustedSubnet == "" {
		return false, ErrTrustedSubnetEmpty
	}

	clientIP := getClientIP(r)
	if clientIP == "" {
		return false, ErrClientIPEmpty
	}

	_, trustedNet, err := net.ParseCIDR(trustedSubnet)
	if err != nil {
		return false, ErrInvalidTrustedSubnetCIDR
	}

	ip := net.ParseIP(clientIP)
	if ip == nil {
		return false, ErrInvalidClientIP
	}

	return trustedNet.Contains(ip), nil
}

// getClientIP gets the client IP from X-Real-IP header or remote address.
func getClientIP(r *http.Request) string {
	// Check X-Real-IP header first
	if ip := r.Header.Get("X-Real-IP"); ip != "" {
		return ip
	}

	// Fallback to X-Forwarded-For
	if ip := r.Header.Get("X-Forwarded-For"); ip != "" {
		return ip
	}

	// Fallback to remote address
	if ip, _, err := net.SplitHostPort(r.RemoteAddr); err == nil {
		return ip
	}

	return r.RemoteAddr
}
