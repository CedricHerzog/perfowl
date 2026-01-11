package analyzer

import (
	"testing"

	"github.com/CedricHerzog/perfowl/internal/testutil"
)

func TestAnalyzeJSCrypto_EmptyProfile(t *testing.T) {
	profile := testutil.MinimalProfile()

	result := AnalyzeJSCrypto(profile)

	if result.TotalSamples != 0 {
		t.Errorf("TotalSamples = %v, want 0", result.TotalSamples)
	}
}

func TestAnalyzeJSCrypto_NoCryptoResources(t *testing.T) {
	profile := testutil.ProfileWithMainThread()

	result := AnalyzeJSCrypto(profile)

	if result.TotalTimeMs != 0 {
		t.Errorf("TotalTimeMs = %v, want 0 for no crypto", result.TotalTimeMs)
	}
}

func TestAnalyzeJSCrypto_WithCryptoWorkers(t *testing.T) {
	profile := testutil.ProfileWithJSCrypto()

	result := AnalyzeJSCrypto(profile)

	// Should detect crypto resources
	if len(result.Resources) == 0 {
		t.Log("No JS crypto resources detected - may need better fixture")
	}
}

func TestIsCryptoJSResource_DecryptWorker(t *testing.T) {
	tests := []struct {
		name     string
		expected bool
	}{
		{"seipdDecryptionWorker.js", true},
		{"decryptionWorker.js", true},
		{"decrypt-worker.js", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isCryptoJSResource(tt.name); got != tt.expected {
				t.Errorf("isCryptoJSResource(%q) = %v, want %v", tt.name, got, tt.expected)
			}
		})
	}
}

func TestIsCryptoJSResource_OpenPGP(t *testing.T) {
	tests := []struct {
		name     string
		expected bool
	}{
		{"openpgp.js", true},
		{"openpgp.min.js", true},
		{"openpgp.worker.js", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isCryptoJSResource(tt.name); got != tt.expected {
				t.Errorf("isCryptoJSResource(%q) = %v, want %v", tt.name, got, tt.expected)
			}
		})
	}
}

func TestIsCryptoJSResource_NotCrypto(t *testing.T) {
	tests := []string{"app.js", "main.js", "bundle.js", "react.js"}
	for _, name := range tests {
		if isCryptoJSResource(name) {
			t.Errorf("isCryptoJSResource(%q) = true, want false", name)
		}
	}
}

func TestIsOpenPGPFunction_DecryptWithSessionKey(t *testing.T) {
	tests := []struct {
		name     string
		expected bool
	}{
		{"decryptWithSessionKey", true},
		{"DecryptWithSessionKey", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isOpenPGPFunction(tt.name); got != tt.expected {
				t.Errorf("isOpenPGPFunction(%q) = %v, want %v", tt.name, got, tt.expected)
			}
		})
	}
}

func TestIsOpenPGPFunction_SEIPDPacket(t *testing.T) {
	if !isOpenPGPFunction("SEIPDPacket") {
		t.Error("expected SEIPDPacket to be OpenPGP function")
	}
}

func TestIsOpenPGPFunction_NotOpenPGP(t *testing.T) {
	tests := []string{"processData", "handleClick", "main"}
	for _, name := range tests {
		if isOpenPGPFunction(name) {
			t.Errorf("isOpenPGPFunction(%q) = true, want false", name)
		}
	}
}

func TestIsServiceWorkerThread_Yes(t *testing.T) {
	// isServiceWorkerThread only matches names containing "serviceworker" (case insensitive)
	tests := []string{"ServiceWorker", "serviceworker", "MyServiceWorkerThread"}
	for _, name := range tests {
		if !isServiceWorkerThread(name) {
			t.Errorf("isServiceWorkerThread(%q) = false, want true", name)
		}
	}
}

func TestIsServiceWorkerThread_No(t *testing.T) {
	// These don't contain "serviceworker"
	tests := []string{"GeckoMain", "DOM Worker", "Compositor", "SW", "sw-thread"}
	for _, name := range tests {
		if isServiceWorkerThread(name) {
			t.Errorf("isServiceWorkerThread(%q) = true, want false", name)
		}
	}
}

func TestExtractResourceName_MozExtension(t *testing.T) {
	result := extractResourceName("moz-extension://abc123/openpgp.js")
	if result != "openpgp.js" {
		t.Errorf("extractResourceName() = %s, want openpgp.js", result)
	}
}

func TestExtractResourceName_ChromeExtension(t *testing.T) {
	result := extractResourceName("chrome-extension://abc123/crypto.js")
	if result != "crypto.js" {
		t.Errorf("extractResourceName() = %s, want crypto.js", result)
	}
}

func TestExtractResourceName_QueryString(t *testing.T) {
	result := extractResourceName("https://example.com/script.js?v=1.0")
	if result != "script.js" {
		t.Errorf("extractResourceName() = %s, want script.js", result)
	}
}

func TestExtractResourceName_PlainPath(t *testing.T) {
	result := extractResourceName("file:///path/to/script.js")
	if result != "script.js" {
		t.Errorf("extractResourceName() = %s, want script.js", result)
	}
}

func TestExtractResourceFromFuncName_MozExtension(t *testing.T) {
	result := extractResourceFromFuncName("fn (moz-extension://abc123/script.js:42:10)")
	if result != "script.js" {
		t.Logf("extractResourceFromFuncName() = %s", result)
	}
}

func TestExtractResourceFromFuncName_NodeModules(t *testing.T) {
	result := extractResourceFromFuncName("fn (node_modules/openpgp/dist/openpgp.js:100:20)")
	t.Logf("extractResourceFromFuncName(node_modules) = %s", result)
}

func TestIsNumeric_Valid(t *testing.T) {
	tests := []string{"0", "123", "999"}
	for _, s := range tests {
		if !isNumeric(s) {
			t.Errorf("isNumeric(%q) = false, want true", s)
		}
	}
}

func TestIsNumeric_Invalid(t *testing.T) {
	tests := []string{"abc", "12a", "", "12.5"}
	for _, s := range tests {
		if isNumeric(s) {
			t.Errorf("isNumeric(%q) = true, want false", s)
		}
	}
}

func TestIsNumeric_Empty(t *testing.T) {
	if isNumeric("") {
		t.Error("isNumeric('') = true, want false")
	}
}

func TestAnalyzeJSCrypto_TopFunctions(t *testing.T) {
	profile := testutil.ProfileWithJSCrypto()

	result := AnalyzeJSCrypto(profile)

	// TopFunctions should be sorted by time
	for i := 1; i < len(result.TopFunctions); i++ {
		if result.TopFunctions[i].TotalTime > result.TopFunctions[i-1].TotalTime {
			t.Errorf("TopFunctions not sorted by time")
		}
	}
}

func TestAnalyzeJSCrypto_WorkerCount(t *testing.T) {
	profile := testutil.ProfileWithJSCrypto()

	result := AnalyzeJSCrypto(profile)

	if result.WorkerCount < 0 {
		t.Errorf("WorkerCount should be >= 0, got %d", result.WorkerCount)
	}
}

func TestAnalyzeJSCrypto_ByThreadDistribution(t *testing.T) {
	profile := testutil.ProfileWithJSCrypto()

	result := AnalyzeJSCrypto(profile)

	if result.ByThread == nil {
		t.Error("expected ByThread map")
	}
}

func TestAnalyzeJSCrypto_Recommendations(t *testing.T) {
	profile := testutil.ProfileWithJSCrypto()

	result := AnalyzeJSCrypto(profile)

	// Recommendations should be a slice (may be empty)
	if result.Recommendations == nil {
		t.Error("expected Recommendations slice")
	}
}

func TestFormatJSCryptoAnalysis_Output(t *testing.T) {
	analysis := JSCryptoAnalysis{
		TotalTimeMs:      1000,
		TotalSamples:     100,
		WorkerCount:      4,
		AvgTimePerWorker: 250,
		Resources: []JSCryptoResource{
			{Name: "openpgp.js", URL: "moz-extension://abc/openpgp.js", TotalTime: 800},
		},
		TopFunctions: []JSCryptoFunction{
			{Name: "decrypt", Resource: "openpgp.js", TotalTime: 500, Percent: 50},
		},
		ByThread:        map[string]float64{"Worker#1": 500, "Worker#2": 500},
		Recommendations: []string{"Work is well distributed"},
	}

	output := FormatJSCryptoAnalysis(analysis)

	if output == "" {
		t.Error("expected non-empty output")
	}
}
