package service_test

import (
	"context"
	"database/sql"
	"strings"
	"testing"
	"time"

	"github.com/mateo/homelab/backend/internal/domain"
	"github.com/mateo/homelab/backend/internal/service"
	"github.com/mateo/homelab/backend/internal/store"

	_ "modernc.org/sqlite"
)

func newSecretService(t *testing.T) (*service.SecretService, *domain.SecretProject, *store.SQLiteStore) {
	t.Helper()
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("sql.Open: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })
	if err := store.Migrate(db); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	s := store.New(db)

	key, _, err := store.ResolveEncryptionKey("service-test-key", "")
	if err != nil {
		t.Fatalf("resolve key: %v", err)
	}
	aead, err := store.NewGCMCipher(key)
	if err != nil {
		t.Fatalf("new cipher: %v", err)
	}
	svc := service.NewSecretService(s, aead)

	ctx := context.Background()
	product, err := svc.CreateProduct(ctx, "Homelab", time.Now())
	if err != nil {
		t.Fatalf("create product: %v", err)
	}
	project, _, err := svc.CreateSecretProject(ctx, product.ID, "api", time.Now())
	if err != nil {
		t.Fatalf("create project: %v", err)
	}
	return svc, project, s
}

func TestSecretService_CreateSecretProject_ProvisionsAllEnvironments(t *testing.T) {
	t.Parallel()
	svc, project, _ := newSecretService(t)
	ctx := context.Background()

	envs, err := svc.ListEnvironments(ctx, project.ID)
	if err != nil {
		t.Fatalf("list environments: %v", err)
	}
	if len(envs) != 3 {
		t.Fatalf("expected 3 environments, got %d", len(envs))
	}
	got := map[string]bool{}
	for _, e := range envs {
		got[e.Name] = true
	}
	for _, want := range domain.ValidEnvironments {
		if !got[want] {
			t.Fatalf("expected environment %q to be provisioned", want)
		}
	}
}

func TestSecretService_CreateSecret_ValidationErrors(t *testing.T) {
	t.Parallel()
	svc, project, _ := newSecretService(t)
	ctx := context.Background()

	tests := []struct {
		name    string
		envName string
		key     string
	}{
		{name: "empty key", envName: domain.EnvDevelopment, key: "  "},
		{name: "unknown environment", envName: "not-an-env", key: "OK"},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			_, err := svc.CreateSecret(ctx, project.ID, tt.envName, tt.key, "value", "api-key")
			if err == nil {
				t.Fatalf("expected error")
			}
		})
	}
}

func TestSecretService_ListAndGet_NeverReturnPlaintext(t *testing.T) {
	t.Parallel()
	svc, project, _ := newSecretService(t)
	ctx := context.Background()

	const plaintext = "s3cr3t-p@ssw0rd"
	if _, err := svc.CreateSecret(ctx, project.ID, domain.EnvDevelopment, "DB_PASSWORD", plaintext, "api-key"); err != nil {
		t.Fatalf("create secret: %v", err)
	}

	list, err := svc.ListSecrets(ctx, project.ID, domain.EnvDevelopment)
	if err != nil {
		t.Fatalf("list secrets: %v", err)
	}
	if len(list) != 1 {
		t.Fatalf("expected 1 secret, got %d", len(list))
	}
	if strings.Contains(list[0].MaskedValue, plaintext) {
		t.Fatalf("list must not leak plaintext")
	}

	meta, err := svc.GetSecret(ctx, project.ID, domain.EnvDevelopment, "DB_PASSWORD")
	if err != nil {
		t.Fatalf("get secret: %v", err)
	}
	if strings.Contains(meta.MaskedValue, plaintext) {
		t.Fatalf("get must not leak plaintext")
	}
}

func TestSecretService_RevealSecret_ReturnsPlaintext(t *testing.T) {
	t.Parallel()
	svc, project, _ := newSecretService(t)
	ctx := context.Background()

	const plaintext = "reveal-me"
	if _, err := svc.CreateSecret(ctx, project.ID, domain.EnvProduction, "TOKEN", plaintext, "api-key"); err != nil {
		t.Fatalf("create secret: %v", err)
	}

	got, err := svc.RevealSecret(ctx, project.ID, domain.EnvProduction, "TOKEN", "jwt")
	if err != nil {
		t.Fatalf("reveal secret: %v", err)
	}
	if got != plaintext {
		t.Fatalf("expected %q, got %q", plaintext, got)
	}
}

func TestSecretService_UpdateSecret_ChangesValue(t *testing.T) {
	t.Parallel()
	svc, project, _ := newSecretService(t)
	ctx := context.Background()

	if _, err := svc.CreateSecret(ctx, project.ID, domain.EnvStaging, "KEY", "old", "api-key"); err != nil {
		t.Fatalf("create: %v", err)
	}
	if _, err := svc.UpdateSecret(ctx, project.ID, domain.EnvStaging, "KEY", "new", "api-key"); err != nil {
		t.Fatalf("update: %v", err)
	}
	got, err := svc.RevealSecret(ctx, project.ID, domain.EnvStaging, "KEY", "api-key")
	if err != nil {
		t.Fatalf("reveal: %v", err)
	}
	if got != "new" {
		t.Fatalf("expected updated value 'new', got %q", got)
	}
}

func TestSecretService_DeleteSecret(t *testing.T) {
	t.Parallel()
	svc, project, _ := newSecretService(t)
	ctx := context.Background()

	if _, err := svc.CreateSecret(ctx, project.ID, domain.EnvDevelopment, "TO_DELETE", "v", "api-key"); err != nil {
		t.Fatalf("create: %v", err)
	}
	if err := svc.DeleteSecret(ctx, project.ID, domain.EnvDevelopment, "TO_DELETE", "api-key"); err != nil {
		t.Fatalf("delete: %v", err)
	}
	if _, err := svc.GetSecret(ctx, project.ID, domain.EnvDevelopment, "TO_DELETE"); err == nil {
		t.Fatalf("expected not found after delete")
	}
}

func TestSecretService_ExportEnvironment_SortedAndEscaped(t *testing.T) {
	t.Parallel()
	svc, project, _ := newSecretService(t)
	ctx := context.Background()

	secrets := map[string]string{
		"ZEBRA":       "simple",
		"ALPHA":       "has spaces",
		"BETA":        `has "quotes" and \backslash\`,
		"EMPTY_VALUE": "",
		"WITH_HASH":   "value#comment-looking",
		"MULTILINE":   "line1\nline2",
	}
	for k, v := range secrets {
		if _, err := svc.CreateSecret(ctx, project.ID, domain.EnvDevelopment, k, v, "api-key"); err != nil {
			t.Fatalf("create secret %q: %v", k, err)
		}
	}

	out, err := svc.ExportEnvironment(ctx, project.ID, domain.EnvDevelopment, "api-key")
	if err != nil {
		t.Fatalf("export: %v", err)
	}

	lines := strings.Split(strings.TrimRight(out, "\n"), "\n")
	if len(lines) != len(secrets) {
		t.Fatalf("expected %d lines, got %d: %q", len(secrets), len(lines), out)
	}

	// Keys must be sorted alphabetically (deterministic ordering).
	wantOrder := []string{"ALPHA", "BETA", "EMPTY_VALUE", "MULTILINE", "WITH_HASH", "ZEBRA"}
	for i, key := range wantOrder {
		if !strings.HasPrefix(lines[i], key+"=") {
			t.Fatalf("expected line %d to start with %q=, got %q", i, key, lines[i])
		}
	}

	if !strings.Contains(out, `ALPHA="has spaces"`) {
		t.Fatalf("expected quoted value with spaces, got %q", out)
	}
	if !strings.Contains(out, `BETA="has \"quotes\" and \\backslash\\"`) {
		t.Fatalf("expected escaped quotes/backslashes, got %q", out)
	}
	if !strings.Contains(out, `EMPTY_VALUE=""`) {
		t.Fatalf("expected empty value quoted, got %q", out)
	}
	if !strings.Contains(out, "MULTILINE=\"line1\\nline2\"") {
		t.Fatalf("expected escaped newline, got %q", out)
	}
	if !strings.Contains(out, `WITH_HASH="value#comment-looking"`) {
		t.Fatalf("expected '#' value quoted, got %q", out)
	}
	if !strings.Contains(out, "ZEBRA=simple") {
		t.Fatalf("expected unquoted simple value, got %q", out)
	}
}

func TestSecretService_Audit_WritesEntriesForEveryAction(t *testing.T) {
	t.Parallel()
	svc, project, s := newSecretService(t)
	ctx := context.Background()

	if _, err := svc.CreateSecret(ctx, project.ID, domain.EnvDevelopment, "AUDITED", "v1", "api-key"); err != nil {
		t.Fatalf("create: %v", err)
	}
	if _, err := svc.UpdateSecret(ctx, project.ID, domain.EnvDevelopment, "AUDITED", "v2", "api-key"); err != nil {
		t.Fatalf("update: %v", err)
	}
	if _, err := svc.RevealSecret(ctx, project.ID, domain.EnvDevelopment, "AUDITED", "jwt"); err != nil {
		t.Fatalf("reveal: %v", err)
	}
	if _, err := svc.ExportEnvironment(ctx, project.ID, domain.EnvDevelopment, "jwt"); err != nil {
		t.Fatalf("export: %v", err)
	}
	if err := svc.DeleteSecret(ctx, project.ID, domain.EnvDevelopment, "AUDITED", "api-key"); err != nil {
		t.Fatalf("delete: %v", err)
	}

	envs, err := svc.ListEnvironments(ctx, project.ID)
	if err != nil {
		t.Fatalf("list environments: %v", err)
	}
	var envID int64
	for _, e := range envs {
		if e.Name == domain.EnvDevelopment {
			envID = e.ID
		}
	}

	logs, err := s.GetAuditLogs(ctx, envID)
	if err != nil {
		t.Fatalf("get audit logs: %v", err)
	}

	wantActions := []string{
		domain.AuditActionCreate,
		domain.AuditActionUpdate,
		domain.AuditActionReveal,
		domain.AuditActionExport,
		domain.AuditActionDelete,
	}
	if len(logs) != len(wantActions) {
		t.Fatalf("expected %d audit logs, got %d", len(wantActions), len(logs))
	}
	for i, action := range wantActions {
		if logs[i].Action != action {
			t.Fatalf("expected audit log[%d].Action=%q, got %q", i, action, logs[i].Action)
		}
	}
}
