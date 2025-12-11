**TLDR:** Question: What should the README contain for first-time users? Answer: Quick start flow (install → init → copy snippet → agentlog errors) plus command reference and schema overview. High confidence (95%) - based on existing codebase, design investigation, and CLI implementation.

---

# Investigation: Write README for agentlog

**Question:** What should the README contain to be clear enough for first-time users?

**Started:** 2025-12-10
**Updated:** 2025-12-10
**Owner:** Worker agent (spawned from agentlog-05z)
**Phase:** Complete
**Next Step:** None
**Status:** Complete
**Confidence:** High (95%)

---

## Findings

### Finding 1: Quick start is the primary user journey

**Evidence:** Design investigation states "README is clear enough for first-time users" as success criteria. The init command auto-detects stack and outputs snippets, making the onboarding path: install → init → copy snippet → view errors.

**Source:** `.kb/investigations/2025-12-10-design-agentlog-architecture.md` (line 231-232)

**Significance:** README structure should lead with quick start, not feature details.

---

### Finding 2: Four language snippets are built into CLI

**Evidence:** init.go contains complete snippets for TypeScript, Go, Python, Rust embedded as const strings. These are output by `agentlog init`.

**Source:** `internal/cmd/init.go:196-358`

**Significance:** README can reference `agentlog init` for snippets rather than duplicating all four.

---

### Finding 3: CLI has AI-first features

**Evidence:** `--ai-help` outputs machine-readable JSON metadata. `--json` flag available globally. `agentlog prime` outputs context for agent injection. These are AI-first CLI patterns.

**Source:** `internal/cmd/root.go:91-143`

**Significance:** README should highlight AI agent benefits alongside human developer benefits.

---

## Implementation

**README created at:** `README.md`

**Structure:**
1. Tagline and positioning
2. Quick Start (4 steps)
3. Why agentlog (for devs + agents)
4. Commands table
5. How it works diagram
6. Supported stacks
7. JSONL schema overview
8. Development setup

**Validation:** Tests pass, README renders correctly.

---

## References

**Files Examined:**
- `CLAUDE.md` - Project context
- `.kb/investigations/2025-12-10-design-agentlog-architecture.md` - Full design doc
- `docs/jsonl-schema.md` - Schema specification
- `internal/cmd/root.go` - CLI structure and --ai-help
- `internal/cmd/init.go` - Snippets and init flow
- `internal/cmd/errors.go` - Errors command flags

---

## Investigation History

**2025-12-10:** Investigation started
- Initial question: What README content is clear for first-time users?
- Context: beads issue agentlog-05z

**2025-12-10:** Investigation completed
- Final confidence: High (95%)
- Status: Complete
- Key outcome: README.md created with quick start guide
