# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project

ptt is a fast, cross-platform git worktree manager written in Go, distributed via npm (`@atarek/ptt`). It supports bash, zsh, and fish shells on Linux and macOS. The Go module is `github.com/a-tarek/ptt`.

## Commands

```bash
# Build
go build .

# Run all tests
go test ./...

# Run a single package's tests
go test ./internal/config/...
go test ./cmd/...

# Run a specific test
go test -run TestParseLine ./internal/config/...

# Verbose with coverage
go test -v -coverprofile=coverage.out ./...

# Vet
go vet ./...

# E2E shell tests (in tests/shell/, builds binary then tests bash/zsh/fish wrappers)
go test ./tests/shell/...
```

## Architecture

### Shell wrapper pattern

Commands that change the shell's directory (like `ptt cd`, `ptt mk`, `ptt eject`) use a two-part design:
1. The Go binary accepts an `--output-path` flag and writes the target directory path to stdout
2. A thin shell wrapper function (installed in the user's RC file) reads that stdout and performs the actual `cd`

This is necessary because a subprocess cannot change the parent shell's working directory. The shell wrapper templates live in `internal/shell/templates/` (bash, zsh, fish) and are embedded at compile time.

### Package layout

- **`cmd/`** — Cobra command definitions. Each command file registers itself in `init()`. Commands that change directories accept `--output-path`.
- **`internal/git/`** — Git operations: repo type detection, worktree path resolution (suffix matching), bare repo root discovery. `BareRepoRoot()` looks for `.bare/` directory with a `.git` pointer file.
- **`internal/config/`** — Parses `.pttconfig/` YAML config files with `create:` and `remove:` lifecycle hooks (actions: `copy`, `symlink`, `run`). Handles CLI flag overrides that merge with file-based config (flags win per-path).
- **`internal/setup/`** — Executes config actions (copy, symlink, run) with rollback on failure.
- **`internal/shell/`** — Shell detection from `$SHELL` and embedded wrapper templates.
- **`internal/installer/`** — RC file modification using marker blocks (`# >>> ptt >>>` / `# <<< ptt <<<`).
- **`internal/init/`** — Bare repo conversion: detects repo type, plans operations, restructures normal clones into ptt bare format.
- **`internal/ui/`** — Task list display with checkmarks for progress output.

### Bare repo structure

ptt supports a nested worktree layout via bare repos:
```
project-bare/
├── .bare/          # Actual git database
├── .git            # Pointer file (gitdir: ./.bare)
├── .pttconfig/     # Shared config
└── main/           # Worktrees live here as siblings
```

### Config resolution

Config root varies by repo type: bare repos use the container directory (parent of `.bare/`), regular repos use the home worktree root. The `ConfigRoot()` function in `internal/git/` handles this.

## Testing patterns

- Unit tests are colocated with source files (`*_test.go`).
- E2E shell wrapper tests are in `tests/shell/shell_test.go` — they compile the binary once via `sync.Once`, then test all three shells.
- CI runs on both `ubuntu-latest` and `macos-latest`.

## Release

Releases are built with goreleaser (`.goreleaser.yaml`), producing flat binaries for darwin/linux on amd64/arm64. Version is injected via ldflags into `cmd.Version`. npm platform packages are staged in `npm/platforms/` and published via `scripts/publish-npm.sh`.

## Zsh pitfall

`path` is a reserved zsh array variable tied to `$PATH`. Never use `local path="..."` in zsh wrapper code — it clobbers PATH lookups. Use a different variable name like `entry` or `cfg_path`.
