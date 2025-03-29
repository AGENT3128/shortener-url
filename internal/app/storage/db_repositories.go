package storage

import (
	"context"
	"time"

	"github.com/AGENT3128/shortener-url/internal/app/db"
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

func (r *URLRepository) Add(shortID, originalURL string) {
	sql := `
		INSERT INTO ` + r.tableName + ` (short_id, original_url, created_at)
		VALUES ($1, $2, $3)
	`

	_, err := r.db.Conn.Exec(
		context.Background(),
		sql,
		shortID,
		originalURL,
		time.Now(),
	)
	if err != nil {
		r.logger.Error("Failed to add URL to database", zap.Error(err))
		return
	}
}

func (r *URLRepository) GetByShortID(shortID string) (string, bool) {
	sql := `
		SELECT original_url
		FROM ` + r.tableName + `
		WHERE short_id = $1
	`

	var url URL
	err := r.db.Conn.QueryRow(
		context.Background(),
		sql,
		shortID,
	).Scan(&url.OriginalURL)

	if err != nil {
		r.logger.Error("Failed to get URL from database", zap.Error(err))
		return "", false
	}
	return url.OriginalURL, true
}

func (r *URLRepository) GetByOriginalURL(originalURL string) (string, bool) {
	sql := `
		SELECT short_id
		FROM ` + r.tableName + `
		WHERE original_url = $1
	`
	var url URL
	err := r.db.Conn.QueryRow(
		context.Background(),
		sql,
		originalURL,
	).Scan(&url.ShortID)

	if err != nil {
		r.logger.Error("Failed to get URL from database", zap.Error(err))
		return "", false
	}
	return url.ShortID, true
}
