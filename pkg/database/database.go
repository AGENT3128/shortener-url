package database

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	// import driver pgx for register in sqlx.
	_ "github.com/jackc/pgx/v5/stdlib"
)

type options struct {
	maxConns          int
	minConns          int
	connMaxLifetime   time.Duration
	connMaxIdleTime   time.Duration
	healthCheckPeriod time.Duration
}

type Option func(options *options) error

func WithMaxConns(maxConns int) Option {
	return func(options *options) error {
		options.maxConns = maxConns
		return nil
	}
}

func WithMinConns(minConns int) Option {
	return func(options *options) error {
		options.minConns = minConns
		return nil
	}
}

func WithConnMaxLifetime(connMaxLifetime time.Duration) Option {
	return func(options *options) error {
		options.connMaxLifetime = connMaxLifetime
		return nil
	}
}

func WithConnMaxIdleTime(connMaxIdleTime time.Duration) Option {
	return func(options *options) error {
		options.connMaxIdleTime = connMaxIdleTime
		return nil
	}
}

func WithHealthCheckPeriod(healthCheckPeriod time.Duration) Option {
	return func(options *options) error {
		options.healthCheckPeriod = healthCheckPeriod
		return nil
	}
}

type Database struct {
	Pool *pgxpool.Pool
}

func New(ctx context.Context, databaseURL string, opts ...Option) (*Database, error) {
	options := &options{}
	for _, opt := range opts {
		if err := opt(options); err != nil {
			return nil, err
		}
	}

	config, err := pgxpool.ParseConfig(databaseURL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse conn string: %w", err)
	}

	config.MaxConns = int32(options.maxConns) //nolint:gosec // maxConns is a positive integer
	config.MinConns = int32(options.minConns) //nolint:gosec // minConns is a positive integer
	config.MaxConnLifetime = options.connMaxLifetime
	config.MaxConnIdleTime = options.connMaxIdleTime
	config.HealthCheckPeriod = options.healthCheckPeriod

	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		return nil, fmt.Errorf("failed to create pool: %w", err)
	}

	if errPing := pool.Ping(ctx); errPing != nil {
		return nil, fmt.Errorf("failed to ping database: %w", errPing)
	}

	return &Database{Pool: pool}, nil
}

func (d *Database) Close() {
	d.Pool.Close()
}
