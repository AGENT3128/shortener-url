package handlers

import (
	"context"

	"github.com/AGENT3128/shortener-url/internal/app/storage"
)

// URLOriginalGetter describes the behavior for retrieving short URL by original URL
type URLOriginalGetter interface {
	GetByOriginalURL(ctx context.Context, originalURL string) (string, bool)
}

// URLSaver describes the behavior for saving short URL by original URL
type URLSaver interface {
	Add(ctx context.Context, shortID, originalURL string) (string, error)
}

// URLRepository combines URL getting and saving capabilities
type URLRepository interface {
	URLOriginalGetter
	URLSaver
}

// TransactionSupport interface for batch operations
type TransactionSupport interface {
	AddBatch(ctx context.Context, urls []storage.URL) error
}
