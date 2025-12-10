# agentlog - Worker Context

---

@.orch/CLAUDE.md

---

## Project Foundation

**Purpose:** AI-native development observability CLI - error visibility for agents in any stack

**Tagline:** "Sentry is for production. DevTools is for humans. agentlog is for AI agents."

**Created:** 2025-12-10

---

## Design Principles

**This is an AI-first CLI.** Follow the rules in:
- `~/orch-knowledge/docs/ai-first-cli-rules.md` (canonical reference)

**Key principles for agentlog:**

1. **Dual output modes** - Human-readable default, `--json` for scripts
2. **TTY detection** - Skip confirmations when non-interactive
3. **Actionable errors** - Tell agent what went wrong AND what to do
4. **Self-describing** - `agentlog --ai-help` outputs machine-readable metadata
5. **Context injection** - `agentlog prime` outputs summary for agent context

---

## Architecture

```
.agentlog/errors.jsonl  ←  Apps write JSONL (15-line snippets)
         ↓
agentlog CLI (Go)       ←  errors, tail, prime, doctor
         ↓
Agent reads terminal    ←  No browser MCP needed
```

**Core decisions (from design investigation):**
- **Language:** Go (single binary, no runtime deps)
- **Primary pattern:** Snippets + file convention (not SDK-first)
- **Scope:** Errors only (v1) - focused, not feature-creeping
- **Storage:** `.agentlog/errors.jsonl` (JSONL format)

---

## Implementation Details

**CLI Commands (MVP):**
```
agentlog init       # Detect stack, create config, print snippet
agentlog errors     # Query .agentlog/errors.jsonl
agentlog tail       # Live watch
agentlog doctor     # Health check
agentlog prime      # Output context for agents
agentlog --ai-help  # Machine-readable command metadata
```

**Snippet languages (priority order):**
1. TypeScript (webapps - primary use case)
2. Go
3. Python
4. Rust

---

## JSONL Schema

**Required fields:**
```json
{
  "timestamp": "2025-12-10T19:19:32.941Z",
  "source": "frontend",
  "error_type": "UNCAUGHT_ERROR",
  "message": "Cannot read property 'foo'"
}
```

**Size limits:**
- Max message: 500 chars
- Max stack_trace: 2KB
- Max entry: 10KB
- Max file: 10MB (then rotate)

---

## Key Files

- `.kb/investigations/2025-12-10-design-agentlog-architecture.md` - Full design doc
- `cmd/` - CLI commands (cobra)
- `internal/` - Core logic
- `snippets/` - Language-specific integration code

---

## Development Setup

```bash
# Prerequisites
go 1.21+

# Build
go build -o agentlog ./cmd/agentlog

# Test
go test ./...

# Run
./agentlog --help
```

---

## Notes

- This tool should dogfood itself - use agentlog for agentlog development
- Dev-only by design - snippets should no-op in production
- Gitignore `.agentlog/errors.jsonl` by default (privacy)
