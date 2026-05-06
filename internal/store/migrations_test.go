package store

import (
	"database/sql"
	"path/filepath"
	"testing"
)

func TestOpenInMemorySQLiteDatabase(t *testing.T) {
	db, err := Open(":memory:")
	if err != nil {
		t.Fatalf("open in-memory database: %v", err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		t.Fatalf("ping in-memory database: %v", err)
	}
}

func TestOpenEnablesForeignKeys(t *testing.T) {
	db := openTestDB(t)

	var enabled int
	if err := db.QueryRow(`PRAGMA foreign_keys`).Scan(&enabled); err != nil {
		t.Fatalf("read foreign_keys pragma: %v", err)
	}
	if enabled != 1 {
		t.Fatalf("foreign_keys = %d, want 1", enabled)
	}
}

func TestRunMigrationsOnce(t *testing.T) {
	db := openTestDB(t)

	if err := RunMigrations(db); err != nil {
		t.Fatalf("run migrations: %v", err)
	}

	if count := migrationCount(t, db); count != 4 {
		t.Fatalf("migration count = %d, want 4", count)
	}

	var tableName string
	err := db.QueryRow(`SELECT name FROM sqlite_master WHERE type = 'table' AND name = 'organizations'`).Scan(&tableName)
	if err != nil {
		t.Fatalf("find organizations table: %v", err)
	}
	if tableName != "organizations" {
		t.Fatalf("table name = %q, want organizations", tableName)
	}
}

func TestRunMigrationsTwiceDoesNotDuplicate(t *testing.T) {
	db := openTestDB(t)

	if err := RunMigrations(db); err != nil {
		t.Fatalf("run migrations first time: %v", err)
	}
	if err := RunMigrations(db); err != nil {
		t.Fatalf("run migrations second time: %v", err)
	}

	if count := migrationCount(t, db); count != 4 {
		t.Fatalf("migration count = %d, want 4", count)
	}
}

func openTestDB(t *testing.T) *sql.DB {
	t.Helper()

	db, err := Open(filepath.Join(t.TempDir(), "test.sqlite"))
	if err != nil {
		t.Fatalf("open test database: %v", err)
	}
	t.Cleanup(func() {
		db.Close()
	})

	return db
}

func migrationCount(t *testing.T, db *sql.DB) int {
	t.Helper()

	var count int
	if err := db.QueryRow(`SELECT COUNT(*) FROM schema_migrations`).Scan(&count); err != nil {
		t.Fatalf("count migrations: %v", err)
	}

	return count
}
