package memento

import (
	"encoding/json"
	"math"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	"memento/internal/indexer"
	"memento/internal/searcher"
)

type queryPair struct {
	Query           string `json:"query"`
	ExpectedSection string `json:"expected_section"`
	ExpectedHeading string `json:"expected_heading"`
	MinRank         int    `json:"min_rank"`
}

func loadQueries(t *testing.T, path string) []queryPair {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read queries file: %v", err)
	}
	var queries []queryPair
	if err := json.Unmarshal(data, &queries); err != nil {
		t.Fatalf("unmarshal queries: %v", err)
	}
	return queries
}

func TestRetrievalBenchmark(t *testing.T) {
	wikiDir := filepath.Join("testdata", "wiki")
	queriesPath := filepath.Join("testdata", "queries.json")
	dbPath := filepath.Join(t.TempDir(), "bench.db")

	if err := indexer.Index(wikiDir, dbPath); err != nil {
		t.Fatalf("index: %v", err)
	}

	s, err := searcher.Open(dbPath)
	if err != nil {
		t.Fatalf("open searcher: %v", err)
	}
	defer s.Close()

	queries := loadQueries(t, queriesPath)
	if len(queries) == 0 {
		t.Fatal("no queries loaded")
	}

	var relevantCount, totalAtK float64
	var mrrSum float64
	k := 5
	queryCount := 0
	passCount := 0

	for _, qp := range queries {
		results, err := s.Search(qp.Query, k)
		if err != nil {
			t.Errorf("search %q failed: %v", qp.Query, err)
			continue
		}
		queryCount++

		if qp.ExpectedSection == "" {
			if len(results) == 0 {
				passCount++
			}
			continue
		}

		found := false
		for rank, r := range results {
			if r.Path == qp.ExpectedSection && strings.EqualFold(r.Heading, qp.ExpectedHeading) {
				found = true
				if rank+1 <= qp.MinRank {
					passCount++
				}
				mrrSum += 1.0 / float64(rank+1)
				totalAtK++
				break
			}
		}

		if !found {
			t.Logf("query %q: expected section %q#%q not found in top %d results", qp.Query, qp.ExpectedSection, qp.ExpectedHeading, k)
		}

		if !found {
			continue
		}
		for _, r := range results {
			_ = r
		}
	}

	relevantCount = totalAtK

	if queryCount > 0 {
		mrr := mrrSum / float64(queryCount)
		recallAtK := relevantCount / float64(queryCount)
		dcgAtK := computeDCG(queries, s, k)
		idcgAtK := float64(queryCount)
		if idcgAtK == 0 {
			idcgAtK = 1
		}
		ndcgAtK := dcgAtK / idcgAtK

		t.Logf("MRR: %.3f", mrr)
		t.Logf("Recall@%d: %.3f", k, recallAtK)
		t.Logf("NDCG@%d: %.3f", k, ndcgAtK)
	}
}

func computeDCG(queries []queryPair, s *searcher.Searcher, k int) float64 {
	var dcg float64
	for _, qp := range queries {
		if qp.ExpectedSection == "" {
			continue
		}
		results, err := s.Search(qp.Query, k)
		if err != nil {
			continue
		}
		for rank, r := range results {
			rel := 0.0
			if r.Path == qp.ExpectedSection && strings.EqualFold(r.Heading, qp.ExpectedHeading) {
				rel = 1.0
			}
			if rel > 0 {
				dcg += rel / math.Log2(float64(rank+2))
			}
		}
	}
	return dcg
}

func TestIndividualQueries(t *testing.T) {
	wikiDir := filepath.Join("testdata", "wiki")
	queriesPath := filepath.Join("testdata", "queries.json")
	dbPath := filepath.Join(t.TempDir(), "indiv.db")

	if err := indexer.Index(wikiDir, dbPath); err != nil {
		t.Fatalf("index: %v", err)
	}

	s, err := searcher.Open(dbPath)
	if err != nil {
		t.Fatalf("open searcher: %v", err)
	}
	defer s.Close()

	queries := loadQueries(t, queriesPath)

	for _, qp := range queries {
		t.Run(qp.Query, func(t *testing.T) {
			results, err := s.Search(qp.Query, 10)
			if err != nil {
				t.Fatalf("search failed: %v", err)
			}

			if qp.ExpectedSection == "" {
				if len(results) > 0 {
					t.Errorf("expected no results, got %d", len(results))
				}
				return
			}

			found := false
			var foundRank int
			for i, r := range results {
				if r.Path == qp.ExpectedSection && strings.EqualFold(r.Heading, qp.ExpectedHeading) {
					found = true
					foundRank = i + 1
					break
				}
			}

			if !found {
				t.Errorf("expected section %q#%q not found in results (got %d results)", qp.ExpectedSection, qp.ExpectedHeading, len(results))
				for i, r := range results {
					t.Logf("  result %d: %s#%s [%s] score=%.2f", i+1, r.Path, r.Anchor, r.Heading, r.Score)
				}
			} else if foundRank > qp.MinRank {
				t.Errorf("expected section %q#%q at rank <= %d, got rank %d", qp.ExpectedSection, qp.ExpectedHeading, qp.MinRank, foundRank)
			}
		})
	}
}

var _ = sort.Ints
