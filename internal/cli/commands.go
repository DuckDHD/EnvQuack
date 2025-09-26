package cli

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/DuckDHD/EnvQuack/internal/checker"
	"github.com/DuckDHD/EnvQuack/internal/parser"
	"github.com/DuckDHD/EnvQuack/internal/quack"
)

var (
	envFile     string
	exampleFile string
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

// syncCmd represents the sync command
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
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "verbose output")
	rootCmd.PersistentFlags().BoolVar(&noColor, "no-color", false, "disable colored output")
	rootCmd.PersistentFlags().BoolVar(&noDuck, "no-duck", false, "disable ASCII duck art")

	// Add commands
	rootCmd.AddCommand(checkCmd)
	rootCmd.AddCommand(syncCmd)
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

func checkFileExists(filename string) error {
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		return fmt.Errorf("file %s does not exist", filename)
	}
	return nil
}
