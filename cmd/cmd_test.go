package cmd

import (
	"os"
	"strings"
	"testing"

	"github.com/CedricHerzog/perfowl/internal/parser"
	"github.com/CedricHerzog/perfowl/internal/testutil"
)

func TestBuildSummary_BasicFields(t *testing.T) {
	profile := testutil.NewProfileBuilder().
		WithMeta(parser.Meta{
			Platform:     "Linux",
			OSCPU:        "x86_64",
			Product:      "Firefox",
			AppBuildID:   "20240101",
			CPUName:      "Intel",
			PhysicalCPUs: 4,
			LogicalCPUs:  8,
		}).
		WithDuration(5000).
		WithThread(testutil.NewThreadBuilder("GeckoMain").AsMainThread().Build()).
		WithThread(testutil.NewThreadBuilder("DOM Worker").Build()).
		Build()

	summary := buildSummary(profile, parser.BrowserFirefox)

	if summary.Duration != 5 {
		t.Errorf("Duration = %v, want 5", summary.Duration)
	}
	if summary.BrowserType != "firefox" {
		t.Errorf("BrowserType = %v, want firefox", summary.BrowserType)
	}
	if summary.Platform != "Linux" {
		t.Errorf("Platform = %v, want Linux", summary.Platform)
	}
	if summary.Product != "Firefox" {
		t.Errorf("Product = %v, want Firefox", summary.Product)
	}
	if summary.ThreadCount != 2 {
		t.Errorf("ThreadCount = %v, want 2", summary.ThreadCount)
	}
	if summary.MainThreadCount != 1 {
		t.Errorf("MainThreadCount = %v, want 1", summary.MainThreadCount)
	}
}

func TestBuildSummary_WithExtensions(t *testing.T) {
	profile := testutil.ProfileWithExtensions()

	summary := buildSummary(profile, parser.BrowserFirefox)

	if summary.ExtensionCount != 2 {
		t.Errorf("ExtensionCount = %v, want 2", summary.ExtensionCount)
	}
	if len(summary.Extensions) != 2 {
		t.Errorf("Extensions length = %v, want 2", len(summary.Extensions))
	}
}

func TestBuildSummary_WithCategories(t *testing.T) {
	profile := testutil.ProfileWithCategories()

	summary := buildSummary(profile, parser.BrowserFirefox)

	if len(summary.Categories) == 0 {
		t.Error("expected categories")
	}
}

func TestBuildSummary_Markers(t *testing.T) {
	mb := testutil.NewMarkerBuilder()
	mb.AddGCMajor(0, 10)
	mb.AddGCMajor(100, 20)
	markers, strings := mb.Build()

	profile := testutil.NewProfileBuilder().
		WithDuration(1000).
		WithThread(testutil.NewThreadBuilder("GeckoMain").
			AsMainThread().
			WithMarkers(markers).
			WithStringArray(strings).
			Build()).
		Build()

	summary := buildSummary(profile, parser.BrowserFirefox)

	if summary.TotalMarkers != 2 {
		t.Errorf("TotalMarkers = %v, want 2", summary.TotalMarkers)
	}
}

func TestBuildSummary_Samples(t *testing.T) {
	sb := testutil.NewSamplesBuilder()
	for i := 0; i < 100; i++ {
		sb.AddSample(0, float64(i))
	}

	profile := testutil.NewProfileBuilder().
		WithDuration(1000).
		WithThread(testutil.NewThreadBuilder("GeckoMain").
			AsMainThread().
			WithSamples(sb.Build()).
			Build()).
		Build()

	summary := buildSummary(profile, parser.BrowserFirefox)

	if summary.TotalSamples != 100 {
		t.Errorf("TotalSamples = %v, want 100", summary.TotalSamples)
	}
}

func TestBuildSummary_Chrome(t *testing.T) {
	profile := testutil.NewProfileBuilder().
		WithDuration(3000).
		WithThread(testutil.NewThreadBuilder("CrBrowserMain").AsMainThread().Build()).
		Build()

	summary := buildSummary(profile, parser.BrowserChrome)

	if summary.BrowserType != "chrome" {
		t.Errorf("BrowserType = %v, want chrome", summary.BrowserType)
	}
}

func TestProfileSummary_Struct(t *testing.T) {
	summary := ProfileSummary{
		BrowserType:     "firefox",
		Duration:        10.5,
		Platform:        "Linux",
		OSCPU:           "x86_64",
		Product:         "Firefox",
		BuildID:         "20240101",
		CPUName:         "Intel",
		PhysicalCPUs:    4,
		LogicalCPUs:     8,
		ThreadCount:     10,
		MainThreadCount: 1,
		ExtensionCount:  2,
		Extensions:      map[string]string{"ext1": "Extension 1"},
		Features:        []string{"feature1", "feature2"},
		Categories:      []string{"JavaScript", "Layout"},
		TotalMarkers:    100,
		TotalSamples:    5000,
	}

	// Verify all fields are accessible
	if summary.BrowserType != "firefox" {
		t.Error("BrowserType field mismatch")
	}
	if summary.Duration != 10.5 {
		t.Error("Duration field mismatch")
	}
	if summary.Platform != "Linux" {
		t.Error("Platform field mismatch")
	}
	if summary.OSCPU != "x86_64" {
		t.Error("OSCPU field mismatch")
	}
	if summary.Product != "Firefox" {
		t.Error("Product field mismatch")
	}
	if summary.BuildID != "20240101" {
		t.Error("BuildID field mismatch")
	}
	if summary.CPUName != "Intel" {
		t.Error("CPUName field mismatch")
	}
	if summary.PhysicalCPUs != 4 {
		t.Error("PhysicalCPUs field mismatch")
	}
	if summary.LogicalCPUs != 8 {
		t.Error("LogicalCPUs field mismatch")
	}
	if summary.ThreadCount != 10 {
		t.Error("ThreadCount field mismatch")
	}
	if summary.MainThreadCount != 1 {
		t.Error("MainThreadCount field mismatch")
	}
	if summary.ExtensionCount != 2 {
		t.Error("ExtensionCount field mismatch")
	}
	if len(summary.Extensions) != 1 {
		t.Error("Extensions field mismatch")
	}
	if len(summary.Features) != 2 {
		t.Error("Features field mismatch")
	}
	if len(summary.Categories) != 2 {
		t.Error("Categories field mismatch")
	}
	if summary.TotalMarkers != 100 {
		t.Error("TotalMarkers field mismatch")
	}
	if summary.TotalSamples != 5000 {
		t.Error("TotalSamples field mismatch")
	}
}

func TestRootCmd_Flags(t *testing.T) {
	// Test that persistent flags are defined
	flag := rootCmd.PersistentFlags().Lookup("profile")
	if flag == nil {
		t.Error("expected 'profile' flag to be defined")
	}
	if flag.Shorthand != "p" {
		t.Errorf("profile flag shorthand = %s, want 'p'", flag.Shorthand)
	}

	flag = rootCmd.PersistentFlags().Lookup("output")
	if flag == nil {
		t.Error("expected 'output' flag to be defined")
	}
	if flag.Shorthand != "o" {
		t.Errorf("output flag shorthand = %s, want 'o'", flag.Shorthand)
	}

	flag = rootCmd.PersistentFlags().Lookup("browser")
	if flag == nil {
		t.Error("expected 'browser' flag to be defined")
	}
	if flag.Shorthand != "b" {
		t.Errorf("browser flag shorthand = %s, want 'b'", flag.Shorthand)
	}
}

func TestRootCmd_Definition(t *testing.T) {
	if rootCmd.Use != "perfowl" {
		t.Errorf("rootCmd.Use = %s, want 'perfowl'", rootCmd.Use)
	}
	if rootCmd.Short == "" {
		t.Error("expected rootCmd.Short to be non-empty")
	}
	if rootCmd.Long == "" {
		t.Error("expected rootCmd.Long to be non-empty")
	}
}

func TestSummaryCmd_Definition(t *testing.T) {
	if summaryCmd.Use != "summary" {
		t.Errorf("summaryCmd.Use = %s, want 'summary'", summaryCmd.Use)
	}
	if summaryCmd.Short == "" {
		t.Error("expected summaryCmd.Short to be non-empty")
	}
}

func TestRunSummary_MissingProfile(t *testing.T) {
	// Save original value
	originalPath := profilePath
	defer func() { profilePath = originalPath }()

	// Set to empty
	profilePath = ""

	err := runSummary(summaryCmd, []string{})
	if err == nil {
		t.Error("expected error for missing profile path")
	}
	if !strings.Contains(err.Error(), "profile path is required") {
		t.Errorf("unexpected error message: %v", err)
	}
}

func TestRunSummary_InvalidProfile(t *testing.T) {
	// Save original values
	originalPath := profilePath
	originalBrowser := browserType
	defer func() {
		profilePath = originalPath
		browserType = originalBrowser
	}()

	profilePath = "/nonexistent/profile.json"
	browserType = "auto"

	err := runSummary(summaryCmd, []string{})
	if err == nil {
		t.Error("expected error for non-existent profile")
	}
	if !strings.Contains(err.Error(), "failed to load profile") {
		t.Errorf("unexpected error message: %v", err)
	}
}

func TestOutputFunctions_NoPanic(t *testing.T) {
	summary := ProfileSummary{
		BrowserType:    "firefox",
		Duration:       5.0,
		Platform:       "Linux",
		OSCPU:          "x86_64",
		Product:        "Firefox",
		BuildID:        "123",
		CPUName:        "Intel",
		PhysicalCPUs:   4,
		LogicalCPUs:    8,
		ThreadCount:    5,
		MainThreadCount: 1,
		ExtensionCount: 1,
		Extensions:     map[string]string{"ext1@test.com": "Test Extension"},
		Features:       []string{"js", "screenshots"},
		Categories:     []string{"JavaScript"},
		TotalMarkers:   100,
		TotalSamples:   1000,
	}

	// These should not panic
	t.Run("outputText", func(t *testing.T) {
		err := outputText(summary)
		if err != nil {
			t.Errorf("outputText returned error: %v", err)
		}
	})

	t.Run("outputMarkdown", func(t *testing.T) {
		err := outputMarkdown(summary)
		if err != nil {
			t.Errorf("outputMarkdown returned error: %v", err)
		}
	})

	t.Run("outputJSON", func(t *testing.T) {
		err := outputJSON(summary)
		if err != nil {
			t.Errorf("outputJSON returned error: %v", err)
		}
	})
}

func TestOutputFunctions_EmptyBrowserType(t *testing.T) {
	summary := ProfileSummary{
		BrowserType: "",
		Duration:    5.0,
		Features:    []string{},
		Categories:  []string{},
	}

	// Should not panic even with empty browser type
	err := outputText(summary)
	if err != nil {
		t.Errorf("outputText returned error: %v", err)
	}

	err = outputMarkdown(summary)
	if err != nil {
		t.Errorf("outputMarkdown returned error: %v", err)
	}
}

func TestOutputFunctions_NoExtensions(t *testing.T) {
	summary := ProfileSummary{
		BrowserType:    "chrome",
		Duration:       10.0,
		ExtensionCount: 0,
		Extensions:     map[string]string{},
		Features:       []string{"js"},
		Categories:     []string{"JavaScript"},
	}

	// Should not panic with no extensions
	err := outputText(summary)
	if err != nil {
		t.Errorf("outputText returned error: %v", err)
	}

	err = outputMarkdown(summary)
	if err != nil {
		t.Errorf("outputMarkdown returned error: %v", err)
	}
}

func TestBuildSummary_EmptyProfile(t *testing.T) {
	profile := testutil.MinimalProfile()

	summary := buildSummary(profile, parser.BrowserUnknown)

	if summary.ThreadCount != 0 {
		t.Errorf("ThreadCount = %v, want 0", summary.ThreadCount)
	}
	if summary.MainThreadCount != 0 {
		t.Errorf("MainThreadCount = %v, want 0", summary.MainThreadCount)
	}
	if summary.TotalMarkers != 0 {
		t.Errorf("TotalMarkers = %v, want 0", summary.TotalMarkers)
	}
	if summary.TotalSamples != 0 {
		t.Errorf("TotalSamples = %v, want 0", summary.TotalSamples)
	}
}

// Test bottlenecks command definitions
func TestBottlenecksCmd_Definition(t *testing.T) {
	if bottlenecksCmd.Use != "bottlenecks" {
		t.Errorf("bottlenecksCmd.Use = %s, want 'bottlenecks'", bottlenecksCmd.Use)
	}
	if bottlenecksCmd.Short == "" {
		t.Error("expected bottlenecksCmd.Short to be non-empty")
	}

	flag := bottlenecksCmd.Flags().Lookup("min-severity")
	if flag == nil {
		t.Error("expected 'min-severity' flag to be defined")
	}
}

// Test markers command definitions
func TestMarkersCmd_Definition(t *testing.T) {
	if markersCmd.Use != "markers" {
		t.Errorf("markersCmd.Use = %s, want 'markers'", markersCmd.Use)
	}
	if markersCmd.Short == "" {
		t.Error("expected markersCmd.Short to be non-empty")
	}
}

// Test workers command definitions
func TestWorkersCmd_Definition(t *testing.T) {
	if workersCmd.Use != "workers" {
		t.Errorf("workersCmd.Use = %s, want 'workers'", workersCmd.Use)
	}
}

// Test crypto command definitions
func TestCryptoCmd_Definition(t *testing.T) {
	if cryptoCmd.Use != "crypto" {
		t.Errorf("cryptoCmd.Use = %s, want 'crypto'", cryptoCmd.Use)
	}
}

// Test contention command definitions
func TestContentionCmd_Definition(t *testing.T) {
	if contentionCmd.Use != "contention" {
		t.Errorf("contentionCmd.Use = %s, want 'contention'", contentionCmd.Use)
	}
}

// Test scaling command definitions
func TestScalingCmd_Definition(t *testing.T) {
	if scalingCmd.Use != "scaling" {
		t.Errorf("scalingCmd.Use = %s, want 'scaling'", scalingCmd.Use)
	}
}

// Test extensions command definitions
func TestExtensionsCmd_Definition(t *testing.T) {
	if extensionsCmd.Use != "extensions" {
		t.Errorf("extensionsCmd.Use = %s, want 'extensions'", extensionsCmd.Use)
	}
}

// Test batch command definitions
func TestBatchCmd_Definition(t *testing.T) {
	if batchCmd.Use != "batch" {
		t.Errorf("batchCmd.Use = %s, want 'batch'", batchCmd.Use)
	}
}

// Test mcp command definitions
func TestMCPCmd_Definition(t *testing.T) {
	if mcpCmd.Use != "mcp" {
		t.Errorf("mcpCmd.Use = %s, want 'mcp'", mcpCmd.Use)
	}
}

// Test jscrypto command definitions
func TestJSCryptoCmd_Definition(t *testing.T) {
	if jscryptoCmd.Use != "jscrypto" {
		t.Errorf("jscryptoCmd.Use = %s, want 'jscrypto'", jscryptoCmd.Use)
	}
}

// Test runBottlenecks error cases
func TestRunBottlenecks_MissingProfile(t *testing.T) {
	originalPath := profilePath
	defer func() { profilePath = originalPath }()

	profilePath = ""

	err := runBottlenecks(bottlenecksCmd, []string{})
	if err == nil {
		t.Error("expected error for missing profile path")
	}
	if !strings.Contains(err.Error(), "profile path is required") {
		t.Errorf("unexpected error message: %v", err)
	}
}

func TestRunBottlenecks_InvalidProfile(t *testing.T) {
	originalPath := profilePath
	originalBrowser := browserType
	defer func() {
		profilePath = originalPath
		browserType = originalBrowser
	}()

	profilePath = "/nonexistent/profile.json"
	browserType = "auto"

	err := runBottlenecks(bottlenecksCmd, []string{})
	if err == nil {
		t.Error("expected error for non-existent profile")
	}
	if !strings.Contains(err.Error(), "failed to load profile") {
		t.Errorf("unexpected error message: %v", err)
	}
}

// Test runMarkers error cases
func TestRunMarkers_MissingProfile(t *testing.T) {
	originalPath := profilePath
	defer func() { profilePath = originalPath }()

	profilePath = ""

	err := runMarkers(markersCmd, []string{})
	if err == nil {
		t.Error("expected error for missing profile path")
	}
}

// Test runWorkers error cases
func TestRunWorkers_MissingProfile(t *testing.T) {
	originalPath := profilePath
	defer func() { profilePath = originalPath }()

	profilePath = ""

	err := runWorkers(workersCmd, []string{})
	if err == nil {
		t.Error("expected error for missing profile path")
	}
}

// Test runCrypto error cases
func TestRunCrypto_MissingProfile(t *testing.T) {
	originalPath := profilePath
	defer func() { profilePath = originalPath }()

	profilePath = ""

	err := runCrypto(cryptoCmd, []string{})
	if err == nil {
		t.Error("expected error for missing profile path")
	}
}

// Test runJSCrypto error cases
func TestRunJSCrypto_MissingProfile(t *testing.T) {
	originalPath := profilePath
	defer func() { profilePath = originalPath }()

	profilePath = ""

	err := runJSCrypto(jscryptoCmd, []string{})
	if err == nil {
		t.Error("expected error for missing profile path")
	}
}

// Test runContention error cases
func TestRunContention_MissingProfile(t *testing.T) {
	originalPath := profilePath
	defer func() { profilePath = originalPath }()

	profilePath = ""

	err := runContention(contentionCmd, []string{})
	if err == nil {
		t.Error("expected error for missing profile path")
	}
}

// Test runScaling error cases
func TestRunScaling_MissingProfile(t *testing.T) {
	originalPath := profilePath
	defer func() { profilePath = originalPath }()

	profilePath = ""

	err := runScaling(scalingCmd, []string{})
	if err == nil {
		t.Error("expected error for missing profile path")
	}
}

// Test runExtensions error cases
func TestRunExtensions_MissingProfile(t *testing.T) {
	originalPath := profilePath
	defer func() { profilePath = originalPath }()

	profilePath = ""

	err := runExtensions(extensionsCmd, []string{})
	if err == nil {
		t.Error("expected error for missing profile path")
	}
}

// Integration tests with actual temp profile files

func TestRunSummary_Success(t *testing.T) {
	profile := testutil.ProfileWithWorkers(2)
	path := testutil.TempProfileFile(t, profile)

	originalPath := profilePath
	originalBrowser := browserType
	originalFormat := outputFormat
	defer func() {
		profilePath = originalPath
		browserType = originalBrowser
		outputFormat = originalFormat
	}()

	profilePath = path
	browserType = "auto"

	// Test text output
	outputFormat = "text"
	err := runSummary(summaryCmd, []string{})
	if err != nil {
		t.Errorf("runSummary text format error: %v", err)
	}

	// Test markdown output
	outputFormat = "markdown"
	err = runSummary(summaryCmd, []string{})
	if err != nil {
		t.Errorf("runSummary markdown format error: %v", err)
	}

	// Test JSON output
	outputFormat = "json"
	err = runSummary(summaryCmd, []string{})
	if err != nil {
		t.Errorf("runSummary json format error: %v", err)
	}
}

func TestRunBottlenecks_Success(t *testing.T) {
	// Use ProfileWithBottlenecks which has all types of bottleneck markers
	profile := testutil.ProfileWithBottlenecks()
	path := testutil.TempProfileFile(t, profile)

	originalPath := profilePath
	originalBrowser := browserType
	originalFormat := outputFormat
	originalSeverity := minSeverity
	defer func() {
		profilePath = originalPath
		browserType = originalBrowser
		outputFormat = originalFormat
		minSeverity = originalSeverity
	}()

	profilePath = path
	browserType = "auto"
	minSeverity = ""

	// Test text output
	outputFormat = "text"
	err := runBottlenecks(bottlenecksCmd, []string{})
	if err != nil {
		t.Errorf("runBottlenecks text format error: %v", err)
	}

	// Test markdown output
	outputFormat = "markdown"
	err = runBottlenecks(bottlenecksCmd, []string{})
	if err != nil {
		t.Errorf("runBottlenecks markdown format error: %v", err)
	}

	// Test JSON output
	outputFormat = "json"
	err = runBottlenecks(bottlenecksCmd, []string{})
	if err != nil {
		t.Errorf("runBottlenecks json format error: %v", err)
	}

	// Test with severity filter
	outputFormat = "text"
	minSeverity = "high"
	err = runBottlenecks(bottlenecksCmd, []string{})
	if err != nil {
		t.Errorf("runBottlenecks with severity filter error: %v", err)
	}
}

func TestRunMarkers_Success(t *testing.T) {
	mb := testutil.NewMarkerBuilder()
	mb.AddGCMajor(0, 10)
	mb.AddGCMajor(100, 20)
	markers, stringArray := mb.Build()

	profile := testutil.NewProfileBuilder().
		WithDuration(1000).
		WithThread(testutil.NewThreadBuilder("GeckoMain").
			AsMainThread().
			WithMarkers(markers).
			WithStringArray(stringArray).
			Build()).
		Build()
	path := testutil.TempProfileFile(t, profile)

	originalPath := profilePath
	originalBrowser := browserType
	originalFormat := outputFormat
	defer func() {
		profilePath = originalPath
		browserType = originalBrowser
		outputFormat = originalFormat
	}()

	profilePath = path
	browserType = "auto"

	// Test text output
	outputFormat = "text"
	err := runMarkers(markersCmd, []string{})
	if err != nil {
		t.Errorf("runMarkers text format error: %v", err)
	}

	// Test markdown output
	outputFormat = "markdown"
	err = runMarkers(markersCmd, []string{})
	if err != nil {
		t.Errorf("runMarkers markdown format error: %v", err)
	}

	// Test JSON output
	outputFormat = "json"
	err = runMarkers(markersCmd, []string{})
	if err != nil {
		t.Errorf("runMarkers json format error: %v", err)
	}
}

func TestRunWorkers_Success(t *testing.T) {
	// Use ProfileWithWorkersData which has comprehensive worker data
	profile := testutil.ProfileWithWorkersData()
	path := testutil.TempProfileFile(t, profile)

	originalPath := profilePath
	originalBrowser := browserType
	originalFormat := outputFormat
	defer func() {
		profilePath = originalPath
		browserType = originalBrowser
		outputFormat = originalFormat
	}()

	profilePath = path
	browserType = "auto"

	// Test text output
	outputFormat = "text"
	err := runWorkers(workersCmd, []string{})
	if err != nil {
		t.Errorf("runWorkers text format error: %v", err)
	}

	// Test markdown output
	outputFormat = "markdown"
	err = runWorkers(workersCmd, []string{})
	if err != nil {
		t.Errorf("runWorkers markdown format error: %v", err)
	}

	// Test JSON output
	outputFormat = "json"
	err = runWorkers(workersCmd, []string{})
	if err != nil {
		t.Errorf("runWorkers json format error: %v", err)
	}
}

func TestRunCrypto_Success(t *testing.T) {
	// Use ProfileWithCrypto which has actual crypto operations
	profile := testutil.ProfileWithCrypto()
	path := testutil.TempProfileFile(t, profile)

	originalPath := profilePath
	originalBrowser := browserType
	originalFormat := outputFormat
	defer func() {
		profilePath = originalPath
		browserType = originalBrowser
		outputFormat = originalFormat
	}()

	profilePath = path
	browserType = "auto"

	// Test text output
	outputFormat = "text"
	err := runCrypto(cryptoCmd, []string{})
	if err != nil {
		t.Errorf("runCrypto text format error: %v", err)
	}

	// Test markdown output
	outputFormat = "markdown"
	err = runCrypto(cryptoCmd, []string{})
	if err != nil {
		t.Errorf("runCrypto markdown format error: %v", err)
	}

	// Test JSON output
	outputFormat = "json"
	err = runCrypto(cryptoCmd, []string{})
	if err != nil {
		t.Errorf("runCrypto json format error: %v", err)
	}
}

func TestRunJSCrypto_Success(t *testing.T) {
	// Use ProfileWithJSCrypto which has openpgp/crypto worker data
	profile := testutil.ProfileWithJSCrypto()
	path := testutil.TempProfileFile(t, profile)

	originalPath := profilePath
	originalBrowser := browserType
	originalFormat := outputFormat
	defer func() {
		profilePath = originalPath
		browserType = originalBrowser
		outputFormat = originalFormat
	}()

	profilePath = path
	browserType = "auto"

	// Test text output
	outputFormat = "text"
	err := runJSCrypto(jscryptoCmd, []string{})
	if err != nil {
		t.Errorf("runJSCrypto text format error: %v", err)
	}

	// Test markdown output
	outputFormat = "markdown"
	err = runJSCrypto(jscryptoCmd, []string{})
	if err != nil {
		t.Errorf("runJSCrypto markdown format error: %v", err)
	}

	// Test JSON output
	outputFormat = "json"
	err = runJSCrypto(jscryptoCmd, []string{})
	if err != nil {
		t.Errorf("runJSCrypto json format error: %v", err)
	}
}

func TestRunContention_Success(t *testing.T) {
	// Use ProfileWithContentionData which has GC/IPC contention
	profile := testutil.ProfileWithContentionData()
	path := testutil.TempProfileFile(t, profile)

	originalPath := profilePath
	originalBrowser := browserType
	originalFormat := outputFormat
	defer func() {
		profilePath = originalPath
		browserType = originalBrowser
		outputFormat = originalFormat
	}()

	profilePath = path
	browserType = "auto"

	// Test text output
	outputFormat = "text"
	err := runContention(contentionCmd, []string{})
	if err != nil {
		t.Errorf("runContention text format error: %v", err)
	}

	// Test markdown output
	outputFormat = "markdown"
	err = runContention(contentionCmd, []string{})
	if err != nil {
		t.Errorf("runContention markdown format error: %v", err)
	}

	// Test JSON output
	outputFormat = "json"
	err = runContention(contentionCmd, []string{})
	if err != nil {
		t.Errorf("runContention json format error: %v", err)
	}
}

func TestRunScaling_Success(t *testing.T) {
	profile := testutil.ProfileWithWorkers(4)
	path := testutil.TempProfileFile(t, profile)

	originalPath := profilePath
	originalBrowser := browserType
	originalFormat := outputFormat
	defer func() {
		profilePath = originalPath
		browserType = originalBrowser
		outputFormat = originalFormat
	}()

	profilePath = path
	browserType = "auto"

	// Test text output
	outputFormat = "text"
	err := runScaling(scalingCmd, []string{})
	if err != nil {
		t.Errorf("runScaling text format error: %v", err)
	}

	// Test markdown output
	outputFormat = "markdown"
	err = runScaling(scalingCmd, []string{})
	if err != nil {
		t.Errorf("runScaling markdown format error: %v", err)
	}

	// Test JSON output
	outputFormat = "json"
	err = runScaling(scalingCmd, []string{})
	if err != nil {
		t.Errorf("runScaling json format error: %v", err)
	}
}

func TestRunExtensions_Success(t *testing.T) {
	// Use ProfileWithExtensionActivity which has actual extension samples
	profile := testutil.ProfileWithExtensionActivity()
	path := testutil.TempProfileFile(t, profile)

	originalPath := profilePath
	originalBrowser := browserType
	originalFormat := outputFormat
	defer func() {
		profilePath = originalPath
		browserType = originalBrowser
		outputFormat = originalFormat
	}()

	profilePath = path
	browserType = "auto"

	// Test text output
	outputFormat = "text"
	err := runExtensions(extensionsCmd, []string{})
	if err != nil {
		t.Errorf("runExtensions text format error: %v", err)
	}

	// Test markdown output
	outputFormat = "markdown"
	err = runExtensions(extensionsCmd, []string{})
	if err != nil {
		t.Errorf("runExtensions markdown format error: %v", err)
	}

	// Test JSON output
	outputFormat = "json"
	err = runExtensions(extensionsCmd, []string{})
	if err != nil {
		t.Errorf("runExtensions json format error: %v", err)
	}
}

// Batch command tests
func TestRunBatch_NoProfiles(t *testing.T) {
	originalConfig := batchConfigFile
	originalProfiles := batchProfiles
	defer func() {
		batchConfigFile = originalConfig
		batchProfiles = originalProfiles
	}()

	batchConfigFile = ""
	batchProfiles = ""

	err := runBatch(batchCmd, []string{})
	if err == nil {
		t.Error("expected error for no profiles")
	}
	if !strings.Contains(err.Error(), "no profiles specified") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestRunBatch_InlineProfiles(t *testing.T) {
	profile := testutil.ProfileWithWorkers(2)
	path := testutil.TempProfileFile(t, profile)

	originalConfig := batchConfigFile
	originalProfiles := batchProfiles
	originalFormat := outputFormat
	originalChartType := batchChartType
	defer func() {
		batchConfigFile = originalConfig
		batchProfiles = originalProfiles
		outputFormat = originalFormat
		batchChartType = originalChartType
	}()

	batchConfigFile = ""
	batchProfiles = path + ":2:Test"
	batchChartType = "wall_clock"

	// Test text output
	outputFormat = "text"
	err := runBatch(batchCmd, []string{})
	if err != nil {
		t.Errorf("runBatch text format error: %v", err)
	}

	// Test markdown output
	outputFormat = "markdown"
	err = runBatch(batchCmd, []string{})
	if err != nil {
		t.Errorf("runBatch markdown format error: %v", err)
	}

	// Test JSON output
	outputFormat = "json"
	err = runBatch(batchCmd, []string{})
	if err != nil {
		t.Errorf("runBatch json format error: %v", err)
	}

	// Test SVG output
	outputFormat = "svg"
	err = runBatch(batchCmd, []string{})
	if err != nil {
		t.Errorf("runBatch svg format error: %v", err)
	}
}

func TestRunBatch_ConfigFile(t *testing.T) {
	profile := testutil.ProfileWithWorkers(2)
	profilePath := testutil.TempProfileFile(t, profile)

	// Create a temp config file
	tmpDir := t.TempDir()
	configPath := tmpDir + "/batch.yaml"
	configContent := `profiles:
  - path: ` + profilePath + `
    workers: 2
    label: Test
`
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("failed to write config file: %v", err)
	}

	originalConfig := batchConfigFile
	originalProfiles := batchProfiles
	originalFormat := outputFormat
	defer func() {
		batchConfigFile = originalConfig
		batchProfiles = originalProfiles
		outputFormat = originalFormat
	}()

	batchConfigFile = configPath
	batchProfiles = ""
	outputFormat = "text"

	err := runBatch(batchCmd, []string{})
	if err != nil {
		t.Errorf("runBatch with config file error: %v", err)
	}
}

func TestRunBatch_SVGFile(t *testing.T) {
	profile := testutil.ProfileWithWorkers(2)
	path := testutil.TempProfileFile(t, profile)

	tmpDir := t.TempDir()
	svgPath := tmpDir + "/chart.svg"

	originalConfig := batchConfigFile
	originalProfiles := batchProfiles
	originalFormat := outputFormat
	originalSVGPath := batchSVGPath
	defer func() {
		batchConfigFile = originalConfig
		batchProfiles = originalProfiles
		outputFormat = originalFormat
		batchSVGPath = originalSVGPath
	}()

	batchConfigFile = ""
	batchProfiles = path + ":2:Test"
	outputFormat = "svg-file"
	batchSVGPath = svgPath

	err := runBatch(batchCmd, []string{})
	if err != nil {
		t.Errorf("runBatch svg-file format error: %v", err)
	}

	// Check file was created
	if _, err := os.Stat(svgPath); os.IsNotExist(err) {
		t.Error("expected SVG file to be created")
	}
}

func TestRunBatch_InvalidInlineFormat(t *testing.T) {
	originalConfig := batchConfigFile
	originalProfiles := batchProfiles
	defer func() {
		batchConfigFile = originalConfig
		batchProfiles = originalProfiles
	}()

	batchConfigFile = ""
	batchProfiles = "invalid-format" // missing :workers:label

	err := runBatch(batchCmd, []string{})
	if err == nil {
		t.Error("expected error for invalid inline format")
	}
}

func TestRunBatch_InvalidConfigFile(t *testing.T) {
	originalConfig := batchConfigFile
	originalProfiles := batchProfiles
	defer func() {
		batchConfigFile = originalConfig
		batchProfiles = originalProfiles
	}()

	batchConfigFile = "/nonexistent/config.yaml"
	batchProfiles = ""

	err := runBatch(batchCmd, []string{})
	if err == nil {
		t.Error("expected error for nonexistent config file")
	}
}

// Scaling comparison tests
func TestRunScaling_WithComparison(t *testing.T) {
	profile1 := testutil.ProfileWithWorkers(2)
	profile2 := testutil.ProfileWithWorkers(4)
	path1 := testutil.TempProfileFile(t, profile1)
	path2 := testutil.TempProfileFile(t, profile2)

	originalPath := profilePath
	originalBrowser := browserType
	originalFormat := outputFormat
	originalCompare := compareProfile
	defer func() {
		profilePath = originalPath
		browserType = originalBrowser
		outputFormat = originalFormat
		compareProfile = originalCompare
	}()

	profilePath = path1
	compareProfile = path2
	browserType = "auto"

	// Test text output
	outputFormat = "text"
	err := runScaling(scalingCmd, []string{})
	if err != nil {
		t.Errorf("runScaling comparison text format error: %v", err)
	}

	// Test markdown output
	outputFormat = "markdown"
	err = runScaling(scalingCmd, []string{})
	if err != nil {
		t.Errorf("runScaling comparison markdown format error: %v", err)
	}

	// Test JSON output
	outputFormat = "json"
	err = runScaling(scalingCmd, []string{})
	if err != nil {
		t.Errorf("runScaling comparison json format error: %v", err)
	}
}
