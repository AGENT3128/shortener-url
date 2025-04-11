package storage

import (
	"context"
	"errors"
	"time"

	"github.com/AGENT3128/shortener-url/internal/app/db"
	"github.com/AGENT3128/shortener-url/internal/app/models"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"go.uber.org/zap"
)

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

func (r *URLRepository) Add(ctx context.Context, userID, shortID, originalURL string) (string, error) {
	r.logger.Debug("Adding URL to database", zap.String("short_id", shortID), zap.String("original_url", originalURL))
	sql := `
		INSERT INTO ` + r.tableName + ` (short_id, original_url, created_at, user_id)
		VALUES ($1, $2, $3, $4)
	`

	_, err := r.db.Conn.Exec(
		ctx,
		sql,
		shortID,
		originalURL,
		time.Now(),
		userID,
	)

	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == pgerrcode.UniqueViolation {
			// get existing shortID
			if existingShortID, exists := r.GetByOriginalURL(ctx, originalURL); exists {
				return existingShortID, models.ErrURLExists
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
		SELECT original_url, is_deleted
		FROM ` + r.tableName + `
		WHERE short_id = $1
	`

	var url models.URL
	var isDeleted bool
	err := r.db.Conn.QueryRow(
		ctx,
		sql,
		shortID,
	).Scan(&url.OriginalURL, &isDeleted)

	if err != nil {
		r.logger.Error("Failed to get URL from database", zap.Error(err))
		return "", false
	}

	// Return deleted status along with the URL
	if isDeleted {
		return url.OriginalURL, false
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
	var url models.URL
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

func (r *URLRepository) AddBatch(ctx context.Context, userID string, urls []models.URL) error {
	r.logger.Debug("Adding batch of URLs to database", zap.Any("urls", urls))
	sql := `
		INSERT INTO ` + r.tableName + ` (short_id, original_url, created_at, user_id)
		VALUES ($1, $2, $3, $4)
	`
	tx, err := r.db.Conn.Begin(ctx)
	if err != nil {
		r.logger.Error("Failed to begin transaction", zap.Error(err))
		return err
	}
	defer tx.Rollback(ctx)

	for _, url := range urls {
		r.logger.Info("Adding URL to database", zap.Any("url", url))
		_, err := tx.Exec(ctx, sql, url.ShortID, url.OriginalURL, time.Now(), userID)
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

func (r *URLRepository) Ping(ctx context.Context) error {
	return r.db.Conn.Ping(ctx)
}

func (r *URLRepository) GetUserURLs(ctx context.Context, userID string) ([]models.URL, error) {
	const method = "GetUserURLs"
	sql := `
		SELECT short_id, original_url
		FROM ` + r.tableName + `
		WHERE user_id = $1
	`
	rows, err := r.db.Conn.Query(ctx, sql, userID)
	if err != nil {
		r.logger.Error("Failed to get user URLs from database", zap.Error(err))
		return nil, err
	}
	defer rows.Close()

	urls := make([]models.URL, 0)
	for rows.Next() {
		var url models.URL
		err := rows.Scan(&url.ShortID, &url.OriginalURL)
		if err != nil {
			r.logger.Error("Failed to scan user URL", zap.Error(err))
			continue
		}
		urls = append(urls, url)
	}
	r.logger.Debug(method, zap.String("userID", userID), zap.Int("count", len(urls)))
	return urls, nil
}

// MarkDeletedBatch marks URLs as deleted in batch
func (r *URLRepository) MarkDeletedBatch(ctx context.Context, userID string, shortIDs []string) error {
	if len(shortIDs) == 0 {
		return nil
	}

	r.logger.Debug("Marking URLs as deleted",
		zap.String("user_id", userID),
		zap.Int("count", len(shortIDs)))

	sql := `
		UPDATE ` + r.tableName + `
		SET is_deleted = true
		WHERE short_id = ANY($1) AND user_id = $2
	`

	_, err := r.db.Conn.Exec(ctx, sql, shortIDs, userID)
	if err != nil {
		r.logger.Error("Failed to mark URLs as deleted",
			zap.String("user_id", userID),
			zap.Int("count", len(shortIDs)),
			zap.Error(err))
		return err
	}

	r.logger.Info("Successfully marked URLs as deleted",
		zap.String("user_id", userID),
		zap.Int("count", len(shortIDs)))
	return nil
}

// IsURLDeleted checks if URL is marked as deleted
func (r *URLRepository) IsURLDeleted(ctx context.Context, shortID string) (bool, error) {
	r.logger.Debug("Checking if URL is deleted", zap.String("short_id", shortID))

	sql := `
		SELECT COALESCE(is_deleted, false)
		FROM ` + r.tableName + `
		WHERE short_id = $1
	`

	var isDeleted bool
	err := r.db.Conn.QueryRow(ctx, sql, shortID).Scan(&isDeleted)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return false, nil
		}
		r.logger.Error("Failed to check if URL is deleted", zap.Error(err))
		return false, err
	}

	return isDeleted, nil
}
