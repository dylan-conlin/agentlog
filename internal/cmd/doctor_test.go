package cmd

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestDoctorCommand_NoAgentlogDir(t *testing.T) {
	// Create temp directory without .agentlog
	tmpDir := t.TempDir()

	// Save and restore working directory
	origDir, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(origDir)

	buf := new(bytes.Buffer)
	doctorCmd.SetOut(buf)
	doctorCmd.SetErr(buf)

	err := runDoctor(doctorCmd, []string{})
	if err != nil {
		t.Fatalf("runDoctor() error = %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "NOT FOUND") && !strings.Contains(output, "not found") {
		t.Errorf("output should indicate directory not found, got: %s", output)
	}
	if !strings.Contains(output, "agentlog init") {
		t.Errorf("output should suggest running 'agentlog init', got: %s", output)
	}
}

func TestDoctorCommand_DirExistsNoFile(t *testing.T) {
	// Create temp directory with .agentlog but no errors.jsonl
	tmpDir := t.TempDir()
	agentlogDir := filepath.Join(tmpDir, ".agentlog")
	os.MkdirAll(agentlogDir, 0755)

	origDir, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(origDir)

	buf := new(bytes.Buffer)
	doctorCmd.SetOut(buf)
	doctorCmd.SetErr(buf)

	err := runDoctor(doctorCmd, []string{})
	if err != nil {
		t.Fatalf("runDoctor() error = %v", err)
	}

	output := buf.String()
	// Should report directory OK but file missing
	if !strings.Contains(output, "OK") && !strings.Contains(output, "ok") {
		t.Errorf("output should show directory check OK, got: %s", output)
	}
}

func TestDoctorCommand_ValidJSONL(t *testing.T) {
	tmpDir := t.TempDir()
	agentlogDir := filepath.Join(tmpDir, ".agentlog")
	os.MkdirAll(agentlogDir, 0755)

	// Write valid JSONL file
	errorsFile := filepath.Join(agentlogDir, "errors.jsonl")
	os.WriteFile(errorsFile, []byte(
		`{"timestamp":"2025-12-10T19:19:32.941Z","source":"frontend","error_type":"UNCAUGHT_ERROR","message":"Error 1"}
{"timestamp":"2025-12-10T19:20:00.000Z","source":"backend","error_type":"DATABASE_ERROR","message":"Error 2"}
`), 0644)

	origDir, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(origDir)

	buf := new(bytes.Buffer)
	doctorCmd.SetOut(buf)
	doctorCmd.SetErr(buf)

	err := runDoctor(doctorCmd, []string{})
	if err != nil {
		t.Fatalf("runDoctor() error = %v", err)
	}

	output := buf.String()
	// Should show all checks passed
	if !strings.Contains(output, "OK") && !strings.Contains(output, "ok") {
		t.Errorf("output should show checks passed, got: %s", output)
	}
	if strings.Contains(output, "ERROR") || strings.Contains(output, "FAIL") {
		t.Errorf("output should not contain errors for valid file, got: %s", output)
	}
}

func TestDoctorCommand_InvalidJSONL(t *testing.T) {
	tmpDir := t.TempDir()
	agentlogDir := filepath.Join(tmpDir, ".agentlog")
	os.MkdirAll(agentlogDir, 0755)

	// Write file with invalid JSON lines
	errorsFile := filepath.Join(agentlogDir, "errors.jsonl")
	os.WriteFile(errorsFile, []byte(
		`{"timestamp":"2025-12-10T19:19:32.941Z","source":"frontend","error_type":"UNCAUGHT_ERROR","message":"Error 1"}
{invalid json line}
not json at all
{"timestamp":"2025-12-10T19:20:00.000Z","source":"backend","error_type":"DATABASE_ERROR","message":"Error 2"}
`), 0644)

	origDir, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(origDir)

	buf := new(bytes.Buffer)
	doctorCmd.SetOut(buf)
	doctorCmd.SetErr(buf)

	err := runDoctor(doctorCmd, []string{})
	if err != nil {
		t.Fatalf("runDoctor() error = %v", err)
	}

	output := buf.String()
	// Should report malformed lines
	if !strings.Contains(output, "malformed") && !strings.Contains(output, "invalid") {
		t.Errorf("output should report malformed lines, got: %s", output)
	}
}

func TestDoctorCommand_FileSizeWarning(t *testing.T) {
	tmpDir := t.TempDir()
	agentlogDir := filepath.Join(tmpDir, ".agentlog")
	os.MkdirAll(agentlogDir, 0755)

	// Write a file larger than 8MB (warning threshold at 80% of 10MB limit)
	errorsFile := filepath.Join(agentlogDir, "errors.jsonl")
	// Create ~9MB file
	line := `{"timestamp":"2025-12-10T19:19:32.941Z","source":"frontend","error_type":"UNCAUGHT_ERROR","message":"` + strings.Repeat("x", 400) + `"}` + "\n"
	var content strings.Builder
	for content.Len() < 9*1024*1024 {
		content.WriteString(line)
	}
	os.WriteFile(errorsFile, []byte(content.String()), 0644)

	origDir, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(origDir)

	buf := new(bytes.Buffer)
	doctorCmd.SetOut(buf)
	doctorCmd.SetErr(buf)

	err := runDoctor(doctorCmd, []string{})
	if err != nil {
		t.Fatalf("runDoctor() error = %v", err)
	}

	output := buf.String()
	// Should warn about file size approaching limit
	if !strings.Contains(output, "WARNING") && !strings.Contains(output, "warning") && !strings.Contains(output, "large") {
		t.Errorf("output should warn about large file size, got: %s", output)
	}
}

func TestDoctorCommand_JSONOutput(t *testing.T) {
	tmpDir := t.TempDir()
	agentlogDir := filepath.Join(tmpDir, ".agentlog")
	os.MkdirAll(agentlogDir, 0755)

	errorsFile := filepath.Join(agentlogDir, "errors.jsonl")
	os.WriteFile(errorsFile, []byte(
		`{"timestamp":"2025-12-10T19:19:32.941Z","source":"frontend","error_type":"UNCAUGHT_ERROR","message":"Error 1"}
`), 0644)

	origDir, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(origDir)

	// Enable JSON output
	jsonOutput = true
	defer func() { jsonOutput = false }()

	buf := new(bytes.Buffer)
	doctorCmd.SetOut(buf)
	doctorCmd.SetErr(buf)

	err := runDoctor(doctorCmd, []string{})
	if err != nil {
		t.Fatalf("runDoctor() error = %v", err)
	}

	output := buf.String()
	// Should be valid JSON
	if !strings.HasPrefix(strings.TrimSpace(output), "{") {
		t.Errorf("JSON output should start with {, got: %s", output)
	}
	if !strings.Contains(output, `"status"`) {
		t.Errorf("JSON output should contain status field, got: %s", output)
	}
}

func TestCheckHealth(t *testing.T) {
	tests := []struct {
		name       string
		setup      func(t *testing.T) string
		wantStatus string
		wantChecks int
	}{
		{
			name: "healthy setup",
			setup: func(t *testing.T) string {
				tmpDir := t.TempDir()
				dir := filepath.Join(tmpDir, ".agentlog")
				os.MkdirAll(dir, 0755)
				f := filepath.Join(dir, "errors.jsonl")
				os.WriteFile(f, []byte(`{"timestamp":"2025-12-10T19:19:32.941Z","source":"frontend","error_type":"UNCAUGHT_ERROR","message":"Error 1"}`+"\n"), 0644)
				return tmpDir
			},
			wantStatus: "healthy",
			wantChecks: 3, // directory, file, jsonl valid
		},
		{
			name: "missing directory",
			setup: func(t *testing.T) string {
				return t.TempDir()
			},
			wantStatus: "unhealthy",
			wantChecks: 1, // only directory check runs
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			baseDir := tt.setup(t)
			result := checkHealth(baseDir)

			if result.Status != tt.wantStatus {
				t.Errorf("checkHealth() Status = %v, want %v", result.Status, tt.wantStatus)
			}
			if len(result.Checks) != tt.wantChecks {
				t.Errorf("checkHealth() Checks count = %v, want %v", len(result.Checks), tt.wantChecks)
			}
		})
	}
}
