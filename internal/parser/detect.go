package parser

import (
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// BrowserType represents the browser that generated a profile
type BrowserType string

const (
	BrowserFirefox BrowserType = "firefox"
	BrowserChrome  BrowserType = "chrome"
	BrowserUnknown BrowserType = "unknown"
)

// ParseBrowserType parses a browser type string
func ParseBrowserType(s string) BrowserType {
	switch strings.ToLower(s) {
	case "firefox":
		return BrowserFirefox
	case "chrome":
		return BrowserChrome
	case "auto", "":
		return BrowserUnknown
	default:
		return BrowserUnknown
	}
}

// profilePeek holds minimal data to detect browser type
type profilePeek struct {
	// Firefox fields
	Meta *struct {
		Product string `json:"product"`
	} `json:"meta"`
	Threads []json.RawMessage `json:"threads"`

	// Chrome fields
	TraceEvents []json.RawMessage `json:"traceEvents"`
}

// DetectBrowserType determines if a profile is Firefox or Chrome
func DetectBrowserType(path string) (BrowserType, error) {
	file, err := os.Open(path)
	if err != nil {
		return BrowserUnknown, fmt.Errorf("failed to open file: %w", err)
	}
	defer func() { _ = file.Close() }()

	var reader io.Reader = file

	// Handle gzip compression
	ext := strings.ToLower(filepath.Ext(path))
	if ext == ".gz" || ext == ".gzip" {
		gzReader, err := gzip.NewReader(file)
		if err != nil {
			return BrowserUnknown, fmt.Errorf("failed to create gzip reader: %w", err)
		}
		defer func() { _ = gzReader.Close() }()
		reader = gzReader
	}

	var peek profilePeek
	decoder := json.NewDecoder(reader)
	if err := decoder.Decode(&peek); err != nil {
		return BrowserUnknown, fmt.Errorf("failed to decode JSON: %w", err)
	}

	return detectFromPeek(&peek), nil
}

func detectFromPeek(peek *profilePeek) BrowserType {
	// Firefox: has meta.product and threads array
	if peek.Meta != nil && len(peek.Threads) > 0 {
		return BrowserFirefox
	}

	// Chrome: has traceEvents array
	if len(peek.TraceEvents) > 0 {
		return BrowserChrome
	}

	return BrowserUnknown
}
