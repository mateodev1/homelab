package store_test

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	"github.com/mateo/homelab/backend/internal/domain"
	"github.com/mateo/homelab/backend/internal/store"

	// Register the modernc SQLite driver.
	_ "modernc.org/sqlite"
)

// openTestDB opens an in-memory SQLite database and runs migrations.
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

// TestCreate_Insert verifies that a Todo is persisted with an assigned ID.
func TestCreate_Insert(t *testing.T) {
	t.Parallel()

	db := openTestDB(t)
	s := store.New(db)
	ctx := context.Background()

	todo := &domain.Todo{
		Title:     "Write tests first",
		Done:      false,
		CreatedAt: time.Now().UTC().Truncate(time.Second),
	}

	if err := s.Create(ctx, todo); err != nil {
		t.Fatalf("Create: %v", err)
	}
	if todo.ID == 0 {
		t.Error("expected Create to assign a non-zero ID")
	}
}

// TestCreate_MultipleTodos verifies each insert gets a distinct ID.
func TestCreate_MultipleTodos(t *testing.T) {
	t.Parallel()

	db := openTestDB(t)
	s := store.New(db)
	ctx := context.Background()

	a := &domain.Todo{Title: "First", CreatedAt: time.Now().UTC()}
	b := &domain.Todo{Title: "Second", CreatedAt: time.Now().UTC()}

	if err := s.Create(ctx, a); err != nil {
		t.Fatalf("Create a: %v", err)
	}
	if err := s.Create(ctx, b); err != nil {
		t.Fatalf("Create b: %v", err)
	}
	if a.ID == b.ID {
		t.Errorf("expected distinct IDs, both got %d", a.ID)
	}
}

// TestGetAll_ReturnsAllInserted verifies GetAll returns every stored Todo.
func TestGetAll_ReturnsAllInserted(t *testing.T) {
	t.Parallel()

	db := openTestDB(t)
	s := store.New(db)
	ctx := context.Background()

	titles := []string{"Alpha", "Beta", "Gamma"}
	for _, title := range titles {
		if err := s.Create(ctx, &domain.Todo{Title: title, CreatedAt: time.Now().UTC()}); err != nil {
			t.Fatalf("Create %q: %v", title, err)
		}
	}

	todos, err := s.GetAll(ctx)
	if err != nil {
		t.Fatalf("GetAll: %v", err)
	}
	if len(todos) != len(titles) {
		t.Errorf("expected %d todos, got %d", len(titles), len(todos))
	}
}

// TestGetAll_EmptyDB verifies GetAll returns an empty slice when no rows exist.
func TestGetAll_EmptyDB(t *testing.T) {
	t.Parallel()

	db := openTestDB(t)
	s := store.New(db)
	ctx := context.Background()

	todos, err := s.GetAll(ctx)
	if err != nil {
		t.Fatalf("GetAll on empty DB: %v", err)
	}
	// Expect empty slice — not nil — after running real GetAll logic.
	if len(todos) != 0 {
		t.Errorf("expected 0 todos in empty DB, got %d", len(todos))
	}
}

// TestGetByID_Found verifies a stored Todo is retrieved by its assigned ID.
func TestGetByID_Found(t *testing.T) {
	t.Parallel()

	db := openTestDB(t)
	s := store.New(db)
	ctx := context.Background()

	original := &domain.Todo{Title: "Find me", Done: false, CreatedAt: time.Now().UTC().Truncate(time.Second)}
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
	if got.ID != original.ID {
		t.Errorf("expected ID %d, got %d", original.ID, got.ID)
	}
}

// TestGetByID_NotFound verifies missing ID maps to domain.ErrNotFound.
func TestGetByID_NotFound(t *testing.T) {
	t.Parallel()

	db := openTestDB(t)
	s := store.New(db)
	ctx := context.Background()

	_, err := s.GetByID(ctx, 99999)
	if !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("expected errors.Is(err, domain.ErrNotFound)=true, got err=%v", err)
	}
}

func TestUpdate_NotFound(t *testing.T) {
	t.Parallel()

	db := openTestDB(t)
	s := store.New(db)
	ctx := context.Background()

	err := s.Update(ctx, &domain.Todo{ID: 99999, Title: "x"})
	if !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("expected errors.Is(err, domain.ErrNotFound)=true, got err=%v", err)
	}
}

// TestUpdate_ChangesTitle verifies that Update persists field changes.
func TestUpdate_ChangesTitle(t *testing.T) {
	t.Parallel()

	db := openTestDB(t)
	s := store.New(db)
	ctx := context.Background()

	todo := &domain.Todo{Title: "Before", Done: false, CreatedAt: time.Now().UTC()}
	if err := s.Create(ctx, todo); err != nil {
		t.Fatalf("Create: %v", err)
	}

	todo.Title = "After"
	todo.Done = true
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
	if !got.Done {
		t.Error("expected Done true after Update")
	}
}

// TestDelete_Removes verifies that Delete makes a Todo unretrievable.
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
	ctx := context.Background()

	err := s.Delete(ctx, 99999)
	if !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("expected errors.Is(err, domain.ErrNotFound)=true, got err=%v", err)
	}
}
