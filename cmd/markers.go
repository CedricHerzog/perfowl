package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/CedricHerzog/perfowl/internal/parser"
	"github.com/spf13/cobra"
)

var (
	markerType     string
	markerCategory string
	minDuration    float64
	limit          int
)

var markersCmd = &cobra.Command{
	Use:   "markers",
	Short: "Extract and analyze markers from the profile",
	Long: `Extract markers from the browser profile, optionally filtered by type or category.

Marker types include: GCMajor, GCMinor, DOMEvent, Styles, UserTiming,
MainThreadLongTask, ChannelMarker, HostResolver, JSActorMessage, etc.

Categories include: JavaScript, Layout, GC / CC, Network, Graphics, DOM, IPC, etc.`,
	RunE: runMarkers,
}

func init() {
	rootCmd.AddCommand(markersCmd)
	markersCmd.Flags().StringVarP(&markerType, "type", "t", "", "Filter by marker type (e.g., GCMajor, DOMEvent)")
	markersCmd.Flags().StringVarP(&markerCategory, "category", "c", "", "Filter by category (e.g., JavaScript, Layout)")
	markersCmd.Flags().Float64VarP(&minDuration, "min-duration", "d", 0, "Minimum duration in ms")
	markersCmd.Flags().IntVarP(&limit, "limit", "l", 0, "Limit number of results (0 = unlimited)")
}

type MarkerOutput struct {
	TotalCount int                     `json:"total_count"`
	FilteredBy string                  `json:"filtered_by,omitempty"`
	Stats      parser.MarkerStats      `json:"stats"`
	Markers    []parser.ParsedMarker   `json:"markers,omitempty"`
}

func runMarkers(cmd *cobra.Command, args []string) error {
	if profilePath == "" {
		return fmt.Errorf("profile path is required (use --profile or -p)")
	}

	bt := parser.ParseBrowserType(browserType)
	profile, _, err := parser.LoadProfileWithType(profilePath, bt)
	if err != nil {
		return fmt.Errorf("failed to load profile: %w", err)
	}

	// Extract all markers from all threads
	var allMarkers []parser.ParsedMarker
	for _, thread := range profile.Threads {
		markers := parser.ExtractMarkers(&thread, profile.Meta.Categories)
		allMarkers = append(allMarkers, markers...)
	}

	// Apply filters
	filtered := allMarkers
	filterDesc := []string{}

	if markerType != "" {
		filtered = parser.FilterMarkersByType(filtered, parser.MarkerType(markerType))
		filterDesc = append(filterDesc, fmt.Sprintf("type=%s", markerType))
	}

	if markerCategory != "" {
		filtered = parser.FilterMarkersByCategory(filtered, markerCategory)
		filterDesc = append(filterDesc, fmt.Sprintf("category=%s", markerCategory))
	}

	if minDuration > 0 {
		filtered = parser.FilterMarkersByDuration(filtered, minDuration)
		filterDesc = append(filterDesc, fmt.Sprintf("min_duration=%.2fms", minDuration))
	}

	// Sort by duration (descending)
	sort.Slice(filtered, func(i, j int) bool {
		return filtered[i].Duration > filtered[j].Duration
	})

	// Apply limit
	if limit > 0 && len(filtered) > limit {
		filtered = filtered[:limit]
	}

	stats := parser.GetMarkerStats(filtered)

	output := MarkerOutput{
		TotalCount: len(allMarkers),
		FilteredBy: strings.Join(filterDesc, ", "),
		Stats:      stats,
		Markers:    filtered,
	}

	switch outputFormat {
	case "json":
		return outputMarkersJSON(output)
	case "markdown":
		return outputMarkersMarkdown(output)
	default:
		return outputMarkersText(output)
	}
}

func outputMarkersJSON(output MarkerOutput) error {
	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	return encoder.Encode(output)
}

func outputMarkersMarkdown(output MarkerOutput) error {
	md := strings.Builder{}

	md.WriteString("# Marker Analysis\n\n")

	if output.FilteredBy != "" {
		md.WriteString(fmt.Sprintf("**Filters applied**: %s\n\n", output.FilteredBy))
	}

	md.WriteString("## Statistics\n\n")
	md.WriteString(fmt.Sprintf("- **Total markers in profile**: %d\n", output.TotalCount))
	md.WriteString(fmt.Sprintf("- **Filtered markers**: %d\n", output.Stats.TotalCount))
	md.WriteString(fmt.Sprintf("- **Total duration**: %.2f ms\n", output.Stats.TotalDuration))
	md.WriteString(fmt.Sprintf("- **Average duration**: %.2f ms\n", output.Stats.AvgDuration))
	md.WriteString(fmt.Sprintf("- **Max duration**: %.2f ms\n", output.Stats.MaxDuration))

	md.WriteString("\n## By Type\n\n")
	md.WriteString("| Type | Count |\n|------|-------|\n")
	for t, count := range output.Stats.ByType {
		md.WriteString(fmt.Sprintf("| %s | %d |\n", t, count))
	}

	md.WriteString("\n## By Category\n\n")
	md.WriteString("| Category | Count |\n|----------|-------|\n")
	for cat, count := range output.Stats.ByCategory {
		md.WriteString(fmt.Sprintf("| %s | %d |\n", cat, count))
	}

	if len(output.Markers) > 0 {
		md.WriteString("\n## Top Markers by Duration\n\n")
		md.WriteString("| Name | Category | Duration (ms) | Thread |\n")
		md.WriteString("|------|----------|---------------|--------|\n")
		displayCount := min(20, len(output.Markers))
		for i := 0; i < displayCount; i++ {
			m := output.Markers[i]
			md.WriteString(fmt.Sprintf("| %s | %s | %.2f | %s |\n", m.Name, m.Category, m.Duration, m.ThreadName))
		}
	}

	fmt.Print(md.String())
	return nil
}

func outputMarkersText(output MarkerOutput) error {
	fmt.Println("Marker Analysis")
	fmt.Println(strings.Repeat("=", 50))
	fmt.Println()

	if output.FilteredBy != "" {
		fmt.Printf("Filters applied: %s\n\n", output.FilteredBy)
	}

	fmt.Println("Statistics:")
	fmt.Printf("  Total markers in profile: %d\n", output.TotalCount)
	fmt.Printf("  Filtered markers:         %d\n", output.Stats.TotalCount)
	fmt.Printf("  Total duration:           %.2f ms\n", output.Stats.TotalDuration)
	fmt.Printf("  Average duration:         %.2f ms\n", output.Stats.AvgDuration)
	fmt.Printf("  Max duration:             %.2f ms\n", output.Stats.MaxDuration)
	fmt.Println()

	fmt.Println("By Type:")
	for t, count := range output.Stats.ByType {
		fmt.Printf("  %-25s %d\n", t, count)
	}
	fmt.Println()

	fmt.Println("By Category:")
	for cat, count := range output.Stats.ByCategory {
		fmt.Printf("  %-25s %d\n", cat, count)
	}

	if len(output.Markers) > 0 {
		fmt.Println()
		fmt.Println("Top Markers by Duration:")
		fmt.Printf("  %-30s %-15s %-12s %s\n", "Name", "Category", "Duration", "Thread")
		fmt.Println("  " + strings.Repeat("-", 80))
		displayCount := min(20, len(output.Markers))
		for i := 0; i < displayCount; i++ {
			m := output.Markers[i]
			name := m.Name
			if len(name) > 28 {
				name = name[:28] + ".."
			}
			fmt.Printf("  %-30s %-15s %8.2f ms  %s\n", name, m.Category, m.Duration, m.ThreadName)
		}
	}

	return nil
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
