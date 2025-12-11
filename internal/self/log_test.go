package self

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// ErrorEntry matches the schema from internal/cmd/errors.go
type ErrorEntry struct {
	Timestamp string                 `json:"timestamp"`
	Source    string                 `json:"source"`
	ErrorType string                 `json:"error_type"`
	Message   string                 `json:"message"`
	Context   map[string]interface{} `json:"context,omitempty"`
}

func TestLogError_WritesToFile(t *testing.T) {
	// Setup: create temp directory with .agentlog
	tmpDir := t.TempDir()
	agentlogDir := filepath.Join(tmpDir, ".agentlog")
	os.MkdirAll(agentlogDir, 0755)
	errorsFile := filepath.Join(agentlogDir, "errors.jsonl")
	os.WriteFile(errorsFile, []byte{}, 0644) // touch file

	// Act: log an error
	LogError(tmpDir, "TEST_ERROR", "test error message")

	// Assert: error was written to file
	content, err := os.ReadFile(errorsFile)
	if err != nil {
		t.Fatalf("failed to read errors file: %v", err)
	}

	if len(content) == 0 {
		t.Fatal("errors file should not be empty")
	}

	// Parse the entry
	var entry ErrorEntry
	if err := json.Unmarshal(content[:len(content)-1], &entry); err != nil { // -1 to remove trailing newline
		t.Fatalf("failed to parse error entry: %v", err)
	}

	if entry.Source != "cli" {
		t.Errorf("Source = %q, want %q", entry.Source, "cli")
	}
	if entry.ErrorType != "TEST_ERROR" {
		t.Errorf("ErrorType = %q, want %q", entry.ErrorType, "TEST_ERROR")
	}
	if entry.Message != "test error message" {
		t.Errorf("Message = %q, want %q", entry.Message, "test error message")
	}
	if entry.Timestamp == "" {
		t.Error("Timestamp should not be empty")
	}
}

func TestLogError_NoOpWhenDirectoryMissing(t *testing.T) {
	// Setup: directory without .agentlog
	tmpDir := t.TempDir()

	// Act: should not panic or error
	LogError(tmpDir, "TEST_ERROR", "should not fail")

	// Assert: no file created
	errorsFile := filepath.Join(tmpDir, ".agentlog", "errors.jsonl")
	if _, err := os.Stat(errorsFile); !os.IsNotExist(err) {
		t.Error("should not create .agentlog directory when it doesn't exist")
	}
}

func TestLogError_NoOpInProduction(t *testing.T) {
	// Setup
	tmpDir := t.TempDir()
	agentlogDir := filepath.Join(tmpDir, ".agentlog")
	os.MkdirAll(agentlogDir, 0755)
	errorsFile := filepath.Join(agentlogDir, "errors.jsonl")
	os.WriteFile(errorsFile, []byte{}, 0644)

	// Set production env
	os.Setenv("PRODUCTION", "1")
	defer os.Unsetenv("PRODUCTION")

	// Act
	LogError(tmpDir, "TEST_ERROR", "should not be logged")

	// Assert: file should still be empty
	content, _ := os.ReadFile(errorsFile)
	if len(content) > 0 {
		t.Error("should not log in production mode")
	}
}

func TestLogError_TruncatesLongMessage(t *testing.T) {
	// Setup
	tmpDir := t.TempDir()
	agentlogDir := filepath.Join(tmpDir, ".agentlog")
	os.MkdirAll(agentlogDir, 0755)
	errorsFile := filepath.Join(agentlogDir, "errors.jsonl")
	os.WriteFile(errorsFile, []byte{}, 0644)

	// Create message longer than 500 chars
	longMessage := strings.Repeat("x", 600)

	// Act
	LogError(tmpDir, "TEST_ERROR", longMessage)

	// Assert: message should be truncated
	content, _ := os.ReadFile(errorsFile)
	var entry ErrorEntry
	json.Unmarshal(content[:len(content)-1], &entry)

	if len(entry.Message) > 500 {
		t.Errorf("Message length = %d, want <= 500", len(entry.Message))
	}
	if !strings.HasSuffix(entry.Message, "...") {
		t.Error("Truncated message should end with '...'")
	}
}

func TestLogErrorWithStack_IncludesStackTrace(t *testing.T) {
	// Setup
	tmpDir := t.TempDir()
	agentlogDir := filepath.Join(tmpDir, ".agentlog")
	os.MkdirAll(agentlogDir, 0755)
	errorsFile := filepath.Join(agentlogDir, "errors.jsonl")
	os.WriteFile(errorsFile, []byte{}, 0644)

	// Act
	LogErrorWithStack(tmpDir, "TEST_ERROR", "error message", "stack trace here")

	// Assert
	content, _ := os.ReadFile(errorsFile)
	var entry ErrorEntry
	json.Unmarshal(content[:len(content)-1], &entry)

	if entry.Context == nil {
		t.Fatal("Context should not be nil")
	}
	stackTrace, ok := entry.Context["stack_trace"].(string)
	if !ok {
		t.Fatal("stack_trace should be a string")
	}
	if stackTrace != "stack trace here" {
		t.Errorf("stack_trace = %q, want %q", stackTrace, "stack trace here")
	}
}

func TestLogErrorWithStack_TruncatesLongStackTrace(t *testing.T) {
	// Setup
	tmpDir := t.TempDir()
	agentlogDir := filepath.Join(tmpDir, ".agentlog")
	os.MkdirAll(agentlogDir, 0755)
	errorsFile := filepath.Join(agentlogDir, "errors.jsonl")
	os.WriteFile(errorsFile, []byte{}, 0644)

	// Create stack trace longer than 2KB
	longStack := strings.Repeat("x", 3000)

	// Act
	LogErrorWithStack(tmpDir, "TEST_ERROR", "error", longStack)

	// Assert
	content, _ := os.ReadFile(errorsFile)
	var entry ErrorEntry
	json.Unmarshal(content[:len(content)-1], &entry)

	stackTrace := entry.Context["stack_trace"].(string)
	if len(stackTrace) > 2048 {
		t.Errorf("stack_trace length = %d, want <= 2048", len(stackTrace))
	}
}

func TestLogError_AppendsToExistingFile(t *testing.T) {
	// Setup
	tmpDir := t.TempDir()
	agentlogDir := filepath.Join(tmpDir, ".agentlog")
	os.MkdirAll(agentlogDir, 0755)
	errorsFile := filepath.Join(agentlogDir, "errors.jsonl")

	// Write an existing entry
	existingEntry := `{"timestamp":"2025-01-01T00:00:00Z","source":"test","error_type":"EXISTING","message":"existing"}` + "\n"
	os.WriteFile(errorsFile, []byte(existingEntry), 0644)

	// Act
	LogError(tmpDir, "NEW_ERROR", "new error")

	// Assert: file should have two lines
	content, _ := os.ReadFile(errorsFile)
	lines := strings.Split(strings.TrimSpace(string(content)), "\n")
	if len(lines) != 2 {
		t.Errorf("expected 2 lines, got %d", len(lines))
	}
}

func TestLogError_SilentlyFailsOnWriteError(t *testing.T) {
	// Setup: directory exists but file is not writable
	tmpDir := t.TempDir()
	agentlogDir := filepath.Join(tmpDir, ".agentlog")
	os.MkdirAll(agentlogDir, 0755)
	errorsFile := filepath.Join(agentlogDir, "errors.jsonl")
	os.WriteFile(errorsFile, []byte{}, 0000) // no permissions
	defer os.Chmod(errorsFile, 0644)         // cleanup

	// Act: should not panic
	LogError(tmpDir, "TEST_ERROR", "should not fail")
	// If we get here without panic, test passes
}
