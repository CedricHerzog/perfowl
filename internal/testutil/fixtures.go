package testutil

import (
	"github.com/CedricHerzog/perfowl/internal/parser"
)

// DefaultCategories returns the standard Firefox profiler categories.
func DefaultCategories() []parser.Category {
	return []parser.Category{
		{Name: "Idle", Color: "transparent"},
		{Name: "Other", Color: "grey"},
		{Name: "JavaScript", Color: "yellow"},
		{Name: "Layout", Color: "purple"},
		{Name: "Graphics", Color: "green"},
		{Name: "DOM", Color: "blue"},
		{Name: "GC / CC", Color: "orange"},
		{Name: "Network", Color: "lightblue"},
	}
}

// MinimalProfile returns an empty but valid profile.
func MinimalProfile() *parser.Profile {
	return NewProfileBuilder().
		WithDuration(1000).
		WithCategories(DefaultCategories()).
		Build()
}

// ProfileWithMainThread returns a profile with a single main thread.
func ProfileWithMainThread() *parser.Profile {
	return NewProfileBuilder().
		WithDuration(1000).
		WithCategories(DefaultCategories()).
		WithThread(NewThreadBuilder("GeckoMain").
			AsMainThread().
			WithProcessType("default").
			Build()).
		Build()
}

// ProfileWithGC returns a profile with GC markers.
func ProfileWithGC() *parser.Profile {
	mb := NewMarkerBuilder()
	mb.AddGCMajor(100, 50).
		AddGCMajor(300, 60).
		AddGCMinor(500, 10).
		AddGCMinor(600, 15)
	markers, strings := mb.Build()

	return NewProfileBuilder().
		WithDuration(1000).
		WithCategories(DefaultCategories()).
		WithThread(NewThreadBuilder("GeckoMain").
			AsMainThread().
			WithMarkers(markers).
			WithStringArray(strings).
			Build()).
		Build()
}

// ProfileWithLongTasks returns a profile with long task markers.
func ProfileWithLongTasks(count int) *parser.Profile {
	mb := NewMarkerBuilder()
	for i := 0; i < count; i++ {
		// Each long task is 100ms, spaced 200ms apart
		mb.AddLongTask(float64(i*200), 100)
	}
	markers, strings := mb.Build()

	return NewProfileBuilder().
		WithDuration(float64(count * 200)).
		WithCategories(DefaultCategories()).
		WithThread(NewThreadBuilder("GeckoMain").
			AsMainThread().
			WithMarkers(markers).
			WithStringArray(strings).
			Build()).
		Build()
}

// ProfileWithWorkers returns a profile with worker threads.
func ProfileWithWorkers(count int) *parser.Profile {
	pb := NewProfileBuilder().
		WithDuration(1000).
		WithCategories(DefaultCategories()).
		WithThread(NewThreadBuilder("GeckoMain").
			AsMainThread().
			Build())

	for i := 0; i < count; i++ {
		// Create samples for the worker
		sb := NewSamplesBuilder()
		for j := 0; j < 100; j++ {
			sb.AddSampleWithCPUDelta(0, float64(j*10), 1000) // 1ms CPU per sample
		}

		pb.WithThread(NewThreadBuilder("DOM Worker").
			WithTID(string(rune('2' + i))).
			WithSamples(sb.Build()).
			Build())
	}

	return pb.Build()
}

// ProfileWithExtensions returns a profile with browser extensions.
func ProfileWithExtensions() *parser.Profile {
	mb := NewMarkerBuilder()
	mb.AddJSActorMessage("extension1", 100, 50).
		AddDOMEvent("click", 200)
	markers, strings := mb.Build()

	return NewProfileBuilder().
		WithDuration(1000).
		WithCategories(DefaultCategories()).
		WithExtension("ext1@example.com", "Test Extension 1", "moz-extension://abc123/").
		WithExtension("ext2@example.com", "Test Extension 2", "moz-extension://def456/").
		WithThread(NewThreadBuilder("GeckoMain").
			AsMainThread().
			WithMarkers(markers).
			WithStringArray(strings).
			Build()).
		Build()
}

// ProfileWithLayoutThrashing returns a profile with rapid layout/reflow markers.
func ProfileWithLayoutThrashing() *parser.Profile {
	mb := NewMarkerBuilder()
	// Add 10 reflows in a 100ms window (thrashing)
	for i := 0; i < 10; i++ {
		mb.AddLayout(float64(i*10), 5)
	}
	markers, strings := mb.Build()

	return NewProfileBuilder().
		WithDuration(1000).
		WithCategories(DefaultCategories()).
		WithThread(NewThreadBuilder("GeckoMain").
			AsMainThread().
			WithMarkers(markers).
			WithStringArray(strings).
			Build()).
		Build()
}

// ProfileWithSyncIPC returns a profile with sync IPC markers.
func ProfileWithSyncIPC() *parser.Profile {
	mb := NewMarkerBuilder()
	mb.AddIPC(100, 50, true).
		AddIPC(200, 30, true).
		AddIPC(300, 20, false) // async
	markers, strings := mb.Build()

	return NewProfileBuilder().
		WithDuration(1000).
		WithCategories(DefaultCategories()).
		WithThread(NewThreadBuilder("GeckoMain").
			AsMainThread().
			WithMarkers(markers).
			WithStringArray(strings).
			Build()).
		Build()
}

// ProfileWithNetworkBlocking returns a profile with slow network markers.
func ProfileWithNetworkBlocking() *parser.Profile {
	mb := NewMarkerBuilder()
	mb.AddNetwork("https://slow.example.com/api", 0, 2000). // 2s - blocking
								AddNetwork("https://fast.example.com/api", 100, 100) // 100ms - ok
	markers, strings := mb.Build()

	return NewProfileBuilder().
		WithDuration(3000).
		WithCategories(DefaultCategories()).
		WithThread(NewThreadBuilder("GeckoMain").
			AsMainThread().
			WithMarkers(markers).
			WithStringArray(strings).
			Build()).
		Build()
}

// ProfileWithCrypto returns a profile with crypto-related samples.
// This profile generates actual crypto analysis data with operations, algorithms, etc.
func ProfileWithCrypto() *parser.Profile {
	// Build string array with crypto function names that trigger detection
	strings := []string{
		"SubtleCrypto.encrypt",     // 0 - triggers encrypt operation
		"SubtleCrypto.decrypt",     // 1 - triggers decrypt operation
		"SubtleCrypto.digest",      // 2 - triggers digest operation
		"AES-GCM-encrypt",          // 3 - triggers AES algorithm
		"SHA-256-digest",           // 4 - triggers SHA-256 algorithm
		"SHA-1-hash",               // 5 - triggers SHA-1 (warning)
		"crypto.js",                // 6 - resource name
		"MD5-checksum",             // 7 - triggers MD5 (warning)
		"SubtleCrypto.generateKey", // 8 - triggers generateKey
		"PBKDF2-deriveKey",         // 9 - triggers key derivation
	}

	// Build stack table - each stack points to a frame
	stb := NewStackTableBuilder()
	for i := 0; i < 10; i++ {
		stb.AddStack(i, 2, -1) // frame i, JavaScript category, no prefix
	}

	// Build frame table - each frame points to a function
	ftb := NewFrameTableBuilder()
	for i := 0; i < 10; i++ {
		ftb.AddFrame(i, 2) // func i, JavaScript category
	}

	// Build function table - each function has a name
	fnb := NewFuncTableBuilder()
	for i := 0; i < 10; i++ {
		fnb.AddFunc(i, true, 0) // name i, isJS, resource 0
	}

	// Build resource table
	rtb := NewResourceTableBuilder()
	rtb.AddResource(-1, 6, -1, 1) // crypto.js

	// Build samples - lots of samples to accumulate time (>100ms for warnings)
	sb := NewSamplesBuilder()
	// 30 samples each for different operations (each 5ms = 150ms total for warnings)
	for i := 0; i < 30; i++ {
		sb.AddSampleWithCPUDelta(0, float64(i*10), 5000)   // encrypt
		sb.AddSampleWithCPUDelta(1, float64(i*10+1), 5000) // decrypt
		sb.AddSampleWithCPUDelta(2, float64(i*10+2), 5000) // digest
		sb.AddSampleWithCPUDelta(3, float64(i*10+3), 5000) // AES
		sb.AddSampleWithCPUDelta(4, float64(i*10+4), 5000) // SHA-256
		sb.AddSampleWithCPUDelta(5, float64(i*10+5), 5000) // SHA-1
		sb.AddSampleWithCPUDelta(7, float64(i*10+6), 5000) // MD5
		sb.AddSampleWithCPUDelta(8, float64(i*10+7), 5000) // generateKey
		sb.AddSampleWithCPUDelta(9, float64(i*10+8), 5000) // PBKDF2
	}

	return NewProfileBuilder().
		WithDuration(1000).
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

// ProfileWithJSCrypto returns a profile with JS crypto worker samples.
// This profile generates actual JS crypto analysis data with openpgp functions.
func ProfileWithJSCrypto() *parser.Profile {
	// String array with openpgp/crypto function names that trigger detection
	strings := []string{
		"decryptWithSessionKey",                        // 0 - openpgp function
		"SEIPDPacket",                                  // 1 - openpgp function
		"processPacket",                                // 2 - openpgp function
		"moz-extension://abc123/openpgp.min.js",        // 3 - crypto resource
		"encryptMessage",                               // 4 - openpgp function
		"armored_decrypt",                              // 5 - decrypt function
		"moz-extension://def456/seipdDecryptWorker.js", // 6 - decrypt worker resource
	}

	// Build stack table - multiple stacks for different functions
	stb := NewStackTableBuilder()
	for i := 0; i < 6; i++ {
		stb.AddStack(i, 2, -1) // frame i, JavaScript category, no prefix
	}

	// Build frame table
	ftb := NewFrameTableBuilder()
	for i := 0; i < 6; i++ {
		ftb.AddFrame(i, 2) // func i, JavaScript category
	}

	// Build function table with file references
	fnb := NewFuncTableBuilder()
	fnb.AddFuncWithFile(0, true, 0, 3, 100). // decryptWithSessionKey in openpgp.min.js
							AddFuncWithFile(1, true, 0, 3, 200). // SEIPDPacket in openpgp.min.js
							AddFuncWithFile(2, true, 0, 3, 300). // processPacket in openpgp.min.js
							AddFuncWithFile(4, true, 0, 3, 400). // encryptMessage in openpgp.min.js
							AddFuncWithFile(5, true, 1, 6, 100). // armored_decrypt in seipdDecryptWorker.js
							AddFunc(3, true, 0)                  // placeholder

	// Build resource tables - two resources for different crypto workers
	rtb := NewResourceTableBuilder()
	rtb.AddResource(-1, 3, -1, 1). // openpgp.min.js
					AddResource(-1, 6, -1, 1) // seipdDecryptWorker.js

	// Build samples for crypto worker - lots of samples
	sb := NewSamplesBuilder()
	for i := 0; i < 100; i++ {
		sb.AddSampleWithCPUDelta(i%5, float64(i*10), 2000) // 2ms per sample
	}

	pb := NewProfileBuilder().
		WithDuration(1000).
		WithCategories(DefaultCategories())

	// Add main thread
	pb.WithThread(NewThreadBuilder("GeckoMain").
		AsMainThread().
		Build())

	// Add first crypto worker
	pb.WithThread(NewThreadBuilder("DOM Worker").
		WithTID("2").
		WithStringArray(strings).
		WithStackTable(stb.Build()).
		WithFrameTable(ftb.Build()).
		WithFuncTable(fnb.Build()).
		WithResourceTable(rtb.Build()).
		WithSamples(sb.Build()).
		Build())

	// Add second crypto worker for multi-worker detection
	sb2 := NewSamplesBuilder()
	for i := 0; i < 50; i++ {
		sb2.AddSampleWithCPUDelta(i%3, float64(i*10), 2000)
	}
	pb.WithThread(NewThreadBuilder("DOM Worker").
		WithTID("3").
		WithStringArray(strings).
		WithStackTable(stb.Build()).
		WithFrameTable(ftb.Build()).
		WithFuncTable(fnb.Build()).
		WithResourceTable(rtb.Build()).
		WithSamples(sb2.Build()).
		Build())

	return pb.Build()
}

// ProfileWithContention returns a profile with contention markers.
func ProfileWithContention() *parser.Profile {
	// Main thread with GC
	mainMb := NewMarkerBuilder()
	mainMb.AddGCMajor(100, 200) // Long GC while workers are active
	mainMarkers, mainStrings := mainMb.Build()

	// Worker with IPC during GC
	workerMb := NewMarkerBuilder()
	workerMb.AddIPC(150, 50, true) // Sync IPC during GC
	workerMarkers, workerStrings := workerMb.Build()

	// Build samples for worker
	sb := NewSamplesBuilder()
	for i := 0; i < 50; i++ {
		sb.AddSampleWithCPUDelta(0, float64(100+i*4), 1000)
	}

	return NewProfileBuilder().
		WithDuration(1000).
		WithCategories(DefaultCategories()).
		WithThread(NewThreadBuilder("GeckoMain").
			AsMainThread().
			WithMarkers(mainMarkers).
			WithStringArray(mainStrings).
			Build()).
		WithThread(NewThreadBuilder("DOM Worker").
			WithTID("2").
			WithMarkers(workerMarkers).
			WithStringArray(workerStrings).
			WithSamples(sb.Build()).
			Build()).
		Build()
}

// ProfileWithDelimiters returns a profile with delimiter markers for operation timing.
func ProfileWithDelimiters() *parser.Profile {
	mb := NewMarkerBuilder()
	mb.AddDOMEventWithDuration("click", 100, 10).
		AddStyles(120, 30).
		AddPaint(160, 20).
		AddUpdateLayoutTree(200, 15)
	markers, strings := mb.Build()

	return NewProfileBuilder().
		WithDuration(1000).
		WithCategories(DefaultCategories()).
		WithThread(NewThreadBuilder("GeckoMain").
			AsMainThread().
			WithMarkers(markers).
			WithStringArray(strings).
			Build()).
		Build()
}

// ProfileWithCallTree returns a profile with a proper call tree for analysis.
func ProfileWithCallTree() *parser.Profile {
	// String array with function names
	strings := []string{
		"main",
		"processData",
		"computeHash",
		"render",
		"updateDOM",
	}

	// Build stack table with call chains
	// Stack 0: main
	// Stack 1: main -> processData
	// Stack 2: main -> processData -> computeHash
	// Stack 3: main -> render
	// Stack 4: main -> render -> updateDOM
	stb := NewStackTableBuilder()
	stb.AddStack(0, 2, -1). // main (no prefix)
				AddStack(1, 2, 0). // processData -> main
				AddStack(2, 2, 1). // computeHash -> processData
				AddStack(3, 2, 0). // render -> main
				AddStack(4, 2, 3)  // updateDOM -> render

	// Build frame table
	ftb := NewFrameTableBuilder()
	for i := 0; i < 5; i++ {
		ftb.AddFrame(i, 2) // All JavaScript category
	}

	// Build function table
	fnb := NewFuncTableBuilder()
	for i := 0; i < 5; i++ {
		fnb.AddFunc(i, true, -1)
	}

	// Build samples with varying stack indices
	sb := NewSamplesBuilder()
	// Spend most time in computeHash
	for i := 0; i < 50; i++ {
		sb.AddSampleWithCPUDelta(2, float64(i*10), 1000) // computeHash
	}
	// Some time in updateDOM
	for i := 0; i < 30; i++ {
		sb.AddSampleWithCPUDelta(4, float64(500+i*10), 1000) // updateDOM
	}
	// Less time in render
	for i := 0; i < 20; i++ {
		sb.AddSampleWithCPUDelta(3, float64(800+i*10), 1000) // render
	}

	return NewProfileBuilder().
		WithDuration(1000).
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

// ProfileWithCategories returns a profile with samples across multiple categories.
func ProfileWithCategories() *parser.Profile {
	strings := []string{
		"jsFunction",
		"layoutFunc",
		"paintFunc",
		"domFunc",
	}

	// Build stack table with different categories
	stb := NewStackTableBuilder()
	stb.AddStack(0, 2, -1). // JavaScript (category 2)
				AddStack(1, 3, -1). // Layout (category 3)
				AddStack(2, 4, -1). // Graphics (category 4)
				AddStack(3, 5, -1)  // DOM (category 5)

	// Build frame table
	ftb := NewFrameTableBuilder()
	ftb.AddFrame(0, 2). // JavaScript
				AddFrame(1, 3). // Layout
				AddFrame(2, 4). // Graphics
				AddFrame(3, 5)  // DOM

	// Build function table
	fnb := NewFuncTableBuilder()
	for i := 0; i < 4; i++ {
		fnb.AddFunc(i, i == 0, -1) // Only first is JS
	}

	// Build samples across categories
	sb := NewSamplesBuilder()
	// 50% JavaScript
	for i := 0; i < 50; i++ {
		sb.AddSampleWithCPUDelta(0, float64(i*10), 1000)
	}
	// 20% Layout
	for i := 0; i < 20; i++ {
		sb.AddSampleWithCPUDelta(1, float64(500+i*10), 1000)
	}
	// 20% Graphics
	for i := 0; i < 20; i++ {
		sb.AddSampleWithCPUDelta(2, float64(700+i*10), 1000)
	}
	// 10% DOM
	for i := 0; i < 10; i++ {
		sb.AddSampleWithCPUDelta(3, float64(900+i*10), 1000)
	}

	return NewProfileBuilder().
		WithDuration(1000).
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

// ProfileWithWakePatterns returns a profile with Awake markers for wake analysis.
func ProfileWithWakePatterns() *parser.Profile {
	mb := NewMarkerBuilder()
	// Regular wake pattern every 100ms
	for i := 0; i < 10; i++ {
		mb.AddAwake(float64(i * 100))
	}
	markers, strings := mb.Build()

	return NewProfileBuilder().
		WithDuration(1000).
		WithCategories(DefaultCategories()).
		WithThread(NewThreadBuilder("GeckoMain").
			AsMainThread().
			WithMarkers(markers).
			WithStringArray(strings).
			Build()).
		Build()
}

// ProfileWithBottlenecks returns a profile with all types of bottleneck markers.
// This triggers all branches in the bottleneck detection output functions.
func ProfileWithBottlenecks() *parser.Profile {
	mb := NewMarkerBuilder()

	// Long tasks (>50ms)
	mb.AddLongTask(0, 100)
	mb.AddLongTask(200, 150)
	mb.AddLongTask(500, 80)

	// GC pressure - multiple GC events
	mb.AddGCMajor(100, 60)
	mb.AddGCMajor(300, 70)
	mb.AddGCMajor(600, 55)
	mb.AddGCMinor(150, 10)
	mb.AddGCMinor(350, 15)

	// Sync IPC markers
	mb.AddIPC(50, 40, true)
	mb.AddIPC(250, 35, true)
	mb.AddIPC(450, 25, true)

	// Layout thrashing - rapid reflows
	for i := 0; i < 8; i++ {
		mb.AddLayout(float64(700+i*10), 5)
	}

	// Network blocking - slow requests
	mb.AddNetwork("https://slow-api.example.com/data", 0, 1500)
	mb.AddNetwork("https://another-slow-api.example.com/endpoint", 100, 2000)

	markers, strings := mb.Build()

	return NewProfileBuilder().
		WithDuration(3000).
		WithCategories(DefaultCategories()).
		WithExtension("ext1@example.com", "Test Extension", "moz-extension://abc123/").
		WithThread(NewThreadBuilder("GeckoMain").
			AsMainThread().
			WithMarkers(markers).
			WithStringArray(strings).
			Build()).
		Build()
}

// ProfileWithExtensionActivity returns a profile with extension activity data.
// This generates actual extension analysis data for output function testing.
func ProfileWithExtensionActivity() *parser.Profile {
	// String array with extension-related URLs
	strings := []string{
		"extensionFunction",                    // 0
		"handleMessage",                        // 1
		"moz-extension://ext123/background.js", // 2 - extension resource
		"chrome-extension://ext456/content.js", // 3 - chrome extension resource
		"processExtensionData",                 // 4
	}

	// Build stack table
	stb := NewStackTableBuilder()
	for i := 0; i < 5; i++ {
		stb.AddStack(i, 2, -1)
	}

	// Build frame table
	ftb := NewFrameTableBuilder()
	for i := 0; i < 5; i++ {
		ftb.AddFrame(i, 2)
	}

	// Build function table with extension resources
	fnb := NewFuncTableBuilder()
	fnb.AddFuncWithFile(0, true, 0, 2, 10).
		AddFuncWithFile(1, true, 0, 2, 20).
		AddFuncWithFile(4, true, 1, 3, 10).
		AddFunc(2, true, 0).
		AddFunc(3, true, 1)

	// Build resource table
	rtb := NewResourceTableBuilder()
	rtb.AddResource(-1, 2, -1, 1). // moz-extension resource
					AddResource(-1, 3, -1, 1) // chrome-extension resource

	// Build samples
	sb := NewSamplesBuilder()
	for i := 0; i < 100; i++ {
		sb.AddSampleWithCPUDelta(i%5, float64(i*10), 2000)
	}

	// Add JSActorMessage markers for extension activity
	mb := NewMarkerBuilder()
	mb.AddJSActorMessage("ExtensionActor", 100, 50)
	mb.AddJSActorMessage("ExtensionActor", 300, 40)
	mb.AddJSActorMessage("ExtensionActor", 500, 60)
	markers, markerStrings := mb.Build()

	// Combine string arrays
	allStrings := append(strings, markerStrings...)

	return NewProfileBuilder().
		WithDuration(1000).
		WithCategories(DefaultCategories()).
		WithExtension("ext123@example.com", "Test Extension 1", "moz-extension://ext123/").
		WithExtension("ext456@example.com", "Test Extension 2", "chrome-extension://ext456/").
		WithThread(NewThreadBuilder("GeckoMain").
			AsMainThread().
			WithStringArray(allStrings).
			WithStackTable(stb.Build()).
			WithFrameTable(ftb.Build()).
			WithFuncTable(fnb.Build()).
			WithResourceTable(rtb.Build()).
			WithSamples(sb.Build()).
			WithMarkers(markers).
			Build()).
		Build()
}

// ProfileWithContentionData returns a profile with comprehensive contention data.
// This triggers all branches in contention output functions.
func ProfileWithContentionData() *parser.Profile {
	// Main thread with GC that overlaps with worker activity
	mainMb := NewMarkerBuilder()
	mainMb.AddGCMajor(100, 200) // Long GC pause
	mainMb.AddGCMajor(400, 150) // Another long GC pause
	mainMarkers, mainStrings := mainMb.Build()

	// Worker with sync IPC during GC periods
	workerMb := NewMarkerBuilder()
	workerMb.AddIPC(120, 80, true) // Sync IPC during first GC
	workerMb.AddIPC(420, 60, true) // Sync IPC during second GC
	workerMb.AddIPC(600, 40, true) // Another sync IPC
	workerMarkers, workerStrings := workerMb.Build()

	// Build samples for workers to show activity during contention
	sb := NewSamplesBuilder()
	for i := 0; i < 100; i++ {
		sb.AddSampleWithCPUDelta(0, float64(100+i*5), 1000)
	}

	return NewProfileBuilder().
		WithDuration(1000).
		WithCategories(DefaultCategories()).
		WithThread(NewThreadBuilder("GeckoMain").
			AsMainThread().
			WithMarkers(mainMarkers).
			WithStringArray(mainStrings).
			Build()).
		WithThread(NewThreadBuilder("DOM Worker").
			WithTID("2").
			WithMarkers(workerMarkers).
			WithStringArray(workerStrings).
			WithSamples(sb.Build()).
			Build()).
		WithThread(NewThreadBuilder("DOM Worker").
			WithTID("3").
			WithMarkers(workerMarkers).
			WithStringArray(workerStrings).
			WithSamples(sb.Build()).
			Build()).
		Build()
}

// ProfileWithWorkersData returns a profile with comprehensive worker data.
// This triggers all branches in worker output functions.
func ProfileWithWorkersData() *parser.Profile {
	// Build samples for workers with varying activity
	sb1 := NewSamplesBuilder()
	for i := 0; i < 200; i++ {
		sb1.AddSampleWithCPUDelta(0, float64(i*5), 1000)
	}

	sb2 := NewSamplesBuilder()
	for i := 0; i < 150; i++ {
		sb2.AddSampleWithCPUDelta(0, float64(i*6), 800)
	}

	sb3 := NewSamplesBuilder()
	for i := 0; i < 100; i++ {
		sb3.AddSampleWithCPUDelta(0, float64(i*8), 600)
	}

	// Add worker markers
	workerMb := NewMarkerBuilder()
	workerMb.AddIPC(100, 20, true) // sync wait
	workerMb.AddIPC(300, 15, true)
	workerMarkers, workerStrings := workerMb.Build()

	return NewProfileBuilder().
		WithDuration(1000).
		WithCategories(DefaultCategories()).
		WithThread(NewThreadBuilder("GeckoMain").
			AsMainThread().
			Build()).
		WithThread(NewThreadBuilder("DOM Worker").
			WithTID("2").
			WithSamples(sb1.Build()).
			WithMarkers(workerMarkers).
			WithStringArray(workerStrings).
			Build()).
		WithThread(NewThreadBuilder("DOM Worker").
			WithTID("3").
			WithSamples(sb2.Build()).
			Build()).
		WithThread(NewThreadBuilder("DOM Worker").
			WithTID("4").
			WithSamples(sb3.Build()).
			Build()).
		WithThread(NewThreadBuilder("DOM Worker").
			WithTID("5").
			WithSamples(sb1.Build()).
			Build()).
		Build()
}
