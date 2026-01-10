package analyzer

import (
	"fmt"
	"sort"
	"strings"

	"github.com/CedricHerzog/perfowl/internal/parser"
)

// JSCryptoResource represents a JavaScript crypto resource/worker
type JSCryptoResource struct {
	Name       string  `json:"name"`
	URL        string  `json:"url,omitempty"`
	TotalTime  float64 `json:"total_time_ms"`
	SampleCount int    `json:"sample_count"`
	ThreadName string  `json:"thread_name"`
}

// JSCryptoFunction represents a function within a crypto resource
type JSCryptoFunction struct {
	Name       string  `json:"name"`
	Resource   string  `json:"resource"`
	TotalTime  float64 `json:"total_time_ms"`
	SampleCount int    `json:"sample_count"`
	Percent    float64 `json:"percent"`
}

// JSCryptoAnalysis contains JavaScript-level crypto analysis
type JSCryptoAnalysis struct {
	TotalTimeMs      float64             `json:"total_time_ms"`
	TotalSamples     int                 `json:"total_samples"`
	Resources        []JSCryptoResource  `json:"resources"`
	TopFunctions     []JSCryptoFunction  `json:"top_functions"`
	ByThread         map[string]float64  `json:"by_thread"`
	WorkerCount      int                 `json:"worker_count"`
	AvgTimePerWorker float64             `json:"avg_time_per_worker_ms"`
	Recommendations  []string            `json:"recommendations,omitempty"`
}

// cryptoResourcePatterns identifies JS files that do crypto work
var cryptoResourcePatterns = []string{
	"decrypt",
	"encrypt",
	"crypto",
	"seipd",
	"openpgp",
	"pgp",
	"aes",
	"rsa",
	"cipher",
	"webcrypto",
}

// isCryptoJSResource checks if a resource name indicates JS crypto work
func isCryptoJSResource(name string) bool {
	lower := strings.ToLower(name)
	// Must be a JS/MJS file or worker
	if !strings.Contains(lower, ".js") && !strings.Contains(lower, ".mjs") && !strings.Contains(lower, "worker") {
		return false
	}
	for _, pattern := range cryptoResourcePatterns {
		if strings.Contains(lower, pattern) {
			return true
		}
	}
	return false
}

// extractResourceName extracts a clean resource name from a URL or path
func extractResourceName(url string) string {
	// Handle moz-extension URLs
	if strings.Contains(url, "moz-extension://") {
		parts := strings.Split(url, "/")
		if len(parts) > 0 {
			return parts[len(parts)-1]
		}
	}
	// Handle chrome-extension URLs
	if strings.Contains(url, "chrome-extension://") {
		parts := strings.Split(url, "/")
		if len(parts) > 0 {
			return parts[len(parts)-1]
		}
	}
	// Handle regular URLs (remove query string first)
	cleanURL := url
	if queryIdx := strings.Index(url, "?"); queryIdx >= 0 {
		cleanURL = url[:queryIdx]
	}
	if idx := strings.LastIndex(cleanURL, "/"); idx >= 0 {
		return cleanURL[idx+1:]
	}
	return cleanURL
}

// extractResourceFromFuncName extracts a JS file path from a Firefox function name
// Firefox embeds paths in function names like:
// "(root scope) moz-extension://abc/workers/seipdDecryptionWorker.min.js"
// "WorkerThreadPrimaryRunnable::Run moz-extension://abc/workers/file.js"
// "../node_modules/.pnpm/openpgp@6.1.1/node_modules/openpgp/dist/openpgp.min.mjs/func"
// Chrome uses chrome-extension:// URLs in FuncTable.FileName
func extractResourceFromFuncName(funcName string) string {
	// Look for moz-extension:// URL pattern
	if idx := strings.Index(funcName, "moz-extension://"); idx >= 0 {
		url := funcName[idx:]
		// Remove trailing line:col info like ":67:93872"
		if colonIdx := strings.LastIndex(url, ":"); colonIdx > 0 {
			// Check if it's line:col format (two colons with numbers)
			beforeColon := url[:colonIdx]
			if secondColonIdx := strings.LastIndex(beforeColon, ":"); secondColonIdx > 0 {
				// Verify both parts after colons are numeric
				part1 := beforeColon[secondColonIdx+1:]
				part2 := url[colonIdx+1:]
				if isNumeric(part1) && isNumeric(part2) {
					url = beforeColon[:secondColonIdx]
				}
			}
		}
		return url
	}
	// Look for chrome-extension:// URL pattern
	if idx := strings.Index(funcName, "chrome-extension://"); idx >= 0 {
		url := funcName[idx:]
		// Remove trailing line:col info
		if colonIdx := strings.LastIndex(url, ":"); colonIdx > 0 {
			beforeColon := url[:colonIdx]
			if secondColonIdx := strings.LastIndex(beforeColon, ":"); secondColonIdx > 0 {
				part1 := beforeColon[secondColonIdx+1:]
				part2 := url[colonIdx+1:]
				if isNumeric(part1) && isNumeric(part2) {
					url = beforeColon[:secondColonIdx]
				}
			}
		}
		return url
	}
	// Look for resource:// URL pattern
	if idx := strings.Index(funcName, "resource://"); idx >= 0 {
		return funcName[idx:]
	}
	// Look for webpack-style paths like "../node_modules/.pnpm/openpgp@6.1.1/.../openpgp.min.mjs/func"
	// Extract the .js or .mjs file name
	if strings.Contains(funcName, "node_modules/") {
		// Find the last .js or .mjs file in the path
		parts := strings.Split(funcName, "/")
		for i := len(parts) - 1; i >= 0; i-- {
			if strings.HasSuffix(parts[i], ".js") || strings.HasSuffix(parts[i], ".mjs") {
				return parts[i]
			}
		}
	}
	// Look for ./src/ style paths
	if strings.Contains(funcName, "./src/") {
		// Extract just the file name
		parts := strings.Split(funcName, "/")
		for i := len(parts) - 1; i >= 0; i-- {
			if strings.HasSuffix(parts[i], ".js") || strings.HasSuffix(parts[i], ".ts") {
				return parts[i]
			}
		}
	}
	// Return as-is if no URL found
	return funcName
}

// isNumeric checks if a string contains only digits
func isNumeric(s string) bool {
	for _, c := range s {
		if c < '0' || c > '9' {
			return false
		}
	}
	return len(s) > 0
}

// AnalyzeJSCrypto analyzes JavaScript-level crypto operations
func AnalyzeJSCrypto(profile *parser.Profile) JSCryptoAnalysis {
	analysis := JSCryptoAnalysis{
		Resources:    make([]JSCryptoResource, 0),
		TopFunctions: make([]JSCryptoFunction, 0),
		ByThread:     make(map[string]float64),
		Recommendations: make([]string, 0),
	}

	interval := profile.Meta.Interval

	// Track resources and functions
	type resourceAgg struct {
		name       string
		url        string
		totalTime  float64
		samples    int
		threadName string
	}
	resourceStats := make(map[string]*resourceAgg) // key: resource URL + thread

	type funcAgg struct {
		name      string
		resource  string
		totalTime float64
		samples   int
	}
	funcStats := make(map[string]*funcAgg) // key: func name + resource

	workerThreads := make(map[string]bool)

	for _, thread := range profile.Threads {
		// Use shared string array if thread's is empty
		stringArray := thread.StringArray
		if len(stringArray) == 0 {
			stringArray = profile.Shared.StringArray
		}

		// Build lookup tables
		// funcIdx -> function name
		funcIdxToName := make(map[int]string)
		for funcIdx, nameIdx := range thread.FuncTable.Name {
			if nameIdx >= 0 && nameIdx < len(stringArray) {
				funcIdxToName[funcIdx] = stringArray[nameIdx]
			}
		}

		// funcIdx -> file name (Chrome profiles store file URLs separately in FuncTable.FileName)
		// Firefox profiles don't use this; they embed paths in function names instead
		funcIdxToFileName := make(map[int]string)
		for funcIdx, fileNameIdx := range thread.FuncTable.FileName {
			if fileNameIdx >= 0 && fileNameIdx < len(stringArray) {
				funcIdxToFileName[funcIdx] = stringArray[fileNameIdx]
			}
		}

		// resourceIdx -> resource name
		resourceIdxToName := make(map[int]string)
		for resIdx, nameIdx := range thread.ResourceTable.Name {
			if nameIdx >= 0 && nameIdx < len(stringArray) {
				resourceIdxToName[resIdx] = stringArray[nameIdx]
			}
		}

		// funcIdx -> resourceIdx
		funcIdxToResource := make(map[int]int)
		for funcIdx, resIdx := range thread.FuncTable.Resource {
			funcIdxToResource[funcIdx] = resIdx
		}

		// First pass: find which resources are crypto JS resources
		// Firefox profiles have functions with embedded paths like:
		// "(root scope) moz-extension://...seipdDecryptionWorker.min.js"
		// Chrome profiles have the file URL in FuncTable.FileName
		// IMPORTANT: Only consider JS functions (isJS=true) to avoid counting native code
		cryptoResourceIdx := make(map[int]string) // resourceIdx -> extracted file name
		cryptoResourceFile := ""                   // Track the crypto file for this thread

		for funcIdx, funcName := range funcIdxToName {
			// Skip non-JS functions - they might reference crypto files but aren't running crypto code
			if funcIdx >= len(thread.FuncTable.IsJS) || !thread.FuncTable.IsJS[funcIdx] {
				continue
			}

			// Check both function name (Firefox) and file name (Chrome) for crypto patterns
			fileName := funcIdxToFileName[funcIdx]
			checkName := funcName
			if fileName != "" && fileName != "(unknown)" {
				// Chrome profile: prefer checking the file name
				checkName = fileName
			}

			if isCryptoJSResource(checkName) {
				extractedFile := extractResourceFromFuncName(checkName)
				if fileName != "" && fileName != "(unknown)" {
					// Chrome: use the file name directly
					extractedFile = extractResourceName(fileName)
				}
				// Also mark the resource this function belongs to as crypto
				if resIdx, ok := funcIdxToResource[funcIdx]; ok && resIdx >= 0 {
					cryptoResourceIdx[resIdx] = extractedFile
					cryptoResourceFile = extractedFile
				} else {
					// No resource index, use function name for tracking
					cryptoResourceFile = extractedFile
				}
			}
		}

		// If no crypto resources found in this thread, skip it
		if len(cryptoResourceIdx) == 0 && cryptoResourceFile == "" {
			continue
		}

		// Second pass: identify JS functions that are actually doing crypto work
		// Only count functions that:
		// 1. Have a crypto file path in their function name or file name
		// 2. OR have a crypto-related function name (decrypt, encrypt, etc.)
		cryptoFuncIdx := make(map[int]string) // funcIdx -> resource file name
		for funcIdx, funcName := range funcIdxToName {
			// Skip non-JS functions (native code)
			if funcIdx >= len(thread.FuncTable.IsJS) || !thread.FuncTable.IsJS[funcIdx] {
				continue
			}

			// Get file name for Chrome profiles
			fileName := funcIdxToFileName[funcIdx]
			checkName := funcName
			if fileName != "" && fileName != "(unknown)" {
				checkName = fileName
			}

			// Check if function/file name indicates a crypto JS resource
			if isCryptoJSResource(checkName) {
				extractedFile := extractResourceFromFuncName(checkName)
				if fileName != "" && fileName != "(unknown)" {
					extractedFile = extractResourceName(fileName)
				}
				cryptoFuncIdx[funcIdx] = extractedFile
				continue
			}

			// Check if this function has a crypto-related name AND belongs to a crypto resource
			// This ensures we only count functions like "decrypt" when they're in a crypto context
			if resIdx, ok := funcIdxToResource[funcIdx]; ok && resIdx >= 0 {
				if extractedFile, isCrypto := cryptoResourceIdx[resIdx]; isCrypto {
					// Only include if the function name itself is crypto-related
					if isCryptoFunction(funcName) {
						cryptoFuncIdx[funcIdx] = extractedFile
					}
				}
			}

			// For Chrome profiles without resource index, check if file is in a known crypto file
			if fileName != "" && fileName != "(unknown)" && isCryptoJSResource(fileName) {
				if isCryptoFunction(funcName) || cryptoResourceFile != "" {
					cryptoFuncIdx[funcIdx] = extractResourceName(fileName)
				}
			}
		}

		// Use the crypto file name for this thread (if found)
		if cryptoResourceFile == "" && len(cryptoResourceIdx) > 0 {
			for _, f := range cryptoResourceIdx {
				cryptoResourceFile = f
				break
			}
		}

		// Map frame -> func
		frameToFunc := make(map[int]int)
		for frameIdx, funcIdx := range thread.FrameTable.Func {
			frameToFunc[frameIdx] = funcIdx
		}

		// Map stack -> frame (leaf)
		stackToFrame := make(map[int]int)
		for stackIdx, frameIdx := range thread.StackTable.Frame {
			stackToFrame[stackIdx] = frameIdx
		}

		// Process samples
		for i := 0; i < thread.Samples.Length; i++ {
			cpuDelta := interval
			if i < len(thread.Samples.ThreadCPUDelta) {
				delta := thread.Samples.ThreadCPUDelta[i]
				if delta > 0 {
					cpuDelta = float64(delta) / 1000.0
				}
			}

			stackIdx := -1
			if i < len(thread.Samples.Stack) {
				stackIdx = thread.Samples.Stack[i]
			}

			if stackIdx >= 0 {
				frameIdx, ok := stackToFrame[stackIdx]
				if !ok {
					continue
				}
				funcIdx, ok := frameToFunc[frameIdx]
				if !ok {
					continue
				}

				// Check if this is a crypto JS function
				resourceURL, isCrypto := cryptoFuncIdx[funcIdx]
				if !isCrypto {
					continue
				}

				funcName := funcIdxToName[funcIdx]
				resourceName := extractResourceName(resourceURL)

				// Track resource stats
				resKey := resourceURL + "|" + thread.Name
				if resourceStats[resKey] == nil {
					resourceStats[resKey] = &resourceAgg{
						name:       resourceName,
						url:        resourceURL,
						threadName: thread.Name,
					}
				}
				resourceStats[resKey].totalTime += cpuDelta
				resourceStats[resKey].samples++

				// Track function stats
				funcKey := funcName + "|" + resourceName
				if funcStats[funcKey] == nil {
					funcStats[funcKey] = &funcAgg{
						name:     funcName,
						resource: resourceName,
					}
				}
				funcStats[funcKey].totalTime += cpuDelta
				funcStats[funcKey].samples++

				// Track thread stats - use name + TID for uniqueness
				threadKey := thread.Name
				if strings.Contains(thread.Name, "Worker") {
					threadKey = fmt.Sprintf("%s (tid:%s)", thread.Name, thread.TID.String())
				}
				analysis.ByThread[threadKey] += cpuDelta
				analysis.TotalTimeMs += cpuDelta
				analysis.TotalSamples++

				// Track worker threads using TID for uniqueness
				if strings.Contains(thread.Name, "Worker") {
					workerThreads[thread.TID.String()] = true
				}
			}
		}
	}

	analysis.WorkerCount = len(workerThreads)
	if analysis.WorkerCount > 0 {
		analysis.AvgTimePerWorker = analysis.TotalTimeMs / float64(analysis.WorkerCount)
	}

	// Convert resource stats to list
	for _, res := range resourceStats {
		analysis.Resources = append(analysis.Resources, JSCryptoResource{
			Name:        res.name,
			URL:         res.url,
			TotalTime:   res.totalTime,
			SampleCount: res.samples,
			ThreadName:  res.threadName,
		})
	}

	// Sort resources by time
	sort.Slice(analysis.Resources, func(i, j int) bool {
		return analysis.Resources[i].TotalTime > analysis.Resources[j].TotalTime
	})

	// Convert function stats to list
	for _, fn := range funcStats {
		pct := 0.0
		if analysis.TotalTimeMs > 0 {
			pct = (fn.totalTime / analysis.TotalTimeMs) * 100
		}
		analysis.TopFunctions = append(analysis.TopFunctions, JSCryptoFunction{
			Name:        fn.name,
			Resource:    fn.resource,
			TotalTime:   fn.totalTime,
			SampleCount: fn.samples,
			Percent:     pct,
		})
	}

	// Sort functions by time
	sort.Slice(analysis.TopFunctions, func(i, j int) bool {
		return analysis.TopFunctions[i].TotalTime > analysis.TopFunctions[j].TotalTime
	})

	// Limit to top 30 functions
	if len(analysis.TopFunctions) > 30 {
		analysis.TopFunctions = analysis.TopFunctions[:30]
	}

	// Generate recommendations
	if analysis.TotalTimeMs > 1000 {
		analysis.Recommendations = append(analysis.Recommendations,
			fmt.Sprintf("Significant JS crypto overhead: %.1fms - consider WebCrypto API for heavy operations", analysis.TotalTimeMs))
	}

	if analysis.WorkerCount == 1 && analysis.TotalTimeMs > 500 {
		analysis.Recommendations = append(analysis.Recommendations,
			"Only 1 crypto worker detected - consider adding more workers for parallelization")
	}

	if analysis.WorkerCount > 1 {
		// Check for uneven distribution
		maxTime := 0.0
		minTime := analysis.TotalTimeMs
		for _, t := range analysis.ByThread {
			if t > maxTime {
				maxTime = t
			}
			if t < minTime {
				minTime = t
			}
		}
		if maxTime > 0 && minTime/maxTime < 0.5 {
			analysis.Recommendations = append(analysis.Recommendations,
				"Uneven work distribution across workers - consider better load balancing")
		}
	}

	return analysis
}

// FormatJSCryptoAnalysis returns a human-readable summary
func FormatJSCryptoAnalysis(analysis JSCryptoAnalysis) string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("JavaScript Crypto Analysis (%d samples)\n", analysis.TotalSamples))
	sb.WriteString(strings.Repeat("=", 60) + "\n\n")

	sb.WriteString(fmt.Sprintf("Total Time:         %.2fms\n", analysis.TotalTimeMs))
	sb.WriteString(fmt.Sprintf("Worker Count:       %d\n", analysis.WorkerCount))
	if analysis.WorkerCount > 0 {
		sb.WriteString(fmt.Sprintf("Avg Time/Worker:    %.2fms\n", analysis.AvgTimePerWorker))
	}
	sb.WriteString("\n")

	if len(analysis.ByThread) > 0 {
		sb.WriteString("Time by Thread:\n")
		sb.WriteString(strings.Repeat("-", 50) + "\n")
		for thread, timeMs := range analysis.ByThread {
			name := thread
			if len(name) > 30 {
				name = name[:27] + "..."
			}
			sb.WriteString(fmt.Sprintf("  %-30s %.2fms\n", name, timeMs))
		}
		sb.WriteString("\n")
	}

	if len(analysis.Resources) > 0 {
		sb.WriteString("Crypto Resources:\n")
		sb.WriteString(strings.Repeat("-", 50) + "\n")
		for i, res := range analysis.Resources {
			if i >= 10 {
				sb.WriteString(fmt.Sprintf("  ... and %d more resources\n", len(analysis.Resources)-10))
				break
			}
			name := res.Name
			if len(name) > 35 {
				name = name[:32] + "..."
			}
			sb.WriteString(fmt.Sprintf("  %-35s %.2fms (%d samples)\n", name, res.TotalTime, res.SampleCount))
		}
		sb.WriteString("\n")
	}

	if len(analysis.TopFunctions) > 0 {
		sb.WriteString("Top Functions:\n")
		sb.WriteString(strings.Repeat("-", 60) + "\n")
		sb.WriteString(fmt.Sprintf("%-40s %10s %6s\n", "Function", "Time", "%"))
		sb.WriteString(strings.Repeat("-", 60) + "\n")

		for i, fn := range analysis.TopFunctions {
			if i >= 15 {
				sb.WriteString(fmt.Sprintf("  ... and %d more functions\n", len(analysis.TopFunctions)-15))
				break
			}
			name := fn.Name
			if len(name) > 40 {
				name = name[:37] + "..."
			}
			sb.WriteString(fmt.Sprintf("%-40s %8.2fms %5.1f%%\n", name, fn.TotalTime, fn.Percent))
		}
		sb.WriteString("\n")
	}

	if len(analysis.Recommendations) > 0 {
		sb.WriteString("Recommendations:\n")
		for _, r := range analysis.Recommendations {
			sb.WriteString(fmt.Sprintf("  - %s\n", r))
		}
	}

	return sb.String()
}
