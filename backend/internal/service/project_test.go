package service_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/mateo/homelab/backend/internal/domain"
	"github.com/mateo/homelab/backend/internal/service"
)

type mockProjectStore struct {
	projects map[int64]*domain.Project
	nextID   int64
	err      error
}

func newMockProjectStore() *mockProjectStore {
	return &mockProjectStore{projects: make(map[int64]*domain.Project), nextID: 1}
}

func (m *mockProjectStore) CreateProject(_ context.Context, p *domain.Project) error {
	if m.err != nil {
		return m.err
	}
	p.ID = m.nextID
	m.nextID++
	cp := *p
	m.projects[p.ID] = &cp
	return nil
}

func (m *mockProjectStore) GetAllProjects(_ context.Context) ([]*domain.Project, error) {
	if m.err != nil {
		return nil, m.err
	}
	out := make([]*domain.Project, 0, len(m.projects))
	for _, p := range m.projects {
		cp := *p
		out = append(out, &cp)
	}
	return out, nil
}

func (m *mockProjectStore) GetProjectByID(_ context.Context, id int64) (*domain.Project, error) {
	if m.err != nil {
		return nil, m.err
	}
	p, ok := m.projects[id]
	if !ok {
		return nil, domain.ErrNotFound
	}
	cp := *p
	return &cp, nil
}

func (m *mockProjectStore) UpdateProject(_ context.Context, p *domain.Project) error {
	if m.err != nil {
		return m.err
	}
	if _, ok := m.projects[p.ID]; !ok {
		return domain.ErrNotFound
	}
	cp := *p
	m.projects[p.ID] = &cp
	return nil
}

func (m *mockProjectStore) DeleteProject(_ context.Context, id int64) error {
	if m.err != nil {
		return m.err
	}
	if _, ok := m.projects[id]; !ok {
		return domain.ErrNotFound
	}
	delete(m.projects, id)
	return nil
}

var _ domain.ProjectStore = (*mockProjectStore)(nil)

func TestCreateProject_AssignsDefaults(t *testing.T) {
	t.Parallel()

	svc := service.NewProjectService(newMockProjectStore())

	got, err := svc.CreateProject(context.Background(), "Homelab", "", time.Now())
	if err != nil {
		t.Fatalf("CreateProject: %v", err)
	}
	if got.ID == 0 {
		t.Error("expected non-zero ID from CreateProject")
	}
	if got.Color != "default" {
		t.Fatalf("expected default color, got %q", got.Color)
	}
}

func TestCreateProject_BlankNameRejected(t *testing.T) {
	t.Parallel()

	svc := service.NewProjectService(newMockProjectStore())

	_, err := svc.CreateProject(context.Background(), "   ", "", time.Now())
	if err == nil {
		t.Fatalf("expected error for blank name")
	}
	if !errors.Is(err, service.ErrValidation) {
		t.Fatalf("expected error to wrap service.ErrValidation, got %v", err)
	}
}

func TestListProjects(t *testing.T) {
	t.Parallel()

	svc := service.NewProjectService(newMockProjectStore())
	ctx := context.Background()

	for _, name := range []string{"A", "B", "C"} {
		if _, err := svc.CreateProject(ctx, name, "", time.Now()); err != nil {
			t.Fatalf("CreateProject %q: %v", name, err)
		}
	}

	projects, err := svc.ListProjects(ctx)
	if err != nil {
		t.Fatalf("ListProjects: %v", err)
	}
	if len(projects) != 3 {
		t.Errorf("expected 3 projects, got %d", len(projects))
	}
}

func TestUpdateProject_ChangesFields(t *testing.T) {
	t.Parallel()

	svc := service.NewProjectService(newMockProjectStore())
	ctx := context.Background()

	created, err := svc.CreateProject(ctx, "Before", "", time.Now())
	if err != nil {
		t.Fatalf("CreateProject: %v", err)
	}

	newName := "After"
	newColor := "blue"
	updated, err := svc.UpdateProject(ctx, created.ID, service.ProjectPatch{Name: &newName, Color: &newColor})
	if err != nil {
		t.Fatalf("UpdateProject: %v", err)
	}
	if updated.Name != newName {
		t.Fatalf("expected name %q, got %q", newName, updated.Name)
	}
	if updated.Color != newColor {
		t.Fatalf("expected color %q, got %q", newColor, updated.Color)
	}
}

func TestUpdateProject_BlankNameRejected(t *testing.T) {
	t.Parallel()

	svc := service.NewProjectService(newMockProjectStore())
	ctx := context.Background()

	created, err := svc.CreateProject(ctx, "Before", "", time.Now())
	if err != nil {
		t.Fatalf("CreateProject: %v", err)
	}

	blank := "   "
	_, err = svc.UpdateProject(ctx, created.ID, service.ProjectPatch{Name: &blank})
	if err == nil {
		t.Fatalf("expected error for blank name")
	}
	if !errors.Is(err, service.ErrValidation) {
		t.Fatalf("expected error to wrap service.ErrValidation, got %v", err)
	}
}

func TestDeleteProject_Removes(t *testing.T) {
	t.Parallel()

	svc := service.NewProjectService(newMockProjectStore())
	ctx := context.Background()

	created, err := svc.CreateProject(ctx, "Delete me", "", time.Now())
	if err != nil {
		t.Fatalf("CreateProject: %v", err)
	}

	if err := svc.DeleteProject(ctx, created.ID); err != nil {
		t.Fatalf("DeleteProject: %v", err)
	}

	_, err = svc.GetProject(ctx, created.ID)
	if err == nil {
		t.Fatal("expected error after DeleteProject")
	}
}
