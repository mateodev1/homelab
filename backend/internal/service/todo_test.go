package service_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/mateo/homelab/backend/internal/domain"
	"github.com/mateo/homelab/backend/internal/service"
)

// --- in-test mock store ---

type mockStore struct {
	todos  map[int64]*domain.Todo
	nextID int64
	err    error // when non-nil every call returns this error
}

func newMockStore() *mockStore {
	return &mockStore{todos: make(map[int64]*domain.Todo), nextID: 1}
}

func (m *mockStore) Create(_ context.Context, t *domain.Todo) error {
	if m.err != nil {
		return m.err
	}
	t.ID = m.nextID
	m.nextID++
	cp := *t
	m.todos[t.ID] = &cp
	return nil
}

func (m *mockStore) GetAll(_ context.Context) ([]*domain.Todo, error) {
	if m.err != nil {
		return nil, m.err
	}
	out := make([]*domain.Todo, 0, len(m.todos))
	for _, t := range m.todos {
		cp := *t
		out = append(out, &cp)
	}
	return out, nil
}

func (m *mockStore) GetByID(_ context.Context, id int64) (*domain.Todo, error) {
	if m.err != nil {
		return nil, m.err
	}
	t, ok := m.todos[id]
	if !ok {
		return nil, errors.New("not found")
	}
	cp := *t
	return &cp, nil
}

func (m *mockStore) Update(_ context.Context, t *domain.Todo) error {
	if m.err != nil {
		return m.err
	}
	if _, ok := m.todos[t.ID]; !ok {
		return errors.New("not found")
	}
	cp := *t
	m.todos[t.ID] = &cp
	return nil
}

func (m *mockStore) Delete(_ context.Context, id int64) error {
	if m.err != nil {
		return m.err
	}
	if _, ok := m.todos[id]; !ok {
		return errors.New("not found")
	}
	delete(m.todos, id)
	return nil
}

// Compile-time assertion: mockStore satisfies domain.TodoStore.
var _ domain.TodoStore = (*mockStore)(nil)

func boolPtr(b bool) *bool { return &b }

// --- tests ---

func TestCreateTodo_AssignsID(t *testing.T) {
	t.Parallel()

	svc := service.NewTodoService(newMockStore())
	ctx := context.Background()

	got, err := svc.CreateTodo(ctx, "Learn TDD", "", "", time.Now())
	if err != nil {
		t.Fatalf("CreateTodo: %v", err)
	}
	if got.ID == 0 {
		t.Error("expected non-zero ID from CreateTodo")
	}
	if got.Title != "Learn TDD" {
		t.Errorf("expected Title \"Learn TDD\", got %q", got.Title)
	}
}

// TestCreateTodo_PropagatesStoreError verifies the service wraps store errors.
func TestCreateTodo_PropagatesStoreError(t *testing.T) {
	t.Parallel()

	ms := newMockStore()
	ms.err = errors.New("db full")
	svc := service.NewTodoService(ms)

	_, err := svc.CreateTodo(context.Background(), "Fail", "", "", time.Now())
	if err == nil {
		t.Fatal("expected error from CreateTodo when store fails")
	}
}

func TestListTodos(t *testing.T) {
	t.Parallel()

	ms := newMockStore()
	svc := service.NewTodoService(ms)
	ctx := context.Background()

	for _, title := range []string{"A", "B", "C"} {
		if _, err := svc.CreateTodo(ctx, title, "", "", time.Now()); err != nil {
			t.Fatalf("CreateTodo %q: %v", title, err)
		}
	}

	if _, err := svc.UpdateTodo(ctx, 1, "A", "", "", false, true); err != nil {
		t.Fatalf("UpdateTodo A done=true: %v", err)
	}

	todos, err := svc.ListTodos(ctx, nil)
	if err != nil {
		t.Fatalf("ListTodos(nil): %v", err)
	}
	if len(todos) != 3 {
		t.Errorf("expected 3 todos, got %d", len(todos))
	}

	doneTodos, err := svc.ListTodos(ctx, boolPtr(true))
	if err != nil {
		t.Fatalf("ListTodos(true): %v", err)
	}
	if len(doneTodos) != 1 {
		t.Errorf("expected 1 done todo, got %d", len(doneTodos))
	}

	notDoneTodos, err := svc.ListTodos(ctx, boolPtr(false))
	if err != nil {
		t.Fatalf("ListTodos(false): %v", err)
	}
	if len(notDoneTodos) != 2 {
		t.Errorf("expected 2 not-done todos, got %d", len(notDoneTodos))
	}
}

func TestListTodos_EmptyStore(t *testing.T) {
	t.Parallel()

	svc := service.NewTodoService(newMockStore())

	todos, err := svc.ListTodos(context.Background(), nil)
	if err != nil {
		t.Fatalf("ListTodos on empty store: %v", err)
	}
	if len(todos) != 0 {
		t.Errorf("expected 0 todos, got %d", len(todos))
	}
}

func TestGetTodo_Found(t *testing.T) {
	t.Parallel()

	svc := service.NewTodoService(newMockStore())
	ctx := context.Background()

	created, err := svc.CreateTodo(ctx, "Find me", "", "", time.Now())
	if err != nil {
		t.Fatalf("CreateTodo: %v", err)
	}

	got, err := svc.GetTodo(ctx, created.ID)
	if err != nil {
		t.Fatalf("GetTodo: %v", err)
	}
	if got.Title != "Find me" {
		t.Errorf("expected \"Find me\", got %q", got.Title)
	}
}

func TestGetTodo_NotFound(t *testing.T) {
	t.Parallel()

	svc := service.NewTodoService(newMockStore())

	_, err := svc.GetTodo(context.Background(), 9999)
	if err == nil {
		t.Fatal("expected error for missing ID")
	}
}

func TestUpdateTodo_PersistsChanges(t *testing.T) {
	t.Parallel()

	svc := service.NewTodoService(newMockStore())
	ctx := context.Background()

	created, err := svc.CreateTodo(ctx, "Original", "", "", time.Now())
	if err != nil {
		t.Fatalf("CreateTodo: %v", err)
	}

	updated, err := svc.UpdateTodo(ctx, created.ID, "Updated", "", "", false, true)
	if err != nil {
		t.Fatalf("UpdateTodo: %v", err)
	}
	if updated.Title != "Updated" {
		t.Errorf("expected \"Updated\", got %q", updated.Title)
	}
	if !updated.Done {
		t.Error("expected Done true after UpdateTodo")
	}
}

func TestUpdateTodo_NotFound(t *testing.T) {
	t.Parallel()

	svc := service.NewTodoService(newMockStore())

	_, err := svc.UpdateTodo(context.Background(), 9999, "Ghost", "", "", false, true)
	if err == nil {
		t.Fatal("expected error updating non-existent todo")
	}
}

func TestDeleteTodo_Removes(t *testing.T) {
	t.Parallel()

	svc := service.NewTodoService(newMockStore())
	ctx := context.Background()

	created, err := svc.CreateTodo(ctx, "Delete me", "", "", time.Now())
	if err != nil {
		t.Fatalf("CreateTodo: %v", err)
	}

	if err := svc.DeleteTodo(ctx, created.ID); err != nil {
		t.Fatalf("DeleteTodo: %v", err)
	}

	_, err = svc.GetTodo(ctx, created.ID)
	if err == nil {
		t.Fatal("expected error after DeleteTodo")
	}
}

func TestDeleteTodo_NotFound(t *testing.T) {
	t.Parallel()

	svc := service.NewTodoService(newMockStore())

	err := svc.DeleteTodo(context.Background(), 9999)
	if err == nil {
		t.Fatal("expected error deleting non-existent todo")
	}
}
