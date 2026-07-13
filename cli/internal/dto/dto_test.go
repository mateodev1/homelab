package dto

import (
	"encoding/json"
	"strings"
	"testing"
)

func strPtr(v string) *string         { return &v }
func intPtr(v int) *int               { return &v }
func int64Ptr(v int64) *int64         { return &v }
func ptrTo[T any](v T) *T             { return &v }

func TestUpdateTodo_OmitsUnsetKeys(t *testing.T) {
	t.Parallel()
	patch := UpdateTodo{} // nothing set
	b, err := json.Marshal(patch)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	out := string(b)
	if out != "{}" {
		t.Fatalf("empty patch marshalled to %q, want {}", out)
	}
}

func TestUpdateTodo_SetValueIncludesKey(t *testing.T) {
	t.Parallel()
	title := "new title"
	status := "done"
	patch := UpdateTodo{Title: &title, Status: &status}
	b, err := json.Marshal(patch)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	out := string(b)
	if !strings.Contains(out, `"title":"new title"`) {
		t.Errorf("missing title key in %q", out)
	}
	if !strings.Contains(out, `"status":"done"`) {
		t.Errorf("missing status key in %q", out)
	}
}

func TestUpdateTodo_NullClearEmitsLiteralJSONNull(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name  string
		build func() UpdateTodo
		want  string
	}{
		{
			name: "due_date clear",
			build: func() UpdateTodo {
				var nilPtr *string
				return UpdateTodo{DueDate: &nilPtr}
			},
			want: `"due_date":null`,
		},
		{
			name: "issue_type clear",
			build: func() UpdateTodo {
				var nilPtr *string
				return UpdateTodo{IssueType: &nilPtr}
			},
			want: `"issue_type":null`,
		},
		{
			name: "project_id clear",
			build: func() UpdateTodo {
				var nilPtr *int64
				return UpdateTodo{ProjectID: &nilPtr}
			},
			want: `"project_id":null`,
		},
	}
	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Helper()
			b, err := json.Marshal(tc.build())
			if err != nil {
				t.Fatalf("marshal: %v", err)
			}
			if !strings.Contains(string(b), tc.want) {
				t.Fatalf("got %q, want substring %q", b, tc.want)
			}
		})
	}
}

func TestUpdateTodo_SetNullableValue(t *testing.T) {
	t.Parallel()
	due := "2024-01-02T00:00:00Z"
	duePtr := &due
	patch := UpdateTodo{DueDate: &duePtr}
	b, err := json.Marshal(patch)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	if !strings.Contains(string(b), `"due_date":"2024-01-02T00:00:00Z"`) {
		t.Errorf("got %q", b)
	}
}

func TestUpdateTodo_PriorityZeroIsIncluded(t *testing.T) {
	t.Parallel()
	p := 0
	patch := UpdateTodo{Priority: &p}
	b, err := json.Marshal(patch)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	if !strings.Contains(string(b), `"priority":0`) {
		t.Errorf("priority 0 must be included when set, got %q", b)
	}
}

func TestTodo_UnmarshalRoundTrip(t *testing.T) {
	t.Parallel()
	raw := `{"id":42,"title":"buy milk","body":"2L","status":"todo","priority":2,"due_date":null,"kind":"note","issue_type":null,"project_id":null,"created_at":"2024-01-01T00:00:00Z","updated_at":"2024-01-01T00:00:00Z"}`
	var t1 Todo
	if err := json.Unmarshal([]byte(raw), &t1); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if t1.ID != 42 || t1.Title != "buy milk" || t1.Priority != 2 || t1.Kind != "note" {
		t.Fatalf("decoded wrong: %+v", t1)
	}
	if t1.DueDate != nil || t1.IssueType != nil || t1.ProjectID != nil {
		t.Fatalf("nullable fields should be nil: %+v", t1)
	}
}

func TestProject_UnmarshalRoundTrip(t *testing.T) {
	t.Parallel()
	raw := `{"id":1,"name":"homelab","color":"#ff0","created_at":"2024-01-01T00:00:00Z"}`
	var p Project
	if err := json.Unmarshal([]byte(raw), &p); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if p.ID != 1 || p.Name != "homelab" || p.Color != "#ff0" {
		t.Fatalf("decoded wrong: %+v", p)
	}
}

func TestCreateTodo_OmitsEmptyOptionals(t *testing.T) {
	t.Parallel()
	c := CreateTodo{Title: "only title"}
	b, err := json.Marshal(c)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	out := string(b)
	if !strings.Contains(out, `"title":"only title"`) {
		t.Errorf("missing title in %q", out)
	}
	for _, k := range []string{`"body"`, `"priority"`, `"due_date"`, `"kind"`, `"issue_type"`, `"project_id"`} {
		if strings.Contains(out, k) {
			t.Errorf("optional key %s should be omitted in %q", k, out)
		}
	}
}

func TestUpdateProject_OmitsUnsetKeys(t *testing.T) {
	t.Parallel()
	b, err := json.Marshal(UpdateProject{})
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	if string(b) != "{}" {
		t.Fatalf("got %q, want {}", b)
	}
}

// keep the helpers referenced so unused linter stays happy across files.
var _ = strPtr
var _ = intPtr
var _ = int64Ptr
var _ = ptrTo[int]