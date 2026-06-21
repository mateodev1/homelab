package main

import (
	"fmt"
	"os"
)

const version = "0.0.1-dev"

func main() {
	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "--version", "-v":
			fmt.Printf("homelab %s\n", version)
			return
		case "--help", "-h", "help":
			printUsage()
			return
		default:
			fmt.Fprintf(os.Stderr, "unknown command: %s\n\n", os.Args[1])
			printUsage()
			os.Exit(1)
		}
	}
	printUsage()
}

func printUsage() {
	fmt.Printf(`homelab CLI — coming soon

Usage:
  homelab [command]

Available Commands:
  (none yet — check back soon)

Flags:
  -h, --help      Show this help message
  -v, --version   Print version information

Version: %s
`, version)
}
