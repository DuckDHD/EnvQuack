package cli

import (
	"fmt"
	"os"
	"strings"

	"github.com/DuckDHD/EnvQuack/internal/checker"
	"github.com/DuckDHD/EnvQuack/internal/parser"
	"github.com/DuckDHD/EnvQuack/internal/quack"
	"github.com/spf13/cobra"
)

var (
	envFile     string
	exampleFile string
	composeFile string
	verbose     bool
	noColor     bool
	noDuck      bool
)

// rootCmd represents the base command
var rootCmd = &cobra.Command{
	Use:   "envquack",
	Short: "Environment Variable Drift Detective 🦆",
	Long:  quack.GetBanner() + "\nEnvQuack helps you keep your environment variables in sync.",
}

// checkCmd represents the check command
var checkCmd = &cobra.Command{
	Use:   "check",
	Short: "Check for differences between .env and .env.example",
	Long: `Check compares your .env file against .env.example and reports any differences.

This includes:
- Missing variables (present in example but not in .env)  
- Extra variables (present in .env but not in example)`,
	RunE: runCheck,
}

// auditCmd represents the audit command
var auditCmd = &cobra.Command{
	Use:   "audit",
	Short: "Comprehensive audit of env files vs docker-compose requirements",
	Long: `Audit performs a comprehensive check across multiple sources:

- Compares .env files against .env.example
- Analyzes docker-compose.yml environment requirements  
- Checks for missing env_file references
- Shows service-by-service breakdown

This gives you a complete picture of your environment configuration.`,
	RunE: runAudit,
}
var syncCmd = &cobra.Command{
	Use:   "sync",
	Short: "Sync missing variables from .env.example to .env",
	Long: `Sync adds missing variables from .env.example to your .env file with empty values.

This helps you quickly scaffold your .env file based on the example.`,
	RunE: runSync,
}

func init() {
	// Global flags
	rootCmd.PersistentFlags().StringVar(&envFile, "env", ".env", "path to .env file")
	rootCmd.PersistentFlags().StringVar(&exampleFile, "example", ".env.example", "path to .env.example file")
	rootCmd.PersistentFlags().StringVar(&composeFile, "compose", "docker-compose.yml", "path to docker-compose file")
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "verbose output")
	rootCmd.PersistentFlags().BoolVar(&noColor, "no-color", false, "disable colored output")
	rootCmd.PersistentFlags().BoolVar(&noDuck, "no-duck", false, "disable ASCII duck art")

	// Add commands
	rootCmd.AddCommand(checkCmd)
	rootCmd.AddCommand(syncCmd)
	rootCmd.AddCommand(auditCmd)
}

// Execute runs the root command
func Execute() error {
	return rootCmd.Execute()
}

func runCheck(cmd *cobra.Command, args []string) error {
	// Check if files exist
	if err := checkFileExists(exampleFile); err != nil {
		return fmt.Errorf("example file error: %w", err)
	}

	if err := checkFileExists(envFile); err != nil {
		return fmt.Errorf("env file error: %w", err)
	}

	// Compare files
	result, err := checker.CompareEnvFiles(envFile, exampleFile)
	if err != nil {
		return fmt.Errorf("failed to compare files: %w", err)
	}

	// Generate and display report
	opts := &checker.ReportOptions{
		ShowDuck: !noDuck,
		Colorize: !noColor,
		Verbose:  verbose,
	}

	report := checker.GenerateReport(result, opts)
	fmt.Print(report)

	// Exit with error code if issues found
	if result.HasIssues() {
		os.Exit(1)
	}

	return nil
}

func runSync(cmd *cobra.Command, args []string) error {
	// Check if example file exists
	if err := checkFileExists(exampleFile); err != nil {
		return fmt.Errorf("example file error: %w", err)
	}

	// Parse example file
	example, err := parser.ParseEnvFile(exampleFile)
	if err != nil {
		return fmt.Errorf("failed to parse example file: %w", err)
	}

	// Parse existing env file (create if doesn't exist)
	var env parser.EnvVars
	if _, err := os.Stat(envFile); os.IsNotExist(err) {
		env = make(parser.EnvVars)
		fmt.Printf("Creating new %s file...\n", envFile)
	} else {
		env, err = parser.ParseEnvFile(envFile)
		if err != nil {
			return fmt.Errorf("failed to parse env file: %w", err)
		}
	}

	// Find missing variables
	result := checker.CompareEnvVars(env, example)

	if len(result.Missing) == 0 {
		fmt.Println("✅ No missing variables to sync.")
		if !noDuck {
			fmt.Println("(Your gopher-duck is already happy!)")
		}
		return nil
	}

	// Show sync message
	if !noDuck {
		fmt.Println(quack.GetSyncMessage())
	}
	fmt.Printf("Adding %d missing variables to %s:\n", len(result.Missing), envFile)

	// Append missing variables to env file
	file, err := os.OpenFile(envFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("failed to open env file for writing: %w", err)
	}
	defer file.Close()

	// Add a separator comment if file already has content
	if len(env) > 0 {
		file.WriteString("\n# Added by envquack sync\n")
	}

	for _, key := range result.Missing {
		line := fmt.Sprintf("%s=\n", key)
		if _, err := file.WriteString(line); err != nil {
			return fmt.Errorf("failed to write variable %s: %w", key, err)
		}
		fmt.Printf("  + %s\n", key)
	}

	fmt.Printf("\n✅ Successfully synced %d variables!\n", len(result.Missing))
	fmt.Println("Don't forget to set the actual values in your .env file.")

	return nil
}

func runAudit(cmd *cobra.Command, args []string) error {
	fmt.Println("🔍 Running comprehensive environment audit...\n")

	hasErrors := false

	// 1. Basic .env vs .env.example check
	if err := checkFileExists(exampleFile); err == nil && fileExists(envFile) {
		fmt.Println("📋 Checking .env vs .env.example:")
		result, err := checker.CompareEnvFiles(envFile, exampleFile)
		if err != nil {
			fmt.Printf("  ❌ Error: %v\n", err)
			hasErrors = true
		} else {
			opts := &checker.ReportOptions{
				ShowDuck: false,
				Colorize: !noColor,
				Verbose:  false,
			}

			if !result.HasIssues() {
				fmt.Println("  ✅ Basic env check passed")
			} else {
				fmt.Print("  " + strings.ReplaceAll(checker.GenerateReport(result, opts), "\n", "\n  "))
				hasErrors = true
			}
		}
		fmt.Println()
	}

	// 2. Docker Compose environment check
	if err := checkFileExists(composeFile); err == nil {
		fmt.Println("🐳 Checking docker-compose environment requirements:")

		envFiles := []string{}
		if fileExists(envFile) {
			envFiles = append(envFiles, envFile)
		}

		composeResult, err := checker.CompareComposeWithEnv(composeFile, envFiles)
		if err != nil {
			fmt.Printf("  ❌ Error parsing compose file: %v\n", err)
			hasErrors = true
		} else {
			opts := &checker.ReportOptions{
				ShowDuck: false,
				Colorize: !noColor,
				Verbose:  verbose,
			}

			if !composeResult.HasIssues() {
				fmt.Println("  ✅ Docker Compose check passed")
			} else {
				report := checker.GenerateComposeReport(composeResult, opts)
				fmt.Print("  " + strings.ReplaceAll(report, "\n", "\n  "))
				hasErrors = true
			}
		}
		fmt.Println()
	} else {
		fmt.Printf("  ℹ️  No docker-compose.yml found, skipping compose check\n\n")
	}

	// 3. Summary
	if !noDuck {
		if hasErrors {
			fmt.Println(quack.GetAngryDuck())
			fmt.Println("QUACK! 🦆 Audit found issues that need attention!")
		} else {
			fmt.Println(quack.GetHappyDuck())
			fmt.Println("✅ Audit passed! Your environment is well organized.")
		}
	} else {
		if hasErrors {
			fmt.Println("❌ Audit found issues that need attention!")
		} else {
			fmt.Println("✅ Audit passed! Your environment is well organized.")
		}
	}

	if hasErrors {
		os.Exit(1)
	}

	return nil
}

func checkFileExists(filename string) error {
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		return fmt.Errorf("file %s does not exist", filename)
	}
	return nil
}

// fileExists is a helper that returns true if file exists, false otherwise
func fileExists(filename string) bool {
	_, err := os.Stat(filename)
	return err == nil
}
