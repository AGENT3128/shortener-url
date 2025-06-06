// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.29.0

package generated

import (
	"context"
)

type Querier interface {
	AddURL(ctx context.Context, arg AddURLParams) (string, error)
	GetURLByOriginalURL(ctx context.Context, originalUrl string) (string, error)
	GetURLByShortURL(ctx context.Context, shortUrl string) (GetURLByShortURLRow, error)
	GetURLsByUserID(ctx context.Context, userID string) ([]Url, error)
	MarkDeletedBatch(ctx context.Context, arg MarkDeletedBatchParams) error
}

var _ Querier = (*Queries)(nil)
