package database_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/cenron/foundry/internal/database"
)

func testDatabaseURL(t *testing.T) string {
	t.Helper()
	url := os.Getenv("DATABASE_URL")
	if url == "" {
		url = "postgres://foundry:foundry@localhost:5433/foundry?sslmode=disable"
	}
	return url
}

func TestConnect(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	db, err := database.Connect(context.Background(), testDatabaseURL(t))
	if err != nil {
		t.Fatalf("Connect() error: %v", err)
	}
	defer func() { _ = db.Close() }()

	var result int
	if err := db.Get(&result, "SELECT 1"); err != nil {
		t.Fatalf("query failed: %v", err)
	}
	if result != 1 {
		t.Errorf("got %d, want 1", result)
	}
}

func TestConnect_InvalidURL(t *testing.T) {
	_, err := database.Connect(context.Background(), "not-a-valid-url")
	if err == nil {
		t.Fatal("expected error for invalid DB URL, got nil")
	}
}

func TestMigrateDown_EmptyDir(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	db, err := database.Connect(context.Background(), testDatabaseURL(t))
	if err != nil {
		t.Fatalf("Connect() error: %v", err)
	}
	defer func() { _ = db.Close() }()

	dir := t.TempDir() // empty dir — no migrations found

	// Should return nil without error when no down migrations exist.
	if err := database.MigrateDown(db, dir); err != nil {
		t.Fatalf("MigrateDown() with empty dir error: %v", err)
	}
}

func TestMigrateDown_VersionNotApplied(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	db, err := database.Connect(context.Background(), testDatabaseURL(t))
	if err != nil {
		t.Fatalf("Connect() error: %v", err)
	}
	defer func() { _ = db.Close() }()

	dir := t.TempDir()
	downSQL := `-- no-op rollback`
	_ = os.WriteFile(filepath.Join(dir, "998_not_applied.down.sql"), []byte(downSQL), 0644)

	// Version 998_not_applied was never applied — MigrateDown should be a no-op.
	_, _ = db.Exec("DELETE FROM schema_migrations WHERE version = '998_not_applied'")

	if err := database.MigrateDown(db, dir); err != nil {
		t.Fatalf("MigrateDown() for unapplied version error: %v", err)
	}
}

func TestMigrateUp_InvalidSQL_ReturnsError(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	db, err := database.Connect(context.Background(), testDatabaseURL(t))
	if err != nil {
		t.Fatalf("Connect() error: %v", err)
	}
	defer func() { _ = db.Close() }()

	dir := t.TempDir()
	invalidSQL := `THIS IS NOT VALID SQL;`
	_ = os.WriteFile(filepath.Join(dir, "997_bad.up.sql"), []byte(invalidSQL), 0644)
	_, _ = db.Exec("DELETE FROM schema_migrations WHERE version = '997_bad'")

	if err := database.MigrateUp(db, dir); err == nil {
		t.Fatal("expected error for invalid SQL migration, got nil")
	}

	// Cleanup in case migration partially applied.
	_, _ = db.Exec("DELETE FROM schema_migrations WHERE version = '997_bad'")
}

func TestMigrateUp_And_Down(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	db, err := database.Connect(context.Background(), testDatabaseURL(t))
	if err != nil {
		t.Fatalf("Connect() error: %v", err)
	}
	defer func() { _ = db.Close() }()

	// Clean up from previous runs
	_, _ = db.Exec("DROP TABLE IF EXISTS test_items")
	_, _ = db.Exec("DELETE FROM schema_migrations WHERE version LIKE '999%'")

	dir := t.TempDir()

	upSQL := `CREATE TABLE test_items (id SERIAL PRIMARY KEY, name TEXT NOT NULL);`
	downSQL := `DROP TABLE IF EXISTS test_items;`

	_ = os.WriteFile(filepath.Join(dir, "999_test.up.sql"), []byte(upSQL), 0644)
	_ = os.WriteFile(filepath.Join(dir, "999_test.down.sql"), []byte(downSQL), 0644)

	// Migrate up
	if err := database.MigrateUp(db, dir); err != nil {
		t.Fatalf("MigrateUp() error: %v", err)
	}

	// Verify table exists
	var exists bool
	err = db.Get(&exists, "SELECT EXISTS(SELECT 1 FROM information_schema.tables WHERE table_name = 'test_items')")
	if err != nil {
		t.Fatalf("checking table: %v", err)
	}
	if !exists {
		t.Fatal("test_items table should exist after migration up")
	}

	// Idempotent — running again should not error
	if err := database.MigrateUp(db, dir); err != nil {
		t.Fatalf("MigrateUp() idempotent error: %v", err)
	}

	// Migrate down
	if err := database.MigrateDown(db, dir); err != nil {
		t.Fatalf("MigrateDown() error: %v", err)
	}

	// Verify table removed
	err = db.Get(&exists, "SELECT EXISTS(SELECT 1 FROM information_schema.tables WHERE table_name = 'test_items')")
	if err != nil {
		t.Fatalf("checking table after down: %v", err)
	}
	if exists {
		t.Fatal("test_items table should not exist after migration down")
	}
}
