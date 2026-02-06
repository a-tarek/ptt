#!/usr/bin/env zsh
# wt — Git Worktree Manager

function wt() {
  local cmd="$1"
  shift 2>/dev/null

  case "$cmd" in
    new)    _wt_new "$@" ;;
    goto)   _wt_goto "$@" ;;
    home)   _wt_home ;;
    eject)  _wt_eject "$@" ;;
    list)   _wt_list ;;
    merge)  _wt_merge "$@" ;;
    rebase) _wt_rebase "$@" ;;
    delete) _wt_delete "$@" ;;
    *)
      echo "Usage: wt <command> [args]"
      echo ""
      echo "Commands:"
      echo "  new [--copy-node-modules] <name> [branch]   Create a new worktree"
      echo "  goto <worktree>                              cd into a worktree"
      echo "  home                                         cd into the main worktree"
      echo "  eject [name]                                 Eject current branch into its own worktree"
      echo "  list                                         List all worktrees"
      echo "  merge <worktree>                             Merge worktree's branch into current"
      echo "  rebase <worktree>                            Rebase current onto worktree's branch"
      echo "  delete <worktree>                            Remove a worktree (keeps branch)"
      return 1
      ;;
  esac
}

function _wt_new() {
  # Parse flags
  local copy_nm=false
  while [[ "$1" == --* ]]; do
    case "$1" in
      --copy-node-modules) copy_nm=true; shift ;;
      *) echo "Unknown flag: $1"; return 1 ;;
    esac
  done

  local name="$1"
  local branch="${2:-$name}"

  if [[ -z "$name" ]]; then
    echo "Usage: wt new [--copy-node-modules] <name> [branch]"
    return 1
  fi

  if ! git rev-parse --git-dir &>/dev/null; then
    echo "Error: not inside a git repository"
    return 1
  fi

  local src_root
  src_root="$(git rev-parse --show-toplevel)"
  local repo_basename="${src_root:t}"
  local target="../${repo_basename}-${name}"
  local target_abs="${src_root:h}/${repo_basename}-${name}"

  if [[ -d "$target_abs" ]]; then
    echo "Error: $target_abs already exists"
    return 1
  fi

  echo "Creating worktree: ${repo_basename}-${name} (branch: $branch)"

  if ! git worktree add "$target_abs" -b "$branch" 2>/dev/null; then
    # Branch might already exist — try without -b
    if ! git worktree add "$target_abs" "$branch"; then
      echo "Error: failed to create worktree"
      return 1
    fi
  fi

  # Copy .env.local if it exists
  if [[ -f "${src_root}/.env.local" ]]; then
    cp "${src_root}/.env.local" "${target_abs}/.env.local"
    echo "Copied .env.local"
  fi

  # node_modules handling
  if [[ -d "${src_root}/node_modules" ]]; then
    if $copy_nm; then
      echo "Copying node_modules (this may take a moment)..."
      cp -r "${src_root}/node_modules" "${target_abs}/node_modules"
      echo "Copied node_modules"
    else
      ln -s "${src_root}/node_modules" "${target_abs}/node_modules"
      echo "Symlinked node_modules"
    fi
  fi

  cd "$target_abs"
  echo ""
  echo "Ready: $(pwd)"
  echo "Branch: $(git branch --show-current)"
}

function _wt_goto() {
  local name="$1"
  if [[ -z "$name" ]]; then
    echo "Usage: wt goto <worktree>"
    return 1
  fi

  local wt_path
  wt_path=$(_wt_resolve_path "$name")
  if [[ -z "$wt_path" ]]; then
    echo "Error: worktree '$name' not found"
    return 1
  fi

  cd "$wt_path"
}

function _wt_home() {
  local main_path
  main_path=$(git worktree list --porcelain 2>/dev/null | head -1 | sed 's/^worktree //')
  if [[ -z "$main_path" ]]; then
    echo "Error: not inside a git repository"
    return 1
  fi
  cd "$main_path"
}

function _wt_eject() {
  if ! git rev-parse --git-dir &>/dev/null; then
    echo "Error: not inside a git repository"
    return 1
  fi

  # 1. Get current branch — error on detached HEAD
  local current_branch
  current_branch="$(git branch --show-current)"
  if [[ -z "$current_branch" ]]; then
    echo "Error: detached HEAD — nothing to eject"
    return 1
  fi

  local src_root
  src_root="$(git rev-parse --show-toplevel)"
  local repo_basename
  repo_basename="$(git worktree list --porcelain 2>/dev/null | head -1 | sed 's/^worktree //')"
  repo_basename="${repo_basename:t}"

  # 2. Determine fallback branch
  local home_path fallback_branch
  home_path="$(git worktree list --porcelain 2>/dev/null | head -1 | sed 's/^worktree //')"

  if [[ "$src_root" == "$home_path" ]]; then
    # Home worktree — fall back to main/master
    if git show-ref --verify --quiet refs/heads/main; then
      fallback_branch="main"
    elif git show-ref --verify --quiet refs/heads/master; then
      fallback_branch="master"
    else
      echo "Error: neither 'main' nor 'master' branch exists"
      return 1
    fi
  else
    # Non-home worktree — fall back to branch matching worktree folder suffix
    local dir_name="${src_root:t}"
    # Strip repo basename prefix to get the suffix (e.g. "web-app-staging" → "staging")
    local suffix="${dir_name#${repo_basename}-}"
    if [[ "$suffix" == "$dir_name" ]]; then
      # No prefix match — use full dir name
      suffix="$dir_name"
    fi
    fallback_branch="$suffix"
    if ! git show-ref --verify --quiet "refs/heads/${fallback_branch}"; then
      echo "Error: fallback branch '${fallback_branch}' does not exist"
      return 1
    fi
    if [[ "$fallback_branch" == "$current_branch" ]]; then
      echo "Error: current branch is already '${fallback_branch}' — nothing to eject"
      return 1
    fi
  fi

  # 3. Error if current branch is the fallback
  if [[ "$current_branch" == "$fallback_branch" ]]; then
    echo "Error: already on '${fallback_branch}' — nothing to eject"
    return 1
  fi

  # 4. Determine new worktree folder name
  local name="${1:-${current_branch//\//-}}"
  local target_abs="${src_root:h}/${repo_basename}-${name}"

  if [[ -d "$target_abs" ]]; then
    echo "Error: $target_abs already exists"
    return 1
  fi

  echo "Ejecting branch '${current_branch}' → ${repo_basename}-${name}"

  # 5. Stash uncommitted changes (including untracked)
  local stash_msg="wt-eject: ${current_branch}"
  local stash_before
  stash_before="$(git stash list | wc -l)"
  git stash push -u -m "$stash_msg" &>/dev/null
  local stash_after
  stash_after="$(git stash list | wc -l)"
  local did_stash=false
  if (( stash_after > stash_before )); then
    did_stash=true
    echo "Stashed uncommitted changes"
  fi

  # 6. Switch current worktree to the fallback branch
  if ! git checkout "$fallback_branch" &>/dev/null; then
    echo "Error: failed to switch to '${fallback_branch}'"
    # Try to pop stash back before aborting
    if $did_stash; then
      git stash pop &>/dev/null
    fi
    return 1
  fi
  echo "Switched to '${fallback_branch}'"

  # 7. Create new worktree for the ejected branch
  if ! git worktree add "$target_abs" "$current_branch" &>/dev/null; then
    echo "Error: failed to create worktree at $target_abs"
    # Roll back: switch back to original branch, pop stash
    git checkout "$current_branch" &>/dev/null
    if $did_stash; then
      git stash pop &>/dev/null
    fi
    return 1
  fi
  echo "Created worktree: ${repo_basename}-${name}"

  # 8. Pop stash in the new worktree
  if $did_stash; then
    git -C "$target_abs" stash pop &>/dev/null
    echo "Restored uncommitted changes in new worktree"
  fi

  # 9. Copy .env.local and symlink node_modules
  if [[ -f "${src_root}/.env.local" ]]; then
    cp "${src_root}/.env.local" "${target_abs}/.env.local"
    echo "Copied .env.local"
  fi

  if [[ -d "${src_root}/node_modules" ]]; then
    ln -s "${src_root}/node_modules" "${target_abs}/node_modules"
    echo "Symlinked node_modules"
  fi

  # 10. cd into the new worktree
  cd "$target_abs"
  echo ""
  echo "Ready: $(pwd)"
  echo "Branch: $(git branch --show-current)"
}

function _wt_list() {
  local base_name current_dir
  base_name=$(git worktree list --porcelain 2>/dev/null | head -1 | sed 's/^worktree //')
  if [[ -z "$base_name" ]]; then
    echo "Error: not inside a git repository"
    return 1
  fi
  current_dir="$(git rev-parse --show-toplevel)"

  local wt_entry="" branch=""
  git worktree list --porcelain 2>/dev/null | while IFS= read -r line; do
    if [[ "$line" == worktree\ * ]]; then
      wt_entry="${line#worktree }"
      branch=""
    elif [[ "$line" == branch\ * ]]; then
      branch="${line#branch refs/heads/}"
    elif [[ -z "$line" && -n "$wt_entry" ]]; then
      local marker=" "
      [[ "$wt_entry" == "$current_dir" ]] && marker="*"
      local dir="${wt_entry:t}"
      printf "%s %-30s %s\n" "$marker" "$dir" "$branch"
      wt_entry=""
      branch=""
    fi
  done
  # Handle last entry (porcelain output may not end with blank line)
  if [[ -n "$wt_entry" ]]; then
    local marker=" "
    [[ "$wt_entry" == "$current_dir" ]] && marker="*"
    local dir="${wt_entry:t}"
    printf "%s %-30s %s\n" "$marker" "$dir" "$branch"
  fi
}

function _wt_merge() {
  local name="$1"
  if [[ -z "$name" ]]; then
    echo "Usage: wt merge <worktree>"
    return 1
  fi

  local branch
  branch=$(_wt_resolve_branch "$name")
  if [[ -z "$branch" ]]; then
    echo "Error: worktree '$name' not found"
    return 1
  fi

  echo "Merging $branch into $(git branch --show-current)..."
  git merge "$branch"
}

function _wt_rebase() {
  local name="$1"
  if [[ -z "$name" ]]; then
    echo "Usage: wt rebase <worktree>"
    return 1
  fi

  local branch
  branch=$(_wt_resolve_branch "$name")
  if [[ -z "$branch" ]]; then
    echo "Error: worktree '$name' not found"
    return 1
  fi

  echo "Rebasing $(git branch --show-current) onto $branch..."
  git rebase "$branch"
}

function _wt_delete() {
  local name="$1"
  if [[ -z "$name" ]]; then
    echo "Usage: wt delete <worktree>"
    return 1
  fi

  local wt_path
  wt_path=$(_wt_resolve_path "$name")
  if [[ -z "$wt_path" ]]; then
    echo "Error: worktree '$name' not found"
    return 1
  fi

  echo "Removing worktree: $wt_path"
  git worktree remove "$wt_path"
}

# --- Helpers ---

# Resolve a worktree name to its path
# Matches against the directory basename suffix after the repo name
function _wt_resolve_path() {
  local name="$1"
  git worktree list --porcelain 2>/dev/null | while read -r line; do
    if [[ "$line" == worktree\ * ]]; then
      local wt_entry="${line#worktree }"
      local dir="${wt_entry:t}"
      if [[ "$dir" == *"-${name}" || "$dir" == "$name" ]]; then
        echo "$wt_entry"
        return 0
      fi
    fi
  done
}

# Resolve a worktree name to its branch
function _wt_resolve_branch() {
  local name="$1"
  local found_path=false
  git worktree list --porcelain 2>/dev/null | while read -r line; do
    if [[ "$line" == worktree\ * ]]; then
      local wt_entry="${line#worktree }"
      local dir="${wt_entry:t}"
      if [[ "$dir" == *"-${name}" || "$dir" == "$name" ]]; then
        found_path=true
      else
        found_path=false
      fi
    elif $found_path && [[ "$line" == branch\ * ]]; then
      echo "${line#branch refs/heads/}"
      return 0
    fi
  done
}

# List worktree short names (suffix after repo basename)
function _wt_list_names() {
  local root
  root="$(git rev-parse --show-toplevel 2>/dev/null)" || return
  local repo_basename="${root:t}"
  # Strip any existing suffix to get the base repo name
  # e.g. "web-app-foo" → base is "web-app", suffix is "foo"
  # We need the base name from the main worktree
  local base_name
  base_name=$(git worktree list --porcelain 2>/dev/null | head -1 | sed 's/^worktree //')
  base_name="${base_name:t}"

  git worktree list --porcelain 2>/dev/null | while read -r line; do
    if [[ "$line" == worktree\ * ]]; then
      local wt_entry="${line#worktree }"
      local dir="${wt_entry:t}"
      if [[ "$dir" == "${base_name}-"* && "$dir" != "$base_name" ]]; then
        echo "${dir#${base_name}-}"
      elif [[ "$dir" != "$base_name" ]]; then
        echo "$dir"
      fi
    fi
  done
}

# --- Zsh Completions ---

function _wt() {
  local -a subcmds
  subcmds=(
    'new:Create a new worktree'
    'goto:cd into a worktree'
    'home:cd into the main worktree'
    'eject:Eject current branch into its own worktree'
    'list:List all worktrees'
    'merge:Merge a worktree branch into current'
    'rebase:Rebase current onto a worktree branch'
    'delete:Remove a worktree'
  )

  if (( CURRENT == 2 )); then
    _describe 'command' subcmds
    return
  fi

  case "${words[2]}" in
    new)
      _arguments \
        '--copy-node-modules[Copy node_modules instead of symlinking]' \
        '1:name:' \
        '2:branch:'
      ;;
    goto|merge|rebase|delete)
      local -a wt_names
      wt_names=(${(f)"$(_wt_list_names)"})
      _describe 'worktree' wt_names
      ;;
  esac
}

compdef _wt wt
