# wt Documentation Refresh

## What This Is

A documentation update for `wt`, a zsh-based Git worktree manager. The codebase has evolved (new commands, override flags, .wtconfig support) and the internal codebase map is stale. The project also lacks user-facing documentation entirely — no README exists.

## Core Value

Accurate, complete documentation that matches the current codebase so users can install, learn, and use every `wt` feature.

## Requirements

### Validated

(None yet — ship to validate)

### Active

- [ ] Full refresh of all 7 `.planning/codebase/` files against current `wt.zsh`
- [ ] README.md at repo root with full documentation:
  - [ ] Installation section (source `wt.zsh` in `.zshrc`)
  - [ ] Every command documented with usage and examples
  - [ ] All flags documented (`--copy`, `--symlink` for `new` and `eject`)
  - [ ] `.wtconfig` format and behavior explained
  - [ ] Container/Docker workflow section (port overrides, `.env` handling, named volumes)
  - [ ] Tab completion mentioned

### Out of Scope

- Plugin manager install docs (oh-my-zsh, zinit, etc.) — just source for now
- Man pages or other doc formats — README.md only
- Code changes — documentation only, no feature work

## Context

- Single-file zsh tool (`wt.zsh`, ~550 lines)
- Recent additions not reflected in codebase map: `wt init`, `wt eject`, `--copy`/`--symlink` override flags, `.wtconfig` support
- No README.md exists yet
- Existing codebase map at `.planning/codebase/` with 7 files (ARCHITECTURE, CONVENTIONS, CONCERNS, INTEGRATIONS, STACK, STRUCTURE, TESTING)

## Constraints

- **Format**: README.md in repo root, codebase docs in `.planning/codebase/`
- **Accuracy**: All docs must match current code exactly — no aspirational features

## Key Decisions

| Decision | Rationale | Outcome |
|----------|-----------|---------|
| Source-only install | Simplest approach, no package manager overhead | -- Pending |
| Include container tips in README | Practical value for real-world usage with Docker apps | -- Pending |

---
*Last updated: 2026-02-07 after initialization*
