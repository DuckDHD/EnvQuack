package parser

import (
	"bufio"
	"os"
	"strings"
)

// EnvVars represents a collection of environment variables
type EnvVars map[string]string

// ParseEnvFile parses a .env file and returns the environment variables
func ParseEnvFile(filename string) (EnvVars, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	vars := make(EnvVars)
	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		// Skip empty lines and comments
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// Split on first = sign
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}

		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])

		// Remove quotes if present
		if len(value) >= 2 {
			if (strings.HasPrefix(value, "\"") && strings.HasSuffix(value, "\"")) ||
				(strings.HasPrefix(value, "'") && strings.HasSuffix(value, "'")) {
				value = value[1 : len(value)-1]
			}
		}

		vars[key] = value
	}

	return vars, scanner.Err()
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
