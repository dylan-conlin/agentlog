**TLDR:** Question: How to implement the 'errors' command for agentlog? Answer: Parse JSONL from .agentlog/errors.jsonl, support --limit/--source/--type/--since filters, dual output (human-readable default, --json for scripts). High confidence (90%) - clear schema spec and established CLI patterns.

---

# Investigation: Implement 'errors' Command

**Question:** How to implement the 'errors' command to query .agentlog/errors.jsonl with filtering?

**Started:** 2025-12-10
**Updated:** 2025-12-10
**Owner:** Worker agent (agentlog-ija)
**Phase:** Complete
**Next Step:** Design phase
**Status:** Complete
**Confidence:** High (90%)

---

## Findings

### Finding 1: CLI Scaffold Already Has Patterns

**Evidence:** Root command in internal/cmd/root.go has global --json flag, IsTTY() helper, and --ai-help metadata describing errors command with flags.

**Source:** internal/cmd/root.go:11-14 (global flags), :79-88 (IsTTY and IsJSONOutput)

**Significance:** Can follow established patterns. Errors command metadata already defined in --ai-help output (lines 107-114) - just need to implement matching behavior.

---

### Finding 2: JSONL Schema is Well-Defined

**Evidence:** docs/jsonl-schema.md (v1.0.0) defines:
- Required fields: timestamp, source, error_type, message
- Optional context object with stack_trace, session_id, etc.
- Size limits: 500 chars message, 2KB stack_trace, 10KB entry, 10MB file

**Source:** docs/jsonl-schema.md

**Significance:** Clear contract for parsing. Need to handle malformed lines gracefully (skip with warning per spec).

---

### Finding 3: Filter Requirements from Metadata

**Evidence:** The --ai-help output already specifies filter flags:
- --limit: Maximum errors to show (default 10)
- --source: Filter by source (frontend, backend, cli, worker, test)
- --type: Filter by error type
- --since: Time filter (e.g., '1h', '30m', '2024-01-01')

**Source:** internal/cmd/root.go:110-114

**Significance:** Clear requirements. --since needs duration parsing (Go's time.ParseDuration for '1h', '30m') plus date parsing for absolute dates.

---

## Synthesis

**Key Insights:**

1. **Follow established patterns** - The scaffold has dual output, TTY detection, consistent flag naming

2. **JSONL parsing is straightforward** - One JSON object per line, skip malformed with warning

3. **Time filtering needs care** - Support both relative durations ('1h') and absolute dates ('2024-01-01')

**Answer to Investigation Question:**

Implement as internal/cmd/errors.go with cobra subcommand. Parse JSONL line-by-line, apply filters, output human-readable table by default or JSON array with --json flag. High confidence given clear spec and patterns.

---

## Confidence Assessment

**Current Confidence:** High (90%)

**Why this level?**

Clear schema spec, established CLI patterns, well-defined filter requirements. Standard Go patterns for file parsing and time handling.

**What's certain:**

- ✅ JSONL schema and file location (.agentlog/errors.jsonl)
- ✅ Filter flags: --limit, --source, --type, --since
- ✅ Dual output pattern (human-readable / --json)

**What's uncertain:**

- ⚠️ Human-readable output format (table vs list?) - will decide in design
- ⚠️ Time parsing edge cases (timezones?)

---

## Implementation Recommendations

### Recommended Approach ⭐

**Standard cobra subcommand with JSONL stream parsing**

**Why this approach:**
- Follows existing patterns in codebase
- Stream parsing handles large files efficiently
- Clear separation: parsing → filtering → formatting

**Implementation sequence:**
1. Create internal/cmd/errors.go with cobra command
2. Implement JSONL parsing with error handling
3. Add filter logic (--since, --type, --source)
4. Implement human-readable and JSON output formatters

---

## References

**Files Examined:**
- internal/cmd/root.go - CLI scaffold with patterns
- docs/jsonl-schema.md - JSONL specification
- .kb/investigations/2025-12-10-inv-scaffold-cli-cobra.md - Prior scaffold work

---

## Deliverables

- `internal/cmd/errors.go` - Errors command implementation
- `internal/cmd/errors_test.go` - Comprehensive tests (unit + integration)
- `docs/designs/2025-12-10-errors-command.md` - Design document

## Investigation History

**2025-12-10:** Investigation started
- Analyzed CLI scaffold structure
- Read JSONL schema specification
- Identified patterns to follow

**2025-12-10:** Implementation complete
- Created design document
- Implemented errors command with TDD
- All tests passing
- CLI verified with real data
