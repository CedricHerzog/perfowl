package analyzer

import (
	"sort"

	"github.com/CedricHerzog/perfowl/internal/parser"
)

// CategoryStats contains timing information for a single category
type CategoryStats struct {
	Name        string  `json:"name"`
	TimeMs      float64 `json:"time_ms"`
	Percent     float64 `json:"percent"`
	SampleCount int     `json:"sample_count"`
}

// CategoryBreakdown contains the full category analysis
type CategoryBreakdown struct {
	TotalTimeMs float64                    `json:"total_time_ms"`
	Categories  []CategoryStats            `json:"categories"`
	ByThread    map[string][]CategoryStats `json:"by_thread,omitempty"`
}

// AnalyzeCategories computes time spent in each profiler category
func AnalyzeCategories(profile *parser.Profile, threadName string) CategoryBreakdown {
	breakdown := CategoryBreakdown{
		Categories: make([]CategoryStats, 0),
		ByThread:   make(map[string][]CategoryStats),
	}

	// Map category index to name
	categoryNames := make(map[int]string)
	for i, cat := range profile.Meta.Categories {
		categoryNames[i] = cat.Name
	}

	// Aggregate sample counts per category per thread
	type threadCategoryData struct {
		sampleCount int
		cpuTime     float64
	}

	threadCategories := make(map[string]map[string]*threadCategoryData)
	globalCategories := make(map[string]*threadCategoryData)

	// Get sampling interval in ms
	interval := profile.Meta.Interval

	for _, thread := range profile.Threads {
		// Filter by thread name if specified
		if threadName != "" && thread.Name != threadName {
			continue
		}

		if threadCategories[thread.Name] == nil {
			threadCategories[thread.Name] = make(map[string]*threadCategoryData)
		}

		// Analyze samples - each sample has a stack with a category
		for i := 0; i < thread.Samples.Length; i++ {
			stackIdx := -1
			if i < len(thread.Samples.Stack) {
				stackIdx = thread.Samples.Stack[i]
			}

			if stackIdx < 0 || stackIdx >= thread.StackTable.Length {
				continue
			}

			// Get category from stack
			// Chrome profiles have category in StackTable.Category (populated by converter)
			// Firefox profiles have category in FrameTable.Category (need to look up via frame)
			catIdx := -1
			if len(thread.StackTable.Category) > 0 && stackIdx < len(thread.StackTable.Category) {
				// Chrome path: category stored directly in StackTable
				catIdx = thread.StackTable.Category[stackIdx]
			} else if stackIdx < len(thread.StackTable.Frame) {
				// Firefox path: get frame index, then look up category from FrameTable
				frameIdx := thread.StackTable.Frame[stackIdx]
				if frameIdx >= 0 && frameIdx < len(thread.FrameTable.Category) {
					catIdx = thread.FrameTable.Category[frameIdx]
				}
			}

			catName := "Unknown"
			if name, ok := categoryNames[catIdx]; ok {
				catName = name
			}

			// Get CPU delta if available
			cpuDelta := 0.0
			if i < len(thread.Samples.ThreadCPUDelta) && thread.Samples.ThreadCPUDelta[i] > 0 {
				cpuDelta = float64(thread.Samples.ThreadCPUDelta[i]) / 1000.0 // Convert to ms
			} else {
				cpuDelta = interval // Use interval as estimate
			}

			// Update thread-level stats
			if threadCategories[thread.Name][catName] == nil {
				threadCategories[thread.Name][catName] = &threadCategoryData{}
			}
			threadCategories[thread.Name][catName].sampleCount++
			threadCategories[thread.Name][catName].cpuTime += cpuDelta

			// Update global stats
			if globalCategories[catName] == nil {
				globalCategories[catName] = &threadCategoryData{}
			}
			globalCategories[catName].sampleCount++
			globalCategories[catName].cpuTime += cpuDelta

			breakdown.TotalTimeMs += cpuDelta
		}
	}

	// Convert global categories to sorted list
	for name, data := range globalCategories {
		percent := 0.0
		if breakdown.TotalTimeMs > 0 {
			percent = (data.cpuTime / breakdown.TotalTimeMs) * 100
		}
		breakdown.Categories = append(breakdown.Categories, CategoryStats{
			Name:        name,
			TimeMs:      data.cpuTime,
			Percent:     percent,
			SampleCount: data.sampleCount,
		})
	}

	// Sort by time descending
	sort.Slice(breakdown.Categories, func(i, j int) bool {
		return breakdown.Categories[i].TimeMs > breakdown.Categories[j].TimeMs
	})

	// Convert per-thread categories
	for threadName, catMap := range threadCategories {
		var threadStats []CategoryStats
		threadTotal := 0.0

		for _, data := range catMap {
			threadTotal += data.cpuTime
		}

		for name, data := range catMap {
			percent := 0.0
			if threadTotal > 0 {
				percent = (data.cpuTime / threadTotal) * 100
			}
			threadStats = append(threadStats, CategoryStats{
				Name:        name,
				TimeMs:      data.cpuTime,
				Percent:     percent,
				SampleCount: data.sampleCount,
			})
		}

		sort.Slice(threadStats, func(i, j int) bool {
			return threadStats[i].TimeMs > threadStats[j].TimeMs
		})

		breakdown.ByThread[threadName] = threadStats
	}

	return breakdown
}
