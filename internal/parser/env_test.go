package parser

import (
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strings"
	"testing"
)

func TestParseEnvFile(t *testing.T) {
	tests := []struct {
		name        string
		content     string
		expected    EnvVars
		expectError bool
	}{
		{
			name: "basic key-value pairs",
			content: `DATABASE_URL=postgres://user:pass@localhost/db
API_KEY=abc123
PORT=3000`,
			expected: EnvVars{
				"DATABASE_URL": "postgres://user:pass@localhost/db",
				"API_KEY":      "abc123",
				"PORT":         "3000",
			},
			expectError: false,
		},
		{
			name: "empty values",
			content: `EMPTY_VAR=
ANOTHER_EMPTY=`,
			expected: EnvVars{
				"EMPTY_VAR":     "",
				"ANOTHER_EMPTY": "",
			},
			expectError: false,
		},
		{
			name: "double quoted values",
			content: `QUOTED="hello world"
WITH_SPACES="value with spaces"
SPECIAL_CHARS="!@#$%^&*()_+-={}[]|\\:;\"'<>?,./"`,
			expected: EnvVars{
				"QUOTED":        "hello world",
				"WITH_SPACES":   "value with spaces",
				"SPECIAL_CHARS": "!@#$%^&*()_+-={}[]|\\:;\"'<>?,./",
			},
			expectError: false,
		},
		{
			name: "single quoted values",
			content: `SINGLE='single quoted'
MIXED_QUOTES='value with "double" quotes inside'`,
			expected: EnvVars{
				"SINGLE":       "single quoted",
				"MIXED_QUOTES": "value with \"double\" quotes inside",
			},
			expectError: false,
		},
		{
			name: "comments and empty lines",
			content: `# This is a comment
DATABASE_URL=postgres://localhost

# Another comment
API_KEY=secret

# Empty line above and below

PORT=3000`,
			expected: EnvVars{
				"DATABASE_URL": "postgres://localhost",
				"API_KEY":      "secret",
				"PORT":         "3000",
			},
			expectError: false,
		},
		{
			name: "values with equals signs",
			content: `EQUATION=x=y+z
BASE64=dGVzdA==
URL_WITH_PARAMS=https://api.example.com?key=value&another=test`,
			expected: EnvVars{
				"EQUATION":        "x=y+z",
				"BASE64":          "dGVzdA==",
				"URL_WITH_PARAMS": "https://api.example.com?key=value&another=test",
			},
			expectError: false,
		},
		{
			name: "whitespace handling",
			content: `  TRIMMED_KEY  =  trimmed_value  
	TAB_KEY	=	tab_value	
NORMAL=value`,
			expected: EnvVars{
				"TRIMMED_KEY": "trimmed_value",
				"TAB_KEY":     "tab_value",
				"NORMAL":      "value",
			},
			expectError: false,
		},
		{
			name: "malformed lines ignored",
			content: `VALID=value
INVALID_NO_EQUALS
ANOTHER_VALID=another_value
=NO_KEY
KEY_WITH_NO_VALUE_NO_EQUALS
FINAL=final_value`,
			expected: EnvVars{
				"VALID":         "value",
				"ANOTHER_VALID": "another_value",
				"FINAL":         "final_value",
			},
			expectError: false,
		},
		{
			name: "quotes edge cases",
			content: `UNMATCHED_QUOTE="unmatched quote
EMPTY_QUOTES=""
EMPTY_SINGLE_QUOTES=''
QUOTE_IN_MIDDLE=val"ue
NESTED_QUOTES="outer 'inner' quotes"`,
			expected: EnvVars{
				"UNMATCHED_QUOTE":     "\"unmatched quote",
				"EMPTY_QUOTES":        "",
				"EMPTY_SINGLE_QUOTES": "",
				"QUOTE_IN_MIDDLE":     "val\"ue",
				"NESTED_QUOTES":       "outer 'inner' quotes",
			},
			expectError: false,
		},
		{
			name: "unicode and special characters",
			content: `UNICODE=café
EMOJI=🦆
JAPANESE=こんにちは
SYMBOLS=αβγδε`,
			expected: EnvVars{
				"UNICODE":  "café",
				"EMOJI":    "🦆",
				"JAPANESE": "こんにちは",
				"SYMBOLS":  "αβγδε",
			},
			expectError: false,
		},
		{
			name:        "empty file",
			content:     "",
			expected:    EnvVars{},
			expectError: false,
		},
		{
			name: "only comments and empty lines",
			content: `# Just comments
# Another comment

# More comments`,
			expected:    EnvVars{},
			expectError: false,
		},
		{
			name: "complex real-world example",
			content: `# Database configuration
DATABASE_URL="postgresql://user:password@localhost:5432/myapp?sslmode=require"
DATABASE_POOL_SIZE=20

# API configuration
API_KEY=sk_test_1234567890abcdef
API_SECRET="super-secret-key-with-special-chars!@#$%"
API_TIMEOUT=30

# Feature flags
FEATURE_NEW_UI=true
FEATURE_ANALYTICS="enabled"
DEBUG=false

# URLs and paths
BASE_URL=https://api.example.com
UPLOAD_PATH="/tmp/uploads"
LOG_LEVEL=info`,
			expected: EnvVars{
				"DATABASE_URL":       "postgresql://user:password@localhost:5432/myapp?sslmode=require",
				"DATABASE_POOL_SIZE": "20",
				"API_KEY":            "sk_test_1234567890abcdef",
				"API_SECRET":         "super-secret-key-with-special-chars!@#$%",
				"API_TIMEOUT":        "30",
				"FEATURE_NEW_UI":     "true",
				"FEATURE_ANALYTICS":  "enabled",
				"DEBUG":              "false",
				"BASE_URL":           "https://api.example.com",
				"UPLOAD_PATH":        "/tmp/uploads",
				"LOG_LEVEL":          "info",
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temporary file
			tmpDir := t.TempDir()
			tmpFile := filepath.Join(tmpDir, ".env")

			err := os.WriteFile(tmpFile, []byte(tt.content), 0644)
			if err != nil {
				t.Fatalf("Failed to create test file: %v", err)
			}

			// Parse the file
			result, err := ParseEnvFile(tmpFile)

			// Check error expectation
			if tt.expectError && err == nil {
				t.Errorf("Expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}

			// Check results
			if !tt.expectError {
				if !reflect.DeepEqual(result, tt.expected) {
					t.Errorf("ParseEnvFile() = %v, want %v", result, tt.expected)
				}
			}
		})
	}
}

func TestParseEnvFile_FileErrors(t *testing.T) {
	tests := []struct {
		name     string
		filename string
		setup    func(string) error // Optional setup function
	}{
		{
			name:     "non-existent file",
			filename: "/non/existent/file.env",
		},
		{
			name:     "directory instead of file",
			filename: "directory.env",
			setup: func(filename string) error {
				return os.Mkdir(filename, 0755)
			},
		},
		{
			name:     "permission denied",
			filename: "permission_denied.env",
			setup: func(filename string) error {
				if err := os.WriteFile(filename, []byte("TEST=value"), 0644); err != nil {
					return err
				}
				return os.Chmod(filename, 0000)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			testFile := filepath.Join(tmpDir, tt.filename)

			// Setup if needed
			if tt.setup != nil {
				if err := tt.setup(testFile); err != nil {
					t.Fatalf("Setup failed: %v", err)
				}
				// Clean up permissions after test
				defer func() {
					os.Chmod(testFile, 0644)
					os.Remove(testFile)
				}()
			}

			// Attempt to parse
			result, err := ParseEnvFile(testFile)

			// Should always error
			if err == nil {
				t.Errorf("Expected error for %s, but got none. Result: %v", tt.name, result)
			}

			// Result should be nil on error
			if result != nil {
				t.Errorf("Expected nil result on error, got: %v", result)
			}
		})
	}
}

func TestEnvVars_GetKeys(t *testing.T) {
	tests := []struct {
		name     string
		envVars  EnvVars
		expected []string
	}{
		{
			name:     "empty env vars",
			envVars:  EnvVars{},
			expected: []string{},
		},
		{
			name: "single key",
			envVars: EnvVars{
				"KEY": "value",
			},
			expected: []string{"KEY"},
		},
		{
			name: "multiple keys",
			envVars: EnvVars{
				"DATABASE_URL": "postgres://localhost",
				"API_KEY":      "secret",
				"PORT":         "3000",
			},
			expected: []string{"API_KEY", "DATABASE_URL", "PORT"}, // Should be sorted
		},
		{
			name: "keys with various formats",
			envVars: EnvVars{
				"SIMPLE":           "value",
				"WITH_UNDERSCORES": "value",
				"123NUMERIC":       "value",
				"MixedCase":        "value",
			},
			expected: []string{"123NUMERIC", "MixedCase", "SIMPLE", "WITH_UNDERSCORES"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.envVars.GetKeys()

			// Sort both slices for comparison since map iteration order is not guaranteed
			sort.Strings(result)
			sort.Strings(tt.expected)

			if !reflect.DeepEqual(result, tt.expected) {
				t.Errorf("GetKeys() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestEnvVars_Has(t *testing.T) {
	envVars := EnvVars{
		"DATABASE_URL": "postgres://localhost",
		"API_KEY":      "secret",
		"PORT":         "3000",
		"EMPTY":        "",
	}

	tests := []struct {
		name     string
		key      string
		expected bool
	}{
		{
			name:     "existing key with value",
			key:      "DATABASE_URL",
			expected: true,
		},
		{
			name:     "existing key with empty value",
			key:      "EMPTY",
			expected: true,
		},
		{
			name:     "non-existing key",
			key:      "NON_EXISTING",
			expected: false,
		},
		{
			name:     "case sensitive - existing key with wrong case",
			key:      "api_key",
			expected: false,
		},
		{
			name:     "empty key",
			key:      "",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := envVars.Has(tt.key)
			if result != tt.expected {
				t.Errorf("Has(%q) = %v, want %v", tt.key, result, tt.expected)
			}
		})
	}
}

func TestParseEnvFile_LargeFile(t *testing.T) {
	// Test with a large number of variables to ensure performance
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, ".env")

	// Generate a large .env file
	var content strings.Builder
	expectedVars := make(EnvVars)

	for i := 0; i < 1000; i++ {
		key := fmt.Sprintf("VAR_%d", i)
		value := fmt.Sprintf("value_%d", i)
		content.WriteString(fmt.Sprintf("%s=%s\n", key, value))
		expectedVars[key] = value
	}

	err := os.WriteFile(tmpFile, []byte(content.String()), 0644)
	if err != nil {
		t.Fatalf("Failed to create large test file: %v", err)
	}

	// Parse the file
	result, err := ParseEnvFile(tmpFile)
	if err != nil {
		t.Fatalf("Failed to parse large file: %v", err)
	}

	// Verify all variables are present
	if len(result) != len(expectedVars) {
		t.Errorf("Expected %d variables, got %d", len(expectedVars), len(result))
	}

	// Spot check a few variables
	testKeys := []string{"VAR_0", "VAR_500", "VAR_999"}
	for _, key := range testKeys {
		if !result.Has(key) {
			t.Errorf("Missing expected key: %s", key)
		}
		if result[key] != expectedVars[key] {
			t.Errorf("Wrong value for %s: got %q, want %q", key, result[key], expectedVars[key])
		}
	}
}

func TestParseEnvFile_EdgeCaseQuotes(t *testing.T) {
	tests := []struct {
		name        string
		content     string
		expectedKey string
		expectedVal string
	}{
		{
			name:        "quote only at start",
			content:     `KEY="value without end quote`,
			expectedKey: "KEY",
			expectedVal: `"value without end quote`,
		},
		{
			name:        "quote only at end",
			content:     `KEY=value without start quote"`,
			expectedKey: "KEY",
			expectedVal: `value without start quote"`,
		},
		{
			name:        "mixed quote types",
			content:     `KEY="single inside' double"`,
			expectedKey: "KEY",
			expectedVal: `single inside' double`,
		},
		{
			name:        "escaped quotes inside",
			content:     `KEY="value with \" escaped quote"`,
			expectedKey: "KEY",
			expectedVal: `value with \" escaped quote`,
		},
		{
			name:        "multiple quotes",
			content:     `KEY="""triple quotes"""`,
			expectedKey: "KEY",
			expectedVal: `""triple quotes""`,
		},
		{
			name:        "single character quoted",
			content:     `KEY="x"`,
			expectedKey: "KEY",
			expectedVal: "x",
		},
		{
			name:        "empty quoted value",
			content:     `KEY=""`,
			expectedKey: "KEY",
			expectedVal: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			tmpFile := filepath.Join(tmpDir, ".env")

			err := os.WriteFile(tmpFile, []byte(tt.content), 0644)
			if err != nil {
				t.Fatalf("Failed to create test file: %v", err)
			}

			result, err := ParseEnvFile(tmpFile)
			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			if !result.Has(tt.expectedKey) {
				t.Errorf("Missing expected key: %s", tt.expectedKey)
			}

			if result[tt.expectedKey] != tt.expectedVal {
				t.Errorf("Wrong value for %s: got %q, want %q", tt.expectedKey, result[tt.expectedKey], tt.expectedVal)
			}
		})
	}
}

// Benchmark tests to ensure good performance
func BenchmarkParseEnvFile_Small(b *testing.B) {
	content := `DATABASE_URL=postgres://localhost
API_KEY=secret
PORT=3000`

	tmpDir := b.TempDir()
	tmpFile := filepath.Join(tmpDir, ".env")
	os.WriteFile(tmpFile, []byte(content), 0644)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := ParseEnvFile(tmpFile)
		if err != nil {
			b.Fatalf("Benchmark failed: %v", err)
		}
	}
}

func BenchmarkParseEnvFile_Large(b *testing.B) {
	// Create a large .env file for benchmarking
	var content strings.Builder
	for i := 0; i < 1000; i++ {
		content.WriteString(fmt.Sprintf("VAR_%d=value_%d\n", i, i))
	}

	tmpDir := b.TempDir()
	tmpFile := filepath.Join(tmpDir, ".env")
	os.WriteFile(tmpFile, []byte(content.String()), 0644)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := ParseEnvFile(tmpFile)
		if err != nil {
			b.Fatalf("Benchmark failed: %v", err)
		}
	}
}

// TestParseEnvFile_PathValidation tests security path validation
func TestParseEnvFile_PathValidation(t *testing.T) {
	// Create a temporary directory for testing
	tmpDir := t.TempDir()

	// Change to temp directory for consistent testing
	originalDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get current directory: %v", err)
	}
	defer os.Chdir(originalDir)

	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("failed to change to temp directory: %v", err)
	}

	// Create a valid test file
	validFile := ".env"
	if err := os.WriteFile(validFile, []byte("TEST=value"), 0644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	tests := []struct {
		name        string
		path        string
		expectError bool
		errorMsg    string
	}{
		{
			name:        "valid relative path in current dir",
			path:        ".env",
			expectError: false,
		},
		{
			name:        "reject parent directory traversal",
			path:        "../.env",
			expectError: true,
			errorMsg:    "invalid file path",
		},
		{
			name:        "reject multiple parent traversal",
			path:        "../../etc/passwd",
			expectError: true,
			errorMsg:    "invalid file path",
		},
		{
			name:        "reject absolute path (Unix style)",
			path:        "/etc/passwd",
			expectError: true,
			errorMsg:    "invalid file path",
		},
		{
			name:        "reject hidden parent traversal",
			path:        "config/../../etc/passwd",
			expectError: true,
			errorMsg:    "invalid file path",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := ParseEnvFile(tt.path)

			if tt.expectError {
				if err == nil {
					t.Errorf("expected error containing '%s', got nil", tt.errorMsg)
				} else if !strings.Contains(err.Error(), tt.errorMsg) {
					t.Errorf("expected error containing '%s', got '%s'", tt.errorMsg, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
			}
		})
	}
}

// TestParseEnvFile_FileSizeValidation tests file size limits
func TestParseEnvFile_FileSizeValidation(t *testing.T) {
	tmpDir := t.TempDir()

	// Change to temp directory
	originalDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get current directory: %v", err)
	}
	defer os.Chdir(originalDir)

	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("failed to change to temp directory: %v", err)
	}

	// Create a file larger than 10MB (the validation limit)
	largeFile := "large.env"
	f, err := os.Create(largeFile)
	if err != nil {
		t.Fatalf("failed to create large file: %v", err)
	}
	// Create an 11MB file
	if err := f.Truncate(11 * 1024 * 1024); err != nil {
		f.Close()
		t.Fatalf("failed to set file size: %v", err)
	}
	f.Close()

	_, err = ParseEnvFile(largeFile)
	if err == nil {
		t.Error("expected error for file too large, got nil")
	} else if !strings.Contains(err.Error(), "file too large") && !strings.Contains(err.Error(), "file validation failed") {
		t.Errorf("expected 'file too large' error, got: %v", err)
	}
}
