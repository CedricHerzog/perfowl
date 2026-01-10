package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/CedricHerzog/perfowl/internal/analyzer"
	"github.com/CedricHerzog/perfowl/internal/chart"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

var (
	batchConfigFile string
	batchProfiles   string
	batchChartType  string
	batchSVGPath    string
)

// BatchConfig is the structure of the YAML/JSON config file
type BatchConfig struct {
	Profiles []analyzer.ProfileEntry `json:"profiles" yaml:"profiles"`
}

var batchCmd = &cobra.Command{
	Use:   "batch",
	Short: "Batch analyze multiple profiles and generate charts",
	Long: `Analyzes multiple browser profiles across worker counts and generates
visualization charts showing scaling behavior.

Profiles can be specified via a config file (YAML or JSON) or inline.

Example usage with config file:
  perfowl batch --config profiles.yaml -o text
  perfowl batch --config profiles.yaml --chart wall_clock -o svg > chart.svg
  perfowl batch --config profiles.yaml --chart efficiency -o svg-file --svg-path chart.svg

Example config file (profiles.yaml):
  profiles:
    - path: profiles/chrome/chrome-sequential.json.gz
      workers: 0
      label: Chrome
    - path: profiles/chrome/chrome-1-core.json.gz
      workers: 1
      label: Chrome
    - path: profiles/firefox/sequential.json.gz
      workers: 0
      label: Firefox

Example inline usage:
  perfowl batch --profiles "path1.json:0:Chrome,path2.json:1:Chrome" -o text`,
	RunE: runBatch,
}

func init() {
	rootCmd.AddCommand(batchCmd)

	batchCmd.Flags().StringVarP(&batchConfigFile, "config", "c", "", "Path to YAML/JSON config file")
	batchCmd.Flags().StringVar(&batchProfiles, "profiles", "", "Inline profiles: path:workers:label,...")
	batchCmd.Flags().StringVar(&batchChartType, "chart", "wall_clock", "Chart type: wall_clock, efficiency, speedup, crypto_time")
	batchCmd.Flags().StringVar(&batchSVGPath, "svg-path", "chart.svg", "Output path for svg-file mode")
}

func runBatch(cmd *cobra.Command, args []string) error {
	// Parse profiles from config file or inline
	profiles, err := parseProfiles()
	if err != nil {
		return err
	}

	if len(profiles) == 0 {
		return fmt.Errorf("no profiles specified. Use --config or --profiles")
	}

	// Run batch analysis
	result, err := analyzer.AnalyzeBatch(profiles)
	if err != nil {
		return fmt.Errorf("batch analysis failed: %w", err)
	}

	// Output based on format
	switch outputFormat {
	case "json":
		return outputBatchJSON(result)
	case "svg":
		return outputBatchSVG(result, false)
	case "svg-file":
		return outputBatchSVG(result, true)
	case "markdown":
		return outputBatchMarkdown(result)
	default:
		return outputBatchText(result)
	}
}

func parseProfiles() ([]analyzer.ProfileEntry, error) {
	var profiles []analyzer.ProfileEntry

	// Try config file first
	if batchConfigFile != "" {
		data, err := os.ReadFile(batchConfigFile)
		if err != nil {
			return nil, fmt.Errorf("failed to read config file: %w", err)
		}

		var config BatchConfig

		// Try YAML first, then JSON
		if err := yaml.Unmarshal(data, &config); err != nil {
			if err := json.Unmarshal(data, &config); err != nil {
				return nil, fmt.Errorf("failed to parse config file as YAML or JSON: %w", err)
			}
		}

		profiles = config.Profiles
	}

	// Parse inline profiles (can be used in addition to config)
	if batchProfiles != "" {
		inline, err := parseInlineProfiles(batchProfiles)
		if err != nil {
			return nil, err
		}
		profiles = append(profiles, inline...)
	}

	return profiles, nil
}

func parseInlineProfiles(input string) ([]analyzer.ProfileEntry, error) {
	var profiles []analyzer.ProfileEntry

	entries := strings.Split(input, ",")
	for _, entry := range entries {
		entry = strings.TrimSpace(entry)
		if entry == "" {
			continue
		}

		parts := strings.Split(entry, ":")
		if len(parts) != 3 {
			return nil, fmt.Errorf("invalid profile entry '%s': expected path:workers:label", entry)
		}

		var workers int
		if _, err := fmt.Sscanf(parts[1], "%d", &workers); err != nil {
			return nil, fmt.Errorf("invalid worker count in '%s': %w", entry, err)
		}

		profiles = append(profiles, analyzer.ProfileEntry{
			Path:        parts[0],
			WorkerCount: workers,
			Label:       parts[2],
		})
	}

	return profiles, nil
}

func outputBatchJSON(result *analyzer.BatchAnalysisResult) error {
	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	return encoder.Encode(result)
}

func outputBatchSVG(result *analyzer.BatchAnalysisResult, toFile bool) error {
	chartType := chart.ChartType(batchChartType)
	svg := chart.GenerateScalingChart(result, chartType)

	if toFile {
		if err := os.WriteFile(batchSVGPath, []byte(svg), 0644); err != nil {
			return fmt.Errorf("failed to write SVG file: %w", err)
		}
		fmt.Printf("Chart saved to: %s\n", batchSVGPath)
		return nil
	}

	fmt.Println(svg)
	return nil
}

func outputBatchText(result *analyzer.BatchAnalysisResult) error {
	fmt.Println("Batch Analysis Results")
	fmt.Println(strings.Repeat("=", 70))
	fmt.Printf("Total Profiles: %d\n", result.Summary.TotalProfiles)
	fmt.Printf("Labels: %s\n\n", strings.Join(result.Summary.Labels, ", "))

	// Print summary per label
	fmt.Println("Summary by Label:")
	fmt.Println(strings.Repeat("-", 70))
	fmt.Printf("%-12s %12s %14s %12s %12s\n", "Label", "Best Workers", "Min WallClock", "Max Speedup", "Peak Eff.")

	for _, label := range result.Summary.Labels {
		fmt.Printf("%-12s %12d %12.1fms %11.2fx %11.1f%%\n",
			label,
			result.Summary.BestWorkers[label],
			result.Summary.MinWallClock[label],
			result.Summary.MaxSpeedup[label],
			result.Summary.PeakEfficiency[label],
		)
	}
	fmt.Println()

	// Print detailed data per label
	for _, label := range result.Summary.Labels {
		points := result.Series[label]
		if len(points) == 0 {
			continue
		}

		fmt.Printf("%s Data Points:\n", label)
		fmt.Println(strings.Repeat("-", 70))
		fmt.Printf("%-8s %12s %12s %10s %10s %12s\n",
			"Workers", "WallClock", "TotalWork", "Speedup", "Efficiency", "CryptoTime")

		for _, p := range points {
			fmt.Printf("%-8d %10.1fms %10.1fms %9.2fx %9.1f%% %10.1fms\n",
				p.WorkerCount,
				p.WallClockMs,
				p.TotalWorkMs,
				p.Speedup,
				p.Efficiency,
				p.CryptoTimeMs,
			)
		}
		fmt.Println()
	}

	return nil
}

func outputBatchMarkdown(result *analyzer.BatchAnalysisResult) error {
	fmt.Println("# Batch Analysis Results")
	fmt.Println()
	fmt.Printf("**Total Profiles:** %d\n\n", result.Summary.TotalProfiles)

	// Summary table
	fmt.Println("## Summary by Label")
	fmt.Println()
	fmt.Println("| Label | Best Workers | Min Wall Clock | Max Speedup | Peak Efficiency |")
	fmt.Println("|-------|--------------|----------------|-------------|-----------------|")

	for _, label := range result.Summary.Labels {
		fmt.Printf("| %s | %d | %.1f ms | %.2fx | %.1f%% |\n",
			label,
			result.Summary.BestWorkers[label],
			result.Summary.MinWallClock[label],
			result.Summary.MaxSpeedup[label],
			result.Summary.PeakEfficiency[label],
		)
	}
	fmt.Println()

	// Detailed tables
	for _, label := range result.Summary.Labels {
		points := result.Series[label]
		if len(points) == 0 {
			continue
		}

		fmt.Printf("## %s Data\n\n", label)
		fmt.Println("| Workers | Wall Clock | Total Work | Speedup | Efficiency | Crypto Time |")
		fmt.Println("|---------|------------|------------|---------|------------|-------------|")

		for _, p := range points {
			fmt.Printf("| %d | %.1f ms | %.1f ms | %.2fx | %.1f%% | %.1f ms |\n",
				p.WorkerCount,
				p.WallClockMs,
				p.TotalWorkMs,
				p.Speedup,
				p.Efficiency,
				p.CryptoTimeMs,
			)
		}
		fmt.Println()
	}

	return nil
}
