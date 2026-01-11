package testutil

import (
	"encoding/json"
	"fmt"

	"github.com/CedricHerzog/perfowl/internal/parser"
)

// SmallProfile returns a profile with ~100 samples for quick benchmarks.
// Useful for measuring per-call overhead.
func SmallProfile() *parser.Profile {
	return ScaledProfile(100, 50, 1)
}

// MediumProfile returns a profile with ~1000 samples for realistic benchmarks.
func MediumProfile() *parser.Profile {
	return ScaledProfile(1000, 500, 2)
}

// LargeProfile returns a profile with ~10000 samples for stress testing.
func LargeProfile() *parser.Profile {
	return ScaledProfile(10000, 5000, 4)
}

// ScaledProfile creates a profile with specified sample count, marker count, and worker count.
func ScaledProfile(sampleCount, markerCount, workerCount int) *parser.Profile {
	// String array with diverse function names
	strings := make([]string, 100)
	for i := 0; i < 100; i++ {
		strings[i] = fmt.Sprintf("function_%d", i)
	}

	// Build stack table with call chains
	stb := NewStackTableBuilder()
	for i := 0; i < 50; i++ {
		prefix := -1
		if i > 0 {
			prefix = i - 1
		}
		stb.AddStack(i%50, 2, prefix)
	}

	// Build frame table
	ftb := NewFrameTableBuilder()
	for i := 0; i < 50; i++ {
		ftb.AddFrame(i, 2)
	}

	// Build function table
	fnb := NewFuncTableBuilder()
	for i := 0; i < 50; i++ {
		fnb.AddFunc(i, true, -1)
	}

	// Build samples
	sb := NewSamplesBuilder()
	for i := 0; i < sampleCount; i++ {
		sb.AddSampleWithCPUDelta(i%50, float64(i), 1000)
	}

	// Build markers
	mb := NewMarkerBuilder()
	for i := 0; i < markerCount; i++ {
		switch i % 6 {
		case 0:
			mb.AddGCMajor(float64(i*2), 10)
		case 1:
			mb.AddLongTask(float64(i*2), 60)
		case 2:
			mb.AddLayout(float64(i*2), 5)
		case 3:
			mb.AddIPC(float64(i*2), 15, i%2 == 0)
		case 4:
			mb.AddDOMEvent("click", float64(i*2))
		case 5:
			mb.AddNetwork(fmt.Sprintf("https://example.com/api/%d", i), float64(i*2), 100)
		}
	}
	markers, markerStrings := mb.Build()

	// Combine string arrays
	allStrings := append(strings, markerStrings...)

	// Build profile with main thread
	pb := NewProfileBuilder().
		WithDuration(float64(sampleCount * 2)).
		WithCategories(DefaultCategories()).
		WithExtension("ext1@example.com", "Test Extension 1", "moz-extension://abc123/").
		WithExtension("ext2@example.com", "Test Extension 2", "moz-extension://def456/")

	// Main thread
	pb.WithThread(NewThreadBuilder("GeckoMain").
		AsMainThread().
		WithStringArray(allStrings).
		WithStackTable(stb.Build()).
		WithFrameTable(ftb.Build()).
		WithFuncTable(fnb.Build()).
		WithSamples(sb.Build()).
		WithMarkers(markers).
		Build())

	// Worker threads
	for w := 0; w < workerCount; w++ {
		workerSb := NewSamplesBuilder()
		workerSamples := sampleCount / (workerCount + 1)
		for i := 0; i < workerSamples; i++ {
			workerSb.AddSampleWithCPUDelta(i%50, float64(i), 1000)
		}

		pb.WithThread(NewThreadBuilder("DOM Worker").
			WithTID(fmt.Sprintf("%d", w+2)).
			WithStringArray(strings).
			WithStackTable(stb.Build()).
			WithFrameTable(ftb.Build()).
			WithFuncTable(fnb.Build()).
			WithSamples(workerSb.Build()).
			Build())
	}

	return pb.Build()
}

// ProfileWithNWorkers creates a profile with N worker threads, each with samplesPerWorker samples.
func ProfileWithNWorkers(workers, samplesPerWorker int) *parser.Profile {
	strings := make([]string, 20)
	for i := 0; i < 20; i++ {
		strings[i] = fmt.Sprintf("workerFunc_%d", i)
	}

	stb := NewStackTableBuilder()
	for i := 0; i < 10; i++ {
		prefix := -1
		if i > 0 {
			prefix = i - 1
		}
		stb.AddStack(i, 2, prefix)
	}

	ftb := NewFrameTableBuilder()
	for i := 0; i < 10; i++ {
		ftb.AddFrame(i, 2)
	}

	fnb := NewFuncTableBuilder()
	for i := 0; i < 10; i++ {
		fnb.AddFunc(i, true, -1)
	}

	pb := NewProfileBuilder().
		WithDuration(float64(samplesPerWorker * workers)).
		WithCategories(DefaultCategories())

	// Main thread (minimal)
	pb.WithThread(NewThreadBuilder("GeckoMain").
		AsMainThread().
		Build())

	// Worker threads
	for w := 0; w < workers; w++ {
		sb := NewSamplesBuilder()
		for i := 0; i < samplesPerWorker; i++ {
			sb.AddSampleWithCPUDelta(i%10, float64(i), 1000)
		}

		pb.WithThread(NewThreadBuilder("DOM Worker").
			WithTID(fmt.Sprintf("%d", w+2)).
			WithStringArray(strings).
			WithStackTable(stb.Build()).
			WithFrameTable(ftb.Build()).
			WithFuncTable(fnb.Build()).
			WithSamples(sb.Build()).
			Build())
	}

	return pb.Build()
}

// ProfileWithNMarkers creates a profile with N markers of various types.
func ProfileWithNMarkers(count int) *parser.Profile {
	mb := NewMarkerBuilder()
	for i := 0; i < count; i++ {
		switch i % 8 {
		case 0:
			mb.AddGCMajor(float64(i), 50)
		case 1:
			mb.AddGCMinor(float64(i), 10)
		case 2:
			mb.AddLongTask(float64(i), 100)
		case 3:
			mb.AddLayout(float64(i), 5)
		case 4:
			mb.AddIPC(float64(i), 20, true)
		case 5:
			mb.AddDOMEvent("click", float64(i))
		case 6:
			mb.AddNetwork(fmt.Sprintf("https://example.com/api/%d", i), float64(i), 100)
		case 7:
			mb.AddStyles(float64(i), 15)
		}
	}
	markers, strings := mb.Build()

	return NewProfileBuilder().
		WithDuration(float64(count * 2)).
		WithCategories(DefaultCategories()).
		WithThread(NewThreadBuilder("GeckoMain").
			AsMainThread().
			WithMarkers(markers).
			WithStringArray(strings).
			Build()).
		Build()
}

// ChromeTraceWithNEvents creates a Chrome trace JSON with N events for conversion benchmarks.
func ChromeTraceWithNEvents(count int) []byte {
	events := make([]map[string]interface{}, count)
	for i := 0; i < count; i++ {
		events[i] = map[string]interface{}{
			"pid":  1,
			"tid":  1 + (i % 4), // Spread across 4 threads
			"ts":   i * 1000,
			"ph":   "X",
			"cat":  "devtools.timeline",
			"name": fmt.Sprintf("Function_%d", i%100),
			"dur":  500 + (i % 500),
		}
	}

	trace := map[string]interface{}{
		"traceEvents": events,
		"metadata": map[string]interface{}{
			"product":     "Chrome",
			"userAgent":   "Mozilla/5.0 Chrome/120.0.0.0",
			"cpuProfile":  true,
			"networkData": true,
		},
	}

	data, _ := json.Marshal(trace)
	return data
}

// ProfileWithDeepCallStack creates a profile with deep call stacks for call tree benchmarks.
func ProfileWithDeepCallStack(depth, sampleCount int) *parser.Profile {
	strings := make([]string, depth)
	for i := 0; i < depth; i++ {
		strings[i] = fmt.Sprintf("func_level_%d", i)
	}

	// Build deep stack chain
	stb := NewStackTableBuilder()
	for i := 0; i < depth; i++ {
		prefix := -1
		if i > 0 {
			prefix = i - 1
		}
		stb.AddStack(i, 2, prefix)
	}

	ftb := NewFrameTableBuilder()
	for i := 0; i < depth; i++ {
		ftb.AddFrame(i, 2)
	}

	fnb := NewFuncTableBuilder()
	for i := 0; i < depth; i++ {
		fnb.AddFunc(i, true, -1)
	}

	// Samples at various stack depths
	sb := NewSamplesBuilder()
	for i := 0; i < sampleCount; i++ {
		// Vary the stack depth - more samples at deeper stacks
		stackIdx := (i * depth / sampleCount) % depth
		if stackIdx >= depth {
			stackIdx = depth - 1
		}
		sb.AddSampleWithCPUDelta(stackIdx, float64(i), 1000)
	}

	return NewProfileBuilder().
		WithDuration(float64(sampleCount)).
		WithCategories(DefaultCategories()).
		WithThread(NewThreadBuilder("GeckoMain").
			AsMainThread().
			WithStringArray(strings).
			WithStackTable(stb.Build()).
			WithFrameTable(ftb.Build()).
			WithFuncTable(fnb.Build()).
			WithSamples(sb.Build()).
			Build()).
		Build()
}

// ProfileWithCryptoOperations creates a profile with crypto function samples for crypto benchmarks.
func ProfileWithCryptoOperations(sampleCount int) *parser.Profile {
	strings := []string{
		"SubtleCrypto.encrypt",
		"SubtleCrypto.decrypt",
		"SubtleCrypto.digest",
		"SubtleCrypto.sign",
		"SubtleCrypto.verify",
		"AES-GCM-encrypt",
		"SHA-256-digest",
		"SHA-1-hash",
		"MD5-checksum",
		"PBKDF2-deriveKey",
		"RSA-OAEP-encrypt",
		"ECDSA-sign",
		"HMAC-sign",
		"crypto.js",
	}

	stb := NewStackTableBuilder()
	for i := 0; i < len(strings)-1; i++ {
		stb.AddStack(i, 2, -1)
	}

	ftb := NewFrameTableBuilder()
	for i := 0; i < len(strings)-1; i++ {
		ftb.AddFrame(i, 2)
	}

	fnb := NewFuncTableBuilder()
	for i := 0; i < len(strings)-1; i++ {
		fnb.AddFunc(i, true, 0)
	}

	rtb := NewResourceTableBuilder()
	rtb.AddResource(-1, 13, -1, 1) // crypto.js

	sb := NewSamplesBuilder()
	for i := 0; i < sampleCount; i++ {
		sb.AddSampleWithCPUDelta(i%(len(strings)-1), float64(i), 1000)
	}

	return NewProfileBuilder().
		WithDuration(float64(sampleCount)).
		WithCategories(DefaultCategories()).
		WithThread(NewThreadBuilder("GeckoMain").
			AsMainThread().
			WithStringArray(strings).
			WithStackTable(stb.Build()).
			WithFrameTable(ftb.Build()).
			WithFuncTable(fnb.Build()).
			WithResourceTable(rtb.Build()).
			WithSamples(sb.Build()).
			Build()).
		Build()
}

// ProfileWithManyExtensions creates a profile with N extensions and activity.
func ProfileWithManyExtensions(extensionCount, sampleCount int) *parser.Profile {
	strings := make([]string, extensionCount*2)
	for i := 0; i < extensionCount; i++ {
		strings[i*2] = fmt.Sprintf("extFunc_%d", i)
		strings[i*2+1] = fmt.Sprintf("moz-extension://ext%d/background.js", i)
	}

	stb := NewStackTableBuilder()
	for i := 0; i < extensionCount; i++ {
		stb.AddStack(i, 2, -1)
	}

	ftb := NewFrameTableBuilder()
	for i := 0; i < extensionCount; i++ {
		ftb.AddFrame(i, 2)
	}

	fnb := NewFuncTableBuilder()
	for i := 0; i < extensionCount; i++ {
		fnb.AddFuncWithFile(i*2, true, i, i*2+1, 10)
	}

	rtb := NewResourceTableBuilder()
	for i := 0; i < extensionCount; i++ {
		rtb.AddResource(-1, i*2+1, -1, 1)
	}

	sb := NewSamplesBuilder()
	for i := 0; i < sampleCount; i++ {
		sb.AddSampleWithCPUDelta(i%extensionCount, float64(i), 1000)
	}

	pb := NewProfileBuilder().
		WithDuration(float64(sampleCount)).
		WithCategories(DefaultCategories())

	for i := 0; i < extensionCount; i++ {
		pb.WithExtension(
			fmt.Sprintf("ext%d@example.com", i),
			fmt.Sprintf("Extension %d", i),
			fmt.Sprintf("moz-extension://ext%d/", i),
		)
	}

	pb.WithThread(NewThreadBuilder("GeckoMain").
		AsMainThread().
		WithStringArray(strings).
		WithStackTable(stb.Build()).
		WithFrameTable(ftb.Build()).
		WithFuncTable(fnb.Build()).
		WithResourceTable(rtb.Build()).
		WithSamples(sb.Build()).
		Build())

	return pb.Build()
}
