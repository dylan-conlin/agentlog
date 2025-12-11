# agentlog

**AI-native development observability** - Error visibility for agents in any stack.

> Sentry is for production. DevTools is for humans. agentlog is for AI agents.

agentlog gives AI coding agents structured access to errors during development - no browser MCP needed, no cloud accounts, works with any language.

## Quick Start

### 1. Install

```bash
# From source (requires Go 1.21+)
go install github.com/agentlog/agentlog/cmd/agentlog@latest

# Or clone and build
git clone https://github.com/agentlog/agentlog.git
cd agentlog
go build -o agentlog ./cmd/agentlog
```

### 2. Initialize in your project

```bash
agentlog init
```

This will:
- Auto-detect your stack (TypeScript, Go, Python, Rust)
- Create `.agentlog/` directory
- Add `.agentlog/errors.jsonl` to `.gitignore`
- Print a code snippet for your language

### 3. Add the snippet to your code

`agentlog init` outputs a snippet for your detected stack. Copy it into your application entry point.

**Example for TypeScript (browser):**
```typescript
// Add to your app's entry point
if (typeof window !== 'undefined' && import.meta.env?.DEV !== false) {
  window.onerror = (msg, src, line, col, err) => {
    fetch('/__agentlog', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({
        timestamp: new Date().toISOString(),
        source: 'frontend',
        error_type: 'UNCAUGHT_ERROR',
        message: String(msg).slice(0, 500),
      }),
    }).catch(() => {});
  };
}
```

Run `agentlog init --stack go|python|rust` for other languages.

### 4. View errors

```bash
# Show recent errors
agentlog errors

# Watch for new errors in real-time
agentlog tail

# Filter errors
agentlog errors --source frontend
agentlog errors --type DATABASE_ERROR
agentlog errors --since 1h
```

## Why agentlog?

**For developers:**
- Zero-config setup
- Works with any language
- Local-first (no cloud, no account)
- 15 lines of code to integrate

**For AI agents:**
- Structured JSON output (`--json`)
- Machine-readable help (`--ai-help`)
- Context injection (`agentlog prime`)
- Errors visible in terminal, no browser tools needed

## Commands

| Command | Description |
|---------|-------------|
| `agentlog init` | Initialize agentlog, detect stack, print snippet |
| `agentlog errors` | Query errors from `.agentlog/errors.jsonl` |
| `agentlog tail` | Watch for errors in real-time |
| `agentlog doctor` | Check configuration health |
| `agentlog prime` | Output context summary for AI agents |
| `agentlog --ai-help` | Machine-readable command metadata |

### Flags

```bash
--json       # Output in JSON format (for scripts and agents)
--ai-help    # Machine-readable command metadata
```

## How It Works

```
Your App ──▶ .agentlog/errors.jsonl ──▶ agentlog CLI ──▶ Agent reads terminal
              (JSONL format)              (query/tail)
```

1. Your app writes errors to `.agentlog/errors.jsonl` (via snippet or SDK)
2. agentlog CLI reads and queries this file
3. AI agent sees errors in the terminal output

**No browser MCP required.** Errors flow from app → file → terminal.

## Supported Stacks

Snippets are provided for:

- **TypeScript** - Browser + Vite/Node dev server
- **Go** - Panic handler
- **Python** - Exception hook
- **Rust** - Panic hook

Any language that can write JSON to a file works with agentlog.

## JSONL Schema

Errors are stored as JSON Lines (one JSON object per line):

```json
{"timestamp":"2025-12-10T19:19:32.941Z","source":"frontend","error_type":"UNCAUGHT_ERROR","message":"Cannot read property 'foo' of undefined"}
```

**Required fields:**
- `timestamp` - ISO 8601 UTC
- `source` - `frontend`, `backend`, `cli`, `worker`, `test`
- `error_type` - Error classification (see docs)
- `message` - Human-readable description (max 500 chars)

See [docs/jsonl-schema.md](docs/jsonl-schema.md) for full specification.

## Development

```bash
# Build
go build -o agentlog ./cmd/agentlog

# Test
go test ./...

# Build with Makefile
make build
make test
```

## License

MIT
