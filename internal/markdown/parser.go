package markdown

import (
	"path/filepath"
	"strings"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/text"
)

type Link struct {
	TargetPath   string
	TargetAnchor string
	Text         string
}

type Section struct {
	Path         string
	Anchor       string
	Heading      string
	HeadingLevel int
	Body         string
	SectionOrder int
	Links        []Link
}

func ParseFile(path string, source []byte) ([]Section, error) {
	reader := text.NewReader(source)
	root := goldmark.DefaultParser().Parse(reader)

	headings := collectHeadingInfos(root, source)

	if len(headings) == 0 {
		body := extractPlainText(root, source)
		links := extractLinks(root, source, path)
		baseName := strings.TrimSuffix(filepath.Base(path), ".md")
		return []Section{{
			Path:         path,
			Anchor:       Slugify(baseName),
			Heading:      baseName,
			HeadingLevel: 0,
			Body:         strings.TrimSpace(body),
			SectionOrder: 0,
			Links:        links,
		}}, nil
	}

	var sections []Section
	baseName := strings.TrimSuffix(filepath.Base(path), ".md")

	if headings[0].offset > 0 {
		preambleSource := source[:headings[0].offset]
		preambleDoc := goldmark.DefaultParser().Parse(text.NewReader(preambleSource))
		preambleBody := extractPlainText(preambleDoc, preambleSource)
		preambleBody = strings.TrimSpace(preambleBody)
		if preambleBody != "" {
			links := extractLinks(preambleDoc, preambleSource, path)
			sections = append(sections, Section{
				Path:         path,
				Anchor:       Slugify(baseName),
				Heading:      baseName,
				HeadingLevel: 0,
				Body:         preambleBody,
				SectionOrder: 0,
				Links:        links,
			})
		}
	}

	for i, h := range headings {
		start := h.bodyOffset
		if start >= len(source) {
			start = h.offset
		}
		var end int
		if i+1 < len(headings) {
			end = headings[i+1].offset
		} else {
			end = len(source)
		}

		if start >= end {
			continue
		}

		sectionSource := source[start:end]
		sectionDoc := goldmark.DefaultParser().Parse(text.NewReader(sectionSource))
		body := extractPlainText(sectionDoc, sectionSource)
		body = strings.TrimSpace(body)
		links := extractLinks(sectionDoc, sectionSource, path)

		sections = append(sections, Section{
			Path:         path,
			Anchor:       Slugify(h.text),
			Heading:      h.text,
			HeadingLevel: h.level,
			Body:         body,
			SectionOrder: len(sections),
			Links:        links,
		})
	}

	return sections, nil
}

type headingInfo struct {
	level      int
	text       string
	offset     int
	bodyOffset int
}

func collectHeadingInfos(root ast.Node, source []byte) []headingInfo {
	var headings []headingInfo
	ast.Walk(root, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
		if !entering {
			return ast.WalkContinue, nil
		}
		if n.Kind() == ast.KindHeading {
			h := n.(*ast.Heading)
			text := extractPlainText(n, source)
			if n.Lines().Len() > 0 {
				offset := n.Lines().At(0).Start
				lastLine := n.Lines().At(n.Lines().Len() - 1)
				bodyOffset := lastLine.Stop + 1
				headings = append(headings, headingInfo{
					level:      h.Level,
					text:       text,
					offset:     offset,
					bodyOffset: bodyOffset,
				})
			}
			return ast.WalkSkipChildren, nil
		}
		return ast.WalkContinue, nil
	})
	return headings
}

func extractPlainText(node ast.Node, source []byte) string {
	var buf strings.Builder
	ast.Walk(node, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
		if !entering {
			return ast.WalkContinue, nil
		}
		if n.Kind() == ast.KindText {
			buf.Write(n.Text(source))
		}
		if n.Kind() == ast.KindCodeSpan {
			for c := n.FirstChild(); c != nil; c = c.NextSibling() {
				if c.Kind() == ast.KindText {
					buf.Write(c.Text(source))
				}
			}
			return ast.WalkSkipChildren, nil
		}
		if n.Kind() == ast.KindLink {
			return ast.WalkContinue, nil
		}
		if n.Kind() == ast.KindImage {
			return ast.WalkSkipChildren, nil
		}
		if n.Kind() == ast.KindHTMLBlock {
			return ast.WalkSkipChildren, nil
		}
		if n.Kind() == ast.KindRawHTML {
			return ast.WalkSkipChildren, nil
		}
		return ast.WalkContinue, nil
	})
	return buf.String()
}

func extractLinks(node ast.Node, source []byte, currentPath string) []Link {
	var links []Link
	ast.Walk(node, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
		if !entering || n.Kind() != ast.KindLink {
			return ast.WalkContinue, nil
		}

		l := n.(*ast.Link)
		dest := string(l.Destination)
		if dest == "" {
			return ast.WalkContinue, nil
		}

		if strings.HasPrefix(dest, "http://") || strings.HasPrefix(dest, "https://") {
			return ast.WalkContinue, nil
		}

		targetPath, targetAnchor := splitLinkDest(dest)

		if targetPath == "" {
			targetPath = currentPath
		} else {
			dir := filepath.Dir(currentPath)
			targetPath = filepath.Join(dir, targetPath)
			targetPath = filepath.ToSlash(targetPath)
		}

		text := extractLinkText(n, source)

		links = append(links, Link{
			TargetPath:   targetPath,
			TargetAnchor: targetAnchor,
			Text:         text,
		})
		return ast.WalkContinue, nil
	})
	return links
}

func splitLinkDest(dest string) (path, anchor string) {
	if idx := strings.IndexByte(dest, '#'); idx >= 0 {
		path = dest[:idx]
		anchor = Slugify(dest[idx+1:])
	} else {
		path = dest
	}
	return path, anchor
}

func extractLinkText(node ast.Node, source []byte) string {
	var buf strings.Builder
	ast.Walk(node, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
		if !entering {
			return ast.WalkContinue, nil
		}
		if n.Kind() == ast.KindText {
			buf.Write(n.Text(source))
		}
		return ast.WalkContinue, nil
	})
	return buf.String()
}
