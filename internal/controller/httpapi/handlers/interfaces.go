package handlers

import (
	"context"

	"github.com/AGENT3128/shortener-url/internal/entity"
)

//go:generate mockgen -source=interfaces.go -destination=./mocks/handlers_mock.go -package=mocks
type URLSaver interface {
	Add(ctx context.Context, userID string, originalURL string) (string, error)
}

// URLGetter is the interface for the URL getter.
type URLGetter interface {
	GetByShortURL(ctx context.Context, shortURL string) (string, error)
}

// Pinger is the interface for the pinger.
type Pinger interface {
	Ping(ctx context.Context) error
}

// BatchURLSaver is the interface for the batch URL saver.
type BatchURLSaver interface {
	AddBatch(ctx context.Context, userID string, urls []entity.URL) ([]entity.URL, error)
}

// UserURLGetter is the interface for the user URL getter.
type UserURLGetter interface {
	GetUserURLs(ctx context.Context, userID string) ([]entity.URL, error)
}

// UserURLDeleter is the interface for the user URL deleter.
type UserURLDeleter interface {
	DeleteUserURLs(ctx context.Context, userID string, shortURLs []string) error
}

// StatsGetter is the interface for getting stats.
type StatsGetter interface {
	GetStats(ctx context.Context) (urlsCount int, usersCount int, err error)
}
