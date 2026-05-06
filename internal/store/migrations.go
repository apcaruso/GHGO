package store

import (
	"database/sql"
	"errors"
	"fmt"
	"io/fs"
	"sort"
	"strings"
	"time"

	"ghgo/migrations"
)

// RunMigrations applies embedded SQL migrations that have not been recorded yet.
func RunMigrations(db *sql.DB) error {
	files, err := migrationFiles()
	if err != nil {
		return err
	}

	applied, err := appliedMigrations(db)
	if err != nil {
		return err
	}

	for _, name := range files {
		if applied[name] {
			continue
		}

		if err := runMigration(db, name); err != nil {
			return err
		}

		applied[name] = true
	}

	return nil
}

func migrationFiles() ([]string, error) {
	entries, err := fs.ReadDir(migrations.Files, ".")
	if err != nil {
		return nil, fmt.Errorf("read embedded migrations: %w", err)
	}

	files := make([]string, 0, len(entries))
	for _, entry := range entries {
		name := entry.Name()
		if entry.IsDir() || !strings.HasSuffix(name, ".sql") {
			continue
		}

		files = append(files, name)
	}
	sort.Strings(files)

	return files, nil
}

func appliedMigrations(db *sql.DB) (map[string]bool, error) {
	applied := make(map[string]bool)

	var tableName string
	err := db.QueryRow(`SELECT name FROM sqlite_master WHERE type = 'table' AND name = 'schema_migrations'`).Scan(&tableName)
	if errors.Is(err, sql.ErrNoRows) {
		return applied, nil
	}
	if err != nil {
		return nil, fmt.Errorf("check schema_migrations table: %w", err)
	}

	rows, err := db.Query(`SELECT version FROM schema_migrations`)
	if err != nil {
		return nil, fmt.Errorf("read applied migrations: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var version string
		if err := rows.Scan(&version); err != nil {
			return nil, fmt.Errorf("scan applied migration: %w", err)
		}
		applied[version] = true
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("read applied migrations rows: %w", err)
	}

	return applied, nil
}

func runMigration(db *sql.DB, name string) error {
	contents, err := fs.ReadFile(migrations.Files, name)
	if err != nil {
		return fmt.Errorf("read migration %s: %w", name, err)
	}

	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("begin migration %s: %w", name, err)
	}

	if _, err := tx.Exec(string(contents)); err != nil {
		return rollback(tx, fmt.Errorf("run migration %s: %w", name, err))
	}

	if _, err := tx.Exec(
		`INSERT INTO schema_migrations (version, applied_at) VALUES (?, ?)`,
		name,
		time.Now().UTC().Format(time.RFC3339),
	); err != nil {
		return rollback(tx, fmt.Errorf("record migration %s: %w", name, err))
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit migration %s: %w", name, err)
	}

	return nil
}

func rollback(tx *sql.Tx, err error) error {
	if rollbackErr := tx.Rollback(); rollbackErr != nil {
		return fmt.Errorf("%w; rollback failed: %v", err, rollbackErr)
	}
	return err
}
