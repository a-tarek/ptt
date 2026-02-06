# Phase 1: Internal Documentation Refresh - Context

**Gathered:** 2026-02-07
**Status:** Ready for planning

<domain>
## Phase Boundary

Update all 7 `.planning/codebase/` files (ARCHITECTURE, CONVENTIONS, STRUCTURE, CONCERNS, STACK, INTEGRATIONS, TESTING) to accurately reflect current `wt.zsh` implementation. This includes recent additions: `wt init`, `wt eject`, `--copy`/`--symlink` override flags, `.wtconfig` support, and the override mechanism in `_wt_setup`.

</domain>

<decisions>
## Implementation Decisions

### Documentation depth
- Full behavior per function: signature, args, return value, side effects, error cases
- Include call chains showing which functions call which and data flow between them (e.g. `wt new` → `_wt_setup` → reads `.wtconfig` → applies overrides)
- Not just a flat list of functions — show how they compose

### Documentation tone
- Both descriptive and prescriptive: describe the current pattern, then state the rule it implies
- Example: "Flag parsing uses a `while [[ "$1" == --* ]]` loop with `case` — new commands with flags should follow this pattern"

### Audience
- Dual audience: structured enough for AI agents (future sessions, GSD workflows), readable enough for human contributors
- Use clear headings, consistent formatting, and complete information — but keep it skimmable

### Claude's Discretion
- Exact formatting and section ordering within each file
- Whether to add diagrams or keep it text-only
- How to handle TESTING.md (no test suite exists — note current state and any testing approach)

</decisions>

<specifics>
## Specific Ideas

No specific requirements — open to standard approaches. The source of truth is `wt.zsh` itself; docs must match the code exactly.

</specifics>

<deferred>
## Deferred Ideas

None — discussion stayed within phase scope.

</deferred>

---

*Phase: 01-internal-documentation-refresh*
*Context gathered: 2026-02-07*
