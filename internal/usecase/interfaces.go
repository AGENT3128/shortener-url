package usecase

import (
	"context"

	"github.com/AGENT3128/shortener-url/internal/entity"
)

//go:generate mockgen -source=interfaces.go -destination=./mocks/usecase_mock.go -package=mocks
type URLRepository interface {
	URLSaver
	URLGetter
	Pinger
	BatchURLSaver
	UserURLGetter
	URLDeleter
}

type URLSaver interface {
	Add(ctx context.Context, userID, shortURL, originalURL string) (string, error)
}

type URLGetter interface {
	GetByOriginalURL(ctx context.Context, originalURL string) (string, error)
	GetByShortURL(ctx context.Context, shortURL string) (string, error)
}

type Pinger interface {
	Ping(ctx context.Context) error
}

type BatchURLSaver interface {
	AddBatch(ctx context.Context, userID string, urls []entity.URL) error
}

type UserURLGetter interface {
	GetUserURLs(ctx context.Context, userID string) ([]entity.URL, error)
}

type URLDeleter interface {
	MarkDeletedBatch(ctx context.Context, userID string, shortURLs []string) error
}
