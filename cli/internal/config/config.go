package config

import "os"

const (
	DefaultBaseURL = "http://localhost:8080"
	EnvBaseURL     = "HOMELAB_BASE_URL"
	EnvAPIKey      = "HOMELAB_API_KEY"
)

// Config holds the resolved runtime configuration for the CLI.
type Config struct {
	BaseURL string
	APIKey  string
}

// Resolve applies the priority order: explicit flag values (when non-empty)
// take precedence, then the corresponding environment variables, then the
// built-in defaults. Only BaseURL has a default; APIKey is left empty when
// nothing provides it so callers can detect a missing key and fail clearly.
//
// Callers should pass the flag values exactly as the user supplied them: an
// empty string means "the user did not pass this flag" and falls back to env
// or default. This keeps the priority chain in a single, testable place.
func Resolve(flagBaseURL, flagAPIKey string) Config {
	cfg := Config{
		BaseURL: DefaultBaseURL,
		APIKey:  "",
	}

	if v := os.Getenv(EnvBaseURL); v != "" {
		cfg.BaseURL = v
	}
	if v := os.Getenv(EnvAPIKey); v != "" {
		cfg.APIKey = v
	}

	if flagBaseURL != "" {
		cfg.BaseURL = flagBaseURL
	}
	if flagAPIKey != "" {
		cfg.APIKey = flagAPIKey
	}

	return cfg
}