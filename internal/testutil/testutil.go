// Package testutil provides test utilities for building profiles and assertions.
package testutil

import (
	"encoding/json"

	"github.com/CedricHerzog/perfowl/internal/parser"
)

// ProfileBuilder provides a fluent API for constructing test profiles.
type ProfileBuilder struct {
	profile *parser.Profile
}

// NewProfileBuilder creates a new ProfileBuilder with sensible defaults.
func NewProfileBuilder() *ProfileBuilder {
	return &ProfileBuilder{
		profile: &parser.Profile{
			Meta: parser.Meta{
				Interval:           1.0,
				StartTime:          0,
				ProfilingStartTime: 0,
				ProfilingEndTime:   1000, // 1 second default
				Product:            "Firefox",
				Version:            1,
				Categories: []parser.Category{
					{Name: "Other", Color: "grey"},
				},
			},
			Threads: []parser.Thread{},
		},
	}
}

// WithMeta sets the profile metadata.
func (b *ProfileBuilder) WithMeta(meta parser.Meta) *ProfileBuilder {
	b.profile.Meta = meta
	return b
}

// WithDuration sets the profile duration in milliseconds.
func (b *ProfileBuilder) WithDuration(ms float64) *ProfileBuilder {
	b.profile.Meta.ProfilingStartTime = 0
	b.profile.Meta.ProfilingEndTime = ms
	return b
}

// WithInterval sets the sampling interval in milliseconds.
func (b *ProfileBuilder) WithInterval(ms float64) *ProfileBuilder {
	b.profile.Meta.Interval = ms
	return b
}

// WithThread adds a thread to the profile.
func (b *ProfileBuilder) WithThread(thread parser.Thread) *ProfileBuilder {
	b.profile.Threads = append(b.profile.Threads, thread)
	return b
}

// WithExtension adds an extension to the profile.
func (b *ProfileBuilder) WithExtension(id, name, baseURL string) *ProfileBuilder {
	b.profile.Meta.Extensions.Length++
	b.profile.Meta.Extensions.ID = append(b.profile.Meta.Extensions.ID, id)
	b.profile.Meta.Extensions.Name = append(b.profile.Meta.Extensions.Name, name)
	b.profile.Meta.Extensions.BaseURL = append(b.profile.Meta.Extensions.BaseURL, baseURL)
	return b
}

// WithCategory adds a category to the profile.
func (b *ProfileBuilder) WithCategory(name, color string) *ProfileBuilder {
	b.profile.Meta.Categories = append(b.profile.Meta.Categories, parser.Category{
		Name:  name,
		Color: color,
	})
	return b
}

// WithCategories sets all categories for the profile.
func (b *ProfileBuilder) WithCategories(categories []parser.Category) *ProfileBuilder {
	b.profile.Meta.Categories = categories
	return b
}

// WithSharedStringArray sets the shared string array.
func (b *ProfileBuilder) WithSharedStringArray(strings []string) *ProfileBuilder {
	b.profile.Shared.StringArray = strings
	return b
}

// Build returns the constructed profile.
func (b *ProfileBuilder) Build() *parser.Profile {
	return b.profile
}

// ThreadBuilder provides a fluent API for constructing test threads.
type ThreadBuilder struct {
	thread parser.Thread
}

// NewThreadBuilder creates a new ThreadBuilder with the given name.
func NewThreadBuilder(name string) *ThreadBuilder {
	return &ThreadBuilder{
		thread: parser.Thread{
			Name:        name,
			ProcessType: "default",
			PID:         "1",
			TID:         "1",
			StringArray: []string{},
		},
	}
}

// AsMainThread marks this thread as a main thread.
func (b *ThreadBuilder) AsMainThread() *ThreadBuilder {
	b.thread.IsMainThread = true
	return b
}

// WithProcessType sets the process type (default, tab, web, etc.).
func (b *ThreadBuilder) WithProcessType(pt string) *ThreadBuilder {
	b.thread.ProcessType = pt
	return b
}

// WithProcessName sets the process name.
func (b *ThreadBuilder) WithProcessName(name string) *ThreadBuilder {
	b.thread.ProcessName = name
	return b
}

// WithPID sets the process ID.
func (b *ThreadBuilder) WithPID(pid string) *ThreadBuilder {
	b.thread.PID = json.Number(pid)
	return b
}

// WithTID sets the thread ID.
func (b *ThreadBuilder) WithTID(tid string) *ThreadBuilder {
	b.thread.TID = json.Number(tid)
	return b
}

// WithSamples sets the thread samples.
func (b *ThreadBuilder) WithSamples(samples parser.Samples) *ThreadBuilder {
	b.thread.Samples = samples
	return b
}

// WithMarkers sets the thread markers.
func (b *ThreadBuilder) WithMarkers(markers parser.Markers) *ThreadBuilder {
	b.thread.Markers = markers
	return b
}

// WithStackTable sets the stack table.
func (b *ThreadBuilder) WithStackTable(st parser.StackTable) *ThreadBuilder {
	b.thread.StackTable = st
	return b
}

// WithFrameTable sets the frame table.
func (b *ThreadBuilder) WithFrameTable(ft parser.FrameTable) *ThreadBuilder {
	b.thread.FrameTable = ft
	return b
}

// WithFuncTable sets the function table.
func (b *ThreadBuilder) WithFuncTable(ft parser.FuncTable) *ThreadBuilder {
	b.thread.FuncTable = ft
	return b
}

// WithResourceTable sets the resource table.
func (b *ThreadBuilder) WithResourceTable(rt parser.ResourceTable) *ThreadBuilder {
	b.thread.ResourceTable = rt
	return b
}

// WithStringArray sets the thread's string array.
func (b *ThreadBuilder) WithStringArray(strings []string) *ThreadBuilder {
	b.thread.StringArray = strings
	return b
}

// Build returns the constructed thread.
func (b *ThreadBuilder) Build() parser.Thread {
	return b.thread
}

// MarkerBuilder provides a fluent API for constructing test markers.
type MarkerBuilder struct {
	markers     parser.Markers
	stringArray []string
	stringMap   map[string]int
}

// NewMarkerBuilder creates a new MarkerBuilder.
func NewMarkerBuilder() *MarkerBuilder {
	return &MarkerBuilder{
		markers: parser.Markers{
			Length:    0,
			Category:  []int{},
			Data:      []json.RawMessage{},
			EndTime:   []interface{}{},
			Name:      []int{},
			Phase:     []int{},
			StartTime: []float64{},
		},
		stringArray: []string{},
		stringMap:   make(map[string]int),
	}
}

// internString adds a string to the string array and returns its index.
func (b *MarkerBuilder) internString(s string) int {
	if idx, ok := b.stringMap[s]; ok {
		return idx
	}
	idx := len(b.stringArray)
	b.stringArray = append(b.stringArray, s)
	b.stringMap[s] = idx
	return idx
}

// addMarker is a helper to add a marker with common fields.
func (b *MarkerBuilder) addMarker(name string, category int, startTime, duration float64, data map[string]interface{}) *MarkerBuilder {
	b.markers.Length++
	b.markers.Name = append(b.markers.Name, b.internString(name))
	b.markers.Category = append(b.markers.Category, category)
	b.markers.StartTime = append(b.markers.StartTime, startTime)
	b.markers.Phase = append(b.markers.Phase, 1) // Interval phase

	if duration > 0 {
		b.markers.EndTime = append(b.markers.EndTime, startTime+duration)
	} else {
		b.markers.EndTime = append(b.markers.EndTime, nil)
	}

	if data != nil {
		dataBytes, _ := json.Marshal(data)
		b.markers.Data = append(b.markers.Data, dataBytes)
	} else {
		b.markers.Data = append(b.markers.Data, nil)
	}

	return b
}

// AddGCMajor adds a major GC marker.
func (b *MarkerBuilder) AddGCMajor(startTime, duration float64) *MarkerBuilder {
	return b.addMarker("GCMajor", 0, startTime, duration, map[string]interface{}{
		"type": "GCMajor",
	})
}

// AddGCMinor adds a minor GC marker.
func (b *MarkerBuilder) AddGCMinor(startTime, duration float64) *MarkerBuilder {
	return b.addMarker("GCMinor", 0, startTime, duration, map[string]interface{}{
		"type": "GCMinor",
	})
}

// AddGCSlice adds a GC slice marker.
func (b *MarkerBuilder) AddGCSlice(startTime, duration float64) *MarkerBuilder {
	return b.addMarker("GCSlice", 0, startTime, duration, map[string]interface{}{
		"type": "GCSlice",
	})
}

// AddLongTask adds a MainThreadLongTask marker.
func (b *MarkerBuilder) AddLongTask(startTime, duration float64) *MarkerBuilder {
	return b.addMarker("MainThreadLongTask", 0, startTime, duration, map[string]interface{}{
		"type": "MainThreadLongTask",
	})
}

// AddDOMEvent adds a DOMEvent marker.
func (b *MarkerBuilder) AddDOMEvent(eventType string, startTime float64) *MarkerBuilder {
	return b.addMarker("DOMEvent", 0, startTime, 0, map[string]interface{}{
		"type":      "DOMEvent",
		"eventType": eventType,
	})
}

// AddDOMEventWithDuration adds a DOMEvent marker with duration.
func (b *MarkerBuilder) AddDOMEventWithDuration(eventType string, startTime, duration float64) *MarkerBuilder {
	return b.addMarker("DOMEvent", 0, startTime, duration, map[string]interface{}{
		"type":      "DOMEvent",
		"eventType": eventType,
	})
}

// AddIPC adds an IPC marker.
func (b *MarkerBuilder) AddIPC(startTime, duration float64, sync bool) *MarkerBuilder {
	return b.addMarker("IPC", 0, startTime, duration, map[string]interface{}{
		"type": "IPC",
		"sync": sync,
	})
}

// AddLayout adds a layout/reflow marker.
func (b *MarkerBuilder) AddLayout(startTime, duration float64) *MarkerBuilder {
	return b.addMarker("Reflow", 0, startTime, duration, map[string]interface{}{
		"type": "Reflow",
	})
}

// AddStyles adds a Styles marker.
func (b *MarkerBuilder) AddStyles(startTime, duration float64) *MarkerBuilder {
	return b.addMarker("Styles", 0, startTime, duration, map[string]interface{}{
		"type": "Styles",
	})
}

// AddNetwork adds a network marker.
func (b *MarkerBuilder) AddNetwork(url string, startTime, duration float64) *MarkerBuilder {
	return b.addMarker("Network", 0, startTime, duration, map[string]interface{}{
		"type": "Network",
		"URI":  url,
	})
}

// AddUserTiming adds a user timing marker.
func (b *MarkerBuilder) AddUserTiming(name string, startTime float64) *MarkerBuilder {
	return b.addMarker(name, 0, startTime, 0, map[string]interface{}{
		"type": "UserTiming",
	})
}

// AddUserTimingWithDuration adds a user timing marker with duration.
func (b *MarkerBuilder) AddUserTimingWithDuration(name string, startTime, duration float64) *MarkerBuilder {
	return b.addMarker(name, 0, startTime, duration, map[string]interface{}{
		"type": "UserTiming",
	})
}

// AddAwake adds an Awake marker.
func (b *MarkerBuilder) AddAwake(startTime float64) *MarkerBuilder {
	return b.addMarker("Awake", 0, startTime, 0, map[string]interface{}{
		"type": "Awake",
	})
}

// AddJSActorMessage adds a JSActorMessage marker.
func (b *MarkerBuilder) AddJSActorMessage(actor string, startTime, duration float64) *MarkerBuilder {
	return b.addMarker("JSActorMessage", 0, startTime, duration, map[string]interface{}{
		"type":  "JSActorMessage",
		"actor": actor,
	})
}

// AddPaint adds a Paint marker.
func (b *MarkerBuilder) AddPaint(startTime, duration float64) *MarkerBuilder {
	return b.addMarker("Paint", 0, startTime, duration, map[string]interface{}{
		"type": "Paint",
	})
}

// AddEventDispatch adds an EventDispatch marker (Chrome-style).
func (b *MarkerBuilder) AddEventDispatch(eventType string, startTime, duration float64) *MarkerBuilder {
	return b.addMarker("EventDispatch", 0, startTime, duration, map[string]interface{}{
		"type": eventType,
	})
}

// AddUpdateLayoutTree adds an UpdateLayoutTree marker (Chrome-style).
func (b *MarkerBuilder) AddUpdateLayoutTree(startTime, duration float64) *MarkerBuilder {
	return b.addMarker("UpdateLayoutTree", 0, startTime, duration, map[string]interface{}{
		"type": "UpdateLayoutTree",
	})
}

// AddV8GCScavenger adds a V8 scavenger GC marker (Chrome-style).
func (b *MarkerBuilder) AddV8GCScavenger(startTime, duration float64) *MarkerBuilder {
	return b.addMarker("V8.GC_SCAVENGER", 0, startTime, duration, map[string]interface{}{
		"type": "V8.GC_SCAVENGER",
	})
}

// AddV8GCMarkCompactor adds a V8 mark-compactor GC marker (Chrome-style).
func (b *MarkerBuilder) AddV8GCMarkCompactor(startTime, duration float64) *MarkerBuilder {
	return b.addMarker("V8.GC_MARK_COMPACTOR", 0, startTime, duration, map[string]interface{}{
		"type": "V8.GC_MARK_COMPACTOR",
	})
}

// AddMajorGC adds a MajorGC marker (Chrome-style).
func (b *MarkerBuilder) AddMajorGC(startTime, duration float64) *MarkerBuilder {
	return b.addMarker("MajorGC", 0, startTime, duration, map[string]interface{}{
		"type": "MajorGC",
	})
}

// AddMinorGC adds a MinorGC marker (Chrome-style).
func (b *MarkerBuilder) AddMinorGC(startTime, duration float64) *MarkerBuilder {
	return b.addMarker("MinorGC", 0, startTime, duration, map[string]interface{}{
		"type": "MinorGC",
	})
}

// AddChannelMarker adds a ChannelMarker (network).
func (b *MarkerBuilder) AddChannelMarker(url string, startTime, duration float64) *MarkerBuilder {
	return b.addMarker("ChannelMarker", 0, startTime, duration, map[string]interface{}{
		"type": "ChannelMarker",
		"URI":  url,
	})
}

// AddHostResolver adds a HostResolver marker.
func (b *MarkerBuilder) AddHostResolver(host string, startTime, duration float64) *MarkerBuilder {
	return b.addMarker("HostResolver", 0, startTime, duration, map[string]interface{}{
		"type": "HostResolver",
		"host": host,
	})
}

// AddText adds a Text marker.
func (b *MarkerBuilder) AddText(text string, startTime float64) *MarkerBuilder {
	return b.addMarker("Text", 0, startTime, 0, map[string]interface{}{
		"type": "Text",
		"name": text,
	})
}

// AddCustom adds a custom marker with arbitrary data.
func (b *MarkerBuilder) AddCustom(name string, category int, startTime, duration float64, data map[string]interface{}) *MarkerBuilder {
	return b.addMarker(name, category, startTime, duration, data)
}

// Build returns the constructed markers and string array.
func (b *MarkerBuilder) Build() (parser.Markers, []string) {
	return b.markers, b.stringArray
}

// BuildMarkers returns just the markers (use with WithMarkers).
func (b *MarkerBuilder) BuildMarkers() parser.Markers {
	return b.markers
}

// BuildForThread builds markers and applies them to a ThreadBuilder.
func (b *MarkerBuilder) BuildForThread(tb *ThreadBuilder) *ThreadBuilder {
	markers, strings := b.Build()
	tb.thread.Markers = markers
	tb.thread.StringArray = append(tb.thread.StringArray, strings...)
	return tb
}

// SamplesBuilder provides a fluent API for constructing test samples.
type SamplesBuilder struct {
	samples parser.Samples
}

// NewSamplesBuilder creates a new SamplesBuilder.
func NewSamplesBuilder() *SamplesBuilder {
	return &SamplesBuilder{
		samples: parser.Samples{
			Length:         0,
			Stack:          []int{},
			Time:           []float64{},
			Weight:         []int{},
			ThreadCPUDelta: []int{},
		},
	}
}

// AddSample adds a sample with the given stack index and time.
func (b *SamplesBuilder) AddSample(stackIdx int, time float64) *SamplesBuilder {
	b.samples.Length++
	b.samples.Stack = append(b.samples.Stack, stackIdx)
	b.samples.Time = append(b.samples.Time, time)
	return b
}

// AddSampleWithCPUDelta adds a sample with CPU delta.
func (b *SamplesBuilder) AddSampleWithCPUDelta(stackIdx int, time float64, cpuDelta int) *SamplesBuilder {
	b.samples.Length++
	b.samples.Stack = append(b.samples.Stack, stackIdx)
	b.samples.Time = append(b.samples.Time, time)
	b.samples.ThreadCPUDelta = append(b.samples.ThreadCPUDelta, cpuDelta)
	return b
}

// AddSampleWithWeight adds a sample with weight.
func (b *SamplesBuilder) AddSampleWithWeight(stackIdx int, time float64, weight int) *SamplesBuilder {
	b.samples.Length++
	b.samples.Stack = append(b.samples.Stack, stackIdx)
	b.samples.Time = append(b.samples.Time, time)
	b.samples.Weight = append(b.samples.Weight, weight)
	return b
}

// WithWeightType sets the weight type.
func (b *SamplesBuilder) WithWeightType(wt string) *SamplesBuilder {
	b.samples.WeightType = wt
	return b
}

// Build returns the constructed samples.
func (b *SamplesBuilder) Build() parser.Samples {
	return b.samples
}

// StackTableBuilder provides a fluent API for constructing stack tables.
type StackTableBuilder struct {
	table parser.StackTable
}

// NewStackTableBuilder creates a new StackTableBuilder.
func NewStackTableBuilder() *StackTableBuilder {
	return &StackTableBuilder{
		table: parser.StackTable{
			Length:   0,
			Frame:    []int{},
			Category: []int{},
			Prefix:   []int{},
		},
	}
}

// AddStack adds a stack entry.
func (b *StackTableBuilder) AddStack(frameIdx, categoryIdx, prefixIdx int) *StackTableBuilder {
	b.table.Length++
	b.table.Frame = append(b.table.Frame, frameIdx)
	b.table.Category = append(b.table.Category, categoryIdx)
	b.table.Prefix = append(b.table.Prefix, prefixIdx)
	return b
}

// Build returns the constructed stack table.
func (b *StackTableBuilder) Build() parser.StackTable {
	return b.table
}

// FrameTableBuilder provides a fluent API for constructing frame tables.
type FrameTableBuilder struct {
	table parser.FrameTable
}

// NewFrameTableBuilder creates a new FrameTableBuilder.
func NewFrameTableBuilder() *FrameTableBuilder {
	return &FrameTableBuilder{
		table: parser.FrameTable{
			Length:         0,
			Address:        []interface{}{},
			InlineDepth:    []int{},
			Category:       []int{},
			Subcategory:    []int{},
			Func:           []int{},
			NativeSymbol:   []interface{}{},
			InnerWindowID:  []interface{}{},
			Implementation: []interface{}{},
			Line:           []interface{}{},
			Column:         []interface{}{},
		},
	}
}

// AddFrame adds a frame entry.
func (b *FrameTableBuilder) AddFrame(funcIdx, categoryIdx int) *FrameTableBuilder {
	b.table.Length++
	b.table.Address = append(b.table.Address, nil)
	b.table.InlineDepth = append(b.table.InlineDepth, 0)
	b.table.Category = append(b.table.Category, categoryIdx)
	b.table.Subcategory = append(b.table.Subcategory, 0)
	b.table.Func = append(b.table.Func, funcIdx)
	b.table.NativeSymbol = append(b.table.NativeSymbol, nil)
	b.table.InnerWindowID = append(b.table.InnerWindowID, nil)
	b.table.Implementation = append(b.table.Implementation, nil)
	b.table.Line = append(b.table.Line, nil)
	b.table.Column = append(b.table.Column, nil)
	return b
}

// Build returns the constructed frame table.
func (b *FrameTableBuilder) Build() parser.FrameTable {
	return b.table
}

// FuncTableBuilder provides a fluent API for constructing function tables.
type FuncTableBuilder struct {
	table parser.FuncTable
}

// NewFuncTableBuilder creates a new FuncTableBuilder.
func NewFuncTableBuilder() *FuncTableBuilder {
	return &FuncTableBuilder{
		table: parser.FuncTable{
			Length:        0,
			Name:          []int{},
			IsJS:          []bool{},
			RelevantForJS: []bool{},
			Resource:      []int{},
			FileName:      []int{},
			LineNumber:    []int{},
			ColumnNumber:  []int{},
		},
	}
}

// AddFunc adds a function entry.
func (b *FuncTableBuilder) AddFunc(nameIdx int, isJS bool, resourceIdx int) *FuncTableBuilder {
	b.table.Length++
	b.table.Name = append(b.table.Name, nameIdx)
	b.table.IsJS = append(b.table.IsJS, isJS)
	b.table.RelevantForJS = append(b.table.RelevantForJS, isJS)
	b.table.Resource = append(b.table.Resource, resourceIdx)
	b.table.FileName = append(b.table.FileName, -1)
	b.table.LineNumber = append(b.table.LineNumber, 0)
	b.table.ColumnNumber = append(b.table.ColumnNumber, 0)
	return b
}

// AddFuncWithFile adds a function entry with file info.
func (b *FuncTableBuilder) AddFuncWithFile(nameIdx int, isJS bool, resourceIdx, fileNameIdx, lineNum int) *FuncTableBuilder {
	b.table.Length++
	b.table.Name = append(b.table.Name, nameIdx)
	b.table.IsJS = append(b.table.IsJS, isJS)
	b.table.RelevantForJS = append(b.table.RelevantForJS, isJS)
	b.table.Resource = append(b.table.Resource, resourceIdx)
	b.table.FileName = append(b.table.FileName, fileNameIdx)
	b.table.LineNumber = append(b.table.LineNumber, lineNum)
	b.table.ColumnNumber = append(b.table.ColumnNumber, 0)
	return b
}

// Build returns the constructed function table.
func (b *FuncTableBuilder) Build() parser.FuncTable {
	return b.table
}

// ResourceTableBuilder provides a fluent API for constructing resource tables.
type ResourceTableBuilder struct {
	table parser.ResourceTable
}

// NewResourceTableBuilder creates a new ResourceTableBuilder.
func NewResourceTableBuilder() *ResourceTableBuilder {
	return &ResourceTableBuilder{
		table: parser.ResourceTable{
			Length: 0,
			Lib:    []int{},
			Name:   []int{},
			Host:   []int{},
			Type:   []int{},
		},
	}
}

// AddResource adds a resource entry.
func (b *ResourceTableBuilder) AddResource(libIdx, nameIdx, hostIdx, typeIdx int) *ResourceTableBuilder {
	b.table.Length++
	b.table.Lib = append(b.table.Lib, libIdx)
	b.table.Name = append(b.table.Name, nameIdx)
	b.table.Host = append(b.table.Host, hostIdx)
	b.table.Type = append(b.table.Type, typeIdx)
	return b
}

// Build returns the constructed resource table.
func (b *ResourceTableBuilder) Build() parser.ResourceTable {
	return b.table
}
