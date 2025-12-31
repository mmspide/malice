package util

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
)

// GetEnv reads an environment variable and returns a default value if not found.
// This is the canonical version - replaces Getopt and GetOpt duplicates.
func GetEnv(key, defaultVal string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultVal
}

// CopyFile copies a file from src to dst, preserving permissions.
// This is the canonical version - replaces duplicate CopyFile implementations.
func CopyFile(src, dst string) error {
	// Validate source
	srcInfo, err := os.Stat(src)
	if err != nil {
		return fmt.Errorf("stat source: %w", err)
	}

	if !srcInfo.Mode().IsRegular() {
		return fmt.Errorf("source is not a regular file")
	}

	// Open source
	srcFile, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("open source: %w", err)
	}
	defer srcFile.Close()

	// Create destination
	dstFile, err := os.Create(dst)
	if err != nil {
		return fmt.Errorf("create destination: %w", err)
	}

	// Copy file contents
	_, err = io.Copy(dstFile, srcFile)
	dstFile.Close() // Close before changing permissions

	if err != nil {
		return fmt.Errorf("copy contents: %w", err)
	}

	// Preserve permissions
	if err := os.Chmod(dst, srcInfo.Mode()); err != nil {
		return fmt.Errorf("chmod destination: %w", err)
	}

	return nil
}

// SafeJoinPath safely joins path components, preventing path traversal attacks.
func SafeJoinPath(base, elem string) (string, error) {
	path := filepath.Join(base, elem)

	// Ensure the result is within base
	absPath, err := filepath.Abs(path)
	if err != nil {
		return "", fmt.Errorf("abs path: %w", err)
	}

	absBase, err := filepath.Abs(base)
	if err != nil {
		return "", fmt.Errorf("abs base: %w", err)
	}

	// Check if path is within base
	relPath, err := filepath.Rel(absBase, absPath)
	if err != nil {
		return "", fmt.Errorf("rel path: %w", err)
	}

	// If relPath starts with "..", it's outside base
	if relPath == ".." || filepath.HasPrefix(relPath, ".."+string(filepath.Separator)) {
		return "", fmt.Errorf("path traversal detected: %s", elem)
	}

	return absPath, nil
}
