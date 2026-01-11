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

var (
	startPattern string
	endPattern   string
	findLast     bool
)

var measureCmd = &cobra.Command{
	Use:   "measure",
	Short: "Measure duration between two markers",
	Long: `Measure the time between two markers identified by patterns.

Pattern formats:
  - "UserTiming:marker-name" - matches UserTiming markers by name
  - "DOMEvent:click" - matches DOMEvent markers by event type
  - "Styles" - matches any marker with type or name "Styles"

Examples:
  # Measure between two user timing markers
  perfowl measure -p profile.json.gz \
    --start "UserTiming:perfowl-login-start" \
    --end "UserTiming:perfowl-resources-decrypted"

  # Measure from click event to paint
  perfowl measure -p profile.json.gz \
    --start "DOMEvent:click" \
    --end "Paint"`,
	RunE: runMeasure,
}

func init() {
	rootCmd.AddCommand(measureCmd)
	measureCmd.Flags().StringVarP(&startPattern, "start", "s", "", "Pattern for start marker (required)")
	measureCmd.Flags().StringVarP(&endPattern, "end", "e", "", "Pattern for end marker (required)")
	measureCmd.Flags().BoolVarP(&findLast, "find-last", "L", false, "Find the last matching end marker instead of first")
	_ = measureCmd.MarkFlagRequired("start")
	_ = measureCmd.MarkFlagRequired("end")
}

type MeasureOutput struct {
	StartMarker     MarkerInfo `json:"start_marker"`
	EndMarker       MarkerInfo `json:"end_marker"`
	OperationTimeMs float64    `json:"operation_time_ms"`
}

type MarkerInfo struct {
	Name       string  `json:"name"`
	Type       string  `json:"type"`
	Category   string  `json:"category"`
	TimeMs     float64 `json:"time_ms"`
	DurationMs float64 `json:"duration_ms,omitempty"`
	Thread     string  `json:"thread"`
}

func runMeasure(cmd *cobra.Command, args []string) error {
	if profilePath == "" {
		return fmt.Errorf("profile path is required (use --profile or -p)")
	}

	bt := parser.ParseBrowserType(browserType)
	profile, _, err := parser.LoadProfileWithType(profilePath, bt)
	if err != nil {
		return fmt.Errorf("failed to load profile: %w", err)
	}

	opts := analyzer.MeasureOptions{
		StartPattern: startPattern,
		EndPattern:   endPattern,
		FindLast:     findLast,
	}

	measurement, err := analyzer.MeasureOperationAdvanced(profile, opts)
	if err != nil {
		return fmt.Errorf("failed to measure operation: %w", err)
	}

	output := MeasureOutput{
		StartMarker: MarkerInfo{
			Name:       measurement.StartMarker.Name,
			Type:       measurement.StartMarker.Type,
			Category:   measurement.StartMarker.Category,
			TimeMs:     measurement.StartMarker.TimeMs,
			DurationMs: measurement.StartMarker.DurationMs,
			Thread:     measurement.StartMarker.Thread,
		},
		EndMarker: MarkerInfo{
			Name:       measurement.EndMarker.Name,
			Type:       measurement.EndMarker.Type,
			Category:   measurement.EndMarker.Category,
			TimeMs:     measurement.EndMarker.TimeMs,
			DurationMs: measurement.EndMarker.DurationMs,
			Thread:     measurement.EndMarker.Thread,
		},
		OperationTimeMs: measurement.OperationTimeMs,
	}

	switch outputFormat {
	case "json":
		return outputMeasureJSON(output)
	case "markdown":
		return outputMeasureMarkdown(output)
	default:
		return outputMeasureText(output)
	}
}

func outputMeasureJSON(output MeasureOutput) error {
	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	return encoder.Encode(output)
}

func outputMeasureMarkdown(output MeasureOutput) error {
	md := strings.Builder{}

	md.WriteString("# Operation Measurement\n\n")

	md.WriteString("## Duration\n\n")
	md.WriteString(fmt.Sprintf("**%.2f ms**\n\n", output.OperationTimeMs))

	md.WriteString("## Start Marker\n\n")
	md.WriteString(fmt.Sprintf("- **Name**: %s\n", output.StartMarker.Name))
	md.WriteString(fmt.Sprintf("- **Type**: %s\n", output.StartMarker.Type))
	md.WriteString(fmt.Sprintf("- **Category**: %s\n", output.StartMarker.Category))
	md.WriteString(fmt.Sprintf("- **Time**: %.2f ms\n", output.StartMarker.TimeMs))
	md.WriteString(fmt.Sprintf("- **Thread**: %s\n", output.StartMarker.Thread))

	md.WriteString("\n## End Marker\n\n")
	md.WriteString(fmt.Sprintf("- **Name**: %s\n", output.EndMarker.Name))
	md.WriteString(fmt.Sprintf("- **Type**: %s\n", output.EndMarker.Type))
	md.WriteString(fmt.Sprintf("- **Category**: %s\n", output.EndMarker.Category))
	md.WriteString(fmt.Sprintf("- **Time**: %.2f ms\n", output.EndMarker.TimeMs))
	md.WriteString(fmt.Sprintf("- **Thread**: %s\n", output.EndMarker.Thread))

	fmt.Print(md.String())
	return nil
}

func outputMeasureText(output MeasureOutput) error {
	fmt.Println("Operation Measurement")
	fmt.Println(strings.Repeat("=", 50))
	fmt.Println()

	fmt.Printf("Duration: %.2f ms\n\n", output.OperationTimeMs)

	fmt.Println("Start Marker:")
	fmt.Printf("  Name:     %s\n", output.StartMarker.Name)
	fmt.Printf("  Type:     %s\n", output.StartMarker.Type)
	fmt.Printf("  Category: %s\n", output.StartMarker.Category)
	fmt.Printf("  Time:     %.2f ms\n", output.StartMarker.TimeMs)
	fmt.Printf("  Thread:   %s\n", output.StartMarker.Thread)

	fmt.Println()
	fmt.Println("End Marker:")
	fmt.Printf("  Name:     %s\n", output.EndMarker.Name)
	fmt.Printf("  Type:     %s\n", output.EndMarker.Type)
	fmt.Printf("  Category: %s\n", output.EndMarker.Category)
	fmt.Printf("  Time:     %.2f ms\n", output.EndMarker.TimeMs)
	fmt.Printf("  Thread:   %s\n", output.EndMarker.Thread)

	return nil
}
