package store

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/mateo/homelab/backend/internal/domain"
)

// Ensure SQLiteStore satisfies the domain.SecretStore interface at compile time.
var _ domain.SecretStore = (*SQLiteStore)(nil)

// CreateProduct inserts a new Product row and sets product.ID to the new row's ID.
func (s *SQLiteStore) CreateProduct(ctx context.Context, product *domain.Product) error {
	const q = `INSERT INTO products (name, created_at) VALUES (?, ?)`

	now := product.CreatedAt.UTC().Format(time.RFC3339)
	res, err := s.db.ExecContext(ctx, q, product.Name, now)
	if err != nil {
		return fmt.Errorf("store.CreateProduct: %w", err)
	}
	id, err := res.LastInsertId()
	if err != nil {
		return fmt.Errorf("store.CreateProduct LastInsertId: %w", err)
	}
	product.ID = id
	return nil
}

// GetAllProducts returns all Product rows ordered by id ASC.
func (s *SQLiteStore) GetAllProducts(ctx context.Context) ([]*domain.Product, error) {
	const q = `SELECT id, name, created_at FROM products ORDER BY id ASC`

	rows, err := s.db.QueryContext(ctx, q)
	if err != nil {
		return nil, fmt.Errorf("store.GetAllProducts: %w", err)
	}
	defer func() { _ = rows.Close() }()

	products := []*domain.Product{}
	for rows.Next() {
		var p domain.Product
		var createdAt string
		if err := rows.Scan(&p.ID, &p.Name, &createdAt); err != nil {
			return nil, fmt.Errorf("store.GetAllProducts scan: %w", err)
		}
		ca, err := time.Parse(time.RFC3339, createdAt)
		if err != nil {
			return nil, fmt.Errorf("parse created_at %q: %w", createdAt, err)
		}
		p.CreatedAt = ca
		products = append(products, &p)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("store.GetAllProducts rows: %w", err)
	}
	return products, nil
}

// GetProductByID returns the Product with the given ID, or domain.ErrNotFound.
func (s *SQLiteStore) GetProductByID(ctx context.Context, id int64) (*domain.Product, error) {
	const q = `SELECT id, name, created_at FROM products WHERE id = ?`

	var p domain.Product
	var createdAt string
	err := s.db.QueryRowContext(ctx, q, id).Scan(&p.ID, &p.Name, &createdAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("get product: %w", domain.ErrNotFound)
		}
		return nil, fmt.Errorf("store.GetProductByID(%d): %w", id, err)
	}
	ca, err := time.Parse(time.RFC3339, createdAt)
	if err != nil {
		return nil, fmt.Errorf("parse created_at %q: %w", createdAt, err)
	}
	p.CreatedAt = ca
	return &p, nil
}

// CreateSecretProjectWithEnvironments creates a SecretProject and its three
// default environments (development, staging, production) in a single
// transaction so a project never exists without its environments.
func (s *SQLiteStore) CreateSecretProjectWithEnvironments(ctx context.Context, project *domain.SecretProject) ([]*domain.SecretEnvironment, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("store.CreateSecretProjectWithEnvironments begin: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	now := project.CreatedAt.UTC().Format(time.RFC3339)
	res, err := tx.ExecContext(ctx, `INSERT INTO secret_projects (product_id, name, created_at) VALUES (?, ?, ?)`,
		project.ProductID, project.Name, now)
	if err != nil {
		return nil, fmt.Errorf("store.CreateSecretProjectWithEnvironments insert project: %w", err)
	}
	projectID, err := res.LastInsertId()
	if err != nil {
		return nil, fmt.Errorf("store.CreateSecretProjectWithEnvironments LastInsertId: %w", err)
	}
	project.ID = projectID

	envs := make([]*domain.SecretEnvironment, 0, len(domain.ValidEnvironments))
	for _, name := range domain.ValidEnvironments {
		envRes, err := tx.ExecContext(ctx, `INSERT INTO secret_environments (project_id, name, created_at) VALUES (?, ?, ?)`,
			projectID, name, now)
		if err != nil {
			return nil, fmt.Errorf("store.CreateSecretProjectWithEnvironments insert environment %q: %w", name, err)
		}
		envID, err := envRes.LastInsertId()
		if err != nil {
			return nil, fmt.Errorf("store.CreateSecretProjectWithEnvironments env LastInsertId: %w", err)
		}
		envs = append(envs, &domain.SecretEnvironment{
			ID:        envID,
			ProjectID: projectID,
			Name:      name,
			CreatedAt: project.CreatedAt.UTC(),
		})
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("store.CreateSecretProjectWithEnvironments commit: %w", err)
	}
	return envs, nil
}

// GetSecretProjectsByProduct returns all SecretProjects for a Product, ordered by id ASC.
func (s *SQLiteStore) GetSecretProjectsByProduct(ctx context.Context, productID int64) ([]*domain.SecretProject, error) {
	const q = `SELECT id, product_id, name, created_at FROM secret_projects WHERE product_id = ? ORDER BY id ASC`

	rows, err := s.db.QueryContext(ctx, q, productID)
	if err != nil {
		return nil, fmt.Errorf("store.GetSecretProjectsByProduct: %w", err)
	}
	defer func() { _ = rows.Close() }()

	projects := []*domain.SecretProject{}
	for rows.Next() {
		p, err := scanSecretProject(rows)
		if err != nil {
			return nil, fmt.Errorf("store.GetSecretProjectsByProduct scan: %w", err)
		}
		projects = append(projects, p)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("store.GetSecretProjectsByProduct rows: %w", err)
	}
	return projects, nil
}

// GetSecretProjectByID returns the SecretProject with the given ID, or domain.ErrNotFound.
func (s *SQLiteStore) GetSecretProjectByID(ctx context.Context, id int64) (*domain.SecretProject, error) {
	const q = `SELECT id, product_id, name, created_at FROM secret_projects WHERE id = ?`

	p, err := scanSecretProject(s.db.QueryRowContext(ctx, q, id))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("get secret project: %w", domain.ErrNotFound)
		}
		return nil, fmt.Errorf("store.GetSecretProjectByID(%d): %w", id, err)
	}
	return p, nil
}

// GetEnvironmentsByProject returns all SecretEnvironments for a SecretProject, ordered by id ASC.
func (s *SQLiteStore) GetEnvironmentsByProject(ctx context.Context, projectID int64) ([]*domain.SecretEnvironment, error) {
	const q = `SELECT id, project_id, name, created_at FROM secret_environments WHERE project_id = ? ORDER BY id ASC`

	rows, err := s.db.QueryContext(ctx, q, projectID)
	if err != nil {
		return nil, fmt.Errorf("store.GetEnvironmentsByProject: %w", err)
	}
	defer func() { _ = rows.Close() }()

	envs := []*domain.SecretEnvironment{}
	for rows.Next() {
		e, err := scanSecretEnvironment(rows)
		if err != nil {
			return nil, fmt.Errorf("store.GetEnvironmentsByProject scan: %w", err)
		}
		envs = append(envs, e)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("store.GetEnvironmentsByProject rows: %w", err)
	}
	return envs, nil
}

// GetEnvironmentByID returns the SecretEnvironment with the given ID, or domain.ErrNotFound.
func (s *SQLiteStore) GetEnvironmentByID(ctx context.Context, id int64) (*domain.SecretEnvironment, error) {
	const q = `SELECT id, project_id, name, created_at FROM secret_environments WHERE id = ?`

	e, err := scanSecretEnvironment(s.db.QueryRowContext(ctx, q, id))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("get environment: %w", domain.ErrNotFound)
		}
		return nil, fmt.Errorf("store.GetEnvironmentByID(%d): %w", id, err)
	}
	return e, nil
}

// CreateSecret inserts a new Secret row and sets secret.ID.
func (s *SQLiteStore) CreateSecret(ctx context.Context, secret *domain.Secret) error {
	const q = `INSERT INTO secrets (environment_id, key, value_encrypted, created_at, updated_at) VALUES (?, ?, ?, ?, ?)`

	now := secret.CreatedAt.UTC().Format(time.RFC3339)
	res, err := s.db.ExecContext(ctx, q, secret.EnvironmentID, secret.Key, secret.ValueEncrypted, now, now)
	if err != nil {
		return fmt.Errorf("store.CreateSecret: %w", err)
	}
	id, err := res.LastInsertId()
	if err != nil {
		return fmt.Errorf("store.CreateSecret LastInsertId: %w", err)
	}
	secret.ID = id
	secret.UpdatedAt = secret.CreatedAt.UTC()
	return nil
}

// GetSecretsByEnvironment returns all Secrets for an environment, ordered by key ASC
// (deterministic ordering used for both listing and export).
func (s *SQLiteStore) GetSecretsByEnvironment(ctx context.Context, environmentID int64) ([]*domain.Secret, error) {
	const q = `SELECT id, environment_id, key, value_encrypted, created_at, updated_at FROM secrets WHERE environment_id = ? ORDER BY key ASC`

	rows, err := s.db.QueryContext(ctx, q, environmentID)
	if err != nil {
		return nil, fmt.Errorf("store.GetSecretsByEnvironment: %w", err)
	}
	defer func() { _ = rows.Close() }()

	secrets := []*domain.Secret{}
	for rows.Next() {
		sec, err := scanSecret(rows)
		if err != nil {
			return nil, fmt.Errorf("store.GetSecretsByEnvironment scan: %w", err)
		}
		secrets = append(secrets, sec)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("store.GetSecretsByEnvironment rows: %w", err)
	}
	return secrets, nil
}

// GetSecretByKey returns the Secret with the given key in an environment, or domain.ErrNotFound.
func (s *SQLiteStore) GetSecretByKey(ctx context.Context, environmentID int64, key string) (*domain.Secret, error) {
	const q = `SELECT id, environment_id, key, value_encrypted, created_at, updated_at FROM secrets WHERE environment_id = ? AND key = ?`

	sec, err := scanSecret(s.db.QueryRowContext(ctx, q, environmentID, key))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("get secret: %w", domain.ErrNotFound)
		}
		return nil, fmt.Errorf("store.GetSecretByKey(%d, %q): %w", environmentID, key, err)
	}
	return sec, nil
}

// UpdateSecret persists a new encrypted value for an existing Secret row.
func (s *SQLiteStore) UpdateSecret(ctx context.Context, secret *domain.Secret) error {
	const q = `UPDATE secrets SET value_encrypted = ?, updated_at = ? WHERE environment_id = ? AND key = ?`

	now := time.Now().UTC().Format(time.RFC3339)
	res, err := s.db.ExecContext(ctx, q, secret.ValueEncrypted, now, secret.EnvironmentID, secret.Key)
	if err != nil {
		return fmt.Errorf("store.UpdateSecret: %w", err)
	}
	n, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("store.UpdateSecret RowsAffected: %w", err)
	}
	if n == 0 {
		return fmt.Errorf("update secret: %w", domain.ErrNotFound)
	}
	return nil
}

// DeleteSecret removes the Secret with the given key from an environment.
func (s *SQLiteStore) DeleteSecret(ctx context.Context, environmentID int64, key string) error {
	res, err := s.db.ExecContext(ctx, `DELETE FROM secrets WHERE environment_id = ? AND key = ?`, environmentID, key)
	if err != nil {
		return fmt.Errorf("store.DeleteSecret: %w", err)
	}
	n, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("store.DeleteSecret RowsAffected: %w", err)
	}
	if n == 0 {
		return fmt.Errorf("delete secret: %w", domain.ErrNotFound)
	}
	return nil
}

// CreateAuditLog inserts a new SecretAuditLog row. Never persists secret values.
func (s *SQLiteStore) CreateAuditLog(ctx context.Context, log *domain.SecretAuditLog) error {
	const q = `INSERT INTO secret_audit_logs (environment_id, secret_key, action, actor, created_at) VALUES (?, ?, ?, ?, ?)`

	now := log.CreatedAt.UTC().Format(time.RFC3339)
	res, err := s.db.ExecContext(ctx, q, log.EnvironmentID, log.SecretKey, log.Action, log.Actor, now)
	if err != nil {
		return fmt.Errorf("store.CreateAuditLog: %w", err)
	}
	id, err := res.LastInsertId()
	if err != nil {
		return fmt.Errorf("store.CreateAuditLog LastInsertId: %w", err)
	}
	log.ID = id
	return nil
}

// GetAuditLogs returns all audit logs for an environment, ordered by id ASC.
func (s *SQLiteStore) GetAuditLogs(ctx context.Context, environmentID int64) ([]*domain.SecretAuditLog, error) {
	const q = `SELECT id, environment_id, secret_key, action, actor, created_at FROM secret_audit_logs WHERE environment_id = ? ORDER BY id ASC`

	rows, err := s.db.QueryContext(ctx, q, environmentID)
	if err != nil {
		return nil, fmt.Errorf("store.GetAuditLogs: %w", err)
	}
	defer func() { _ = rows.Close() }()

	logs := []*domain.SecretAuditLog{}
	for rows.Next() {
		var l domain.SecretAuditLog
		var createdAt string
		if err := rows.Scan(&l.ID, &l.EnvironmentID, &l.SecretKey, &l.Action, &l.Actor, &createdAt); err != nil {
			return nil, fmt.Errorf("store.GetAuditLogs scan: %w", err)
		}
		ca, err := time.Parse(time.RFC3339, createdAt)
		if err != nil {
			return nil, fmt.Errorf("parse created_at %q: %w", createdAt, err)
		}
		l.CreatedAt = ca
		logs = append(logs, &l)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("store.GetAuditLogs rows: %w", err)
	}
	return logs, nil
}

func scanSecretProject(s scanner) (*domain.SecretProject, error) {
	var (
		p         domain.SecretProject
		createdAt string
	)
	if err := s.Scan(&p.ID, &p.ProductID, &p.Name, &createdAt); err != nil {
		return nil, err
	}
	ca, err := time.Parse(time.RFC3339, createdAt)
	if err != nil {
		return nil, fmt.Errorf("parse created_at %q: %w", createdAt, err)
	}
	p.CreatedAt = ca
	return &p, nil
}

func scanSecretEnvironment(s scanner) (*domain.SecretEnvironment, error) {
	var (
		e         domain.SecretEnvironment
		createdAt string
	)
	if err := s.Scan(&e.ID, &e.ProjectID, &e.Name, &createdAt); err != nil {
		return nil, err
	}
	ca, err := time.Parse(time.RFC3339, createdAt)
	if err != nil {
		return nil, fmt.Errorf("parse created_at %q: %w", createdAt, err)
	}
	e.CreatedAt = ca
	return &e, nil
}

func scanSecret(s scanner) (*domain.Secret, error) {
	var (
		sec       domain.Secret
		createdAt string
		updatedAt string
	)
	if err := s.Scan(&sec.ID, &sec.EnvironmentID, &sec.Key, &sec.ValueEncrypted, &createdAt, &updatedAt); err != nil {
		return nil, err
	}
	ca, err := time.Parse(time.RFC3339, createdAt)
	if err != nil {
		return nil, fmt.Errorf("parse created_at %q: %w", createdAt, err)
	}
	ua, err := time.Parse(time.RFC3339, updatedAt)
	if err != nil {
		return nil, fmt.Errorf("parse updated_at %q: %w", updatedAt, err)
	}
	sec.CreatedAt = ca
	sec.UpdatedAt = ua
	return &sec, nil
}
