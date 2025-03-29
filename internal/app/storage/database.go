package storage

import (
	"context"

	"github.com/AGENT3128/shortener-url/internal/app/config"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Database struct {
	pool *pgxpool.Pool
}

func NewDatabase(ctx context.Context, cfg *config.Config) (*Database, error) {
	pool, err := pgxpool.New(ctx, cfg.DatabaseDSN)
	if err != nil {
		return nil, err
	}
	// defer pool.Close()

	return &Database{pool: pool}, nil
}

func (d *Database) Close() {
	d.pool.Close()
}

func (d *Database) Ping(ctx context.Context) error {
	return d.pool.Ping(ctx)
}
