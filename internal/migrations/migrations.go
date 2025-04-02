package migrations

import (
	"embed"
)

//go:embed *.sql
var EmbedMigrations embed.FS

// GetMigrationsFS returns the embedded FS with migrations
func GetMigrationsFS() embed.FS {
	return EmbedMigrations
}

// MigrationsPath returns the path to migrations inside the embedded FS
func MigrationsPath() string {
	return "."
}
