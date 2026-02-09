---
status: diagnosed
phase: 03-09-full
source: 03-01-SUMMARY.md, 03-02-SUMMARY.md, 04-01-SUMMARY.md, 04-02-SUMMARY.md, 05-01-SUMMARY.md, 05-02-SUMMARY.md, 05-03-SUMMARY.md, 06-01-SUMMARY.md, 06-02-SUMMARY.md, 07-01-SUMMARY.md, 07-02-SUMMARY.md, 08-01-SUMMARY.md, 08-02-SUMMARY.md, 09-01-SUMMARY.md, 09-02-SUMMARY.md, 09-03-SUMMARY.md, 09-04-SUMMARY.md
started: 2026-02-08T12:00:00Z
updated: 2026-02-08T13:30:00Z
---

## Current Test

[testing complete]

## Tests

### 1. Version Output
expected: Running `./wt-bin --version` prints a version string (e.g., "wt version dev")
result: pass

### 2. Help Shows All Commands
expected: Running `./wt-bin --help` lists all commands: list, init, delete, goto, home, new, eject, merge, rebase, install, uninstall, completion
result: pass

### 3. List Worktrees
expected: Running `./wt-bin list` shows your current worktree(s) with branch name, dirty/clean status (~), and current marker (*)
result: pass

### 4. Init Creates Config Template
expected: Running `./wt-bin init` in a git repo creates a .wtconfig file at the repo root with commented example actions (copy, symlink, run)
result: pass

### 5. Init Named Config
expected: Running `./wt-bin init --name ci` creates a .wtconfig-ci file at the repo root
result: pass

### 6. Delete Worktree Safety
expected: Running `./wt-bin delete` with no args shows an error. Running `./wt-bin delete nonexistent` shows "worktree not found" error with a suggestion if a similar name exists
result: pass

### 7. Shell Init Output
expected: Running `./wt-bin shell-init` outputs a shell wrapper function for your detected shell (bash/zsh). The output contains a `wt()` function definition
result: pass

### 8. Shell Wrapper CD
expected: After running `eval "$(./wt-bin shell-init)"`, the `wt` shell function is available. Running `wt list` works and shows the same output as `./wt-bin list`
result: issue
reported: "After eval, wt gives 'zsh: command not found: wt'. The wrapper defines wt() but uses 'command wt' internally which looks for a wt binary in PATH that doesn't exist."
severity: major

### 9. Goto Worktree
expected: If you have multiple worktrees, running `wt goto <name>` (with shell wrapper active) changes your shell directory to that worktree. If only one worktree, `./wt-bin goto nonexistent` shows a "not found" error
result: pass

### 10. Home Command
expected: Running `wt home` (with shell wrapper active) changes directory to your main/home worktree. If already there, prints "Already home" to stderr
result: pass

### 11. New Worktree Creation
expected: Running `wt new test-uat` creates a new worktree (sibling directory for regular repos, nested for bare repos) and changes directory to it. The worktree has a new branch named test-uat
result: issue
reported: "Worktree created but cd didn't happen - stayed in ~/code/wt. Error: wt:cd:7: no such file or directory: Symlinked wt-bin\n/Users/ahmed.tarek/code/wt-test-uat. Output path is garbled with stderr messages mixed in."
severity: major

### 12. Tab Completion Generation
expected: Running `./wt-bin completion zsh` outputs a zsh completion script (200+ lines). Same for `./wt-bin completion bash` and `./wt-bin completion fish`
result: pass

### 13. Fuzzy Match Error Suggestions
expected: Running `./wt-bin goto <misspelled-name>` (e.g., a name 1-2 chars off from an existing worktree) shows "Did you mean '<correct-name>'?" in the error message
result: issue
reported: "wt goto feat suggested 'wt' instead of 'wt-feat-1' or 'wt-feat-2'. Fuzzy matching picks wrong candidate when closer matches exist."
severity: major

### 14. Colored Error Output
expected: When running in a terminal (not piped), error messages from `./wt-bin goto nonexistent` display in red. When piped (`./wt-bin goto nonexistent 2>&1 | cat`), no color codes appear
result: pass

### 15. Install Command Preview
expected: Running `./wt-bin install` detects your shell, shows the RC file path, and previews what will be added (marker block with eval line). You can decline and get manual instructions
result: pass

### 16. Uninstall Command
expected: Running `./wt-bin uninstall` detects your shell and checks for existing wt integration in your RC file. If not installed, says "not configured"
result: pass

### 17. npm Package Structure
expected: The `npm/` directory contains: package.json (with @potato/wt name), bin/wt (Node.js wrapper script), and platforms/ with 4 subdirectories (darwin-arm64, darwin-amd64, linux-amd64, linux-arm64) each with their own package.json
result: pass

### 18. Build Scripts Exist
expected: `scripts/build-npm.sh` and `scripts/publish-npm.sh` exist and are executable. Running `bash scripts/build-npm.sh --help` or checking the file shows it handles goreleaser output staging
result: pass

### 19. CI/CD Workflows
expected: `.github/workflows/ci.yml` and `.github/workflows/release.yml` exist. CI runs tests on ubuntu+macos matrix. Release triggers on version tags
result: pass

### 20. README v2 Documentation
expected: README.md covers: installation via npm, all 9 commands (list, init, delete, goto, home, new, eject, merge, rebase), .wtconfig configuration, shell support, and troubleshooting
result: pass

### 21. All Tests Pass
expected: Running `go test ./...` passes all unit tests (config, setup, installer, shell E2E tests)
result: pass

## Summary

total: 21
passed: 18
issues: 3
pending: 0
skipped: 0

## Gaps

- truth: "After eval of shell-init output, wt shell function is available and wt list works"
  status: failed
  reason: "User reported: After eval, wt gives 'zsh: command not found: wt'. The wrapper defines wt() but uses 'command wt' internally which looks for a wt binary in PATH that doesn't exist."
  severity: major
  test: 8
  root_cause: "Wrapper templates hardcode 'command wt' instead of resolving actual binary path. shell_init.go does not call os.Executable(). embed.go has no placeholder replacement. Also: --output-path passed to all subcommands, not just cd-requiring ones."
  artifacts:
    - path: "internal/shell/templates/wrapper.zsh"
      issue: "Hardcoded 'command wt' instead of resolved binary path"
    - path: "internal/shell/templates/wrapper.bash"
      issue: "Same hardcoded binary name"
    - path: "internal/shell/templates/wrapper.fish"
      issue: "Same hardcoded binary name"
    - path: "internal/shell/embed.go"
      issue: "No placeholder replacement mechanism"
    - path: "cmd/shell_init.go"
      issue: "Does not resolve binary path via os.Executable()"
  missing:
    - "Templates need __WT_BIN__ placeholder"
    - "shell_init.go needs os.Executable() resolution"
    - "GetWrapper needs to accept and inject binary path"
    - "Templates should differentiate cd-requiring subcommands from pass-through"
  debug_session: ".planning/debug/shell-wrapper-cd.md"

- truth: "wt new creates worktree and changes directory to it"
  status: failed
  reason: "User reported: Worktree created but cd didn't happen - stayed in ~/code/wt. Error: wt:cd:7: no such file or directory: Symlinked wt-bin\\n/path. Output path garbled with stderr messages."
  severity: major
  test: 11
  root_cause: "internal/setup/executor.go prints progress messages (Copied/Symlinked/Running) to stdout via fmt.Printf(). Shell wrapper captures all stdout as the cd target path. Also run.go pipes run-action output to stdout."
  artifacts:
    - path: "internal/setup/executor.go"
      issue: "Lines 41, 49, 52: fmt.Printf should be fmt.Fprintf(os.Stderr)"
    - path: "internal/setup/run.go"
      issue: "Line 14: cmd.Stdout = os.Stdout should be cmd.Stdout = os.Stderr"
  missing:
    - "Change all fmt.Printf in executor.go to fmt.Fprintf(os.Stderr)"
    - "Change cmd.Stdout in run.go from os.Stdout to os.Stderr"
  debug_session: ".planning/debug/output-path-garbling.md"

- truth: "wt new --run 'cmd' executes inline run commands in new worktree"
  status: failed
  reason: "User reported: --run flag not available on wt new. BuildActionsFromFlags supports it but flag never registered."
  severity: major
  test: N/A
  root_cause: "cmd/new.go and cmd/eject.go pass nil for runCommands parameter to BuildActionsFromFlags. No --run flag registered in init()."
  artifacts:
    - path: "cmd/new.go"
      issue: "No runFlags var, no --run flag in init(), nil passed to BuildActionsFromFlags"
    - path: "cmd/eject.go"
      issue: "Same missing --run flag"
  missing:
    - "Register --run StringSliceVar flag on both commands"
    - "Wire runFlags into BuildActionsFromFlags call"
  debug_session: ""

- truth: "Fuzzy match suggests closest worktree name when misspelled"
  status: failed
  reason: "User reported: wt goto feat suggested 'wt' instead of 'wt-feat-1' or 'wt-feat-2'. Fuzzy matching picks wrong candidate when closer matches exist."
  severity: major
  test: 13
  root_cause: "findClosestMatch() in internal/git/resolve.go uses pure Levenshtein with maxDistance=3. Worktree basenames like wt-feat-1 have distance 5 from 'feat' (exceeds threshold) while 'wt' has distance 3 (within threshold). No substring awareness."
  artifacts:
    - path: "internal/git/resolve.go"
      issue: "Lines 62-78: Pure Levenshtein with hard maxDistance=3, no substring matching"
  missing:
    - "Add substring/segment-aware matching (split on -, compare segments)"
    - "Give bonus to candidates containing input as substring"
    - "Or raise threshold to account for repo-prefix pattern"
  debug_session: ".planning/debug/fuzzy-match.md"
