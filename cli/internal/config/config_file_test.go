package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

// These tests exercise the file primitives directly with an explicit dir
// argument (t.TempDir). Because Load/Save/Delete/ConfigPath all take dir as a
// parameter, they never consult the package-level configDirFn seam here, so
// these tests never touch the developer's real ~/.config/homelab/config.json
// (REQ-CFG-007) and can run fully parallel without racing on the seam var. The
// configDirFn seam is only swapped where production code calls it — the cmd
// package integration tests (see cli/internal/cmd/cmd_test.go, login_test.go).

// TestConfigPath asserts ConfigPath joins the dir with the conventional
// "config.json" basename.
func TestConfigPath(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()

	got := ConfigPath(dir)
	want := filepath.Join(dir, "config.json")
	if got != want {
		t.Fatalf("ConfigPath(%q) = %q, want %q", dir, got, want)
	}
}

// TestLoad_MissingFileReturnsZero covers SC-CFG-001: with no file on disk Load
// returns a zero Config and no error — resolution then falls back to defaults.
func TestLoad_MissingFileReturnsZero(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()

	cfg, err := Load(dir)
	if err != nil {
		t.Fatalf("Load missing file: unexpected error %v", err)
	}
	if cfg.BaseURL != "" || cfg.APIKey != "" || cfg.Env != "" {
		t.Fatalf("Load missing file = %+v, want zero Config", cfg)
	}
	if cfg.RequireAuth {
		t.Fatal("RequireAuth must be false for a zero (unloaded) Config")
	}
}

// TestLoad_InvalidJSONReturnsError: a corrupt config file yields a wrapped
// error rather than silently treating the file as missing.
func TestLoad_InvalidJSONReturnsError(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()

	if err := os.WriteFile(ConfigPath(dir), []byte("{not json"), 0o600); err != nil {
		t.Fatalf("seed corrupt file: %v", err)
	}

	_, err := Load(dir)
	if err == nil {
		t.Fatal("Load of invalid JSON must return an error")
	}
}

// TestSave_RoundTrip verifies SC-CFG-002 (persisted fields, no require_auth):
// Save writes base_url/api_key/env and Load reads them back, and the file on
// disk contains no require_auth key.
func TestSave_RoundTrip(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()

	in := Config{BaseURL: "http://file.example", APIKey: "file-secret", Env: "production"}
	if err := Save(dir, in); err != nil {
		t.Fatalf("Save: %v", err)
	}

	out, err := Load(dir)
	if err != nil {
		t.Fatalf("Load after Save: %v", err)
	}
	if out.BaseURL != in.BaseURL || out.APIKey != in.APIKey || out.Env != in.Env {
		t.Fatalf("round trip mismatch: got %+v, want %+v", out, in)
	}
	// Load does not derive RequireAuth (that is the resolver's job); the file
	// value round-trips as raw fields, derivation happens in ResolveWithFile.
	if out.RequireAuth {
		t.Fatal("Load must not set RequireAuth from the file")
	}

	// Inspect the raw file: require_auth MUST NOT be persisted (SC-CFG-002).
	raw, err := os.ReadFile(ConfigPath(dir))
	if err != nil {
		t.Fatalf("read back file: %v", err)
	}
	var rawFields map[string]json.RawMessage
	if err := json.Unmarshal(raw, &rawFields); err != nil {
		t.Fatalf("file is not a JSON object: %v\n%s", err, raw)
	}
	if _, ok := rawFields["require_auth"]; ok {
		t.Errorf("require_auth must not be persisted; file=%s", raw)
	}
	for _, key := range []string{"base_url", "api_key", "env"} {
		if _, ok := rawFields[key]; !ok {
			t.Errorf("persisted file missing key %q; file=%s", key, raw)
		}
	}
}

// TestSave_CreatesDir0700: when the config directory does not exist, Save
// creates it (and any missing parents) with mode 0700.
func TestSave_CreatesDir0700(t *testing.T) {
	t.Parallel()
	if runtime.GOOS == "windows" {
		t.Skip("perm bits are best-effort on Windows; mkdir mode asserted on POSIX")
	}
	base := t.TempDir()
	dir := filepath.Join(base, "nested", "homelab")

	if err := Save(dir, Config{Env: "development"}); err != nil {
		t.Fatalf("Save into missing dir: %v", err)
	}
	info, err := os.Stat(dir)
	if err != nil {
		t.Fatalf("dir not created: %v", err)
	}
	if got := info.Mode().Perm(); got != 0o700 {
		t.Fatalf("created dir perm = %o, want 0700", got)
	}
}

// TestSave_FilePerm0600_Linux covers SC-CFG-005a / REQ-CFG-001: the written
// config.json has owner-only read/write permissions (0600). Asserted on POSIX
// only — on Windows the perm bit is best-effort.
func TestSave_FilePerm0600_Linux(t *testing.T) {
	t.Parallel()
	if runtime.GOOS == "windows" {
		t.Skip("0600 perm is best-effort on Windows; asserted on POSIX/CI")
	}
	dir := t.TempDir()

	if err := Save(dir, Config{Env: "development"}); err != nil {
		t.Fatalf("Save: %v", err)
	}
	info, err := os.Stat(ConfigPath(dir))
	if err != nil {
		t.Fatalf("stat config: %v", err)
	}
	if got := info.Mode().Perm(); got != 0o600 {
		t.Fatalf("config file perm = %o, want 0600", got)
	}
}

// TestDelete_Idempotent: Delete is a no-op when the file is missing, removes it
// when present, and is again a no-op on a second call.
func TestDelete_Idempotent(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()

	if err := Delete(dir); err != nil {
		t.Fatalf("Delete missing file: unexpected error %v", err)
	}
	if err := Save(dir, Config{Env: "development"}); err != nil {
		t.Fatalf("Save: %v", err)
	}
	if err := Delete(dir); err != nil {
		t.Fatalf("Delete existing file: %v", err)
	}
	if _, err := os.Stat(ConfigPath(dir)); !os.IsNotExist(err) {
		t.Fatalf("file still present after Delete: stat err=%v", err)
	}
	if err := Delete(dir); err != nil {
		t.Fatalf("Delete after delete (idempotent): unexpected error %v", err)
	}
}

// TestSave_AtomicRenameNoTmpOnFailure verifies D7: when the atomic rename step
// fails, Save removes the temp file and returns an error — no orphan tmp is
// left behind. We force the rename to fail by pre-creating the target path as
// a directory, which makes rename(file, dir) return ENOTDIR.
func TestSave_AtomicRenameNoTmpOnFailure(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()

	// Pre-create the target as a directory so the rename into config.json fails.
	if err := os.Mkdir(ConfigPath(dir), 0o700); err != nil {
		t.Fatalf("seed blocking dir: %v", err)
	}

	err := Save(dir, Config{Env: "development"})
	if err == nil {
		t.Fatal("Save must fail when the rename target is a directory")
	}

	// No temp file may remain after the failure.
	matches, globErr := filepath.Glob(filepath.Join(dir, ".config.json.tmp.*"))
	if globErr != nil {
		t.Fatalf("glob leftover tmp: %v", globErr)
	}
	if len(matches) != 0 {
		t.Errorf("orphan temp file left behind after failed Save: %v", matches)
	}
	// The blocking directory must be untouched (clean state, no clobber).
	info, statErr := os.Stat(ConfigPath(dir))
	if statErr != nil {
		t.Fatalf("stat config path after failure: %v", statErr)
	}
	if !info.IsDir() {
		t.Errorf("blocking directory was replaced/clobbered after failed Save")
	}
}

