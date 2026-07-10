package store

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/mateo/homelab/backend/internal/domain"
)

// Ensure SQLiteStore satisfies the domain.ProjectStore interface at compile time.
var _ domain.ProjectStore = (*SQLiteStore)(nil)

// CreateProject inserts a new Project row and sets project.ID to the new row's ID.
func (s *SQLiteStore) CreateProject(ctx context.Context, project *domain.Project) error {
	const q = `INSERT INTO projects (name, color, created_at) VALUES (?, ?, ?)`

	if project.Color == "" {
		project.Color = "default"
	}

	now := project.CreatedAt.UTC().Format(time.RFC3339)
	res, err := s.db.ExecContext(ctx, q, project.Name, project.Color, now)
	if err != nil {
		return fmt.Errorf("store.CreateProject: %w", err)
	}

	id, err := res.LastInsertId()
	if err != nil {
		return fmt.Errorf("store.CreateProject LastInsertId: %w", err)
	}

	project.ID = id
	return nil
}

// GetAllProjects returns all Project rows ordered by id ASC.
func (s *SQLiteStore) GetAllProjects(ctx context.Context) ([]*domain.Project, error) {
	const q = `SELECT id, name, color, created_at FROM projects ORDER BY id ASC`

	rows, err := s.db.QueryContext(ctx, q)
	if err != nil {
		return nil, fmt.Errorf("store.GetAllProjects: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var projects []*domain.Project
	for rows.Next() {
		project, err := scanProject(rows)
		if err != nil {
			return nil, fmt.Errorf("store.GetAllProjects scan: %w", err)
		}
		projects = append(projects, project)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("store.GetAllProjects rows: %w", err)
	}

	if projects == nil {
		projects = []*domain.Project{}
	}
	return projects, nil
}

// GetProjectByID returns the Project with the given ID, or an error if not found.
func (s *SQLiteStore) GetProjectByID(ctx context.Context, id int64) (*domain.Project, error) {
	const q = `SELECT id, name, color, created_at FROM projects WHERE id = ?`

	row := s.db.QueryRowContext(ctx, q, id)
	project, err := scanProjectRow(row)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("get project: %w", domain.ErrNotFound)
		}
		return nil, fmt.Errorf("store.GetProjectByID(%d): %w", id, err)
	}
	return project, nil
}

// UpdateProject persists changes to an existing Project row.
func (s *SQLiteStore) UpdateProject(ctx context.Context, project *domain.Project) error {
	const q = `UPDATE projects SET name = ?, color = ? WHERE id = ?`

	res, err := s.db.ExecContext(ctx, q, project.Name, project.Color, project.ID)
	if err != nil {
		return fmt.Errorf("store.UpdateProject: %w", err)
	}

	n, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("store.UpdateProject RowsAffected: %w", err)
	}
	if n == 0 {
		return fmt.Errorf("update project: %w", domain.ErrNotFound)
	}
	return nil
}

// DeleteProject removes the Project with the given ID, first clearing
// project_id on any Todos that referenced it to avoid dangling references.
func (s *SQLiteStore) DeleteProject(ctx context.Context, id int64) error {
	if _, err := s.db.ExecContext(ctx, `UPDATE todos SET project_id = NULL WHERE project_id = ?`, id); err != nil {
		return fmt.Errorf("store.DeleteProject clear todos: %w", err)
	}

	res, err := s.db.ExecContext(ctx, `DELETE FROM projects WHERE id = ?`, id)
	if err != nil {
		return fmt.Errorf("store.DeleteProject: %w", err)
	}

	n, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("store.DeleteProject RowsAffected: %w", err)
	}
	if n == 0 {
		return fmt.Errorf("delete project: %w", domain.ErrNotFound)
	}
	return nil
}

func scanProject(s scanner) (*domain.Project, error) {
	var (
		project   domain.Project
		createdAt string
	)
	if err := s.Scan(&project.ID, &project.Name, &project.Color, &createdAt); err != nil {
		return nil, err
	}
	ca, err := time.Parse(time.RFC3339, createdAt)
	if err != nil {
		return nil, fmt.Errorf("parse created_at %q: %w", createdAt, err)
	}
	project.CreatedAt = ca
	return &project, nil
}

func scanProjectRow(row *sql.Row) (*domain.Project, error) {
	var (
		project   domain.Project
		createdAt string
	)
	if err := row.Scan(&project.ID, &project.Name, &project.Color, &createdAt); err != nil {
		return nil, err
	}
	ca, err := time.Parse(time.RFC3339, createdAt)
	if err != nil {
		return nil, fmt.Errorf("parse created_at %q: %w", createdAt, err)
	}
	project.CreatedAt = ca
	return &project, nil
}
