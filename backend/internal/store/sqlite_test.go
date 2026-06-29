package store_test

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	"github.com/mateo/homelab/backend/internal/domain"
	"github.com/mateo/homelab/backend/internal/store"

	_ "modernc.org/sqlite"
)

func openTestDB(t *testing.T) *sql.DB {
	t.Helper()

	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("sql.Open: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })

	if err := store.Migrate(db); err != nil {
		t.Fatalf("store.Migrate: %v", err)
	}

	return db
}

func TestCreate_Insert(t *testing.T) {
	t.Parallel()

	db := openTestDB(t)
	s := store.New(db)
	ctx := context.Background()

	todo := &domain.Todo{
		Title:     "Write tests first",
		Body:      "Use good assertions",
		Priority:  2,
		CreatedAt: time.Now().UTC().Truncate(time.Second),
	}

	if err := s.Create(ctx, todo); err != nil {
		t.Fatalf("Create: %v", err)
	}
	if todo.ID == 0 {
		t.Error("expected Create to assign a non-zero ID")
	}
	if todo.Status != domain.TodoStatusTodo {
		t.Fatalf("expected default status todo, got %q", todo.Status)
	}
}

func TestGetByID_Found(t *testing.T) {
	t.Parallel()

	db := openTestDB(t)
	s := store.New(db)
	ctx := context.Background()

	dueDate := "2026-07-01"
	original := &domain.Todo{
		Title:     "Find me",
		Body:      "Body",
		Status:    domain.TodoStatusInProgress,
		Priority:  3,
		DueDate:   &dueDate,
		CreatedAt: time.Now().UTC().Truncate(time.Second),
	}
	if err := s.Create(ctx, original); err != nil {
		t.Fatalf("Create: %v", err)
	}

	got, err := s.GetByID(ctx, original.ID)
	if err != nil {
		t.Fatalf("GetByID: %v", err)
	}
	if got.Title != original.Title {
		t.Errorf("expected Title %q, got %q", original.Title, got.Title)
	}
	if got.Status != domain.TodoStatusInProgress {
		t.Errorf("expected status in_progress, got %q", got.Status)
	}
	if got.Priority != 3 {
		t.Errorf("expected priority 3, got %d", got.Priority)
	}
	if got.DueDate == nil || *got.DueDate != dueDate {
		t.Fatalf("expected due date %q, got %#v", dueDate, got.DueDate)
	}
}

func TestGetByID_NotFound(t *testing.T) {
	t.Parallel()

	db := openTestDB(t)
	s := store.New(db)

	_, err := s.GetByID(context.Background(), 99999)
	if !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("expected errors.Is(err, domain.ErrNotFound)=true, got err=%v", err)
	}
}

func TestGetAll_EmptyDB(t *testing.T) {
	t.Parallel()

	db := openTestDB(t)
	s := store.New(db)

	todos, err := s.GetAll(context.Background())
	if err != nil {
		t.Fatalf("GetAll on empty DB: %v", err)
	}
	if len(todos) != 0 {
		t.Errorf("expected 0 todos in empty DB, got %d", len(todos))
	}
}

func TestUpdate_ChangesFields(t *testing.T) {
	t.Parallel()

	db := openTestDB(t)
	s := store.New(db)
	ctx := context.Background()

	todo := &domain.Todo{
		Title:     "Before",
		Body:      "Body",
		Status:    domain.TodoStatusTodo,
		Priority:  1,
		CreatedAt: time.Now().UTC(),
	}
	if err := s.Create(ctx, todo); err != nil {
		t.Fatalf("Create: %v", err)
	}

	dueDate := "2026-08-01"
	todo.Title = "After"
	todo.Status = domain.TodoStatusDone
	todo.Priority = 3
	todo.DueDate = &dueDate
	if err := s.Update(ctx, todo); err != nil {
		t.Fatalf("Update: %v", err)
	}

	got, err := s.GetByID(ctx, todo.ID)
	if err != nil {
		t.Fatalf("GetByID after Update: %v", err)
	}
	if got.Title != "After" {
		t.Errorf("expected Title \"After\", got %q", got.Title)
	}
	if got.Status != domain.TodoStatusDone {
		t.Errorf("expected status done, got %q", got.Status)
	}
	if got.Priority != 3 {
		t.Errorf("expected priority 3, got %d", got.Priority)
	}
	if got.DueDate == nil || *got.DueDate != dueDate {
		t.Fatalf("expected due date %q, got %#v", dueDate, got.DueDate)
	}
}

func TestUpdate_NotFound(t *testing.T) {
	t.Parallel()

	db := openTestDB(t)
	s := store.New(db)

	err := s.Update(context.Background(), &domain.Todo{ID: 99999, Title: "x"})
	if !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("expected errors.Is(err, domain.ErrNotFound)=true, got err=%v", err)
	}
}

func TestDelete_Removes(t *testing.T) {
	t.Parallel()

	db := openTestDB(t)
	s := store.New(db)
	ctx := context.Background()

	todo := &domain.Todo{Title: "Delete me", CreatedAt: time.Now().UTC()}
	if err := s.Create(ctx, todo); err != nil {
		t.Fatalf("Create: %v", err)
	}

	if err := s.Delete(ctx, todo.ID); err != nil {
		t.Fatalf("Delete: %v", err)
	}

	_, err := s.GetByID(ctx, todo.ID)
	if err == nil {
		t.Error("expected error after Delete, got nil")
	}
}

func TestDelete_NotFound(t *testing.T) {
	t.Parallel()

	db := openTestDB(t)
	s := store.New(db)

	err := s.Delete(context.Background(), 99999)
	if !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("expected errors.Is(err, domain.ErrNotFound)=true, got err=%v", err)
	}
}

func TestTodo_DueDateNullRoundtrip(t *testing.T) {
	t.Parallel()

	db := openTestDB(t)
	s := store.New(db)
	ctx := context.Background()

	todo := &domain.Todo{
		Title:     "No date",
		Status:    domain.TodoStatusTodo,
		Priority:  0,
		DueDate:   nil,
		CreatedAt: time.Now().UTC().Truncate(time.Second),
	}
	if err := s.Create(ctx, todo); err != nil {
		t.Fatalf("Create: %v", err)
	}

	got, err := s.GetByID(ctx, todo.ID)
	if err != nil {
		t.Fatalf("GetByID: %v", err)
	}
	if got.DueDate != nil {
		t.Fatalf("expected due_date to be nil, got %q", *got.DueDate)
	}
}

func TestTodo_StatusDefaultsFromMigration(t *testing.T) {
	t.Parallel()

	db := openTestDB(t)
	ctx := context.Background()

	createdAt := time.Now().UTC().Format(time.RFC3339)
	if _, err := db.ExecContext(ctx, `INSERT INTO todos (title, done, created_at, body, color, pinned, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?)`,
		"legacy row", 0, createdAt, "", "default", 0, createdAt,
	); err != nil {
		t.Fatalf("insert legacy row: %v", err)
	}

	s := store.New(db)
	todos, err := s.GetAll(ctx)
	if err != nil {
		t.Fatalf("GetAll: %v", err)
	}
	if len(todos) != 1 {
		t.Fatalf("expected 1 todo, got %d", len(todos))
	}
	if todos[0].Status != domain.TodoStatusTodo {
		t.Fatalf("expected default status todo, got %q", todos[0].Status)
	}
}

func TestTodo_SelectAndScanDoNotDependOnDoneColorPinned(t *testing.T) {
	t.Parallel()

	db := openTestDB(t)
	ctx := context.Background()

	createdAt := time.Now().UTC().Format(time.RFC3339)
	if _, err := db.ExecContext(ctx, `
		INSERT INTO todos (title, done, created_at, body, color, pinned, updated_at, status, priority, due_date)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, "scan guard", 1, createdAt, "body", "red", 1, createdAt, domain.TodoStatusInProgress, 2, nil); err != nil {
		t.Fatalf("insert row: %v", err)
	}

	s := store.New(db)
	todos, err := s.GetAll(ctx)
	if err != nil {
		t.Fatalf("GetAll: %v", err)
	}
	if len(todos) != 1 {
		t.Fatalf("expected 1 todo, got %d", len(todos))
	}
	if todos[0].Status != domain.TodoStatusInProgress {
		t.Fatalf("expected status in_progress, got %q", todos[0].Status)
	}
	if todos[0].Priority != 2 {
		t.Fatalf("expected priority 2, got %d", todos[0].Priority)
	}
}
