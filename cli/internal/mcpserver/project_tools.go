package mcpserver

import (
	"context"
	"fmt"
	"strings"

	"github.com/mateo/homelab/cli/internal/dto"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type projectListInput struct{}

type projectIDInput struct {
	ID int64 `json:"id" jsonschema:"project id"`
}

type projectCreateInput struct {
	Name  string `json:"name" jsonschema:"required project name"`
	Color string `json:"color,omitempty" jsonschema:"optional color"`
}

type projectUpdateInput struct {
	ID    int64   `json:"id" jsonschema:"project id"`
	Name  *string `json:"name,omitempty" jsonschema:"set name"`
	Color *string `json:"color,omitempty" jsonschema:"set color"`
}

type projectListOutput struct {
	Projects []dto.Project `json:"projects" jsonschema:"list of projects"`
}

func errPositiveID(field string) error {
	return fmt.Errorf("%s must be a positive integer", field)
}

func (s *Server) projectList(ctx context.Context, _ *mcp.CallToolRequest, _ projectListInput) (*mcp.CallToolResult, projectListOutput, error) {
	raw, _, err := s.c.Do(ctx, "GET", "/api/projects", nil)
	if err != nil {
		return nil, projectListOutput{}, err
	}
	projects, err := decodeJSON[[]dto.Project](raw)
	if err != nil {
		return nil, projectListOutput{}, err
	}
	if projects == nil {
		projects = []dto.Project{}
	}
	return nil, projectListOutput{Projects: projects}, nil
}

func (s *Server) projectGet(ctx context.Context, _ *mcp.CallToolRequest, in projectIDInput) (*mcp.CallToolResult, dto.Project, error) {
	if in.ID <= 0 {
		return nil, dto.Project{}, errPositiveID("id")
	}
	raw, _, err := s.c.Do(ctx, "GET", projectPath(in.ID), nil)
	if err != nil {
		return nil, dto.Project{}, err
	}
	p, err := decodeJSON[dto.Project](raw)
	if err != nil {
		return nil, dto.Project{}, err
	}
	return nil, p, nil
}

func (s *Server) projectCreate(ctx context.Context, _ *mcp.CallToolRequest, in projectCreateInput) (*mcp.CallToolResult, dto.Project, error) {
	if strings.TrimSpace(in.Name) == "" {
		return nil, dto.Project{}, fmt.Errorf("name is required")
	}
	body, err := encodeJSON(dto.CreateProject{Name: in.Name, Color: in.Color})
	if err != nil {
		return nil, dto.Project{}, err
	}
	raw, _, err := s.c.Do(ctx, "POST", "/api/projects", body)
	if err != nil {
		return nil, dto.Project{}, err
	}
	p, err := decodeJSON[dto.Project](raw)
	if err != nil {
		return nil, dto.Project{}, err
	}
	return nil, p, nil
}

func (s *Server) projectUpdate(ctx context.Context, _ *mcp.CallToolRequest, in projectUpdateInput) (*mcp.CallToolResult, dto.Project, error) {
	if in.ID <= 0 {
		return nil, dto.Project{}, errPositiveID("id")
	}
	if in.Name == nil && in.Color == nil {
		return nil, dto.Project{}, fmt.Errorf("no fields specified: pass name and/or color")
	}
	if in.Name != nil && strings.TrimSpace(*in.Name) == "" {
		return nil, dto.Project{}, fmt.Errorf("name is required")
	}
	body, err := encodeJSON(dto.UpdateProject{Name: in.Name, Color: in.Color})
	if err != nil {
		return nil, dto.Project{}, err
	}
	raw, _, err := s.c.Do(ctx, "PUT", projectPath(in.ID), body)
	if err != nil {
		return nil, dto.Project{}, err
	}
	p, err := decodeJSON[dto.Project](raw)
	if err != nil {
		return nil, dto.Project{}, err
	}
	return nil, p, nil
}

func (s *Server) projectDelete(ctx context.Context, _ *mcp.CallToolRequest, in projectIDInput) (*mcp.CallToolResult, deleteOutput, error) {
	if in.ID <= 0 {
		return nil, deleteOutput{}, errPositiveID("id")
	}
	if _, _, err := s.c.Do(ctx, "DELETE", projectPath(in.ID), nil); err != nil {
		return nil, deleteOutput{}, err
	}
	return nil, deleteOutput{Deleted: true, ID: in.ID}, nil
}
