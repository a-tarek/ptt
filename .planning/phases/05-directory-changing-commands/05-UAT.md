---
status: complete
phase: 05-directory-changing-commands
source: [05-01-SUMMARY.md, 05-02-SUMMARY.md, 05-03-SUMMARY.md]
started: 2026-02-07T23:00:00Z
updated: 2026-02-07T23:10:00Z
---

## Current Test

[testing complete]

## Tests

### 1. wt goto switches to worktree
expected: Run `wt goto <worktree-name>` (or a suffix match like just "staging" for "repo-staging"). Binary outputs target path and confirmation to stderr showing branch name and dirty status.
result: pass
note: Binary outputs correctly. cd requires shell wrapper (Phase 6).

### 2. wt goto already-there is a no-op
expected: Run `wt goto <name>` when already in that worktree. Prints "Already in <name>" to stderr and exits 0 (no error).
result: pass

### 3. wt home switches to main worktree
expected: Run `wt home`. Binary outputs the home worktree path. Confirmation to stderr shows branch name (or "(bare)" for bare repos). Already-there case prints "Already home".
result: pass

### 4. wt merge resolves worktree to branch
expected: Run `wt merge <worktree-name>`. Resolves the worktree name to its branch, then runs `git merge <branch>`. Shows "Merging <branch> into current branch..." to stderr.
result: skipped
reason: Requires directory changing (shell wrapper from Phase 6) for proper manual testing

### 5. wt rebase resolves worktree to branch
expected: Run `wt rebase <worktree-name>`. Resolves the worktree name to its branch, then runs `git rebase <branch>`. Shows "Rebasing onto <branch>..." to stderr.
result: skipped
reason: Requires directory changing (shell wrapper from Phase 6) for proper manual testing

### 6. wt new creates worktree with correct placement
expected: Run `wt new <name>`. Creates a new worktree — nested under bare repo root (e.g., `repo.git/name/`) or as sibling for regular repos (e.g., `repo-name/`). Creates a new branch named `<name>`. Prints "Created worktree <name> (branch: <name>)" to stderr.
result: skipped
reason: Requires directory changing for proper manual testing; covered by automated tests

### 7. wt new applies .wtconfig automatically
expected: With a .wtconfig file at repo root, run `wt new <name>`. Config actions (copy/symlink/run) are applied to the new worktree. Shows "Applied .wtconfig (N actions)" to stderr. Missing .wtconfig is silently skipped.
result: skipped
reason: Requires directory changing for proper manual testing; covered by automated tests

### 8. wt new config flag overrides
expected: `--config ci` loads `.wtconfig-ci` instead of `.wtconfig`. `--skip-config` skips file-based config. `--copy`/`--symlink` flags merge on top and apply even with `--skip-config`.
result: skipped
reason: Requires directory changing for proper manual testing; covered by automated tests

### 9. wt new existing branch reuse
expected: Run `wt new <name>` where `<name>` matches an existing branch. Instead of creating a new branch, checks out the existing branch into the new worktree.
result: skipped
reason: Requires directory changing for proper manual testing; covered by automated tests

### 10. wt eject moves current branch to new worktree
expected: Run `wt eject <name>` from a worktree. Stashes any dirty changes, switches current worktree to a fallback branch (main/master for home, directory-suffix for others), creates new worktree for the ejected branch, and pops stash in the new worktree.
result: skipped
reason: Requires directory changing for proper manual testing; covered by automated tests

### 11. wt eject fallback branch detection
expected: From home worktree, eject uses main or master as fallback. From non-home worktree (e.g., `repo-staging`), eject uses the directory name suffix (e.g., "staging") as fallback branch.
result: skipped
reason: Requires directory changing for proper manual testing; covered by automated tests

### 12. wt eject stash conflict handling
expected: If stash pop has merge conflicts, eject prints a warning but doesn't fail. The stash was popped (with conflicts), and the user can resolve.
result: skipped
reason: Requires directory changing for proper manual testing; covered by automated tests

### 13. Tests pass
expected: Run `go test ./...` from the project root. All tests pass, including the new Phase 5 tests for goto, home, merge, rebase, new, and eject.
result: pass

## Summary

total: 13
passed: 4
issues: 0
pending: 0
skipped: 9

## Gaps

[none yet]
