package security

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

// ValidateFilePath ensures the file path is safe to access
// Prevents directory traversal and restricts access to current working directory
func ValidateFilePath(path string) error {
	// Check for empty path
	if path == "" {
		return fmt.Errorf("path cannot be empty")
	}

	// Clean the path to normalize it (removes .., ., etc.)
	cleanPath := filepath.Clean(path)

	// Check for parent directory traversal patterns
	if strings.Contains(cleanPath, "..") {
		return fmt.Errorf("path traversal not allowed: path contains '..'")
	}

	// Check if path is absolute
	if filepath.IsAbs(cleanPath) {
		return fmt.Errorf("absolute paths not allowed: %s", cleanPath)
	}

	// Get current working directory
	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get current working directory: %w", err)
	}

	// Convert relative path to absolute
	absPath, err := filepath.Abs(cleanPath)
	if err != nil {
		return fmt.Errorf("failed to resolve absolute path: %w", err)
	}

	// Resolve symlinks to get the real path
	realPath, err := filepath.EvalSymlinks(absPath)
	if err != nil {
		// If file doesn't exist yet (for write operations), check the directory
		if os.IsNotExist(err) {
			// Check if parent directory exists and is within CWD
			dir := filepath.Dir(absPath)
			realPath, err = filepath.EvalSymlinks(dir)
			if err != nil {
				// Parent directory doesn't exist, validate the absolute path directly
				realPath = absPath
			} else {
				// Reconstruct path with resolved directory and filename
				realPath = filepath.Join(realPath, filepath.Base(absPath))
			}
		} else {
			return fmt.Errorf("failed to resolve symbolic links: %w", err)
		}
	}

	// Normalize both paths for comparison
	cwdNorm := filepath.Clean(cwd)
	realPathNorm := filepath.Clean(realPath)

	// On Windows, ensure consistent path separators and case
	if runtime.GOOS == "windows" {
		cwdNorm = strings.ToLower(filepath.ToSlash(cwdNorm))
		realPathNorm = strings.ToLower(filepath.ToSlash(realPathNorm))
	}

	// Check if the real path is within or equal to the current working directory
	// Use filepath.Rel to check if realPath is relative to cwd
	relPath, err := filepath.Rel(cwdNorm, realPathNorm)
	if err != nil {
		return fmt.Errorf("failed to determine relative path: %w", err)
	}

	// If relative path starts with "..", it's outside the CWD
	if strings.HasPrefix(relPath, "..") {
		return fmt.Errorf("path outside working directory: %s resolves to %s", path, realPath)
	}

	return nil
}

// ValidateFileSize ensures the file isn't too large (prevent OOM attacks)
func ValidateFileSize(path string, maxSizeMB int) error {
	info, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			// File doesn't exist yet (may be created later)
			return nil
		}
		return fmt.Errorf("failed to stat file: %w", err)
	}

	maxBytes := int64(maxSizeMB * 1024 * 1024)
	if info.Size() > maxBytes {
		return fmt.Errorf("file too large: %s is %d bytes (max %dMB)", path, info.Size(), maxSizeMB)
	}

	return nil
}

// ValidateDirectoryWritable checks if a directory exists and is writable
func ValidateDirectoryWritable(dirPath string) error {
	// First validate the directory path itself
	if err := ValidateFilePath(dirPath); err != nil {
		return err
	}

	// Check if directory exists
	info, err := os.Stat(dirPath)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("directory does not exist: %s", dirPath)
		}
		return fmt.Errorf("failed to stat directory: %w", err)
	}

	// Check if it's actually a directory
	if !info.IsDir() {
		return fmt.Errorf("path is not a directory: %s", dirPath)
	}

	// Try to create a temporary file to test write permissions
	testFile := filepath.Join(dirPath, ".envquack_write_test")
	f, err := os.Create(testFile)
	if err != nil {
		return fmt.Errorf("directory is not writable: %s", dirPath)
	}
	f.Close()
	os.Remove(testFile)

	return nil
}
