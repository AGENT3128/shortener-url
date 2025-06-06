package main

import (
	"github.com/AGENT3128/shortener-url/internal/app"
	"github.com/AGENT3128/shortener-url/internal/config"
)

func main() {
	cfg, err := config.NewConfig()
	if err != nil {
		panic(err)
	}

	if errApp := app.Run(cfg); errApp != nil {
		panic(errApp)
	}
}
