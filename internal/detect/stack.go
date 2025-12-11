package detect

import (
	"os"
	"path/filepath"
)

// Stack represents a detected tech stack
type Stack string

const (
	TypeScript Stack = "typescript"
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
			return DetectionResult{
				Stack:      marker.stack,
				Detected:   true,
				MarkerFile: markerFile,
			}
		}
	}
	return DetectionResult{Detected: false}
}
