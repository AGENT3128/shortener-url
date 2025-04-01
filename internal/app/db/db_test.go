package db

import (
	"context"
	"testing"
	"time"
)

func TestMigration(t *testing.T) {
	ctx := context.Background()

	dsn := "postgres://postgres:postgres@localhost:5432/postgres?sslmode=disable"

	ctxDB, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	db, err := NewDatabase(ctxDB, dsn)
	if err != nil {
		t.Fatalf("Failed to create database: %v", err)
	}
	defer db.Conn.Close()

	ctxMigration, cancelMigration := context.WithTimeout(ctx, 1*time.Minute)
	defer cancelMigration()
	err = db.MigrateWithContext(ctxMigration)
	if err != nil {
		t.Fatalf("Failed to migrate: %v", err)
	}
}
