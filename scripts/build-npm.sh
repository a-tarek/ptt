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
# goreleaser uses _v1 suffix for amd64, _v8.0 suffix for arm64 builds

# Define platform mappings (platform|source_path)
PLATFORMS=(
  "darwin-arm64|dist/wt_darwin_arm64_v8.0/wt"
  "darwin-amd64|dist/wt_darwin_amd64_v1/wt"
  "linux-arm64|dist/wt_linux_arm64_v8.0/wt"
  "linux-amd64|dist/wt_linux_amd64_v1/wt"
)

STAGED_COUNT=0

for entry in "${PLATFORMS[@]}"; do
  platform="${entry%%|*}"
  src="${entry##*|}"
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
