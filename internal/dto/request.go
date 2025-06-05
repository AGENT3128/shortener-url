package dto

// ShortenRequest represents the request for shortening a URL.
type ShortenRequest struct {
	URL string `json:"url"`
}

// ShortenBatchRequest represents an item in the batch shortening request.
type ShortenBatchRequest struct {
	CorrelationID string `json:"correlation_id"`
	OriginalURL   string `json:"original_url"`
}
