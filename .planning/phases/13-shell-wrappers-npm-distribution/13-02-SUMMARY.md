---
phase: 13-shell-wrappers-npm-distribution
plan: 02
subsystem: distribution
tags: [npm, nodejs, package-management, branding]

# Dependency graph
requires:
  - phase: 11-go-module-binary-rename
    provides: Binary renamed from wt to ptt, Go module path updated
provides:
  - npm package @a-tarek/ptt with platform-specific optional dependencies
  - Platform packages @a-tarek/ptt-{platform}-{arch} for all 4 supported platforms
  - Node.js binary wrapper resolving platform-specific binaries
  - Repository URLs pointing to github:a-tarek/ptt
affects: [14-shell-integration-rebrand, release, documentation]

# Tech tracking
tech-stack:
  added: []
  patterns:
    - "npm scoped packages with @a-tarek scope for personal distribution"
    - "Platform-specific optional dependencies for multi-platform binaries"

key-files:
  created:
    - npm/bin/ptt
  modified:
    - npm/package.json
    - npm/platforms/darwin-arm64/package.json
    - npm/platforms/darwin-amd64/package.json
    - npm/platforms/linux-amd64/package.json
    - npm/platforms/linux-arm64/package.json

key-decisions:
  - "@a-tarek/ptt npm scope ensures package name availability and personal ownership"
  - "Platform packages use amd64 suffix (Go convention) while Node.js arch keys use x64"
  - "Binary wrapper file renamed to match bin field (npm/bin/ptt)"

patterns-established:
  - "Scoped npm packages with platform-specific optional dependencies pattern"
  - "Node.js binary wrapper resolving native platform binaries via require.resolve"

# Metrics
duration: 1.8min
completed: 2026-02-08
---

# Phase 13 Plan 02: npm Package Rebrand Summary

**npm distribution rebranded from @potato/wt to @a-tarek/ptt with updated package names, bin fields, and binary wrapper**

## Performance

- **Duration:** 1.8 min (107 seconds)
- **Started:** 2026-02-08T21:21:04Z
- **Completed:** 2026-02-08T21:22:51Z
- **Tasks:** 2
- **Files modified:** 6

## Accomplishments
- Main npm package rebranded to @a-tarek/ptt with ptt bin field
- All 4 platform packages rebranded to @a-tarek/ptt-{platform}-{arch} scope
- Binary wrapper renamed from npm/bin/wt to npm/bin/ptt with updated platform package references
- Repository URLs updated to github:a-tarek/ptt across all packages
- Descriptions updated to reference "ptt: a potato worktree manager" brand

## Task Commits

Each task was committed atomically:

1. **Task 1: Rebrand npm package.json files** - `eaadcff` (feat)
2. **Task 2: Rename and rebrand npm binary wrapper** - `2565991` (feat)

**Plan metadata:** (pending at completion)

## Files Created/Modified
- `npm/package.json` - Main package definition with @a-tarek/ptt name and ptt bin field
- `npm/platforms/darwin-arm64/package.json` - macOS ARM64 platform package
- `npm/platforms/darwin-amd64/package.json` - macOS x64 platform package
- `npm/platforms/linux-amd64/package.json` - Linux x64 platform package
- `npm/platforms/linux-arm64/package.json` - Linux ARM64 platform package
- `npm/bin/ptt` - Node.js binary wrapper (renamed from npm/bin/wt)

## Decisions Made

**Platform package architecture naming:**
- Platform packages use `amd64` suffix (e.g., @a-tarek/ptt-darwin-amd64) matching Go's GOARCH convention
- Node.js platform map keys still use `x64` (e.g., 'darwin-x64') matching process.arch
- This dual convention is correct and intentional - the map bridges Node.js naming to Go package naming

**Binary wrapper renaming:**
- Binary wrapper file renamed from npm/bin/wt to npm/bin/ptt to match bin field
- Maintains consistency between package.json bin field and actual file path
- Simplifies debugging and package structure understanding

**Version preservation:**
- Kept version at 0.0.0 in all packages
- scripts/publish-npm.sh handles version updates at publish time
- Prevents version drift between manual edits and publish automation

## Deviations from Plan

None - plan executed exactly as written.

## Issues Encountered

None - all files updated cleanly, verification checks passed, and no conflicts with existing code.

## User Setup Required

None - no external service configuration required.

## Next Phase Readiness

**Ready for Phase 14 (Shell Integration Rebrand):**
- npm packages fully rebranded to @a-tarek/ptt scope
- Binary wrapper correctly references platform packages
- Build scripts (scripts/build-npm.sh, scripts/publish-npm.sh) already reference ptt paths
- Release workflow (.github/workflows/release.yml) already references ptt paths
- No blockers for shell integration rebranding

**Pre-release readiness:**
- npm distribution structure complete
- Package names secured (@a-tarek scope)
- Ready for first npm publish after shell integration rebrand completes

**Note:** Old compiled Go binaries still exist in npm/platforms/*/bin/wt directories. These are artifacts from previous builds and will be replaced when scripts/build-npm.sh runs, which copies ptt binaries to bin/ptt paths. This is expected and not a blocker.

## Self-Check: PASSED

All key-files and commits verified:
- All 6 files exist (5 package.json + 1 binary wrapper)
- Both commits present in git history (eaadcff, 2565991)

---
*Phase: 13-shell-wrappers-npm-distribution*
*Completed: 2026-02-08*
