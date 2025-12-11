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

func TestErrorEntry_ParseJSON(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    ErrorEntry
		wantErr bool
	}{
		{
			name:  "minimal valid entry",
			input: `{"timestamp":"2025-12-10T19:19:32.941Z","source":"frontend","error_type":"UNCAUGHT_ERROR","message":"Cannot read property 'foo'"}`,
			want: ErrorEntry{
				Timestamp: "2025-12-10T19:19:32.941Z",
				Source:    "frontend",
				ErrorType: "UNCAUGHT_ERROR",
				Message:   "Cannot read property 'foo'",
			},
			wantErr: false,
		},
		{
			name:  "entry with context",
			input: `{"timestamp":"2025-12-10T19:19:32.941Z","source":"backend","error_type":"DATABASE_ERROR","message":"Connection refused","context":{"endpoint":"/api/users"}}`,
			want: ErrorEntry{
				Timestamp: "2025-12-10T19:19:32.941Z",
				Source:    "backend",
				ErrorType: "DATABASE_ERROR",
				Message:   "Connection refused",
				Context:   map[string]interface{}{"endpoint": "/api/users"},
			},
			wantErr: false,
		},
		{
			name:    "invalid JSON",
			input:   `{invalid json}`,
			want:    ErrorEntry{},
			wantErr: true,
		},
		{
			name:    "empty string",
			input:   "",
			want:    ErrorEntry{},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var got ErrorEntry
			err := json.Unmarshal([]byte(tt.input), &got)
			if (err != nil) != tt.wantErr {
				t.Errorf("json.Unmarshal() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if got.Timestamp != tt.want.Timestamp {
					t.Errorf("Timestamp = %v, want %v", got.Timestamp, tt.want.Timestamp)
				}
				if got.Source != tt.want.Source {
					t.Errorf("Source = %v, want %v", got.Source, tt.want.Source)
				}
				if got.ErrorType != tt.want.ErrorType {
					t.Errorf("ErrorType = %v, want %v", got.ErrorType, tt.want.ErrorType)
				}
				if got.Message != tt.want.Message {
					t.Errorf("Message = %v, want %v", got.Message, tt.want.Message)
				}
			}
		})
	}
}

func TestParseSince(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
		check   func(t time.Time) bool
	}{
		{
			name:    "duration 1h",
			input:   "1h",
			wantErr: false,
			check: func(t time.Time) bool {
				diff := time.Since(t)
				return diff > 55*time.Minute && diff < 65*time.Minute
			},
		},
		{
			name:    "duration 30m",
			input:   "30m",
			wantErr: false,
			check: func(t time.Time) bool {
				diff := time.Since(t)
				return diff > 25*time.Minute && diff < 35*time.Minute
			},
		},
		{
			name:    "date format YYYY-MM-DD",
			input:   "2024-01-01",
			wantErr: false,
			check: func(parsed time.Time) bool {
				return parsed.Year() == 2024 && parsed.Month() == 1 && parsed.Day() == 1
			},
		},
		{
			name:    "invalid format",
			input:   "invalid",
			wantErr: true,
			check:   nil,
		},
		{
			name:    "empty string",
			input:   "",
			wantErr: true,
			check:   nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseSince(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseSince(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
				return
			}
			if !tt.wantErr && tt.check != nil && !tt.check(got) {
				t.Errorf("parseSince(%q) = %v, failed check", tt.input, got)
			}
		})
	}
}

func TestFilterErrors(t *testing.T) {
	now := time.Now()
	oneHourAgo := now.Add(-1 * time.Hour).Format(time.RFC3339)
	twoHoursAgo := now.Add(-2 * time.Hour).Format(time.RFC3339)

	entries := []ErrorEntry{
		{Timestamp: oneHourAgo, Source: "frontend", ErrorType: "UNCAUGHT_ERROR", Message: "Error 1"},
		{Timestamp: twoHoursAgo, Source: "backend", ErrorType: "DATABASE_ERROR", Message: "Error 2"},
		{Timestamp: oneHourAgo, Source: "frontend", ErrorType: "NETWORK_ERROR", Message: "Error 3"},
	}

	tests := []struct {
		name    string
		entries []ErrorEntry
		source  string
		errType string
		since   time.Time
		wantLen int
	}{
		{
			name:    "no filters",
			entries: entries,
			source:  "",
			errType: "",
			wantLen: 3,
		},
		{
			name:    "filter by source",
			entries: entries,
			source:  "frontend",
			errType: "",
			wantLen: 2,
		},
		{
			name:    "filter by type",
			entries: entries,
			source:  "",
			errType: "DATABASE_ERROR",
			wantLen: 1,
		},
		{
			name:    "filter by source and type",
			entries: entries,
			source:  "frontend",
			errType: "UNCAUGHT_ERROR",
			wantLen: 1,
		},
		{
			name:    "filter by since (90 min ago)",
			entries: entries,
			source:  "",
			errType: "",
			since:   now.Add(-90 * time.Minute),
			wantLen: 2, // only entries from 1 hour ago
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := filterErrors(tt.entries, tt.source, tt.errType, tt.since)
			if len(got) != tt.wantLen {
				t.Errorf("filterErrors() returned %d entries, want %d", len(got), tt.wantLen)
			}
		})
	}
}

func TestReadErrors(t *testing.T) {
	tests := []struct {
		name    string
		setup   func(t *testing.T) string // returns base directory
		wantLen int
		wantErr bool
	}{
		{
			name: "valid JSONL file",
			setup: func(t *testing.T) string {
				tmpDir := t.TempDir()
				dir := filepath.Join(tmpDir, ".agentlog")
				os.MkdirAll(dir, 0755)
				f := filepath.Join(dir, "errors.jsonl")
				os.WriteFile(f, []byte(
					`{"timestamp":"2025-12-10T19:19:32.941Z","source":"frontend","error_type":"UNCAUGHT_ERROR","message":"Error 1"}
{"timestamp":"2025-12-10T19:20:00.000Z","source":"backend","error_type":"DATABASE_ERROR","message":"Error 2"}
`), 0644)
				return tmpDir
			},
			wantLen: 2,
			wantErr: false,
		},
		{
			name: "file with malformed lines",
			setup: func(t *testing.T) string {
				tmpDir := t.TempDir()
				dir := filepath.Join(tmpDir, ".agentlog")
				os.MkdirAll(dir, 0755)
				f := filepath.Join(dir, "errors.jsonl")
				os.WriteFile(f, []byte(
					`{"timestamp":"2025-12-10T19:19:32.941Z","source":"frontend","error_type":"UNCAUGHT_ERROR","message":"Error 1"}
{invalid json line}
{"timestamp":"2025-12-10T19:20:00.000Z","source":"backend","error_type":"DATABASE_ERROR","message":"Error 2"}
`), 0644)
				return tmpDir
			},
			wantLen: 2, // skips malformed line
			wantErr: false,
		},
		{
			name: "missing file",
			setup: func(t *testing.T) string {
				tmpDir := t.TempDir()
				return tmpDir
			},
			wantLen: 0,
			wantErr: true,
		},
		{
			name: "empty file",
			setup: func(t *testing.T) string {
				tmpDir := t.TempDir()
				dir := filepath.Join(tmpDir, ".agentlog")
				os.MkdirAll(dir, 0755)
				f := filepath.Join(dir, "errors.jsonl")
				os.WriteFile(f, []byte(""), 0644)
				return tmpDir
			},
			wantLen: 0,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			baseDir := tt.setup(t)
			got, err := readErrors(baseDir)
			if (err != nil) != tt.wantErr {
				t.Errorf("readErrors() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if len(got) != tt.wantLen {
				t.Errorf("readErrors() returned %d entries, want %d", len(got), tt.wantLen)
			}
		})
	}
}

func TestFormatHuman(t *testing.T) {
	entries := []ErrorEntry{
		{
			Timestamp: "2025-12-10T19:19:32.941Z",
			Source:    "frontend",
			ErrorType: "UNCAUGHT_ERROR",
			Message:   "Cannot read property 'foo' of undefined",
		},
	}

	output := formatHuman(entries, 10)

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

func TestFormatJSON(t *testing.T) {
	entries := []ErrorEntry{
		{
			Timestamp: "2025-12-10T19:19:32.941Z",
			Source:    "frontend",
			ErrorType: "UNCAUGHT_ERROR",
			Message:   "Test error",
		},
	}

	output := formatJSON(entries)

	// Verify valid JSON array
	var parsed []ErrorEntry
	if err := json.Unmarshal([]byte(output), &parsed); err != nil {
		t.Errorf("formatJSON() output is not valid JSON: %v", err)
	}

	if len(parsed) != 1 {
		t.Errorf("expected 1 entry, got %d", len(parsed))
	}

	if parsed[0].Message != "Test error" {
		t.Errorf("expected message 'Test error', got '%s'", parsed[0].Message)
	}
}

func TestErrorsCommand_PathFlag(t *testing.T) {
	// Create temp directory with test data in a subdirectory
	tmpDir := t.TempDir()
	subDir := filepath.Join(tmpDir, "packages", "api")
	agentlogDir := filepath.Join(subDir, ".agentlog")
	os.MkdirAll(agentlogDir, 0755)

	// Write test errors in the subdirectory
	errorsFile := filepath.Join(agentlogDir, "errors.jsonl")
	os.WriteFile(errorsFile, []byte(
		`{"timestamp":"2025-12-10T19:19:32.941Z","source":"backend","error_type":"API_ERROR","message":"Custom path error"}
`), 0644)

	// Save and restore original state
	originalPath := pathOverride
	defer func() { pathOverride = originalPath }()

	// Set path override to the subdirectory
	pathOverride = subDir

	// Reset other flags
	errorsLimit = 10
	errorsSource = ""
	errorsType = ""
	errorsSince = ""
	jsonOutput = false

	buf := new(bytes.Buffer)
	errorsCmd.SetOut(buf)
	errorsCmd.SetErr(buf)

	err := runErrors(errorsCmd, []string{})
	if err != nil {
		t.Fatalf("runErrors() error = %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "Custom path error") {
		t.Errorf("output should contain error from custom path, got: %s", output)
	}
}

func TestErrorsCommand_Integration(t *testing.T) {
	// Create temp directory with test data
	tmpDir := t.TempDir()
	agentlogDir := filepath.Join(tmpDir, ".agentlog")
	os.MkdirAll(agentlogDir, 0755)

	// Write test errors
	errorsFile := filepath.Join(agentlogDir, "errors.jsonl")
	os.WriteFile(errorsFile, []byte(
		`{"timestamp":"2025-12-10T19:19:32.941Z","source":"frontend","error_type":"UNCAUGHT_ERROR","message":"Error 1"}
{"timestamp":"2025-12-10T19:20:00.000Z","source":"backend","error_type":"DATABASE_ERROR","message":"Error 2"}
{"timestamp":"2025-12-10T19:21:00.000Z","source":"frontend","error_type":"NETWORK_ERROR","message":"Error 3"}
`), 0644)

	// Save and restore working directory
	origDir, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(origDir)

	tests := []struct {
		name       string
		limit      int
		source     string
		errType    string
		since      string
		useJSON    bool
		wantInOut  []string
		wantNotOut []string
	}{
		{
			name:      "default output",
			limit:     10,
			wantInOut: []string{"Error 1", "Error 2", "Error 3"},
		},
		{
			name:       "filter by source",
			limit:      10,
			source:    "frontend",
			wantInOut:  []string{"Error 1", "Error 3"},
			wantNotOut: []string{"Error 2"},
		},
		{
			name:       "filter by type",
			limit:      10,
			errType:   "DATABASE_ERROR",
			wantInOut:  []string{"Error 2"},
			wantNotOut: []string{"Error 1", "Error 3"},
		},
		{
			name:      "limit results",
			limit:     2,
			wantInOut: []string{"Error"}, // at least some errors
		},
		{
			name:      "json output",
			limit:     10,
			useJSON:   true,
			wantInOut: []string{`"timestamp"`, `"source"`, `"error_type"`},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset global flags
			errorsLimit = tt.limit
			errorsSource = tt.source
			errorsType = tt.errType
			errorsSince = tt.since
			jsonOutput = tt.useJSON

			buf := new(bytes.Buffer)
			errorsCmd.SetOut(buf)
			errorsCmd.SetErr(buf)

			// Call RunE directly instead of Execute
			err := runErrors(errorsCmd, []string{})
			if err != nil {
				t.Fatalf("runErrors() error = %v", err)
			}

			output := buf.String()
			for _, want := range tt.wantInOut {
				if !strings.Contains(output, want) {
					t.Errorf("output should contain %q, got: %s", want, output)
				}
			}
			for _, notWant := range tt.wantNotOut {
				if strings.Contains(output, notWant) {
					t.Errorf("output should NOT contain %q, got: %s", notWant, output)
				}
			}

			// Reset jsonOutput for next test
			jsonOutput = false
		})
	}
}
