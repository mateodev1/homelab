package store_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/mateo/homelab/backend/internal/domain"
	"github.com/mateo/homelab/backend/internal/store"
)

func TestCreateProject_Insert(t *testing.T) {
	t.Parallel()

	db := openTestDB(t)
	s := store.New(db)
	ctx := context.Background()

	project := &domain.Project{
		Name:      "Homelab",
		CreatedAt: time.Now().UTC().Truncate(time.Second),
	}

	if err := s.CreateProject(ctx, project); err != nil {
		t.Fatalf("CreateProject: %v", err)
	}
	if project.ID == 0 {
		t.Error("expected CreateProject to assign a non-zero ID")
	}
	if project.Color != "default" {
		t.Fatalf("expected default color, got %q", project.Color)
	}
}

func TestGetProjectByID_NotFound(t *testing.T) {
	t.Parallel()

	db := openTestDB(t)
	s := store.New(db)

	_, err := s.GetProjectByID(context.Background(), 99999)
	if !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("expected errors.Is(err, domain.ErrNotFound)=true, got err=%v", err)
	}
}

func TestGetAllProjects_EmptyDB(t *testing.T) {
	t.Parallel()

	db := openTestDB(t)
	s := store.New(db)

	projects, err := s.GetAllProjects(context.Background())
	if err != nil {
		t.Fatalf("GetAllProjects on empty DB: %v", err)
	}
	if len(projects) != 0 {
		t.Errorf("expected 0 projects in empty DB, got %d", len(projects))
	}
}

func TestUpdateProject_ChangesFields(t *testing.T) {
	t.Parallel()

	db := openTestDB(t)
	s := store.New(db)
	ctx := context.Background()

	project := &domain.Project{Name: "Before", CreatedAt: time.Now().UTC()}
	if err := s.CreateProject(ctx, project); err != nil {
		t.Fatalf("CreateProject: %v", err)
	}

	project.Name = "After"
	project.Color = "blue"
	if err := s.UpdateProject(ctx, project); err != nil {
		t.Fatalf("UpdateProject: %v", err)
	}

	got, err := s.GetProjectByID(ctx, project.ID)
	if err != nil {
		t.Fatalf("GetProjectByID after Update: %v", err)
	}
	if got.Name != "After" {
		t.Errorf("expected Name \"After\", got %q", got.Name)
	}
	if got.Color != "blue" {
		t.Errorf("expected color blue, got %q", got.Color)
	}
}

func TestUpdateProject_NotFound(t *testing.T) {
	t.Parallel()

	db := openTestDB(t)
	s := store.New(db)

	err := s.UpdateProject(context.Background(), &domain.Project{ID: 99999, Name: "x"})
	if !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("expected errors.Is(err, domain.ErrNotFound)=true, got err=%v", err)
	}
}

func TestDeleteProject_ClearsTodoProjectID(t *testing.T) {
	t.Parallel()

	db := openTestDB(t)
	s := store.New(db)
	ctx := context.Background()

	project := &domain.Project{Name: "Homelab", CreatedAt: time.Now().UTC()}
	if err := s.CreateProject(ctx, project); err != nil {
		t.Fatalf("CreateProject: %v", err)
	}

	todo := &domain.Todo{
		Title:     "In project",
		ProjectID: &project.ID,
		CreatedAt: time.Now().UTC(),
	}
	if err := s.Create(ctx, todo); err != nil {
		t.Fatalf("Create todo: %v", err)
	}

	if err := s.DeleteProject(ctx, project.ID); err != nil {
		t.Fatalf("DeleteProject: %v", err)
	}

	got, err := s.GetByID(ctx, todo.ID)
	if err != nil {
		t.Fatalf("GetByID after DeleteProject: %v", err)
	}
	if got.ProjectID != nil {
		t.Fatalf("expected ProjectID to be cleared, got %v", *got.ProjectID)
	}
}

func TestDeleteProject_NotFound(t *testing.T) {
	t.Parallel()

	db := openTestDB(t)
	s := store.New(db)

	err := s.DeleteProject(context.Background(), 99999)
	if !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("expected errors.Is(err, domain.ErrNotFound)=true, got err=%v", err)
	}
}

func TestTodo_ProjectIDNullRoundtrip(t *testing.T) {
	t.Parallel()

	db := openTestDB(t)
	s := store.New(db)
	ctx := context.Background()

	todo := &domain.Todo{
		Title:     "No project",
		CreatedAt: time.Now().UTC().Truncate(time.Second),
	}
	if err := s.Create(ctx, todo); err != nil {
		t.Fatalf("Create: %v", err)
	}

	got, err := s.GetByID(ctx, todo.ID)
	if err != nil {
		t.Fatalf("GetByID: %v", err)
	}
	if got.ProjectID != nil {
		t.Fatalf("expected ProjectID to be nil, got %v", *got.ProjectID)
	}
}
