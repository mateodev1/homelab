package mcpserver

import (
	"context"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// healthOutput mirrors GET /api/health.
type healthOutput struct {
	Status string `json:"status" jsonschema:"health status string, typically ok"`
	DBOk   bool   `json:"db_ok" jsonschema:"whether the backend database ping succeeded"`
}

type emptyInput struct{}

func (s *Server) health(ctx context.Context, _ *mcp.CallToolRequest, _ emptyInput) (*mcp.CallToolResult, healthOutput, error) {
	raw, _, err := s.c.Do(ctx, "GET", "/api/health", nil)
	if err != nil {
		return nil, healthOutput{}, err
	}
	out, err := decodeJSON[healthOutput](raw)
	if err != nil {
		return nil, healthOutput{}, err
	}
	return nil, out, nil
}
