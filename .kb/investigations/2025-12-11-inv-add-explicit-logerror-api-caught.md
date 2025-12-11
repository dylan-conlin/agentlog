**TLDR:** Task: Add explicit logError() API to TypeScript snippet for caught errors. Currently, the TypeScript snippet only captures uncaught errors via window.onerror/onunhandledrejection. Node.js snippet already has logError() - need to add equivalent to browser snippet. High confidence (90%).

---

# Investigation: Add Explicit logError() API for Caught Errors

**Question:** How should we add an explicit logError() API to the TypeScript snippet so developers can log caught errors (try/catch scenarios)?

**Started:** 2025-12-11
**Updated:** 2025-12-11
**Owner:** worker-agent
**Phase:** Complete
**Next Step:** None
**Status:** Complete
**Confidence:** High (90%)

---

## Findings

### Finding 1: TypeScript snippet lacks explicit logging API

**Evidence:** Current `snippetTypeScript` (init.go:485-524) only has a private `log` function inside the if block, used internally by window.onerror and window.onunhandledrejection handlers. No exported function for manual error logging.

**Source:** `/Users/dylanconlin/Documents/personal/agentlog/internal/cmd/init.go:485-524`

**Significance:** Developers cannot log caught errors in try/catch blocks - only uncaught errors are captured.

---

### Finding 2: Node.js snippet already has logError() pattern

**Evidence:** `snippetNode` (init.go:526-627) exports `logError(errorType: string, message: string, context?: Record<string, unknown>)` function that developers can call directly. This is the pattern to follow.

**Source:** `/Users/dylanconlin/Documents/personal/agentlog/internal/cmd/init.go:544-590`

**Significance:** Proven API design exists - just need to adapt for browser context.

---

### Finding 3: Test patterns exist for snippet validation

**Evidence:** Tests use `strings.Contains()` to verify snippet behavior - check for required fields, error handlers, dev mode checks. Tests for Node.js snippet verify `logError` export.

**Source:** `/Users/dylanconlin/Documents/personal/agentlog/internal/cmd/init_test.go:282-335`

**Significance:** Can follow same test patterns for new logError() API.

---

## Implementation Recommendations

### Recommended Approach: Export logError function from TypeScript snippet

**Why this approach:**
- Matches existing Node.js snippet API
- Users can call `logError('CAUGHT_ERROR', error.message, { stack_trace: error.stack })`
- Maintains existing window.onerror/onunhandledrejection behavior

**Implementation sequence:**
1. Write failing test for logError export in TypeScript snippet
2. Add exported logError function to snippetTypeScript
3. Ensure existing tests still pass

**Trade-offs accepted:**
- Function added to window global (unavoidable in browser context without bundler)
- Consistent with current snippet design pattern

---

## References

**Files Examined:**
- `/Users/dylanconlin/Documents/personal/agentlog/internal/cmd/init.go:485-627` - TypeScript and Node.js snippets
- `/Users/dylanconlin/Documents/personal/agentlog/internal/cmd/init_test.go` - Test patterns

---

## Investigation History

**2025-12-11:** Investigation started
- Initial question: How to add logError() API for caught errors
- Context: Currently only uncaught errors captured in TypeScript snippet

**2025-12-11:** Analysis complete
- Found Node.js pattern to follow
- Ready for TDD implementation
