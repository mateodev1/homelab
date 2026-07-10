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

// ErrValidation wraps errors caused by invalid input (as opposed to
// infrastructure/store failures) so callers can distinguish 400 from 500.
var ErrValidation = errors.New("validation error")

type TodoPatch struct {
	Title     *string
	Body      *string
	Status    *string
	Priority  *int
	DueDate   **string
	Kind      *string
	IssueType **string
	ProjectID **int64
}

// NewTodoService creates a new TodoService with the given store.
func NewTodoService(store domain.TodoStore) *TodoService {
	return &TodoService{store: store}
}

// validateKindAndIssueType applies the "note" default when kind is empty and
// enforces that issue_type is only set when kind == "issue".
func validateKindAndIssueType(kind string, issueType *string) (string, error) {
	if kind == "" {
		kind = domain.TodoKindNote
	}
	if !domain.ValidKinds[kind] {
		return "", fmt.Errorf("kind must be one of: note, issue: %w", ErrValidation)
	}
	if issueType != nil {
		if kind != domain.TodoKindIssue {
			return "", fmt.Errorf("issue_type can only be set when kind is issue: %w", ErrValidation)
		}
		if !domain.ValidIssueTypes[*issueType] {
			return "", fmt.Errorf("issue_type must be one of: feature, bug, improvement: %w", ErrValidation)
		}
	}
	return kind, nil
}

// CreateTodo creates a new Todo with the given fields and persists it.
func (s *TodoService) CreateTodo(ctx context.Context, title, body string, priority int, dueDate *string, kind string, issueType *string, projectID *int64, createdAt time.Time) (*domain.Todo, error) {
	if priority < 0 || priority > 3 {
		return nil, errors.New("priority must be between 0 and 3")
	}
	kind, err := validateKindAndIssueType(kind, issueType)
	if err != nil {
		return nil, err
	}
	todo := &domain.Todo{
		Title:     title,
		Body:      body,
		Status:    domain.TodoStatusTodo,
		Priority:  priority,
		DueDate:   dueDate,
		Kind:      kind,
		IssueType: issueType,
		ProjectID: projectID,
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
	if patch.Kind != nil {
		if !domain.ValidKinds[*patch.Kind] {
			return nil, fmt.Errorf("kind must be one of: note, issue: %w", ErrValidation)
		}
		todo.Kind = *patch.Kind
		if todo.Kind == domain.TodoKindNote && patch.IssueType == nil {
			todo.IssueType = nil
		}
	}
	if patch.IssueType != nil {
		todo.IssueType = *patch.IssueType
	}
	if patch.ProjectID != nil {
		todo.ProjectID = *patch.ProjectID
	}
	if todo.IssueType != nil {
		if todo.Kind != domain.TodoKindIssue {
			return nil, fmt.Errorf("issue_type can only be set when kind is issue: %w", ErrValidation)
		}
		if !domain.ValidIssueTypes[*todo.IssueType] {
			return nil, fmt.Errorf("issue_type must be one of: feature, bug, improvement: %w", ErrValidation)
		}
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
