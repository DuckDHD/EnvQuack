package checker

import (
	"fmt"
	"sort"
	"strings"

	"github.com/DuckDHD/EnvQuack/internal/parser"
	"github.com/DuckDHD/EnvQuack/internal/quack"
)

// DockerfileDiffResult represents comparison between env files and Dockerfile
type DockerfileDiffResult struct {
	MissingInEnv       []string // Variables in Dockerfile but not in env files
	ExtraInEnv         []string // Variables in env files but not used in Dockerfile
	UnusedArgs         []string // ARG variables not referenced anywhere
	HardcodedEnvs      []string // ENV variables with hardcoded values (might need to be configurable)
	MissingArgDefaults []string // ARG variables without default values
}

// HasIssues returns true if there are any issues
func (d *DockerfileDiffResult) HasIssues() bool {
	return len(d.MissingInEnv) > 0 ||
		len(d.ExtraInEnv) > 0 ||
		len(d.UnusedArgs) > 0 ||
		len(d.HardcodedEnvs) > 0
}

// CompareDockerfileWithEnv compares Dockerfile requirements against env files
func CompareDockerfileWithEnv(dockerfilePath string, envFiles []string) (*DockerfileDiffResult, error) {
	// Parse Dockerfile
	dockerfileInfo, err := parser.ParseDockerfile(dockerfilePath)
	if err != nil {
		return nil, fmt.Errorf("failed to parse Dockerfile: %w", err)
	}

	// Parse all env files
	allEnvVars := make(parser.EnvVars)
	for _, envFile := range envFiles {
		envVars, err := parser.ParseEnvFile(envFile)
		if err != nil {
			// Skip missing files, we'll report them separately
			continue
		}

		// Merge env vars
		for k, v := range envVars {
			allEnvVars[k] = v
		}
	}

	return compareDockerfileWithEnvVars(dockerfileInfo, allEnvVars), nil
}

// compareDockerfileWithEnvVars performs the actual comparison logic
func compareDockerfileWithEnvVars(dockerfileInfo *parser.DockerfileEnvInfo, envVars parser.EnvVars) *DockerfileDiffResult {
	result := &DockerfileDiffResult{
		MissingInEnv:       []string{},
		ExtraInEnv:         []string{},
		UnusedArgs:         []string{},
		HardcodedEnvs:      []string{},
		MissingArgDefaults: []string{},
	}

	// Get all variables referenced in Dockerfile
	dockerfileVars := dockerfileInfo.GetAllVars()
	dockerfileVarSet := make(map[string]bool)
	for _, v := range dockerfileVars {
		dockerfileVarSet[v] = true
	}

	// Find missing variables (referenced in Dockerfile but not in env)
	for _, dockerVar := range dockerfileVars {
		// Check if variable is defined as ENV in Dockerfile with a non-variable value (has a default)
		if envValue, hasEnv := dockerfileInfo.EnvVars[dockerVar]; hasEnv {
			// If the ENV value is not a variable reference, it has a default
			if !isObviousConstant(envValue) {
				// This ENV has a hardcoded value, so it's not required in .env
				continue
			}
		}

		// Skip if variable is an ARG with a default value
		if argValue, isArg := dockerfileInfo.ArgVars[dockerVar]; isArg && argValue != "" {
			continue
		}

		// Check if it's in env files
		if !envVars.Has(dockerVar) {
			result.MissingInEnv = append(result.MissingInEnv, dockerVar)
		}
	}

	// Find extra variables (in env but not used in Dockerfile)
	for envVar := range envVars {
		if !dockerfileVarSet[envVar] {
			result.ExtraInEnv = append(result.ExtraInEnv, envVar)
		}
	}

	// Find unused ARG variables
	for argVar := range dockerfileInfo.ArgVars {
		// Check if ARG is referenced anywhere in variable references
		isUsed := false
		for _, ref := range dockerfileInfo.VariableRefs {
			if ref == argVar {
				isUsed = true
				break
			}
		}
		if !isUsed {
			result.UnusedArgs = append(result.UnusedArgs, argVar)
		}
	}

	// Find hardcoded ENV variables (might be better as configurable)
	for envVar, value := range dockerfileInfo.EnvVars {
		// Skip empty values and obvious constants
		if value != "" && !isObviousConstant(value) {
			result.HardcodedEnvs = append(result.HardcodedEnvs, envVar)
		}
	}

	// Find ARG variables without default values
	for argVar, value := range dockerfileInfo.ArgVars {
		if value == "" {
			result.MissingArgDefaults = append(result.MissingArgDefaults, argVar)
		}
	}

	// Sort results
	sort.Strings(result.MissingInEnv)
	sort.Strings(result.ExtraInEnv)
	sort.Strings(result.UnusedArgs)
	sort.Strings(result.HardcodedEnvs)
	sort.Strings(result.MissingArgDefaults)

	return result
}

// isObviousConstant checks if a value looks like a constant rather than config
// Returns true for values that are obviously not configuration (should be filtered out)
func isObviousConstant(value string) bool {
	// Variable references should always be filtered
	if strings.HasPrefix(value, "${") || strings.HasPrefix(value, "$") {
		return true
	}

	lowerValue := strings.ToLower(value)

	// Boolean and simple numeric values
	if lowerValue == "true" || lowerValue == "false" || lowerValue == "0" || lowerValue == "1" {
		return true
	}

	// Encoding/locale constants
	encodingLocales := []string{
		"utf8", "utf-8", "en_us", "c", "posix",
	}
	for _, locale := range encodingLocales {
		if lowerValue == locale {
			return true
		}
	}

	// Common log levels (these are typically constants)
	logLevels := []string{
		"debug", "info", "warn", "warning", "error", "fatal", "trace",
	}
	for _, level := range logLevels {
		if lowerValue == level {
			return true
		}
	}

	// Generic system paths (common Docker paths)
	commonPaths := []string{
		"/app", "/usr/bin", "/usr/local/bin", "/bin", "/tmp", "/var", "/etc",
		"/usr/local", "/opt", "/home", "/root",
	}
	for _, path := range commonPaths {
		if value == path {
			return true
		}
	}

	// Check if it's a simple numeric port that's very common
	commonPorts := []string{
		"80", "443", "3000", "5000", "8000", "8080", "9000",
	}
	for _, port := range commonPorts {
		if value == port {
			return true
		}
	}

	return false
}

// GenerateDockerfileReport creates a formatted report for Dockerfile comparison
func GenerateDockerfileReport(result *DockerfileDiffResult, opts *ReportOptions) string {
	if opts == nil {
		opts = DefaultReportOptions()
	}

	var report strings.Builder

	if !result.HasIssues() {
		report.WriteString("✅ Dockerfile environment is aligned.\n")
		if opts.ShowDuck {
			report.WriteString("(Your gopher-duck approves of your containerized setup!)\n")
		}
		return report.String()
	}

	// Header with duck
	if opts.ShowDuck {
		report.WriteString(quack.GetAngryDuck() + "\n")
		report.WriteString("QUACK! 🦆 Dockerfile environment issues detected:\n\n")
	}

	// Missing variables
	if len(result.MissingInEnv) > 0 {
		if opts.Colorize {
			report.WriteString("🔴 Variables required by Dockerfile but missing in env files:\n")
		} else {
			report.WriteString("Missing variables:\n")
		}

		for _, key := range result.MissingInEnv {
			report.WriteString(fmt.Sprintf("  - %s\n", key))
		}
		report.WriteString("\n")
	}

	// Unused ARG variables
	if len(result.UnusedArgs) > 0 {
		if opts.Colorize {
			report.WriteString("🟠 ARG variables declared but never used:\n")
		} else {
			report.WriteString("Unused ARG variables:\n")
		}

		for _, key := range result.UnusedArgs {
			report.WriteString(fmt.Sprintf("  - %s\n", key))
		}
		report.WriteString("\n")
	}

	// Hardcoded ENV variables (warnings)
	if len(result.HardcodedEnvs) > 0 && opts.Verbose {
		if opts.Colorize {
			report.WriteString("🟡 ENV variables with hardcoded values (consider making configurable):\n")
		} else {
			report.WriteString("Hardcoded ENV variables:\n")
		}

		for _, key := range result.HardcodedEnvs {
			report.WriteString(fmt.Sprintf("  - %s\n", key))
		}
		report.WriteString("\n")
	}

	// ARG variables without defaults
	if len(result.MissingArgDefaults) > 0 && opts.Verbose {
		if opts.Colorize {
			report.WriteString("⚠️  ARG variables without default values:\n")
		} else {
			report.WriteString("ARG variables without defaults:\n")
		}

		for _, key := range result.MissingArgDefaults {
			report.WriteString(fmt.Sprintf("  - %s\n", key))
		}
		report.WriteString("\n")
	}

	// Extra variables (usually less critical)
	if len(result.ExtraInEnv) > 0 {
		if opts.Colorize {
			report.WriteString("🔵 Variables in env files but not used in Dockerfile:\n")
		} else {
			report.WriteString("Unused variables:\n")
		}

		for _, key := range result.ExtraInEnv {
			report.WriteString(fmt.Sprintf("  - %s\n", key))
		}
		report.WriteString("\n")
	}

	// Footer with duck message
	if opts.ShowDuck {
		report.WriteString("(Your gopher-duck is confused by your Dockerfile setup!)\n")
	}

	return report.String()
}
