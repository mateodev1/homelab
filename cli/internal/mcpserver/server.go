// Package mcpserver implements the homelab Model Context Protocol server.
//
// Tools mirror the CLI surface (health, todos, projects) and talk to the
// backend through cli/internal/client using the same config resolution as the
// CLI. The AI never receives the API key.
package mcpserver

import (
	"context"
	"log"

	"github.com/mateo/homelab/cli/internal/client"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// Server holds the shared HTTP client used by every tool handler.
type Server struct {
	c *client.Client
}

// Run resolves config, builds the MCP server with all tools, and serves over
// stdio until the client disconnects.
func Run(ctx context.Context, version string) error {
	c, err := resolveClient()
	if err != nil {
		return err
	}
	s := &Server{c: c}
	srv := s.NewMCPServer(version)
	log.Printf("homelab-mcp %s ready (stdio)", version)
	return srv.Run(ctx, &mcp.StdioTransport{})
}

// NewMCPServer builds an mcp.Server with every homelab tool registered. Exported
// for tests that drive tools without the stdio transport.
func (s *Server) NewMCPServer(version string) *mcp.Server {
	srv := mcp.NewServer(&mcp.Implementation{
		Name:    "homelab",
		Version: version,
	}, nil)
	s.registerTools(srv)
	return srv
}

// NewForTest constructs a Server bound to an already-built client (typically
// pointed at an httptest backend). Used by unit tests.
func NewForTest(c *client.Client) *Server {
	return &Server{c: c}
}

func (s *Server) registerTools(srv *mcp.Server) {
	mcp.AddTool(srv, &mcp.Tool{
		Name:        "health",
		Description: "Check backend health (no auth required)",
	}, s.health)

	mcp.AddTool(srv, &mcp.Tool{
		Name:        "todo_list",
		Description: "List todos with optional filters (kind, issue_type, project_id)",
	}, s.todoList)
	mcp.AddTool(srv, &mcp.Tool{
		Name:        "todo_get",
		Description: "Get a single todo by id",
	}, s.todoGet)
	mcp.AddTool(srv, &mcp.Tool{
		Name:        "todo_create",
		Description: "Create a todo (title required; priority 0-3; kind note|issue)",
	}, s.todoCreate)
	mcp.AddTool(srv, &mcp.Tool{
		Name:        "todo_update",
		Description: "Partial-update a todo. Use clear_due_date / clear_issue_type / clear_project_id to null nullable fields. Do not set a value and its clear_* flag together.",
	}, s.todoUpdate)
	mcp.AddTool(srv, &mcp.Tool{
		Name:        "todo_done",
		Description: "Mark a todo status as done",
	}, s.todoDone)
	mcp.AddTool(srv, &mcp.Tool{
		Name:        "todo_delete",
		Description: "Delete a todo by id",
	}, s.todoDelete)

	mcp.AddTool(srv, &mcp.Tool{
		Name:        "project_list",
		Description: "List all projects",
	}, s.projectList)
	mcp.AddTool(srv, &mcp.Tool{
		Name:        "project_get",
		Description: "Get a single project by id",
	}, s.projectGet)
	mcp.AddTool(srv, &mcp.Tool{
		Name:        "project_create",
		Description: "Create a project (name required)",
	}, s.projectCreate)
	mcp.AddTool(srv, &mcp.Tool{
		Name:        "project_update",
		Description: "Partial-update a project (name and/or color)",
	}, s.projectUpdate)
	mcp.AddTool(srv, &mcp.Tool{
		Name:        "project_delete",
		Description: "Delete a project by id (orphans referencing todos by nulling their project_id)",
	}, s.projectDelete)
}
