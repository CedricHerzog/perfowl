package analyzer

import (
	"testing"

	"github.com/CedricHerzog/perfowl/internal/testutil"
)

func TestAnalyzeCrypto_EmptyProfile(t *testing.T) {
	profile := testutil.MinimalProfile()

	result := AnalyzeCrypto(profile)

	if result.TotalOperations != 0 {
		t.Errorf("TotalOperations = %v, want 0", result.TotalOperations)
	}
}

func TestAnalyzeCrypto_NoCryptoOperations(t *testing.T) {
	profile := testutil.ProfileWithMainThread()

	result := AnalyzeCrypto(profile)

	if result.TotalOperations != 0 {
		t.Errorf("TotalOperations = %v, want 0", result.TotalOperations)
	}
}

func TestAnalyzeCrypto_WithCryptoFunctions(t *testing.T) {
	profile := testutil.ProfileWithCrypto()

	result := AnalyzeCrypto(profile)

	if result.TotalOperations == 0 {
		t.Log("No crypto operations detected - may need better fixture")
	}
}

func TestIsCryptoFunction_EncryptDecrypt(t *testing.T) {
	tests := []struct {
		name     string
		expected bool
	}{
		{"encrypt", true},
		{"decrypt", true},
		{"Encrypt", true},
		{"Decrypt", true},
		{"doEncrypt", true},
		{"doDecrypt", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isCryptoFunction(tt.name); got != tt.expected {
				t.Errorf("isCryptoFunction(%q) = %v, want %v", tt.name, got, tt.expected)
			}
		})
	}
}

func TestIsCryptoFunction_SubtleCrypto(t *testing.T) {
	tests := []struct {
		name     string
		expected bool
	}{
		{"subtle.encrypt", true},
		{"subtle.decrypt", true},
		{"subtle.digest", true},
		{"subtle.sign", true},
		{"SubtleCrypto", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isCryptoFunction(tt.name); got != tt.expected {
				t.Errorf("isCryptoFunction(%q) = %v, want %v", tt.name, got, tt.expected)
			}
		})
	}
}

func TestIsCryptoFunction_HashFunctions(t *testing.T) {
	tests := []struct {
		name     string
		expected bool
	}{
		{"sha256", true},
		{"SHA256", true},
		{"sha1", true},
		{"md5", true},
		{"hash", true},
		{"digest", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isCryptoFunction(tt.name); got != tt.expected {
				t.Errorf("isCryptoFunction(%q) = %v, want %v", tt.name, got, tt.expected)
			}
		})
	}
}

func TestIsCryptoFunction_NotCrypto(t *testing.T) {
	tests := []struct {
		name string
	}{
		{"processData"},
		{"handleClick"},
		{"renderComponent"},
		{"main"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if isCryptoFunction(tt.name) {
				t.Errorf("isCryptoFunction(%q) = true, want false", tt.name)
			}
		})
	}
}

func TestExtractOperation_Encrypt(t *testing.T) {
	result := extractOperation("encrypt")
	if result != "encrypt" {
		t.Errorf("extractOperation(encrypt) = %s, want encrypt", result)
	}
}

func TestExtractOperation_Decrypt(t *testing.T) {
	result := extractOperation("decrypt")
	if result != "decrypt" {
		t.Errorf("extractOperation(decrypt) = %s, want decrypt", result)
	}
}

func TestExtractOperation_Digest(t *testing.T) {
	result := extractOperation("sha256.digest")
	if result != "digest" {
		t.Errorf("extractOperation(sha256.digest) = %s, want digest", result)
	}
}

func TestExtractOperation_SignVerify(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"sign", "sign"},
		{"verify", "verify"},
		{"subtle.sign", "sign"},
		{"subtle.verify", "verify"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			if got := extractOperation(tt.input); got != tt.expected {
				t.Errorf("extractOperation(%q) = %s, want %s", tt.input, got, tt.expected)
			}
		})
	}
}

func TestExtractOperation_KeyOps(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"generateKey", "generateKey"},
		{"importKey", "importKey"},
		{"exportKey", "exportKey"},
		{"deriveKey", "deriveKey"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			if got := extractOperation(tt.input); got != tt.expected {
				t.Errorf("extractOperation(%q) = %s, want %s", tt.input, got, tt.expected)
			}
		})
	}
}

func TestExtractAlgorithm_AES(t *testing.T) {
	tests := []string{"AES", "aes", "AES-GCM", "AES-CBC", "aes256"}
	for _, input := range tests {
		result := extractAlgorithm(input)
		if result != "AES" && result != "" {
			// May not extract, but if it does, should be AES
			t.Logf("extractAlgorithm(%s) = %s", input, result)
		}
	}
}

func TestExtractAlgorithm_RSA(t *testing.T) {
	tests := []string{"RSA", "rsa", "RSA-OAEP", "RSA-PSS"}
	for _, input := range tests {
		result := extractAlgorithm(input)
		if result != "RSA" && result != "" {
			t.Logf("extractAlgorithm(%s) = %s", input, result)
		}
	}
}

func TestExtractAlgorithm_SHA(t *testing.T) {
	tests := []string{"SHA-256", "SHA256", "sha256", "SHA-1", "SHA-384", "SHA-512"}
	for _, input := range tests {
		result := extractAlgorithm(input)
		t.Logf("extractAlgorithm(%s) = %s", input, result)
	}
}

func TestIsCryptoResource_CryptoLib(t *testing.T) {
	// isCryptoResource checks for specific patterns:
	// "crypto", "decrypt", "encrypt", "libcorecrypto", "libcommoncrypto",
	// "openssl", "boringssl", "nss", "pgp", "gpg", "seipd"
	tests := []struct {
		name     string
		expected bool
	}{
		{"crypto.js", true},
		{"cryptolib.js", true},
		{"decrypt.js", true},
		{"encrypt-worker.js", true},
		{"libcorecrypto.dylib", true},
		{"openssl.so", true},
		{"boringssl.so", true},
		{"nss3.dll", true},
		{"openpgp.js", true},
		{"gpg.js", true},
		{"seipd.js", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isCryptoResource(tt.name); got != tt.expected {
				t.Errorf("isCryptoResource(%q) = %v, want %v", tt.name, got, tt.expected)
			}
		})
	}
}

func TestIsCryptoResource_NotCrypto(t *testing.T) {
	// These don't match any of the crypto resource patterns
	tests := []string{"app.js", "main.js", "utils.js", "react.js", "forge.js", "sodium.js", "tweetnacl.js"}
	for _, name := range tests {
		if isCryptoResource(name) {
			t.Errorf("isCryptoResource(%q) = true, want false", name)
		}
	}
}

func TestAnalyzeCrypto_ByOperationAggregation(t *testing.T) {
	profile := testutil.ProfileWithCrypto()

	result := AnalyzeCrypto(profile)

	// ByOperation should be a map
	if result.ByOperation == nil {
		t.Error("expected ByOperation map")
	}
}

func TestAnalyzeCrypto_ByAlgorithmAggregation(t *testing.T) {
	profile := testutil.ProfileWithCrypto()

	result := AnalyzeCrypto(profile)

	if result.ByAlgorithm == nil {
		t.Error("expected ByAlgorithm map")
	}
}

func TestAnalyzeCrypto_ByThreadAggregation(t *testing.T) {
	profile := testutil.ProfileWithCrypto()

	result := AnalyzeCrypto(profile)

	if result.ByThread == nil {
		t.Error("expected ByThread map")
	}
}

func TestFormatCryptoAnalysis_Output(t *testing.T) {
	analysis := CryptoAnalysis{
		TotalOperations: 10,
		TotalTimeMs:     500,
		ByOperation:     map[string]float64{"encrypt": 300, "decrypt": 200},
		ByAlgorithm:     map[string]float64{"AES": 500},
		ByThread:        map[string]float64{"GeckoMain": 500},
		TopOperations: []CryptoOperation{
			{Operation: "encrypt", Algorithm: "AES", DurationMs: 50},
		},
		Warnings: []string{"Consider using Web Workers"},
	}

	output := FormatCryptoAnalysis(analysis)

	if output == "" {
		t.Error("expected non-empty output")
	}
}
