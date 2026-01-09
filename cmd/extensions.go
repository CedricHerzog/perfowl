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
	extensionID string
)

var extensionsCmd = &cobra.Command{
	Use:   "extensions",
	Short: "Analyze extension performance impact",
	Long: `Analyzes the performance impact of browser extensions in the profile.

Shows metrics for each extension including:
- Total time spent in extension code
- Number of markers/events triggered
- DOM events caused
- IPC messages sent
- Overall impact score`,
	RunE: runExtensions,
}

func init() {
	rootCmd.AddCommand(extensionsCmd)
	extensionsCmd.Flags().StringVarP(&extensionID, "extension", "e", "", "Filter by specific extension ID")
}

func runExtensions(cmd *cobra.Command, args []string) error {
	if profilePath == "" {
		return fmt.Errorf("profile path is required (use --profile or -p)")
	}

	profile, err := parser.LoadProfile(profilePath)
	if err != nil {
		return fmt.Errorf("failed to load profile: %w", err)
	}

	// Analyze extensions
	report := analyzer.AnalyzeExtensions(profile)

	// Filter if specific extension requested
	if extensionID != "" {
		var filtered []analyzer.ExtensionReport
		for _, ext := range report.Extensions {
			if ext.ID == extensionID || strings.Contains(ext.ID, extensionID) {
				filtered = append(filtered, ext)
			}
		}
		report.Extensions = filtered
	}

	switch outputFormat {
	case "json":
		return outputExtensionsJSON(report)
	case "markdown":
		return outputExtensionsMarkdown(report)
	default:
		return outputExtensionsText(report)
	}
}

func outputExtensionsJSON(report analyzer.ExtensionsAnalysis) error {
	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	return encoder.Encode(report)
}

func outputExtensionsMarkdown(report analyzer.ExtensionsAnalysis) error {
	md := strings.Builder{}

	md.WriteString("# Extension Performance Analysis\n\n")

	md.WriteString("## Summary\n\n")
	md.WriteString(fmt.Sprintf("- **Total Extensions**: %d\n", report.TotalExtensions))
	md.WriteString(fmt.Sprintf("- **Combined Duration**: %.2f ms\n", report.TotalDuration))
	md.WriteString(fmt.Sprintf("- **Total Events**: %d\n", report.TotalEvents))

	if len(report.Extensions) > 0 {
		md.WriteString("\n## Extension Details\n\n")

		// Sort by duration
		sort.Slice(report.Extensions, func(i, j int) bool {
			return report.Extensions[i].TotalDuration > report.Extensions[j].TotalDuration
		})

		for _, ext := range report.Extensions {
			impactEmoji := map[string]string{
				"low":    "🟢",
				"medium": "🟡",
				"high":   "🔴",
			}[ext.ImpactScore]

			md.WriteString(fmt.Sprintf("### %s %s\n\n", impactEmoji, ext.Name))
			md.WriteString(fmt.Sprintf("- **ID**: `%s`\n", ext.ID))
			md.WriteString(fmt.Sprintf("- **Impact**: %s\n", ext.ImpactScore))
			md.WriteString(fmt.Sprintf("- **Total Duration**: %.2f ms\n", ext.TotalDuration))
			md.WriteString(fmt.Sprintf("- **Markers/Events**: %d\n", ext.MarkersCount))
			md.WriteString(fmt.Sprintf("- **DOM Events**: %d\n", ext.DOMEvents))
			md.WriteString(fmt.Sprintf("- **IPC Messages**: %d\n", ext.IPCMessages))

			if len(ext.TopMarkers) > 0 {
				md.WriteString("\n**Top Activity:**\n")
				for _, marker := range ext.TopMarkers {
					md.WriteString(fmt.Sprintf("- %s (%.2fms)\n", marker.Name, marker.Duration))
				}
			}
			md.WriteString("\n")
		}
	} else {
		md.WriteString("\n## No Extension Activity Detected\n\n")
		md.WriteString("No significant extension activity was captured in this profile.\n")
	}

	fmt.Print(md.String())
	return nil
}

func outputExtensionsText(report analyzer.ExtensionsAnalysis) error {
	fmt.Println("Extension Performance Analysis")
	fmt.Println(strings.Repeat("=", 50))
	fmt.Println()

	fmt.Println("Summary:")
	fmt.Printf("  Total Extensions: %d\n", report.TotalExtensions)
	fmt.Printf("  Combined Duration: %.2f ms\n", report.TotalDuration)
	fmt.Printf("  Total Events: %d\n", report.TotalEvents)
	fmt.Println()

	if len(report.Extensions) > 0 {
		// Sort by duration
		sort.Slice(report.Extensions, func(i, j int) bool {
			return report.Extensions[i].TotalDuration > report.Extensions[j].TotalDuration
		})

		fmt.Println("Extension Details:")
		fmt.Println(strings.Repeat("-", 50))

		for _, ext := range report.Extensions {
			impactStr := map[string]string{
				"low":    "[LOW]   ",
				"medium": "[MEDIUM]",
				"high":   "[HIGH]  ",
			}[ext.ImpactScore]

			fmt.Printf("\n%s %s\n", impactStr, ext.Name)
			fmt.Printf("  ID:             %s\n", ext.ID)
			fmt.Printf("  Total Duration: %.2f ms\n", ext.TotalDuration)
			fmt.Printf("  Markers/Events: %d\n", ext.MarkersCount)
			fmt.Printf("  DOM Events:     %d\n", ext.DOMEvents)
			fmt.Printf("  IPC Messages:   %d\n", ext.IPCMessages)

			if len(ext.TopMarkers) > 0 {
				fmt.Println("  Top Activity:")
				for _, marker := range ext.TopMarkers {
					fmt.Printf("    - %s (%.2fms)\n", marker.Name, marker.Duration)
				}
			}
		}
	} else {
		fmt.Println("No Extension Activity Detected")
		fmt.Println("  No significant extension activity was captured in this profile.")
	}

	return nil
}
