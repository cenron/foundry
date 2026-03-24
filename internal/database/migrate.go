package database

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/jmoiron/sqlx"
)

const createMigrationsTable = `
CREATE TABLE IF NOT EXISTS schema_migrations (
	version TEXT PRIMARY KEY,
	applied_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);`

func MigrateUp(db *sqlx.DB, migrationsDir string) error {
	if _, err := db.Exec(createMigrationsTable); err != nil {
		return fmt.Errorf("creating migrations table: %w", err)
	}

	files, err := findMigrations(migrationsDir, "up")
	if err != nil {
		return err
	}

	for _, f := range files {
		version := extractVersion(f)

		var exists bool
		if err := db.Get(&exists, "SELECT EXISTS(SELECT 1 FROM schema_migrations WHERE version = $1)", version); err != nil {
			return fmt.Errorf("checking migration %s: %w", version, err)
		}
		if exists {
			continue
		}

		sql, err := os.ReadFile(f)
		if err != nil {
			return fmt.Errorf("reading migration %s: %w", f, err)
		}

		tx, err := db.Beginx()
		if err != nil {
			return fmt.Errorf("beginning transaction for %s: %w", version, err)
		}

		if _, err := tx.Exec(string(sql)); err != nil {
			_ = tx.Rollback()
			return fmt.Errorf("executing migration %s: %w", version, err)
		}

		if _, err := tx.Exec("INSERT INTO schema_migrations (version) VALUES ($1)", version); err != nil {
			_ = tx.Rollback()
			return fmt.Errorf("recording migration %s: %w", version, err)
		}

		if err := tx.Commit(); err != nil {
			return fmt.Errorf("committing migration %s: %w", version, err)
		}
	}

	return nil
}

func MigrateDown(db *sqlx.DB, migrationsDir string) error {
	if _, err := db.Exec(createMigrationsTable); err != nil {
		return fmt.Errorf("creating migrations table: %w", err)
	}

	files, err := findMigrations(migrationsDir, "down")
	if err != nil {
		return err
	}
	if len(files) == 0 {
		return nil
	}

	// Sort descending — rollback the latest first
	sort.Sort(sort.Reverse(sort.StringSlice(files)))
	latest := files[0]
	version := extractVersion(latest)

	var exists bool
	if err := db.Get(&exists, "SELECT EXISTS(SELECT 1 FROM schema_migrations WHERE version = $1)", version); err != nil {
		return fmt.Errorf("checking migration %s: %w", version, err)
	}
	if !exists {
		return nil
	}

	sql, err := os.ReadFile(latest)
	if err != nil {
		return fmt.Errorf("reading migration %s: %w", latest, err)
	}

	tx, err := db.Beginx()
	if err != nil {
		return fmt.Errorf("beginning transaction for %s: %w", version, err)
	}

	if _, err := tx.Exec(string(sql)); err != nil {
		_ = tx.Rollback()
		return fmt.Errorf("executing rollback %s: %w", version, err)
	}

	if _, err := tx.Exec("DELETE FROM schema_migrations WHERE version = $1", version); err != nil {
		_ = tx.Rollback()
		return fmt.Errorf("removing migration record %s: %w", version, err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("committing rollback %s: %w", version, err)
	}

	return nil
}

func findMigrations(dir, direction string) ([]string, error) {
	pattern := filepath.Join(dir, fmt.Sprintf("*.%s.sql", direction))
	files, err := filepath.Glob(pattern)
	if err != nil {
		return nil, fmt.Errorf("finding %s migrations: %w", direction, err)
	}
	sort.Strings(files)
	return files, nil
}

func extractVersion(path string) string {
	base := filepath.Base(path)
	// Format: 001_name.up.sql → version is "001_name"
	parts := strings.SplitN(base, ".", 2)
	return parts[0]
}
