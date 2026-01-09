package analyzer

import (
	"fmt"
	"sort"
	"strings"

	"github.com/CedricHerzog/perfowl/internal/parser"
)

// CryptoOperation represents a single crypto operation found in the profile
type CryptoOperation struct {
	Operation  string  `json:"operation"`
	Algorithm  string  `json:"algorithm,omitempty"`
	DurationMs float64 `json:"duration_ms"`
	ThreadName string  `json:"thread_name"`
	StartTime  float64 `json:"start_time_ms"`
	FuncName   string  `json:"function_name,omitempty"`
}

// CryptoAnalysis contains the full crypto operation analysis
type CryptoAnalysis struct {
	TotalOperations int               `json:"total_operations"`
	TotalTimeMs     float64           `json:"total_time_ms"`
	ByOperation     map[string]float64 `json:"by_operation"`
	ByAlgorithm     map[string]float64 `json:"by_algorithm"`
	ByThread        map[string]float64 `json:"by_thread"`
	TopOperations   []CryptoOperation `json:"top_operations,omitempty"`
	Serialized      bool              `json:"possibly_serialized"`
	Warnings        []string          `json:"warnings,omitempty"`
}

// cryptoKeywords are function name patterns that indicate crypto operations
var cryptoKeywords = []string{
	"crypto", "subtlecrypto", "encrypt", "decrypt", "digest",
	"sign", "verify", "hash", "pbkdf", "hmac", "aes", "rsa",
	"sha-", "sha1", "sha256", "sha384", "sha512", "md5",
	"derivebits", "derivekey", "generatekey", "importkey", "exportkey",
	"wrapkey", "unwrapkey",
}

// operationNames maps function patterns to friendly operation names
var operationNames = map[string]string{
	"encrypt":     "encrypt",
	"decrypt":     "decrypt",
	"digest":      "digest",
	"sign":        "sign",
	"verify":      "verify",
	"hash":        "hash",
	"pbkdf":       "key derivation",
	"hmac":        "HMAC",
	"derivebits":  "deriveBits",
	"derivekey":   "deriveKey",
	"generatekey": "generateKey",
	"importkey":   "importKey",
	"exportkey":   "exportKey",
	"wrapkey":     "wrapKey",
	"unwrapkey":   "unwrapKey",
}

// algorithmNames maps patterns to algorithm names
var algorithmNames = map[string]string{
	"aes":    "AES",
	"rsa":    "RSA",
	"sha-1":  "SHA-1",
	"sha1":   "SHA-1",
	"sha256": "SHA-256",
	"sha384": "SHA-384",
	"sha512": "SHA-512",
	"md5":    "MD5",
	"pbkdf":  "PBKDF2",
	"hmac":   "HMAC",
	"ecdsa":  "ECDSA",
	"ecdh":   "ECDH",
}

// isCryptoFunction checks if a function name is crypto-related
func isCryptoFunction(name string) bool {
	lower := strings.ToLower(name)
	for _, keyword := range cryptoKeywords {
		if strings.Contains(lower, keyword) {
			return true
		}
	}
	return false
}

// extractOperation extracts the operation type from a function name
func extractOperation(name string) string {
	lower := strings.ToLower(name)
	for pattern, opName := range operationNames {
		if strings.Contains(lower, pattern) {
			return opName
		}
	}
	// Default to the function name or "crypto" if generic
	if strings.Contains(lower, "crypto") || strings.Contains(lower, "subtlecrypto") {
		return "crypto (generic)"
	}
	return "unknown"
}

// extractAlgorithm extracts the algorithm from a function name
func extractAlgorithm(name string) string {
	lower := strings.ToLower(name)
	for pattern, algoName := range algorithmNames {
		if strings.Contains(lower, pattern) {
			return algoName
		}
	}
	return ""
}

// isCryptoResource checks if a resource/file name is crypto-related
func isCryptoResource(name string) bool {
	lower := strings.ToLower(name)
	cryptoResources := []string{
		"crypto", "decrypt", "encrypt", "libcorecrypto", "libcommoncrypto",
		"openssl", "boringssl", "nss", "pgp", "gpg", "seipd",
	}
	for _, keyword := range cryptoResources {
		if strings.Contains(lower, keyword) {
			return true
		}
	}
	return false
}

// AnalyzeCrypto performs crypto operation analysis on the profile
// Uses an optimized approach: pre-index crypto functions and resources, then only check leaf frames
func AnalyzeCrypto(profile *parser.Profile) CryptoAnalysis {
	analysis := CryptoAnalysis{
		ByOperation:   make(map[string]float64),
		ByAlgorithm:   make(map[string]float64),
		ByThread:      make(map[string]float64),
		TopOperations: make([]CryptoOperation, 0),
		Warnings:      make([]string, 0),
	}

	interval := profile.Meta.Interval

	// Pre-identify crypto library indices from profile.Libs
	cryptoLibIndices := make(map[int]string) // libIdx -> lib name
	for i, lib := range profile.Libs {
		libName := strings.ToLower(lib.Name)
		if strings.Contains(libName, "crypto") || strings.Contains(libName, "ssl") ||
			strings.Contains(libName, "nss") || strings.Contains(libName, "gpg") {
			cryptoLibIndices[i] = lib.Name
		}
	}

	// Track crypto operations for aggregation
	type cryptoAgg struct {
		operation  string
		algorithm  string
		funcName   string
		threadName string
		totalTime  float64
		count      int
		firstTime  float64
	}
	cryptoFuncs := make(map[string]*cryptoAgg) // key: funcName+threadName

	// Track time windows for serialization detection
	type timeWindow struct {
		threadName string
		start      float64
		end        float64
	}
	var cryptoWindows []timeWindow

	for _, thread := range profile.Threads {
		// Use shared string array if thread's is empty (Firefox profile optimization)
		stringArray := thread.StringArray
		if len(stringArray) == 0 {
			stringArray = profile.Shared.StringArray
		}

		// Step 1: Pre-build lookup tables
		// Map funcIdx -> function name string
		funcIdxToName := make(map[int]string)
		for funcIdx, nameIdx := range thread.FuncTable.Name {
			if nameIdx >= 0 && nameIdx < len(stringArray) {
				funcIdxToName[funcIdx] = stringArray[nameIdx]
			}
		}

		// Map funcIdx -> file name string
		funcIdxToFile := make(map[int]string)
		for funcIdx, fileIdx := range thread.FuncTable.FileName {
			if fileIdx >= 0 && fileIdx < len(stringArray) {
				funcIdxToFile[funcIdx] = stringArray[fileIdx]
			}
		}

		// Map resourceIdx -> resource name (from StringArray)
		resourceIdxToName := make(map[int]string)
		for resIdx, nameIdx := range thread.ResourceTable.Name {
			if nameIdx >= 0 && nameIdx < len(stringArray) {
				resourceIdxToName[resIdx] = stringArray[nameIdx]
			}
		}

		// Map resourceIdx -> libIdx (from ResourceTable.Lib)
		resourceIdxToLib := make(map[int]int)
		for resIdx, libIdx := range thread.ResourceTable.Lib {
			resourceIdxToLib[resIdx] = libIdx
		}

		// Map funcIdx -> resourceIdx
		funcIdxToResource := make(map[int]int)
		for funcIdx, resIdx := range thread.FuncTable.Resource {
			funcIdxToResource[funcIdx] = resIdx
		}

		// Step 2: Pre-identify which funcIdx are crypto-related
		// Check: function name, file name, resource name, or linked library
		cryptoFuncIdx := make(map[int]string) // funcIdx -> display name for crypto
		for funcIdx := range funcIdxToName {
			funcName := funcIdxToName[funcIdx]

			// Check function name
			if isCryptoFunction(funcName) {
				cryptoFuncIdx[funcIdx] = funcName
				continue
			}

			// Check file name
			if fileName, ok := funcIdxToFile[funcIdx]; ok && isCryptoResource(fileName) {
				displayName := fileName
				if funcName != "" && funcName != "(unknown)" {
					displayName = funcName + " (" + fileName + ")"
				}
				cryptoFuncIdx[funcIdx] = displayName
				continue
			}

			// Check resource name
			if resIdx, ok := funcIdxToResource[funcIdx]; ok {
				if resName, ok := resourceIdxToName[resIdx]; ok && isCryptoResource(resName) {
					displayName := resName
					if funcName != "" && funcName != "(unknown)" {
						displayName = funcName + " (" + resName + ")"
					}
					cryptoFuncIdx[funcIdx] = displayName
					continue
				}

				// Check if resource links to a crypto library (via profile.Libs)
				if libIdx, ok := resourceIdxToLib[resIdx]; ok {
					if libName, isCrypto := cryptoLibIndices[libIdx]; isCrypto {
						displayName := libName
						if funcName != "" && funcName != "(unknown)" && len(funcName) > 2 {
							displayName = funcName + " (" + libName + ")"
						}
						cryptoFuncIdx[funcIdx] = displayName
					}
				}
			}
		}

		// Step 3: Map frameIdx -> funcIdx for quick lookup
		frameToFunc := make(map[int]int)
		for frameIdx, funcIdx := range thread.FrameTable.Func {
			frameToFunc[frameIdx] = funcIdx
		}

		// Step 4: Map stackIdx -> frameIdx (leaf frame)
		stackToFrame := make(map[int]int)
		for stackIdx, frameIdx := range thread.StackTable.Frame {
			stackToFrame[stackIdx] = frameIdx
		}

		// Step 5: Process samples - only check leaf function
		sampleTime := 0.0
		for i := 0; i < thread.Samples.Length; i++ {
			cpuDelta := interval
			if i < len(thread.Samples.ThreadCPUDelta) && thread.Samples.ThreadCPUDelta[i] > 0 {
				cpuDelta = float64(thread.Samples.ThreadCPUDelta[i]) / 1000.0
			}

			stackIdx := -1
			if i < len(thread.Samples.Stack) {
				stackIdx = thread.Samples.Stack[i]
			}

			if stackIdx >= 0 {
				// Get leaf frame
				frameIdx, ok := stackToFrame[stackIdx]
				if ok {
					funcIdx, ok := frameToFunc[frameIdx]
					if ok {
						// Check if this function is crypto-related
						displayName, isCrypto := cryptoFuncIdx[funcIdx]
						if isCrypto {
							key := displayName + "|" + thread.Name

							if cryptoFuncs[key] == nil {
								cryptoFuncs[key] = &cryptoAgg{
									operation:  extractOperation(displayName),
									algorithm:  extractAlgorithm(displayName),
									funcName:   displayName,
									threadName: thread.Name,
									firstTime:  sampleTime,
								}
							}
							cryptoFuncs[key].totalTime += cpuDelta
							cryptoFuncs[key].count++

							// Track window for serialization detection
							cryptoWindows = append(cryptoWindows, timeWindow{
								threadName: thread.Name,
								start:      sampleTime,
								end:        sampleTime + cpuDelta,
							})
						}
					}
				}
			}

			sampleTime += cpuDelta
		}
	}

	// Aggregate results
	threadSet := make(map[string]bool)
	for _, agg := range cryptoFuncs {
		analysis.TotalOperations += agg.count
		analysis.TotalTimeMs += agg.totalTime
		analysis.ByOperation[agg.operation] += agg.totalTime
		analysis.ByThread[agg.threadName] += agg.totalTime
		threadSet[agg.threadName] = true

		if agg.algorithm != "" {
			analysis.ByAlgorithm[agg.algorithm] += agg.totalTime
		}

		// Add to top operations
		analysis.TopOperations = append(analysis.TopOperations, CryptoOperation{
			Operation:  agg.operation,
			Algorithm:  agg.algorithm,
			DurationMs: agg.totalTime,
			ThreadName: agg.threadName,
			StartTime:  agg.firstTime,
			FuncName:   agg.funcName,
		})
	}

	// Sort operations by duration (descending)
	sort.Slice(analysis.TopOperations, func(i, j int) bool {
		return analysis.TopOperations[i].DurationMs > analysis.TopOperations[j].DurationMs
	})

	// Limit to top 20
	if len(analysis.TopOperations) > 20 {
		analysis.TopOperations = analysis.TopOperations[:20]
	}

	// Detect serialization: check if crypto windows overlap across threads
	if len(threadSet) > 1 && len(cryptoWindows) > 1 {
		sort.Slice(cryptoWindows, func(i, j int) bool {
			return cryptoWindows[i].start < cryptoWindows[j].start
		})

		hasOverlap := false
		for i := 1; i < len(cryptoWindows); i++ {
			// Check if different threads have overlapping crypto
			if cryptoWindows[i].threadName != cryptoWindows[i-1].threadName {
				if cryptoWindows[i].start < cryptoWindows[i-1].end {
					hasOverlap = true
					break
				}
			}
		}

		if !hasOverlap && analysis.TotalOperations > 5 {
			analysis.Serialized = true
			analysis.Warnings = append(analysis.Warnings,
				"Crypto operations appear serialized despite multiple threads - consider parallelizing")
		}
	}

	// Generate warnings
	if analysis.TotalTimeMs > 100 {
		analysis.Warnings = append(analysis.Warnings,
			fmt.Sprintf("Significant crypto overhead: %.1fms total", analysis.TotalTimeMs))
	}

	if time, ok := analysis.ByAlgorithm["SHA-1"]; ok && time > 0 {
		analysis.Warnings = append(analysis.Warnings,
			"SHA-1 usage detected - consider upgrading to SHA-256 for security")
	}

	if time, ok := analysis.ByAlgorithm["MD5"]; ok && time > 0 {
		analysis.Warnings = append(analysis.Warnings,
			"MD5 usage detected - MD5 is cryptographically broken, use SHA-256")
	}

	return analysis
}

// FormatCryptoAnalysis returns a human-readable summary
func FormatCryptoAnalysis(analysis CryptoAnalysis) string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("Crypto Operation Analysis (%d samples)\n", analysis.TotalOperations))
	sb.WriteString(strings.Repeat("=", 60) + "\n\n")

	sb.WriteString(fmt.Sprintf("Total Time: %.2fms\n", analysis.TotalTimeMs))
	if analysis.Serialized {
		sb.WriteString("Serialization: Possibly serialized (no parallel crypto detected)\n")
	}
	sb.WriteString("\n")

	if len(analysis.ByOperation) > 0 {
		sb.WriteString("Time by Operation:\n")
		sb.WriteString(strings.Repeat("-", 40) + "\n")
		for op, time := range analysis.ByOperation {
			sb.WriteString(fmt.Sprintf("  %-20s %.2fms\n", op, time))
		}
		sb.WriteString("\n")
	}

	if len(analysis.ByAlgorithm) > 0 {
		sb.WriteString("Time by Algorithm:\n")
		sb.WriteString(strings.Repeat("-", 40) + "\n")
		for algo, time := range analysis.ByAlgorithm {
			sb.WriteString(fmt.Sprintf("  %-20s %.2fms\n", algo, time))
		}
		sb.WriteString("\n")
	}

	if len(analysis.ByThread) > 0 {
		sb.WriteString("Time by Thread:\n")
		sb.WriteString(strings.Repeat("-", 40) + "\n")
		for thread, time := range analysis.ByThread {
			name := thread
			if len(name) > 30 {
				name = name[:27] + "..."
			}
			sb.WriteString(fmt.Sprintf("  %-30s %.2fms\n", name, time))
		}
		sb.WriteString("\n")
	}

	if len(analysis.Warnings) > 0 {
		sb.WriteString("Warnings:\n")
		for _, w := range analysis.Warnings {
			sb.WriteString(fmt.Sprintf("  - %s\n", w))
		}
	}

	return sb.String()
}
