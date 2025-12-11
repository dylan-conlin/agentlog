**TLDR:** Question: What is the formal JSONL schema for agentlog errors? Answer: Created comprehensive spec at docs/jsonl-schema.md defining required fields (timestamp, source, error_type, message), error type taxonomy, optional context fields, and size limits. High confidence (95%) - based on architecture design decisions already made.

---

# Investigation: Write JSONL Schema Spec

**Question:** Document the errors.jsonl format: required fields, optional context fields, and size limits as the universal contract.

**Started:** 2025-12-10
**Updated:** 2025-12-10
**Owner:** Worker agent (spawned by orchestrator)
**Phase:** Complete
**Next Step:** None
**Status:** Complete
**Confidence:** Very High (95%)

---

## Findings

### Finding 1: Schema requirements already defined in architecture doc

**Evidence:** The architecture investigation at `.kb/investigations/2025-12-10-design-agentlog-architecture.md` already defined:
- Required fields: timestamp, source, error_type, message
- Error type taxonomy with categories for universal, frontend, backend, CLI, and runtime
- Size limits: 500 char message, 2KB stack trace, 10KB entry, 10MB file

**Source:** `.kb/investigations/2025-12-10-design-agentlog-architecture.md:250-296`

**Significance:** The design work was already done. This task was documentation/formalization of existing decisions.

---

### Finding 2: Spec document structure needs to be comprehensive

**Evidence:** Created a formal specification that includes:
- Required field definitions with types, formats, and examples
- Complete error type taxonomy table
- Optional context fields with max sizes
- Validation rules (required vs recommended)
- Complete examples for different source types
- Compatibility notes for JSON formatting and production mode
- Versioning (1.0.0) for future evolution

**Source:** `docs/jsonl-schema.md` (created)

**Significance:** This spec serves as the universal contract for any application integrating with agentlog. It's language-agnostic and provides enough detail for snippet/SDK implementers.

---

## Synthesis

**Key Insights:**

1. **Spec formalizes existing design** - The architecture investigation already made the key decisions. This spec documents them in a format usable by implementers.

2. **Balance of strictness and flexibility** - Required fields are minimal (4 fields). Context is flexible to support many use cases without over-constraining.

3. **Size limits prevent abuse** - The limits (500 char message, 2KB stack, 10KB entry, 10MB file) are practical for development errors while preventing runaway logging.

**Answer to Investigation Question:**

The JSONL schema spec is now documented at `docs/jsonl-schema.md`. It defines:
- **Required fields:** timestamp (ISO 8601 UTC), source, error_type (from taxonomy), message
- **Error type taxonomy:** 15 types across universal, frontend, backend, CLI, and runtime categories
- **Optional context:** session_id, stack_trace, url, endpoint, command, component, etc.
- **Size limits:** 500 char message, 2KB stack trace, 10KB total entry, 10MB file with rotation

---

## Confidence Assessment

**Current Confidence:** Very High (95%)

**Why this level?**

The spec directly transcribes and formalizes decisions already made in the architecture investigation. No new design decisions were required.

**What's certain:**

- Required fields are correct (from architecture doc)
- Size limits are sensible (from architecture doc)
- Error taxonomy covers common cases

**What's uncertain:**

- Real-world usage may reveal missing error types
- Size limits may need tuning based on actual data

**What would increase confidence to 99%:**

- Validate spec against real snippet implementations
- Get feedback from users of beads-ui-svelte console bridge

---

## Deliverables

**Created:**
- `docs/jsonl-schema.md` - Comprehensive JSONL schema specification (v1.0.0)

**Specification includes:**
1. Required fields table with types and examples
2. Error type taxonomy with 15 types across 5 categories
3. Optional context fields with size limits
4. Complete examples for frontend, backend, CLI errors
5. Validation rules (required vs recommended)
6. Compatibility notes
7. Changelog for versioning

---

## References

**Files Examined:**
- `.kb/investigations/2025-12-10-design-agentlog-architecture.md` - Architecture decisions for schema

**Commands Run:**
```bash
# Create docs directory
mkdir -p docs
```

**Related Artifacts:**
- **Investigation:** `.kb/investigations/2025-12-10-design-agentlog-architecture.md` - Source of schema decisions
- **Deliverable:** `docs/jsonl-schema.md` - Output spec document

---

## Investigation History

**2025-12-10 18:40:** Investigation started
- Initial question: Document the errors.jsonl format
- Context: Part of agentlog MVP - task 4 from architecture investigation

**2025-12-10 18:42:** Spec document created
- Read architecture investigation for schema requirements
- Created comprehensive spec at docs/jsonl-schema.md

**2025-12-10 18:45:** Investigation completed
- Final confidence: Very High (95%)
- Status: Complete
- Key outcome: JSONL schema spec v1.0.0 documented at docs/jsonl-schema.md
