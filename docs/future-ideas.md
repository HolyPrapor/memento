# Future Ideas

Improvements to evaluate after MVP is validated.

## Query Quality

**Synonyms.** Expand query terms via `.memento/synonyms.yaml`:

```yaml
flaky: [intermittent, nondeterministic, unstable]
runner: [executor, worker]
storage: [persistence, database]
```

Preprocess query before FTS5 — OR-join synonyms for each matched term. Low complexity, high potential impact for domain-specific vocabulary.

**Query cleanup.** Strip filler words from natural language queries ("why is", "how do I", "the") before search. Unknown whether FTS5 BM25 handles this well enough natively.

**Fuzzy / typo-tolerant search.** FTS5 doesn't do fuzzy matching natively. Could add trigram-based fallback or spellcheck layer for common typos.

## Output

**YAML output.** Add `--yaml` flag alongside `--json`. Trivial to add, but JSON covers agent consumption.

**Rich terminal output.** Colors, bolded headings, syntax-highlighted snippets. Nice-to-have for human users browsing in terminal.

## Indexing

**Auto-reindex (file watcher).** `wiki-watch` command that watches the wiki directory and reindexes on changes. Useful during wiki editing sessions.

**Incremental indexing.** Currently reindexes everything. For large wikis (1000+ files), track file mtimes and only reindex changed files.

**Frontmatter metadata.** Parse YAML frontmatter in wiki files for optional section metadata (tags, weight, aliases). Could feed into ranking.

## Distribution & Integration

**Self-update.** `wiki-update` command that downloads the latest GitHub release binary. Requires release infrastructure and checksum verification.

**Install script.** One-liner `curl | sh` or PowerShell script that downloads the platform binary and puts it on PATH.

**opencode/Claude Code skill.** Distribute a SKILL.md so coding agents automatically know to search the wiki. See dedicated `docs/skill-design.md` for details.

**CI/CD integration.** GitHub Action to verify wiki is indexed and searchable on PR. Catch stale or broken wiki sections.

## Search

**Vector search fallback.** If keyword search returns few or no results, fall back to an embedding-based semantic search. Requires an embeddings model (local, no SaaS). Only if FTS5 proves insufficient for natural language queries.

**Cross-file references.** During indexing, extract Markdown links (`[text](path#anchor)`) from section bodies into a `section_links` table.

Backlink search — `wiki-search "topic" --backlinks` — appends a "Referenced by:" field listing sections that link to each result. Useful for understanding impact or finding related context.

Wiki-style links (`[[runner#worker-pool]]`) as a shorthand that resolves by slug/anchor without full paths. More resilient to renames than relative paths.

**Link validation.** `wiki-lint` checks that all `[text](path)` and `[[wikilinks]]` resolve to existing sections. Run in CI to catch stale references after renames or deletions.

**Git-aware search.** Show which git commit last touched a section. Filter results to a specific branch or date range. Answer "when did this rationale change?"

**Search in source code, linked.** Return wiki sections that reference specific source files or symbols. Tightens the wiki-code connection.

## Wiki Content

**Templates / scaffolding.** `wiki-new architecture|decision|debugging <name>` that creates a pre-structured Markdown file with suggested headings.

**Section merge conflict detection.** After `git merge`, detect wiki sections that may contain contradictory or outdated information. Better suited for manual review than automation.

**Wiki health dashboard.** `wiki-stats` showing section counts, file sizes, last-updated dates, orphan pages. Helps maintainers spot neglected areas.

## Benchmarks

**Retrieval evaluation suite.** Set of known-good queries mapped to expected best-matching sections. Necessary for tuning ranking, evaluating synonyms, and comparing search strategies. Metrics: MRR, NDCG@5, Recall@5.

**Phase 1 — hand-written golden set.** A small test wiki (10-20 sections across a few files) with ~20 query → expected-section pairs, hand-written to test specific retrieval scenarios:
- Exact keyword match on heading
- Synonym-like match (query uses different words than body)
- Multi-word query where terms appear across different parts of a section
- Query that should match multiple sections (ranking order matters)
- Query that should return no results (nil handling)

This is high-quality but small. Good enough for MVP tuning.

**Phase 2 — real-world corpus.** Gather documentation from well-maintained open-source projects as a larger, realistic evaluation corpus:

* `github.com/rust-lang/rfcs` — design RFCs with clear rationale sections
* `github.com/kubernetes/community/tree/master/design-proposals` — Kubernetes design proposals
* `github.com/reactjs/rfcs` — React RFCs
* `github.com/golang/proposal` — Go proposals

Self-retrieval test: for each section, use its heading + first sentence as the query and verify the section ranks #1. This is a weak signal (biased), but can catch indexing bugs and severe ranking failures at scale.

LLM-assisted evaluation: use an LLM to generate plausible natural-language queries for each section, then manually spot-check a sample for quality. Faster than writing all query pairs by hand, but requires verification to avoid GIGO.

The benchmark runner takes a JSON file:

```json
[
  {
    "query": "why is the runner separated from the scheduler",
    "expected_section": "docs/agent-wiki/architecture/runner.md#design-rationale",
    "min_rank": 3
  }
]
```

…indexes the wiki, runs each query, and reports metrics. New ranking heuristics or index tweaks can be validated by re-running and comparing scores.
