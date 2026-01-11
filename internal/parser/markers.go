package parser

import (
	"encoding/json"
	"time"
)

// MarkerType represents known marker types
type MarkerType string

const (
	MarkerTypeGCMajor            MarkerType = "GCMajor"
	MarkerTypeGCMinor            MarkerType = "GCMinor"
	MarkerTypeGCSlice            MarkerType = "GCSlice"
	MarkerTypeCC                 MarkerType = "CC"
	MarkerTypeCCSlice            MarkerType = "CCSlice"
	MarkerTypeDOMEvent           MarkerType = "DOMEvent"
	MarkerTypeStyles             MarkerType = "Styles"
	MarkerTypeUserTiming         MarkerType = "UserTiming"
	MarkerTypeMainThreadLongTask MarkerType = "MainThreadLongTask"
	MarkerTypeTracing            MarkerType = "tracing"
	MarkerTypeChannelMarker      MarkerType = "ChannelMarker"
	MarkerTypeHostResolver       MarkerType = "HostResolver"
	MarkerTypeJSActorMessage     MarkerType = "JSActorMessage"
	MarkerTypeFrameMessage       MarkerType = "FrameMessage"
	MarkerTypeAwake              MarkerType = "Awake"
	MarkerTypeText               MarkerType = "Text"
	MarkerTypePreference         MarkerType = "Preference"
	MarkerTypeIPC                MarkerType = "IPC"
)

// ParsedMarker represents a parsed marker with type-specific data
type ParsedMarker struct {
	Name       string
	Type       MarkerType
	Category   string
	StartTime  float64
	EndTime    float64
	Duration   float64
	Phase      int
	Data       map[string]interface{}
	ThreadName string
	ThreadPID  string
}

// IsDuration returns true if this marker spans a duration
func (m *ParsedMarker) IsDuration() bool {
	return m.Duration > 0
}

// DurationMs returns the duration in milliseconds
func (m *ParsedMarker) DurationMs() float64 {
	return m.Duration
}

// DurationTime returns the duration as a time.Duration
func (m *ParsedMarker) DurationTime() time.Duration {
	return time.Duration(m.Duration * float64(time.Millisecond))
}

// ExtractMarkers extracts and parses all markers from a thread
func ExtractMarkers(thread *Thread, categories []Category) []ParsedMarker {
	markers := make([]ParsedMarker, 0, thread.Markers.Length)

	for i := 0; i < thread.Markers.Length; i++ {
		marker := ParsedMarker{
			Phase:      getPhase(thread.Markers.Phase, i),
			ThreadName: thread.Name,
			ThreadPID:  thread.PID.String(),
		}

		// Get marker name from string array
		if nameIdx := getInt(thread.Markers.Name, i); nameIdx >= 0 && nameIdx < len(thread.StringArray) {
			marker.Name = thread.StringArray[nameIdx]
		}

		// Get category
		if catIdx := getInt(thread.Markers.Category, i); catIdx >= 0 && catIdx < len(categories) {
			marker.Category = categories[catIdx].Name
		}

		// Get timing
		marker.StartTime = getFloat(thread.Markers.StartTime, i)
		if endTime := getEndTime(thread.Markers.EndTime, i); endTime != nil {
			marker.EndTime = *endTime
			// Only calculate positive durations
			if marker.EndTime > marker.StartTime {
				marker.Duration = marker.EndTime - marker.StartTime
			}
		}

		// Parse marker data
		if i < len(thread.Markers.Data) && thread.Markers.Data[i] != nil {
			var data map[string]interface{}
			if err := json.Unmarshal(thread.Markers.Data[i], &data); err == nil {
				marker.Data = data
				// Extract type from data if present
				if typeStr, ok := data["type"].(string); ok {
					marker.Type = MarkerType(typeStr)
				}
			}
		}

		// Infer type from name if not set
		if marker.Type == "" {
			marker.Type = inferMarkerType(marker.Name)
		}

		markers = append(markers, marker)
	}

	return markers
}

// FilterMarkersByType returns markers matching the given type
func FilterMarkersByType(markers []ParsedMarker, markerType MarkerType) []ParsedMarker {
	var filtered []ParsedMarker
	for _, m := range markers {
		if m.Type == markerType || MarkerType(m.Name) == markerType {
			filtered = append(filtered, m)
		}
	}
	return filtered
}

// FilterMarkersByCategory returns markers matching the given category
func FilterMarkersByCategory(markers []ParsedMarker, category string) []ParsedMarker {
	var filtered []ParsedMarker
	for _, m := range markers {
		if m.Category == category {
			filtered = append(filtered, m)
		}
	}
	return filtered
}

// FilterMarkersByDuration returns markers with duration >= minMs
func FilterMarkersByDuration(markers []ParsedMarker, minMs float64) []ParsedMarker {
	var filtered []ParsedMarker
	for _, m := range markers {
		if m.Duration >= minMs {
			filtered = append(filtered, m)
		}
	}
	return filtered
}

// GetMarkerStats returns statistics about markers
type MarkerStats struct {
	TotalCount    int
	ByType        map[string]int
	ByCategory    map[string]int
	TotalDuration float64
	AvgDuration   float64
	MaxDuration   float64
	MinDuration   float64
}

func GetMarkerStats(markers []ParsedMarker) MarkerStats {
	stats := MarkerStats{
		TotalCount:  len(markers),
		ByType:      make(map[string]int),
		ByCategory:  make(map[string]int),
		MinDuration: -1,
	}

	for _, m := range markers {
		stats.ByType[string(m.Type)]++
		stats.ByCategory[m.Category]++

		if m.Duration > 0 {
			stats.TotalDuration += m.Duration
			if m.Duration > stats.MaxDuration {
				stats.MaxDuration = m.Duration
			}
			if stats.MinDuration < 0 || m.Duration < stats.MinDuration {
				stats.MinDuration = m.Duration
			}
		}
	}

	if stats.TotalCount > 0 {
		stats.AvgDuration = stats.TotalDuration / float64(stats.TotalCount)
	}

	return stats
}

// Helper functions

func getInt(slice []int, index int) int {
	if index >= 0 && index < len(slice) {
		return slice[index]
	}
	return -1
}

func getFloat(slice []float64, index int) float64 {
	if index >= 0 && index < len(slice) {
		return slice[index]
	}
	return 0
}

func getPhase(slice []int, index int) int {
	if index >= 0 && index < len(slice) {
		return slice[index]
	}
	return 0
}

func getEndTime(slice []interface{}, index int) *float64 {
	if index >= 0 && index < len(slice) {
		if val, ok := slice[index].(float64); ok {
			return &val
		}
	}
	return nil
}

func inferMarkerType(name string) MarkerType {
	switch name {
	case "GCMajor":
		return MarkerTypeGCMajor
	case "GCMinor":
		return MarkerTypeGCMinor
	case "GCSlice":
		return MarkerTypeGCSlice
	case "CC":
		return MarkerTypeCC
	case "CCSlice":
		return MarkerTypeCCSlice
	case "DOMEvent":
		return MarkerTypeDOMEvent
	case "Styles":
		return MarkerTypeStyles
	case "UserTiming":
		return MarkerTypeUserTiming
	case "MainThreadLongTask":
		return MarkerTypeMainThreadLongTask
	case "tracing":
		return MarkerTypeTracing
	case "ChannelMarker":
		return MarkerTypeChannelMarker
	case "HostResolver":
		return MarkerTypeHostResolver
	case "JSActorMessage":
		return MarkerTypeJSActorMessage
	case "FrameMessage":
		return MarkerTypeFrameMessage
	case "Awake":
		return MarkerTypeAwake
	default:
		return MarkerType(name)
	}
}
