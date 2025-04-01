package storage

import (
	"context"
	"errors"
	"time"

	"github.com/AGENT3128/shortener-url/internal/app/db"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
	"go.uber.org/zap"
)

type URL struct {
	ShortID     string    `json:"short_id"`
	OriginalURL string    `json:"original_url"`
	CreatedAt   time.Time `json:"created_at"`
}

type URLRepository struct {
	db        *db.Database
	logger    *zap.Logger
	tableName string
}

func NewURLRepository(db *db.Database, logger *zap.Logger) *URLRepository {
	logger = logger.With(zap.String("storage", "database"))
	return &URLRepository{db: db, logger: logger, tableName: "urls"}
}

// WithTableName allows to change table name (useful for testing)
func (r *URLRepository) WithTableName(name string) *URLRepository {
	r.tableName = name
	return r
}

func (r *URLRepository) Add(ctx context.Context, shortID, originalURL string) (string, error) {
	r.logger.Debug("Adding URL to database", zap.String("short_id", shortID), zap.String("original_url", originalURL))
	sql := `
		INSERT INTO ` + r.tableName + ` (short_id, original_url, created_at)
		VALUES ($1, $2, $3)
	`

	_, err := r.db.Conn.Exec(
		ctx,
		sql,
		shortID,
		originalURL,
		time.Now(),
	)

	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == pgerrcode.UniqueViolation {
			// get existing shortID
			if existingShortID, exists := r.GetByOriginalURL(ctx, originalURL); exists {
				return existingShortID, ErrURLExists
			}
		}
		r.logger.Error("Failed to add URL to database", zap.Error(err))
		return "", err
	}

	return shortID, nil
}

func (r *URLRepository) GetByShortID(ctx context.Context, shortID string) (string, bool) {
	r.logger.Debug("Getting URL from database", zap.String("short_id", shortID))
	sql := `
		SELECT original_url
		FROM ` + r.tableName + `
		WHERE short_id = $1
	`

	var url URL
	err := r.db.Conn.QueryRow(
		ctx,
		sql,
		shortID,
	).Scan(&url.OriginalURL)

	if err != nil {
		r.logger.Error("Failed to get URL from database", zap.Error(err))
		return "", false
	}
	return url.OriginalURL, true
}

func (r *URLRepository) GetByOriginalURL(ctx context.Context, originalURL string) (string, bool) {
	r.logger.Debug("Getting URL from database", zap.String("original_url", originalURL))
	sql := `
		SELECT short_id
		FROM ` + r.tableName + `
		WHERE original_url = $1
	`
	var url URL
	err := r.db.Conn.QueryRow(
		ctx,
		sql,
		originalURL,
	).Scan(&url.ShortID)

	if err != nil {
		r.logger.Error("Failed to get URL from database", zap.Error(err))
		return "", false
	}
	return url.ShortID, true
}

func (r *URLRepository) AddBatch(ctx context.Context, urls []URL) error {
	r.logger.Debug("Adding batch of URLs to database", zap.Any("urls", urls))
	sql := `
		INSERT INTO ` + r.tableName + ` (short_id, original_url, created_at)
		VALUES ($1, $2, $3)
	`
	tx, err := r.db.Conn.Begin(ctx)
	if err != nil {
		r.logger.Error("Failed to begin transaction", zap.Error(err))
		return err
	}
	defer tx.Rollback(ctx)

	for _, url := range urls {
		r.logger.Info("Adding URL to database", zap.Any("url", url))
		_, err := tx.Exec(ctx, sql, url.ShortID, url.OriginalURL, time.Now())
		if err != nil {
			r.logger.Error("Failed to add URL to database", zap.Error(err))
			return err
		}
	}

	err = tx.Commit(ctx)
	if err != nil {
		r.logger.Error("Failed to commit transaction", zap.Error(err))
		return err
	}
	return nil
}

// Close closes the database connection
func (r *URLRepository) Close() error {
	if r.db != nil {
		return r.db.Close()
	}
	return nil
}
