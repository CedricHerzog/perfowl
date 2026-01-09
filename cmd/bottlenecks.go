package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/CedricHerzog/perfowl/internal/analyzer"
	"github.com/CedricHerzog/perfowl/internal/parser"
	"github.com/spf13/cobra"
)

var (
	minSeverity string
)

var bottlenecksCmd = &cobra.Command{
	Use:   "bottlenecks",
	Short: "Detect performance bottlenecks in the profile",
	Long: `Analyzes the Firefox profile to identify performance bottlenecks.

Detects:
- Long tasks (main thread blocking > 50ms)
- Synchronous IPC calls
- GC pressure (frequent or long garbage collection)
- Layout thrashing (rapid reflow cycles)
- Slow JavaScript execution
- Network blocking
- Extension overhead`,
	RunE: runBottlenecks,
}

func init() {
	rootCmd.AddCommand(bottlenecksCmd)
	bottlenecksCmd.Flags().StringVarP(&minSeverity, "min-severity", "s", "", "Minimum severity to report: low, medium, high")
}

func runBottlenecks(cmd *cobra.Command, args []string) error {
	if profilePath == "" {
		return fmt.Errorf("profile path is required (use --profile or -p)")
	}

	profile, err := parser.LoadProfile(profilePath)
	if err != nil {
		return fmt.Errorf("failed to load profile: %w", err)
	}

	// Run bottleneck detection
	bottlenecks := analyzer.DetectBottlenecks(profile)

	// Filter by severity if specified
	if minSeverity != "" {
		minSev := analyzer.ParseSeverity(minSeverity)
		var filtered []analyzer.Bottleneck
		for _, b := range bottlenecks {
			if b.Severity >= minSev {
				filtered = append(filtered, b)
			}
		}
		bottlenecks = filtered
	}

	// Sort by severity (highest first)
	sort.Slice(bottlenecks, func(i, j int) bool {
		return bottlenecks[i].Severity > bottlenecks[j].Severity
	})

	// Calculate overall score
	score := analyzer.CalculateScore(bottlenecks)

	output := analyzer.BottleneckReport{
		Score:       score,
		Bottlenecks: bottlenecks,
		Summary:     analyzer.GenerateSummary(bottlenecks, profile),
	}

	switch outputFormat {
	case "json":
		return outputBottlenecksJSON(output)
	case "markdown":
		return outputBottlenecksMarkdown(output, profile)
	default:
		return outputBottlenecksText(output, profile)
	}
}

func outputBottlenecksJSON(output analyzer.BottleneckReport) error {
	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	return encoder.Encode(output)
}

func outputBottlenecksMarkdown(output analyzer.BottleneckReport, profile *parser.Profile) error {
	md := strings.Builder{}

	md.WriteString("# Firefox Profile Bottleneck Analysis\n\n")

	md.WriteString("## Overview\n\n")
	md.WriteString(fmt.Sprintf("- **Duration**: %.2f seconds\n", profile.DurationSeconds()))
	md.WriteString(fmt.Sprintf("- **Platform**: %s (%s)\n", profile.Meta.Platform, profile.Meta.CPUName))
	md.WriteString(fmt.Sprintf("- **Performance Score**: %d/100\n\n", output.Score))

	if output.Score >= 80 {
		md.WriteString("**Status**: Good performance, minor issues detected.\n\n")
	} else if output.Score >= 60 {
		md.WriteString("**Status**: Moderate performance issues detected.\n\n")
	} else {
		md.WriteString("**Status**: Significant performance issues detected.\n\n")
	}

	md.WriteString("## Summary\n\n")
	md.WriteString(output.Summary + "\n\n")

	if len(output.Bottlenecks) > 0 {
		md.WriteString("## Bottlenecks Detected\n\n")

		for _, b := range output.Bottlenecks {
			sevEmoji := map[analyzer.Severity]string{
				analyzer.SeverityHigh:   "🔴",
				analyzer.SeverityMedium: "🟡",
				analyzer.SeverityLow:    "🟢",
			}[b.Severity]

			md.WriteString(fmt.Sprintf("### %s %s (%s)\n\n", sevEmoji, b.Type, b.Severity))
			md.WriteString(fmt.Sprintf("**Description**: %s\n\n", b.Description))
			md.WriteString(fmt.Sprintf("- **Occurrences**: %d\n", b.Count))
			md.WriteString(fmt.Sprintf("- **Total Duration**: %.2f ms\n", b.TotalDuration))

			if b.Recommendation != "" {
				md.WriteString(fmt.Sprintf("\n**Recommendation**: %s\n", b.Recommendation))
			}

			if len(b.Locations) > 0 {
				md.WriteString("\n**Locations**:\n")
				displayCount := min(5, len(b.Locations))
				for i := 0; i < displayCount; i++ {
					md.WriteString(fmt.Sprintf("- %s\n", b.Locations[i]))
				}
				if len(b.Locations) > 5 {
					md.WriteString(fmt.Sprintf("- ... and %d more\n", len(b.Locations)-5))
				}
			}
			md.WriteString("\n")
		}

		md.WriteString("## Recommendations Summary\n\n")
		for i, b := range output.Bottlenecks {
			if b.Recommendation != "" {
				md.WriteString(fmt.Sprintf("%d. %s\n", i+1, b.Recommendation))
			}
		}
	} else {
		md.WriteString("## No Significant Bottlenecks Detected\n\n")
		md.WriteString("The profile shows good performance characteristics.\n")
	}

	fmt.Print(md.String())
	return nil
}

func outputBottlenecksText(output analyzer.BottleneckReport, profile *parser.Profile) error {
	fmt.Println("Firefox Profile Bottleneck Analysis")
	fmt.Println(strings.Repeat("=", 50))
	fmt.Println()

	fmt.Println("Overview:")
	fmt.Printf("  Duration:          %.2f seconds\n", profile.DurationSeconds())
	fmt.Printf("  Platform:          %s (%s)\n", profile.Meta.Platform, profile.Meta.CPUName)
	fmt.Printf("  Performance Score: %d/100\n", output.Score)
	fmt.Println()

	fmt.Println("Summary:")
	fmt.Printf("  %s\n", output.Summary)
	fmt.Println()

	if len(output.Bottlenecks) > 0 {
		fmt.Println("Bottlenecks Detected:")
		fmt.Println(strings.Repeat("-", 50))

		for _, b := range output.Bottlenecks {
			sevStr := map[analyzer.Severity]string{
				analyzer.SeverityHigh:   "[HIGH]  ",
				analyzer.SeverityMedium: "[MEDIUM]",
				analyzer.SeverityLow:    "[LOW]   ",
			}[b.Severity]

			fmt.Printf("\n%s %s\n", sevStr, b.Type)
			fmt.Printf("  Description: %s\n", b.Description)
			fmt.Printf("  Occurrences: %d\n", b.Count)
			fmt.Printf("  Total Duration: %.2f ms\n", b.TotalDuration)

			if b.Recommendation != "" {
				fmt.Printf("  Recommendation: %s\n", b.Recommendation)
			}
		}

		fmt.Println()
		fmt.Println("Recommendations Summary:")
		for i, b := range output.Bottlenecks {
			if b.Recommendation != "" {
				fmt.Printf("  %d. %s\n", i+1, b.Recommendation)
			}
		}
	} else {
		fmt.Println("No Significant Bottlenecks Detected")
		fmt.Println("  The profile shows good performance characteristics.")
	}

	return nil
}
