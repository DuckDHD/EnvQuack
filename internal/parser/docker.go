package parser

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
	"sort"
	"strings"
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
	file, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to open Dockerfile: %w", err)
	}
	defer file.Close()

	info := &DockerfileEnvInfo{
		EnvVars:      make(EnvVars),
		ArgVars:      make(EnvVars),
		VariableRefs: []string{},
	}

	scanner := bufio.NewScanner(file)
	lineNum := 0
	var currentInstruction strings.Builder

	for scanner.Scan() {
		lineNum++
		line := strings.TrimSpace(scanner.Text())

		// Skip empty lines and comments
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// Handle line continuation with backslash
		if strings.HasSuffix(line, "\\") {
			currentInstruction.WriteString(strings.TrimSuffix(line, "\\"))
			currentInstruction.WriteString(" ")
			continue
		}

		// Complete instruction (either single line or end of multi-line)
		if currentInstruction.Len() > 0 {
			line = currentInstruction.String() + line
			currentInstruction.Reset()
		}

		// Parse the instruction
		if err := parseDockerfileInstruction(line, info); err != nil {
			// Log warning but continue parsing
			fmt.Printf("Warning: line %d - %v\n", lineNum, err)
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading Dockerfile: %w", err)
	}

	// Extract variable references from all content
	file.Seek(0, 0) // Reset file pointer
	content, _ := os.ReadFile(filename)
	info.VariableRefs = extractDockerfileVariableRefs(string(content))

	return info, nil
}

// parseDockerfileInstruction parses a single Dockerfile instruction
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
