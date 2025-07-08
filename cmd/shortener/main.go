package main

import (
	"fmt"

	"github.com/AGENT3128/shortener-url/internal/app"
	"github.com/AGENT3128/shortener-url/internal/config"
)

var (
	buildVersion = "N/A"
	buildCommit  = "N/A"
	buildDate    = "N/A"
)

func main() {
	fmt.Printf("Build version: %s\n", buildVersion)
	fmt.Printf("Build date: %s\n", buildDate)
	fmt.Printf("Build commit: %s\n", buildCommit)

	cfg, err := config.NewConfig()
	if err != nil {
		panic(err)
	}

	if errApp := app.Run(cfg); errApp != nil {
		panic(errApp)
	}
}
