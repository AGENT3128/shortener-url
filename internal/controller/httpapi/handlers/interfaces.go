package handlers

import (
	"context"

	"github.com/AGENT3128/shortener-url/internal/entity"
)

//go:generate mockgen -source=interfaces.go -destination=./mocks/handlers_mock.go -package=mocks
type URLSaver interface {
	Add(ctx context.Context, userID string, originalURL string) (string, error)
}

type URLGetter interface {
	GetByShortURL(ctx context.Context, shortURL string) (string, error)
}

type Pinger interface {
	Ping(ctx context.Context) error
}

type BatchURLSaver interface {
	AddBatch(ctx context.Context, userID string, urls []entity.URL) ([]entity.URL, error)
}

type UserURLGetter interface {
	GetUserURLs(ctx context.Context, userID string) ([]entity.URL, error)
}

type UserURLDeleter interface {
	DeleteUserURLs(ctx context.Context, userID string, shortURLs []string) error
}
