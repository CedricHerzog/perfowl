package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/CedricHerzog/perfowl/internal/parser"
	"github.com/spf13/cobra"
)

var summaryCmd = &cobra.Command{
	Use:   "summary",
	Short: "Display a summary of the browser profile",
	Long:  `Shows key information about the profile including duration, threads, extensions, and captured features.`,
	RunE:  runSummary,
}

func init() {
	rootCmd.AddCommand(summaryCmd)
}

type ProfileSummary struct {
	BrowserType     string             `json:"browser_type"`
	Duration        float64            `json:"duration_seconds"`
	Platform        string             `json:"platform"`
	OSCPU           string             `json:"os_cpu"`
	Product         string             `json:"product"`
	BuildID         string             `json:"build_id"`
	CPUName         string             `json:"cpu_name"`
	PhysicalCPUs    int                `json:"physical_cpus"`
	LogicalCPUs     int                `json:"logical_cpus"`
	ThreadCount     int                `json:"thread_count"`
	MainThreadCount int                `json:"main_thread_count"`
	ExtensionCount  int                `json:"extension_count"`
	Extensions      map[string]string  `json:"extensions"`
	Features        []string           `json:"features"`
	Categories      []string           `json:"categories"`
	TotalMarkers    int                `json:"total_markers"`
	TotalSamples    int                `json:"total_samples"`
}

func runSummary(cmd *cobra.Command, args []string) error {
	if profilePath == "" {
		return fmt.Errorf("profile path is required (use --profile or -p)")
	}

	bt := parser.ParseBrowserType(browserType)
	profile, detectedType, err := parser.LoadProfileWithType(profilePath, bt)
	if err != nil {
		return fmt.Errorf("failed to load profile: %w", err)
	}

	summary := buildSummary(profile, detectedType)

	switch outputFormat {
	case "json":
		return outputJSON(summary)
	case "markdown":
		return outputMarkdown(summary)
	default:
		return outputText(summary)
	}
}

func buildSummary(profile *parser.Profile, bt parser.BrowserType) ProfileSummary {
	summary := ProfileSummary{
		BrowserType:    string(bt),
		Duration:       profile.DurationSeconds(),
		Platform:       profile.Meta.Platform,
		OSCPU:          profile.Meta.OSCPU,
		Product:        profile.Meta.Product,
		BuildID:        profile.Meta.AppBuildID,
		CPUName:        profile.Meta.CPUName,
		PhysicalCPUs:   profile.Meta.PhysicalCPUs,
		LogicalCPUs:    profile.Meta.LogicalCPUs,
		ThreadCount:    profile.ThreadCount(),
		ExtensionCount: profile.ExtensionCount(),
		Extensions:     profile.GetExtensions(),
		Features:       profile.Meta.Configuration.Features,
	}

	// Count main threads
	for _, t := range profile.Threads {
		if t.IsMainThread {
			summary.MainThreadCount++
		}
		summary.TotalMarkers += t.Markers.Length
		summary.TotalSamples += t.Samples.Length
	}

	// Get category names
	for _, cat := range profile.Meta.Categories {
		summary.Categories = append(summary.Categories, cat.Name)
	}

	return summary
}

func outputJSON(summary ProfileSummary) error {
	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	return encoder.Encode(summary)
}

func outputMarkdown(summary ProfileSummary) error {
	md := strings.Builder{}

	browserName := strings.Title(summary.BrowserType)
	if browserName == "" {
		browserName = "Browser"
	}
	md.WriteString(fmt.Sprintf("# %s Profile Summary\n\n", browserName))

	md.WriteString("## Overview\n\n")
	md.WriteString(fmt.Sprintf("- **Browser**: %s\n", browserName))
	md.WriteString(fmt.Sprintf("- **Duration**: %.2f seconds\n", summary.Duration))
	md.WriteString(fmt.Sprintf("- **Platform**: %s (%s)\n", summary.Platform, summary.OSCPU))
	md.WriteString(fmt.Sprintf("- **Product**: %s (Build: %s)\n", summary.Product, summary.BuildID))
	md.WriteString(fmt.Sprintf("- **CPU**: %s (%d physical, %d logical cores)\n", summary.CPUName, summary.PhysicalCPUs, summary.LogicalCPUs))

	md.WriteString("\n## Profiling Data\n\n")
	md.WriteString(fmt.Sprintf("- **Threads**: %d total (%d main threads)\n", summary.ThreadCount, summary.MainThreadCount))
	md.WriteString(fmt.Sprintf("- **Total Markers**: %d\n", summary.TotalMarkers))
	md.WriteString(fmt.Sprintf("- **Total Samples**: %d\n", summary.TotalSamples))

	md.WriteString("\n## Captured Features\n\n")
	for _, feature := range summary.Features {
		md.WriteString(fmt.Sprintf("- %s\n", feature))
	}

	if summary.ExtensionCount > 0 {
		md.WriteString("\n## Extensions\n\n")
		md.WriteString(fmt.Sprintf("**%d extensions active during profiling:**\n\n", summary.ExtensionCount))
		for id, name := range summary.Extensions {
			md.WriteString(fmt.Sprintf("- **%s** (`%s`)\n", name, id))
		}
	}

	fmt.Print(md.String())
	return nil
}

func outputText(summary ProfileSummary) error {
	browserName := strings.Title(summary.BrowserType)
	if browserName == "" {
		browserName = "Browser"
	}
	fmt.Printf("%s Profile Summary\n", browserName)
	fmt.Println(strings.Repeat("=", 50))
	fmt.Println()

	fmt.Println("Overview:")
	fmt.Printf("  Browser:      %s\n", browserName)
	fmt.Printf("  Duration:     %.2f seconds\n", summary.Duration)
	fmt.Printf("  Platform:     %s (%s)\n", summary.Platform, summary.OSCPU)
	fmt.Printf("  Product:      %s\n", summary.Product)
	fmt.Printf("  Build ID:     %s\n", summary.BuildID)
	fmt.Printf("  CPU:          %s\n", summary.CPUName)
	fmt.Printf("  Cores:        %d physical, %d logical\n", summary.PhysicalCPUs, summary.LogicalCPUs)
	fmt.Println()

	fmt.Println("Profiling Data:")
	fmt.Printf("  Threads:      %d total (%d main)\n", summary.ThreadCount, summary.MainThreadCount)
	fmt.Printf("  Markers:      %d\n", summary.TotalMarkers)
	fmt.Printf("  Samples:      %d\n", summary.TotalSamples)
	fmt.Println()

	fmt.Println("Features Captured:")
	for _, feature := range summary.Features {
		fmt.Printf("  - %s\n", feature)
	}
	fmt.Println()

	if summary.ExtensionCount > 0 {
		fmt.Printf("Extensions (%d):\n", summary.ExtensionCount)
		for id, name := range summary.Extensions {
			fmt.Printf("  - %s\n", name)
			fmt.Printf("    ID: %s\n", id)
		}
	}

	return nil
}
