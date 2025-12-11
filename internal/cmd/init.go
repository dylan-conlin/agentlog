package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/agentlog/agentlog/internal/detect"
	"github.com/spf13/cobra"
)

var (
	initForce    bool
	initStack    string
)

// InitResult contains the result of the init command
type InitResult struct {
	Stack       string `json:"stack"`
	Detected    bool   `json:"detected"`
	MarkerFile  string `json:"marker_file,omitempty"`
	DirCreated  bool   `json:"dir_created"`
	GitIgnored  bool   `json:"gitignore_updated"`
	SnippetLang string `json:"snippet_language"`
	Snippet     string `json:"snippet"`
}

// initCmd represents the init command
var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize agentlog in your project",
	Long: `Initialize agentlog in the current project.

This command will:
  1. Detect your project's tech stack (TypeScript, Go, Python, Rust)
  2. Create the .agentlog/ directory
  3. Add .agentlog/errors.jsonl to .gitignore
  4. Print a code snippet to capture errors in your detected language

Examples:
  agentlog init              # Auto-detect stack and initialize
  agentlog init --stack go   # Force Go stack
  agentlog init --json       # Output result as JSON`,
	RunE: func(cmd *cobra.Command, args []string) error {
		cwd, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("failed to get current directory: %w", err)
		}

		result, err := runInit(cwd, initForce, initStack)
		if err != nil {
			return err
		}

		if IsJSONOutput() {
			output, _ := json.MarshalIndent(result, "", "  ")
			fmt.Println(string(output))
			return nil
		}

		// Human-readable output
		printInitResult(result)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(initCmd)
	initCmd.Flags().BoolVar(&initForce, "force", false, "Reinitialize even if .agentlog/ already exists")
	initCmd.Flags().StringVar(&initStack, "stack", "", "Override stack detection (typescript, go, python, rust)")
}

// runInit performs the init operation and returns the result
func runInit(dir string, force bool, stackOverride string) (*InitResult, error) {
	result := &InitResult{}

	// Detect or override stack
	if stackOverride != "" {
		result.Stack = strings.ToLower(stackOverride)
		result.Detected = false
	} else {
		detection := detect.DetectStack(dir)
		result.Stack = detection.Stack.String()
		result.Detected = detection.Detected
		result.MarkerFile = detection.MarkerFile
	}
	result.SnippetLang = result.Stack

	// Create .agentlog directory
	agentlogDir := filepath.Join(dir, ".agentlog")
	if _, err := os.Stat(agentlogDir); os.IsNotExist(err) {
		if err := os.MkdirAll(agentlogDir, 0755); err != nil {
			return nil, fmt.Errorf("failed to create .agentlog directory: %w", err)
		}
		result.DirCreated = true
	}

	// Create errors.jsonl file (touch)
	errorsFile := filepath.Join(agentlogDir, "errors.jsonl")
	if _, err := os.Stat(errorsFile); os.IsNotExist(err) {
		if err := os.WriteFile(errorsFile, []byte{}, 0644); err != nil {
			return nil, fmt.Errorf("failed to create errors.jsonl: %w", err)
		}
	}

	// Update .gitignore
	gitignorePath := filepath.Join(dir, ".gitignore")
	gitignoreEntry := ".agentlog/errors.jsonl"

	gitignoreContent, err := os.ReadFile(gitignorePath)
	if err != nil && !os.IsNotExist(err) {
		return nil, fmt.Errorf("failed to read .gitignore: %w", err)
	}

	if !strings.Contains(string(gitignoreContent), gitignoreEntry) {
		var newContent string
		if len(gitignoreContent) == 0 {
			newContent = gitignoreEntry + "\n"
		} else {
			content := string(gitignoreContent)
			if !strings.HasSuffix(content, "\n") {
				content += "\n"
			}
			newContent = content + gitignoreEntry + "\n"
		}

		if err := os.WriteFile(gitignorePath, []byte(newContent), 0644); err != nil {
			return nil, fmt.Errorf("failed to update .gitignore: %w", err)
		}
		result.GitIgnored = true
	}

	// Get snippet
	result.Snippet = getSnippet(result.Stack)

	return result, nil
}

// printInitResult prints the init result in human-readable format
func printInitResult(result *InitResult) {
	// Stack detection
	if result.Detected {
		fmt.Printf("Detected stack: %s (from %s)\n\n", capitalize(result.Stack), result.MarkerFile)
	} else if result.Stack != "" {
		fmt.Printf("Using stack: %s\n\n", capitalize(result.Stack))
	}

	// Directory creation
	if result.DirCreated {
		fmt.Println("Created .agentlog/ directory")
	} else {
		fmt.Println(".agentlog/ directory already exists")
	}

	// Gitignore update
	if result.GitIgnored {
		fmt.Println("Added .agentlog/errors.jsonl to .gitignore")
	}

	fmt.Println()

	// Snippet
	fmt.Printf("Add this snippet to your %s code:\n\n", capitalize(result.Stack))
	fmt.Println("---")
	fmt.Println(result.Snippet)
	fmt.Println("---")
	fmt.Println()
	fmt.Println("Done! Run 'agentlog tail' to watch for errors.")
}

func capitalize(s string) string {
	if s == "" {
		return s
	}
	return strings.ToUpper(s[:1]) + s[1:]
}

// getSnippet returns the error capture snippet for the given stack
func getSnippet(stack string) string {
	switch stack {
	case "typescript":
		return snippetTypeScript
	case "go":
		return snippetGo
	case "python":
		return snippetPython
	case "rust":
		return snippetRust
	default:
		return snippetTypeScript
	}
}

const snippetTypeScript = `// === BROWSER (add to app entry point) ===
if (typeof window !== 'undefined' && import.meta.env?.DEV !== false) {
  const log = (type: string, msg: unknown, ctx?: object) =>
    fetch('/__agentlog', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({
        timestamp: new Date().toISOString(),
        source: 'frontend',
        error_type: type,
        message: String(msg).slice(0, 500),
        context: ctx,
      }),
    }).catch(() => {});

  window.onerror = (msg, src, line, col, err) =>
    log('UNCAUGHT_ERROR', msg, { file: src, line, column: col, stack_trace: err?.stack?.slice(0, 2048) });

  window.onunhandledrejection = (e) =>
    log('UNHANDLED_REJECTION', e.reason, { stack_trace: e.reason?.stack?.slice(0, 2048) });
}

// === DEV SERVER (vite.config.ts or similar) ===
// Add this plugin to handle /__agentlog POST requests:
import { appendFileSync, mkdirSync } from 'fs';
export const agentlogPlugin = () => ({
  name: 'agentlog',
  configureServer(server) {
    server.middlewares.use('/__agentlog', (req, res) => {
      if (req.method !== 'POST') return res.end();
      let body = '';
      req.on('data', c => body += c);
      req.on('end', () => {
        mkdirSync('.agentlog', { recursive: true });
        appendFileSync('.agentlog/errors.jsonl', body + '\n');
        res.end('ok');
      });
    });
  },
});`

const snippetGo = `// agentlog error handler - add to your main.go
package main

import (
	"encoding/json"
	"fmt"
	"os"
	"runtime/debug"
	"time"
)

func initAgentlog() {
	if os.Getenv("PRODUCTION") != "" {
		return // no-op in production
	}

	defer func() {
		if r := recover(); r != nil {
			logAgentError("PANIC", fmt.Sprintf("%v", r), string(debug.Stack()))
			panic(r) // re-panic after logging
		}
	}()
}

func logAgentError(errType, message, stackTrace string) {
	entry := map[string]interface{}{
		"timestamp":  time.Now().UTC().Format(time.RFC3339Nano),
		"source":     "backend",
		"error_type": errType,
		"message":    truncate(message, 500),
	}
	if stackTrace != "" {
		entry["context"] = map[string]string{"stack_trace": truncate(stackTrace, 2048)}
	}

	data, _ := json.Marshal(entry)
	f, _ := os.OpenFile(".agentlog/errors.jsonl", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	defer f.Close()
	f.WriteString(string(data) + "\n")
}

func truncate(s string, max int) string {
	if len(s) <= max { return s }
	return s[:max-3] + "..."
}`

const snippetPython = `# agentlog error handler - add to your main module
import sys
import os
import json
import traceback
from datetime import datetime, timezone

def init_agentlog():
    if os.environ.get('ENV') == 'production':
        return  # no-op in production

    original_excepthook = sys.excepthook

    def agentlog_excepthook(exc_type, exc_value, exc_tb):
        entry = {
            "timestamp": datetime.now(timezone.utc).isoformat(),
            "source": "backend",
            "error_type": "EXCEPTION",
            "message": str(exc_value)[:500],
            "context": {
                "stack_trace": "".join(traceback.format_exception(exc_type, exc_value, exc_tb))[:2048]
            }
        }

        os.makedirs('.agentlog', exist_ok=True)
        with open('.agentlog/errors.jsonl', 'a') as f:
            f.write(json.dumps(entry) + '\n')

        original_excepthook(exc_type, exc_value, exc_tb)

    sys.excepthook = agentlog_excepthook

# Call at application startup
init_agentlog()`

const snippetRust = `// agentlog error handler - add to your main.rs
use std::fs::{OpenOptions, create_dir_all};
use std::io::Write;
use std::panic;
use chrono::Utc;
use serde_json::json;

pub fn init_agentlog() {
    if std::env::var("PRODUCTION").is_ok() {
        return; // no-op in production
    }

    panic::set_hook(Box::new(|panic_info| {
        let message = panic_info.to_string();
        let location = panic_info.location()
            .map(|l| format!("{}:{}:{}", l.file(), l.line(), l.column()))
            .unwrap_or_default();

        let entry = json!({
            "timestamp": Utc::now().to_rfc3339(),
            "source": "backend",
            "error_type": "PANIC",
            "message": &message[..message.len().min(500)],
            "context": {
                "file": location
            }
        });

        let _ = create_dir_all(".agentlog");
        if let Ok(mut file) = OpenOptions::new()
            .create(true)
            .append(true)
            .open(".agentlog/errors.jsonl")
        {
            let _ = writeln!(file, "{}", entry);
        }
    }));
}

// Call at application startup
// fn main() { init_agentlog(); ... }`
