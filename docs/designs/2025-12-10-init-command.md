# Design: agentlog init Command

**Status:** Draft
**Author:** Worker agent (agentlog-wad)
**Date:** 2025-12-10

---

## Problem Statement

The `agentlog init` command needs to initialize agentlog in a project by:
1. Detecting the project's tech stack (TypeScript, Go, Python, Rust)
2. Creating the `.agentlog/` directory
3. Adding `.agentlog/` to `.gitignore` (if not already present)
4. Printing an appropriate error capture snippet for the detected language

**Success Criteria:**
- Command works in projects with different stacks
- Idempotent (safe to run multiple times)
- Clear, actionable output for users and AI agents
- JSON output mode for programmatic use

---

## Approach

### Stack Detection Strategy

Detect stack by checking for marker files in order of priority:

| Marker File | Detected Stack | Snippet Language |
|-------------|---------------|------------------|
| `package.json` | TypeScript/JavaScript | TypeScript |
| `go.mod` | Go | Go |
| `pyproject.toml` | Python | Python |
| `requirements.txt` | Python | Python |
| `Cargo.toml` | Rust | Rust |
| (none found) | Unknown | TypeScript (default) |

**Priority order:** Check in the order listed. First match wins.

**Why TypeScript default:** Per CLAUDE.md, TypeScript is the primary use case for webapps.

### Directory Creation

1. Create `.agentlog/` directory if it doesn't exist
2. Create empty `.agentlog/errors.jsonl` file (touch)
3. Create `.agentlog/.gitkeep` to ensure directory is tracked

### Gitignore Management

1. Check if `.gitignore` exists
2. If exists, check if it already contains `.agentlog/`
3. If not present, append `.agentlog/` on a new line
4. If `.gitignore` doesn't exist, create it with `.agentlog/`

**Important:** Only add `.agentlog/` (the directory), not the whole path. The `errors.jsonl` file should be ignored but the directory structure preserved via `.gitkeep`.

Actually, per CLAUDE.md: "Gitignore `.agentlog/errors.jsonl` by default (privacy)" - so we should add the specific file pattern.

Corrected approach:
- Add `.agentlog/errors.jsonl` to `.gitignore` (not the whole directory)
- This keeps config files visible in git while ignoring the error log

### Snippet Output

Print a language-specific snippet that users can copy-paste into their application. Snippets should:
- Be minimal (15-20 lines max)
- Be copy-paste ready
- Include comments explaining what to do
- No-op in production (check environment)

---

## Data Model

### Init Result (for JSON output)

```go
type InitResult struct {
    Stack        string `json:"stack"`
    Detected     bool   `json:"detected"`         // true if auto-detected, false if default
    DirCreated   bool   `json:"dir_created"`
    GitIgnored   bool   `json:"gitignore_updated"`
    SnippetLang  string `json:"snippet_language"`
}
```

---

## Command Interface

```
agentlog init [flags]

Flags:
  --force    Reinitialize even if .agentlog/ already exists
  --stack    Override auto-detection (typescript, go, python, rust)

Global Flags:
  --json     Output in JSON format
```

### Human Output Example

```
Detected stack: TypeScript (from package.json)

Created .agentlog/ directory
Added .agentlog/errors.jsonl to .gitignore

Add this snippet to your frontend code:

---[TypeScript Snippet]---
// Error handler for agentlog
// Add this to your app's entry point (e.g., src/main.ts)

if (process.env.NODE_ENV !== 'production') {
  window.onerror = (msg, src, line, col, err) => {
    const entry = {
      timestamp: new Date().toISOString(),
      source: 'frontend',
      error_type: 'UNCAUGHT_ERROR',
      message: String(msg).slice(0, 500),
      context: { file: src, line, column: col }
    };
    // Write to .agentlog/errors.jsonl (requires backend endpoint or fs access)
    console.log('[agentlog]', JSON.stringify(entry));
  };
}
---

Done! Run 'agentlog tail' to watch for errors.
```

### JSON Output Example

```json
{
  "stack": "typescript",
  "detected": true,
  "dir_created": true,
  "gitignore_updated": true,
  "snippet_language": "typescript",
  "snippet": "// Error handler for agentlog..."
}
```

---

## Testing Strategy

1. **Unit tests for stack detection:**
   - Test each marker file detection
   - Test priority (package.json + go.mod should detect TypeScript)
   - Test fallback to default

2. **Unit tests for directory operations:**
   - Test directory creation
   - Test idempotency (run twice, no errors)

3. **Unit tests for gitignore:**
   - Test adding to empty file
   - Test adding to file with other entries
   - Test skipping if already present
   - Test creating new gitignore

4. **Integration test:**
   - Test full init flow in temp directory

---

## File Structure

```
internal/
  cmd/
    init.go       # Init command implementation
    init_test.go  # Tests for init command
  detect/
    stack.go      # Stack detection logic
    stack_test.go # Stack detection tests
snippets/
  typescript.txt  # TypeScript snippet
  go.txt          # Go snippet
  python.txt      # Python snippet
  rust.txt        # Rust snippet
```

---

## Security Considerations

- Only write to `.agentlog/` directory within cwd
- Validate that we're in a reasonable project directory (not root, not home)
- Don't execute any user code, just create files

---

## Alternatives Considered

### Option B: SDK-first approach
Create full SDK packages instead of snippets.
- **Pros:** Better DX, more features
- **Cons:** Much more work, delays MVP
- **Why not:** Per design investigation, snippets-first is the MVP approach

### Option C: Config file approach
Create `.agentlog/config.yaml` with detected settings.
- **Pros:** More customizable
- **Cons:** Overcomplicates MVP
- **Why not:** Can add later, start simple

---

## Open Questions

None - design is straightforward.

---

## References

- CLAUDE.md - Project context and priorities
- docs/jsonl-schema.md - JSONL format specification
- .kb/investigations/2025-12-10-design-agentlog-architecture.md - Architecture decisions
