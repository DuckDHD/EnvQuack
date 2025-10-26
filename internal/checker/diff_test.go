package checker

import (
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"testing"

	"github.com/DuckDHD/EnvQuack/internal/parser"
)

func TestCompareEnvFiles(t *testing.T) {
	tests := []struct {
		name           string
		envContent     string
		exampleContent string
		expected       *DiffResult
		expectError    bool
	}{
		{
			name: "identical files",
			envContent: `DATABASE_URL=postgres://localhost/myapp
API_KEY=secret123
PORT=3000`,
			exampleContent: `DATABASE_URL=postgres://localhost/example
API_KEY=example_key
PORT=8080`,
			expected: &DiffResult{
				Missing: []string{},
				Extra:   []string{},
			},
			expectError: false,
		},
		{
			name: "missing variables in env",
			envContent: `DATABASE_URL=postgres://localhost/myapp
PORT=3000`,
			exampleContent: `DATABASE_URL=postgres://localhost/example
API_KEY=example_key
PORT=8080
REDIS_URL=redis://localhost`,
			expected: &DiffResult{
				Missing: []string{"API_KEY", "REDIS_URL"},
				Extra:   []string{},
			},
			expectError: false,
		},
		{
			name: "extra variables in env",
			envContent: `DATABASE_URL=postgres://localhost/myapp
API_KEY=secret123
PORT=3000
DEBUG=true
LOG_LEVEL=info`,
			exampleContent: `DATABASE_URL=postgres://localhost/example
API_KEY=example_key
PORT=8080`,
			expected: &DiffResult{
				Missing: []string{},
				Extra:   []string{"DEBUG", "LOG_LEVEL"},
			},
			expectError: false,
		},
		{
			name: "both missing and extra variables",
			envContent: `DATABASE_URL=postgres://localhost/myapp
PORT=3000
DEBUG=true
CUSTOM_VAR=custom_value`,
			exampleContent: `DATABASE_URL=postgres://localhost/example
API_KEY=example_key
PORT=8080
REDIS_URL=redis://localhost`,
			expected: &DiffResult{
				Missing: []string{"API_KEY", "REDIS_URL"},
				Extra:   []string{"CUSTOM_VAR", "DEBUG"},
			},
			expectError: false,
		},
		{
			name:       "empty env file",
			envContent: "",
			exampleContent: `API_KEY=example_key
PORT=8080`,
			expected: &DiffResult{
				Missing: []string{"API_KEY", "PORT"},
				Extra:   []string{},
			},
			expectError: false,
		},
		{
			name: "empty example file",
			envContent: `DATABASE_URL=postgres://localhost/myapp
PORT=3000`,
			exampleContent: "",
			expected: &DiffResult{
				Missing: []string{},
				Extra:   []string{"DATABASE_URL", "PORT"},
			},
			expectError: false,
		},
		{
			name:           "both files empty",
			envContent:     "",
			exampleContent: "",
			expected: &DiffResult{
				Missing: []string{},
				Extra:   []string{},
			},
			expectError: false,
		},
		{
			name: "files with comments and empty lines",
			envContent: `# Production environment
DATABASE_URL=postgres://localhost/myapp

# API Configuration
API_KEY=secret123
PORT=3000`,
			exampleContent: `# Example environment file
DATABASE_URL=postgres://localhost/example

# API Configuration
API_KEY=example_key
PORT=8080

# Optional features
REDIS_URL=redis://localhost`,
			expected: &DiffResult{
				Missing: []string{"REDIS_URL"},
				Extra:   []string{},
			},
			expectError: false,
		},
		{
			name: "variables with special characters",
			envContent: `DATABASE_URL="postgres://user:pass@localhost/db?sslmode=require"
API_KEY=sk_test_123abc
JWT_SECRET="super-secret-key!@#$%"`,
			exampleContent: `DATABASE_URL="postgres://user:pass@example.com/db?sslmode=require"
API_KEY=sk_example_456def
JWT_SECRET="example-secret-key!@#$%"
OAUTH_SECRET="oauth-secret-xyz"`,
			expected: &DiffResult{
				Missing: []string{"OAUTH_SECRET"},
				Extra:   []string{},
			},
			expectError: false,
		},
		{
			name: "case sensitive variable names",
			envContent: `database_url=postgres://localhost/myapp
API_KEY=secret123
Port=3000`,
			exampleContent: `DATABASE_URL=postgres://localhost/example
api_key=example_key
PORT=8080`,
			expected: &DiffResult{
				Missing: []string{"DATABASE_URL", "PORT", "api_key"},
				Extra:   []string{"API_KEY", "Port", "database_url"},
			},
			expectError: false,
		},
		{
			name: "variables with equals in values",
			envContent: `EQUATION=x=y+z
BASE64_KEY=dGVzdA==
URL_PARAMS=https://api.com?key=value&other=test`,
			exampleContent: `EQUATION=a=b+c
BASE64_KEY=ZXhhbXBsZQ==
URL_PARAMS=https://example.com?key=value&other=test
ANOTHER_EQUATION=1+1=2`,
			expected: &DiffResult{
				Missing: []string{"ANOTHER_EQUATION"},
				Extra:   []string{},
			},
			expectError: false,
		},
		{
			name: "mixed quote types",
			envContent: `SINGLE_QUOTED='single value'
DOUBLE_QUOTED="double value"
UNQUOTED=unquoted_value`,
			exampleContent: `SINGLE_QUOTED="example single"
DOUBLE_QUOTED='example double'
UNQUOTED=example_unquoted`,
			expected: &DiffResult{
				Missing: []string{},
				Extra:   []string{},
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temporary directory and files
			tmpDir := t.TempDir()
			envFile := filepath.Join(tmpDir, ".env")
			exampleFile := filepath.Join(tmpDir, ".env.example")

			// Write test files
			err := os.WriteFile(envFile, []byte(tt.envContent), 0644)
			if err != nil {
				t.Fatalf("Failed to create .env file: %v", err)
			}

			err = os.WriteFile(exampleFile, []byte(tt.exampleContent), 0644)
			if err != nil {
				t.Fatalf("Failed to create .env.example file: %v", err)
			}

			// Run the comparison
			result, err := CompareEnvFiles(envFile, exampleFile)

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
			if !compareDiffResults(result, tt.expected) {
				t.Errorf("CompareEnvFiles() result mismatch")
				t.Errorf("Got Missing: %v", result.Missing)
				t.Errorf("Want Missing: %v", tt.expected.Missing)
				t.Errorf("Got Extra: %v", result.Extra)
				t.Errorf("Want Extra: %v", tt.expected.Extra)
			}
		})
	}
}

func TestCompareEnvFiles_FileErrors(t *testing.T) {
	tests := []struct {
		name         string
		setupEnv     func(string) error
		setupExample func(string) error
		expectError  bool
	}{
		{
			name: "missing env file",
			setupEnv: func(filename string) error {
				// Don't create the file
				return nil
			},
			setupExample: func(filename string) error {
				return os.WriteFile(filename, []byte("API_KEY=example"), 0644)
			},
			expectError: true,
		},
		{
			name: "missing example file",
			setupEnv: func(filename string) error {
				return os.WriteFile(filename, []byte("API_KEY=secret"), 0644)
			},
			setupExample: func(filename string) error {
				// Don't create the file
				return nil
			},
			expectError: true,
		},
		{
			name: "both files missing",
			setupEnv: func(filename string) error {
				return nil // Don't create
			},
			setupExample: func(filename string) error {
				return nil // Don't create
			},
			expectError: true,
		},
		{
			name: "env file is directory",
			setupEnv: func(filename string) error {
				return os.Mkdir(filename, 0755)
			},
			setupExample: func(filename string) error {
				return os.WriteFile(filename, []byte("API_KEY=example"), 0644)
			},
			expectError: true,
		},
		{
			name: "example file is directory",
			setupEnv: func(filename string) error {
				return os.WriteFile(filename, []byte("API_KEY=secret"), 0644)
			},
			setupExample: func(filename string) error {
				return os.Mkdir(filename, 0755)
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			envFile := filepath.Join(tmpDir, ".env")
			exampleFile := filepath.Join(tmpDir, ".env.example")

			// Setup files
			if err := tt.setupEnv(envFile); err != nil {
				t.Fatalf("Failed to setup env file: %v", err)
			}
			if err := tt.setupExample(exampleFile); err != nil {
				t.Fatalf("Failed to setup example file: %v", err)
			}

			// Clean up after test
			defer func() {
				os.RemoveAll(envFile)
				os.RemoveAll(exampleFile)
			}()

			// Run comparison
			result, err := CompareEnvFiles(envFile, exampleFile)

			if tt.expectError && err == nil {
				t.Errorf("Expected error but got none. Result: %v", result)
			}
			if !tt.expectError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
		})
	}
}

func TestCompareEnvVars(t *testing.T) {
	tests := []struct {
		name     string
		env      parser.EnvVars
		example  parser.EnvVars
		expected *DiffResult
	}{
		{
			name: "identical env vars",
			env: parser.EnvVars{
				"DATABASE_URL": "postgres://localhost/myapp",
				"API_KEY":      "secret123",
				"PORT":         "3000",
			},
			example: parser.EnvVars{
				"DATABASE_URL": "postgres://localhost/example",
				"API_KEY":      "example_key",
				"PORT":         "8080",
			},
			expected: &DiffResult{
				Missing: []string{},
				Extra:   []string{},
			},
		},
		{
			name: "missing variables",
			env: parser.EnvVars{
				"DATABASE_URL": "postgres://localhost/myapp",
				"PORT":         "3000",
			},
			example: parser.EnvVars{
				"DATABASE_URL": "postgres://localhost/example",
				"API_KEY":      "example_key",
				"PORT":         "8080",
				"REDIS_URL":    "redis://localhost",
			},
			expected: &DiffResult{
				Missing: []string{"API_KEY", "REDIS_URL"},
				Extra:   []string{},
			},
		},
		{
			name: "extra variables",
			env: parser.EnvVars{
				"DATABASE_URL": "postgres://localhost/myapp",
				"API_KEY":      "secret123",
				"PORT":         "3000",
				"DEBUG":        "true",
				"LOG_LEVEL":    "info",
			},
			example: parser.EnvVars{
				"DATABASE_URL": "postgres://localhost/example",
				"API_KEY":      "example_key",
				"PORT":         "8080",
			},
			expected: &DiffResult{
				Missing: []string{},
				Extra:   []string{"DEBUG", "LOG_LEVEL"},
			},
		},
		{
			name: "both missing and extra",
			env: parser.EnvVars{
				"DATABASE_URL": "postgres://localhost/myapp",
				"PORT":         "3000",
				"DEBUG":        "true",
				"CUSTOM_VAR":   "custom_value",
			},
			example: parser.EnvVars{
				"DATABASE_URL": "postgres://localhost/example",
				"API_KEY":      "example_key",
				"PORT":         "8080",
				"REDIS_URL":    "redis://localhost",
			},
			expected: &DiffResult{
				Missing: []string{"API_KEY", "REDIS_URL"},
				Extra:   []string{"CUSTOM_VAR", "DEBUG"},
			},
		},
		{
			name: "empty env",
			env:  parser.EnvVars{},
			example: parser.EnvVars{
				"API_KEY": "example_key",
				"PORT":    "8080",
			},
			expected: &DiffResult{
				Missing: []string{"API_KEY", "PORT"},
				Extra:   []string{},
			},
		},
		{
			name: "empty example",
			env: parser.EnvVars{
				"DATABASE_URL": "postgres://localhost/myapp",
				"PORT":         "3000",
			},
			example: parser.EnvVars{},
			expected: &DiffResult{
				Missing: []string{},
				Extra:   []string{"DATABASE_URL", "PORT"},
			},
		},
		{
			name:    "both empty",
			env:     parser.EnvVars{},
			example: parser.EnvVars{},
			expected: &DiffResult{
				Missing: []string{},
				Extra:   []string{},
			},
		},
		{
			name: "case sensitivity",
			env: parser.EnvVars{
				"database_url": "postgres://localhost/myapp",
				"Api_Key":      "secret123",
			},
			example: parser.EnvVars{
				"DATABASE_URL": "postgres://localhost/example",
				"API_KEY":      "example_key",
			},
			expected: &DiffResult{
				Missing: []string{"API_KEY", "DATABASE_URL"},
				Extra:   []string{"Api_Key", "database_url"},
			},
		},
		{
			name: "variables with empty values",
			env: parser.EnvVars{
				"DATABASE_URL": "postgres://localhost/myapp",
				"API_KEY":      "",
				"DEBUG":        "",
			},
			example: parser.EnvVars{
				"DATABASE_URL": "postgres://localhost/example",
				"API_KEY":      "example_key",
				"PORT":         "",
			},
			expected: &DiffResult{
				Missing: []string{"PORT"},
				Extra:   []string{"DEBUG"},
			},
		},
		{
			name: "large number of variables",
			env: func() parser.EnvVars {
				vars := make(parser.EnvVars)
				for i := 0; i < 100; i++ {
					if i%2 == 0 { // Even numbers
						vars[fmt.Sprintf("VAR_%d", i)] = fmt.Sprintf("value_%d", i)
					}
				}
				return vars
			}(),
			example: func() parser.EnvVars {
				vars := make(parser.EnvVars)
				for i := 0; i < 100; i++ {
					if i%3 == 0 { // Multiples of 3
						vars[fmt.Sprintf("VAR_%d", i)] = fmt.Sprintf("example_%d", i)
					}
				}
				return vars
			}(),
			expected: func() *DiffResult {
				missing := []string{}
				extra := []string{}

				for i := 0; i < 100; i++ {
					varName := fmt.Sprintf("VAR_%d", i)
					inEnv := i%2 == 0
					inExample := i%3 == 0

					if inExample && !inEnv {
						missing = append(missing, varName)
					}
					if inEnv && !inExample {
						extra = append(extra, varName)
					}
				}

				sort.Strings(missing)
				sort.Strings(extra)

				return &DiffResult{Missing: missing, Extra: extra}
			}(),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CompareEnvVars(tt.env, tt.example)

			if !compareDiffResults(result, tt.expected) {
				t.Errorf("CompareEnvVars() result mismatch")
				t.Errorf("Got Missing: %v", result.Missing)
				t.Errorf("Want Missing: %v", tt.expected.Missing)
				t.Errorf("Got Extra: %v", result.Extra)
				t.Errorf("Want Extra: %v", tt.expected.Extra)
			}
		})
	}
}

func TestDiffResult_HasIssues(t *testing.T) {
	tests := []struct {
		name     string
		result   *DiffResult
		expected bool
	}{
		{
			name: "no issues",
			result: &DiffResult{
				Missing: []string{},
				Extra:   []string{},
			},
			expected: false,
		},
		{
			name: "has missing variables",
			result: &DiffResult{
				Missing: []string{"API_KEY", "DATABASE_URL"},
				Extra:   []string{},
			},
			expected: true,
		},
		{
			name: "has extra variables",
			result: &DiffResult{
				Missing: []string{},
				Extra:   []string{"DEBUG", "LOG_LEVEL"},
			},
			expected: true,
		},
		{
			name: "has both missing and extra",
			result: &DiffResult{
				Missing: []string{"API_KEY"},
				Extra:   []string{"DEBUG"},
			},
			expected: true,
		},
		{
			name: "nil slices (treated as empty)",
			result: &DiffResult{
				Missing: nil,
				Extra:   nil,
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

func TestCompareEnvVars_SortingConsistency(t *testing.T) {
	// Test that results are consistently sorted regardless of input order
	env := parser.EnvVars{
		"ZEBRA": "value",
		"ALPHA": "value",
		"BETA":  "value",
		"GAMMA": "value",
	}

	example := parser.EnvVars{
		"CHARLIE": "value",
		"DELTA":   "value",
		"ECHO":    "value",
		"ALPHA":   "value",
	}

	// Run comparison multiple times to ensure consistency
	var results []*DiffResult
	for i := 0; i < 5; i++ {
		result := CompareEnvVars(env, example)
		results = append(results, result)
	}

	// All results should be identical
	for i := 1; i < len(results); i++ {
		if !compareDiffResults(results[0], results[i]) {
			t.Errorf("Results not consistent between runs %d and %d", 0, i)
		}
	}

	// Check that results are sorted
	expectedMissing := []string{"CHARLIE", "DELTA", "ECHO"}
	expectedExtra := []string{"BETA", "GAMMA", "ZEBRA"}

	if !reflect.DeepEqual(results[0].Missing, expectedMissing) {
		t.Errorf("Missing not sorted correctly: got %v, want %v", results[0].Missing, expectedMissing)
	}

	if !reflect.DeepEqual(results[0].Extra, expectedExtra) {
		t.Errorf("Extra not sorted correctly: got %v, want %v", results[0].Extra, expectedExtra)
	}
}

func TestCompareEnvVars_EdgeCases(t *testing.T) {
	tests := []struct {
		name     string
		env      parser.EnvVars
		example  parser.EnvVars
		expected *DiffResult
	}{
		{
			name: "duplicate values different keys",
			env: parser.EnvVars{
				"KEY1": "same_value",
				"KEY2": "different_value",
			},
			example: parser.EnvVars{
				"KEY3": "same_value",
				"KEY4": "different_value",
			},
			expected: &DiffResult{
				Missing: []string{"KEY3", "KEY4"},
				Extra:   []string{"KEY1", "KEY2"},
			},
		},
		{
			name: "keys with special characters",
			env: parser.EnvVars{
				"KEY_WITH_UNDERSCORES": "value1",
				"KEY-WITH-DASHES":      "value2",
				"KEY.WITH.DOTS":        "value3",
				"123NUMERIC_START":     "value4",
			},
			example: parser.EnvVars{
				"KEY_WITH_UNDERSCORES": "example1",
				"KEY-WITH-DASHES":      "example2",
				"DIFFERENT_KEY":        "example3",
			},
			expected: &DiffResult{
				Missing: []string{"DIFFERENT_KEY"},
				Extra:   []string{"123NUMERIC_START", "KEY.WITH.DOTS"},
			},
		},
		{
			name: "very long variable names",
			env: parser.EnvVars{
				"THIS_IS_A_VERY_LONG_VARIABLE_NAME_THAT_EXCEEDS_NORMAL_LENGTH_EXPECTATIONS": "value1",
				"NORMAL_VAR": "value2",
			},
			example: parser.EnvVars{
				"THIS_IS_A_VERY_LONG_VARIABLE_NAME_THAT_EXCEEDS_NORMAL_LENGTH_EXPECTATIONS": "example1",
				"ANOTHER_NORMAL_VAR": "example2",
			},
			expected: &DiffResult{
				Missing: []string{"ANOTHER_NORMAL_VAR"},
				Extra:   []string{"NORMAL_VAR"},
			},
		},
		{
			name: "unicode variable names",
			env: parser.EnvVars{
				"CAFÉ_VAR":    "value1",
				"测试_VAR":      "value2",
				"ÉMOJI_🦆_VAR": "value3",
			},
			example: parser.EnvVars{
				"CAFÉ_VAR":   "example1",
				"ДРУГОЙ_VAR": "example2",
				"NORMAL_VAR": "example3",
			},
			expected: &DiffResult{
				Missing: []string{"NORMAL_VAR", "ДРУГОЙ_VAR"},
				Extra:   []string{"ÉMOJI_🦆_VAR", "测试_VAR"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CompareEnvVars(tt.env, tt.example)

			if !compareDiffResults(result, tt.expected) {
				t.Errorf("CompareEnvVars() result mismatch")
				t.Errorf("Got Missing: %v", result.Missing)
				t.Errorf("Want Missing: %v", tt.expected.Missing)
				t.Errorf("Got Extra: %v", result.Extra)
				t.Errorf("Want Extra: %v", tt.expected.Extra)
			}
		})
	}
}

func TestCompareEnvFiles_RealWorldScenarios(t *testing.T) {
	tests := []struct {
		name            string
		envContent      string
		exampleContent  string
		expectedMissing int
		expectedExtra   int
	}{
		{
			name: "typical web application",
			envContent: `# Database
DATABASE_URL=postgres://user:pass@localhost:5432/myapp_production
DATABASE_POOL_SIZE=20

# API Keys
STRIPE_SECRET_KEY=sk_live_123...
STRIPE_PUBLISHABLE_KEY=pk_live_456...
SENDGRID_API_KEY=SG.789...

# Application
APP_ENV=production
BASE_URL=https://myapp.com
SESSION_SECRET=super-secret-session-key

# Feature Flags
FEATURE_ANALYTICS=true
FEATURE_CHAT=false

# Monitoring
SENTRY_DSN=https://123@sentry.io/456
NEW_RELIC_LICENSE_KEY=abc123...`,
			exampleContent: `# Database
DATABASE_URL=postgres://user:pass@localhost:5432/myapp_development
DATABASE_POOL_SIZE=5

# API Keys
STRIPE_SECRET_KEY=sk_test_123...
STRIPE_PUBLISHABLE_KEY=pk_test_456...
SENDGRID_API_KEY=SG.test_789...
TWILIO_AUTH_TOKEN=your_twilio_token

# Application
APP_ENV=development
BASE_URL=http://localhost:3000
SESSION_SECRET=development-session-secret
JWT_SECRET=your-jwt-secret

# Feature Flags
FEATURE_ANALYTICS=false
FEATURE_CHAT=true
FEATURE_BETA_UI=false

# Monitoring
SENTRY_DSN=https://test@sentry.io/test
DATADOG_API_KEY=your_datadog_key`,
			expectedMissing: 4, // TWILIO_AUTH_TOKEN, JWT_SECRET, FEATURE_BETA_UI, DATADOG_API_KEY
			expectedExtra:   1, // NEW_RELIC_LICENSE_KEY
		},
		{
			name: "microservice configuration",
			envContent: `# Service Configuration
SERVICE_NAME=user-service
SERVICE_VERSION=1.2.3
PORT=8080
HEALTH_CHECK_PATH=/health

# Database
DB_HOST=user-db.internal
DB_PORT=5432
DB_NAME=users
DB_USER=service_user
DB_PASSWORD=secure_password

# Message Queue
RABBITMQ_URL=amqp://service:pass@queue.internal:5672
QUEUE_NAME=user.events

# External Services
AUTH_SERVICE_URL=http://auth-service:8081
NOTIFICATION_SERVICE_URL=http://notification-service:8082
USER_CACHE_REDIS_URL=redis://cache.internal:6379

# Observability
LOG_LEVEL=info
METRICS_PORT=9090
TRACING_ENABLED=true`,
			exampleContent: `# Service Configuration
SERVICE_NAME=example-service
SERVICE_VERSION=0.1.0
PORT=3000
HEALTH_CHECK_PATH=/ping
GRACEFUL_SHUTDOWN_TIMEOUT=30s

# Database
DB_HOST=localhost
DB_PORT=5432
DB_NAME=example_db
DB_USER=postgres
DB_PASSWORD=password
DB_SSL_MODE=disable

# Message Queue
RABBITMQ_URL=amqp://guest:guest@localhost:5672
QUEUE_NAME=example.events
DEAD_LETTER_QUEUE=example.dlq

# External Services
AUTH_SERVICE_URL=http://localhost:8081
EMAIL_SERVICE_URL=http://localhost:8083

# Observability
LOG_LEVEL=debug
METRICS_PORT=9090
JAEGER_ENDPOINT=http://localhost:14268/api/traces`,
			expectedMissing: 5, // GRACEFUL_SHUTDOWN_TIMEOUT, DB_SSL_MODE, DEAD_LETTER_QUEUE, EMAIL_SERVICE_URL, JAEGER_ENDPOINT
			expectedExtra:   3, // NOTIFICATION_SERVICE_URL, USER_CACHE_REDIS_URL, TRACING_ENABLED
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			envFile := filepath.Join(tmpDir, ".env")
			exampleFile := filepath.Join(tmpDir, ".env.example")

			err := os.WriteFile(envFile, []byte(tt.envContent), 0644)
			if err != nil {
				t.Fatalf("Failed to create .env file: %v", err)
			}

			err = os.WriteFile(exampleFile, []byte(tt.exampleContent), 0644)
			if err != nil {
				t.Fatalf("Failed to create .env.example file: %v", err)
			}

			result, err := CompareEnvFiles(envFile, exampleFile)
			if err != nil {
				t.Fatalf("CompareEnvFiles failed: %v", err)
			}

			if len(result.Missing) != tt.expectedMissing {
				t.Errorf("Expected %d missing variables, got %d: %v",
					tt.expectedMissing, len(result.Missing), result.Missing)
			}

			if len(result.Extra) != tt.expectedExtra {
				t.Errorf("Expected %d extra variables, got %d: %v",
					tt.expectedExtra, len(result.Extra), result.Extra)
			}

			// Verify HasIssues works correctly
			expectedHasIssues := tt.expectedMissing > 0 || tt.expectedExtra > 0
			if result.HasIssues() != expectedHasIssues {
				t.Errorf("HasIssues() = %v, want %v", result.HasIssues(), expectedHasIssues)
			}
		})
	}
}

// Helper function to compare DiffResult structs
func compareDiffResults(a, b *DiffResult) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}

	// Sort slices for comparison since order might vary
	aMissingCopy := make([]string, len(a.Missing))
	copy(aMissingCopy, a.Missing)
	sort.Strings(aMissingCopy)

	bMissingCopy := make([]string, len(b.Missing))
	copy(bMissingCopy, b.Missing)
	sort.Strings(bMissingCopy)

	aExtraCopy := make([]string, len(a.Extra))
	copy(aExtraCopy, a.Extra)
	sort.Strings(aExtraCopy)

	bExtraCopy := make([]string, len(b.Extra))
	copy(bExtraCopy, b.Extra)
	sort.Strings(bExtraCopy)

	return reflect.DeepEqual(aMissingCopy, bMissingCopy) &&
		reflect.DeepEqual(aExtraCopy, bExtraCopy)
}

// Benchmark tests
func BenchmarkCompareEnvVars_Small(b *testing.B) {
	env := parser.EnvVars{
		"DATABASE_URL": "postgres://localhost",
		"API_KEY":      "secret",
		"PORT":         "3000",
	}
	example := parser.EnvVars{
		"DATABASE_URL": "postgres://example",
		"API_KEY":      "example",
		"REDIS_URL":    "redis://localhost",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		CompareEnvVars(env, example)
	}
}

func BenchmarkCompareEnvVars_Large(b *testing.B) {
	// Create large env maps
	env := make(parser.EnvVars)
	example := make(parser.EnvVars)

	for i := 0; i < 1000; i++ {
		if i%2 == 0 {
			env[fmt.Sprintf("VAR_%d", i)] = fmt.Sprintf("value_%d", i)
		}
		if i%3 == 0 {
			example[fmt.Sprintf("VAR_%d", i)] = fmt.Sprintf("example_%d", i)
		}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		CompareEnvVars(env, example)
	}
}

func BenchmarkCompareEnvFiles(b *testing.B) {
	// Create test files
	envContent := `DATABASE_URL=postgres://localhost/myapp
API_KEY=secret123
PORT=3000
DEBUG=true
LOG_LEVEL=info`

	exampleContent := `DATABASE_URL=postgres://localhost/example
API_KEY=example_key
PORT=8080
REDIS_URL=redis://localhost`

	tmpDir := b.TempDir()
	envFile := filepath.Join(tmpDir, ".env")
	exampleFile := filepath.Join(tmpDir, ".env.example")

	os.WriteFile(envFile, []byte(envContent), 0644)
	os.WriteFile(exampleFile, []byte(exampleContent), 0644)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := CompareEnvFiles(envFile, exampleFile)
		if err != nil {
			b.Fatalf("Benchmark failed: %v", err)
		}
	}
}
