package checker

import (
	"fmt"
	"sort"
	"strings"

	"github.com/DuckDHD/EnvQuack/internal/parser"
	"github.com/DuckDHD/EnvQuack/internal/quack"
)

// ComposeDiffResult represents comparison between env files and compose file
type ComposeDiffResult struct {
	MissingInEnv     []string            // Variables in compose but not in env files
	ExtraInEnv       []string            // Variables in env files but not used in compose
	MissingEnvFiles  []string            // env_file references that don't exist
	ServiceBreakdown map[string][]string // Missing variables by service
}

// HasIssues returns true if there are any issues
func (c *ComposeDiffResult) HasIssues() bool {
	return len(c.MissingInEnv) > 0 ||
		len(c.ExtraInEnv) > 0 ||
		len(c.MissingEnvFiles) > 0
}

// CompareComposeWithEnv compares docker-compose requirements against env files
func CompareComposeWithEnv(composeFile string, envFiles []string) (*ComposeDiffResult, error) {
	// Parse compose file
	composeInfo, err := parser.ParseComposeFile(composeFile)
	if err != nil {
		return nil, fmt.Errorf("failed to parse compose file: %w", err)
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

	return compareComposeWithEnvVars(composeInfo, allEnvVars), nil
}

// compareComposeWithEnvVars performs the actual comparison logic
func compareComposeWithEnvVars(composeInfo *parser.ComposeEnvInfo, envVars parser.EnvVars) *ComposeDiffResult {
	result := &ComposeDiffResult{
		MissingInEnv:     []string{},
		ExtraInEnv:       []string{},
		MissingEnvFiles:  []string{},
		ServiceBreakdown: make(map[string][]string),
	}

	// Get all variables referenced in compose
	composeVars := composeInfo.GetAllEnvVars()
	composeVarSet := make(map[string]bool)
	for _, v := range composeVars {
		composeVarSet[v] = true
	}

	// Find missing variables (in compose but not in env)
	for _, composeVar := range composeVars {
		if !envVars.Has(composeVar) {
			result.MissingInEnv = append(result.MissingInEnv, composeVar)
		}
	}

	// Find extra variables (in env but not in compose)
	for envVar := range envVars {
		if !composeVarSet[envVar] {
			result.ExtraInEnv = append(result.ExtraInEnv, envVar)
		}
	}

	// Check service-specific breakdowns
	for serviceName, serviceVars := range composeInfo.ServiceVars {
		missing := []string{}
		for varName := range serviceVars {
			if !envVars.Has(varName) {
				missing = append(missing, varName)
			}
		}
		if len(missing) > 0 {
			sort.Strings(missing)
			result.ServiceBreakdown[serviceName] = missing
		}
	}

	// Check for missing env files referenced in compose
	for _, envFile := range composeInfo.EnvFiles {
		if _, err := parser.ParseEnvFile(envFile); err != nil {
			result.MissingEnvFiles = append(result.MissingEnvFiles, envFile)
		}
	}

	// Sort results
	sort.Strings(result.MissingInEnv)
	sort.Strings(result.ExtraInEnv)
	sort.Strings(result.MissingEnvFiles)

	return result
}

// GenerateComposeReport creates a formatted report for compose comparison
func GenerateComposeReport(result *ComposeDiffResult, opts *ReportOptions) string {
	if opts == nil {
		opts = DefaultReportOptions()
	}

	var report strings.Builder

	if !result.HasIssues() {
		report.WriteString("✅ Docker Compose environment is aligned.\n")
		if opts.ShowDuck {
			report.WriteString("(Your gopher-duck approves of your container setup!)\n")
		}
		return report.String()
	}

	// Header with duck
	if opts.ShowDuck {
		report.WriteString(quack.GetAngryDuck() + "\n")
		report.WriteString("QUACK! 🦆 Docker Compose environment issues detected:\n\n")
	}

	// Missing env files
	if len(result.MissingEnvFiles) > 0 {
		if opts.Colorize {
			report.WriteString("💥 Missing env_files referenced in compose:\n")
		} else {
			report.WriteString("Missing env_files:\n")
		}

		for _, file := range result.MissingEnvFiles {
			report.WriteString(fmt.Sprintf("  - %s\n", file))
		}
		report.WriteString("\n")
	}

	// Missing variables
	if len(result.MissingInEnv) > 0 {
		if opts.Colorize {
			report.WriteString("🔴 Variables required by compose but missing in env files:\n")
		} else {
			report.WriteString("Missing variables:\n")
		}

		for _, key := range result.MissingInEnv {
			report.WriteString(fmt.Sprintf("  - %s\n", key))
		}
		report.WriteString("\n")
	}

	// Service breakdown
	if len(result.ServiceBreakdown) > 0 && opts.Verbose {
		report.WriteString("📋 Service breakdown:\n")
		for serviceName, missing := range result.ServiceBreakdown {
			report.WriteString(fmt.Sprintf("  %s:\n", serviceName))
			for _, varName := range missing {
				report.WriteString(fmt.Sprintf("    - %s\n", varName))
			}
		}
		report.WriteString("\n")
	}

	// Extra variables (usually less critical)
	if len(result.ExtraInEnv) > 0 {
		if opts.Colorize {
			report.WriteString("🟡 Variables in env files but not used in compose:\n")
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
		report.WriteString("(Your gopher-duck is confused by your container setup!)\n")
	}

	return report.String()
}
