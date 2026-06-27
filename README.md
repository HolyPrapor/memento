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
memento search "runner separation"  # search with ranked results (JSON)
memento search --limit 5 "query"    # limit results
```

## How It Works

1. **Index** walks a directory of `.md` files, splits them into sections at
   each heading (`#`...`######`), and rebuilds a SQLite FTS5 index.
2. **Search** queries the index with BM25 ranking, returns JSON with snippets,
   relevance scores, and backlinks — sections that reference each result.
   Headings are weighted 3x over body text for ranking.
3. **Backlinks** are extracted from Markdown links (`[text](path.md#heading)`)
   during indexing. Use them to discover related docs and change impact.

## Configuration

Create `.memento/config.yaml` next to your database to downrank sections
by path glob (first match wins). No reindex needed.

```yaml
downrank:
  - paths: ["Obsolete/*"]
    factor: 0.1
  - paths: ["Migration/*"]
    factor: 0.3
```

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
