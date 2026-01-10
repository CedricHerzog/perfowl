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

var jscryptoCmd = &cobra.Command{
	Use:   "jscrypto",
	Short: "Analyze JavaScript-level crypto operations",
	Long: `Analyzes JavaScript crypto workers and decryption functions including:
- Crypto worker files (e.g., seipdDecryptionWorker.js, openpgp.js)
- Time spent in each crypto resource/worker
- Top functions by CPU time
- Per-worker time distribution

This complements the 'crypto' command which focuses on native crypto libraries.
Use this to understand application-level crypto performance.`,
	RunE: runJSCrypto,
}

func init() {
	rootCmd.AddCommand(jscryptoCmd)
}

func runJSCrypto(cmd *cobra.Command, args []string) error {
	if profilePath == "" {
		return fmt.Errorf("profile path is required (use --profile or -p)")
	}

	bt := parser.ParseBrowserType(browserType)
	profile, _, err := parser.LoadProfileWithType(profilePath, bt)
	if err != nil {
		return fmt.Errorf("failed to load profile: %w", err)
	}

	analysis := analyzer.AnalyzeJSCrypto(profile)

	switch outputFormat {
	case "json":
		return outputJSCryptoJSON(analysis)
	case "markdown":
		return outputJSCryptoMarkdown(analysis)
	default:
		return outputJSCryptoText(analysis)
	}
}

func outputJSCryptoJSON(analysis analyzer.JSCryptoAnalysis) error {
	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	return encoder.Encode(analysis)
}

func outputJSCryptoMarkdown(analysis analyzer.JSCryptoAnalysis) error {
	md := strings.Builder{}

	md.WriteString("# JavaScript Crypto Analysis\n\n")

	md.WriteString("## Summary\n\n")
	md.WriteString(fmt.Sprintf("- **Total Time**: %.2f ms\n", analysis.TotalTimeMs))
	md.WriteString(fmt.Sprintf("- **Total Samples**: %d\n", analysis.TotalSamples))
	md.WriteString(fmt.Sprintf("- **Worker Count**: %d\n", analysis.WorkerCount))
	if analysis.WorkerCount > 0 {
		md.WriteString(fmt.Sprintf("- **Avg Time/Worker**: %.2f ms\n", analysis.AvgTimePerWorker))
	}

	if len(analysis.ByThread) > 0 {
		md.WriteString("\n## Time by Thread\n\n")
		md.WriteString("| Thread | Time |\n")
		md.WriteString("|--------|------|\n")

		for thread, timeMs := range analysis.ByThread {
			name := thread
			if len(name) > 30 {
				name = name[:27] + "..."
			}
			md.WriteString(fmt.Sprintf("| %s | %.2fms |\n", name, timeMs))
		}
	}

	if len(analysis.Resources) > 0 {
		md.WriteString("\n## Crypto Resources\n\n")
		md.WriteString("| Resource | Thread | Time | Samples |\n")
		md.WriteString("|----------|--------|------|--------|\n")

		displayCount := len(analysis.Resources)
		if displayCount > 15 {
			displayCount = 15
		}

		for i := 0; i < displayCount; i++ {
			res := analysis.Resources[i]
			name := res.Name
			if len(name) > 30 {
				name = name[:27] + "..."
			}
			threadName := res.ThreadName
			if len(threadName) > 15 {
				threadName = threadName[:12] + "..."
			}
			md.WriteString(fmt.Sprintf("| %s | %s | %.2fms | %d |\n",
				name, threadName, res.TotalTime, res.SampleCount))
		}

		if len(analysis.Resources) > 15 {
			md.WriteString(fmt.Sprintf("\n... and %d more resources\n", len(analysis.Resources)-15))
		}
	}

	if len(analysis.TopFunctions) > 0 {
		md.WriteString("\n## Top Functions\n\n")
		md.WriteString("| Function | Resource | Time | % |\n")
		md.WriteString("|----------|----------|------|---|\n")

		displayCount := len(analysis.TopFunctions)
		if displayCount > 20 {
			displayCount = 20
		}

		for i := 0; i < displayCount; i++ {
			fn := analysis.TopFunctions[i]
			name := fn.Name
			if len(name) > 35 {
				name = name[:32] + "..."
			}
			res := fn.Resource
			if len(res) > 20 {
				res = res[:17] + "..."
			}
			md.WriteString(fmt.Sprintf("| %s | %s | %.2fms | %.1f%% |\n",
				name, res, fn.TotalTime, fn.Percent))
		}

		if len(analysis.TopFunctions) > 20 {
			md.WriteString(fmt.Sprintf("\n... and %d more functions\n", len(analysis.TopFunctions)-20))
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

func outputJSCryptoText(analysis analyzer.JSCryptoAnalysis) error {
	fmt.Println("JavaScript Crypto Analysis")
	fmt.Println(strings.Repeat("=", 60))
	fmt.Println()

	fmt.Println("Summary:")
	fmt.Printf("  Total Time:       %.2f ms\n", analysis.TotalTimeMs)
	fmt.Printf("  Total Samples:    %d\n", analysis.TotalSamples)
	fmt.Printf("  Worker Count:     %d\n", analysis.WorkerCount)
	if analysis.WorkerCount > 0 {
		fmt.Printf("  Avg Time/Worker:  %.2f ms\n", analysis.AvgTimePerWorker)
	}
	fmt.Println()

	if len(analysis.ByThread) > 0 {
		fmt.Println("Time by Thread:")
		fmt.Println(strings.Repeat("-", 50))
		for thread, timeMs := range analysis.ByThread {
			name := thread
			if len(name) > 30 {
				name = name[:27] + "..."
			}
			fmt.Printf("  %-30s %.2fms\n", name, timeMs)
		}
		fmt.Println()
	}

	if len(analysis.Resources) > 0 {
		fmt.Println("Crypto Resources:")
		fmt.Println(strings.Repeat("-", 60))
		fmt.Printf("%-35s %10s %8s\n", "Resource", "Time", "Samples")
		fmt.Println(strings.Repeat("-", 60))

		displayCount := len(analysis.Resources)
		if displayCount > 10 {
			displayCount = 10
		}

		for i := 0; i < displayCount; i++ {
			res := analysis.Resources[i]
			name := res.Name
			if len(name) > 35 {
				name = name[:32] + "..."
			}
			fmt.Printf("%-35s %8.2fms %8d\n", name, res.TotalTime, res.SampleCount)
		}

		if len(analysis.Resources) > 10 {
			fmt.Printf("  ... and %d more resources\n", len(analysis.Resources)-10)
		}
		fmt.Println()
	}

	if len(analysis.TopFunctions) > 0 {
		fmt.Println("Top Functions:")
		fmt.Println(strings.Repeat("-", 70))
		fmt.Printf("%-45s %10s %6s\n", "Function", "Time", "%")
		fmt.Println(strings.Repeat("-", 70))

		displayCount := len(analysis.TopFunctions)
		if displayCount > 15 {
			displayCount = 15
		}

		for i := 0; i < displayCount; i++ {
			fn := analysis.TopFunctions[i]
			name := fn.Name
			if len(name) > 45 {
				name = name[:42] + "..."
			}
			fmt.Printf("%-45s %8.2fms %5.1f%%\n", name, fn.TotalTime, fn.Percent)
		}

		if len(analysis.TopFunctions) > 15 {
			fmt.Printf("  ... and %d more functions\n", len(analysis.TopFunctions)-15)
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
