package store_test

import (
	"database/sql"
	"testing"

	"github.com/mateo/homelab/backend/internal/store"

	_ "modernc.org/sqlite"
)

func TestMigration_Idempotent(t *testing.T) {
	t.Parallel()

	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("sql.Open: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })

	if err := store.Migrate(db); err != nil {
		t.Fatalf("first migrate: %v", err)
	}
	if err := store.Migrate(db); err != nil {
		t.Fatalf("second migrate should be no-op safe, got: %v", err)
	}
}

func TestMigration_DoneToStatusDataUpdate(t *testing.T) {
	t.Parallel()

	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("sql.Open: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })

	if _, err := db.Exec(`
		CREATE TABLE todos (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			title TEXT NOT NULL,
			done INTEGER NOT NULL DEFAULT 0,
			created_at TEXT NOT NULL
		)
	`); err != nil {
		t.Fatalf("create legacy todos: %v", err)
	}

	if _, err := db.Exec(`INSERT INTO todos (title, done, created_at) VALUES ('legacy done', 1, '2026-06-01T10:00:00Z')`); err != nil {
		t.Fatalf("insert legacy row: %v", err)
	}

	if err := store.Migrate(db); err != nil {
		t.Fatalf("migrate: %v", err)
	}

	var status string
	if err := db.QueryRow(`SELECT status FROM todos LIMIT 1`).Scan(&status); err != nil {
		t.Fatalf("query status: %v", err)
	}
	if status != "done" {
		t.Fatalf("expected migrated status done, got %q", status)
	}
}
