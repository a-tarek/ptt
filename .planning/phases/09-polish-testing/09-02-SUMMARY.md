---
phase: 09
plan: 02
subsystem: ci-cd
status: complete
tags: [github-actions, ci, cd, testing, automation, goreleaser, npm-publish]
dependencies:
  requires: [07-02, 08-02]
  provides: [automated-ci, automated-release]
  affects: []
tech-stack:
  added: [github-actions]
  patterns: [ci-matrix-testing, tag-triggered-release]
key-files:
  created:
    - .github/workflows/ci.yml
    - .github/workflows/release.yml
  modified: []
decisions:
  - id: ci-matrix-os
    choice: ubuntu-latest and macos-latest only (no Windows)
    rationale: Target platforms for wt; Windows support deferred
  - id: coverage-tracking
    choice: Display coverage in output but don't enforce threshold
    rationale: Track metrics without failing builds on coverage drops
  - id: go-version-strategy
    choice: Use exact version from go.mod (1.25.7)
    rationale: Ensures CI uses same version as development
  - id: no-caching
    choice: No explicit Go module caching
    rationale: Built-in caching in actions/setup-go is sufficient for small project
  - id: lean-ci
    choice: Only go vet, no golangci-lint
    rationale: Keep CI simple and fast
  - id: npm-publish-automation
    choice: Release workflow stages binaries and runs publish-npm.sh
    rationale: Single git tag push triggers complete release pipeline
metrics:
  duration: 1.6 min
  completed: 2026-02-08
---

# Phase 9 Plan 02: GitHub Actions CI/CD Summary

**One-liner:** Automated CI with Linux/macOS matrix testing and tag-triggered release pipeline using goreleaser + npm publish

## What Was Built

### CI Pipeline (.github/workflows/ci.yml)

**Trigger:** Push to any branch, pull requests to main

**Matrix Testing:**
- OS: ubuntu-latest, macos-latest
- Go version: 1.25.7 (from go.mod)

**Steps:**
1. Checkout code
2. Set up Go
3. Run `go vet ./...` for static analysis
4. Run `go test -v -coverprofile=coverage.out ./...`
5. Display coverage summary with `go tool cover -func=coverage.out`

**Philosophy:** Keep it lean - no caching, no linting beyond go vet, just fast and simple.

### Release Pipeline (.github/workflows/release.yml)

**Trigger:** Push tags matching `v*` (e.g., v2.0.0, v2.1.0-beta)

**Two-stage process:**

**Stage 1: Test Job** (runs in parallel on Linux + macOS)
- Runs full test suite before allowing release
- Ensures broken builds never get released

**Stage 2: Release Job** (runs on ubuntu-latest after tests pass)
1. Checkout with full git history (fetch-depth: 0 for goreleaser)
2. Set up Go and Node.js
3. Install goreleaser
4. Run `goreleaser release --clean` (builds all platform binaries)
5. Stage npm binaries:
   - Copy from `dist/wt_darwin_arm64_v8.0/wt` to `npm/platforms/darwin-arm64/bin/wt`
   - Copy from `dist/wt_darwin_amd64_v1/wt` to `npm/platforms/darwin-amd64/bin/wt`
   - Copy from `dist/wt_linux_arm64_v8.0/wt` to `npm/platforms/linux-arm64/bin/wt`
   - Copy from `dist/wt_linux_amd64_v1/wt` to `npm/platforms/linux-amd64/bin/wt`
   - Make all binaries executable
6. Extract version from tag: `VERSION=${GITHUB_REF_NAME#v}` (strips leading 'v')
7. Run `bash scripts/publish-npm.sh $VERSION` to publish all 5 npm packages

**Secrets Required:**
- `NPM_TOKEN`: npm automation token for publishing @potato/* packages
  - Create at: https://www.npmjs.com/settings/username/tokens
  - Add to: Repository Settings → Secrets and variables → Actions

**Permissions:**
- `contents: write` for goreleaser to create GitHub releases

## Task Commits

| Task | Description | Commit | Files |
|------|-------------|--------|-------|
| 1 | CI workflow for testing on push and PR | b277047 | .github/workflows/ci.yml |
| 2 | Release workflow triggered by version tags | 2565734 | .github/workflows/release.yml |

## Decisions Made

**CI Matrix Strategy:**
- **Decision:** Test on ubuntu-latest and macos-latest only
- **Context:** These are the target platforms for wt distribution
- **Impact:** Windows support can be added later if needed

**Coverage Policy:**
- **Decision:** Track coverage in CI output but don't enforce thresholds
- **Context:** User wants visibility without build failures on coverage drops
- **Impact:** Coverage visible in logs, can add enforcement later if desired

**Go Version Pinning:**
- **Decision:** Use exact version from go.mod (1.25.7)
- **Context:** Ensures CI environment matches development environment
- **Alternative considered:** Use `stable` for latest Go version
- **Impact:** More predictable builds, explicit when upgrading Go version

**Lean CI Philosophy:**
- **Decision:** Only use go vet, skip golangci-lint and other linters
- **Context:** Keep CI fast and simple for a small project
- **Impact:** ~30 second faster CI runs, can add more linting later if needed

**Single-Command Release:**
- **Decision:** Full automation on tag push
- **Context:** Push tag → tests run → build → npm publish, all automatic
- **Impact:** Releasing v2.0.0 is just `git tag v2.0.0 && git push origin v2.0.0`

## Key Integrations

**CI → Tests:**
- Every push runs full test suite on target platforms
- PR checks prevent merging broken code

**Release → goreleaser:**
- goreleaser builds all 4 platform binaries (darwin-arm64, darwin-amd64, linux-arm64, linux-amd64)
- Configured in .goreleaser.yaml (created in phase 7)

**Release → npm:**
- Binaries staged to npm/platforms/*/bin/ after goreleaser
- scripts/publish-npm.sh updates versions and publishes all 5 packages
- Platform packages published first, then main @potato/wt package

**Tag → Version Extraction:**
- `${GITHUB_REF_NAME#v}` strips 'v' prefix (v2.0.0 → 2.0.0)
- Version passed to publish-npm.sh for package.json updates

## Deviations from Plan

None - plan executed exactly as written.

## Next Phase Readiness

**Blockers:** None

**Concerns:**
1. **NPM_TOKEN secret:** Must be configured in repository settings before first release
   - Without it, release workflow will fail at npm publish step
   - Documentation included in workflow comments

2. **First release test:** Recommend testing with a beta tag (v2.0.0-beta.1) before v2.0.0
   - Verifies full pipeline works
   - Can unpublish beta versions if something goes wrong

**Opportunities:**
- CI runs on every push - developers get immediate feedback
- Release process is fully automated - no manual steps beyond pushing a tag
- Matrix testing catches platform-specific issues early

## Files Created

### .github/workflows/ci.yml
Continuous integration workflow with matrix testing on Linux and macOS. Runs go vet and go test with coverage on every push and PR.

### .github/workflows/release.yml
Automated release pipeline triggered by version tags. Tests on both platforms, builds with goreleaser, stages npm binaries, and publishes all 5 packages to npm.

## How to Use

### Running CI
CI runs automatically on every push and PR. No manual action needed.

**Local testing before push:**
```bash
go vet ./...
go test -v -coverprofile=coverage.out ./...
go tool cover -func=coverage.out
```

### Releasing a New Version

**1. Tag the release:**
```bash
git tag v2.0.0
git push origin v2.0.0
```

**2. Watch the release:**
- GitHub Actions will automatically run tests, build, and publish
- Check: Repository → Actions → Release workflow

**3. Verify publication:**
```bash
npm view @potato/wt versions
```

**First-time setup:**
- Add NPM_TOKEN secret to repository (Settings → Secrets and variables → Actions)
- Recommended: Test with beta tag first (v2.0.0-beta.1)

## Testing Notes

**CI Workflow:**
- Tested YAML syntax (valid)
- Verified matrix configuration (ubuntu-latest, macos-latest)
- Confirmed go vet and go test steps present
- Coverage output configured correctly

**Release Workflow:**
- Tested YAML syntax (valid)
- Verified tag trigger pattern (`v*`)
- Confirmed test job dependency (release needs test)
- Verified NPM_TOKEN secret reference
- Confirmed goreleaser integration
- Verified binary staging paths match goreleaser output

**Integration Points:**
- goreleaser config exists (.goreleaser.yaml)
- publish-npm.sh script exists and accepts VERSION argument
- npm/ directory structure matches binary staging paths

## Self-Check: PASSED

All files created:
- .github/workflows/ci.yml ✓
- .github/workflows/release.yml ✓

All commits exist:
- b277047 ✓
- 2565734 ✓
