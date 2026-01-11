package analyzer

import (
	"fmt"
	"strings"

	"github.com/CedricHerzog/perfowl/internal/parser"
)

// Severity represents the severity level of a bottleneck
type Severity int

const (
	SeverityLow Severity = iota
	SeverityMedium
	SeverityHigh
)

func (s Severity) String() string {
	switch s {
	case SeverityHigh:
		return "high"
	case SeverityMedium:
		return "medium"
	default:
		return "low"
	}
}

func (s Severity) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf(`"%s"`, s.String())), nil
}

// ParseSeverity parses a severity string
func ParseSeverity(s string) Severity {
	switch strings.ToLower(s) {
	case "high":
		return SeverityHigh
	case "medium":
		return SeverityMedium
	default:
		return SeverityLow
	}
}

// Bottleneck represents a detected performance issue
type Bottleneck struct {
	Type           string   `json:"type"`
	Severity       Severity `json:"severity"`
	Count          int      `json:"count"`
	TotalDuration  float64  `json:"total_duration_ms"`
	AvgDuration    float64  `json:"avg_duration_ms,omitempty"`
	MaxDuration    float64  `json:"max_duration_ms,omitempty"`
	Description    string   `json:"description"`
	Recommendation string   `json:"recommendation,omitempty"`
	Locations      []string `json:"locations,omitempty"`
}

// BottleneckReport contains the analysis results
type BottleneckReport struct {
	Score       int          `json:"score"`
	Summary     string       `json:"summary"`
	Bottlenecks []Bottleneck `json:"bottlenecks"`
}

// Thresholds for bottleneck detection
const (
	LongTaskThresholdMs     = 50.0   // Tasks blocking main thread > 50ms
	GCPauseLongMs           = 100.0  // GC pause considered long
	GCFrequencyHighPerSec   = 2.0    // More than 2 GC/sec is high
	LayoutThrashingWindowMs = 100.0  // Window to detect rapid reflows
	LayoutThrashingMinCount = 5      // Min reflows in window
	SyncIPCThresholdMs      = 10.0   // Sync IPC calls > 10ms
	NetworkBlockingMs       = 1000.0 // Network blocking > 1s
)

// DetectBottlenecks analyzes the profile and returns detected bottlenecks
func DetectBottlenecks(profile *parser.Profile) []Bottleneck {
	var bottlenecks []Bottleneck

	// Collect all markers
	var allMarkers []parser.ParsedMarker
	for _, thread := range profile.Threads {
		markers := parser.ExtractMarkers(&thread, profile.Meta.Categories)
		allMarkers = append(allMarkers, markers...)
	}

	// Detect long tasks
	if b := detectLongTasks(allMarkers); b != nil {
		bottlenecks = append(bottlenecks, *b)
	}

	// Detect GC pressure
	if b := detectGCPressure(allMarkers, profile.DurationSeconds()); b != nil {
		bottlenecks = append(bottlenecks, *b)
	}

	// Detect sync IPC
	if b := detectSyncIPC(allMarkers); b != nil {
		bottlenecks = append(bottlenecks, *b)
	}

	// Detect layout thrashing
	if b := detectLayoutThrashing(allMarkers); b != nil {
		bottlenecks = append(bottlenecks, *b)
	}

	// Detect network blocking
	if b := detectNetworkBlocking(allMarkers); b != nil {
		bottlenecks = append(bottlenecks, *b)
	}

	// Detect extension overhead
	if b := detectExtensionOverhead(allMarkers, profile); b != nil {
		bottlenecks = append(bottlenecks, *b)
	}

	return bottlenecks
}

func detectLongTasks(markers []parser.ParsedMarker) *Bottleneck {
	var longTasks []parser.ParsedMarker
	var totalDuration, maxDuration float64

	// Max reasonable task duration (10 seconds) - anything longer is likely a profile span marker
	const maxReasonableDuration = 10000.0

	for _, m := range markers {
		// Only consider MainThreadLongTask markers - these are actual long task indicators
		if m.Name == "MainThreadLongTask" {
			if m.Duration > LongTaskThresholdMs && m.Duration < maxReasonableDuration {
				longTasks = append(longTasks, m)
				totalDuration += m.Duration
				if m.Duration > maxDuration {
					maxDuration = m.Duration
				}
			}
		}
	}

	if len(longTasks) == 0 {
		return nil
	}

	severity := SeverityLow
	if len(longTasks) > 10 || maxDuration > 200 {
		severity = SeverityHigh
	} else if len(longTasks) > 5 || maxDuration > 100 {
		severity = SeverityMedium
	}

	locations := make([]string, 0, min(10, len(longTasks)))
	for i := 0; i < min(10, len(longTasks)); i++ {
		locations = append(locations, fmt.Sprintf("%.2fms in %s thread", longTasks[i].Duration, longTasks[i].ThreadName))
	}

	return &Bottleneck{
		Type:           "Long Tasks",
		Severity:       severity,
		Count:          len(longTasks),
		TotalDuration:  totalDuration,
		AvgDuration:    totalDuration / float64(len(longTasks)),
		MaxDuration:    maxDuration,
		Description:    fmt.Sprintf("%d tasks blocked the main thread for >%.0fms", len(longTasks), LongTaskThresholdMs),
		Recommendation: "Investigate long-running JavaScript and consider breaking into smaller chunks or using Web Workers",
		Locations:      locations,
	}
}

func detectGCPressure(markers []parser.ParsedMarker, durationSec float64) *Bottleneck {
	var gcMarkers []parser.ParsedMarker
	var totalDuration, maxDuration float64

	// Max reasonable GC duration (5 seconds)
	const maxReasonableGCDuration = 5000.0

	for _, m := range markers {
		// Match Firefox GC markers (GCMajor, GCMinor, etc.) and Chrome GC events (V8.GC_*)
		isGC := strings.HasPrefix(m.Name, "GC") ||
			strings.HasPrefix(m.Name, "V8.GC") ||
			m.Category == "GC / CC"
		if isGC {
			// Only count markers with reasonable positive durations
			if m.Duration > 0 && m.Duration < maxReasonableGCDuration {
				gcMarkers = append(gcMarkers, m)
				totalDuration += m.Duration
				if m.Duration > maxDuration {
					maxDuration = m.Duration
				}
			}
		}
	}

	if len(gcMarkers) == 0 {
		return nil
	}

	gcPerSec := float64(len(gcMarkers)) / durationSec

	severity := SeverityLow
	if gcPerSec > GCFrequencyHighPerSec*2 || maxDuration > GCPauseLongMs*2 {
		severity = SeverityHigh
	} else if gcPerSec > GCFrequencyHighPerSec || maxDuration > GCPauseLongMs {
		severity = SeverityMedium
	}

	// If severity is low and total duration is minimal, skip reporting
	if severity == SeverityLow && totalDuration < 500 {
		return nil
	}

	locations := make([]string, 0, min(5, len(gcMarkers)))
	for i := 0; i < min(5, len(gcMarkers)); i++ {
		locations = append(locations, fmt.Sprintf("%s: %.2fms", gcMarkers[i].Name, gcMarkers[i].Duration))
	}

	return &Bottleneck{
		Type:           "GC Pressure",
		Severity:       severity,
		Count:          len(gcMarkers),
		TotalDuration:  totalDuration,
		AvgDuration:    totalDuration / float64(len(gcMarkers)),
		MaxDuration:    maxDuration,
		Description:    fmt.Sprintf("%.1f GC events/sec with %.0fms total pause time (max: %.2fms)", gcPerSec, totalDuration, maxDuration),
		Recommendation: "Reduce object allocations, consider object pooling, and avoid creating unnecessary temporary objects",
		Locations:      locations,
	}
}

func detectSyncIPC(markers []parser.ParsedMarker) *Bottleneck {
	var syncIPCMarkers []parser.ParsedMarker
	var totalDuration, maxDuration float64

	// Max reasonable IPC duration (2 seconds)
	const maxReasonableIPCDuration = 2000.0

	for _, m := range markers {
		if m.Category == "IPC" || strings.Contains(m.Name, "IPC") {
			// Only count markers with reasonable positive durations
			if m.Duration <= 0 || m.Duration > maxReasonableIPCDuration {
				continue
			}

			// Check for sync indicator in data
			if sync, ok := m.Data["sync"].(bool); ok && sync {
				syncIPCMarkers = append(syncIPCMarkers, m)
				totalDuration += m.Duration
				if m.Duration > maxDuration {
					maxDuration = m.Duration
				}
			} else if m.Duration > SyncIPCThresholdMs {
				// Long IPC calls are likely sync
				syncIPCMarkers = append(syncIPCMarkers, m)
				totalDuration += m.Duration
				if m.Duration > maxDuration {
					maxDuration = m.Duration
				}
			}
		}
	}

	if len(syncIPCMarkers) == 0 {
		return nil
	}

	severity := SeverityLow
	if len(syncIPCMarkers) > 20 || totalDuration > 500 {
		severity = SeverityHigh
	} else if len(syncIPCMarkers) > 10 || totalDuration > 200 {
		severity = SeverityMedium
	}

	locations := make([]string, 0, min(5, len(syncIPCMarkers)))
	for i := 0; i < min(5, len(syncIPCMarkers)); i++ {
		locations = append(locations, fmt.Sprintf("%.2fms in %s", syncIPCMarkers[i].Duration, syncIPCMarkers[i].ThreadName))
	}

	return &Bottleneck{
		Type:           "Synchronous IPC",
		Severity:       severity,
		Count:          len(syncIPCMarkers),
		TotalDuration:  totalDuration,
		MaxDuration:    maxDuration,
		Description:    fmt.Sprintf("%d synchronous IPC calls blocking threads for %.0fms total", len(syncIPCMarkers), totalDuration),
		Recommendation: "Consider using async IPC alternatives to avoid blocking the main thread",
		Locations:      locations,
	}
}

func detectLayoutThrashing(markers []parser.ParsedMarker) *Bottleneck {
	var layoutMarkers []parser.ParsedMarker
	var totalDuration, maxDuration float64

	// Max reasonable layout duration (1 second)
	const maxReasonableLayoutDuration = 1000.0

	for _, m := range markers {
		if m.Category == "Layout" || m.Name == "Styles" || strings.Contains(m.Name, "Reflow") {
			// Only count markers with reasonable positive durations
			if m.Duration > 0 && m.Duration < maxReasonableLayoutDuration {
				layoutMarkers = append(layoutMarkers, m)
				totalDuration += m.Duration
				if m.Duration > maxDuration {
					maxDuration = m.Duration
				}
			}
		}
	}

	if len(layoutMarkers) < LayoutThrashingMinCount {
		return nil
	}

	// Simple thrashing detection: many layout events in succession
	thrashingCount := 0
	for i := 1; i < len(layoutMarkers); i++ {
		if layoutMarkers[i].StartTime-layoutMarkers[i-1].StartTime < LayoutThrashingWindowMs {
			thrashingCount++
		}
	}

	if thrashingCount < LayoutThrashingMinCount {
		return nil
	}

	severity := SeverityLow
	if thrashingCount > 50 || totalDuration > 1000 {
		severity = SeverityHigh
	} else if thrashingCount > 20 || totalDuration > 500 {
		severity = SeverityMedium
	}

	return &Bottleneck{
		Type:           "Layout Thrashing",
		Severity:       severity,
		Count:          thrashingCount,
		TotalDuration:  totalDuration,
		MaxDuration:    maxDuration,
		Description:    fmt.Sprintf("%d rapid layout recalculations detected, causing %.0fms of layout work", thrashingCount, totalDuration),
		Recommendation: "Batch DOM reads and writes separately, use requestAnimationFrame for visual updates",
		Locations:      []string{},
	}
}

func detectNetworkBlocking(markers []parser.ParsedMarker) *Bottleneck {
	var networkMarkers []parser.ParsedMarker
	var totalDuration, maxDuration float64

	// Max reasonable network duration (30 seconds)
	const maxReasonableNetworkDuration = 30000.0

	for _, m := range markers {
		if m.Category == "Network" || m.Name == "ChannelMarker" || m.Name == "HostResolver" {
			// Only count markers with reasonable positive durations above threshold
			if m.Duration > NetworkBlockingMs && m.Duration < maxReasonableNetworkDuration {
				networkMarkers = append(networkMarkers, m)
				totalDuration += m.Duration
				if m.Duration > maxDuration {
					maxDuration = m.Duration
				}
			}
		}
	}

	if len(networkMarkers) == 0 {
		return nil
	}

	severity := SeverityLow
	if len(networkMarkers) > 5 || maxDuration > 5000 {
		severity = SeverityMedium
	}

	locations := make([]string, 0, min(5, len(networkMarkers)))
	for i := 0; i < min(5, len(networkMarkers)); i++ {
		url := ""
		if u, ok := networkMarkers[i].Data["url"].(string); ok {
			url = u
			if len(url) > 50 {
				url = url[:50] + "..."
			}
		}
		locations = append(locations, fmt.Sprintf("%.0fms: %s", networkMarkers[i].Duration, url))
	}

	return &Bottleneck{
		Type:           "Network Blocking",
		Severity:       severity,
		Count:          len(networkMarkers),
		TotalDuration:  totalDuration,
		MaxDuration:    maxDuration,
		Description:    fmt.Sprintf("%d slow network requests (>%.0fms each)", len(networkMarkers), NetworkBlockingMs),
		Recommendation: "Optimize slow requests, consider caching, or load resources asynchronously",
		Locations:      locations,
	}
}

func detectExtensionOverhead(markers []parser.ParsedMarker, profile *parser.Profile) *Bottleneck {
	extensionURLs := profile.GetExtensionBaseURLs()
	if len(extensionURLs) == 0 {
		return nil
	}

	extensionActivity := make(map[string]struct {
		count    int
		duration float64
	})

	for _, m := range markers {
		// Check if marker is related to extension
		if url, ok := m.Data["url"].(string); ok {
			for extID, baseURL := range extensionURLs {
				if strings.HasPrefix(url, baseURL) {
					act := extensionActivity[extID]
					act.count++
					act.duration += m.Duration
					extensionActivity[extID] = act
					break
				}
			}
		}

		// Check JSActorMessage markers for extension
		if m.Name == "JSActorMessage" {
			if actor, ok := m.Data["actor"].(string); ok {
				for extID := range extensionURLs {
					if strings.Contains(actor, extID) || strings.Contains(actor, "WebExtensions") {
						act := extensionActivity[extID]
						act.count++
						act.duration += m.Duration
						extensionActivity[extID] = act
						break
					}
				}
			}
		}
	}

	// Calculate total extension overhead
	var totalDuration float64
	var totalCount int
	var locations []string

	extensions := profile.GetExtensions()
	for extID, act := range extensionActivity {
		totalCount += act.count
		totalDuration += act.duration
		name := extensions[extID]
		if name == "" {
			name = extID
		}
		locations = append(locations, fmt.Sprintf("%s: %d events, %.0fms", name, act.count, act.duration))
	}

	if totalCount == 0 {
		return nil
	}

	// Only report if significant
	if totalDuration < 100 && totalCount < 50 {
		return nil
	}

	severity := SeverityLow
	if totalDuration > 1000 {
		severity = SeverityMedium
	}

	return &Bottleneck{
		Type:           "Extension Overhead",
		Severity:       severity,
		Count:          totalCount,
		TotalDuration:  totalDuration,
		Description:    fmt.Sprintf("%d extension-related events consuming %.0fms", totalCount, totalDuration),
		Recommendation: "Review extension activity, consider disabling extensions during performance-critical operations",
		Locations:      locations,
	}
}

// CalculateScore returns a performance score from 0-100
func CalculateScore(bottlenecks []Bottleneck) int {
	score := 100

	for _, b := range bottlenecks {
		switch b.Severity {
		case SeverityHigh:
			score -= 20
		case SeverityMedium:
			score -= 10
		case SeverityLow:
			score -= 5
		}
	}

	if score < 0 {
		score = 0
	}

	return score
}

// GenerateSummary creates a human-readable summary of the bottlenecks
func GenerateSummary(bottlenecks []Bottleneck, profile *parser.Profile) string {
	if len(bottlenecks) == 0 {
		return "No significant performance bottlenecks detected. The profile shows healthy performance characteristics."
	}

	var highCount, mediumCount, lowCount int
	for _, b := range bottlenecks {
		switch b.Severity {
		case SeverityHigh:
			highCount++
		case SeverityMedium:
			mediumCount++
		case SeverityLow:
			lowCount++
		}
	}

	parts := []string{}
	if highCount > 0 {
		parts = append(parts, fmt.Sprintf("%d high severity", highCount))
	}
	if mediumCount > 0 {
		parts = append(parts, fmt.Sprintf("%d medium severity", mediumCount))
	}
	if lowCount > 0 {
		parts = append(parts, fmt.Sprintf("%d low severity", lowCount))
	}

	return fmt.Sprintf("Detected %d bottleneck(s): %s. See details below for recommendations.",
		len(bottlenecks), strings.Join(parts, ", "))
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
