package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/CedricHerzog/perfowl/internal/analyzer"
	"github.com/CedricHerzog/perfowl/internal/parser"
	"github.com/spf13/cobra"
)

var contentionCmd = &cobra.Command{
	Use:   "contention",
	Short: "Detect thread contention issues",
	Long: `Detects contention issues between threads including:
- GC pauses affecting active workers
- Synchronous IPC blocking multiple threads
- Lock contention patterns

Provides severity assessment and recommendations for:
- Reducing GC pressure
- Converting sync IPC to async
- Optimizing shared resource access`,
	RunE: runContention,
}

func init() {
	rootCmd.AddCommand(contentionCmd)
}

func runContention(cmd *cobra.Command, args []string) error {
	if profilePath == "" {
		return fmt.Errorf("profile path is required (use --profile or -p)")
	}

	profile, err := parser.LoadProfile(profilePath)
	if err != nil {
		return fmt.Errorf("failed to load profile: %w", err)
	}

	analysis := analyzer.AnalyzeContention(profile)

	switch outputFormat {
	case "json":
		return outputContentionJSON(analysis)
	case "markdown":
		return outputContentionMarkdown(analysis)
	default:
		return outputContentionText(analysis)
	}
}

func outputContentionJSON(analysis analyzer.ContentionAnalysis) error {
	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	return encoder.Encode(analysis)
}

func outputContentionMarkdown(analysis analyzer.ContentionAnalysis) error {
	md := strings.Builder{}

	md.WriteString("# Contention Analysis\n\n")

	// Severity badge
	severityEmoji := map[string]string{
		"minimal": "✅",
		"low":     "🔵",
		"medium":  "🟡",
		"high":    "🔴",
		"unknown": "❓",
	}

	md.WriteString("## Summary\n\n")
	md.WriteString(fmt.Sprintf("- **Severity**: %s %s\n", severityEmoji[analysis.Severity], analysis.Severity))
	md.WriteString(fmt.Sprintf("- **Total Events**: %d\n", analysis.TotalEvents))
	md.WriteString(fmt.Sprintf("- **Total Impact**: %.2f ms\n", analysis.TotalImpactMs))
	md.WriteString(fmt.Sprintf("- **GC Contention**: %d events\n", analysis.GCContention))
	md.WriteString(fmt.Sprintf("- **IPC Contention**: %d events\n", analysis.IPCContention))
	md.WriteString(fmt.Sprintf("- **Lock Contention**: %d events\n", analysis.LockContention))

	if len(analysis.Events) > 0 {
		md.WriteString("\n## Top Contention Events\n\n")
		md.WriteString("| Time | Type | Duration | Threads | Description |\n")
		md.WriteString("|------|------|----------|---------|-------------|\n")

		displayCount := len(analysis.Events)
		if displayCount > 15 {
			displayCount = 15
		}

		for i := 0; i < displayCount; i++ {
			e := analysis.Events[i]
			md.WriteString(fmt.Sprintf("| %.2fms | %s | %.2fms | %d | %s |\n",
				e.StartTime, e.Type, e.Duration, len(e.Threads), e.Description))
		}

		if len(analysis.Events) > 15 {
			md.WriteString(fmt.Sprintf("\n... and %d more events\n", len(analysis.Events)-15))
		}
	}

	if len(analysis.Recommendations) > 0 {
		md.WriteString("\n## Recommendations\n\n")
		for _, r := range analysis.Recommendations {
			md.WriteString(fmt.Sprintf("- 💡 %s\n", r))
		}
	}

	fmt.Print(md.String())
	return nil
}

func outputContentionText(analysis analyzer.ContentionAnalysis) error {
	fmt.Println("Contention Analysis")
	fmt.Println(strings.Repeat("=", 60))
	fmt.Println()

	fmt.Println("Summary:")
	fmt.Printf("  Severity:         %s\n", strings.ToUpper(analysis.Severity))
	fmt.Printf("  Total Events:     %d\n", analysis.TotalEvents)
	fmt.Printf("  Total Impact:     %.2f ms\n", analysis.TotalImpactMs)
	fmt.Printf("  GC Contention:    %d events\n", analysis.GCContention)
	fmt.Printf("  IPC Contention:   %d events\n", analysis.IPCContention)
	fmt.Printf("  Lock Contention:  %d events\n", analysis.LockContention)
	fmt.Println()

	if len(analysis.Events) > 0 {
		fmt.Println("Top Contention Events:")
		fmt.Println(strings.Repeat("-", 60))
		fmt.Printf("%-10s %-12s %10s %8s  %s\n", "Time", "Type", "Duration", "Threads", "Description")
		fmt.Println(strings.Repeat("-", 60))

		displayCount := len(analysis.Events)
		if displayCount > 10 {
			displayCount = 10
		}

		for i := 0; i < displayCount; i++ {
			e := analysis.Events[i]
			desc := e.Description
			if len(desc) > 25 {
				desc = desc[:22] + "..."
			}
			fmt.Printf("%8.2fms %-12s %8.2fms %8d  %s\n",
				e.StartTime, e.Type, e.Duration, len(e.Threads), desc)
		}

		if len(analysis.Events) > 10 {
			fmt.Printf("\n  ... and %d more events\n", len(analysis.Events)-10)
		}
		fmt.Println()
	}

	if len(analysis.Recommendations) > 0 {
		fmt.Println("Recommendations:")
		for _, r := range analysis.Recommendations {
			fmt.Printf("  - %s\n", r)
		}
	}

	return nil
}
