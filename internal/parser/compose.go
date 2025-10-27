package parser

import (
	"fmt"
	"io"
	"os"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/DuckDHD/EnvQuack/internal/errors"
	"github.com/DuckDHD/EnvQuack/internal/security"
	"gopkg.in/yaml.v3"
)

// ComposeService represents a service in docker-compose
type ComposeService struct {
	Environment interface{} `yaml:"environment"`
	EnvFile     interface{} `yaml:"env_file"`
}

// ComposeFile represents the structure of a docker-compose.yml
type ComposeFile struct {
	Version  string                    `yaml:"version"`
	Services map[string]ComposeService `yaml:"services"`
}

// ComposeEnvInfo contains environment information extracted from compose file
type ComposeEnvInfo struct {
	Variables    EnvVars            // All environment variables found
	ServiceVars  map[string]EnvVars // Variables by service name
	EnvFiles     []string           // Referenced env_file paths
	VariableRefs []string           // Variables referenced as ${VAR} or $VAR
}

// ParseComposeFile parses a docker-compose.yml file and extracts environment variables
func ParseComposeFile(filename string) (*ComposeEnvInfo, error) {
	// Validate file path for security
	if err := security.ValidateFilePath(filename); err != nil {
		return nil, fmt.Errorf("invalid file path: %w", err)
	}

	// Validate file size to prevent OOM attacks
	if err := security.ValidateFileSize(filename, 10); err != nil {
		return nil, err
	}

	file, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to open compose file: %w", err)
	}
	defer file.Close()

	data, err := io.ReadAll(file)
	if err != nil {
		return nil, fmt.Errorf("failed to read compose file: %w", err)
	}

	return ParseComposeDataWithFilename(data, filename)
}

// ParseComposeData parses docker-compose YAML data (for backward compatibility)
func ParseComposeData(data []byte) (*ComposeEnvInfo, error) {
	return ParseComposeDataWithFilename(data, "docker-compose.yml")
}

// ParseComposeDataWithFilename parses docker-compose YAML data with enhanced error messages
func ParseComposeDataWithFilename(data []byte, filename string) (*ComposeEnvInfo, error) {
	var compose ComposeFile
	if err := yaml.Unmarshal(data, &compose); err != nil {
		// Try to extract line number from YAML error
		if parseErr := formatYAMLError(filename, string(data), err); parseErr != nil {
			return nil, parseErr
		}
		return nil, fmt.Errorf("failed to parse YAML: %w", err)
	}

	info := &ComposeEnvInfo{
		Variables:    make(EnvVars),
		ServiceVars:  make(map[string]EnvVars),
		EnvFiles:     []string{},
		VariableRefs: []string{},
	}

	// Extract variables from each service
	for serviceName, service := range compose.Services {
		serviceVars := make(EnvVars)

		// Parse environment variables
		envVars := parseEnvironmentSection(service.Environment)
		for k, v := range envVars {
			info.Variables[k] = v
			serviceVars[k] = v
		}

		// Parse env_file references
		envFiles := parseEnvFileSection(service.EnvFile)
		info.EnvFiles = append(info.EnvFiles, envFiles...)

		// Store service-specific variables
		if len(serviceVars) > 0 {
			info.ServiceVars[serviceName] = serviceVars
		}
	}

	// Extract variable references from the entire YAML content
	info.VariableRefs = extractVariableReferences(string(data))

	// Remove duplicates from env files
	info.EnvFiles = removeDuplicates(info.EnvFiles)
	sort.Strings(info.EnvFiles)

	return info, nil
}

// parseEnvironmentSection handles different formats of environment sections
func parseEnvironmentSection(env interface{}) EnvVars {
	vars := make(EnvVars)
	if env == nil {
		return vars
	}

	switch e := env.(type) {
	case []interface{}:
		// Array format: ["VAR1=value1", "VAR2=value2"]
		for _, item := range e {
			if str, ok := item.(string); ok {
				key, value := parseEnvString(str)
				if key != "" {
					vars[key] = value
				}
			}
		}
	case map[string]interface{}:
		// Object format: {VAR1: value1, VAR2: value2}
		for key, value := range e {
			if value == nil {
				vars[key] = ""
			} else {
				vars[key] = fmt.Sprintf("%v", value)
			}
		}
	}

	return vars
}

// parseEnvFileSection handles different formats of env_file sections
func parseEnvFileSection(envFile interface{}) []string {
	var files []string
	if envFile == nil {
		return files
	}

	switch e := envFile.(type) {
	case string:
		// Single file: env_file: .env
		files = append(files, e)
	case []interface{}:
		// Array format: env_file: [.env, .env.local]
		for _, item := range e {
			if str, ok := item.(string); ok {
				files = append(files, str)
			}
		}
	}

	return files
}

// parseEnvString parses "KEY=value" or "KEY" format
func parseEnvString(str string) (string, string) {
	str = strings.TrimSpace(str)
	if str == "" {
		return "", ""
	}

	// Handle "KEY=value" format
	if strings.Contains(str, "=") {
		parts := strings.SplitN(str, "=", 2)
		return strings.TrimSpace(parts[0]), strings.TrimSpace(parts[1])
	}

	// Handle "KEY" format (no value)
	return str, ""
}

// extractVariableReferences finds ${VAR} and $VAR references in the compose file
func extractVariableReferences(content string) []string {
	// Regex patterns for variable references
	patterns := []*regexp.Regexp{
		regexp.MustCompile(`\$\{([A-Z_][A-Z0-9_]*)\}`),        // ${VAR_NAME}
		regexp.MustCompile(`\$\{([A-Z_][A-Z0-9_]*):?[^}]*\}`), // ${VAR_NAME:-default}
		regexp.MustCompile(`\$([A-Z_][A-Z0-9_]*)`),            // $VAR_NAME
	}

	varSet := make(map[string]bool)

	for _, pattern := range patterns {
		matches := pattern.FindAllStringSubmatch(content, -1)
		for _, match := range matches {
			if len(match) > 1 {
				varName := match[1]
				// Filter out common docker variables that aren't typically in .env
				if !isDockerInternalVar(varName) {
					varSet[varName] = true
				}
			}
		}
	}

	// Convert to sorted slice
	vars := make([]string, 0, len(varSet))
	for varName := range varSet {
		vars = append(vars, varName)
	}
	sort.Strings(vars)

	return vars
}

// isDockerInternalVar checks if a variable is a Docker/Compose internal variable
func isDockerInternalVar(varName string) bool {
	internalVars := map[string]bool{
		"COMPOSE_PROJECT_NAME":   true,
		"COMPOSE_FILE":           true,
		"COMPOSE_PATH_SEPARATOR": true,
		"DOCKER_HOST":            true,
		"DOCKER_TLS_VERIFY":      true,
		"DOCKER_CERT_PATH":       true,
		"HOSTNAME":               true,
		"USER":                   true,
		"HOME":                   true,
		"PATH":                   true,
		"PWD":                    true,
	}
	return internalVars[varName]
}

// removeDuplicates removes duplicate strings from a slice
func removeDuplicates(slice []string) []string {
	seen := make(map[string]bool)
	result := []string{}

	for _, item := range slice {
		if !seen[item] {
			seen[item] = true
			result = append(result, item)
		}
	}

	return result
}

// GetAllEnvVars returns all unique environment variable names from compose info
func (c *ComposeEnvInfo) GetAllEnvVars() []string {
	varSet := make(map[string]bool)

	// Add explicitly defined variables
	for key := range c.Variables {
		varSet[key] = true
	}

	// Add referenced variables
	for _, ref := range c.VariableRefs {
		varSet[ref] = true
	}

	vars := make([]string, 0, len(varSet))
	for key := range varSet {
		vars = append(vars, key)
	}
	sort.Strings(vars)

	return vars
}

// GetServiceVars returns variables for a specific service
func (c *ComposeEnvInfo) GetServiceVars(serviceName string) EnvVars {
	if vars, exists := c.ServiceVars[serviceName]; exists {
		return vars
	}
	return make(EnvVars)
}

// HasService checks if a service exists in the compose file
func (c *ComposeEnvInfo) HasService(serviceName string) bool {
	_, exists := c.ServiceVars[serviceName]
	return exists
}

// GetServices returns all service names that have environment variables
func (c *ComposeEnvInfo) GetServices() []string {
	services := make([]string, 0, len(c.ServiceVars))
	for service := range c.ServiceVars {
		services = append(services, service)
	}
	sort.Strings(services)
	return services
}

// formatYAMLError extracts line numbers from YAML errors and creates enhanced error messages
func formatYAMLError(filename string, content string, err error) error {
	if err == nil {
		return nil
	}

	errMsg := err.Error()
	lines := strings.Split(content, "\n")

	// Try to extract line number from error message
	// YAML errors typically look like: "yaml: line 15: mapping values are not allowed here"
	lineNum := extractLineNumber(errMsg)
	if lineNum > 0 && lineNum <= len(lines) {
		// Extract a meaningful error message
		message := extractYAMLErrorMessage(errMsg)

		return errors.NewParseError(filename, lineNum, message).
			WithContext(lines, lineNum, 2).
			WithHint("Check YAML syntax - ensure proper indentation and structure")
	}

	// If we can't extract line number, return a generic parse error
	return nil
}

// extractLineNumber extracts the line number from a YAML error message
func extractLineNumber(errMsg string) int {
	// Try different patterns to extract line number
	patterns := []string{
		"line ",    // "yaml: line 15: ..."
		"at line ", // "error at line 15"
	}

	for _, pattern := range patterns {
		idx := strings.Index(errMsg, pattern)
		if idx >= 0 {
			// Extract the number after the pattern
			start := idx + len(pattern)
			end := start
			for end < len(errMsg) && errMsg[end] >= '0' && errMsg[end] <= '9' {
				end++
			}
			if end > start {
				numStr := errMsg[start:end]
				if num, err := strconv.Atoi(numStr); err == nil {
					return num
				}
			}
		}
	}

	return 0
}

// extractYAMLErrorMessage extracts a clean error message from YAML error
func extractYAMLErrorMessage(errMsg string) string {
	// Remove "yaml: " prefix if present
	errMsg = strings.TrimPrefix(errMsg, "yaml: ")

	// If there's a line number, extract the message after it
	if idx := strings.Index(errMsg, "line "); idx >= 0 {
		// Find the colon after the line number
		if colonIdx := strings.Index(errMsg[idx:], ":"); colonIdx >= 0 {
			msg := strings.TrimSpace(errMsg[idx+colonIdx+1:])
			if msg != "" {
				return msg
			}
		}
	}

	return errMsg
}
