package config

import (
	"flag"
	"fmt"
	"time"

	"github.com/caarlos0/env/v11"
)

// Config is the configuration for the application.
type Config struct {
	BaseURLAddress              string        `env:"BASE_URL"                        envDefault:"http://localhost:8080"` // base url for shortened urls
	ReleaseMode                 string        `env:"RELEASE_MODE"                    envDefault:"debug"`                 // release mode. Available options: debug, release, test
	LogLevel                    string        `env:"LOG_LEVEL"                       envDefault:"info"`                  // log level
	FileStoragePath             string        `env:"FILE_STORAGE_PATH"               envDefault:""`                      // file storage path
	DatabaseDSN                 string        `env:"DATABASE_DSN"                    envDefault:""`                      // database dsn
	HTTPServerAddress           string        `env:"HTTP_SERVER_ADDRESS"             envDefault:"localhost:8080"`        // http server address
	DatabaseMaxConns            int           `env:"DATABASE_MAX_CONNS"              envDefault:"10"`                    // database max conns
	DatabaseMinConns            int           `env:"DATABASE_MIN_CONNS"              envDefault:"2"`                     // database min conns
	DatabaseConnMaxLifetime     time.Duration `env:"DATABASE_CONN_MAX_LIFETIME"      envDefault:"10s"`                   // database connection max lifetime
	DatabaseConnMaxIdleTime     time.Duration `env:"DATABASE_CONN_MAX_IDLE_TIME"     envDefault:"10s"`                   // database connection max idle time
	DatabaseHealthCheckPeriod   time.Duration `env:"DATABASE_HEALTH_CHECK_PERIOD"    envDefault:"10s"`                   // database health check period
	HTTPServerIdleTimeout       time.Duration `env:"HTTP_SERVER_IDLE_TIMEOUT"        envDefault:"30s"`                   // http server idle timeout
	HTTPServerReadTimeout       time.Duration `env:"HTTP_SERVER_READ_TIMEOUT"        envDefault:"15s"`                   // http server read timeout
	HTTPServerReadHeaderTimeout time.Duration `env:"HTTP_SERVER_READ_HEADER_TIMEOUT" envDefault:"15s"`                   // http server read header timeout
	HTTPServerWriteTimeout      time.Duration `env:"HTTP_SERVER_WRITE_TIMEOUT"       envDefault:"10s"`                   // http server write timeout
	GracefulShutdownTimeout     time.Duration `env:"GRACEFUL_SHUTDOWN_TIMEOUT"       envDefault:"20s"`                   // graceful shutdown timeout
}

// NewConfig creates a new Config instance.
func NewConfig() (*Config, error) {
	// Create new config instance
	cfg := &Config{}

	// Parse environment variables
	if err := env.Parse(cfg); err != nil {
		return nil, fmt.Errorf("failed to parse environment variables: %w", err)
	}

	// Define and parse command line flags
	flag.StringVar(&cfg.HTTPServerAddress, "a", cfg.HTTPServerAddress, "HTTP server address")
	flag.StringVar(&cfg.BaseURLAddress, "b", cfg.BaseURLAddress, "Base URL for shortened URLs")
	flag.StringVar(&cfg.ReleaseMode, "r", cfg.ReleaseMode, "Release mode. Available options: debug, release, test")
	flag.StringVar(&cfg.LogLevel, "l", cfg.LogLevel, "Log level")
	flag.StringVar(&cfg.FileStoragePath, "f", cfg.FileStoragePath, "File storage path")
	flag.StringVar(&cfg.DatabaseDSN, "d", cfg.DatabaseDSN, "Database DSN")
	flag.DurationVar(
		&cfg.HTTPServerIdleTimeout,
		"http-server-idle-timeout",
		cfg.HTTPServerIdleTimeout,
		"HTTP server idle timeout",
	)
	flag.DurationVar(
		&cfg.HTTPServerReadTimeout,
		"http-server-read-timeout",
		cfg.HTTPServerReadTimeout,
		"HTTP server read timeout",
	)
	flag.DurationVar(
		&cfg.HTTPServerReadHeaderTimeout,
		"http-server-read-header-timeout",
		cfg.HTTPServerReadHeaderTimeout,
		"HTTP server read header timeout",
	)
	flag.DurationVar(
		&cfg.HTTPServerWriteTimeout,
		"http-server-write-timeout",
		cfg.HTTPServerWriteTimeout,
		"HTTP server write timeout",
	)
	flag.DurationVar(
		&cfg.GracefulShutdownTimeout,
		"http-graceful-shutdown-timeout",
		cfg.GracefulShutdownTimeout,
		"HTTP graceful shutdown timeout",
	)
	flag.IntVar(
		&cfg.DatabaseMaxConns,
		"database-max-conns",
		cfg.DatabaseMaxConns,
		"Database max conns",
	)
	flag.IntVar(
		&cfg.DatabaseMinConns,
		"database-min-conns",
		cfg.DatabaseMinConns,
		"Database min conns",
	)
	flag.DurationVar(
		&cfg.DatabaseConnMaxLifetime,
		"database-conn-max-lifetime",
		cfg.DatabaseConnMaxLifetime,
		"Database connection max lifetime",
	)
	flag.DurationVar(
		&cfg.DatabaseConnMaxIdleTime,
		"database-conn-max-idle-time",
		cfg.DatabaseConnMaxIdleTime,
		"Database connection max idle time",
	)
	flag.DurationVar(
		&cfg.DatabaseHealthCheckPeriod,
		"database-health-check-period",
		cfg.DatabaseHealthCheckPeriod,
		"Database health check period",
	)
	flag.Parse()

	return cfg, nil
}
