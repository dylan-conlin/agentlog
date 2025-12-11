package cmd

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
)

func TestRootCommand_Help(t *testing.T) {
	// Reset for test
	rootCmd.SetArgs([]string{"--help"})

	buf := new(bytes.Buffer)
	rootCmd.SetOut(buf)
	rootCmd.SetErr(buf)

	// Execute should not error for help
	_ = rootCmd.Execute()

	output := buf.String()
	if !strings.Contains(output, "agentlog") {
		t.Errorf("help output should contain 'agentlog', got: %s", output)
	}
	if !strings.Contains(output, "AI agents") {
		t.Errorf("help output should contain 'AI agents', got: %s", output)
	}
}

func TestRootCommand_AIHelp(t *testing.T) {
	// Test the actual printAIHelp output
	buf := new(bytes.Buffer)
	printAIHelpTo(buf)
	output := buf.String()

	// Verify output is valid JSON
	var parsed CommandMetadata
	if err := json.Unmarshal([]byte(output), &parsed); err != nil {
		t.Fatalf("output is not valid JSON: %v\nOutput was: %s", err, output)
	}

	// Verify required fields
	if parsed.Name != "agentlog" {
		t.Errorf("expected name 'agentlog', got '%s'", parsed.Name)
	}
	if parsed.Version == "" {
		t.Error("version should not be empty")
	}
	if parsed.Description == "" {
		t.Error("description should not be empty")
	}

	// Verify commands are included
	if len(parsed.Commands) == 0 {
		t.Error("commands should not be empty")
	}

	// Verify expected commands exist
	expectedCommands := []string{"init", "errors", "tail", "doctor", "prime"}
	for _, expected := range expectedCommands {
		found := false
		for _, cmd := range parsed.Commands {
			if cmd.Name == expected {
				found = true
				if cmd.Description == "" {
					t.Errorf("command '%s' should have a description", expected)
				}
				if cmd.Usage == "" {
					t.Errorf("command '%s' should have usage", expected)
				}
				break
			}
		}
		if !found {
			t.Errorf("expected command '%s' not found in output", expected)
		}
	}

	// Verify global flags
	if parsed.GlobalFlags == nil || len(parsed.GlobalFlags) == 0 {
		t.Error("global_flags should not be empty")
	}
	if _, ok := parsed.GlobalFlags["--json"]; !ok {
		t.Error("global_flags should include --json")
	}
	if _, ok := parsed.GlobalFlags["--ai-help"]; !ok {
		t.Error("global_flags should include --ai-help")
	}
}

func TestAIHelp_CommandFlags(t *testing.T) {
	// Verify the errors command has its flags documented
	buf := new(bytes.Buffer)
	printAIHelpTo(buf)

	var parsed CommandMetadata
	if err := json.Unmarshal(buf.Bytes(), &parsed); err != nil {
		t.Fatalf("failed to parse output: %v", err)
	}

	// Find errors command
	var errorsCmd *CommandInfo
	for i := range parsed.Commands {
		if parsed.Commands[i].Name == "errors" {
			errorsCmd = &parsed.Commands[i]
			break
		}
	}

	if errorsCmd == nil {
		t.Fatal("errors command not found")
	}

	// Verify errors command has expected flags
	expectedFlags := []string{"--limit", "--source", "--type", "--since"}
	for _, flag := range expectedFlags {
		if _, ok := errorsCmd.Flags[flag]; !ok {
			t.Errorf("errors command should have flag '%s'", flag)
		}
	}
}

func TestIsTTY(t *testing.T) {
	// IsTTY should return a boolean (we can't easily mock stdout)
	result := IsTTY()
	// Just verify it doesn't panic and returns a boolean
	_ = result
}

func TestIsJSONOutput(t *testing.T) {
	// Default should be false
	if IsJSONOutput() {
		t.Error("JSON output should be false by default")
	}
}

func TestGetBaseDir(t *testing.T) {
	// Save original state
	originalPath := pathOverride

	tests := []struct {
		name         string
		pathOverride string
		want         string
	}{
		{
			name:         "no override returns empty (caller uses cwd)",
			pathOverride: "",
			want:         "",
		},
		{
			name:         "override returns custom path",
			pathOverride: "/custom/project/path",
			want:         "/custom/project/path",
		},
		{
			name:         "relative path works",
			pathOverride: "./subdir",
			want:         "./subdir",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set path override
			pathOverride = tt.pathOverride

			got := GetPathOverride()
			if got != tt.want {
				t.Errorf("GetPathOverride() = %v, want %v", got, tt.want)
			}
		})
	}

	// Restore original state
	pathOverride = originalPath
}

func TestGetErrorsPath(t *testing.T) {
	tests := []struct {
		name    string
		baseDir string
		want    string
	}{
		{
			name:    "constructs standard path",
			baseDir: "/project",
			want:    "/project/.agentlog/errors.jsonl",
		},
		{
			name:    "handles trailing slash",
			baseDir: "/project/",
			want:    "/project/.agentlog/errors.jsonl",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GetErrorsPath(tt.baseDir)
			if got != tt.want {
				t.Errorf("GetErrorsPath(%q) = %v, want %v", tt.baseDir, got, tt.want)
			}
		})
	}
}

func TestPathFlagInGlobalFlags(t *testing.T) {
	// Verify --path is documented in AI help output
	buf := new(bytes.Buffer)
	printAIHelpTo(buf)

	var parsed CommandMetadata
	if err := json.Unmarshal(buf.Bytes(), &parsed); err != nil {
		t.Fatalf("failed to parse output: %v", err)
	}

	if _, ok := parsed.GlobalFlags["--path"]; !ok {
		t.Error("global_flags should include --path")
	}
}
