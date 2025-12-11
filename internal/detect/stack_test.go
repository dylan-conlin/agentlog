package detect

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDetectStack(t *testing.T) {
	tests := []struct {
		name           string
		files          []string
		expectedStack  Stack
		expectedDetect bool
	}{
		{
			name:           "package.json detected as TypeScript",
			files:          []string{"package.json"},
			expectedStack:  TypeScript,
			expectedDetect: true,
		},
		{
			name:           "go.mod detected as Go",
			files:          []string{"go.mod"},
			expectedStack:  Go,
			expectedDetect: true,
		},
		{
			name:           "pyproject.toml detected as Python",
			files:          []string{"pyproject.toml"},
			expectedStack:  Python,
			expectedDetect: true,
		},
		{
			name:           "requirements.txt detected as Python",
			files:          []string{"requirements.txt"},
			expectedStack:  Python,
			expectedDetect: true,
		},
		{
			name:           "Cargo.toml detected as Rust",
			files:          []string{"Cargo.toml"},
			expectedStack:  Rust,
			expectedDetect: true,
		},
		{
			name:           "Gemfile detected as Ruby",
			files:          []string{"Gemfile"},
			expectedStack:  Ruby,
			expectedDetect: true,
		},
		{
			name:           "config/routes.rb detected as Ruby (Rails-specific)",
			files:          []string{"config/routes.rb"},
			expectedStack:  Ruby,
			expectedDetect: true,
		},
		{
			name:           "config/routes.rb takes priority over package.json (Rails with npm)",
			files:          []string{"package.json", "config/routes.rb"},
			expectedStack:  Ruby,
			expectedDetect: true,
		},
		{
			name:           "package.json takes priority over go.mod",
			files:          []string{"package.json", "go.mod"},
			expectedStack:  TypeScript,
			expectedDetect: true,
		},
		{
			name:           "pyproject.toml takes priority over requirements.txt",
			files:          []string{"pyproject.toml", "requirements.txt"},
			expectedStack:  Python,
			expectedDetect: true,
		},
		{
			name:           "no marker files defaults to TypeScript",
			files:          []string{},
			expectedStack:  TypeScript,
			expectedDetect: false,
		},
		{
			name:           "unknown file does not trigger detection",
			files:          []string{"random.txt"},
			expectedStack:  TypeScript,
			expectedDetect: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Create temp directory
			tmpDir := t.TempDir()

			// Create marker files (including parent directories for nested paths)
			for _, file := range tc.files {
				path := filepath.Join(tmpDir, file)
				// Create parent directory if needed (for paths like config/routes.rb)
				if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
					t.Fatalf("failed to create parent directory for %s: %v", file, err)
				}
				if err := os.WriteFile(path, []byte(""), 0644); err != nil {
					t.Fatalf("failed to create test file %s: %v", file, err)
				}
			}

			// Test detection
			result := DetectStack(tmpDir)

			if result.Stack != tc.expectedStack {
				t.Errorf("expected stack %s, got %s", tc.expectedStack, result.Stack)
			}

			if result.Detected != tc.expectedDetect {
				t.Errorf("expected detected=%v, got %v", tc.expectedDetect, result.Detected)
			}
		})
	}
}

func TestStackString(t *testing.T) {
	tests := []struct {
		stack    Stack
		expected string
	}{
		{TypeScript, "typescript"},
		{Go, "go"},
		{Python, "python"},
		{Rust, "rust"},
		{Ruby, "ruby"},
	}

	for _, tc := range tests {
		t.Run(tc.expected, func(t *testing.T) {
			if tc.stack.String() != tc.expected {
				t.Errorf("expected %s, got %s", tc.expected, tc.stack.String())
			}
		})
	}
}

func TestDetectStackMonorepoSubdirectories(t *testing.T) {
	tests := []struct {
		name           string
		files          []string
		expectedStack  Stack
		expectedDetect bool
		expectedMarker string
	}{
		{
			name:           "backend/ with go.mod detected as Go",
			files:          []string{"backend/go.mod"},
			expectedStack:  Go,
			expectedDetect: true,
			expectedMarker: "backend/go.mod",
		},
		{
			name:           "api/ with go.mod detected as Go",
			files:          []string{"api/go.mod"},
			expectedStack:  Go,
			expectedDetect: true,
			expectedMarker: "api/go.mod",
		},
		{
			name:           "server/ with requirements.txt detected as Python",
			files:          []string{"server/requirements.txt"},
			expectedStack:  Python,
			expectedDetect: true,
			expectedMarker: "server/requirements.txt",
		},
		{
			name:           "backend/ with config/routes.rb detected as Ruby (Rails)",
			files:          []string{"backend/config/routes.rb"},
			expectedStack:  Ruby,
			expectedDetect: true,
			expectedMarker: "backend/config/routes.rb",
		},
		{
			name:           "root level takes priority over subdirectory",
			files:          []string{"package.json", "backend/go.mod"},
			expectedStack:  TypeScript,
			expectedDetect: true,
			expectedMarker: "package.json",
		},
		{
			name:           "first subdirectory match wins (backend before api)",
			files:          []string{"backend/go.mod", "api/requirements.txt"},
			expectedStack:  Go,
			expectedDetect: true,
			expectedMarker: "backend/go.mod",
		},
		{
			name:           "no detection when subdirectory has no marker files",
			files:          []string{"backend/random.txt"},
			expectedStack:  TypeScript,
			expectedDetect: false,
			expectedMarker: "",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Create temp directory
			tmpDir := t.TempDir()

			// Create marker files (including parent directories for nested paths)
			for _, file := range tc.files {
				path := filepath.Join(tmpDir, file)
				// Create parent directory if needed
				if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
					t.Fatalf("failed to create parent directory for %s: %v", file, err)
				}
				if err := os.WriteFile(path, []byte(""), 0644); err != nil {
					t.Fatalf("failed to create test file %s: %v", file, err)
				}
			}

			// Test detection
			result := DetectStack(tmpDir)

			if result.Stack != tc.expectedStack {
				t.Errorf("expected stack %s, got %s", tc.expectedStack, result.Stack)
			}

			if result.Detected != tc.expectedDetect {
				t.Errorf("expected detected=%v, got %v", tc.expectedDetect, result.Detected)
			}

			if result.MarkerFile != tc.expectedMarker {
				t.Errorf("expected marker file %q, got %q", tc.expectedMarker, result.MarkerFile)
			}
		})
	}
}

func TestStackMarkerFile(t *testing.T) {
	tests := []struct {
		stack    Stack
		expected string
	}{
		{TypeScript, "package.json"},
		{Go, "go.mod"},
		{Python, "pyproject.toml"},
		{Rust, "Cargo.toml"},
		{Ruby, "Gemfile"},
	}

	for _, tc := range tests {
		t.Run(tc.stack.String(), func(t *testing.T) {
			result := DetectStack(t.TempDir())
			// Create a file to trigger detection
			tmpDir := t.TempDir()
			os.WriteFile(filepath.Join(tmpDir, tc.expected), []byte(""), 0644)
			result = DetectStack(tmpDir)

			if result.MarkerFile != tc.expected {
				t.Errorf("expected marker file %s, got %s", tc.expected, result.MarkerFile)
			}
		})
	}
}
