# Design: agentlog prime command

**Status:** Pending Approval
**Author:** Worker agent
**Date:** 2025-12-10

---

## Problem Statement

AI agents working on development tasks need visibility into recent errors. The `agentlog prime` command outputs a concise context summary suitable for injection into agent prompts via hooks.

**Success criteria:**
1. Parses `.agentlog/errors.jsonl` reliably
2. Outputs human-readable summary by default, JSON with `--json`
3. Summary is concise enough for prompt context injection
4. Handles edge cases gracefully (missing file, empty file, malformed entries)

---

## Approach

### Overview

Create `internal/cmd/prime.go` following existing CLI patterns. Parse JSONL file, aggregate errors by type/source/timeframe, output summary.

### Data Structures

```go
// ErrorEntry matches JSONL schema required fields
type ErrorEntry struct {
    Timestamp string `json:"timestamp"`
    Source    string `json:"source"`
    ErrorType string `json:"error_type"`
    Message   string `json:"message"`
}

// PrimeSummary is the output structure for prime command
type PrimeSummary struct {
    TotalErrors     int              `json:"total_errors"`
    Last24hErrors   int              `json:"last_24h_errors"`
    LastHourErrors  int              `json:"last_hour_errors"`
    TopErrorTypes   []ErrorTypeCount `json:"top_error_types"`
    TopSources      []SourceCount    `json:"top_sources"`
    ActionableTip   string           `json:"actionable_tip"`
    GeneratedAt     string           `json:"generated_at"`
}

type ErrorTypeCount struct {
    ErrorType string `json:"error_type"`
    Count     int    `json:"count"`
}

type SourceCount struct {
    Source string `json:"source"`
    Count  int    `json:"count"`
}
```

### Output Formats

**Human-readable (default):**
```
agentlog: 12 errors (5 in last hour)
  Top types: UNCAUGHT_ERROR (7), NETWORK_ERROR (3), VALIDATION_ERROR (2)
  Sources: frontend (8), backend (4)
  Tip: Focus on UNCAUGHT_ERROR in frontend - 58% of recent errors
```

**JSON (`--json`):**
```json
{
  "total_errors": 12,
  "last_24h_errors": 12,
  "last_hour_errors": 5,
  "top_error_types": [
    {"error_type": "UNCAUGHT_ERROR", "count": 7},
    {"error_type": "NETWORK_ERROR", "count": 3}
  ],
  "top_sources": [
    {"source": "frontend", "count": 8},
    {"source": "backend", "count": 4}
  ],
  "actionable_tip": "Focus on UNCAUGHT_ERROR in frontend - 58% of recent errors",
  "generated_at": "2025-12-10T10:30:00Z"
}
```

**No errors case:**
```
agentlog: No errors logged
```

**Missing file case:**
```
agentlog: No error log found (.agentlog/errors.jsonl)
  Run 'agentlog init' to set up error tracking
```

### Implementation Steps

1. **ErrorEntry parsing** - Read JSONL line by line, unmarshal to ErrorEntry, skip malformed
2. **Time filtering** - Parse ISO 8601 timestamps, filter to last 24h/1h
3. **Aggregation** - Count by error_type and source, sort by frequency
4. **Actionable tip** - Generate tip based on highest frequency error type + source combination
5. **Output formatting** - Human-readable or JSON based on `IsJSONOutput()`
6. **Command registration** - Add to root command in `init()`

### File Structure

```
internal/
  cmd/
    root.go       (existing)
    prime.go      (new - command implementation)
    prime_test.go (new - tests)
```

---

## Testing Strategy

Using TDD - write failing tests first.

### Test Cases

1. **TestPrimeCommand_NoFile** - Missing `.agentlog/errors.jsonl` returns helpful message
2. **TestPrimeCommand_EmptyFile** - Empty file returns "No errors logged"
3. **TestPrimeCommand_SingleError** - Parses and displays single error
4. **TestPrimeCommand_MultipleErrors** - Aggregates correctly, shows top types/sources
5. **TestPrimeCommand_JSONOutput** - `--json` flag produces valid JSON
6. **TestPrimeCommand_MalformedLine** - Skips malformed JSONL lines gracefully
7. **TestParsePrimeSummary** - Unit test for parsing/aggregation logic

### Test Fixtures

Create test JSONL files in `testdata/` directory:
- `testdata/empty.jsonl` - Empty file
- `testdata/single.jsonl` - One error
- `testdata/multiple.jsonl` - Multiple errors, various types/sources
- `testdata/malformed.jsonl` - Mix of valid and invalid lines

---

## Security Considerations

- File path is hardcoded (`.agentlog/errors.jsonl`) - no path traversal risk
- Read-only operation - no modification to error log
- No sensitive data in output (error messages may be truncated per schema)

---

## Performance

- Files limited to 10MB per schema spec
- Line-by-line parsing, no full file load to memory
- O(n) parsing, O(k log k) sorting where k = unique error types/sources (small)

---

## Alternatives Considered

**Option B: Stream parser with channel**
- Pros: More memory efficient for huge files
- Cons: Over-engineered for 10MB limit
- **Not chosen:** File size limit makes simple buffered read sufficient

**Option C: SQLite for aggregation**
- Pros: More complex queries possible
- Cons: Adds dependency, over-engineered for v1
- **Not chosen:** Simple in-memory aggregation sufficient for use case

---

## Open Questions

None - straightforward implementation following established patterns.

---

## References

- JSONL Schema: `docs/jsonl-schema.md`
- CLI patterns: `internal/cmd/root.go`
- Investigation: `.kb/investigations/2025-12-10-inv-implement-prime-command.md`
