package cmd

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestPrimeCommand_NoFile(t *testing.T) {
	// Setup: temp dir without errors.jsonl
	tmpDir := t.TempDir()
	origDir, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(origDir)

	// Execute prime command
	summary, err := generatePrimeSummary()

	// Should return empty summary with specific indication
	if err != nil {
		t.Fatalf("expected no error for missing file, got: %v", err)
	}
	if summary.TotalErrors != 0 {
		t.Errorf("expected 0 total errors for missing file, got %d", summary.TotalErrors)
	}
	if !summary.NoLogFile {
		t.Error("expected NoLogFile to be true")
	}
}

func TestPrimeCommand_EmptyFile(t *testing.T) {
	// Setup: temp dir with empty errors.jsonl
	tmpDir := t.TempDir()
	agentlogDir := filepath.Join(tmpDir, ".agentlog")
	os.MkdirAll(agentlogDir, 0755)
	os.WriteFile(filepath.Join(agentlogDir, "errors.jsonl"), []byte(""), 0644)

	origDir, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(origDir)

	// Execute
	summary, err := generatePrimeSummary()

	if err != nil {
		t.Fatalf("expected no error for empty file, got: %v", err)
	}
	if summary.TotalErrors != 0 {
		t.Errorf("expected 0 total errors for empty file, got %d", summary.TotalErrors)
	}
	if summary.NoLogFile {
		t.Error("expected NoLogFile to be false for empty file")
	}
}

func TestPrimeCommand_SingleError(t *testing.T) {
	// Setup: temp dir with single error
	tmpDir := t.TempDir()
	agentlogDir := filepath.Join(tmpDir, ".agentlog")
	os.MkdirAll(agentlogDir, 0755)

	now := time.Now().UTC().Format(time.RFC3339Nano)
	errorLine := `{"timestamp":"` + now + `","source":"frontend","error_type":"UNCAUGHT_ERROR","message":"Test error"}`
	os.WriteFile(filepath.Join(agentlogDir, "errors.jsonl"), []byte(errorLine+"\n"), 0644)

	origDir, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(origDir)

	// Execute
	summary, err := generatePrimeSummary()

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if summary.TotalErrors != 1 {
		t.Errorf("expected 1 total error, got %d", summary.TotalErrors)
	}
	if summary.LastHourErrors != 1 {
		t.Errorf("expected 1 last hour error, got %d", summary.LastHourErrors)
	}
	if len(summary.TopErrorTypes) != 1 || summary.TopErrorTypes[0].ErrorType != "UNCAUGHT_ERROR" {
		t.Errorf("expected UNCAUGHT_ERROR as top type, got %v", summary.TopErrorTypes)
	}
	if len(summary.TopSources) != 1 || summary.TopSources[0].Source != "frontend" {
		t.Errorf("expected frontend as top source, got %v", summary.TopSources)
	}
}

func TestPrimeCommand_MultipleErrors(t *testing.T) {
	// Setup: temp dir with multiple errors
	tmpDir := t.TempDir()
	agentlogDir := filepath.Join(tmpDir, ".agentlog")
	os.MkdirAll(agentlogDir, 0755)

	now := time.Now().UTC()
	errors := []string{
		`{"timestamp":"` + now.Format(time.RFC3339Nano) + `","source":"frontend","error_type":"UNCAUGHT_ERROR","message":"Error 1"}`,
		`{"timestamp":"` + now.Add(-30*time.Minute).Format(time.RFC3339Nano) + `","source":"frontend","error_type":"UNCAUGHT_ERROR","message":"Error 2"}`,
		`{"timestamp":"` + now.Add(-1*time.Hour).Format(time.RFC3339Nano) + `","source":"backend","error_type":"NETWORK_ERROR","message":"Error 3"}`,
		`{"timestamp":"` + now.Add(-2*time.Hour).Format(time.RFC3339Nano) + `","source":"frontend","error_type":"NETWORK_ERROR","message":"Error 4"}`,
		`{"timestamp":"` + now.Add(-3*time.Hour).Format(time.RFC3339Nano) + `","source":"backend","error_type":"VALIDATION_ERROR","message":"Error 5"}`,
	}
	os.WriteFile(filepath.Join(agentlogDir, "errors.jsonl"), []byte(strings.Join(errors, "\n")+"\n"), 0644)

	origDir, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(origDir)

	// Execute
	summary, err := generatePrimeSummary()

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if summary.TotalErrors != 5 {
		t.Errorf("expected 5 total errors, got %d", summary.TotalErrors)
	}
	// Errors in last hour: Error 1, Error 2 (within 60 min)
	if summary.LastHourErrors != 2 {
		t.Errorf("expected 2 last hour errors, got %d", summary.LastHourErrors)
	}
	// Top error types should be sorted by count
	if len(summary.TopErrorTypes) < 2 {
		t.Fatalf("expected at least 2 error types, got %d", len(summary.TopErrorTypes))
	}
	// UNCAUGHT_ERROR: 2, NETWORK_ERROR: 2, VALIDATION_ERROR: 1
	// With ties, order may vary but both should have count 2
	if summary.TopErrorTypes[0].Count < 2 {
		t.Errorf("expected top error type count >= 2, got %d", summary.TopErrorTypes[0].Count)
	}
	// Top sources: frontend: 3, backend: 2
	if summary.TopSources[0].Source != "frontend" || summary.TopSources[0].Count != 3 {
		t.Errorf("expected frontend with 3 as top source, got %v", summary.TopSources[0])
	}
}

func TestPrimeCommand_JSONOutput(t *testing.T) {
	// Setup: temp dir with errors
	tmpDir := t.TempDir()
	agentlogDir := filepath.Join(tmpDir, ".agentlog")
	os.MkdirAll(agentlogDir, 0755)

	now := time.Now().UTC().Format(time.RFC3339Nano)
	errorLine := `{"timestamp":"` + now + `","source":"frontend","error_type":"UNCAUGHT_ERROR","message":"Test error"}`
	os.WriteFile(filepath.Join(agentlogDir, "errors.jsonl"), []byte(errorLine+"\n"), 0644)

	origDir, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(origDir)

	// Execute
	summary, err := generatePrimeSummary()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Format as JSON
	output := formatPrimeSummaryJSON(summary)

	// Should be valid JSON
	var parsed PrimeSummary
	if err := json.Unmarshal([]byte(output), &parsed); err != nil {
		t.Fatalf("JSON output is not valid: %v\nOutput: %s", err, output)
	}
	if parsed.TotalErrors != 1 {
		t.Errorf("expected 1 total error in JSON, got %d", parsed.TotalErrors)
	}
}

func TestPrimeCommand_MalformedLine(t *testing.T) {
	// Setup: temp dir with mix of valid and invalid lines
	tmpDir := t.TempDir()
	agentlogDir := filepath.Join(tmpDir, ".agentlog")
	os.MkdirAll(agentlogDir, 0755)

	now := time.Now().UTC().Format(time.RFC3339Nano)
	lines := []string{
		`{"timestamp":"` + now + `","source":"frontend","error_type":"UNCAUGHT_ERROR","message":"Valid error"}`,
		`{invalid json`,
		`{"timestamp":"` + now + `","source":"backend","error_type":"NETWORK_ERROR","message":"Another valid"}`,
		`not json at all`,
		``,
	}
	os.WriteFile(filepath.Join(agentlogDir, "errors.jsonl"), []byte(strings.Join(lines, "\n")+"\n"), 0644)

	origDir, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(origDir)

	// Execute
	summary, err := generatePrimeSummary()

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Should only count valid lines (2 valid errors)
	if summary.TotalErrors != 2 {
		t.Errorf("expected 2 valid errors (skipping malformed), got %d", summary.TotalErrors)
	}
}

func TestPrimeCommand_HumanReadableOutput(t *testing.T) {
	summary := PrimeSummary{
		TotalErrors:    12,
		Last24hErrors:  12,
		LastHourErrors: 5,
		TopErrorTypes: []ErrorTypeCount{
			{ErrorType: "UNCAUGHT_ERROR", Count: 7},
			{ErrorType: "NETWORK_ERROR", Count: 3},
		},
		TopSources: []SourceCount{
			{Source: "frontend", Count: 8},
			{Source: "backend", Count: 4},
		},
		ActionableTip: "Focus on UNCAUGHT_ERROR in frontend",
		GeneratedAt:   "2025-12-10T10:30:00Z",
	}

	output := formatPrimeSummaryHuman(summary)

	if !strings.Contains(output, "12 errors") {
		t.Errorf("expected '12 errors' in output, got: %s", output)
	}
	if !strings.Contains(output, "5 in last hour") {
		t.Errorf("expected '5 in last hour' in output, got: %s", output)
	}
	if !strings.Contains(output, "UNCAUGHT_ERROR") {
		t.Errorf("expected 'UNCAUGHT_ERROR' in output, got: %s", output)
	}
	if !strings.Contains(output, "frontend") {
		t.Errorf("expected 'frontend' in output, got: %s", output)
	}
}

func TestPrimeCommand_NoErrorsHumanOutput(t *testing.T) {
	summary := PrimeSummary{
		TotalErrors: 0,
		NoLogFile:   false,
	}

	output := formatPrimeSummaryHuman(summary)

	if !strings.Contains(output, "No errors logged") {
		t.Errorf("expected 'No errors logged' in output, got: %s", output)
	}
}

func TestPrimeCommand_NoLogFileHumanOutput(t *testing.T) {
	summary := PrimeSummary{
		TotalErrors: 0,
		NoLogFile:   true,
	}

	output := formatPrimeSummaryHuman(summary)

	if !strings.Contains(output, "No error log found") {
		t.Errorf("expected 'No error log found' in output, got: %s", output)
	}
	if !strings.Contains(output, "agentlog init") {
		t.Errorf("expected 'agentlog init' suggestion in output, got: %s", output)
	}
}

func TestPrimeCommand_Integration(t *testing.T) {
	// Test the full command execution
	tmpDir := t.TempDir()
	agentlogDir := filepath.Join(tmpDir, ".agentlog")
	os.MkdirAll(agentlogDir, 0755)

	now := time.Now().UTC().Format(time.RFC3339Nano)
	errorLine := `{"timestamp":"` + now + `","source":"frontend","error_type":"UNCAUGHT_ERROR","message":"Test error"}`
	os.WriteFile(filepath.Join(agentlogDir, "errors.jsonl"), []byte(errorLine+"\n"), 0644)

	origDir, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(origDir)

	// Execute command by calling the Run function directly
	buf := new(bytes.Buffer)
	primeCmd.SetOut(buf)
	primeCmd.SetErr(buf)

	// Call Run directly with empty args
	primeCmd.Run(primeCmd, []string{})

	output := buf.String()
	if !strings.Contains(output, "1 error") {
		t.Errorf("expected '1 error' in output, got: %s", output)
	}
}

func TestTopN_SortsCorrectly(t *testing.T) {
	counts := map[string]int{
		"NETWORK_ERROR":    5,
		"UNCAUGHT_ERROR":   10,
		"VALIDATION_ERROR": 3,
		"DATABASE_ERROR":   7,
	}

	result := topN(counts, 3)

	if len(result) != 3 {
		t.Fatalf("expected 3 results, got %d", len(result))
	}
	if result[0].ErrorType != "UNCAUGHT_ERROR" || result[0].Count != 10 {
		t.Errorf("expected first to be UNCAUGHT_ERROR (10), got %v", result[0])
	}
	if result[1].ErrorType != "DATABASE_ERROR" || result[1].Count != 7 {
		t.Errorf("expected second to be DATABASE_ERROR (7), got %v", result[1])
	}
	if result[2].ErrorType != "NETWORK_ERROR" || result[2].Count != 5 {
		t.Errorf("expected third to be NETWORK_ERROR (5), got %v", result[2])
	}
}

func TestGenerateTip(t *testing.T) {
	summary := PrimeSummary{
		TotalErrors: 10,
		TopErrorTypes: []ErrorTypeCount{
			{ErrorType: "NETWORK_ERROR", Count: 6},
		},
		TopSources: []SourceCount{
			{Source: "frontend", Count: 8},
		},
	}

	tip := generateTip(summary)

	if !strings.Contains(tip, "NETWORK_ERROR") {
		t.Errorf("tip should mention top error type, got: %s", tip)
	}
	if !strings.Contains(tip, "frontend") {
		t.Errorf("tip should mention top source, got: %s", tip)
	}
	if !strings.Contains(tip, "60%") {
		t.Errorf("tip should mention percentage, got: %s", tip)
	}
}
