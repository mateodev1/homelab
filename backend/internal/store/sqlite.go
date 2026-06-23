package store

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/mateo/homelab/backend/internal/domain"
)

// SQLiteStore is a domain.TodoStore backed by a SQLite database.
type SQLiteStore struct {
	db *sql.DB
}

// New creates a new SQLiteStore using the provided database connection.
// Call store.Migrate(db) before New to ensure the schema exists.
func New(db *sql.DB) *SQLiteStore {
	return &SQLiteStore{db: db}
}

// Ensure SQLiteStore satisfies the domain.TodoStore interface at compile time.
var _ domain.TodoStore = (*SQLiteStore)(nil)

// Create inserts a new Todo row and sets todo.ID to the new row's ID.
func (s *SQLiteStore) Create(ctx context.Context, todo *domain.Todo) error {
	const q = `INSERT INTO todos (title, body, color, pinned, done, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?)`

	now := todo.CreatedAt.UTC().Format(time.RFC3339)
	res, err := s.db.ExecContext(ctx, q,
		todo.Title, todo.Body, todo.Color, boolToInt(todo.Pinned),
		boolToInt(todo.Done), now, now,
	)
	if err != nil {
		return fmt.Errorf("store.Create: %w", err)
	}

	id, err := res.LastInsertId()
	if err != nil {
		return fmt.Errorf("store.Create LastInsertId: %w", err)
	}

	todo.ID = id
	todo.UpdatedAt = todo.CreatedAt
	return nil
}

// GetAll returns all Todo rows ordered by pinned DESC, id ASC.
func (s *SQLiteStore) GetAll(ctx context.Context) ([]*domain.Todo, error) {
	const q = `SELECT id, title, body, color, pinned, done, created_at, updated_at FROM todos ORDER BY pinned DESC, id ASC`

	rows, err := s.db.QueryContext(ctx, q)
	if err != nil {
		return nil, fmt.Errorf("store.GetAll: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var todos []*domain.Todo
	for rows.Next() {
		todo, err := scanTodo(rows)
		if err != nil {
			return nil, fmt.Errorf("store.GetAll scan: %w", err)
		}
		todos = append(todos, todo)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("store.GetAll rows: %w", err)
	}

	if todos == nil {
		todos = []*domain.Todo{}
	}
	return todos, nil
}

// GetByID returns the Todo with the given ID, or an error if not found.
func (s *SQLiteStore) GetByID(ctx context.Context, id int64) (*domain.Todo, error) {
	const q = `SELECT id, title, body, color, pinned, done, created_at, updated_at FROM todos WHERE id = ?`

	row := s.db.QueryRowContext(ctx, q, id)
	todo, err := scanTodoRow(row)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("get todo: %w", domain.ErrNotFound)
		}
		return nil, fmt.Errorf("store.GetByID(%d): %w", id, err)
	}
	return todo, nil
}

// Update persists changes to an existing Todo row.
func (s *SQLiteStore) Update(ctx context.Context, todo *domain.Todo) error {
	const q = `UPDATE todos SET title = ?, body = ?, color = ?, pinned = ?, done = ?, updated_at = ? WHERE id = ?`

	todo.UpdatedAt = time.Now().UTC()
	res, err := s.db.ExecContext(ctx, q,
		todo.Title, todo.Body, todo.Color, boolToInt(todo.Pinned),
		boolToInt(todo.Done), todo.UpdatedAt.Format(time.RFC3339), todo.ID,
	)
	if err != nil {
		return fmt.Errorf("store.Update: %w", err)
	}

	n, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("store.Update RowsAffected: %w", err)
	}
	if n == 0 {
		return fmt.Errorf("update todo: %w", domain.ErrNotFound)
	}
	return nil
}

// Delete removes the Todo with the given ID.
func (s *SQLiteStore) Delete(ctx context.Context, id int64) error {
	const q = `DELETE FROM todos WHERE id = ?`

	res, err := s.db.ExecContext(ctx, q, id)
	if err != nil {
		return fmt.Errorf("store.Delete: %w", err)
	}

	n, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("store.Delete RowsAffected: %w", err)
	}
	if n == 0 {
		return fmt.Errorf("delete todo: %w", domain.ErrNotFound)
	}
	return nil
}

// --- helpers ---

func boolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
}

type scanner interface {
	Scan(dest ...any) error
}

func scanTodo(s scanner) (*domain.Todo, error) {
	var (
		todo      domain.Todo
		pinned    int
		done      int
		createdAt string
		updatedAt string
	)
	if err := s.Scan(&todo.ID, &todo.Title, &todo.Body, &todo.Color, &pinned, &done, &createdAt, &updatedAt); err != nil {
		return nil, err
	}
	todo.Pinned = pinned != 0
	todo.Done = done != 0
	ca, err := time.Parse(time.RFC3339, createdAt)
	if err != nil {
		return nil, fmt.Errorf("parse created_at %q: %w", createdAt, err)
	}
	todo.CreatedAt = ca
	if updatedAt != "" {
		ua, err := time.Parse(time.RFC3339, updatedAt)
		if err != nil {
			return nil, fmt.Errorf("parse updated_at %q: %w", updatedAt, err)
		}
		todo.UpdatedAt = ua
	} else {
		todo.UpdatedAt = ca
	}
	return &todo, nil
}

func scanTodoRow(row *sql.Row) (*domain.Todo, error) {
	var (
		todo      domain.Todo
		pinned    int
		done      int
		createdAt string
		updatedAt string
	)
	if err := row.Scan(&todo.ID, &todo.Title, &todo.Body, &todo.Color, &pinned, &done, &createdAt, &updatedAt); err != nil {
		return nil, err
	}
	todo.Pinned = pinned != 0
	todo.Done = done != 0
	ca, err := time.Parse(time.RFC3339, createdAt)
	if err != nil {
		return nil, fmt.Errorf("parse created_at %q: %w", createdAt, err)
	}
	todo.CreatedAt = ca
	if updatedAt != "" {
		ua, err := time.Parse(time.RFC3339, updatedAt)
		if err != nil {
			return nil, fmt.Errorf("parse updated_at %q: %w", updatedAt, err)
		}
		todo.UpdatedAt = ua
	} else {
		todo.UpdatedAt = ca
	}
	return &todo, nil
}
