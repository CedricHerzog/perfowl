package parser

import (
	"testing"
)

func TestProfile_Duration(t *testing.T) {
	tests := []struct {
		name     string
		start    float64
		end      float64
		expected float64
	}{
		{"positive duration", 0, 1000, 1000},
		{"zero duration", 100, 100, 0},
		{"large duration", 0, 60000, 60000},
		{"non-zero start", 500, 1500, 1000},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &Profile{
				Meta: Meta{
					ProfilingStartTime: tt.start,
					ProfilingEndTime:   tt.end,
				},
			}
			if got := p.Duration(); got != tt.expected {
				t.Errorf("Duration() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestProfile_DurationSeconds(t *testing.T) {
	tests := []struct {
		name     string
		start    float64
		end      float64
		expected float64
	}{
		{"one second", 0, 1000, 1.0},
		{"half second", 0, 500, 0.5},
		{"zero", 0, 0, 0},
		{"ten seconds", 0, 10000, 10.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &Profile{
				Meta: Meta{
					ProfilingStartTime: tt.start,
					ProfilingEndTime:   tt.end,
				},
			}
			if got := p.DurationSeconds(); got != tt.expected {
				t.Errorf("DurationSeconds() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestProfile_ExtensionCount(t *testing.T) {
	tests := []struct {
		name     string
		length   int
		expected int
	}{
		{"no extensions", 0, 0},
		{"one extension", 1, 1},
		{"multiple extensions", 5, 5},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &Profile{
				Meta: Meta{
					Extensions: Extensions{
						Length: tt.length,
					},
				},
			}
			if got := p.ExtensionCount(); got != tt.expected {
				t.Errorf("ExtensionCount() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestProfile_GetExtensions(t *testing.T) {
	t.Run("no extensions", func(t *testing.T) {
		p := &Profile{
			Meta: Meta{
				Extensions: Extensions{Length: 0},
			},
		}
		exts := p.GetExtensions()
		if len(exts) != 0 {
			t.Errorf("expected empty map, got %v", exts)
		}
	})

	t.Run("single extension", func(t *testing.T) {
		p := &Profile{
			Meta: Meta{
				Extensions: Extensions{
					Length:  1,
					ID:      []string{"ext1@example.com"},
					Name:    []string{"Test Extension"},
					BaseURL: []string{"moz-extension://abc123/"},
				},
			},
		}
		exts := p.GetExtensions()
		if len(exts) != 1 {
			t.Errorf("expected 1 extension, got %d", len(exts))
		}
		if exts["ext1@example.com"] != "Test Extension" {
			t.Errorf("unexpected extension name: %v", exts["ext1@example.com"])
		}
	})

	t.Run("multiple extensions", func(t *testing.T) {
		p := &Profile{
			Meta: Meta{
				Extensions: Extensions{
					Length:  2,
					ID:      []string{"ext1@example.com", "ext2@example.com"},
					Name:    []string{"Extension 1", "Extension 2"},
					BaseURL: []string{"moz-extension://abc/", "moz-extension://def/"},
				},
			},
		}
		exts := p.GetExtensions()
		if len(exts) != 2 {
			t.Errorf("expected 2 extensions, got %d", len(exts))
		}
		if exts["ext1@example.com"] != "Extension 1" {
			t.Errorf("unexpected extension 1 name: %v", exts["ext1@example.com"])
		}
		if exts["ext2@example.com"] != "Extension 2" {
			t.Errorf("unexpected extension 2 name: %v", exts["ext2@example.com"])
		}
	})

	t.Run("mismatched lengths", func(t *testing.T) {
		p := &Profile{
			Meta: Meta{
				Extensions: Extensions{
					Length:  3,
					ID:      []string{"ext1@example.com", "ext2@example.com"},
					Name:    []string{"Extension 1"},
					BaseURL: []string{},
				},
			},
		}
		// Should only return extensions where both ID and Name exist
		exts := p.GetExtensions()
		if len(exts) != 1 {
			t.Errorf("expected 1 extension (due to mismatched lengths), got %d", len(exts))
		}
	})
}

func TestProfile_GetExtensionBaseURLs(t *testing.T) {
	t.Run("no extensions", func(t *testing.T) {
		p := &Profile{
			Meta: Meta{
				Extensions: Extensions{Length: 0},
			},
		}
		urls := p.GetExtensionBaseURLs()
		if len(urls) != 0 {
			t.Errorf("expected empty map, got %v", urls)
		}
	})

	t.Run("with extensions", func(t *testing.T) {
		p := &Profile{
			Meta: Meta{
				Extensions: Extensions{
					Length:  2,
					ID:      []string{"ext1@example.com", "ext2@example.com"},
					Name:    []string{"Extension 1", "Extension 2"},
					BaseURL: []string{"moz-extension://abc/", "moz-extension://def/"},
				},
			},
		}
		urls := p.GetExtensionBaseURLs()
		if len(urls) != 2 {
			t.Errorf("expected 2 URLs, got %d", len(urls))
		}
		if urls["ext1@example.com"] != "moz-extension://abc/" {
			t.Errorf("unexpected URL for ext1: %v", urls["ext1@example.com"])
		}
		if urls["ext2@example.com"] != "moz-extension://def/" {
			t.Errorf("unexpected URL for ext2: %v", urls["ext2@example.com"])
		}
	})

	t.Run("mismatched lengths", func(t *testing.T) {
		p := &Profile{
			Meta: Meta{
				Extensions: Extensions{
					Length:  3,
					ID:      []string{"ext1@example.com", "ext2@example.com"},
					Name:    []string{"Extension 1", "Extension 2"},
					BaseURL: []string{"moz-extension://abc/"},
				},
			},
		}
		// Should only return URLs where both ID and BaseURL exist
		urls := p.GetExtensionBaseURLs()
		if len(urls) != 1 {
			t.Errorf("expected 1 URL (due to mismatched lengths), got %d", len(urls))
		}
	})
}

func TestProfile_GetCategoryByIndex(t *testing.T) {
	categories := []Category{
		{Name: "Idle", Color: "transparent"},
		{Name: "JavaScript", Color: "yellow"},
		{Name: "Layout", Color: "purple"},
	}

	p := &Profile{
		Meta: Meta{
			Categories: categories,
		},
	}

	t.Run("valid index 0", func(t *testing.T) {
		cat := p.GetCategoryByIndex(0)
		if cat == nil {
			t.Fatal("expected category, got nil")
		}
		if cat.Name != "Idle" {
			t.Errorf("expected Idle, got %s", cat.Name)
		}
	})

	t.Run("valid index 1", func(t *testing.T) {
		cat := p.GetCategoryByIndex(1)
		if cat == nil {
			t.Fatal("expected category, got nil")
		}
		if cat.Name != "JavaScript" {
			t.Errorf("expected JavaScript, got %s", cat.Name)
		}
	})

	t.Run("valid index 2", func(t *testing.T) {
		cat := p.GetCategoryByIndex(2)
		if cat == nil {
			t.Fatal("expected category, got nil")
		}
		if cat.Name != "Layout" {
			t.Errorf("expected Layout, got %s", cat.Name)
		}
	})

	t.Run("negative index", func(t *testing.T) {
		cat := p.GetCategoryByIndex(-1)
		if cat != nil {
			t.Errorf("expected nil for negative index, got %v", cat)
		}
	})

	t.Run("index out of bounds", func(t *testing.T) {
		cat := p.GetCategoryByIndex(3)
		if cat != nil {
			t.Errorf("expected nil for out of bounds index, got %v", cat)
		}
	})

	t.Run("large index", func(t *testing.T) {
		cat := p.GetCategoryByIndex(1000)
		if cat != nil {
			t.Errorf("expected nil for large index, got %v", cat)
		}
	})
}

func TestProfile_ThreadCount(t *testing.T) {
	tests := []struct {
		name     string
		threads  []Thread
		expected int
	}{
		{"no threads", nil, 0},
		{"empty slice", []Thread{}, 0},
		{"one thread", []Thread{{Name: "Main"}}, 1},
		{"multiple threads", []Thread{{Name: "Main"}, {Name: "Worker1"}, {Name: "Worker2"}}, 3},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &Profile{Threads: tt.threads}
			if got := p.ThreadCount(); got != tt.expected {
				t.Errorf("ThreadCount() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestProfile_GetMainThreads(t *testing.T) {
	t.Run("no threads", func(t *testing.T) {
		p := &Profile{Threads: nil}
		mains := p.GetMainThreads()
		if len(mains) != 0 {
			t.Errorf("expected no main threads, got %d", len(mains))
		}
	})

	t.Run("no main threads", func(t *testing.T) {
		p := &Profile{
			Threads: []Thread{
				{Name: "Worker1", IsMainThread: false},
				{Name: "Worker2", IsMainThread: false},
			},
		}
		mains := p.GetMainThreads()
		if len(mains) != 0 {
			t.Errorf("expected no main threads, got %d", len(mains))
		}
	})

	t.Run("single main thread", func(t *testing.T) {
		p := &Profile{
			Threads: []Thread{
				{Name: "GeckoMain", IsMainThread: true},
				{Name: "Worker1", IsMainThread: false},
			},
		}
		mains := p.GetMainThreads()
		if len(mains) != 1 {
			t.Errorf("expected 1 main thread, got %d", len(mains))
		}
		if mains[0].Name != "GeckoMain" {
			t.Errorf("expected GeckoMain, got %s", mains[0].Name)
		}
	})

	t.Run("multiple main threads", func(t *testing.T) {
		p := &Profile{
			Threads: []Thread{
				{Name: "GeckoMain", IsMainThread: true},
				{Name: "ContentMain", IsMainThread: true},
				{Name: "Worker1", IsMainThread: false},
			},
		}
		mains := p.GetMainThreads()
		if len(mains) != 2 {
			t.Errorf("expected 2 main threads, got %d", len(mains))
		}
	})

	t.Run("all main threads", func(t *testing.T) {
		p := &Profile{
			Threads: []Thread{
				{Name: "Main1", IsMainThread: true},
				{Name: "Main2", IsMainThread: true},
			},
		}
		mains := p.GetMainThreads()
		if len(mains) != 2 {
			t.Errorf("expected 2 main threads, got %d", len(mains))
		}
	})
}
