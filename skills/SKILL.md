# Memento — Agent Memory Skill

## Purpose

Memento searches a project's agent wiki — a collection of Markdown files that
explain how the codebase works, why decisions were made, and where things
live. It provides better-than-grep search over heading-based sections,
returning ranked results with snippets.

## When to Use This Skill

**Before exploring or changing code**, search the wiki first. It may already
contain the context you need. Trigger this skill whenever you need to
understand something about the codebase:

- How does a system or feature work? (e.g. "game fighting system",
  "data generation pipeline", "auth flow", "payment retry logic")
- What is the architecture of a component or module?
- Why was a particular design decision made?
- Where should I look to find code related to X?
- What invariants or constraints does the system rely on?
- Are there known pitfalls, edge cases, or debugging notes about an area?
- Has this been changed before, and what was the migration process?

Think of the wiki as a guided tour of the codebase. Before diving into
unfamiliar code, ask the wiki: *what should I know about this area?*

## Workflow

### 1. Orient First

Every session starts with orientation. Before opening any source file, run:

```bash
memento search "codebase map architecture glossary"
```

This surfaces the orientation documents — codebase map (what lives where,
what's current vs. obsolete), architecture overview (how components connect),
and glossary (domain terms). These tell you which parts of the codebase are
*relevant* and which are *safe to skip* before you waste time reading dead
code.

If these docs don't exist yet, note it — the codebase lacks orientation
materials, so you'll need to map the area yourself as you work.

### 2. Search Before Acting

When asked to understand or modify something specific, follow up with a
targeted search:

```bash
memento search <query>
```

For structured results:

```bash
memento search --json "<query>"
```

Frame your query around what you want to understand. Good examples:
- `memento search "game fighting system"`
- `memento search "save checkpoint persistence"`
- `memento search "card merge interaction rules"`

### 3. Watch for Stale Signals

Wiki sections are not guaranteed current. When reading results, look for:

- **Status markers** — sections labeled "OBSOLETE", "Planned", "Phase X" tell
  you whether the content reflects the current codebase or aspirations
- **File references** — sections that cite specific source files are more
  trustworthy than vague overviews; verify at least one reference against the
  actual code
- **Date or version mentions** — if a section references a version or milestone
  that has passed, check whether the described work actually landed

If a section is clearly outdated and you can't verify it, say so rather than
relying on it.

### 4. Treat Results as Context, Not Authority

Wiki results provide orientation and rationale. They are not a substitute for
reading the source code. Always verify claims against the actual code — the
source code is authoritative for current behavior. The wiki tells you *why*
and *where*, the code tells you *what* and *how*.

Use search results to:
- Identify which files and modules are relevant to your task
- Understand the design rationale before proposing changes
- Find related systems you might also need to touch
- Learn about constraints you must respect
- Know which systems are dead code — don't modify them

### 5. Contribute Back

After completing a task, if you learned something durable about the codebase,
add it to the wiki. Good contributions:

- A non-obvious interaction between components
- A design constraint that isn't clear from the types alone
- A migration gotcha or required ordering
- A subsystem overview you wish you'd had when starting
- A debugging note that saved you hours
- A status update on a section that has become outdated

Keep updates factual, small, and reviewable. Add to the smallest relevant
section. Never store personal memory, transient notes, or generated summaries.

### 6. Reindex After Editing

After editing wiki files:

```bash
memento index docs/agent-wiki
```

### 7. Command Reference

```
memento index [--db .memento/wiki.db] <wiki-dir>
memento search [--db .memento/wiki.db] [--json] [--limit 10] <query>
```

Options:
- `--db` — database path (default: .memento/wiki.db)
- `--json` — output results as JSON
- `--limit` — max results (default: 10)
