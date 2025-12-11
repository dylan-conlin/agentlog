**TLDR:** Task: Create Ruby/Rails snippet for agentlog with test coverage. Outcome: Implemented Rack middleware snippet with Rails.env.development? check, writes to .agentlog/errors.jsonl with source='backend'. Added 10 tests (7 snippet + 3 detection). All 84 tests pass. High confidence (95%) - follows established patterns.

---

# Investigation: Create Ruby/Rails Snippet

**Question:** How to implement Ruby/Rails error logging snippet that integrates with Rails exception handling?

**Started:** 2025-12-10
**Updated:** 2025-12-10
**Owner:** Worker agent (agentlog-bhe)
**Phase:** Complete
**Next Step:** None
**Status:** Complete
**Confidence:** Very High (95%)

---

## Findings

### Finding 1: Existing snippet patterns are well-established

**Evidence:** All existing snippets (TypeScript, Go, Python, Rust) follow the same pattern:
- Production mode check (env var or config)
- Required JSONL fields (timestamp, source, error_type, message)
- Context with stack_trace (truncated to 2048 chars)
- Message truncation to 500 chars
- Writes to `.agentlog/errors.jsonl`
- Uses only stdlib/core libraries

**Source:** `internal/cmd/init.go:202-364`

**Significance:** Ruby snippet must follow this established pattern for consistency.

---

### Finding 2: Stack detection requires Gemfile marker

**Evidence:** Stack detection uses marker files:
- TypeScript: `package.json`
- Go: `go.mod`
- Python: `pyproject.toml` or `requirements.txt`
- Rust: `Cargo.toml`

Ruby needs `Gemfile` as its marker file.

**Source:** `internal/detect/stack.go:30-40`

**Significance:** Need to add Ruby to stack detection system.

---

### Finding 3: Test coverage patterns are established

**Evidence:** Each snippet has dedicated tests checking:
- Required JSONL fields present
- Exception/panic handler mechanism
- Dev mode check
- Stdlib-only dependencies
- Correct file path

**Source:** `internal/cmd/init_test.go:280-523`

**Significance:** Ruby snippet tests should follow the same pattern with 5+ tests.

---

## Design Decision

### Approach: Rack Middleware

**Why middleware over rescue_from or Notifications:**
- **rescue_from**: Only captures controller errors, misses routing errors, middleware errors, background jobs
- **ActiveSupport::Notifications**: Event-based, more complex, requires subscribing to specific events
- **Rack middleware**: Catches ALL request-level exceptions at the lowest level

**Rails integration pattern:**
```ruby
# config/initializers/agentlog.rb
Rails.application.config.middleware.insert(0, Agentlog::ExceptionCatcher) if Rails.env.development?
```

**Snippet components:**
1. `Rails.env.development?` check (standard Rails idiom)
2. Rack middleware class wrapping `app.call(env)`
3. JSON serialization using stdlib `json` gem
4. File writing to `.agentlog/errors.jsonl`
5. Stack trace from exception backtrace

---

## Implementation Plan

**Files to modify:**
1. `internal/detect/stack.go` - Add Ruby stack constant and Gemfile marker
2. `internal/cmd/init.go` - Add `snippetRuby` constant and update `getSnippet`
3. `internal/cmd/init_test.go` - Add Ruby snippet tests
4. `internal/detect/stack_test.go` - Add Ruby detection test

**TDD sequence:**
1. Write failing tests for Ruby snippet (RED)
2. Implement Ruby snippet to pass tests (GREEN)
3. Refactor if needed (REFACTOR)

---

## References

**Files Examined:**
- `internal/cmd/init.go` - Existing snippet implementations
- `internal/cmd/init_test.go` - Test patterns for snippets
- `internal/detect/stack.go` - Stack detection system
- `docs/jsonl-schema.md` - JSONL format specification

**Related Artifacts:**
- **Investigation:** `.kb/investigations/2025-12-10-inv-create-python-snippet.md` - Similar task for Python

---

## Investigation History

**2025-12-10:** Investigation started
- Task: Create Ruby/Rails snippet with test coverage
- Context: Part of MVP snippet language support (TypeScript, Go, Python, Rust, now Ruby)

**2025-12-10:** Design completed
- Approach: Rack middleware for comprehensive exception capture
- Following TDD pattern established by other snippets

**2025-12-10:** Implementation completed (TDD)
- Added Ruby stack constant and Gemfile marker to detection
- Created snippetRuby constant with Rack middleware
- Added 7 Ruby snippet tests + 3 detection tests
- All 84 tests pass
- Final confidence: Very High (95%)
- Status: Complete
