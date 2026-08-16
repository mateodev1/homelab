package mcpserver

import (
	"context"
	"net/http"
	"strings"
	"testing"
)

func TestSecretProductTools(t *testing.T) {
	t.Parallel()

	t.Run("list", func(t *testing.T) {
		t.Parallel()
		srv, cap := newTestBackend(t, http.StatusOK, []byte(`[]`), nil)
		s := testServer(t, srv.URL, false, "")
		_, out, err := s.secretProductList(context.Background(), nil, secretProductListInput{})
		if err != nil {
			t.Fatal(err)
		}
		if out.Products == nil {
			t.Fatal("expected non-nil products")
		}
		if cap.path != "/api/products" {
			t.Fatalf("path = %q", cap.path)
		}
	})

	t.Run("create requires name", func(t *testing.T) {
		t.Parallel()
		srv, _ := newTestBackend(t, http.StatusOK, []byte(`{}`), nil)
		s := testServer(t, srv.URL, false, "")
		_, _, err := s.secretProductCreate(context.Background(), nil, secretProductCreateInput{Name: "  "})
		if err == nil || !strings.Contains(err.Error(), "name is required") {
			t.Fatalf("expected name required error, got %v", err)
		}
	})

	t.Run("create posts to /api/products", func(t *testing.T) {
		t.Parallel()
		respJSON := []byte(`{"id":1,"name":"acme","created_at":"t"}`)
		srv, cap := newTestBackend(t, http.StatusCreated, respJSON, nil)
		s := testServer(t, srv.URL, false, "")
		_, got, err := s.secretProductCreate(context.Background(), nil, secretProductCreateInput{Name: "acme"})
		if err != nil {
			t.Fatal(err)
		}
		if got.Name != "acme" {
			t.Fatalf("got %+v", got)
		}
		if cap.method != http.MethodPost || cap.path != "/api/products" {
			t.Fatalf("request %s %s", cap.method, cap.path)
		}
	})
}

func TestSecretProjectTools(t *testing.T) {
	t.Parallel()

	t.Run("list requires positive product id", func(t *testing.T) {
		t.Parallel()
		srv, _ := newTestBackend(t, http.StatusOK, []byte(`[]`), nil)
		s := testServer(t, srv.URL, false, "")
		_, _, err := s.secretProjectList(context.Background(), nil, secretProjectListInput{ProductID: 0})
		if err == nil || !strings.Contains(err.Error(), "product_id") {
			t.Fatalf("expected product_id error, got %v", err)
		}
	})

	t.Run("list builds nested path", func(t *testing.T) {
		t.Parallel()
		srv, cap := newTestBackend(t, http.StatusOK, []byte(`[]`), nil)
		s := testServer(t, srv.URL, false, "")
		_, _, err := s.secretProjectList(context.Background(), nil, secretProjectListInput{ProductID: 7})
		if err != nil {
			t.Fatal(err)
		}
		if cap.path != "/api/products/7/projects" {
			t.Fatalf("path = %q", cap.path)
		}
	})

	t.Run("create returns environments", func(t *testing.T) {
		t.Parallel()
		respJSON := []byte(`{"id":2,"product_id":7,"name":"webapp","created_at":"t","environments":[{"id":1,"project_id":2,"name":"development","created_at":"t"}]}`)
		srv, cap := newTestBackend(t, http.StatusCreated, respJSON, nil)
		s := testServer(t, srv.URL, false, "")
		_, got, err := s.secretProjectCreate(context.Background(), nil, secretProjectCreateInput{ProductID: 7, Name: "webapp"})
		if err != nil {
			t.Fatal(err)
		}
		if len(got.Environments) != 1 || got.Environments[0].Name != "development" {
			t.Fatalf("got %+v", got)
		}
		if cap.path != "/api/products/7/projects" {
			t.Fatalf("path = %q", cap.path)
		}
	})
}

func TestSecretEnvironmentList(t *testing.T) {
	t.Parallel()
	srv, cap := newTestBackend(t, http.StatusOK, []byte(`[]`), nil)
	s := testServer(t, srv.URL, false, "")
	_, _, err := s.secretEnvironmentList(context.Background(), nil, secretEnvironmentListInput{ProductID: 1, ProjectID: 2})
	if err != nil {
		t.Fatal(err)
	}
	if cap.path != "/api/products/1/projects/2/environments" {
		t.Fatalf("path = %q", cap.path)
	}
}

func TestValidateSecretScope(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name        string
		productID   int64
		projectID   int64
		environment string
		wantErr     string
	}{
		{"valid", 1, 2, "production", ""},
		{"bad product id", 0, 2, "production", "product_id"},
		{"bad project id", 1, 0, "production", "project_id"},
		{"empty environment", 1, 2, "", "environment is required"},
		{"invalid environment", 1, 2, "prod", "invalid environment"},
	}
	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			err := validateSecretScope(tc.productID, tc.projectID, tc.environment)
			if tc.wantErr == "" {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				return
			}
			if err == nil || !strings.Contains(err.Error(), tc.wantErr) {
				t.Fatalf("error = %v, want containing %q", err, tc.wantErr)
			}
		})
	}
}

func TestSecretCRUDTools(t *testing.T) {
	t.Parallel()

	t.Run("list", func(t *testing.T) {
		t.Parallel()
		srv, cap := newTestBackend(t, http.StatusOK, []byte(`[]`), nil)
		s := testServer(t, srv.URL, false, "")
		_, out, err := s.secretList(context.Background(), nil, secretListInput{ProductID: 1, ProjectID: 2, Environment: "staging"})
		if err != nil {
			t.Fatal(err)
		}
		if out.Secrets == nil {
			t.Fatal("expected non-nil secrets")
		}
		if cap.path != "/api/products/1/projects/2/environments/staging/secrets" {
			t.Fatalf("path = %q", cap.path)
		}
	})

	t.Run("create requires key", func(t *testing.T) {
		t.Parallel()
		srv, _ := newTestBackend(t, http.StatusOK, []byte(`{}`), nil)
		s := testServer(t, srv.URL, false, "")
		_, _, err := s.secretCreate(context.Background(), nil, secretCreateInput{ProductID: 1, ProjectID: 2, Environment: "development", Key: ""})
		if err == nil || !strings.Contains(err.Error(), "key is required") {
			t.Fatalf("expected key required error, got %v", err)
		}
	})

	t.Run("create posts key and value", func(t *testing.T) {
		t.Parallel()
		respJSON := []byte(`{"environment_id":1,"key":"API_KEY","value":"***","created_at":"t","updated_at":"t"}`)
		srv, cap := newTestBackend(t, http.StatusCreated, respJSON, nil)
		s := testServer(t, srv.URL, false, "")
		_, got, err := s.secretCreate(context.Background(), nil, secretCreateInput{
			ProductID: 1, ProjectID: 2, Environment: "development", Key: "API_KEY", Value: "secret",
		})
		if err != nil {
			t.Fatal(err)
		}
		if got.Key != "API_KEY" {
			t.Fatalf("got %+v", got)
		}
		if cap.method != http.MethodPost || cap.path != "/api/products/1/projects/2/environments/development/secrets" {
			t.Fatalf("request %s %s", cap.method, cap.path)
		}
		if !strings.Contains(cap.body, `"value":"secret"`) {
			t.Fatalf("body missing plaintext value: %q", cap.body)
		}
	})

	t.Run("reveal returns plaintext and hits reveal path", func(t *testing.T) {
		t.Parallel()
		respJSON := []byte(`{"key":"API_KEY","value":"plaintext-value"}`)
		srv, cap := newTestBackend(t, http.StatusOK, respJSON, nil)
		s := testServer(t, srv.URL, false, "")
		_, got, err := s.secretReveal(context.Background(), nil, secretKeyInput{
			ProductID: 1, ProjectID: 2, Environment: "development", Key: "API_KEY",
		})
		if err != nil {
			t.Fatal(err)
		}
		if got.Value != "plaintext-value" {
			t.Fatalf("got %+v", got)
		}
		if cap.path != "/api/products/1/projects/2/environments/development/secrets/API_KEY/reveal" {
			t.Fatalf("path = %q", cap.path)
		}
	})

	t.Run("update sends value", func(t *testing.T) {
		t.Parallel()
		respJSON := []byte(`{"environment_id":1,"key":"API_KEY","value":"***","created_at":"t","updated_at":"t"}`)
		srv, cap := newTestBackend(t, http.StatusOK, respJSON, nil)
		s := testServer(t, srv.URL, false, "")
		_, _, err := s.secretUpdate(context.Background(), nil, secretUpdateInput{
			ProductID: 1, ProjectID: 2, Environment: "development", Key: "API_KEY", Value: "new",
		})
		if err != nil {
			t.Fatal(err)
		}
		if cap.method != http.MethodPut || cap.path != "/api/products/1/projects/2/environments/development/secrets/API_KEY" {
			t.Fatalf("request %s %s", cap.method, cap.path)
		}
	})

	t.Run("delete", func(t *testing.T) {
		t.Parallel()
		srv, cap := newTestBackend(t, http.StatusNoContent, nil, nil)
		s := testServer(t, srv.URL, false, "")
		_, got, err := s.secretDelete(context.Background(), nil, secretKeyInput{
			ProductID: 1, ProjectID: 2, Environment: "development", Key: "API_KEY",
		})
		if err != nil {
			t.Fatal(err)
		}
		if !got.Deleted || got.Key != "API_KEY" {
			t.Fatalf("got %+v", got)
		}
		if cap.method != http.MethodDelete || cap.path != "/api/products/1/projects/2/environments/development/secrets/API_KEY" {
			t.Fatalf("request %s %s", cap.method, cap.path)
		}
	})

	t.Run("export returns dotenv text", func(t *testing.T) {
		t.Parallel()
		dotenv := "API_KEY=abc\nDB_URL=postgres://x\n"
		srv, cap := newTestBackend(t, http.StatusOK, []byte(dotenv), nil)
		s := testServer(t, srv.URL, false, "")
		_, got, err := s.secretExport(context.Background(), nil, secretExportInput{
			ProductID: 1, ProjectID: 2, Environment: "development",
		})
		if err != nil {
			t.Fatal(err)
		}
		if got.Dotenv != dotenv {
			t.Fatalf("got %q, want %q", got.Dotenv, dotenv)
		}
		if cap.path != "/api/products/1/projects/2/environments/development/export" {
			t.Fatalf("path = %q", cap.path)
		}
	})

	t.Run("rejects invalid environment", func(t *testing.T) {
		t.Parallel()
		srv, _ := newTestBackend(t, http.StatusOK, []byte(`[]`), nil)
		s := testServer(t, srv.URL, false, "")
		_, _, err := s.secretList(context.Background(), nil, secretListInput{ProductID: 1, ProjectID: 2, Environment: "prod"})
		if err == nil || !strings.Contains(err.Error(), "invalid environment") {
			t.Fatalf("expected invalid environment error, got %v", err)
		}
	})
}
