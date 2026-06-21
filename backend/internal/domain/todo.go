package domain

import (
	"context"
	"time"
)

// Todo represents a single to-do item.
type Todo struct {
	ID        int64
	Title     string
	Done      bool
	CreatedAt time.Time
}

// TodoStore defines the persistence contract for Todo entities.
// All concrete implementations (e.g. SQLiteStore) must satisfy this interface.
type TodoStore interface {
	Create(ctx context.Context, todo *Todo) error
	GetAll(ctx context.Context) ([]*Todo, error)
	GetByID(ctx context.Context, id int64) (*Todo, error)
	Update(ctx context.Context, todo *Todo) error
	Delete(ctx context.Context, id int64) error
}
