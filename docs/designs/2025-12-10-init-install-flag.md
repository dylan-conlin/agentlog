# Design: agentlog init --install Flag

**Status:** Draft
**Created:** 2025-12-10
**Author:** worker-agent

---

## Problem Statement

Currently, `agentlog init` prints code snippets to stdout that users must manually copy/paste into their project files. This is error-prone and tedious, especially for Rails where the snippet involves 4 different files.

**Success Criteria:**
- `agentlog init --install` writes files directly to the project
- Reports what was done, not what user should do
- Idempotent (running twice doesn't duplicate code)
- Safe (doesn't overwrite user code without merging)

---

## Approach

### Stack-Aware Installation

Different stacks have different installation strategies based on their conventions:

| Stack | Strategy | Files Created/Modified |
|-------|----------|----------------------|
| Ruby (Rails) | Full automation | 4 files (controller, initializer, route, JS) |
| TypeScript | Importable files | `.agentlog/capture.ts` |
| Go | Importable files | `.agentlog/capture.go` |
| Python | Importable files | `.agentlog/capture.py` |
| Rust | Importable files | `.agentlog/capture.rs` |

**Rationale:** Rails has strong conventions (known file paths). Other stacks have variable entry points, so we create files users import.

---

## Data Model

### InstallAction

Represents a single file operation:

```go
type InstallAction struct {
    Path      string      // Target file path (relative)
    Operation string      // "create", "append", "insert"
    Content   string      // Content to write/append
    Marker    string      // For idempotency check
    Before    string      // For "insert" - insert before this pattern
}
```

### InitResult Changes

Add installation results to existing struct:

```go
type InitResult struct {
    // ... existing fields ...
    Installed      bool             `json:"installed"`
    InstallActions []InstallAction  `json:"install_actions,omitempty"`
}
```

---

## Architecture

### File Operations

**Create:** Write new file (fail if exists and no `--force`)
```go
func createFile(path, content string) error
```

**Append:** Add content to end of file (create if doesn't exist)
```go
func appendToFile(path, content, marker string) error
```

**Insert:** Insert content before a pattern (for routes.rb)
```go
func insertBefore(path, content, pattern, marker string) error
```

### Idempotency

Each operation includes a `marker` string that's checked before writing:
- For create: Check if file exists with marker comment
- For append: Check if file already contains marker
- For insert: Same as append

Marker format: `# agentlog:installed` or `// agentlog:installed`

---

## Installation Specs by Stack

### Ruby (Rails)

**1. Controller** - `app/controllers/agentlog_controller.rb`
- Operation: create
- Content: Full controller class

**2. Initializer** - `config/initializers/agentlog.rb`
- Operation: create
- Content: Middleware setup

**3. Route** - `config/routes.rb`
- Operation: insert
- Before: `end` (last one - close of `Rails.application.routes.draw`)
- Content: `post '/__agentlog', to: 'agentlog#create' if Rails.env.development?`

**4. Frontend JS** - `app/javascript/application.js`
- Operation: append
- Content: Browser error capture script

### TypeScript

**1. Capture file** - `.agentlog/capture.ts`
- Operation: create
- Content: Browser-side error capture (window.onerror, etc.)

**Output message:** "Created .agentlog/capture.ts - import in your app entry point"

### Go

**1. Capture file** - `.agentlog/capture.go`
- Operation: create
- Content: Go error handler package

**Output message:** "Created .agentlog/capture.go - import and call InitAgentlog() in main()"

### Python

**1. Capture file** - `.agentlog/capture.py`
- Operation: create
- Content: Python exception hook

**Output message:** "Created .agentlog/capture.py - import and call init_agentlog() at startup"

### Rust

**1. Capture file** - `.agentlog/capture.rs`
- Operation: create
- Content: Rust panic hook

**Output message:** "Created .agentlog/capture.rs - add as module and call init_agentlog() in main()"

---

## Testing Strategy

### Unit Tests

1. **Rails installation tests:**
   - Creates all 4 files in correct locations
   - Inserts route correctly (before `end`)
   - Idempotent (no duplicates on second run)
   - Handles missing directories (creates them)

2. **Other stack tests:**
   - Creates `.agentlog/capture.<ext>` file
   - File contains correct content
   - Idempotent

3. **Edge cases:**
   - Target file doesn't exist (create vs error)
   - Target file exists with marker (skip)
   - No write permission (error message)

### Integration Tests

- Run in temp directory with package.json → TypeScript behavior
- Run in temp directory with config/routes.rb → Rails behavior

---

## Security Considerations

- Only writes to known paths (no user input in paths)
- Marker check prevents injection via crafted existing files
- No external network calls during installation

---

## Output Changes

### Without --install (unchanged)
```
Detected stack: Ruby (from config/routes.rb)

Created .agentlog/ directory
Added .agentlog/errors.jsonl to .gitignore

Add this snippet to your Ruby code:
---
[snippet]
---

Done! Run 'agentlog tail' to watch for errors.
```

### With --install (new)
```
Detected stack: Ruby (from config/routes.rb)

Created .agentlog/ directory
Added .agentlog/errors.jsonl to .gitignore

Installed agentlog to your project:
  Created: app/controllers/agentlog_controller.rb
  Created: config/initializers/agentlog.rb
  Modified: config/routes.rb (added agentlog route)
  Modified: app/javascript/application.js (added error capture)

Done! Run 'agentlog tail' to watch for errors.
```

### JSON Output with --install
```json
{
  "stack": "ruby",
  "detected": true,
  "marker_file": "config/routes.rb",
  "dir_created": true,
  "gitignore_updated": true,
  "installed": true,
  "install_actions": [
    {"path": "app/controllers/agentlog_controller.rb", "operation": "create"},
    {"path": "config/initializers/agentlog.rb", "operation": "create"},
    {"path": "config/routes.rb", "operation": "insert"},
    {"path": "app/javascript/application.js", "operation": "append"}
  ]
}
```

---

## Rollout Plan

1. Add `--install` flag (opt-in)
2. Test with Rails projects
3. Test with TypeScript projects
4. Consider making `--install` default in v2

---

## Open Questions

1. Should `--install` be the default in future versions?
2. Should we add `--dry-run` to preview changes?
3. Should we support `--install --force` to overwrite existing files?
