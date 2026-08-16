package cmd

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"os"

	"github.com/mateo/homelab/cli/internal/dto"
	"github.com/spf13/cobra"
)

// newSecretCmd wires the "secret" command tree: products -> secret projects ->
// environments -> secrets. Flag names are shared across subcommands so users
// only learn one vocabulary: --product-id, --project-id, --environment,
// --key/--value.
func newSecretCmd() *cobra.Command {
	secret := &cobra.Command{
		Use:          "secret",
		Short:        "Manage products, secret projects, environments, and secrets",
		Args:         cobra.NoArgs,
		SilenceUsage: true,
	}
	secret.AddCommand(newSecretProductListCmd())
	secret.AddCommand(newSecretProductCreateCmd())
	secret.AddCommand(newSecretProjectListCmd())
	secret.AddCommand(newSecretProjectCreateCmd())
	secret.AddCommand(newSecretEnvironmentListCmd())
	secret.AddCommand(newSecretListCmd())
	secret.AddCommand(newSecretCreateCmd())
	secret.AddCommand(newSecretRevealCmd())
	secret.AddCommand(newSecretUpdateCmd())
	secret.AddCommand(newSecretDeleteCmd())
	secret.AddCommand(newSecretExportCmd())
	return secret
}

// productsPath builds /api/products.
func productsPath() string {
	return "/api/products"
}

// projectsPath builds /api/products/{productID}/projects.
func projectsPath(productID int64) string {
	return fmt.Sprintf("/api/products/%d/projects", productID)
}

// environmentsPath builds /api/products/{productID}/projects/{projectID}/environments.
func environmentsPath(productID, projectID int64) string {
	return fmt.Sprintf("/api/products/%d/projects/%d/environments", productID, projectID)
}

// secretsPath builds .../environments/{envName}/secrets.
func secretsPath(productID, projectID int64, envName string) string {
	return fmt.Sprintf("/api/products/%d/projects/%d/environments/%s/secrets", productID, projectID, url.PathEscape(envName))
}

// secretKeyPath builds .../secrets/{key}.
func secretKeyPath(productID, projectID int64, envName, key string) string {
	return secretsPath(productID, projectID, envName) + "/" + url.PathEscape(key)
}

// secretRevealPath builds .../secrets/{key}/reveal.
func secretRevealPath(productID, projectID int64, envName, key string) string {
	return secretKeyPath(productID, projectID, envName, key) + "/reveal"
}

// secretExportPath builds .../environments/{envName}/export.
func secretExportPath(productID, projectID int64, envName string) string {
	return fmt.Sprintf("/api/products/%d/projects/%d/environments/%s/export", productID, projectID, url.PathEscape(envName))
}

func newSecretProductListCmd() *cobra.Command {
	return &cobra.Command{
		Use:          "product-list",
		Short:        "List products",
		Args:         cobra.NoArgs,
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			c, err := resolveClient(cmd, true)
			if err != nil {
				return err
			}
			raw, _, err := c.Do(cmd.Context(), "GET", productsPath(), nil)
			if err != nil {
				return err
			}
			var products []dto.Product
			if err := json.Unmarshal(raw, &products); err != nil {
				return fmt.Errorf("parse products: %w", err)
			}
			if products == nil {
				products = []dto.Product{}
			}
			return printJSON(cmd, products)
		},
	}
}

func newSecretProductCreateCmd() *cobra.Command {
	var name string
	c := &cobra.Command{
		Use:          "product-create",
		Short:        "Create a product",
		Args:         cobra.NoArgs,
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			if name == "" {
				return errors.New("name is required")
			}
			body, err := json.Marshal(dto.CreateProduct{Name: name})
			if err != nil {
				return fmt.Errorf("encode request: %w", err)
			}
			cl, err := resolveClient(cmd, true)
			if err != nil {
				return err
			}
			raw, _, err := cl.Do(cmd.Context(), "POST", productsPath(), body)
			if err != nil {
				return err
			}
			var p dto.Product
			if err := json.Unmarshal(raw, &p); err != nil {
				return fmt.Errorf("parse product: %w", err)
			}
			return printJSON(cmd, p)
		},
	}
	c.Flags().StringVar(&name, "name", "", "product name (required)")
	return c
}

func newSecretProjectListCmd() *cobra.Command {
	var productID int64
	c := &cobra.Command{
		Use:          "project-list",
		Short:        "List secret projects under a product",
		Args:         cobra.NoArgs,
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			if productID <= 0 {
				return errors.New("--product-id must be a positive integer")
			}
			cl, err := resolveClient(cmd, true)
			if err != nil {
				return err
			}
			raw, _, err := cl.Do(cmd.Context(), "GET", projectsPath(productID), nil)
			if err != nil {
				return err
			}
			var projects []dto.SecretProject
			if err := json.Unmarshal(raw, &projects); err != nil {
				return fmt.Errorf("parse secret projects: %w", err)
			}
			if projects == nil {
				projects = []dto.SecretProject{}
			}
			return printJSON(cmd, projects)
		},
	}
	c.Flags().Int64Var(&productID, "product-id", 0, "product id (required)")
	return c
}

func newSecretProjectCreateCmd() *cobra.Command {
	var productID int64
	var name string
	c := &cobra.Command{
		Use:          "project-create",
		Short:        "Create a secret project under a product (auto-provisions environments)",
		Args:         cobra.NoArgs,
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			if productID <= 0 {
				return errors.New("--product-id must be a positive integer")
			}
			if name == "" {
				return errors.New("name is required")
			}
			body, err := json.Marshal(dto.CreateSecretProject{Name: name})
			if err != nil {
				return fmt.Errorf("encode request: %w", err)
			}
			cl, err := resolveClient(cmd, true)
			if err != nil {
				return err
			}
			raw, _, err := cl.Do(cmd.Context(), "POST", projectsPath(productID), body)
			if err != nil {
				return err
			}
			var p dto.CreateSecretProjectResponse
			if err := json.Unmarshal(raw, &p); err != nil {
				return fmt.Errorf("parse secret project: %w", err)
			}
			return printJSON(cmd, p)
		},
	}
	c.Flags().Int64Var(&productID, "product-id", 0, "product id (required)")
	c.Flags().StringVar(&name, "name", "", "secret project name (required)")
	return c
}

func newSecretEnvironmentListCmd() *cobra.Command {
	var productID, projectID int64
	c := &cobra.Command{
		Use:          "environment-list",
		Short:        "List environments under a secret project",
		Args:         cobra.NoArgs,
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			if productID <= 0 {
				return errors.New("--product-id must be a positive integer")
			}
			if projectID <= 0 {
				return errors.New("--project-id must be a positive integer")
			}
			cl, err := resolveClient(cmd, true)
			if err != nil {
				return err
			}
			raw, _, err := cl.Do(cmd.Context(), "GET", environmentsPath(productID, projectID), nil)
			if err != nil {
				return err
			}
			var envs []dto.Environment
			if err := json.Unmarshal(raw, &envs); err != nil {
				return fmt.Errorf("parse environments: %w", err)
			}
			if envs == nil {
				envs = []dto.Environment{}
			}
			return printJSON(cmd, envs)
		},
	}
	c.Flags().Int64Var(&productID, "product-id", 0, "product id (required)")
	c.Flags().Int64Var(&projectID, "project-id", 0, "secret project id (required)")
	return c
}

// secretScopeFlags holds the flags shared by every secret-level subcommand.
type secretScopeFlags struct {
	productID   int64
	projectID   int64
	environment string
}

func addSecretScopeFlags(c *cobra.Command, f *secretScopeFlags) {
	c.Flags().Int64Var(&f.productID, "product-id", 0, "product id (required)")
	c.Flags().Int64Var(&f.projectID, "project-id", 0, "secret project id (required)")
	c.Flags().StringVar(&f.environment, "environment", "", "environment name: development|staging|production (required)")
}

func (f *secretScopeFlags) validate() error {
	if f.productID <= 0 {
		return errors.New("--product-id must be a positive integer")
	}
	if f.projectID <= 0 {
		return errors.New("--project-id must be a positive integer")
	}
	if f.environment == "" {
		return errors.New("--environment is required")
	}
	return nil
}

func newSecretListCmd() *cobra.Command {
	f := &secretScopeFlags{}
	c := &cobra.Command{
		Use:          "list",
		Short:        "List secret metadata (values masked) in an environment",
		Args:         cobra.NoArgs,
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := f.validate(); err != nil {
				return err
			}
			cl, err := resolveClient(cmd, true)
			if err != nil {
				return err
			}
			raw, _, err := cl.Do(cmd.Context(), "GET", secretsPath(f.productID, f.projectID, f.environment), nil)
			if err != nil {
				return err
			}
			var secrets []dto.SecretMetadata
			if err := json.Unmarshal(raw, &secrets); err != nil {
				return fmt.Errorf("parse secrets: %w", err)
			}
			if secrets == nil {
				secrets = []dto.SecretMetadata{}
			}
			return printJSON(cmd, secrets)
		},
	}
	addSecretScopeFlags(c, f)
	return c
}

func newSecretCreateCmd() *cobra.Command {
	f := &secretScopeFlags{}
	var key, value string
	c := &cobra.Command{
		Use:          "create",
		Short:        "Create a secret in an environment",
		Args:         cobra.NoArgs,
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := f.validate(); err != nil {
				return err
			}
			if key == "" {
				return errors.New("--key is required")
			}
			body, err := json.Marshal(dto.CreateSecret{Key: key, Value: value})
			if err != nil {
				return fmt.Errorf("encode request: %w", err)
			}
			cl, err := resolveClient(cmd, true)
			if err != nil {
				return err
			}
			raw, _, err := cl.Do(cmd.Context(), "POST", secretsPath(f.productID, f.projectID, f.environment), body)
			if err != nil {
				return err
			}
			var meta dto.SecretMetadata
			if err := json.Unmarshal(raw, &meta); err != nil {
				return fmt.Errorf("parse secret: %w", err)
			}
			return printJSON(cmd, meta)
		},
	}
	addSecretScopeFlags(c, f)
	c.Flags().StringVar(&key, "key", "", "secret key (required)")
	c.Flags().StringVar(&value, "value", "", "secret value")
	return c
}

func newSecretRevealCmd() *cobra.Command {
	f := &secretScopeFlags{}
	var key string
	c := &cobra.Command{
		Use:          "reveal",
		Short:        "Reveal the plaintext value of a secret",
		Args:         cobra.NoArgs,
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := f.validate(); err != nil {
				return err
			}
			if key == "" {
				return errors.New("--key is required")
			}
			cl, err := resolveClient(cmd, true)
			if err != nil {
				return err
			}
			raw, _, err := cl.Do(cmd.Context(), "GET", secretRevealPath(f.productID, f.projectID, f.environment, key), nil)
			if err != nil {
				return err
			}
			var resp dto.SecretRevealResponse
			if err := json.Unmarshal(raw, &resp); err != nil {
				return fmt.Errorf("parse reveal response: %w", err)
			}
			return printJSON(cmd, resp)
		},
	}
	addSecretScopeFlags(c, f)
	c.Flags().StringVar(&key, "key", "", "secret key (required)")
	return c
}

func newSecretUpdateCmd() *cobra.Command {
	f := &secretScopeFlags{}
	var key, value string
	c := &cobra.Command{
		Use:          "update",
		Short:        "Update a secret's value",
		Args:         cobra.NoArgs,
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := f.validate(); err != nil {
				return err
			}
			if key == "" {
				return errors.New("--key is required")
			}
			body, err := json.Marshal(dto.UpdateSecret{Value: value})
			if err != nil {
				return fmt.Errorf("encode request: %w", err)
			}
			cl, err := resolveClient(cmd, true)
			if err != nil {
				return err
			}
			raw, _, err := cl.Do(cmd.Context(), "PUT", secretKeyPath(f.productID, f.projectID, f.environment, key), body)
			if err != nil {
				return err
			}
			var meta dto.SecretMetadata
			if err := json.Unmarshal(raw, &meta); err != nil {
				return fmt.Errorf("parse secret: %w", err)
			}
			return printJSON(cmd, meta)
		},
	}
	addSecretScopeFlags(c, f)
	c.Flags().StringVar(&key, "key", "", "secret key (required)")
	c.Flags().StringVar(&value, "value", "", "new secret value (required)")
	return c
}

func newSecretDeleteCmd() *cobra.Command {
	f := &secretScopeFlags{}
	var key string
	var yes bool
	c := &cobra.Command{
		Use:          "delete",
		Short:        "Delete a secret",
		Args:         cobra.NoArgs,
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := f.validate(); err != nil {
				return err
			}
			if key == "" {
				return errors.New("--key is required")
			}
			if !yes {
				var resp string
				_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "delete secret %q? [y/N]: ", key)
				if _, err := fmt.Fscanln(cmd.InOrStdin(), &resp); err != nil {
					return fmt.Errorf("aborted: %w", err)
				}
				if resp != "y" && resp != "Y" && resp != "yes" {
					return errors.New("aborted")
				}
			}
			cl, err := resolveClient(cmd, true)
			if err != nil {
				return err
			}
			if _, _, err := cl.Do(cmd.Context(), "DELETE", secretKeyPath(f.productID, f.projectID, f.environment, key), nil); err != nil {
				return err
			}
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "deleted secret %q\n", key)
			return nil
		},
	}
	addSecretScopeFlags(c, f)
	c.Flags().StringVar(&key, "key", "", "secret key (required)")
	c.Flags().BoolVarP(&yes, "yes", "y", false, "skip confirmation prompt")
	return c
}

func newSecretExportCmd() *cobra.Command {
	f := &secretScopeFlags{}
	var output string
	c := &cobra.Command{
		Use:          "export",
		Short:        "Export all secrets in an environment as dotenv",
		Args:         cobra.NoArgs,
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := f.validate(); err != nil {
				return err
			}
			cl, err := resolveClient(cmd, true)
			if err != nil {
				return err
			}
			raw, _, err := cl.Do(cmd.Context(), "GET", secretExportPath(f.productID, f.projectID, f.environment), nil)
			if err != nil {
				return err
			}
			if output != "" {
				if err := os.WriteFile(output, raw, 0o600); err != nil {
					return fmt.Errorf("write output file: %w", err)
				}
				_, _ = fmt.Fprintf(cmd.OutOrStdout(), "wrote %s\n", output)
				return nil
			}
			_, _ = fmt.Fprint(cmd.OutOrStdout(), string(raw))
			return nil
		},
	}
	addSecretScopeFlags(c, f)
	c.Flags().StringVar(&output, "output", "", "write dotenv output to this file instead of stdout")
	return c
}
