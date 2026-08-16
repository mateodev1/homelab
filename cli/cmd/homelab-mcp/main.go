// Command homelab-mcp is the Model Context Protocol server for the homelab backend.
//
// It reuses the CLI config file and HTTP client so the AI never sees the API key.
// Auth and base URL come from the same resolution chain as the homelab CLI
// (env > config file > defaults). Run `homelab login` once to persist credentials.
//
// Logs go to stderr only; stdout is reserved for the MCP stdio transport.
package main

import (
	"context"
	"log"
	"os"

	"github.com/mateo/homelab/cli/internal/mcpserver"
)

// version is the build-supplied MCP server version string.
var version = "0.1.0"

func main() {
	log.SetOutput(os.Stderr)
	log.SetFlags(0)

	if err := mcpserver.Run(context.Background(), version); err != nil {
		log.Fatalf("homelab-mcp: %v", err)
	}
}
