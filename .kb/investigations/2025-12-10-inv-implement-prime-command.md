**TLDR:** Question: How should the `agentlog prime` command output context for AI agent injection? Answer: Parse `.agentlog/errors.jsonl`, aggregate by error_type and source, output concise summary (human-readable by default, JSON with `--json`). High confidence (90%) - clear requirements, well-defined JSONL schema, straightforward implementation.

---

# Investigation: Implement prime command

**Question:** What should the `agentlog prime` command output and how should it be structured for AI agent context injection?

**Started:** 2025-12-10
**Updated:** 2025-12-10
**Owner:** Worker agent
**Phase:** Complete
**Next Step:** Create design document, proceed to TDD implementation
**Status:** Complete
**Confidence:** High (90%)

---

## Findings

### Finding 1: CLI scaffold uses Cobra with established patterns

**Evidence:** Root command at `internal/cmd/root.go` includes:
- `--json` flag for JSON output mode (`IsJSONOutput()`)
- `--ai-help` flag for machine-readable metadata (`printAIHelp()`)
- `IsTTY()` helper for terminal detection
- CommandMetadata struct pattern for AI-readable output

**Source:** `internal/cmd/root.go:11-137`

**Significance:** The prime command should follow these established patterns - use `IsJSONOutput()` to determine output format, support both human-readable and JSON modes.

---

### Finding 2: JSONL schema is well-defined

**Evidence:** Schema at `docs/jsonl-schema.md` specifies:
- Required fields: `timestamp`, `source`, `error_type`, `message`
- Optional `context` object with `stack_trace`, `session_id`, etc.
- Error type taxonomy: UNCAUGHT_ERROR, UNHANDLED_REJECTION, NETWORK_ERROR, etc.
- Source values: frontend, backend, cli, worker, test
- File location: `.agentlog/errors.jsonl`

**Source:** `docs/jsonl-schema.md:1-313`

**Significance:** The prime command can reliably parse the JSONL file and aggregate by `error_type` and `source` fields. Clear schema means no ambiguity in parsing.

---

### Finding 3: Prime command purpose is context injection

**Evidence:** From CLAUDE.md and spawn context:
- "Output context summary for agent injection"
- "Recent error count, top error types, actionable summary"
- "Used by orchestration hooks to inject context"

**Source:** Project CLAUDE.md, beads issue agentlog-7jx

**Significance:** Output must be concise (fits in prompt context), actionable (tells agent what to focus on), and parseable (works in both TTY and non-TTY scenarios).

---

## Synthesis

**Key Insights:**

1. **Follow existing patterns** - Use `--json` flag, `IsTTY()` detection, same file structure as root.go

2. **Aggregate meaningfully** - Group by error_type and source, sort by frequency, show recent timeframe

3. **Keep output concise** - Prime output goes into agent context, so brief but actionable

**Answer to Investigation Question:**

The `agentlog prime` command should:
1. Read `.agentlog/errors.jsonl` (handle missing file gracefully)
2. Aggregate errors by `error_type` and `source`
3. Output recent error count (last 24h), top 3 error types, top 3 sources
4. Human-readable by default, JSON with `--json` flag
5. Include actionable summary like "3 errors in last hour - focus on NETWORK_ERROR in frontend"

---

## Confidence Assessment

**Current Confidence:** High (90%)

**Why this level?**

Well-defined JSONL schema, clear CLI patterns to follow, straightforward aggregation logic.

**What's certain:**

- ✅ JSONL schema with required fields (timestamp, source, error_type, message)
- ✅ CLI patterns (--json flag, IsTTY detection)
- ✅ Purpose: concise context injection for AI agents

**What's uncertain:**

- ⚠️ Exact output format preferences (may need iteration)
- ⚠️ How large error files should be handled (streaming vs loading all)

**What would increase confidence to Very High:**

- User feedback on output format
- Testing with real-world error files

---

## Implementation Recommendations

**Purpose:** Bridge from investigation findings to actionable implementation.

### Recommended Approach: Single command file with JSONL parser

**Why this approach:**
- Follows existing CLI structure (`internal/cmd/prime.go`)
- Reuses global flags (`--json`)
- Self-contained parsing logic (no external deps)

**Trade-offs accepted:**
- Load entire file vs streaming (files rarely exceed 10MB per schema)
- Simple aggregation vs advanced analytics (v1 focused on core use case)

**Implementation sequence:**
1. Create ErrorEntry struct matching JSONL schema
2. Write JSONL parser function (handle malformed lines gracefully)
3. Write aggregation logic (by error_type, source, timeframe)
4. Create PrimeSummary struct for output
5. Implement human-readable and JSON formatters
6. Register command with Cobra

---

## References

**Files Examined:**
- `internal/cmd/root.go` - CLI patterns, global flags
- `internal/cmd/root_test.go` - Testing patterns
- `docs/jsonl-schema.md` - JSONL schema specification
- `cmd/agentlog/main.go` - Entry point

**Commands Run:**
```bash
# Check module structure
cat go.mod

# List Go files
find . -name "*.go"
```

---

## Investigation History

**2025-12-10 10:00:** Investigation started
- Initial question: How should prime command be implemented?
- Context: Spawned to implement prime command for agentlog CLI

**2025-12-10 10:15:** Investigation completed
- Final confidence: High (90%)
- Status: Complete
- Key outcome: Clear implementation path identified - follow existing CLI patterns, parse JSONL, aggregate and output summary
