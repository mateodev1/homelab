package mcpserver

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestBuildTodoListPath(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name      string
		kind      string
		issueType string
		projectID string
		wantPath  string
		wantErr   string
	}{
		{"no filters", "", "", "", "/api/todos", ""},
		{"kind only", "issue", "", "", "/api/todos?kind=issue", ""},
		{"all filters", "issue", "bug", "3", "/api/todos?kind=issue&issue_type=bug&project_id=3", ""},
		{"project none", "", "", "none", "/api/todos?project_id=none", ""},
		{"bad kind", "other", "", "", "", "invalid kind"},
		{"issue_type without issue kind", "note", "bug", "", "", "issue_type can only be set"},
		{"bad project id", "", "", "abc", "", "invalid project_id"},
	}
	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got, err := buildTodoListPath(tc.kind, tc.issueType, tc.projectID)
			if tc.wantErr != "" {
				if err == nil {
					t.Fatalf("expected error containing %q, got nil", tc.wantErr)
				}
				if !strings.Contains(err.Error(), tc.wantErr) {
					t.Fatalf("error %q missing %q", err, tc.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tc.wantPath {
				t.Fatalf("path = %q, want %q", got, tc.wantPath)
			}
		})
	}
}

func TestBuildTodoUpdatePatch_ClearFlags(t *testing.T) {
	t.Parallel()

	due := "2026-08-15"
	issue := "bug"
	pid := int64(7)

	t.Run("clear due date emits null", func(t *testing.T) {
		t.Parallel()
		patch, err := buildTodoUpdatePatch(todoUpdateInput{ID: 1, ClearDueDate: true})
		if err != nil {
			t.Fatal(err)
		}
		b, err := json.Marshal(patch)
		if err != nil {
			t.Fatal(err)
		}
		if !strings.Contains(string(b), `"due_date":null`) {
			t.Fatalf("body %s missing due_date null", b)
		}
	})

	t.Run("set due date", func(t *testing.T) {
		t.Parallel()
		patch, err := buildTodoUpdatePatch(todoUpdateInput{ID: 1, DueDate: &due})
		if err != nil {
			t.Fatal(err)
		}
		b, err := json.Marshal(patch)
		if err != nil {
			t.Fatal(err)
		}
		if !strings.Contains(string(b), `"due_date":"2026-08-15"`) {
			t.Fatalf("body %s missing due_date value", b)
		}
	})

	t.Run("clear and value conflict", func(t *testing.T) {
		t.Parallel()
		_, err := buildTodoUpdatePatch(todoUpdateInput{ID: 1, DueDate: &due, ClearDueDate: true})
		if err == nil || !strings.Contains(err.Error(), "clear_due_date") {
			t.Fatalf("expected clear conflict, got %v", err)
		}
		_, err = buildTodoUpdatePatch(todoUpdateInput{ID: 1, IssueType: &issue, ClearIssueType: true})
		if err == nil || !strings.Contains(err.Error(), "clear_issue_type") {
			t.Fatalf("expected clear conflict, got %v", err)
		}
		_, err = buildTodoUpdatePatch(todoUpdateInput{ID: 1, ProjectID: &pid, ClearProjectID: true})
		if err == nil || !strings.Contains(err.Error(), "clear_project_id") {
			t.Fatalf("expected clear conflict, got %v", err)
		}
	})

	t.Run("no fields", func(t *testing.T) {
		t.Parallel()
		_, err := buildTodoUpdatePatch(todoUpdateInput{ID: 1})
		if err == nil || !strings.Contains(err.Error(), "no fields") {
			t.Fatalf("expected no fields error, got %v", err)
		}
	})

	t.Run("invalid status", func(t *testing.T) {
		t.Parallel()
		bad := "nope"
		_, err := buildTodoUpdatePatch(todoUpdateInput{ID: 1, Status: &bad})
		if err == nil || !strings.Contains(err.Error(), "invalid status") {
			t.Fatalf("expected invalid status, got %v", err)
		}
	})
}

func TestValidateCreateTodo(t *testing.T) {
	t.Parallel()
	if err := validateCreateTodo(todoCreateInput{}); err == nil || !strings.Contains(err.Error(), "title") {
		t.Fatalf("expected title required, got %v", err)
	}
	if err := validateCreateTodo(todoCreateInput{Title: "x", Priority: 9}); err == nil || !strings.Contains(err.Error(), "priority") {
		t.Fatalf("expected priority error, got %v", err)
	}
	it := "bug"
	if err := validateCreateTodo(todoCreateInput{Title: "x", Kind: "note", IssueType: &it}); err == nil || !strings.Contains(err.Error(), "issue_type") {
		t.Fatalf("expected issue_type/kind error, got %v", err)
	}
	if err := validateCreateTodo(todoCreateInput{Title: "ok", Kind: "issue", IssueType: &it, Priority: 2}); err != nil {
		t.Fatalf("unexpected: %v", err)
	}
}
