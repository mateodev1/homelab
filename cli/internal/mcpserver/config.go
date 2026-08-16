package mcpserver

import (
	"fmt"

	"github.com/mateo/homelab/cli/internal/client"
	"github.com/mateo/homelab/cli/internal/config"
)

// resolveClient loads the persistent config file (if any), merges it with
// HOMELAB_* env vars via the same precedence as the CLI (env > file > defaults),
// and returns an HTTP client. In production a missing API key fails fast with
// an actionable message pointing at `homelab login`.
func resolveClient() (*client.Client, error) {
	var file config.Config
	if dir, err := config.ConfigDirFn(); err == nil {
		// Missing/corrupt file -> zero Config; fall back to env/defaults.
		file, _ = config.Load(dir)
	}
	cfg := config.ResolveWithFile(file, "", "", "")

	if cfg.RequireAuth && cfg.APIKey == "" {
		return nil, fmt.Errorf(
			"api key required for production: run `homelab login --env production` or set %s",
			config.EnvAPIKey,
		)
	}
	return client.New(cfg.BaseURL, cfg.APIKey, cfg.RequireAuth), nil
}
