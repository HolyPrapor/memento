package searcher

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	sqlite "memento/internal/sqlite"
)

type Backlink struct {
	Path    string `json:"path"`
	Heading string `json:"heading"`
}

type Result struct {
	Path      string     `json:"path"`
	Heading   string     `json:"heading,omitempty"`
	Relevance float64    `json:"relevance"`
	Snippet   string     `json:"snippet"`
	Backlinks []Backlink `json:"backlinks,omitempty"`
}

type Searcher struct {
	DB *sql.DB
}

func Open(dbPath string) (*Searcher, error) {
	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("database not found at %s — run 'memento index <wiki-dir>' first", dbPath)
	}

	db, err := sqlite.OpenDB(dbPath)
	if err != nil {
		return nil, err
	}

	return &Searcher{DB: db}, nil
}

func (s *Searcher) Close() error {
	return s.DB.Close()
}

func (s *Searcher) Search(query string, limit int) ([]Result, error) {
	escaped := escapeFTS5Query(query)
	ftsQuery := buildFTSQuery(escaped)

	rows, err := s.DB.Query(`
		SELECT
			s.path,
			s.anchor,
			CASE WHEN s.heading_level = 0 THEN '' ELSE s.heading END AS heading,
			-bm25(sections_fts) AS relevance,
			snippet(sections_fts, 0, '<mark>', '</mark>', '...', 64) AS snippet
		FROM sections_fts
		JOIN sections s ON s.id = sections_fts.rowid
		WHERE sections_fts MATCH ?
		ORDER BY relevance DESC
		LIMIT ?
	`, ftsQuery, limit)
	if err != nil {
		return nil, fmt.Errorf("search query: %w", err)
	}
	defer rows.Close()

	var results []Result
	var targetPairs []struct{ path, anchor string }
	for rows.Next() {
		var r Result
		var anchor string
		if err := rows.Scan(&r.Path, &anchor, &r.Heading, &r.Relevance, &r.Snippet); err != nil {
			return nil, fmt.Errorf("scan result: %w", err)
		}
		results = append(results, r)
		targetPairs = append(targetPairs, struct{ path, anchor string }{r.Path, anchor})
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows iteration: %w", err)
	}

	if len(results) > 0 {
		backlinks, err := s.getBacklinks(targetPairs)
		if err != nil {
			return nil, fmt.Errorf("backlinks: %w", err)
		}
		for i := range results {
			key := targetPairs[i].path + "\x00" + targetPairs[i].anchor
			results[i].Backlinks = backlinks[key]
		}
	}

	if results == nil {
		results = []Result{}
	}

	return results, nil
}

func (s *Searcher) SearchJSON(query string, limit int) error {
	results, err := s.Search(query, limit)
	if err != nil {
		return err
	}

	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	return encoder.Encode(results)
}

func (s *Searcher) getBacklinks(pairs []struct{ path, anchor string }) (map[string][]Backlink, error) {
	if len(pairs) == 0 {
		return nil, nil
	}

	seen := make(map[string]bool)
	var queryPairs []struct{ path, anchor string }
	for _, p := range pairs {
		key := p.path + "\x00" + ""
		if !seen[key] {
			seen[key] = true
			queryPairs = append(queryPairs, struct{ path, anchor string }{p.path, ""})
		}
		key = p.path + "\x00" + p.anchor
		if !seen[key] && p.anchor != "" {
			seen[key] = true
			queryPairs = append(queryPairs, p)
		}
	}

	placeholders := make([]string, len(queryPairs))
	args := make([]interface{}, 0, len(queryPairs)*2)
	for i, p := range queryPairs {
		placeholders[i] = "(?, ?)"
		args = append(args, p.path, p.anchor)
	}

	query := fmt.Sprintf(`
		SELECT sl.target_path, sl.target_anchor,
		       s.path, s.heading
		FROM section_links sl
		JOIN sections s ON s.id = sl.source_section_id
		WHERE (sl.target_path, sl.target_anchor) IN (%s)
		ORDER BY s.path, s.heading
	`, strings.Join(placeholders, ", "))

	rows, err := s.DB.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("query backlinks: %w", err)
	}
	defer rows.Close()

	raw := make(map[string][]Backlink)
	for rows.Next() {
		var targetPath, targetAnchor, sourcePath, sourceHeading string
		if err := rows.Scan(&targetPath, &targetAnchor, &sourcePath, &sourceHeading); err != nil {
			return nil, fmt.Errorf("scan backlink: %w", err)
		}
		key := targetPath + "\x00" + targetAnchor
		raw[key] = append(raw[key], Backlink{Path: sourcePath, Heading: sourceHeading})
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	result := make(map[string][]Backlink)
	for _, p := range pairs {
		key := p.path + "\x00" + p.anchor
		fileKey := p.path + "\x00"

		if links, ok := raw[key]; ok {
			result[key] = links
		}
		if fileLinks, ok := raw[fileKey]; ok {
			result[key] = append(result[key], fileLinks...)
		}
	}

	return result, nil
}

func escapeFTS5Query(query string) string {
	special := "\"*^(){}[]+-~&|!<>=':\\"
	var buf strings.Builder
	for _, r := range query {
		if strings.ContainsRune(special, r) {
			buf.WriteByte(' ')
			continue
		}
		buf.WriteRune(r)
	}
	return strings.TrimSpace(buf.String())
}

func buildFTSQuery(query string) string {
	query = strings.TrimSpace(query)
	if query == "" {
		return ""
	}

	terms := strings.Fields(query)
	if len(terms) == 0 {
		return ""
	}

	return strings.Join(terms, " OR ")
}
