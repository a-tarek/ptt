# Codebase Concerns

**Analysis Date:** 2026-02-07

## Tech Debt

**Subshell/Piping Data Loss in Helper Functions:**
- Issue: Functions `_wt_resolve_path()` and `_wt_resolve_branch()` use piped `while` loops (lines 354, 370), which execute in subshells. Return values from within subshells don't propagate to the caller, and local variable modifications are lost.
- Files: `wt.zsh` (lines 352-384)
- Impact: While the `echo` output is captured correctly, any variable state changes or return codes from within the subshell are discarded. This limits future refactoring and makes the code fragile.
- Fix approach: Replace piped `while` loops with `while` loops that read from process substitution (`while ... < <(...)`) or use array iteration. This keeps execution in the parent shell context.

**Duplicated Logic Across Helper Functions:**
- Issue: `_wt_resolve_path()` (lines 352-364) and `_wt_resolve_branch()` (lines 367-384) both parse `git worktree list` output with nearly identical matching logic. When the matching algorithm changes (e.g., worktree naming convention), both must be updated.
- Files: `wt.zsh` (lines 352-384)
- Impact: Maintenance burden increases with each feature that uses worktree resolution. Bug fixes in one function may not be applied to the other.
- Fix approach: Extract common worktree-matching logic into a single helper function that returns matching worktree entry and branch, then have `_wt_resolve_path()` and `_wt_resolve_branch()` call it to extract their respective values.

**Inconsistent Error Exit Codes:**
- Issue: Throughout the codebase, functions return `1` on error, but return codes from failed git commands are suppressed with `&>/dev/null` (9 instances at lines 70, 72, 204, 214, 218, 225, 228, 230, 238). When operations fail silently, callers can't distinguish between different failure modes.
- Files: `wt.zsh` (lines 34-446)
- Impact: Debugging is harder when error messages are lost. Users see generic "Error: failed to..." messages without the underlying git error details.
- Fix approach: Capture stderr separately and log it on failure. Reserve `&>/dev/null` for expected failures only (like checking branch existence).

## Known Bugs

**Stash State Race Condition in `_wt_eject`:**
- Symptoms: If multiple `wt eject` commands run on the same repository simultaneously in different worktrees, stash counting (lines 203-206) can give false positives due to shared git stash state.
- Files: `wt.zsh` (lines 200-211)
- Trigger: Run `wt eject` in two worktrees at nearly the same time against the same repo
- Workaround: Ensure sequential execution of `wt eject`; don't parallelize across worktrees of the same repo

**Detached HEAD Stash Pop Failure Not Handled:**
- Symptoms: If stash pop fails during `_wt_eject` (line 238), the error is silenced but uncommitted changes are not restored to the user's working directory.
- Files: `wt.zsh` (lines 236-240)
- Trigger: Create a stash conflict scenario during eject (e.g., uncommitted changes conflict with restored branch state)
- Workaround: Manually inspect git stash list and pop the stash manually with conflict resolution

**Variable Scope Leak in Parsing Loop:**
- Symptoms: In `_wt_list()` (lines 270-284), variables `wt_entry` and `branch` persist across loop iterations in the subshell context. While `wt_entry=""` is explicitly reset (line 281), this relies on explicit cleanup that could be forgotten in future modifications.
- Files: `wt.zsh` (lines 260-292)
- Trigger: Any refactoring of the output parsing logic without clearing variables
- Workaround: Carefully review loop variable initialization on each iteration

## Security Considerations

**Unvalidated Paths in File Operations:**
- Risk: `cp` and `ln -s` operations use paths constructed from git-derived strings without validation. A maliciously crafted `.env.local` filename or worktree path containing special characters could lead to unintended file operations.
- Files: `wt.zsh` (lines 78-94, 242-251)
- Current mitigation: Paths use quoted variables and come from git, which validates directory names
- Recommendations: Add explicit path validation before `cp` and `ln -s` operations. Consider using safer alternatives like `install` for copying or absolute path construction with normalization.

**Environment Variable Exposure:**
- Risk: `.env.local` is copied verbatim from source to target worktree without filtering. Secrets intended for a specific worktree context may leak to sibling worktrees.
- Files: `wt.zsh` (lines 78-82, 242-246)
- Current mitigation: File exists only if user explicitly created it
- Recommendations: Document that `.env.local` should not contain environment-specific secrets. Consider adding a `--no-env` flag to skip env copying for sensitive use cases.

**Symlinked node_modules Shared State:**
- Risk: When `node_modules` is symlinked between worktrees (default behavior, line 91, 249), package installation in one worktree can corrupt or delete packages needed by the other. Worktrees with different dependency versions will conflict.
- Files: `wt.zsh` (lines 85-94, 248-251)
- Current mitigation: `--copy-node-modules` flag available as opt-in alternative
- Recommendations: Consider defaulting to copy instead of symlink, or add a `--shared-modules` flag to explicitly opt-in to symlinks. Document the risks clearly in help text.

## Performance Bottlenecks

**`cp -r node_modules` on Large Monorepos:**
- Problem: Copying `node_modules` recursively (line 88) with `--copy-node-modules` flag is extremely slow for large projects (> 10GB). No progress indicator or cancellation mechanism.
- Files: `wt.zsh` (lines 85-94)
- Cause: Shell `cp -r` has no parallel copy, no deduplication, and no resume on interruption
- Improvement path: Use `rsync -a --progress` instead of `cp -r`, or implement hardlink copying with `cp -al` (if filesystem supports it). Add `--quiet` flag to suppress output, and allow user to cancel with Ctrl+C.

**Repeated `git worktree list` Parsing:**
- Problem: Multiple functions call `git worktree list --porcelain` repeatedly (lines 121, 146, 151, 262, 354, 370, 395, 398). Each call spawns a git subprocess and parses output independently.
- Files: `wt.zsh` (lines 119-292, 352-409)
- Cause: No caching of worktree metadata across function calls
- Improvement path: Cache `git worktree list` output in a variable at function entry if called multiple times in sequence, or refactor to build a lookup table once per command invocation.

**`git stash list | wc -l` Counts All Stashes:**
- Problem: Line 203 and 206 count all stashes to detect if a new one was created. For repos with hundreds of stashes, this is O(n) and slow.
- Files: `wt.zsh` (lines 202-211)
- Cause: No built-in way to detect stash creation without counting
- Improvement path: Use `git stash create` to get the stash ID before and after, then check if IDs differ. Or use `git stash push` return code more reliably to detect success.

## Fragile Areas

**`_wt_eject` Multi-Step Transaction Without Rollback Guarantee:**
- Files: `wt.zsh` (lines 129-258)
- Why fragile: The eject operation involves 7+ git commands (stash, checkout, worktree add, stash pop). Partial failure at any step leaves the repo in an inconsistent state. Rollback attempts (lines 216-221, 227-233) may themselves fail or lose state.
- Safe modification: Do not change the order of stash/checkout/worktree/pop steps without understanding full state dependencies. Add comprehensive error logging to track which step failed. Consider wrapping the entire operation in a git transaction equivalent (not available in git, so add manual state snapshots before critical steps).
- Test coverage: Gaps in failure scenarios (e.g., what happens if stash pop fails after worktree is created). Test with intentional failures at each step.

**Worktree Name Resolution Ambiguity:**
- Files: `wt.zsh` (lines 352-409)
- Why fragile: The name matching (lines 358, 374, 402) uses suffix matching: `*"-${name}"` or exact match. A user naming a worktree "foo" could match repo "my-foo" unintentionally. Adding two worktrees with overlapping suffixes creates ambiguous resolution.
- Safe modification: Clarify naming constraints in documentation. Consider requiring exact match by default and adding a `--partial` flag for suffix matching.
- Test coverage: Gaps in multi-match scenarios. No tests for repos named similarly to worktree names.

**Fallback Branch Detection in `_wt_eject`:**
- Files: `wt.zsh` (lines 149-181)
- Why fragile: Fallback branch detection uses directory naming convention (line 167: `suffix="${dir_name#${repo_basename}-}"`) but this assumes consistent naming. A worktree created outside `wt` tool will not follow this pattern. The logic falls back to the full directory name as a branch name (line 170), which may not exist.
- Safe modification: Add validation that fallback branch actually exists before proceeding. Add verbose logging of branch resolution choices.
- Test coverage: Gap in non-standard worktree naming created by git command directly.

**Symlink Breakage Across Filesystems:**
- Files: `wt.zsh` (lines 91, 249)
- Why fragile: Symlinked `node_modules` breaks if source and target worktrees are moved to different filesystems or machines. No validation that symlink target is still valid.
- Safe modification: Add a check after symlink creation that `ls -la` confirms the symlink resolves. Document that worktrees must stay on same filesystem as main repo.
- Test coverage: No tests for symlink validity after worktree operations.

## Scaling Limits

**Max Worktrees Per Repository:**
- Current capacity: Tested up to ~20 worktrees; likely hits filesystem limits or git performance issues
- Limit: Shell loops over all worktrees on each `wt` command (lines 354, 370, 398). With hundreds of worktrees, `git worktree list --porcelain` becomes slow (O(n) parsing).
- Scaling path: Implement worktree filtering by prefix or pattern early in pipeline. Cache worktree list globally. Consider git-worktree native speedups as they're released (git 2.36+).

**Large node_modules Symlinks:**
- Current capacity: Works for typical projects (< 100K files in node_modules)
- Limit: Symlinked `node_modules` breaks with monorepos having multiple independent node_modules trees or workspace setups
- Scaling path: Support per-workspace symlink strategy. Add `--workspace` flag to selectively symlink specific workspace folders.

## Dependencies at Risk

**Git Worktree Feature Maturity:**
- Risk: Git worktree is relatively young feature. Breaking changes or behavior shifts between git versions (especially 2.35 - 2.40 range) could break the tool.
- Impact: Tool stops working with newer git versions that change worktree semantics
- Migration plan: Add git version check at startup (`git --version`). Document minimum required git version. Have fallback for deprecated flags or behaviors.

**Zsh-Specific Syntax:**
- Risk: Code uses zsh-specific features (parameter expansion with `:t` suffix, `(( ... ))` arithmetic, `${...#...}` pattern removal) that don't work in bash or POSIX sh.
- Impact: Code cannot be ported to bash or used in non-zsh shells
- Migration plan: If portability becomes a goal, refactor to POSIX-compatible syntax or add shell detection with fallback implementation.

## Missing Critical Features

**No Dry-Run Mode:**
- Problem: `wt delete`, `wt eject`, and `wt new` make immediate changes with no `--dry-run` flag to preview what will happen
- Blocks: Users cannot safely test commands before executing; no way to validate naming/routing without doing the operation

**No Stash Recovery Help After Eject Failure:**
- Problem: If `_wt_eject` fails after creating stash, the stash is left dangling. Users must manually find and pop it.
- Blocks: Unclear state after failure; users may lose work

**No Batch Operations:**
- Problem: Deleting multiple worktrees requires running `wt delete` N times
- Blocks: Cleanup of many worktrees is tedious

**No Completion for Branch Names:**
- Problem: `wt new [--copy-node-modules] <name> [branch]` doesn't offer completion for existing branches
- Blocks: Users must type branch names manually

## Test Coverage Gaps

**Error Path in Worktree Creation:**
- What's not tested: Behavior when `git worktree add` fails on second attempt (line 72). No test for when branch exists but worktree path doesn't.
- Files: `wt.zsh` (lines 70-76)
- Risk: Silent failure if both branch-creation and branch-checkout fail in unexpected ways
- Priority: High

**Stash Conflict Resolution:**
- What's not tested: Behavior when `git stash pop` encounters merge conflicts during `_wt_eject` (line 238). Current code silences stderr.
- Files: `wt.zsh` (lines 236-240)
- Risk: User loses visibility into conflicted stash state
- Priority: High

**Symlink Validation:**
- What's not tested: Whether symlinked `node_modules` persists after worktree is moved or filesystem is remounted
- Files: `wt.zsh` (lines 91, 249)
- Risk: Silent symlink breakage goes undetected
- Priority: Medium

**Ambiguous Name Resolution:**
- What's not tested: Behavior when multiple worktrees have overlapping suffix patterns (e.g., "main-foo" and "foo" both exist, and user runs `wt goto foo`)
- Files: `wt.zsh` (lines 358, 374, 402)
- Risk: Unpredictable worktree selection
- Priority: Medium

**Non-Standard Worktree Naming:**
- What's not tested: Behavior when `wt eject` is run in a worktree created outside the tool (e.g., via `git worktree add` directly with non-standard naming)
- Files: `wt.zsh` (lines 164-181)
- Risk: Fallback branch detection fails or selects wrong branch
- Priority: Medium

---

*Concerns audit: 2026-02-07*
