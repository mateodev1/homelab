package cmd

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/mateo/homelab/cli/internal/dto"
	"github.com/spf13/cobra"
)

// Validated domain enums mirrored from the backend. Keeping these as sets means
// the CLI fails fast with a clear message instead of round-tripping a 400.
var (
	validStatus    = map[string]bool{"todo": true, "in_progress": true, "done": true, "cancelled": true}
	validKind      = map[string]bool{"note": true, "issue": true}
	validIssueType = map[string]bool{"feature": true, "bug": true, "improvement": true}
)

var errIssueTypeNeedsIssue = errors.New("issue_type can only be set when kind is issue")

func newTodoCmd() *cobra.Command {
	todo := &cobra.Command{
		Use:          "todo",
		Short:        "Manage todos",
		Args:         cobra.NoArgs,
		SilenceUsage: true,
	}
	todo.AddCommand(newTodoListCmd())
	todo.AddCommand(newTodoGetCmd())
	todo.AddCommand(newTodoAddCmd())
	todo.AddCommand(newTodoUpdateCmd())
	todo.AddCommand(newTodoDoneCmd())
	todo.AddCommand(newTodoDeleteCmd())
	return todo
}

func newTodoListCmd() *cobra.Command {
	var kind, issueType, projectID string
	c := &cobra.Command{
		Use:          "list",
		Short:        "List todos (optional filters)",
		Args:         cobra.NoArgs,
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			if kind != "" && !validKind[kind] {
				return fmt.Errorf("invalid --kind %q: must be note or issue", kind)
			}
			if issueType != "" && !validIssueType[issueType] {
				return fmt.Errorf("invalid --issue-type %q: must be feature, bug, or improvement", issueType)
			}
			if issueType != "" && kind != "" && kind != "issue" {
				return errIssueTypeNeedsIssue
			}
			path := "/api/todos"
			q := []string{}
			if kind != "" {
				q = append(q, "kind="+kind)
			}
			if issueType != "" {
				q = append(q, "issue_type="+issueType)
			}
			if projectID != "" {
				q = append(q, "project_id="+projectID)
			}
			if len(q) > 0 {
				path += "?" + strings.Join(q, "&")
			}

			c, err := resolveClient(cmd, true)
			if err != nil {
				return err
			}
			raw, _, err := c.Do(cmd.Context(), "GET", path, nil)
			if err != nil {
				return err
			}
			var todos []dto.Todo
			if err := json.Unmarshal(raw, &todos); err != nil {
				return fmt.Errorf("parse todos: %w", err)
			}
			if todos == nil {
				todos = []dto.Todo{}
			}
			return printJSON(cmd,todos)
		},
	}
	c.Flags().StringVar(&kind, "kind", "", "filter by kind: note|issue")
	c.Flags().StringVar(&issueType, "issue-type", "", "filter by issue_type: feature|bug|improvement")
	c.Flags().StringVar(&projectID, "project-id", "", "filter by project_id: <n> or none")
	return c
}

func newTodoGetCmd() *cobra.Command {
	return &cobra.Command{
		Use:          "get <id>",
		Short:        "Get a single todo by id",
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
			raw, _, err := c.Do(cmd.Context(), "GET", "/api/todos/"+strconv.FormatInt(id, 10), nil)
			if err != nil {
				return err
			}
			var t dto.Todo
			if err := json.Unmarshal(raw, &t); err != nil {
				return fmt.Errorf("parse todo: %w", err)
			}
			return printJSON(cmd,t)
		},
	}
}

func newTodoAddCmd() *cobra.Command {
	var (
		title, body, dueDate, kind, issueType string
		priority                              int
		projectID                             string
	)
	c := &cobra.Command{
		Use:          "add",
		Short:        "Create a todo",
		Args:         cobra.NoArgs,
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			if title == "" {
				return errors.New("title is required")
			}
			if priority < 0 || priority > 3 {
				return fmt.Errorf("invalid --priority %d: must be 0..3", priority)
			}
			if kind != "" && !validKind[kind] {
				return fmt.Errorf("invalid --kind %q: must be note or issue", kind)
			}
			if issueType != "" && !validIssueType[issueType] {
				return fmt.Errorf("invalid --issue-type %q: must be feature, bug, or improvement", issueType)
			}
			if issueType != "" && kind != "" && kind != "issue" {
				return errIssueTypeNeedsIssue
			}

			payload := dto.CreateTodo{
				Title:    title,
				Body:     body,
				Priority: priority,
				Kind:     kind,
			}
			if dueDate != "" {
				payload.DueDate = &dueDate
			}
			if issueType != "" {
				payload.IssueType = &issueType
			}
			if projectID != "" {
				pid, err := parseID(projectID)
				if err != nil {
					return fmt.Errorf("invalid --project-id %q: must be an integer", projectID)
				}
				payload.ProjectID = &pid
			}

			reqBody, err := json.Marshal(payload)
			if err != nil {
				return fmt.Errorf("encode request: %w", err)
			}
			c, err := resolveClient(cmd, true)
			if err != nil {
				return err
			}
			raw, _, err := c.Do(cmd.Context(), "POST", "/api/todos", reqBody)
			if err != nil {
				return err
			}
			var t dto.Todo
			if err := json.Unmarshal(raw, &t); err != nil {
				return fmt.Errorf("parse todo: %w", err)
			}
			return printJSON(cmd,t)
		},
	}
	c.Flags().StringVar(&title, "title", "", "todo title (required)")
	c.Flags().StringVar(&body, "body", "", "todo body")
	c.Flags().IntVar(&priority, "priority", 0, "priority 0..3")
	c.Flags().StringVar(&dueDate, "due-date", "", "due date (RFC3339)")
	c.Flags().StringVar(&kind, "kind", "", "kind: note|issue")
	c.Flags().StringVar(&issueType, "issue-type", "", "issue_type: feature|bug|improvement (requires kind=issue)")
	c.Flags().StringVar(&projectID, "project-id", "", "project id to associate")
	return c
}

func newTodoUpdateCmd() *cobra.Command {
	var (
		title, body, status, dueDate, kind, issueType, projectID string
		priorityStr                                               string
	)
	c := &cobra.Command{
		Use:          "update <id>",
		Short:        "Update a todo (partial patch)",
		Args:         cobra.ExactArgs(1),
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			id, err := parseID(args[0])
			if err != nil {
				return err
			}

			changed := false
			patch := dto.UpdateTodo{}

			if cmd.Flags().Changed("title") {
				patch.Title = &title
				changed = true
			}
			if cmd.Flags().Changed("body") {
				patch.Body = &body
				changed = true
			}
			if cmd.Flags().Changed("status") {
				if !validStatus[status] {
					return fmt.Errorf("invalid --status %q: must be todo, in_progress, done, or cancelled", status)
				}
				patch.Status = &status
				changed = true
			}
			if cmd.Flags().Changed("priority") {
				p, err := strconv.Atoi(priorityStr)
				if err != nil || p < 0 || p > 3 {
					return fmt.Errorf("invalid --priority %q: must be 0..3", priorityStr)
				}
				patch.Priority = &p
				changed = true
			}
			if cmd.Flags().Changed("kind") {
				if !validKind[kind] {
					return fmt.Errorf("invalid --kind %q: must be note or issue", kind)
				}
				patch.Kind = &kind
				changed = true
			}
if cmd.Flags().Changed("due-date") {
			if dueDate == "null" {
				var nilPtr *string
				patch.DueDate = &nilPtr
			} else {
				v := dueDate
				vPtr := &v
				patch.DueDate = &vPtr
			}
			changed = true
		}
		if cmd.Flags().Changed("issue-type") {
			if issueType != "null" && !validIssueType[issueType] {
				return fmt.Errorf("invalid --issue-type %q: must be feature, bug, improvement, or null", issueType)
			}
			if issueType == "null" {
				var nilPtr *string
				patch.IssueType = &nilPtr
			} else {
				v := issueType
				vPtr := &v
				patch.IssueType = &vPtr
			}
			changed = true
		}
		if cmd.Flags().Changed("project-id") {
			if projectID == "null" {
				var nilPtr *int64
				patch.ProjectID = &nilPtr
			} else {
				pid, err := parseID(projectID)
				if err != nil {
					return fmt.Errorf("invalid --project-id %q: must be an integer or null", projectID)
				}
				pidPtr := &pid
				patch.ProjectID = &pidPtr
			}
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
			raw, _, err := c.Do(cmd.Context(), "PUT", "/api/todos/"+strconv.FormatInt(id, 10), body)
			if err != nil {
				return err
			}
			var t dto.Todo
			if err := json.Unmarshal(raw, &t); err != nil {
				return fmt.Errorf("parse todo: %w", err)
			}
			return printJSON(cmd,t)
		},
	}
	c.Flags().StringVar(&title, "title", "", "set title")
	c.Flags().StringVar(&body, "body", "", "set body")
	c.Flags().StringVar(&status, "status", "", "set status: todo|in_progress|done|cancelled")
	c.Flags().StringVar(&priorityStr, "priority", "", "set priority 0..3")
	c.Flags().StringVar(&dueDate, "due-date", "", "set due date (RFC3339) or null to clear")
	c.Flags().StringVar(&kind, "kind", "", "set kind: note|issue")
	c.Flags().StringVar(&issueType, "issue-type", "", "set issue_type (feature|bug|improvement) or null to clear")
	c.Flags().StringVar(&projectID, "project-id", "", "set project id or null to clear")
	return c
}

func newTodoDoneCmd() *cobra.Command {
	return &cobra.Command{
		Use:          "done <id>",
		Short:        "Mark a todo done",
		Args:         cobra.ExactArgs(1),
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			id, err := parseID(args[0])
			if err != nil {
				return err
			}
			status := "done"
			patch := dto.UpdateTodo{Status: &status}
			body, err := json.Marshal(patch)
			if err != nil {
				return fmt.Errorf("encode request: %w", err)
			}
			c, err := resolveClient(cmd, true)
			if err != nil {
				return err
			}
			raw, _, err := c.Do(cmd.Context(), "PUT", "/api/todos/"+strconv.FormatInt(id, 10), body)
			if err != nil {
				return err
			}
			var t dto.Todo
			if err := json.Unmarshal(raw, &t); err != nil {
				return fmt.Errorf("parse todo: %w", err)
			}
			return printJSON(cmd,t)
		},
	}
}

func newTodoDeleteCmd() *cobra.Command {
	return &cobra.Command{
		Use:          "delete <id>",
		Short:        "Delete a todo",
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
			if _, _, err := c.Do(cmd.Context(), "DELETE", "/api/todos/"+strconv.FormatInt(id, 10), nil); err != nil {
				return err
			}
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "deleted todo %d\n", id)
			return nil
		},
	}
}

// parseID converts a numeric string into int64 with a clear CLI error.
func parseID(s string) (int64, error) {
	id, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid id %q: must be an integer", s)
	}
	return id, nil
}

