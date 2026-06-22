package domain_test

import (
	"context"
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/mateo/homelab/backend/internal/domain"
)

// mockStore is an in-test implementation used solely to prove the interface compiles.
type mockStore struct{}

func (m *mockStore) Create(_ context.Context, _ *domain.Todo) error   { return nil }
func (m *mockStore) GetAll(_ context.Context) ([]*domain.Todo, error) { return nil, nil }
func (m *mockStore) GetByID(_ context.Context, _ int64) (*domain.Todo, error) {
	return nil, nil
}
func (m *mockStore) Update(_ context.Context, _ *domain.Todo) error { return nil }
func (m *mockStore) Delete(_ context.Context, _ int64) error        { return nil }

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

func TestErrNotFound(t *testing.T) {
	t.Parallel()

	if domain.ErrNotFound == nil {
		t.Fatal("expected ErrNotFound to be non-nil")
	}
	if !strings.Contains(domain.ErrNotFound.Error(), "not found") {
		t.Errorf("expected ErrNotFound message to contain %q, got %q", "not found", domain.ErrNotFound.Error())
	}
}

func TestHealthStatus(t *testing.T) {
	t.Parallel()

	hs := domain.HealthStatus{Status: "ok", DBOk: true}
	if hs.Status != "ok" {
		t.Errorf("expected Status %q, got %q", "ok", hs.Status)
	}
	if !hs.DBOk {
		t.Error("expected DBOk true, got false")
	}

	b, err := json.Marshal(hs)
	if err != nil {
		t.Fatalf("json.Marshal: %v", err)
	}
	jsonOut := string(b)
	if !strings.Contains(jsonOut, `"status":"ok"`) {
		t.Errorf("expected marshaled JSON to contain %q, got %q", `"status":"ok"`, jsonOut)
	}
	if !strings.Contains(jsonOut, `"db_ok":true`) {
		t.Errorf("expected marshaled JSON to contain %q, got %q", `"db_ok":true`, jsonOut)
	}
}
