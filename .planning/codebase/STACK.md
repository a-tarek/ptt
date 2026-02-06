# Technology Stack

**Analysis Date:** 2026-02-07

## Languages

**Primary:**
- Zsh - Shell scripting language used for the entire codebase

## Runtime

**Environment:**
- Zsh shell (typically zsh 5.0 or higher)

**Package Manager:**
- None - No external package dependencies

## Frameworks

**Core:**
- None - Pure shell implementation with no framework dependencies

**Testing:**
- Not applicable - Manual testing via shell commands

**Build/Dev:**
- None - Direct shell script execution

## Key Dependencies

**Critical:**
- None - Project has zero external dependencies

**Infrastructure:**
- Git (system dependency) - Used for all version control and worktree operations

## Configuration

**Environment:**
- Configured via Git configuration and environment variables passed from parent shell
- `.wtconfig` file (optional) - Plain text file defining copy/symlink actions for worktree creation
  - Created by `wt init` command
  - Format: `<action> <path>` (e.g., `copy .env.local`, `symlink node_modules`)
  - Read during worktree setup to automate file management
- No configuration files required for basic operation

**Build:**
- No build step required
- Script is directly executable: `#!/usr/bin/env zsh`

## Platform Requirements

**Development:**
- Zsh shell installed and available
- Git installed (2.7.0 or later for git worktree support)
- Unix-like operating system (Linux, macOS)
- Read/write access to repository directory
- Permission to run Git commands

**Production:**
- Same as development - the tool functions as a shell utility for developers
- Can be sourced into `.zshrc` for permanent availability

**Installation:**
- Source the `wt.zsh` file in shell profile:
  ```bash
  source /path/to/wt.zsh
  ```
- Or invoke directly:
  ```bash
  zsh /path/to/wt.zsh
  ```

---

*Stack analysis: 2026-02-07*
