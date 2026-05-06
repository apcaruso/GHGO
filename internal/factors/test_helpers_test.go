package factors

import (
	"path/filepath"
	"testing"

	"ghgo/internal/store"
)

func newTestStore(t *testing.T) *store.Store {
	t.Helper()

	db, err := store.Open(filepath.Join(t.TempDir(), "factors.sqlite"))
	if err != nil {
		t.Fatalf("open test database: %v", err)
	}
	t.Cleanup(func() {
		db.Close()
	})
	if err := store.RunMigrations(db); err != nil {
		t.Fatalf("run migrations: %v", err)
	}

	return store.New(db)
}
