**TLDR:** How should agentlog log its own errors for dogfooding? Create an `internal/self/` package with self-logging functions that write to `.agentlog/errors.jsonl` with source="cli", called from cmd handlers at error points. Must avoid infinite loops by not logging self-logging failures.

---

# Investigation: Dogfood agentlog error logging

**Question:** How should agentlog CLI log its own errors to `.agentlog/errors.jsonl`?

**Started:** 2025-12-10
**Updated:** 2025-12-10
**Owner:** Worker agent
**Phase:** Complete
**Next Step:** None
**Status:** Complete
**Confidence:** High (90%)

---

## Findings

### Finding 1: Go snippet is a template, not reusable code

**Evidence:** The `snippetGo` constant in `internal/cmd/init.go:237-281` is a string template meant to be copied by users, not a callable function.

**Source:** `internal/cmd/init.go:237-281`

**Significance:** Need to create a real implementation in a new package that agentlog can use internally.

---

### Finding 2: Error points are concentrated in cmd handlers

**Evidence:** All commands (init, errors, tail, doctor, prime) follow pattern of:
1. `os.Getwd()` failure - all commands
2. File I/O failures - `os.Open()`, `os.WriteFile()`, etc.
3. JSON parsing failures - `json.Unmarshal()` in errors.go, tail.go, doctor.go

**Source:**
- `internal/cmd/init.go:49,95,104,114,130`
- `internal/cmd/errors.go:63,130,140`
- `internal/cmd/tail.go:41,118,124`
- `internal/cmd/doctor.go:60`

**Significance:** Can add self-logging at these existing error return points.

---

### Finding 3: JSONL schema already defined

**Evidence:** Schema documented in `docs/jsonl-schema.md` with:
- Required fields: timestamp, source, error_type, message
- Size limits: message 500 chars, stack_trace 2KB
- Atomic writes, file rotation at 10MB

**Source:** `docs/jsonl-schema.md`, snippet constants in `init.go`

**Significance:** Self-logging should use source="cli" and match existing schema exactly.

---

## Synthesis

**Key Insights:**

1. **New package needed** - Create `internal/self/log.go` with actual Go functions (not template strings)

2. **No-op on failure** - If self-logging fails (e.g., .agentlog doesn't exist), must silently fail - no infinite loops

3. **Source distinction** - Use source="cli" to distinguish agentlog's own errors from user app errors

**Answer to Investigation Question:**

Create `internal/self/log.go` with:
- `LogError(errType, message string, err error)` - main entry point
- No-op if `.agentlog/` doesn't exist or if logging fails
- Uses source="cli" in JSONL
- Called from cmd handlers before returning errors

---

## Implementation Recommendations

### Recommended Approach ⭐

**TDD implementation of internal/self package**

**Why this approach:**
- Reusable across all cmd handlers
- Isolated from cmd logic - easy to test
- Follows project TDD mode requirement

**Implementation sequence:**
1. Write tests for self-logging in `internal/self/log_test.go`
2. Implement `internal/self/log.go` to pass tests
3. Add self-logging calls to cmd handlers (init, errors, tail, doctor, prime)
4. Validate with test run

---

## References

**Files Examined:**
- `internal/cmd/init.go` - Go snippet template and error patterns
- `internal/cmd/errors.go` - JSON parsing and file reading errors
- `internal/cmd/doctor.go` - Health check error patterns
- `docs/jsonl-schema.md` - JSONL schema specification

---

## Investigation History

**2025-12-10 19:30:** Investigation started
- Initial question: How to dogfood agentlog error logging
- Context: Beads issue agentlog-pxr

**2025-12-10 19:35:** Codebase exploration complete
- Found Go snippet is template, not callable
- Identified all error points in cmd handlers
- Determined need for internal/self package

**2025-12-10 19:40:** Investigation complete, moving to implementation
- Final confidence: High (90%)
- Key outcome: Create internal/self/log.go with TDD
