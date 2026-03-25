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
	if err := os.WriteFile(filepath.Join(dir, "998_not_applied.down.sql"), []byte(downSQL), 0644); err != nil {
		t.Fatalf("writing migration file: %v", err)
	}

	// Version 998_not_applied was never applied — MigrateDown should be a no-op.
	_, _ = db.Exec("DELETE FROM schema_migrations WHERE version = '998_not_applied'")

	if err := database.MigrateDown(db, dir); err != nil {
		t.Fatalf("MigrateDown() for unapplied version error: %v", err)
	}
}

func TestMigrateUp_AlreadyApplied_IsNoOp(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	db, err := database.Connect(context.Background(), testDatabaseURL(t))
	if err != nil {
		t.Fatalf("Connect() error: %v", err)
	}
	defer func() { _ = db.Close() }()

	dir := t.TempDir()
	upSQL := `CREATE TABLE IF NOT EXISTS test_already_applied (id SERIAL PRIMARY KEY);`
	if err := os.WriteFile(filepath.Join(dir, "995_already.up.sql"), []byte(upSQL), 0644); err != nil {
		t.Fatalf("writing migration file: %v", err)
	}

	_, _ = db.Exec("DROP TABLE IF EXISTS test_already_applied")
	_, _ = db.Exec("DELETE FROM schema_migrations WHERE version = '995_already'")

	// Apply once.
	if err := database.MigrateUp(db, dir); err != nil {
		t.Fatalf("first MigrateUp() error: %v", err)
	}

	// Apply again — already recorded, should be skipped without error.
	if err := database.MigrateUp(db, dir); err != nil {
		t.Fatalf("second MigrateUp() (idempotent) error: %v", err)
	}

	// Cleanup.
	_, _ = db.Exec("DROP TABLE IF EXISTS test_already_applied")
	_, _ = db.Exec("DELETE FROM schema_migrations WHERE version = '995_already'")
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
	if err := os.WriteFile(filepath.Join(dir, "997_bad.up.sql"), []byte(invalidSQL), 0644); err != nil {
		t.Fatalf("writing migration file: %v", err)
	}
	_, _ = db.Exec("DELETE FROM schema_migrations WHERE version = '997_bad'")

	if err := database.MigrateUp(db, dir); err == nil {
		t.Fatal("expected error for invalid SQL migration, got nil")
	}

	// Cleanup in case migration partially applied.
	_, _ = db.Exec("DELETE FROM schema_migrations WHERE version = '997_bad'")
}

func TestMigrateDown_InvalidSQL_ReturnsError(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	db, err := database.Connect(context.Background(), testDatabaseURL(t))
	if err != nil {
		t.Fatalf("Connect() error: %v", err)
	}
	defer func() { _ = db.Close() }()

	dir := t.TempDir()

	// Create matching up/down pair — apply the up first.
	upSQL := `CREATE TABLE IF NOT EXISTS test_down_invalid (id SERIAL PRIMARY KEY);`
	downSQL := `THIS IS NOT VALID SQL;`

	if err := os.WriteFile(filepath.Join(dir, "996_bad_down.up.sql"), []byte(upSQL), 0644); err != nil {
		t.Fatalf("writing up migration file: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, "996_bad_down.down.sql"), []byte(downSQL), 0644); err != nil {
		t.Fatalf("writing down migration file: %v", err)
	}

	// Clean state.
	_, _ = db.Exec("DROP TABLE IF EXISTS test_down_invalid")
	_, _ = db.Exec("DELETE FROM schema_migrations WHERE version = '996_bad_down'")

	// Apply the up migration so MigrateDown sees it as applied.
	if err := database.MigrateUp(db, dir); err != nil {
		t.Fatalf("MigrateUp() setup error: %v", err)
	}

	// MigrateDown with invalid SQL should return an error.
	if err := database.MigrateDown(db, dir); err == nil {
		t.Fatal("expected error for invalid down SQL, got nil")
	}

	// Cleanup.
	_, _ = db.Exec("DROP TABLE IF EXISTS test_down_invalid")
	_, _ = db.Exec("DELETE FROM schema_migrations WHERE version = '996_bad_down'")
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

	if err := os.WriteFile(filepath.Join(dir, "999_test.up.sql"), []byte(upSQL), 0644); err != nil {
		t.Fatalf("writing up migration file: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, "999_test.down.sql"), []byte(downSQL), 0644); err != nil {
		t.Fatalf("writing down migration file: %v", err)
	}

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

func TestMigrateUp_MultipleMigrations(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	db, err := database.Connect(context.Background(), testDatabaseURL(t))
	if err != nil {
		t.Fatalf("Connect() error: %v", err)
	}
	defer func() { _ = db.Close() }()

	dir := t.TempDir()

	// Write two sequential migrations.
	for _, f := range []struct {
		name, sql string
	}{
		{"991_multi_a.up.sql", "CREATE TABLE IF NOT EXISTS multi_a (id SERIAL PRIMARY KEY);"},
		{"992_multi_b.up.sql", "CREATE TABLE IF NOT EXISTS multi_b (id SERIAL PRIMARY KEY);"},
	} {
		if err := os.WriteFile(filepath.Join(dir, f.name), []byte(f.sql), 0644); err != nil {
			t.Fatalf("writing %s: %v", f.name, err)
		}
	}

	// Clean state.
	_, _ = db.Exec("DROP TABLE IF EXISTS multi_a, multi_b")
	_, _ = db.Exec("DELETE FROM schema_migrations WHERE version IN ('991_multi_a','992_multi_b')")

	if err := database.MigrateUp(db, dir); err != nil {
		t.Fatalf("MigrateUp() error: %v", err)
	}

	for _, table := range []string{"multi_a", "multi_b"} {
		var exists bool
		_ = db.Get(&exists,
			"SELECT EXISTS(SELECT 1 FROM information_schema.tables WHERE table_name = $1)", table)
		if !exists {
			t.Errorf("expected table %q to exist after migration", table)
		}
	}

	// Cleanup.
	_, _ = db.Exec("DROP TABLE IF EXISTS multi_a, multi_b")
	_, _ = db.Exec("DELETE FROM schema_migrations WHERE version IN ('991_multi_a','992_multi_b')")
}

func TestMigrateUp_NonExistentDir(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	db, err := database.Connect(context.Background(), testDatabaseURL(t))
	if err != nil {
		t.Fatalf("Connect() error: %v", err)
	}
	defer func() { _ = db.Close() }()

	// Non-existent directory — filepath.Glob returns empty, so MigrateUp is a no-op (not an error).
	err = database.MigrateUp(db, filepath.Join(t.TempDir(), "does-not-exist"))
	if err != nil {
		t.Errorf("MigrateUp() with non-existent dir error: %v", err)
	}
}
