package analyzer

import (
	"sort"
	"strings"

	"github.com/CedricHerzog/perfowl/internal/parser"
)

// ExtensionReport contains analysis for a single extension
type ExtensionReport struct {
	ID            string         `json:"id"`
	Name          string         `json:"name"`
	BaseURL       string         `json:"base_url"`
	TotalDuration float64        `json:"total_duration_ms"`
	MarkersCount  int            `json:"markers_count"`
	DOMEvents     int            `json:"dom_events"`
	IPCMessages   int            `json:"ipc_messages"`
	ImpactScore   string         `json:"impact_score"`
	TopMarkers    []MarkerSummary `json:"top_markers,omitempty"`
}

// MarkerSummary is a simplified marker for reporting
type MarkerSummary struct {
	Name     string  `json:"name"`
	Category string  `json:"category"`
	Duration float64 `json:"duration_ms"`
}

// ExtensionsAnalysis contains the overall extension analysis
type ExtensionsAnalysis struct {
	TotalExtensions int               `json:"total_extensions"`
	TotalDuration   float64           `json:"total_duration_ms"`
	TotalEvents     int               `json:"total_events"`
	Extensions      []ExtensionReport `json:"extensions"`
}

// AnalyzeExtensions analyzes extension performance in the profile
func AnalyzeExtensions(profile *parser.Profile) ExtensionsAnalysis {
	analysis := ExtensionsAnalysis{
		TotalExtensions: profile.ExtensionCount(),
		Extensions:      make([]ExtensionReport, 0),
	}

	extensions := profile.GetExtensions()
	extensionURLs := profile.GetExtensionBaseURLs()

	// Initialize reports for all extensions
	extReports := make(map[string]*ExtensionReport)
	for id, name := range extensions {
		extReports[id] = &ExtensionReport{
			ID:      id,
			Name:    name,
			BaseURL: extensionURLs[id],
		}
	}

	// Collect all markers
	var allMarkers []parser.ParsedMarker
	for _, thread := range profile.Threads {
		markers := parser.ExtractMarkers(&thread, profile.Meta.Categories)
		allMarkers = append(allMarkers, markers...)
	}

	// Analyze markers for extension activity
	for _, m := range allMarkers {
		// Skip markers with unreasonable durations
		if m.Duration < 0 || m.Duration > 10000 {
			continue
		}

		extID := matchExtension(m, extensionURLs)
		if extID == "" {
			continue
		}

		report := extReports[extID]
		if report == nil {
			continue
		}

		report.MarkersCount++
		report.TotalDuration += m.Duration

		// Count specific event types
		if m.Name == "DOMEvent" || strings.Contains(m.Category, "DOM") {
			report.DOMEvents++
		}
		if strings.Contains(m.Name, "IPC") || strings.Contains(m.Name, "Message") ||
			m.Name == "JSActorMessage" || m.Name == "FrameMessage" {
			report.IPCMessages++
		}

		// Track top markers by duration
		if m.Duration > 1 { // Only track markers > 1ms
			report.TopMarkers = append(report.TopMarkers, MarkerSummary{
				Name:     m.Name,
				Category: m.Category,
				Duration: m.Duration,
			})
		}
	}

	// Process extension reports
	for _, report := range extReports {
		if report.MarkersCount == 0 && report.TotalDuration == 0 {
			continue // Skip extensions with no activity
		}

		// Sort and limit top markers
		sort.Slice(report.TopMarkers, func(i, j int) bool {
			return report.TopMarkers[i].Duration > report.TopMarkers[j].Duration
		})
		if len(report.TopMarkers) > 5 {
			report.TopMarkers = report.TopMarkers[:5]
		}

		// Calculate impact score
		report.ImpactScore = calculateImpactScore(report)

		analysis.Extensions = append(analysis.Extensions, *report)
		analysis.TotalDuration += report.TotalDuration
		analysis.TotalEvents += report.MarkersCount
	}

	// Sort by total duration
	sort.Slice(analysis.Extensions, func(i, j int) bool {
		return analysis.Extensions[i].TotalDuration > analysis.Extensions[j].TotalDuration
	})

	return analysis
}

// matchExtension checks if a marker is related to an extension
func matchExtension(m parser.ParsedMarker, extensionURLs map[string]string) string {
	// Check URL in marker data
	if url, ok := m.Data["url"].(string); ok {
		for extID, baseURL := range extensionURLs {
			if strings.HasPrefix(url, baseURL) || strings.Contains(url, "moz-extension://") {
				return extID
			}
		}
	}

	// Check target in DOMEvent
	if target, ok := m.Data["target"].(string); ok {
		for extID, baseURL := range extensionURLs {
			if strings.Contains(target, baseURL) {
				return extID
			}
		}
	}

	// Check JSActorMessage actor field
	if m.Name == "JSActorMessage" || m.Type == "JSActorMessage" {
		if actor, ok := m.Data["actor"].(string); ok {
			// WebExtensions actors are related to extensions
			if strings.Contains(actor, "WebExtension") {
				// Return first extension (generic WebExtension activity)
				for extID := range extensionURLs {
					return extID
				}
			}
			// Conduits are used for extension port messaging
			if actor == "Conduits" {
				if msgName, ok := m.Data["name"].(string); ok {
					if strings.Contains(msgName, "PortMessage") || strings.Contains(msgName, "Port") {
						// Generic extension messaging
						for extID := range extensionURLs {
							return extID
						}
					}
				}
			}
		}
	}

	// Check FrameMessage for extension-related messages
	if m.Name == "FrameMessage" || m.Type == "FrameMessage" {
		if msgName, ok := m.Data["name"].(string); ok {
			if strings.Contains(msgName, "Extension") || strings.Contains(msgName, "WebExt") ||
				strings.Contains(msgName, "addons") {
				for extID := range extensionURLs {
					return extID
				}
			}
		}
	}

	// Check for extension-specific text markers
	if m.Name == "Text" || m.Type == "Text" {
		if name, ok := m.Data["name"].(string); ok {
			if strings.Contains(name, "moz-extension://") {
				for extID, baseURL := range extensionURLs {
					if strings.Contains(name, baseURL) {
						return extID
					}
				}
				// Generic extension activity
				for extID := range extensionURLs {
					return extID
				}
			}
		}
	}

	return ""
}

// calculateImpactScore determines the performance impact level
func calculateImpactScore(report *ExtensionReport) string {
	// Scoring based on duration and event count
	score := 0

	// Duration scoring
	if report.TotalDuration > 1000 {
		score += 3
	} else if report.TotalDuration > 500 {
		score += 2
	} else if report.TotalDuration > 100 {
		score += 1
	}

	// Event count scoring
	if report.MarkersCount > 1000 {
		score += 3
	} else if report.MarkersCount > 500 {
		score += 2
	} else if report.MarkersCount > 100 {
		score += 1
	}

	// IPC message scoring (IPC is expensive)
	if report.IPCMessages > 100 {
		score += 2
	} else if report.IPCMessages > 50 {
		score += 1
	}

	// Convert to impact level
	if score >= 5 {
		return "high"
	} else if score >= 3 {
		return "medium"
	}
	return "low"
}
