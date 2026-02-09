# Requirements: ptt v0.1.2

**Defined:** 2026-02-09
**Core Value:** A single `ptt` command that works in any shell on any platform with full autocompletion

## v0.1.2 Requirements

### README

- [ ] **README-01**: README displays SVG banner from `assets/banner.svg`
- [ ] **README-02**: README has one-line project description
- [ ] **README-03**: README has npm install instructions
- [ ] **README-04**: README documents core commands (mk, cd, rm, ls) with one-liner examples
- [ ] **README-05**: README documents `ptt init` with bare repo conversion workflow
- [ ] **README-06**: README fits approximately one screen — no verbose explanations, no separate docs

### Release

- [ ] **REL-01**: Git tag `v0.1.2` created and pushed
- [ ] **REL-02**: GitHub release created via goreleaser with platform binaries
- [ ] **REL-03**: `@a-tarek/ptt` published to npm registry

## Future Requirements

### Post-v0.1.2

- **COLOR-01**: Colored output with NO_COLOR support
- **CI-01**: Non-interactive mode flags for CI/scripting
- **ERR-01**: Enhanced error messages with suggestions

## Out of Scope

| Feature | Reason |
|---------|--------|
| Separate docs files | Everything fits on one README screen |
| Detailed troubleshooting | Discoverable via `--help` and error messages |
| eject/merge/rebase in README | Discoverable via `ptt --help`; keep README lean |
| Git history cleanup | Deferred; .planning already gitignored |

## Traceability

| Requirement | Phase | Status |
|-------------|-------|--------|
| README-01 | — | Pending |
| README-02 | — | Pending |
| README-03 | — | Pending |
| README-04 | — | Pending |
| README-05 | — | Pending |
| README-06 | — | Pending |
| REL-01 | — | Pending |
| REL-02 | — | Pending |
| REL-03 | — | Pending |

**Coverage:**
- v0.1.2 requirements: 9 total
- Mapped to phases: 0
- Unmapped: 9 ⚠️

---
*Requirements defined: 2026-02-09*
*Last updated: 2026-02-09 after initial definition*
