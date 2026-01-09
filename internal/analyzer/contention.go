package analyzer

import (
	"fmt"
	"sort"
	"strings"

	"github.com/CedricHerzog/perfowl/internal/parser"
)

// ContentionEvent represents a contention event affecting threads
type ContentionEvent struct {
	Type        string   `json:"type"`
	StartTime   float64  `json:"start_time_ms"`
	Duration    float64  `json:"duration_ms"`
	Threads     []string `json:"affected_threads"`
	Description string   `json:"description"`
}

// ContentionAnalysis contains the full contention analysis
type ContentionAnalysis struct {
	TotalEvents     int               `json:"total_events"`
	TotalImpactMs   float64           `json:"total_impact_ms"`
	GCContention    int               `json:"gc_contention_events"`
	IPCContention   int               `json:"ipc_contention_events"`
	LockContention  int               `json:"lock_contention_events"`
	Events          []ContentionEvent `json:"events,omitempty"`
	Severity        string            `json:"severity"`
	Recommendations []string          `json:"recommendations,omitempty"`
}

// AnalyzeContention performs contention detection analysis
func AnalyzeContention(profile *parser.Profile) ContentionAnalysis {
	analysis := ContentionAnalysis{
		Events:          make([]ContentionEvent, 0),
		Recommendations: make([]string, 0),
	}

	profileDuration := profile.DurationSeconds() * 1000 // Convert to ms

	// Track active threads at each time point
	type threadActivity struct {
		name      string
		startTime float64
		endTime   float64
		isWorker  bool
	}
	var activities []threadActivity

	// Collect GC events
	type gcEvent struct {
		startTime float64
		duration  float64
		gcType    string
	}
	var gcEvents []gcEvent

	// Collect sync IPC events
	type ipcEvent struct {
		startTime  float64
		duration   float64
		threadName string
	}
	var syncIPCEvents []ipcEvent

	// Process each thread
	for _, thread := range profile.Threads {
		isWorker := isWorkerThread(&thread)

		// Estimate thread active periods from samples
		interval := profile.Meta.Interval
		sampleTime := 0.0

		for i := 0; i < thread.Samples.Length; i++ {
			cpuDelta := interval
			if i < len(thread.Samples.ThreadCPUDelta) && thread.Samples.ThreadCPUDelta[i] > 0 {
				cpuDelta = float64(thread.Samples.ThreadCPUDelta[i]) / 1000.0
			}

			if cpuDelta > 0 {
				activities = append(activities, threadActivity{
					name:      thread.Name,
					startTime: sampleTime,
					endTime:   sampleTime + cpuDelta,
					isWorker:  isWorker,
				})
			}
			sampleTime += interval
		}

		// Process markers for GC and IPC events
		markers := parser.ExtractMarkers(&thread, profile.Meta.Categories)
		for _, m := range markers {
			// Track GC events
			if m.Name == "GCMajor" || m.Name == "GCMinor" || m.Name == "GCSlice" {
				gcEvents = append(gcEvents, gcEvent{
					startTime: m.StartTime,
					duration:  m.Duration,
					gcType:    m.Name,
				})
			}

			// Track synchronous IPC
			if m.Category == "IPC" {
				if sync, ok := m.Data["sync"].(bool); ok && sync {
					syncIPCEvents = append(syncIPCEvents, ipcEvent{
						startTime:  m.StartTime,
						duration:   m.Duration,
						threadName: thread.Name,
					})
				}
			}
		}
	}

	// Detect GC contention: GC events while workers are active
	for _, gc := range gcEvents {
		affectedThreads := make(map[string]bool)

		for _, activity := range activities {
			if activity.isWorker {
				// Check if activity overlaps with GC
				if activity.startTime < gc.startTime+gc.duration &&
					activity.endTime > gc.startTime {
					affectedThreads[activity.name] = true
				}
			}
		}

		if len(affectedThreads) > 0 {
			threads := make([]string, 0, len(affectedThreads))
			for t := range affectedThreads {
				threads = append(threads, t)
			}

			analysis.Events = append(analysis.Events, ContentionEvent{
				Type:        "gc_pause",
				StartTime:   gc.startTime,
				Duration:    gc.duration,
				Threads:     threads,
				Description: fmt.Sprintf("%s paused %d worker threads", gc.gcType, len(threads)),
			})
			analysis.GCContention++
			analysis.TotalImpactMs += gc.duration * float64(len(threads))
		}
	}

	// Detect IPC contention: multiple threads doing sync IPC around same time
	sort.Slice(syncIPCEvents, func(i, j int) bool {
		return syncIPCEvents[i].startTime < syncIPCEvents[j].startTime
	})

	for i := 0; i < len(syncIPCEvents); i++ {
		ipc := syncIPCEvents[i]
		involvedThreads := map[string]bool{ipc.threadName: true}
		maxDuration := ipc.duration

		// Look for nearby sync IPC (within 10ms window)
		for j := i + 1; j < len(syncIPCEvents) && syncIPCEvents[j].startTime-ipc.startTime < 10; j++ {
			involvedThreads[syncIPCEvents[j].threadName] = true
			if syncIPCEvents[j].duration > maxDuration {
				maxDuration = syncIPCEvents[j].duration
			}
		}

		if len(involvedThreads) > 1 {
			threads := make([]string, 0, len(involvedThreads))
			for t := range involvedThreads {
				threads = append(threads, t)
			}

			analysis.Events = append(analysis.Events, ContentionEvent{
				Type:        "ipc_wait",
				StartTime:   ipc.startTime,
				Duration:    maxDuration,
				Threads:     threads,
				Description: fmt.Sprintf("Sync IPC contention between %d threads", len(threads)),
			})
			analysis.IPCContention++
			analysis.TotalImpactMs += maxDuration

			// Skip events we've already included
			for i++; i < len(syncIPCEvents) && syncIPCEvents[i].startTime < ipc.startTime+10; i++ {
			}
			i--
		}
	}

	analysis.TotalEvents = len(analysis.Events)

	// Calculate severity
	if profileDuration > 0 {
		contentionPercent := (analysis.TotalImpactMs / profileDuration) * 100
		switch {
		case contentionPercent > 10:
			analysis.Severity = "high"
		case contentionPercent > 5:
			analysis.Severity = "medium"
		case contentionPercent > 1:
			analysis.Severity = "low"
		default:
			analysis.Severity = "minimal"
		}
	} else {
		analysis.Severity = "unknown"
	}

	// Generate recommendations
	if analysis.GCContention > 5 {
		analysis.Recommendations = append(analysis.Recommendations,
			"High GC contention detected - consider reducing memory allocations in hot paths")
	}

	if analysis.IPCContention > 5 {
		analysis.Recommendations = append(analysis.Recommendations,
			"Frequent sync IPC contention - consider using async messaging between threads")
	}

	if analysis.TotalImpactMs > 100 {
		analysis.Recommendations = append(analysis.Recommendations,
			fmt.Sprintf("Total contention impact: %.1fms - significant opportunity for optimization", analysis.TotalImpactMs))
	}

	if analysis.Severity == "high" {
		analysis.Recommendations = append(analysis.Recommendations,
			"Consider profiling with GC/CC categories enabled for more detailed analysis")
	}

	// Sort events by duration (descending)
	sort.Slice(analysis.Events, func(i, j int) bool {
		return analysis.Events[i].Duration > analysis.Events[j].Duration
	})

	// Limit events to top 50
	if len(analysis.Events) > 50 {
		analysis.Events = analysis.Events[:50]
	}

	return analysis
}

// FormatContentionAnalysis returns a human-readable summary
func FormatContentionAnalysis(analysis ContentionAnalysis) string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("Contention Analysis (%d events, severity: %s)\n",
		analysis.TotalEvents, analysis.Severity))
	sb.WriteString(strings.Repeat("=", 60) + "\n\n")

	sb.WriteString(fmt.Sprintf("Total Impact: %.2fms\n", analysis.TotalImpactMs))
	sb.WriteString(fmt.Sprintf("GC Contention Events: %d\n", analysis.GCContention))
	sb.WriteString(fmt.Sprintf("IPC Contention Events: %d\n", analysis.IPCContention))
	sb.WriteString(fmt.Sprintf("Lock Contention Events: %d\n\n", analysis.LockContention))

	if len(analysis.Events) > 0 {
		sb.WriteString("Top Contention Events:\n")
		sb.WriteString(strings.Repeat("-", 60) + "\n")

		displayCount := len(analysis.Events)
		if displayCount > 10 {
			displayCount = 10
		}

		for i := 0; i < displayCount; i++ {
			e := analysis.Events[i]
			sb.WriteString(fmt.Sprintf("  %.2fms: %s (%.2fms, %d threads)\n",
				e.StartTime, e.Type, e.Duration, len(e.Threads)))
		}

		if len(analysis.Events) > 10 {
			sb.WriteString(fmt.Sprintf("\n  ... and %d more events\n", len(analysis.Events)-10))
		}
		sb.WriteString("\n")
	}

	if len(analysis.Recommendations) > 0 {
		sb.WriteString("Recommendations:\n")
		for _, r := range analysis.Recommendations {
			sb.WriteString(fmt.Sprintf("  - %s\n", r))
		}
	}

	return sb.String()
}
