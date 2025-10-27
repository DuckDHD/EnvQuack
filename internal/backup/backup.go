package backup

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

// BackupFile creates a timestamped backup of the specified file.
// Returns the backup filename and any error.
// If the original file doesn't exist, returns empty string and no error.
func BackupFile(filename string) (string, error) {
	// Check if original file exists
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		// Nothing to backup, not an error
		return "", nil
	}
	if err != nil {
		return "", fmt.Errorf("failed to stat file %s: %w", filename, err)
	}

	// Read original file contents
	content, err := os.ReadFile(filename)
	if err != nil {
		return "", fmt.Errorf("failed to read file %s: %w", filename, err)
	}

	// Get original file permissions
	permissions := info.Mode()

	// Generate backup filename with timestamp
	backupFilename := generateBackupFilename(filename)

	// Ensure we don't have a collision (unlikely but possible)
	for fileExists(backupFilename) {
		time.Sleep(1 * time.Millisecond)
		backupFilename = generateBackupFilename(filename)
	}

	// Get directory for temporary file (same directory as original for atomic rename)
	dir := filepath.Dir(filename)

	// Write to temporary file first (atomic operation)
	tempFile, err := os.CreateTemp(dir, ".envquack-backup-*.tmp")
	if err != nil {
		return "", fmt.Errorf("failed to create temp file: %w", err)
	}
	tempFilename := tempFile.Name()

	// Clean up temp file on error
	defer func() {
		if err != nil {
			os.Remove(tempFilename)
		}
	}()

	// Write contents
	if _, err = tempFile.Write(content); err != nil {
		tempFile.Close()
		return "", fmt.Errorf("failed to write backup: %w", err)
	}

	// Close before changing permissions
	if err = tempFile.Close(); err != nil {
		return "", fmt.Errorf("failed to close temp file: %w", err)
	}

	// Set permissions to match original
	if err = os.Chmod(tempFilename, permissions); err != nil {
		return "", fmt.Errorf("failed to set permissions: %w", err)
	}

	// Rename temp file to final backup name (atomic operation)
	if err = os.Rename(tempFilename, backupFilename); err != nil {
		return "", fmt.Errorf("failed to rename backup: %w", err)
	}

	return backupFilename, nil
}

// RestoreBackup restores a file from a backup.
func RestoreBackup(backupFile, originalFile string) error {
	// Validate backup file exists
	if _, err := os.Stat(backupFile); os.IsNotExist(err) {
		return fmt.Errorf("backup file does not exist: %s", backupFile)
	} else if err != nil {
		return fmt.Errorf("failed to stat backup file: %w", err)
	}

	// Read backup contents
	content, err := os.ReadFile(backupFile)
	if err != nil {
		return fmt.Errorf("failed to read backup file: %w", err)
	}

	// Get backup file permissions
	info, err := os.Stat(backupFile)
	if err != nil {
		return fmt.Errorf("failed to get backup file info: %w", err)
	}
	permissions := info.Mode()

	// Get directory for temporary file (same directory for atomic rename)
	dir := filepath.Dir(originalFile)

	// Write to temporary file first
	tempFile, err := os.CreateTemp(dir, ".envquack-restore-*.tmp")
	if err != nil {
		return fmt.Errorf("failed to create temp file: %w", err)
	}
	tempFilename := tempFile.Name()

	// Clean up temp file on error
	defer func() {
		if err != nil {
			os.Remove(tempFilename)
		}
	}()

	// Write contents
	if _, err = tempFile.Write(content); err != nil {
		tempFile.Close()
		return fmt.Errorf("failed to write restore: %w", err)
	}

	// Close before changing permissions
	if err = tempFile.Close(); err != nil {
		return fmt.Errorf("failed to close temp file: %w", err)
	}

	// Set permissions to match backup
	if err = os.Chmod(tempFilename, permissions); err != nil {
		return fmt.Errorf("failed to set permissions: %w", err)
	}

	// Rename temp file to original file (atomic operation)
	if err = os.Rename(tempFilename, originalFile); err != nil {
		return fmt.Errorf("failed to rename restore: %w", err)
	}

	return nil
}

// ListBackups lists all backup files for a given file.
// Returns backups sorted by timestamp (newest first).
func ListBackups(filename string) ([]string, error) {
	// Get directory and base filename
	dir := filepath.Dir(filename)
	baseName := filepath.Base(filename)

	// List all files in directory
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("failed to read directory: %w", err)
	}

	// Filter for backup pattern: basename.backup.*
	var backups []string
	prefix := baseName + ".backup."

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		name := entry.Name()
		if strings.HasPrefix(name, prefix) {
			fullPath := filepath.Join(dir, name)
			backups = append(backups, fullPath)
		}
	}

	// Sort by timestamp (newest first)
	// Backup format: .env.backup.20241027-143052
	// Lexicographic sort works because of YYYYMMDD-HHMMSS format
	sort.Slice(backups, func(i, j int) bool {
		return backups[i] > backups[j] // Descending order (newest first)
	})

	return backups, nil
}

// CleanOldBackups removes backups older than specified days, keeping at least minKeep backups.
// maxAgeDays: maximum age in days for backups to keep
// minKeep: minimum number of backups to keep regardless of age
func CleanOldBackups(filename string, maxAgeDays int, minKeep int) error {
	// List all backups (sorted newest first)
	backups, err := ListBackups(filename)
	if err != nil {
		return err
	}

	// Nothing to clean
	if len(backups) <= minKeep {
		return nil
	}

	// Calculate cutoff time
	cutoffTime := time.Now().AddDate(0, 0, -maxAgeDays)

	// Always keep the minKeep most recent backups
	backupsToConsider := backups[minKeep:]

	// Delete backups older than cutoff
	for _, backup := range backupsToConsider {
		info, err := os.Stat(backup)
		if err != nil {
			// Skip if we can't stat (might have been deleted)
			continue
		}

		if info.ModTime().Before(cutoffTime) {
			if err := os.Remove(backup); err != nil {
				// Log error but continue with other backups
				return fmt.Errorf("failed to remove backup %s: %w", backup, err)
			}
		}
	}

	return nil
}

// generateBackupFilename creates a timestamped backup filename.
// Format: original.backup.20241027-143052
func generateBackupFilename(filename string) string {
	timestamp := time.Now().Format("20060102-150405")
	return fmt.Sprintf("%s.backup.%s", filename, timestamp)
}

// fileExists checks if a file exists.
func fileExists(filename string) bool {
	_, err := os.Stat(filename)
	return err == nil
}
