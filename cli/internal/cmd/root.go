// Package cmd wires the homelab CLI subcommands onto cobra.
//
// Each file in this package registers one resource group. The root command is
// built in NewRootCommand, which installs the persistent flags, the version
// flag, and the auth resolution that runs in PersistentPreRunE.
package cmd

import (
	"encoding/json"
	"fmt"

	"github.com/mateo/homelab/cli/internal/client"
	"github.com/mateo/homelab/cli/internal/config"
	"github.com/spf13/cobra"
)

// flag names shared across commands.
const (
	flagBaseURL = "base-url"
	flagAPIKey  = "api-key"
	flagEnv     = "env"
)

// NewRootCommand builds the root homelab command with persistent flags and all
// subcommands attached. version is the build-supplied version string.
func NewRootCommand(version string) *cobra.Command {
	root := &cobra.Command{
		Use:   "homelab",
		Short: "homelab CLI — talk to the homelab backend",
		Long: "homelab is a small client for the homelab backend HTTP API.\n\n" +
			"Configure the target server with --base-url or HOMELAB_BASE_URL,\n" +
			"and authenticate with --api-key or HOMELAB_API_KEY (required in\n" +
			"production only; development is auth-free). Select the environment\n" +
			"with --env or HOMELAB_ENV (development|production, default\n" +
			"development).",
		SilenceUsage:  true,
		SilenceErrors: true,
	}

	root.PersistentFlags().String(flagBaseURL, "", "backend base URL (default http://localhost:8080, env HOMELAB_BASE_URL)")
	root.PersistentFlags().String(flagAPIKey, "", "API key sent as Bearer token in production (env HOMELAB_API_KEY)")
	root.PersistentFlags().String(flagEnv, "", "target environment: development|production (default development, env HOMELAB_ENV)")
	// Setting Version makes cobra wire --version automatically.
	root.Version = version

	root.AddCommand(newHealthCmd())
	root.AddCommand(newTodoCmd())
	root.AddCommand(newProjectCmd())

	return root
}

// resolveClient reads the persistent flags from cmd, resolves config, and
// returns an HTTP client. requireKey reflects the per-command auth need:
// health passes false (always auth-free), every other command passes true.
// The actual "api key required" error fires only when the command needs a key
// (requireKey) AND the resolved environment requires auth (cfg.RequireAuth,
// i.e. production) AND no key is available. In development RequireAuth is
// false, so no command ever requires a key — dev is fully open, matching the
// backend.
func resolveClient(cmd *cobra.Command, requireKey bool) (*client.Client, error) {
	baseURL, _ := cmd.Flags().GetString(flagBaseURL)
	apiKey, _ := cmd.Flags().GetString(flagAPIKey)
	env, _ := cmd.Flags().GetString(flagEnv)
	cfg := config.Resolve(baseURL, apiKey, env)

	if requireKey && cfg.RequireAuth && cfg.APIKey == "" {
		return nil, fmt.Errorf("api key required: pass --api-key or set %s", config.EnvAPIKey)
	}
	return client.New(cfg.BaseURL, cfg.APIKey, cfg.RequireAuth), nil
}

// printJSON pretty-prints v to the command's stdout as indented JSON. Using
// cmd.OutOrStdout() lets tests capture output by setting cobra's writer.
func printJSON(cmd *cobra.Command, v any) error {
	b, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return fmt.Errorf("format output: %w", err)
	}
	_, _ = fmt.Fprintln(cmd.OutOrStdout(), string(b))
	return nil
}