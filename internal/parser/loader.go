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

// LoadProfile loads a Firefox Profiler JSON file (supports gzip compression)
func LoadProfile(path string) (*Profile, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open profile: %w", err)
	}
	defer func() { _ = file.Close() }()

	var reader io.Reader = file

	// Check if gzip compressed
	ext := strings.ToLower(filepath.Ext(path))
	if ext == ".gz" || ext == ".gzip" {
		gzReader, err := gzip.NewReader(file)
		if err != nil {
			return nil, fmt.Errorf("failed to create gzip reader: %w", err)
		}
		defer func() { _ = gzReader.Close() }()
		reader = gzReader
	}

	var profile Profile
	decoder := json.NewDecoder(reader)
	if err := decoder.Decode(&profile); err != nil {
		return nil, fmt.Errorf("failed to decode profile JSON: %w", err)
	}

	return &profile, nil
}

// LoadProfileFromReader loads a profile from an io.Reader
func LoadProfileFromReader(reader io.Reader) (*Profile, error) {
	var profile Profile
	decoder := json.NewDecoder(reader)
	if err := decoder.Decode(&profile); err != nil {
		return nil, fmt.Errorf("failed to decode profile JSON: %w", err)
	}
	return &profile, nil
}

// LoadProfileAuto loads a profile with automatic browser detection
func LoadProfileAuto(path string) (*Profile, BrowserType, error) {
	browserType, err := DetectBrowserType(path)
	if err != nil {
		return nil, BrowserUnknown, fmt.Errorf("failed to detect browser type: %w", err)
	}

	return LoadProfileWithType(path, browserType)
}

// LoadProfileWithType loads a profile with explicit browser type
func LoadProfileWithType(path string, browserType BrowserType) (*Profile, BrowserType, error) {
	switch browserType {
	case BrowserFirefox:
		profile, err := LoadProfile(path)
		return profile, BrowserFirefox, err

	case BrowserChrome:
		chromeProfile, err := LoadChromeProfile(path)
		if err != nil {
			return nil, BrowserChrome, err
		}
		profile, err := ConvertChromeToProfile(chromeProfile)
		return profile, BrowserChrome, err

	default:
		// Auto-detect: try to detect browser type first
		detectedType, detectErr := DetectBrowserType(path)
		if detectErr == nil && detectedType != BrowserUnknown {
			return LoadProfileWithType(path, detectedType)
		}

		// Fallback: try Firefox first, then Chrome
		profile, err := LoadProfile(path)
		if err == nil {
			return profile, BrowserFirefox, nil
		}

		chromeProfile, chromeErr := LoadChromeProfile(path)
		if chromeErr == nil {
			profile, err := ConvertChromeToProfile(chromeProfile)
			return profile, BrowserChrome, err
		}

		return nil, BrowserUnknown, fmt.Errorf("failed to parse as Firefox or Chrome profile")
	}
}
