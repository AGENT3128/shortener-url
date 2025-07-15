package usecase

import (
	"context"

	"github.com/AGENT3128/shortener-url/internal/entity"
)

//go:generate mockgen -source=interfaces.go -destination=./mocks/usecase_mock.go -package=mocks

// URLRepository is the interface for the URLRepository.
type URLRepository interface {
	URLSaver
	URLGetter
	Pinger
	BatchURLSaver
	UserURLGetter
	URLDeleter
	Closer
	StatsGetter
}

// StatsGetter is the interface for the StatsGetter.
type StatsGetter interface {
	GetStats(ctx context.Context) (urlsCount int, usersCount int, err error)
}

// URLSaver is the interface for the URLSaver.
type URLSaver interface {
	Add(ctx context.Context, userID, shortURL, originalURL string) (string, error)
}

// URLGetter is the interface for the URLGetter.
type URLGetter interface {
	GetByOriginalURL(ctx context.Context, originalURL string) (string, error)
	GetByShortURL(ctx context.Context, shortURL string) (string, error)
}

// Pinger is the interface for the Pinger.
type Pinger interface {
	Ping(ctx context.Context) error
}

// BatchURLSaver is the interface for the BatchURLSaver.
type BatchURLSaver interface {
	AddBatch(ctx context.Context, userID string, urls []entity.URL) error
}

// UserURLGetter is the interface for the UserURLGetter.
type UserURLGetter interface {
	GetUserURLs(ctx context.Context, userID string) ([]entity.URL, error)
}

// URLDeleter is the interface for the URLDeleter.
type URLDeleter interface {
	MarkDeletedBatch(ctx context.Context, userID string, shortURLs []string) error
}

// Closer is the interface for the Closer.
type Closer interface {
	Close() error
}
