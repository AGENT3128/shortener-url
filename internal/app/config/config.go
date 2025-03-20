package config

import (
	"flag"

	"github.com/caarlos0/env/v11"
)

var Config *config

type config struct {
	ServerAddress  string `env:"SERVER_ADDRESS" envDefault:"localhost:8080"`  // server address
	BaseURLAddress string `env:"BASE_URL" envDefault:"http://localhost:8080"` // base url for shortened urls
}

func InitConfig() {
	// Create new config instance
	cfg := &config{}

	// Parse environment variables
	if err := env.Parse(cfg); err != nil {
		panic(err)
	}

	// Define and parse command line flags
	flag.StringVar(&cfg.ServerAddress, "a", cfg.ServerAddress, "HTTP server address")
	flag.StringVar(&cfg.BaseURLAddress, "b", cfg.BaseURLAddress, "Base URL for shortened URLs")
	flag.Parse()

	// Assign to global Config variable
	Config = cfg
}
