package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/CedricHerzog/perfowl/internal/analyzer"
	"github.com/CedricHerzog/perfowl/internal/chart"
	"github.com/CedricHerzog/perfowl/internal/format/toon"
	"github.com/CedricHerzog/perfowl/internal/parser"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

// PerfOwlServer wraps the MCP server for browser performance trace analysis
type PerfOwlServer struct {
	server *server.MCPServer
}

// NewServer creates a new PerfOwl MCP server
func NewServer() *PerfOwlServer {
	s := server.NewMCPServer(
		"PerfOwl - Optimization Workbench & Lab",
		"1.0.0",
		server.WithToolCapabilities(true),
	)

	pos := &PerfOwlServer{server: s}
	pos.registerTools()

	return pos
}

// registerTools adds all profile analysis tools to the server
func (pos *PerfOwlServer) registerTools() {
	// get_summary tool
	summaryTool := mcp.NewTool("get_summary",
		mcp.WithDescription("Get a summary of the browser profile including duration, platform, threads, and extensions"),
		mcp.WithString("path", mcp.Required(), mcp.Description("Path to the profile JSON file (gzip supported)")),
	)
	pos.server.AddTool(summaryTool, pos.handleGetSummary)

	// get_bottlenecks tool
	bottlenecksTool := mcp.NewTool("get_bottlenecks",
		mcp.WithDescription("Detect performance bottlenecks in the profile including long tasks, GC pressure, sync IPC, layout thrashing, and network blocking"),
		mcp.WithString("path", mcp.Required(), mcp.Description("Path to the profile JSON file")),
		mcp.WithString("min_severity", mcp.Description("Minimum severity to report: low, medium, high")),
	)
	pos.server.AddTool(bottlenecksTool, pos.handleGetBottlenecks)

	// get_markers tool
	markersTool := mcp.NewTool("get_markers",
		mcp.WithDescription("Extract and analyze markers from the profile, optionally filtered by type, category, or name"),
		mcp.WithString("path", mcp.Required(), mcp.Description("Path to the profile JSON file")),
		mcp.WithString("type", mcp.Description("Filter by marker type (e.g., GCMajor, DOMEvent, JSActorMessage)")),
		mcp.WithString("category", mcp.Description("Filter by category (e.g., JavaScript, Layout, Network)")),
		mcp.WithString("name", mcp.Description("Filter by marker name pattern (case-insensitive substring match)")),
		mcp.WithNumber("min_duration", mcp.Description("Minimum duration in milliseconds")),
		mcp.WithNumber("limit", mcp.Description("Maximum number of markers to return")),
	)
	pos.server.AddTool(markersTool, pos.handleGetMarkers)

	// analyze_extension tool
	extensionTool := mcp.NewTool("analyze_extension",
		mcp.WithDescription("Analyze extension performance impact including duration, events, DOM interactions, and IPC messages"),
		mcp.WithString("path", mcp.Required(), mcp.Description("Path to the profile JSON file")),
		mcp.WithString("extension_id", mcp.Description("Filter by specific extension ID (optional)")),
	)
	pos.server.AddTool(extensionTool, pos.handleAnalyzeExtension)

	// analyze_profile tool (comprehensive analysis)
	analyzeTool := mcp.NewTool("analyze_profile",
		mcp.WithDescription("Perform a comprehensive analysis of the profile including summary, bottlenecks, and extension impact"),
		mcp.WithString("path", mcp.Required(), mcp.Description("Path to the profile JSON file")),
	)
	pos.server.AddTool(analyzeTool, pos.handleAnalyzeProfile)

	// get_call_tree tool
	callTreeTool := mcp.NewTool("get_call_tree",
		mcp.WithDescription("Analyze call tree to find hot functions by self time and running time, with hot path detection"),
		mcp.WithString("path", mcp.Required(), mcp.Description("Path to the profile JSON file")),
		mcp.WithString("thread", mcp.Description("Filter by thread name (optional)")),
		mcp.WithNumber("limit", mcp.Description("Maximum number of functions/paths to return (default 20)")),
	)
	pos.server.AddTool(callTreeTool, pos.handleGetCallTree)

	// get_category_breakdown tool
	categoryTool := mcp.NewTool("get_category_breakdown",
		mcp.WithDescription("Get time spent per profiler category (JavaScript, Layout, GC/CC, Network, Graphics, DOM, etc.)"),
		mcp.WithString("path", mcp.Required(), mcp.Description("Path to the profile JSON file")),
		mcp.WithString("thread", mcp.Description("Filter by thread name (optional)")),
	)
	pos.server.AddTool(categoryTool, pos.handleGetCategoryBreakdown)

	// get_thread_analysis tool
	threadTool := mcp.NewTool("get_thread_analysis",
		mcp.WithDescription("Analyze all threads including CPU time, sample counts, wake patterns, and category distribution"),
		mcp.WithString("path", mcp.Required(), mcp.Description("Path to the profile JSON file")),
	)
	pos.server.AddTool(threadTool, pos.handleGetThreadAnalysis)

	// compare_profiles tool
	compareTool := mcp.NewTool("compare_profiles",
		mcp.WithDescription("Compare two profiles to identify performance improvements or regressions"),
		mcp.WithString("baseline", mcp.Required(), mcp.Description("Path to the baseline profile JSON file")),
		mcp.WithString("comparison", mcp.Required(), mcp.Description("Path to the comparison profile JSON file")),
	)
	pos.server.AddTool(compareTool, pos.handleCompareProfiles)

	// analyze_workers tool
	workersTool := mcp.NewTool("analyze_workers",
		mcp.WithDescription("Analyze worker thread performance including CPU time, idle time, messaging, and synchronization points"),
		mcp.WithString("path", mcp.Required(), mcp.Description("Path to the profile JSON file")),
	)
	pos.server.AddTool(workersTool, pos.handleAnalyzeWorkers)

	// analyze_crypto tool
	cryptoTool := mcp.NewTool("analyze_crypto",
		mcp.WithDescription("Analyze cryptographic operations including SubtleCrypto API usage, algorithm detection, and serialization issues"),
		mcp.WithString("path", mcp.Required(), mcp.Description("Path to the profile JSON file")),
	)
	pos.server.AddTool(cryptoTool, pos.handleAnalyzeCrypto)

	// analyze_jscrypto tool
	jsCryptoTool := mcp.NewTool("analyze_jscrypto",
		mcp.WithDescription("Analyze JavaScript-level crypto operations including crypto worker files (seipdDecryptionWorker, openpgp.js), per-worker time distribution, and top functions"),
		mcp.WithString("path", mcp.Required(), mcp.Description("Path to the profile JSON file")),
	)
	pos.server.AddTool(jsCryptoTool, pos.handleAnalyzeJSCrypto)

	// analyze_contention tool
	contentionTool := mcp.NewTool("analyze_contention",
		mcp.WithDescription("Detect thread contention issues including GC pauses affecting workers, sync IPC blocking, and lock contention"),
		mcp.WithString("path", mcp.Required(), mcp.Description("Path to the profile JSON file")),
	)
	pos.server.AddTool(contentionTool, pos.handleAnalyzeContention)

	// analyze_scaling tool
	scalingTool := mcp.NewTool("analyze_scaling",
		mcp.WithDescription("Analyze parallel scaling efficiency including worker utilization, speedup, and bottleneck identification"),
		mcp.WithString("path", mcp.Required(), mcp.Description("Path to the profile JSON file")),
	)
	pos.server.AddTool(scalingTool, pos.handleAnalyzeScaling)

	// compare_scaling tool
	compareScalingTool := mcp.NewTool("compare_scaling",
		mcp.WithDescription("Compare parallel scaling efficiency between two profiles"),
		mcp.WithString("baseline", mcp.Required(), mcp.Description("Path to the baseline profile JSON file")),
		mcp.WithString("comparison", mcp.Required(), mcp.Description("Path to the comparison profile JSON file")),
	)
	pos.server.AddTool(compareScalingTool, pos.handleCompareScaling)

	// batch_analyze tool
	batchTool := mcp.NewTool("batch_analyze",
		mcp.WithDescription("Analyze multiple profiles across worker counts and return aggregated results for charting. Provide a JSON array of profile entries."),
		mcp.WithString("profiles", mcp.Required(), mcp.Description(`JSON array of profile entries. Each entry: {"path": "file.json.gz", "workers": 4, "label": "Chrome", "start_pattern": "EventDispatch", "end_pattern": "UpdateLayoutTree", "start_min_duration": 0, "end_min_duration": 1000}`)),
	)
	pos.server.AddTool(batchTool, pos.handleBatchAnalyze)

	// generate_chart tool
	chartTool := mcp.NewTool("generate_chart",
		mcp.WithDescription("Generate SVG chart from batch analysis of multiple profiles"),
		mcp.WithString("profiles", mcp.Required(), mcp.Description(`JSON array of profile entries. Each entry: {"path": "file.json.gz", "workers": 4, "label": "Chrome", "start_pattern": "EventDispatch", "end_pattern": "UpdateLayoutTree", "start_min_duration": 0, "end_min_duration": 1000}`)),
		mcp.WithString("chart_type", mcp.Description("Chart type: wall_clock, efficiency, speedup, crypto_time, operation_time (default: wall_clock)")),
		mcp.WithString("output", mcp.Description("Output mode: inline (returns SVG), file (saves to path)")),
		mcp.WithString("output_path", mcp.Description("File path for 'file' output mode (default: chart.svg)")),
	)
	pos.server.AddTool(chartTool, pos.handleGenerateChart)

	// get_delimiter_markers tool
	delimitersTool := mcp.NewTool("get_delimiter_markers",
		mcp.WithDescription("List markers that can be used as operation start/end delimiters (click events, DOM updates, paint events, etc.). Use this to identify events for measuring actual operation time."),
		mcp.WithString("path", mcp.Required(), mcp.Description("Path to the profile JSON file")),
		mcp.WithString("categories", mcp.Description("Filter by categories (comma-separated, e.g., 'DOM,Layout,Graphics')")),
		mcp.WithNumber("limit", mcp.Description("Maximum number of markers to return (default: all)")),
	)
	pos.server.AddTool(delimitersTool, pos.handleGetDelimiterMarkers)

	// measure_operation tool
	measureTool := mcp.NewTool("measure_operation",
		mcp.WithDescription("Measure time between two marker patterns to get actual operation duration (e.g., click to paint). Pattern format: 'type' or 'type:subtype' (e.g., 'DOMEvent:click', 'Styles', 'Paint')."),
		mcp.WithString("path", mcp.Required(), mcp.Description("Path to the profile JSON file")),
		mcp.WithString("start_pattern", mcp.Description("Pattern to match start marker (e.g., 'DOMEvent:click')")),
		mcp.WithString("end_pattern", mcp.Description("Pattern to match end marker (e.g., 'Styles' or 'Paint')")),
		mcp.WithNumber("start_after_ms", mcp.Description("Only consider markers after this time (optional)")),
		mcp.WithNumber("end_before_ms", mcp.Description("Only consider markers before this time (optional)")),
		mcp.WithNumber("start_index", mcp.Description("Alternative: use marker index instead of pattern for start")),
		mcp.WithNumber("end_index", mcp.Description("Alternative: use marker index instead of pattern for end")),
		mcp.WithBoolean("find_last", mcp.Description("If true, find the LAST matching end marker instead of the first (for full operation time)")),
		mcp.WithNumber("start_min_duration", mcp.Description("Only match start markers with duration >= this value in ms (optional)")),
		mcp.WithNumber("end_min_duration", mcp.Description("Only match end markers with duration >= this value in ms (optional)")),
	)
	pos.server.AddTool(measureTool, pos.handleMeasureOperation)
}

// Serve starts the MCP server on stdio
func (pos *PerfOwlServer) Serve() error {
	return server.ServeStdio(pos.server)
}

// Tool handlers

func (pos *PerfOwlServer) handleGetSummary(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	path, err := req.RequireString("path")
	if err != nil {
		return nil, fmt.Errorf("path is required: %w", err)
	}

	profile, _, err := parser.LoadProfileAuto(path)
	if err != nil {
		return nil, fmt.Errorf("failed to load profile: %w", err)
	}

	summary := buildSummary(profile)
	output, err := toon.Encode(summary)
	if err != nil {
		return nil, fmt.Errorf("failed to encode summary: %w", err)
	}

	return mcp.NewToolResultText(output), nil
}

func (pos *PerfOwlServer) handleGetBottlenecks(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	path, err := req.RequireString("path")
	if err != nil {
		return nil, fmt.Errorf("path is required: %w", err)
	}

	profile, _, err := parser.LoadProfileAuto(path)
	if err != nil {
		return nil, fmt.Errorf("failed to load profile: %w", err)
	}

	bottlenecks := analyzer.DetectBottlenecks(profile)

	// Filter by min severity if provided
	if minSev, err := req.RequireString("min_severity"); err == nil && minSev != "" {
		sev := analyzer.ParseSeverity(minSev)
		var filtered []analyzer.Bottleneck
		for _, b := range bottlenecks {
			if b.Severity >= sev {
				filtered = append(filtered, b)
			}
		}
		bottlenecks = filtered
	}

	report := analyzer.BottleneckReport{
		Score:       analyzer.CalculateScore(bottlenecks),
		Summary:     analyzer.GenerateSummary(bottlenecks, profile),
		Bottlenecks: bottlenecks,
	}

	output, err := toon.Encode(report)
	if err != nil {
		return nil, fmt.Errorf("failed to encode bottlenecks: %w", err)
	}

	return mcp.NewToolResultText(output), nil
}

func (pos *PerfOwlServer) handleGetMarkers(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	path, err := req.RequireString("path")
	if err != nil {
		return nil, fmt.Errorf("path is required: %w", err)
	}

	profile, _, err := parser.LoadProfileAuto(path)
	if err != nil {
		return nil, fmt.Errorf("failed to load profile: %w", err)
	}

	// Extract all markers
	var allMarkers []parser.ParsedMarker
	for _, thread := range profile.Threads {
		markers := parser.ExtractMarkers(&thread, profile.Meta.Categories)
		allMarkers = append(allMarkers, markers...)
	}

	filtered := allMarkers

	// Apply filters
	if markerType, err := req.RequireString("type"); err == nil && markerType != "" {
		filtered = parser.FilterMarkersByType(filtered, parser.MarkerType(markerType))
	}

	if category, err := req.RequireString("category"); err == nil && category != "" {
		filtered = parser.FilterMarkersByCategory(filtered, category)
	}

	if name, err := req.RequireString("name"); err == nil && name != "" {
		filtered = parser.FilterMarkersByName(filtered, name)
	}

	if minDur, err := req.RequireFloat("min_duration"); err == nil && minDur > 0 {
		filtered = parser.FilterMarkersByDuration(filtered, minDur)
	}

	// Apply limit
	if limit, err := req.RequireFloat("limit"); err == nil && limit > 0 {
		limitInt := int(limit)
		if len(filtered) > limitInt {
			filtered = filtered[:limitInt]
		}
	}

	stats := parser.GetMarkerStats(filtered)

	markerOutput := map[string]interface{}{
		"total_count": len(allMarkers),
		"filtered":    len(filtered),
		"stats":       stats,
		"markers":     filtered,
	}

	output, err := toon.Encode(markerOutput)
	if err != nil {
		return nil, fmt.Errorf("failed to encode markers: %w", err)
	}

	return mcp.NewToolResultText(output), nil
}

func (pos *PerfOwlServer) handleAnalyzeExtension(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	path, err := req.RequireString("path")
	if err != nil {
		return nil, fmt.Errorf("path is required: %w", err)
	}

	profile, _, err := parser.LoadProfileAuto(path)
	if err != nil {
		return nil, fmt.Errorf("failed to load profile: %w", err)
	}

	report := analyzer.AnalyzeExtensions(profile)

	// Filter by extension ID if provided
	if extID, err := req.RequireString("extension_id"); err == nil && extID != "" {
		var filtered []analyzer.ExtensionReport
		for _, ext := range report.Extensions {
			if ext.ID == extID || len(extID) > 0 && ext.ID == extID {
				filtered = append(filtered, ext)
			}
		}
		report.Extensions = filtered
	}

	output, err := toon.Encode(report)
	if err != nil {
		return nil, fmt.Errorf("failed to encode extension report: %w", err)
	}

	return mcp.NewToolResultText(output), nil
}

func (pos *PerfOwlServer) handleAnalyzeProfile(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	path, err := req.RequireString("path")
	if err != nil {
		return nil, fmt.Errorf("path is required: %w", err)
	}

	profile, _, err := parser.LoadProfileAuto(path)
	if err != nil {
		return nil, fmt.Errorf("failed to load profile: %w", err)
	}

	// Build comprehensive analysis
	summary := buildSummary(profile)
	bottlenecks := analyzer.DetectBottlenecks(profile)
	bottleneckReport := analyzer.BottleneckReport{
		Score:       analyzer.CalculateScore(bottlenecks),
		Summary:     analyzer.GenerateSummary(bottlenecks, profile),
		Bottlenecks: bottlenecks,
	}
	extensions := analyzer.AnalyzeExtensions(profile)

	fullReport := map[string]interface{}{
		"summary":     summary,
		"bottlenecks": bottleneckReport,
		"extensions":  extensions,
	}

	output, err := toon.Encode(fullReport)
	if err != nil {
		return nil, fmt.Errorf("failed to encode full report: %w", err)
	}

	return mcp.NewToolResultText(output), nil
}

// Helper types and functions

type ProfileSummary struct {
	DurationSeconds float64           `json:"duration_seconds"`
	Platform        string            `json:"platform"`
	OSCPU           string            `json:"os_cpu"`
	Product         string            `json:"product"`
	BuildID         string            `json:"build_id"`
	CPUName         string            `json:"cpu_name"`
	PhysicalCPUs    int               `json:"physical_cpus"`
	LogicalCPUs     int               `json:"logical_cpus"`
	ThreadCount     int               `json:"thread_count"`
	MainThreadCount int               `json:"main_thread_count"`
	ExtensionCount  int               `json:"extension_count"`
	Extensions      map[string]string `json:"extensions"`
	Features        []string          `json:"features"`
	TotalMarkers    int               `json:"total_markers"`
	TotalSamples    int               `json:"total_samples"`
}

func (pos *PerfOwlServer) handleGetCallTree(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	path, err := req.RequireString("path")
	if err != nil {
		return nil, fmt.Errorf("path is required: %w", err)
	}

	profile, _, err := parser.LoadProfileAuto(path)
	if err != nil {
		return nil, fmt.Errorf("failed to load profile: %w", err)
	}

	threadName := ""
	if t, err := req.RequireString("thread"); err == nil {
		threadName = t
	}

	limit := 20
	if l, err := req.RequireFloat("limit"); err == nil && l > 0 {
		limit = int(l)
	}

	analysis := analyzer.AnalyzeCallTree(profile, threadName, limit)

	output, err := toon.Encode(analysis)
	if err != nil {
		return nil, fmt.Errorf("failed to encode call tree: %w", err)
	}

	return mcp.NewToolResultText(output), nil
}

func (pos *PerfOwlServer) handleGetCategoryBreakdown(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	path, err := req.RequireString("path")
	if err != nil {
		return nil, fmt.Errorf("path is required: %w", err)
	}

	profile, _, err := parser.LoadProfileAuto(path)
	if err != nil {
		return nil, fmt.Errorf("failed to load profile: %w", err)
	}

	threadName := ""
	if t, err := req.RequireString("thread"); err == nil {
		threadName = t
	}

	breakdown := analyzer.AnalyzeCategories(profile, threadName)

	output, err := toon.Encode(breakdown)
	if err != nil {
		return nil, fmt.Errorf("failed to encode category breakdown: %w", err)
	}

	return mcp.NewToolResultText(output), nil
}

func (pos *PerfOwlServer) handleGetThreadAnalysis(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	path, err := req.RequireString("path")
	if err != nil {
		return nil, fmt.Errorf("path is required: %w", err)
	}

	profile, _, err := parser.LoadProfileAuto(path)
	if err != nil {
		return nil, fmt.Errorf("failed to load profile: %w", err)
	}

	analysis := analyzer.AnalyzeThreads(profile)

	output, err := toon.Encode(analysis)
	if err != nil {
		return nil, fmt.Errorf("failed to encode thread analysis: %w", err)
	}

	return mcp.NewToolResultText(output), nil
}

func (pos *PerfOwlServer) handleCompareProfiles(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	baselinePath, err := req.RequireString("baseline")
	if err != nil {
		return nil, fmt.Errorf("baseline path is required: %w", err)
	}

	comparisonPath, err := req.RequireString("comparison")
	if err != nil {
		return nil, fmt.Errorf("comparison path is required: %w", err)
	}

	baseline, _, err := parser.LoadProfileAuto(baselinePath)
	if err != nil {
		return nil, fmt.Errorf("failed to load baseline profile: %w", err)
	}

	comparison, _, err := parser.LoadProfileAuto(comparisonPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load comparison profile: %w", err)
	}

	diff := analyzer.CompareProfiles(baseline, comparison)

	output, err := toon.Encode(diff)
	if err != nil {
		return nil, fmt.Errorf("failed to encode comparison: %w", err)
	}

	return mcp.NewToolResultText(output), nil
}

// Helper types and functions

func buildSummary(profile *parser.Profile) ProfileSummary {
	summary := ProfileSummary{
		DurationSeconds: profile.DurationSeconds(),
		Platform:        profile.Meta.Platform,
		OSCPU:           profile.Meta.OSCPU,
		Product:         profile.Meta.Product,
		BuildID:         profile.Meta.AppBuildID,
		CPUName:         profile.Meta.CPUName,
		PhysicalCPUs:    profile.Meta.PhysicalCPUs,
		LogicalCPUs:     profile.Meta.LogicalCPUs,
		ThreadCount:     profile.ThreadCount(),
		ExtensionCount:  profile.ExtensionCount(),
		Extensions:      profile.GetExtensions(),
		Features:        profile.Meta.Configuration.Features,
	}

	for _, t := range profile.Threads {
		if t.IsMainThread {
			summary.MainThreadCount++
		}
		summary.TotalMarkers += t.Markers.Length
		summary.TotalSamples += t.Samples.Length
	}

	return summary
}

func (pos *PerfOwlServer) handleAnalyzeWorkers(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	path, err := req.RequireString("path")
	if err != nil {
		return nil, fmt.Errorf("path is required: %w", err)
	}

	profile, _, err := parser.LoadProfileAuto(path)
	if err != nil {
		return nil, fmt.Errorf("failed to load profile: %w", err)
	}

	analysis := analyzer.AnalyzeWorkers(profile)

	output, err := toon.Encode(analysis)
	if err != nil {
		return nil, fmt.Errorf("failed to encode worker analysis: %w", err)
	}

	return mcp.NewToolResultText(output), nil
}

func (pos *PerfOwlServer) handleAnalyzeCrypto(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	path, err := req.RequireString("path")
	if err != nil {
		return nil, fmt.Errorf("path is required: %w", err)
	}

	profile, _, err := parser.LoadProfileAuto(path)
	if err != nil {
		return nil, fmt.Errorf("failed to load profile: %w", err)
	}

	analysis := analyzer.AnalyzeCrypto(profile)

	output, err := toon.Encode(analysis)
	if err != nil {
		return nil, fmt.Errorf("failed to encode crypto analysis: %w", err)
	}

	return mcp.NewToolResultText(output), nil
}

func (pos *PerfOwlServer) handleAnalyzeJSCrypto(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	path, err := req.RequireString("path")
	if err != nil {
		return nil, fmt.Errorf("path is required: %w", err)
	}

	profile, _, err := parser.LoadProfileAuto(path)
	if err != nil {
		return nil, fmt.Errorf("failed to load profile: %w", err)
	}

	analysis := analyzer.AnalyzeJSCrypto(profile)

	output, err := toon.Encode(analysis)
	if err != nil {
		return nil, fmt.Errorf("failed to encode JS crypto analysis: %w", err)
	}

	return mcp.NewToolResultText(output), nil
}

func (pos *PerfOwlServer) handleAnalyzeContention(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	path, err := req.RequireString("path")
	if err != nil {
		return nil, fmt.Errorf("path is required: %w", err)
	}

	profile, _, err := parser.LoadProfileAuto(path)
	if err != nil {
		return nil, fmt.Errorf("failed to load profile: %w", err)
	}

	analysis := analyzer.AnalyzeContention(profile)

	output, err := toon.Encode(analysis)
	if err != nil {
		return nil, fmt.Errorf("failed to encode contention analysis: %w", err)
	}

	return mcp.NewToolResultText(output), nil
}

func (pos *PerfOwlServer) handleAnalyzeScaling(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	path, err := req.RequireString("path")
	if err != nil {
		return nil, fmt.Errorf("path is required: %w", err)
	}

	profile, _, err := parser.LoadProfileAuto(path)
	if err != nil {
		return nil, fmt.Errorf("failed to load profile: %w", err)
	}

	analysis := analyzer.AnalyzeScaling(profile)

	output, err := toon.Encode(analysis)
	if err != nil {
		return nil, fmt.Errorf("failed to encode scaling analysis: %w", err)
	}

	return mcp.NewToolResultText(output), nil
}

func (pos *PerfOwlServer) handleCompareScaling(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	baselinePath, err := req.RequireString("baseline")
	if err != nil {
		return nil, fmt.Errorf("baseline path is required: %w", err)
	}

	comparisonPath, err := req.RequireString("comparison")
	if err != nil {
		return nil, fmt.Errorf("comparison path is required: %w", err)
	}

	baseline, _, err := parser.LoadProfileAuto(baselinePath)
	if err != nil {
		return nil, fmt.Errorf("failed to load baseline profile: %w", err)
	}

	comparison, _, err := parser.LoadProfileAuto(comparisonPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load comparison profile: %w", err)
	}

	result := analyzer.CompareScaling(baseline, comparison)

	output, err := toon.Encode(result)
	if err != nil {
		return nil, fmt.Errorf("failed to encode scaling comparison: %w", err)
	}

	return mcp.NewToolResultText(output), nil
}

func (pos *PerfOwlServer) handleBatchAnalyze(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	profilesJSON, err := req.RequireString("profiles")
	if err != nil {
		return nil, fmt.Errorf("profiles is required: %w", err)
	}

	var profiles []analyzer.ProfileEntry
	if err := json.Unmarshal([]byte(profilesJSON), &profiles); err != nil {
		return nil, fmt.Errorf("failed to parse profiles JSON: %w", err)
	}

	if len(profiles) == 0 {
		return nil, fmt.Errorf("no profiles provided")
	}

	result, err := analyzer.AnalyzeBatch(profiles)
	if err != nil {
		return nil, fmt.Errorf("batch analysis failed: %w", err)
	}

	output, err := toon.Encode(result)
	if err != nil {
		return nil, fmt.Errorf("failed to encode batch result: %w", err)
	}

	return mcp.NewToolResultText(output), nil
}

func (pos *PerfOwlServer) handleGenerateChart(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	profilesJSON, err := req.RequireString("profiles")
	if err != nil {
		return nil, fmt.Errorf("profiles is required: %w", err)
	}

	var profiles []analyzer.ProfileEntry
	if err := json.Unmarshal([]byte(profilesJSON), &profiles); err != nil {
		return nil, fmt.Errorf("failed to parse profiles JSON: %w", err)
	}

	if len(profiles) == 0 {
		return nil, fmt.Errorf("no profiles provided")
	}

	chartType := chart.ChartWallClock
	if ct, err := req.RequireString("chart_type"); err == nil && ct != "" {
		chartType = chart.ChartType(ct)
	}

	outputMode := "inline"
	if om, err := req.RequireString("output"); err == nil && om != "" {
		outputMode = om
	}

	result, err := analyzer.AnalyzeBatch(profiles)
	if err != nil {
		return nil, fmt.Errorf("batch analysis failed: %w", err)
	}

	svg := chart.GenerateScalingChart(result, chartType)

	if outputMode == "file" {
		outputPath := "chart.svg"
		if op, err := req.RequireString("output_path"); err == nil && op != "" {
			outputPath = op
		}
		if err := os.WriteFile(outputPath, []byte(svg), 0644); err != nil {
			return nil, fmt.Errorf("failed to write SVG file: %w", err)
		}
		return mcp.NewToolResultText(fmt.Sprintf("Chart saved to: %s", outputPath)), nil
	}

	return mcp.NewToolResultText(svg), nil
}

func (pos *PerfOwlServer) handleGetDelimiterMarkers(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	path, err := req.RequireString("path")
	if err != nil {
		return nil, fmt.Errorf("path is required: %w", err)
	}

	profile, _, err := parser.LoadProfileAuto(path)
	if err != nil {
		return nil, fmt.Errorf("failed to load profile: %w", err)
	}

	var categories []string
	if cats, err := req.RequireString("categories"); err == nil && cats != "" {
		categories = append(categories, splitAndTrim(cats)...)
	}

	limit := 0
	if l, err := req.RequireFloat("limit"); err == nil && l > 0 {
		limit = int(l)
	}

	report := analyzer.GetDelimiterMarkersReport(profile, categories, limit)

	output, err := toon.Encode(report)
	if err != nil {
		return nil, fmt.Errorf("failed to encode delimiter markers: %w", err)
	}

	return mcp.NewToolResultText(output), nil
}

func (pos *PerfOwlServer) handleMeasureOperation(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	path, err := req.RequireString("path")
	if err != nil {
		return nil, fmt.Errorf("path is required: %w", err)
	}

	profile, _, err := parser.LoadProfileAuto(path)
	if err != nil {
		return nil, fmt.Errorf("failed to load profile: %w", err)
	}

	// Check if using index-based measurement
	startIdx, startIdxErr := req.RequireFloat("start_index")
	endIdx, endIdxErr := req.RequireFloat("end_index")

	if startIdxErr == nil && endIdxErr == nil {
		// Index-based measurement
		measurement, err := analyzer.MeasureOperationByIndex(profile, int(startIdx), int(endIdx))
		if err != nil {
			return nil, fmt.Errorf("measurement failed: %w", err)
		}

		output, err := toon.Encode(measurement)
		if err != nil {
			return nil, fmt.Errorf("failed to encode measurement: %w", err)
		}
		return mcp.NewToolResultText(output), nil
	}

	// Pattern-based measurement
	startPattern, err := req.RequireString("start_pattern")
	if err != nil {
		return nil, fmt.Errorf("start_pattern is required: %w", err)
	}

	endPattern, err := req.RequireString("end_pattern")
	if err != nil {
		return nil, fmt.Errorf("end_pattern is required: %w", err)
	}

	opts := analyzer.MeasureOptions{
		StartPattern: startPattern,
		EndPattern:   endPattern,
	}

	if s, err := req.RequireFloat("start_after_ms"); err == nil {
		opts.StartAfterMs = s
	}

	if e, err := req.RequireFloat("end_before_ms"); err == nil {
		opts.EndBeforeMs = e
	}

	if fl, err := req.RequireBool("find_last"); err == nil {
		opts.FindLast = fl
	}

	if sd, err := req.RequireFloat("start_min_duration"); err == nil {
		opts.StartMinDurationMs = sd
	}

	if ed, err := req.RequireFloat("end_min_duration"); err == nil {
		opts.EndMinDurationMs = ed
	}

	measurement, err := analyzer.MeasureOperationAdvanced(profile, opts)
	if err != nil {
		return nil, fmt.Errorf("measurement failed: %w", err)
	}

	output, err := toon.Encode(measurement)
	if err != nil {
		return nil, fmt.Errorf("failed to encode measurement: %w", err)
	}

	return mcp.NewToolResultText(output), nil
}

// splitAndTrim splits a string by comma and trims whitespace
func splitAndTrim(s string) []string {
	var result []string
	for _, part := range strings.Split(s, ",") {
		trimmed := strings.TrimSpace(part)
		if trimmed != "" {
			result = append(result, trimmed)
		}
	}
	return result
}
