package analyzer

import (
	"fmt"
	"math"

	"github.com/CedricHerzog/perfowl/internal/parser"
)

// ProfileDiff contains comparison between two profiles
type ProfileDiff struct {
	Baseline    ProfileSummary   `json:"baseline"`
	Comparison  ProfileSummary   `json:"comparison"`
	Changes     DiffChanges      `json:"changes"`
	Improved    []string         `json:"improved"`
	Regressed   []string         `json:"regressed"`
	Unchanged   []string         `json:"unchanged"`
}

// ProfileSummary contains key metrics for comparison
type ProfileSummary struct {
	Name           string  `json:"name"`
	DurationMs     float64 `json:"duration_ms"`
	TotalSamples   int     `json:"total_samples"`
	ThreadCount    int     `json:"thread_count"`
	GCMajorCount   int     `json:"gc_major_count"`
	GCMinorCount   int     `json:"gc_minor_count"`
	GCTotalTimeMs  float64 `json:"gc_total_time_ms"`
	SyncIPCCount   int     `json:"sync_ipc_count"`
	LongTaskCount  int     `json:"long_task_count"`
	LayoutCount    int     `json:"layout_count"`
	ExtensionCount int     `json:"extension_count"`
}

// DiffChanges contains the delta between profiles
type DiffChanges struct {
	DurationChangeMs      float64 `json:"duration_change_ms"`
	DurationChangePercent float64 `json:"duration_change_percent"`
	SampleCountChange     int     `json:"sample_count_change"`
	ThreadCountChange     int     `json:"thread_count_change"`
	GCMajorChange         int     `json:"gc_major_change"`
	GCMinorChange         int     `json:"gc_minor_change"`
	GCTimeChangeMs        float64 `json:"gc_time_change_ms"`
	GCTimeChangePercent   float64 `json:"gc_time_change_percent"`
	SyncIPCChange         int     `json:"sync_ipc_change"`
	LongTaskChange        int     `json:"long_task_change"`
	LayoutChange          int     `json:"layout_change"`
}

// CompareProfiles compares two profiles and returns differences
func CompareProfiles(baseline, comparison *parser.Profile) ProfileDiff {
	baseSummary := extractSummary(baseline, "baseline")
	compSummary := extractSummary(comparison, "comparison")

	diff := ProfileDiff{
		Baseline:   baseSummary,
		Comparison: compSummary,
		Improved:   make([]string, 0),
		Regressed:  make([]string, 0),
		Unchanged:  make([]string, 0),
	}

	// Calculate changes
	diff.Changes = DiffChanges{
		DurationChangeMs:      compSummary.DurationMs - baseSummary.DurationMs,
		DurationChangePercent: percentChange(baseSummary.DurationMs, compSummary.DurationMs),
		SampleCountChange:     compSummary.TotalSamples - baseSummary.TotalSamples,
		ThreadCountChange:     compSummary.ThreadCount - baseSummary.ThreadCount,
		GCMajorChange:         compSummary.GCMajorCount - baseSummary.GCMajorCount,
		GCMinorChange:         compSummary.GCMinorCount - baseSummary.GCMinorCount,
		GCTimeChangeMs:        compSummary.GCTotalTimeMs - baseSummary.GCTotalTimeMs,
		GCTimeChangePercent:   percentChange(baseSummary.GCTotalTimeMs, compSummary.GCTotalTimeMs),
		SyncIPCChange:         compSummary.SyncIPCCount - baseSummary.SyncIPCCount,
		LongTaskChange:        compSummary.LongTaskCount - baseSummary.LongTaskCount,
		LayoutChange:          compSummary.LayoutCount - baseSummary.LayoutCount,
	}

	// Determine improvements and regressions
	// Lower is better for these metrics
	threshold := 5.0 // 5% change threshold

	// Duration
	durChange := diff.Changes.DurationChangePercent
	if durChange < -threshold {
		diff.Improved = append(diff.Improved, "Duration reduced by "+formatPercent(-durChange))
	} else if durChange > threshold {
		diff.Regressed = append(diff.Regressed, "Duration increased by "+formatPercent(durChange))
	} else {
		diff.Unchanged = append(diff.Unchanged, "Duration similar")
	}

	// GC Time
	gcChange := diff.Changes.GCTimeChangePercent
	if gcChange < -threshold {
		diff.Improved = append(diff.Improved, "GC time reduced by "+formatPercent(-gcChange))
	} else if gcChange > threshold {
		diff.Regressed = append(diff.Regressed, "GC time increased by "+formatPercent(gcChange))
	} else {
		diff.Unchanged = append(diff.Unchanged, "GC time similar")
	}

	// GC Major count
	if diff.Changes.GCMajorChange < -2 {
		diff.Improved = append(diff.Improved, "Fewer major GC events")
	} else if diff.Changes.GCMajorChange > 2 {
		diff.Regressed = append(diff.Regressed, "More major GC events")
	}

	// Sync IPC
	if diff.Changes.SyncIPCChange < -2 {
		diff.Improved = append(diff.Improved, "Fewer sync IPC calls")
	} else if diff.Changes.SyncIPCChange > 2 {
		diff.Regressed = append(diff.Regressed, "More sync IPC calls")
	}

	// Long tasks
	if diff.Changes.LongTaskChange < -2 {
		diff.Improved = append(diff.Improved, "Fewer long tasks")
	} else if diff.Changes.LongTaskChange > 2 {
		diff.Regressed = append(diff.Regressed, "More long tasks")
	}

	// Layout
	if diff.Changes.LayoutChange < -5 {
		diff.Improved = append(diff.Improved, "Fewer layout operations")
	} else if diff.Changes.LayoutChange > 5 {
		diff.Regressed = append(diff.Regressed, "More layout operations")
	}

	return diff
}

func extractSummary(profile *parser.Profile, name string) ProfileSummary {
	summary := ProfileSummary{
		Name:        name,
		ThreadCount: len(profile.Threads),
	}

	// Calculate duration from meta
	if profile.Meta.ProfilingEndTime > profile.Meta.ProfilingStartTime {
		summary.DurationMs = profile.Meta.ProfilingEndTime - profile.Meta.ProfilingStartTime
	}

	// Count extensions
	summary.ExtensionCount = len(profile.Meta.Extensions.BaseURL)

	// Analyze markers across all threads
	for _, thread := range profile.Threads {
		summary.TotalSamples += thread.Samples.Length

		markers := parser.ExtractMarkers(&thread, profile.Meta.Categories)
		for _, m := range markers {
			switch m.Type {
			case "GCMajor":
				summary.GCMajorCount++
				summary.GCTotalTimeMs += m.Duration
			case "GCMinor":
				summary.GCMinorCount++
				summary.GCTotalTimeMs += m.Duration
			case "IPC":
				if m.Data != nil {
					if sync, ok := m.Data["sync"].(bool); ok && sync {
						summary.SyncIPCCount++
					}
				}
			case "Reflow", "ForceReflow":
				summary.LayoutCount++
			}

			// Long task detection (>50ms)
			if m.Duration > 50 {
				if m.Category == "JavaScript" || m.Type == "eventProcessing" {
					summary.LongTaskCount++
				}
			}
		}
	}

	return summary
}

func percentChange(baseline, comparison float64) float64 {
	if baseline == 0 {
		if comparison == 0 {
			return 0
		}
		return 100 // Infinite increase represented as 100%
	}
	return ((comparison - baseline) / baseline) * 100
}

func formatPercent(p float64) string {
	return fmt.Sprintf("%.1f%%", math.Abs(p))
}
