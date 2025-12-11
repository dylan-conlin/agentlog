**TLDR:** Task: Add --path flag for monorepo/subdir support. Approach: Add global --path flag to root.go with centralized path resolver, update all commands (errors, tail, doctor, prime) to use it. High confidence - pattern matches existing --json flag implementation.

---

# Investigation: Add --path Flag for Monorepo/Subdir Support

**Question:** How to implement --path flag so users can point agentlog at a subdirectory in a monorepo?

**Started:** 2025-12-11
**Updated:** 2025-12-11
**Owner:** Worker agent
**Phase:** Complete
**Next Step:** None
**Status:** Complete
**Confidence:** Very High (95%)

---

## Findings

### Finding 1: Global flags pattern established in root.go

**Evidence:** Existing global flags `--json` and `--ai-help` use this pattern:
```go
var (
  jsonOutput bool
  aiHelp     bool
)

func init() {
  rootCmd.PersistentFlags().BoolVar(&jsonOutput, "json", false, "...")
}

func IsJSONOutput() bool {
  return jsonOutput
}
```

**Source:** `internal/cmd/root.go:12-83`

**Significance:** Clear pattern to follow for adding `--path` flag.

---

### Finding 2: Path construction is hardcoded in multiple places

**Evidence:** Each command constructs path independently:
- `errors.go:113`: `filepath.Join(baseDir, ".agentlog", "errors.jsonl")`
- `tail.go:89`: `filepath.Join(baseDir, ".agentlog", "errors.jsonl")`
- `doctor.go:83-84`: Constructs both dir and file paths

**Source:** Multiple command files

**Significance:** Need centralized path resolver function.

---

### Finding 3: Commands use cwd as baseDir

**Evidence:** All commands start with `cwd, err := os.Getwd()` then pass to path construction.

**Source:** `errors.go:61-66`, `tail.go:40-44`, `doctor.go:59-63`

**Significance:** --path flag can override this baseDir value.

---

## Implementation Recommendations

### Recommended Approach ⭐

**Add global --path flag with centralized path functions**

**Implementation sequence:**
1. Add `pathOverride` variable and `--path` flag in root.go
2. Add `GetBaseDir()` and `GetErrorsPath()` helper functions
3. Update all commands to use helpers instead of hardcoded paths
4. Write tests for the new functionality

**Trade-offs accepted:**
- Slight complexity increase in root.go
- Need to update multiple files

**Success criteria:**
- ✅ `agentlog errors --path /other/project` reads from that path
- ✅ All existing tests pass
- ✅ New tests for --path flag pass

---

## References

**Files Examined:**
- `internal/cmd/root.go` - Global flag patterns
- `internal/cmd/errors.go` - Path usage in errors command
- `internal/cmd/tail.go` - Path usage in tail command
- `internal/cmd/doctor.go` - Path usage in doctor command
- `internal/cmd/errors_test.go` - Test patterns

---

## Investigation History

**2025-12-11 13:15:** Investigation started
- Initial question: How to add --path flag for monorepo support
- Context: Users in monorepos need to point agentlog at subdirectories

**2025-12-11 13:20:** Pre-impl exploration complete
- Found global flag pattern in root.go
- Identified 6 files needing modification
- Determined centralized path resolver approach

**2025-12-11 13:30:** Implementation complete
- Added `--path` global flag in root.go
- Added `GetPathOverride()` and `GetErrorsPath()` helper functions
- Updated all commands (errors, tail, doctor, prime) to use --path override
- All tests pass including new tests for --path functionality
- Manual verification: `agentlog errors --path /custom/path` works correctly
