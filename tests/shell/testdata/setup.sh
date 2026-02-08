#!/usr/bin/env bash
# setup.sh - Creates temporary git repos with worktrees for shell E2E testing
# Usage: setup.sh <tmpdir>

set -euo pipefail

if [ $# -ne 1 ]; then
    echo "Usage: $0 <tmpdir>" >&2
    exit 1
fi

TMPDIR="$1"

# Create bare repo
BARE="$TMPDIR/repo.git"
git init --bare "$BARE" >/dev/null 2>&1

# Create main worktree with initial commit
MAIN="$BARE/main"
git worktree add "$MAIN" -b main >/dev/null 2>&1
cd "$MAIN"
echo "Initial content" > README.md
git add README.md
git config user.email "test@example.com"
git config user.name "Test User"
git commit -m "Initial commit" >/dev/null 2>&1

# Create feature worktree
FEATURE="$BARE/feature"
cd "$BARE"
git worktree add "$FEATURE" -b feature >/dev/null 2>&1
cd "$FEATURE"
echo "Feature content" > feature.txt
git add feature.txt
git commit -m "Add feature" >/dev/null 2>&1

# Print paths in parseable format
echo "BARE=$BARE"
echo "MAIN=$MAIN"
echo "FEATURE=$FEATURE"
