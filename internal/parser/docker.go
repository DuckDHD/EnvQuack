package parser

import (
	"fmt"
	"os"
	"regexp"
	"sort"
	"strings"

	"github.com/DuckDHD/EnvQuack/internal/errors"
	"github.com/DuckDHD/EnvQuack/internal/security"
)

// DockerfileEnvInfo contains environment information extracted from Dockerfile
type DockerfileEnvInfo struct {
	EnvVars      EnvVars  // ENV instructions
	ArgVars      EnvVars  // ARG instructions
	VariableRefs []string // Variables referenced as ${VAR} or $VAR
}

// Dockerfile instruction patterns
var (
	envInstructionRegex = regexp.MustCompile(`^ENV\s+(.+)$`)
	argInstructionRegex = regexp.MustCompile(`^ARG\s+(.+)$`)
	varRefRegex         = regexp.MustCompile(`\$\{?([A-Z_][A-Z0-9_]*)\}?`)
)

// ParseDockerfile parses a Dockerfile and extracts environment variables
func ParseDockerfile(filename string) (*DockerfileEnvInfo, error) {
	// Validate file path for security
	if err := security.ValidateFilePath(filename); err != nil {
		return nil, fmt.Errorf("invalid file path: %w", err)
	}

	// Validate file size to prevent OOM attacks
	if err := security.ValidateFileSize(filename, 10); err != nil {
		return nil, err
	}

	// Read entire file to track line numbers
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to read Dockerfile: %w", err)
	}

	content := string(data)
	lines := strings.Split(content, "\n")

	info := &DockerfileEnvInfo{
		EnvVars:      make(EnvVars),
		ArgVars:      make(EnvVars),
		VariableRefs: []string{},
	}

	var currentInstruction strings.Builder
	var instructionStartLine int

	for lineNum, line := range lines {
		lineNumber := lineNum + 1 // 1-indexed for users
		trimmedLine := strings.TrimSpace(line)

		// Skip empty lines and comments
		if trimmedLine == "" || strings.HasPrefix(trimmedLine, "#") {
			continue
		}

		// Track the start of multi-line instructions
		if currentInstruction.Len() == 0 {
			instructionStartLine = lineNumber
		}

		// Handle line continuation with backslash
		if strings.HasSuffix(trimmedLine, "\\") {
			currentInstruction.WriteString(strings.TrimSuffix(trimmedLine, "\\"))
			currentInstruction.WriteString(" ")
			continue
		}

		// Complete instruction (either single line or end of multi-line)
		var fullInstruction string
		if currentInstruction.Len() > 0 {
			fullInstruction = currentInstruction.String() + trimmedLine
			currentInstruction.Reset()
		} else {
			fullInstruction = trimmedLine
		}

		// Parse the instruction
		if err := parseDockerfileInstructionWithLine(fullInstruction, info, filename, lines, instructionStartLine); err != nil {
			return nil, err
		}
	}

	// Extract variable references from all content
	info.VariableRefs = extractDockerfileVariableRefs(content)

	return info, nil
}

// parseDockerfileInstructionWithLine parses a single Dockerfile instruction with line tracking
func parseDockerfileInstructionWithLine(line string, info *DockerfileEnvInfo, filename string, allLines []string, lineNum int) error {
	line = strings.TrimSpace(line)
	upperLine := strings.ToUpper(line)

	// Parse ENV instructions
	if envMatch := envInstructionRegex.FindStringSubmatch(upperLine); envMatch != nil {
		envContent := strings.TrimSpace(line[4:]) // Remove "ENV " prefix from original line
		if err := parseEnvInstruction(envContent, info.EnvVars); err != nil {
			return errors.NewParseError(filename, lineNum, err.Error()).
				WithContext(allLines, lineNum, 2).
				WithHint("ENV format: ENV KEY=value or ENV KEY1=value1 KEY2=value2")
		}
		return nil
	}

	// Parse ARG instructions
	if argMatch := argInstructionRegex.FindStringSubmatch(upperLine); argMatch != nil {
		argContent := strings.TrimSpace(line[4:]) // Remove "ARG " prefix from original line
		if err := parseArgInstruction(argContent, info.ArgVars); err != nil {
			return errors.NewParseError(filename, lineNum, err.Error()).
				WithContext(allLines, lineNum, 2).
				WithHint("ARG format: ARG NAME or ARG NAME=defaultvalue")
		}
		return nil
	}

	// Check for common instruction typos (only if line starts with a letter)
	if len(line) > 0 && ((line[0] >= 'A' && line[0] <= 'Z') || (line[0] >= 'a' && line[0] <= 'z')) {
		firstWord := strings.Fields(upperLine)
		if len(firstWord) > 0 {
			instruction := firstWord[0]
			// Check for common typos
			switch instruction {
			case "ARGS":
				return errors.NewParseError(filename, lineNum, "Invalid instruction 'ARGS'").
					WithContext(allLines, lineNum, 2).
					WithHint("Did you mean 'ARG'? See https://docs.docker.com/engine/reference/builder/")
			case "ENVS":
				return errors.NewParseError(filename, lineNum, "Invalid instruction 'ENVS'").
					WithContext(allLines, lineNum, 2).
					WithHint("Did you mean 'ENV'? See https://docs.docker.com/engine/reference/builder/")
			}
		}
	}

	return nil
}

// parseDockerfileInstruction parses a single Dockerfile instruction (legacy, for backward compatibility)
func parseDockerfileInstruction(line string, info *DockerfileEnvInfo) error {
	line = strings.TrimSpace(line)
	upperLine := strings.ToUpper(line)

	// Parse ENV instructions
	if envMatch := envInstructionRegex.FindStringSubmatch(upperLine); envMatch != nil {
		envContent := strings.TrimSpace(line[4:]) // Remove "ENV " prefix from original line
		return parseEnvInstruction(envContent, info.EnvVars)
	}

	// Parse ARG instructions
	if argMatch := argInstructionRegex.FindStringSubmatch(upperLine); argMatch != nil {
		argContent := strings.TrimSpace(line[4:]) // Remove "ARG " prefix from original line
		return parseArgInstruction(argContent, info.ArgVars)
	}

	return nil
}

// parseEnvInstruction parses ENV instruction content
func parseEnvInstruction(content string, envVars EnvVars) error {
	// ENV can have multiple formats:
	// ENV key=value
	// ENV key1=value1 key2=value2
	// ENV key value (deprecated but still valid)

	// Try key=value format first
	if strings.Contains(content, "=") {
		return parseKeyValuePairs(content, envVars)
	}

	// Handle legacy "ENV key value" format
	parts := strings.Fields(content)
	if len(parts) >= 2 {
		key := parts[0]
		value := strings.Join(parts[1:], " ")
		envVars[key] = value
		return nil
	}

	return fmt.Errorf("invalid ENV instruction format: %s", content)
}

// parseArgInstruction parses ARG instruction content
func parseArgInstruction(content string, argVars EnvVars) error {
	// ARG can have formats:
	// ARG name
	// ARG name=defaultvalue

	if strings.Contains(content, "=") {
		return parseKeyValuePairs(content, argVars)
	}

	// ARG without default value
	parts := strings.Fields(content)
	if len(parts) == 1 {
		argVars[parts[0]] = ""
		return nil
	}

	return fmt.Errorf("invalid ARG instruction format: %s", content)
}

// parseKeyValuePairs parses "key1=value1 key2=value2" format
func parseKeyValuePairs(content string, vars EnvVars) error {
	// Handle quoted values and spaces properly
	var pairs []string
	var current strings.Builder
	inQuotes := false
	quoteChar := byte(0)

	for i := 0; i < len(content); i++ {
		char := content[i]

		switch char {
		case '"', '\'':
			if !inQuotes {
				inQuotes = true
				quoteChar = char
			} else if char == quoteChar {
				inQuotes = false
				quoteChar = 0
			}
			current.WriteByte(char)
		case ' ':
			if inQuotes {
				current.WriteByte(char)
			} else {
				if current.Len() > 0 {
					pairs = append(pairs, current.String())
					current.Reset()
				}
			}
		default:
			current.WriteByte(char)
		}
	}

	// Add the last pair
	if current.Len() > 0 {
		pairs = append(pairs, current.String())
	}

	// Parse each key=value pair
	for _, pair := range pairs {
		if strings.Contains(pair, "=") {
			kv := strings.SplitN(pair, "=", 2)
			key := strings.TrimSpace(kv[0])
			value := strings.TrimSpace(kv[1])

			// Remove quotes from value
			if len(value) >= 2 {
				if (strings.HasPrefix(value, "\"") && strings.HasSuffix(value, "\"")) ||
					(strings.HasPrefix(value, "'") && strings.HasSuffix(value, "'")) {
					value = value[1 : len(value)-1]
				}
			}

			vars[key] = value
		}
	}

	return nil
}

// extractDockerfileVariableRefs finds variable references in Dockerfile content
func extractDockerfileVariableRefs(content string) []string {
	varSet := make(map[string]bool)

	// Find all variable references
	matches := varRefRegex.FindAllStringSubmatch(content, -1)
	for _, match := range matches {
		if len(match) > 1 {
			varName := match[1]
			// Filter out common system variables
			if !isSystemVar(varName) {
				varSet[varName] = true
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

// isSystemVar checks if a variable is a common system variable
func isSystemVar(varName string) bool {
	systemVars := map[string]bool{
		"PATH":     true,
		"HOME":     true,
		"USER":     true,
		"SHELL":    true,
		"TERM":     true,
		"PWD":      true,
		"OLDPWD":   true,
		"HOSTNAME": true,
		"UID":      true,
		"GID":      true,
	}
	return systemVars[varName]
}

// GetAllVars returns all environment variable names from Dockerfile
func (d *DockerfileEnvInfo) GetAllVars() []string {
	varSet := make(map[string]bool)

	// Add ENV vars
	for key := range d.EnvVars {
		varSet[key] = true
	}

	// Add ARG vars
	for key := range d.ArgVars {
		varSet[key] = true
	}

	// Add referenced vars
	for _, ref := range d.VariableRefs {
		varSet[ref] = true
	}

	vars := make([]string, 0, len(varSet))
	for key := range varSet {
		vars = append(vars, key)
	}
	sort.Strings(vars)

	return vars
}

// GetEnvVars returns only ENV instruction variables
func (d *DockerfileEnvInfo) GetEnvVars() []string {
	return d.EnvVars.GetKeys()
}

// GetArgVars returns only ARG instruction variables
func (d *DockerfileEnvInfo) GetArgVars() []string {
	return d.ArgVars.GetKeys()
}

// HasVar checks if a variable exists in any form (ENV, ARG, or referenced)
func (d *DockerfileEnvInfo) HasVar(varName string) bool {
	if d.EnvVars.Has(varName) || d.ArgVars.Has(varName) {
		return true
	}

	for _, ref := range d.VariableRefs {
		if ref == varName {
			return true
		}
	}

	return false
}
