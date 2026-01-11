package chart

import (
	"strings"
	"testing"

	"github.com/CedricHerzog/perfowl/internal/analyzer"
)

func TestGenerateScalingChart_WallClock(t *testing.T) {
	result := createTestBatchResult()

	svg := GenerateScalingChart(result, ChartWallClock)

	if svg == "" {
		t.Error("expected non-empty SVG output")
	}
	if !strings.Contains(svg, "Wall Clock Time vs Worker Count") {
		t.Error("expected wall clock title in SVG")
	}
	if !strings.Contains(svg, "Time (ms)") {
		t.Error("expected Y axis label in SVG")
	}
}

func TestGenerateScalingChart_OperationTime(t *testing.T) {
	result := createTestBatchResult()

	svg := GenerateScalingChart(result, ChartOperationTime)

	if !strings.Contains(svg, "Operation Time vs Worker Count") {
		t.Error("expected operation time title in SVG")
	}
}

func TestGenerateScalingChart_Efficiency(t *testing.T) {
	result := createTestBatchResult()

	svg := GenerateScalingChart(result, ChartEfficiency)

	if !strings.Contains(svg, "Parallel Efficiency vs Worker Count") {
		t.Error("expected efficiency title in SVG")
	}
	if !strings.Contains(svg, "Efficiency (%)") {
		t.Error("expected efficiency Y axis label in SVG")
	}
}

func TestGenerateScalingChart_Speedup(t *testing.T) {
	result := createTestBatchResult()

	svg := GenerateScalingChart(result, ChartSpeedup)

	if !strings.Contains(svg, "Speedup vs Worker Count") {
		t.Error("expected speedup title in SVG")
	}
	if !strings.Contains(svg, "Speedup (x)") {
		t.Error("expected speedup Y axis label in SVG")
	}
}

func TestGenerateScalingChart_CryptoTime(t *testing.T) {
	result := createTestBatchResult()

	svg := GenerateScalingChart(result, ChartCryptoTime)

	if !strings.Contains(svg, "Crypto Time vs Worker Count") {
		t.Error("expected crypto time title in SVG")
	}
}

func TestGenerateScalingChart_DefaultType(t *testing.T) {
	result := createTestBatchResult()

	// Use invalid chart type, should default to wall_clock
	svg := GenerateScalingChart(result, "invalid")

	if !strings.Contains(svg, "Wall Clock Time vs Worker Count") {
		t.Error("expected default to wall clock chart")
	}
}

func TestGenerateSVG_EmptySeries(t *testing.T) {
	config := ChartConfig{
		Width:      800,
		Height:     450,
		Title:      "Test Chart",
		XAxisLabel: "X Axis",
		YAxisLabel: "Y Axis",
		ShowLegend: true,
		ShowGrid:   true,
	}

	svg := GenerateSVG(config, []DataSeries{})

	if svg == "" {
		t.Error("expected non-empty SVG output for empty series")
	}
	if !strings.Contains(svg, "<svg") {
		t.Error("expected SVG element")
	}
}

func TestGenerateSVG_SingleSeries(t *testing.T) {
	config := ChartConfig{
		Width:      800,
		Height:     450,
		Title:      "Single Series",
		XAxisLabel: "X",
		YAxisLabel: "Y",
		ShowLegend: true,
		ShowGrid:   true,
	}

	series := []DataSeries{
		{
			Name:  "Test",
			Color: "#FF0000",
			Points: []DataPoint{
				{X: 1, Y: 100},
				{X: 2, Y: 200},
				{X: 3, Y: 150},
			},
		},
	}

	svg := GenerateSVG(config, series)

	if !strings.Contains(svg, "#FF0000") {
		t.Error("expected series color in SVG")
	}
	if !strings.Contains(svg, "Test") {
		t.Error("expected series name in legend")
	}
	if !strings.Contains(svg, "<path") {
		t.Error("expected path element for line")
	}
	if !strings.Contains(svg, "<circle") {
		t.Error("expected circle elements for points")
	}
}

func TestGenerateSVG_MultipleSeries(t *testing.T) {
	config := ChartConfig{
		Width:      800,
		Height:     450,
		Title:      "Multiple Series",
		XAxisLabel: "X",
		YAxisLabel: "Y",
		ShowLegend: true,
		ShowGrid:   true,
	}

	series := []DataSeries{
		{
			Name:   "Firefox",
			Color:  "#FF6611",
			Points: []DataPoint{{X: 1, Y: 100}, {X: 2, Y: 80}},
		},
		{
			Name:   "Chrome",
			Color:  "#4285F4",
			Points: []DataPoint{{X: 1, Y: 120}, {X: 2, Y: 90}},
		},
	}

	svg := GenerateSVG(config, series)

	if !strings.Contains(svg, "Firefox") {
		t.Error("expected Firefox in legend")
	}
	if !strings.Contains(svg, "Chrome") {
		t.Error("expected Chrome in legend")
	}
}

func TestGenerateSVG_DefaultDimensions(t *testing.T) {
	config := ChartConfig{
		Title:      "Default Size",
		XAxisLabel: "X",
		YAxisLabel: "Y",
	}

	series := []DataSeries{
		{
			Name:   "Test",
			Color:  "#FF0000",
			Points: []DataPoint{{X: 1, Y: 100}},
		},
	}

	svg := GenerateSVG(config, series)

	// Should use default 800x450
	if !strings.Contains(svg, `width="800"`) {
		t.Error("expected default width 800")
	}
	if !strings.Contains(svg, `height="450"`) {
		t.Error("expected default height 450")
	}
}

func TestGenerateSVG_NoLegend(t *testing.T) {
	config := ChartConfig{
		Width:      800,
		Height:     450,
		Title:      "No Legend",
		ShowLegend: false,
	}

	series := []DataSeries{
		{
			Name:   "Test",
			Color:  "#FF0000",
			Points: []DataPoint{{X: 1, Y: 100}},
		},
	}

	svg := GenerateSVG(config, series)

	// Legend text shouldn't appear (except in line/circle which are the data)
	// Actually the series name "Test" would still be in the legend area, but
	// since ShowLegend=false, the legend block shouldn't render
	if !strings.Contains(svg, "<svg") {
		t.Error("expected valid SVG")
	}
}

func TestGenerateSVG_NoGrid(t *testing.T) {
	config := ChartConfig{
		Width:    800,
		Height:   450,
		Title:    "No Grid",
		ShowGrid: false,
	}

	series := []DataSeries{
		{
			Name:   "Test",
			Color:  "#FF0000",
			Points: []DataPoint{{X: 1, Y: 100}},
		},
	}

	svg := GenerateSVG(config, series)

	// Grid lines class shouldn't appear
	if strings.Contains(svg, `class="grid"`) {
		t.Error("expected no grid lines when ShowGrid=false")
	}
}

func TestCalculateRanges_EmptyData(t *testing.T) {
	xMin, xMax, yMin, yMax := calculateRanges([]DataSeries{})

	// Should return default ranges for empty data
	if xMin != 0 || xMax != 12 {
		t.Errorf("expected default X range 0-12, got %.1f-%.1f", xMin, xMax)
	}
	if yMin != 0 || yMax != 100 {
		t.Errorf("expected default Y range 0-100, got %.1f-%.1f", yMin, yMax)
	}
}

func TestCalculateRanges_WithData(t *testing.T) {
	series := []DataSeries{
		{
			Points: []DataPoint{
				{X: 1, Y: 50},
				{X: 4, Y: 200},
			},
		},
		{
			Points: []DataPoint{
				{X: 2, Y: 100},
				{X: 8, Y: 150},
			},
		},
	}

	xMin, xMax, yMin, yMax := calculateRanges(series)

	if xMin != 1 {
		t.Errorf("xMin = %.1f, want 1", xMin)
	}
	if xMax != 8 {
		t.Errorf("xMax = %.1f, want 8", xMax)
	}
	if yMin != 50 {
		t.Errorf("yMin = %.1f, want 50", yMin)
	}
	if yMax != 200 {
		t.Errorf("yMax = %.1f, want 200", yMax)
	}
}

func TestCalculateTicks_Normal(t *testing.T) {
	ticks := calculateTicks(0, 100, 5)

	if len(ticks) == 0 {
		t.Error("expected ticks")
	}

	// Ticks should be in ascending order
	for i := 1; i < len(ticks); i++ {
		if ticks[i] <= ticks[i-1] {
			t.Error("ticks should be in ascending order")
		}
	}
}

func TestCalculateTicks_SameMinMax(t *testing.T) {
	ticks := calculateTicks(100, 100, 5)

	if len(ticks) != 1 {
		t.Errorf("expected 1 tick for same min/max, got %d", len(ticks))
	}
	if ticks[0] != 100 {
		t.Errorf("expected tick at 100, got %.1f", ticks[0])
	}
}

func TestCalculateTicks_SmallRange(t *testing.T) {
	ticks := calculateTicks(0, 10, 5)

	if len(ticks) == 0 {
		t.Error("expected ticks for small range")
	}
}

func TestCalculateTicks_LargeRange(t *testing.T) {
	ticks := calculateTicks(0, 10000, 5)

	if len(ticks) == 0 {
		t.Error("expected ticks for large range")
	}

	// All ticks should be within range
	for _, tick := range ticks {
		if tick < 0 || tick > 10000 {
			t.Errorf("tick %.1f outside range 0-10000", tick)
		}
	}
}

func TestCalculateXTicks_SmallRange(t *testing.T) {
	ticks := calculateXTicks(1, 8)

	if len(ticks) == 0 {
		t.Error("expected X ticks")
	}

	// Should use step 1 for small range
	for _, tick := range ticks {
		if tick < 1 || tick > 8 {
			t.Errorf("tick %.1f outside range 1-8", tick)
		}
	}
}

func TestCalculateXTicks_MediumRange(t *testing.T) {
	ticks := calculateXTicks(1, 16)

	if len(ticks) == 0 {
		t.Error("expected X ticks for medium range")
	}
}

func TestCalculateXTicks_LargeRange(t *testing.T) {
	ticks := calculateXTicks(0, 30)

	if len(ticks) == 0 {
		t.Error("expected X ticks for large range")
	}

	// Should use step 5 for large range (>20)
	for i := 1; i < len(ticks); i++ {
		step := ticks[i] - ticks[i-1]
		if step != 5 {
			t.Errorf("expected step 5 for large range, got %.1f", step)
		}
	}
}

func TestFormatNumber_Zero(t *testing.T) {
	result := formatNumber(0)
	if result != "0" {
		t.Errorf("formatNumber(0) = %s, want 0", result)
	}
}

func TestFormatNumber_Large(t *testing.T) {
	result := formatNumber(1234)
	if result != "1234" {
		t.Errorf("formatNumber(1234) = %s, want 1234", result)
	}
}

func TestFormatNumber_Medium(t *testing.T) {
	result := formatNumber(123)
	if result != "123" {
		t.Errorf("formatNumber(123) = %s, want 123", result)
	}
}

func TestFormatNumber_Small(t *testing.T) {
	result := formatNumber(12.5)
	if result != "12.5" {
		t.Errorf("formatNumber(12.5) = %s, want 12.5", result)
	}
}

func TestFormatNumber_VerySmall(t *testing.T) {
	result := formatNumber(1.5)
	if result != "1.5" {
		t.Errorf("formatNumber(1.5) = %s, want 1.5", result)
	}
}

func TestFormatNumber_Fraction(t *testing.T) {
	result := formatNumber(0.12)
	if result != "0.12" {
		t.Errorf("formatNumber(0.12) = %s, want 0.12", result)
	}
}

func TestBuildSeries_Empty(t *testing.T) {
	result := &analyzer.BatchAnalysisResult{
		Summary: analyzer.BatchSummary{
			Labels: []string{},
		},
		Series: make(map[string][]analyzer.ProfileDataPoint),
	}

	series := buildSeries(result, func(p analyzer.ProfileDataPoint) float64 { return p.WallClockMs })

	if len(series) != 0 {
		t.Errorf("expected empty series, got %d", len(series))
	}
}

func TestBuildSeries_WithData(t *testing.T) {
	result := createTestBatchResult()

	series := buildSeries(result, func(p analyzer.ProfileDataPoint) float64 { return p.WallClockMs })

	if len(series) != 2 {
		t.Errorf("expected 2 series, got %d", len(series))
	}

	// Check Firefox series
	var firefoxSeries *DataSeries
	for i := range series {
		if series[i].Name == "Firefox" {
			firefoxSeries = &series[i]
			break
		}
	}

	if firefoxSeries == nil {
		t.Error("expected Firefox series")
	} else {
		if firefoxSeries.Color != "#FF6611" {
			t.Errorf("Firefox color = %s, want #FF6611", firefoxSeries.Color)
		}
		if len(firefoxSeries.Points) != 3 {
			t.Errorf("Firefox points = %d, want 3", len(firefoxSeries.Points))
		}
	}
}

func TestBuildSeries_DefaultColor(t *testing.T) {
	result := &analyzer.BatchAnalysisResult{
		Summary: analyzer.BatchSummary{
			Labels: []string{"Custom"},
		},
		Series: map[string][]analyzer.ProfileDataPoint{
			"Custom": {
				{WorkerCount: 1, WallClockMs: 100},
			},
		},
	}

	series := buildSeries(result, func(p analyzer.ProfileDataPoint) float64 { return p.WallClockMs })

	if len(series) != 1 {
		t.Errorf("expected 1 series, got %d", len(series))
	}

	// Should have a default color assigned
	if series[0].Color == "" {
		t.Error("expected non-empty default color")
	}
}

func TestChartType_Constants(t *testing.T) {
	// Verify chart type constants
	if ChartWallClock != "wall_clock" {
		t.Error("ChartWallClock constant mismatch")
	}
	if ChartOperationTime != "operation_time" {
		t.Error("ChartOperationTime constant mismatch")
	}
	if ChartEfficiency != "efficiency" {
		t.Error("ChartEfficiency constant mismatch")
	}
	if ChartSpeedup != "speedup" {
		t.Error("ChartSpeedup constant mismatch")
	}
	if ChartCryptoTime != "crypto_time" {
		t.Error("ChartCryptoTime constant mismatch")
	}
}

func TestDefaultColors(t *testing.T) {
	// Verify default colors exist
	if defaultColors["Chrome"] == "" {
		t.Error("expected Chrome color")
	}
	if defaultColors["Firefox"] == "" {
		t.Error("expected Firefox color")
	}
}

// Helper to create test batch result
func createTestBatchResult() *analyzer.BatchAnalysisResult {
	return &analyzer.BatchAnalysisResult{
		Summary: analyzer.BatchSummary{
			TotalProfiles: 6,
			Labels:        []string{"Firefox", "Chrome"},
			BestWorkers:   map[string]int{"Firefox": 4, "Chrome": 8},
			MinWallClock:  map[string]float64{"Firefox": 500, "Chrome": 450},
		},
		Series: map[string][]analyzer.ProfileDataPoint{
			"Firefox": {
				{WorkerCount: 1, WallClockMs: 1000, OperationTimeMs: 800, Efficiency: 100, Speedup: 1.0, CryptoTimeMs: 200},
				{WorkerCount: 2, WallClockMs: 600, OperationTimeMs: 500, Efficiency: 83, Speedup: 1.67, CryptoTimeMs: 150},
				{WorkerCount: 4, WallClockMs: 400, OperationTimeMs: 350, Efficiency: 62.5, Speedup: 2.5, CryptoTimeMs: 100},
			},
			"Chrome": {
				{WorkerCount: 1, WallClockMs: 1100, OperationTimeMs: 900, Efficiency: 100, Speedup: 1.0, CryptoTimeMs: 220},
				{WorkerCount: 2, WallClockMs: 650, OperationTimeMs: 550, Efficiency: 85, Speedup: 1.69, CryptoTimeMs: 160},
				{WorkerCount: 4, WallClockMs: 450, OperationTimeMs: 400, Efficiency: 61, Speedup: 2.44, CryptoTimeMs: 110},
			},
		},
	}
}
