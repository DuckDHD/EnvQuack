package parser

import (
	"fmt"
	"os"
	"strings"

	"github.com/DuckDHD/EnvQuack/internal/errors"
	"github.com/DuckDHD/EnvQuack/internal/security"
)

// EnvVars represents a collection of environment variables
type EnvVars map[string]string

// ParseEnvFile parses a .env file and returns the environment variables
func ParseEnvFile(filename string) (EnvVars, error) {
	// Validate file path for security
	if err := security.ValidateFilePath(filename); err != nil {
		return nil, fmt.Errorf("invalid file path: %w", err)
	}

	// Validate file size to prevent OOM attacks
	if err := security.ValidateFileSize(filename, 10); err != nil {
		return nil, fmt.Errorf("file validation failed: %w", err)
	}

	// Read entire file to track line numbers
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	// Split into lines for line number tracking
	lines := strings.Split(string(data), "\n")
	vars := make(EnvVars)

	for lineNum, line := range lines {
		lineNumber := lineNum + 1 // 1-indexed for users

		// Trim only spaces, preserve tabs for error display
		line = strings.TrimSpace(line)

		// Skip empty lines and comments
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// Parse the line
		key, value, err := parseLine(line)
		if err != nil {
			// Return enhanced error with line number and context
			parseErr := errors.NewParseError(filename, lineNumber, err.Error()).
				WithContext(lines, lineNumber, 2).
				WithHint("Each line should be in format KEY=VALUE")

			return nil, parseErr
		}

		vars[key] = value
	}

	return vars, nil
}

// parseLine parses a single line of .env file
func parseLine(line string) (key, value string, err error) {
	// Find the first '=' sign
	idx := strings.Index(line, "=")
	if idx == -1 {
		return "", "", fmt.Errorf("missing '=' separator")
	}

	key = strings.TrimSpace(line[:idx])
	if key == "" {
		return "", "", fmt.Errorf("empty variable name before '='")
	}

	// Validate key format (should be alphanumeric + underscore)
	if !isValidKey(key) {
		return "", "", fmt.Errorf("invalid variable name '%s' (use A-Z, 0-9, _)", key)
	}

	value = strings.TrimSpace(line[idx+1:])

	// Check for unclosed quotes
	if err := validateQuotes(value); err != nil {
		return "", "", err
	}

	// Remove quotes if present and properly balanced
	value = unquoteValue(value)

	return key, value, nil
}

// isValidKey checks if a variable name is valid
// Valid keys: [A-Z_][A-Z0-9_]* (uppercase recommended but not enforced)
func isValidKey(key string) bool {
	if len(key) == 0 {
		return false
	}

	for i, ch := range key {
		// First character: A-Z, a-z, or _
		if i == 0 {
			if !((ch >= 'A' && ch <= 'Z') || (ch >= 'a' && ch <= 'z') || ch == '_') {
				return false
			}
		} else {
			// Subsequent characters: A-Z, a-z, 0-9, or _
			if !((ch >= 'A' && ch <= 'Z') || (ch >= 'a' && ch <= 'z') || (ch >= '0' && ch <= '9') || ch == '_') {
				return false
			}
		}
	}
	return true
}

// validateQuotes checks if quotes are properly balanced
func validateQuotes(value string) error {
	if len(value) == 0 {
		return nil
	}

	// Check for unclosed double quotes
	if strings.HasPrefix(value, "\"") && !strings.HasSuffix(value, "\"") {
		return fmt.Errorf("unclosed double quote")
	}

	// Check for unclosed single quotes
	if strings.HasPrefix(value, "'") && !strings.HasSuffix(value, "'") {
		return fmt.Errorf("unclosed single quote")
	}

	return nil
}

// unquoteValue removes surrounding quotes from a value
func unquoteValue(value string) string {
	if len(value) >= 2 {
		if (strings.HasPrefix(value, "\"") && strings.HasSuffix(value, "\"")) ||
			(strings.HasPrefix(value, "'") && strings.HasSuffix(value, "'")) {
			return value[1 : len(value)-1]
		}
	}
	return value
}

// GetKeys returns all the keys from the environment variables
func (e EnvVars) GetKeys() []string {
	keys := make([]string, 0, len(e))
	for key := range e {
		keys = append(keys, key)
	}
	return keys
}

// Has checks if a key exists in the environment variables
func (e EnvVars) Has(key string) bool {
	_, exists := e[key]
	return exists
}
