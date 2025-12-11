**TLDR:** Question: Why don't Node.js snippets auto-update .gitignore when creating .agentlog directory? Answer: The snippetNode and nodeCapture constants create the .agentlog directory but don't update .gitignore, meaning users who copy-paste the snippets (instead of using `agentlog init`) will accidentally commit their errors.jsonl file. High confidence (90%) - code review confirms this gap.

---

# Investigation: Node.js Snippet Auto-Update .gitignore

**Question:** Should the Node.js code snippets (snippetNode and nodeCapture) auto-update .gitignore when creating the .agentlog directory?

**Started:** 2025-12-11
**Updated:** 2025-12-11
**Owner:** Dylan Conlin
**Phase:** Complete
**Next Step:** None
**Status:** Complete
**Confidence:** Very High (95%)

---

## Findings

### Finding 1: Node.js snippets create .agentlog directory but don't update .gitignore

**Evidence:** Both `snippetNode` (lines 567-569) and `nodeCapture` (lines 981-983) in internal/cmd/init.go contain code to create the .agentlog directory:
```typescript
if (!existsSync('.agentlog')) {
  mkdirSync('.agentlog', { recursive: true });
}
```
However, neither snippet includes code to update the user's .gitignore file.

**Source:** internal/cmd/init.go:567-569, internal/cmd/init.go:981-983

**Significance:** Users who copy-paste these snippets (instead of running `agentlog init`) will have .agentlog/errors.jsonl tracked by git, which is a privacy/security issue as error logs may contain sensitive data.

---

### Finding 2: Only `agentlog init` command updates .gitignore

**Evidence:** The `runInit` function (lines 127-154) is the only place that updates .gitignore:
```go
gitignoreEntry := ".agentlog/errors.jsonl"
if !strings.Contains(string(gitignoreContent), gitignoreEntry) {
  // ... update logic ...
}
```
The install functions (installTypeScriptSnippets, installNodeSnippets, etc.) create .agentlog but don't update .gitignore because `runInit` handles it before calling them.

**Source:** internal/cmd/init.go:127-154

**Significance:** This confirms the gap - the CLI handles .gitignore, but the copy-paste snippets don't.

---

### Finding 3: Two usage patterns exist: CLI init vs copy-paste snippets

**Evidence:**
- Pattern 1: `agentlog init` (with or without `--install`) - .gitignore IS updated by `runInit`
- Pattern 2: Copy-paste snippet - .gitignore IS NOT updated because the snippet code doesn't include that logic

**Source:** internal/cmd/init.go workflow analysis

**Significance:** The bug only affects Pattern 2 (copy-paste users), which is a significant UX gap since the README and docs promote copy-paste snippets as a quick-start option.

---

## Synthesis

**Key Insights:**

1. **Two usage patterns, inconsistent behavior** - Users can either run `agentlog init` (which updates .gitignore) or copy-paste snippets (which didn't), creating inconsistent privacy/security posture.

2. **Node.js snippets have filesystem access** - Unlike browser TypeScript snippets, Node.js snippets use the `fs` module and can safely update .gitignore when creating the .agentlog directory.

3. **Fix is straightforward** - Add .gitignore update logic after directory creation, wrapped in the same try-catch to fail silently if .gitignore is readonly or missing.

**Answer to Investigation Question:**

Yes, Node.js snippets should auto-update .gitignore when creating the .agentlog directory. The fix adds readFileSync/writeFileSync logic to both `snippetNode` and `nodeCapture` constants, checking if .agentlog/errors.jsonl is already present before appending it. This ensures users who copy-paste the snippets get the same privacy protection as users who run `agentlog init`.

---

## Confidence Assessment

**Current Confidence:** Very High (95%)

**Why this level?**

The fix is tested, all tests pass, and the implementation matches the existing .gitignore update pattern from `runInit`.

**What's certain:**

- ✅ Node.js snippets create .agentlog directory but didn't update .gitignore (confirmed by code review)
- ✅ Only `agentlog init` updated .gitignore before this fix (confirmed by flow analysis)
- ✅ Tests verify the snippets now contain .gitignore update logic (2 new tests pass)
- ✅ No regressions - all 79 existing tests still pass

**What's uncertain:**

- ⚠️ Edge case: What if .gitignore is readonly? (Answer: try-catch makes it fail silently, which is acceptable)
- ⚠️ Performance impact of reading .gitignore on every first error log (minimal - only happens once when .agentlog is first created)

**What would increase confidence to Very High:**

- Integration test in a real Node.js project (manual validation recommended but not blocking)

---

## Implementation Recommendations

**Purpose:** The fix has been implemented and tested.

### Implemented Solution ⭐

**Node.js Snippet .gitignore Auto-Update** - Added .gitignore update logic to both `snippetNode` and `nodeCapture` constants.

**Changes made:**
1. Added `readFileSync, writeFileSync` to imports in both snippets
2. Added .gitignore update logic after `mkdirSync('.agentlog')`
3. Logic checks if .agentlog/errors.jsonl is already present before appending
4. Wrapped in existing try-catch to fail silently on errors

**Files modified:**
- `internal/cmd/init.go` (lines 528, 570-584, 958, 1000-1014)
- `internal/cmd/init_test.go` (added 2 new tests)

**Test results:**
- 2 new tests pass (TestNodeSnippet_UpdatesGitignore, TestNodeCapture_UpdatesGitignore)
- All 79 existing tests still pass
- No regressions detected

---

### Finding 2: [Brief, descriptive title]

**Evidence:** [Concrete observations, data, examples]

**Source:** [File paths with line numbers, commands run, specific artifacts examined]

**Significance:** [Why this matters, what it tells us, implications for the investigation question]

---

### Finding 3: [Brief, descriptive title]

**Evidence:** [Concrete observations, data, examples]

**Source:** [File paths with line numbers, commands run, specific artifacts examined]

**Significance:** [Why this matters, what it tells us, implications for the investigation question]

---

## Synthesis

**Key Insights:**

1. **[Insight title]** - [Explanation of the insight, connecting multiple findings]

2. **[Insight title]** - [Explanation of the insight, connecting multiple findings]

3. **[Insight title]** - [Explanation of the insight, connecting multiple findings]

**Answer to Investigation Question:**

[Clear, direct answer to the question posed at the top of this investigation. Reference specific findings that support this answer. Acknowledge any limitations or gaps.]

---

## Confidence Assessment

**Current Confidence:** [Level] ([Percentage])

**Why this level?**

[Explanation of why you chose this confidence level - what evidence supports it, what's strong vs uncertain]

**What's certain:**

- ✅ [Thing you're confident about with supporting evidence]
- ✅ [Thing you're confident about with supporting evidence]
- ✅ [Thing you're confident about with supporting evidence]

**What's uncertain:**

- ⚠️ [Area of uncertainty or limitation]
- ⚠️ [Area of uncertainty or limitation]
- ⚠️ [Area of uncertainty or limitation]

**What would increase confidence to [next level]:**

- [Specific additional investigation or evidence needed]
- [Specific additional investigation or evidence needed]
- [Specific additional investigation or evidence needed]

**Confidence levels guide:**
- **Very High (95%+):** Strong evidence, minimal uncertainty, unlikely to change
- **High (80-94%):** Solid evidence, minor uncertainties, confident to act
- **Medium (60-79%):** Reasonable evidence, notable gaps, validate before major commitment
- **Low (40-59%):** Limited evidence, high uncertainty, proceed with caution
- **Very Low (<40%):** Highly speculative, more investigation needed

---

## Implementation Recommendations

**Purpose:** Bridge from investigation findings to actionable implementation using directive guidance pattern (strong recommendations + visible reasoning).

### Recommended Approach ⭐

**[Approach Name]** - [One sentence stating the recommended implementation]

**Why this approach:**
- [Key benefit 1 based on findings]
- [Key benefit 2 based on findings]
- [How this directly addresses investigation findings]

**Trade-offs accepted:**
- [What we're giving up or deferring]
- [Why that's acceptable given findings]

**Implementation sequence:**
1. [First step - why it's foundational]
2. [Second step - why it comes next]
3. [Third step - builds on previous]

### Alternative Approaches Considered

**Option B: [Alternative approach]**
- **Pros:** [Benefits]
- **Cons:** [Why not recommended - reference findings]
- **When to use instead:** [Conditions where this might be better]

**Option C: [Alternative approach]**
- **Pros:** [Benefits]
- **Cons:** [Why not recommended - reference findings]
- **When to use instead:** [Conditions where this might be better]

**Rationale for recommendation:** [Brief synthesis of why Option A beats alternatives given investigation findings]

---

### Implementation Details

**What to implement first:**
- [Highest priority change based on findings]
- [Quick wins or foundational work]
- [Dependencies that need to be addressed early]

**Things to watch out for:**
- ⚠️ [Edge cases or gotchas discovered during investigation]
- ⚠️ [Areas of uncertainty that need validation during implementation]
- ⚠️ [Performance, security, or compatibility concerns to address]

**Areas needing further investigation:**
- [Questions that arose but weren't in scope]
- [Uncertainty areas that might affect implementation]
- [Optional deep-dives that could improve the solution]

**Success criteria:**
- ✅ [How to know the implementation solved the investigated problem]
- ✅ [What to test or validate]
- ✅ [Metrics or observability to add]

---

## References

**Files Examined:**
- [File path] - [What you looked at and why]
- [File path] - [What you looked at and why]

**Commands Run:**
```bash
# [Command description]
[command]

# [Command description]
[command]
```

**External Documentation:**
- [Link or reference] - [What it is and relevance]

**Related Artifacts:**
- **Decision:** [Path to related decision document] - [How it relates]
- **Investigation:** [Path to related investigation] - [How it relates]
- **Workspace:** [Path to related workspace] - [How it relates]

---

## Investigation History

**[YYYY-MM-DD HH:MM]:** Investigation started
- Initial question: [Original question as posed]
- Context: [Why this investigation was initiated]

**[YYYY-MM-DD HH:MM]:** [Milestone or significant finding]
- [Description of what happened or was discovered]

**[YYYY-MM-DD HH:MM]:** Investigation completed
- Final confidence: [Level] ([Percentage])
- Status: [Complete/Paused with reason]
- Key outcome: [One sentence summary of result]
