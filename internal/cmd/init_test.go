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
		{"Ruby", "Gemfile", "ruby"},
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
		{"ruby", "rescue"},
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

// Rust snippet tests

func TestRustSnippet_UsesSerde(t *testing.T) {
	snippet := getSnippet("rust")

	// Must use serde_json for serialization (per task spec)
	if !strings.Contains(snippet, "serde_json") {
		t.Error("Rust snippet must use serde_json for serialization")
	}
}

func TestRustSnippet_RequiredJSONLFields(t *testing.T) {
	snippet := getSnippet("rust")

	// Must include all required JSONL fields per schema
	requiredFields := []string{"timestamp", "source", "error_type", "message"}
	for _, field := range requiredFields {
		if !strings.Contains(snippet, field) {
			t.Errorf("Rust snippet must include required JSONL field: %s", field)
		}
	}
}

func TestRustSnippet_PanicHandler(t *testing.T) {
	snippet := getSnippet("rust")

	// Must capture panics
	if !strings.Contains(snippet, "panic::set_hook") {
		t.Error("Rust snippet must capture panics via panic::set_hook")
	}
}

func TestRustSnippet_ProductionNoOp(t *testing.T) {
	snippet := getSnippet("rust")

	// Must check for production mode (should no-op in production)
	hasProductionCheck := strings.Contains(snippet, "PRODUCTION") ||
		strings.Contains(snippet, "production")
	if !hasProductionCheck {
		t.Error("Rust snippet should check for production mode and no-op")
	}
}

func TestRustSnippet_WritesToCorrectFile(t *testing.T) {
	snippet := getSnippet("rust")

	// Must write to correct file
	if !strings.Contains(snippet, ".agentlog/errors.jsonl") {
		t.Error("Rust snippet must write to .agentlog/errors.jsonl")
	}
}

// Python snippet tests

func TestPythonSnippet_RequiredJSONLFields(t *testing.T) {
	snippet := getSnippet("python")

	// Must include all required JSONL fields per schema
	requiredFields := []string{"timestamp", "source", "error_type", "message"}
	for _, field := range requiredFields {
		if !strings.Contains(snippet, field) {
			t.Errorf("Python snippet must include required JSONL field: %s", field)
		}
	}
}

func TestPythonSnippet_ExceptionHandler(t *testing.T) {
	snippet := getSnippet("python")

	// Must capture exceptions via sys.excepthook
	if !strings.Contains(snippet, "sys.excepthook") {
		t.Error("Python snippet must capture exceptions via sys.excepthook")
	}
}

func TestPythonSnippet_DevModeCheck(t *testing.T) {
	snippet := getSnippet("python")

	// Must check for production mode (should no-op in production)
	hasDevCheck := strings.Contains(snippet, "ENV") ||
		strings.Contains(snippet, "production") ||
		strings.Contains(snippet, "PRODUCTION")
	if !hasDevCheck {
		t.Error("Python snippet should check for production mode")
	}
}

func TestPythonSnippet_StdlibOnly(t *testing.T) {
	snippet := getSnippet("python")

	// Should only use stdlib modules (json, sys, os, traceback, datetime, pathlib)
	// No external deps like requests, logging frameworks, etc.
	bannedImports := []string{"import requests", "import httpx", "import aiohttp"}
	for _, banned := range bannedImports {
		if strings.Contains(snippet, banned) {
			t.Errorf("Python snippet should not use external dependency: %s", banned)
		}
	}

	// Must have json import for JSONL writing
	if !strings.Contains(snippet, "import json") {
		t.Error("Python snippet must import json")
	}
}

func TestPythonSnippet_WritesToCorrectPath(t *testing.T) {
	snippet := getSnippet("python")

	// Must write to .agentlog/errors.jsonl
	if !strings.Contains(snippet, ".agentlog") || !strings.Contains(snippet, "errors.jsonl") {
		t.Error("Python snippet must write to .agentlog/errors.jsonl")
	}
}

// Go snippet tests

func TestGoSnippet_RequiredJSONLFields(t *testing.T) {
	snippet := getSnippet("go")

	// Must include all required JSONL fields per schema
	requiredFields := []string{"timestamp", "source", "error_type", "message"}
	for _, field := range requiredFields {
		if !strings.Contains(snippet, field) {
			t.Errorf("Go snippet must include required JSONL field: %s", field)
		}
	}
}

func TestGoSnippet_PanicRecovery(t *testing.T) {
	snippet := getSnippet("go")

	// Must capture panics via recover()
	if !strings.Contains(snippet, "recover()") {
		t.Error("Go snippet must capture panics via recover()")
	}

	// Must use debug.Stack() for stack traces
	if !strings.Contains(snippet, "debug.Stack()") {
		t.Error("Go snippet must capture stack traces via debug.Stack()")
	}
}

func TestGoSnippet_DevModeCheck(t *testing.T) {
	snippet := getSnippet("go")

	// Must check for production mode (should no-op in production)
	hasProductionCheck := strings.Contains(snippet, "PRODUCTION") ||
		strings.Contains(snippet, "production")
	if !hasProductionCheck {
		t.Error("Go snippet should check for production mode to no-op")
	}
}

func TestGoSnippet_StackTraceCapture(t *testing.T) {
	snippet := getSnippet("go")

	// Must capture stack traces in context
	if !strings.Contains(snippet, "stack_trace") {
		t.Error("Go snippet must include stack_trace in context")
	}

	// Must truncate stack trace per schema (2048 bytes)
	if !strings.Contains(snippet, "2048") {
		t.Error("Go snippet must truncate stack_trace to 2048 bytes per schema")
	}
}

func TestGoSnippet_MessageTruncation(t *testing.T) {
	snippet := getSnippet("go")

	// Must truncate message per schema (500 chars)
	if !strings.Contains(snippet, "500") {
		t.Error("Go snippet must truncate message to 500 characters per schema")
	}
}

func TestGoSnippet_FileWriting(t *testing.T) {
	snippet := getSnippet("go")

	// Must write to .agentlog/errors.jsonl
	if !strings.Contains(snippet, ".agentlog/errors.jsonl") {
		t.Error("Go snippet must write to .agentlog/errors.jsonl")
	}

	// Must use append mode
	if !strings.Contains(snippet, "O_APPEND") {
		t.Error("Go snippet must use O_APPEND mode for file writing")
	}
}

// Ruby snippet tests

func TestRubySnippet_RequiredJSONLFields(t *testing.T) {
	snippet := getSnippet("ruby")

	// Must include all required JSONL fields per schema
	requiredFields := []string{"timestamp", "source", "error_type", "message"}
	for _, field := range requiredFields {
		if !strings.Contains(snippet, field) {
			t.Errorf("Ruby snippet must include required JSONL field: %s", field)
		}
	}
}

func TestRubySnippet_ExceptionHandler(t *testing.T) {
	snippet := getSnippet("ruby")

	// Must capture exceptions via middleware or rescue
	hasExceptionCapture := strings.Contains(snippet, "rescue") ||
		strings.Contains(snippet, "Exception")
	if !hasExceptionCapture {
		t.Error("Ruby snippet must capture exceptions via rescue or Exception handling")
	}
}

func TestRubySnippet_RailsDevModeCheck(t *testing.T) {
	snippet := getSnippet("ruby")

	// Must check for Rails development environment
	hasRailsEnvCheck := strings.Contains(snippet, "Rails.env") ||
		strings.Contains(snippet, "development")
	if !hasRailsEnvCheck {
		t.Error("Ruby snippet should check for Rails development environment")
	}
}

func TestRubySnippet_StdlibOnly(t *testing.T) {
	snippet := getSnippet("ruby")

	// Should only use stdlib/Rails core gems
	// No external gems like sentry-ruby, rollbar, etc.
	bannedGems := []string{"require 'sentry'", "require 'rollbar'", "require 'bugsnag'"}
	for _, banned := range bannedGems {
		if strings.Contains(snippet, banned) {
			t.Errorf("Ruby snippet should not use external gem: %s", banned)
		}
	}

	// Must have json require for JSONL writing
	if !strings.Contains(snippet, "json") {
		t.Error("Ruby snippet must use json for JSONL writing")
	}
}

func TestRubySnippet_WritesToCorrectPath(t *testing.T) {
	snippet := getSnippet("ruby")

	// Must write to .agentlog/errors.jsonl
	if !strings.Contains(snippet, ".agentlog") || !strings.Contains(snippet, "errors.jsonl") {
		t.Error("Ruby snippet must write to .agentlog/errors.jsonl")
	}
}

func TestRubySnippet_SourceIsBackend(t *testing.T) {
	snippet := getSnippet("ruby")

	// Must set source to 'backend'
	if !strings.Contains(snippet, "backend") {
		t.Error("Ruby snippet must set source to 'backend'")
	}
}

func TestRubySnippet_StackTraceCapture(t *testing.T) {
	snippet := getSnippet("ruby")

	// Must capture stack traces (backtrace in Ruby)
	if !strings.Contains(snippet, "backtrace") && !strings.Contains(snippet, "stack_trace") {
		t.Error("Ruby snippet must capture stack traces")
	}
}

// Rails/Turbo frontend tests - for browser-side error capture

func TestRubySnippet_FrontendErrorCapture(t *testing.T) {
	snippet := getSnippet("ruby")

	// Must include frontend JavaScript for browser-side error capture
	if !strings.Contains(snippet, "window.onerror") {
		t.Error("Ruby snippet must include frontend JavaScript with window.onerror for browser-side error capture")
	}

	if !strings.Contains(snippet, "onunhandledrejection") {
		t.Error("Ruby snippet must include frontend JavaScript with onunhandledrejection for promise errors")
	}
}

func TestRubySnippet_AgentlogRoute(t *testing.T) {
	snippet := getSnippet("ruby")

	// Must include Rails route for /__agentlog endpoint
	if !strings.Contains(snippet, "__agentlog") {
		t.Error("Ruby snippet must include /__agentlog route for frontend error posts")
	}
}

func TestRubySnippet_FrontendNoVite(t *testing.T) {
	snippet := getSnippet("ruby")

	// Frontend JavaScript should NOT use Vite-specific APIs
	if strings.Contains(snippet, "import.meta.env") {
		t.Error("Ruby snippet frontend should not use import.meta.env (Vite-specific)")
	}

	if strings.Contains(snippet, "configureServer") {
		t.Error("Ruby snippet should not include Vite plugin code")
	}
}

func TestRubySnippet_FrontendFetch(t *testing.T) {
	snippet := getSnippet("ruby")

	// Frontend must use fetch API to POST errors
	if !strings.Contains(snippet, "fetch") {
		t.Error("Ruby snippet frontend must use fetch API")
	}
}

func TestRubySnippet_RailsController(t *testing.T) {
	snippet := getSnippet("ruby")

	// Must include Rails controller for handling /__agentlog endpoint
	hasController := strings.Contains(snippet, "AgentlogController") ||
		strings.Contains(snippet, "controller")
	if !hasController {
		t.Error("Ruby snippet must include Rails controller or route handler")
	}
}
