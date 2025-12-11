# Design: Rails/Turbo TypeScript Snippet

**Status:** Approved (Implemented)
**Author:** Worker Agent
**Created:** 2025-12-10

---

## Problem Statement

Current TypeScript snippet assumes Vite:
- Browser uses `import.meta.env?.DEV` (Vite-specific)
- Server uses Vite plugin with `configureServer()` hook

Rails apps with Turbo/Hotwire need different integration:
- No Vite by default (uses Importmaps or jsbundling-rails)
- Entry point is `app/javascript/application.js`
- Server needs Rails controller + route, not Vite plugin

**Success Criteria:**
- [ ] Rails projects detected via marker file
- [ ] Rails-specific snippet output by `agentlog init`
- [ ] Snippet works with standard Rails 7 + Turbo setup
- [ ] All existing tests pass

---

## Approach

Add Rails as a new stack type following existing patterns:

### 1. Stack Detection Changes

**File:** `internal/detect/stack.go`

Add `Rails` constant and detection marker:

```go
const (
    TypeScript Stack = "typescript"
    Rails      Stack = "rails"      // NEW
    Go         Stack = "go"
    Python     Stack = "python"
    Rust       Stack = "rust"
)

var markerPriority = []struct {
    file  string
    stack Stack
}{
    {"config/routes.rb", Rails},    // NEW - Rails-specific, checked first
    {"package.json", TypeScript},
    {"go.mod", Go},
    {"pyproject.toml", Python},
    {"requirements.txt", Python},
    {"Cargo.toml", Rust},
}
```

**Design Decision:** Use `config/routes.rb` instead of `Gemfile` because:
- `config/routes.rb` is Rails-specific (unambiguous)
- Rails apps often have `package.json` for npm dependencies
- Checking `config/routes.rb` first ensures Rails detection takes priority

### 2. Rails Snippet

**File:** `internal/cmd/init.go`

Add `snippetRails` constant with browser + server code:

```javascript
// === BROWSER (add to app/javascript/application.js) ===
// Error capture for agentlog - sends errors to /__agentlog endpoint
(function() {
  const log = (type, msg, ctx) =>
    fetch('/__agentlog', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({
        timestamp: new Date().toISOString(),
        source: 'frontend',
        error_type: type,
        message: String(msg).slice(0, 500),
        context: ctx,
      }),
    }).catch(() => {});

  window.onerror = (msg, src, line, col, err) =>
    log('UNCAUGHT_ERROR', msg, { file: src, line, column: col, stack_trace: err?.stack?.slice(0, 2048) });

  window.onunhandledrejection = (e) =>
    log('UNHANDLED_REJECTION', e.reason, { stack_trace: e.reason?.stack?.slice(0, 2048) });
})();
```

```ruby
# === RAILS CONTROLLER (app/controllers/agentlog_controller.rb) ===
class AgentlogController < ApplicationController
  skip_before_action :verify_authenticity_token, only: :create

  def create
    return head :not_found unless Rails.env.development?

    FileUtils.mkdir_p('.agentlog')
    File.open('.agentlog/errors.jsonl', 'a') do |f|
      f.puts(request.raw_post)
    end

    head :ok
  end
end

# === ROUTE (add to config/routes.rb) ===
post '/__agentlog', to: 'agentlog#create' if Rails.env.development?
```

**Design Decisions:**

1. **IIFE wrapper for browser code** - No module system assumption (works with Importmaps and jsbundling)
2. **No dev-mode check in browser** - The Rails route only exists in development, so browser always sends
3. **skip_before_action for CSRF** - JSON POST from JavaScript won't have CSRF token
4. **Rails.env.development? on server** - Standard Rails idiom for dev-only features
5. **Conditional route** - Route only added in development environment

### 3. Init Command Changes

**File:** `internal/cmd/init.go`

Add case to `getSnippet()`:

```go
func getSnippet(stack string) string {
    switch stack {
    case "typescript":
        return snippetTypeScript
    case "rails":           // NEW
        return snippetRails // NEW
    case "go":
        return snippetGo
    // ...
    }
}
```

---

## Data Model

No database changes. Uses existing `.agentlog/errors.jsonl` file.

---

## Testing Strategy

### Unit Tests

1. **Stack Detection Tests** (`internal/detect/stack_test.go`):
   - `config/routes.rb` detected as Rails
   - `config/routes.rb` takes priority over `package.json`
   - Rails constant has correct string value

2. **Snippet Tests** (`internal/cmd/init_test.go`):
   - Rails snippet returned for rails stack
   - Snippet contains required components (browser code, controller, route)

### Integration Tests

- Run `agentlog init` in directory with `config/routes.rb`
- Verify Rails detection and snippet output

---

## Security Considerations

1. **CSRF bypass is intentional** - The agentlog endpoint only writes to local file in development, no security risk
2. **No user data** - Error logs may contain stack traces but no sensitive user data by design
3. **Dev-only route** - Conditional route ensures endpoint doesn't exist in production

---

## Performance Requirements

- No performance impact on production (route doesn't exist)
- File append is synchronous but acceptable for dev-mode error logging
- No changes to agentlog CLI performance

---

## Rollout Plan

1. Implement with TDD
2. Test in actual Rails 7 + Turbo project (manual validation)
3. Update README with Rails example

---

## Alternatives Considered

### Option B: TypeScript sub-variants
`--stack typescript-vite` / `--stack typescript-rails`

**Rejected because:** Complicates detection, Rails isn't really a TypeScript variant

### Option C: Separate browser/server snippet selection
**Rejected because:** Overengineered for current needs, can add later if demand arises

---

## Open Questions

1. ~~Which marker file for Rails detection?~~ **Resolved:** `config/routes.rb` (unambiguous)
2. Should we support Rails with Vite (vite_rails gem)? **Deferred:** Users can use `--stack typescript` if using Vite with Rails

---

## References

- Investigation: `.kb/investigations/2025-12-10-inv-add-rails-turbo-typescript-snippet.md`
- Architecture: `.kb/investigations/2025-12-10-design-agentlog-architecture.md`
- JSONL Schema: `docs/jsonl-schema.md`
