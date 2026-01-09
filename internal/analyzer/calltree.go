package analyzer

import (
	"fmt"
	"sort"
	"strings"

	"github.com/CedricHerzog/perfowl/internal/parser"
)

// FunctionStats contains timing information for a single function
type FunctionStats struct {
	Name          string  `json:"name"`
	File          string  `json:"file,omitempty"`
	SelfTimeMs    float64 `json:"self_time_ms"`
	RunningTimeMs float64 `json:"running_time_ms"`
	SelfPercent   float64 `json:"self_percent"`
	TotalPercent  float64 `json:"total_percent"`
	SampleCount   int     `json:"sample_count"`
	CallPath      string  `json:"call_path,omitempty"`
}

// HotPath represents a frequently executed call path
type HotPath struct {
	Path       string  `json:"path"`
	SelfTimeMs float64 `json:"self_time_ms"`
	Percent    float64 `json:"percent"`
	Count      int     `json:"count"`
}

// CallTreeAnalysis contains the call tree analysis results
type CallTreeAnalysis struct {
	TotalTimeMs   float64         `json:"total_time_ms"`
	TotalSamples  int             `json:"total_samples"`
	TopFunctions  []FunctionStats `json:"top_functions"`
	HotPaths      []HotPath       `json:"hot_paths"`
	ThreadName    string          `json:"thread_name,omitempty"`
}

// AnalyzeCallTree builds a call tree and identifies hot paths
func AnalyzeCallTree(profile *parser.Profile, threadName string, limit int) CallTreeAnalysis {
	// Use shared string array as fallback (Firefox profile optimization)
	sharedStrings := profile.Shared.StringArray
	analysis := CallTreeAnalysis{
		TopFunctions: make([]FunctionStats, 0),
		HotPaths:     make([]HotPath, 0),
		ThreadName:   threadName,
	}

	if limit <= 0 {
		limit = 20
	}

	interval := profile.Meta.Interval

	// Aggregate function stats across all threads by function name
	type funcData struct {
		selfTime    float64
		runningTime float64
		sampleCount int
		file        string
	}
	globalFuncStats := make(map[string]*funcData)

	// Aggregate hot paths
	type pathData struct {
		time  float64
		count int
	}
	globalHotPaths := make(map[string]*pathData)

	for _, thread := range profile.Threads {
		// Filter by thread name if specified
		if threadName != "" && thread.Name != threadName {
			continue
		}

		stackTable := &thread.StackTable
		frameTable := &thread.FrameTable
		funcTable := &thread.FuncTable
		samples := &thread.Samples

		// Use thread's string array, fall back to shared if empty
		stringArray := thread.StringArray
		if len(stringArray) == 0 {
			stringArray = sharedStrings
		}

		// Per-thread tracking using indices (fast)
		leafFrameTime := make(map[int]float64)
		leafFrameCount := make(map[int]int)
		stackTime := make(map[int]float64)
		stackCount := make(map[int]int)

		// First pass: collect self time by leaf frame and stack time
		for i := 0; i < samples.Length; i++ {
			stackIdx := -1
			if i < len(samples.Stack) {
				stackIdx = samples.Stack[i]
			}
			if stackIdx < 0 {
				continue
			}

			cpuDelta := interval
			if i < len(samples.ThreadCPUDelta) && samples.ThreadCPUDelta[i] > 0 {
				cpuDelta = float64(samples.ThreadCPUDelta[i]) / 1000.0
			}

			analysis.TotalTimeMs += cpuDelta
			analysis.TotalSamples++

			// Get leaf frame (self time)
			if stackIdx < len(stackTable.Frame) {
				frameIdx := stackTable.Frame[stackIdx]
				if frameIdx >= 0 {
					leafFrameTime[frameIdx] += cpuDelta
					leafFrameCount[frameIdx]++
				}
			}

			// Track stack for hot paths
			stackTime[stackIdx] += cpuDelta
			stackCount[stackIdx]++
		}

		// Compute running time by walking unique stacks
		frameRunningTime := make(map[int]float64)
		for stackIdx, stackCpuTime := range stackTime {
			currentStack := stackIdx
			depth := 0
			maxDepth := 50

			for currentStack >= 0 && currentStack < stackTable.Length && depth < maxDepth {
				if currentStack < len(stackTable.Frame) {
					frameIdx := stackTable.Frame[currentStack]
					if frameIdx >= 0 {
						frameRunningTime[frameIdx] += stackCpuTime
					}
				}

				if currentStack < len(stackTable.Prefix) {
					currentStack = stackTable.Prefix[currentStack]
				} else {
					break
				}
				depth++
			}
		}

		// Helper to get function name from frame index
		getFuncName := func(frameIdx int) (string, string) {
			if frameIdx < 0 || frameIdx >= frameTable.Length || frameIdx >= len(frameTable.Func) {
				return "(unknown)", ""
			}
			funcIdx := frameTable.Func[frameIdx]
			if funcIdx < 0 || funcIdx >= funcTable.Length {
				return "(unknown)", ""
			}

			funcName := "(unknown)"
			fileName := ""

			if funcIdx < len(funcTable.Name) {
				nameIdx := funcTable.Name[funcIdx]
				if nameIdx >= 0 && nameIdx < len(stringArray) {
					funcName = stringArray[nameIdx]
				}
			}
			if funcIdx < len(funcTable.FileName) {
				fileIdx := funcTable.FileName[funcIdx]
				if fileIdx >= 0 && fileIdx < len(stringArray) {
					fileName = stringArray[fileIdx]
				}
			}

			return funcName, fileName
		}

		// Convert leaf frame stats to function stats (aggregate by name)
		for frameIdx, selfTime := range leafFrameTime {
			funcName, fileName := getFuncName(frameIdx)

			if globalFuncStats[funcName] == nil {
				globalFuncStats[funcName] = &funcData{file: fileName}
			}
			globalFuncStats[funcName].selfTime += selfTime
			globalFuncStats[funcName].sampleCount += leafFrameCount[frameIdx]
		}

		// Add running time
		for frameIdx, runTime := range frameRunningTime {
			funcName, fileName := getFuncName(frameIdx)

			if globalFuncStats[funcName] == nil {
				globalFuncStats[funcName] = &funcData{file: fileName}
			}
			globalFuncStats[funcName].runningTime += runTime
		}

		// Build hot paths from top stacks
		type stackEntry struct {
			idx   int
			time  float64
			count int
		}
		var topStacks []stackEntry
		for idx, time := range stackTime {
			topStacks = append(topStacks, stackEntry{idx, time, stackCount[idx]})
		}
		sort.Slice(topStacks, func(i, j int) bool {
			return topStacks[i].time > topStacks[j].time
		})

		// Convert top stacks to paths and aggregate
		maxPaths := limit * 2 // Get more to aggregate
		if len(topStacks) < maxPaths {
			maxPaths = len(topStacks)
		}

		for i := 0; i < maxPaths; i++ {
			entry := topStacks[i]
			path := buildStackPathWithThread(&thread, stringArray, entry.idx, 5)

			if globalHotPaths[path] == nil {
				globalHotPaths[path] = &pathData{}
			}
			globalHotPaths[path].time += entry.time
			globalHotPaths[path].count += entry.count
		}
	}

	// Convert global function stats to output
	for name, data := range globalFuncStats {
		selfPercent := 0.0
		totalPercent := 0.0
		if analysis.TotalTimeMs > 0 {
			selfPercent = (data.selfTime / analysis.TotalTimeMs) * 100
			totalPercent = (data.runningTime / analysis.TotalTimeMs) * 100
		}

		analysis.TopFunctions = append(analysis.TopFunctions, FunctionStats{
			Name:          name,
			File:          data.file,
			SelfTimeMs:    data.selfTime,
			RunningTimeMs: data.runningTime,
			SelfPercent:   selfPercent,
			TotalPercent:  totalPercent,
			SampleCount:   data.sampleCount,
		})
	}

	// Sort by self time descending
	sort.Slice(analysis.TopFunctions, func(i, j int) bool {
		return analysis.TopFunctions[i].SelfTimeMs > analysis.TopFunctions[j].SelfTimeMs
	})

	if len(analysis.TopFunctions) > limit {
		analysis.TopFunctions = analysis.TopFunctions[:limit]
	}

	// Convert hot paths to output
	for path, data := range globalHotPaths {
		percent := 0.0
		if analysis.TotalTimeMs > 0 {
			percent = (data.time / analysis.TotalTimeMs) * 100
		}
		analysis.HotPaths = append(analysis.HotPaths, HotPath{
			Path:       path,
			SelfTimeMs: data.time,
			Percent:    percent,
			Count:      data.count,
		})
	}

	sort.Slice(analysis.HotPaths, func(i, j int) bool {
		return analysis.HotPaths[i].SelfTimeMs > analysis.HotPaths[j].SelfTimeMs
	})

	if len(analysis.HotPaths) > limit {
		analysis.HotPaths = analysis.HotPaths[:limit]
	}

	return analysis
}

// buildStackPathWithThread builds a human-readable path from a stack index
func buildStackPathWithThread(thread *parser.Thread, stringArray []string, stackIdx int, maxDepth int) string {
	var names []string
	currentStack := stackIdx
	depth := 0

	for currentStack >= 0 && currentStack < thread.StackTable.Length && depth < maxDepth {
		frameIdx := -1
		if currentStack < len(thread.StackTable.Frame) {
			frameIdx = thread.StackTable.Frame[currentStack]
		}

		if frameIdx >= 0 && frameIdx < thread.FrameTable.Length {
			funcIdx := -1
			if frameIdx < len(thread.FrameTable.Func) {
				funcIdx = thread.FrameTable.Func[frameIdx]
			}

			if funcIdx >= 0 && funcIdx < thread.FuncTable.Length {
				nameIdx := -1
				if funcIdx < len(thread.FuncTable.Name) {
					nameIdx = thread.FuncTable.Name[funcIdx]
				}

				funcName := "(unknown)"
				if nameIdx >= 0 && nameIdx < len(stringArray) {
					funcName = stringArray[nameIdx]
				}

				// Truncate long names
				if len(funcName) > 50 {
					funcName = funcName[:47] + "..."
				}
				names = append(names, funcName)
			}
		}

		if currentStack < len(thread.StackTable.Prefix) {
			currentStack = thread.StackTable.Prefix[currentStack]
		} else {
			break
		}
		depth++
	}

	// Reverse to show root -> leaf
	for i, j := 0, len(names)-1; i < j; i, j = i+1, j-1 {
		names[i], names[j] = names[j], names[i]
	}

	return strings.Join(names, " → ")
}

// FormatCallTree returns a human-readable call tree summary
func FormatCallTree(analysis CallTreeAnalysis) string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("Call Tree Analysis (Total: %.2fms, %d samples)\n",
		analysis.TotalTimeMs, analysis.TotalSamples))
	sb.WriteString(strings.Repeat("=", 60) + "\n\n")

	sb.WriteString("Top Functions by Self Time:\n")
	sb.WriteString(strings.Repeat("-", 60) + "\n")
	sb.WriteString(fmt.Sprintf("%-40s %10s %8s\n", "Function", "Self Time", "Self %"))
	sb.WriteString(strings.Repeat("-", 60) + "\n")

	for _, f := range analysis.TopFunctions {
		name := f.Name
		if len(name) > 40 {
			name = name[:37] + "..."
		}
		sb.WriteString(fmt.Sprintf("%-40s %8.2fms %7.1f%%\n", name, f.SelfTimeMs, f.SelfPercent))
	}

	sb.WriteString("\nHot Paths:\n")
	sb.WriteString(strings.Repeat("-", 60) + "\n")

	for i, hp := range analysis.HotPaths {
		if i >= 10 {
			break
		}
		sb.WriteString(fmt.Sprintf("%.1f%% (%.2fms): %s\n", hp.Percent, hp.SelfTimeMs, hp.Path))
	}

	return sb.String()
}
