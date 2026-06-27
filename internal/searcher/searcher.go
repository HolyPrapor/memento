package searcher

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	sqlite "memento/internal/sqlite"
)

type Result struct {
	Path         string  `json:"path"`
	Anchor       string  `json:"anchor"`
	Heading      string  `json:"heading"`
	HeadingLevel int     `json:"heading_level"`
	Score        float64 `json:"score"`
	Snippet      string  `json:"snippet"`
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
			s.heading,
			s.heading_level,
			bm25(sections_fts) AS score,
			snippet(sections_fts, 0, '<mark>', '</mark>', '...', 64) AS snippet
		FROM sections_fts
		JOIN sections s ON s.id = sections_fts.rowid
		WHERE sections_fts MATCH ?
		ORDER BY score
		LIMIT ?
	`, ftsQuery, limit)
	if err != nil {
		return nil, fmt.Errorf("search query: %w", err)
	}
	defer rows.Close()

	var results []Result
	for rows.Next() {
		var r Result
		if err := rows.Scan(&r.Path, &r.Anchor, &r.Heading, &r.HeadingLevel, &r.Score, &r.Snippet); err != nil {
			return nil, fmt.Errorf("scan result: %w", err)
		}
		results = append(results, r)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows iteration: %w", err)
	}

	return results, nil
}

func (s *Searcher) SearchJSON(query string, limit int) error {
	results, err := s.Search(query, limit)
	if err != nil {
		return err
	}

	if results == nil {
		results = []Result{}
	}

	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	return encoder.Encode(results)
}

func (s *Searcher) SearchText(query string, limit int) error {
	results, err := s.Search(query, limit)
	if err != nil {
		return err
	}

	if len(results) == 0 {
		fmt.Println("No results found.")
		return nil
	}

	for i, r := range results {
		if i > 0 {
			fmt.Println()
		}
		fmt.Printf("%s#%s\n", r.Path, r.Anchor)
		if r.Heading != "" {
			fmt.Printf("  %s\n", r.Heading)
		}
		fmt.Printf("  score: %.2f\n", r.Score)
		fmt.Printf("  %s\n", r.Snippet)
	}

	return nil
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
