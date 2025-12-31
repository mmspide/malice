package commands

import (
	"os"
	"path/filepath"
	"testing"
)

func TestValidateAndNormalizePath(t *testing.T) {
	// Create a temporary test directory and file
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.bin")
	if err := os.WriteFile(testFile, []byte("test content"), 0644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	tests := []struct {
		name      string
		path      string
		wantError bool
		errorMsg  string
	}{
		{
			name:      "valid file",
			path:      testFile,
			wantError: false,
		},
		{
			name:      "nonexistent file",
			path:      filepath.Join(tmpDir, "nonexist.bin"),
			wantError: true,
			errorMsg:  "file not found",
		},
		{
			name:      "directory instead of file",
			path:      tmpDir,
			wantError: true,
			errorMsg:  "not a regular file",
		},
		{
			name:      "empty path",
			path:      "",
			wantError: true,
			errorMsg:  "cannot resolve path",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateAndNormalizePath(tt.path)
			if (err != nil) != tt.wantError {
				t.Errorf("got error %v, want error %v", err != nil, tt.wantError)
			}
			if err != nil && tt.errorMsg != "" {
				if !containsString(err.Error(), tt.errorMsg) {
					t.Errorf("error message %q does not contain %q", err.Error(), tt.errorMsg)
				}
			}
		})
	}
}

func containsString(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && s[0:len(substr)] == substr || findInString(s, substr))
}

func findInString(s, substr string) bool {
	for i := 0; i < len(s)-len(substr)+1; i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

func TestCmdScan(t *testing.T) {
	tests := []struct {
		name      string
		path      string
		wantError bool
	}{
		{
			name:      "empty path",
			path:      "",
			wantError: true,
		},
		{
			name:      "nonexistent file",
			path:      "/tmp/nonexistent_file_12345.bin",
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := cmdScan(tt.path, false)
			if (err != nil) != tt.wantError {
				t.Errorf("got error %v, want error %v", err != nil, tt.wantError)
			}
		})
	}
}
