package postgres

import (
	"context"
	"time"

	"github.com/AGENT3128/shortener-url/internal/entity"
	"github.com/AGENT3128/shortener-url/internal/repository/postgres/generated"

	"go.uber.org/zap"

	"github.com/AGENT3128/shortener-url/pkg/database"
)

//go:generate sqlc generate

type URLRepository struct {
	db      *database.Database
	logger  *zap.Logger
	queries *generated.Queries
}

func NewURLRepository(db *database.Database, logger *zap.Logger) *URLRepository {
	return &URLRepository{
		db:      db,
		logger:  logger.With(zap.String("repository", "url")),
		queries: generated.New(db.Pool),
	}
}

func (r *URLRepository) Add(ctx context.Context, userID, shortURL, originalURL string) (string, error) {
	shortURL, err := r.queries.AddURL(ctx, generated.AddURLParams{
		UserID:      userID,
		ShortUrl:    shortURL,
		OriginalUrl: originalURL,
		CreatedAt:   time.Now(),
	})
	if err != nil {
		return shortURL, err
	}
	return shortURL, nil
}

func (r *URLRepository) GetByOriginalURL(ctx context.Context, originalURL string) (string, error) {
	shortURL, err := r.queries.GetURLByOriginalURL(ctx, originalURL)
	if err != nil {
		return "", err
	}
	r.logger.Info("original URL found", zap.String("original_url", originalURL), zap.String("short_url", shortURL))
	return shortURL, nil
}

func (r *URLRepository) GetByShortURL(ctx context.Context, shortURL string) (string, error) {
	row, err := r.queries.GetURLByShortURL(ctx, shortURL)
	if err != nil {
		return "", err
	}
	if row.IsDeleted {
		return row.OriginalUrl, entity.ErrURLDeleted
	}
	return row.OriginalUrl, nil
}

func (r *URLRepository) Ping(ctx context.Context) error {
	return r.db.Pool.Ping(ctx)
}

func (r *URLRepository) AddBatch(ctx context.Context, userID string, urls []entity.URL) error {
	tx, err := r.db.Pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	qtx := r.queries.WithTx(tx)
	now := time.Now()

	for _, url := range urls {
		_, err := qtx.AddURL(ctx, generated.AddURLParams{
			UserID:      userID,
			ShortUrl:    url.ShortURL,
			OriginalUrl: url.OriginalURL,
			CreatedAt:   now,
		})
		if err != nil {
			return err
		}
	}

	return tx.Commit(ctx)
}

func (r *URLRepository) GetUserURLs(ctx context.Context, userID string) ([]entity.URL, error) {
	urls, err := r.queries.GetURLsByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}
	result := make([]entity.URL, 0, len(urls))
	for _, url := range urls {
		result = append(result, entity.URL{
			ShortURL:    url.ShortUrl,
			OriginalURL: url.OriginalUrl,
		})
	}
	return result, nil
}

func (r *URLRepository) MarkDeletedBatch(ctx context.Context, userID string, shortURLs []string) error {
	err := r.queries.MarkDeletedBatch(ctx, generated.MarkDeletedBatchParams{
		UserID:  userID,
		Column2: shortURLs,
	})
	r.logger.Info("marked deleted batch", zap.String("userID", userID), zap.Any("shortURLs", shortURLs))
	if err != nil {
		return err
	}
	return nil
}
