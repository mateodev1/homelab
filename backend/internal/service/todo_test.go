package service_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/mateo/homelab/backend/internal/domain"
	"github.com/mateo/homelab/backend/internal/service"
)

type mockStore struct {
	todos  map[int64]*domain.Todo
	nextID int64
	err    error
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
		return nil, domain.ErrNotFound
	}
	cp := *t
	return &cp, nil
}

func (m *mockStore) Update(_ context.Context, t *domain.Todo) error {
	if m.err != nil {
		return m.err
	}
	if _, ok := m.todos[t.ID]; !ok {
		return domain.ErrNotFound
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
		return domain.ErrNotFound
	}
	delete(m.todos, id)
	return nil
}

var _ domain.TodoStore = (*mockStore)(nil)

func TestCreateTodo_AssignsDefaults(t *testing.T) {
	t.Parallel()

	svc := service.NewTodoService(newMockStore())
	ctx := context.Background()

	got, err := svc.CreateTodo(ctx, "Learn TDD", "", 3, nil, time.Now())
	if err != nil {
		t.Fatalf("CreateTodo: %v", err)
	}
	if got.ID == 0 {
		t.Error("expected non-zero ID from CreateTodo")
	}
	if got.Status != domain.TodoStatusTodo {
		t.Fatalf("expected default status todo, got %q", got.Status)
	}
	if got.Priority != 3 {
		t.Fatalf("expected priority 3, got %d", got.Priority)
	}
}

func TestCreateTodo_InvalidPriority(t *testing.T) {
	t.Parallel()

	svc := service.NewTodoService(newMockStore())

	_, err := svc.CreateTodo(context.Background(), "x", "", 7, nil, time.Now())
	if err == nil {
		t.Fatalf("expected invalid priority error")
	}
}

func TestListTodos(t *testing.T) {
	t.Parallel()

	ms := newMockStore()
	svc := service.NewTodoService(ms)
	ctx := context.Background()

	for _, title := range []string{"A", "B", "C"} {
		if _, err := svc.CreateTodo(ctx, title, "", 0, nil, time.Now()); err != nil {
			t.Fatalf("CreateTodo %q: %v", title, err)
		}
	}

	todos, err := svc.ListTodos(ctx)
	if err != nil {
		t.Fatalf("ListTodos: %v", err)
	}
	if len(todos) != 3 {
		t.Errorf("expected 3 todos, got %d", len(todos))
	}
}

func TestUpdateTodo_PatchMergeMatrix(t *testing.T) {
	t.Parallel()

	ms := newMockStore()
	svc := service.NewTodoService(ms)
	ctx := context.Background()

	dueDate := "2026-07-01"
	created, err := svc.CreateTodo(ctx, "Original", "Body", 1, &dueDate, time.Now())
	if err != nil {
		t.Fatalf("CreateTodo: %v", err)
	}

	newTitle := "Updated"
	newBody := "Body 2"
	newStatus := domain.TodoStatusInProgress
	newPriority := 3
	newDueDate := "2026-08-01"
	updated, err := svc.UpdateTodo(ctx, created.ID, service.TodoPatch{
		Title:    &newTitle,
		Body:     &newBody,
		Status:   &newStatus,
		Priority: &newPriority,
		DueDate:  ptrPtr(newDueDate),
	})
	if err != nil {
		t.Fatalf("UpdateTodo: %v", err)
	}
	if updated.Title != newTitle || updated.Body != newBody {
		t.Fatalf("expected updated title/body, got %q/%q", updated.Title, updated.Body)
	}
	if updated.Status != domain.TodoStatusInProgress {
		t.Fatalf("expected status in_progress, got %q", updated.Status)
	}
	if updated.Priority != 3 {
		t.Fatalf("expected priority 3, got %d", updated.Priority)
	}
	if updated.DueDate == nil || *updated.DueDate != newDueDate {
		t.Fatalf("expected due date %q, got %#v", newDueDate, updated.DueDate)
	}
}

func TestUpdateTodo_DueDateSemantics(t *testing.T) {
	t.Parallel()

	ms := newMockStore()
	svc := service.NewTodoService(ms)
	ctx := context.Background()

	dueDate := "2026-07-01"
	created, err := svc.CreateTodo(ctx, "Original", "Body", 1, &dueDate, time.Now())
	if err != nil {
		t.Fatalf("CreateTodo: %v", err)
	}

	// nil patch field => unchanged
	unchanged, err := svc.UpdateTodo(ctx, created.ID, service.TodoPatch{})
	if err != nil {
		t.Fatalf("UpdateTodo unchanged: %v", err)
	}
	if unchanged.DueDate == nil || *unchanged.DueDate != dueDate {
		t.Fatalf("expected due date unchanged %q", dueDate)
	}

	// &nil => clear
	cleared, err := svc.UpdateTodo(ctx, created.ID, service.TodoPatch{DueDate: nilDueDatePatch()})
	if err != nil {
		t.Fatalf("UpdateTodo clear due date: %v", err)
	}
	if cleared.DueDate != nil {
		t.Fatalf("expected due date to be nil after clear")
	}

	// &"date" => set
	setDate := "2026-10-15"
	set, err := svc.UpdateTodo(ctx, created.ID, service.TodoPatch{DueDate: ptrPtr(setDate)})
	if err != nil {
		t.Fatalf("UpdateTodo set due date: %v", err)
	}
	if set.DueDate == nil || *set.DueDate != setDate {
		t.Fatalf("expected due date %q, got %#v", setDate, set.DueDate)
	}
}

func TestUpdateTodo_ValidationPaths(t *testing.T) {
	t.Parallel()

	ms := newMockStore()
	svc := service.NewTodoService(ms)
	ctx := context.Background()

	created, err := svc.CreateTodo(ctx, "Original", "Body", 1, nil, time.Now())
	if err != nil {
		t.Fatalf("CreateTodo: %v", err)
	}

	badStatus := "blocked"
	_, err = svc.UpdateTodo(ctx, created.ID, service.TodoPatch{Status: &badStatus})
	if err == nil {
		t.Fatalf("expected invalid status error")
	}

	badPriority := 7
	_, err = svc.UpdateTodo(ctx, created.ID, service.TodoPatch{Priority: &badPriority})
	if err == nil {
		t.Fatalf("expected invalid priority error")
	}
}

func TestDeleteTodo_Removes(t *testing.T) {
	t.Parallel()

	svc := service.NewTodoService(newMockStore())
	ctx := context.Background()

	created, err := svc.CreateTodo(ctx, "Delete me", "", 0, nil, time.Now())
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

func TestCreateTodo_PropagatesStoreError(t *testing.T) {
	t.Parallel()

	ms := newMockStore()
	ms.err = errors.New("db full")
	svc := service.NewTodoService(ms)

	_, err := svc.CreateTodo(context.Background(), "Fail", "", 0, nil, time.Now())
	if err == nil {
		t.Fatal("expected error from CreateTodo when store fails")
	}
}

func ptrPtr(v string) **string {
	p := &v
	return &p
}

func nilDueDatePatch() **string {
	var p *string
	return &p
}
