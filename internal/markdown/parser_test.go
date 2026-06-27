package markdown

import (
	"strings"
	"testing"
)

func TestSlugify(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"Storage migration race", "storage-migration-race"},
		{"Why SQLite?", "why-sqlite"},
		{"Task Lifecycle", "task-lifecycle"},
		{"Hello, World!", "hello-world"},
		{"  Spaces  Everywhere  ", "spaces-everywhere"},
		{"CamelCase AND CAPS", "camelcase-and-caps"},
		{"123 Numbers", "123-numbers"},
		{"Special @#$% Chars", "special-chars"},
	}

	for _, tc := range tests {
		t.Run(tc.input, func(t *testing.T) {
			got := Slugify(tc.input)
			if got != tc.expected {
				t.Errorf("Slugify(%q) = %q, want %q", tc.input, got, tc.expected)
			}
		})
	}
}

func TestParseFile(t *testing.T) {
	source := []byte(`Preamble text before any heading.

# Getting Started

This is the getting started section.

More content here.

## Prerequisites

You need to install things.

## Configuration

Set up your config file.

# Architecture

The system is designed around modules.
`)

	sections, err := ParseFile("docs/wiki/readme.md", source)
	if err != nil {
		t.Fatalf("ParseFile failed: %v", err)
	}

	if len(sections) != 5 {
		t.Fatalf("expected 5 sections, got %d", len(sections))
	}

	if sections[0].HeadingLevel != 0 {
		t.Errorf("section 0: expected heading_level=0 (preamble), got %d", sections[0].HeadingLevel)
	}
	if sections[0].Heading != "readme" {
		t.Errorf("section 0: expected heading='readme', got %q", sections[0].Heading)
	}

	if sections[1].HeadingLevel != 1 {
		t.Errorf("section 1: expected heading_level=1, got %d", sections[1].HeadingLevel)
	}
	if sections[1].Heading != "Getting Started" {
		t.Errorf("section 1: expected heading='Getting Started', got %q", sections[1].Heading)
	}

	if sections[2].Heading != "Prerequisites" {
		t.Errorf("section 2: expected heading='Prerequisites', got %q", sections[2].Heading)
	}
	if sections[2].HeadingLevel != 2 {
		t.Errorf("section 2: expected heading_level=2, got %d", sections[2].HeadingLevel)
	}

	if sections[3].Heading != "Configuration" {
		t.Errorf("section 3: expected heading='Configuration', got %q", sections[3].Heading)
	}

	if sections[4].Heading != "Architecture" {
		t.Errorf("section 4: expected heading='Architecture', got %q", sections[4].Heading)
	}
	if sections[4].HeadingLevel != 1 {
		t.Errorf("section 4: expected heading_level=1, got %d", sections[4].HeadingLevel)
	}

	for i, s := range sections {
		if s.SectionOrder != i {
			t.Errorf("section %d: expected section_order=%d, got %d", i, i, s.SectionOrder)
		}
		if s.Body == "" {
			t.Errorf("section %d: body is empty", i)
		}
		if s.Path != "docs/wiki/readme.md" {
			t.Errorf("section %d: expected path='docs/wiki/readme.md', got %q", i, s.Path)
		}
	}
}

func TestParseFileNoHeadings(t *testing.T) {
	source := []byte(`This file has no headings at all.
Just a paragraph of text.`)

	sections, err := ParseFile("docs/wiki/plain.md", source)
	if err != nil {
		t.Fatalf("ParseFile failed: %v", err)
	}

	if len(sections) != 1 {
		t.Fatalf("expected 1 section, got %d", len(sections))
	}

	if sections[0].HeadingLevel != 0 {
		t.Errorf("expected heading_level=0, got %d", sections[0].HeadingLevel)
	}
	if sections[0].Heading != "plain" {
		t.Errorf("expected heading='plain', got %q", sections[0].Heading)
	}
}

func TestParseFileEmptyPreamble(t *testing.T) {
	source := []byte(`# Only Heading

Body text here.`)

	sections, err := ParseFile("docs/wiki/onlyheading.md", source)
	if err != nil {
		t.Fatalf("ParseFile failed: %v", err)
	}

	if len(sections) != 1 {
		t.Fatalf("expected 1 section, got %d", len(sections))
	}

	if sections[0].HeadingLevel != 1 {
		t.Errorf("expected heading_level=1, got %d", sections[0].HeadingLevel)
	}
}

func TestParseFileCodeFence(t *testing.T) {
	source := []byte("# Real Heading\n\n```\n# This is not a heading\n```\n\nBody text.")

	sections, err := ParseFile("docs/wiki/code.md", source)
	if err != nil {
		t.Fatalf("ParseFile failed: %v", err)
	}

	if len(sections) != 1 {
		t.Fatalf("expected 1 section, got %d", len(sections))
	}

	if sections[0].Heading != "Real Heading" {
		t.Errorf("expected heading='Real Heading', got %q", sections[0].Heading)
	}

	if strings.Contains(sections[0].Body, "This is not a heading") {
		t.Log("code fence content may appear in body (goldmark extracts text nodes)")
	}
}

func TestParseFileBodyContent(t *testing.T) {
	source := []byte("# Hello\n\nThis is **bold** and *italic* text with `code`.\n\nEnd.")

	sections, err := ParseFile("docs/wiki/formatting.md", source)
	if err != nil {
		t.Fatalf("ParseFile failed: %v", err)
	}

	body := sections[0].Body
	if !strings.Contains(body, "bold") {
		t.Errorf("expected 'bold' in body, got: %s", body)
	}
	if !strings.Contains(body, "italic") {
		t.Errorf("expected 'italic' in body, got: %s", body)
	}
	if !strings.Contains(body, "code") {
		t.Errorf("expected 'code' in body, got: %s", body)
	}
	if strings.Contains(body, "**") {
		t.Errorf("expected no markdown formatting in body, got: %s", body)
	}
}

func TestExtractLinks(t *testing.T) {
	source := []byte(`# Links

See [Runner](architecture/runner.md#design-rationale) for details.
Also [storage](decisions/storage-v2.md) and [same file](#links).
External [Google](https://google.com) is skipped.
`)

	sections, err := ParseFile("docs/wiki/links.md", source)
	if err != nil {
		t.Fatalf("ParseFile failed: %v", err)
	}

	if len(sections) != 1 {
		t.Fatalf("expected 1 section, got %d", len(sections))
	}

	links := sections[0].Links
	if len(links) != 3 {
		t.Fatalf("expected 3 links (external skipped), got %d", len(links))
	}

	if links[0].TargetPath != "docs/wiki/architecture/runner.md" {
		t.Errorf("expected resolved path 'docs/wiki/architecture/runner.md', got %q", links[0].TargetPath)
	}
	if links[0].TargetAnchor != "design-rationale" {
		t.Errorf("expected anchor 'design-rationale', got %q", links[0].TargetAnchor)
	}
	if links[0].Text != "Runner" {
		t.Errorf("expected text 'Runner', got %q", links[0].Text)
	}

	if links[1].TargetPath != "docs/wiki/decisions/storage-v2.md" {
		t.Errorf("expected 'docs/wiki/decisions/storage-v2.md', got %q", links[1].TargetPath)
	}
	if links[1].TargetAnchor != "" {
		t.Errorf("expected empty anchor, got %q", links[1].TargetAnchor)
	}
	if links[1].Text != "storage" {
		t.Errorf("expected text 'storage', got %q", links[1].Text)
	}

	if links[2].TargetPath != "docs/wiki/links.md" {
		t.Errorf("expected same-file path 'docs/wiki/links.md', got %q", links[2].TargetPath)
	}
	if links[2].TargetAnchor != "links" {
		t.Errorf("expected anchor 'links', got %q", links[2].TargetAnchor)
	}
	if links[2].Text != "same file" {
		t.Errorf("expected text 'same file', got %q", links[2].Text)
	}
}
