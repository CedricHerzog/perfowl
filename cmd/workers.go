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

var workersCmd = &cobra.Command{
	Use:   "workers",
	Short: "Analyze worker thread performance",
	Long: `Analyzes worker thread performance including:
- CPU time and idle time per worker
- Worker utilization percentage
- Message counts (sent/received)
- Synchronization waits
- Sync points between workers

Identifies potential issues like:
- Worker starvation (low utilization)
- Excessive synchronous waits
- Contention between workers`,
	RunE: runWorkers,
}

func init() {
	rootCmd.AddCommand(workersCmd)
}

func runWorkers(cmd *cobra.Command, args []string) error {
	if profilePath == "" {
		return fmt.Errorf("profile path is required (use --profile or -p)")
	}

	bt := parser.ParseBrowserType(browserType)
	profile, _, err := parser.LoadProfileWithType(profilePath, bt)
	if err != nil {
		return fmt.Errorf("failed to load profile: %w", err)
	}

	analysis := analyzer.AnalyzeWorkers(profile)

	switch outputFormat {
	case "json":
		return outputWorkersJSON(analysis)
	case "markdown":
		return outputWorkersMarkdown(analysis)
	default:
		return outputWorkersText(analysis)
	}
}

func outputWorkersJSON(analysis analyzer.WorkerAnalysis) error {
	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	return encoder.Encode(analysis)
}

func outputWorkersMarkdown(analysis analyzer.WorkerAnalysis) error {
	md := strings.Builder{}

	md.WriteString("# Worker Thread Analysis\n\n")

	md.WriteString("## Summary\n\n")
	md.WriteString(fmt.Sprintf("- **Total Workers**: %d\n", analysis.TotalWorkers))
	md.WriteString(fmt.Sprintf("- **Active Workers**: %d\n", analysis.ActiveWorkers))
	md.WriteString(fmt.Sprintf("- **Overall Efficiency**: %.1f%%\n", analysis.OverallEfficiency))
	md.WriteString(fmt.Sprintf("- **Total CPU Time**: %.2f ms\n", analysis.TotalCPUTimeMs))
	md.WriteString(fmt.Sprintf("- **Total Idle Time**: %.2f ms\n", analysis.TotalIdleTimeMs))

	if len(analysis.Workers) > 0 {
		md.WriteString("\n## Worker Details\n\n")
		md.WriteString("| Worker | CPU Time | Idle Time | Active % | Messages | Sync Waits |\n")
		md.WriteString("|--------|----------|-----------|----------|----------|------------|\n")

		for _, w := range analysis.Workers {
			name := w.ThreadName
			if len(name) > 20 {
				name = name[:17] + "..."
			}
			md.WriteString(fmt.Sprintf("| %s | %.2fms | %.2fms | %.1f%% | %d | %d |\n",
				name, w.CPUTimeMs, w.IdleTimeMs, w.ActivePercent, w.MessagesSent, w.SyncWaitCount))
		}
	}

	if len(analysis.SyncPoints) > 0 {
		md.WriteString("\n## Synchronization Points\n\n")
		md.WriteString(fmt.Sprintf("Found **%d** sync points where multiple workers blocked simultaneously.\n\n", len(analysis.SyncPoints)))

		displayCount := len(analysis.SyncPoints)
		if displayCount > 10 {
			displayCount = 10
		}

		for i := 0; i < displayCount; i++ {
			sp := analysis.SyncPoints[i]
			md.WriteString(fmt.Sprintf("- **%.2fms**: %s (%.2fms duration)\n", sp.Time, sp.Description, sp.Duration))
			md.WriteString(fmt.Sprintf("  - Threads: %s\n", strings.Join(sp.Threads, ", ")))
		}

		if len(analysis.SyncPoints) > 10 {
			md.WriteString(fmt.Sprintf("\n... and %d more sync points\n", len(analysis.SyncPoints)-10))
		}
	}

	if len(analysis.Warnings) > 0 {
		md.WriteString("\n## Warnings\n\n")
		for _, w := range analysis.Warnings {
			md.WriteString(fmt.Sprintf("- %s\n", w))
		}
	}

	fmt.Print(md.String())
	return nil
}

func outputWorkersText(analysis analyzer.WorkerAnalysis) error {
	fmt.Println("Worker Thread Analysis")
	fmt.Println(strings.Repeat("=", 60))
	fmt.Println()

	fmt.Println("Summary:")
	fmt.Printf("  Total Workers:      %d\n", analysis.TotalWorkers)
	fmt.Printf("  Active Workers:     %d\n", analysis.ActiveWorkers)
	fmt.Printf("  Overall Efficiency: %.1f%%\n", analysis.OverallEfficiency)
	fmt.Printf("  Total CPU Time:     %.2f ms\n", analysis.TotalCPUTimeMs)
	fmt.Printf("  Total Idle Time:    %.2f ms\n", analysis.TotalIdleTimeMs)
	fmt.Println()

	if len(analysis.Workers) > 0 {
		fmt.Println("Worker Details:")
		fmt.Println(strings.Repeat("-", 60))
		fmt.Printf("%-22s %10s %10s %8s %6s\n", "Name", "CPU Time", "Idle Time", "Active%", "Msgs")
		fmt.Println(strings.Repeat("-", 60))

		for _, w := range analysis.Workers {
			name := w.ThreadName
			if len(name) > 22 {
				name = name[:19] + "..."
			}
			fmt.Printf("%-22s %8.2fms %8.2fms %7.1f%% %6d\n",
				name, w.CPUTimeMs, w.IdleTimeMs, w.ActivePercent, w.MessagesSent)
		}
		fmt.Println()
	}

	if len(analysis.SyncPoints) > 0 {
		fmt.Printf("Synchronization Points: %d detected\n", len(analysis.SyncPoints))
		fmt.Println(strings.Repeat("-", 60))

		displayCount := len(analysis.SyncPoints)
		if displayCount > 5 {
			displayCount = 5
		}

		for i := 0; i < displayCount; i++ {
			sp := analysis.SyncPoints[i]
			fmt.Printf("  %.2fms: %s (%.2fms)\n", sp.Time, sp.Description, sp.Duration)
		}

		if len(analysis.SyncPoints) > 5 {
			fmt.Printf("  ... and %d more\n", len(analysis.SyncPoints)-5)
		}
		fmt.Println()
	}

	if len(analysis.Warnings) > 0 {
		fmt.Println("Warnings:")
		for _, w := range analysis.Warnings {
			fmt.Printf("  - %s\n", w)
		}
	}

	return nil
}
