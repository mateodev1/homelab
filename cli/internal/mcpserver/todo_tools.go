package mcpserver

import (
	"context"

	"github.com/mateo/homelab/cli/internal/dto"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type todoListInput struct {
	Kind      string `json:"kind,omitempty" jsonschema:"optional filter: note or issue"`
	IssueType string `json:"issue_type,omitempty" jsonschema:"optional filter: feature, bug, or improvement"`
	ProjectID string `json:"project_id,omitempty" jsonschema:"optional filter: numeric project id or none for unassigned"`
}

type todoIDInput struct {
	ID int64 `json:"id" jsonschema:"todo id"`
}

type todoCreateInput struct {
	Title     string  `json:"title" jsonschema:"required todo title"`
	Body      string  `json:"body,omitempty" jsonschema:"optional body text"`
	Priority  int     `json:"priority,omitempty" jsonschema:"priority 0-3, default 0"`
	DueDate   *string `json:"due_date,omitempty" jsonschema:"optional due date string"`
	Kind      string  `json:"kind,omitempty" jsonschema:"note or issue; defaults to note on the server"`
	IssueType *string `json:"issue_type,omitempty" jsonschema:"feature, bug, or improvement; only when kind is issue"`
	ProjectID *int64  `json:"project_id,omitempty" jsonschema:"optional project id to associate"`
}

type todoUpdateInput struct {
	ID             int64   `json:"id" jsonschema:"todo id"`
	Title          *string `json:"title,omitempty" jsonschema:"set title"`
	Body           *string `json:"body,omitempty" jsonschema:"set body"`
	Status         *string `json:"status,omitempty" jsonschema:"todo, in_progress, done, or cancelled"`
	Priority       *int    `json:"priority,omitempty" jsonschema:"priority 0-3"`
	Kind           *string `json:"kind,omitempty" jsonschema:"note or issue"`
	DueDate        *string `json:"due_date,omitempty" jsonschema:"set due date; do not combine with clear_due_date"`
	IssueType      *string `json:"issue_type,omitempty" jsonschema:"feature, bug, or improvement; do not combine with clear_issue_type"`
	ProjectID      *int64  `json:"project_id,omitempty" jsonschema:"set project id; do not combine with clear_project_id"`
	ClearDueDate   bool    `json:"clear_due_date,omitempty" jsonschema:"if true, set due_date to null"`
	ClearIssueType bool    `json:"clear_issue_type,omitempty" jsonschema:"if true, set issue_type to null"`
	ClearProjectID bool    `json:"clear_project_id,omitempty" jsonschema:"if true, set project_id to null"`
}

type todoListOutput struct {
	Todos []dto.Todo `json:"todos" jsonschema:"list of todos"`
}

type deleteOutput struct {
	Deleted bool  `json:"deleted" jsonschema:"true when the resource was deleted"`
	ID      int64 `json:"id" jsonschema:"id of the deleted resource"`
}

func (s *Server) todoList(ctx context.Context, _ *mcp.CallToolRequest, in todoListInput) (*mcp.CallToolResult, todoListOutput, error) {
	path, err := buildTodoListPath(in.Kind, in.IssueType, in.ProjectID)
	if err != nil {
		return nil, todoListOutput{}, err
	}
	raw, _, err := s.c.Do(ctx, "GET", path, nil)
	if err != nil {
		return nil, todoListOutput{}, err
	}
	todos, err := decodeJSON[[]dto.Todo](raw)
	if err != nil {
		return nil, todoListOutput{}, err
	}
	if todos == nil {
		todos = []dto.Todo{}
	}
	return nil, todoListOutput{Todos: todos}, nil
}

func (s *Server) todoGet(ctx context.Context, _ *mcp.CallToolRequest, in todoIDInput) (*mcp.CallToolResult, dto.Todo, error) {
	if in.ID <= 0 {
		return nil, dto.Todo{}, errPositiveID("id")
	}
	raw, _, err := s.c.Do(ctx, "GET", todoPath(in.ID), nil)
	if err != nil {
		return nil, dto.Todo{}, err
	}
	t, err := decodeJSON[dto.Todo](raw)
	if err != nil {
		return nil, dto.Todo{}, err
	}
	return nil, t, nil
}

func (s *Server) todoCreate(ctx context.Context, _ *mcp.CallToolRequest, in todoCreateInput) (*mcp.CallToolResult, dto.Todo, error) {
	if err := validateCreateTodo(in); err != nil {
		return nil, dto.Todo{}, err
	}
	payload := dto.CreateTodo{
		Title:     in.Title,
		Body:      in.Body,
		Priority:  in.Priority,
		DueDate:   in.DueDate,
		Kind:      in.Kind,
		IssueType: in.IssueType,
		ProjectID: in.ProjectID,
	}
	body, err := encodeJSON(payload)
	if err != nil {
		return nil, dto.Todo{}, err
	}
	raw, _, err := s.c.Do(ctx, "POST", "/api/todos", body)
	if err != nil {
		return nil, dto.Todo{}, err
	}
	t, err := decodeJSON[dto.Todo](raw)
	if err != nil {
		return nil, dto.Todo{}, err
	}
	return nil, t, nil
}

func (s *Server) todoUpdate(ctx context.Context, _ *mcp.CallToolRequest, in todoUpdateInput) (*mcp.CallToolResult, dto.Todo, error) {
	if in.ID <= 0 {
		return nil, dto.Todo{}, errPositiveID("id")
	}
	patch, err := buildTodoUpdatePatch(in)
	if err != nil {
		return nil, dto.Todo{}, err
	}
	body, err := encodeJSON(patch)
	if err != nil {
		return nil, dto.Todo{}, err
	}
	raw, _, err := s.c.Do(ctx, "PUT", todoPath(in.ID), body)
	if err != nil {
		return nil, dto.Todo{}, err
	}
	t, err := decodeJSON[dto.Todo](raw)
	if err != nil {
		return nil, dto.Todo{}, err
	}
	return nil, t, nil
}

func (s *Server) todoDone(ctx context.Context, _ *mcp.CallToolRequest, in todoIDInput) (*mcp.CallToolResult, dto.Todo, error) {
	if in.ID <= 0 {
		return nil, dto.Todo{}, errPositiveID("id")
	}
	status := "done"
	body, err := encodeJSON(dto.UpdateTodo{Status: &status})
	if err != nil {
		return nil, dto.Todo{}, err
	}
	raw, _, err := s.c.Do(ctx, "PUT", todoPath(in.ID), body)
	if err != nil {
		return nil, dto.Todo{}, err
	}
	t, err := decodeJSON[dto.Todo](raw)
	if err != nil {
		return nil, dto.Todo{}, err
	}
	return nil, t, nil
}

func (s *Server) todoDelete(ctx context.Context, _ *mcp.CallToolRequest, in todoIDInput) (*mcp.CallToolResult, deleteOutput, error) {
	if in.ID <= 0 {
		return nil, deleteOutput{}, errPositiveID("id")
	}
	if _, _, err := s.c.Do(ctx, "DELETE", todoPath(in.ID), nil); err != nil {
		return nil, deleteOutput{}, err
	}
	return nil, deleteOutput{Deleted: true, ID: in.ID}, nil
}
