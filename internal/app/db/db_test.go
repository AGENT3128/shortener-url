package db

import (
	"testing"
)

func TestMigration(t *testing.T) {
	dsn := "postgres://postgres:postgres@localhost:5432/postgres?sslmode=disable"

	db, err := NewDatabase(dsn)
	if err != nil {
		t.Fatalf("Failed to create database: %v", err)
	}
	defer db.Conn.Close()

	err = db.Migrate()
	if err != nil {
		t.Fatalf("Failed to migrate: %v", err)
	}
}
