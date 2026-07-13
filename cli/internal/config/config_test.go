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