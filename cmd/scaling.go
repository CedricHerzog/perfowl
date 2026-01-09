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
	compareProfile string
)

var scalingCmd = &cobra.Command{
	Use:   "scaling",
	Short: "Analyze parallel scaling efficiency",
	Long: `Analyzes parallel scaling efficiency including:
- Worker thread utilization
- Theoretical vs actual speedup
- Parallel efficiency percentage
- Bottleneck type identification

Use --compare to compare scaling between two profiles:
  perfowl scaling -p baseline.json.gz --compare optimized.json.gz`,
	RunE: runScaling,
}

func init() {
	rootCmd.AddCommand(scalingCmd)
	scalingCmd.Flags().StringVar(&compareProfile, "compare", "", "Compare with another profile")
}

func runScaling(cmd *cobra.Command, args []string) error {
	if profilePath == "" {
		return fmt.Errorf("profile path is required (use --profile or -p)")
	}

	profile, err := parser.LoadProfile(profilePath)
	if err != nil {
		return fmt.Errorf("failed to load profile: %w", err)
	}

	// Check if we're doing a comparison
	if compareProfile != "" {
		compProfile, err := parser.LoadProfile(compareProfile)
		if err != nil {
			return fmt.Errorf("failed to load comparison profile: %w", err)
		}

		comparison := analyzer.CompareScaling(profile, compProfile)

		switch outputFormat {
		case "json":
			return outputScalingComparisonJSON(comparison)
		case "markdown":
			return outputScalingComparisonMarkdown(comparison)
		default:
			return outputScalingComparisonText(comparison)
		}
	}

	analysis := analyzer.AnalyzeScaling(profile)

	switch outputFormat {
	case "json":
		return outputScalingJSON(analysis)
	case "markdown":
		return outputScalingMarkdown(analysis)
	default:
		return outputScalingText(analysis)
	}
}

func outputScalingJSON(analysis analyzer.ScalingAnalysis) error {
	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	return encoder.Encode(analysis)
}

func outputScalingComparisonJSON(comparison analyzer.ScalingComparison) error {
	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	return encoder.Encode(comparison)
}

func outputScalingMarkdown(analysis analyzer.ScalingAnalysis) error {
	md := strings.Builder{}

	md.WriteString("# Scaling Analysis\n\n")

	md.WriteString("## Summary\n\n")
	md.WriteString(fmt.Sprintf("- **Worker Count**: %d\n", analysis.WorkerCount))
	md.WriteString(fmt.Sprintf("- **Wall Clock Time**: %.2f ms\n", analysis.WallClockMs))
	md.WriteString(fmt.Sprintf("- **Total CPU Work**: %.2f ms\n", analysis.TotalWorkMs))

	md.WriteString("\n## Speedup Analysis\n\n")
	md.WriteString(fmt.Sprintf("- **Theoretical Speedup**: %.2fx\n", analysis.TheoreticalSpeedup))
	md.WriteString(fmt.Sprintf("- **Actual Speedup**: %.2fx\n", analysis.ActualSpeedup))
	md.WriteString(fmt.Sprintf("- **Parallel Efficiency**: %.1f%%\n", analysis.Efficiency))

	// Efficiency gauge
	md.WriteString("\n### Efficiency Rating\n\n")
	effEmoji := "🔴"
	if analysis.Efficiency >= 90 {
		effEmoji = "🟢"
	} else if analysis.Efficiency >= 70 {
		effEmoji = "🟡"
	} else if analysis.Efficiency >= 50 {
		effEmoji = "🟠"
	}
	md.WriteString(fmt.Sprintf("%s **%.1f%%** parallel efficiency\n", effEmoji, analysis.Efficiency))

	if analysis.BottleneckType != "" {
		md.WriteString(fmt.Sprintf("\n**Bottleneck Type**: %s\n", analysis.BottleneckType))
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

func outputScalingComparisonMarkdown(comparison analyzer.ScalingComparison) error {
	md := strings.Builder{}

	md.WriteString("# Scaling Comparison\n\n")

	// Improvement summary
	improvementEmoji := "➡️"
	if comparison.Improvement > 5 {
		improvementEmoji = "📈"
	} else if comparison.Improvement < -5 {
		improvementEmoji = "📉"
	}
	md.WriteString(fmt.Sprintf("## Overall: %s %+.1f%% improvement\n\n", improvementEmoji, comparison.Improvement))

	md.WriteString("## Comparison Table\n\n")
	md.WriteString("| Metric | Baseline | Comparison | Change |\n")
	md.WriteString("|--------|----------|------------|--------|\n")

	md.WriteString(fmt.Sprintf("| Worker Count | %d | %d | %+d |\n",
		comparison.Baseline.WorkerCount,
		comparison.Comparison.WorkerCount,
		comparison.Comparison.WorkerCount-comparison.Baseline.WorkerCount))

	md.WriteString(fmt.Sprintf("| Wall Clock | %.1fms | %.1fms | %+.1fms |\n",
		comparison.Baseline.WallClockMs,
		comparison.Comparison.WallClockMs,
		comparison.Comparison.WallClockMs-comparison.Baseline.WallClockMs))

	md.WriteString(fmt.Sprintf("| Total Work | %.1fms | %.1fms | %+.1fms |\n",
		comparison.Baseline.TotalWorkMs,
		comparison.Comparison.TotalWorkMs,
		comparison.Comparison.TotalWorkMs-comparison.Baseline.TotalWorkMs))

	md.WriteString(fmt.Sprintf("| Efficiency | %.1f%% | %.1f%% | %+.1f%% |\n",
		comparison.Baseline.Efficiency,
		comparison.Comparison.Efficiency,
		comparison.Comparison.Efficiency-comparison.Baseline.Efficiency))

	md.WriteString(fmt.Sprintf("| Bottleneck | %s | %s | - |\n",
		comparison.Baseline.BottleneckType,
		comparison.Comparison.BottleneckType))

	md.WriteString("\n## Analysis\n\n")
	md.WriteString(comparison.Analysis + "\n")

	fmt.Print(md.String())
	return nil
}

func outputScalingText(analysis analyzer.ScalingAnalysis) error {
	fmt.Println("Scaling Analysis")
	fmt.Println(strings.Repeat("=", 60))
	fmt.Println()

	fmt.Println("Summary:")
	fmt.Printf("  Worker Count:        %d\n", analysis.WorkerCount)
	fmt.Printf("  Wall Clock Time:     %.2f ms\n", analysis.WallClockMs)
	fmt.Printf("  Total CPU Work:      %.2f ms\n", analysis.TotalWorkMs)
	fmt.Println()

	fmt.Println("Speedup Analysis:")
	fmt.Printf("  Theoretical Speedup: %.2fx\n", analysis.TheoreticalSpeedup)
	fmt.Printf("  Actual Speedup:      %.2fx\n", analysis.ActualSpeedup)
	fmt.Printf("  Parallel Efficiency: %.1f%%\n", analysis.Efficiency)
	fmt.Println()

	if analysis.BottleneckType != "" {
		fmt.Printf("Bottleneck Type: %s\n\n", analysis.BottleneckType)
	}

	if len(analysis.Recommendations) > 0 {
		fmt.Println("Recommendations:")
		for _, r := range analysis.Recommendations {
			fmt.Printf("  - %s\n", r)
		}
	}

	return nil
}

func outputScalingComparisonText(comparison analyzer.ScalingComparison) error {
	fmt.Println("Scaling Comparison")
	fmt.Println(strings.Repeat("=", 60))
	fmt.Println()

	fmt.Printf("Overall Improvement: %+.1f%%\n\n", comparison.Improvement)

	fmt.Println("                        Baseline    Comparison    Change")
	fmt.Println(strings.Repeat("-", 60))

	fmt.Printf("Worker Count:           %8d    %10d    %+d\n",
		comparison.Baseline.WorkerCount,
		comparison.Comparison.WorkerCount,
		comparison.Comparison.WorkerCount-comparison.Baseline.WorkerCount)

	fmt.Printf("Wall Clock (ms):        %8.1f    %10.1f    %+.1f\n",
		comparison.Baseline.WallClockMs,
		comparison.Comparison.WallClockMs,
		comparison.Comparison.WallClockMs-comparison.Baseline.WallClockMs)

	fmt.Printf("Total Work (ms):        %8.1f    %10.1f    %+.1f\n",
		comparison.Baseline.TotalWorkMs,
		comparison.Comparison.TotalWorkMs,
		comparison.Comparison.TotalWorkMs-comparison.Baseline.TotalWorkMs)

	fmt.Printf("Efficiency (%%):         %8.1f    %10.1f    %+.1f\n",
		comparison.Baseline.Efficiency,
		comparison.Comparison.Efficiency,
		comparison.Comparison.Efficiency-comparison.Baseline.Efficiency)

	fmt.Printf("\nBottleneck:             %-10s  %-10s\n",
		comparison.Baseline.BottleneckType,
		comparison.Comparison.BottleneckType)

	fmt.Println()
	fmt.Println("Analysis:")
	fmt.Printf("  %s\n", comparison.Analysis)

	return nil
}
