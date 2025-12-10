**TLDR:** Question: How should language/framework choice change when AI writes the code? Answer: Optimize for artifact quality, not developer experience. CLIs → Go/Rust (single binary), frontends → modern stack (Bun/Svelte), backends → Python/Node unless performance proven necessary. High confidence (85%) - framework validated through interactive exploration.

---

# Investigation: AI-Native Technology Choice

**Question:** How does language and framework choice change when AI handles development, and what should the decision framework be?

**Started:** 2025-12-10
**Updated:** 2025-12-10
**Owner:** Dylan + Architect agent (interactive)
**Phase:** Complete
**Next Step:** None - framework externalized to orch-knowledge
**Status:** Complete
**Confidence:** High (85%)

---

## Findings

### Finding 1: The paradigm shift is real - artifact vs process

**Evidence:** When AI writes code, the human no longer experiences the development process. Languages that were "too annoying" (Rust's borrow checker, Go's verbosity) become viable because AI absorbs the friction.

**Source:** Interactive exploration of Dylan's thesis on AI-native development

**Significance:** This fundamentally changes the decision calculus. Old model: `f(team expertise, readability, ecosystem vibes)`. New model: `f(artifact quality, compiler strictness, deployment story)`.

---

### Finding 2: CLIs have the clearest win for Go/Rust

**Evidence:**
- Python CLI: requires `pip install`, Python version management, virtual environments
- Go/Rust CLI: single binary download, no dependencies, works everywhere
- Binary size: Go ~5-10MB, Rust ~1-3MB (both acceptable)
- Startup: Go/Rust <10ms vs Python 100-300ms

**Source:** Analysis of artifact characteristics across languages

**Significance:** For CLIs, the artifact IS the product. Users don't see source code, they see installation friction and runtime behavior. Clear case for Go/Rust.

---

### Finding 3: Frontends benefit from "annoying" modern tooling

**Evidence:**
- Svelte: smaller bundles, no virtual DOM, but "different mental model" for humans
- Bun: faster builds, mostly npm-compatible, but "unfamiliar" to humans
- shadcn: owned components, but "more setup" for humans
- With AI: all the "costs" disappear, benefits remain

**Source:** Analysis of frontend framework tradeoffs

**Significance:** The Bun/Svelte/shadcn stack produces better artifacts (smaller bundles, faster runtime) with complexity handled by AI.

---

### Finding 4: Backend ecosystem still matters more than raw performance

**Evidence:**
- Auth libraries: NextAuth, Django auth are battle-tested; Go/Rust equivalents less mature
- ORMs: SQLAlchemy, Prisma vs sqlx (manual), Diesel (complex)
- Deployment: Vercel, Railway have first-class Node/Python; Go works; Rust is friction
- Performance gap (30ms vs 5ms) invisible to users at typical scale

**Source:** Analysis of ecosystem richness vs performance characteristics

**Significance:** For most backends, ecosystem richness > raw performance. The threshold where performance matters is higher than intuition suggests.

---

### Finding 5: Performance thresholds are measurable

**Evidence:** Concrete heuristics for when backend performance matters:
- Requests/sec: Python fine < 1,000, consider Go/Rust > 10,000
- Hosting bill: Fine < $100/mo, reconsider > $500/mo
- Cold starts: Matter for user-facing serverless with infrequent invocations
- CPU-bound: Python 10-100x slower for compute-intensive work

**Source:** Analysis of when artifact improvements justify ecosystem tradeoffs

**Significance:** "Measure first" is the answer. Most performance intuitions are wrong. Bottleneck is usually database or external APIs, not application code.

---

## Synthesis

**Key Insights:**

1. **"AI absorbs the annoyance"** - Languages with better artifacts but higher human friction become viable. You get Rust's benefits without paying Rust's learning curve.

2. **Project type determines tradeoff** - CLIs: artifact matters. Frontends: artifact matters. Backends: ecosystem matters. Scripts: iteration matters.

3. **Ecosystem is still a constraint** - Even with AI writing code, a weaker ecosystem means worse tools, not just harder development. This limits Go/Rust backend adoption.

**Answer to Investigation Question:**

Language choice should shift based on what the artifact needs:
- **CLIs:** Go/Rust (single binary wins)
- **Frontends:** Modern stack like Bun/Svelte (better bundles)
- **Backends:** Python/Node until measurement proves performance matters
- **Scripts:** Python (iteration matters, artifact doesn't)

The framework is: "Optimize for artifact when artifact matters. Optimize for ecosystem when ecosystem matters."

---

## Confidence Assessment

**Current Confidence:** High (85%)

**Why this level?**

Framework validated through interactive exploration with Dylan. Reasoning is sound and maps to concrete heuristics. Not yet validated through multiple real projects.

**What's certain:**

- ✅ CLIs benefit from Go/Rust (artifact story is dramatically better)
- ✅ Frontend modern tooling is viable when AI handles complexity
- ✅ Backend ecosystem tradeoff is real and measurable
- ✅ "Measure first" is correct for performance decisions

**What's uncertain:**

- ⚠️ AI code quality across languages (is AI-written Rust as good as AI-written Python?)
- ⚠️ Debugging story when AI-written code breaks
- ⚠️ How much ecosystem gap AI can bridge

**What would increase confidence to Very High:**

- Apply framework to 3+ real projects
- Measure AI code quality across languages empirically
- Validate that "measure first" heuristics hold in practice

---

## Implementation Recommendations

**Purpose:** This investigation produced a general framework, not project-specific recommendations.

### Recommended Approach ⭐

**Framework externalized to orch-knowledge** - Created `~/orch-knowledge/docs/ai-native-technology-choice.md` as reusable reference.

**Why this approach:**
- Applies across all projects, not just agentlog
- Referenceable by future Claude sessions
- Living document that can evolve with experience

**Deliverable:** `~/orch-knowledge/docs/ai-native-technology-choice.md`

---

## References

**Files Examined:**
- `~/orch-knowledge/docs/ai-first-cli-rules.md` - AI-first CLI interface design principles
- `.kb/investigations/2025-12-10-design-agentlog-architecture.md` - Original agentlog design (chose Go)

**Related Artifacts:**
- **Output:** `~/orch-knowledge/docs/ai-native-technology-choice.md` - The framework document

---

## Investigation History

**2025-12-10 ~14:00:** Investigation started
- Initial question: Is Go still right for agentlog given AI-native paradigm?
- Context: Dylan's thesis that AI changes language choice calculus

**2025-12-10 ~14:15:** Scope expanded
- Dylan requested general discussion, not agentlog-specific
- Pivoted from "Go vs Rust for agentlog" to "general framework for AI-native tech choice"

**2025-12-10 ~14:30:** Framework developed
- Mapped decision criteria by project type
- Identified "when performance matters" heuristics
- Defined hybrid pattern (CLIs: Go, Frontend: modern, Backend: Python/Node)

**2025-12-10 ~14:45:** Investigation completed
- Final confidence: High (85%)
- Status: Complete
- Key outcome: Framework externalized to orch-knowledge/docs/ai-native-technology-choice.md
