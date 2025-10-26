package checker

import (
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strings"
	"testing"

	"github.com/DuckDHD/EnvQuack/internal/parser"
)

func TestCompareComposeWithEnv(t *testing.T) {
	tests := []struct {
		name           string
		composeContent string
		envFiles       map[string]string // filename -> content
		expected       *ComposeDiffResult
		expectError    bool
	}{
		{
			name: "all variables present in env",
			composeContent: `version: '3.8'
services:
  web:
    environment:
      - NODE_ENV=production
      - DATABASE_URL=${DATABASE_URL}
      - PORT=${PORT:-3000}`,
			envFiles: map[string]string{
				".env": `DATABASE_URL=postgres://localhost/myapp
PORT=8080
NODE_ENV=production`,
			},
			expected: &ComposeDiffResult{
				MissingInEnv:     []string{},
				ExtraInEnv:       []string{},
				MissingEnvFiles:  []string{},
				ServiceBreakdown: map[string][]string{},
			},
			expectError: false,
		},
		{
			name: "missing variables in env",
			composeContent: `version: '3.8'
services:
  web:
    environment:
      - NODE_ENV=${NODE_ENV}
      - DATABASE_URL=${DATABASE_URL}
      - API_KEY=${API_KEY}
      - REDIS_URL=${REDIS_URL}`,
			envFiles: map[string]string{
				".env": `NODE_ENV=production
DATABASE_URL=postgres://localhost/myapp`,
			},
			expected: &ComposeDiffResult{
				MissingInEnv:    []string{"API_KEY", "REDIS_URL"},
				ExtraInEnv:      []string{},
				MissingEnvFiles: []string{},
				ServiceBreakdown: map[string][]string{
					"web": {"API_KEY", "REDIS_URL"},
				},
			},
			expectError: false,
		},
		{
			name: "extra variables in env",
			composeContent: `version: '3.8'
services:
  web:
    environment:
      - NODE_ENV=${NODE_ENV}
      - PORT=${PORT}`,
			envFiles: map[string]string{
				".env": `NODE_ENV=production
PORT=3000
DATABASE_URL=postgres://localhost/myapp
API_KEY=secret123
DEBUG=true`,
			},
			expected: &ComposeDiffResult{
				MissingInEnv:     []string{},
				ExtraInEnv:       []string{"API_KEY", "DATABASE_URL", "DEBUG"},
				MissingEnvFiles:  []string{},
				ServiceBreakdown: map[string][]string{},
			},
			expectError: false,
		},
		{
			name: "multiple services with different missing variables",
			composeContent: `version: '3.8'
services:
  web:
    environment:
      - NODE_ENV=${NODE_ENV}
      - DATABASE_URL=${DATABASE_URL}
      - WEB_PORT=${WEB_PORT}
  worker:
    environment:
      NODE_ENV: ${NODE_ENV}
      DATABASE_URL: ${DATABASE_URL}
      QUEUE_NAME: ${QUEUE_NAME}
      WORKER_CONCURRENCY: ${WORKER_CONCURRENCY}
  db:
    environment:
      POSTGRES_DB: ${DB_NAME}
      POSTGRES_USER: ${DB_USER}
      POSTGRES_PASSWORD: ${DB_PASSWORD}`,
			envFiles: map[string]string{
				".env": `NODE_ENV=production
DATABASE_URL=postgres://localhost/myapp
WEB_PORT=3000
DB_NAME=myapp
DB_USER=postgres
DB_PASSWORD=supersecret`,
			},
			expected: &ComposeDiffResult{
				MissingInEnv:    []string{"QUEUE_NAME", "WORKER_CONCURRENCY"},
				ExtraInEnv:      []string{},
				MissingEnvFiles: []string{},
				ServiceBreakdown: map[string][]string{
					"worker": {"QUEUE_NAME", "WORKER_CONCURRENCY"},
				},
			},
			expectError: false,
		},
		{
			name: "simple env_file test",
			composeContent: `version: '3.8'
services:
  web:
    environment:
      - NODE_ENV=${NODE_ENV}
      - PORT=${PORT}`,
			envFiles: map[string]string{
				".env": `NODE_ENV=production
PORT=3000
DATABASE_URL=postgres://localhost/myapp
DEBUG=true`,
			},
			expected: &ComposeDiffResult{
				MissingInEnv:     []string{},
				ExtraInEnv:       []string{"DATABASE_URL", "DEBUG"},
				MissingEnvFiles:  []string{},
				ServiceBreakdown: map[string][]string{},
			},
			expectError: false,
		},
		{
			name: "variables with default values",
			composeContent: `version: '3.8'
services:
  web:
    environment:
      - NODE_ENV=${NODE_ENV:-production}
      - PORT=${PORT:-3000}
      - DEBUG=${DEBUG:-false}
    ports:
      - "${WEB_PORT:-8080}:${PORT:-3000}"`,
			envFiles: map[string]string{
				".env": `NODE_ENV=development
WEB_PORT=9090`,
			},
			expected: &ComposeDiffResult{
				MissingInEnv:    []string{"DEBUG", "PORT"},
				ExtraInEnv:      []string{},
				MissingEnvFiles: []string{},
				ServiceBreakdown: map[string][]string{
					"web": {"DEBUG", "PORT"},
				},
			},
			expectError: false,
		},
		{
			name: "mixed variable reference formats",
			composeContent: `version: '3.8'
services:
  app:
    image: myapp:$VERSION
    environment:
      - DATABASE_URL=${DATABASE_URL}
      - API_KEY=$API_KEY
      - PORT=${PORT:-3000}
    volumes:
      - "${DATA_DIR:-./data}:/app/data"`,
			envFiles: map[string]string{
				".env": `VERSION=1.0.0
DATABASE_URL=postgres://localhost/myapp
DATA_DIR=/opt/data`,
			},
			expected: &ComposeDiffResult{
				MissingInEnv:    []string{"API_KEY", "PORT"},
				ExtraInEnv:      []string{},
				MissingEnvFiles: []string{},
				ServiceBreakdown: map[string][]string{
					"app": {"API_KEY", "PORT"},
				},
			},
			expectError: false,
		},
		{
			name: "empty compose file",
			composeContent: `version: '3.8'
services: {}`,
			envFiles: map[string]string{
				".env": `NODE_ENV=production
DATABASE_URL=postgres://localhost/myapp`,
			},
			expected: &ComposeDiffResult{
				MissingInEnv:     []string{},
				ExtraInEnv:       []string{"DATABASE_URL", "NODE_ENV"},
				MissingEnvFiles:  []string{},
				ServiceBreakdown: map[string][]string{},
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temporary directory and files
			tmpDir := t.TempDir()
			composeFile := filepath.Join(tmpDir, "docker-compose.yml")

			// Write compose file
			err := os.WriteFile(composeFile, []byte(tt.composeContent), 0644)
			if err != nil {
				t.Fatalf("Failed to create compose file: %v", err)
			}

			// Write env files
			var envFilePaths []string
			for filename, content := range tt.envFiles {
				envPath := filepath.Join(tmpDir, filename)
				err := os.WriteFile(envPath, []byte(content), 0644)
				if err != nil {
					t.Fatalf("Failed to create env file %s: %v", filename, err)
				}
				envFilePaths = append(envFilePaths, envPath)
			}

			// Run the comparison
			result, err := CompareComposeWithEnv(composeFile, envFilePaths)

			// Check error expectation
			if tt.expectError && err == nil {
				t.Errorf("Expected error but got none")
				return
			}
			if !tt.expectError && err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if tt.expectError {
				return // Don't check results if we expected an error
			}

			// Compare results
			if !compareComposeDiffResults(result, tt.expected) {
				t.Errorf("CompareComposeWithEnv() result mismatch")
				t.Errorf("Got MissingInEnv: %v", result.MissingInEnv)
				t.Errorf("Want MissingInEnv: %v", tt.expected.MissingInEnv)
				t.Errorf("Got ExtraInEnv: %v", result.ExtraInEnv)
				t.Errorf("Want ExtraInEnv: %v", tt.expected.ExtraInEnv)
				t.Errorf("Got MissingEnvFiles: %v", result.MissingEnvFiles)
				t.Errorf("Want MissingEnvFiles: %v", tt.expected.MissingEnvFiles)
				t.Errorf("Got ServiceBreakdown: %v", result.ServiceBreakdown)
				t.Errorf("Want ServiceBreakdown: %v", tt.expected.ServiceBreakdown)
			}
		})
	}
}

func TestCompareComposeWithEnv_FileErrors(t *testing.T) {
	tests := []struct {
		name         string
		setupCompose func(string) error
		setupEnv     func(string) error
		expectError  bool
	}{
		{
			name: "missing compose file",
			setupCompose: func(filename string) error {
				// Don't create the file
				return nil
			},
			setupEnv: func(filename string) error {
				return os.WriteFile(filename, []byte("NODE_ENV=production"), 0644)
			},
			expectError: true,
		},
		{
			name: "compose file is directory",
			setupCompose: func(filename string) error {
				return os.Mkdir(filename, 0755)
			},
			setupEnv: func(filename string) error {
				return os.WriteFile(filename, []byte("NODE_ENV=production"), 0644)
			},
			expectError: true,
		},
		{
			name: "invalid compose YAML",
			setupCompose: func(filename string) error {
				return os.WriteFile(filename, []byte("invalid: yaml: content: [unclosed"), 0644)
			},
			setupEnv: func(filename string) error {
				return os.WriteFile(filename, []byte("NODE_ENV=production"), 0644)
			},
			expectError: true,
		},
		{
			name: "missing env files are handled gracefully",
			setupCompose: func(filename string) error {
				content := `version: '3.8'
services:
  web:
    environment:
      - NODE_ENV=${NODE_ENV}`
				return os.WriteFile(filename, []byte(content), 0644)
			},
			setupEnv: func(filename string) error {
				// Don't create env file - should be handled gracefully
				return nil
			},
			expectError: false, // Missing env files should not cause error in compare function
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			composeFile := filepath.Join(tmpDir, "docker-compose.yml")
			envFile := filepath.Join(tmpDir, ".env")

			// Setup files
			if err := tt.setupCompose(composeFile); err != nil {
				t.Fatalf("Failed to setup compose file: %v", err)
			}
			if err := tt.setupEnv(envFile); err != nil {
				t.Fatalf("Failed to setup env file: %v", err)
			}

			// Clean up after test
			defer func() {
				os.RemoveAll(composeFile)
				os.RemoveAll(envFile)
			}()

			// Run comparison
			result, err := CompareComposeWithEnv(composeFile, []string{envFile})

			if tt.expectError && err == nil {
				t.Errorf("Expected error but got none. Result: %v", result)
			}
			if !tt.expectError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
		})
	}
}

func TestCompareComposeWithEnvVars(t *testing.T) {
	tests := []struct {
		name        string
		composeInfo *parser.ComposeEnvInfo
		envVars     parser.EnvVars
		expected    *ComposeDiffResult
	}{
		{
			name: "perfect match",
			composeInfo: &parser.ComposeEnvInfo{
				Variables: parser.EnvVars{
					"NODE_ENV":     "production",
					"DATABASE_URL": "${DATABASE_URL}",
				},
				ServiceVars: map[string]parser.EnvVars{
					"web": {
						"NODE_ENV":     "production",
						"DATABASE_URL": "${DATABASE_URL}",
					},
				},
				VariableRefs: []string{"DATABASE_URL"},
				EnvFiles:     []string{},
			},
			envVars: parser.EnvVars{
				"NODE_ENV":     "production",
				"DATABASE_URL": "postgres://localhost/myapp",
			},
			expected: &ComposeDiffResult{
				MissingInEnv:     []string{},
				ExtraInEnv:       []string{},
				MissingEnvFiles:  []string{},
				ServiceBreakdown: map[string][]string{},
			},
		},
		{
			name: "missing variables with service breakdown",
			composeInfo: &parser.ComposeEnvInfo{
				Variables: parser.EnvVars{
					"NODE_ENV":     "${NODE_ENV}",
					"DATABASE_URL": "${DATABASE_URL}",
					"API_KEY":      "${API_KEY}",
					"REDIS_URL":    "${REDIS_URL}",
				},
				ServiceVars: map[string]parser.EnvVars{
					"web": {
						"NODE_ENV":     "${NODE_ENV}",
						"DATABASE_URL": "${DATABASE_URL}",
						"API_KEY":      "${API_KEY}",
					},
					"worker": {
						"NODE_ENV":     "${NODE_ENV}",
						"DATABASE_URL": "${DATABASE_URL}",
						"REDIS_URL":    "${REDIS_URL}",
					},
				},
				VariableRefs: []string{"NODE_ENV", "DATABASE_URL", "API_KEY", "REDIS_URL"},
				EnvFiles:     []string{},
			},
			envVars: parser.EnvVars{
				"NODE_ENV":     "production",
				"DATABASE_URL": "postgres://localhost/myapp",
			},
			expected: &ComposeDiffResult{
				MissingInEnv:    []string{"API_KEY", "REDIS_URL"},
				ExtraInEnv:      []string{},
				MissingEnvFiles: []string{},
				ServiceBreakdown: map[string][]string{
					"web":    {"API_KEY"},
					"worker": {"REDIS_URL"},
				},
			},
		},
		{
			name: "extra variables in env",
			composeInfo: &parser.ComposeEnvInfo{
				Variables:    parser.EnvVars{"NODE_ENV": "${NODE_ENV}"},
				ServiceVars:  map[string]parser.EnvVars{"web": {"NODE_ENV": "${NODE_ENV}"}},
				VariableRefs: []string{"NODE_ENV"},
				EnvFiles:     []string{},
			},
			envVars: parser.EnvVars{
				"NODE_ENV":     "production",
				"DATABASE_URL": "postgres://localhost/myapp",
				"API_KEY":      "secret123",
				"DEBUG":        "true",
			},
			expected: &ComposeDiffResult{
				MissingInEnv:     []string{},
				ExtraInEnv:       []string{"API_KEY", "DATABASE_URL", "DEBUG"},
				MissingEnvFiles:  []string{},
				ServiceBreakdown: map[string][]string{},
			},
		},
		{
			name: "empty compose info",
			composeInfo: &parser.ComposeEnvInfo{
				Variables:    parser.EnvVars{},
				ServiceVars:  map[string]parser.EnvVars{},
				VariableRefs: []string{},
				EnvFiles:     []string{},
			},
			envVars: parser.EnvVars{
				"NODE_ENV":     "production",
				"DATABASE_URL": "postgres://localhost/myapp",
			},
			expected: &ComposeDiffResult{
				MissingInEnv:     []string{},
				ExtraInEnv:       []string{"DATABASE_URL", "NODE_ENV"},
				MissingEnvFiles:  []string{},
				ServiceBreakdown: map[string][]string{},
			},
		},
		{
			name: "empty env vars",
			composeInfo: &parser.ComposeEnvInfo{
				Variables: parser.EnvVars{
					"NODE_ENV":     "${NODE_ENV}",
					"DATABASE_URL": "${DATABASE_URL}",
				},
				ServiceVars: map[string]parser.EnvVars{
					"web": {
						"NODE_ENV":     "${NODE_ENV}",
						"DATABASE_URL": "${DATABASE_URL}",
					},
				},
				VariableRefs: []string{"NODE_ENV", "DATABASE_URL"},
				EnvFiles:     []string{},
			},
			envVars: parser.EnvVars{},
			expected: &ComposeDiffResult{
				MissingInEnv:    []string{"DATABASE_URL", "NODE_ENV"},
				ExtraInEnv:      []string{},
				MissingEnvFiles: []string{},
				ServiceBreakdown: map[string][]string{
					"web": {"DATABASE_URL", "NODE_ENV"},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := compareComposeWithEnvVars(tt.composeInfo, tt.envVars)

			if !compareComposeDiffResults(result, tt.expected) {
				t.Errorf("compareComposeWithEnvVars() result mismatch")
				t.Errorf("Got MissingInEnv: %v", result.MissingInEnv)
				t.Errorf("Want MissingInEnv: %v", tt.expected.MissingInEnv)
				t.Errorf("Got ExtraInEnv: %v", result.ExtraInEnv)
				t.Errorf("Want ExtraInEnv: %v", tt.expected.ExtraInEnv)
				t.Errorf("Got MissingEnvFiles: %v", result.MissingEnvFiles)
				t.Errorf("Want MissingEnvFiles: %v", tt.expected.MissingEnvFiles)
				t.Errorf("Got ServiceBreakdown: %v", result.ServiceBreakdown)
				t.Errorf("Want ServiceBreakdown: %v", tt.expected.ServiceBreakdown)
			}
		})
	}
}

func TestComposeDiffResult_HasIssues(t *testing.T) {
	tests := []struct {
		name     string
		result   *ComposeDiffResult
		expected bool
	}{
		{
			name: "no issues",
			result: &ComposeDiffResult{
				MissingInEnv:     []string{},
				ExtraInEnv:       []string{},
				MissingEnvFiles:  []string{},
				ServiceBreakdown: map[string][]string{},
			},
			expected: false,
		},
		{
			name: "missing variables",
			result: &ComposeDiffResult{
				MissingInEnv:     []string{"API_KEY", "DATABASE_URL"},
				ExtraInEnv:       []string{},
				MissingEnvFiles:  []string{},
				ServiceBreakdown: map[string][]string{},
			},
			expected: true,
		},
		{
			name: "extra variables",
			result: &ComposeDiffResult{
				MissingInEnv:     []string{},
				ExtraInEnv:       []string{"DEBUG", "LOG_LEVEL"},
				MissingEnvFiles:  []string{},
				ServiceBreakdown: map[string][]string{},
			},
			expected: true,
		},
		{
			name: "missing env files",
			result: &ComposeDiffResult{
				MissingInEnv:     []string{},
				ExtraInEnv:       []string{},
				MissingEnvFiles:  []string{".env.local", ".env.production"},
				ServiceBreakdown: map[string][]string{},
			},
			expected: true,
		},
		{
			name: "all types of issues",
			result: &ComposeDiffResult{
				MissingInEnv:    []string{"API_KEY"},
				ExtraInEnv:      []string{"DEBUG"},
				MissingEnvFiles: []string{".env.local"},
				ServiceBreakdown: map[string][]string{
					"web": {"API_KEY"},
				},
			},
			expected: true,
		},
		{
			name: "nil slices treated as empty",
			result: &ComposeDiffResult{
				MissingInEnv:     nil,
				ExtraInEnv:       nil,
				MissingEnvFiles:  nil,
				ServiceBreakdown: nil,
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.result.HasIssues()
			if result != tt.expected {
				t.Errorf("HasIssues() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestGenerateComposeReport(t *testing.T) {
	tests := []struct {
		name        string
		result      *ComposeDiffResult
		opts        *ReportOptions
		contains    []string // Strings that should be in the report
		notContains []string // Strings that should NOT be in the report
	}{
		{
			name: "no issues report",
			result: &ComposeDiffResult{
				MissingInEnv:     []string{},
				ExtraInEnv:       []string{},
				MissingEnvFiles:  []string{},
				ServiceBreakdown: map[string][]string{},
			},
			opts: &ReportOptions{ShowDuck: true, Colorize: true, Verbose: false},
			contains: []string{
				"✅ Docker Compose environment is aligned",
				"gopher-duck approves",
			},
			notContains: []string{
				"QUACK!",
				"Missing",
				"Extra",
			},
		},
		{
			name: "missing variables report",
			result: &ComposeDiffResult{
				MissingInEnv:    []string{"API_KEY", "DATABASE_URL"},
				ExtraInEnv:      []string{},
				MissingEnvFiles: []string{},
				ServiceBreakdown: map[string][]string{
					"web": {"API_KEY", "DATABASE_URL"},
				},
			},
			opts: &ReportOptions{ShowDuck: true, Colorize: true, Verbose: true},
			contains: []string{
				"QUACK!",
				"🔴 Variables required by compose but missing",
				"API_KEY",
				"DATABASE_URL",
				"📋 Service breakdown",
				"web:",
			},
			notContains: []string{
				"✅",
				"Extra",
			},
		},
		{
			name: "missing env files report",
			result: &ComposeDiffResult{
				MissingInEnv:     []string{},
				ExtraInEnv:       []string{},
				MissingEnvFiles:  []string{".env.local", ".env.production"},
				ServiceBreakdown: map[string][]string{},
			},
			opts: &ReportOptions{ShowDuck: true, Colorize: true, Verbose: false},
			contains: []string{
				"💥 Missing env_files referenced in compose",
				".env.local",
				".env.production",
			},
			notContains: []string{
				"✅",
				"Variables required",
				"Service breakdown",
			},
		},
		{
			name: "extra variables report",
			result: &ComposeDiffResult{
				MissingInEnv:     []string{},
				ExtraInEnv:       []string{"DEBUG", "LOG_LEVEL"},
				MissingEnvFiles:  []string{},
				ServiceBreakdown: map[string][]string{},
			},
			opts: &ReportOptions{ShowDuck: false, Colorize: false, Verbose: false},
			contains: []string{
				"Unused variables:",
				"DEBUG",
				"LOG_LEVEL",
			},
			notContains: []string{
				"🦆",
				"gopher-duck",
				"🟡",
				"✅",
			},
		},
		{
			name: "comprehensive report with all issue types",
			result: &ComposeDiffResult{
				MissingInEnv:    []string{"API_KEY", "JWT_SECRET"},
				ExtraInEnv:      []string{"DEBUG", "LOG_LEVEL"},
				MissingEnvFiles: []string{".env.local"},
				ServiceBreakdown: map[string][]string{
					"web":    {"API_KEY"},
					"worker": {"JWT_SECRET"},
				},
			},
			opts: &ReportOptions{ShowDuck: true, Colorize: true, Verbose: true},
			contains: []string{
				"QUACK!",
				"💥 Missing env_files",
				".env.local",
				"🔴 Variables required by compose but missing",
				"API_KEY",
				"JWT_SECRET",
				"📋 Service breakdown",
				"web:",
				"worker:",
				"🟡 Variables in env files but not used",
				"DEBUG",
				"LOG_LEVEL",
				"confused by your container setup",
			},
			notContains: []string{
				"✅",
				"approves",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			report := GenerateComposeReport(tt.result, tt.opts)

			// Check that required strings are present
			for _, required := range tt.contains {
				if !strings.Contains(report, required) {
					t.Errorf("Report should contain %q but doesn't. Report:\n%s", required, report)
				}
			}

			// Check that forbidden strings are not present
			for _, forbidden := range tt.notContains {
				if strings.Contains(report, forbidden) {
					t.Errorf("Report should not contain %q but does. Report:\n%s", forbidden, report)
				}
			}
		})
	}
}

func TestGenerateComposeReport_DefaultOptions(t *testing.T) {
	result := &ComposeDiffResult{
		MissingInEnv:    []string{"API_KEY"},
		ExtraInEnv:      []string{},
		MissingEnvFiles: []string{},
		ServiceBreakdown: map[string][]string{
			"web": {"API_KEY"},
		},
	}

	// Test with nil options (should use defaults)
	report := GenerateComposeReport(result, nil)

	// Should contain default behavior
	if !strings.Contains(report, "API_KEY") {
		t.Errorf("Report should contain API_KEY with default options")
	}
}

// Helper function to compare ComposeDiffResult structs
func compareComposeDiffResults(a, b *ComposeDiffResult) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}

	// Compare MissingInEnv
	aMissingCopy := make([]string, len(a.MissingInEnv))
	copy(aMissingCopy, a.MissingInEnv)
	sort.Strings(aMissingCopy)

	bMissingCopy := make([]string, len(b.MissingInEnv))
	copy(bMissingCopy, b.MissingInEnv)
	sort.Strings(bMissingCopy)

	if !reflect.DeepEqual(aMissingCopy, bMissingCopy) {
		return false
	}

	// Compare ExtraInEnv
	aExtraCopy := make([]string, len(a.ExtraInEnv))
	copy(aExtraCopy, a.ExtraInEnv)
	sort.Strings(aExtraCopy)

	bExtraCopy := make([]string, len(b.ExtraInEnv))
	copy(bExtraCopy, b.ExtraInEnv)
	sort.Strings(bExtraCopy)

	if !reflect.DeepEqual(aExtraCopy, bExtraCopy) {
		return false
	}

	// Compare MissingEnvFiles
	aFilesCopy := make([]string, len(a.MissingEnvFiles))
	copy(aFilesCopy, a.MissingEnvFiles)
	sort.Strings(aFilesCopy)

	bFilesCopy := make([]string, len(b.MissingEnvFiles))
	copy(bFilesCopy, b.MissingEnvFiles)
	sort.Strings(bFilesCopy)

	if !reflect.DeepEqual(aFilesCopy, bFilesCopy) {
		return false
	}

	// Compare ServiceBreakdown
	if len(a.ServiceBreakdown) != len(b.ServiceBreakdown) {
		return false
	}

	for service, aVars := range a.ServiceBreakdown {
		bVars, exists := b.ServiceBreakdown[service]
		if !exists {
			return false
		}

		aVarsCopy := make([]string, len(aVars))
		copy(aVarsCopy, aVars)
		sort.Strings(aVarsCopy)

		bVarsCopy := make([]string, len(bVars))
		copy(bVarsCopy, bVars)
		sort.Strings(bVarsCopy)

		if !reflect.DeepEqual(aVarsCopy, bVarsCopy) {
			return false
		}
	}

	return true
}

// Benchmark tests
func BenchmarkCompareComposeWithEnv_Small(b *testing.B) {
	composeContent := `version: '3.8'
services:
  web:
    environment:
      - NODE_ENV=${NODE_ENV}
      - DATABASE_URL=${DATABASE_URL}
      - PORT=${PORT}`

	envContent := `NODE_ENV=production
DATABASE_URL=postgres://localhost
PORT=3000`

	tmpDir := b.TempDir()
	composeFile := filepath.Join(tmpDir, "docker-compose.yml")
	envFile := filepath.Join(tmpDir, ".env")

	os.WriteFile(composeFile, []byte(composeContent), 0644)
	os.WriteFile(envFile, []byte(envContent), 0644)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := CompareComposeWithEnv(composeFile, []string{envFile})
		if err != nil {
			b.Fatalf("Benchmark failed: %v", err)
		}
	}
}

func BenchmarkCompareComposeWithEnv_Large(b *testing.B) {
	// Create large compose with many services and variables
	var composeContent strings.Builder
	var envContent strings.Builder

	composeContent.WriteString("version: '3.8'\nservices:\n")

	for i := 0; i < 20; i++ {
		composeContent.WriteString(fmt.Sprintf("  service_%d:\n", i))
		composeContent.WriteString("    environment:\n")
		for j := 0; j < 10; j++ {
			varName := fmt.Sprintf("VAR_%d_%d", i, j)
			composeContent.WriteString(fmt.Sprintf("      - %s=${%s}\n", varName, varName))
			envContent.WriteString(fmt.Sprintf("%s=value_%d_%d\n", varName, i, j))
		}
	}

	tmpDir := b.TempDir()
	composeFile := filepath.Join(tmpDir, "docker-compose.yml")
	envFile := filepath.Join(tmpDir, ".env")

	os.WriteFile(composeFile, []byte(composeContent.String()), 0644)
	os.WriteFile(envFile, []byte(envContent.String()), 0644)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := CompareComposeWithEnv(composeFile, []string{envFile})
		if err != nil {
			b.Fatalf("Benchmark failed: %v", err)
		}
	}
}
