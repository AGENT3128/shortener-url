package config

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/caarlos0/env/v11"
)

// Config is the configuration for the application.
// Priority:
// 1. Command line flags
// 2. Environment variables
// 3. Config file.
type Config struct {
	ConfigPath                  string        `json:"-"                                         env:"CONFIG_PATH"                     envDefault:""`                      // config path
	BaseURLAddress              string        `json:"base_url,omitempty"                        env:"BASE_URL"                        envDefault:"http://localhost:8080"` // base url for shortened urls
	ReleaseMode                 string        `json:"release_mode,omitempty"                    env:"RELEASE_MODE"                    envDefault:"debug"`                 // release mode. Available options: debug, release, test
	LogLevel                    string        `json:"log_level,omitempty"                       env:"LOG_LEVEL"                       envDefault:"info"`                  // log level
	FileStoragePath             string        `json:"file_storage_path,omitempty"               env:"FILE_STORAGE_PATH"               envDefault:""`                      // file storage path
	DatabaseDSN                 string        `json:"database_dsn,omitempty"                    env:"DATABASE_DSN"                    envDefault:""`                      // database dsn
	HTTPServerAddress           string        `json:"http_server_address,omitempty"             env:"HTTP_SERVER_ADDRESS"             envDefault:"localhost:8080"`        // http server address
	TLSCertPath                 string        `json:"tls_cert_path,omitempty"                   env:"TLS_CERT_PATH"                   envDefault:""`                      // tls cert path
	TLSKeyPath                  string        `json:"tls_key_path,omitempty"                    env:"TLS_KEY_PATH"                    envDefault:""`                      // tls key path
	TrustedSubnet               string        `json:"trusted_subnet,omitempty"                  env:"TRUSTED_SUBNET"                  envDefault:""`                      // trusted subnet
	DatabaseMaxConns            int           `json:"database_max_conns,omitempty"              env:"DATABASE_MAX_CONNS"              envDefault:"10"`                    // database max conns
	DatabaseMinConns            int           `json:"database_min_conns,omitempty"              env:"DATABASE_MIN_CONNS"              envDefault:"2"`                     // database min conns
	DatabaseConnMaxLifetime     time.Duration `json:"database_conn_max_lifetime,omitempty"      env:"DATABASE_CONN_MAX_LIFETIME"      envDefault:"10s"`                   // database connection max lifetime
	DatabaseConnMaxIdleTime     time.Duration `json:"database_conn_max_idle_time,omitempty"     env:"DATABASE_CONN_MAX_IDLE_TIME"     envDefault:"10s"`                   // database connection max idle time
	DatabaseHealthCheckPeriod   time.Duration `json:"database_health_check_period,omitempty"    env:"DATABASE_HEALTH_CHECK_PERIOD"    envDefault:"10s"`                   // database health check period
	HTTPServerIdleTimeout       time.Duration `json:"http_server_idle_timeout,omitempty"        env:"HTTP_SERVER_IDLE_TIMEOUT"        envDefault:"30s"`                   // http server idle timeout
	HTTPServerReadTimeout       time.Duration `json:"http_server_read_timeout,omitempty"        env:"HTTP_SERVER_READ_TIMEOUT"        envDefault:"15s"`                   // http server read timeout
	HTTPServerReadHeaderTimeout time.Duration `json:"http_server_read_header_timeout,omitempty" env:"HTTP_SERVER_READ_HEADER_TIMEOUT" envDefault:"15s"`                   // http server read header timeout
	HTTPServerWriteTimeout      time.Duration `json:"http_server_write_timeout,omitempty"       env:"HTTP_SERVER_WRITE_TIMEOUT"       envDefault:"10s"`                   // http server write timeout
	GracefulShutdownTimeout     time.Duration `json:"graceful_shutdown_timeout,omitempty"       env:"GRACEFUL_SHUTDOWN_TIMEOUT"       envDefault:"20s"`                   // graceful shutdown timeout
	EnableHTTPS                 bool          `json:"enable_https,omitempty"                    env:"ENABLE_HTTPS"                    envDefault:""`                      // enable https
}

// NewConfig creates a new Config instance.
func NewConfig() (*Config, error) {
	// Create new config instance
	cfg := &Config{}

	// Load config from file
	configPath, exists := os.LookupEnv("CONFIG_PATH")
	if exists && configPath != "" {
		if err := cfg.loadFromJSONFile(configPath); err != nil {
			return nil, fmt.Errorf("failed to load config from file: %w", err)
		}
	}

	for i, arg := range os.Args {
		if arg == "-c" || arg == "--config" {
			if err := cfg.loadFromJSONFile(os.Args[i+1]); err != nil {
				return nil, fmt.Errorf("failed to load config from file: %w", err)
			}
		}
	}
	// Parse environment variables
	if err := env.Parse(cfg); err != nil {
		return nil, fmt.Errorf("failed to parse environment variables: %w", err)
	}

	parseFlags(cfg)
	return cfg, nil
}

// parseFlags parses command line flags.
func parseFlags(cfg *Config) {
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
	flag.IntVar(&cfg.DatabaseMaxConns, "database-max-conns", cfg.DatabaseMaxConns, "Database max conns")
	flag.IntVar(&cfg.DatabaseMinConns, "database-min-conns", cfg.DatabaseMinConns, "Database min conns")
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
	flag.BoolVar(&cfg.EnableHTTPS, "enable-https", cfg.EnableHTTPS, "Enable HTTPS")
	flag.StringVar(&cfg.TLSCertPath, "tls-cert-path", cfg.TLSCertPath, "TLS cert path")
	flag.StringVar(&cfg.TLSKeyPath, "tls-key-path", cfg.TLSKeyPath, "TLS key path")
	flag.StringVar(&cfg.ConfigPath, "c", cfg.ConfigPath, "Path to config file")
	flag.StringVar(&cfg.ConfigPath, "config", cfg.ConfigPath, "Path to config file")
	flag.StringVar(&cfg.TrustedSubnet, "t", cfg.TrustedSubnet, "Trusted subnet")
	flag.Parse()
}

// loadFromJSONFile loads configuration from JSON file.
func (c *Config) loadFromJSONFile(filePath string) error {
	if filePath == "" {
		return nil
	}

	file, err := os.Open(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("config file %s does not exist", filePath)
		}
		return fmt.Errorf("failed to open config file %s: %w", filePath, err)
	}
	defer file.Close()

	decoder := json.NewDecoder(file)
	if errDecode := decoder.Decode(c); errDecode != nil {
		return fmt.Errorf("failed to decode config file %s: %w", filePath, errDecode)
	}

	return nil
}
