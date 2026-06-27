# Wiki Authoring Guidelines

## Core Principles

### Headings Are Queries

Make headings descriptive and keyword-rich. A reader should be able to guess the
content from the heading alone. Think of each heading as what someone would
search for to find this information.

**Good:**
- `# Why the Runner Is Separated from the Scheduler`
- `# Storage Migration from Flat Files to SQLite`

**Bad:**
- `# Overview`
- `# Some Notes`

### Sections Are Search Units

The search indexer splits content at each heading level (`#` through `######`).
Each section becomes an independent search result. Keep sections self-contained
and focused on a single topic.

- A section should make sense when read in isolation (with its heading)
- Avoid relying on context from previous sections
- Target 3-15 lines per section; if a section is longer, consider splitting it

### Many Small Files Over One Mega-File

- Easier to diff and review
- Better search granularity
- Simpler to find the right section to update
- Allows natural organization through directory structure

### Use Relative Markdown Links

Link to other sections using standard relative paths:
```markdown
See [Runner Architecture](architecture/runner.md#worker-pool) for details.
```

The system resolves anchors by GitHub slug conventions (lowercase, hyphens).

### Include Concrete Details

- Code references (file paths, function names)
- Tradeoffs with rationale
- Decision dates and context
- Known limitations and workarounds
- Migration steps and ordering

### When to Add a Section

Add a section when you learn something non-obvious that someone would waste time
rediscovering. Good signals:
- A bug took hours to find because of a non-obvious interaction
- A design decision has tradeoffs that aren't clear from code alone
- A migration requires specific ordering or has pitfalls
- An invariant must be maintained but isn't enforced by the type system

### What NOT to Store

- Personal memory, preferences, or user instructions (use `AGENTS.md`)
- Transient scratch notes or WIP thoughts
- Generated summaries of code anyone can read
- Duplicated documentation that belongs in code comments or README
- Outdated information — prefer deleting over keeping stale content
