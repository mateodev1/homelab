package dto

// Todo mirrors the backend wire shape for a todo.
//
// Nullable fields are modelled with pointers so that "absent" (nil) and
// "explicit clear" (non-nil pointer to nil) are distinguishable on the patch
// path. On read responses the same struct decodes incoming JSON directly.
type Todo struct {
	ID        int64   `json:"id"`
	Title     string  `json:"title"`
	Body      string  `json:"body"`
	Status    string  `json:"status"`
	Priority  int     `json:"priority"`
	DueDate   *string `json:"due_date"`
	Kind      string  `json:"kind"`
	IssueType *string `json:"issue_type"`
	ProjectID *int64  `json:"project_id"`
	CreatedAt string  `json:"created_at"`
	UpdatedAt string  `json:"updated_at"`
}

// CreateTodo is the request body for POST /api/todos. Pointer fields keep nil
// distinct from zero values so the server only receives the keys the caller
// intends to set.
type CreateTodo struct {
	Title     string  `json:"title"`
	Body      string  `json:"body,omitempty"`
	Priority  int     `json:"priority,omitempty"`
	DueDate   *string `json:"due_date,omitempty"`
	Kind      string  `json:"kind,omitempty"`
	IssueType *string `json:"issue_type,omitempty"`
	ProjectID *int64  `json:"project_id,omitempty"`
}

// UpdateTodo is the PATCH-style body for PUT /api/todos/{id}. It mirrors the
// backend's three-state semantics: a nil field means "key omitted" (unchanged),
// a non-nil field pointing at nil means "send literal JSON null" (clear), and a
// non-nil field pointing at a value means "set". Using double pointers lets
// encoding/json emit the key only when the outer pointer is set while still
// producing a bare `null` when the inner pointer is nil.
//
// DueDate/IssueType/ProjectID are double pointers because they are nullable.
// Title/Body/Status/Kind single-pointer: set or omitted. Priority is a single
// pointer because integer zero is a valid set value.
type UpdateTodo struct {
	Title     *string  `json:"title,omitempty"`
	Body      *string  `json:"body,omitempty"`
	Status    *string  `json:"status,omitempty"`
	Priority  *int     `json:"priority,omitempty"`
	DueDate   **string `json:"due_date,omitempty"`
	Kind      *string  `json:"kind,omitempty"`
	IssueType **string `json:"issue_type,omitempty"`
	ProjectID **int64  `json:"project_id,omitempty"`
}

