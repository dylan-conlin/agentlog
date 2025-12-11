# agentlog JSONL Schema Specification

**Version:** 1.0.0
**Status:** Draft
**Last Updated:** 2025-12-10

This document defines the JSONL format for `.agentlog/errors.jsonl`. This is the universal contract between any application and the agentlog CLI.

---

## Overview

agentlog uses JSONL (JSON Lines) format - one JSON object per line. Each line represents a single error event.

**File Location:** `.agentlog/errors.jsonl` (project root)

**Encoding:** UTF-8

**Line Terminator:** `\n` (LF)

---

## Required Fields

Every error entry MUST include these fields:

| Field | Type | Description | Example |
|-------|------|-------------|---------|
| `timestamp` | string | ISO 8601 UTC datetime | `"2025-12-10T19:19:32.941Z"` |
| `source` | string | Origin of the error | `"frontend"` |
| `error_type` | string | Error classification | `"UNCAUGHT_ERROR"` |
| `message` | string | Human-readable description | `"Cannot read property 'foo'"` |

### Field Specifications

#### timestamp

- **Format:** ISO 8601 with milliseconds in UTC
- **Pattern:** `YYYY-MM-DDTHH:mm:ss.sssZ`
- **Required:** Yes
- **Example:** `"2025-12-10T19:19:32.941Z"`

```json
{"timestamp": "2025-12-10T19:19:32.941Z"}
```

#### source

- **Type:** String (enum recommended)
- **Required:** Yes
- **Recommended Values:**
  - `frontend` - Browser/client-side errors
  - `backend` - Server-side errors
  - `cli` - Command-line tool errors
  - `worker` - Background job/worker errors
  - `test` - Test runner errors
- **Custom:** Applications may use custom source values (e.g., `api-gateway`, `auth-service`)

```json
{"source": "frontend"}
```

#### error_type

- **Type:** String
- **Required:** Yes
- **Must be:** One of the values from the Error Type Taxonomy (see below)

```json
{"error_type": "UNCAUGHT_ERROR"}
```

#### message

- **Type:** String
- **Required:** Yes
- **Max Length:** 500 characters
- **Truncation:** Messages exceeding 500 chars SHOULD be truncated with `...` suffix

```json
{"message": "Cannot read property 'foo' of undefined"}
```

---

## Error Type Taxonomy

Use these standardized error types for consistency across languages and frameworks.

### Universal (All Sources)

| Type | When to Use |
|------|-------------|
| `UNEXPECTED_ERROR` | Catch-all for unclassified errors |
| `VALIDATION_ERROR` | Input/data validation failures |

### Frontend-Specific

| Type | When to Use |
|------|-------------|
| `UNCAUGHT_ERROR` | Uncaught exceptions (`window.onerror`) |
| `UNHANDLED_REJECTION` | Unhandled promise rejections |
| `NETWORK_ERROR` | Fetch/XHR failures, API errors |
| `RENDER_ERROR` | React/Vue/Svelte component render errors |

### Backend-Specific

| Type | When to Use |
|------|-------------|
| `REQUEST_ERROR` | HTTP request handling errors |
| `WEBSOCKET_ERROR` | WebSocket connection/message errors |
| `DATABASE_ERROR` | Database query/connection errors |

### CLI-Specific

| Type | When to Use |
|------|-------------|
| `COMMAND_ERROR` | Command execution failures |
| `CONFIG_ERROR` | Configuration parsing/loading errors |

### Runtime-Specific

| Type | When to Use |
|------|-------------|
| `PANIC` | Go panics, Rust panics |
| `EXCEPTION` | Language exceptions |
| `TIMEOUT` | Operation timeouts |

---

## Optional Context Fields

Additional context MAY be included in a `context` object:

```json
{
  "timestamp": "2025-12-10T19:19:32.941Z",
  "source": "frontend",
  "error_type": "UNCAUGHT_ERROR",
  "message": "Cannot read property 'foo'",
  "context": {
    "session_id": "m1a2b3c4d5",
    "stack_trace": "Error: Cannot read property...\n    at foo.js:42:15",
    "url": "/dashboard",
    "component": "UserList"
  }
}
```

### Common Context Fields

| Field | Type | Max Size | When to Use |
|-------|------|----------|-------------|
| `session_id` | string | 64 chars | Correlate errors across requests/pages |
| `stack_trace` | string | 2KB | Full error stack trace |
| `url` | string | 500 chars | Frontend: current page URL |
| `endpoint` | string | 500 chars | Backend: API endpoint |
| `command` | string | 500 chars | CLI: command that failed |
| `component` | string | 100 chars | UI component name |
| `user_id` | string | 100 chars | User identifier (if applicable) |
| `request_id` | string | 100 chars | HTTP request correlation ID |
| `file` | string | 200 chars | Source file path |
| `line` | integer | - | Line number |
| `column` | integer | - | Column number |

### Custom Context

Applications MAY include additional context fields. Custom fields SHOULD:
- Use snake_case naming
- Be JSON-serializable primitives (string, number, boolean, null)
- Avoid nested objects deeper than 2 levels

```json
{
  "context": {
    "session_id": "m1a2b3c4d5",
    "custom_field": "value",
    "feature_flag": true,
    "attempt_count": 3
  }
}
```

---

## Size Limits

These limits ensure predictable memory usage and prevent runaway log files:

| Limit | Value | Behavior |
|-------|-------|----------|
| Max `message` length | 500 characters | Truncate with `...` |
| Max `stack_trace` length | 2KB (2048 bytes) | Truncate with `...` |
| Max total entry size | 10KB (10240 bytes) | Reject/drop entry |
| Max file size | 10MB (10485760 bytes) | Rotate to `errors.1.jsonl` |

### Truncation Rules

When truncating fields:
1. Truncate to limit - 3 characters
2. Append `...` suffix
3. Ensure UTF-8 boundary is not broken

```javascript
// Example truncation (500 char limit)
const truncated = message.length > 500
  ? message.slice(0, 497) + '...'
  : message;
```

### File Rotation

When `.agentlog/errors.jsonl` exceeds 10MB:
1. Rename to `.agentlog/errors.1.jsonl` (overwrite if exists)
2. Create new empty `.agentlog/errors.jsonl`
3. Continue writing to new file

Rotation is handled by the agentlog CLI during `tail` or write operations, not by snippets.

---

## Complete Examples

### Minimal Entry (Required Fields Only)

```json
{"timestamp":"2025-12-10T19:19:32.941Z","source":"frontend","error_type":"UNCAUGHT_ERROR","message":"Cannot read property 'foo' of undefined"}
```

### Frontend Error with Context

```json
{"timestamp":"2025-12-10T19:19:32.941Z","source":"frontend","error_type":"UNHANDLED_REJECTION","message":"Failed to fetch user data","context":{"session_id":"m1a2b3c4d5","url":"/dashboard","stack_trace":"Error: Failed to fetch user data\n    at UserList.svelte:42:15\n    at async load (routes/dashboard/+page.ts:12:3)","component":"UserList"}}
```

### Backend Error with Context

```json
{"timestamp":"2025-12-10T19:20:15.123Z","source":"backend","error_type":"DATABASE_ERROR","message":"Connection refused to database","context":{"endpoint":"/api/users","request_id":"req_abc123","stack_trace":"Error: Connection refused\n    at Pool.connect (pg.js:123:5)"}}
```

### CLI Error with Context

```json
{"timestamp":"2025-12-10T19:21:00.456Z","source":"cli","error_type":"COMMAND_ERROR","message":"Unknown command 'foo'","context":{"command":"agentlog foo --verbose"}}
```

### Network Error (Frontend)

```json
{"timestamp":"2025-12-10T19:22:30.789Z","source":"frontend","error_type":"NETWORK_ERROR","message":"POST /api/users failed: 500 Internal Server Error","context":{"session_id":"m1a2b3c4d5","url":"/settings","endpoint":"/api/users"}}
```

---

## Validation Rules

Implementations SHOULD validate entries before writing:

### Required
- [ ] `timestamp` is present and valid ISO 8601 UTC
- [ ] `source` is present and non-empty
- [ ] `error_type` is present and from taxonomy (warn if unknown)
- [ ] `message` is present and non-empty

### Recommended
- [ ] `message` does not exceed 500 characters
- [ ] `context.stack_trace` does not exceed 2KB
- [ ] Total entry size does not exceed 10KB
- [ ] All strings are valid UTF-8

### Invalid Entry Handling

If an entry fails validation:
1. **Snippets:** Log warning to stderr, skip writing invalid entry
2. **SDKs:** Attempt to fix (truncate, default values), warn if unable
3. **CLI:** When reading, skip malformed lines with warning

---

## Compatibility Notes

### JSON Formatting

- One JSON object per line (no pretty-printing)
- No trailing commas
- Double quotes for strings only
- No comments

### File System

- Ensure atomic writes (write to temp file, rename)
- Handle concurrent writes gracefully (file locking optional)
- Create `.agentlog/` directory if missing

### Production Mode

Snippets SHOULD no-op in production:
- Check `NODE_ENV !== 'production'` (Node.js)
- Check `!os.Getenv("PRODUCTION")` (Go)
- Check `os.environ.get('ENV') != 'production'` (Python)

---

## Changelog

### 1.0.0 (2025-12-10)
- Initial specification
- Required fields: timestamp, source, error_type, message
- Error type taxonomy defined
- Size limits established
- Context field guidelines
