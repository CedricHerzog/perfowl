package analyzer

import (
	"sort"

	"github.com/CedricHerzog/perfowl/internal/parser"
)

// ThreadStats contains analysis for a single thread
type ThreadStats struct {
	Name          string  `json:"name"`
	ProcessType   string  `json:"process_type"`
	ProcessName   string  `json:"process_name"`
	PID           string  `json:"pid"`
	TID           string  `json:"tid"`
	IsMainThread  bool    `json:"is_main_thread"`
	CPUTimeMs     float64 `json:"cpu_time_ms"`
	SampleCount   int     `json:"sample_count"`
	MarkerCount   int     `json:"marker_count"`
	WakeCount     int     `json:"wake_count"`
	AvgWakeIntervalMs float64 `json:"avg_wake_interval_ms,omitempty"`
	TopCategories []CategoryStats `json:"top_categories,omitempty"`
}

// ThreadAnalysis contains the full thread analysis
type ThreadAnalysis struct {
	TotalThreads     int           `json:"total_threads"`
	MainThreadCount  int           `json:"main_thread_count"`
	ParentProcessThreads int       `json:"parent_process_threads"`
	ContentProcessThreads int      `json:"content_process_threads"`
	Threads          []ThreadStats `json:"threads"`
}

// AnalyzeThreads performs detailed thread analysis
func AnalyzeThreads(profile *parser.Profile) ThreadAnalysis {
	analysis := ThreadAnalysis{
		TotalThreads: len(profile.Threads),
		Threads:      make([]ThreadStats, 0, len(profile.Threads)),
	}

	// Map category index to name
	categoryNames := make(map[int]string)
	for i, cat := range profile.Meta.Categories {
		categoryNames[i] = cat.Name
	}

	interval := profile.Meta.Interval

	for _, thread := range profile.Threads {
		stats := ThreadStats{
			Name:         thread.Name,
			ProcessType:  thread.ProcessType,
			ProcessName:  thread.ProcessName,
			PID:          thread.PID.String(),
			TID:          thread.TID.String(),
			IsMainThread: thread.IsMainThread,
			SampleCount:  thread.Samples.Length,
			MarkerCount:  thread.Markers.Length,
		}

		// Count process types
		if thread.ProcessType == "default" || thread.ProcessType == "parent" {
			analysis.ParentProcessThreads++
		} else if thread.ProcessType == "tab" || thread.ProcessType == "web" {
			analysis.ContentProcessThreads++
		}

		if thread.IsMainThread {
			analysis.MainThreadCount++
		}

		// Calculate CPU time from samples
		categoryTime := make(map[string]float64)
		for i := 0; i < thread.Samples.Length; i++ {
			cpuDelta := 0.0
			if i < len(thread.Samples.ThreadCPUDelta) && thread.Samples.ThreadCPUDelta[i] > 0 {
				cpuDelta = float64(thread.Samples.ThreadCPUDelta[i]) / 1000.0
			} else {
				cpuDelta = interval
			}
			stats.CPUTimeMs += cpuDelta

			// Track category distribution
			stackIdx := -1
			if i < len(thread.Samples.Stack) {
				stackIdx = thread.Samples.Stack[i]
			}
			if stackIdx >= 0 && stackIdx < len(thread.StackTable.Category) {
				catIdx := thread.StackTable.Category[stackIdx]
				catName := categoryNames[catIdx]
				if catName == "" {
					catName = "Unknown"
				}
				categoryTime[catName] += cpuDelta
			}
		}

		// Count Awake markers for wake analysis
		markers := parser.ExtractMarkers(&thread, profile.Meta.Categories)
		var awakeTimes []float64
		for _, m := range markers {
			if m.Name == "Awake" || m.Type == "Awake" {
				stats.WakeCount++
				awakeTimes = append(awakeTimes, m.StartTime)
			}
		}

		// Calculate average wake interval
		if len(awakeTimes) > 1 {
			sort.Float64s(awakeTimes)
			totalInterval := 0.0
			for i := 1; i < len(awakeTimes); i++ {
				totalInterval += awakeTimes[i] - awakeTimes[i-1]
			}
			stats.AvgWakeIntervalMs = totalInterval / float64(len(awakeTimes)-1)
		}

		// Get top 5 categories
		var cats []CategoryStats
		for name, time := range categoryTime {
			percent := 0.0
			if stats.CPUTimeMs > 0 {
				percent = (time / stats.CPUTimeMs) * 100
			}
			cats = append(cats, CategoryStats{
				Name:    name,
				TimeMs:  time,
				Percent: percent,
			})
		}
		sort.Slice(cats, func(i, j int) bool {
			return cats[i].TimeMs > cats[j].TimeMs
		})
		if len(cats) > 5 {
			cats = cats[:5]
		}
		stats.TopCategories = cats

		analysis.Threads = append(analysis.Threads, stats)
	}

	// Sort threads by CPU time descending
	sort.Slice(analysis.Threads, func(i, j int) bool {
		return analysis.Threads[i].CPUTimeMs > analysis.Threads[j].CPUTimeMs
	})

	return analysis
}
