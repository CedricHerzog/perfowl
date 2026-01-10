package mcp

import (
	"strings"
	"testing"

	"github.com/CedricHerzog/perfowl/internal/format/toon"
	"github.com/CedricHerzog/perfowl/internal/parser"
)

func TestTOONOutputFormat(t *testing.T) {
	// Load a profile
	profile, _, err := parser.LoadProfileAuto("../../profiles/firefox/new-4-core.json.gz")
	if err != nil {
		t.Skipf("Skipping test: could not load profile: %v", err)
	}

	// Build summary using the same function as the MCP handler
	summary := buildSummary(profile)

	// Encode to TOON
	output, err := toon.Encode(summary)
	if err != nil {
		t.Fatalf("Failed to encode summary: %v", err)
	}

	// Verify it's TOON format, not JSON
	if strings.HasPrefix(strings.TrimSpace(output), "{") {
		t.Errorf("Output looks like JSON, expected TOON format:\n%s", output)
	}

	// Verify expected TOON fields are present
	if !strings.Contains(output, "duration_seconds:") {
		t.Errorf("Missing 'duration_seconds:' field in output:\n%s", output)
	}
	if !strings.Contains(output, "platform:") {
		t.Errorf("Missing 'platform:' field in output:\n%s", output)
	}
	if !strings.Contains(output, "features[") {
		t.Errorf("Missing 'features[' array in output:\n%s", output)
	}

	t.Logf("TOON Output:\n%s", output)
}

func TestTOONTabularFormat(t *testing.T) {
	// Test that arrays of structs use tabular format
	type Item struct {
		Name  string `json:"name"`
		Value int    `json:"value"`
	}

	type Report struct {
		Items []Item `json:"items"`
	}

	report := Report{
		Items: []Item{
			{Name: "first", Value: 1},
			{Name: "second", Value: 2},
		},
	}

	output, err := toon.Encode(report)
	if err != nil {
		t.Fatalf("Failed to encode report: %v", err)
	}

	// Should use tabular format
	if !strings.Contains(output, "items[2]{name,value}:") {
		t.Errorf("Expected tabular header 'items[2]{name,value}:', got:\n%s", output)
	}

	t.Logf("TOON Tabular Output:\n%s", output)
}
