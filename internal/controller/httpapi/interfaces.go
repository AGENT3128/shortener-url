package httpapi

import (
	"context"
	"net/http"

	"github.com/AGENT3128/shortener-url/internal/entity"
)

// PatternGetter is the interface for the pattern getter.
type PatternGetter interface {
	Pattern() string
}

// MethodGetter is the interface for the method getter.
type MethodGetter interface {
	Method() string
}

// HandlerFuncGetter is the interface for the handler func getter.
type HandlerFuncGetter interface {
	HandlerFunc() http.HandlerFunc
}

type handler interface {
	PatternGetter
	MethodGetter
	HandlerFuncGetter
}

// URLSaver is the interface for the URL saver.
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

// URLusecase is the interface for the URL usecase.
type URLusecase interface {
	URLSaver
	URLGetter
	Pinger
	BatchURLSaver
	UserURLGetter
	UserURLDeleter
}
