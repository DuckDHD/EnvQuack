package parser

import (
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"testing"
)

func TestParseComposeFile(t *testing.T) {
	tests := []struct {
		name        string
		content     string
		expected    *ComposeEnvInfo
		expectError bool
	}{
		{
			name: "basic compose with environment array",
			content: `version: '3.8'
services:
  web:
    environment:
      - NODE_ENV=production
      - API_KEY=secret123
      - PORT=3000`,
			expected: &ComposeEnvInfo{
				Variables: EnvVars{
					"NODE_ENV": "production",
					"API_KEY":  "secret123",
					"PORT":     "3000",
				},
				ServiceVars: map[string]EnvVars{
					"web": {
						"NODE_ENV": "production",
						"API_KEY":  "secret123",
						"PORT":     "3000",
					},
				},
				EnvFiles:     []string{},
				VariableRefs: []string{},
			},
			expectError: false,
		},
		{
			name: "environment object format",
			content: `version: '3.8'
services:
  app:
    environment:
      DATABASE_URL: postgres://localhost/myapp
      DEBUG: true
      PORT: 8080`,
			expected: &ComposeEnvInfo{
				Variables: EnvVars{
					"DATABASE_URL": "postgres://localhost/myapp",
					"DEBUG":        "true",
					"PORT":         "8080",
				},
				ServiceVars: map[string]EnvVars{
					"app": {
						"DATABASE_URL": "postgres://localhost/myapp",
						"DEBUG":        "true",
						"PORT":         "8080",
					},
				},
				EnvFiles:     []string{},
				VariableRefs: []string{},
			},
			expectError: false,
		},
		{
			name: "variable references ${VAR} format",
			content: `version: '3.8'
services:
  web:
    environment:
      - DATABASE_URL=${DATABASE_URL}
      - API_SECRET=${API_SECRET:-default_secret}
    ports:
      - "${WEB_PORT:-8080}:8080"`,
			expected: &ComposeEnvInfo{
				Variables: EnvVars{
					"DATABASE_URL": "${DATABASE_URL}",
				},
				ServiceVars: map[string]EnvVars{
					"web": {
						"DATABASE_URL": "${DATABASE_URL}",
					},
				},
				EnvFiles:     []string{},
				VariableRefs: []string{"API_SECRET", "DATABASE_URL", "WEB_PORT"},
			},
			expectError: false,
		},
		{
			name: "variable references $VAR format",
			content: `version: '3.8'
services:
  worker:
    image: worker:$VERSION
    environment:
      REDIS_URL: $REDIS_URL
      NODE_ENV: $NODE_ENV`,
			expected: &ComposeEnvInfo{
				Variables: EnvVars{
					"REDIS_URL": "$REDIS_URL",
					"NODE_ENV":  "$NODE_ENV",
				},
				ServiceVars: map[string]EnvVars{
					"worker": {
						"REDIS_URL": "$REDIS_URL",
						"NODE_ENV":  "$NODE_ENV",
					},
				},
				EnvFiles:     []string{},
				VariableRefs: []string{"NODE_ENV", "REDIS_URL", "VERSION"},
			},
			expectError: false,
		},
		{
			name: "env_file single file",
			content: `version: '3.8'
services:
  app:
    env_file: .env
    environment:
      - OVERRIDE=value`,
			expected: &ComposeEnvInfo{
				Variables: EnvVars{
					"OVERRIDE": "value",
				},
				ServiceVars: map[string]EnvVars{
					"app": {
						"OVERRIDE": "value",
					},
				},
				EnvFiles:     []string{".env"},
				VariableRefs: []string{},
			},
			expectError: false,
		},
		{
			name: "env_file array format",
			content: `version: '3.8'
services:
  web:
    env_file:
      - .env
      - .env.local
      - config/.env.production`,
			expected: &ComposeEnvInfo{
				Variables:    EnvVars{},
				ServiceVars:  map[string]EnvVars{},
				EnvFiles:     []string{".env", ".env.local", "config/.env.production"},
				VariableRefs: []string{},
			},
			expectError: false,
		},
		{
			name: "multiple services",
			content: `version: '3.8'
services:
  web:
    environment:
      - WEB_PORT=3000
      - API_URL=${API_URL}
  worker:
    environment:
      WORKER_CONCURRENCY: 5
      REDIS_URL: ${REDIS_URL}
  db:
    environment:
      POSTGRES_DB: myapp
      POSTGRES_PASSWORD: ${DB_PASSWORD}`,
			expected: &ComposeEnvInfo{
				Variables: EnvVars{
					"WEB_PORT":           "3000",
					"API_URL":            "${API_URL}",
					"WORKER_CONCURRENCY": "5",
					"REDIS_URL":          "${REDIS_URL}",
					"POSTGRES_DB":        "myapp",
					"POSTGRES_PASSWORD":  "${DB_PASSWORD}",
				},
				ServiceVars: map[string]EnvVars{
					"web": {
						"WEB_PORT": "3000",
						"API_URL":  "${API_URL}",
					},
					"worker": {
						"WORKER_CONCURRENCY": "5",
						"REDIS_URL":          "${REDIS_URL}",
					},
					"db": {
						"POSTGRES_DB":       "myapp",
						"POSTGRES_PASSWORD": "${DB_PASSWORD}",
					},
				},
				EnvFiles:     []string{},
				VariableRefs: []string{"API_URL", "DB_PASSWORD", "REDIS_URL"},
			},
			expectError: false,
		},
		{
			name: "empty values and null values",
			content: `version: '3.8'
services:
  app:
    environment:
      - EMPTY_VAR=
      - NULL_VAR
      OBJECT_EMPTY: ""
      OBJECT_NULL:`,
			expected: &ComposeEnvInfo{
				Variables: EnvVars{
					"EMPTY_VAR":    "",
					"NULL_VAR":     "",
					"OBJECT_EMPTY": "",
					"OBJECT_NULL":  "",
				},
				ServiceVars: map[string]EnvVars{
					"app": {
						"EMPTY_VAR":    "",
						"NULL_VAR":     "",
						"OBJECT_EMPTY": "",
						"OBJECT_NULL":  "",
					},
				},
				EnvFiles:     []string{},
				VariableRefs: []string{},
			},
			expectError: false,
		},
		{
			name: "docker internal variables filtered",
			content: `version: '3.8'
services:
  app:
    environment:
      - COMPOSE_PROJECT_NAME=myproject
      - DOCKER_HOST=unix:///var/run/docker.sock
      - MY_VAR=${MY_VAR}
      - HOME=/home/user
    image: app:${VERSION}`,
			expected: &ComposeEnvInfo{
				Variables: EnvVars{
					"COMPOSE_PROJECT_NAME": "myproject",
					"DOCKER_HOST":          "unix:///var/run/docker.sock",
					"MY_VAR":               "${MY_VAR}",
					"HOME":                 "/home/user",
				},
				ServiceVars: map[string]EnvVars{
					"app": {
						"COMPOSE_PROJECT_NAME": "myproject",
						"DOCKER_HOST":          "unix:///var/run/docker.sock",
						"MY_VAR":               "${MY_VAR}",
						"HOME":                 "/home/user",
					},
				},
				EnvFiles:     []string{},
				VariableRefs: []string{"MY_VAR", "VERSION"}, // Internal vars filtered out
			},
			expectError: false,
		},
		{
			name: "complex real-world example",
			content: `version: '3.8'
services:
  web:
    build:
      context: .
      args:
        - NODE_ENV=${NODE_ENV:-production}
    environment:
      - NODE_ENV=${NODE_ENV:-production}
      - PORT=${WEB_PORT:-3000}
      - DATABASE_URL=${DATABASE_URL}
      - REDIS_URL=${REDIS_URL}
      - JWT_SECRET=${JWT_SECRET}
    env_file:
      - .env
      - .env.local
    ports:
      - "${WEB_PORT:-3000}:3000"
    depends_on:
      - db
      - redis

  worker:
    build:
      context: .
      dockerfile: Dockerfile.worker
    environment:
      NODE_ENV: ${NODE_ENV:-production}
      DATABASE_URL: ${DATABASE_URL}
      REDIS_URL: ${REDIS_URL}
      WORKER_CONCURRENCY: ${WORKER_CONCURRENCY:-1}
    env_file: .env

  db:
    image: postgres:13
    environment:
      POSTGRES_DB: ${DB_NAME:-myapp}
      POSTGRES_USER: ${DB_USER:-postgres}
      POSTGRES_PASSWORD: ${DB_PASSWORD}
    volumes:
      - postgres_data:/var/lib/postgresql/data

  redis:
    image: redis:6-alpine

volumes:
  postgres_data:`,
			expected: &ComposeEnvInfo{
				Variables: EnvVars{
					"NODE_ENV":           "${NODE_ENV:-production}",
					"PORT":               "${WEB_PORT:-3000}",
					"DATABASE_URL":       "${DATABASE_URL}",
					"REDIS_URL":          "${REDIS_URL}",
					"JWT_SECRET":         "${JWT_SECRET}",
					"POSTGRES_DB":        "${DB_NAME:-myapp}",
					"POSTGRES_USER":      "${DB_USER:-postgres}",
					"POSTGRES_PASSWORD":  "${DB_PASSWORD}",
					"WORKER_CONCURRENCY": "${WORKER_CONCURRENCY:-1}",
				},
				ServiceVars: map[string]EnvVars{
					"web": {
						"NODE_ENV":     "${NODE_ENV:-production}",
						"PORT":         "${WEB_PORT:-3000}",
						"DATABASE_URL": "${DATABASE_URL}",
						"REDIS_URL":    "${REDIS_URL}",
						"JWT_SECRET":   "${JWT_SECRET}",
					},
					"worker": {
						"NODE_ENV":           "${NODE_ENV:-production}",
						"DATABASE_URL":       "${DATABASE_URL}",
						"REDIS_URL":          "${REDIS_URL}",
						"WORKER_CONCURRENCY": "${WORKER_CONCURRENCY:-1}",
					},
					"db": {
						"POSTGRES_DB":       "${DB_NAME:-myapp}",
						"POSTGRES_USER":     "${DB_USER:-postgres}",
						"POSTGRES_PASSWORD": "${DB_PASSWORD}",
					},
				},
				EnvFiles: []string{".env", ".env.local"},
				VariableRefs: []string{
					"DATABASE_URL", "DB_NAME", "DB_PASSWORD", "DB_USER",
					"JWT_SECRET", "NODE_ENV", "REDIS_URL", "WEB_PORT", "WORKER_CONCURRENCY",
				},
			},
			expectError: false,
		},
		{
			name:        "empty compose file",
			content:     "",
			expected:    nil,
			expectError: true,
		},
		{
			name: "minimal valid compose",
			content: `version: '3.8'
services: {}`,
			expected: &ComposeEnvInfo{
				Variables:    EnvVars{},
				ServiceVars:  map[string]EnvVars{},
				EnvFiles:     []string{},
				VariableRefs: []string{},
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temporary file
			tmpDir := t.TempDir()
			tmpFile := filepath.Join(tmpDir, "docker-compose.yml")

			err := os.WriteFile(tmpFile, []byte(tt.content), 0644)
			if err != nil {
				t.Fatalf("Failed to create test file: %v", err)
			}

			// Parse the file
			result, err := ParseComposeFile(tmpFile)

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
			if !compareComposeEnvInfo(result, tt.expected) {
				t.Errorf("ParseComposeFile() result mismatch")
				t.Errorf("Got Variables: %v", result.Variables)
				t.Errorf("Want Variables: %v", tt.expected.Variables)
				t.Errorf("Got ServiceVars: %v", result.ServiceVars)
				t.Errorf("Want ServiceVars: %v", tt.expected.ServiceVars)
				t.Errorf("Got EnvFiles: %v", result.EnvFiles)
				t.Errorf("Want EnvFiles: %v", tt.expected.EnvFiles)
				t.Errorf("Got VariableRefs: %v", result.VariableRefs)
				t.Errorf("Want VariableRefs: %v", tt.expected.VariableRefs)
			}
		})
	}
}

func TestParseComposeData(t *testing.T) {
	tests := []struct {
		name        string
		data        []byte
		expectError bool
	}{
		{
			name: "valid YAML",
			data: []byte(`version: '3.8'
services:
  app:
    environment:
      - TEST=value`),
			expectError: false,
		},
		{
			name:        "invalid YAML - malformed",
			data:        []byte(`invalid: yaml: content: [unclosed bracket`),
			expectError: true,
		},
		{
			name:        "invalid YAML - tabs and spaces mixed",
			data:        []byte("version: '3.8'\nservices:\n\tapp:\n  environment:\n    - TEST=value"),
			expectError: true,
		},
		{
			name:        "empty data",
			data:        []byte(""),
			expectError: true,
		},
		{
			name:        "only comments",
			data:        []byte("# This is just a comment"),
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ParseComposeData(tt.data)

			if tt.expectError && err == nil {
				t.Errorf("Expected error but got none. Result: %v", result)
			}
			if !tt.expectError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
		})
	}
}

func TestParseComposeFile_FileErrors(t *testing.T) {
	tests := []struct {
		name     string
		filename string
		setup    func(string) error
	}{
		{
			name:     "non-existent file",
			filename: "/non/existent/docker-compose.yml",
		},
		{
			name:     "directory instead of file",
			filename: "directory-compose",
			setup: func(filename string) error {
				return os.Mkdir(filename, 0755)
			},
		},
		{
			name:     "permission denied",
			filename: "permission_denied_compose.yml",
			setup: func(filename string) error {
				content := `version: '3.8'
services:
  app:
    environment:
      - TEST=value`
				if err := os.WriteFile(filename, []byte(content), 0644); err != nil {
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

			if tt.setup != nil {
				if err := tt.setup(testFile); err != nil {
					t.Fatalf("Setup failed: %v", err)
				}
				defer func() {
					os.Chmod(testFile, 0644)
					os.Remove(testFile)
				}()
			}

			result, err := ParseComposeFile(testFile)

			if err == nil {
				t.Errorf("Expected error for %s, but got none. Result: %v", tt.name, result)
			}

			if result != nil {
				t.Errorf("Expected nil result on error, got: %v", result)
			}
		})
	}
}

func TestParseEnvironmentSection(t *testing.T) {
	tests := []struct {
		name     string
		input    interface{}
		expected EnvVars
	}{
		{
			name:     "nil input",
			input:    nil,
			expected: EnvVars{},
		},
		{
			name: "array format with key=value",
			input: []interface{}{
				"NODE_ENV=production",
				"PORT=3000",
				"DEBUG=true",
			},
			expected: EnvVars{
				"NODE_ENV": "production",
				"PORT":     "3000",
				"DEBUG":    "true",
			},
		},
		{
			name: "array format with key only",
			input: []interface{}{
				"NODE_ENV=production",
				"DEBUG",
				"PORT=3000",
			},
			expected: EnvVars{
				"NODE_ENV": "production",
				"DEBUG":    "",
				"PORT":     "3000",
			},
		},
		{
			name: "object format",
			input: map[string]interface{}{
				"NODE_ENV": "production",
				"PORT":     3000,
				"DEBUG":    true,
				"EMPTY":    nil,
			},
			expected: EnvVars{
				"NODE_ENV": "production",
				"PORT":     "3000",
				"DEBUG":    "true",
				"EMPTY":    "",
			},
		},
		{
			name: "array with non-string values (should be ignored)",
			input: []interface{}{
				"VALID=value",
				123,
				true,
				"ANOTHER=valid",
			},
			expected: EnvVars{
				"VALID":   "value",
				"ANOTHER": "valid",
			},
		},
		{
			name:     "empty array",
			input:    []interface{}{},
			expected: EnvVars{},
		},
		{
			name:     "empty object",
			input:    map[string]interface{}{},
			expected: EnvVars{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parseEnvironmentSection(tt.input)
			if !reflect.DeepEqual(result, tt.expected) {
				t.Errorf("parseEnvironmentSection() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestParseEnvFileSection(t *testing.T) {
	tests := []struct {
		name     string
		input    interface{}
		expected []string
	}{
		{
			name:     "nil input",
			input:    nil,
			expected: []string{},
		},
		{
			name:     "single file string",
			input:    ".env",
			expected: []string{".env"},
		},
		{
			name: "array of files",
			input: []interface{}{
				".env",
				".env.local",
				"config/.env.production",
			},
			expected: []string{".env", ".env.local", "config/.env.production"},
		},
		{
			name: "array with non-string values (should be ignored)",
			input: []interface{}{
				".env",
				123,
				".env.local",
				true,
			},
			expected: []string{".env", ".env.local"},
		},
		{
			name:     "empty array",
			input:    []interface{}{},
			expected: []string{},
		},
		{
			name:     "empty string",
			input:    "",
			expected: []string{""},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parseEnvFileSection(tt.input)
			if !reflect.DeepEqual(result, tt.expected) {
				t.Errorf("parseEnvFileSection() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestParseEnvString(t *testing.T) {
	tests := []struct {
		name          string
		input         string
		expectedKey   string
		expectedValue string
	}{
		{
			name:          "key=value format",
			input:         "DATABASE_URL=postgres://localhost",
			expectedKey:   "DATABASE_URL",
			expectedValue: "postgres://localhost",
		},
		{
			name:          "key only format",
			input:         "DEBUG",
			expectedKey:   "DEBUG",
			expectedValue: "",
		},
		{
			name:          "empty value",
			input:         "EMPTY=",
			expectedKey:   "EMPTY",
			expectedValue: "",
		},
		{
			name:          "value with equals",
			input:         "EQUATION=x=y+z",
			expectedKey:   "EQUATION",
			expectedValue: "x=y+z",
		},
		{
			name:          "whitespace trimming",
			input:         "  KEY  =  value  ",
			expectedKey:   "KEY",
			expectedValue: "value",
		},
		{
			name:          "empty string",
			input:         "",
			expectedKey:   "",
			expectedValue: "",
		},
		{
			name:          "whitespace only",
			input:         "   ",
			expectedKey:   "",
			expectedValue: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			key, value := parseEnvString(tt.input)
			if key != tt.expectedKey || value != tt.expectedValue {
				t.Errorf("parseEnvString(%q) = (%q, %q), want (%q, %q)",
					tt.input, key, value, tt.expectedKey, tt.expectedValue)
			}
		})
	}
}

func TestExtractVariableReferences(t *testing.T) {
	tests := []struct {
		name     string
		content  string
		expected []string
	}{
		{
			name: "standard ${VAR} format",
			content: `version: '3.8'
services:
  app:
    environment:
      - DATABASE_URL=${DATABASE_URL}
      - API_KEY=${API_SECRET}`,
			expected: []string{"API_SECRET", "DATABASE_URL"},
		},
		{
			name: "with default values",
			content: `version: '3.8'
services:
  app:
    ports:
      - "${WEB_PORT:-3000}:3000"
    environment:
      - NODE_ENV=${NODE_ENV:-production}`,
			expected: []string{"NODE_ENV", "WEB_PORT"},
		},
		{
			name: "shell style $VAR format",
			content: `version: '3.8'
services:
  app:
    image: app:$VERSION
    environment:
      - REDIS_URL=$REDIS_URL`,
			expected: []string{"REDIS_URL", "VERSION"},
		},
		{
			name: "mixed formats",
			content: `version: '3.8'
services:
  app:
    image: app:$VERSION
    environment:
      - DATABASE_URL=${DATABASE_URL}
      - PORT=${WEB_PORT:-3000}
      - DEBUG=$DEBUG_MODE`,
			expected: []string{"DATABASE_URL", "DEBUG_MODE", "VERSION", "WEB_PORT"},
		},
		{
			name: "docker internal variables filtered",
			content: `version: '3.8'
services:
  app:
    environment:
      - COMPOSE_PROJECT_NAME=${COMPOSE_PROJECT_NAME}
      - DOCKER_HOST=${DOCKER_HOST}
      - MY_VAR=${MY_VAR}
      - PATH=${PATH}`,
			expected: []string{"MY_VAR"}, // Internal variables filtered out
		},
		{
			name: "no variables",
			content: `version: '3.8'
services:
  app:
    environment:
      - STATIC=value
      - ANOTHER=constant`,
			expected: []string{},
		},
		{
			name:     "empty content",
			content:  "",
			expected: []string{},
		},
		{
			name: "invalid variable names ignored",
			content: `version: '3.8'
services:
  app:
    environment:
      - VALID_VAR=${VALID_VAR}
      - INVALID=${123INVALID}
      - ANOTHER_VALID=${ANOTHER_VALID}`,
			expected: []string{"ANOTHER_VALID", "VALID_VAR"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := extractVariableReferences(tt.content)
			sort.Strings(result)
			sort.Strings(tt.expected)

			if !reflect.DeepEqual(result, tt.expected) {
				t.Errorf("extractVariableReferences() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestIsDockerInternalVar(t *testing.T) {
	tests := []struct {
		name     string
		varName  string
		expected bool
	}{
		{
			name:     "COMPOSE_PROJECT_NAME",
			varName:  "COMPOSE_PROJECT_NAME",
			expected: true,
		},
		{
			name:     "DOCKER_HOST",
			varName:  "DOCKER_HOST",
			expected: true,
		},
		{
			name:     "PATH",
			varName:  "PATH",
			expected: true,
		},
		{
			name:     "HOME",
			varName:  "HOME",
			expected: true,
		},
		{
			name:     "custom variable",
			varName:  "MY_CUSTOM_VAR",
			expected: false,
		},
		{
			name:     "DATABASE_URL",
			varName:  "DATABASE_URL",
			expected: false,
		},
		{
			name:     "empty string",
			varName:  "",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isDockerInternalVar(tt.varName)
			if result != tt.expected {
				t.Errorf("isDockerInternalVar(%q) = %v, want %v", tt.varName, result, tt.expected)
			}
		})
	}
}

func TestComposeEnvInfo_Methods(t *testing.T) {
	info := &ComposeEnvInfo{
		Variables: EnvVars{
			"DATABASE_URL": "postgres://localhost",
			"API_KEY":      "secret",
		},
		ServiceVars: map[string]EnvVars{
			"web": {
				"DATABASE_URL": "postgres://localhost",
				"WEB_PORT":     "3000",
			},
			"worker": {
				"DATABASE_URL": "postgres://localhost",
				"QUEUE_NAME":   "default",
			},
		},
		EnvFiles:     []string{".env", ".env.local"},
		VariableRefs: []string{"REDIS_URL", "SECRET_KEY"},
	}

	t.Run("GetAllEnvVars", func(t *testing.T) {
		result := info.GetAllEnvVars()
		expected := []string{"API_KEY", "DATABASE_URL", "REDIS_URL", "SECRET_KEY"}

		sort.Strings(result)
		if !reflect.DeepEqual(result, expected) {
			t.Errorf("GetAllEnvVars() = %v, want %v", result, expected)
		}
	})

	t.Run("GetServiceVars existing service", func(t *testing.T) {
		result := info.GetServiceVars("web")
		expected := EnvVars{
			"DATABASE_URL": "postgres://localhost",
			"WEB_PORT":     "3000",
		}

		if !reflect.DeepEqual(result, expected) {
			t.Errorf("GetServiceVars('web') = %v, want %v", result, expected)
		}
	})

	t.Run("GetServiceVars non-existing service", func(t *testing.T) {
		result := info.GetServiceVars("non-existing")
		expected := make(EnvVars)

		if !reflect.DeepEqual(result, expected) {
			t.Errorf("GetServiceVars('non-existing') = %v, want %v", result, expected)
		}
	})

	t.Run("HasService", func(t *testing.T) {
		tests := []struct {
			service  string
			expected bool
		}{
			{"web", true},
			{"worker", true},
			{"non-existing", false},
			{"", false},
		}

		for _, tt := range tests {
			result := info.HasService(tt.service)
			if result != tt.expected {
				t.Errorf("HasService(%q) = %v, want %v", tt.service, result, tt.expected)
			}
		}
	})

	t.Run("GetServices", func(t *testing.T) {
		result := info.GetServices()
		expected := []string{"web", "worker"}

		sort.Strings(result)
		if !reflect.DeepEqual(result, expected) {
			t.Errorf("GetServices() = %v, want %v", result, expected)
		}
	})
}

func TestRemoveDuplicates(t *testing.T) {
	tests := []struct {
		name     string
		input    []string
		expected []string
	}{
		{
			name:     "no duplicates",
			input:    []string{"a", "b", "c"},
			expected: []string{"a", "b", "c"},
		},
		{
			name:     "with duplicates",
			input:    []string{"a", "b", "a", "c", "b"},
			expected: []string{"a", "b", "c"},
		},
		{
			name:     "empty slice",
			input:    []string{},
			expected: []string{},
		},
		{
			name:     "all same",
			input:    []string{"a", "a", "a"},
			expected: []string{"a"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := removeDuplicates(tt.input)

			// Sort both for comparison since order might vary
			sort.Strings(result)
			sort.Strings(tt.expected)

			if !reflect.DeepEqual(result, tt.expected) {
				t.Errorf("removeDuplicates() = %v, want %v", result, tt.expected)
			}
		})
	}
}

// Helper function to compare ComposeEnvInfo structs
func compareComposeEnvInfo(a, b *ComposeEnvInfo) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}

	// Compare Variables
	if !reflect.DeepEqual(a.Variables, b.Variables) {
		return false
	}

	// Compare ServiceVars
	if !reflect.DeepEqual(a.ServiceVars, b.ServiceVars) {
		return false
	}

	// Compare EnvFiles (order matters after sorting)
	aCopy := make([]string, len(a.EnvFiles))
	copy(aCopy, a.EnvFiles)
	sort.Strings(aCopy)

	bCopy := make([]string, len(b.EnvFiles))
	copy(bCopy, b.EnvFiles)
	sort.Strings(bCopy)

	if !reflect.DeepEqual(aCopy, bCopy) {
		return false
	}

	// Compare VariableRefs (order matters after sorting)
	aRefsCopy := make([]string, len(a.VariableRefs))
	copy(aRefsCopy, a.VariableRefs)
	sort.Strings(aRefsCopy)

	bRefsCopy := make([]string, len(b.VariableRefs))
	copy(bRefsCopy, b.VariableRefs)
	sort.Strings(bRefsCopy)

	return reflect.DeepEqual(aRefsCopy, bRefsCopy)
}

// Benchmark tests
func BenchmarkParseComposeFile(b *testing.B) {
	content := `version: '3.8'
services:
  web:
    environment:
      - NODE_ENV=${NODE_ENV:-production}
      - DATABASE_URL=${DATABASE_URL}
      - PORT=${WEB_PORT:-3000}
    env_file: .env
  worker:
    environment:
      - NODE_ENV=${NODE_ENV:-production}
      - REDIS_URL=${REDIS_URL}
    env_file: .env`

	tmpDir := b.TempDir()
	tmpFile := filepath.Join(tmpDir, "docker-compose.yml")
	os.WriteFile(tmpFile, []byte(content), 0644)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := ParseComposeFile(tmpFile)
		if err != nil {
			b.Fatalf("Benchmark failed: %v", err)
		}
	}
}

func BenchmarkExtractVariableReferences(b *testing.B) {
	content := `version: '3.8'
services:
  web:
    environment:
      - NODE_ENV=${NODE_ENV:-production}
      - DATABASE_URL=${DATABASE_URL}
      - PORT=${WEB_PORT:-3000}
      - API_KEY=$API_KEY
    ports:
      - "${WEB_PORT:-3000}:3000"`

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		extractVariableReferences(content)
	}
}
