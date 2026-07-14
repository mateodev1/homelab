package cmd

import (
	"bufio"
	"bytes"
	"errors"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/mateo/homelab/cli/internal/config"
)

// stubSecretReader returns a readSecretFn that reads one line from r (mirroring
// the non-TTY bufio.Scanner contract of readSecretFromTTY). An empty/EOF read
// yields ("", nil) so the command's own empty-input check rejects it — this
// keeps the unit test focused on the command wiring, not on the TTY path.
func stubSecretReader(r io.Reader) func() (string, error) {
	return func() (string, error) {
		sc := bufio.NewScanner(r)
		if !sc.Scan() {
			if err := sc.Err(); err != nil {
				return "", err
			}
			return "", nil
		}
		return strings.TrimSpace(sc.Text()), nil
	}
}

// swapReadSecret replaces the package-level readSecretFn seam for the test and
// restores it on cleanup. Login MUST swap this so prod-path tests never touch
// the real stdin or require a TTY.
func swapReadSecret(t *testing.T, fn func() (string, error)) {
	t.Helper()
	prev := readSecretFn
	readSecretFn = fn
	t.Cleanup(func() { readSecretFn = prev })
}

// loadLoggedConfig reads the config file from dir (assumes the seam points there).
func loadLoggedConfig(t *testing.T, dir string) config.Config {
	t.Helper()
	cfg, err := config.Load(dir)
	if err != nil {
		t.Fatalf("load config after login: %v", err)
	}
	return cfg
}

// TestLogin_DevNoPromptStoresEmptyKey covers SC-LOGIN-001 / REQ-LOGIN-001: with
// no --env and no config file, login resolves env="development", does NOT
// prompt for an API key (readSecretFn is never called), writes api_key="" and
// env="development", and exits 0.
func TestLogin_DevNoPromptStoresEmptyKey(t *testing.T) {
	dir := pinConfigDir(t)
	swapReadSecret(t, func() (string, error) {
		t.Error("readSecretFn must not be called in development (no prompt)")
		return "", errors.New("unexpected prompt in dev")
	})

	root := NewRootCommand("test")
	root.SetArgs([]string{"login"})
	if err := root.Execute(); err != nil {
		t.Fatalf("dev login: %v", err)
	}

	cfg := loadLoggedConfig(t, dir)
	if cfg.Env != "development" {
		t.Fatalf("env = %q, want development", cfg.Env)
	}
	if cfg.APIKey != "" {
		t.Fatalf("api key = %q, want empty in dev", cfg.APIKey)
	}
	if cfg.BaseURL != config.DefaultBaseURL {
		t.Fatalf("base url = %q, want default %q", cfg.BaseURL, config.DefaultBaseURL)
	}
	if cfg.RequireAuth {
		t.Fatal("RequireAuth must be false for dev, but it is not persisted anyway")
	}
}

// TestLogin_ProdNonTTYReadsStdin covers SC-LOGIN-004 / REQ-LOGIN-004: with
// --env=production and a piped (non-TTY) stdin of "secret\n", login reads the
// key via the readSecretFn seam (bufio.Scanner contract) and stores it.
func TestLogin_ProdNonTTYReadsStdin(t *testing.T) {
	dir := pinConfigDir(t)
	swapReadSecret(t, stubSecretReader(strings.NewReader("secret\n")))

	root := NewRootCommand("test")
	root.SetArgs([]string{"login", "--env", "production"})
	if err := root.Execute(); err != nil {
		t.Fatalf("prod login: %v", err)
	}

	cfg := loadLoggedConfig(t, dir)
	if cfg.Env != "production" {
		t.Fatalf("env = %q, want production", cfg.Env)
	}
	if cfg.APIKey != "secret" {
		t.Fatalf("api key = %q, want 'secret'", cfg.APIKey)
	}
}

// TestLogin_ProdOverridesFileEnv covers SC-LOGIN-005 / REQ-LOGIN-005: an
// explicit --env=production overrides a config file whose env="development".
func TestLogin_ProdOverridesFileEnv(t *testing.T) {
	dir := pinConfigDir(t)
	if err := config.Save(dir, config.Config{Env: "development", BaseURL: "http://file"}); err != nil {
		t.Fatalf("seed file: %v", err)
	}
	swapReadSecret(t, stubSecretReader(strings.NewReader("prodkey\n")))

	root := NewRootCommand("test")
	root.SetArgs([]string{"login", "--env", "production"})
	if err := root.Execute(); err != nil {
		t.Fatalf("login --env=production over file dev: %v", err)
	}

	cfg := loadLoggedConfig(t, dir)
	if cfg.Env != "production" {
		t.Fatalf("env = %q, want production (flag overrides file)", cfg.Env)
	}
	if cfg.APIKey != "prodkey" {
		t.Fatalf("api key = %q, want prodkey", cfg.APIKey)
	}
}

// TestLogin_OmittedEnvFallsBackToFileEnv covers SC-LOGIN-002 / REQ-LOGIN-002:
// when --env is NOT passed and a config file has env="production", login uses
// the file's env and reads the key.
func TestLogin_OmittedEnvFallsBackToFileEnv(t *testing.T) {
	dir := pinConfigDir(t)
	if err := config.Save(dir, config.Config{Env: "production"}); err != nil {
		t.Fatalf("seed file: %v", err)
	}
	swapReadSecret(t, stubSecretReader(strings.NewReader("fileenvkey\n")))

	root := NewRootCommand("test")
	root.SetArgs([]string{"login"})
	if err := root.Execute(); err != nil {
		t.Fatalf("login with file env=production: %v", err)
	}

	cfg := loadLoggedConfig(t, dir)
	if cfg.Env != "production" {
		t.Fatalf("env = %q, want production (from file)", cfg.Env)
	}
	if cfg.APIKey != "fileenvkey" {
		t.Fatalf("api key = %q, want fileenvkey", cfg.APIKey)
	}
}

// TestLogin_WritesFilePerm0600 covers SC-LOGIN-006 / SC-CFG-005a /
// REQ-LOGIN-006: a successful (dev) login writes config.json with mode 0600 on
// a POSIX platform.
func TestLogin_WritesFilePerm0600(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("0600 perm is best-effort on Windows; asserted on POSIX/CI")
	}
	dir := pinConfigDir(t)
	swapReadSecret(t, func() (string, error) {
		t.Error("dev login must not prompt")
		return "", errors.New("unexpected")
	})

	root := NewRootCommand("test")
	root.SetArgs([]string{"login"})
	if err := root.Execute(); err != nil {
		t.Fatalf("dev login: %v", err)
	}

	info, err := os.Stat(filepath.Join(dir, "config.json"))
	if err != nil {
		t.Fatalf("config file not written: %v", err)
	}
	if got := info.Mode().Perm(); got != 0o600 {
		t.Fatalf("config file perm = %o, want 0600", got)
	}
}

// TestLogin_NoBackendContact covers SC-LOGIN-007 / REQ-LOGIN-007: login MUST
// NOT import the client package or net/http, and running it succeeds with no
// backend present (login is a local-file-only operation).
func TestLogin_NoBackendContact(t *testing.T) {
	// Static guard: login.go must not import client or net/http.
	src, err := os.ReadFile("login.go")
	if err != nil {
		t.Fatalf("read login.go: %v", err)
	}
	if bytes.Contains(src, []byte(`"net/http"`)) {
		t.Error(`login.go imports net/http — LOGIN-007 forbids backend contact`)
	}
	if bytes.Contains(src, []byte("homelab/cli/internal/client")) {
		t.Error("login.go imports the client package — LOGIN-007 forbids backend contact")
	}

	// Dynamic guard: run dev login with no backend; it must succeed and write
	// an empty-key dev config entirely locally.
	dir := pinConfigDir(t)
	swapReadSecret(t, func() (string, error) {
		t.Error("dev login must not prompt")
		return "", errors.New("unexpected")
	})
	root := NewRootCommand("test")
	root.SetArgs([]string{"login"})
	if err := root.Execute(); err != nil {
		t.Fatalf("dev login with no backend: %v", err)
	}
	cfg := loadLoggedConfig(t, dir)
	if cfg.Env != "development" || cfg.APIKey != "" {
		t.Fatalf("dev login wrote %+v, want development/empty-key", cfg)
	}
}

// TestLogin_ProdEmptyInputExitsNonZero covers SC-LOGIN-009 / REQ-LOGIN-009:
// env=production with an empty line on stdin must error (non-zero), and must
// NOT write a config file.
func TestLogin_ProdEmptyInputExitsNonZero(t *testing.T) {
	dir := pinConfigDir(t)
	swapReadSecret(t, stubSecretReader(strings.NewReader("\n")))

	root := NewRootCommand("test")
	root.SetArgs([]string{"login", "--env", "production"})
	err := root.Execute()
	if err == nil {
		t.Fatal("expected non-zero exit for empty production input, got nil")
	}

	// No config file must be written on failure.
	if _, statErr := os.Stat(filepath.Join(dir, "config.json")); !os.IsNotExist(statErr) {
		t.Fatalf("config file must NOT be written on login failure: stat err=%v", statErr)
	}
}