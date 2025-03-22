package config

import (
	"flag"

	"github.com/caarlos0/env/v11"
)

type Config struct {
	ServerAddress  string `env:"SERVER_ADDRESS" envDefault:"localhost:8080"`  // server address
	BaseURLAddress string `env:"BASE_URL" envDefault:"http://localhost:8080"` // base url for shortened urls
	ReleaseMode    string `env:"RELEASE_MODE" envDefault:"debug"`             // release mode. Available options: debug, release, test
}

func NewConfig() *Config {
	// Create new config instance
	cfg := &Config{}

	// Parse environment variables
	if err := env.Parse(cfg); err != nil {
		panic(err)
	}

	// Define and parse command line flags
	flag.StringVar(&cfg.ServerAddress, "a", cfg.ServerAddress, "HTTP server address")
	flag.StringVar(&cfg.BaseURLAddress, "b", cfg.BaseURLAddress, "Base URL for shortened URLs")
	flag.Parse()

	return cfg
}
