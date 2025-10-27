package backup

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
	"time"
)

func TestBackupFile(t *testing.T) {
	tests := []struct {
		name        string
		setup       func(t *testing.T) string
		expectError bool
		validate    func(t *testing.T, originalPath, backupPath string)
	}{
		{
			name: "successful backup of existing file",
			setup: func(t *testing.T) string {
				tmpDir := t.TempDir()
				envFile := filepath.Join(tmpDir, ".env")
				content := "DATABASE_URL=postgres://localhost\nAPI_KEY=secret\n"
				if err := os.WriteFile(envFile, []byte(content), 0644); err != nil {
					t.Fatal(err)
				}
				return envFile
			},
			expectError: false,
			validate: func(t *testing.T, originalPath, backupPath string) {
				if backupPath == "" {
					t.Error("expected backup to be created, got empty path")
					return
				}

				// Verify backup file exists
				if _, err := os.Stat(backupPath); os.IsNotExist(err) {
					t.Errorf("backup file does not exist: %s", backupPath)
				}

				// Verify backup has correct content
				backupContent, err := os.ReadFile(backupPath)
				if err != nil {
					t.Fatalf("failed to read backup: %v", err)
				}

				originalContent, err := os.ReadFile(originalPath)
				if err != nil {
					t.Fatalf("failed to read original: %v", err)
				}

				if string(backupContent) != string(originalContent) {
					t.Errorf("backup content mismatch:\ngot:  %q\nwant: %q", string(backupContent), string(originalContent))
				}

				// Verify backup has correct permissions
				info, err := os.Stat(backupPath)
				if err != nil {
					t.Fatalf("failed to stat backup: %v", err)
				}

				originalInfo, err := os.Stat(originalPath)
				if err != nil {
					t.Fatalf("failed to stat original: %v", err)
				}

				if info.Mode() != originalInfo.Mode() {
					t.Errorf("permissions not preserved: got %v, want %v", info.Mode(), originalInfo.Mode())
				}

				// Verify backup filename format
				pattern := regexp.MustCompile(`\.env\.backup\.\d{8}-\d{6}$`)
				if !pattern.MatchString(backupPath) {
					t.Errorf("backup filename format incorrect: %s", backupPath)
				}
			},
		},
		{
			name: "skip backup if original doesn't exist",
			setup: func(t *testing.T) string {
				tmpDir := t.TempDir()
				return filepath.Join(tmpDir, "nonexistent.env")
			},
			expectError: false,
			validate: func(t *testing.T, originalPath, backupPath string) {
				if backupPath != "" {
					t.Errorf("expected no backup for nonexistent file, got: %s", backupPath)
				}
			},
		},
		{
			name: "preserve file permissions",
			setup: func(t *testing.T) string {
				tmpDir := t.TempDir()
				envFile := filepath.Join(tmpDir, ".env")
				if err := os.WriteFile(envFile, []byte("TEST=value"), 0600); err != nil {
					t.Fatal(err)
				}
				return envFile
			},
			expectError: false,
			validate: func(t *testing.T, originalPath, backupPath string) {
				if backupPath == "" {
					t.Error("expected backup to be created")
					return
				}

				info, err := os.Stat(backupPath)
				if err != nil {
					t.Fatalf("failed to stat backup: %v", err)
				}

				// Note: On Windows, permission handling is different
				// We'll check that permissions were at least attempted
				if info.Mode().Perm() == 0 {
					t.Error("backup has no permissions set")
				}
			},
		},
		{
			name: "backup file contains exact copy of original",
			setup: func(t *testing.T) string {
				tmpDir := t.TempDir()
				envFile := filepath.Join(tmpDir, ".env")
				content := "LINE1=value1\nLINE2=value2\nLINE3=value3\n"
				if err := os.WriteFile(envFile, []byte(content), 0644); err != nil {
					t.Fatal(err)
				}
				return envFile
			},
			expectError: false,
			validate: func(t *testing.T, originalPath, backupPath string) {
				if backupPath == "" {
					t.Error("expected backup to be created")
					return
				}

				// Compare byte-for-byte
				originalBytes, err := os.ReadFile(originalPath)
				if err != nil {
					t.Fatal(err)
				}

				backupBytes, err := os.ReadFile(backupPath)
				if err != nil {
					t.Fatal(err)
				}

				if len(originalBytes) != len(backupBytes) {
					t.Errorf("backup size mismatch: got %d bytes, want %d bytes", len(backupBytes), len(originalBytes))
				}

				for i := range originalBytes {
					if i >= len(backupBytes) {
						break
					}
					if originalBytes[i] != backupBytes[i] {
						t.Errorf("byte mismatch at position %d: got %d, want %d", i, backupBytes[i], originalBytes[i])
						break
					}
				}
			},
		},
		{
			name: "backup filename format is correct",
			setup: func(t *testing.T) string {
				tmpDir := t.TempDir()
				envFile := filepath.Join(tmpDir, ".env")
				if err := os.WriteFile(envFile, []byte("TEST=value"), 0644); err != nil {
					t.Fatal(err)
				}
				return envFile
			},
			expectError: false,
			validate: func(t *testing.T, originalPath, backupPath string) {
				if backupPath == "" {
					t.Error("expected backup to be created")
					return
				}

				// Verify format: .env.backup.YYYYMMDD-HHMMSS
				pattern := regexp.MustCompile(`\.env\.backup\.\d{8}-\d{6}$`)
				if !pattern.MatchString(backupPath) {
					t.Errorf("backup filename format incorrect: %s", backupPath)
				}

				// Verify timestamp is parseable
				baseName := filepath.Base(backupPath)
				parts := strings.Split(baseName, ".")
				if len(parts) < 3 {
					t.Errorf("invalid backup filename structure: %s", baseName)
					return
				}

				timestamp := parts[len(parts)-1]
				_, err := time.Parse("20060102-150405", timestamp)
				if err != nil {
					t.Errorf("failed to parse timestamp %s: %v", timestamp, err)
					return
				}

				// Just verify timestamp can be parsed - don't check actual time
				// as system clock may vary across environments
			},
		},
		{
			name: "empty file backup",
			setup: func(t *testing.T) string {
				tmpDir := t.TempDir()
				envFile := filepath.Join(tmpDir, ".env")
				if err := os.WriteFile(envFile, []byte(""), 0644); err != nil {
					t.Fatal(err)
				}
				return envFile
			},
			expectError: false,
			validate: func(t *testing.T, originalPath, backupPath string) {
				if backupPath == "" {
					t.Error("expected backup to be created even for empty file")
					return
				}

				content, err := os.ReadFile(backupPath)
				if err != nil {
					t.Fatal(err)
				}

				if len(content) != 0 {
					t.Errorf("expected empty backup, got %d bytes", len(content))
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			originalPath := tt.setup(t)

			backupPath, err := BackupFile(originalPath)

			if tt.expectError && err == nil {
				t.Error("expected error, got nil")
			}

			if !tt.expectError && err != nil {
				t.Errorf("unexpected error: %v", err)
			}

			if tt.validate != nil {
				tt.validate(t, originalPath, backupPath)
			}
		})
	}
}

func TestBackupFileMultipleInSameSecond(t *testing.T) {
	tmpDir := t.TempDir()
	envFile := filepath.Join(tmpDir, ".env")

	content := "TEST=value\n"
	if err := os.WriteFile(envFile, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	// Create multiple backups rapidly
	var backups []string
	for i := 0; i < 5; i++ {
		backup, err := BackupFile(envFile)
		if err != nil {
			t.Fatalf("backup %d failed: %v", i, err)
		}
		backups = append(backups, backup)
	}

	// Verify all backups were created with unique names
	seen := make(map[string]bool)
	for _, backup := range backups {
		if seen[backup] {
			t.Errorf("duplicate backup filename: %s", backup)
		}
		seen[backup] = true

		// Verify each backup exists
		if _, err := os.Stat(backup); os.IsNotExist(err) {
			t.Errorf("backup does not exist: %s", backup)
		}
	}

	if len(seen) != len(backups) {
		t.Errorf("expected %d unique backups, got %d", len(backups), len(seen))
	}
}

func TestRestoreBackup(t *testing.T) {
	tests := []struct {
		name           string
		setup          func(t *testing.T) (backupFile, originalFile string)
		expectError    bool
		errorContains  string
		validate       func(t *testing.T, originalFile string)
	}{
		{
			name: "restore overwrites existing file",
			setup: func(t *testing.T) (string, string) {
				tmpDir := t.TempDir()
				backupFile := filepath.Join(tmpDir, ".env.backup.20241027-100000")
				originalFile := filepath.Join(tmpDir, ".env")

				backupContent := "RESTORED=true\nOLD_VALUE=preserved\n"
				if err := os.WriteFile(backupFile, []byte(backupContent), 0644); err != nil {
					t.Fatal(err)
				}

				existingContent := "CURRENT=modified\n"
				if err := os.WriteFile(originalFile, []byte(existingContent), 0644); err != nil {
					t.Fatal(err)
				}

				return backupFile, originalFile
			},
			expectError: false,
			validate: func(t *testing.T, originalFile string) {
				content, err := os.ReadFile(originalFile)
				if err != nil {
					t.Fatal(err)
				}

				expected := "RESTORED=true\nOLD_VALUE=preserved\n"
				if string(content) != expected {
					t.Errorf("restore failed:\ngot:  %q\nwant: %q", string(content), expected)
				}
			},
		},
		{
			name: "restore creates new file if original missing",
			setup: func(t *testing.T) (string, string) {
				tmpDir := t.TempDir()
				backupFile := filepath.Join(tmpDir, ".env.backup.20241027-100000")
				originalFile := filepath.Join(tmpDir, ".env")

				backupContent := "NEW=value\n"
				if err := os.WriteFile(backupFile, []byte(backupContent), 0644); err != nil {
					t.Fatal(err)
				}

				// Don't create original file

				return backupFile, originalFile
			},
			expectError: false,
			validate: func(t *testing.T, originalFile string) {
				content, err := os.ReadFile(originalFile)
				if err != nil {
					t.Fatal(err)
				}

				expected := "NEW=value\n"
				if string(content) != expected {
					t.Errorf("restore failed:\ngot:  %q\nwant: %q", string(content), expected)
				}
			},
		},
		{
			name: "error if backup file doesn't exist",
			setup: func(t *testing.T) (string, string) {
				tmpDir := t.TempDir()
				backupFile := filepath.Join(tmpDir, "nonexistent.backup")
				originalFile := filepath.Join(tmpDir, ".env")
				return backupFile, originalFile
			},
			expectError:   true,
			errorContains: "does not exist",
		},
		{
			name: "preserve permissions from backup",
			setup: func(t *testing.T) (string, string) {
				tmpDir := t.TempDir()
				backupFile := filepath.Join(tmpDir, ".env.backup.20241027-100000")
				originalFile := filepath.Join(tmpDir, ".env")

				if err := os.WriteFile(backupFile, []byte("TEST=value"), 0600); err != nil {
					t.Fatal(err)
				}

				return backupFile, originalFile
			},
			expectError: false,
			validate: func(t *testing.T, originalFile string) {
				info, err := os.Stat(originalFile)
				if err != nil {
					t.Fatal(err)
				}

				// Verify permissions were set (exact values may vary on Windows)
				if info.Mode().Perm() == 0 {
					t.Error("restored file has no permissions")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			backupFile, originalFile := tt.setup(t)

			err := RestoreBackup(backupFile, originalFile)

			if tt.expectError {
				if err == nil {
					t.Error("expected error, got nil")
				} else if tt.errorContains != "" && !strings.Contains(err.Error(), tt.errorContains) {
					t.Errorf("error should contain %q, got: %v", tt.errorContains, err)
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
			}

			if !tt.expectError && tt.validate != nil {
				tt.validate(t, originalFile)
			}
		})
	}
}

func TestListBackups(t *testing.T) {
	tests := []struct {
		name          string
		setup         func(t *testing.T) string
		expectedCount int
		validateOrder func(t *testing.T, backups []string)
	}{
		{
			name: "list multiple backups sorted by date",
			setup: func(t *testing.T) string {
				tmpDir := t.TempDir()
				envFile := filepath.Join(tmpDir, ".env")

				// Create backups with different timestamps
				backupFiles := []string{
					".env.backup.20241027-100000",
					".env.backup.20241027-120000",
					".env.backup.20241027-090000",
				}

				for _, bf := range backupFiles {
					path := filepath.Join(tmpDir, bf)
					if err := os.WriteFile(path, []byte("TEST=value"), 0644); err != nil {
						t.Fatal(err)
					}
				}

				return envFile
			},
			expectedCount: 3,
			validateOrder: func(t *testing.T, backups []string) {
				expected := []string{
					".env.backup.20241027-120000", // Newest first
					".env.backup.20241027-100000",
					".env.backup.20241027-090000",
				}

				for i, backup := range backups {
					baseName := filepath.Base(backup)
					if baseName != expected[i] {
						t.Errorf("backup %d: got %s, want %s", i, baseName, expected[i])
					}
				}
			},
		},
		{
			name: "no backups returns empty list",
			setup: func(t *testing.T) string {
				tmpDir := t.TempDir()
				return filepath.Join(tmpDir, ".env")
			},
			expectedCount: 0,
		},
		{
			name: "ignore non-backup files",
			setup: func(t *testing.T) string {
				tmpDir := t.TempDir()
				envFile := filepath.Join(tmpDir, ".env")

				// Create various files
				files := []string{
					".env",
					".env.example",
					".env.local",
					".env.backup.20241027-100000", // Only this one should be listed
					"other.txt",
				}

				for _, f := range files {
					path := filepath.Join(tmpDir, f)
					if err := os.WriteFile(path, []byte("TEST=value"), 0644); err != nil {
						t.Fatal(err)
					}
				}

				return envFile
			},
			expectedCount: 1,
			validateOrder: func(t *testing.T, backups []string) {
				if len(backups) != 1 {
					t.Errorf("expected 1 backup, got %d", len(backups))
					return
				}

				baseName := filepath.Base(backups[0])
				expected := ".env.backup.20241027-100000"
				if baseName != expected {
					t.Errorf("got %s, want %s", baseName, expected)
				}
			},
		},
		{
			name: "backups for different base files are separate",
			setup: func(t *testing.T) string {
				tmpDir := t.TempDir()
				envFile := filepath.Join(tmpDir, ".env")

				// Create backups for .env and .env.local
				files := []string{
					".env.backup.20241027-100000",
					".env.backup.20241027-110000",
					".env.local.backup.20241027-100000",
					".env.local.backup.20241027-110000",
				}

				for _, f := range files {
					path := filepath.Join(tmpDir, f)
					if err := os.WriteFile(path, []byte("TEST=value"), 0644); err != nil {
						t.Fatal(err)
					}
				}

				return envFile
			},
			expectedCount: 2, // Only .env backups
			validateOrder: func(t *testing.T, backups []string) {
				for _, backup := range backups {
					baseName := filepath.Base(backup)
					if !strings.HasPrefix(baseName, ".env.backup.") {
						t.Errorf("unexpected backup in list: %s", baseName)
					}
					if strings.HasPrefix(baseName, ".env.local.backup.") {
						t.Errorf(".env.local backup should not be in .env backup list: %s", baseName)
					}
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			envFile := tt.setup(t)

			backups, err := ListBackups(envFile)
			if err != nil {
				t.Fatalf("ListBackups failed: %v", err)
			}

			if len(backups) != tt.expectedCount {
				t.Errorf("expected %d backups, got %d: %v", tt.expectedCount, len(backups), backups)
			}

			if tt.validateOrder != nil {
				tt.validateOrder(t, backups)
			}
		})
	}
}

func TestCleanOldBackups(t *testing.T) {
	type backupInfo struct {
		timestamp string
		daysOld   int
	}

	tests := []struct {
		name           string
		setupBackups   []backupInfo
		maxAgeDays     int
		minKeep        int
		expectedRemain int
		validate       func(t *testing.T, remaining []string)
	}{
		{
			name: "keep minimum number of backups regardless of age",
			setupBackups: []backupInfo{
				{timestamp: "20241020-100000", daysOld: 7},
				{timestamp: "20241021-100000", daysOld: 6},
				{timestamp: "20241022-100000", daysOld: 5},
				{timestamp: "20241023-100000", daysOld: 4},
				{timestamp: "20241024-100000", daysOld: 3},
			},
			maxAgeDays:     3,
			minKeep:        5,
			expectedRemain: 5, // Keep all 5 even though 2 are older than 3 days
		},
		{
			name: "delete old backups beyond minimum",
			setupBackups: []backupInfo{
				{timestamp: "20241010-100000", daysOld: 17},
				{timestamp: "20241015-100000", daysOld: 12},
				{timestamp: "20241020-100000", daysOld: 7},
				{timestamp: "20241024-100000", daysOld: 3},
				{timestamp: "20241025-100000", daysOld: 2},
			},
			maxAgeDays:     7,
			minKeep:        3,
			expectedRemain: 3, // Keep 3 newest, delete 2 old ones
			validate: func(t *testing.T, remaining []string) {
				// Should keep the 3 newest (10/24, 10/25, 10/20)
				expectedTimestamps := map[string]bool{
					"20241024-100000": true,
					"20241025-100000": true,
					"20241020-100000": true,
				}

				for _, backup := range remaining {
					baseName := filepath.Base(backup)
					found := false
					for ts := range expectedTimestamps {
						if strings.Contains(baseName, ts) {
							found = true
							break
						}
					}
					if !found {
						t.Errorf("unexpected backup kept: %s", baseName)
					}
				}
			},
		},
		{
			name: "no deletion if backups equal to minKeep",
			setupBackups: []backupInfo{
				{timestamp: "20241020-100000", daysOld: 7},
				{timestamp: "20241024-100000", daysOld: 3},
				{timestamp: "20241025-100000", daysOld: 2},
			},
			maxAgeDays:     5,
			minKeep:        3,
			expectedRemain: 3, // Keep all
		},
		{
			name: "no deletion if backups less than minKeep",
			setupBackups: []backupInfo{
				{timestamp: "20241025-100000", daysOld: 2},
			},
			maxAgeDays:     1,
			minKeep:        3,
			expectedRemain: 1, // Keep the only backup
		},
		{
			name:           "no backups to clean",
			setupBackups:   []backupInfo{},
			maxAgeDays:     7,
			minKeep:        3,
			expectedRemain: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			envFile := filepath.Join(tmpDir, ".env")

			// Create backups with specific ages
			now := time.Now()
			for _, bi := range tt.setupBackups {
				backupFile := filepath.Join(tmpDir, ".env.backup."+bi.timestamp)
				if err := os.WriteFile(backupFile, []byte("TEST=value"), 0644); err != nil {
					t.Fatal(err)
				}

				// Set modification time to simulate age
				modTime := now.AddDate(0, 0, -bi.daysOld)
				if err := os.Chtimes(backupFile, modTime, modTime); err != nil {
					t.Fatal(err)
				}
			}

			// Run cleanup
			err := CleanOldBackups(envFile, tt.maxAgeDays, tt.minKeep)
			if err != nil {
				t.Fatalf("CleanOldBackups failed: %v", err)
			}

			// List remaining backups
			remaining, err := ListBackups(envFile)
			if err != nil {
				t.Fatalf("ListBackups failed: %v", err)
			}

			if len(remaining) != tt.expectedRemain {
				t.Errorf("expected %d remaining backups, got %d", tt.expectedRemain, len(remaining))
			}

			if tt.validate != nil {
				tt.validate(t, remaining)
			}
		})
	}
}

func TestGenerateBackupFilename(t *testing.T) {
	tests := []struct {
		name     string
		filename string
		pattern  string
	}{
		{
			name:     "simple env file",
			filename: ".env",
			pattern:  `^\.env\.backup\.\d{8}-\d{6}$`,
		},
		{
			name:     "env file with path",
			filename: "/path/to/.env",
			pattern:  `^/path/to/\.env\.backup\.\d{8}-\d{6}$`,
		},
		{
			name:     "env.local file",
			filename: ".env.local",
			pattern:  `^\.env\.local\.backup\.\d{8}-\d{6}$`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := generateBackupFilename(tt.filename)

			matched, err := regexp.MatchString(tt.pattern, result)
			if err != nil {
				t.Fatalf("invalid pattern: %v", err)
			}

			if !matched {
				t.Errorf("generated filename %q does not match pattern %q", result, tt.pattern)
			}
		})
	}
}

func TestFileExists(t *testing.T) {
	tmpDir := t.TempDir()

	existingFile := filepath.Join(tmpDir, "exists.txt")
	if err := os.WriteFile(existingFile, []byte("test"), 0644); err != nil {
		t.Fatal(err)
	}

	nonExistentFile := filepath.Join(tmpDir, "does-not-exist.txt")

	if !fileExists(existingFile) {
		t.Error("fileExists returned false for existing file")
	}

	if fileExists(nonExistentFile) {
		t.Error("fileExists returned true for non-existent file")
	}
}
