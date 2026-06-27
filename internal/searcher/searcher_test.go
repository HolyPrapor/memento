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

func setupLinkedDB(t *testing.T) string {
	t.Helper()
	wikiDir := t.TempDir()

	writeFile := func(name, content string) {
		path := filepath.Join(wikiDir, name)
		dir := filepath.Dir(path)
		if err := os.MkdirAll(dir, 0755); err != nil {
			t.Fatalf("mkdir: %v", err)
		}
		if err := os.WriteFile(path, []byte(content), 0644); err != nil {
			t.Fatalf("write file: %v", err)
		}
	}

	writeFile("index.md", `# Index

See [runner design](arch/runner.md#design-rationale) and [storage](arch/storage.md).
`)

	writeFile("arch/runner.md", `# Runner

## Design Rationale

Chosen because [storage v2](storage.md#why-sqlite).
`)

	writeFile("arch/storage.md", `# Storage

Some preamble.

## Why SQLite

Because SQLite.
`)

	dbPath := filepath.Join(t.TempDir(), "test-links.db")
	if err := indexer.Index(wikiDir, dbPath); err != nil {
		t.Fatalf("Index failed: %v", err)
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

	if results[0].Relevance <= 0 {
		t.Errorf("expected positive relevance, got %.2f", results[0].Relevance)
	}
}

func TestSearchJSON(t *testing.T) {
	dbPath := setupTestDB(t)

	s, err := Open(dbPath)
	if err != nil {
		t.Fatalf("Open failed: %v", err)
	}
	defer s.Close()

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

func TestSearchBacklinks(t *testing.T) {
	dbPath := setupLinkedDB(t)

	s, err := Open(dbPath)
	if err != nil {
		t.Fatalf("Open failed: %v", err)
	}
	defer s.Close()

	results, err := s.Search("SQLite", 10)
	if err != nil {
		t.Fatalf("Search failed: %v", err)
	}

	if len(results) == 0 {
		t.Fatal("expected results")
	}

	var sqliteResult *Result
	for i := range results {
		if results[i].Path == "arch/storage.md" && results[i].Heading == "Why SQLite" {
			sqliteResult = &results[i]
			break
		}
	}
	if sqliteResult == nil {
		t.Fatal("expected arch/storage.md Why SQLite in results")
	}

	if len(sqliteResult.Backlinks) < 1 {
		t.Fatalf("expected backlinks, got %d", len(sqliteResult.Backlinks))
	}

	hasRunner := false
	for _, bl := range sqliteResult.Backlinks {
		if bl.Path == "arch/runner.md" && bl.Heading == "Design Rationale" {
			hasRunner = true
		}
	}
	if !hasRunner {
		t.Error("expected backlink from arch/runner.md Design Rationale")
	}
}

func TestSearchFileLevelBacklinks(t *testing.T) {
	dbPath := setupLinkedDB(t)

	s, err := Open(dbPath)
	if err != nil {
		t.Fatalf("Open failed: %v", err)
	}
	defer s.Close()

	results, err := s.Search("storage", 10)
	if err != nil {
		t.Fatalf("Search failed: %v", err)
	}

	if len(results) < 2 {
		t.Fatal("expected at least 2 results for 'storage'")
	}

	for _, r := range results {
		if r.Path != "arch/storage.md" {
			continue
		}
		hasIndex := false
		for _, bl := range r.Backlinks {
			if bl.Path == "index.md" {
				hasIndex = true
			}
		}
		if !hasIndex {
			t.Errorf("storage section %q missing file-level backlink from index.md", r.Heading)
		}
	}
}
