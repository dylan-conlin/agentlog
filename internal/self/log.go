// Package self provides error logging for agentlog itself (dogfooding).
// These functions log agentlog CLI errors to .agentlog/errors.jsonl
// so agentlog can observe its own errors.
package self

import (
	"encoding/json"
	"os"
	"path/filepath"
	"time"
)

// LogError logs an error to .agentlog/errors.jsonl with source="cli".
// It silently no-ops if:
// - .agentlog directory doesn't exist (no auto-creation)
// - PRODUCTION environment variable is set
// - Any error occurs during logging (no infinite loops)
func LogError(baseDir, errType, message string) {
	LogErrorWithStack(baseDir, errType, message, "")
}

// LogErrorWithStack logs an error with a stack trace.
func LogErrorWithStack(baseDir, errType, message, stackTrace string) {
	// No-op in production
	if os.Getenv("PRODUCTION") != "" {
		return
	}

	// Check if .agentlog directory exists (don't create it)
	agentlogDir := filepath.Join(baseDir, ".agentlog")
	if _, err := os.Stat(agentlogDir); os.IsNotExist(err) {
		return
	}

	errorsFile := filepath.Join(agentlogDir, "errors.jsonl")

	// Build entry
	entry := map[string]interface{}{
		"timestamp":  time.Now().UTC().Format(time.RFC3339Nano),
		"source":     "cli",
		"error_type": errType,
		"message":    truncate(message, 500),
	}

	if stackTrace != "" {
		entry["context"] = map[string]string{
			"stack_trace": truncate(stackTrace, 2048),
		}
	}

	// Marshal to JSON
	data, err := json.Marshal(entry)
	if err != nil {
		return // silently fail
	}

	// Append to file
	f, err := os.OpenFile(errorsFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return // silently fail
	}
	defer f.Close()

	f.WriteString(string(data) + "\n")
}

// truncate truncates a string to max length with "..." suffix
func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max-3] + "..."
}
