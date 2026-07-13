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
	})
	cfg := Resolve("", "")
	if cfg.BaseURL != DefaultBaseURL {
		t.Fatalf("base url = %q, want %q", cfg.BaseURL, DefaultBaseURL)
	}
	if cfg.APIKey != "" {
		t.Fatalf("api key = %q, want empty", cfg.APIKey)
	}
}

func TestResolveEnvFallback(t *testing.T) {
	setEnv(t, map[string]string{
		EnvBaseURL: "http://example.test:9090",
		EnvAPIKey:  "env-secret",
	})
	cfg := Resolve("", "")
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
	})
	cfg := Resolve("http://flag.example", "flag-secret")
	if cfg.BaseURL != "http://flag.example" {
		t.Fatalf("base url = %q", cfg.BaseURL)
	}
	if cfg.APIKey != "flag-secret" {
		t.Fatalf("api key = %q", cfg.APIKey)
	}
}