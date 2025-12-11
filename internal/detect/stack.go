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
var markerPriority = []struct {
	file  string
	stack Stack
}{
	{"package.json", TypeScript},
	{"go.mod", Go},
	{"pyproject.toml", Python},
	{"requirements.txt", Python},
	{"Cargo.toml", Rust},
	{"Gemfile", Ruby},
}

// DetectStack detects the project's tech stack based on marker files
func DetectStack(dir string) DetectionResult {
	for _, marker := range markerPriority {
		path := filepath.Join(dir, marker.file)
		if _, err := os.Stat(path); err == nil {
			return DetectionResult{
				Stack:      marker.stack,
				Detected:   true,
				MarkerFile: marker.file,
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
