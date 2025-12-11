package cmd

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestInitCommand_CreatesAgentlogDir(t *testing.T) {
	tmpDir := t.TempDir()
	origDir, _ := os.Getwd()
	defer os.Chdir(origDir)
	os.Chdir(tmpDir)

	// Create package.json to trigger TypeScript detection
	os.WriteFile(filepath.Join(tmpDir, "package.json"), []byte("{}"), 0644)

	result, err := runInit(tmpDir, false, "")
	if err != nil {
		t.Fatalf("init failed: %v", err)
	}

	// Check directory was created
	agentlogDir := filepath.Join(tmpDir, ".agentlog")
	if _, err := os.Stat(agentlogDir); os.IsNotExist(err) {
		t.Error(".agentlog directory was not created")
	}

	if !result.DirCreated {
		t.Error("DirCreated should be true")
	}
}

func TestInitCommand_CreatesErrorsFile(t *testing.T) {
	tmpDir := t.TempDir()

	_, err := runInit(tmpDir, false, "")
	if err != nil {
		t.Fatalf("init failed: %v", err)
	}

	errorsFile := filepath.Join(tmpDir, ".agentlog", "errors.jsonl")
	if _, err := os.Stat(errorsFile); os.IsNotExist(err) {
		t.Error(".agentlog/errors.jsonl was not created")
	}
}

func TestInitCommand_UpdatesGitignore_NewFile(t *testing.T) {
	tmpDir := t.TempDir()

	result, err := runInit(tmpDir, false, "")
	if err != nil {
		t.Fatalf("init failed: %v", err)
	}

	// Check gitignore was created
	gitignore := filepath.Join(tmpDir, ".gitignore")
	content, err := os.ReadFile(gitignore)
	if err != nil {
		t.Fatalf("failed to read .gitignore: %v", err)
	}

	if !strings.Contains(string(content), ".agentlog/errors.jsonl") {
		t.Error(".gitignore does not contain .agentlog/errors.jsonl")
	}

	if !result.GitIgnored {
		t.Error("GitIgnored should be true")
	}
}

func TestInitCommand_UpdatesGitignore_ExistingFile(t *testing.T) {
	tmpDir := t.TempDir()

	// Create existing gitignore
	gitignore := filepath.Join(tmpDir, ".gitignore")
	os.WriteFile(gitignore, []byte("node_modules/\n.env\n"), 0644)

	_, err := runInit(tmpDir, false, "")
	if err != nil {
		t.Fatalf("init failed: %v", err)
	}

	content, err := os.ReadFile(gitignore)
	if err != nil {
		t.Fatalf("failed to read .gitignore: %v", err)
	}

	// Should contain original content
	if !strings.Contains(string(content), "node_modules/") {
		t.Error("original gitignore content was lost")
	}

	// Should contain new entry
	if !strings.Contains(string(content), ".agentlog/errors.jsonl") {
		t.Error(".agentlog/errors.jsonl not added to .gitignore")
	}
}

func TestInitCommand_SkipsGitignore_AlreadyPresent(t *testing.T) {
	tmpDir := t.TempDir()

	// Create gitignore with agentlog already present
	gitignore := filepath.Join(tmpDir, ".gitignore")
	os.WriteFile(gitignore, []byte(".agentlog/errors.jsonl\n"), 0644)

	result, err := runInit(tmpDir, false, "")
	if err != nil {
		t.Fatalf("init failed: %v", err)
	}

	// GitIgnored should be false since no change was made
	if result.GitIgnored {
		t.Error("GitIgnored should be false when already present")
	}

	// Content should not be duplicated
	content, _ := os.ReadFile(gitignore)
	count := strings.Count(string(content), ".agentlog/errors.jsonl")
	if count != 1 {
		t.Errorf("expected 1 occurrence, found %d", count)
	}
}

func TestInitCommand_DetectsStack(t *testing.T) {
	tests := []struct {
		name          string
		markerFile    string
		expectedStack string
	}{
		{"TypeScript", "package.json", "typescript"},
		{"Go", "go.mod", "go"},
		{"Python", "pyproject.toml", "python"},
		{"Rust", "Cargo.toml", "rust"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			os.WriteFile(filepath.Join(tmpDir, tc.markerFile), []byte(""), 0644)

			result, err := runInit(tmpDir, false, "")
			if err != nil {
				t.Fatalf("init failed: %v", err)
			}

			if result.Stack != tc.expectedStack {
				t.Errorf("expected stack %s, got %s", tc.expectedStack, result.Stack)
			}

			if !result.Detected {
				t.Error("Detected should be true")
			}
		})
	}
}

func TestInitCommand_DefaultsToTypeScript(t *testing.T) {
	tmpDir := t.TempDir()

	result, err := runInit(tmpDir, false, "")
	if err != nil {
		t.Fatalf("init failed: %v", err)
	}

	if result.Stack != "typescript" {
		t.Errorf("expected default stack typescript, got %s", result.Stack)
	}

	if result.Detected {
		t.Error("Detected should be false for default")
	}
}

func TestInitCommand_StackOverride(t *testing.T) {
	tmpDir := t.TempDir()
	// Create package.json but override to Go
	os.WriteFile(filepath.Join(tmpDir, "package.json"), []byte("{}"), 0644)

	result, err := runInit(tmpDir, false, "go")
	if err != nil {
		t.Fatalf("init failed: %v", err)
	}

	if result.Stack != "go" {
		t.Errorf("expected overridden stack go, got %s", result.Stack)
	}

	if result.Detected {
		t.Error("Detected should be false when overridden")
	}
}

func TestInitCommand_Idempotent(t *testing.T) {
	tmpDir := t.TempDir()

	// Run init twice
	result1, err := runInit(tmpDir, false, "")
	if err != nil {
		t.Fatalf("first init failed: %v", err)
	}

	result2, err := runInit(tmpDir, false, "")
	if err != nil {
		t.Fatalf("second init failed: %v", err)
	}

	// First run should create dir
	if !result1.DirCreated {
		t.Error("first run should create dir")
	}

	// Second run should not create (already exists)
	if result2.DirCreated {
		t.Error("second run should not report dir created")
	}

	// Second run should not update gitignore (already present)
	if result2.GitIgnored {
		t.Error("second run should not update gitignore")
	}
}

func TestInitCommand_JSONOutput(t *testing.T) {
	tmpDir := t.TempDir()
	os.WriteFile(filepath.Join(tmpDir, "package.json"), []byte("{}"), 0644)

	result, err := runInit(tmpDir, false, "")
	if err != nil {
		t.Fatalf("init failed: %v", err)
	}

	// Verify JSON marshaling works
	output := &bytes.Buffer{}
	if err := json.NewEncoder(output).Encode(result); err != nil {
		t.Fatalf("failed to encode result as JSON: %v", err)
	}

	// Verify it can be decoded back
	var decoded InitResult
	if err := json.NewDecoder(bytes.NewReader(output.Bytes())).Decode(&decoded); err != nil {
		t.Fatalf("failed to decode JSON result: %v", err)
	}

	if decoded.Stack != "typescript" {
		t.Errorf("expected stack typescript in JSON, got %s", decoded.Stack)
	}
}

func TestInitCommand_ReturnsSnippet(t *testing.T) {
	tests := []struct {
		stack           string
		expectedContain string
	}{
		{"typescript", "window.onerror"},
		{"go", "recover()"},
		{"python", "sys.excepthook"},
		{"rust", "panic::set_hook"},
	}

	for _, tc := range tests {
		t.Run(tc.stack, func(t *testing.T) {
			tmpDir := t.TempDir()

			result, err := runInit(tmpDir, false, tc.stack)
			if err != nil {
				t.Fatalf("init failed: %v", err)
			}

			if !strings.Contains(result.Snippet, tc.expectedContain) {
				t.Errorf("snippet for %s should contain %s", tc.stack, tc.expectedContain)
			}
		})
	}
}

func TestTypeScriptSnippet_BrowserCompatible(t *testing.T) {
	snippet := getSnippet("typescript")

	// Should NOT use Node.js fs module (doesn't work in browser)
	if strings.Contains(snippet, "require('fs')") {
		t.Error("TypeScript snippet should not use require('fs') - not browser compatible")
	}
	if strings.Contains(snippet, "require('path')") {
		t.Error("TypeScript snippet should not use require('path') - not browser compatible")
	}

	// Should use fetch API for browser compatibility
	if !strings.Contains(snippet, "fetch") {
		t.Error("TypeScript snippet should use fetch API for browser compatibility")
	}
}

func TestTypeScriptSnippet_RequiredJSONLFields(t *testing.T) {
	snippet := getSnippet("typescript")

	// Must include all required JSONL fields per schema
	requiredFields := []string{"timestamp", "source", "error_type", "message"}
	for _, field := range requiredFields {
		if !strings.Contains(snippet, field) {
			t.Errorf("TypeScript snippet must include required JSONL field: %s", field)
		}
	}
}

func TestTypeScriptSnippet_ErrorHandlers(t *testing.T) {
	snippet := getSnippet("typescript")

	// Must capture uncaught errors
	if !strings.Contains(snippet, "window.onerror") {
		t.Error("TypeScript snippet must capture uncaught errors via window.onerror")
	}

	// Must capture unhandled promise rejections
	if !strings.Contains(snippet, "onunhandledrejection") {
		t.Error("TypeScript snippet must capture unhandled promise rejections")
	}
}

func TestTypeScriptSnippet_DevModeCheck(t *testing.T) {
	snippet := getSnippet("typescript")

	// Must check for development mode (should no-op in production)
	hasDevCheck := strings.Contains(snippet, "NODE_ENV") ||
		strings.Contains(snippet, "DEV") ||
		strings.Contains(snippet, "development")
	if !hasDevCheck {
		t.Error("TypeScript snippet should check for development mode")
	}
}
