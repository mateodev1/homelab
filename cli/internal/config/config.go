package config

import (
	"crypto/rand"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

const (
	DefaultBaseURL = "http://localhost:8080"
	EnvBaseURL    = "HOMELAB_BASE_URL"
	EnvAPIKey     = "HOMELAB_API_KEY"
	EnvEnv        = "HOMELAB_ENV"
	DefaultEnv    = "development"
	EnvProduction = "production"
)

// Config holds the resolved runtime configuration for the CLI.
type Config struct {
	BaseURL string
	APIKey  string
	Env     string
	// RequireAuth mirrors the backend auth gate: true in production, false in
	// development. When false the CLI never sends an Authorization header and
	// never requires an API key — dev is fully open, matching the backend.
	RequireAuth bool
}

// Resolve applies the priority order: explicit flag values (when non-empty)
// take precedence, then the corresponding environment variables, then the
// built-in defaults. BaseURL and Env have defaults; APIKey is left empty when
// nothing provides it so callers can detect a missing key in production and
// fail clearly. RequireAuth is derived from Env (production => true).
//
// Callers should pass the flag values exactly as the user supplied them: an
// empty string means "the user did not pass this flag" and falls back to env
// or default. This keeps the priority chain in a single, testable place.
func Resolve(flagBaseURL, flagAPIKey, flagEnv string) Config {
	cfg := Config{
		BaseURL: DefaultBaseURL,
		APIKey:  "",
		Env:     DefaultEnv,
	}

	if v := os.Getenv(EnvBaseURL); v != "" {
		cfg.BaseURL = v
	}
	if v := os.Getenv(EnvAPIKey); v != "" {
		cfg.APIKey = v
	}
	if v := os.Getenv(EnvEnv); v != "" {
		cfg.Env = v
	}

	if flagBaseURL != "" {
		cfg.BaseURL = flagBaseURL
	}
	if flagAPIKey != "" {
		cfg.APIKey = flagAPIKey
	}
	if flagEnv != "" {
		cfg.Env = flagEnv
	}

	cfg.RequireAuth = cfg.Env == EnvProduction
	return cfg
}

// fileConfig is the on-disk JSON representation of the persistent config. It
// holds exactly the three persisted fields; require_auth is NEVER persisted
// (REQ-CFG-002) — it is always derived from Env by the resolver.
type fileConfig struct {
	BaseURL string `json:"base_url"`
	APIKey  string `json:"api_key"`
	Env     string `json:"env"`
}

// configDirFn is the swappable seam for the config directory (REQ-CFG-007 /
// SC-CFG-008). Tests replace it with a function returning t.TempDir() and
// restore it via t.Cleanup so no test ever touches the developer's real
// ~/.config/homelab/config.json. defaultConfigDir honors HOMELAB_CONFIG_DIR
// for ad-hoc dev convenience only — tests MUST swap the var, not rely on env.
//
// The seam is consumed by cmd.resolveClient (cli/internal/cmd/root.go); until
// that wiring lands it is intentionally unused at the package level.
var configDirFn = defaultConfigDir //nolint:unused // wired in cmd.resolveClient (T3)

// defaultConfigDir returns the directory holding config.json: HOMELAB_CONFIG_DIR
// if set, otherwise os.UserConfigDir()+"/homelab".
func defaultConfigDir() (string, error) { //nolint:unused // wired in cmd.resolveClient (T3)
	if v := os.Getenv("HOMELAB_CONFIG_DIR"); v != "" {
		return v, nil
	}
	base, err := os.UserConfigDir()
	if err != nil {
		return "", fmt.Errorf("resolve user config dir: %w", err)
	}
	return filepath.Join(base, "homelab"), nil
}

// ConfigPath returns the absolute path to the config file inside dir.
func ConfigPath(dir string) string {
	return filepath.Join(dir, "config.json")
}

// Load reads and parses the config file at ConfigPath(dir). A missing file is
// not an error: it yields a zero Config so the resolver can fall back to
// defaults (SC-CFG-001). A corrupt file yields a wrapped error.
func Load(dir string) (Config, error) {
	b, err := os.ReadFile(ConfigPath(dir))
	if err != nil {
		if os.IsNotExist(err) {
			return Config{}, nil
		}
		return Config{}, fmt.Errorf("read config file: %w", err)
	}
	var fc fileConfig
	if err := json.Unmarshal(b, &fc); err != nil {
		return Config{}, fmt.Errorf("parse config file: %w", err)
	}
	return Config{BaseURL: fc.BaseURL, APIKey: fc.APIKey, Env: fc.Env}, nil
}

// Save writes cfg to ConfigPath(dir) atomically: it marshals to a temp file in
// the same directory (0600), then os.Renames it over the target. Same-directory
// rename is atomic on POSIX and best-effort on Windows, so a crash never leaves
// a torn config.json (D7). The directory is created with 0700 if missing. On
// any error after the temp file is created, the temp is removed and a wrapped
// error is returned — no orphan tmp is left behind.
func Save(dir string, cfg Config) error {
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return fmt.Errorf("create config dir: %w", err)
	}

	b, err := json.MarshalIndent(fileConfig{
		BaseURL: cfg.BaseURL,
		APIKey:  cfg.APIKey,
		Env:     cfg.Env,
	}, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal config: %w", err)
	}
	b = append(b, '\n')

	var suffix [4]byte
	if _, err := rand.Read(suffix[:]); err != nil {
		return fmt.Errorf("generate temp suffix: %w", err)
	}
	tmp := filepath.Join(dir, ".config.json.tmp."+fmt.Sprintf("%x", suffix[:]))

	f, err := os.OpenFile(tmp, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0o600)
	if err != nil {
		return fmt.Errorf("create temp config: %w", err)
	}
	if _, err := f.Write(b); err != nil {
		_ = f.Close()
		_ = os.Remove(tmp)
		return fmt.Errorf("write temp config: %w", err)
	}
	if err := f.Close(); err != nil {
		_ = os.Remove(tmp)
		return fmt.Errorf("close temp config: %w", err)
	}

	if err := os.Rename(tmp, ConfigPath(dir)); err != nil {
		_ = os.Remove(tmp)
		return fmt.Errorf("rename temp config: %w", err)
	}
	return nil
}

// Delete removes the config file. A missing file is not an error (idempotent).
func Delete(dir string) error {
	err := os.Remove(ConfigPath(dir))
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("remove config file: %w", err)
	}
	return nil
}