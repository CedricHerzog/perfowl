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
	defer file.Close()

	var reader io.Reader = file

	// Check if gzip compressed
	ext := strings.ToLower(filepath.Ext(path))
	if ext == ".gz" || ext == ".gzip" {
		gzReader, err := gzip.NewReader(file)
		if err != nil {
			return nil, fmt.Errorf("failed to create gzip reader: %w", err)
		}
		defer gzReader.Close()
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
