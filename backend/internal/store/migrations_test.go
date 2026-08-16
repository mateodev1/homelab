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

func TestMigration_SeedsProductsFromExistingProjects(t *testing.T) {
	t.Parallel()

	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("sql.Open: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })

	// First migrate to create the projects table, then insert a pre-existing
	// todo project the way an older deployment would have it.
	if err := store.Migrate(db); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	if _, err := db.Exec(`INSERT INTO projects (name, color, created_at) VALUES ('Homelab', 'blue', '2026-01-01T00:00:00Z')`); err != nil {
		t.Fatalf("insert legacy project: %v", err)
	}

	// Re-run migrate: it must seed a product named 'Homelab' idempotently.
	if err := store.Migrate(db); err != nil {
		t.Fatalf("second migrate: %v", err)
	}
	if err := store.Migrate(db); err != nil {
		t.Fatalf("third migrate: %v", err)
	}

	var count int
	if err := db.QueryRow(`SELECT COUNT(*) FROM products WHERE name = 'Homelab'`).Scan(&count); err != nil {
		t.Fatalf("query products: %v", err)
	}
	if count != 1 {
		t.Fatalf("expected exactly one seeded product 'Homelab', got %d", count)
	}
}

func TestMigration_SecretTablesExist(t *testing.T) {
	t.Parallel()

	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("sql.Open: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })

	if err := store.Migrate(db); err != nil {
		t.Fatalf("migrate: %v", err)
	}

	tables := []string{"products", "secret_projects", "secret_environments", "secrets", "secret_audit_logs"}
	for _, table := range tables {
		var name string
		err := db.QueryRow(`SELECT name FROM sqlite_master WHERE type='table' AND name = ?`, table).Scan(&name)
		if err != nil {
			t.Fatalf("expected table %q to exist: %v", table, err)
		}
	}

	// projects (todo hierarchy) must remain untouched by the secret migration.
	var name string
	if err := db.QueryRow(`SELECT name FROM sqlite_master WHERE type='table' AND name = 'projects'`).Scan(&name); err != nil {
		t.Fatalf("expected existing 'projects' table to remain: %v", err)
	}
}
