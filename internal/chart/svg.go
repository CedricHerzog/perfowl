package chart

import (
	"fmt"
	"math"
	"strings"

	"github.com/CedricHerzog/perfowl/internal/analyzer"
)

// ChartType specifies what kind of chart to generate
type ChartType string

const (
	ChartWallClock     ChartType = "wall_clock"
	ChartOperationTime ChartType = "operation_time"
	ChartEfficiency    ChartType = "efficiency"
	ChartSpeedup       ChartType = "speedup"
	ChartCryptoTime    ChartType = "crypto_time"
)

// DataPoint is a single x,y coordinate
type DataPoint struct {
	X float64
	Y float64
}

// DataSeries represents a single line in the chart
type DataSeries struct {
	Name   string
	Color  string
	Points []DataPoint
}

// ChartConfig defines chart appearance
type ChartConfig struct {
	Width      int
	Height     int
	Title      string
	XAxisLabel string
	YAxisLabel string
	ShowLegend bool
	ShowGrid   bool
}

// Default colors for common series
var defaultColors = map[string]string{
	"Chrome":  "#4285F4", // Google blue
	"Firefox": "#FF6611", // Firefox orange
}

// GenerateScalingChart creates an SVG chart from batch analysis results
func GenerateScalingChart(result *analyzer.BatchAnalysisResult, chartType ChartType) string {
	config := ChartConfig{
		Width:      800,
		Height:     450,
		ShowLegend: true,
		ShowGrid:   true,
	}

	var series []DataSeries

	switch chartType {
	case ChartWallClock:
		config.Title = "Wall Clock Time vs Worker Count"
		config.XAxisLabel = "Worker Count"
		config.YAxisLabel = "Time (ms)"
		series = buildSeries(result, func(p analyzer.ProfileDataPoint) float64 { return p.WallClockMs })

	case ChartOperationTime:
		config.Title = "Operation Time vs Worker Count"
		config.XAxisLabel = "Worker Count"
		config.YAxisLabel = "Time (ms)"
		series = buildSeries(result, func(p analyzer.ProfileDataPoint) float64 { return p.OperationTimeMs })

	case ChartEfficiency:
		config.Title = "Parallel Efficiency vs Worker Count"
		config.XAxisLabel = "Worker Count"
		config.YAxisLabel = "Efficiency (%)"
		series = buildSeries(result, func(p analyzer.ProfileDataPoint) float64 { return p.Efficiency })

	case ChartSpeedup:
		config.Title = "Speedup vs Worker Count"
		config.XAxisLabel = "Worker Count"
		config.YAxisLabel = "Speedup (x)"
		series = buildSeries(result, func(p analyzer.ProfileDataPoint) float64 { return p.Speedup })

	case ChartCryptoTime:
		config.Title = "Crypto Time vs Worker Count"
		config.XAxisLabel = "Worker Count"
		config.YAxisLabel = "Time (ms)"
		series = buildSeries(result, func(p analyzer.ProfileDataPoint) float64 { return p.CryptoTimeMs })

	default:
		config.Title = "Wall Clock Time vs Worker Count"
		config.XAxisLabel = "Worker Count"
		config.YAxisLabel = "Time (ms)"
		series = buildSeries(result, func(p analyzer.ProfileDataPoint) float64 { return p.WallClockMs })
	}

	return GenerateSVG(config, series)
}

// buildSeries creates data series from batch results using a metric extractor
func buildSeries(result *analyzer.BatchAnalysisResult, getMetric func(analyzer.ProfileDataPoint) float64) []DataSeries {
	var series []DataSeries

	for _, label := range result.Summary.Labels {
		points, ok := result.Series[label]
		if !ok || len(points) == 0 {
			continue
		}

		ds := DataSeries{
			Name:   label,
			Color:  defaultColors[label],
			Points: make([]DataPoint, 0, len(points)),
		}

		if ds.Color == "" {
			// Assign a default color based on index
			colors := []string{"#4285F4", "#FF6611", "#34A853", "#EA4335", "#FBBC05"}
			idx := len(series) % len(colors)
			ds.Color = colors[idx]
		}

		for _, p := range points {
			ds.Points = append(ds.Points, DataPoint{
				X: float64(p.WorkerCount),
				Y: getMetric(p),
			})
		}

		series = append(series, ds)
	}

	return series
}

// GenerateSVG creates an SVG chart from data series
func GenerateSVG(config ChartConfig, series []DataSeries) string {
	var sb strings.Builder

	// Defaults
	if config.Width == 0 {
		config.Width = 800
	}
	if config.Height == 0 {
		config.Height = 450
	}

	margin := struct{ top, right, bottom, left int }{60, 100, 70, 80}
	chartWidth := config.Width - margin.left - margin.right
	chartHeight := config.Height - margin.top - margin.bottom

	// Calculate ranges
	xMin, xMax, yMin, yMax := calculateRanges(series)

	// Add padding to Y range
	yRange := yMax - yMin
	if yRange == 0 {
		yRange = 1
	}
	yMin = math.Max(0, yMin-yRange*0.05)
	yMax = yMax + yRange*0.05

	// Scale functions
	scaleX := func(x float64) float64 {
		if xMax == xMin {
			return float64(margin.left) + float64(chartWidth)/2
		}
		return float64(margin.left) + (x-xMin)/(xMax-xMin)*float64(chartWidth)
	}
	scaleY := func(y float64) float64 {
		if yMax == yMin {
			return float64(config.Height-margin.bottom) - float64(chartHeight)/2
		}
		return float64(config.Height-margin.bottom) - (y-yMin)/(yMax-yMin)*float64(chartHeight)
	}

	// SVG header
	sb.WriteString(fmt.Sprintf(`<?xml version="1.0" encoding="UTF-8"?>
<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 %d %d" width="%d" height="%d">
<style>
  .chart-bg { fill: #fafafa; }
  .axis { stroke: #333; stroke-width: 1.5; fill: none; }
  .grid { stroke: #e0e0e0; stroke-width: 0.5; stroke-dasharray: 4,4; }
  .title { font: bold 18px system-ui, -apple-system, sans-serif; fill: #222; }
  .axis-label { font: 13px system-ui, -apple-system, sans-serif; fill: #555; }
  .axis-title { font: 13px system-ui, -apple-system, sans-serif; fill: #333; }
  .legend-text { font: 12px system-ui, -apple-system, sans-serif; fill: #333; }
  .data-line { fill: none; stroke-width: 2.5; stroke-linecap: round; stroke-linejoin: round; }
  .data-point { stroke: white; stroke-width: 2; }
</style>
`, config.Width, config.Height, config.Width, config.Height))

	// Background
	sb.WriteString(fmt.Sprintf(`<rect class="chart-bg" x="0" y="0" width="%d" height="%d"/>
`, config.Width, config.Height))

	// Title
	if config.Title != "" {
		sb.WriteString(fmt.Sprintf(`<text class="title" x="%d" y="35" text-anchor="middle">%s</text>
`, config.Width/2, config.Title))
	}

	// Grid lines (horizontal)
	if config.ShowGrid {
		yTicks := calculateTicks(yMin, yMax, 5)
		for _, tick := range yTicks {
			y := scaleY(tick)
			sb.WriteString(fmt.Sprintf(`<line class="grid" x1="%d" y1="%.1f" x2="%d" y2="%.1f"/>
`, margin.left, y, config.Width-margin.right, y))
		}
	}

	// X axis
	sb.WriteString(fmt.Sprintf(`<line class="axis" x1="%d" y1="%d" x2="%d" y2="%d"/>
`, margin.left, config.Height-margin.bottom, config.Width-margin.right, config.Height-margin.bottom))

	// Y axis
	sb.WriteString(fmt.Sprintf(`<line class="axis" x1="%d" y1="%d" x2="%d" y2="%d"/>
`, margin.left, margin.top, margin.left, config.Height-margin.bottom))

	// X axis labels (worker counts)
	xTicks := calculateXTicks(xMin, xMax)
	for _, x := range xTicks {
		xPos := scaleX(x)
		label := fmt.Sprintf("%.0f", x)
		sb.WriteString(fmt.Sprintf(`<text class="axis-label" x="%.1f" y="%d" text-anchor="middle">%s</text>
`, xPos, config.Height-margin.bottom+20, label))
	}

	// X axis title
	sb.WriteString(fmt.Sprintf(`<text class="axis-title" x="%d" y="%d" text-anchor="middle">%s</text>
`, config.Width/2, config.Height-15, config.XAxisLabel))

	// Y axis labels
	yTicks := calculateTicks(yMin, yMax, 5)
	for _, tick := range yTicks {
		y := scaleY(tick)
		label := formatNumber(tick)
		sb.WriteString(fmt.Sprintf(`<text class="axis-label" x="%d" y="%.1f" text-anchor="end" dominant-baseline="middle">%s</text>
`, margin.left-10, y, label))
	}

	// Y axis title (rotated)
	sb.WriteString(fmt.Sprintf(`<text class="axis-title" x="20" y="%d" text-anchor="middle" transform="rotate(-90, 20, %d)">%s</text>
`, (config.Height-margin.top-margin.bottom)/2+margin.top, (config.Height-margin.top-margin.bottom)/2+margin.top, config.YAxisLabel))

	// Data lines and points
	for _, s := range series {
		if len(s.Points) == 0 {
			continue
		}

		// Draw line
		var pathData strings.Builder
		for i, p := range s.Points {
			x := scaleX(p.X)
			y := scaleY(p.Y)
			if i == 0 {
				pathData.WriteString(fmt.Sprintf("M%.1f,%.1f", x, y))
			} else {
				pathData.WriteString(fmt.Sprintf(" L%.1f,%.1f", x, y))
			}
		}
		sb.WriteString(fmt.Sprintf(`<path class="data-line" stroke="%s" d="%s"/>
`, s.Color, pathData.String()))

		// Draw points
		for _, p := range s.Points {
			x := scaleX(p.X)
			y := scaleY(p.Y)
			sb.WriteString(fmt.Sprintf(`<circle class="data-point" cx="%.1f" cy="%.1f" r="5" fill="%s"/>
`, x, y, s.Color))
		}
	}

	// Legend
	if config.ShowLegend && len(series) > 0 {
		legendX := config.Width - margin.right + 15
		legendY := margin.top + 10

		for i, s := range series {
			y := legendY + i*25
			sb.WriteString(fmt.Sprintf(`<line x1="%d" y1="%d" x2="%d" y2="%d" stroke="%s" stroke-width="3"/>
`, legendX, y, legendX+20, y, s.Color))
			sb.WriteString(fmt.Sprintf(`<circle cx="%d" cy="%d" r="4" fill="%s" stroke="white" stroke-width="1"/>
`, legendX+10, y, s.Color))
			sb.WriteString(fmt.Sprintf(`<text class="legend-text" x="%d" y="%d" dominant-baseline="middle">%s</text>
`, legendX+28, y, s.Name))
		}
	}

	sb.WriteString(`</svg>`)
	return sb.String()
}

// calculateRanges finds min/max for x and y across all series
func calculateRanges(series []DataSeries) (xMin, xMax, yMin, yMax float64) {
	xMin = math.MaxFloat64
	xMax = -math.MaxFloat64
	yMin = math.MaxFloat64
	yMax = -math.MaxFloat64

	for _, s := range series {
		for _, p := range s.Points {
			if p.X < xMin {
				xMin = p.X
			}
			if p.X > xMax {
				xMax = p.X
			}
			if p.Y < yMin {
				yMin = p.Y
			}
			if p.Y > yMax {
				yMax = p.Y
			}
		}
	}

	// Handle empty data
	if xMin == math.MaxFloat64 {
		xMin, xMax = 0, 12
		yMin, yMax = 0, 100
	}

	return xMin, xMax, yMin, yMax
}

// calculateTicks generates nice tick values for the Y axis
func calculateTicks(min, max float64, count int) []float64 {
	if max <= min {
		return []float64{min}
	}

	range_ := max - min
	step := range_ / float64(count)

	// Round step to nice value
	magnitude := math.Pow(10, math.Floor(math.Log10(step)))
	normalized := step / magnitude

	var niceStep float64
	switch {
	case normalized <= 1.5:
		niceStep = magnitude
	case normalized <= 3:
		niceStep = 2 * magnitude
	case normalized <= 7:
		niceStep = 5 * magnitude
	default:
		niceStep = 10 * magnitude
	}

	ticks := make([]float64, 0)
	start := math.Ceil(min/niceStep) * niceStep
	for t := start; t <= max; t += niceStep {
		ticks = append(ticks, t)
	}

	return ticks
}

// calculateXTicks generates tick values for X axis (worker counts)
func calculateXTicks(min, max float64) []float64 {
	ticks := make([]float64, 0)

	// For worker counts, use integer values
	start := math.Ceil(min)
	end := math.Floor(max)

	step := 1.0
	range_ := end - start
	if range_ > 12 {
		step = 2.0
	}
	if range_ > 20 {
		step = 5.0
	}

	for x := start; x <= end; x += step {
		ticks = append(ticks, x)
	}

	return ticks
}

// formatNumber formats a number for display
func formatNumber(n float64) string {
	if n == 0 {
		return "0"
	}
	if math.Abs(n) >= 1000 {
		return fmt.Sprintf("%.0f", n)
	}
	if math.Abs(n) >= 100 {
		return fmt.Sprintf("%.0f", n)
	}
	if math.Abs(n) >= 10 {
		return fmt.Sprintf("%.1f", n)
	}
	if math.Abs(n) >= 1 {
		return fmt.Sprintf("%.1f", n)
	}
	return fmt.Sprintf("%.2f", n)
}
