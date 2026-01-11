package parser

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"time"
)

// chromeCategoryMap maps Chrome categories to Firefox-style categories
var chromeCategoryMap = map[string]string{
	"devtools.timeline":                           "JavaScript",
	"disabled-by-default-devtools.timeline":       "Other",
	"disabled-by-default-devtools.timeline.frame": "Graphics",
	"disabled-by-default-devtools.timeline.stack": "JavaScript",
	"v8":                                  "JavaScript",
	"v8.execute":                          "JavaScript",
	"v8.compile":                          "JavaScript",
	"disabled-by-default-v8.gc":           "GC / CC",
	"disabled-by-default-v8.cpu_profiler": "JavaScript",
	"blink":                               "Layout",
	"blink.user_timing":                   "UserTiming",
	"blink.console":                       "JavaScript",
	"loading":                             "Network",
	"net":                                 "Network",
	"netlog":                              "Network",
	"gpu":                                 "Graphics",
	"cc":                                  "Graphics",
	"viz":                                 "Graphics",
	"benchmark":                           "Other",
	"rail":                                "Other",
	"__metadata":                          "Other",
	"toplevel":                            "Other",
	"ipc":                                 "IPC",
}

// chromeConverter handles conversion of Chrome profiles to Firefox format
type chromeConverter struct {
	chrome       *ChromeProfile
	threads      map[string]*threadBuilder
	processNames map[int]string
	stringMap    map[string]int
	stringArray  []string
	categories   []Category
	categoryMap  map[string]int

	// Track start/end times
	minTime float64
	maxTime float64

	// Profile ID to target thread mapping (from Profile events)
	// Profile events (ph="P") indicate which thread is being profiled
	// ProfileChunk events reference this via their "id" field
	profileTargets map[string]profileTarget

	// Track extensions discovered from chrome-extension:// URLs
	extensions map[string]bool // extension ID -> seen
}

// profileTarget stores the thread being profiled by a V8 CPU profile session
type profileTarget struct {
	pid int
	tid int
}

// threadBuilder accumulates data for a single thread
type threadBuilder struct {
	name        string
	pid         int
	tid         int
	processName string

	// Markers (from duration events)
	markerStartTimes []float64
	markerEndTimes   []any
	markerNames      []int
	markerCategories []int
	markerData       []json.RawMessage
	markerPhases     []int

	// Samples (from V8 ProfileChunk)
	sampleStacks    []int
	sampleTimes     []float64
	sampleWeights   []int
	sampleCPUDeltas []int

	// Stack/Frame/Func tables
	stackFrames     []int
	stackCategories []int
	stackPrefixes   []int

	frameAddresses    []any
	frameInlineDepths []int
	frameCategories   []int
	frameSubcats      []int
	frameFuncs        []int
	frameSymbols      []any
	frameWindowIDs    []any
	frameImpls        []any
	frameLines        []any
	frameCols         []any

	funcNames       []int
	funcIsJS        []bool
	funcRelevant    []bool
	funcResources   []int
	funcFileNames   []int
	funcLineNumbers []int
	funcColNumbers  []int

	// Helper maps for deduplication
	funcMap  map[string]int // key: "name|url|line" -> funcIndex
	frameMap map[string]int // key: "funcIdx|category" -> frameIndex
	stackMap map[string]int // key: "frameIdx|prefixIdx" -> stackIndex
}

// ConvertChromeToProfile converts a Chrome trace to Firefox Profile structure
func ConvertChromeToProfile(chrome *ChromeProfile) (*Profile, error) {
	c := &chromeConverter{
		chrome:         chrome,
		threads:        make(map[string]*threadBuilder),
		processNames:   make(map[int]string),
		stringMap:      make(map[string]int),
		stringArray:    make([]string, 0),
		categories:     defaultCategories(),
		categoryMap:    make(map[string]int),
		minTime:        -1,
		maxTime:        0,
		profileTargets: make(map[string]profileTarget),
		extensions:     make(map[string]bool),
	}

	// Build category lookup
	for i, cat := range c.categories {
		c.categoryMap[cat.Name] = i
	}

	return c.convert()
}

func defaultCategories() []Category {
	return []Category{
		{Name: "Idle", Color: "transparent", Subcategories: []string{"Other"}},
		{Name: "Other", Color: "grey", Subcategories: []string{"Other"}},
		{Name: "Layout", Color: "purple", Subcategories: []string{"Other"}},
		{Name: "JavaScript", Color: "yellow", Subcategories: []string{"Other"}},
		{Name: "GC / CC", Color: "orange", Subcategories: []string{"Other"}},
		{Name: "Network", Color: "lightblue", Subcategories: []string{"Other"}},
		{Name: "Graphics", Color: "green", Subcategories: []string{"Other"}},
		{Name: "DOM", Color: "blue", Subcategories: []string{"Other"}},
		{Name: "UserTiming", Color: "yellow", Subcategories: []string{"Other"}},
		{Name: "IPC", Color: "lightgreen", Subcategories: []string{"Other"}},
	}
}

func (c *chromeConverter) convert() (*Profile, error) {
	// Phase 1: Extract metadata events (thread/process names)
	c.parseMetadataEvents()

	// Phase 2: Process all events to build markers
	c.processEvents()

	// Phase 3: Build Profile id -> thread mapping from Profile events
	c.buildProfileTargetMapping()

	// Phase 4: Extract V8 CPU profiles using the target mapping
	c.extractCPUProfiles()

	// Phase 5: Assemble final Profile
	return c.buildProfile()
}

func (c *chromeConverter) parseMetadataEvents() {
	for _, evt := range c.chrome.TraceEvents {
		if evt.Ph != PhaseMetadata {
			continue
		}

		switch evt.Name {
		case "thread_name":
			var args ThreadNameArgs
			if err := json.Unmarshal(evt.Args, &args); err == nil && args.Name != "" {
				tb := c.getOrCreateThread(evt.Pid, evt.Tid)
				tb.name = args.Name
			}

		case "process_name":
			var args ProcessNameArgs
			if err := json.Unmarshal(evt.Args, &args); err == nil && args.Name != "" {
				c.processNames[evt.Pid] = args.Name
			}
		}
	}
}

func (c *chromeConverter) processEvents() {
	// First pass: find TracingStartedInBrowser marker for accurate start time
	// Chrome profiles often include events from before the actual recording started
	foundTracingStart := false
	for _, evt := range c.chrome.TraceEvents {
		if evt.Name == "TracingStartedInBrowser" && evt.Ts > 0 {
			c.minTime = evt.Ts
			foundTracingStart = true
			break
		}
	}

	for _, evt := range c.chrome.TraceEvents {
		// Track time range (skip metadata events with ts=0)
		if evt.Ts > 0 {
			// Only update minTime from events if TracingStartedInBrowser wasn't found
			if !foundTracingStart && (c.minTime < 0 || evt.Ts < c.minTime) {
				c.minTime = evt.Ts
			}
			endTs := evt.Ts + evt.Dur
			if endTs > c.maxTime {
				c.maxTime = endTs
			}
		}

		switch evt.Ph {
		case PhaseDuration: // X - complete duration event
			c.handleDurationEvent(&evt)
		case PhaseInstant: // I - instant event
			c.handleInstantEvent(&evt)
		case PhaseMark: // R - mark event
			c.handleMarkEvent(&evt)
		}
	}
}

func (c *chromeConverter) handleDurationEvent(evt *ChromeEvent) {
	tb := c.getOrCreateThread(evt.Pid, evt.Tid)

	// Convert timestamps from microseconds to milliseconds
	startTime := (evt.Ts - c.minTime) / 1000.0
	duration := evt.Dur / 1000.0
	endTime := startTime + duration

	nameIdx := c.internString(evt.Name)
	catIdx := c.mapCategory(evt.Cat)

	tb.markerStartTimes = append(tb.markerStartTimes, startTime)
	tb.markerEndTimes = append(tb.markerEndTimes, endTime)
	tb.markerNames = append(tb.markerNames, nameIdx)
	tb.markerCategories = append(tb.markerCategories, catIdx)
	tb.markerPhases = append(tb.markerPhases, 1) // IntervalStart

	// Store args as marker data
	if len(evt.Args) > 0 && string(evt.Args) != "{}" {
		tb.markerData = append(tb.markerData, evt.Args)
	} else {
		tb.markerData = append(tb.markerData, nil)
	}
}

func (c *chromeConverter) handleInstantEvent(evt *ChromeEvent) {
	tb := c.getOrCreateThread(evt.Pid, evt.Tid)

	startTime := (evt.Ts - c.minTime) / 1000.0
	nameIdx := c.internString(evt.Name)
	catIdx := c.mapCategory(evt.Cat)

	tb.markerStartTimes = append(tb.markerStartTimes, startTime)
	tb.markerEndTimes = append(tb.markerEndTimes, nil) // Instant events have no end time
	tb.markerNames = append(tb.markerNames, nameIdx)
	tb.markerCategories = append(tb.markerCategories, catIdx)
	tb.markerPhases = append(tb.markerPhases, 0) // Instant

	if len(evt.Args) > 0 && string(evt.Args) != "{}" {
		tb.markerData = append(tb.markerData, evt.Args)
	} else {
		tb.markerData = append(tb.markerData, nil)
	}
}

func (c *chromeConverter) handleMarkEvent(evt *ChromeEvent) {
	// Mark events are similar to instant events
	c.handleInstantEvent(evt)
}

// buildProfileTargetMapping extracts Profile events (phase "P") which indicate
// which thread is being profiled. ProfileChunk events reference these via their "id" field.
func (c *chromeConverter) buildProfileTargetMapping() {
	for _, evt := range c.chrome.TraceEvents {
		// Profile events have phase "P" and name "Profile"
		if evt.Ph != PhaseSample || evt.Name != "Profile" {
			continue
		}

		// Get the profile ID (can be string like "0x1" or number)
		idStr := c.eventIDToString(evt.ID)
		if idStr == "" {
			continue
		}

		// The tid/pid on the Profile event is the thread being profiled
		c.profileTargets[idStr] = profileTarget{
			pid: evt.Pid,
			tid: evt.Tid,
		}
	}
}

// eventIDToString converts an event ID (which can be string or number) to string
func (c *chromeConverter) eventIDToString(id any) string {
	if id == nil {
		return ""
	}
	switch v := id.(type) {
	case string:
		return v
	case float64:
		return fmt.Sprintf("%d", int(v))
	default:
		return fmt.Sprintf("%v", v)
	}
}

func (c *chromeConverter) extractCPUProfiles() {
	for _, evt := range c.chrome.TraceEvents {
		// Only process ProfileChunk events (Profile events were handled in buildProfileTargetMapping)
		if evt.Name != "ProfileChunk" {
			continue
		}

		var args ProfileChunkArgs
		if err := json.Unmarshal(evt.Args, &args); err != nil {
			continue
		}

		cpuProfile := args.Data.CPUProfile
		if len(cpuProfile.Nodes) == 0 && len(cpuProfile.Samples) == 0 {
			continue
		}

		// Look up the target thread from the Profile id mapping
		// ProfileChunk events are emitted by v8:ProfEvntProc threads, but the samples
		// belong to the thread specified by the corresponding Profile event with matching id
		idStr := c.eventIDToString(evt.ID)
		targetPid, targetTid := evt.Pid, evt.Tid // Default to event's pid/tid
		if target, ok := c.profileTargets[idStr]; ok {
			targetPid = target.pid
			targetTid = target.tid
		}

		tb := c.getOrCreateThread(targetPid, targetTid)
		c.processCPUProfile(tb, &cpuProfile, &args.Data, evt.Ts)
	}
}

func (c *chromeConverter) processCPUProfile(tb *threadBuilder, cp *V8CPUProfile, data *ProfileChunkData, baseTs float64) {
	if tb.funcMap == nil {
		tb.funcMap = make(map[string]int)
		tb.frameMap = make(map[string]int)
		tb.stackMap = make(map[string]int)
	}

	// Build node ID to stack index mapping
	nodeToStack := make(map[int]int)

	// Process nodes in order (parent before children due to tree structure)
	for _, node := range cp.Nodes {
		funcIdx := c.getOrCreateFunc(tb, &node.CallFrame)
		catIdx := c.getCategoryForCallFrame(&node.CallFrame)
		frameIdx := c.getOrCreateFrame(tb, funcIdx, catIdx)

		// Find parent stack index
		prefixIdx := -1
		if node.Parent > 0 {
			if parentStack, ok := nodeToStack[node.Parent]; ok {
				prefixIdx = parentStack
			}
		}

		stackIdx := c.getOrCreateStack(tb, frameIdx, prefixIdx, catIdx)
		nodeToStack[node.ID] = stackIdx
	}

	// Process samples
	timeDeltas := data.TimeDeltas
	if len(timeDeltas) == 0 {
		timeDeltas = cp.TimeDeltas
	}

	currentTime := (baseTs - c.minTime) / 1000.0 // Convert to ms relative to profile start
	for i, nodeID := range cp.Samples {
		stackIdx, ok := nodeToStack[nodeID]
		if !ok {
			stackIdx = -1
		}

		var delta float64
		if i < len(timeDeltas) {
			delta = float64(timeDeltas[i]) / 1000.0 // Convert to ms
		}

		tb.sampleStacks = append(tb.sampleStacks, stackIdx)
		tb.sampleTimes = append(tb.sampleTimes, currentTime)
		tb.sampleWeights = append(tb.sampleWeights, 1)
		tb.sampleCPUDeltas = append(tb.sampleCPUDeltas, int(delta*1000)) // Store as microseconds

		currentTime += delta
	}
}

func (c *chromeConverter) getOrCreateFunc(tb *threadBuilder, cf *V8CallFrame) int {
	url := cf.URL
	if url == "" {
		url = "(unknown)"
	}
	key := fmt.Sprintf("%s|%s|%d", cf.FunctionName, url, cf.LineNumber)

	if idx, ok := tb.funcMap[key]; ok {
		return idx
	}

	idx := len(tb.funcNames)
	tb.funcMap[key] = idx

	nameIdx := c.internString(cf.FunctionName)
	fileIdx := c.internString(url)

	tb.funcNames = append(tb.funcNames, nameIdx)
	tb.funcIsJS = append(tb.funcIsJS, true)
	tb.funcRelevant = append(tb.funcRelevant, true)
	tb.funcResources = append(tb.funcResources, -1)
	tb.funcFileNames = append(tb.funcFileNames, fileIdx)
	tb.funcLineNumbers = append(tb.funcLineNumbers, cf.LineNumber)
	tb.funcColNumbers = append(tb.funcColNumbers, cf.ColumnNumber)

	return idx
}

func (c *chromeConverter) getOrCreateFrame(tb *threadBuilder, funcIdx, catIdx int) int {
	key := fmt.Sprintf("%d|%d", funcIdx, catIdx)

	if idx, ok := tb.frameMap[key]; ok {
		return idx
	}

	idx := len(tb.frameFuncs)
	tb.frameMap[key] = idx

	tb.frameAddresses = append(tb.frameAddresses, nil)
	tb.frameInlineDepths = append(tb.frameInlineDepths, 0)
	tb.frameCategories = append(tb.frameCategories, catIdx)
	tb.frameSubcats = append(tb.frameSubcats, 0)
	tb.frameFuncs = append(tb.frameFuncs, funcIdx)
	tb.frameSymbols = append(tb.frameSymbols, nil)
	tb.frameWindowIDs = append(tb.frameWindowIDs, nil)
	tb.frameImpls = append(tb.frameImpls, nil)
	tb.frameLines = append(tb.frameLines, nil)
	tb.frameCols = append(tb.frameCols, nil)

	return idx
}

func (c *chromeConverter) getOrCreateStack(tb *threadBuilder, frameIdx, prefixIdx, catIdx int) int {
	key := fmt.Sprintf("%d|%d", frameIdx, prefixIdx)

	if idx, ok := tb.stackMap[key]; ok {
		return idx
	}

	idx := len(tb.stackFrames)
	tb.stackMap[key] = idx

	tb.stackFrames = append(tb.stackFrames, frameIdx)
	tb.stackPrefixes = append(tb.stackPrefixes, prefixIdx)
	tb.stackCategories = append(tb.stackCategories, catIdx)

	return idx
}

func (c *chromeConverter) getCategoryForCallFrame(cf *V8CallFrame) int {
	// Determine category based on URL or function name
	url := cf.URL
	if strings.Contains(url, "chrome-extension://") {
		// Extract and track extension ID
		c.extractExtensionID(url)
		return c.categoryMap["Other"]
	}
	if strings.HasPrefix(url, "http") || strings.HasPrefix(url, "file") {
		return c.categoryMap["JavaScript"]
	}
	if cf.CodeType == "other" || cf.FunctionName == "(root)" || cf.FunctionName == "(program)" {
		return c.categoryMap["Other"]
	}
	return c.categoryMap["JavaScript"]
}

// extractExtensionID extracts and tracks extension IDs from chrome-extension:// URLs
func (c *chromeConverter) extractExtensionID(url string) {
	// URL format: chrome-extension://EXTENSION_ID/path/to/file.js
	const prefix = "chrome-extension://"
	if !strings.HasPrefix(url, prefix) {
		return
	}
	rest := url[len(prefix):]
	// Find the end of the extension ID (next slash or end of string)
	slashIdx := strings.Index(rest, "/")
	var extID string
	if slashIdx > 0 {
		extID = rest[:slashIdx]
	} else {
		extID = rest
	}
	if extID != "" && len(extID) == 32 { // Chrome extension IDs are 32 chars
		c.extensions[extID] = true
	}
}

// buildExtensions creates the Extensions struct from discovered extension IDs
func (c *chromeConverter) buildExtensions() Extensions {
	if len(c.extensions) == 0 {
		return Extensions{}
	}

	ids := make([]string, 0, len(c.extensions))
	names := make([]string, 0, len(c.extensions))
	baseURLs := make([]string, 0, len(c.extensions))

	for extID := range c.extensions {
		ids = append(ids, extID)
		names = append(names, extID) // Use ID as name since we don't have the actual name
		baseURLs = append(baseURLs, "chrome-extension://"+extID+"/")
	}

	return Extensions{
		Length:  len(ids),
		ID:      ids,
		Name:    names,
		BaseURL: baseURLs,
	}
}

func (c *chromeConverter) getOrCreateThread(pid, tid int) *threadBuilder {
	key := fmt.Sprintf("%d:%d", pid, tid)
	if tb, ok := c.threads[key]; ok {
		return tb
	}

	tb := &threadBuilder{
		pid: pid,
		tid: tid,
	}
	c.threads[key] = tb
	return tb
}

func (c *chromeConverter) internString(s string) int {
	if idx, ok := c.stringMap[s]; ok {
		return idx
	}
	idx := len(c.stringArray)
	c.stringMap[s] = idx
	c.stringArray = append(c.stringArray, s)
	return idx
}

func (c *chromeConverter) mapCategory(chromeCategories string) int {
	// Chrome categories are comma-separated
	cats := strings.Split(chromeCategories, ",")
	for _, cat := range cats {
		cat = strings.TrimSpace(cat)
		if firefoxCat, ok := chromeCategoryMap[cat]; ok {
			if idx, ok := c.categoryMap[firefoxCat]; ok {
				return idx
			}
		}
	}
	// Default to "Other"
	if idx, ok := c.categoryMap["Other"]; ok {
		return idx
	}
	return 1 // Fallback to index 1 (Other)
}

func (c *chromeConverter) buildProfile() (*Profile, error) {
	// Apply process names to threads
	for _, tb := range c.threads {
		if pname, ok := c.processNames[tb.pid]; ok {
			tb.processName = pname
		}
	}

	// Sort threads for consistent ordering
	threadKeys := make([]string, 0, len(c.threads))
	for k := range c.threads {
		threadKeys = append(threadKeys, k)
	}
	sort.Strings(threadKeys)

	threads := make([]Thread, 0, len(c.threads))
	for _, key := range threadKeys {
		tb := c.threads[key]
		thread := c.buildThread(tb)
		threads = append(threads, thread)
	}

	// Calculate duration
	duration := (c.maxTime - c.minTime) / 1000.0 // Convert to ms

	// Parse start time from metadata
	var startTimeMs float64
	if c.chrome.Metadata.StartTime != "" {
		if t, err := time.Parse(time.RFC3339, c.chrome.Metadata.StartTime); err == nil {
			startTimeMs = float64(t.UnixMilli())
		}
	}

	// Build extensions metadata from discovered extension IDs
	extensions := c.buildExtensions()

	profile := &Profile{
		Meta: Meta{
			Interval:           1.0, // Default 1ms interval
			StartTime:          startTimeMs,
			ProfilingStartTime: 0,
			ProfilingEndTime:   duration,
			Product:            "Chrome",
			Version:            1,
			Platform:           "Chrome DevTools",
			Categories:         c.categories,
			Extensions:         extensions,
		},
		Threads: threads,
		Shared: Shared{
			StringArray: c.stringArray,
		},
	}

	return profile, nil
}

func (c *chromeConverter) buildThread(tb *threadBuilder) Thread {
	// Determine if this is a main thread
	isMain := strings.Contains(tb.name, "Main") ||
		tb.name == "CrBrowserMain" ||
		tb.name == "CrRendererMain"

	// Determine process type from process name
	processType := "tab"
	if strings.Contains(tb.processName, "Browser") {
		processType = "default"
	} else if strings.Contains(tb.processName, "GPU") {
		processType = "gpu"
	} else if strings.Contains(tb.processName, "Extension") {
		processType = "extension"
	}

	thread := Thread{
		Name:         tb.name,
		IsMainThread: isMain,
		ProcessType:  processType,
		ProcessName:  tb.processName,
		PID:          json.Number(fmt.Sprintf("%d", tb.pid)),
		TID:          json.Number(fmt.Sprintf("%d", tb.tid)),
		StringArray:  c.stringArray,
		Samples: Samples{
			Length:         len(tb.sampleStacks),
			Stack:          tb.sampleStacks,
			Time:           tb.sampleTimes,
			Weight:         tb.sampleWeights,
			WeightType:     "samples",
			ThreadCPUDelta: tb.sampleCPUDeltas,
		},
		Markers: Markers{
			Length:    len(tb.markerNames),
			Category:  tb.markerCategories,
			Data:      tb.markerData,
			EndTime:   toInterfaceSlice(tb.markerEndTimes),
			Name:      tb.markerNames,
			Phase:     tb.markerPhases,
			StartTime: tb.markerStartTimes,
		},
		StackTable: StackTable{
			Length:   len(tb.stackFrames),
			Frame:    tb.stackFrames,
			Category: tb.stackCategories,
			Prefix:   tb.stackPrefixes,
		},
		FrameTable: FrameTable{
			Length:         len(tb.frameFuncs),
			Address:        tb.frameAddresses,
			InlineDepth:    tb.frameInlineDepths,
			Category:       tb.frameCategories,
			Subcategory:    tb.frameSubcats,
			Func:           tb.frameFuncs,
			NativeSymbol:   tb.frameSymbols,
			InnerWindowID:  tb.frameWindowIDs,
			Implementation: tb.frameImpls,
			Line:           tb.frameLines,
			Column:         tb.frameCols,
		},
		FuncTable: FuncTable{
			Length:        len(tb.funcNames),
			Name:          tb.funcNames,
			IsJS:          tb.funcIsJS,
			RelevantForJS: tb.funcRelevant,
			Resource:      tb.funcResources,
			FileName:      tb.funcFileNames,
			LineNumber:    tb.funcLineNumbers,
			ColumnNumber:  tb.funcColNumbers,
		},
		ResourceTable: ResourceTable{
			Length: 0,
		},
		NativeSymbols: NativeSymbols{
			Length: 0,
		},
	}

	return thread
}

func toInterfaceSlice(slice []any) []any {
	if slice == nil {
		return []any{}
	}
	return slice
}
