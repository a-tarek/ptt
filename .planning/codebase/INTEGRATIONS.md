# External Integrations

**Analysis Date:** 2026-02-07

## APIs & External Services

**None detected** - This is a standalone shell utility that does not integrate with external APIs or cloud services.

## Data Storage

**Databases:**
- None - No database integration

**File Storage:**
- Local filesystem only
  - Worktree directories stored in filesystem
  - Git metadata stored in `.git` directory

**Caching:**
- None - No caching layer

## Authentication & Identity

**Auth Provider:**
- None required

**Implementation:**
- Tool operates with the user's current Git credentials and filesystem permissions
- No separate authentication mechanism

## Monitoring & Observability

**Error Tracking:**
- None - Errors reported to stdout/stderr

**Logs:**
- All output is directed to stdout/stderr
- No persistent logging infrastructure
- Error messages are user-facing and printed directly to terminal

## CI/CD & Deployment

**Hosting:**
- Not applicable - This is a development utility, not a deployed service

**CI Pipeline:**
- Not applicable

**Distribution:**
- Source code distributed via Git repository
- Installation by sourcing shell script directly

## Environment Configuration

**Required env vars:**
- None - Tool uses Git's own environment and inherits parent shell variables

**Optional env vars:**
- None explicitly required

**Secrets location:**
- Not applicable - No secrets are stored or managed by this tool
- User's Git credentials are handled by Git itself (SSH keys, credential helpers, etc.)

## External Command Dependencies

**System Commands Used:**
- `git` - For all version control operations (worktree management, branch operations, stash handling)
  - Minimum version: 2.7.0 (git worktree support introduced)
- `sed` - For text parsing in worktree list operations
- `wc` - For counting stash entries
- Standard shell builtins: `cd`, `echo`, `test`, `case`, `read`, etc.

## Webhooks & Callbacks

**Incoming:**
- None

**Outgoing:**
- None

---

*Integration audit: 2026-02-07*
