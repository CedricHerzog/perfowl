package parser

import (
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// ChromeProfile represents a Chrome DevTools Performance trace
type ChromeProfile struct {
	Metadata    ChromeMetadata `json:"metadata"`
	TraceEvents []ChromeEvent  `json:"traceEvents"`
}

// ChromeMetadata contains profile metadata
type ChromeMetadata struct {
	EnhancedTraceVersion int     `json:"enhancedTraceVersion"`
	Source               string  `json:"source"`
	StartTime            string  `json:"startTime"`
	DataOrigin           string  `json:"dataOrigin"`
	HostDPR              float64 `json:"hostDPR"`
	SourceMaps           []any   `json:"sourceMaps"`
	Resources            []any   `json:"resources"`
	Modifications        any     `json:"modifications"` // Can be object or array
}

// ChromeEvent represents a single trace event
type ChromeEvent struct {
	Name  string          `json:"name"`            // Event name
	Cat   string          `json:"cat"`             // Category (comma-separated)
	Ph    string          `json:"ph"`              // Phase: B/E/X/M/I/P/etc.
	Ts    float64         `json:"ts"`              // Timestamp (microseconds)
	Dur   float64         `json:"dur,omitempty"`   // Duration (for X events, microseconds)
	TDur  float64         `json:"tdur,omitempty"`  // Thread clock duration
	Pid   int             `json:"pid"`             // Process ID
	Tid   int             `json:"tid"`             // Thread ID
	Tts   float64         `json:"tts,omitempty"`   // Thread timestamp
	Args  json.RawMessage `json:"args,omitempty"`  // Event-specific data
	ID    any             `json:"id,omitempty"`    // Event ID (for async events, can be string or number)
	Scope string          `json:"scope,omitempty"` // Event scope
	Bp    string          `json:"bp,omitempty"`    // Bind point
}

// Chrome event phase constants
const (
	PhaseBegin      = "B" // Duration event begin
	PhaseEnd        = "E" // Duration event end
	PhaseDuration   = "X" // Complete duration event
	PhaseMetadata   = "M" // Metadata event
	PhaseInstant    = "I" // Instant event
	PhaseCounter    = "C" // Counter event
	PhaseAsyncStart = "S" // Async event start (deprecated, use b)
	PhaseAsyncEnd   = "F" // Async event end (deprecated, use e)
	PhaseAsyncBegin = "b" // Async nestable begin
	PhaseAsyncEnd2  = "e" // Async nestable end
	PhaseAsyncStep  = "n" // Async nestable step
	PhaseFlowStart  = "s" // Flow event start
	PhaseFlowEnd    = "f" // Flow event end
	PhaseSample     = "P" // Sample event (V8 profiler)
	PhaseObject     = "O" // Object snapshot
	PhaseCreate     = "N" // Object created
	PhaseDestroy    = "D" // Object destroyed
	PhaseMark       = "R" // Mark event
)

// V8CPUProfile represents V8's CPU profile data embedded in ProfileChunk events
type V8CPUProfile struct {
	Nodes      []V8Node `json:"nodes"`
	Samples    []int    `json:"samples"`
	TimeDeltas []int    `json:"timeDeltas,omitempty"`
	StartTime  int64    `json:"startTime,omitempty"`
	EndTime    int64    `json:"endTime,omitempty"`
}

// V8Node represents a node in the V8 CPU profile tree
type V8Node struct {
	ID        int         `json:"id"`
	CallFrame V8CallFrame `json:"callFrame"`
	HitCount  int         `json:"hitCount,omitempty"`
	Children  []int       `json:"children,omitempty"`
	Parent    int         `json:"parent,omitempty"`
}

// V8CallFrame represents a call frame in V8 profile
type V8CallFrame struct {
	FunctionName string `json:"functionName"`
	ScriptID     any    `json:"scriptId"` // Can be int or string
	URL          string `json:"url,omitempty"`
	LineNumber   int    `json:"lineNumber,omitempty"`
	ColumnNumber int    `json:"columnNumber,omitempty"`
	CodeType     string `json:"codeType,omitempty"`
}

// ProfileChunkArgs represents args for ProfileChunk events
type ProfileChunkArgs struct {
	Data ProfileChunkData `json:"data"`
}

// ProfileChunkData represents the data field in ProfileChunk args
type ProfileChunkData struct {
	CPUProfile V8CPUProfile `json:"cpuProfile"`
	TimeDeltas []int        `json:"timeDeltas,omitempty"`
}

// ThreadNameArgs represents args for thread_name metadata events
type ThreadNameArgs struct {
	Name string `json:"name"`
}

// ProcessNameArgs represents args for process_name metadata events
type ProcessNameArgs struct {
	Name string `json:"name"`
}

// LoadChromeProfile loads a Chrome DevTools Performance trace
func LoadChromeProfile(path string) (*ChromeProfile, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open profile: %w", err)
	}
	defer file.Close()

	var reader io.Reader = file

	// Check if gzip compressed
	ext := strings.ToLower(filepath.Ext(path))
	if ext == ".gz" || ext == ".gzip" {
		gzReader, err := gzip.NewReader(file)
		if err != nil {
			return nil, fmt.Errorf("failed to create gzip reader: %w", err)
		}
		defer gzReader.Close()
		reader = gzReader
	}

	var profile ChromeProfile
	decoder := json.NewDecoder(reader)
	if err := decoder.Decode(&profile); err != nil {
		return nil, fmt.Errorf("failed to decode Chrome profile JSON: %w", err)
	}

	return &profile, nil
}
