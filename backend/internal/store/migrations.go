package store

import "database/sql"

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

	_, err := db.Exec(schema)
	return err
}
