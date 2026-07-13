package dto

// Project mirrors the backend wire shape for a project.
type Project struct {
	ID        int64  `json:"id"`
	Name      string `json:"name"`
	Color     string `json:"color"`
	CreatedAt string `json:"created_at"`
}

// CreateProject is the request body for POST /api/projects.
type CreateProject struct {
	Name  string `json:"name"`
	Color string `json:"color,omitempty"`
}

// UpdateProject is the PATCH-style body for PUT /api/projects/{id}. Omitted
// keys (nil) leave the field unchanged; present keys set the new value.
type UpdateProject struct {
	Name  *string `json:"name,omitempty"`
	Color *string `json:"color,omitempty"`
}