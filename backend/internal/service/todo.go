package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/mateo/homelab/backend/internal/domain"
)

// TodoService implements all business logic for Todo operations.
// It depends exclusively on the domain.TodoStore interface — no direct DB access.
type TodoService struct {
	store domain.TodoStore
}

type TodoPatch struct {
	Title    *string
	Body     *string
	Status   *string
	Priority *int
	DueDate  **string
}

// NewTodoService creates a new TodoService with the given store.
func NewTodoService(store domain.TodoStore) *TodoService {
	return &TodoService{store: store}
}

// CreateTodo creates a new Todo with the given fields and persists it.
func (s *TodoService) CreateTodo(ctx context.Context, title, body string, priority int, dueDate *string, createdAt time.Time) (*domain.Todo, error) {
	if priority < 0 || priority > 3 {
		return nil, errors.New("priority must be between 0 and 3")
	}
	todo := &domain.Todo{
		Title:     title,
		Body:      body,
		Status:    domain.TodoStatusTodo,
		Priority:  priority,
		DueDate:   dueDate,
		CreatedAt: createdAt.UTC(),
	}
	if err := s.store.Create(ctx, todo); err != nil {
		return nil, fmt.Errorf("service.CreateTodo: %w", err)
	}
	return todo, nil
}

// ListTodos returns all stored Todos.
func (s *TodoService) ListTodos(ctx context.Context) ([]*domain.Todo, error) {
	todos, err := s.store.GetAll(ctx)
	if err != nil {
		return nil, fmt.Errorf("service.ListTodos: %w", err)
	}
	return todos, nil
}

// GetTodo returns the Todo with the given ID.
func (s *TodoService) GetTodo(ctx context.Context, id int64) (*domain.Todo, error) {
	todo, err := s.store.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("service.GetTodo(%d): %w", id, err)
	}
	return todo, nil
}

// UpdateTodo updates fields of an existing Todo.
func (s *TodoService) UpdateTodo(ctx context.Context, id int64, patch TodoPatch) (*domain.Todo, error) {
	todo, err := s.store.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("service.UpdateTodo(%d): %w", id, err)
	}

	if patch.Title != nil {
		todo.Title = *patch.Title
	}
	if patch.Body != nil {
		todo.Body = *patch.Body
	}
	if patch.Status != nil {
		if !domain.ValidStatuses[*patch.Status] {
			return nil, errors.New("status must be one of: todo, in_progress, done, cancelled")
		}
		todo.Status = *patch.Status
	}
	if patch.Priority != nil {
		if *patch.Priority < 0 || *patch.Priority > 3 {
			return nil, errors.New("priority must be between 0 and 3")
		}
		todo.Priority = *patch.Priority
	}
	if patch.DueDate != nil {
		todo.DueDate = *patch.DueDate
	}

	if err := s.store.Update(ctx, todo); err != nil {
		return nil, fmt.Errorf("service.UpdateTodo(%d) persist: %w", id, err)
	}
	return todo, nil
}

// DeleteTodo removes the Todo with the given ID.
func (s *TodoService) DeleteTodo(ctx context.Context, id int64) error {
	if err := s.store.Delete(ctx, id); err != nil {
		return fmt.Errorf("service.DeleteTodo(%d): %w", id, err)
	}
	return nil
}
