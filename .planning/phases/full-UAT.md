---
status: testing
phase: 03-09-full
source: 03-01-SUMMARY.md, 03-02-SUMMARY.md, 04-01-SUMMARY.md, 04-02-SUMMARY.md, 05-01-SUMMARY.md, 05-02-SUMMARY.md, 05-03-SUMMARY.md, 06-01-SUMMARY.md, 06-02-SUMMARY.md, 07-01-SUMMARY.md, 07-02-SUMMARY.md, 08-01-SUMMARY.md, 08-02-SUMMARY.md, 09-01-SUMMARY.md, 09-02-SUMMARY.md, 09-03-SUMMARY.md, 09-04-SUMMARY.md
started: 2026-02-08T12:00:00Z
updated: 2026-02-08T12:00:00Z
---

## Current Test
<!-- OVERWRITE each test - shows where we are -->

number: 1
name: Version Output
expected: |
  Running `./wt-bin --version` prints a version string (e.g., "wt version dev")
awaiting: user response

## Tests

### 1. Version Output
expected: Running `./wt-bin --version` prints a version string (e.g., "wt version dev")
result: [pending]

### 2. Help Shows All Commands
expected: Running `./wt-bin --help` lists all commands: list, init, delete, goto, home, new, eject, merge, rebase, install, uninstall, completion
result: [pending]

### 3. List Worktrees
expected: Running `./wt-bin list` shows your current worktree(s) with branch name, dirty/clean status (~), and current marker (*)
result: [pending]

### 4. Init Creates Config Template
expected: Running `./wt-bin init` in a git repo creates a .wtconfig file at the repo root with commented example actions (copy, symlink, run)
result: [pending]

### 5. Init Named Config
expected: Running `./wt-bin init --name ci` creates a .wtconfig-ci file at the repo root
result: [pending]

### 6. Delete Worktree Safety
expected: Running `./wt-bin delete` with no args shows an error. Running `./wt-bin delete nonexistent` shows "worktree not found" error with a suggestion if a similar name exists
result: [pending]

### 7. Shell Init Output
expected: Running `./wt-bin shell-init` outputs a shell wrapper function for your detected shell (bash/zsh). The output contains a `wt()` function definition
result: [pending]

### 8. Shell Wrapper CD
expected: After running `eval "$(./wt-bin shell-init)"`, the `wt` shell function is available. Running `wt list` works and shows the same output as `./wt-bin list`
result: [pending]

### 9. Goto Worktree
expected: If you have multiple worktrees, running `wt goto <name>` (with shell wrapper active) changes your shell directory to that worktree. If only one worktree, `./wt-bin goto nonexistent` shows a "not found" error
result: [pending]

### 10. Home Command
expected: Running `wt home` (with shell wrapper active) changes directory to your main/home worktree. If already there, prints "Already home" to stderr
result: [pending]

### 11. New Worktree Creation
expected: Running `wt new test-uat` creates a new worktree (sibling directory for regular repos, nested for bare repos) and changes directory to it. The worktree has a new branch named test-uat
result: [pending]

### 12. Tab Completion Generation
expected: Running `./wt-bin completion zsh` outputs a zsh completion script (200+ lines). Same for `./wt-bin completion bash` and `./wt-bin completion fish`
result: [pending]

### 13. Fuzzy Match Error Suggestions
expected: Running `./wt-bin goto <misspelled-name>` (e.g., a name 1-2 chars off from an existing worktree) shows "Did you mean '<correct-name>'?" in the error message
result: [pending]

### 14. Colored Error Output
expected: When running in a terminal (not piped), error messages from `./wt-bin goto nonexistent` display in red. When piped (`./wt-bin goto nonexistent 2>&1 | cat`), no color codes appear
result: [pending]

### 15. Install Command Preview
expected: Running `./wt-bin install` detects your shell, shows the RC file path, and previews what will be added (marker block with eval line). You can decline and get manual instructions
result: [pending]

### 16. Uninstall Command
expected: Running `./wt-bin uninstall` detects your shell and checks for existing wt integration in your RC file. If not installed, says "not configured"
result: [pending]

### 17. npm Package Structure
expected: The `npm/` directory contains: package.json (with @potato/wt name), bin/wt (Node.js wrapper script), and platforms/ with 4 subdirectories (darwin-arm64, darwin-amd64, linux-amd64, linux-arm64) each with their own package.json
result: [pending]

### 18. Build Scripts Exist
expected: `scripts/build-npm.sh` and `scripts/publish-npm.sh` exist and are executable. Running `bash scripts/build-npm.sh --help` or checking the file shows it handles goreleaser output staging
result: [pending]

### 19. CI/CD Workflows
expected: `.github/workflows/ci.yml` and `.github/workflows/release.yml` exist. CI runs tests on ubuntu+macos matrix. Release triggers on version tags
result: [pending]

### 20. README v2 Documentation
expected: README.md covers: installation via npm, all 9 commands (list, init, delete, goto, home, new, eject, merge, rebase), .wtconfig configuration, shell support, and troubleshooting
result: [pending]

### 21. All Tests Pass
expected: Running `go test ./...` passes all unit tests (config, setup, installer, shell E2E tests)
result: [pending]

## Summary

total: 21
passed: 0
issues: 0
pending: 21
skipped: 0

## Gaps

[none yet]
