package detect

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
)

// Stack represents a detected tech stack
type Stack string

const (
	TypeScript Stack = "typescript"
	Node       Stack = "node"
	Go         Stack = "go"
	Python     Stack = "python"
	Rust       Stack = "rust"
	Ruby       Stack = "ruby"
)

// String returns the string representation of the stack
func (s Stack) String() string {
	return string(s)
}

// DetectionResult contains the result of stack detection
type DetectionResult struct {
	Stack      Stack  // The detected or default stack
	Detected   bool   // Whether the stack was auto-detected
	MarkerFile string // The file that triggered detection (empty if not detected)
}

// markerPriority defines the order of marker file checks
// Order matters: config/routes.rb before package.json ensures Rails apps
// with npm dependencies are detected as Ruby, not TypeScript
var markerPriority = []struct {
	file  string
	stack Stack
}{
	{"config/routes.rb", Ruby}, // Rails-specific, takes priority over package.json
	{"package.json", TypeScript},
	{"go.mod", Go},
	{"pyproject.toml", Python},
	{"requirements.txt", Python},
	{"Cargo.toml", Rust},
	{"Gemfile", Ruby},
}

// monorepoSubdirs are common subdirectory patterns in monorepos
// Order matters: backend is checked before api, server
var monorepoSubdirs = []string{
	"backend",
	"api",
	"server",
}

// DetectStack detects the project's tech stack based on marker files
func DetectStack(dir string) DetectionResult {
	// First, check root level
	if result := detectInDir(dir, ""); result.Detected {
		return result
	}

	// Then, check common monorepo subdirectories
	for _, subdir := range monorepoSubdirs {
		subdirPath := filepath.Join(dir, subdir)
		if info, err := os.Stat(subdirPath); err == nil && info.IsDir() {
			if result := detectInDir(subdirPath, subdir); result.Detected {
				return result
			}
		}
	}

	// Default to TypeScript if no marker found
	return DetectionResult{
		Stack:      TypeScript,
		Detected:   false,
		MarkerFile: "",
	}
}

// detectInDir checks for marker files in a specific directory
// prefix is prepended to MarkerFile (e.g., "backend" -> "backend/go.mod")
func detectInDir(dir, prefix string) DetectionResult {
	for _, marker := range markerPriority {
		path := filepath.Join(dir, marker.file)
		if _, err := os.Stat(path); err == nil {
			markerFile := marker.file
			if prefix != "" {
				markerFile = filepath.Join(prefix, marker.file)
			}

			// For TypeScript (package.json), apply Node.js vs browser heuristics
			stack := marker.stack
			if marker.stack == TypeScript {
				stack = detectTypeScriptVariant(dir)
			}

			return DetectionResult{
				Stack:      stack,
				Detected:   true,
				MarkerFile: markerFile,
			}
		}
	}
	return DetectionResult{Detected: false}
}

// browserFrameworks are frontend framework dependencies that indicate a browser project
var browserFrameworks = []string{
	"react",
	"vue",
	"svelte",
	"@angular/core",
	"next",
	"nuxt",
	"solid-js",
	"preact",
	"@remix-run/react",
	"gatsby",
}

// nodeFrameworks are backend framework dependencies that indicate a Node.js project
var nodeFrameworks = []string{
	"express",
	"fastify",
	"hono",
	"koa",
	"nestjs",
	"@nestjs/core",
	"bullmq",
	"bull",
	"bee-queue",
	"agenda",
}

// nodeModuleSettings are tsconfig module values that indicate Node.js
var nodeModuleSettings = []string{
	"commonjs",
	"nodenext",
	"node16",
}

// detectTypeScriptVariant determines if a TypeScript project is Node.js or browser
// Returns TypeScript for browser projects, Node for server-side Node.js projects
func detectTypeScriptVariant(dir string) Stack {
	// Priority 1: Check for explicit browser indicators (files)
	browserFiles := []string{
		"vite.config.ts",
		"vite.config.js",
		"vite.config.mts",
		"vite.config.mjs",
		"src/App.tsx",
		"src/App.jsx",
		"next.config.js",
		"next.config.mjs",
		"nuxt.config.ts",
		"nuxt.config.js",
	}

	for _, file := range browserFiles {
		path := filepath.Join(dir, file)
		if _, err := os.Stat(path); err == nil {
			return TypeScript
		}
	}

	// Priority 2: Check package.json for dependencies
	packageJSONPath := filepath.Join(dir, "package.json")
	packageJSON, err := os.ReadFile(packageJSONPath)
	if err == nil {
		var pkg struct {
			Dependencies    map[string]string `json:"dependencies"`
			DevDependencies map[string]string `json:"devDependencies"`
			Scripts         map[string]string `json:"scripts"`
		}
		if err := json.Unmarshal(packageJSON, &pkg); err == nil {
			// Check for browser framework dependencies
			allDeps := make(map[string]bool)
			for dep := range pkg.Dependencies {
				allDeps[dep] = true
			}
			for dep := range pkg.DevDependencies {
				allDeps[dep] = true
			}

			for _, framework := range browserFrameworks {
				if allDeps[framework] {
					return TypeScript
				}
			}

			// Check for Node.js framework dependencies
			for _, framework := range nodeFrameworks {
				if allDeps[framework] {
					return Node
				}
			}

			// Check scripts for Node.js indicators
			for _, script := range pkg.Scripts {
				scriptLower := strings.ToLower(script)
				// Check for ts-node, tsx, or node commands
				if strings.Contains(scriptLower, "ts-node") ||
					strings.Contains(scriptLower, "tsx ") ||
					strings.HasPrefix(scriptLower, "tsx ") ||
					strings.Contains(scriptLower, " tsx") ||
					strings.HasPrefix(scriptLower, "node ") ||
					strings.Contains(scriptLower, " node ") {
					return Node
				}
			}
		}
	}

	// Priority 3: Check tsconfig.json for module settings
	tsconfigPath := filepath.Join(dir, "tsconfig.json")
	tsconfigJSON, err := os.ReadFile(tsconfigPath)
	if err == nil {
		var tsconfig struct {
			CompilerOptions struct {
				Module string `json:"module"`
			} `json:"compilerOptions"`
		}
		if err := json.Unmarshal(tsconfigJSON, &tsconfig); err == nil {
			moduleLower := strings.ToLower(tsconfig.CompilerOptions.Module)
			for _, nodeModule := range nodeModuleSettings {
				if moduleLower == nodeModule {
					return Node
				}
			}
		}
	}

	// Default: TypeScript (browser) - safer default for typical web projects
	return TypeScript
}
