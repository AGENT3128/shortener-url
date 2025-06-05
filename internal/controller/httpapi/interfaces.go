package httpapi

import (
	"context"
	"net/http"

	"github.com/AGENT3128/shortener-url/internal/entity"
)

type PatternGetter interface {
	Pattern() string
}

type MethodGetter interface {
	Method() string
}

type HandlerFuncGetter interface {
	HandlerFunc() http.HandlerFunc
}

type handler interface {
	PatternGetter
	MethodGetter
	HandlerFuncGetter
}

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

type URLusecase interface {
	URLSaver
	URLGetter
	Pinger
	BatchURLSaver
	UserURLGetter
	UserURLDeleter
}
