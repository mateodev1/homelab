package cmd

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"

	"github.com/mateo/homelab/cli/internal/dto"
	"github.com/spf13/cobra"
)

func newProjectCmd() *cobra.Command {
	project := &cobra.Command{
		Use:          "project",
		Short:        "Manage projects",
		Args:         cobra.NoArgs,
		SilenceUsage: true,
	}
	project.AddCommand(newProjectListCmd())
	project.AddCommand(newProjectGetCmd())
	project.AddCommand(newProjectAddCmd())
	project.AddCommand(newProjectUpdateCmd())
	project.AddCommand(newProjectDeleteCmd())
	return project
}

func newProjectListCmd() *cobra.Command {
	return &cobra.Command{
		Use:          "list",
		Short:        "List projects",
		Args:         cobra.NoArgs,
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			c, err := resolveClient(cmd, true)
			if err != nil {
				return err
			}
			raw, _, err := c.Do(cmd.Context(), "GET", "/api/projects", nil)
			if err != nil {
				return err
			}
			var projects []dto.Project
			if err := json.Unmarshal(raw, &projects); err != nil {
				return fmt.Errorf("parse projects: %w", err)
			}
			if projects == nil {
				projects = []dto.Project{}
			}
			return printJSON(cmd,projects)
		},
	}
}

func newProjectGetCmd() *cobra.Command {
	return &cobra.Command{
		Use:          "get <id>",
		Short:        "Get a single project by id",
		Args:         cobra.ExactArgs(1),
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			id, err := parseID(args[0])
			if err != nil {
				return err
			}
			c, err := resolveClient(cmd, true)
			if err != nil {
				return err
			}
			raw, _, err := c.Do(cmd.Context(), "GET", "/api/projects/"+strconv.FormatInt(id, 10), nil)
			if err != nil {
				return err
			}
			var p dto.Project
			if err := json.Unmarshal(raw, &p); err != nil {
				return fmt.Errorf("parse project: %w", err)
			}
			return printJSON(cmd,p)
		},
	}
}

func newProjectAddCmd() *cobra.Command {
	var name, color string
	c := &cobra.Command{
		Use:          "add",
		Short:        "Create a project",
		Args:         cobra.NoArgs,
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			if name == "" {
				return errors.New("name is required")
			}
			payload := dto.CreateProject{Name: name, Color: color}
			body, err := json.Marshal(payload)
			if err != nil {
				return fmt.Errorf("encode request: %w", err)
			}
			c, err := resolveClient(cmd, true)
			if err != nil {
				return err
			}
			raw, _, err := c.Do(cmd.Context(), "POST", "/api/projects", body)
			if err != nil {
				return err
			}
			var p dto.Project
			if err := json.Unmarshal(raw, &p); err != nil {
				return fmt.Errorf("parse project: %w", err)
			}
			return printJSON(cmd,p)
		},
	}
	c.Flags().StringVar(&name, "name", "", "project name (required)")
	c.Flags().StringVar(&color, "color", "", "project color")
	return c
}

func newProjectUpdateCmd() *cobra.Command {
	var name, color string
	c := &cobra.Command{
		Use:          "update <id>",
		Short:        "Update a project (partial patch)",
		Args:         cobra.ExactArgs(1),
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			id, err := parseID(args[0])
			if err != nil {
				return err
			}
			patch := dto.UpdateProject{}
			changed := false
			if cmd.Flags().Changed("name") {
				patch.Name = &name
				changed = true
			}
			if cmd.Flags().Changed("color") {
				patch.Color = &color
				changed = true
			}
			if !changed {
				return errors.New("no fields specified: pass at least one flag to update")
			}
			body, err := json.Marshal(patch)
			if err != nil {
				return fmt.Errorf("encode request: %w", err)
			}
			c, err := resolveClient(cmd, true)
			if err != nil {
				return err
			}
			raw, _, err := c.Do(cmd.Context(), "PUT", "/api/projects/"+strconv.FormatInt(id, 10), body)
			if err != nil {
				return err
			}
			var p dto.Project
			if err := json.Unmarshal(raw, &p); err != nil {
				return fmt.Errorf("parse project: %w", err)
			}
			return printJSON(cmd,p)
		},
	}
	c.Flags().StringVar(&name, "name", "", "set name")
	c.Flags().StringVar(&color, "color", "", "set color")
	return c
}

func newProjectDeleteCmd() *cobra.Command {
	var yes bool
	c := &cobra.Command{
		Use:          "delete <id>",
		Short:        "Delete a project (orphaning its todos)",
		Args:         cobra.ExactArgs(1),
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			id, err := parseID(args[0])
			if err != nil {
				return err
			}
			_, _ = fmt.Fprintln(cmd.ErrOrStderr(), "warning: deleting a project nulls project_id on todos that reference it")
			if !yes {
				var resp string
				_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "delete project %d? [y/N]: ", id)
				if _, err := fmt.Fscanln(cmd.InOrStdin(), &resp); err != nil {
					return fmt.Errorf("aborted: %w", err)
				}
				if resp != "y" && resp != "Y" && resp != "yes" {
					return errors.New("aborted")
				}
			}
			c, err := resolveClient(cmd, true)
			if err != nil {
				return err
			}
			if _, _, err := c.Do(cmd.Context(), "DELETE", "/api/projects/"+strconv.FormatInt(id, 10), nil); err != nil {
				return err
			}
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "deleted project %d\n", id)
			return nil
		},
	}
	c.Flags().BoolVarP(&yes, "yes", "y", false, "skip confirmation prompt")
	return c
}