# Phase 8: Interactive Installer - Context

**Gathered:** 2026-02-07
**Status:** Ready for planning

<domain>
## Phase Boundary

Provide an npx-based installer (`npx @potato/wt install`) that safely modifies shell rc files with user confirmation. Includes an uninstall command (`wt uninstall`) to cleanly remove rc file entries. The installer handles shell detection, v1-to-v2 migration, idempotent installation, and rollback on failure.

</domain>

<decisions>
## Implementation Decisions

### Installation flow
- Guided walkthrough experience — show each step with explanations: detecting shell, showing changes, asking confirmation, reporting success
- After installation completes, print instructions to restart shell or `source ~/.zshrc` — no magic sourcing
- No pre-check of binary — trust that npm installed correctly, keep it simple
- When user declines installation, print a ready-to-paste copy block of the exact lines for their rc file

### RC file modifications
- Marker block format: Claude's discretion (pick a clear, standard format like conda or certbot style)
- Content inside marker block: `eval "$(wt shell-init zsh)"` — one eval line, always current, handles both wrapper functions and completions
- If wt v1 entry detected (`source wt.zsh` or similar): automatically comment out the v1 line and add the v2 block, showing the user what changed
- Placement: always append to end of file
- Idempotent: running installer twice must not duplicate entries (detect existing marker block)

### Shell detection & multi-shell
- Auto-detect current shell from `$SHELL`, show it to user, ask to confirm before proceeding
- Configure current shell only — don't offer multi-shell setup
- RC file selection: Claude's discretion based on shell conventions (e.g., bash may need .bashrc vs .bash_profile consideration)
- Unsupported shell handling: Claude's discretion

### Uninstall command
- Built-in `wt uninstall` command — removes marker block from rc file, prints npm uninstall instructions
- RC file backup before modification: Claude's discretion on whether to backup
- npm uninstall scope: Claude's discretion (rc cleanup only vs full cleanup including npm uninstall)
- On failure mid-installation: rollback any partial changes, undo modifications, print what went wrong and how to fix

### Claude's Discretion
- Marker block format (comment style)
- RC file selection logic per shell
- Unsupported shell error handling
- Whether to backup rc files before modification
- Whether `wt uninstall` also runs `npm uninstall` or just cleans rc files
- Exact wording of walkthrough steps and confirmation prompts

</decisions>

<specifics>
## Specific Ideas

- The eval approach mirrors how zoxide and other modern CLI tools handle shell integration — one line, always up to date
- v1 migration is important: existing wt.zsh users should have a smooth upgrade path where the installer handles commenting out the old entry
- Copy-paste manual instructions should be complete enough that a user can set up without the installer at all

</specifics>

<deferred>
## Deferred Ideas

None — discussion stayed within phase scope

</deferred>

---

*Phase: 08-interactive-installer*
*Context gathered: 2026-02-07*
