package cmd

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestFormatTailEntry_Human(t *testing.T) {
	entry := ErrorEntry{
		Timestamp: "2025-12-10T19:19:32.941Z",
		Source:    "frontend",
		ErrorType: "UNCAUGHT_ERROR",
		Message:   "Cannot read property 'foo' of undefined",
	}

	output := formatTailEntry(entry, false)

	// Check key elements are present
	if !strings.Contains(output, "Cannot read property") {
		t.Error("output should contain error message")
	}
	if !strings.Contains(output, "frontend") {
		t.Error("output should contain source")
	}
	if !strings.Contains(output, "UNCAUGHT_ERROR") {
		t.Error("output should contain error type")
	}
}

func TestFormatTailEntry_JSON(t *testing.T) {
	entry := ErrorEntry{
		Timestamp: "2025-12-10T19:19:32.941Z",
		Source:    "frontend",
		ErrorType: "UNCAUGHT_ERROR",
		Message:   "Test error",
	}

	output := formatTailEntry(entry, true)

	// Verify JSON structure
	if !strings.Contains(output, `"timestamp"`) {
		t.Error("JSON output should contain timestamp field")
	}
	if !strings.Contains(output, `"source"`) {
		t.Error("JSON output should contain source field")
	}
	if !strings.HasSuffix(strings.TrimSpace(output), "}") {
		t.Error("JSON output should be a valid JSON object")
	}
}

func TestTailFile_NoFile(t *testing.T) {
	tmpDir := t.TempDir()
	buf := new(bytes.Buffer)

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	err := tailFile(ctx, tmpDir, buf, false)
	if err == nil {
		t.Error("tailFile should return error for missing file")
	}
	if !os.IsNotExist(err) {
		t.Errorf("expected os.IsNotExist error, got: %v", err)
	}
}

func TestTailFile_ExistingEntries(t *testing.T) {
	// Setup temp directory with existing errors
	tmpDir := t.TempDir()
	agentlogDir := filepath.Join(tmpDir, ".agentlog")
	os.MkdirAll(agentlogDir, 0755)

	errorsFile := filepath.Join(agentlogDir, "errors.jsonl")
	os.WriteFile(errorsFile, []byte(
		`{"timestamp":"2025-12-10T19:19:32.941Z","source":"frontend","error_type":"UNCAUGHT_ERROR","message":"Error 1"}
{"timestamp":"2025-12-10T19:20:00.000Z","source":"backend","error_type":"DATABASE_ERROR","message":"Error 2"}
`), 0644)

	buf := new(bytes.Buffer)
	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()

	// Run tail - it should show existing entries then wait
	err := tailFile(ctx, tmpDir, buf, false)
	if err != nil && err != context.DeadlineExceeded {
		t.Errorf("tailFile returned unexpected error: %v", err)
	}

	output := buf.String()
	// Should have both existing errors
	if !strings.Contains(output, "Error 1") {
		t.Error("output should contain existing Error 1")
	}
	if !strings.Contains(output, "Error 2") {
		t.Error("output should contain existing Error 2")
	}
}

func TestTailFile_NewEntries(t *testing.T) {
	// Setup temp directory with initial file
	tmpDir := t.TempDir()
	agentlogDir := filepath.Join(tmpDir, ".agentlog")
	os.MkdirAll(agentlogDir, 0755)

	errorsFile := filepath.Join(agentlogDir, "errors.jsonl")
	os.WriteFile(errorsFile, []byte(
		`{"timestamp":"2025-12-10T19:19:32.941Z","source":"frontend","error_type":"UNCAUGHT_ERROR","message":"Initial Error"}
`), 0644)

	buf := new(bytes.Buffer)
	ctx, cancel := context.WithTimeout(context.Background(), 800*time.Millisecond)
	defer cancel()

	// Start tail in goroutine
	done := make(chan error, 1)
	go func() {
		done <- tailFile(ctx, tmpDir, buf, false)
	}()

	// Wait a bit then append new entry
	time.Sleep(200 * time.Millisecond)
	f, _ := os.OpenFile(errorsFile, os.O_APPEND|os.O_WRONLY, 0644)
	f.WriteString(`{"timestamp":"2025-12-10T19:21:00.000Z","source":"backend","error_type":"NEW_ERROR","message":"New Error Added"}` + "\n")
	f.Close()

	// Wait for tail to complete
	<-done

	output := buf.String()
	// Should have both initial and new error
	if !strings.Contains(output, "Initial Error") {
		t.Error("output should contain Initial Error")
	}
	if !strings.Contains(output, "New Error Added") {
		t.Error("output should contain New Error Added")
	}
}

func TestTailFile_JSONOutput(t *testing.T) {
	tmpDir := t.TempDir()
	agentlogDir := filepath.Join(tmpDir, ".agentlog")
	os.MkdirAll(agentlogDir, 0755)

	errorsFile := filepath.Join(agentlogDir, "errors.jsonl")
	os.WriteFile(errorsFile, []byte(
		`{"timestamp":"2025-12-10T19:19:32.941Z","source":"frontend","error_type":"UNCAUGHT_ERROR","message":"Test Error"}
`), 0644)

	buf := new(bytes.Buffer)
	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()

	err := tailFile(ctx, tmpDir, buf, true) // JSON mode
	if err != nil && err != context.DeadlineExceeded {
		t.Errorf("tailFile returned unexpected error: %v", err)
	}

	output := buf.String()
	// Should be JSON formatted
	if !strings.Contains(output, `"timestamp"`) {
		t.Error("JSON output should contain timestamp field")
	}
	if !strings.Contains(output, `"message":"Test Error"`) {
		t.Error("JSON output should contain the error message")
	}
}

func TestTailCommand_PathFlag(t *testing.T) {
	// Create temp directory with test data in a subdirectory (monorepo scenario)
	tmpDir := t.TempDir()
	subDir := filepath.Join(tmpDir, "packages", "worker")
	agentlogDir := filepath.Join(subDir, ".agentlog")
	os.MkdirAll(agentlogDir, 0755)

	// Write test errors in the subdirectory
	errorsFile := filepath.Join(agentlogDir, "errors.jsonl")
	os.WriteFile(errorsFile, []byte(
		`{"timestamp":"2025-12-10T19:19:32.941Z","source":"worker","error_type":"QUEUE_ERROR","message":"Custom path tail error"}
`), 0644)

	// Save and restore original state
	originalPath := pathOverride
	defer func() { pathOverride = originalPath }()

	// Set path override to the subdirectory
	pathOverride = subDir

	buf := new(bytes.Buffer)
	ctx, cancel := context.WithTimeout(context.Background(), 300*time.Millisecond)
	defer cancel()

	// Use tailFile directly with the resolved base directory
	baseDir := GetPathOverride()
	err := tailFile(ctx, baseDir, buf, false)
	if err != nil && err != context.DeadlineExceeded {
		t.Fatalf("tailFile() error = %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "Custom path tail error") {
		t.Errorf("output should contain error from custom path, got: %s", output)
	}
}
