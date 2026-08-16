package mcpserver

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strings"

	"github.com/mateo/homelab/cli/internal/dto"
)

// Domain enums mirrored from the backend / CLI. Validating here fails fast with
// a clear tool error instead of round-tripping a 400.
var (
	validStatus    = map[string]bool{"todo": true, "in_progress": true, "done": true, "cancelled": true}
	validKind      = map[string]bool{"note": true, "issue": true}
	validIssueType = map[string]bool{"feature": true, "bug": true, "improvement": true}
)

func decodeJSON[T any](raw []byte) (T, error) {
	var v T
	if err := json.Unmarshal(raw, &v); err != nil {
		return v, fmt.Errorf("parse response: %w", err)
	}
	return v, nil
}

func encodeJSON(v any) ([]byte, error) {
	b, err := json.Marshal(v)
	if err != nil {
		return nil, fmt.Errorf("encode request: %w", err)
	}
	return b, nil
}

func todoPath(id int64) string {
	return fmt.Sprintf("/api/todos/%d", id)
}

func projectPath(id int64) string {
	return fmt.Sprintf("/api/projects/%d", id)
}

// productsPath builds /api/products.
func productsPath() string {
	return "/api/products"
}

// secretProjectsPath builds /api/products/{productID}/projects.
func secretProjectsPath(productID int64) string {
	return fmt.Sprintf("/api/products/%d/projects", productID)
}

// secretEnvironmentsPath builds
// /api/products/{productID}/projects/{projectID}/environments.
func secretEnvironmentsPath(productID, projectID int64) string {
	return fmt.Sprintf("/api/products/%d/projects/%d/environments", productID, projectID)
}

// secretsPath builds .../environments/{envName}/secrets. envName is
// URL-escaped defensively even though it is validated against a fixed set.
func secretsPath(productID, projectID int64, envName string) string {
	return fmt.Sprintf("/api/products/%d/projects/%d/environments/%s/secrets", productID, projectID, url.PathEscape(envName))
}

// secretKeyPath builds .../secrets/{key}. key is URL-escaped since it is
// arbitrary caller-supplied input.
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

// validEnvironmentName mirrors the backend's fixed environment set. Kept
// lenient (not enforced everywhere) since the backend is the source of truth,
// but used to fail fast on obviously wrong input from an AI caller.
var validEnvironmentName = map[string]bool{"development": true, "staging": true, "production": true}

// validateSecretScope validates the common productID/projectID/environment
// triple shared by every secret-level MCP tool.
func validateSecretScope(productID, projectID int64, environment string) error {
	if productID <= 0 {
		return errPositiveID("product_id")
	}
	if projectID <= 0 {
		return errPositiveID("project_id")
	}
	if strings.TrimSpace(environment) == "" {
		return fmt.Errorf("environment is required")
	}
	if !validEnvironmentName[environment] {
		return fmt.Errorf("invalid environment %q: must be development, staging, or production", environment)
	}
	return nil
}

// buildTodoListPath appends optional query filters. project_id accepts a
// numeric id or the literal "none" (unassigned), matching the CLI.
func buildTodoListPath(kind, issueType, projectID string) (string, error) {
	if kind != "" && !validKind[kind] {
		return "", fmt.Errorf("invalid kind %q: must be note or issue", kind)
	}
	if issueType != "" && !validIssueType[issueType] {
		return "", fmt.Errorf("invalid issue_type %q: must be feature, bug, or improvement", issueType)
	}
	if issueType != "" && kind != "" && kind != "issue" {
		return "", fmt.Errorf("issue_type can only be set when kind is issue")
	}
	if projectID != "" && projectID != "none" {
		// numeric check only; backend rejects bad ids
		var n int64
		if _, err := fmt.Sscan(projectID, &n); err != nil {
			return "", fmt.Errorf("invalid project_id %q: must be an integer or none", projectID)
		}
	}

	path := "/api/todos"
	q := make([]string, 0, 3)
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
	return path, nil
}

// buildTodoUpdatePatch maps the MCP todo_update input (optional values +
// clear_* booleans) onto the backend's three-state dto.UpdateTodo. A clear_*
// flag and its value field must not both be set.
func buildTodoUpdatePatch(in todoUpdateInput) (dto.UpdateTodo, error) {
	if in.ClearDueDate && in.DueDate != nil {
		return dto.UpdateTodo{}, fmt.Errorf("cannot set due_date and clear_due_date together")
	}
	if in.ClearIssueType && in.IssueType != nil {
		return dto.UpdateTodo{}, fmt.Errorf("cannot set issue_type and clear_issue_type together")
	}
	if in.ClearProjectID && in.ProjectID != nil {
		return dto.UpdateTodo{}, fmt.Errorf("cannot set project_id and clear_project_id together")
	}

	patch := dto.UpdateTodo{}
	changed := false

	if in.Title != nil {
		patch.Title = in.Title
		changed = true
	}
	if in.Body != nil {
		patch.Body = in.Body
		changed = true
	}
	if in.Status != nil {
		if !validStatus[*in.Status] {
			return dto.UpdateTodo{}, fmt.Errorf("invalid status %q: must be todo, in_progress, done, or cancelled", *in.Status)
		}
		patch.Status = in.Status
		changed = true
	}
	if in.Priority != nil {
		if *in.Priority < 0 || *in.Priority > 3 {
			return dto.UpdateTodo{}, fmt.Errorf("invalid priority %d: must be 0..3", *in.Priority)
		}
		patch.Priority = in.Priority
		changed = true
	}
	if in.Kind != nil {
		if !validKind[*in.Kind] {
			return dto.UpdateTodo{}, fmt.Errorf("invalid kind %q: must be note or issue", *in.Kind)
		}
		patch.Kind = in.Kind
		changed = true
	}
	if in.ClearDueDate {
		var nilPtr *string
		patch.DueDate = &nilPtr
		changed = true
	} else if in.DueDate != nil {
		v := *in.DueDate
		vPtr := &v
		patch.DueDate = &vPtr
		changed = true
	}
	if in.ClearIssueType {
		var nilPtr *string
		patch.IssueType = &nilPtr
		changed = true
	} else if in.IssueType != nil {
		if !validIssueType[*in.IssueType] {
			return dto.UpdateTodo{}, fmt.Errorf("invalid issue_type %q: must be feature, bug, or improvement", *in.IssueType)
		}
		v := *in.IssueType
		vPtr := &v
		patch.IssueType = &vPtr
		changed = true
	}
	if in.ClearProjectID {
		var nilPtr *int64
		patch.ProjectID = &nilPtr
		changed = true
	} else if in.ProjectID != nil {
		pid := *in.ProjectID
		pidPtr := &pid
		patch.ProjectID = &pidPtr
		changed = true
	}

	if !changed {
		return dto.UpdateTodo{}, fmt.Errorf("no fields specified: pass at least one field or clear_* flag")
	}
	return patch, nil
}

func validateCreateTodo(in todoCreateInput) error {
	if strings.TrimSpace(in.Title) == "" {
		return fmt.Errorf("title is required")
	}
	if in.Priority < 0 || in.Priority > 3 {
		return fmt.Errorf("invalid priority %d: must be 0..3", in.Priority)
	}
	if in.Kind != "" && !validKind[in.Kind] {
		return fmt.Errorf("invalid kind %q: must be note or issue", in.Kind)
	}
	if in.IssueType != nil && *in.IssueType != "" && !validIssueType[*in.IssueType] {
		return fmt.Errorf("invalid issue_type %q: must be feature, bug, or improvement", *in.IssueType)
	}
	if in.IssueType != nil && *in.IssueType != "" && in.Kind != "" && in.Kind != "issue" {
		return fmt.Errorf("issue_type can only be set when kind is issue")
	}
	return nil
}
