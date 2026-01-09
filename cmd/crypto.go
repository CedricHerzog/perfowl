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

var cryptoCmd = &cobra.Command{
	Use:   "crypto",
	Short: "Analyze crypto operations in the profile",
	Long: `Analyzes cryptographic operations including:
- SubtleCrypto API usage (encrypt, decrypt, sign, verify, digest)
- Key management operations (generateKey, importKey, deriveKey)
- Algorithm detection (AES, RSA, SHA-*, PBKDF2, etc.)

Identifies potential issues like:
- Serialized crypto operations that could be parallelized
- Usage of weak algorithms (MD5, SHA-1)
- Excessive crypto overhead`,
	RunE: runCrypto,
}

func init() {
	rootCmd.AddCommand(cryptoCmd)
}

func runCrypto(cmd *cobra.Command, args []string) error {
	if profilePath == "" {
		return fmt.Errorf("profile path is required (use --profile or -p)")
	}

	profile, err := parser.LoadProfile(profilePath)
	if err != nil {
		return fmt.Errorf("failed to load profile: %w", err)
	}

	analysis := analyzer.AnalyzeCrypto(profile)

	switch outputFormat {
	case "json":
		return outputCryptoJSON(analysis)
	case "markdown":
		return outputCryptoMarkdown(analysis)
	default:
		return outputCryptoText(analysis)
	}
}

func outputCryptoJSON(analysis analyzer.CryptoAnalysis) error {
	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	return encoder.Encode(analysis)
}

func outputCryptoMarkdown(analysis analyzer.CryptoAnalysis) error {
	md := strings.Builder{}

	md.WriteString("# Crypto Operation Analysis\n\n")

	md.WriteString("## Summary\n\n")
	md.WriteString(fmt.Sprintf("- **Total Operations**: %d\n", analysis.TotalOperations))
	md.WriteString(fmt.Sprintf("- **Total Time**: %.2f ms\n", analysis.TotalTimeMs))
	if analysis.Serialized {
		md.WriteString("- **Serialization**: ⚠️ Possibly serialized (no parallel crypto detected)\n")
	}

	if len(analysis.ByOperation) > 0 {
		md.WriteString("\n## Time by Operation\n\n")
		md.WriteString("| Operation | Time |\n")
		md.WriteString("|-----------|------|\n")

		for op, timeMs := range analysis.ByOperation {
			md.WriteString(fmt.Sprintf("| %s | %.2fms |\n", op, timeMs))
		}
	}

	if len(analysis.ByAlgorithm) > 0 {
		md.WriteString("\n## Time by Algorithm\n\n")
		md.WriteString("| Algorithm | Time |\n")
		md.WriteString("|-----------|------|\n")

		for algo, timeMs := range analysis.ByAlgorithm {
			md.WriteString(fmt.Sprintf("| %s | %.2fms |\n", algo, timeMs))
		}
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

	if len(analysis.TopOperations) > 0 {
		md.WriteString("\n## Top Operations by Duration\n\n")
		md.WriteString("| Operation | Algorithm | Duration | Thread |\n")
		md.WriteString("|-----------|-----------|----------|--------|\n")

		displayCount := len(analysis.TopOperations)
		if displayCount > 10 {
			displayCount = 10
		}

		for i := 0; i < displayCount; i++ {
			op := analysis.TopOperations[i]
			algo := op.Algorithm
			if algo == "" {
				algo = "-"
			}
			threadName := op.ThreadName
			if len(threadName) > 20 {
				threadName = threadName[:17] + "..."
			}
			md.WriteString(fmt.Sprintf("| %s | %s | %.2fms | %s |\n",
				op.Operation, algo, op.DurationMs, threadName))
		}

		if len(analysis.TopOperations) > 10 {
			md.WriteString(fmt.Sprintf("\n... and %d more operations\n", len(analysis.TopOperations)-10))
		}
	}

	if len(analysis.Warnings) > 0 {
		md.WriteString("\n## Warnings\n\n")
		for _, w := range analysis.Warnings {
			md.WriteString(fmt.Sprintf("- ⚠️ %s\n", w))
		}
	}

	fmt.Print(md.String())
	return nil
}

func outputCryptoText(analysis analyzer.CryptoAnalysis) error {
	fmt.Println("Crypto Operation Analysis")
	fmt.Println(strings.Repeat("=", 60))
	fmt.Println()

	fmt.Println("Summary:")
	fmt.Printf("  Total Operations: %d\n", analysis.TotalOperations)
	fmt.Printf("  Total Time:       %.2f ms\n", analysis.TotalTimeMs)
	if analysis.Serialized {
		fmt.Println("  Serialization:    Possibly serialized (no parallel crypto)")
	}
	fmt.Println()

	if len(analysis.ByOperation) > 0 {
		fmt.Println("Time by Operation:")
		fmt.Println(strings.Repeat("-", 40))
		for op, timeMs := range analysis.ByOperation {
			fmt.Printf("  %-20s %.2fms\n", op, timeMs)
		}
		fmt.Println()
	}

	if len(analysis.ByAlgorithm) > 0 {
		fmt.Println("Time by Algorithm:")
		fmt.Println(strings.Repeat("-", 40))
		for algo, timeMs := range analysis.ByAlgorithm {
			fmt.Printf("  %-20s %.2fms\n", algo, timeMs)
		}
		fmt.Println()
	}

	if len(analysis.ByThread) > 0 {
		fmt.Println("Time by Thread:")
		fmt.Println(strings.Repeat("-", 40))
		for thread, timeMs := range analysis.ByThread {
			name := thread
			if len(name) > 30 {
				name = name[:27] + "..."
			}
			fmt.Printf("  %-30s %.2fms\n", name, timeMs)
		}
		fmt.Println()
	}

	if len(analysis.TopOperations) > 0 {
		fmt.Println("Top Operations by Duration:")
		fmt.Println(strings.Repeat("-", 60))
		fmt.Printf("%-15s %-12s %10s  %s\n", "Operation", "Algorithm", "Duration", "Thread")
		fmt.Println(strings.Repeat("-", 60))

		displayCount := len(analysis.TopOperations)
		if displayCount > 10 {
			displayCount = 10
		}

		for i := 0; i < displayCount; i++ {
			op := analysis.TopOperations[i]
			algo := op.Algorithm
			if algo == "" {
				algo = "-"
			}
			threadName := op.ThreadName
			if len(threadName) > 20 {
				threadName = threadName[:17] + "..."
			}
			fmt.Printf("%-15s %-12s %8.2fms  %s\n",
				op.Operation, algo, op.DurationMs, threadName)
		}

		if len(analysis.TopOperations) > 10 {
			fmt.Printf("\n  ... and %d more operations\n", len(analysis.TopOperations)-10)
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
