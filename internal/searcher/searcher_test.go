package searcher

import (
	"os"
	"path/filepath"
	"testing"

	"memento/internal/indexer"
)

func setupTestDB(t *testing.T) string {
	t.Helper()
	wikiDir := filepath.Join("..", "..", "testdata", "wiki")
	dbPath := filepath.Join(t.TempDir(), "test.db")
	if err := indexer.Index(wikiDir, dbPath); err != nil {
		t.Fatalf("failed to index test wiki: %v", err)
	}
	return dbPath
}

func TestOpenNonExistentDB(t *testing.T) {
	_, err := Open("/nonexistent/path/to/db.db")
	if err == nil {
		t.Error("expected error for nonexistent database")
	}
}

func TestSearchResults(t *testing.T) {
	dbPath := setupTestDB(t)

	s, err := Open(dbPath)
	if err != nil {
		t.Fatalf("Open failed: %v", err)
	}
	defer s.Close()

	results, err := s.Search("runner", 10)
	if err != nil {
		t.Fatalf("Search failed: %v", err)
	}

	if len(results) == 0 {
		t.Error("expected results for 'runner' query")
	}

	found := false
	for _, r := range results {
		if r.Path == "architecture/runner.md" {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected runner.md in results")
	}
}

func TestSearchNoResults(t *testing.T) {
	dbPath := setupTestDB(t)

	s, err := Open(dbPath)
	if err != nil {
		t.Fatalf("Open failed: %v", err)
	}
	defer s.Close()

	results, err := s.Search("xyznonexistentterm", 10)
	if err != nil {
		t.Fatalf("Search failed: %v", err)
	}

	if len(results) != 0 {
		t.Errorf("expected 0 results, got %d", len(results))
	}
}

func TestSearchHeadingMatch(t *testing.T) {
	dbPath := setupTestDB(t)

	s, err := Open(dbPath)
	if err != nil {
		t.Fatalf("Open failed: %v", err)
	}
	defer s.Close()

	results, err := s.Search("worker pool", 10)
	if err != nil {
		t.Fatalf("Search failed: %v", err)
	}

	if len(results) == 0 {
		t.Fatal("expected results")
	}

	if results[0].Path != "architecture/runner.md" {
		t.Errorf("expected architecture/runner.md, got %s", results[0].Path)
	}
}

func TestSearchJSON(t *testing.T) {
	os.Args = append(os.Args, "test")

	dbPath := setupTestDB(t)

	s, err := Open(dbPath)
	if err != nil {
		t.Fatalf("Open failed: %v", err)
	}
	defer s.Close()

	// Just verify it doesn't error
	err = s.SearchJSON("runner", 10)
	if err != nil {
		t.Fatalf("SearchJSON failed: %v", err)
	}
}

func TestEscapeFTS5Query(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"simple query", "simple query"},
		{"query with * star", "query with   star"},
		{"hello (world)", "hello  world"},
	}

	for _, tc := range tests {
		got := escapeFTS5Query(tc.input)
		if got != tc.expected {
			t.Errorf("escapeFTS5Query(%q) = %q, want %q", tc.input, got, tc.expected)
		}
	}
}
