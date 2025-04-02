package db

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/AGENT3128/shortener-url/internal/migrations"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
)

type Database struct {
	Conn *pgxpool.Pool
}

func NewDatabase(ctx context.Context, databaseDSN string) (*Database, error) {
	pool, err := NewConnectionPool(ctx, databaseDSN)
	if err != nil {
		return nil, err
	}

	return &Database{Conn: pool}, nil
}

func NewConnectionPool(ctx context.Context, databaseDSN string) (*pgxpool.Pool, error) {
	dbConfig, err := pgxpool.ParseConfig(databaseDSN)
	if err != nil {
		return nil, err
	}

	conn, err := pgxpool.NewWithConfig(ctx, dbConfig)
	if err != nil {
		return nil, err
	}

	err = conn.Ping(ctx)
	if err != nil {
		return nil, fmt.Errorf("error connecting to the database: %w", err)
	}

	return conn, nil
}

// GetSQLDB returns a standard sql.DB from the pgx connection pool
func (d *Database) GetSQLDB() (*sql.DB, error) {
	dbConfig := d.Conn.Config().ConnConfig
	dsn := stdlib.RegisterConnConfig(dbConfig)
	return sql.Open("pgx", dsn)
}

// MigrateWithContext performs migrations using embedded SQL files with provided context
func (d *Database) MigrateWithContext(ctx context.Context) error {
	db, err := d.GetSQLDB()
	if err != nil {
		return fmt.Errorf("failed to open database: %w", err)
	}
	defer db.Close()

	// Set the migration file system
	goose.SetBaseFS(migrations.GetMigrationsFS())

	if err := goose.SetDialect("postgres"); err != nil {
		return fmt.Errorf("failed to set dialect: %w", err)
	}

	if err := goose.Up(db, migrations.MigrationsPath()); err != nil {
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	return nil
}

// Close closes the database connection
func (d *Database) Close() error {
	if d.Conn != nil {
		d.Conn.Close()
	}
	return nil
}
