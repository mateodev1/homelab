package domain_test

import (
	"testing"
	"time"

	"github.com/mateo/homelab/shared/pkg/domain"
)

// TestTodoStructFields verifies that the Todo struct exposes the required fields
// with the correct types. This is the TDD RED step — types.go does not exist yet.
func TestTodoStructFields(t *testing.T) {
	now := time.Now()
	todo := domain.Todo{
		ID:        1,
		Title:     "Buy groceries",
		Done:      false,
		CreatedAt: now,
	}

	if todo.ID != 1 {
		t.Errorf("expected ID=1, got %d", todo.ID)
	}
	if todo.Title != "Buy groceries" {
		t.Errorf("expected Title='Buy groceries', got %q", todo.Title)
	}
	if todo.Done != false {
		t.Errorf("expected Done=false, got %v", todo.Done)
	}
	if !todo.CreatedAt.Equal(now) {
		t.Errorf("expected CreatedAt=%v, got %v", now, todo.CreatedAt)
	}
}

// TestTodoStructDoneTrue verifies the Done field can be set to true (triangulation).
func TestTodoStructDoneTrue(t *testing.T) {
	todo := domain.Todo{
		ID:        42,
		Title:     "Completed task",
		Done:      true,
		CreatedAt: time.Now(),
	}

	if !todo.Done {
		t.Errorf("expected Done=true, got %v", todo.Done)
	}
	if todo.ID != 42 {
		t.Errorf("expected ID=42, got %d", todo.ID)
	}
}
