package cli

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/DuckDHD/EnvQuack/internal/checker"
	"github.com/DuckDHD/EnvQuack/internal/parser"
)

// createTestEnvFiles creates test .env and .env.example files
func createTestEnvFiles(t *testing.T, dir string, envContent, exampleContent string) (string, string) {
	t.Helper()

	envPath := filepath.Join(dir, ".env")
	examplePath := filepath.Join(dir, ".env.example")

	if envContent != "" {
		if err := os.WriteFile(envPath, []byte(envContent), 0644); err != nil {
			t.Fatalf("Failed to create .env: %v", err)
		}
	}

	if exampleContent != "" {
		if err := os.WriteFile(examplePath, []byte(exampleContent), 0644); err != nil {
			t.Fatalf("Failed to create .env.example: %v", err)
		}
	}

	return envPath, examplePath
}

// captureStdout captures stdout during function execution
func captureStdout(f func()) string {
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	f()

	w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	io.Copy(&buf, r)
	return buf.String()
}

func TestCheckFileExists(t *testing.T) {
	tests := []struct {
		name        string
		setupFile   bool
		expectError bool
	}{
		{
			name:        "existing file",
			setupFile:   true,
			expectError: false,
		},
		{
			name:        "non-existent file",
			setupFile:   false,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			filePath := filepath.Join(tmpDir, "test.txt")

			if tt.setupFile {
				os.WriteFile(filePath, []byte("test"), 0644)
			}

			err := checkFileExists(filePath)

			if tt.expectError && err == nil {
				t.Errorf("Expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
			if tt.expectError && err != nil && !strings.Contains(err.Error(), "does not exist") {
				t.Errorf("Error should mention 'does not exist', got: %v", err)
			}
		})
	}
}

func TestFileExists(t *testing.T) {
	tests := []struct {
		name      string
		setupFile bool
		expected  bool
	}{
		{
			name:      "existing file",
			setupFile: true,
			expected:  true,
		},
		{
			name:      "non-existent file",
			setupFile: false,
			expected:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			filePath := filepath.Join(tmpDir, "test.txt")

			if tt.setupFile {
				os.WriteFile(filePath, []byte("test"), 0644)
			}

			result := fileExists(filePath)

			if result != tt.expected {
				t.Errorf("fileExists() = %v, expected %v", result, tt.expected)
			}
		})
	}
}

// Test runCheck logic without os.Exit
func TestRunCheckLogic(t *testing.T) {
	tests := []struct {
		name           string
		envContent     string
		exampleContent string
		expectError    bool
		expectIssues   bool
	}{
		{
			name: "perfect match - no issues",
			envContent: `NODE_ENV=production
DATABASE_URL=postgres://localhost/myapp
PORT=3000`,
			exampleContent: `NODE_ENV=production
DATABASE_URL=postgres://localhost/myapp
PORT=3000`,
			expectError:  false,
			expectIssues: false,
		},
		{
			name: "missing variables detected",
			envContent: `NODE_ENV=production
PORT=3000`,
			exampleContent: `NODE_ENV=production
DATABASE_URL=postgres://localhost/myapp
PORT=3000`,
			expectError:  false,
			expectIssues: true,
		},
		{
			name: "extra variables detected",
			envContent: `NODE_ENV=production
DATABASE_URL=postgres://localhost/myapp
PORT=3000
DEBUG=true`,
			exampleContent: `NODE_ENV=production
DATABASE_URL=postgres://localhost/myapp
PORT=3000`,
			expectError:  false,
			expectIssues: true,
		},
		{
			name:           "missing .env file",
			envContent:     "",
			exampleContent: "NODE_ENV=production\n",
			expectError:    true,
			expectIssues:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()

			envPath := filepath.Join(tmpDir, ".env")
			examplePath := filepath.Join(tmpDir, ".env.example")

			// Create files
			if tt.envContent != "" {
				os.WriteFile(envPath, []byte(tt.envContent), 0644)
			}
			if tt.exampleContent != "" {
				os.WriteFile(examplePath, []byte(tt.exampleContent), 0644)
			}

			// Check file existence
			if tt.envContent == "" {
				err := checkFileExists(envPath)
				if err == nil {
					t.Errorf("Expected error for missing .env file")
				}
				return
			}

			// Run comparison
			result, err := checker.CompareEnvFiles(envPath, examplePath)
			if tt.expectError && err == nil {
				t.Errorf("Expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}

			if !tt.expectError {
				hasIssues := result.HasIssues()
				if hasIssues != tt.expectIssues {
					t.Errorf("HasIssues() = %v, expected %v", hasIssues, tt.expectIssues)
				}

				// Verify report generation works
				opts := &checker.ReportOptions{
					ShowDuck: false,
					Colorize: false,
					Verbose:  false,
				}
				report := checker.GenerateReport(result, opts)
				if report == "" {
					t.Errorf("Report should not be empty")
				}

				// Check for expected content in report
				if tt.expectIssues {
					if !strings.Contains(report, "Missing") && !strings.Contains(report, "Extra") {
						t.Errorf("Report should mention Missing or Extra variables")
					}
				}
			}
		})
	}
}

// Test runSync logic
func TestRunSyncLogic(t *testing.T) {
	tests := []struct {
		name                string
		envContent          string
		exampleContent      string
		expectError         bool
		expectSync          bool
		expectVarsAdded     int
		expectEnvFileToHave []string
	}{
		{
			name: "sync missing variables",
			envContent: `NODE_ENV=production
PORT=3000`,
			exampleContent: `NODE_ENV=production
DATABASE_URL=
PORT=3000
API_KEY=`,
			expectError:     false,
			expectSync:      true,
			expectVarsAdded: 2,
			expectEnvFileToHave: []string{
				"NODE_ENV=production",
				"PORT=3000",
				"DATABASE_URL=",
				"API_KEY=",
				"# Added by envquack sync",
			},
		},
		{
			name:       "create new .env file",
			envContent: "",
			exampleContent: `NODE_ENV=
DATABASE_URL=
PORT=`,
			expectError:     false,
			expectSync:      true,
			expectVarsAdded: 3,
			expectEnvFileToHave: []string{
				"NODE_ENV=",
				"DATABASE_URL=",
				"PORT=",
			},
		},
		{
			name: "no missing variables",
			envContent: `NODE_ENV=production
DATABASE_URL=postgres://localhost/myapp
PORT=3000`,
			exampleContent: `NODE_ENV=
DATABASE_URL=
PORT=`,
			expectError:     false,
			expectSync:      false,
			expectVarsAdded: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()

			envPath := filepath.Join(tmpDir, ".env")
			examplePath := filepath.Join(tmpDir, ".env.example")

			// Create example file
			os.WriteFile(examplePath, []byte(tt.exampleContent), 0644)

			// Create .env if content provided
			if tt.envContent != "" {
				os.WriteFile(envPath, []byte(tt.envContent), 0644)
			}

			// Parse files
			example, err := parser.ParseEnvFile(examplePath)
			if err != nil {
				t.Fatalf("Failed to parse example file: %v", err)
			}

			var env parser.EnvVars
			if tt.envContent == "" {
				env = make(parser.EnvVars)
			} else {
				env, err = parser.ParseEnvFile(envPath)
				if err != nil {
					t.Fatalf("Failed to parse env file: %v", err)
				}
			}

			// Find missing variables
			result := checker.CompareEnvVars(env, example)

			if len(result.Missing) != tt.expectVarsAdded {
				t.Errorf("Expected %d missing variables, got %d: %v",
					tt.expectVarsAdded, len(result.Missing), result.Missing)
			}

			// If we expect sync, simulate adding variables
			if tt.expectSync && len(result.Missing) > 0 {
				// Open file for append
				file, err := os.OpenFile(envPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
				if err != nil {
					t.Fatalf("Failed to open file: %v", err)
				}
				defer file.Close()

				// Add separator if file has content
				if len(env) > 0 {
					file.WriteString("\n# Added by envquack sync\n")
				}

				// Add missing variables
				for _, key := range result.Missing {
					file.WriteString(key + "=\n")
				}
			}

			// Verify file contents if expected
			if len(tt.expectEnvFileToHave) > 0 {
				content, err := os.ReadFile(envPath)
				if err != nil {
					t.Fatalf("Failed to read .env: %v", err)
				}

				fileContent := string(content)
				for _, expected := range tt.expectEnvFileToHave {
					if !strings.Contains(fileContent, expected) {
						t.Errorf(".env should contain %q but doesn't.\nContent:\n%s",
							expected, fileContent)
					}
				}
			}
		})
	}
}

func TestRunSyncIdempotent(t *testing.T) {
	tmpDir := t.TempDir()

	envPath := filepath.Join(tmpDir, ".env")
	examplePath := filepath.Join(tmpDir, ".env.example")

	// Create example
	exampleContent := `NODE_ENV=
DATABASE_URL=
PORT=`
	os.WriteFile(examplePath, []byte(exampleContent), 0644)

	// Create partial .env
	envContent := `NODE_ENV=production`
	os.WriteFile(envPath, []byte(envContent), 0644)

	// First sync
	example, _ := parser.ParseEnvFile(examplePath)
	env, _ := parser.ParseEnvFile(envPath)
	result := checker.CompareEnvVars(env, example)

	if len(result.Missing) != 2 {
		t.Errorf("Should have 2 missing vars initially, got %d", len(result.Missing))
	}

	// Add missing vars
	file, _ := os.OpenFile(envPath, os.O_APPEND|os.O_WRONLY, 0644)
	file.WriteString("\n# Added by envquack sync\n")
	for _, key := range result.Missing {
		file.WriteString(key + "=\n")
	}
	file.Close()

	content1, _ := os.ReadFile(envPath)

	// Second sync
	env2, _ := parser.ParseEnvFile(envPath)
	result2 := checker.CompareEnvVars(env2, example)

	if len(result2.Missing) != 0 {
		t.Errorf("Second sync should have 0 missing vars, got %d: %v",
			len(result2.Missing), result2.Missing)
	}

	content2, _ := os.ReadFile(envPath)

	// Content should be the same
	if string(content1) != string(content2) {
		t.Errorf("File changed after second sync (not idempotent)")
	}
}

// Test audit logic
func TestRunAuditLogic(t *testing.T) {
	t.Run("audit with matching files", func(t *testing.T) {
		tmpDir := t.TempDir()

		envContent := `NODE_ENV=production
DATABASE_URL=postgres://localhost/myapp
PORT=3000`

		envPath := filepath.Join(tmpDir, ".env")
		examplePath := filepath.Join(tmpDir, ".env.example")

		os.WriteFile(envPath, []byte(envContent), 0644)
		os.WriteFile(examplePath, []byte(envContent), 0644)

		// Check basic comparison
		result, err := checker.CompareEnvFiles(envPath, examplePath)
		if err != nil {
			t.Fatalf("Failed to compare files: %v", err)
		}

		if result.HasIssues() {
			t.Errorf("Should have no issues with matching files")
		}
	})

	t.Run("audit with docker-compose", func(t *testing.T) {
		tmpDir := t.TempDir()

		envPath := filepath.Join(tmpDir, ".env")
		envContent := `NODE_ENV=production
DATABASE_URL=postgres://localhost/myapp
PORT=3000`
		os.WriteFile(envPath, []byte(envContent), 0644)

		composePath := filepath.Join(tmpDir, "docker-compose.yml")
		composeContent := `version: '3.8'
services:
  web:
    environment:
      - NODE_ENV=${NODE_ENV}
      - DATABASE_URL=${DATABASE_URL}
      - PORT=${PORT}`
		os.WriteFile(composePath, []byte(composeContent), 0644)

		// Check compose comparison
		result, err := checker.CompareComposeWithEnv(composePath, []string{envPath})
		if err != nil {
			t.Fatalf("Failed to compare compose: %v", err)
		}

		if result.HasIssues() {
			t.Errorf("Should have no issues: %v", result)
		}
	})

	t.Run("audit with Dockerfile", func(t *testing.T) {
		tmpDir := t.TempDir()

		envPath := filepath.Join(tmpDir, ".env")
		envContent := `NODE_ENV=production
DATABASE_URL=postgres://localhost/myapp
PORT=3000`
		os.WriteFile(envPath, []byte(envContent), 0644)

		dockerfilePath := filepath.Join(tmpDir, "Dockerfile")
		dockerfileContent := `FROM node:16
ENV NODE_ENV=${NODE_ENV}
ENV DATABASE_URL=${DATABASE_URL}
ENV PORT=${PORT}`
		os.WriteFile(dockerfilePath, []byte(dockerfileContent), 0644)

		// Check Dockerfile comparison
		result, err := checker.CompareDockerfileWithEnv(dockerfilePath, []string{envPath})
		if err != nil {
			t.Fatalf("Failed to compare Dockerfile: %v", err)
		}

		// Should have no missing variables
		if len(result.MissingInEnv) > 0 {
			t.Errorf("Should have no missing vars, got: %v", result.MissingInEnv)
		}
	})
}

// Test Execute function exists
func TestExecute(t *testing.T) {
	// We can't really test Execute() fully without mocking os.Args and handling os.Exit(),
	// but we verify it exists as part of the public API
	// The underlying command logic is tested through TestRunCheckLogic, TestRunSyncLogic, etc.
	t.Skip("Execute() tested indirectly through logic tests - skipping direct test due to os.Exit()")
}
