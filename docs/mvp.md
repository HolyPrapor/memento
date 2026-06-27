# MVP Plan

## Goal

A minimal Go CLI that indexes a directory of Markdown files into SQLite FTS5
and provides better-than-grep search over heading-based sections.

## Non-goals (explicitly deferred)

- Synonyms / query expansion
- Query cleanup / filler word stripping
- YAML output
- Self-update command
- Auto-reindex / file watcher
- Link validation
- Vector search
- Any SaaS, server, or network dependency

## Deliverables

1. `memento` Go binary with two subcommands:
   - `memento index <dir>` — walk a directory, parse Markdown into sections, rebuild FTS5 index
   - `memento search <query>` — query FTS5, print ranked results (plain text or `--json`)
2. Test wiki + retrieval benchmark harness
3. opencode skill file (SKILL.md)
4. Wiki authoring guidelines doc

## CLI Spec

```
memento index [--db .memento/wiki.db] <wiki-dir>
memento search [--db .memento/wiki.db] [--json] [--limit 10] <query>
```

- `index` does a full reindex every run. No incremental mode in MVP.
- `search` fails with a clear message if the DB hasn't been built.
- Default DB path: `.memento/wiki.db` (resolved relative to the wiki directory's
  parent, which is assumed to be the repo root).
- `--json` outputs the structured schema from the pitch. Plain text is the
  default for humans.
- `--limit` controls max results (default 10).

## Section Model

```
docs/agent-wiki/
  index.md
  architecture/
    runner.md
  debugging/
    flaky-tests.md
```

Each Markdown file is split on all heading levels (`#` through `######`).

Content before the first heading becomes a *preamble section* (heading =
filename stem without extension, heading_level = 0).

Each section gets:
- **path** — repo-relative path to the file
- **anchor** — GitHub-style slugified heading
- **heading** — raw heading text
- **heading_level** — 0 for preamble, 1-6 otherwise
- **body** — section content (everything between this heading and the next)
- **section_order** — position within the file

## FTS5 Schema

```sql
CREATE VIRTUAL TABLE sections_fts USING fts5(
  heading,
  body,
  path,
  anchor,
  heading_level,
  section_order,
  tokenize='porter',
  content='',
  content_rowid='id'
);

CREATE TABLE sections(
  id INTEGER PRIMARY KEY,
  path TEXT NOT NULL,
  anchor TEXT NOT NULL,
  heading TEXT,
  heading_level INTEGER NOT NULL,
  body TEXT NOT NULL,
  section_order INTEGER NOT NULL
);
```

Content column: insert heading text 3 times concatenated with body text. This
weights heading matches higher in BM25 ranking without custom ranking functions.

```
content_for_fts = heading + "\n" + heading + "\n" + heading + "\n" + body
```

## Markdown Parsing

Use `github.com/yuin/goldmark` (pure Go CommonMark parser).

Iterate the AST to find heading nodes. Split the source Markdown into sections
by heading boundaries. Use `goldmark` to strip inline formatting for FTS
content while preserving link info for future use.

Anchor generation: lowercase, replace non-alphanumeric sequences with `-`,
strip leading/trailing `-`. Follow GitHub conventions (remove punctuation,
compact hyphens).

Handling edge cases:
- Headings inside code fences are NOT section boundaries (goldmark handles this
  since code blocks are separate AST nodes).
- HTML comments with `#` are NOT headings.
- Empty sections: include them, they still have a path + anchor + heading.
- Non-Markdown files in the wiki dir: skip them (no `.md` extension).

## Search

FTS5 query construction:
1. Escape FTS5 special characters in the user query.
2. Wrap in a simple FTS5 boolean expression.
3. Use `snippet(body, '<mark>', '</mark>', '...', 64)` for excerpt generation.
4. Use `bm25()` for ranking.

Results structure (JSON):
```json
[
  {
    "path": "docs/agent-wiki/debugging/flaky-tests.md",
    "anchor": "storage-migration-race",
    "heading": "Storage migration race",
    "heading_level": 2,
    "score": -4.72,
    "snippet": "...after the storage migration because..."
  }
]
```

Scores are negative (BM25 convention). Lower absolute value = better match.

## Project Structure

```
memento/
  go.mod
  go.sum
  cmd/
    memento/
      main.go              # CLI entry point, command dispatch
  internal/
    indexer/
      indexer.go           # Walk directory, parse MD, write to SQLite
      indexer_test.go
    searcher/
      searcher.go           # Query FTS5, format results
      searcher_test.go
    markdown/
      parser.go             # goldmark AST traversal, section splitting
      anchor.go             # Slugify heading -> anchor
      parser_test.go
    sqlite/
      schema.go             # CREATE TABLE statements, DB open/init
  testdata/
    wiki/
      index.md
      architecture/
        runner.md
      decisions/
        storage-v2.md
      debugging/
        flaky-tests.md
    queries.json            # query -> expected_section benchmark pairs
  docs/
    pitch.md
    mvp.md
    future-ideas.md
    authoring-guide.md       # How to write wiki sections that search well
    skill-design.md          # How the opencode skill works
  skills/
    SKILL.md                # opencode skill definition
```

## Retrieval Benchmarks

A table-driven Go test that:
1. Indexes `testdata/wiki/` into an in-memory SQLite DB.
2. Reads `testdata/queries.json`.
3. For each query, runs search and checks the expected section appears within
   the specified `min_rank`.
4. Reports MRR, Recall@5, NDCG@5.

`testdata/queries.json` format:
```json
[
  {
    "query": "why is the runner separated from the scheduler",
    "expected_section": "testdata/wiki/architecture/runner.md#design-rationale",
    "min_rank": 3
  },
  {
    "query": "flaky indexer tests after storage changes",
    "expected_section": "testdata/wiki/debugging/flaky-tests.md#storage-migration-race",
    "min_rank": 1
  }
]
```

Test scenarios to cover:
- Exact keyword match on heading (expect rank 1)
- Keyword match only in body (expect ranked lower than heading match)
- Natural language query (filler words, question form)
- Multi-word query with terms spread across the section
- Query with terms matching multiple sections (check ranking order)
- Query that should return no results

## Go Dependencies

```
github.com/yuin/goldmark    — Markdown parser
modernc.org/sqlite           — Pure-Go SQLite (no cgo, FTS5 enabled)
```

`modernc.org/sqlite` is the right choice: it bundles SQLite in Go, supports
FTS5, and cross-compiles without cgo. The cost is a larger binary (~15 MB after
stripping) but zero compile-time dependencies.

## SKILL.md (opencode Integration)

The skill auto-loads from `~/.claude/skills/memento/SKILL.md` (no config
changes needed). The install script copies it there. For project-local use,
`--local` writes into `.opencode/skills/memento/SKILL.md`.

The skill body instructs agents to:
1. Run `memento search <query>` before architectural, migration, storage,
   testing, or broad debugging work.
2. Use `memento search --json "query"` for structured results when needed.
3. Treat results as context and rationale, never as source of truth over source
   code.
4. Update the smallest relevant Markdown section when a task reveals durable
   project knowledge.
5. Run `memento index <wiki-dir>` after editing wiki files.
6. Keep updates factual, small, reviewable, and committed as normal Markdown
   changes.

## Wiki Authoring Guidelines

Document in `docs/authoring-guide.md`:
- Headings are queries — make them descriptive and keyword-rich
- Sections are search units — keep each self-contained and focused
- Many small files > one mega-file (easier to diff, review, and search)
- Use relative Markdown links to reference other sections
- Include concrete facts, tradeoffs, and code references
- Add a section when you learn something non-obvious that someone would
  waste time rediscovering
- Never store personal memory, scratch notes, or generated summaries

## Build and Release

```
go build -ldflags="-s -w" -o memento ./cmd/memento
```

Push to GitHub, use `goreleaser` to build binaries for linux/amd64, linux/arm64,
darwin/amd64, darwin/arm64, windows/amd64. Tagged releases trigger the build.

Install script: shell one-liner that detects OS/arch, downloads the right
binary, and puts it on PATH.

## What "MVP Done" Looks Like

1. `go build ./cmd/memento` succeeds
2. `./memento index docs/agent-wiki` populates `.memento/wiki.db`
3. `./memento search "runner separation"` returns ranked sections with snippets
4. `./memento search --json "migration"` returns valid JSON
5. Searching with no DB prints a clear error
6. `go test ./...` passes, including retrieval benchmark assertions
7. Non-Markdown files and code-fenced `#` are handled correctly
8. SKILL.md exists and is installable via the install script
