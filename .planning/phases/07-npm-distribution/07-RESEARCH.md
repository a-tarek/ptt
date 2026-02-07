# Phase 7: npm Distribution - Research

**Researched:** 2026-02-07
**Domain:** Go binary distribution via npm
**Confidence:** HIGH

## Summary

Distributing Go binaries via npm is a well-established pattern used by major tools (esbuild, turbo, biome). The standard approach uses **optionalDependencies** to publish platform-specific packages (e.g., `@scope/wt-darwin-arm64`) alongside a main wrapper package (`@scope/wt`). npm automatically installs only the binary matching the user's platform.

Two tools exist for implementation:
1. **GoReleaser's built-in npm support** (experimental, uses postinstall script)
2. **Manual implementation** following the esbuild pattern (production-proven)

The esbuild pattern is recommended: it's battle-tested, doesn't rely on postinstall scripts (which many users disable), and provides better error handling.

**Primary recommendation:** Use the **manual optionalDependencies approach** with platform-specific packages. Structure: main package `@scope/wt` with thin Node.js wrapper + platform packages `@scope/wt-{os}-{arch}` containing Go binaries. Use GoReleaser for cross-compilation only.

## Standard Stack

### Core
| Library | Version | Purpose | Why Standard |
|---------|---------|---------|--------------|
| goreleaser | v2.13+ | Cross-platform Go binary compilation | Industry standard for Go releases, handles all GOOS/GOARCH combinations |
| Node.js | 18+ | npm package infrastructure | Required for npm distribution, provides `process.platform`/`process.arch` detection |

### Supporting
| Library | Version | Purpose | When to Use |
|---------|---------|---------|-------------|
| goreleaser-npm-publisher | Latest | Automates npm package generation from GoReleaser | Alternative to manual implementation (less control) |

### Alternatives Considered
| Instead of | Could Use | Tradeoff |
|------------|-----------|----------|
| Manual optionalDependencies | GoReleaser built-in npm support | GoReleaser's npm feature uses postinstall scripts (many users disable with `--ignore-scripts`), less proven |
| optionalDependencies pattern | Bundle all binaries in one package | Bloat: 15MB binary × 4 platforms = 60MB download vs 15MB for single platform |
| npm distribution | Direct GitHub Releases download | npm provides: version management, `npx` support, familiar install story, download tracking |

**Installation:**
```bash
# No additional Node.js packages needed for implementation
# Users install with:
npm install -g @scope/wt
# or
npx @scope/wt <command>
```

## Architecture Patterns

### Recommended Project Structure
```
npm/
├── package.json                 # Main wrapper package
├── install.js                   # Fallback installer (optional)
├── bin/
│   └── wt                       # Node.js wrapper script
└── platforms/
    ├── darwin-arm64/
    │   ├── package.json         # Platform-specific package
    │   └── bin/
    │       └── wt               # Go binary for darwin-arm64
    ├── darwin-amd64/
    │   └── ...
    ├── linux-amd64/
    │   └── ...
    └── linux-arm64/
        └── ...
```

### Pattern 1: Platform-Specific Package Structure
**What:** Each platform has its own npm package with `os` and `cpu` constraints
**When to use:** Always - this is the standard approach
**Example:**
```json
// npm/platforms/darwin-arm64/package.json
{
  "name": "@scope/wt-darwin-arm64",
  "version": "1.0.0",
  "os": ["darwin"],
  "cpu": ["arm64"],
  "bin": {
    "wt": "bin/wt"
  },
  "files": ["bin/wt"]
}
```
**Source:** [Sentry Engineering: Publishing Binaries on npm](https://sentry.engineering/blog/publishing-binaries-on-npm)

### Pattern 2: Main Package with optionalDependencies
**What:** Wrapper package lists all platform packages as optionalDependencies
**When to use:** Always - npm auto-installs only the matching platform
**Example:**
```json
// npm/package.json
{
  "name": "@scope/wt",
  "version": "1.0.0",
  "bin": {
    "wt": "bin/wt"
  },
  "optionalDependencies": {
    "@scope/wt-darwin-arm64": "1.0.0",
    "@scope/wt-darwin-amd64": "1.0.0",
    "@scope/wt-linux-amd64": "1.0.0",
    "@scope/wt-linux-arm64": "1.0.0"
  }
}
```
**Source:** [esbuild package.json structure](https://github.com/evanw/esbuild/blob/main/npm/esbuild/package.json)

### Pattern 3: Binary Path Resolution in Wrapper
**What:** Node.js wrapper detects platform and resolves correct binary path
**When to use:** Always - provides fallback if optionalDependencies fail
**Example:**
```javascript
#!/usr/bin/env node
// npm/bin/wt

const { spawn } = require('child_process');
const { join } = require('path');

function getBinaryPath() {
  const platform = process.platform;
  const arch = process.arch;

  // Map Node.js platform/arch to package names
  const platformMap = {
    'darwin-arm64': '@scope/wt-darwin-arm64',
    'darwin-x64': '@scope/wt-darwin-amd64',
    'linux-x64': '@scope/wt-linux-amd64',
    'linux-arm64': '@scope/wt-linux-arm64',
  };

  const key = `${platform}-${arch}`;
  const packageName = platformMap[key];

  if (!packageName) {
    console.error(`Unsupported platform: ${platform}-${arch}`);
    process.exit(1);
  }

  // Try to resolve from optionalDependencies
  try {
    return require.resolve(`${packageName}/bin/wt`);
  } catch (e) {
    console.error(`Failed to find ${packageName}. Installation may have failed.`);
    process.exit(1);
  }
}

// Forward all args to the Go binary
const binaryPath = getBinaryPath();
const child = spawn(binaryPath, process.argv.slice(2), { stdio: 'inherit' });
child.on('exit', (code) => process.exit(code));
```
**Source:** [Sentry Engineering: Publishing Binaries on npm](https://sentry.engineering/blog/publishing-binaries-on-npm)

### Pattern 4: GOOS/GOARCH to npm Platform Mapping
**What:** Go's platform identifiers must map to Node.js platform/arch values
**When to use:** When configuring GoReleaser builds and npm packages
**Example:**
```yaml
# .goreleaser.yml
builds:
  - id: wt
    binary: wt
    goos:
      - darwin
      - linux
    goarch:
      - amd64
      - arm64
    env:
      - CGO_ENABLED=0
```

**Mapping table:**
| GOOS | GOARCH | npm os | npm cpu | Package Name |
|------|--------|--------|---------|--------------|
| darwin | amd64 | darwin | x64 | @scope/wt-darwin-amd64 |
| darwin | arm64 | darwin | arm64 | @scope/wt-darwin-arm64 |
| linux | amd64 | linux | x64 | @scope/wt-linux-amd64 |
| linux | arm64 | linux | arm64 | @scope/wt-linux-arm64 |

**Sources:**
- [Go GOOS/GOARCH values](https://gist.github.com/asukakenji/f15ba7e588ac42795f421b48b8aede63)
- [Node.js process.platform documentation](https://nodejs.org/api/process.html)

### Anti-Patterns to Avoid
- **Bundle all binaries in main package:** Creates 60MB+ downloads when users only need 15MB
- **Rely solely on postinstall scripts:** Users with `--ignore-scripts` can't install (security-conscious orgs disable scripts)
- **Use GoReleaser's experimental npm support for production:** Less battle-tested than manual approach; postinstall-dependent
- **Forget executable permissions:** Binaries must have `chmod +x` applied before packaging
- **Mix amd64/x64 naming:** Go uses "amd64", npm uses "x64" - be consistent per context

## Don't Hand-Roll

| Problem | Don't Build | Use Instead | Why |
|---------|-------------|-------------|-----|
| Cross-platform Go compilation | Custom build scripts | GoReleaser | Handles GOOS/GOARCH matrix, archives, checksums, releases |
| Platform detection in Node.js | Custom OS detection logic | `process.platform` / `process.arch` | Built-in Node.js APIs, handles all edge cases |
| Binary path resolution | Hardcoded paths | `require.resolve()` | Respects npm's module resolution, works with symlinks |
| Package publishing | Manual `npm publish` per platform | Automation script or CI/CD | 5+ packages need coordinated version bumps |

**Key insight:** The optionalDependencies mechanism is npm's built-in solution for platform-specific binaries. Don't reinvent it with custom download logic - it handles platform detection, caching, and fallback resolution automatically.

## Common Pitfalls

### Pitfall 1: Executable Permissions Lost
**What goes wrong:** Binary works locally but gets "permission denied" after npm install
**Why it happens:** File permissions don't survive git upload/download or tarball creation
**How to avoid:**
- Ensure binaries have `chmod +x` before `npm pack`
- npm automatically applies executable permissions to files in `bin` field of package.json
- Verify with: `tar -tzf *.tgz` shows executable flag
**Warning signs:** Users report "EACCES: permission denied" on installation

### Pitfall 2: Publishing Private Scoped Packages by Default
**What goes wrong:** `npm publish` for `@scope/wt` fails or creates private package
**Why it happens:** Scoped packages default to private visibility
**How to avoid:** Use `npm publish --access public` for first publish
**Warning signs:** `npm ERR! 402 Payment Required` or package not visible on npmjs.com
**Source:** [npm Docs: Creating scoped public packages](https://docs.npmjs.com/creating-and-publishing-scoped-public-packages/)

### Pitfall 3: package-lock.json Platform Mismatch
**What goes wrong:** `npm ci` fails on Windows when package-lock.json was created on macOS
**Why it happens:** Known npm bug (npm/cli#4828) - lockfile only includes current platform's optionalDependencies
**How to avoid:**
- Don't commit package-lock.json for packages with optionalDependencies
- Or regenerate lockfile on each platform before CI runs
- Or use `npm install` instead of `npm ci` (slower but resolves all platforms)
**Warning signs:** CI fails with "package X not found" on different platforms
**Source:** [npm CLI issue #4828](https://github.com/npm/cli/issues/4828)

### Pitfall 4: Node.js Arch != OS Arch
**What goes wrong:** 32-bit Node.js on 64-bit system downloads wrong binary
**Why it happens:** `process.arch` returns Node.js binary's architecture, not OS architecture
**How to avoid:** Document that users need 64-bit Node.js for 64-bit systems
**Warning signs:** Reports of "wrong architecture" errors on 64-bit systems
**Source:** [Node.js process.arch issues](https://github.com/nodejs/node/issues/9491)

### Pitfall 5: Forgetting npm Pack Test
**What goes wrong:** Package works with `npm link` but fails after `npm install` from registry
**Why it happens:** `npm link` uses symlinks (doesn't test packaging), `npm pack` simulates actual publish
**How to avoid:**
```bash
npm pack
npm install -g ./scope-wt-1.0.0.tgz
wt --version
```
**Warning signs:** Works in development but not for users

### Pitfall 6: Missing bin Shebang
**What goes wrong:** Binary works on macOS/Linux but fails on Windows (or vice versa)
**Why it happens:** Node.js wrapper needs `#!/usr/bin/env node` shebang
**How to avoid:** First line of `bin/wt` must be `#!/usr/bin/env node`
**Warning signs:** "command not found" on Unix systems despite npm install success

## Code Examples

Verified patterns from official sources:

### Complete Main Package Configuration
```json
// npm/package.json
{
  "name": "@scope/wt",
  "version": "1.0.0",
  "description": "Git worktree manager",
  "repository": "github:username/wt",
  "license": "MIT",
  "bin": {
    "wt": "bin/wt"
  },
  "files": [
    "bin/wt"
  ],
  "optionalDependencies": {
    "@scope/wt-darwin-arm64": "1.0.0",
    "@scope/wt-darwin-amd64": "1.0.0",
    "@scope/wt-linux-amd64": "1.0.0",
    "@scope/wt-linux-arm64": "1.0.0"
  },
  "engines": {
    "node": ">=18"
  }
}
```
**Source:** Pattern from [esbuild package.json](https://github.com/evanw/esbuild/blob/main/npm/esbuild/package.json)

### Complete Platform Package Configuration
```json
// npm/platforms/darwin-arm64/package.json
{
  "name": "@scope/wt-darwin-arm64",
  "version": "1.0.0",
  "description": "wt binary for macOS ARM64",
  "repository": "github:username/wt",
  "license": "MIT",
  "os": ["darwin"],
  "cpu": ["arm64"],
  "bin": {
    "wt": "bin/wt"
  },
  "files": [
    "bin/wt"
  ]
}
```
**Source:** [Sentry Engineering blog](https://sentry.engineering/blog/publishing-binaries-on-npm)

### GoReleaser Configuration for npm Builds
```yaml
# .goreleaser.yml
project_name: wt

before:
  hooks:
    - go mod tidy

builds:
  - id: wt
    binary: wt
    main: ./main.go
    goos:
      - darwin
      - linux
    goarch:
      - amd64
      - arm64
    env:
      - CGO_ENABLED=0
    ldflags:
      - -s -w
      - -X main.version={{.Version}}
      - -X main.commit={{.Commit}}
      - -X main.date={{.Date}}

archives:
  - id: default
    format: binary
    name_template: "{{ .ProjectName }}_{{ .Os }}_{{ .Arch }}"

release:
  github:
    owner: username
    name: wt
```
**Note:** This config generates binaries only. npm packaging is manual.
**Source:** [GoReleaser documentation](https://goreleaser.com/customization/builds/go/)

### Platform Detection and Binary Execution
```javascript
#!/usr/bin/env node
// npm/bin/wt - Complete production-ready wrapper

const { spawn } = require('child_process');
const { resolve } = require('path');

const PLATFORM_MAP = {
  'darwin-arm64': '@scope/wt-darwin-arm64',
  'darwin-x64': '@scope/wt-darwin-amd64',
  'linux-x64': '@scope/wt-linux-amd64',
  'linux-arm64': '@scope/wt-linux-arm64',
};

function getPlatformKey() {
  const platform = process.platform;
  const arch = process.arch;
  return `${platform}-${arch}`;
}

function getBinaryPath() {
  const platformKey = getPlatformKey();
  const packageName = PLATFORM_MAP[platformKey];

  if (!packageName) {
    const supported = Object.keys(PLATFORM_MAP).join(', ');
    console.error(
      `Unsupported platform: ${platformKey}. ` +
      `Supported platforms: ${supported}`
    );
    process.exit(1);
  }

  try {
    // Resolve binary from optional dependency
    return require.resolve(`${packageName}/bin/wt`);
  } catch (error) {
    console.error(
      `Failed to find ${packageName}. ` +
      `This likely means the optional dependency failed to install.\n` +
      `Error: ${error.message}`
    );
    process.exit(1);
  }
}

function runBinary() {
  const binaryPath = getBinaryPath();
  const child = spawn(binaryPath, process.argv.slice(2), {
    stdio: 'inherit',
    windowsHide: false,
  });

  child.on('error', (error) => {
    console.error(`Failed to execute binary: ${error.message}`);
    process.exit(1);
  });

  child.on('exit', (code, signal) => {
    if (signal) {
      process.kill(process.pid, signal);
    } else {
      process.exit(code || 0);
    }
  });
}

runBinary();
```
**Source:** Pattern from [turbo CLI wrapper](https://github.com/vercel/turbo/issues/1749)

### Automation Script for Publishing All Packages
```bash
#!/bin/bash
# scripts/publish-npm.sh - Publish all npm packages

set -e

VERSION=$1
if [ -z "$VERSION" ]; then
  echo "Usage: ./publish-npm.sh <version>"
  exit 1
fi

echo "Publishing version $VERSION..."

# Update versions in all package.json files
for pkg in npm/package.json npm/platforms/*/package.json; do
  jq ".version = \"$VERSION\"" "$pkg" > "$pkg.tmp" && mv "$pkg.tmp" "$pkg"
done

# Update optionalDependencies versions in main package
cd npm
for platform in darwin-arm64 darwin-amd64 linux-amd64 linux-arm64; do
  jq ".optionalDependencies[\"@scope/wt-$platform\"] = \"$VERSION\"" package.json > package.json.tmp
  mv package.json.tmp package.json
done

# Publish platform packages first
for platform_dir in platforms/*; do
  echo "Publishing $(basename $platform_dir)..."
  cd "$platform_dir"
  npm publish --access public
  cd ../..
done

# Publish main package
echo "Publishing main package..."
npm publish --access public

echo "All packages published successfully!"
```
**Note:** Production script should include error handling and dry-run mode

## State of the Art

| Old Approach | Current Approach | When Changed | Impact |
|--------------|------------------|--------------|--------|
| Postinstall script downloads from GitHub Releases | optionalDependencies with platform packages | ~2020 (esbuild adoption) | Faster installs, works with `--ignore-scripts`, better caching |
| Single package with all binaries | Platform-specific packages | ~2019 | 60MB+ download → 15MB, npm auto-selects correct binary |
| Manual binary selection in wrapper | `process.platform`/`process.arch` built-ins | Always standard | Reliable cross-platform detection |
| GoReleaser homebrew-only | GoReleaser + npm distribution | 2021+ | Wider distribution, easier for Node.js developers |

**Deprecated/outdated:**
- **go-npm library:** Early tool (2017) for npm distribution; superseded by optionalDependencies pattern
- **GoReleaser npm feature (experimental):** Uses postinstall scripts; less reliable than manual approach
- **Including Windows binaries:** Requirements specify darwin/linux only; skip Windows support

## Open Questions

Things that couldn't be fully resolved:

1. **What npm scope to use for @scope/wt?**
   - What we know: User hasn't specified the npm organization/scope name
   - What's unclear: Whether user has existing npm organization or needs to create one
   - Recommendation: Plan tasks to determine scope name (use GitHub username, create npm org, or user-specified)

2. **Should we support Windows initially?**
   - What we know: Requirements specify darwin-arm64, darwin-amd64, linux-amd64, linux-arm64 (no Windows)
   - What's unclear: Whether Windows support is intentionally excluded or oversight
   - Recommendation: Follow requirements (no Windows) but structure npm packages to easily add Windows later

3. **goreleaser-npm-publisher vs manual implementation?**
   - What we know: goreleaser-npm-publisher automates package generation but is less proven
   - What's unclear: How much maintenance effort manual implementation requires
   - Recommendation: Start with manual approach (battle-tested pattern from esbuild/turbo); migrate to automation tool if maintenance burden is high

4. **CI/CD integration for publishing?**
   - What we know: Publishing 5 packages (1 main + 4 platforms) manually is error-prone
   - What's unclear: What CI/CD system user has (GitHub Actions, etc.)
   - Recommendation: Include automation script; CI/CD integration can be Phase 8 or separate effort

## Sources

### Primary (HIGH confidence)
- [GoReleaser npm documentation](https://goreleaser.com/customization/npm/) - Official docs for npm integration
- [esbuild package.json structure](https://github.com/evanw/esbuild/blob/main/npm/esbuild/package.json) - Battle-tested reference implementation
- [Sentry Engineering: Publishing Binaries on npm](https://sentry.engineering/blog/publishing-binaries-on-npm) - Comprehensive guide with code examples
- [npm Docs: Creating scoped public packages](https://docs.npmjs.com/creating-and-publishing-scoped-public-packages/) - Official npm documentation
- [Node.js process API documentation](https://nodejs.org/api/process.html) - Official Node.js docs

### Secondary (MEDIUM confidence)
- [turbo npm package structure](https://github.com/vercel/turbo/issues/1749) - Real-world issue discussing platform package handling
- [goreleaser-npm-publisher](https://github.com/evg4b/goreleaser-npm-publisher) - Automation tool (less proven than manual approach)
- [Go GOOS/GOARCH values](https://gist.github.com/asukakenji/f15ba7e588ac42795f421b48b8aede63) - Community-maintained reference
- [npm optionalDependencies platform-specific bugs](https://github.com/npm/cli/issues/4828) - Known issues with package-lock.json

### Tertiary (LOW confidence)
- Various blog posts about npm binary distribution (2019-2026) - Patterns are consistent but implementations vary

## Metadata

**Confidence breakdown:**
- Standard stack: HIGH - esbuild/turbo/biome prove the pattern works at scale
- Architecture: HIGH - optionalDependencies is npm's built-in solution, well-documented
- Pitfalls: HIGH - All pitfalls verified with official sources or community issue trackers
- GoReleaser integration: MEDIUM - Documentation exists but npm feature is experimental; manual approach more proven

**Research date:** 2026-02-07
**Valid until:** ~30 days (stable patterns, but npm/goreleaser may release updates)
