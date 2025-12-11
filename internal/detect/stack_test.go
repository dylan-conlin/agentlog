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
		{Node, "node"},
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

func TestDetectNodeVsBrowserTypeScript(t *testing.T) {
	tests := []struct {
		name           string
		files          map[string]string // filename -> content
		expectedStack  Stack
		expectedDetect bool
	}{
		// Browser indicators
		{
			name: "vite.config.ts indicates browser TypeScript",
			files: map[string]string{
				"package.json":   `{"name": "app"}`,
				"vite.config.ts": `export default {}`,
			},
			expectedStack:  TypeScript,
			expectedDetect: true,
		},
		{
			name: "vite.config.js indicates browser TypeScript",
			files: map[string]string{
				"package.json":   `{"name": "app"}`,
				"vite.config.js": `export default {}`,
			},
			expectedStack:  TypeScript,
			expectedDetect: true,
		},
		{
			name: "src/App.tsx indicates browser TypeScript",
			files: map[string]string{
				"package.json": `{"name": "app"}`,
				"src/App.tsx":  `export default function App() {}`,
			},
			expectedStack:  TypeScript,
			expectedDetect: true,
		},
		{
			name: "react dependency indicates browser TypeScript",
			files: map[string]string{
				"package.json": `{"name": "app", "dependencies": {"react": "^18.0.0"}}`,
			},
			expectedStack:  TypeScript,
			expectedDetect: true,
		},
		{
			name: "vue dependency indicates browser TypeScript",
			files: map[string]string{
				"package.json": `{"name": "app", "dependencies": {"vue": "^3.0.0"}}`,
			},
			expectedStack:  TypeScript,
			expectedDetect: true,
		},
		{
			name: "svelte dependency indicates browser TypeScript",
			files: map[string]string{
				"package.json": `{"name": "app", "devDependencies": {"svelte": "^4.0.0"}}`,
			},
			expectedStack:  TypeScript,
			expectedDetect: true,
		},
		{
			name: "@angular/core dependency indicates browser TypeScript",
			files: map[string]string{
				"package.json": `{"name": "app", "dependencies": {"@angular/core": "^17.0.0"}}`,
			},
			expectedStack:  TypeScript,
			expectedDetect: true,
		},
		{
			name: "next dependency indicates browser TypeScript",
			files: map[string]string{
				"package.json": `{"name": "app", "dependencies": {"next": "^14.0.0"}}`,
			},
			expectedStack:  TypeScript,
			expectedDetect: true,
		},
		// Node.js indicators
		{
			name: "ts-node in scripts indicates Node.js",
			files: map[string]string{
				"package.json": `{"name": "worker", "scripts": {"start": "ts-node src/index.ts"}}`,
			},
			expectedStack:  Node,
			expectedDetect: true,
		},
		{
			name: "tsx in scripts indicates Node.js",
			files: map[string]string{
				"package.json": `{"name": "worker", "scripts": {"dev": "tsx watch src/index.ts"}}`,
			},
			expectedStack:  Node,
			expectedDetect: true,
		},
		{
			name: "node in scripts indicates Node.js",
			files: map[string]string{
				"package.json": `{"name": "api", "scripts": {"start": "node dist/index.js"}}`,
			},
			expectedStack:  Node,
			expectedDetect: true,
		},
		{
			name: "tsconfig with module commonjs indicates Node.js",
			files: map[string]string{
				"package.json":  `{"name": "service"}`,
				"tsconfig.json": `{"compilerOptions": {"module": "commonjs"}}`,
			},
			expectedStack:  Node,
			expectedDetect: true,
		},
		{
			name: "tsconfig with module CommonJS (case insensitive) indicates Node.js",
			files: map[string]string{
				"package.json":  `{"name": "service"}`,
				"tsconfig.json": `{"compilerOptions": {"module": "CommonJS"}}`,
			},
			expectedStack:  Node,
			expectedDetect: true,
		},
		{
			name: "tsconfig with module nodenext indicates Node.js",
			files: map[string]string{
				"package.json":  `{"name": "service"}`,
				"tsconfig.json": `{"compilerOptions": {"module": "nodenext"}}`,
			},
			expectedStack:  Node,
			expectedDetect: true,
		},
		{
			name: "tsconfig with module node16 indicates Node.js",
			files: map[string]string{
				"package.json":  `{"name": "service"}`,
				"tsconfig.json": `{"compilerOptions": {"module": "node16"}}`,
			},
			expectedStack:  Node,
			expectedDetect: true,
		},
		{
			name: "bullmq dependency indicates Node.js",
			files: map[string]string{
				"package.json": `{"name": "worker", "dependencies": {"bullmq": "^5.0.0"}}`,
			},
			expectedStack:  Node,
			expectedDetect: true,
		},
		{
			name: "express dependency indicates Node.js",
			files: map[string]string{
				"package.json": `{"name": "api", "dependencies": {"express": "^4.0.0"}}`,
			},
			expectedStack:  Node,
			expectedDetect: true,
		},
		{
			name: "fastify dependency indicates Node.js",
			files: map[string]string{
				"package.json": `{"name": "api", "dependencies": {"fastify": "^4.0.0"}}`,
			},
			expectedStack:  Node,
			expectedDetect: true,
		},
		{
			name: "hono dependency indicates Node.js",
			files: map[string]string{
				"package.json": `{"name": "api", "dependencies": {"hono": "^4.0.0"}}`,
			},
			expectedStack:  Node,
			expectedDetect: true,
		},
		// Mixed - browser indicators take priority
		{
			name: "express with vite.config uses browser TypeScript",
			files: map[string]string{
				"package.json":   `{"name": "fullstack", "dependencies": {"express": "^4.0.0"}}`,
				"vite.config.ts": `export default {}`,
			},
			expectedStack:  TypeScript,
			expectedDetect: true,
		},
		{
			name: "express with react uses browser TypeScript",
			files: map[string]string{
				"package.json": `{"name": "fullstack", "dependencies": {"express": "^4.0.0", "react": "^18.0.0"}}`,
			},
			expectedStack:  TypeScript,
			expectedDetect: true,
		},
		// Default case - no indicators defaults to browser TypeScript (safer default)
		{
			name: "package.json only defaults to browser TypeScript",
			files: map[string]string{
				"package.json": `{"name": "unknown-app"}`,
			},
			expectedStack:  TypeScript,
			expectedDetect: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Create temp directory
			tmpDir := t.TempDir()

			// Create files with content
			for filename, content := range tc.files {
				path := filepath.Join(tmpDir, filename)
				// Create parent directory if needed
				if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
					t.Fatalf("failed to create parent directory for %s: %v", filename, err)
				}
				if err := os.WriteFile(path, []byte(content), 0644); err != nil {
					t.Fatalf("failed to create test file %s: %v", filename, err)
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
