package parser

import "encoding/json"

// Profile represents the top-level Firefox Profiler JSON structure
type Profile struct {
	Meta    Meta     `json:"meta"`
	Libs    []Lib    `json:"libs"`
	Threads []Thread `json:"threads"`
	Shared  Shared   `json:"shared"`
}

// Shared contains shared data across threads (Firefox profiler optimization)
type Shared struct {
	StringArray []string `json:"stringArray"`
}

// Meta contains profile metadata
type Meta struct {
	Interval                   float64        `json:"interval"`
	StartTime                  float64        `json:"startTime"`
	ProfilingStartTime         float64        `json:"profilingStartTime"`
	ProfilingEndTime           float64        `json:"profilingEndTime"`
	ABI                        string         `json:"abi"`
	OSCPU                      string         `json:"oscpu"`
	Platform                   string         `json:"platform"`
	ProcessType                int            `json:"processType"`
	Product                    string         `json:"product"`
	Version                    int            `json:"version"`
	Stackwalk                  int            `json:"stackwalk"`
	Debug                      bool           `json:"debug"`
	Toolkit                    string         `json:"toolkit"`
	CPUName                    string         `json:"CPUName"`
	PhysicalCPUs               int            `json:"physicalCPUs"`
	LogicalCPUs                int            `json:"logicalCPUs"`
	Symbolicated               bool           `json:"symbolicated"`
	UpdateChannel              string         `json:"updateChannel"`
	AppBuildID                 string         `json:"appBuildID"`
	SourceURL                  string         `json:"sourceURL"`
	PreprocessedProfileVersion int            `json:"preprocessedProfileVersion"`
	Extensions                 Extensions     `json:"extensions"`
	Categories                 []Category     `json:"categories"`
	MarkerSchema               []MarkerSchema `json:"markerSchema"`
	Configuration              Configuration  `json:"configuration"`
	SampleUnits                SampleUnits    `json:"sampleUnits"`
}

// Extensions contains information about installed browser extensions
type Extensions struct {
	Length  int      `json:"length"`
	ID      []string `json:"id"`
	Name    []string `json:"name"`
	BaseURL []string `json:"baseURL"`
}

// Category represents a profiling category (JavaScript, Layout, etc.)
type Category struct {
	Name          string   `json:"name"`
	Color         string   `json:"color"`
	Subcategories []string `json:"subcategories"`
}

// MarkerSchema defines the structure of marker types
type MarkerSchema struct {
	Name         string                   `json:"name"`
	Display      []string                 `json:"display"`
	Fields       []map[string]interface{} `json:"fields"`
	TooltipLabel string                   `json:"tooltipLabel,omitempty"`
	TableLabel   string                   `json:"tableLabel,omitempty"`
	ChartLabel   string                   `json:"chartLabel,omitempty"`
	Description  string                   `json:"description,omitempty"`
	IsStackBased bool                     `json:"isStackBased,omitempty"`
}

// Configuration contains profiling configuration
type Configuration struct {
	Features    []string `json:"features"`
	Threads     []string `json:"threads"`
	Interval    float64  `json:"interval"`
	Capacity    int      `json:"capacity"`
	ActiveTabID int      `json:"activeTabID"`
}

// SampleUnits defines the units for sample data
type SampleUnits struct {
	Time           string `json:"time"`
	EventDelay     string `json:"eventDelay"`
	ThreadCPUDelta string `json:"threadCPUDelta"`
}

// Lib represents a loaded library
type Lib struct {
	Arch       string `json:"arch"`
	Name       string `json:"name"`
	Path       string `json:"path"`
	DebugName  string `json:"debugName"`
	DebugPath  string `json:"debugPath"`
	BreakpadID string `json:"breakpadId"`
	CodeID     string `json:"codeId"`
}

// Thread represents a profiled thread
type Thread struct {
	Name                string        `json:"name"`
	IsMainThread        bool          `json:"isMainThread"`
	ProcessType         string        `json:"processType"`
	ProcessName         string        `json:"processName"`
	ProcessStartupTime  float64       `json:"processStartupTime"`
	ProcessShutdownTime *float64      `json:"processShutdownTime"`
	RegisterTime        float64       `json:"registerTime"`
	UnregisterTime      *float64      `json:"unregisterTime"`
	PID                 json.Number   `json:"pid"`
	TID                 json.Number   `json:"tid"`
	Samples             Samples       `json:"samples"`
	Markers             Markers       `json:"markers"`
	StackTable          StackTable    `json:"stackTable"`
	FrameTable          FrameTable    `json:"frameTable"`
	StringArray         []string      `json:"stringArray"`
	FuncTable           FuncTable     `json:"funcTable"`
	ResourceTable       ResourceTable `json:"resourceTable"`
	NativeSymbols       NativeSymbols `json:"nativeSymbols"`
}

// Samples contains sample data
type Samples struct {
	Length         int       `json:"length"`
	Stack          []int     `json:"stack"`
	Time           []float64 `json:"time"`
	Weight         []int     `json:"weight,omitempty"`
	WeightType     string    `json:"weightType,omitempty"`
	ThreadCPUDelta []int     `json:"threadCPUDelta,omitempty"`
}

// Markers contains marker data
type Markers struct {
	Length    int               `json:"length"`
	Category  []int             `json:"category"`
	Data      []json.RawMessage `json:"data"`
	EndTime   []interface{}     `json:"endTime"`
	Name      []int             `json:"name"`
	Phase     []int             `json:"phase"`
	StartTime []float64         `json:"startTime"`
}

// StackTable contains stack frame information
type StackTable struct {
	Length   int   `json:"length"`
	Frame    []int `json:"frame"`
	Category []int `json:"category"`
	Prefix   []int `json:"prefix"`
}

// FrameTable contains frame information
// Note: Some fields use interface{} to handle large numbers that overflow int64
type FrameTable struct {
	Length         int           `json:"length"`
	Address        []interface{} `json:"address"`
	InlineDepth    []int         `json:"inlineDepth"`
	Category       []int         `json:"category"`
	Subcategory    []int         `json:"subcategory"`
	Func           []int         `json:"func"`
	NativeSymbol   []interface{} `json:"nativeSymbol"`
	InnerWindowID  []interface{} `json:"innerWindowID"`
	Implementation []interface{} `json:"implementation"`
	Line           []interface{} `json:"line"`
	Column         []interface{} `json:"column"`
}

// FuncTable contains function information
type FuncTable struct {
	Length        int    `json:"length"`
	Name          []int  `json:"name"`
	IsJS          []bool `json:"isJS"`
	RelevantForJS []bool `json:"relevantForJS"`
	Resource      []int  `json:"resource"`
	FileName      []int  `json:"fileName"`
	LineNumber    []int  `json:"lineNumber"`
	ColumnNumber  []int  `json:"columnNumber"`
}

// ResourceTable contains resource information
type ResourceTable struct {
	Length int   `json:"length"`
	Lib    []int `json:"lib"`
	Name   []int `json:"name"`
	Host   []int `json:"host"`
	Type   []int `json:"type"`
}

// NativeSymbols contains native symbol information
type NativeSymbols struct {
	Length       int           `json:"length"`
	Address      []interface{} `json:"address"`
	FunctionSize []interface{} `json:"functionSize"`
	LibIndex     []int         `json:"libIndex"`
	Name         []int         `json:"name"`
}

// Duration returns the profile duration in milliseconds
func (p *Profile) Duration() float64 {
	return p.Meta.ProfilingEndTime - p.Meta.ProfilingStartTime
}

// DurationSeconds returns the profile duration in seconds
func (p *Profile) DurationSeconds() float64 {
	return p.Duration() / 1000.0
}

// ExtensionCount returns the number of extensions
func (p *Profile) ExtensionCount() int {
	return p.Meta.Extensions.Length
}

// GetExtensions returns a map of extension ID to name
func (p *Profile) GetExtensions() map[string]string {
	extensions := make(map[string]string)
	for i := 0; i < p.Meta.Extensions.Length; i++ {
		if i < len(p.Meta.Extensions.ID) && i < len(p.Meta.Extensions.Name) {
			extensions[p.Meta.Extensions.ID[i]] = p.Meta.Extensions.Name[i]
		}
	}
	return extensions
}

// GetExtensionBaseURLs returns a map of extension ID to base URL
func (p *Profile) GetExtensionBaseURLs() map[string]string {
	urls := make(map[string]string)
	for i := 0; i < p.Meta.Extensions.Length; i++ {
		if i < len(p.Meta.Extensions.ID) && i < len(p.Meta.Extensions.BaseURL) {
			urls[p.Meta.Extensions.ID[i]] = p.Meta.Extensions.BaseURL[i]
		}
	}
	return urls
}

// GetCategoryByIndex returns a category by its index
func (p *Profile) GetCategoryByIndex(index int) *Category {
	if index >= 0 && index < len(p.Meta.Categories) {
		return &p.Meta.Categories[index]
	}
	return nil
}

// ThreadCount returns the number of threads
func (p *Profile) ThreadCount() int {
	return len(p.Threads)
}

// GetMainThreads returns all main threads
func (p *Profile) GetMainThreads() []Thread {
	var mainThreads []Thread
	for _, t := range p.Threads {
		if t.IsMainThread {
			mainThreads = append(mainThreads, t)
		}
	}
	return mainThreads
}
