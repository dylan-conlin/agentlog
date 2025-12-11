**TLDR:** Task: Create browser-compatible TypeScript snippet for agentlog error capture. Outcome: Created ~30-line snippet with browser error handlers (fetch-based) and Vite dev server plugin. High confidence (95%) - tests pass, follows JSONL schema.

---

# Investigation: Create TypeScript Snippet

**Question:** How to create a browser-compatible TypeScript snippet for capturing errors in webapps?

**Started:** 2025-12-10
**Updated:** 2025-12-10
**Owner:** Worker agent (agentlog-n7o)
**Phase:** Complete
**Next Step:** None
**Status:** Complete
**Confidence:** Very High (95%)

---

## Findings

### Finding 1: Existing snippet was browser-incompatible

**Evidence:** The original TypeScript snippet in init.go used `require('fs')` and `require('path')` which are Node.js modules not available in browsers.

**Source:** `internal/cmd/init.go:196-225` (original snippet)

**Significance:** The snippet needed complete rewrite to work in actual webapp environments. Browser code cannot directly write to filesystem.

---

### Finding 2: Fetch API + dev server pattern is the solution

**Evidence:** Modern webapps use dev servers (Vite, Next.js, etc.). Browser code can POST to a local endpoint, and a dev server middleware can write to the JSONL file.

**Source:** JSONL schema docs, Vite plugin API documentation

**Significance:** This pattern provides complete end-to-end error capture: browser captures errors, dev server writes to file.

---

## Implementation

**Files modified:**
- `internal/cmd/init.go` - Updated `snippetTypeScript` constant with browser + server code
- `internal/cmd/init_test.go` - Added 4 new tests for TypeScript snippet validation

**Tests added:**
1. `TestTypeScriptSnippet_BrowserCompatible` - Verifies no Node.js requires, uses fetch
2. `TestTypeScriptSnippet_RequiredJSONLFields` - Verifies all JSONL required fields present
3. `TestTypeScriptSnippet_ErrorHandlers` - Verifies window.onerror and unhandledrejection
4. `TestTypeScriptSnippet_DevModeCheck` - Verifies production no-op check

**Snippet structure:**
1. **Browser section (~15 lines):** `window.onerror` and `onunhandledrejection` handlers that POST to `/__agentlog`
2. **Server section (~15 lines):** Vite plugin that handles POST and appends to `.agentlog/errors.jsonl`

---

## References

**Files Examined:**
- `docs/jsonl-schema.md` - JSONL format specification
- `internal/cmd/init.go` - Existing snippet implementation
- `internal/cmd/init_test.go` - Existing test patterns

**Commands Run:**
```bash
# Run tests (TDD approach)
go test -v -run "TestTypeScript|TestInit" ./internal/cmd/...

# Build and verify output
go build -o /tmp/agentlog ./cmd/agentlog && /tmp/agentlog init --stack typescript
```

---

## Investigation History

**2025-12-10 19:00:** Investigation started
- Task: Create browser-compatible TypeScript snippet
- Context: Original snippet used Node.js fs module (not browser compatible)

**2025-12-10 19:03:** TDD tests written (RED)
- Added 4 tests verifying browser compatibility requirements
- Tests failed as expected (original snippet used require('fs'))

**2025-12-10 19:04:** Implementation complete (GREEN)
- Updated snippetTypeScript with fetch-based browser code
- Added Vite plugin for dev server file writing
- All tests pass

**2025-12-10 19:05:** Investigation completed
- Final confidence: Very High (95%)
- Status: Complete
- Key outcome: Browser-compatible TypeScript snippet with dev server plugin
