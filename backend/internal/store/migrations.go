package store

import (
	"database/sql"
	"strings"
)

// Migrate applies all schema migrations to the provided database.
// It is idempotent: safe to call on an already-migrated database.
func Migrate(db *sql.DB) error {
	const schema = `
CREATE TABLE IF NOT EXISTS todos (
	id         INTEGER PRIMARY KEY AUTOINCREMENT,
	title      TEXT    NOT NULL,
	done       INTEGER NOT NULL DEFAULT 0,
	created_at TEXT    NOT NULL
);`

	const projectsSchema = `
CREATE TABLE IF NOT EXISTS projects (
	id         INTEGER PRIMARY KEY AUTOINCREMENT,
	name       TEXT    NOT NULL UNIQUE,
	color      TEXT    NOT NULL DEFAULT 'default',
	created_at TEXT    NOT NULL
);`

	if _, err := db.Exec(schema); err != nil {
		return err
	}
	if _, err := db.Exec(projectsSchema); err != nil {
		return err
	}

	alterations := []string{
		`ALTER TABLE todos ADD COLUMN body       TEXT    NOT NULL DEFAULT ''`,
		`ALTER TABLE todos ADD COLUMN color      TEXT    NOT NULL DEFAULT 'default'`,
		`ALTER TABLE todos ADD COLUMN pinned     INTEGER NOT NULL DEFAULT 0`,
		`ALTER TABLE todos ADD COLUMN updated_at TEXT    NOT NULL DEFAULT ''`,
		`ALTER TABLE todos ADD COLUMN status     TEXT    NOT NULL DEFAULT 'todo'`,
		`ALTER TABLE todos ADD COLUMN priority   INTEGER NOT NULL DEFAULT 0`,
		`ALTER TABLE todos ADD COLUMN due_date   TEXT    NULL`,
		`ALTER TABLE todos ADD COLUMN kind       TEXT    NOT NULL DEFAULT 'note'`,
		`ALTER TABLE todos ADD COLUMN issue_type TEXT    NULL`,
		`ALTER TABLE todos ADD COLUMN project_id INTEGER NULL`,
	}

	for _, stmt := range alterations {
		if _, err := db.Exec(stmt); err != nil {
			if !strings.Contains(err.Error(), "duplicate column name") {
				return err
			}
		}
	}

	if _, err := db.Exec(`UPDATE todos SET status = 'done' WHERE done = 1 AND status = 'todo'`); err != nil {
		return err
	}

	if err := migrateSecretsSchema(db); err != nil {
		return err
	}

	return nil
}

// migrateSecretsSchema creates the secret-manager hierarchy
// (Product -> SecretProject -> SecretEnvironment -> Secret) as tables
// distinct from the existing todo `projects` table so that hierarchy is
// never touched. It is idempotent and non-destructive: it never drops or
// rewrites existing data, and it seeds `products` from existing `projects`
// names only when a product with that name does not already exist.
func migrateSecretsSchema(db *sql.DB) error {
	const schema = `
CREATE TABLE IF NOT EXISTS products (
	id         INTEGER PRIMARY KEY AUTOINCREMENT,
	name       TEXT    NOT NULL UNIQUE,
	created_at TEXT    NOT NULL
);

CREATE TABLE IF NOT EXISTS secret_projects (
	id         INTEGER PRIMARY KEY AUTOINCREMENT,
	product_id INTEGER NOT NULL REFERENCES products(id),
	name       TEXT    NOT NULL,
	created_at TEXT    NOT NULL,
	UNIQUE(product_id, name)
);

CREATE TABLE IF NOT EXISTS secret_environments (
	id         INTEGER PRIMARY KEY AUTOINCREMENT,
	project_id INTEGER NOT NULL REFERENCES secret_projects(id),
	name       TEXT    NOT NULL,
	created_at TEXT    NOT NULL,
	UNIQUE(project_id, name)
);

CREATE TABLE IF NOT EXISTS secrets (
	id              INTEGER PRIMARY KEY AUTOINCREMENT,
	environment_id  INTEGER NOT NULL REFERENCES secret_environments(id),
	key             TEXT    NOT NULL,
	value_encrypted TEXT    NOT NULL,
	created_at      TEXT    NOT NULL,
	updated_at      TEXT    NOT NULL,
	UNIQUE(environment_id, key)
);

CREATE TABLE IF NOT EXISTS secret_audit_logs (
	id             INTEGER PRIMARY KEY AUTOINCREMENT,
	environment_id INTEGER NOT NULL REFERENCES secret_environments(id),
	secret_key     TEXT    NOT NULL,
	action         TEXT    NOT NULL,
	actor          TEXT    NOT NULL,
	created_at     TEXT    NOT NULL
);`

	if _, err := db.Exec(schema); err != nil {
		return err
	}

	// Seed products from existing todo projects (best-effort, non-destructive):
	// only insert names that don't already exist as a product.
	const seed = `
INSERT INTO products (name, created_at)
SELECT p.name, p.created_at
FROM projects p
WHERE NOT EXISTS (SELECT 1 FROM products WHERE products.name = p.name)`
	if _, err := db.Exec(seed); err != nil {
		return err
	}

	return nil
}
