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

	if _, err := db.Exec(schema); err != nil {
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

	return nil
}
