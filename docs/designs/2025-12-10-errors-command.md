# Design: errors Command

**Status:** Approved
**Created:** 2025-12-10
**Author:** Worker agent (agentlog-ija)

---

## Problem Statement

Implement the `agentlog errors` command to query `.agentlog/errors.jsonl` with filtering capabilities. This is the core command for agentlog - agents need to quickly query recent errors.

**Success Criteria:**
- Query errors from `.agentlog/errors.jsonl`
- Filter by: --limit, --source, --type, --since
- Human-readable output by default
- JSON output with --json flag
- Graceful handling of missing file, malformed entries

---

## Approach

### Architecture

```
errors.go (cobra cmd)
    ↓
ErrorEntry struct (data model)
    ↓
readErrors() → filterErrors() → formatOutput()
```

### File Structure

```
internal/cmd/errors.go       # Cobra command + core logic
internal/cmd/errors_test.go  # Tests
```

### Data Model

```go
// ErrorEntry represents a single error from errors.jsonl
type ErrorEntry struct {
    Timestamp  string                 `json:"timestamp"`
    Source     string                 `json:"source"`
    ErrorType  string                 `json:"error_type"`
    Message    string                 `json:"message"`
    Context    map[string]interface{} `json:"context,omitempty"`
}
```

---

## Flags

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| --limit | int | 10 | Maximum errors to show |
| --source | string | "" | Filter by source (frontend, backend, etc.) |
| --type | string | "" | Filter by error_type |
| --since | string | "" | Show errors since time |

### --since Parsing

Support two formats:
1. **Duration:** `1h`, `30m`, `24h` → Parse with time.ParseDuration
2. **Date:** `2024-01-01` → Parse as RFC3339 date

```go
func parseSince(since string) (time.Time, error) {
    // Try duration first (relative)
    if dur, err := time.ParseDuration(since); err == nil {
        return time.Now().Add(-dur), nil
    }
    // Try date formats
    formats := []string{
        "2006-01-02",           // YYYY-MM-DD
        "2006-01-02T15:04:05Z", // RFC3339
    }
    for _, f := range formats {
        if t, err := time.Parse(f, since); err == nil {
            return t, nil
        }
    }
    return time.Time{}, fmt.Errorf("invalid time format: %s", since)
}
```

---

## Output Formats

### Human-Readable (Default)

```
Error: Cannot read property 'foo' of undefined
  Source: frontend | Type: UNCAUGHT_ERROR
  Time: 2025-12-10T19:19:32Z

Error: Connection refused to database
  Source: backend | Type: DATABASE_ERROR
  Time: 2025-12-10T19:20:15Z

Showing 2 of 42 errors (use --limit to see more)
```

### JSON Output (--json)

```json
[
  {
    "timestamp": "2025-12-10T19:19:32.941Z",
    "source": "frontend",
    "error_type": "UNCAUGHT_ERROR",
    "message": "Cannot read property 'foo' of undefined"
  }
]
```

---

## Error Handling

| Scenario | Behavior |
|----------|----------|
| File missing | "No errors file found. Run 'agentlog init' to set up." |
| File empty | "No errors recorded yet." |
| Malformed line | Skip with warning to stderr, continue |
| Invalid --since | Error with usage hint |

---

## Testing Strategy

1. **Unit tests for parsing:**
   - parseSince() with durations and dates
   - ErrorEntry JSON parsing
   - Filter logic

2. **Integration tests:**
   - Command with various flags
   - Missing file handling
   - Malformed JSONL handling

3. **Test fixtures:**
   - Valid JSONL with multiple entries
   - JSONL with malformed lines
   - Empty file

---

## Implementation Plan

1. Create ErrorEntry struct and parsing
2. Implement readErrors() - read and parse JSONL
3. Implement filterErrors() - apply --since, --source, --type
4. Implement formatOutput() - human-readable and JSON
5. Wire up cobra command with flags
6. Add to root command

---

## Security Considerations

- No execution of file contents
- Bound memory usage (read line-by-line, not whole file)
- No path traversal (fixed file location)

---

## Alternatives Considered

**Option B: Separate parser package**
- Pros: Reusable parser
- Cons: Over-engineering for single command
- Decision: Keep in errors.go, extract later if needed

**Option C: SQLite storage**
- Pros: Better querying
- Cons: Adds dependency, breaks JSONL convention
- Decision: Stick with JSONL per architecture decision
