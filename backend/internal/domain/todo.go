package domain

import (
	"context"
	"errors"
	"time"
)

var ErrNotFound = errors.New("not found")

const (
	TodoStatusTodo       = "todo"
	TodoStatusInProgress = "in_progress"
	TodoStatusDone       = "done"
	TodoStatusCancelled  = "cancelled"
)

var ValidStatuses = map[string]bool{
	TodoStatusTodo:       true,
	TodoStatusInProgress: true,
	TodoStatusDone:       true,
	TodoStatusCancelled:  true,
}

// Todo represents a single to-do item.
type Todo struct {
	ID        int64
	Title     string
	Body      string
	Status    string
	Priority  int
	DueDate   *string
	CreatedAt time.Time
	UpdatedAt time.Time
}

type HealthStatus struct {
	Status string `json:"status"`
	DBOk   bool   `json:"db_ok"`
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
