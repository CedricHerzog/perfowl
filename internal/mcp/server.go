package mcp

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/CedricHerzog/perfowl/internal/analyzer"
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
		mcp.WithDescription("Extract and analyze markers from the profile, optionally filtered by type or category"),
		mcp.WithString("path", mcp.Required(), mcp.Description("Path to the profile JSON file")),
		mcp.WithString("type", mcp.Description("Filter by marker type (e.g., GCMajor, DOMEvent, JSActorMessage)")),
		mcp.WithString("category", mcp.Description("Filter by category (e.g., JavaScript, Layout, Network)")),
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
	jsonBytes, err := json.MarshalIndent(summary, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("failed to marshal summary: %w", err)
	}

	return mcp.NewToolResultText(string(jsonBytes)), nil
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

	jsonBytes, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("failed to marshal bottlenecks: %w", err)
	}

	return mcp.NewToolResultText(string(jsonBytes)), nil
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

	output := map[string]interface{}{
		"total_count": len(allMarkers),
		"filtered":    len(filtered),
		"stats":       stats,
		"markers":     filtered,
	}

	jsonBytes, err := json.MarshalIndent(output, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("failed to marshal markers: %w", err)
	}

	return mcp.NewToolResultText(string(jsonBytes)), nil
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

	jsonBytes, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("failed to marshal extension report: %w", err)
	}

	return mcp.NewToolResultText(string(jsonBytes)), nil
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

	jsonBytes, err := json.MarshalIndent(fullReport, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("failed to marshal full report: %w", err)
	}

	return mcp.NewToolResultText(string(jsonBytes)), nil
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

	jsonBytes, err := json.MarshalIndent(analysis, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("failed to marshal call tree: %w", err)
	}

	return mcp.NewToolResultText(string(jsonBytes)), nil
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

	jsonBytes, err := json.MarshalIndent(breakdown, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("failed to marshal category breakdown: %w", err)
	}

	return mcp.NewToolResultText(string(jsonBytes)), nil
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

	jsonBytes, err := json.MarshalIndent(analysis, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("failed to marshal thread analysis: %w", err)
	}

	return mcp.NewToolResultText(string(jsonBytes)), nil
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

	jsonBytes, err := json.MarshalIndent(diff, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("failed to marshal comparison: %w", err)
	}

	return mcp.NewToolResultText(string(jsonBytes)), nil
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

	jsonBytes, err := json.MarshalIndent(analysis, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("failed to marshal worker analysis: %w", err)
	}

	return mcp.NewToolResultText(string(jsonBytes)), nil
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

	jsonBytes, err := json.MarshalIndent(analysis, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("failed to marshal crypto analysis: %w", err)
	}

	return mcp.NewToolResultText(string(jsonBytes)), nil
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

	jsonBytes, err := json.MarshalIndent(analysis, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("failed to marshal JS crypto analysis: %w", err)
	}

	return mcp.NewToolResultText(string(jsonBytes)), nil
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

	jsonBytes, err := json.MarshalIndent(analysis, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("failed to marshal contention analysis: %w", err)
	}

	return mcp.NewToolResultText(string(jsonBytes)), nil
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

	jsonBytes, err := json.MarshalIndent(analysis, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("failed to marshal scaling analysis: %w", err)
	}

	return mcp.NewToolResultText(string(jsonBytes)), nil
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

	jsonBytes, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("failed to marshal scaling comparison: %w", err)
	}

	return mcp.NewToolResultText(string(jsonBytes)), nil
}
