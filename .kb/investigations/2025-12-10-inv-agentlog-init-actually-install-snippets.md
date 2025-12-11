**TLDR:** How should `agentlog init --install` write snippets to project files? Rails has strong conventions (known file paths), other stacks don't (variable entry points). Recommend: full automation for Rails, create importable files for others. High confidence (85%) - codebase analysis complete.

---

# Investigation: agentlog init --install Implementation

**Question:** How should `agentlog init` actually install snippets instead of just printing them?

**Started:** 2025-12-10
**Updated:** 2025-12-10
**Owner:** worker-agent
**Phase:** Complete
**Next Step:** None - implementation complete
**Status:** Complete
**Confidence:** High (90%)

---

## Findings

### Finding 1: Current Implementation Structure

**Evidence:** `internal/cmd/init.go` has:
- `runInit()` returns `InitResult` with snippet as string
- `printInitResult()` outputs snippet with "Add this snippet to your code" message
- Snippets defined as const strings in same file (lines 204-453)
- No file-writing for snippets, only for `.agentlog/` dir and `.gitignore`

**Source:** `internal/cmd/init.go:141-143` (snippet assignment), `internal/cmd/init.go:170-176` (print output)

**Significance:** Easy to extend - add `--install` flag and `installSnippets()` function. Snippets are already separated by stack.

---

### Finding 2: Ruby (Rails) Has Well-Defined Target Paths

**Evidence:** Rails snippet (lines 368-453) explicitly mentions:
- `app/javascript/application.js` - frontend JS
- `app/controllers/agentlog_controller.rb` - controller
- `config/routes.rb` - route
- `config/initializers/agentlog.rb` - middleware

Rails follows convention-over-configuration - these paths are standard across all Rails apps.

**Source:** `internal/cmd/init.go:368-453`, `internal/detect/stack.go:38` (detects via `config/routes.rb`)

**Significance:** Can fully automate installation for Rails - known target paths allow safe file creation/modification.

---

### Finding 3: Other Stacks Have Variable Entry Points

**Evidence:**
- TypeScript: Entry could be `src/index.ts`, `src/App.tsx`, `pages/_app.tsx`, `app/page.tsx`, etc.
- Dev server varies: Vite, Webpack, Next.js, etc.
- Go: Entry could be `main.go`, `cmd/*/main.go`, etc.
- Python: Entry could be `main.py`, `app.py`, `src/__main__.py`, etc.
- Rust: Entry is `main.rs` but may be in `src/` or `src/bin/`

**Source:** Industry knowledge, codebase analysis

**Significance:** Cannot safely auto-detect entry points for non-Rails stacks. Need different strategy.

---

## Synthesis

**Key Insights:**

1. **Rails is special** - Strong conventions mean we know exactly where files go. Can safely create controller, initializer, append to routes.rb and application.js.

2. **Other stacks need importable files** - Since entry points vary, create standalone files that user can import. E.g., `.agentlog/capture.ts` with instructions "import this in your entry point".

3. **Merge strategy matters** - Routes.rb and application.js need append/insert, not overwrite. Controller and initializer can be created fresh (with existence check).

**Answer to Investigation Question:**

For `--install` flag:
- **Rails:** Full automation - create controller, initializer, append to routes.rb and application.js
- **Other stacks:** Partial automation - create `.agentlog/capture.<ext>` file, print import instructions

---

## Confidence Assessment

**Current Confidence:** High (85%)

**Why this level?**

Strong evidence from codebase analysis. Rails convention is well-understood. Uncertainty only in edge cases (unusual project structures).

**What's certain:**

- ✅ Current snippet structure supports splitting into installable parts
- ✅ Rails paths are standardized (convention-over-configuration)
- ✅ Detection already works correctly for all stacks

**What's uncertain:**

- ⚠️ How to handle existing content in target files (merge strategy)
- ⚠️ How to detect if snippets are already installed
- ⚠️ How to handle non-standard Rails app structures

**What would increase confidence to Very High:**

- Test on real-world Rails apps with varying structures
- Validate TypeScript importable file approach works with various bundlers

---

## Implementation Recommendations

### Recommended Approach ⭐

**Stack-Aware Installation** - Full automation for Rails, importable files for others

**Why this approach:**
- Rails convention makes full automation safe and valuable
- Other stacks benefit from partial automation (less manual typing)
- Single `--install` flag works for all stacks, behavior varies appropriately

**Trade-offs accepted:**
- TypeScript/Go/Python/Rust still require manual import step
- This is acceptable because entry point detection is unreliable

**Implementation sequence:**
1. Add `--install` flag to init command
2. Split Ruby snippet into installable parts with target paths
3. Create `.agentlog/capture.<ext>` for other stacks
4. Update output messages to report what was done

### Alternative Approaches Considered

**Option B: Full automation for all stacks**
- **Pros:** Complete hands-off experience
- **Cons:** Unreliable entry point detection could break projects
- **When to use instead:** If we add interactive mode asking "Where is your entry point?"

**Option C: Just create files, never merge**
- **Pros:** Simpler implementation
- **Cons:** User still has to manually add route to config/routes.rb
- **When to use instead:** If merge logic proves too complex

**Rationale for recommendation:** Option A balances automation value with safety. Rails users get full automation (most complex snippet), others get significant time savings.

---

### Implementation Details

**What to implement first:**
- Add `--install` flag to cobra command
- Create `InstallationTarget` struct to represent file operations
- Implement Rails installation first (most complex, proves the pattern)

**Things to watch out for:**
- ⚠️ File permissions - check writable before attempting
- ⚠️ Existing content - don't duplicate if already installed
- ⚠️ Route insertion - need to find right location in routes.rb

**Success criteria:**
- ✅ `agentlog init --install` in Rails project creates all 4 files/modifications
- ✅ `agentlog init --install` in TypeScript project creates `.agentlog/capture.ts`
- ✅ Running twice is idempotent (no duplicates)
- ✅ Output reports what was done, not what user should do

---

## References

**Files Examined:**
- `internal/cmd/init.go` - Current init implementation
- `internal/cmd/init_test.go` - Existing tests
- `internal/detect/stack.go` - Stack detection logic

**Related Artifacts:**
- **Investigation:** `.kb/investigations/2025-12-10-design-agentlog-architecture.md` - Original design
