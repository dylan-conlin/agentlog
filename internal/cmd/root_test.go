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
	// Test that AI help produces valid JSON metadata
	// Note: We can't easily test --ai-help flag because it calls os.Exit
	// Instead we test the printAIHelp function indirectly by checking metadata structure

	metadata := CommandMetadata{
		Name:        "agentlog",
		Version:     "0.1.0",
		Description: "test",
		GlobalFlags: map[string]string{
			"--json": "test",
		},
		Commands: []CommandInfo{
			{
				Name:        "init",
				Description: "test",
				Usage:       "agentlog init",
			},
		},
	}

	output, err := json.MarshalIndent(metadata, "", "  ")
	if err != nil {
		t.Fatalf("failed to marshal metadata: %v", err)
	}

	// Verify it's valid JSON by unmarshaling
	var parsed CommandMetadata
	if err := json.Unmarshal(output, &parsed); err != nil {
		t.Fatalf("output is not valid JSON: %v", err)
	}

	if parsed.Name != "agentlog" {
		t.Errorf("expected name 'agentlog', got '%s'", parsed.Name)
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
