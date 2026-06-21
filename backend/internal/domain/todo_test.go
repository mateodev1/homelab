package domain_test

import (
	"context"
	"testing"
	"time"

	"github.com/mateo/homelab/backend/internal/domain"
)

// mockStore is an in-test implementation used solely to prove the interface compiles.
type mockStore struct{}

func (m *mockStore) Create(_ context.Context, _ *domain.Todo) error   { return nil }
func (m *mockStore) GetAll(_ context.Context) ([]*domain.Todo, error)  { return nil, nil }
func (m *mockStore) GetByID(_ context.Context, _ int64) (*domain.Todo, error) {
	return nil, nil
}
func (m *mockStore) Update(_ context.Context, _ *domain.Todo) error   { return nil }
func (m *mockStore) Delete(_ context.Context, _ int64) error           { return nil }

// Compile-time assertion: mockStore satisfies domain.TodoStore.
var _ domain.TodoStore = (*mockStore)(nil)

// TestTodoZeroValue verifies that the zero value of Todo has the expected field types.
func TestTodoZeroValue(t *testing.T) {
	t.Parallel()

	var todo domain.Todo

	if todo.ID != 0 {
		t.Errorf("expected ID zero value 0, got %d", todo.ID)
	}
	if todo.Title != "" {
		t.Errorf("expected Title zero value \"\", got %q", todo.Title)
	}
	if todo.Done != false {
		t.Error("expected Done zero value false, got true")
	}
	if !todo.CreatedAt.IsZero() {
		t.Errorf("expected CreatedAt zero value, got %v", todo.CreatedAt)
	}
}

// TestTodoFieldAssignment verifies fields can be set and read back correctly.
func TestTodoFieldAssignment(t *testing.T) {
	t.Parallel()

	now := time.Now()
	todo := domain.Todo{
		ID:        42,
		Title:     "Buy milk",
		Done:      true,
		CreatedAt: now,
	}

	if todo.ID != 42 {
		t.Errorf("expected ID 42, got %d", todo.ID)
	}
	if todo.Title != "Buy milk" {
		t.Errorf("expected Title \"Buy milk\", got %q", todo.Title)
	}
	if !todo.Done {
		t.Error("expected Done true, got false")
	}
	if !todo.CreatedAt.Equal(now) {
		t.Errorf("expected CreatedAt %v, got %v", now, todo.CreatedAt)
	}
}

// TestHealthStatusFields verifies HealthStatus fields exist and have correct zero values.
func TestHealthStatusFields(t *testing.T) {
	t.Parallel()

	var hs domain.HealthStatus

	if hs.Status != "" {
		t.Errorf("expected Status zero value \"\", got %q", hs.Status)
	}
	if hs.DBOk != false {
		t.Error("expected DBOk zero value false, got true")
	}

	// Triangulation: set non-zero values.
	hs2 := domain.HealthStatus{Status: "ok", DBOk: true}
	if hs2.Status != "ok" {
		t.Errorf("expected Status \"ok\", got %q", hs2.Status)
	}
	if !hs2.DBOk {
		t.Error("expected DBOk true, got false")
	}
}
