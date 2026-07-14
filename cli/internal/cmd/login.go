package cmd

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/mateo/homelab/cli/internal/config"
	"github.com/spf13/cobra"
	"golang.org/x/term"
)

// readSecretFn is the testable seam for reading the API key in production. The
// default (readSecretFromTTY) reads from os.Stdin: on a TTY it uses
// term.ReadPassword for no-echo; off a TTY it reads one line via bufio.Scanner
// (the testable piped path). Tests swap this var to a reader-backed stub so
// they never touch the real stdin and never require a TTY.
//
// The seam is package-level shared state, so prod-path login tests that swap
// it are serial within the (already serial) cmd suite.
var readSecretFn = readSecretFromTTY

// readSecretFromTTY reads a single API key from os.Stdin. On a TTY it uses
// term.ReadPassword so the key is not echoed (REQ-LOGIN-003); the no-echo path
// is exercised only via smoke/manual (design risk #1). Off a TTY (piped/
// redirected input — the testable, CI-friendly path, REQ-LOGIN-004) it reads
// one line with bufio.Scanner. An empty line / EOF yields an error so the
// caller's empty-input check (non-zero exit, REQ-LOGIN-009) fires.
func readSecretFromTTY() (string, error) {
	fd := int(os.Stdin.Fd())
	if term.IsTerminal(fd) {
		b, err := term.ReadPassword(fd)
		if err != nil {
			return "", fmt.Errorf("read api key from tty: %w", err)
		}
		return strings.TrimRight(string(b), "\r\n"), nil
	}
	sc := bufio.NewScanner(os.Stdin)
	if !sc.Scan() {
		if err := sc.Err(); err != nil {
			return "", fmt.Errorf("read api key from stdin: %w", err)
		}
		return "", errors.New("api key is required in production: no input on stdin")
	}
	return strings.TrimSpace(sc.Text()), nil
}

// newLoginCmd builds the `homelab login` command. login persists the CLI
// config file (~/.config/homelab/config.json, 0600) with env and, in
// production, the API key read (no-echo on TTY, one-line on piped stdin).
// In development no key is read and api_key is stored empty. login NEVER
// contacts the backend (REQ-LOGIN-007): it is a local-file-only operation.
//
// login inherits the root persistent flags (--env, --base-url, --api-key). It
// honors --env (Changed) and, in T5, --base-url; it ignores --api-key with a
// stderr warning (T5). Redefining the inherited flags here would shadow/panic,
// so login only reads them.
func newLoginCmd() *cobra.Command {
	return &cobra.Command{
		Use:          "login",
		Short:        "Persist CLI config (env; API key from stdin in production)",
		Long: "homelab login persists the CLI config file so other commands\n" +
			"read base_url/api_key/env from it. In development (the default) no\n" +
			"API key is read and api_key is stored empty. In production the API\n" +
			"key is read from stdin (no echo on a TTY via x/term; one line on\n" +
			"piped input). login never contacts the backend.",
		Args:         cobra.NoArgs,
		SilenceUsage: true,
		RunE:         runLogin,
	}
}

// runLogin resolves the environment, optionally reads the API key from stdin
// (production only), and writes the config file.
func runLogin(cmd *cobra.Command, _ []string) error {
	// Env resolution (REQ-LOGIN-001/002/005): explicit --env (Changed) wins;
	// else the existing config file's env; else the built-in default development.
	env := config.DefaultEnv
	if cmd.Flags().Changed(flagEnv) {
		if v, _ := cmd.Flags().GetString(flagEnv); v != "" {
			env = v
		}
	} else if dir, dirErr := config.ConfigDirFn(); dirErr == nil {
		if file, loadErr := config.Load(dir); loadErr == nil && file.Env != "" {
			env = file.Env
		}
	}
	if env != config.DefaultEnv && env != config.EnvProduction {
		return fmt.Errorf("invalid env %q: must be development or production", env)
	}

	var apiKey string
	if env == config.EnvProduction {
		key, err := readSecretFn()
		if err != nil {
			return err
		}
		key = strings.TrimSpace(key)
		if key == "" {
			return errors.New("api key is required in production: empty input on stdin")
		}
		apiKey = key
	}
	// development: no prompt, api_key stays "".

	dir, err := config.ConfigDirFn()
	if err != nil {
		return fmt.Errorf("resolve config dir: %w", err)
	}
	return config.Save(dir, config.Config{BaseURL: config.DefaultBaseURL, APIKey: apiKey, Env: env})
}