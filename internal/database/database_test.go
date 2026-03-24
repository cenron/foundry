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
