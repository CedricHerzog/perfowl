package analyzer

import (
	"fmt"
	"sort"
	"strings"

	"github.com/CedricHerzog/perfowl/internal/parser"
)

// DelimiterMarker represents a potential start/end point for operation timing
type DelimiterMarker struct {
	TimeMs     float64                `json:"time_ms"`
	DurationMs float64                `json:"duration_ms,omitempty"`
	Name       string                 `json:"name"`
	Type       string                 `json:"type"`
	Category   string                 `json:"category"`
	Thread     string                 `json:"thread"`
	Data       map[string]interface{} `json:"data,omitempty"`
}

// OperationMeasurement is the result of measuring between two delimiters
type OperationMeasurement struct {
	StartMarker     DelimiterMarker `json:"start_marker"`
	EndMarker       DelimiterMarker `json:"end_marker"`
	OperationTimeMs float64         `json:"operation_time_ms"`
}

// Delimiter marker types that are useful for operation timing
var delimiterTypes = map[string]bool{
	"DOMEvent":           true, // click, input, submit - operation start (Firefox)
	"EventDispatch":      true, // click, input, submit - operation start (Chrome)
	"UserTiming":         true, // performance.mark() - explicit markers
	"Styles":             true, // DOM style recalculation (Firefox)
	"UpdateLayoutTree":   true, // DOM style recalculation (Chrome)
	"Reflow":             true, // Layout reflow
	"Paint":              true, // Visual paint
	"Composite":          true, // GPU compositing
	"MainThreadLongTask": true, // Long tasks
	"Navigation":         true, // Page navigation events
	"Load":               true, // Load events
}

// GetDelimiterMarkers returns markers suitable for operation timing from all threads
func GetDelimiterMarkers(profile *parser.Profile, categories []string) []DelimiterMarker {
	var allMarkers []DelimiterMarker
	categoryFilter := make(map[string]bool)
	for _, c := range categories {
		categoryFilter[strings.TrimSpace(c)] = true
	}

	for _, thread := range profile.Threads {
		markers := parser.ExtractMarkers(&thread, profile.Meta.Categories)

		for _, m := range markers {
			// Check if marker type is useful for delimiting
			if !isDelimiterMarker(m) {
				continue
			}

			// Apply category filter if specified
			if len(categoryFilter) > 0 && !categoryFilter[m.Category] {
				continue
			}

			dm := DelimiterMarker{
				TimeMs:     m.StartTime,
				DurationMs: m.Duration,
				Name:       m.Name,
				Type:       string(m.Type),
				Category:   m.Category,
				Thread:     m.ThreadName,
				Data:       m.Data,
			}

			allMarkers = append(allMarkers, dm)
		}
	}

	// Sort by time
	sort.Slice(allMarkers, func(i, j int) bool {
		return allMarkers[i].TimeMs < allMarkers[j].TimeMs
	})

	return allMarkers
}

// isDelimiterMarker checks if a marker is useful for operation timing
func isDelimiterMarker(m parser.ParsedMarker) bool {
	// Check by type
	if delimiterTypes[string(m.Type)] {
		return true
	}

	// Check by name
	if delimiterTypes[m.Name] {
		return true
	}

	// Check by category - include Layout, Graphics, DOM, and UserTiming categories
	if m.Category == "Layout" || m.Category == "Graphics" || m.Category == "DOM" || m.Category == "UserTiming" {
		return true
	}

	return false
}

// MeasureOptions contains options for MeasureOperationAdvanced
type MeasureOptions struct {
	StartPattern       string  // Pattern to match start marker
	EndPattern         string  // Pattern to match end marker
	StartAfterMs       float64 // Only consider markers after this time
	EndBeforeMs        float64 // Only consider markers before this time
	FindLast           bool    // If true, find the LAST matching end marker
	StartMinDurationMs float64 // Only match start markers with duration >= this value
	EndMinDurationMs   float64 // Only match end markers with duration >= this value
}

// MeasureOperation finds start/end markers matching patterns and returns duration
func MeasureOperation(profile *parser.Profile, startPattern, endPattern string, startAfterMs, endBeforeMs float64) (*OperationMeasurement, error) {
	return MeasureOperationAdvanced(profile, MeasureOptions{
		StartPattern: startPattern,
		EndPattern:   endPattern,
		StartAfterMs: startAfterMs,
		EndBeforeMs:  endBeforeMs,
		FindLast:     false,
	})
}

// MeasureOperationLast finds the first start marker and LAST end marker matching patterns
func MeasureOperationLast(profile *parser.Profile, startPattern, endPattern string, startAfterMs, endBeforeMs float64) (*OperationMeasurement, error) {
	return MeasureOperationAdvanced(profile, MeasureOptions{
		StartPattern: startPattern,
		EndPattern:   endPattern,
		StartAfterMs: startAfterMs,
		EndBeforeMs:  endBeforeMs,
		FindLast:     true,
	})
}

// MeasureOperationWithOptions finds start/end markers with option to find last end marker
// Deprecated: Use MeasureOperationAdvanced for new code
func MeasureOperationWithOptions(profile *parser.Profile, startPattern, endPattern string, startAfterMs, endBeforeMs float64, findLast bool) (*OperationMeasurement, error) {
	return MeasureOperationAdvanced(profile, MeasureOptions{
		StartPattern: startPattern,
		EndPattern:   endPattern,
		StartAfterMs: startAfterMs,
		EndBeforeMs:  endBeforeMs,
		FindLast:     findLast,
	})
}

// MeasureOperationAdvanced finds start/end markers with full control over matching options
func MeasureOperationAdvanced(profile *parser.Profile, opts MeasureOptions) (*OperationMeasurement, error) {
	allMarkers := GetDelimiterMarkers(profile, nil)

	if len(allMarkers) == 0 {
		return nil, fmt.Errorf("no delimiter markers found in profile")
	}

	// Find start marker
	var startMarker *DelimiterMarker
	for i := range allMarkers {
		m := &allMarkers[i]
		if opts.StartAfterMs > 0 && m.TimeMs < opts.StartAfterMs {
			continue
		}
		if opts.StartMinDurationMs > 0 && m.DurationMs < opts.StartMinDurationMs {
			continue
		}
		if MatchMarkerPattern(*m, opts.StartPattern) {
			startMarker = m
			break
		}
	}

	if startMarker == nil {
		return nil, fmt.Errorf("no marker matching start pattern '%s' found", opts.StartPattern)
	}

	// Find end marker (after start marker)
	var endMarker *DelimiterMarker
	for i := range allMarkers {
		m := &allMarkers[i]
		if m.TimeMs <= startMarker.TimeMs {
			continue
		}
		if opts.EndBeforeMs > 0 && m.TimeMs > opts.EndBeforeMs {
			continue
		}
		if opts.EndMinDurationMs > 0 && m.DurationMs < opts.EndMinDurationMs {
			continue
		}
		if MatchMarkerPattern(*m, opts.EndPattern) {
			if opts.FindLast {
				// Keep updating to find the last match
				endMarker = &allMarkers[i]
			} else {
				// Return first match
				endMarker = m
				break
			}
		}
	}

	if endMarker == nil {
		return nil, fmt.Errorf("no marker matching end pattern '%s' found after start marker", opts.EndPattern)
	}

	return &OperationMeasurement{
		StartMarker:     *startMarker,
		EndMarker:       *endMarker,
		OperationTimeMs: endMarker.TimeMs - startMarker.TimeMs,
	}, nil
}

// MeasureOperationByIndex finds start/end markers by index and returns duration
// This is useful when the LLM has identified specific markers by their index
func MeasureOperationByIndex(profile *parser.Profile, startIndex, endIndex int) (*OperationMeasurement, error) {
	allMarkers := GetDelimiterMarkers(profile, nil)

	if startIndex < 0 || startIndex >= len(allMarkers) {
		return nil, fmt.Errorf("start index %d out of range (0-%d)", startIndex, len(allMarkers)-1)
	}
	if endIndex < 0 || endIndex >= len(allMarkers) {
		return nil, fmt.Errorf("end index %d out of range (0-%d)", endIndex, len(allMarkers)-1)
	}
	if endIndex <= startIndex {
		return nil, fmt.Errorf("end index %d must be greater than start index %d", endIndex, startIndex)
	}

	startMarker := allMarkers[startIndex]
	endMarker := allMarkers[endIndex]

	return &OperationMeasurement{
		StartMarker:     startMarker,
		EndMarker:       endMarker,
		OperationTimeMs: endMarker.TimeMs - startMarker.TimeMs,
	}, nil
}

// MatchMarkerPattern checks if a marker matches a pattern like "DOMEvent:click"
// Pattern formats:
//   - "DOMEvent" - matches any marker with type or name "DOMEvent"
//   - "DOMEvent:click" - matches DOMEvent with data.type="click"
//   - "UserTiming:decrypt-start" - matches UserTiming with name containing "decrypt-start"
func MatchMarkerPattern(marker DelimiterMarker, pattern string) bool {
	parts := strings.SplitN(pattern, ":", 2)
	mainType := parts[0]

	// Match main type against Type or Name
	typeMatch := strings.EqualFold(marker.Type, mainType) || strings.EqualFold(marker.Name, mainType)
	if !typeMatch {
		return false
	}

	// If no subtype, we're done
	if len(parts) == 1 {
		return true
	}

	subtype := strings.ToLower(parts[1])

	// Check data.type field (common for DOMEvent)
	if dataType, ok := marker.Data["type"].(string); ok {
		if strings.EqualFold(dataType, subtype) {
			return true
		}
	}

	// Check data.eventType field
	if eventType, ok := marker.Data["eventType"].(string); ok {
		if strings.EqualFold(eventType, subtype) {
			return true
		}
	}

	// Check marker name contains subtype
	if strings.Contains(strings.ToLower(marker.Name), subtype) {
		return true
	}

	// Check data.name field (for UserTiming)
	if dataName, ok := marker.Data["name"].(string); ok {
		if strings.Contains(strings.ToLower(dataName), subtype) {
			return true
		}
	}

	return false
}

// GetDelimiterMarkersReport generates a human-readable report of delimiter markers
type DelimiterMarkersReport struct {
	TotalCount int               `json:"total_count"`
	Markers    []DelimiterMarker `json:"markers"`
	ByType     map[string]int    `json:"by_type"`
	ByCategory map[string]int    `json:"by_category"`
}

func GetDelimiterMarkersReport(profile *parser.Profile, categories []string, limit int) *DelimiterMarkersReport {
	markers := GetDelimiterMarkers(profile, categories)

	report := &DelimiterMarkersReport{
		TotalCount: len(markers),
		ByType:     make(map[string]int),
		ByCategory: make(map[string]int),
	}

	for _, m := range markers {
		report.ByType[m.Type]++
		report.ByCategory[m.Category]++
	}

	// Apply limit
	if limit > 0 && len(markers) > limit {
		markers = markers[:limit]
	}

	report.Markers = markers

	return report
}
