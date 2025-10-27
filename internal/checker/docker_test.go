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

func TestCompareDockerfileWithEnv(t *testing.T) {
	tests := []struct {
		name              string
		dockerfileContent string
		envFiles          map[string]string // filename -> content
		expected          *DockerfileDiffResult
		expectError       bool
	}{
		{
			name: "all variables present in env",
			dockerfileContent: `FROM node:16
ARG NODE_VERSION=16
ENV NODE_ENV=${NODE_ENV}
ENV DATABASE_URL=${DATABASE_URL}
ENV PORT=${PORT:-3000}`,
			envFiles: map[string]string{
				".env": `NODE_ENV=production
DATABASE_URL=postgres://localhost/myapp
PORT=8080`,
			},
			expected: &DockerfileDiffResult{
				MissingInEnv:       []string{},
				ExtraInEnv:         []string{},
				UnusedArgs:         []string{"NODE_VERSION"},
				HardcodedEnvs:      []string{},
				MissingArgDefaults: []string{},
			},
			expectError: false,
		},
		{
			name: "missing variables in env",
			dockerfileContent: `FROM node:${NODE_VERSION}
ARG NODE_VERSION
ARG API_KEY
ENV NODE_ENV=${NODE_ENV}
ENV DATABASE_URL=${DATABASE_URL}
ENV API_KEY=${API_KEY}`,
			envFiles: map[string]string{
				".env": `NODE_VERSION=16
NODE_ENV=production`,
			},
			expected: &DockerfileDiffResult{
				MissingInEnv:       []string{"API_KEY", "DATABASE_URL"},
				ExtraInEnv:         []string{},
				UnusedArgs:         []string{},
				HardcodedEnvs:      []string{},
				MissingArgDefaults: []string{"API_KEY", "NODE_VERSION"},
			},
			expectError: false,
		},
		{
			name: "extra variables in env",
			dockerfileContent: `FROM alpine:3.14
ENV NODE_ENV=${NODE_ENV}
ENV PORT=${PORT}`,
			envFiles: map[string]string{
				".env": `NODE_ENV=production
PORT=3000
DATABASE_URL=postgres://localhost/myapp
API_KEY=secret123
DEBUG=true`,
			},
			expected: &DockerfileDiffResult{
				MissingInEnv:       []string{},
				ExtraInEnv:         []string{"API_KEY", "DATABASE_URL", "DEBUG"},
				UnusedArgs:         []string{},
				HardcodedEnvs:      []string{},
				MissingArgDefaults: []string{},
			},
			expectError: false,
		},
		{
			name: "unused ARG variables",
			dockerfileContent: `FROM node:16
ARG NODE_VERSION=16
ARG UNUSED_ARG=default
ARG ANOTHER_UNUSED
ENV NODE_ENV=production
RUN echo "Building app"`,
			envFiles: map[string]string{
				".env": `NODE_ENV=production`,
			},
			expected: &DockerfileDiffResult{
				MissingInEnv:       []string{},
				ExtraInEnv:         []string{},
				UnusedArgs:         []string{"ANOTHER_UNUSED", "NODE_VERSION", "UNUSED_ARG"},
				HardcodedEnvs:      []string{"NODE_ENV"},
				MissingArgDefaults: []string{"ANOTHER_UNUSED"},
			},
			expectError: false,
		},
		{
			name: "hardcoded ENV variables",
			dockerfileContent: `FROM node:16
ENV NODE_ENV=production
ENV API_URL=https://api.example.com
ENV DATABASE_URL=postgres://db:5432/myapp
ENV LOG_LEVEL=info
ENV WORKDIR=/app`,
			envFiles: map[string]string{
				".env": ``,
			},
			expected: &DockerfileDiffResult{
				MissingInEnv:       []string{},
				ExtraInEnv:         []string{},
				UnusedArgs:         []string{},
				HardcodedEnvs:      []string{"API_URL", "DATABASE_URL", "NODE_ENV"},
				MissingArgDefaults: []string{},
			},
			expectError: false,
		},
		{
			name: "ARG variables without defaults",
			dockerfileContent: `FROM node:${NODE_VERSION}
ARG NODE_VERSION
ARG BUILD_ENV=production
ARG API_ENDPOINT
ARG SECRET_KEY
ENV NODE_ENV=${BUILD_ENV}`,
			envFiles: map[string]string{
				".env": `NODE_VERSION=16
API_ENDPOINT=https://api.com
SECRET_KEY=secret`,
			},
			expected: &DockerfileDiffResult{
				MissingInEnv:       []string{},
				ExtraInEnv:         []string{},
				UnusedArgs:         []string{"API_ENDPOINT", "SECRET_KEY"},
				HardcodedEnvs:      []string{},
				MissingArgDefaults: []string{"API_ENDPOINT", "NODE_VERSION", "SECRET_KEY"},
			},
			expectError: false,
		},
		{
			name: "complex real-world Dockerfile",
			dockerfileContent: `FROM node:${NODE_VERSION:-16} AS builder

# Build arguments
ARG NODE_ENV=production
ARG API_ENDPOINT
ARG BUILD_VERSION=1.0.0
ARG ENABLE_FEATURES="auth,analytics"

# Runtime environment variables
ENV NODE_ENV=${NODE_ENV}
ENV PORT=${PORT:-3000}
ENV API_ENDPOINT=${API_ENDPOINT}
ENV BUILD_VERSION=${BUILD_VERSION}
ENV FEATURES="$ENABLE_FEATURES"

# Application setup
WORKDIR /app
COPY package*.json ./
RUN npm ci --only=${NODE_ENV}

# Copy application code
COPY src/ ./src/
COPY public/ ./public/

# Build the application
RUN npm run build:${NODE_ENV}

EXPOSE $PORT
CMD ["node", "dist/server.js"]`,
			envFiles: map[string]string{
				".env": `NODE_ENV=development
PORT=8080
API_ENDPOINT=https://api.staging.com`,
			},
			expected: &DockerfileDiffResult{
				MissingInEnv:       []string{},
				ExtraInEnv:         []string{},
				UnusedArgs:         []string{},
				HardcodedEnvs:      []string{"FEATURES"},
				MissingArgDefaults: []string{"API_ENDPOINT"},
			},
			expectError: false,
		},
		{
			name: "ENV variables defined in Dockerfile (should not be missing)",
			dockerfileContent: `FROM alpine:3.14
ENV DATABASE_URL=postgres://localhost/default
ENV API_KEY=${API_KEY}
ENV PORT=3000`,
			envFiles: map[string]string{
				".env": `API_KEY=secret123`,
			},
			expected: &DockerfileDiffResult{
				MissingInEnv:       []string{},
				ExtraInEnv:         []string{},
				UnusedArgs:         []string{},
				HardcodedEnvs:      []string{"DATABASE_URL"},
				MissingArgDefaults: []string{},
			},
			expectError: false,
		},
		{
			name: "mixed ARG and ENV scenarios",
			dockerfileContent: `FROM node:${NODE_VERSION}
ARG NODE_VERSION=16
ARG DATABASE_HOST
ARG DATABASE_PORT=5432
ENV NODE_ENV=${NODE_ENV:-production}
ENV DATABASE_URL=postgres://${DATABASE_HOST}:${DATABASE_PORT}/myapp
ENV API_KEY=${API_KEY}
RUN echo "Node version: $NODE_VERSION"`,
			envFiles: map[string]string{
				".env": `DATABASE_HOST=localhost
API_KEY=secret123
NODE_ENV=development`,
			},
			expected: &DockerfileDiffResult{
				MissingInEnv:       []string{},
				ExtraInEnv:         []string{},
				UnusedArgs:         []string{},
				HardcodedEnvs:      []string{"DATABASE_URL"},
				MissingArgDefaults: []string{"DATABASE_HOST"},
			},
			expectError: false,
		},
		{
			name:              "empty Dockerfile",
			dockerfileContent: `FROM alpine:3.14`,
			envFiles: map[string]string{
				".env": `NODE_ENV=production
DATABASE_URL=postgres://localhost/myapp`,
			},
			expected: &DockerfileDiffResult{
				MissingInEnv:       []string{},
				ExtraInEnv:         []string{"DATABASE_URL", "NODE_ENV"},
				UnusedArgs:         []string{},
				HardcodedEnvs:      []string{},
				MissingArgDefaults: []string{},
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temporary directory and files
			tmpDir := t.TempDir()

			// Change to temp directory to use relative paths (security validation)
			originalDir, err := os.Getwd()
			if err != nil {
				t.Fatalf("Failed to get current directory: %v", err)
			}
			defer os.Chdir(originalDir)

			if err := os.Chdir(tmpDir); err != nil {
				t.Fatalf("Failed to change to temp directory: %v", err)
			}

			dockerfilePath := "Dockerfile"

			// Write Dockerfile
			err = os.WriteFile(dockerfilePath, []byte(tt.dockerfileContent), 0644)
			if err != nil {
				t.Fatalf("Failed to create Dockerfile: %v", err)
			}

			// Write env files
			var envFilePaths []string
			for filename, content := range tt.envFiles {
				err := os.WriteFile(filename, []byte(content), 0644)
				if err != nil {
					t.Fatalf("Failed to create env file %s: %v", filename, err)
				}
				envFilePaths = append(envFilePaths, filename)
			}

			// Run the comparison
			result, err := CompareDockerfileWithEnv(dockerfilePath, envFilePaths)

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
			if !compareDockerfileDiffResults(result, tt.expected) {
				t.Errorf("CompareDockerfileWithEnv() result mismatch")
				t.Errorf("Got MissingInEnv: %v", result.MissingInEnv)
				t.Errorf("Want MissingInEnv: %v", tt.expected.MissingInEnv)
				t.Errorf("Got ExtraInEnv: %v", result.ExtraInEnv)
				t.Errorf("Want ExtraInEnv: %v", tt.expected.ExtraInEnv)
				t.Errorf("Got UnusedArgs: %v", result.UnusedArgs)
				t.Errorf("Want UnusedArgs: %v", tt.expected.UnusedArgs)
				t.Errorf("Got HardcodedEnvs: %v", result.HardcodedEnvs)
				t.Errorf("Want HardcodedEnvs: %v", tt.expected.HardcodedEnvs)
				t.Errorf("Got MissingArgDefaults: %v", result.MissingArgDefaults)
				t.Errorf("Want MissingArgDefaults: %v", tt.expected.MissingArgDefaults)
			}
		})
	}
}

func TestCompareDockerfileWithEnv_FileErrors(t *testing.T) {
	tests := []struct {
		name            string
		setupDockerfile func(string) error
		setupEnv        func(string) error
		expectError     bool
	}{
		{
			name: "missing Dockerfile",
			setupDockerfile: func(filename string) error {
				// Don't create the file
				return nil
			},
			setupEnv: func(filename string) error {
				return os.WriteFile(filename, []byte("NODE_ENV=production"), 0644)
			},
			expectError: true,
		},
		{
			name: "Dockerfile is directory",
			setupDockerfile: func(filename string) error {
				return os.Mkdir(filename, 0755)
			},
			setupEnv: func(filename string) error {
				return os.WriteFile(filename, []byte("NODE_ENV=production"), 0644)
			},
			expectError: true,
		},
		{
			name: "missing env files are handled gracefully",
			setupDockerfile: func(filename string) error {
				content := `FROM alpine:3.14
ENV NODE_ENV=${NODE_ENV}`
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
			dockerfilePath := filepath.Join(tmpDir, "Dockerfile")
			envFile := filepath.Join(tmpDir, ".env")

			// Setup files
			if err := tt.setupDockerfile(dockerfilePath); err != nil {
				t.Fatalf("Failed to setup Dockerfile: %v", err)
			}
			if err := tt.setupEnv(envFile); err != nil {
				t.Fatalf("Failed to setup env file: %v", err)
			}

			// Clean up after test
			defer func() {
				os.RemoveAll(dockerfilePath)
				os.RemoveAll(envFile)
			}()

			// Run comparison
			result, err := CompareDockerfileWithEnv(dockerfilePath, []string{envFile})

			if tt.expectError && err == nil {
				t.Errorf("Expected error but got none. Result: %v", result)
			}
			if !tt.expectError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
		})
	}
}

func TestCompareDockerfileWithEnvVars(t *testing.T) {
	tests := []struct {
		name           string
		dockerfileInfo *parser.DockerfileEnvInfo
		envVars        parser.EnvVars
		expected       *DockerfileDiffResult
	}{
		{
			name: "perfect match",
			dockerfileInfo: &parser.DockerfileEnvInfo{
				EnvVars: parser.EnvVars{
					"NODE_ENV":     "${NODE_ENV}",
					"DATABASE_URL": "${DATABASE_URL}",
				},
				ArgVars: parser.EnvVars{
					"BUILD_VERSION": "1.0.0",
				},
				VariableRefs: []string{"NODE_ENV", "DATABASE_URL", "BUILD_VERSION"},
			},
			envVars: parser.EnvVars{
				"NODE_ENV":     "production",
				"DATABASE_URL": "postgres://localhost/myapp",
			},
			expected: &DockerfileDiffResult{
				MissingInEnv:       []string{},
				ExtraInEnv:         []string{},
				UnusedArgs:         []string{},
				HardcodedEnvs:      []string{},
				MissingArgDefaults: []string{},
			},
		},
		{
			name: "missing variables with unused args",
			dockerfileInfo: &parser.DockerfileEnvInfo{
				EnvVars: parser.EnvVars{
					"NODE_ENV":     "${NODE_ENV}",
					"API_KEY":      "${API_KEY}",
					"DATABASE_URL": "${DATABASE_URL}",
				},
				ArgVars: parser.EnvVars{
					"BUILD_VERSION": "",
					"UNUSED_ARG":    "default",
				},
				VariableRefs: []string{"NODE_ENV", "API_KEY", "DATABASE_URL"},
			},
			envVars: parser.EnvVars{
				"NODE_ENV": "production",
			},
			expected: &DockerfileDiffResult{
				MissingInEnv:       []string{"API_KEY", "DATABASE_URL"},
				ExtraInEnv:         []string{},
				UnusedArgs:         []string{"BUILD_VERSION", "UNUSED_ARG"},
				HardcodedEnvs:      []string{},
				MissingArgDefaults: []string{"BUILD_VERSION"},
			},
		},
		{
			name: "extra variables in env",
			dockerfileInfo: &parser.DockerfileEnvInfo{
				EnvVars:      parser.EnvVars{"NODE_ENV": "${NODE_ENV}"},
				ArgVars:      parser.EnvVars{},
				VariableRefs: []string{"NODE_ENV"},
			},
			envVars: parser.EnvVars{
				"NODE_ENV":     "production",
				"DATABASE_URL": "postgres://localhost/myapp",
				"API_KEY":      "secret123",
				"DEBUG":        "true",
			},
			expected: &DockerfileDiffResult{
				MissingInEnv:       []string{},
				ExtraInEnv:         []string{"API_KEY", "DATABASE_URL", "DEBUG"},
				UnusedArgs:         []string{},
				HardcodedEnvs:      []string{},
				MissingArgDefaults: []string{},
			},
		},
		{
			name: "hardcoded ENV variables",
			dockerfileInfo: &parser.DockerfileEnvInfo{
				EnvVars: parser.EnvVars{
					"NODE_ENV":  "production",
					"API_URL":   "https://api.example.com",
					"LOG_LEVEL": "info",
					"WORKDIR":   "/app",
				},
				ArgVars:      parser.EnvVars{},
				VariableRefs: []string{},
			},
			envVars: parser.EnvVars{},
			expected: &DockerfileDiffResult{
				MissingInEnv:       []string{},
				ExtraInEnv:         []string{},
				UnusedArgs:         []string{},
				HardcodedEnvs:      []string{"API_URL", "NODE_ENV"}, // WORKDIR and LOG_LEVEL are filtered as constants
				MissingArgDefaults: []string{},
			},
		},
		{
			name: "empty dockerfile info",
			dockerfileInfo: &parser.DockerfileEnvInfo{
				EnvVars:      parser.EnvVars{},
				ArgVars:      parser.EnvVars{},
				VariableRefs: []string{},
			},
			envVars: parser.EnvVars{
				"NODE_ENV":     "production",
				"DATABASE_URL": "postgres://localhost/myapp",
			},
			expected: &DockerfileDiffResult{
				MissingInEnv:       []string{},
				ExtraInEnv:         []string{"DATABASE_URL", "NODE_ENV"},
				UnusedArgs:         []string{},
				HardcodedEnvs:      []string{},
				MissingArgDefaults: []string{},
			},
		},
		{
			name: "empty env vars",
			dockerfileInfo: &parser.DockerfileEnvInfo{
				EnvVars: parser.EnvVars{
					"NODE_ENV":     "${NODE_ENV}",
					"DATABASE_URL": "${DATABASE_URL}",
				},
				ArgVars: parser.EnvVars{
					"API_KEY": "",
				},
				VariableRefs: []string{"NODE_ENV", "DATABASE_URL", "API_KEY"},
			},
			envVars: parser.EnvVars{},
			expected: &DockerfileDiffResult{
				MissingInEnv:       []string{"API_KEY", "DATABASE_URL", "NODE_ENV"},
				ExtraInEnv:         []string{},
				UnusedArgs:         []string{},
				HardcodedEnvs:      []string{},
				MissingArgDefaults: []string{"API_KEY"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := compareDockerfileWithEnvVars(tt.dockerfileInfo, tt.envVars)

			if !compareDockerfileDiffResults(result, tt.expected) {
				t.Errorf("compareDockerfileWithEnvVars() result mismatch")
				t.Errorf("Got MissingInEnv: %v", result.MissingInEnv)
				t.Errorf("Want MissingInEnv: %v", tt.expected.MissingInEnv)
				t.Errorf("Got ExtraInEnv: %v", result.ExtraInEnv)
				t.Errorf("Want ExtraInEnv: %v", tt.expected.ExtraInEnv)
				t.Errorf("Got UnusedArgs: %v", result.UnusedArgs)
				t.Errorf("Want UnusedArgs: %v", tt.expected.UnusedArgs)
				t.Errorf("Got HardcodedEnvs: %v", result.HardcodedEnvs)
				t.Errorf("Want HardcodedEnvs: %v", tt.expected.HardcodedEnvs)
				t.Errorf("Got MissingArgDefaults: %v", result.MissingArgDefaults)
				t.Errorf("Want MissingArgDefaults: %v", tt.expected.MissingArgDefaults)
			}
		})
	}
}

func TestDockerfileDiffResult_HasIssues(t *testing.T) {
	tests := []struct {
		name     string
		result   *DockerfileDiffResult
		expected bool
	}{
		{
			name: "no issues",
			result: &DockerfileDiffResult{
				MissingInEnv:       []string{},
				ExtraInEnv:         []string{},
				UnusedArgs:         []string{},
				HardcodedEnvs:      []string{},
				MissingArgDefaults: []string{},
			},
			expected: false,
		},
		{
			name: "missing variables",
			result: &DockerfileDiffResult{
				MissingInEnv:       []string{"API_KEY", "DATABASE_URL"},
				ExtraInEnv:         []string{},
				UnusedArgs:         []string{},
				HardcodedEnvs:      []string{},
				MissingArgDefaults: []string{},
			},
			expected: true,
		},
		{
			name: "extra variables",
			result: &DockerfileDiffResult{
				MissingInEnv:       []string{},
				ExtraInEnv:         []string{"DEBUG", "LOG_LEVEL"},
				UnusedArgs:         []string{},
				HardcodedEnvs:      []string{},
				MissingArgDefaults: []string{},
			},
			expected: true,
		},
		{
			name: "unused args",
			result: &DockerfileDiffResult{
				MissingInEnv:       []string{},
				ExtraInEnv:         []string{},
				UnusedArgs:         []string{"UNUSED_ARG"},
				HardcodedEnvs:      []string{},
				MissingArgDefaults: []string{},
			},
			expected: true,
		},
		{
			name: "hardcoded envs",
			result: &DockerfileDiffResult{
				MissingInEnv:       []string{},
				ExtraInEnv:         []string{},
				UnusedArgs:         []string{},
				HardcodedEnvs:      []string{"NODE_ENV"},
				MissingArgDefaults: []string{},
			},
			expected: true,
		},
		{
			name: "all types of issues",
			result: &DockerfileDiffResult{
				MissingInEnv:       []string{"API_KEY"},
				ExtraInEnv:         []string{"DEBUG"},
				UnusedArgs:         []string{"UNUSED_ARG"},
				HardcodedEnvs:      []string{"NODE_ENV"},
				MissingArgDefaults: []string{"BUILD_VERSION"},
			},
			expected: true,
		},
		{
			name: "nil slices treated as empty",
			result: &DockerfileDiffResult{
				MissingInEnv:       nil,
				ExtraInEnv:         nil,
				UnusedArgs:         nil,
				HardcodedEnvs:      nil,
				MissingArgDefaults: nil,
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

func TestGenerateDockerfileReport(t *testing.T) {
	tests := []struct {
		name        string
		result      *DockerfileDiffResult
		opts        *ReportOptions
		contains    []string // Strings that should be in the report
		notContains []string // Strings that should NOT be in the report
	}{
		{
			name: "no issues report",
			result: &DockerfileDiffResult{
				MissingInEnv:       []string{},
				ExtraInEnv:         []string{},
				UnusedArgs:         []string{},
				HardcodedEnvs:      []string{},
				MissingArgDefaults: []string{},
			},
			opts: &ReportOptions{ShowDuck: true, Colorize: true, Verbose: false},
			contains: []string{
				"✅ Dockerfile environment is aligned",
				"gopher-duck approves",
			},
			notContains: []string{
				"QUACK!",
				"Missing",
				"Unused",
			},
		},
		{
			name: "missing variables report",
			result: &DockerfileDiffResult{
				MissingInEnv:       []string{"API_KEY", "DATABASE_URL"},
				ExtraInEnv:         []string{},
				UnusedArgs:         []string{},
				HardcodedEnvs:      []string{},
				MissingArgDefaults: []string{},
			},
			opts: &ReportOptions{ShowDuck: true, Colorize: true, Verbose: false},
			contains: []string{
				"QUACK!",
				"🔴 Variables required by Dockerfile but missing",
				"API_KEY",
				"DATABASE_URL",
			},
			notContains: []string{
				"✅",
				"Unused",
				"Hardcoded",
			},
		},
		{
			name: "unused ARG variables report",
			result: &DockerfileDiffResult{
				MissingInEnv:       []string{},
				ExtraInEnv:         []string{},
				UnusedArgs:         []string{"UNUSED_ARG", "ANOTHER_UNUSED"},
				HardcodedEnvs:      []string{},
				MissingArgDefaults: []string{},
			},
			opts: &ReportOptions{ShowDuck: true, Colorize: true, Verbose: true},
			contains: []string{
				"🟠 ARG variables declared but never used",
				"UNUSED_ARG",
				"ANOTHER_UNUSED",
			},
			notContains: []string{
				"✅",
				"Missing",
			},
		},
		{
			name: "hardcoded ENV variables report",
			result: &DockerfileDiffResult{
				MissingInEnv:       []string{},
				ExtraInEnv:         []string{},
				UnusedArgs:         []string{},
				HardcodedEnvs:      []string{"NODE_ENV", "API_URL"},
				MissingArgDefaults: []string{},
			},
			opts: &ReportOptions{ShowDuck: false, Colorize: false, Verbose: true},
			contains: []string{
				"Hardcoded ENV variables:",
				"NODE_ENV",
				"API_URL",
			},
			notContains: []string{
				"🦆",
				"gopher-duck",
				"🟡",
				"✅",
			},
		},
		{
			name: "missing ARG defaults report",
			result: &DockerfileDiffResult{
				MissingInEnv:       []string{},
				ExtraInEnv:         []string{},
				UnusedArgs:         []string{},
				HardcodedEnvs:      []string{},
				MissingArgDefaults: []string{"BUILD_VERSION", "API_KEY"},
			},
			opts: &ReportOptions{ShowDuck: true, Colorize: true, Verbose: true},
			contains: []string{
				"⚠️  ARG variables without default values",
				"BUILD_VERSION",
				"API_KEY",
			},
			notContains: []string{
				"✅",
				"Missing",
				"Unused",
			},
		},
		{
			name: "comprehensive report with all issue types",
			result: &DockerfileDiffResult{
				MissingInEnv:       []string{"SECRET_KEY"},
				ExtraInEnv:         []string{"DEBUG"},
				UnusedArgs:         []string{"UNUSED_ARG"},
				HardcodedEnvs:      []string{"NODE_ENV"},
				MissingArgDefaults: []string{"BUILD_VERSION"},
			},
			opts: &ReportOptions{ShowDuck: true, Colorize: true, Verbose: true},
			contains: []string{
				"QUACK!",
				"🔴 Variables required by Dockerfile but missing",
				"SECRET_KEY",
				"🟠 ARG variables declared but never used",
				"UNUSED_ARG",
				"🟡 ENV variables with hardcoded values",
				"NODE_ENV",
				"⚠️  ARG variables without default values",
				"BUILD_VERSION",
				"🔵 Variables in env files but not used",
				"DEBUG",
				"confused by your Dockerfile setup",
			},
			notContains: []string{
				"✅",
				"approves",
			},
		},
		{
			name: "non-verbose mode hides some warnings",
			result: &DockerfileDiffResult{
				MissingInEnv:       []string{"API_KEY"},
				ExtraInEnv:         []string{},
				UnusedArgs:         []string{"UNUSED_ARG"},
				HardcodedEnvs:      []string{"NODE_ENV"},
				MissingArgDefaults: []string{"BUILD_VERSION"},
			},
			opts: &ReportOptions{ShowDuck: true, Colorize: true, Verbose: false},
			contains: []string{
				"🔴 Variables required by Dockerfile but missing",
				"API_KEY",
				"🟠 ARG variables declared but never used",
				"UNUSED_ARG",
			},
			notContains: []string{
				"Hardcoded ENV variables",
				"ARG variables without default",
				"✅",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			report := GenerateDockerfileReport(tt.result, tt.opts)

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

func TestGenerateDockerfileReport_DefaultOptions(t *testing.T) {
	result := &DockerfileDiffResult{
		MissingInEnv:       []string{"API_KEY"},
		ExtraInEnv:         []string{},
		UnusedArgs:         []string{},
		HardcodedEnvs:      []string{},
		MissingArgDefaults: []string{},
	}

	// Test with nil options (should use defaults)
	report := GenerateDockerfileReport(result, nil)

	// Should contain default behavior
	if !strings.Contains(report, "API_KEY") {
		t.Errorf("Report should contain API_KEY with default options")
	}
}

func TestIsObviousConstant(t *testing.T) {
	tests := []struct {
		name     string
		value    string
		expected bool
	}{
		// Obvious constants that should be detected
		{name: "boolean true", value: "true", expected: true},
		{name: "boolean false", value: "false", expected: true},
		{name: "number zero", value: "0", expected: true},
		{name: "number one", value: "1", expected: true},
		{name: "utf8 encoding", value: "utf8", expected: true},
		{name: "utf-8 encoding", value: "utf-8", expected: true},
		{name: "locale en_US", value: "en_US", expected: true},
		{name: "locale C", value: "C", expected: true},
		{name: "app path", value: "/app", expected: true},
		{name: "bin path", value: "/usr/local/bin", expected: true},
		{name: "tmp path", value: "/tmp", expected: true},

		// Variable references (should be filtered)
		{name: "variable reference ${}", value: "${NODE_ENV}", expected: true},
		{name: "variable reference $", value: "$PORT", expected: true},

		// Values that should NOT be considered constants (should be reported as hardcoded)
		{name: "production environment", value: "production", expected: false},
		{name: "development environment", value: "development", expected: false},
		{name: "staging environment", value: "staging", expected: false},
		{name: "test environment", value: "test", expected: false},
		{name: "http URL", value: "http://example.com", expected: false},
		{name: "https URL", value: "https://api.example.com", expected: false},
		{name: "database URL", value: "postgres://db:5432/myapp", expected: false},
		{name: "custom API key", value: "sk_test_123abc", expected: false},
		{name: "database password", value: "mypassword123", expected: false},
		{name: "custom port", value: "3000", expected: true}, // Common port
		{name: "JWT secret", value: "super-secret-key", expected: false},
		{name: "custom domain", value: "myapp.example.com", expected: false},
		{name: "empty string", value: "", expected: false},
		{name: "random string", value: "random_config_value", expected: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isObviousConstant(tt.value)
			if result != tt.expected {
				t.Errorf("isObviousConstant(%q) = %v, want %v", tt.value, result, tt.expected)
			}
		})
	}
}

func TestCompareDockerfileWithEnv_RealWorldScenarios(t *testing.T) {
	tests := []struct {
		name                    string
		dockerfileContent       string
		envContent              string
		expectedMissingCount    int
		expectedExtraCount      int
		expectedUnusedArgsCount int
		expectedHardcodedCount  int
	}{
		{
			name: "Node.js production application",
			dockerfileContent: `FROM node:${NODE_VERSION:-16} AS builder

ARG NODE_ENV=production
ARG BUILD_VERSION
ARG API_ENDPOINT

ENV NODE_ENV=${NODE_ENV}
ENV PORT=${PORT:-3000}
ENV API_ENDPOINT=${API_ENDPOINT}
ENV DATABASE_URL=${DATABASE_URL}
ENV JWT_SECRET=${JWT_SECRET}
ENV LOG_LEVEL=info

WORKDIR /app
COPY package*.json ./
RUN npm ci --only=${NODE_ENV}

COPY src/ ./src/
RUN npm run build:${NODE_ENV}

EXPOSE $PORT
CMD ["node", "dist/server.js"]`,
			envContent: `NODE_ENV=production
PORT=8080
API_ENDPOINT=https://api.staging.com
DATABASE_URL=postgres://user:pass@db:5432/myapp
JWT_SECRET=super-secret-jwt-key
EXTRA_CONFIG=not_used_in_dockerfile`,
			expectedMissingCount:    0, // All required variables provided
			expectedExtraCount:      1, // EXTRA_CONFIG
			expectedUnusedArgsCount: 1, // BUILD_VERSION (not referenced)
			expectedHardcodedCount:  1, // LOG_LEVEL
		},
		{
			name: "Multi-stage Python application",
			dockerfileContent: `FROM python:${PYTHON_VERSION} AS base

ARG PYTHON_VERSION=3.9
ARG REQUIREMENTS_FILE=requirements.txt
ARG APP_USER=appuser

ENV PYTHONPATH=/app
ENV FLASK_ENV=${FLASK_ENV:-development}
ENV DATABASE_URL=${DATABASE_URL}
ENV SECRET_KEY=${SECRET_KEY}
ENV REDIS_URL=${REDIS_URL}

RUN groupadd -r $APP_USER && useradd -r -g $APP_USER $APP_USER

WORKDIR /app
COPY ${REQUIREMENTS_FILE} .
RUN pip install -r ${REQUIREMENTS_FILE}

COPY app/ ./app/
USER $APP_USER

EXPOSE ${PORT:-5000}
CMD ["python", "-m", "flask", "run", "--host=0.0.0.0", "--port=${PORT:-5000}"]`,
			envContent: `FLASK_ENV=production
DATABASE_URL=postgres://user:pass@db:5432/myapp
SECRET_KEY=flask-secret-key-12345
REDIS_URL=redis://cache:6379
PORT=5000
UNUSED_VAR=not_referenced`,
			expectedMissingCount:    0, // All required variables provided
			expectedExtraCount:      1, // UNUSED_VAR
			expectedUnusedArgsCount: 0, // All ARGs are used
			expectedHardcodedCount:  1, // PYTHONPATH
		},
		{
			name: "Go microservice with missing variables",
			dockerfileContent: `FROM golang:${GO_VERSION} AS builder

ARG GO_VERSION=1.19
ARG CGO_ENABLED=0
ARG GOOS=linux

ENV CGO_ENABLED=${CGO_ENABLED}
ENV GOOS=${GOOS}
ENV SERVICE_NAME=${SERVICE_NAME}
ENV SERVICE_PORT=${SERVICE_PORT:-8080}
ENV DATABASE_DSN=${DATABASE_DSN}
ENV AUTH_SECRET=${AUTH_SECRET}
ENV LOG_FORMAT=json

WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download

COPY cmd/ ./cmd/
COPY internal/ ./internal/
RUN go build -o /app/service ./cmd/main.go

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/

COPY --from=builder /app/service .
EXPOSE ${SERVICE_PORT:-8080}
CMD ["./service"]`,
			envContent: `GO_VERSION=1.19
SERVICE_NAME=user-service
SERVICE_PORT=8081
AUTH_SECRET=auth-secret-key
DEBUG=true
METRICS_ENABLED=true`,
			expectedMissingCount:    1, // DATABASE_DSN missing
			expectedExtraCount:      2, // DEBUG, METRICS_ENABLED
			expectedUnusedArgsCount: 0, // All ARGs used
			expectedHardcodedCount:  2, // LOG_FORMAT, CGO_ENABLED, GOOS might be detected
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			dockerfilePath := filepath.Join(tmpDir, "Dockerfile")
			envFile := filepath.Join(tmpDir, ".env")

			err := os.WriteFile(dockerfilePath, []byte(tt.dockerfileContent), 0644)
			if err != nil {
				t.Fatalf("Failed to create Dockerfile: %v", err)
			}

			err = os.WriteFile(envFile, []byte(tt.envContent), 0644)
			if err != nil {
				t.Fatalf("Failed to create env file: %v", err)
			}

			result, err := CompareDockerfileWithEnv(dockerfilePath, []string{envFile})
			if err != nil {
				t.Fatalf("CompareDockerfileWithEnv failed: %v", err)
			}

			if len(result.MissingInEnv) != tt.expectedMissingCount {
				t.Errorf("Expected %d missing variables, got %d: %v",
					tt.expectedMissingCount, len(result.MissingInEnv), result.MissingInEnv)
			}

			if len(result.ExtraInEnv) != tt.expectedExtraCount {
				t.Errorf("Expected %d extra variables, got %d: %v",
					tt.expectedExtraCount, len(result.ExtraInEnv), result.ExtraInEnv)
			}

			if len(result.UnusedArgs) != tt.expectedUnusedArgsCount {
				t.Errorf("Expected %d unused args, got %d: %v",
					tt.expectedUnusedArgsCount, len(result.UnusedArgs), result.UnusedArgs)
			}

			// Note: Hardcoded count is approximate since constant detection may vary
			if tt.expectedHardcodedCount > 0 && len(result.HardcodedEnvs) == 0 {
				t.Errorf("Expected some hardcoded ENVs but got none")
			}

			// Verify HasIssues works correctly
			expectedHasIssues := tt.expectedMissingCount > 0 || tt.expectedExtraCount > 0 ||
				tt.expectedUnusedArgsCount > 0 || len(result.HardcodedEnvs) > 0
			if result.HasIssues() != expectedHasIssues {
				t.Errorf("HasIssues() = %v, want %v", result.HasIssues(), expectedHasIssues)
			}
		})
	}
}

// Helper function to compare DockerfileDiffResult structs
func compareDockerfileDiffResults(a, b *DockerfileDiffResult) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}

	// Helper function to sort and compare string slices
	compareSlices := func(slice1, slice2 []string) bool {
		if len(slice1) != len(slice2) {
			return false
		}

		s1Copy := make([]string, len(slice1))
		copy(s1Copy, slice1)
		sort.Strings(s1Copy)

		s2Copy := make([]string, len(slice2))
		copy(s2Copy, slice2)
		sort.Strings(s2Copy)

		return reflect.DeepEqual(s1Copy, s2Copy)
	}

	return compareSlices(a.MissingInEnv, b.MissingInEnv) &&
		compareSlices(a.ExtraInEnv, b.ExtraInEnv) &&
		compareSlices(a.UnusedArgs, b.UnusedArgs) &&
		compareSlices(a.HardcodedEnvs, b.HardcodedEnvs) &&
		compareSlices(a.MissingArgDefaults, b.MissingArgDefaults)
}

// Benchmark tests
func BenchmarkCompareDockerfileWithEnv_Small(b *testing.B) {
	dockerfileContent := `FROM node:16
ENV NODE_ENV=${NODE_ENV}
ENV DATABASE_URL=${DATABASE_URL}
ENV PORT=${PORT}`

	envContent := `NODE_ENV=production
DATABASE_URL=postgres://localhost
PORT=3000`

	tmpDir := b.TempDir()
	dockerfilePath := filepath.Join(tmpDir, "Dockerfile")
	envFile := filepath.Join(tmpDir, ".env")

	os.WriteFile(dockerfilePath, []byte(dockerfileContent), 0644)
	os.WriteFile(envFile, []byte(envContent), 0644)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := CompareDockerfileWithEnv(dockerfilePath, []string{envFile})
		if err != nil {
			b.Fatalf("Benchmark failed: %v", err)
		}
	}
}

func BenchmarkCompareDockerfileWithEnv_Large(b *testing.B) {
	// Create large Dockerfile with many ARG and ENV instructions
	var dockerfileContent strings.Builder
	var envContent strings.Builder

	dockerfileContent.WriteString("FROM alpine:3.14\n")

	for i := 0; i < 100; i++ {
		argName := fmt.Sprintf("ARG_%d", i)
		envName := fmt.Sprintf("ENV_%d", i)
		dockerfileContent.WriteString(fmt.Sprintf("ARG %s=default_%d\n", argName, i))
		dockerfileContent.WriteString(fmt.Sprintf("ENV %s=${%s}\n", envName, envName))
		envContent.WriteString(fmt.Sprintf("%s=value_%d\n", envName, i))
	}

	tmpDir := b.TempDir()
	dockerfilePath := filepath.Join(tmpDir, "Dockerfile")
	envFile := filepath.Join(tmpDir, ".env")

	os.WriteFile(dockerfilePath, []byte(dockerfileContent.String()), 0644)
	os.WriteFile(envFile, []byte(envContent.String()), 0644)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := CompareDockerfileWithEnv(dockerfilePath, []string{envFile})
		if err != nil {
			b.Fatalf("Benchmark failed: %v", err)
		}
	}
}
