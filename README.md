# memento

A minimal Go CLI for repo-local codebase knowledge wikis. Searches heading-based
Markdown sections with SQLite FTS5 — better than grep.

## Install

**Windows (PowerShell):**
```powershell
powershell -c "irm https://raw.githubusercontent.com/HolyPrapor/memento/master/install.ps1 | iex"
```

**Linux / macOS:**
```bash
curl -sSL https://raw.githubusercontent.com/HolyPrapor/memento/master/install.sh | bash
```

Or pick a binary from [releases](https://github.com/HolyPrapor/memento/releases).

## Usage

```bash
memento index docs/agent-wiki       # build the search index
memento search "runner separation"  # search with ranked results
memento search --json "migration"   # structured output for scripts
memento search --limit 5 "query"    # limit results
```

## How It Works

1. **Index** walks a directory of `.md` files, splits them into sections at
   each heading (`#`...`######`), and rebuilds a SQLite FTS5 index.
2. **Search** queries the index with BM25 ranking, returns snippets with
   highlighted matches, paths, headings, and relevance scores.
3. Sections are the search unit — not whole files, not grep lines. Headings
   are weighted 3x over body text.

## Writing Wiki Pages

- Make headings descriptive and keyword-rich (they are queries)
- Keep sections self-contained and focused (they are search results)
- Prefer many small files over one mega-file
- Use relative Markdown links between sections
- See [`docs/authoring-guide.md`](docs/authoring-guide.md) for full guidelines

## Build From Source

```bash
go build -ldflags="-s -w" -o memento ./cmd/memento
```

Requires Go 1.21+.
