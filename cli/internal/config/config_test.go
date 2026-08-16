package config

import (
	"os"
	"testing"
)

// setEnv sets env vars for the duration of a test and restores them on cleanup.
func setEnv(t *testing.T, kv map[string]string) {
	t.Helper()
	for k, v := range kv {
		prev, ok := os.LookupEnv(k)
		if err := os.Setenv(k, v); err != nil {
			t.Fatalf("setenv %s: %v", k, err)
		}
		if ok {
			t.Cleanup(func() { _ = os.Setenv(k, prev) })
		} else {
			t.Cleanup(func() { _ = os.Unsetenv(k) })
		}
	}
}

func TestResolveDefaults(t *testing.T) {
	setEnv(t, map[string]string{
		EnvBaseURL: "",
		EnvAPIKey:  "",
		EnvEnv:     "",
	})
	cfg := Resolve("", "", "")
	if cfg.BaseURL != DefaultBaseURL {
		t.Fatalf("base url = %q, want %q", cfg.BaseURL, DefaultBaseURL)
	}
	if cfg.APIKey != "" {
		t.Fatalf("api key = %q, want empty", cfg.APIKey)
	}
	if cfg.Env != DefaultEnv {
		t.Fatalf("env = %q, want %q", cfg.Env, DefaultEnv)
	}
	if cfg.RequireAuth {
		t.Fatalf("require auth = true, want false in development")
	}
}

func TestResolveEnvFallback(t *testing.T) {
	setEnv(t, map[string]string{
		EnvBaseURL: "http://example.test:9090",
		EnvAPIKey:  "env-secret",
		EnvEnv:     "",
	})
	cfg := Resolve("", "", "")
	if cfg.BaseURL != "http://example.test:9090" {
		t.Fatalf("base url = %q", cfg.BaseURL)
	}
	if cfg.APIKey != "env-secret" {
		t.Fatalf("api key = %q", cfg.APIKey)
	}
}

func TestResolveFlagOverridesEnv(t *testing.T) {
	setEnv(t, map[string]string{
		EnvBaseURL: "http://env.example",
		EnvAPIKey:  "env-secret",
		EnvEnv:     "production",
	})
	cfg := Resolve("http://flag.example", "flag-secret", "development")
	if cfg.BaseURL != "http://flag.example" {
		t.Fatalf("base url = %q", cfg.BaseURL)
	}
	if cfg.APIKey != "flag-secret" {
		t.Fatalf("api key = %q", cfg.APIKey)
	}
	// flag wins over env for Env too.
	if cfg.Env != "development" {
		t.Fatalf("env = %q, want development (flag wins)", cfg.Env)
	}
	if cfg.RequireAuth {
		t.Fatalf("require auth = true, want false for development")
	}
}

func TestResolveEnvFallbackForEnv(t *testing.T) {
	setEnv(t, map[string]string{
		EnvEnv: "production",
	})
	cfg := Resolve("", "", "")
	if cfg.Env != "production" {
		t.Fatalf("env = %q, want production (env fallback)", cfg.Env)
	}
	if !cfg.RequireAuth {
		t.Fatalf("require auth = false, want true for production")
	}
}

func TestResolveProductionRequiresAuth(t *testing.T) {
	setEnv(t, map[string]string{EnvEnv: ""})
	cases := []struct {
		name    string
		flagEnv string
		want    bool
	}{
		{"development default -> no auth", "", false},
		{"explicit development -> no auth", "development", false},
		{"explicit production -> auth", "production", true},
		{"unknown env -> no auth (only prod requires)", "staging", false},
	}
	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Helper()
			cfg := Resolve("", "", tc.flagEnv)
			if cfg.RequireAuth != tc.want {
				t.Fatalf("require auth = %v, want %v", cfg.RequireAuth, tc.want)
			}
		})
	}
}

func TestResolveAPIKeyOptionalInDev(t *testing.T) {
	setEnv(t, map[string]string{EnvEnv: ""})
	// Dev with no API key: config resolves cleanly with an empty key; the
	// gating (key required only in prod) is the caller's responsibility.
	cfg := Resolve("", "", "")
	if cfg.APIKey != "" {
		t.Fatalf("api key = %q, want empty", cfg.APIKey)
	}
	if cfg.RequireAuth {
		t.Fatal("dev must not require auth")
	}
}

// ResolveWithFile applies the precedence chain defaults -> file -> env -> flags
// (REQ-CFG-004) and derives RequireAuth = (Env == "production") (REQ-CFG-003).
// Empty env strings are treated as unset so the file value wins (REQ-CFG-005).
// These tests are pure-function (ResolveWithFile takes the file as an explicit
// Config and reads env via os.Getenv); they never touch the filesystem, so the
// configDirFn seam is not consulted here and needs no pin (REQ-CFG-007 applies
// to tests that invoke Load/Save/Execute, not the pure resolver).

// TestResolveWithFile_FileFillsDefaults covers SC-CFG-007 and the lower half
// of SC-CFG-004: with no env and no flags, the file values fill the defaults.
func TestResolveWithFile_FileFillsDefaults(t *testing.T) {
	setEnv(t, map[string]string{EnvBaseURL: "", EnvAPIKey: "", EnvEnv: ""})
	file := Config{BaseURL: "http://file", APIKey: "filekey", Env: "development"}

	cfg := ResolveWithFile(file, "", "", "")

	if cfg.BaseURL != "http://file" {
		t.Fatalf("base url = %q, want http://file (file fills default)", cfg.BaseURL)
	}
	if cfg.APIKey != "filekey" {
		t.Fatalf("api key = %q, want filekey", cfg.APIKey)
	}
	if cfg.Env != "development" {
		t.Fatalf("env = %q, want development", cfg.Env)
	}
	if cfg.RequireAuth {
		t.Fatal("development must not require auth")
	}
}

// TestResolveWithFile_EnvOverridesFile covers SC-CFG-004: a non-empty HOMELAB_*
// env var wins over the file value, while fields not set in env keep the file.
func TestResolveWithFile_EnvOverridesFile(t *testing.T) {
	setEnv(t, map[string]string{EnvBaseURL: "http://env", EnvAPIKey: "", EnvEnv: ""})
	file := Config{BaseURL: "http://file", APIKey: "filekey", Env: "development"}

	cfg := ResolveWithFile(file, "", "", "")

	if cfg.BaseURL != "http://env" {
		t.Fatalf("base url = %q, want http://env (env overrides file)", cfg.BaseURL)
	}
	if cfg.APIKey != "filekey" {
		t.Fatalf("api key = %q, want filekey (env unset, file wins)", cfg.APIKey)
	}
	if cfg.Env != "development" {
		t.Fatalf("env = %q, want development (env unset, file wins)", cfg.Env)
	}
}

// TestResolveWithFile_FlagOverridesEnv covers SC-CFG-004 top half: explicit
// flags win over env and file.
func TestResolveWithFile_FlagOverridesEnv(t *testing.T) {
	setEnv(t, map[string]string{EnvBaseURL: "http://env", EnvAPIKey: "env-secret", EnvEnv: "production"})
	file := Config{BaseURL: "http://file", APIKey: "filekey", Env: "development"}

	cfg := ResolveWithFile(file, "http://flag", "flag-secret", "development")

	if cfg.BaseURL != "http://flag" {
		t.Fatalf("base url = %q, want http://flag (flag overrides env+file)", cfg.BaseURL)
	}
	if cfg.APIKey != "flag-secret" {
		t.Fatalf("api key = %q, want flag-secret", cfg.APIKey)
	}
	if cfg.Env != "development" {
		t.Fatalf("env = %q, want development (flag wins)", cfg.Env)
	}
	if cfg.RequireAuth {
		t.Fatal("development must not require auth even when env file said production")
	}
}

// TestResolveWithFile_EmptyEnvIsUnset covers SC-CFG-006 / REQ-CFG-005: an
// explicit empty HOMELAB_ENV is treated as unset, so the file env wins.
func TestResolveWithFile_EmptyEnvIsUnset(t *testing.T) {
	setEnv(t, map[string]string{EnvBaseURL: "", EnvAPIKey: "", EnvEnv: ""})
	// Simulate "HOMELAB_ENV set to empty string": setEnv with "" already clears
	// the var above, so also assert the with-empty-string path explicitly by
	// setting it to empty and expecting the file value to win.
	_ = os.Setenv(EnvEnv, "")
	t.Cleanup(func() { _ = os.Unsetenv(EnvEnv) })
	file := Config{Env: "production"}

	cfg := ResolveWithFile(file, "", "", "")

	if cfg.Env != "production" {
		t.Fatalf("env = %q, want production (empty env = unset, file wins)", cfg.Env)
	}
	if !cfg.RequireAuth {
		t.Fatal("require auth = false, want true (file env=production)")
	}
}

// TestResolveWithFile_SetsRequireAuth covers SC-CFG-003: RequireAuth is derived
// purely from the resolved Env == "production", regardless of which layer
// supplied it (flag, env, or file).
func TestResolveWithFile_SetsRequireAuth(t *testing.T) {
	setEnv(t, map[string]string{EnvBaseURL: "", EnvAPIKey: "", EnvEnv: ""})
	cases := []struct {
		name string
		file Config
		env  string
		flag string
		want bool
	}{
		{"defaults -> development, no auth", Config{}, "", "", false},
		{"file env=production -> auth", Config{Env: "production"}, "", "", true},
		{"env=production overrides file dev -> auth", Config{Env: "development"}, "production", "", true},
		{"flag=production overrides all -> auth", Config{Env: "development"}, "", "production", true},
		{"flag=development overrides prod env -> no auth", Config{}, "production", "development", false},
		{"unknown env -> no auth", Config{Env: "staging"}, "", "", false},
	}
	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			setEnv(t, map[string]string{EnvEnv: tc.env})
			cfg := ResolveWithFile(tc.file, "", "", tc.flag)
			if cfg.Env != "production" && cfg.Env != "development" && cfg.Env != "staging" && cfg.Env != "" {
				t.Fatalf("unexpected env %q", cfg.Env)
			}
			if cfg.RequireAuth != tc.want {
				t.Fatalf("require auth = %v, want %v (env=%q)", cfg.RequireAuth, tc.want, cfg.Env)
			}
		})
	}
}