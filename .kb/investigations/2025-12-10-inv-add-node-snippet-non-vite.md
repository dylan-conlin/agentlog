**TLDR:** Task: Add Node.js snippet for non-Vite TypeScript (BullMQ workers, scrapers, etc). Solution: Added new "node" stack type with snippet that writes directly to `.agentlog/errors.jsonl`, captures `process.on('uncaughtException')` and `process.on('unhandledRejection')`, and provides `logError()` function for pino/logger integration. High confidence (95%) - all 12 tests pass.

---

# Investigation: Add Node.js Snippet (Non-Vite TypeScript)

**Question:** How to add error capture for Node.js services (BullMQ workers, scrapers) that don't use Vite dev server?

**Started:** 2025-12-10
**Updated:** 2025-12-10
**Owner:** Worker agent
**Phase:** Complete
**Next Step:** None
**Status:** Complete
**Confidence:** Very High (95%)

---

## Findings

### Finding 1: Current TypeScript snippet assumes browser environment

**Evidence:** The existing `snippetTypeScript` uses browser APIs (`window.onerror`, `window.onunhandledrejection`) and POSTs to `/__agentlog` endpoint expecting Vite dev server middleware.

**Source:** `internal/cmd/init.go:461-500` (snippetTypeScript constant)

**Significance:** Node.js services can't use this snippet - no window object, no Vite dev server to handle POSTs.

---

### Finding 2: Other backend snippets (Go, Python, Rust) write directly to file

**Evidence:** Go, Python, and Rust snippets all write directly to `.agentlog/errors.jsonl` using native file I/O, not HTTP POSTs.

**Source:**
- `internal/cmd/init.go:502-546` (snippetGo)
- `internal/cmd/init.go:548-581` (snippetPython)
- `internal/cmd/init.go:583-623` (snippetRust)

**Significance:** Node.js snippet should follow same pattern - direct file write using `fs` module.

---

### Finding 3: Stack override mechanism already exists

**Evidence:** The `--stack` flag allows overriding auto-detection. Adding "node" as a valid stack value fits the existing pattern.

**Source:** `internal/cmd/init.go:97-105` (stack override logic)

**Significance:** Users can specify `--stack node` to get the Node.js snippet even if package.json is detected.

---

## Synthesis

**Key Insights:**

1. **Browser vs Server dichotomy** - TypeScript is unique in being used both in browsers (Vite/React) and servers (Node.js). Other languages are backend-only, so their snippets naturally use file I/O.

2. **No auto-detection needed** - Reliably distinguishing browser vs Node.js TypeScript projects is complex (no single marker file). Using `--stack node` explicit override is cleaner.

3. **Logger integration is key** - Node.js services often use structured loggers (pino, winston). The snippet should expose a `logError()` function that can be called from logger hooks.

**Answer to Investigation Question:**

Added new "node" stack type with:
- Direct file write to `.agentlog/errors.jsonl` using Node.js `fs` module
- `process.on('uncaughtException')` handler
- `process.on('unhandledRejection')` handler
- Exported `logError()` function for pino/winston integration
- `NODE_ENV === 'production'` check to no-op in production

---

## Confidence Assessment

**Current Confidence:** Very High (95%)

**Why this level?**

All 12 new tests pass, full test suite passes with no regressions, CLI builds and works correctly.

**What's certain:**

- ✅ Node.js snippet correctly captures uncaught exceptions
- ✅ Node.js snippet correctly captures unhandled rejections
- ✅ Direct file write works (tested via CLI)
- ✅ `--stack node` override works
- ✅ `--install` creates correct capture.ts file

**What's uncertain:**

- ⚠️ Real-world pino integration not tested (only documented pattern)

---

## Implementation

### Files Changed:

1. **`internal/cmd/init.go`**:
   - Added `snippetNode` constant (lines 502-587)
   - Added `nodeCapture` constant for installable version (lines 938-1026)
   - Added "node" case in `getSnippet()` (line 448-449)
   - Added "node" case in `installSnippets()` (line 179-180)
   - Added `installNodeSnippets()` function (lines 299-317)

2. **`internal/cmd/init_test.go`**:
   - Added 12 new tests for Node.js snippet (lines 1013-1201)

### Usage:

```bash
# Print snippet for manual copy
agentlog init --stack node

# Install capture.ts file directly
agentlog init --stack node --install
```

### Node.js Snippet Features:

- `logError(type, message, context)` - callable function for logger integration
- `initAgentlog()` - sets up process error handlers
- Writes to `.agentlog/errors.jsonl`
- No-ops when `NODE_ENV === 'production'`
- Truncates message (500 chars) and stack_trace (2048 bytes) per schema

---

## References

**Files Examined:**
- `internal/cmd/init.go` - Snippet definitions and init logic
- `internal/cmd/init_test.go` - Existing test patterns
- `internal/detect/stack.go` - Stack detection mechanism
- `docs/jsonl-schema.md` - JSONL schema requirements

---

## Investigation History

**2025-12-10 15:00:** Investigation started
- Read spawn context and existing codebase
- Identified that TypeScript snippet uses browser APIs

**2025-12-10 15:15:** Implementation started (TDD)
- Wrote 12 failing tests first
- Implemented snippetNode constant
- Implemented nodeCapture constant
- Added installNodeSnippets function
- All tests passing

**2025-12-10 15:30:** Investigation completed
- Final confidence: Very High (95%)
- Status: Complete
- Key outcome: Node.js snippet added with `--stack node` support
