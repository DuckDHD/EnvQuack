package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/DuckDHD/EnvQuack/internal/backup"
	"github.com/DuckDHD/EnvQuack/internal/checker"
	"github.com/DuckDHD/EnvQuack/internal/errors"
	"github.com/DuckDHD/EnvQuack/internal/masking"
	"github.com/DuckDHD/EnvQuack/internal/parser"
	"github.com/DuckDHD/EnvQuack/internal/quack"
	"github.com/DuckDHD/EnvQuack/internal/security"
	"github.com/spf13/cobra"
)

var (
	envFile          string
	exampleFile      string
	composeFile      string
	dockerfileFile   string
	verbose          bool
	noColor          bool
	noDuck           bool
	allowUnsafePaths bool
	noBackup         bool
	cleanupBackups   bool
	showSecrets      bool
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
	Short: "Comprehensive audit of env files vs docker-compose and Dockerfile requirements",
	Long: `Audit performs a comprehensive check across multiple sources:

- Compares .env files against .env.example
- Analyzes docker-compose.yml environment requirements  
- Checks for missing env_file references
- Analyzes Dockerfile ARG and ENV instructions
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

var restoreCmd = &cobra.Command{
	Use:   "restore [backup-file]",
	Short: "Restore .env from a backup file",
	Long: `Restore your .env file from a previous backup.

Examples:
  # List available backups
  envquack restore --list

  # Restore from most recent backup
  envquack restore --latest

  # Restore from specific backup
  envquack restore .env.backup.20241027-143052`,
	RunE: runRestore,
}

var (
	listBackupsFlag bool
	latestBackup    bool
)

func init() {
	// Global flags
	rootCmd.PersistentFlags().StringVar(&envFile, "env", ".env", "path to .env file")
	rootCmd.PersistentFlags().StringVar(&exampleFile, "example", ".env.example", "path to .env.example file")
	rootCmd.PersistentFlags().StringVar(&composeFile, "compose", "docker-compose.yml", "path to docker-compose file")
	rootCmd.PersistentFlags().StringVar(&dockerfileFile, "dockerfile", "Dockerfile", "path to Dockerfile")
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "verbose output")
	rootCmd.PersistentFlags().BoolVar(&noColor, "no-color", false, "disable colored output")
	rootCmd.PersistentFlags().BoolVar(&noDuck, "no-duck", false, "disable ASCII duck art")
	rootCmd.PersistentFlags().BoolVar(&allowUnsafePaths, "allow-unsafe-paths", false, "allow access to files outside current directory (USE WITH CAUTION)")
	rootCmd.PersistentFlags().BoolVar(&showSecrets, "show-secrets", false, "show actual sensitive values (WARNING: insecure, use only for debugging)")

	// Sync command specific flags
	syncCmd.Flags().BoolVar(&noBackup, "no-backup", false, "skip creating backup before sync (not recommended)")
	syncCmd.Flags().BoolVar(&cleanupBackups, "cleanup-backups", false, "automatically cleanup old backups (keeps last 5, deletes >30 days)")

	// Restore command specific flags
	restoreCmd.Flags().BoolVar(&listBackupsFlag, "list", false, "list all available backups")
	restoreCmd.Flags().BoolVar(&latestBackup, "latest", false, "restore from most recent backup")

	// Add commands
	rootCmd.AddCommand(checkCmd)
	rootCmd.AddCommand(syncCmd)
	rootCmd.AddCommand(auditCmd)
	rootCmd.AddCommand(restoreCmd)
}

// Execute runs the root command
func Execute() error {
	return rootCmd.Execute()
}

// validatePathOrSkip validates a file path unless allowUnsafePaths is set
func validatePathOrSkip(path string) error {
	if allowUnsafePaths {
		fmt.Fprintf(os.Stderr, "⚠️  WARNING: Path validation disabled for: %s\n", path)
		return nil
	}
	return security.ValidateFilePath(path)
}

func runCheck(cmd *cobra.Command, args []string) error {
	// Set masking based on flag
	masking.SetMaskingEnabled(!showSecrets)

	// Warn if showing secrets
	if showSecrets {
		fmt.Fprintln(os.Stderr, "⚠️  WARNING: Showing actual sensitive values (insecure)")
	}

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
		displayError(err)
		return fmt.Errorf("failed to compare files")
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
	// Validate paths for security before any write operations
	if err := validatePathOrSkip(envFile); err != nil {
		return fmt.Errorf("invalid env file path: %w", err)
	}

	// Verify parent directory is writable
	dir := filepath.Dir(envFile)
	if dir == "" || dir == "." {
		// Current directory
		dir = "."
	}
	if err := security.ValidateDirectoryWritable(dir); err != nil {
		if !allowUnsafePaths {
			return fmt.Errorf("cannot write to directory: %w", err)
		}
		// If unsafe paths allowed, just warn
		fmt.Fprintf(os.Stderr, "⚠️  WARNING: Directory write validation bypassed\n")
	}

	// Check if example file exists
	if err := checkFileExists(exampleFile); err != nil {
		return fmt.Errorf("example file error: %w", err)
	}

	// Parse example file
	example, err := parser.ParseEnvFile(exampleFile)
	if err != nil {
		displayError(err)
		return fmt.Errorf("failed to parse example file")
	}

	// Parse existing env file (create if doesn't exist)
	var env parser.EnvVars
	if _, err := os.Stat(envFile); os.IsNotExist(err) {
		env = make(parser.EnvVars)
		fmt.Printf("Creating new %s file...\n", envFile)
	} else {
		env, err = parser.ParseEnvFile(envFile)
		if err != nil {
			displayError(err)
			return fmt.Errorf("failed to parse env file")
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

	// Create backup before modifying file
	if !noBackup {
		backupPath, err := backup.BackupFile(envFile)
		if err != nil {
			return fmt.Errorf("failed to create backup: %w", err)
		}

		if backupPath != "" {
			fmt.Printf("📦 Backup created: %s\n", filepath.Base(backupPath))
		}
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

	// Cleanup old backups if requested
	if cleanupBackups {
		if err := backup.CleanOldBackups(envFile, 30, 5); err != nil {
			fmt.Fprintf(os.Stderr, "⚠️  Warning: failed to cleanup old backups: %v\n", err)
		} else {
			backups, _ := backup.ListBackups(envFile)
			fmt.Printf("🧹 Cleaned up old backups (kept %d most recent)\n", len(backups))
		}
	}

	return nil
}

func runRestore(cmd *cobra.Command, args []string) error {
	// List backups mode
	if listBackupsFlag {
		backups, err := backup.ListBackups(envFile)
		if err != nil {
			return fmt.Errorf("failed to list backups: %w", err)
		}

		if len(backups) == 0 {
			fmt.Println("No backups found for", envFile)
			return nil
		}

		fmt.Printf("Available backups for %s:\n", envFile)
		for i, b := range backups {
			info, err := os.Stat(b)
			if err != nil {
				continue
			}
			fmt.Printf("  %d. %s (%s, %d bytes)\n",
				i+1,
				filepath.Base(b),
				info.ModTime().Format("2006-01-02 15:04:05"),
				info.Size())
		}
		return nil
	}

	// Restore from latest backup
	if latestBackup {
		backups, err := backup.ListBackups(envFile)
		if err != nil {
			return fmt.Errorf("failed to list backups: %w", err)
		}

		if len(backups) == 0 {
			return fmt.Errorf("no backups found for %s", envFile)
		}

		backupFile := backups[0]
		fmt.Printf("Restoring from latest backup: %s\n", filepath.Base(backupFile))

		if err := backup.RestoreBackup(backupFile, envFile); err != nil {
			return fmt.Errorf("failed to restore backup: %w", err)
		}

		fmt.Printf("✅ Successfully restored %s from %s\n", envFile, filepath.Base(backupFile))
		return nil
	}

	// Restore from specific backup file
	if len(args) == 0 {
		return fmt.Errorf("please specify a backup file or use --list or --latest\n\nUsage:\n  envquack restore --list\n  envquack restore --latest\n  envquack restore <backup-file>")
	}

	backupFile := args[0]

	// Validate backup file exists
	if _, err := os.Stat(backupFile); os.IsNotExist(err) {
		return fmt.Errorf("backup file does not exist: %s", backupFile)
	}

	fmt.Printf("Restoring %s from %s\n", envFile, filepath.Base(backupFile))

	if err := backup.RestoreBackup(backupFile, envFile); err != nil {
		return fmt.Errorf("failed to restore backup: %w", err)
	}

	fmt.Printf("✅ Successfully restored %s from %s\n", envFile, filepath.Base(backupFile))
	return nil
}

func runAudit(cmd *cobra.Command, args []string) error {
	// Set masking based on flag
	masking.SetMaskingEnabled(!showSecrets)

	// Warn if showing secrets
	if showSecrets {
		fmt.Fprintln(os.Stderr, "⚠️  WARNING: Showing actual sensitive values (insecure)")
	}

	fmt.Println("🔍 Running comprehensive environment audit...")

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
			fmt.Print("  ❌ Error parsing compose file:\n  ")
			// Indent the error output
			if parseErr, ok := err.(*errors.ParseError); ok {
				errorOutput := parseErr.Error()
				if !noColor {
					errorOutput = parseErr.ErrorWithColor()
				}
				fmt.Fprintln(os.Stderr, strings.ReplaceAll(errorOutput, "\n", "\n  "))
			} else {
				fmt.Fprintf(os.Stderr, "%v\n", err)
			}
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

	// 3. Dockerfile environment check
	if err := checkFileExists(dockerfileFile); err == nil {
		fmt.Println("🐋 Checking Dockerfile environment requirements:")

		envFiles := []string{}
		if fileExists(envFile) {
			envFiles = append(envFiles, envFile)
		}

		dockerfileResult, err := checker.CompareDockerfileWithEnv(dockerfileFile, envFiles)
		if err != nil {
			fmt.Print("  ❌ Error parsing Dockerfile:\n  ")
			// Indent the error output
			if parseErr, ok := err.(*errors.ParseError); ok {
				errorOutput := parseErr.Error()
				if !noColor {
					errorOutput = parseErr.ErrorWithColor()
				}
				fmt.Fprintln(os.Stderr, strings.ReplaceAll(errorOutput, "\n", "\n  "))
			} else {
				fmt.Fprintf(os.Stderr, "%v\n", err)
			}
			hasErrors = true
		} else {
			opts := &checker.ReportOptions{
				ShowDuck: false,
				Colorize: !noColor,
				Verbose:  verbose,
			}

			if !dockerfileResult.HasIssues() {
				fmt.Println("  ✅ Dockerfile check passed")
			} else {
				report := checker.GenerateDockerfileReport(dockerfileResult, opts)
				fmt.Print("  " + strings.ReplaceAll(report, "\n", "\n  "))
				hasErrors = true
			}
		}
		fmt.Println()
	} else {
		fmt.Printf("  ℹ️  No Dockerfile found, skipping Dockerfile check\n\n")
	}

	// 4. Summary
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

// displayError displays an error with enhanced formatting if it's a ParseError
func displayError(err error) {
	if parseErr, ok := err.(*errors.ParseError); ok {
		// Use colored output unless --no-color flag is set
		if noColor {
			fmt.Fprintln(os.Stderr, parseErr.Error())
		} else {
			fmt.Fprintln(os.Stderr, parseErr.ErrorWithColor())
		}
	} else {
		// For non-ParseError errors, just print normally
		fmt.Fprintln(os.Stderr, err)
	}
}
