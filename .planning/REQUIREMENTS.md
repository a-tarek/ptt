# Requirements: wt Documentation Refresh

**Defined:** 2026-02-07
**Core Value:** Accurate, complete documentation that matches the current codebase

## v1 Requirements

### Codebase Map Refresh

- [x] **MAP-01**: ARCHITECTURE.md reflects current function structure (all 15+ functions)
- [x] **MAP-02**: CONVENTIONS.md documents current patterns (flag parsing, override mechanism)
- [x] **MAP-03**: STRUCTURE.md matches current file layout and line ranges
- [x] **MAP-04**: CONCERNS.md updated with current known issues (zsh reserved vars, etc.)
- [x] **MAP-05**: STACK.md, INTEGRATIONS.md, TESTING.md refreshed against current code

### README — Installation

- [ ] **DOC-01**: Clone/download instructions
- [ ] **DOC-02**: Source-in-zshrc setup with example

### README — Command Reference

- [ ] **DOC-03**: `wt new` documented with all flags (`--copy`, `--symlink`), positional args, examples
- [ ] **DOC-04**: `wt eject` documented with all flags, examples
- [ ] **DOC-05**: `wt goto`, `wt home`, `wt list` documented
- [ ] **DOC-06**: `wt merge`, `wt rebase`, `wt delete` documented
- [ ] **DOC-07**: `wt init` documented

### README — Configuration

- [ ] **DOC-08**: `.wtconfig` format explained (syntax, actions, comments)
- [ ] **DOC-09**: Override flags explained with examples (precedence, one-offs)

### README — Workflows

- [ ] **DOC-10**: Container/Docker workflow section (port overrides, `.env` handling, named volumes)
- [ ] **DOC-11**: Tab completion mentioned

## v2 Requirements

### Future Docs

- **DOC-12**: Plugin manager install docs (oh-my-zsh, zinit, antigen)
- **DOC-13**: Man page generation

## Out of Scope

| Feature | Reason |
|---------|--------|
| Code changes | Documentation only — no feature work |
| Separate docs site | README.md is sufficient for a single-file tool |

## Traceability

| Requirement | Phase | Status |
|-------------|-------|--------|
| MAP-01 | Phase 1 | Complete |
| MAP-02 | Phase 1 | Complete |
| MAP-03 | Phase 1 | Complete |
| MAP-04 | Phase 1 | Complete |
| MAP-05 | Phase 1 | Complete |
| DOC-01 | Phase 2 | Pending |
| DOC-02 | Phase 2 | Pending |
| DOC-03 | Phase 2 | Pending |
| DOC-04 | Phase 2 | Pending |
| DOC-05 | Phase 2 | Pending |
| DOC-06 | Phase 2 | Pending |
| DOC-07 | Phase 2 | Pending |
| DOC-08 | Phase 2 | Pending |
| DOC-09 | Phase 2 | Pending |
| DOC-10 | Phase 2 | Pending |
| DOC-11 | Phase 2 | Pending |

**Coverage:**
- v1 requirements: 16 total
- Mapped to phases: 16
- Unmapped: 0

---
*Requirements defined: 2026-02-07*
*Last updated: 2026-02-07 after roadmap creation*
