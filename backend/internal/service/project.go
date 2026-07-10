package service

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/mateo/homelab/backend/internal/domain"
)

// ProjectService implements all business logic for Project operations.
// It depends exclusively on the domain.ProjectStore interface — no direct DB access.
type ProjectService struct {
	store domain.ProjectStore
}

// ProjectPatch describes optional field updates for a Project.
type ProjectPatch struct {
	Name  *string
	Color *string
}

// NewProjectService creates a new ProjectService with the given store.
func NewProjectService(store domain.ProjectStore) *ProjectService {
	return &ProjectService{store: store}
}

// CreateProject creates a new Project with the given fields and persists it.
func (s *ProjectService) CreateProject(ctx context.Context, name, color string, createdAt time.Time) (*domain.Project, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return nil, fmt.Errorf("name is required: %w", ErrValidation)
	}
	if color == "" {
		color = "default"
	}

	project := &domain.Project{
		Name:      name,
		Color:     color,
		CreatedAt: createdAt.UTC(),
	}
	if err := s.store.CreateProject(ctx, project); err != nil {
		return nil, fmt.Errorf("service.CreateProject: %w", err)
	}
	return project, nil
}

// ListProjects returns all stored Projects.
func (s *ProjectService) ListProjects(ctx context.Context) ([]*domain.Project, error) {
	projects, err := s.store.GetAllProjects(ctx)
	if err != nil {
		return nil, fmt.Errorf("service.ListProjects: %w", err)
	}
	return projects, nil
}

// GetProject returns the Project with the given ID.
func (s *ProjectService) GetProject(ctx context.Context, id int64) (*domain.Project, error) {
	project, err := s.store.GetProjectByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("service.GetProject(%d): %w", id, err)
	}
	return project, nil
}

// UpdateProject updates fields of an existing Project.
func (s *ProjectService) UpdateProject(ctx context.Context, id int64, patch ProjectPatch) (*domain.Project, error) {
	project, err := s.store.GetProjectByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("service.UpdateProject(%d): %w", id, err)
	}

	if patch.Name != nil {
		name := strings.TrimSpace(*patch.Name)
		if name == "" {
			return nil, fmt.Errorf("name is required: %w", ErrValidation)
		}
		project.Name = name
	}
	if patch.Color != nil {
		project.Color = *patch.Color
	}

	if err := s.store.UpdateProject(ctx, project); err != nil {
		return nil, fmt.Errorf("service.UpdateProject(%d) persist: %w", id, err)
	}
	return project, nil
}

// DeleteProject removes the Project with the given ID.
func (s *ProjectService) DeleteProject(ctx context.Context, id int64) error {
	if err := s.store.DeleteProject(ctx, id); err != nil {
		return fmt.Errorf("service.DeleteProject(%d): %w", id, err)
	}
	return nil
}
