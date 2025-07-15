package dto

// ShortenResponse represents the response for a shortened URL.
type ShortenResponse struct {
	Result string `json:"result"`
}

// ShortenBatchResponse represents an item in the batch shortening response.
type ShortenBatchResponse struct {
	CorrelationID string `json:"correlation_id"`
	ShortURL      string `json:"short_url"`
}

// UserURLsResponse represents individual URL in the response.
type UserURLsResponse struct {
	ShortURL    string `json:"short_url"`
	OriginalURL string `json:"original_url"`
}

// StatsResponse represents the response for stats endpoint.
type StatsResponse struct {
	URLs  int `json:"urls"`
	Users int `json:"users"`
}
