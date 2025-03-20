package config

import "flag"

type Config struct {
	ServerAddress  string // server address
	BaseURLAddress string // base url for shortened urls
}

func NewConfig() *Config {
	cfg := &Config{}

	flag.StringVar(&cfg.ServerAddress, "a", "localhost:8080", "HTTP server address")
	flag.StringVar(&cfg.BaseURLAddress, "b", "http://localhost:8080", "Base URL for shortened URLs")
	flag.Parse()

	return cfg
}
