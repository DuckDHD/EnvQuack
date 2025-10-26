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

func TestParseDockerfile(t *testing.T) {
	tests := []struct {
		name        string
		content     string
		expected    *DockerfileEnvInfo
		expectError bool
	}{
		{
			name: "basic ENV and ARG instructions",
			content: `FROM node:16
ARG NODE_ENV=production
ENV PORT=3000
ENV DATABASE_URL=postgres://localhost/myapp`,
			expected: &DockerfileEnvInfo{
				EnvVars: EnvVars{
					"PORT":         "3000",
					"DATABASE_URL": "postgres://localhost/myapp",
				},
				ArgVars: EnvVars{
					"NODE_ENV": "production",
				},
				VariableRefs: []string{},
			},
			expectError: false,
		},
		{
			name: "ARG without default values",
			content: `FROM alpine:3.14
ARG BUILD_VERSION
ARG API_KEY
ARG NODE_ENV=development`,
			expected: &DockerfileEnvInfo{
				EnvVars: EnvVars{},
				ArgVars: EnvVars{
					"BUILD_VERSION": "",
					"API_KEY":       "",
					"NODE_ENV":      "development",
				},
				VariableRefs: []string{},
			},
			expectError: false,
		},
		{
			name: "ENV legacy format (ENV key value)",
			content: `FROM ubuntu:20.04
ENV NODE_ENV production
ENV LOG_LEVEL info
ENV WORKDIR /app`,
			expected: &DockerfileEnvInfo{
				EnvVars: EnvVars{
					"NODE_ENV":  "production",
					"LOG_LEVEL": "info",
					"WORKDIR":   "/app",
				},
				ArgVars:      EnvVars{},
				VariableRefs: []string{},
			},
			expectError: false,
		},
		{
			name: "ENV modern format (ENV key=value)",
			content: `FROM node:16
ENV NODE_ENV=production
ENV PORT=3000
ENV DEBUG=false`,
			expected: &DockerfileEnvInfo{
				EnvVars: EnvVars{
					"NODE_ENV": "production",
					"PORT":     "3000",
					"DEBUG":    "false",
				},
				ArgVars:      EnvVars{},
				VariableRefs: []string{},
			},
			expectError: false,
		},
		{
			name: "ENV multi-value format",
			content: `FROM alpine:3.14
ENV NODE_ENV=production PORT=3000 DEBUG=false
ENV API_URL=https://api.example.com LOG_LEVEL=info`,
			expected: &DockerfileEnvInfo{
				EnvVars: EnvVars{
					"NODE_ENV":  "production",
					"PORT":      "3000",
					"DEBUG":     "false",
					"API_URL":   "https://api.example.com",
					"LOG_LEVEL": "info",
				},
				ArgVars:      EnvVars{},
				VariableRefs: []string{},
			},
			expectError: false,
		},
		{
			name: "quoted values with spaces",
			content: `FROM node:16
ENV APP_NAME="My Application"
ENV DESCRIPTION='This is a test application'
ENV COMPLEX="value with 'single' quotes"
ARG BUILD_ARGS="--production --verbose"`,
			expected: &DockerfileEnvInfo{
				EnvVars: EnvVars{
					"APP_NAME":    "My Application",
					"DESCRIPTION": "This is a test application",
					"COMPLEX":     "value with 'single' quotes",
				},
				ArgVars: EnvVars{
					"BUILD_ARGS": "--production --verbose",
				},
				VariableRefs: []string{},
			},
			expectError: false,
		},
		{
			name: "multi-line instructions with backslash continuation",
			content: `FROM node:16
ENV NODE_ENV=production \
    PORT=3000 \
    DEBUG=false

ARG BUILD_VERSION=1.0.0 \
    API_ENDPOINT=https://api.example.com`,
			expected: &DockerfileEnvInfo{
				EnvVars: EnvVars{
					"NODE_ENV": "production",
					"PORT":     "3000",
					"DEBUG":    "false",
				},
				ArgVars: EnvVars{
					"BUILD_VERSION": "1.0.0",
					"API_ENDPOINT":  "https://api.example.com",
				},
				VariableRefs: []string{},
			},
			expectError: false,
		},
		{
			name: "variable references ${VAR} and $VAR",
			content: `FROM node:${NODE_VERSION:-16}
ARG NODE_VERSION
ARG API_KEY
ENV NODE_ENV=${NODE_ENV}
ENV API_ENDPOINT=${API_ENDPOINT:-https://api.default.com}
RUN echo "Version: $NODE_VERSION"
COPY package.json ${WORKDIR}/package.json
WORKDIR $WORKDIR`,
			expected: &DockerfileEnvInfo{
				EnvVars: EnvVars{
					"NODE_ENV":     "${NODE_ENV}",
					"API_ENDPOINT": "${API_ENDPOINT:-https://api.default.com}",
				},
				ArgVars: EnvVars{
					"NODE_VERSION": "",
					"API_KEY":      "",
				},
				VariableRefs: []string{"API_ENDPOINT", "NODE_ENV", "NODE_VERSION", "WORKDIR"},
			},
			expectError: false,
		},
		{
			name: "system variables filtered out",
			content: `FROM alpine:3.14
ENV PATH=/usr/local/bin:$PATH
ENV HOME=/home/app
ENV USER=appuser
ENV MY_VAR=$MY_VAR
WORKDIR $HOME/app`,
			expected: &DockerfileEnvInfo{
				EnvVars: EnvVars{
					"PATH":   "/usr/local/bin:$PATH",
					"HOME":   "/home/app",
					"USER":   "appuser",
					"MY_VAR": "$MY_VAR",
				},
				ArgVars:      EnvVars{},
				VariableRefs: []string{"MY_VAR"}, // System vars filtered out
			},
			expectError: false,
		},
		{
			name: "comments and empty lines",
			content: `# This is a comment
FROM node:16

# Set environment variables
ENV NODE_ENV=production
# ENV COMMENTED_OUT=value

ARG BUILD_VERSION
# Another comment

ENV PORT=3000`,
			expected: &DockerfileEnvInfo{
				EnvVars: EnvVars{
					"NODE_ENV": "production",
					"PORT":     "3000",
				},
				ArgVars: EnvVars{
					"BUILD_VERSION": "",
				},
				VariableRefs: []string{},
			},
			expectError: false,
		},
		{
			name: "complex real-world Dockerfile",
			content: `FROM node:${NODE_VERSION:-16} AS builder

# Build arguments
ARG NODE_ENV=production
ARG API_ENDPOINT
ARG BUILD_VERSION
ARG ENABLE_FEATURES="auth,analytics"

# Runtime environment variables
ENV NODE_ENV=${NODE_ENV}
ENV PORT=${PORT:-3000}
ENV API_ENDPOINT=${API_ENDPOINT}
ENV LOG_LEVEL=${LOG_LEVEL:-info}
ENV FEATURES="$ENABLE_FEATURES"

# Application setup
WORKDIR /app
COPY package*.json ./
RUN npm ci --only=production && \
    npm cache clean --force

# Copy application code
COPY src/ ./src/
COPY public/ ./public/

# Build the application
RUN npm run build:${NODE_ENV}

# Health check with variable
HEALTHCHECK --interval=30s --timeout=3s \
  CMD curl -f http://localhost:$PORT/health || exit 1

EXPOSE $PORT
CMD ["node", "dist/server.js"]`,
			expected: &DockerfileEnvInfo{
				EnvVars: EnvVars{
					"NODE_ENV":     "${NODE_ENV}",
					"PORT":         "${PORT:-3000}",
					"API_ENDPOINT": "${API_ENDPOINT}",
					"LOG_LEVEL":    "${LOG_LEVEL:-info}",
					"FEATURES":     "$ENABLE_FEATURES",
				},
				ArgVars: EnvVars{
					"NODE_ENV":        "production",
					"API_ENDPOINT":    "",
					"BUILD_VERSION":   "",
					"ENABLE_FEATURES": "auth,analytics",
				},
				VariableRefs: []string{
					"API_ENDPOINT", "ENABLE_FEATURES", "LOG_LEVEL",
					"NODE_ENV", "NODE_VERSION", "PORT",
				},
			},
			expectError: false,
		},
		{
			name: "edge case: ENV with equals in value",
			content: `FROM alpine:3.14
ENV EQUATION="x=y+z"
ENV BASE64_KEY="dGVzdA=="
ENV URL_PARAMS="https://api.com?key=value&other=test"`,
			expected: &DockerfileEnvInfo{
				EnvVars: EnvVars{
					"EQUATION":   "x=y+z",
					"BASE64_KEY": "dGVzdA==",
					"URL_PARAMS": "https://api.com?key=value&other=test",
				},
				ArgVars:      EnvVars{},
				VariableRefs: []string{},
			},
			expectError: false,
		},
		{
			name: "edge case: empty values",
			content: `FROM alpine:3.14
ENV EMPTY_VAR=""
ENV EMPTY_VAR2=
ARG EMPTY_ARG=
ARG EMPTY_ARG2`,
			expected: &DockerfileEnvInfo{
				EnvVars: EnvVars{
					"EMPTY_VAR":  "",
					"EMPTY_VAR2": "",
				},
				ArgVars: EnvVars{
					"EMPTY_ARG":  "",
					"EMPTY_ARG2": "",
				},
				VariableRefs: []string{},
			},
			expectError: false,
		},
		{
			name: "malformed instructions ignored",
			content: `FROM alpine:3.14
ENV VALID=value
ENV # This is malformed
ARG VALID_ARG=test
ARG # This is also malformed
ENV ANOTHER_VALID=another`,
			expected: &DockerfileEnvInfo{
				EnvVars: EnvVars{
					"VALID":         "value",
					"ANOTHER_VALID": "another",
				},
				ArgVars: EnvVars{
					"VALID_ARG": "test",
				},
				VariableRefs: []string{},
			},
			expectError: false,
		},
		{
			name:        "empty Dockerfile",
			content:     "",
			expected:    &DockerfileEnvInfo{EnvVars: EnvVars{}, ArgVars: EnvVars{}, VariableRefs: []string{}},
			expectError: false,
		},
		{
			name: "only comments",
			content: `# This is just a comment
# Another comment
# No actual instructions`,
			expected:    &DockerfileEnvInfo{EnvVars: EnvVars{}, ArgVars: EnvVars{}, VariableRefs: []string{}},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temporary file
			tmpDir := t.TempDir()
			tmpFile := filepath.Join(tmpDir, "Dockerfile")

			err := os.WriteFile(tmpFile, []byte(tt.content), 0644)
			if err != nil {
				t.Fatalf("Failed to create test file: %v", err)
			}

			// Parse the file
			result, err := ParseDockerfile(tmpFile)

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
			if !compareDockerfileEnvInfo(result, tt.expected) {
				t.Errorf("ParseDockerfile() result mismatch")
				t.Errorf("Got EnvVars: %v", result.EnvVars)
				t.Errorf("Want EnvVars: %v", tt.expected.EnvVars)
				t.Errorf("Got ArgVars: %v", result.ArgVars)
				t.Errorf("Want ArgVars: %v", tt.expected.ArgVars)
				t.Errorf("Got VariableRefs: %v", result.VariableRefs)
				t.Errorf("Want VariableRefs: %v", tt.expected.VariableRefs)
			}
		})
	}
}

func TestParseDockerfile_FileErrors(t *testing.T) {
	tests := []struct {
		name     string
		filename string
		setup    func(string) error
	}{
		{
			name:     "non-existent file",
			filename: "/non/existent/Dockerfile",
		},
		{
			name:     "directory instead of file",
			filename: "directory-dockerfile",
			setup: func(filename string) error {
				return os.Mkdir(filename, 0755)
			},
		},
		{
			name:     "permission denied",
			filename: "permission_denied_dockerfile",
			setup: func(filename string) error {
				content := `FROM alpine:3.14
ENV TEST=value`
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

			result, err := ParseDockerfile(testFile)

			if err == nil {
				t.Errorf("Expected error for %s, but got none. Result: %v", tt.name, result)
			}

			if result != nil {
				t.Errorf("Expected nil result on error, got: %v", result)
			}
		})
	}
}

func TestParseDockerfileInstruction(t *testing.T) {
	tests := []struct {
		name           string
		instruction    string
		expectedEnvVar string
		expectedEnvVal string
		expectedArgVar string
		expectedArgVal string
		expectError    bool
	}{
		{
			name:           "ENV key=value",
			instruction:    "ENV NODE_ENV=production",
			expectedEnvVar: "NODE_ENV",
			expectedEnvVal: "production",
		},
		{
			name:           "ENV key value",
			instruction:    "ENV NODE_ENV production",
			expectedEnvVar: "NODE_ENV",
			expectedEnvVal: "production",
		},
		{
			name:           "ARG with default",
			instruction:    "ARG NODE_VERSION=16",
			expectedArgVar: "NODE_VERSION",
			expectedArgVal: "16",
		},
		{
			name:           "ARG without default",
			instruction:    "ARG API_KEY",
			expectedArgVar: "API_KEY",
			expectedArgVal: "",
		},
		{
			name:        "non-ENV/ARG instruction",
			instruction: "FROM alpine:3.14",
			expectError: false, // Should not error, just ignore
		},
		{
			name:        "RUN instruction",
			instruction: "RUN apt-get update",
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			info := &DockerfileEnvInfo{
				EnvVars:      make(EnvVars),
				ArgVars:      make(EnvVars),
				VariableRefs: []string{},
			}

			err := parseDockerfileInstruction(tt.instruction, info)

			if tt.expectError && err == nil {
				t.Errorf("Expected error but got none")
				return
			}
			if !tt.expectError && err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			// Check ENV variables
			if tt.expectedEnvVar != "" {
				if val, exists := info.EnvVars[tt.expectedEnvVar]; !exists {
					t.Errorf("Expected ENV var %s not found", tt.expectedEnvVar)
				} else if val != tt.expectedEnvVal {
					t.Errorf("ENV var %s = %q, want %q", tt.expectedEnvVar, val, tt.expectedEnvVal)
				}
			}

			// Check ARG variables
			if tt.expectedArgVar != "" {
				if val, exists := info.ArgVars[tt.expectedArgVar]; !exists {
					t.Errorf("Expected ARG var %s not found", tt.expectedArgVar)
				} else if val != tt.expectedArgVal {
					t.Errorf("ARG var %s = %q, want %q", tt.expectedArgVar, val, tt.expectedArgVal)
				}
			}
		})
	}
}

func TestParseEnvInstruction(t *testing.T) {
	tests := []struct {
		name        string
		content     string
		expected    EnvVars
		expectError bool
	}{
		{
			name:    "single key=value",
			content: "NODE_ENV=production",
			expected: EnvVars{
				"NODE_ENV": "production",
			},
		},
		{
			name:    "multiple key=value pairs",
			content: "NODE_ENV=production PORT=3000 DEBUG=false",
			expected: EnvVars{
				"NODE_ENV": "production",
				"PORT":     "3000",
				"DEBUG":    "false",
			},
		},
		{
			name:    "legacy key value format",
			content: "NODE_ENV production",
			expected: EnvVars{
				"NODE_ENV": "production",
			},
		},
		{
			name:    "legacy with spaces in value",
			content: "APP_NAME My Application Name",
			expected: EnvVars{
				"APP_NAME": "My Application Name",
			},
		},
		{
			name:    "quoted values",
			content: `APP_NAME="My Application" VERSION='1.0.0'`,
			expected: EnvVars{
				"APP_NAME": "My Application",
				"VERSION":  "1.0.0",
			},
		},
		{
			name:    "values with equals signs",
			content: "EQUATION=x=y+z BASE64=dGVzdA==",
			expected: EnvVars{
				"EQUATION": "x=y+z",
				"BASE64":   "dGVzdA==",
			},
		},
		{
			name:    "empty values",
			content: "EMPTY= ANOTHER_EMPTY=",
			expected: EnvVars{
				"EMPTY":         "",
				"ANOTHER_EMPTY": "",
			},
		},
		{
			name:        "malformed - no content",
			content:     "",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			envVars := make(EnvVars)
			err := parseEnvInstruction(tt.content, envVars)

			if tt.expectError && err == nil {
				t.Errorf("Expected error but got none")
				return
			}
			if !tt.expectError && err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if !tt.expectError && !reflect.DeepEqual(envVars, tt.expected) {
				t.Errorf("parseEnvInstruction() = %v, want %v", envVars, tt.expected)
			}
		})
	}
}

func TestParseArgInstruction(t *testing.T) {
	tests := []struct {
		name        string
		content     string
		expected    EnvVars
		expectError bool
	}{
		{
			name:    "ARG with default value",
			content: "NODE_VERSION=16",
			expected: EnvVars{
				"NODE_VERSION": "16",
			},
		},
		{
			name:    "ARG without default value",
			content: "API_KEY",
			expected: EnvVars{
				"API_KEY": "",
			},
		},
		{
			name:    "multiple ARG with defaults",
			content: "NODE_VERSION=16 API_ENDPOINT=https://api.com",
			expected: EnvVars{
				"NODE_VERSION": "16",
				"API_ENDPOINT": "https://api.com",
			},
		},
		{
			name:    "quoted default values",
			content: `BUILD_ARGS="--production --verbose"`,
			expected: EnvVars{
				"BUILD_ARGS": "--production --verbose",
			},
		},
		{
			name:    "empty default value",
			content: "EMPTY_ARG=",
			expected: EnvVars{
				"EMPTY_ARG": "",
			},
		},
		{
			name:        "malformed - no content",
			content:     "",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			argVars := make(EnvVars)
			err := parseArgInstruction(tt.content, argVars)

			if tt.expectError && err == nil {
				t.Errorf("Expected error but got none")
				return
			}
			if !tt.expectError && err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if !tt.expectError && !reflect.DeepEqual(argVars, tt.expected) {
				t.Errorf("parseArgInstruction() = %v, want %v", argVars, tt.expected)
			}
		})
	}
}

func TestParseKeyValuePairs(t *testing.T) {
	tests := []struct {
		name     string
		content  string
		expected EnvVars
	}{
		{
			name:    "simple key=value",
			content: "KEY=value",
			expected: EnvVars{
				"KEY": "value",
			},
		},
		{
			name:    "multiple pairs",
			content: "KEY1=value1 KEY2=value2 KEY3=value3",
			expected: EnvVars{
				"KEY1": "value1",
				"KEY2": "value2",
				"KEY3": "value3",
			},
		},
		{
			name:    "quoted values with spaces",
			content: `NAME="John Doe" LOCATION="New York"`,
			expected: EnvVars{
				"NAME":     "John Doe",
				"LOCATION": "New York",
			},
		},
		{
			name:    "single quoted values",
			content: `APP='My App' VERSION='1.0.0'`,
			expected: EnvVars{
				"APP":     "My App",
				"VERSION": "1.0.0",
			},
		},
		{
			name:    "mixed quotes",
			content: `APP="My App" VERSION='1.0.0' DEBUG=true`,
			expected: EnvVars{
				"APP":     "My App",
				"VERSION": "1.0.0",
				"DEBUG":   "true",
			},
		},
		{
			name:    "values with equals signs",
			content: "EQUATION=x=y+z URL=https://api.com?key=value",
			expected: EnvVars{
				"EQUATION": "x=y+z",
				"URL":      "https://api.com?key=value",
			},
		},
		{
			name:    "quoted values with spaces inside quotes",
			content: `COMPLEX="value with 'single' quotes" ANOTHER='value with "double" quotes'`,
			expected: EnvVars{
				"COMPLEX": "value with 'single' quotes",
				"ANOTHER": `value with "double" quotes`,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			vars := make(EnvVars)
			err := parseKeyValuePairs(tt.content, vars)

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if !reflect.DeepEqual(vars, tt.expected) {
				t.Errorf("parseKeyValuePairs() = %v, want %v", vars, tt.expected)
			}
		})
	}
}

func TestExtractDockerfileVariableRefs(t *testing.T) {
	tests := []struct {
		name     string
		content  string
		expected []string
	}{
		{
			name: "basic variable references",
			content: `FROM node:$NODE_VERSION
ENV API_URL=${API_URL}
WORKDIR ${WORKDIR}`,
			expected: []string{"API_URL", "NODE_VERSION", "WORKDIR"},
		},
		{
			name: "system variables filtered",
			content: `FROM alpine:3.14
ENV PATH=/usr/local/bin:$PATH
ENV HOME=/home/app
WORKDIR $HOME
ENV MY_VAR=$MY_VAR`,
			expected: []string{"MY_VAR"}, // System vars filtered out
		},
		{
			name: "mixed variable formats",
			content: `FROM node:${NODE_VERSION}
RUN echo $BUILD_VERSION
COPY app ${APP_DIR}/
ENV API_KEY=${API_KEY}`,
			expected: []string{"API_KEY", "APP_DIR", "BUILD_VERSION", "NODE_VERSION"},
		},
		{
			name: "variables with default values",
			content: `FROM node:${NODE_VERSION:-16}
ENV PORT=${PORT:-3000}
EXPOSE ${WEB_PORT:-8080}`,
			expected: []string{"NODE_VERSION", "PORT", "WEB_PORT"},
		},
		{
			name:     "no variables",
			content:  "FROM alpine:3.14\nRUN apk add --no-cache curl",
			expected: []string{},
		},
		{
			name: "invalid variable names ignored",
			content: `FROM alpine:3.14
ENV VALID=${VALID_VAR}
ENV INVALID=${123INVALID}
ENV ANOTHER=${ANOTHER_VALID}`,
			expected: []string{"ANOTHER_VALID", "VALID_VAR"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := extractDockerfileVariableRefs(tt.content)
			sort.Strings(result)
			sort.Strings(tt.expected)

			if !reflect.DeepEqual(result, tt.expected) {
				t.Errorf("extractDockerfileVariableRefs() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestIsSystemVar(t *testing.T) {
	tests := []struct {
		name     string
		varName  string
		expected bool
	}{
		{
			name:     "PATH system variable",
			varName:  "PATH",
			expected: true,
		},
		{
			name:     "HOME system variable",
			varName:  "HOME",
			expected: true,
		},
		{
			name:     "USER system variable",
			varName:  "USER",
			expected: true,
		},
		{
			name:     "PWD system variable",
			varName:  "PWD",
			expected: true,
		},
		{
			name:     "custom variable",
			varName:  "MY_CUSTOM_VAR",
			expected: false,
		},
		{
			name:     "API_KEY custom variable",
			varName:  "API_KEY",
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
			result := isSystemVar(tt.varName)
			if result != tt.expected {
				t.Errorf("isSystemVar(%q) = %v, want %v", tt.varName, result, tt.expected)
			}
		})
	}
}

func TestDockerfileEnvInfo_Methods(t *testing.T) {
	info := &DockerfileEnvInfo{
		EnvVars: EnvVars{
			"NODE_ENV": "production",
			"PORT":     "3000",
		},
		ArgVars: EnvVars{
			"BUILD_VERSION": "1.0.0",
			"API_KEY":       "",
		},
		VariableRefs: []string{"DATABASE_URL", "REDIS_URL"},
	}

	t.Run("GetAllVars", func(t *testing.T) {
		result := info.GetAllVars()
		expected := []string{"API_KEY", "BUILD_VERSION", "DATABASE_URL", "NODE_ENV", "PORT", "REDIS_URL"}

		sort.Strings(result)
		if !reflect.DeepEqual(result, expected) {
			t.Errorf("GetAllVars() = %v, want %v", result, expected)
		}
	})

	t.Run("GetEnvVars", func(t *testing.T) {
		result := info.GetEnvVars()
		expected := []string{"NODE_ENV", "PORT"}

		sort.Strings(result)
		sort.Strings(expected)
		if !reflect.DeepEqual(result, expected) {
			t.Errorf("GetEnvVars() = %v, want %v", result, expected)
		}
	})

	t.Run("GetArgVars", func(t *testing.T) {
		result := info.GetArgVars()
		expected := []string{"API_KEY", "BUILD_VERSION"}

		sort.Strings(result)
		sort.Strings(expected)
		if !reflect.DeepEqual(result, expected) {
			t.Errorf("GetArgVars() = %v, want %v", result, expected)
		}
	})

	t.Run("HasVar", func(t *testing.T) {
		tests := []struct {
			varName  string
			expected bool
		}{
			{"NODE_ENV", true},      // In EnvVars
			{"BUILD_VERSION", true}, // In ArgVars
			{"DATABASE_URL", true},  // In VariableRefs
			{"NON_EXISTENT", false}, // Not found anywhere
			{"", false},             // Empty string
		}

		for _, tt := range tests {
			result := info.HasVar(tt.varName)
			if result != tt.expected {
				t.Errorf("HasVar(%q) = %v, want %v", tt.varName, result, tt.expected)
			}
		}
	})
}

func TestParseDockerfile_MultilineContinuation(t *testing.T) {
	content := `FROM node:16

# Multi-line ENV instruction
ENV NODE_ENV=production \
    PORT=3000 \
    DEBUG=false \
    API_URL=https://api.example.com

# Multi-line ARG instruction  
ARG BUILD_VERSION=1.0.0 \
    API_ENDPOINT=https://api.production.com \
    ENABLE_CACHE=true

# Multi-line RUN with variables
RUN npm install && \
    npm run build:$NODE_ENV && \
    npm cache clean --force

EXPOSE $PORT`

	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "Dockerfile")
	err := os.WriteFile(tmpFile, []byte(content), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	result, err := ParseDockerfile(tmpFile)
	if err != nil {
		t.Fatalf("Failed to parse Dockerfile: %v", err)
	}

	expectedEnvVars := EnvVars{
		"NODE_ENV": "production",
		"PORT":     "3000",
		"DEBUG":    "false",
		"API_URL":  "https://api.example.com",
	}

	expectedArgVars := EnvVars{
		"BUILD_VERSION": "1.0.0",
		"API_ENDPOINT":  "https://api.production.com",
		"ENABLE_CACHE":  "true",
	}

	if !reflect.DeepEqual(result.EnvVars, expectedEnvVars) {
		t.Errorf("EnvVars = %v, want %v", result.EnvVars, expectedEnvVars)
	}

	if !reflect.DeepEqual(result.ArgVars, expectedArgVars) {
		t.Errorf("ArgVars = %v, want %v", result.ArgVars, expectedArgVars)
	}

	// Check that NODE_ENV and PORT are in VariableRefs
	sort.Strings(result.VariableRefs)
	if !stringInSlice("NODE_ENV", result.VariableRefs) {
		t.Errorf("Expected NODE_ENV in VariableRefs, got: %v", result.VariableRefs)
	}
	if !stringInSlice("PORT", result.VariableRefs) {
		t.Errorf("Expected PORT in VariableRefs, got: %v", result.VariableRefs)
	}
}

func TestParseDockerfile_LargeFile(t *testing.T) {
	// Generate a large Dockerfile with many ENV and ARG instructions
	var content strings.Builder
	content.WriteString("FROM alpine:3.14\n\n")

	expectedEnvVars := make(EnvVars)
	expectedArgVars := make(EnvVars)

	// Generate 100 ENV instructions
	for i := 0; i < 100; i++ {
		envVar := fmt.Sprintf("ENV_VAR_%d", i)
		envVal := fmt.Sprintf("env_value_%d", i)
		content.WriteString(fmt.Sprintf("ENV %s=%s\n", envVar, envVal))
		expectedEnvVars[envVar] = envVal
	}

	// Generate 50 ARG instructions
	for i := 0; i < 50; i++ {
		argVar := fmt.Sprintf("ARG_VAR_%d", i)
		argVal := fmt.Sprintf("arg_value_%d", i)
		content.WriteString(fmt.Sprintf("ARG %s=%s\n", argVar, argVal))
		expectedArgVars[argVar] = argVal
	}

	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "Dockerfile")
	err := os.WriteFile(tmpFile, []byte(content.String()), 0644)
	if err != nil {
		t.Fatalf("Failed to create large test file: %v", err)
	}

	result, err := ParseDockerfile(tmpFile)
	if err != nil {
		t.Fatalf("Failed to parse large Dockerfile: %v", err)
	}

	if len(result.EnvVars) != 100 {
		t.Errorf("Expected 100 ENV vars, got %d", len(result.EnvVars))
	}

	if len(result.ArgVars) != 50 {
		t.Errorf("Expected 50 ARG vars, got %d", len(result.ArgVars))
	}

	// Spot check a few variables
	testEnvKeys := []string{"ENV_VAR_0", "ENV_VAR_50", "ENV_VAR_99"}
	for _, key := range testEnvKeys {
		if !result.EnvVars.Has(key) {
			t.Errorf("Missing expected ENV key: %s", key)
		}
	}

	testArgKeys := []string{"ARG_VAR_0", "ARG_VAR_25", "ARG_VAR_49"}
	for _, key := range testArgKeys {
		if !result.ArgVars.Has(key) {
			t.Errorf("Missing expected ARG key: %s", key)
		}
	}
}

// Helper function to compare DockerfileEnvInfo structs
func compareDockerfileEnvInfo(a, b *DockerfileEnvInfo) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}

	// Compare EnvVars
	if !reflect.DeepEqual(a.EnvVars, b.EnvVars) {
		return false
	}

	// Compare ArgVars
	if !reflect.DeepEqual(a.ArgVars, b.ArgVars) {
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

// Helper function to check if string is in slice
func stringInSlice(str string, slice []string) bool {
	for _, s := range slice {
		if s == str {
			return true
		}
	}
	return false
}

// Benchmark tests
func BenchmarkParseDockerfile_Small(b *testing.B) {
	content := `FROM node:16
ENV NODE_ENV=production
ENV PORT=3000
ARG API_KEY`

	tmpDir := b.TempDir()
	tmpFile := filepath.Join(tmpDir, "Dockerfile")
	os.WriteFile(tmpFile, []byte(content), 0644)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := ParseDockerfile(tmpFile)
		if err != nil {
			b.Fatalf("Benchmark failed: %v", err)
		}
	}
}

func BenchmarkParseDockerfile_Large(b *testing.B) {
	// Create a large Dockerfile for benchmarking
	var content strings.Builder
	content.WriteString("FROM alpine:3.14\n")
	for i := 0; i < 500; i++ {
		content.WriteString(fmt.Sprintf("ENV VAR_%d=value_%d\n", i, i))
		content.WriteString(fmt.Sprintf("ARG ARG_%d=default_%d\n", i, i))
	}

	tmpDir := b.TempDir()
	tmpFile := filepath.Join(tmpDir, "Dockerfile")
	os.WriteFile(tmpFile, []byte(content.String()), 0644)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := ParseDockerfile(tmpFile)
		if err != nil {
			b.Fatalf("Benchmark failed: %v", err)
		}
	}
}

func BenchmarkExtractDockerfileVariableRefs(b *testing.B) {
	content := `FROM node:${NODE_VERSION}
ENV NODE_ENV=${NODE_ENV}
ENV PORT=${PORT:-3000}
ARG API_KEY=$API_KEY
RUN echo $BUILD_VERSION
WORKDIR ${APP_DIR}`

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		extractDockerfileVariableRefs(content)
	}
}
