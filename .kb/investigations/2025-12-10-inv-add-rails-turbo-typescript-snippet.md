**TLDR:** Task: Add Rails/Turbo TypeScript snippet since current snippet assumes Vite. Approach: Add Rails as new stack type with Gemfile detection, separate browser (Turbo-compatible) and server (Rails controller) snippets. High confidence (90%) - straightforward extension of existing patterns.

---

# Investigation: Add Rails/Turbo TypeScript Snippet

**Question:** How to add Rails/Turbo support to agentlog when the current TypeScript snippet assumes Vite?

**Started:** 2025-12-10
**Updated:** 2025-12-10
**Owner:** Worker agent
**Phase:** Complete
**Next Step:** None - implementation complete
**Status:** Complete
**Confidence:** High (90%)

---

## Findings

### Finding 1: Current TypeScript snippet is Vite-specific

**Evidence:**
- Browser code uses `import.meta.env?.DEV` for development check (Vite-specific API)
- Server code provides a Vite plugin with `configureServer()` hook
- Comment explicitly says "vite.config.ts or similar"

**Source:** `internal/cmd/init.go:202-241` - snippetTypeScript constant

**Significance:** Rails apps don't use Vite by default. They use Turbo/Hotwire with Rails asset pipeline or Importmaps. Both the browser and server snippets need Rails-specific versions.

---

### Finding 2: Stack detection is modular and extensible

**Evidence:**
- `detect.Stack` is a simple string type with constants
- Detection uses marker file list with priority order
- Adding a new stack requires: add constant, add marker to priority list

**Source:** `internal/detect/stack.go:8-40`

**Significance:** Adding Rails detection is straightforward - add `Rails` constant and check for `Gemfile` or `config/routes.rb`.

---

### Finding 3: Snippet selection is switch-based, easily extensible

**Evidence:**
- `getSnippet()` function uses simple switch statement on stack name
- Each stack has a separate constant for its snippet
- JSON output includes both `stack` and `snippet_language` fields

**Source:** `internal/cmd/init.go:187-200`

**Significance:** Adding Rails snippet follows the same pattern - add `snippetRails` constant, add case to switch.

---

## Synthesis

**Key Insights:**

1. **Rails needs both browser and server snippets** - Browser code captures errors and POSTs to `/__agentlog`, server code (Rails controller) receives the POST and writes to `.agentlog/errors.jsonl`.

2. **Rails detection should check for Gemfile with rails dependency** - More reliable than just Gemfile presence (could be non-Rails Ruby project). Alternative: check for `config/routes.rb` which is Rails-specific.

3. **Browser snippet should be framework-agnostic where possible** - The core error capture logic is the same, only the dev-mode check differs (Vite: `import.meta.env.DEV`, Rails: check for specific condition or always-on in dev).

**Answer to Investigation Question:**

Add Rails as a new stack type with:
1. Detection via `Gemfile` marker file (priority after `package.json`, before `go.mod`)
2. Separate `snippetRails` constant with:
   - Browser: Similar error capture but without Vite-specific checks
   - Server: Rails controller + route instead of Vite plugin

---

## Confidence Assessment

**Current Confidence:** High (90%)

**Why this level?**

The approach follows existing patterns exactly. Rails/Turbo integration is well-documented. No architectural changes needed.

**What's certain:**

- ✅ Current snippet is Vite-specific (explicit in code)
- ✅ Stack detection is extensible (proven by existing 4 stacks)
- ✅ Rails uses different entry points and no Vite plugin

**What's uncertain:**

- ⚠️ Best marker file for Rails detection (Gemfile vs config/routes.rb)
- ⚠️ Whether to check Gemfile contents or just presence
- ⚠️ Exact Rails development mode check on browser side

**What would increase confidence to Very High:**

- Validate on actual Rails/Turbo project
- Confirm snippet works with both Importmaps and jsbundling-rails

---

## Implementation Recommendations

**Purpose:** Bridge from investigation findings to actionable implementation.

### Recommended Approach ⭐

**Add Rails as new Stack type** - Follow existing patterns to add Rails stack detection and snippet.

**Why this approach:**
- Maintains consistency with existing TypeScript/Go/Python/Rust pattern
- Clear separation between build tools (Rails isn't really "TypeScript")
- Allows Rails-specific snippet customization

**Trade-offs accepted:**
- Rails is bundled framework, not language like TypeScript - acceptable since goal is integration-specific snippets
- Gemfile detection could false-positive on non-Rails Ruby - mitigated by priority order (package.json first)

**Implementation sequence:**
1. Add `Rails` constant and detection in `internal/detect/stack.go`
2. Add `snippetRails` constant in `internal/cmd/init.go`
3. Add tests for Rails detection and snippet output

### Alternative Approaches Considered

**Option B: TypeScript sub-variants (--stack typescript-vite / typescript-rails)**
- **Pros:** Keeps TypeScript as single language, explicit variant selection
- **Cons:** Complicates detection logic, awkward naming
- **When to use instead:** If many more variants emerge (Next.js, Remix, etc.)

**Option C: Separate browser/server snippets with mix-and-match**
- **Pros:** Maximum flexibility, users pick browser + server separately
- **Cons:** More complex, requires multiple flags or prompts
- **When to use instead:** When users request more granular control

**Rationale for recommendation:** Option A is simplest, follows existing patterns, and solves the immediate need. Can evolve to B/C later if needed.

---

### Implementation Details

**What to implement first:**
1. Rails detection (Gemfile marker)
2. Rails snippet (browser + controller + route)
3. Tests for both detection and init output

**Things to watch out for:**
- ⚠️ Gemfile priority relative to package.json (Rails often has package.json too)
- ⚠️ Browser snippet needs to work with Turbo Drive page navigation
- ⚠️ Rails route needs to only exist in development environment

**Areas needing further investigation:**
- Importmaps vs jsbundling-rails entry point differences
- Turbo Streams error handling (if different from standard errors)

**Success criteria:**
- ✅ `agentlog init` detects Rails projects and outputs Rails snippet
- ✅ Snippet compiles/runs in Rails 7 project with Turbo
- ✅ All existing tests continue to pass
- ✅ New tests cover Rails detection and snippet content

---

## References

**Files Examined:**
- `internal/detect/stack.go` - Stack detection logic
- `internal/detect/stack_test.go` - Detection test patterns
- `internal/cmd/init.go` - Init command and snippet constants
- `docs/jsonl-schema.md` - JSONL format specification

**Commands Run:**
```bash
# Explored codebase structure
find . -name "*.go" | head -20
ls -la internal/
```

**Related Artifacts:**
- **Investigation:** `.kb/investigations/2025-12-10-design-agentlog-architecture.md` - Original design

---

## Investigation History

**2025-12-10 21:35:** Investigation started
- Initial question: How to add Rails/Turbo support?
- Context: Current TypeScript snippet assumes Vite, Rails apps need different integration

**2025-12-10 21:40:** Codebase analysis complete
- Found stack detection in internal/detect/stack.go
- Found snippet constants in internal/cmd/init.go
- Confirmed extension pattern is straightforward

**2025-12-10 21:45:** Investigation completed
- Final confidence: High (90%)
- Status: Complete
- Key outcome: Add Rails as new stack type with Gemfile detection and Rails-specific snippet
