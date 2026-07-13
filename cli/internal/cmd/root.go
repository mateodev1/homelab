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
)

// NewRootCommand builds the root homelab command with persistent flags and all
// subcommands attached. version is the build-supplied version string.
func NewRootCommand(version string) *cobra.Command {
	root := &cobra.Command{
		Use:   "homelab",
		Short: "homelab CLI — talk to the homelab backend",
		Long: "homelab is a small client for the homelab backend HTTP API.\n\n" +
			"Configure the target server with --base-url or HOMELAB_BASE_URL,\n" +
			"and authenticate with --api-key or HOMELAB_API_KEY (required for\n" +
			"every command except 'health').",
		SilenceUsage:  true,
		SilenceErrors: true,
	}

	root.PersistentFlags().String(flagBaseURL, "", "backend base URL (default http://localhost:8080, env HOMELAB_BASE_URL)")
	root.PersistentFlags().String(flagAPIKey, "", "API key sent as Bearer token (env HOMELAB_API_KEY)")
	// Setting Version makes cobra wire --version automatically.
	root.Version = version

	root.AddCommand(newHealthCmd())
	root.AddCommand(newTodoCmd())
	root.AddCommand(newProjectCmd())

	return root
}

// resolveClient reads the persistent flags from cmd, resolves config, and
// returns an HTTP client. When requireKey is true and no API key is available
// it returns an error explaining how to supply one.
func resolveClient(cmd *cobra.Command, requireKey bool) (*client.Client, error) {
	baseURL, _ := cmd.Flags().GetString(flagBaseURL)
	apiKey, _ := cmd.Flags().GetString(flagAPIKey)
	cfg := config.Resolve(baseURL, apiKey)

	if requireKey && cfg.APIKey == "" {
		return nil, fmt.Errorf("api key required: pass --api-key or set %s", config.EnvAPIKey)
	}
	return client.New(cfg.BaseURL, cfg.APIKey), nil
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