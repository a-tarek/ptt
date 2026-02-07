# Phase 9: Polish & Testing - Context

**Gathered:** 2026-02-07
**Status:** Ready for planning

<domain>
## Phase Boundary

Improve error messages with actionable suggestions, add CI/CD pipeline with GitHub Actions, comprehensive shell wrapper testing across bash/zsh/fish versions, and rewrite README for v2. No new features — hardening and documentation for what's already built.

</domain>

<decisions>
## Implementation Decisions

### Error message style
- Single-line format with inline hint: `Error: worktree 'foo' not found. Did you mean 'foo-bar'?`
- Color output with auto-detection: red for errors, yellow for warnings; plain when piped or NO_COLOR is set
- Fuzzy match suggestions on every not-found error (always suggest closest match, like git)
- Every error ends with `Run 'wt help <cmd>' for details` footer — discoverable for new users

### CI/CD pipeline
- GitHub Actions as CI platform
- OS matrix: Linux (ubuntu-latest) + macOS (macos-latest) — no Windows
- Full release pipeline: tag push triggers goreleaser build + npm publish
- Track test coverage in CI output but don't enforce a minimum threshold (no fail on drops)

### Shell testing strategy
- Full end-to-end shell tests: source wrapper, run command, verify directory actually changed
- Real git repos in tests: create temporary bare repos and worktrees, no mocking git operations
- Test actual cd behavior in real shell sessions for high confidence

### Claude's Discretion
- Shell version tier split: which versions are must-pass vs best-effort (bash 3.2, bash 5+, zsh 5.8+, fish 3.0+)
- Shell testing infrastructure: Docker containers vs native CI shells vs hybrid approach
- Troubleshooting in README: dedicated section vs inline tips vs skip entirely

### README
- Thorough walkthrough style — more explanation, use cases, and context (like zoxide or starship)
- No v1-to-v2 migration guide — v2 is documented as a fresh standalone tool
- Rewrite from scratch — clean slate, no v1 artifacts carried over

</decisions>

<specifics>
## Specific Ideas

- Error format should feel like git's error messages: compact but helpful
- README tone like zoxide or starship — thorough but not overwhelming
- Release pipeline should be one-command: push a tag and everything builds + publishes

</specifics>

<deferred>
## Deferred Ideas

None — discussion stayed within phase scope

</deferred>

---

*Phase: 09-polish-testing*
*Context gathered: 2026-02-07*
