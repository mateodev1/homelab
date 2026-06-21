package service

import (
	"context"
	"fmt"
	"time"

	"github.com/mateo/homelab/backend/internal/domain"
)

// TodoService implements all business logic for Todo operations.
// It depends exclusively on the domain.TodoStore interface — no direct DB access.
type TodoService struct {
	store domain.TodoStore
}

// NewTodoService creates a new TodoService with the given store.
func NewTodoService(store domain.TodoStore) *TodoService {
	return &TodoService{store: store}
}

// CreateTodo creates a new Todo with the given title and persists it.
func (s *TodoService) CreateTodo(ctx context.Context, title string, createdAt time.Time) (*domain.Todo, error) {
	todo := &domain.Todo{
		Title:     title,
		Done:      false,
		CreatedAt: createdAt.UTC(),
	}
	if err := s.store.Create(ctx, todo); err != nil {
		return nil, fmt.Errorf("service.CreateTodo: %w", err)
	}
	return todo, nil
}

// ListTodos returns stored Todos, optionally filtered by done status.
func (s *TodoService) ListTodos(ctx context.Context, done *bool) ([]*domain.Todo, error) {
	todos, err := s.store.GetAll(ctx)
	if err != nil {
		return nil, fmt.Errorf("service.ListTodos: %w", err)
	}
	if done != nil {
		filtered := make([]*domain.Todo, 0, len(todos))
		for _, t := range todos {
			if t.Done == *done {
				filtered = append(filtered, t)
			}
		}
		return filtered, nil
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

// UpdateTodo updates the title and done status of an existing Todo.
func (s *TodoService) UpdateTodo(ctx context.Context, id int64, title string, done bool) (*domain.Todo, error) {
	todo, err := s.store.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("service.UpdateTodo(%d): %w", id, err)
	}
	todo.Title = title
	todo.Done = done
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
