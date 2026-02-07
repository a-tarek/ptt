#!/usr/bin/env bash
set -euo pipefail

# Build script for wt npm distribution
# Runs goreleaser cross-compilation and stages binaries in npm platform directories

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"

echo "==> Building wt with goreleaser..."
cd "$PROJECT_ROOT"

# Run goreleaser build (snapshot mode for local builds, no git tags required)
goreleaser build --snapshot --clean

echo ""
echo "==> Staging binaries in npm platform directories..."

# Map goreleaser output to npm platform directories
# goreleaser uses _v1 suffix for amd64 builds (GOAMD64 default level)
declare -A PLATFORM_MAP=(
  ["darwin-arm64"]="dist/wt_darwin_arm64/wt"
  ["darwin-amd64"]="dist/wt_darwin_amd64_v1/wt"
  ["linux-arm64"]="dist/wt_linux_arm64/wt"
  ["linux-amd64"]="dist/wt_linux_amd64_v1/wt"
)

STAGED_COUNT=0

for platform in "${!PLATFORM_MAP[@]}"; do
  src="${PLATFORM_MAP[$platform]}"
  dest="npm/platforms/$platform/bin/wt"

  if [[ ! -f "$src" ]]; then
    echo "ERROR: Expected binary not found: $src"
    echo "Available files in dist/:"
    ls -la dist/ || true
    exit 1
  fi

  # Copy binary to npm platform directory
  cp "$src" "$dest"

  # Make executable (critical for npm package functionality)
  chmod +x "$dest"

  # Verify
  if [[ -x "$dest" ]]; then
    echo "  ✓ $platform: staged and executable"
    ((STAGED_COUNT++))
  else
    echo "  ✗ $platform: ERROR - not executable"
    exit 1
  fi
done

echo ""
echo "==> Build complete!"
echo "    Staged $STAGED_COUNT binaries across 4 platforms"
echo "    Ready for 'npm pack' or 'npm publish'"
