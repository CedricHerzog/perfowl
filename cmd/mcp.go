package cmd

import (
	"fmt"

	mcpserver "github.com/CedricHerzog/perfowl/internal/mcp"
	"github.com/spf13/cobra"
)

var mcpCmd = &cobra.Command{
	Use:   "mcp",
	Short: "Start the MCP server for AI tool integration",
	Long: `Starts the Model Context Protocol (MCP) server that exposes
profile analysis capabilities as tools for AI assistants like Claude.

The server communicates over stdio and provides the following tools:
- get_summary: Get profile summary information
- get_bottlenecks: Detect performance bottlenecks
- get_markers: Extract and filter markers
- analyze_extension: Analyze extension performance
- analyze_profile: Comprehensive profile analysis

To use with Docker MCP Toolkit:
1. Build Docker image: docker build -t profile-analyzer-mcp .
2. Register in Docker Desktop MCP Toolkit
3. Connect Claude Desktop or Claude Code`,
	RunE: runMCP,
}

func init() {
	rootCmd.AddCommand(mcpCmd)
}

func runMCP(cmd *cobra.Command, args []string) error {
	server := mcpserver.NewServer()

	// Serve blocks until stdin is closed
	if err := server.Serve(); err != nil {
		return fmt.Errorf("MCP server error: %w", err)
	}

	return nil
}
