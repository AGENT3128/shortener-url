package main

import (
	"fmt"
	"os"

	"github.com/AGENT3128/shortener-url/internal/app"
	"github.com/AGENT3128/shortener-url/internal/config"
)

var (
	buildVersion = "N/A" //nolint:gochecknoglobals // build version - tag or branch name
	buildCommit  = "N/A" //nolint:gochecknoglobals // commit hash
	buildDate    = "N/A" //nolint:gochecknoglobals // build date
)

func main() {
	_, _ = fmt.Fprintf(os.Stdout, "Build version: %s\n", buildVersion)
	_, _ = fmt.Fprintf(os.Stdout, "Build date: %s\n", buildDate)
	_, _ = fmt.Fprintf(os.Stdout, "Build commit: %s\n", buildCommit)

	cfg, err := config.NewConfig()
	if err != nil {
		panic(err)
	}

	if errApp := app.Run(cfg); errApp != nil {
		panic(errApp)
	}
}
