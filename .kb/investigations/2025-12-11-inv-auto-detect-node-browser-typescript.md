**TLDR:** How to auto-detect Node.js vs browser TypeScript projects? Added heuristics checking browser indicators (vite.config.*, src/App.tsx, react/vue/svelte deps) vs Node.js indicators (express/fastify/bullmq deps, ts-node/tsx scripts, tsconfig module:commonjs/node16/nodenext). Browser indicators take priority in mixed cases. Very High confidence (95%) - 22 tests pass.

---

# Investigation: Auto-detect Node.js vs Browser TypeScript

**Question:** How to auto-detect whether a TypeScript project (detected via package.json) is browser-based or Node.js server-side?

**Started:** 2025-12-11
**Updated:** 2025-12-11
**Owner:** Worker agent
**Phase:** Complete
**Next Step:** None
**Status:** Complete
**Confidence:** Very High (95%)

---

## Findings

### Finding 1: Browser projects have distinctive markers

**Evidence:** Browser/frontend TypeScript projects consistently have one or more of:
- Config files: `vite.config.ts`, `next.config.js`, `nuxt.config.ts`
- Entry points: `src/App.tsx`, `src/App.jsx`
- Framework deps: `react`, `vue`, `svelte`, `@angular/core`, `next`, `nuxt`

**Source:** Analysis of common project structures and spawn context requirements

**Significance:** These markers are highly reliable - having any of them almost certainly means browser project.

---

### Finding 2: Node.js projects have server-side indicators

**Evidence:** Node.js/server-side TypeScript projects consistently have:
- Backend framework deps: `express`, `fastify`, `hono`, `koa`, `@nestjs/core`
- Queue/worker deps: `bullmq`, `bull`, `bee-queue`, `agenda`
- Scripts: Commands using `ts-node`, `tsx`, or `node`
- tsconfig module: `commonjs`, `nodenext`, `node16`

**Source:** Analysis of spawn context and common Node.js project patterns

**Significance:** These indicators are reliable for pure server-side projects.

---

### Finding 3: Mixed projects should default to browser

**Evidence:** Full-stack projects may have both frontend and backend deps (e.g., React + Express). In these cases:
- The frontend typically uses the dev server (Vite middleware for error capture)
- The backend runs separately and may need manual configuration

**Source:** spawn context guidance and investigation of full-stack patterns

**Significance:** Browser indicators should take priority to avoid breaking frontend error capture in full-stack projects.

---

## Synthesis

**Key Insights:**

1. **Priority-based detection** - Check browser indicators first (files, then deps), then Node.js indicators (deps, scripts, tsconfig). Browser wins in conflicts.

2. **Safe default** - Unknown projects default to TypeScript (browser) since that's the more common use case and Vite middleware handles the capture cleanly.

3. **Existing "node" stack support** - The CLI already had `--stack node` manual override and Node.js snippets. This just adds auto-detection.

**Answer to Investigation Question:**

Added `detectTypeScriptVariant(dir string) Stack` function in `internal/detect/stack.go` that:
1. Checks for browser config files (vite.config.*, next.config.*, etc.)
2. Checks package.json for browser framework deps
3. Checks package.json for Node.js framework deps
4. Checks npm scripts for ts-node/tsx/node commands
5. Checks tsconfig.json module setting
6. Defaults to TypeScript (browser)

---

## Confidence Assessment

**Current Confidence:** Very High (95%)

**Why this level?**

All 22 new tests pass, full test suite passes with no regressions, CLI correctly detects all tested project types.

**What's certain:**

- ✅ Browser framework deps correctly detected (react, vue, svelte, angular, next)
- ✅ Node.js framework deps correctly detected (express, fastify, hono, bullmq)
- ✅ Browser indicators take priority over Node.js in mixed cases
- ✅ tsconfig module settings correctly detected
- ✅ npm script commands correctly detected

**What's uncertain:**

- ⚠️ Edge cases with unconventional project structures
- ⚠️ Monorepo subdirectory detection may need refinement

---

## Implementation

### Files Changed:

1. **`internal/detect/stack.go`**:
   - Added `Node` Stack constant
   - Added `browserFrameworks` slice (react, vue, svelte, etc.)
   - Added `nodeFrameworks` slice (express, fastify, bullmq, etc.)
   - Added `nodeModuleSettings` slice (commonjs, nodenext, node16)
   - Added `detectTypeScriptVariant(dir string) Stack` function
   - Modified `detectInDir` to call variant detection for TypeScript

2. **`internal/detect/stack_test.go`**:
   - Added `TestDetectNodeVsBrowserTypeScript` with 22 test cases
   - Updated `TestStackString` to include Node

### Detection Priority:

1. Browser config files (vite.config.*, next.config.*, src/App.tsx)
2. Browser framework dependencies (react, vue, svelte, @angular/core, next)
3. Node.js framework dependencies (express, fastify, bullmq, hono)
4. npm scripts (ts-node, tsx, node commands)
5. tsconfig.json module setting (commonjs, nodenext, node16)
6. Default: TypeScript (browser)

---

## References

**Files Examined:**
- `internal/detect/stack.go` - Stack detection logic
- `internal/detect/stack_test.go` - Existing test patterns
- `internal/cmd/init.go` - How stacks map to snippets
- `.kb/investigations/2025-12-10-inv-add-node-snippet-non-vite.md` - Prior Node.js snippet work

---

## Investigation History

**2025-12-11 09:10:** Investigation started
- Read spawn context and codebase
- Identified existing stack detection mechanism

**2025-12-11 09:15:** TDD implementation started
- Wrote 22 failing tests for Node.js vs browser detection
- Implemented detectTypeScriptVariant function
- All tests passing

**2025-12-11 09:22:** Validation complete
- Full test suite passes with no regressions
- CLI correctly detects BullMQ → node, React → typescript, Express → node
- Investigation complete
