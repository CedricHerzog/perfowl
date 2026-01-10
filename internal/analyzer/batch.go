package analyzer

import (
	"fmt"
	"sort"

	"github.com/CedricHerzog/perfowl/internal/parser"
)

// ProfileEntry defines a single profile to analyze with its metadata
type ProfileEntry struct {
	Path         string `json:"path" yaml:"path"`
	WorkerCount  int    `json:"workers" yaml:"workers"`
	Label        string `json:"label" yaml:"label"`
	StartPattern string `json:"start_pattern,omitempty" yaml:"start_pattern,omitempty"`
	EndPattern   string `json:"end_pattern,omitempty" yaml:"end_pattern,omitempty"`
}

// ProfileDataPoint represents metrics for a single profile
type ProfileDataPoint struct {
	WorkerCount     int     `json:"worker_count"`
	Label           string  `json:"label"`
	FilePath        string  `json:"file_path"`
	WallClockMs     float64 `json:"wall_clock_ms"`
	OperationTimeMs float64 `json:"operation_time_ms,omitempty"`
	TotalWorkMs     float64 `json:"total_work_ms"`
	Efficiency      float64 `json:"efficiency_percent"`
	Speedup         float64 `json:"speedup"`
	CryptoTimeMs    float64 `json:"crypto_time_ms"`
}

// BatchSummary provides high-level insights across all profiles
type BatchSummary struct {
	TotalProfiles    int                `json:"total_profiles"`
	Labels           []string           `json:"labels"`
	BestWorkers      map[string]int     `json:"best_workers"`       // Label -> optimal worker count
	MinWallClock     map[string]float64 `json:"min_wall_clock"`     // Label -> minimum wall clock time
	MinOperationTime map[string]float64 `json:"min_operation_time"` // Label -> minimum operation time
	MaxSpeedup       map[string]float64 `json:"max_speedup"`        // Label -> maximum speedup achieved
	PeakEfficiency   map[string]float64 `json:"peak_efficiency"`    // Label -> peak efficiency
}

// BatchAnalysisResult contains aggregated results from batch analysis
type BatchAnalysisResult struct {
	Series  map[string][]ProfileDataPoint `json:"series"`
	Summary BatchSummary                  `json:"summary"`
}

// AnalyzeBatch performs batch analysis across multiple profiles
func AnalyzeBatch(profiles []ProfileEntry) (*BatchAnalysisResult, error) {
	result := &BatchAnalysisResult{
		Series: make(map[string][]ProfileDataPoint),
		Summary: BatchSummary{
			TotalProfiles:    len(profiles),
			Labels:           make([]string, 0),
			BestWorkers:      make(map[string]int),
			MinWallClock:     make(map[string]float64),
			MinOperationTime: make(map[string]float64),
			MaxSpeedup:       make(map[string]float64),
			PeakEfficiency:   make(map[string]float64),
		},
	}

	if len(profiles) == 0 {
		return result, nil
	}

	// Track unique labels
	labelSet := make(map[string]bool)

	// Process each profile
	for _, entry := range profiles {
		profile, _, err := parser.LoadProfileAuto(entry.Path)
		if err != nil {
			return nil, fmt.Errorf("failed to load profile %s: %w", entry.Path, err)
		}

		// Run scaling analysis
		scaling := AnalyzeScaling(profile)

		// Run JS crypto analysis
		crypto := AnalyzeJSCrypto(profile)

		// Create data point
		point := ProfileDataPoint{
			WorkerCount:  entry.WorkerCount,
			Label:        entry.Label,
			FilePath:     entry.Path,
			WallClockMs:  scaling.WallClockMs,
			TotalWorkMs:  scaling.TotalWorkMs,
			Efficiency:   scaling.Efficiency,
			Speedup:      scaling.ActualSpeedup,
			CryptoTimeMs: crypto.TotalTimeMs,
		}

		// Measure operation time if patterns are provided (uses find_last for full operation duration)
		if entry.StartPattern != "" && entry.EndPattern != "" {
			measurement, err := MeasureOperationLast(profile, entry.StartPattern, entry.EndPattern, 0, 0)
			if err == nil {
				point.OperationTimeMs = measurement.OperationTimeMs
			}
			// Silently ignore errors - operation time will be 0 if measurement fails
		}

		// Add to series
		if result.Series[entry.Label] == nil {
			result.Series[entry.Label] = make([]ProfileDataPoint, 0)
		}
		result.Series[entry.Label] = append(result.Series[entry.Label], point)
		labelSet[entry.Label] = true
	}

	// Sort each series by worker count
	for label := range result.Series {
		sort.Slice(result.Series[label], func(i, j int) bool {
			return result.Series[label][i].WorkerCount < result.Series[label][j].WorkerCount
		})
	}

	// Build summary
	for label := range labelSet {
		result.Summary.Labels = append(result.Summary.Labels, label)
	}
	sort.Strings(result.Summary.Labels)

	// Calculate best metrics per label
	for label, points := range result.Series {
		if len(points) == 0 {
			continue
		}

		// Find best wall clock time (minimum)
		bestWallClock := points[0].WallClockMs
		bestWorkers := points[0].WorkerCount

		// Find peak efficiency and max speedup
		peakEfficiency := 0.0
		maxSpeedup := 0.0

		// Track min operation time (only if measured)
		minOpTime := 0.0
		hasOpTime := false

		for _, p := range points {
			if p.WallClockMs < bestWallClock {
				bestWallClock = p.WallClockMs
				bestWorkers = p.WorkerCount
			}
			if p.Efficiency > peakEfficiency {
				peakEfficiency = p.Efficiency
			}
			if p.Speedup > maxSpeedup {
				maxSpeedup = p.Speedup
			}
			if p.OperationTimeMs > 0 {
				if !hasOpTime || p.OperationTimeMs < minOpTime {
					minOpTime = p.OperationTimeMs
					hasOpTime = true
				}
			}
		}

		result.Summary.BestWorkers[label] = bestWorkers
		result.Summary.MinWallClock[label] = bestWallClock
		result.Summary.PeakEfficiency[label] = peakEfficiency
		result.Summary.MaxSpeedup[label] = maxSpeedup
		if hasOpTime {
			result.Summary.MinOperationTime[label] = minOpTime
		}
	}

	return result, nil
}
