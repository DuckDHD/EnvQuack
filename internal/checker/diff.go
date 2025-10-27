package checker

import (
	"sort"

	"github.com/DuckDHD/EnvQuack/internal/parser"
)

// DiffResult represents the difference between two sets of environment variables
type DiffResult struct {
	Missing       []string          // Keys present in example but missing in env
	Extra         []string          // Keys present in env but not in example
	ExampleValues map[string]string // Example values for missing keys
	EnvValues     map[string]string // Actual values for extra keys
}

// HasIssues returns true if there are any differences
func (d *DiffResult) HasIssues() bool {
	return len(d.Missing) > 0 || len(d.Extra) > 0
}

// CompareEnvFiles compares .env file against .env.example
func CompareEnvFiles(envFile, exampleFile string) (*DiffResult, error) {
	env, err := parser.ParseEnvFile(envFile)
	if err != nil {
		return nil, err
	}

	example, err := parser.ParseEnvFile(exampleFile)
	if err != nil {
		return nil, err
	}

	return CompareEnvVars(env, example), nil
}

// CompareEnvVars compares two sets of environment variables
func CompareEnvVars(env, example parser.EnvVars) *DiffResult {
	result := &DiffResult{
		Missing:       []string{},
		Extra:         []string{},
		ExampleValues: make(map[string]string),
		EnvValues:     make(map[string]string),
	}

	// Find missing vars (in example but not in env)
	for key := range example {
		if !env.Has(key) {
			result.Missing = append(result.Missing, key)
			// Store the example value for verbose display
			result.ExampleValues[key] = example[key]
		}
	}

	// Find extra vars (in env but not in example)
	for key := range env {
		if !example.Has(key) {
			result.Extra = append(result.Extra, key)
			// Store the actual value for verbose display
			result.EnvValues[key] = env[key]
		}
	}

	// Sort for consistent output
	sort.Strings(result.Missing)
	sort.Strings(result.Extra)

	return result
}
