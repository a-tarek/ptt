#!/usr/bin/env bash
set -euo pipefail

# Publish script for wt npm distribution
# Handles version updates and coordinated publishing of all 5 packages

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"

# Parse arguments
VERSION=""
DRY_RUN=false

while [[ $# -gt 0 ]]; do
  case $1 in
    --dry-run)
      DRY_RUN=true
      shift
      ;;
    *)
      if [[ -z "$VERSION" ]]; then
        VERSION="$1"
        shift
      else
        echo "ERROR: Unexpected argument: $1"
        exit 1
      fi
      ;;
  esac
done

if [[ -z "$VERSION" ]]; then
  echo "ERROR: VERSION required"
  echo "Usage: $0 VERSION [--dry-run]"
  echo "Example: $0 2.0.0"
  echo "         $0 2.0.0 --dry-run"
  exit 1
fi

# Validate version format (basic semver check)
if ! [[ "$VERSION" =~ ^[0-9]+\.[0-9]+\.[0-9]+(-[a-zA-Z0-9.]+)?$ ]]; then
  echo "ERROR: Invalid version format: $VERSION"
  echo "Expected semver format: X.Y.Z or X.Y.Z-prerelease"
  exit 1
fi

cd "$PROJECT_ROOT/npm"

echo "==> Updating version to $VERSION in all package.json files..."

# Define all package paths
MAIN_PKG="package.json"
PLATFORM_PKGS=(
  "platforms/darwin-arm64/package.json"
  "platforms/darwin-amd64/package.json"
  "platforms/linux-arm64/package.json"
  "platforms/linux-amd64/package.json"
)

# Update version in all packages
for pkg in "$MAIN_PKG" "${PLATFORM_PKGS[@]}"; do
  if [[ ! -f "$pkg" ]]; then
    echo "ERROR: Package file not found: $pkg"
    exit 1
  fi

  # Use node to update version in JSON
  node -e "
    const fs = require('fs');
    const pkg = JSON.parse(fs.readFileSync('$pkg', 'utf8'));
    pkg.version = '$VERSION';
    fs.writeFileSync('$pkg', JSON.stringify(pkg, null, 2) + '\n');
  "

  echo "  ✓ Updated $pkg"
done

# Update optionalDependencies versions in main package.json
echo "  ✓ Updating optionalDependencies in main package..."
node -e "
  const fs = require('fs');
  const pkg = JSON.parse(fs.readFileSync('$MAIN_PKG', 'utf8'));
  if (pkg.optionalDependencies) {
    for (const dep in pkg.optionalDependencies) {
      pkg.optionalDependencies[dep] = '$VERSION';
    }
    fs.writeFileSync('$MAIN_PKG', JSON.stringify(pkg, null, 2) + '\n');
  }
"

echo ""

# Prepare publish command
PUBLISH_CMD="npm publish --access public"
if [[ "$DRY_RUN" == true ]]; then
  PUBLISH_CMD="npm publish --dry-run"
  echo "==> DRY RUN MODE - No actual publishing"
else
  echo "==> Publishing packages to npm..."
fi

echo ""

# Publish platform packages FIRST (main package depends on them)
for platform_dir in platforms/darwin-arm64 platforms/darwin-amd64 platforms/linux-arm64 platforms/linux-amd64; do
  echo "Publishing $platform_dir..."
  (cd "$platform_dir" && $PUBLISH_CMD)
  echo "  ✓ Published $(basename $platform_dir)"
  echo ""
done

# Publish main package LAST
echo "Publishing main package @potato/wt..."
$PUBLISH_CMD
echo "  ✓ Published @potato/wt"

echo ""
if [[ "$DRY_RUN" == true ]]; then
  echo "==> DRY RUN complete - no packages were published"
else
  echo "==> Publish complete!"
  echo "    All 5 packages published at version $VERSION"
  echo "    Install with: npm install @potato/wt@$VERSION"
fi
