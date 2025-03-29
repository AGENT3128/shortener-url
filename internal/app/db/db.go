package db

import (
	"context"
	"fmt"
	"os"

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
	query, err := os.ReadFile("migrations/urls.sql")
	if err != nil {
		return fmt.Errorf("failed to read migration file: %w", err)
	}

	// Execute migration query
	_, err = d.Conn.Exec(context.Background(), string(query))
	if err != nil {
		return fmt.Errorf("failed to execute migration: %w", err)
	}

	return nil
}
