package indexer

import (
	"os"
	"path/filepath"
	"testing"
)

func TestBuildFTSContent(t *testing.T) {
	tests := []struct {
		heading  string
		body     string
		expected string
	}{
		{
			heading:  "Getting Started",
			body:     "This is the body.",
			expected: "Getting Started\nGetting Started\nGetting Started\nThis is the body.",
		},
		{
			heading:  "",
			body:     "Just body.",
			expected: "Just body.",
		},
	}

	for _, tc := range tests {
		got := buildFTSContent(tc.heading, tc.body)
		if got != tc.expected {
			t.Errorf("buildFTSContent(%q, %q) = %q, want %q", tc.heading, tc.body, got, tc.expected)
		}
	}
}

func TestIndex(t *testing.T) {
	wikiDir := filepath.Join("..", "..", "testdata", "wiki")
	dbPath := filepath.Join(t.TempDir(), "test.db")

	err := Index(wikiDir, dbPath)
	if err != nil {
		t.Fatalf("Index failed: %v", err)
	}

	if _, err := os.Stat(dbPath); err != nil {
		t.Errorf("database file not created: %v", err)
	}
}

func TestIndexNonExistentDir(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "nonexistent.db")
	err := Index("/nonexistent/path/to/wiki", dbPath)
	if err == nil {
		t.Error("expected error for nonexistent directory")
	}
}
