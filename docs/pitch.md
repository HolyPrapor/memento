I want to design a simple project for maintaining a **repo-local codebase knowledge wiki** that coding agents can search and update.

The goal is **not** personal memory. User preferences and personal instructions are handled separately via `AGENTS.md`.

The goal is to help coding agents like Codex, Claude Code, or similar tools understand and maintain durable **project knowledge**, such as:

* architecture
* design rationale
* important tradeoffs
* invariants
* debugging history
* migration notes
* “why things are this way”

This is **not** a replacement for reading the source code. The source code remains authoritative for current behavior. The wiki should only provide supporting context and rationale.

I want the project to be extremely simple to use.

The source of truth should be repo-local Markdown files, for example:

```text
docs/agent-wiki/
  index.md
  architecture/
  decisions/
  invariants/
  debugging/
  migrations/
  glossary.md
```

But the system should also work if the Markdown files are not perfectly structured.

The main feature I want is **better-than-grep search** over these Markdown docs.

My preferred design is:

```text
Markdown files
  -> split into heading-based sections
  -> index sections in SQLite FTS5
  -> expose a tiny CLI search tool
  -> return ranked results with paths, headings, snippets, and scores
```

The CLI should be minimal:

```bash
wiki-index docs/agent-wiki
wiki-search "why is the runner separated from the scheduler?"
wiki-search --json "storage migration flaky tests"
```

No complicated taxonomy. No required `--kind` flags. No database server. No SaaS. No Obsidian requirement. No vector DB unless there is a very strong reason.

Search results should look roughly like:

```json
[
  {
    "path": "docs/agent-wiki/debugging/flaky-indexer-tests.md",
    "anchor": "storage-migration-race",
    "heading": "Storage migration race",
    "score": -4.72,
    "snippet": "The indexer tests became flaky after the storage migration because..."
  }
]
```

or the same for yaml.

The important implementation idea is that the search unit should be a **Markdown section**, not a whole file and not a grep line.

The system should improve on grep by:

* ranking results
* returning relevant sections
* showing snippets
* weighting headings higher than body text
* supporting JSON output for agents
* optionally doing simple query cleanup
* optionally using a small synonyms file

Example synonyms file:

```yaml
flaky:
  - intermittent
  - nondeterministic
  - unstable

runner:
  - executor
  - worker

storage:
  - persistence
  - database
```

Agents should be instructed through `AGENTS.md` to:

1. Search the wiki before architectural, migration, storage, API, testing, or broad debugging work.
2. Treat wiki results as context and rationale, not as source of truth over code.
3. Read the source code before trusting any statement about current behavior.
4. Update the smallest relevant Markdown section when a task reveals durable project knowledge.
5. Keep wiki updates factual, small, reviewable, and committed as normal Markdown changes.
6. Never store personal memory, transient scratch notes, or vague generated summaries in the wiki.

I am looking for a practical design and implementation plan for this project.

Please keep the design boring and minimal. Prefer SQLite FTS5, a small Markdown section indexer, and a tiny CLI. Avoid obscure “AI memory” projects, opaque SaaS, large platforms, or overengineered knowledge systems.

