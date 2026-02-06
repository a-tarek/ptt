# Testing Patterns

**Analysis Date:** 2026-02-07

## Test Framework

**Current State:**
- No testing framework detected
- No test files found in codebase
- No test configuration files (jest.config.js, vitest.config.ts, etc.)
- No testing dependencies in package.json (no package.json present)

## Code Testing Approach

**Manual Testing:**
The codebase appears designed for manual testing through zsh shell execution. The tool is a shell script intended for interactive use.

**Testing Surface:**
The following operations should be manually validated:

**Basic Commands:**
- `wt new <name> [branch]` - Create worktree with default or specified branch
- `wt new --copy-node-modules <name>` - Create worktree with node_modules copied
- `wt goto <worktree>` - Navigate to worktree
- `wt home` - Return to main worktree
- `wt list` - List all worktrees
- `wt delete <worktree>` - Remove worktree
- `wt merge <worktree>` - Merge worktree branch
- `wt rebase <worktree>` - Rebase onto worktree branch
- `wt eject [name]` - Eject current branch to new worktree

**State-Based Testing:**
- Pre-conditions: Repository state, existing worktrees, branch existence
- Post-conditions: Directory creation, branch checkouts, file copies
- Recovery: Stash management, rollback on failure

## Manual Test Scenarios

**New Worktree Creation:**
1. Initialize git repo with multiple branches
2. Run `wt new feature-x`
3. Verify: New directory created, correct branch checked out
4. Test with `--copy-node-modules` flag
5. Verify: node_modules copied, symlink when flag omitted

**Navigation:**
1. From main worktree: `wt goto feature-x`
2. Verify: Current directory changed to worktree
3. Test with non-existent worktree
4. Verify: Error message displayed

**Eject Functionality (Complex):**
1. Create worktree with topic branch
2. Switch back to main, create different branch
3. `wt eject` in new branch
4. Verify: New worktree created, original falls back, stash preserved
5. Test on detached HEAD - should error
6. Test on main worktree fallback logic

**List Command:**
1. Create multiple worktrees
2. Run `wt list`
3. Verify: Current worktree marked with `*`
4. Verify: All worktrees listed with branches

**Merge/Rebase:**
1. Create worktree with changes
2. From another worktree: `wt merge <name>`
3. Verify: Git merge executed with correct branch
4. Test rebase similarly

**Error Conditions:**
1. Run commands outside git repository - expect error
2. Try to create worktree that already exists - expect error
3. Try to delete non-existent worktree - expect error
4. Run eject with detached HEAD - expect error

## Testing Considerations

**Git State Verification:**
Tests should verify git command execution by checking:
- Working directory after operation
- Branch state: `git branch --show-current`
- Worktree list output: `git worktree list --porcelain`
- Stash state: `git stash list`

**File System Verification:**
- Directory existence checks
- Symlink resolution (for node_modules)
- File copy success (.env.local)

**Error Path Testing:**
- Invalid repository
- Missing required arguments
- Conflicting branch states
- Permission errors (if possible)

## Code Inspection Points

**Critical Paths:**
`_wt_eject` (function at lines 129-258 in `/Users/ahmed.tarek/code/wt/wt.zsh`):
- Most complex function with multi-step state management
- Stash tracking: Needs validation that stash is correctly counted before/after
- Rollback logic: Verify git checkout failure triggers stash pop
- Branch fallback logic: Test main vs non-main worktree fallback paths

`_wt_new` (function at lines 34-100 in `/Users/ahmed.tarek/code/wt/wt.zsh`):
- Branch creation logic: Test with existing vs new branches
- node_modules handling: Copy vs symlink paths
- .env.local propagation: Verify file is copied when present

`_wt_resolve_path` and `_wt_resolve_branch` (lines 352-384 in `/Users/ahmed.tarek/code/wt/wt.zsh`):
- Matching logic: Test fuzzy matching (suffix matching)
- Edge cases: Worktree name matches multiple repos
- Output parsing: Verify correct branch extraction from porcelain format

**Input Validation:**
- Empty/missing arguments handling
- Special characters in branch/worktree names
- Path traversal safety (not a concern for worktree paths)

## Regression Test Checklist

When making changes to `/Users/ahmed.tarek/code/wt/wt.zsh`, manually verify:

1. Basic worktree creation works
2. Navigation between worktrees works
3. Eject handles stash correctly
4. List shows current worktree marker
5. Delete removes worktree without branch
6. Merge/rebase target correct branches
7. Error messages display for invalid input
8. Completion suggestions work in zsh

## Integration Points to Test

**With Git:**
- Git command execution (all commands silenced with `2>/dev/null`)
- Porcelain format parsing for `worktree list`
- Branch reference validation

**With File System:**
- Directory creation via `git worktree add`
- Symlink creation for node_modules
- File copy for .env.local
- Directory removal via `git worktree remove`

**With Zsh:**
- Function scoping (local variables)
- Completion system (`compdef` registration)
- Command substitution
- Conditional execution

## Notes on Testing This Codebase

**Why No Framework:**
This is a thin shell wrapper around git worktrees. A testing framework would add significant overhead for a ~450-line script. Manual testing of git commands is more practical.

**Recommended Additions (If Scaling):**
If this grows significantly:
1. Consider BATS (Bash Automated Testing System) for shell scripts
2. Create fixtures (test repos) in temp directories
3. Mock git commands if testing error paths becomes complex
4. Use shellcheck for static analysis

**Current Quality Assurance:**
- Manual testing by users
- Code review of logic (especially eject, merge, rebase)
- Zsh completion system catches many parameter errors

---

*Testing analysis: 2026-02-07*
