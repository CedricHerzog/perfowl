package analyzer

import (
	"fmt"
	"strings"

	"github.com/CedricHerzog/perfowl/internal/parser"
)

// ScalingAnalysis contains parallel scaling analysis results
type ScalingAnalysis struct {
	WorkerCount        int      `json:"worker_count"`
	TotalWorkMs        float64  `json:"total_work_ms"`
	WallClockMs        float64  `json:"wall_clock_ms"`
	TheoreticalSpeedup float64  `json:"theoretical_speedup"`
	ActualSpeedup      float64  `json:"actual_speedup"`
	Efficiency         float64  `json:"parallel_efficiency_percent"`
	BottleneckType     string   `json:"bottleneck_type,omitempty"`
	Recommendations    []string `json:"recommendations,omitempty"`
}

// ScalingComparison compares scaling between two profiles
type ScalingComparison struct {
	Baseline    ScalingAnalysis `json:"baseline"`
	Comparison  ScalingAnalysis `json:"comparison"`
	Improvement float64         `json:"improvement_percent"`
	Analysis    string          `json:"analysis"`
}

// AnalyzeScaling performs parallel scaling analysis
func AnalyzeScaling(profile *parser.Profile) ScalingAnalysis {
	analysis := ScalingAnalysis{
		Recommendations: make([]string, 0),
	}

	analysis.WallClockMs = profile.DurationSeconds() * 1000

	// Count worker threads and calculate total work
	interval := profile.Meta.Interval
	workerCPUTime := 0.0
	mainThreadCPUTime := 0.0

	for _, thread := range profile.Threads {
		threadCPU := 0.0
		for i := 0; i < thread.Samples.Length; i++ {
			cpuDelta := interval
			if i < len(thread.Samples.ThreadCPUDelta) && thread.Samples.ThreadCPUDelta[i] > 0 {
				cpuDelta = float64(thread.Samples.ThreadCPUDelta[i]) / 1000.0
			}
			threadCPU += cpuDelta
		}

		if isWorkerThread(&thread) {
			analysis.WorkerCount++
			workerCPUTime += threadCPU
		} else if thread.IsMainThread {
			mainThreadCPUTime = threadCPU
		}
	}

	// Total work is all CPU time spent
	analysis.TotalWorkMs = workerCPUTime + mainThreadCPUTime

	// Calculate theoretical speedup (Amdahl's Law estimate)
	// Assuming worker work could be parallelized
	if analysis.WallClockMs > 0 && analysis.WorkerCount > 0 {
		// Theoretical: if all work was perfectly parallel
		analysis.TheoreticalSpeedup = analysis.TotalWorkMs / analysis.WallClockMs

		// Actual speedup estimation
		// If we had 1 worker, the wall clock would be approximately TotalWorkMs
		// Current speedup = TotalWorkMs / WallClockMs
		analysis.ActualSpeedup = analysis.TotalWorkMs / analysis.WallClockMs

		// Parallel efficiency = actual speedup / theoretical max (worker count)
		if analysis.WorkerCount > 0 {
			analysis.Efficiency = (analysis.ActualSpeedup / float64(analysis.WorkerCount)) * 100
			if analysis.Efficiency > 100 {
				// Can happen due to measurement artifacts, cap at 100
				analysis.Efficiency = 100
			}
		}
	}

	// Determine bottleneck type based on patterns
	if analysis.Efficiency < 30 {
		// Very low efficiency - likely serialization
		analysis.BottleneckType = "serialization"
		analysis.Recommendations = append(analysis.Recommendations,
			"Very low parallel efficiency - work may be serialized on main thread")
	} else if analysis.Efficiency < 50 {
		// Medium efficiency - contention likely
		analysis.BottleneckType = "contention"
		analysis.Recommendations = append(analysis.Recommendations,
			"Medium parallel efficiency - possible contention or synchronization overhead")
	} else if analysis.Efficiency < 70 {
		// Decent efficiency - some overhead
		analysis.BottleneckType = "overhead"
		analysis.Recommendations = append(analysis.Recommendations,
			"Good parallel efficiency with some overhead - consider reducing sync points")
	} else if analysis.Efficiency < 90 {
		// Good efficiency
		analysis.BottleneckType = "minimal"
		analysis.Recommendations = append(analysis.Recommendations,
			"Good parallel efficiency - minor optimizations possible")
	} else {
		analysis.BottleneckType = "none"
		analysis.Recommendations = append(analysis.Recommendations,
			"Excellent parallel efficiency")
	}

	// Additional recommendations based on worker count
	if analysis.WorkerCount == 0 {
		analysis.Recommendations = append(analysis.Recommendations,
			"No worker threads detected - consider using Web Workers for parallel processing")
	} else if analysis.WorkerCount == 1 && analysis.TotalWorkMs > 100 {
		analysis.Recommendations = append(analysis.Recommendations,
			"Only 1 worker detected - additional workers may improve throughput")
	} else if analysis.WorkerCount > 4 && analysis.Efficiency < 50 {
		analysis.Recommendations = append(analysis.Recommendations,
			fmt.Sprintf("%d workers with low efficiency - consider reducing worker count or improving work distribution", analysis.WorkerCount))
	}

	return analysis
}

// CompareScaling compares scaling between two profiles
func CompareScaling(baseline, comparison *parser.Profile) ScalingComparison {
	result := ScalingComparison{
		Baseline:   AnalyzeScaling(baseline),
		Comparison: AnalyzeScaling(comparison),
	}

	// Calculate improvement in efficiency
	if result.Baseline.Efficiency > 0 {
		result.Improvement = ((result.Comparison.Efficiency - result.Baseline.Efficiency) / result.Baseline.Efficiency) * 100
	}

	// Generate analysis text
	var analysis strings.Builder

	// Worker count change
	if result.Comparison.WorkerCount != result.Baseline.WorkerCount {
		analysis.WriteString(fmt.Sprintf("Worker count changed from %d to %d. ",
			result.Baseline.WorkerCount, result.Comparison.WorkerCount))
	}

	// Efficiency change
	effDiff := result.Comparison.Efficiency - result.Baseline.Efficiency
	if effDiff > 5 {
		analysis.WriteString(fmt.Sprintf("Parallel efficiency improved by %.1f%%. ", effDiff))
	} else if effDiff < -5 {
		analysis.WriteString(fmt.Sprintf("Parallel efficiency decreased by %.1f%%. ", -effDiff))
	} else {
		analysis.WriteString("Parallel efficiency remained stable. ")
	}

	// Wall clock change
	wallDiff := result.Comparison.WallClockMs - result.Baseline.WallClockMs
	wallPercent := 0.0
	if result.Baseline.WallClockMs > 0 {
		wallPercent = (wallDiff / result.Baseline.WallClockMs) * 100
	}

	if wallPercent < -5 {
		analysis.WriteString(fmt.Sprintf("Wall clock time improved by %.1f%% (%.1fms faster). ",
			-wallPercent, -wallDiff))
	} else if wallPercent > 5 {
		analysis.WriteString(fmt.Sprintf("Wall clock time regressed by %.1f%% (%.1fms slower). ",
			wallPercent, wallDiff))
	}

	// Bottleneck change
	if result.Comparison.BottleneckType != result.Baseline.BottleneckType {
		analysis.WriteString(fmt.Sprintf("Bottleneck changed from '%s' to '%s'. ",
			result.Baseline.BottleneckType, result.Comparison.BottleneckType))
	}

	result.Analysis = analysis.String()
	if result.Analysis == "" {
		result.Analysis = "No significant changes detected between profiles."
	}

	return result
}

// FormatScalingAnalysis returns a human-readable summary
func FormatScalingAnalysis(analysis ScalingAnalysis) string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("Scaling Analysis (%d workers)\n", analysis.WorkerCount))
	sb.WriteString(strings.Repeat("=", 60) + "\n\n")

	sb.WriteString(fmt.Sprintf("Wall Clock Time:       %.2fms\n", analysis.WallClockMs))
	sb.WriteString(fmt.Sprintf("Total CPU Work:        %.2fms\n", analysis.TotalWorkMs))
	sb.WriteString(fmt.Sprintf("Theoretical Speedup:   %.2fx\n", analysis.TheoreticalSpeedup))
	sb.WriteString(fmt.Sprintf("Actual Speedup:        %.2fx\n", analysis.ActualSpeedup))
	sb.WriteString(fmt.Sprintf("Parallel Efficiency:   %.1f%%\n", analysis.Efficiency))
	sb.WriteString(fmt.Sprintf("Bottleneck Type:       %s\n\n", analysis.BottleneckType))

	if len(analysis.Recommendations) > 0 {
		sb.WriteString("Recommendations:\n")
		for _, r := range analysis.Recommendations {
			sb.WriteString(fmt.Sprintf("  - %s\n", r))
		}
	}

	return sb.String()
}

// FormatScalingComparison returns a human-readable comparison
func FormatScalingComparison(comparison ScalingComparison) string {
	var sb strings.Builder

	sb.WriteString("Scaling Comparison\n")
	sb.WriteString(strings.Repeat("=", 60) + "\n\n")

	sb.WriteString("                        Baseline    Comparison    Change\n")
	sb.WriteString(strings.Repeat("-", 60) + "\n")

	sb.WriteString(fmt.Sprintf("Worker Count:           %8d    %10d    %+d\n",
		comparison.Baseline.WorkerCount,
		comparison.Comparison.WorkerCount,
		comparison.Comparison.WorkerCount-comparison.Baseline.WorkerCount))

	sb.WriteString(fmt.Sprintf("Wall Clock (ms):        %8.1f    %10.1f    %+.1f\n",
		comparison.Baseline.WallClockMs,
		comparison.Comparison.WallClockMs,
		comparison.Comparison.WallClockMs-comparison.Baseline.WallClockMs))

	sb.WriteString(fmt.Sprintf("Total Work (ms):        %8.1f    %10.1f    %+.1f\n",
		comparison.Baseline.TotalWorkMs,
		comparison.Comparison.TotalWorkMs,
		comparison.Comparison.TotalWorkMs-comparison.Baseline.TotalWorkMs))

	sb.WriteString(fmt.Sprintf("Efficiency (%%):         %8.1f    %10.1f    %+.1f\n",
		comparison.Baseline.Efficiency,
		comparison.Comparison.Efficiency,
		comparison.Comparison.Efficiency-comparison.Baseline.Efficiency))

	sb.WriteString(fmt.Sprintf("\nBottleneck:             %-10s  %-10s\n",
		comparison.Baseline.BottleneckType,
		comparison.Comparison.BottleneckType))

	sb.WriteString(fmt.Sprintf("\nOverall Improvement: %+.1f%%\n\n", comparison.Improvement))

	sb.WriteString("Analysis:\n")
	sb.WriteString(fmt.Sprintf("  %s\n", comparison.Analysis))

	return sb.String()
}
