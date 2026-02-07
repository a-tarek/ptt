# Milestones

## v1.0: Documentation Refresh (Complete)

**Completed:** 2026-02-07
**Phases:** 1-2 (6 plans total)

### What shipped:
- Internal codebase map refreshed (7 files: ARCHITECTURE, CONVENTIONS, CONCERNS, INTEGRATIONS, STACK, STRUCTURE, TESTING)
- README.md with full user-facing documentation
  - Installation (source wt.zsh in .zshrc)
  - All 9 commands documented with usage and examples
  - All flags documented (--copy, --symlink for new and eject)
  - .wtconfig format and behavior explained
  - Container/Docker workflow section
  - Tab completion mentioned

### Requirements delivered:
- MAP-01 through MAP-05: Codebase map refresh ✓
- DOC-01 through DOC-11: User-facing documentation ✓

### Key decisions:
- Source-only install (simplest approach)
- Quick start shows 5-command workflow (init, new, goto, home, delete)
- Container tips included in README
- Document flag override precedence (flags override .wtconfig)
