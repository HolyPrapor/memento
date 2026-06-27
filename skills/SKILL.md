# Memento — Agent Memory Skill

Search a project's agent wiki for architecture, rationale, and guided
codebase exploration.

## Workflow

### 1. Orient
```bash
memento search "codebase map architecture glossary"
```

### 2. Topic search
```bash
memento search "fighting system"
memento search "save checkpoint"
```

### 3. Verify
Wiki provides context and rationale. Source code is authoritative for
current behavior. Cross-reference at least one file path from results
against the actual code.

### 4. Contribute
When you learn something durable — an invariant, a gotcha, a subsystem
overview you wish you'd had — add it to the smallest relevant section.
Keep it factual and small.

### 5. Reindex
```bash
memento index docs/agent-wiki
```

## Commands

```
memento index [--db .memento/wiki.db] <wiki-dir>
memento search [--db .memento/wiki.db] [--limit 10] <query>
memento version
memento update
```

## Configuration

Place `.memento/config.yaml` next to the wiki database to downrank sections
by path glob. First matching rule wins. No reindex required.

```yaml
downrank:
  - paths: ["Obsolete/*"]
    factor: 0.1
```
