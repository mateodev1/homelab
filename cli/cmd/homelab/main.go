// Command homelab is the CLI client for the homelab backend HTTP API.
package main

import (
	"fmt"
	"os"

	"github.com/mateo/homelab/cli/internal/cmd"
)

// version is the build-supplied CLI version string.
var version = "0.1.0"

func main() {
	root := cmd.NewRootCommand(version)
	if err := root.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}