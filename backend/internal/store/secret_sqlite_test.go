package store_test

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/mateo/homelab/backend/internal/domain"
	"github.com/mateo/homelab/backend/internal/store"

	_ "modernc.org/sqlite"
)

func newSecretTestStore(t *testing.T) *store.SQLiteStore {
	t.Helper()
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("sql.Open: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })
	if err := store.Migrate(db); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	return store.New(db)
}

func TestSecretStore_CreateSecretProjectWithEnvironments(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	s := newSecretTestStore(t)

	product := &domain.Product{Name: "Homelab", CreatedAt: time.Now()}
	if err := s.CreateProduct(ctx, product); err != nil {
		t.Fatalf("create product: %v", err)
	}

	project := &domain.SecretProject{ProductID: product.ID, Name: "api", CreatedAt: time.Now()}
	envs, err := s.CreateSecretProjectWithEnvironments(ctx, project)
	if err != nil {
		t.Fatalf("create project with environments: %v", err)
	}
	if project.ID == 0 {
		t.Fatalf("expected project ID to be set")
	}
	if len(envs) != len(domain.ValidEnvironments) {
		t.Fatalf("expected %d environments, got %d", len(domain.ValidEnvironments), len(envs))
	}
	for i, want := range domain.ValidEnvironments {
		if envs[i].Name != want {
			t.Fatalf("expected environment[%d]=%q, got %q", i, want, envs[i].Name)
		}
		if envs[i].ProjectID != project.ID {
			t.Fatalf("expected environment project_id=%d, got %d", project.ID, envs[i].ProjectID)
		}
	}

	fromDB, err := s.GetEnvironmentsByProject(ctx, project.ID)
	if err != nil {
		t.Fatalf("get environments: %v", err)
	}
	if len(fromDB) != 3 {
		t.Fatalf("expected 3 persisted environments, got %d", len(fromDB))
	}
}

func TestSecretStore_ProductNotFound(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	s := newSecretTestStore(t)

	if _, err := s.GetProductByID(ctx, 999); err == nil {
		t.Fatalf("expected not found error")
	}
}

func TestSecretStore_SecretCRUD(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	s := newSecretTestStore(t)

	product := &domain.Product{Name: "Homelab", CreatedAt: time.Now()}
	if err := s.CreateProduct(ctx, product); err != nil {
		t.Fatalf("create product: %v", err)
	}
	project := &domain.SecretProject{ProductID: product.ID, Name: "api", CreatedAt: time.Now()}
	envs, err := s.CreateSecretProjectWithEnvironments(ctx, project)
	if err != nil {
		t.Fatalf("create project: %v", err)
	}
	env := envs[0]

	secret := &domain.Secret{EnvironmentID: env.ID, Key: "DB_PASSWORD", ValueEncrypted: "ciphertext-a", CreatedAt: time.Now()}
	if err := s.CreateSecret(ctx, secret); err != nil {
		t.Fatalf("create secret: %v", err)
	}
	if secret.ID == 0 {
		t.Fatalf("expected secret ID to be set")
	}

	fetched, err := s.GetSecretByKey(ctx, env.ID, "DB_PASSWORD")
	if err != nil {
		t.Fatalf("get secret: %v", err)
	}
	if fetched.ValueEncrypted != "ciphertext-a" {
		t.Fatalf("expected ciphertext-a, got %q", fetched.ValueEncrypted)
	}

	if err := s.UpdateSecret(ctx, &domain.Secret{EnvironmentID: env.ID, Key: "DB_PASSWORD", ValueEncrypted: "ciphertext-b"}); err != nil {
		t.Fatalf("update secret: %v", err)
	}
	updated, err := s.GetSecretByKey(ctx, env.ID, "DB_PASSWORD")
	if err != nil {
		t.Fatalf("get updated secret: %v", err)
	}
	if updated.ValueEncrypted != "ciphertext-b" {
		t.Fatalf("expected ciphertext-b after update, got %q", updated.ValueEncrypted)
	}

	if err := s.DeleteSecret(ctx, env.ID, "DB_PASSWORD"); err != nil {
		t.Fatalf("delete secret: %v", err)
	}
	if _, err := s.GetSecretByKey(ctx, env.ID, "DB_PASSWORD"); err == nil {
		t.Fatalf("expected not found after delete")
	}
}

func TestSecretStore_AuditLog(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	s := newSecretTestStore(t)

	product := &domain.Product{Name: "Homelab", CreatedAt: time.Now()}
	_ = s.CreateProduct(ctx, product)
	project := &domain.SecretProject{ProductID: product.ID, Name: "api", CreatedAt: time.Now()}
	envs, _ := s.CreateSecretProjectWithEnvironments(ctx, project)
	env := envs[0]

	if err := s.CreateAuditLog(ctx, &domain.SecretAuditLog{
		EnvironmentID: env.ID,
		SecretKey:     "DB_PASSWORD",
		Action:        domain.AuditActionCreate,
		Actor:         "api-key",
		CreatedAt:     time.Now(),
	}); err != nil {
		t.Fatalf("create audit log: %v", err)
	}

	logs, err := s.GetAuditLogs(ctx, env.ID)
	if err != nil {
		t.Fatalf("get audit logs: %v", err)
	}
	if len(logs) != 1 {
		t.Fatalf("expected 1 audit log, got %d", len(logs))
	}
	if logs[0].Action != domain.AuditActionCreate || logs[0].Actor != "api-key" {
		t.Fatalf("unexpected audit log: %+v", logs[0])
	}
}
