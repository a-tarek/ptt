# Roadmap: wt Documentation Refresh

## Overview

A two-phase documentation project: first refresh the internal codebase map to match current wt.zsh features (new commands, override flags, .wtconfig support), then write comprehensive user-facing documentation in README.md. Phase 2 depends on Phase 1 to ensure user docs reference accurate internal knowledge.

## Phases

**Phase Numbering:**
- Integer phases (1, 2): Planned milestone work
- Decimal phases (1.1, 1.2): Urgent insertions (marked with INSERTED)

Decimal phases appear between their surrounding integers in numeric order.

- [ ] **Phase 1: Internal Documentation Refresh** - Update codebase map files to match current code
- [ ] **Phase 2: User-Facing Documentation** - Write complete README.md with all features documented

## Phase Details

### Phase 1: Internal Documentation Refresh
**Goal**: Accurate codebase map reflecting all current wt.zsh features and structure
**Depends on**: Nothing (first phase)
**Requirements**: MAP-01, MAP-02, MAP-03, MAP-04, MAP-05
**Success Criteria** (what must be TRUE):
  1. ARCHITECTURE.md reflects all 15+ functions currently in wt.zsh
  2. CONVENTIONS.md documents flag parsing patterns and override mechanism
  3. STRUCTURE.md line ranges match current file layout
  4. CONCERNS.md lists all known issues (zsh reserved vars, edge cases)
  5. STACK.md, INTEGRATIONS.md, TESTING.md accurately reflect current state
**Plans**: TBD

Plans:
- [ ] (Plans will be created during plan-phase)

### Phase 2: User-Facing Documentation
**Goal**: Complete README.md enabling users to install, learn, and use every wt feature
**Depends on**: Phase 1
**Requirements**: DOC-01, DOC-02, DOC-03, DOC-04, DOC-05, DOC-06, DOC-07, DOC-08, DOC-09, DOC-10, DOC-11
**Success Criteria** (what must be TRUE):
  1. User can clone repo and source wt.zsh in .zshrc following README instructions
  2. User can find complete usage for every command (new, eject, goto, home, list, merge, rebase, delete, init)
  3. User understands all flags (--copy, --symlink) and their use cases
  4. User can create and customize .wtconfig following documented format
  5. User knows container/Docker workflows (port overrides, .env handling, named volumes)
**Plans**: TBD

Plans:
- [ ] (Plans will be created during plan-phase)

## Progress

**Execution Order:**
Phases execute in numeric order: 1 → 2

| Phase | Plans Complete | Status | Completed |
|-------|----------------|--------|-----------|
| 1. Internal Documentation Refresh | 0/TBD | Not started | - |
| 2. User-Facing Documentation | 0/TBD | Not started | - |
