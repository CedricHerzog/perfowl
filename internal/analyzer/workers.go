package analyzer

import (
	"fmt"
	"sort"
	"strings"

	"github.com/CedricHerzog/perfowl/internal/parser"
)

// WorkerStats contains analysis for a single worker thread
type WorkerStats struct {
	ThreadName       string          `json:"thread_name"`
	ThreadID         string          `json:"thread_id"`
	ProcessID        string          `json:"process_id"`
	CPUTimeMs        float64         `json:"cpu_time_ms"`
	IdleTimeMs       float64         `json:"idle_time_ms"`
	ActivePercent    float64         `json:"active_percent"`
	MessagesSent     int             `json:"messages_sent"`
	MessagesReceived int             `json:"messages_received"`
	SyncWaitCount    int             `json:"sync_wait_count"`
	SyncWaitTimeMs   float64         `json:"sync_wait_time_ms"`
	TopCategories    []CategoryStats `json:"top_categories,omitempty"`
}

// SyncPoint represents a synchronization event affecting multiple threads
type SyncPoint struct {
	Time        float64  `json:"time_ms"`
	Type        string   `json:"type"`
	Description string   `json:"description"`
	Threads     []string `json:"threads_involved"`
	Duration    float64  `json:"duration_ms"`
}

// WorkerAnalysis contains the full worker thread analysis
type WorkerAnalysis struct {
	TotalWorkers      int           `json:"total_workers"`
	ActiveWorkers     int           `json:"active_workers"`
	TotalCPUTimeMs    float64       `json:"total_cpu_time_ms"`
	TotalIdleTimeMs   float64       `json:"total_idle_time_ms"`
	OverallEfficiency float64       `json:"overall_efficiency_percent"`
	Workers           []WorkerStats `json:"workers"`
	SyncPoints        []SyncPoint   `json:"sync_points,omitempty"`
	Warnings          []string      `json:"warnings,omitempty"`
}

// isWorkerThread checks if a thread is a web worker thread (not browser-internal threads)
func isWorkerThread(thread *parser.Thread) bool {
	name := strings.ToLower(thread.Name)

	// Exclude browser-internal worker threads that aren't web workers
	excludePatterns := []string{
		"threadpoolforegroundworker", // Chrome's internal thread pool
		"threadpoolbackgroundworker", // Chrome's internal thread pool
		"compositortileworker",       // Chrome's compositor tile worker
		"audioworklet",               // Audio worklet (different from web workers)
		"paintworklet",               // Paint worklet
		"v8:profevntproc",            // V8 profiler event processor
	}
	for _, pattern := range excludePatterns {
		if strings.Contains(name, pattern) {
			return false
		}
	}

	// Check for web worker patterns
	if strings.Contains(name, "dom worker") || // Firefox web workers
		strings.Contains(name, "dedicatedworker") || // Chrome dedicated workers
		strings.Contains(name, "sharedworker") || // Shared workers
		strings.Contains(name, "serviceworker") { // Service workers
		return true
	}

	// Generic "worker" in name, but not main thread and not excluded
	if strings.Contains(name, "worker") && !thread.IsMainThread {
		return true
	}

	return false
}

// AnalyzeWorkers performs detailed worker thread analysis
func AnalyzeWorkers(profile *parser.Profile) WorkerAnalysis {
	analysis := WorkerAnalysis{
		Workers:    make([]WorkerStats, 0),
		SyncPoints: make([]SyncPoint, 0),
		Warnings:   make([]string, 0),
	}

	// Map category index to name
	categoryNames := make(map[int]string)
	for i, cat := range profile.Meta.Categories {
		categoryNames[i] = cat.Name
	}

	interval := profile.Meta.Interval
	profileDuration := profile.DurationSeconds() * 1000 // Convert to ms

	// Track all worker messaging events for sync point detection
	type messageEvent struct {
		time       float64
		threadName string
		isSync     bool
		duration   float64
	}
	var messageEvents []messageEvent

	for _, thread := range profile.Threads {
		if !isWorkerThread(&thread) {
			continue
		}

		stats := WorkerStats{
			ThreadName: thread.Name,
			ThreadID:   thread.TID.String(),
			ProcessID:  thread.PID.String(),
		}

		// Calculate CPU time from samples
		categoryTime := make(map[string]float64)
		for i := 0; i < thread.Samples.Length; i++ {
			cpuDelta := interval
			if i < len(thread.Samples.ThreadCPUDelta) && thread.Samples.ThreadCPUDelta[i] > 0 {
				cpuDelta = float64(thread.Samples.ThreadCPUDelta[i]) / 1000.0
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

		// Calculate idle time (profile duration - CPU time)
		if profileDuration > 0 {
			stats.IdleTimeMs = profileDuration - stats.CPUTimeMs
			if stats.IdleTimeMs < 0 {
				stats.IdleTimeMs = 0
			}
			stats.ActivePercent = (stats.CPUTimeMs / profileDuration) * 100
		}

		// Analyze markers for messaging and synchronization
		markers := parser.ExtractMarkers(&thread, profile.Meta.Categories)
		for _, m := range markers {
			// Track messages
			switch m.Name {
			case "JSActorMessage", "FrameMessage":
				stats.MessagesSent++
				isSync := false
				if sync, ok := m.Data["sync"].(bool); ok && sync {
					isSync = true
					stats.SyncWaitCount++
					stats.SyncWaitTimeMs += m.Duration
				}
				messageEvents = append(messageEvents, messageEvent{
					time:       m.StartTime,
					threadName: thread.Name,
					isSync:     isSync,
					duration:   m.Duration,
				})
			case "postMessage":
				stats.MessagesSent++
			}

			// Track IPC waits
			if m.Category == "IPC" {
				if sync, ok := m.Data["sync"].(bool); ok && sync {
					stats.SyncWaitCount++
					stats.SyncWaitTimeMs += m.Duration
				}
			}
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

		analysis.Workers = append(analysis.Workers, stats)
		analysis.TotalCPUTimeMs += stats.CPUTimeMs
		analysis.TotalIdleTimeMs += stats.IdleTimeMs
	}

	analysis.TotalWorkers = len(analysis.Workers)

	// Count active workers (>5% CPU utilization)
	for _, w := range analysis.Workers {
		if w.ActivePercent > 5 {
			analysis.ActiveWorkers++
		}
	}

	// Calculate overall efficiency
	if analysis.TotalCPUTimeMs+analysis.TotalIdleTimeMs > 0 {
		analysis.OverallEfficiency = (analysis.TotalCPUTimeMs / (analysis.TotalCPUTimeMs + analysis.TotalIdleTimeMs)) * 100
	}

	// Detect synchronization points (multiple workers blocked within 10ms window)
	sort.Slice(messageEvents, func(i, j int) bool {
		return messageEvents[i].time < messageEvents[j].time
	})

	for i := 0; i < len(messageEvents); i++ {
		if !messageEvents[i].isSync {
			continue
		}

		// Find nearby sync events (within 10ms)
		involvedThreads := make(map[string]bool)
		involvedThreads[messageEvents[i].threadName] = true
		maxDuration := messageEvents[i].duration

		for j := i + 1; j < len(messageEvents) && messageEvents[j].time-messageEvents[i].time < 10; j++ {
			if messageEvents[j].isSync {
				involvedThreads[messageEvents[j].threadName] = true
				if messageEvents[j].duration > maxDuration {
					maxDuration = messageEvents[j].duration
				}
			}
		}

		if len(involvedThreads) > 1 {
			threads := make([]string, 0, len(involvedThreads))
			for t := range involvedThreads {
				threads = append(threads, t)
			}
			analysis.SyncPoints = append(analysis.SyncPoints, SyncPoint{
				Time:        messageEvents[i].time,
				Type:        "sync_message",
				Description: fmt.Sprintf("Synchronous messaging between %d workers", len(involvedThreads)),
				Threads:     threads,
				Duration:    maxDuration,
			})
			// Skip to next non-overlapping event
			for i++; i < len(messageEvents) && messageEvents[i].time < messageEvents[i-1].time+10; i++ {
			}
			i--
		}
	}

	// Generate warnings
	if analysis.TotalWorkers > 0 && analysis.ActiveWorkers == 0 {
		analysis.Warnings = append(analysis.Warnings, "All worker threads appear idle - check if work is being dispatched")
	}

	for _, w := range analysis.Workers {
		if w.ActivePercent < 10 && w.CPUTimeMs > 0 {
			analysis.Warnings = append(analysis.Warnings,
				fmt.Sprintf("Worker '%s' is mostly idle (%.1f%% active) - possible worker starvation", w.ThreadName, w.ActivePercent))
		}
		if w.SyncWaitTimeMs > 100 {
			analysis.Warnings = append(analysis.Warnings,
				fmt.Sprintf("Worker '%s' spent %.1fms in synchronous waits - consider async alternatives", w.ThreadName, w.SyncWaitTimeMs))
		}
	}

	if len(analysis.SyncPoints) > 5 {
		analysis.Warnings = append(analysis.Warnings,
			fmt.Sprintf("Detected %d synchronization points - workers may be contending for shared resources", len(analysis.SyncPoints)))
	}

	// Sort workers by CPU time descending
	sort.Slice(analysis.Workers, func(i, j int) bool {
		return analysis.Workers[i].CPUTimeMs > analysis.Workers[j].CPUTimeMs
	})

	return analysis
}

// FormatWorkerAnalysis returns a human-readable summary
func FormatWorkerAnalysis(analysis WorkerAnalysis) string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("Worker Thread Analysis (%d workers, %d active)\n",
		analysis.TotalWorkers, analysis.ActiveWorkers))
	sb.WriteString(strings.Repeat("=", 60) + "\n\n")

	sb.WriteString(fmt.Sprintf("Overall Efficiency: %.1f%%\n", analysis.OverallEfficiency))
	sb.WriteString(fmt.Sprintf("Total CPU Time: %.2fms\n", analysis.TotalCPUTimeMs))
	sb.WriteString(fmt.Sprintf("Total Idle Time: %.2fms\n\n", analysis.TotalIdleTimeMs))

	if len(analysis.Workers) > 0 {
		sb.WriteString("Workers by CPU Time:\n")
		sb.WriteString(strings.Repeat("-", 60) + "\n")
		sb.WriteString(fmt.Sprintf("%-25s %10s %10s %8s\n", "Name", "CPU Time", "Idle Time", "Active%"))
		sb.WriteString(strings.Repeat("-", 60) + "\n")

		for _, w := range analysis.Workers {
			name := w.ThreadName
			if len(name) > 25 {
				name = name[:22] + "..."
			}
			sb.WriteString(fmt.Sprintf("%-25s %8.2fms %8.2fms %7.1f%%\n",
				name, w.CPUTimeMs, w.IdleTimeMs, w.ActivePercent))
		}
	}

	if len(analysis.SyncPoints) > 0 {
		sb.WriteString(fmt.Sprintf("\nSynchronization Points: %d\n", len(analysis.SyncPoints)))
	}

	if len(analysis.Warnings) > 0 {
		sb.WriteString("\nWarnings:\n")
		for _, w := range analysis.Warnings {
			sb.WriteString(fmt.Sprintf("  - %s\n", w))
		}
	}

	return sb.String()
}
