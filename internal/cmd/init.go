package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/agentlog/agentlog/internal/detect"
	"github.com/agentlog/agentlog/internal/self"
	"github.com/spf13/cobra"
)

var (
	initForce   bool
	initStack   string
	initInstall bool
)

// InstallAction represents a file operation performed during installation
type InstallAction struct {
	Path      string `json:"path"`
	Operation string `json:"operation"` // "create", "append", "insert"
}

// InitResult contains the result of the init command
type InitResult struct {
	Stack          string          `json:"stack"`
	Detected       bool            `json:"detected"`
	MarkerFile     string          `json:"marker_file,omitempty"`
	DirCreated     bool            `json:"dir_created"`
	GitIgnored     bool            `json:"gitignore_updated"`
	SnippetLang    string          `json:"snippet_language"`
	Snippet        string          `json:"snippet"`
	Installed      bool            `json:"installed"`
	InstallActions []InstallAction `json:"install_actions,omitempty"`
}

// initCmd represents the init command
var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize agentlog in your project",
	Long: `Initialize agentlog in the current project.

This command will:
  1. Detect your project's tech stack (TypeScript, Go, Python, Rust, Ruby)
  2. Create the .agentlog/ directory
  3. Add .agentlog/errors.jsonl to .gitignore
  4. Print a code snippet to capture errors in your detected language

With --install flag, agentlog will write files directly to your project:
  - Rails: Creates controller, initializer, adds route, appends to application.js
  - Other stacks: Creates .agentlog/capture.<ext> file you can import

Examples:
  agentlog init              # Auto-detect stack and print snippet
  agentlog init --install    # Auto-detect and install files
  agentlog init --stack go   # Force Go stack
  agentlog init --json       # Output result as JSON`,
	RunE: func(cmd *cobra.Command, args []string) error {
		cwd, err := os.Getwd()
		if err != nil {
			self.LogError(".", "GETWD_ERROR", err.Error())
			return fmt.Errorf("failed to get current directory: %w", err)
		}

		result, err := runInit(cwd, initForce, initStack, initInstall)
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
	initCmd.Flags().StringVar(&initStack, "stack", "", "Override stack detection (typescript, go, python, rust, ruby)")
	initCmd.Flags().BoolVar(&initInstall, "install", false, "Install snippets directly to project files")
}

// runInit performs the init operation and returns the result
func runInit(dir string, force bool, stackOverride string, install bool) (*InitResult, error) {
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
			self.LogError(dir, "MKDIR_ERROR", fmt.Sprintf("failed to create .agentlog directory: %v", err))
			return nil, fmt.Errorf("failed to create .agentlog directory: %w", err)
		}
		result.DirCreated = true
	}

	// Create errors.jsonl file (touch)
	errorsFile := filepath.Join(agentlogDir, "errors.jsonl")
	if _, err := os.Stat(errorsFile); os.IsNotExist(err) {
		if err := os.WriteFile(errorsFile, []byte{}, 0644); err != nil {
			self.LogError(dir, "FILE_CREATE_ERROR", fmt.Sprintf("failed to create errors.jsonl: %v", err))
			return nil, fmt.Errorf("failed to create errors.jsonl: %w", err)
		}
	}

	// Update .gitignore
	gitignorePath := filepath.Join(dir, ".gitignore")
	gitignoreEntry := ".agentlog/errors.jsonl"

	gitignoreContent, err := os.ReadFile(gitignorePath)
	if err != nil && !os.IsNotExist(err) {
		self.LogError(dir, "FILE_READ_ERROR", fmt.Sprintf("failed to read .gitignore: %v", err))
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
			self.LogError(dir, "FILE_WRITE_ERROR", fmt.Sprintf("failed to update .gitignore: %v", err))
			return nil, fmt.Errorf("failed to update .gitignore: %w", err)
		}
		result.GitIgnored = true
	}

	// Get snippet
	result.Snippet = getSnippet(result.Stack)

	// Install snippets if requested
	if install {
		actions, err := installSnippets(dir, result.Stack)
		if err != nil {
			return nil, err
		}
		result.Installed = true
		result.InstallActions = actions
	}

	return result, nil
}

// installSnippets writes snippet files to the project
func installSnippets(dir string, stack string) ([]InstallAction, error) {
	switch stack {
	case "ruby":
		return installRubySnippets(dir)
	case "typescript":
		return installTypeScriptSnippets(dir)
	case "node":
		return installNodeSnippets(dir)
	case "go":
		return installGoSnippets(dir)
	case "python":
		return installPythonSnippets(dir)
	case "rust":
		return installRustSnippets(dir)
	default:
		return installTypeScriptSnippets(dir)
	}
}

// installRubySnippets installs Rails-specific files
func installRubySnippets(dir string) ([]InstallAction, error) {
	var actions []InstallAction

	// 1. Create controller
	controllerDir := filepath.Join(dir, "app", "controllers")
	if err := os.MkdirAll(controllerDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create controllers directory: %w", err)
	}

	controllerPath := filepath.Join(controllerDir, "agentlog_controller.rb")
	if _, err := os.Stat(controllerPath); os.IsNotExist(err) {
		if err := os.WriteFile(controllerPath, []byte(rubyController), 0644); err != nil {
			return nil, fmt.Errorf("failed to create controller: %w", err)
		}
		actions = append(actions, InstallAction{Path: "app/controllers/agentlog_controller.rb", Operation: "create"})
	}

	// 2. Create initializer
	initializerDir := filepath.Join(dir, "config", "initializers")
	if err := os.MkdirAll(initializerDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create initializers directory: %w", err)
	}

	initializerPath := filepath.Join(initializerDir, "agentlog.rb")
	if _, err := os.Stat(initializerPath); os.IsNotExist(err) {
		if err := os.WriteFile(initializerPath, []byte(rubyInitializer), 0644); err != nil {
			return nil, fmt.Errorf("failed to create initializer: %w", err)
		}
		actions = append(actions, InstallAction{Path: "config/initializers/agentlog.rb", Operation: "create"})
	}

	// 3. Add route to config/routes.rb
	routesPath := filepath.Join(dir, "config", "routes.rb")
	routesContent, err := os.ReadFile(routesPath)
	if err == nil && !strings.Contains(string(routesContent), "__agentlog") {
		// Insert route before the final "end"
		newContent := insertRouteIntoRailsRoutes(string(routesContent))
		if err := os.WriteFile(routesPath, []byte(newContent), 0644); err != nil {
			return nil, fmt.Errorf("failed to update routes.rb: %w", err)
		}
		actions = append(actions, InstallAction{Path: "config/routes.rb", Operation: "insert"})
	}

	// 4. Append frontend JS to app/javascript/application.js
	jsPath := filepath.Join(dir, "app", "javascript", "application.js")
	jsContent, err := os.ReadFile(jsPath)
	if err == nil && !strings.Contains(string(jsContent), "window.onerror") {
		newContent := string(jsContent) + "\n" + rubyFrontendJS
		if err := os.WriteFile(jsPath, []byte(newContent), 0644); err != nil {
			return nil, fmt.Errorf("failed to update application.js: %w", err)
		}
		actions = append(actions, InstallAction{Path: "app/javascript/application.js", Operation: "append"})
	}

	return actions, nil
}

// insertRouteIntoRailsRoutes inserts the agentlog route before the final 'end'
func insertRouteIntoRailsRoutes(content string) string {
	lines := strings.Split(content, "\n")
	var result []string

	// Find the last 'end' line and insert before it
	lastEndIdx := -1
	for i := len(lines) - 1; i >= 0; i-- {
		if strings.TrimSpace(lines[i]) == "end" {
			lastEndIdx = i
			break
		}
	}

	if lastEndIdx == -1 {
		// No 'end' found, just append
		return content + "\n" + rubyRoute
	}

	for i, line := range lines {
		if i == lastEndIdx {
			result = append(result, "  "+rubyRoute)
		}
		result = append(result, line)
	}

	return strings.Join(result, "\n")
}

// installTypeScriptSnippets creates a capture.ts file
func installTypeScriptSnippets(dir string) ([]InstallAction, error) {
	var actions []InstallAction

	agentlogDir := filepath.Join(dir, ".agentlog")
	if err := os.MkdirAll(agentlogDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create .agentlog directory: %w", err)
	}

	capturePath := filepath.Join(agentlogDir, "capture.ts")
	if _, err := os.Stat(capturePath); os.IsNotExist(err) {
		if err := os.WriteFile(capturePath, []byte(typescriptCapture), 0644); err != nil {
			return nil, fmt.Errorf("failed to create capture.ts: %w", err)
		}
		actions = append(actions, InstallAction{Path: ".agentlog/capture.ts", Operation: "create"})
	}

	return actions, nil
}

// installNodeSnippets creates a capture.ts file for Node.js
func installNodeSnippets(dir string) ([]InstallAction, error) {
	var actions []InstallAction

	agentlogDir := filepath.Join(dir, ".agentlog")
	if err := os.MkdirAll(agentlogDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create .agentlog directory: %w", err)
	}

	capturePath := filepath.Join(agentlogDir, "capture.ts")
	if _, err := os.Stat(capturePath); os.IsNotExist(err) {
		if err := os.WriteFile(capturePath, []byte(nodeCapture), 0644); err != nil {
			return nil, fmt.Errorf("failed to create capture.ts: %w", err)
		}
		actions = append(actions, InstallAction{Path: ".agentlog/capture.ts", Operation: "create"})
	}

	return actions, nil
}

// installGoSnippets creates a capture.go file
func installGoSnippets(dir string) ([]InstallAction, error) {
	var actions []InstallAction

	agentlogDir := filepath.Join(dir, ".agentlog")
	if err := os.MkdirAll(agentlogDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create .agentlog directory: %w", err)
	}

	capturePath := filepath.Join(agentlogDir, "capture.go")
	if _, err := os.Stat(capturePath); os.IsNotExist(err) {
		if err := os.WriteFile(capturePath, []byte(snippetGo), 0644); err != nil {
			return nil, fmt.Errorf("failed to create capture.go: %w", err)
		}
		actions = append(actions, InstallAction{Path: ".agentlog/capture.go", Operation: "create"})
	}

	return actions, nil
}

// installPythonSnippets creates a capture.py file
func installPythonSnippets(dir string) ([]InstallAction, error) {
	var actions []InstallAction

	agentlogDir := filepath.Join(dir, ".agentlog")
	if err := os.MkdirAll(agentlogDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create .agentlog directory: %w", err)
	}

	capturePath := filepath.Join(agentlogDir, "capture.py")
	if _, err := os.Stat(capturePath); os.IsNotExist(err) {
		if err := os.WriteFile(capturePath, []byte(snippetPython), 0644); err != nil {
			return nil, fmt.Errorf("failed to create capture.py: %w", err)
		}
		actions = append(actions, InstallAction{Path: ".agentlog/capture.py", Operation: "create"})
	}

	return actions, nil
}

// installRustSnippets creates a capture.rs file
func installRustSnippets(dir string) ([]InstallAction, error) {
	var actions []InstallAction

	agentlogDir := filepath.Join(dir, ".agentlog")
	if err := os.MkdirAll(agentlogDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create .agentlog directory: %w", err)
	}

	capturePath := filepath.Join(agentlogDir, "capture.rs")
	if _, err := os.Stat(capturePath); os.IsNotExist(err) {
		if err := os.WriteFile(capturePath, []byte(snippetRust), 0644); err != nil {
			return nil, fmt.Errorf("failed to create capture.rs: %w", err)
		}
		actions = append(actions, InstallAction{Path: ".agentlog/capture.rs", Operation: "create"})
	}

	return actions, nil
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

	// Installation results
	if result.Installed {
		fmt.Println("Installed agentlog to your project:")
		for _, action := range result.InstallActions {
			switch action.Operation {
			case "create":
				fmt.Printf("  Created: %s\n", action.Path)
			case "append":
				fmt.Printf("  Modified: %s (appended error capture)\n", action.Path)
			case "insert":
				fmt.Printf("  Modified: %s (added route)\n", action.Path)
			}
		}

		// Stack-specific follow-up instructions
		fmt.Println()
		switch result.Stack {
		case "ruby":
			fmt.Println("Done! Run 'agentlog tail' to watch for errors.")
		case "typescript":
			fmt.Println("Import the capture file in your app entry point:")
			fmt.Println("  import './.agentlog/capture';")
			fmt.Println()
			fmt.Println("Done! Run 'agentlog tail' to watch for errors.")
		case "go":
			fmt.Println("Add to your main.go:")
			fmt.Println("  // import \".agentlog\"")
			fmt.Println("  // call initAgentlog() at startup")
			fmt.Println()
			fmt.Println("Done! Run 'agentlog tail' to watch for errors.")
		case "python":
			fmt.Println("Add to your main module:")
			fmt.Println("  from .agentlog.capture import init_agentlog")
			fmt.Println("  init_agentlog()")
			fmt.Println()
			fmt.Println("Done! Run 'agentlog tail' to watch for errors.")
		case "rust":
			fmt.Println("Add to your main.rs:")
			fmt.Println("  mod agentlog { include!(\".agentlog/capture.rs\"); }")
			fmt.Println("  agentlog::init_agentlog();")
			fmt.Println()
			fmt.Println("Done! Run 'agentlog tail' to watch for errors.")
		default:
			fmt.Println("Done! Run 'agentlog tail' to watch for errors.")
		}
	} else {
		// No installation - print snippet for manual copy/paste
		fmt.Printf("Add this snippet to your %s code:\n\n", capitalize(result.Stack))
		fmt.Println("---")
		fmt.Println(result.Snippet)
		fmt.Println("---")
		fmt.Println()
		fmt.Println("Done! Run 'agentlog tail' to watch for errors.")
	}
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
	case "node":
		return snippetNode
	case "go":
		return snippetGo
	case "python":
		return snippetPython
	case "rust":
		return snippetRust
	case "ruby":
		return snippetRuby
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

const snippetNode = `// agentlog error handler for Node.js - add to your app entry point
// Works with BullMQ workers, scrapers, CLI tools, and any Node.js service
import { appendFileSync, mkdirSync, existsSync, readFileSync, writeFileSync } from 'fs';

const AGENTLOG_FILE = '.agentlog/errors.jsonl';

// Skip in production
const isProduction = process.env.NODE_ENV === 'production';

interface AgentlogEntry {
  timestamp: string;
  source: string;
  error_type: string;
  message: string;
  context?: Record<string, unknown>;
}

// Log an error to agentlog - call this directly or use with your logger (pino, winston, etc.)
export function logError(
  errorType: string,
  message: string,
  context?: Record<string, unknown>
): void {
  if (isProduction) return;

  const entry: AgentlogEntry = {
    timestamp: new Date().toISOString(),
    source: 'worker',
    error_type: errorType,
    message: String(message).slice(0, 500),
  };

  if (context) {
    // Truncate stack_trace if present
    if (typeof context.stack_trace === 'string') {
      context.stack_trace = context.stack_trace.slice(0, 2048);
    }
    entry.context = context;
  }

  try {
    if (!existsSync('.agentlog')) {
      mkdirSync('.agentlog', { recursive: true });

      // Update .gitignore
      const gitignorePath = '.gitignore';
      const gitignoreEntry = '.agentlog/errors.jsonl';
      let gitignoreContent = '';

      if (existsSync(gitignorePath)) {
        gitignoreContent = readFileSync(gitignorePath, 'utf-8');
      }

      if (!gitignoreContent.includes(gitignoreEntry)) {
        const newContent = gitignoreContent === ''
          ? gitignoreEntry + '\n'
          : gitignoreContent + (gitignoreContent.endsWith('\n') ? '' : '\n') + gitignoreEntry + '\n';
        writeFileSync(gitignorePath, newContent);
      }
    }
    appendFileSync(AGENTLOG_FILE, JSON.stringify(entry) + '\n');
  } catch {
    // Silently fail - don't crash the app for logging
  }
}

// Initialize agentlog: captures uncaught exceptions and unhandled rejections
export function initAgentlog(): void {
  if (isProduction) return;

  process.on('uncaughtException', (err: Error) => {
    logError('UNCAUGHT_EXCEPTION', err.message, {
      stack_trace: err.stack,
    });
    // Re-throw to let the process crash as expected
    throw err;
  });

  process.on('unhandledRejection', (reason: unknown) => {
    const message = reason instanceof Error ? reason.message : String(reason);
    const stack = reason instanceof Error ? reason.stack : undefined;
    logError('UNHANDLED_REJECTION', message, {
      stack_trace: stack,
    });
  });
}

// Pino integration example:
// import pino from 'pino';
// const logger = pino({
//   hooks: {
//     logMethod(args, method, level) {
//       if (level >= 50) { // error level
//         logError('LOG_ERROR', args[0]?.msg || String(args[0]));
//       }
//       method.apply(this, args);
//     }
//   }
// });

// Call at application startup
initAgentlog();`

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

const snippetRuby = `# === BROWSER (add to app/javascript/application.js) ===
// Error capture for agentlog - sends frontend errors to /__agentlog endpoint
(function() {
  const log = (type, msg, ctx) =>
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
})();

# === RAILS CONTROLLER (app/controllers/agentlog_controller.rb) ===
class AgentlogController < ApplicationController
  skip_before_action :verify_authenticity_token, only: :create

  def create
    return head :not_found unless Rails.env.development?

    FileUtils.mkdir_p('.agentlog')
    File.open('.agentlog/errors.jsonl', 'a') do |f|
      f.puts(request.raw_post)
    end

    head :ok
  end
end

# === ROUTE (add to config/routes.rb) ===
post '/__agentlog', to: 'agentlog#create' if Rails.env.development?

# === BACKEND MIDDLEWARE (add to config/initializers/agentlog.rb) ===
require 'json'
require 'fileutils'

module Agentlog
  class ExceptionCatcher
    def initialize(app)
      @app = app
    end

    def call(env)
      @app.call(env)
    rescue Exception => e
      log_error(e, env)
      raise
    end

    private

    def log_error(exception, env)
      entry = {
        timestamp: Time.now.utc.iso8601(3),
        source: 'backend',
        error_type: 'REQUEST_ERROR',
        message: exception.message.to_s[0, 500],
        context: {
          stack_trace: exception.backtrace&.join("\n")&.slice(0, 2048),
          endpoint: env['REQUEST_PATH'] || env['PATH_INFO'],
          request_id: env['action_dispatch.request_id']
        }.compact
      }

      FileUtils.mkdir_p('.agentlog')
      File.open('.agentlog/errors.jsonl', 'a') do |f|
        f.puts(entry.to_json)
      end
    end
  end
end

# Add to middleware stack (only in development)
if defined?(Rails) && Rails.env.development?
  Rails.application.config.middleware.insert(0, Agentlog::ExceptionCatcher)
end`

// Installable snippet parts for --install flag

const rubyController = `# agentlog:installed
class AgentlogController < ApplicationController
  skip_before_action :verify_authenticity_token, only: :create

  def create
    return head :not_found unless Rails.env.development?

    FileUtils.mkdir_p('.agentlog')
    File.open('.agentlog/errors.jsonl', 'a') do |f|
      f.puts(request.raw_post)
    end

    head :ok
  end
end
`

const rubyInitializer = `# agentlog:installed
require 'json'
require 'fileutils'

module Agentlog
  class ExceptionCatcher
    def initialize(app)
      @app = app
    end

    def call(env)
      @app.call(env)
    rescue Exception => e
      log_error(e, env)
      raise
    end

    private

    def log_error(exception, env)
      entry = {
        timestamp: Time.now.utc.iso8601(3),
        source: 'backend',
        error_type: 'REQUEST_ERROR',
        message: exception.message.to_s[0, 500],
        context: {
          stack_trace: exception.backtrace&.join("\n")&.slice(0, 2048),
          endpoint: env['REQUEST_PATH'] || env['PATH_INFO'],
          request_id: env['action_dispatch.request_id']
        }.compact
      }

      FileUtils.mkdir_p('.agentlog')
      File.open('.agentlog/errors.jsonl', 'a') do |f|
        f.puts(entry.to_json)
      end
    end
  end
end

# Add to middleware stack (only in development)
if defined?(Rails) && Rails.env.development?
  Rails.application.config.middleware.insert(0, Agentlog::ExceptionCatcher)
end
`

const rubyRoute = `post '/__agentlog', to: 'agentlog#create' if Rails.env.development?`

const rubyFrontendJS = `// agentlog:installed - Error capture for agentlog
(function() {
  const log = (type, msg, ctx) =>
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
})();
`

const typescriptCapture = `// agentlog:installed - Import this in your app entry point
// Usage: import './.agentlog/capture';

if (typeof window !== 'undefined') {
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
`

const nodeCapture = `// agentlog:installed - Import this in your Node.js app entry point
// Usage: import './.agentlog/capture';
// Works with BullMQ workers, scrapers, CLI tools, and any Node.js service

import { appendFileSync, mkdirSync, existsSync, readFileSync, writeFileSync } from 'fs';

const AGENTLOG_FILE = '.agentlog/errors.jsonl';

// Skip in production
const isProduction = process.env.NODE_ENV === 'production';

interface AgentlogEntry {
  timestamp: string;
  source: string;
  error_type: string;
  message: string;
  context?: Record<string, unknown>;
}

// Log an error to agentlog - call this directly or use with your logger (pino, winston, etc.)
export function logError(
  errorType: string,
  message: string,
  context?: Record<string, unknown>
): void {
  if (isProduction) return;

  const entry: AgentlogEntry = {
    timestamp: new Date().toISOString(),
    source: 'worker',
    error_type: errorType,
    message: String(message).slice(0, 500),
  };

  if (context) {
    // Truncate stack_trace if present
    if (typeof context.stack_trace === 'string') {
      context.stack_trace = context.stack_trace.slice(0, 2048);
    }
    entry.context = context;
  }

  try {
    if (!existsSync('.agentlog')) {
      mkdirSync('.agentlog', { recursive: true });

      // Update .gitignore
      const gitignorePath = '.gitignore';
      const gitignoreEntry = '.agentlog/errors.jsonl';
      let gitignoreContent = '';

      if (existsSync(gitignorePath)) {
        gitignoreContent = readFileSync(gitignorePath, 'utf-8');
      }

      if (!gitignoreContent.includes(gitignoreEntry)) {
        const newContent = gitignoreContent === ''
          ? gitignoreEntry + '\n'
          : gitignoreContent + (gitignoreContent.endsWith('\n') ? '' : '\n') + gitignoreEntry + '\n';
        writeFileSync(gitignorePath, newContent);
      }
    }
    appendFileSync(AGENTLOG_FILE, JSON.stringify(entry) + '\n');
  } catch {
    // Silently fail - don't crash the app for logging
  }
}

// Initialize agentlog: captures uncaught exceptions and unhandled rejections
export function initAgentlog(): void {
  if (isProduction) return;

  process.on('uncaughtException', (err: Error) => {
    logError('UNCAUGHT_EXCEPTION', err.message, {
      stack_trace: err.stack,
    });
    // Re-throw to let the process crash as expected
    throw err;
  });

  process.on('unhandledRejection', (reason: unknown) => {
    const message = reason instanceof Error ? reason.message : String(reason);
    const stack = reason instanceof Error ? reason.stack : undefined;
    logError('UNHANDLED_REJECTION', message, {
      stack_trace: stack,
    });
  });
}

// Pino integration example:
// import pino from 'pino';
// const logger = pino({
//   hooks: {
//     logMethod(args, method, level) {
//       if (level >= 50) { // error level
//         logError('LOG_ERROR', args[0]?.msg || String(args[0]));
//       }
//       method.apply(this, args);
//     }
//   }
// });

// Call at application startup
initAgentlog();
`
