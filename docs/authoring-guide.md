# Wiki Authoring Guidelines

## Orientation Docs

Every wiki needs three reference documents that agents search before
anything else:

- **`CODEBASE_MAP.md`** — what lives where. List each major system, its
  key files, and its status. Mark dead code explicitly: `Battleground
  (OBSOLETE — do not modify)`. Agents waste time on dead code unless
  you tell them it's dead.
- **`ARCHITECTURE.md`** — how components connect. High-level diagram of
  data flow and dependencies.
- **`GLOSSARY.md`** — domain terms an agent must know to read the code.
  Define `Card`, `Merge`, `Biom`, `CodedCard`, etc.

Without these, every session starts blind.

## Headings Are Queries

Make headings descriptive and keyword-rich. An agent searches for what
it wants to understand.

Good: `# Why the Runner Is Separated from the Scheduler`
Bad: `# Overview`

## Sections Are Search Results

Each section becomes an independent result. Keep sections self-contained
and focused. Prefer many small files over one mega-file.

## Cite Source Files

Sections that reference specific files are more trustworthy than vague
overviews. Include at least one path: `Assets/Scripts/Mechanics/SaveCheckpoint.cs`.

## Mark Status

Prefix sections with status when it matters:

- `CURRENT` — reflects actual code
- `OBSOLETE` — dead code, do not modify
- `PLANNED` — not yet implemented

```
# Save System (CURRENT)
# Battleground (OBSOLETE)
# Multiplayer (PLANNED)
```

Move obsolete files to an `Obsolete/` directory and use
[`.memento/config.yaml`](config) to downrank them so they don't clutter
search results.

If a roadmap doc mixes done and planned work, add a status line at the top:
`Status: Phase 3 done, Phase 4 not yet landed. See ResumeRouter.cs.`

## When to Add

Add a section when you learn something non-obvious that someone would
waste time rediscovering — a hidden invariant, a migration gotcha, a
system interaction the types don't reveal.

## What Not to Store

- Personal memory (use `AGENTS.md`)
- Generated summaries of code anyone can read
- Outdated info — delete it, don't archive it
