package indexer

import (
	"database/sql"
	"os"
	"path/filepath"
	"testing"
)

func TestBuildHeadingWeight(t *testing.T) {
	tests := []struct {
		heading  string
		expected string
	}{
		{
			heading:  "Getting Started",
			expected: "Getting Started\nGetting Started\nGetting Started",
		},
		{
			heading:  "",
			expected: "",
		},
	}

	for _, tc := range tests {
		got := buildHeadingWeight(tc.heading)
		if got != tc.expected {
			t.Errorf("buildHeadingWeight(%q) = %q, want %q", tc.heading, got, tc.expected)
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

func TestLinksStored(t *testing.T) {
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

	dbPath := filepath.Join(t.TempDir(), "test.db")
	if err := Index(wikiDir, dbPath); err != nil {
		t.Fatalf("Index failed: %v", err)
	}

	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	defer db.Close()

	rows, err := db.Query(`
		SELECT sl.target_path, sl.target_anchor, sl.link_text,
		       s.path AS source_path, s.heading
		FROM section_links sl
		JOIN sections s ON s.id = sl.source_section_id
		ORDER BY s.path, s.heading
	`)
	if err != nil {
		t.Fatalf("query links: %v", err)
	}
	defer rows.Close()

	type row struct {
		targetPath, targetAnchor, linkText, sourcePath, heading string
	}
	var links []row
	for rows.Next() {
		var r row
		if err := rows.Scan(&r.targetPath, &r.targetAnchor, &r.linkText, &r.sourcePath, &r.heading); err != nil {
			t.Fatalf("scan: %v", err)
		}
		links = append(links, r)
	}

	if len(links) != 3 {
		t.Fatalf("expected 3 links, got %d", len(links))
	}

	findLink := func(targetPath, targetAnchor string) (row, bool) {
		for _, l := range links {
			if l.targetPath == targetPath && l.targetAnchor == targetAnchor {
				return l, true
			}
		}
		return row{}, false
	}

	r, ok := findLink("arch/runner.md", "design-rationale")
	if !ok {
		t.Error("missing link to arch/runner.md#design-rationale")
	} else {
		if r.sourcePath != "index.md" {
			t.Errorf("link to runner: expected source index.md, got %s", r.sourcePath)
		}
		if r.linkText != "runner design" {
			t.Errorf("link to runner: expected text 'runner design', got %q", r.linkText)
		}
	}

	r, ok = findLink("arch/storage.md", "")
	if !ok {
		t.Error("missing link to arch/storage.md")
	} else {
		if r.sourcePath != "index.md" {
			t.Errorf("link to storage: expected source index.md, got %s", r.sourcePath)
		}
	}

	r, ok = findLink("arch/storage.md", "why-sqlite")
	if !ok {
		t.Error("missing link to arch/storage.md#why-sqlite")
	} else {
		if r.sourcePath != "arch/runner.md" {
			t.Errorf("link to storage#why-sqlite: expected source arch/runner.md, got %s", r.sourcePath)
		}
		if r.linkText != "storage v2" {
			t.Errorf("link to storage#why-sqlite: expected text 'storage v2', got %q", r.linkText)
		}
	}
}
