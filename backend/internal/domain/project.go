package domain

import (
	"context"
	"time"
)

// Project represents a grouping label that Todos can optionally belong to.
type Project struct {
	ID        int64
	Name      string
	Color     string
	CreatedAt time.Time
}

// ProjectStore defines the persistence contract for Project entities.
// Method names are suffixed with "Project" so a single concrete store (e.g.
// SQLiteStore) can implement both TodoStore and ProjectStore without name clashes.
type ProjectStore interface {
	CreateProject(ctx context.Context, project *Project) error
	GetAllProjects(ctx context.Context) ([]*Project, error)
	GetProjectByID(ctx context.Context, id int64) (*Project, error)
	UpdateProject(ctx context.Context, project *Project) error
	DeleteProject(ctx context.Context, id int64) error
}
