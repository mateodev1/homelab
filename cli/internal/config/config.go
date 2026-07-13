package config

import "os"

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