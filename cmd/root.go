package cmd

import (
	"fmt"
	"os"

	"github.com/CedricHerzog/perfowl/internal/version"
	"github.com/spf13/cobra"
)

var (
	profilePath  string
	outputFormat string
	browserType  string
)

var rootCmd = &cobra.Command{
	Use:     "perfowl",
	Version: version.Version,
	Short:   "PerfOwl - Optimization Workbench & Lab for browser performance traces",
	Long: `PerfOwl (Optimization Workbench & Lab) - A wise tool for analyzing browser
profiler exports to identify performance bottlenecks and provide AI-friendly insights.

Features:
- Sharp-eyed bottleneck detection (GC pressure, layout thrashing, sync IPC, long tasks)
- Web extension performance analysis
- Marker extraction and filtering
- MCP server integration for Claude and other AI assistants
- Multiple output formats (text, JSON, markdown)

Supported browsers:
- Firefox Profiler exports
- Chrome DevTools Performance traces`,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().StringVarP(&profilePath, "profile", "p", "", "Path to browser profile JSON (gzip supported)")
	rootCmd.PersistentFlags().StringVarP(&outputFormat, "output", "o", "text", "Output format: text, json, markdown")
	rootCmd.PersistentFlags().StringVarP(&browserType, "browser", "b", "auto", "Browser type: auto, firefox, chrome")
}
