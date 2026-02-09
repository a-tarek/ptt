#!/usr/bin/env bash
set -euo pipefail

# Auto-increment version, tag, and push to trigger CI release
# Usage: ./scripts/release.sh [patch|minor|major]
# Default: patch

BUMP="${1:-patch}"

# Get latest tag
LATEST=$(git tag --sort=-v:refname | head -1)

if [[ -z "$LATEST" ]]; then
  echo "No existing tags found. Starting at v0.1.0"
  LATEST="v0.0.0"
fi

# Strip leading v and any prerelease suffix (e.g., v0.1.1-beta -> 0.1.1)
VERSION="${LATEST#v}"
VERSION="${VERSION%%-*}"

IFS='.' read -r MAJOR MINOR PATCH <<< "$VERSION"

case "$BUMP" in
  major)
    MAJOR=$((MAJOR + 1))
    MINOR=0
    PATCH=0
    ;;
  minor)
    MINOR=$((MINOR + 1))
    PATCH=0
    ;;
  patch)
    PATCH=$((PATCH + 1))
    ;;
  *)
    echo "Usage: $0 [patch|minor|major]"
    exit 1
    ;;
esac

NEW_VERSION="v${MAJOR}.${MINOR}.${PATCH}"

echo "Current: $LATEST"
echo "Next:    $NEW_VERSION"
echo ""

# Ensure we're on main branch
BRANCH=$(git branch --show-current)
if [[ "$BRANCH" != "main" ]]; then
  echo "WARNING: You're on '$BRANCH', not 'main'"
  read -rp "Continue anyway? (y/N) " confirm
  [[ "$confirm" =~ ^[Yy]$ ]] || exit 0
fi

# Ensure working tree is clean
if [[ -n "$(git status --porcelain)" ]]; then
  echo "ERROR: Working tree is dirty. Commit or stash changes first."
  exit 1
fi

read -rp "Tag and push $NEW_VERSION? (y/N) " confirm
[[ "$confirm" =~ ^[Yy]$ ]] || exit 0

git tag "$NEW_VERSION"
git push origin "$NEW_VERSION"

echo ""
echo "Tagged and pushed $NEW_VERSION"
echo "CI will build and publish to npm automatically."
