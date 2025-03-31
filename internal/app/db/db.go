package db

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Database struct {
	Conn *pgxpool.Pool
}

func NewDatabase(databaseDSN string) (*Database, error) {
	pool, err := NewConnectionPool(databaseDSN)
	if err != nil {
		return nil, err
	}

	return &Database{Conn: pool}, nil
}

func NewConnectionPool(databaseDSN string) (*pgxpool.Pool, error) {
	dbConfig, err := pgxpool.ParseConfig(databaseDSN)
	if err != nil {
		return nil, err
	}

	conn, err := pgxpool.NewWithConfig(context.Background(), dbConfig)
	if err != nil {
		return nil, err
	}

	err = conn.Ping(context.Background())
	if err != nil {
		return nil, fmt.Errorf("error connecting to the database: %w", err)
	}

	return conn, nil
}

func (d *Database) Migrate() error {
	// Read and execute migration file
	query := `
		CREATE TABLE IF NOT EXISTS urls (
		id SERIAL PRIMARY KEY,
		short_id TEXT NOT NULL UNIQUE,
		original_url TEXT NOT NULL,
		created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP
	);

	CREATE INDEX IF NOT EXISTS idx_urls_short_id ON urls(short_id);
	CREATE UNIQUE INDEX IF NOT EXISTS urls_original_url_idx ON urls (original_url);
	`
	// Execute migration query
	_, err := d.Conn.Exec(context.Background(), query)
	if err != nil {
		return fmt.Errorf("failed to execute migration: %w", err)
	}

	return nil
}
