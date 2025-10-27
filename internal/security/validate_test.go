package security

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestValidateFilePath(t *testing.T) {
	// Create a temporary directory for testing
	tempDir := t.TempDir()

	// Change to temp directory for consistent testing
	originalDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get current directory: %v", err)
	}
	defer os.Chdir(originalDir)

	if err := os.Chdir(tempDir); err != nil {
		t.Fatalf("failed to change to temp directory: %v", err)
	}

	// Create test files and directories
	os.WriteFile(".env", []byte("TEST=value"), 0644)
	os.Mkdir("config", 0755)
	os.WriteFile("config/.env", []byte("TEST=value"), 0644)
	os.Mkdir("subdir", 0755)
	os.WriteFile("subdir/file.txt", []byte("content"), 0644)

	// Create a file with spaces in the name
	os.Mkdir("my config", 0755)
	os.WriteFile("my config/.env file", []byte("TEST=value"), 0644)

	tests := []struct {
		name        string
		path        string
		setup       func() string
		cleanup     func()
		expectError bool
		errorMsg    string
	}{
		{
			name:        "valid relative path in current dir",
			path:        ".env",
			expectError: false,
		},
		{
			name:        "valid relative path in subdirectory",
			path:        "config/.env",
			expectError: false,
		},
		{
			name:        "valid path with subdirectory",
			path:        "subdir/file.txt",
			expectError: false,
		},
		{
			name:        "reject parent directory traversal",
			path:        "../.env",
			expectError: true,
			errorMsg:    "path traversal not allowed",
		},
		{
			name:        "reject multiple parent traversal",
			path:        "../../etc/passwd",
			expectError: true,
			errorMsg:    "path traversal not allowed",
		},
		{
			name:        "reject absolute path (Unix style)",
			path:        "/etc/passwd",
			expectError: true,
			errorMsg:    "path", // On Windows this might be "path outside", on Unix "absolute paths"
		},
		{
			name:        "reject hidden parent traversal",
			path:        "config/../../etc/passwd",
			expectError: true,
			errorMsg:    "path", // Can be either "path traversal" or "path outside"
		},
		{
			name:        "reject empty path",
			path:        "",
			expectError: true,
			errorMsg:    "path cannot be empty",
		},
		{
			name:        "handle paths with spaces",
			path:        "my config/.env file",
			expectError: false,
		},
		{
			name:        "reject path with multiple parent traversals",
			path:        "subdir/../../../etc/passwd",
			expectError: true,
			errorMsg:    "path", // Can be either "path traversal" or "path outside"
		},
		{
			name:        "accept path with single dot",
			path:        "./config/.env",
			expectError: false,
		},
		{
			name:        "accept nested subdirectories",
			path:        "config/subdir/nested/file.env",
			expectError: false,
		},
	}

	// Windows-specific tests
	if runtime.GOOS == "windows" {
		tests = append(tests, []struct {
			name        string
			path        string
			setup       func() string
			cleanup     func()
			expectError bool
			errorMsg    string
		}{
			{
				name:        "reject absolute path (Windows style with drive)",
				path:        "C:\\Windows\\System32\\config",
				expectError: true,
				errorMsg:    "absolute paths not allowed",
			},
			{
				name:        "reject absolute path (Windows style alternate)",
				path:        "C:/Windows/System32/config",
				expectError: true,
				errorMsg:    "absolute paths not allowed",
			},
		}...)
	}

	// Test for symbolic links
	symlinkTest := struct {
		name        string
		path        string
		setup       func() string
		cleanup     func()
		expectError bool
		errorMsg    string
	}{
		name: "handle symbolic links within directory",
		path: "symlink_test",
		setup: func() string {
			target := filepath.Join(tempDir, "config", ".env")
			link := filepath.Join(tempDir, "symlink_test")
			os.Symlink(target, link)
			return link
		},
		cleanup: func() {
			os.Remove(filepath.Join(tempDir, "symlink_test"))
		},
		expectError: false,
	}

	// Add symlink test if not on Windows (Windows requires admin for symlinks)
	if runtime.GOOS != "windows" {
		tests = append(tests, symlinkTest)
	}

	// Test for symlink pointing outside directory
	symlinkOutsideTest := struct {
		name        string
		path        string
		setup       func() string
		cleanup     func()
		expectError bool
		errorMsg    string
	}{
		name: "reject symbolic link pointing outside directory",
		path: "symlink_outside",
		setup: func() string {
			target := filepath.Join(tempDir, "..", "outside.txt")
			link := filepath.Join(tempDir, "symlink_outside")
			// Create the target file first
			os.WriteFile(target, []byte("outside"), 0644)
			os.Symlink(target, link)
			return link
		},
		cleanup: func() {
			os.Remove(filepath.Join(tempDir, "symlink_outside"))
			os.Remove(filepath.Join(tempDir, "..", "outside.txt"))
		},
		expectError: true,
		errorMsg:    "path outside working directory",
	}

	if runtime.GOOS != "windows" {
		tests = append(tests, symlinkOutsideTest)
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setup != nil {
				tt.setup()
			}
			if tt.cleanup != nil {
				defer tt.cleanup()
			}

			err := ValidateFilePath(tt.path)

			if tt.expectError {
				if err == nil {
					t.Errorf("expected error containing '%s', got nil", tt.errorMsg)
				} else if !strings.Contains(err.Error(), tt.errorMsg) {
					t.Errorf("expected error containing '%s', got '%s'", tt.errorMsg, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
			}
		})
	}
}

func TestValidateFileSize(t *testing.T) {
	tempDir := t.TempDir()

	tests := []struct {
		name        string
		fileSize    int64
		maxSizeMB   int
		expectError bool
	}{
		{
			name:        "small file accepted",
			fileSize:    1024, // 1KB
			maxSizeMB:   10,
			expectError: false,
		},
		{
			name:        "file at limit accepted",
			fileSize:    10 * 1024 * 1024, // 10MB
			maxSizeMB:   10,
			expectError: false,
		},
		{
			name:        "file just under limit accepted",
			fileSize:    10*1024*1024 - 1, // 10MB - 1 byte
			maxSizeMB:   10,
			expectError: false,
		},
		{
			name:        "file over limit rejected",
			fileSize:    11 * 1024 * 1024, // 11MB
			maxSizeMB:   10,
			expectError: true,
		},
		{
			name:        "empty file accepted",
			fileSize:    0,
			maxSizeMB:   10,
			expectError: false,
		},
		{
			name:        "very large file rejected",
			fileSize:    100 * 1024 * 1024, // 100MB
			maxSizeMB:   10,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a test file with specified size
			testFile := filepath.Join(tempDir, tt.name+".txt")
			f, err := os.Create(testFile)
			if err != nil {
				t.Fatalf("failed to create test file: %v", err)
			}

			// Efficiently create a file of specific size
			if tt.fileSize > 0 {
				err = f.Truncate(tt.fileSize)
				if err != nil {
					f.Close()
					t.Fatalf("failed to set file size: %v", err)
				}
			}
			f.Close()

			err = ValidateFileSize(testFile, tt.maxSizeMB)

			if tt.expectError {
				if err == nil {
					t.Error("expected error for large file, got nil")
				} else if !strings.Contains(err.Error(), "file too large") {
					t.Errorf("expected 'file too large' error, got: %v", err)
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
			}
		})
	}
}

func TestValidateFileSize_NonExistentFile(t *testing.T) {
	// Test that non-existent files return nil (for write operations)
	err := ValidateFileSize("nonexistent.txt", 10)
	if err != nil {
		t.Errorf("expected nil for non-existent file, got: %v", err)
	}
}

func TestValidateDirectoryWritable(t *testing.T) {
	tempDir := t.TempDir()

	// Change to temp directory
	originalDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get current directory: %v", err)
	}
	defer os.Chdir(originalDir)

	if err := os.Chdir(tempDir); err != nil {
		t.Fatalf("failed to change to temp directory: %v", err)
	}

	// Create a writable directory
	writableDir := "writable"
	if err := os.Mkdir(writableDir, 0755); err != nil {
		t.Fatalf("failed to create writable directory: %v", err)
	}

	// Create a file (not a directory)
	testFile := "notadir.txt"
	if err := os.WriteFile(testFile, []byte("test"), 0644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	tests := []struct {
		name        string
		dirPath     string
		setup       func()
		expectError bool
		errorMsg    string
	}{
		{
			name:        "valid writable directory",
			dirPath:     writableDir,
			expectError: false,
		},
		{
			name:        "current directory is writable",
			dirPath:     ".",
			expectError: false,
		},
		{
			name:        "non-existent directory",
			dirPath:     "nonexistent",
			expectError: true,
			errorMsg:    "directory does not exist",
		},
		{
			name:        "path is a file not directory",
			dirPath:     testFile,
			expectError: true,
			errorMsg:    "path is not a directory",
		},
		{
			name:        "reject parent directory traversal",
			dirPath:     "../",
			expectError: true,
			errorMsg:    "path traversal not allowed",
		},
	}

	// Skip read-only test on Windows as it behaves differently
	if runtime.GOOS != "windows" {
		// Create a read-only directory (Unix-only test)
		readonlyDir := "readonly"
		tests = append(tests, struct {
			name        string
			dirPath     string
			setup       func()
			expectError bool
			errorMsg    string
		}{
			name:    "read-only directory",
			dirPath: readonlyDir,
			setup: func() {
				os.Mkdir(readonlyDir, 0555) // read + execute, no write
			},
			expectError: true,
			errorMsg:    "directory is not writable",
		})
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setup != nil {
				tt.setup()
			}

			err := ValidateDirectoryWritable(tt.dirPath)

			if tt.expectError {
				if err == nil {
					t.Errorf("expected error containing '%s', got nil", tt.errorMsg)
				} else if !strings.Contains(err.Error(), tt.errorMsg) {
					t.Errorf("expected error containing '%s', got '%s'", tt.errorMsg, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
			}
		})
	}
}
