package util

import (
	"os"
	"path/filepath"
	"testing"
)

func TestGetEnv(t *testing.T) {
	// Set an environment variable for testing
	testKey := "TEST_GETENV_VAR_12345"
	testValue := "test_value"
	os.Setenv(testKey, testValue)
	defer os.Unsetenv(testKey)

	tests := []struct {
		name       string
		key        string
		defaultVal string
		want       string
	}{
		{
			name:       "env var exists",
			key:        testKey,
			defaultVal: "default",
			want:       testValue,
		},
		{
			name:       "env var not exists, use default",
			key:        "NONEXISTENT_VAR_XYZ",
			defaultVal: "default_value",
			want:       "default_value",
		},
		{
			name:       "empty env var returns default",
			key:        "EMPTY_VAR_TEST",
			defaultVal: "default",
			want:       "default",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GetEnv(tt.key, tt.defaultVal)
			if got != tt.want {
				t.Errorf("GetEnv(%q, %q) = %q, want %q", tt.key, tt.defaultVal, got, tt.want)
			}
		})
	}
}

func TestCopyFile(t *testing.T) {
	// Create a temporary directory for testing
	tmpDir := t.TempDir()

	// Create a source file
	srcFile := filepath.Join(tmpDir, "source.txt")
	srcContent := []byte("test content")
	if err := os.WriteFile(srcFile, srcContent, 0644); err != nil {
		t.Fatalf("failed to create source file: %v", err)
	}

	// Set specific permissions
	if err := os.Chmod(srcFile, 0755); err != nil {
		t.Fatalf("failed to set permissions: %v", err)
	}

	dstFile := filepath.Join(tmpDir, "destination.txt")

	tests := []struct {
		name      string
		src       string
		dst       string
		wantError bool
	}{
		{
			name:      "copy valid file",
			src:       srcFile,
			dst:       dstFile,
			wantError: false,
		},
		{
			name:      "copy nonexistent file",
			src:       filepath.Join(tmpDir, "nonexistent.txt"),
			dst:       filepath.Join(tmpDir, "output.txt"),
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Use unique destination for each test
			testDst := filepath.Join(tmpDir, tt.name+".txt")

			err := CopyFile(tt.src, testDst)
			if (err != nil) != tt.wantError {
				t.Errorf("CopyFile() error = %v, wantError %v", err, tt.wantError)
			}

			if !tt.wantError {
				// Verify file contents
				dstContent, err := os.ReadFile(testDst)
				if err != nil {
					t.Fatalf("failed to read destination file: %v", err)
				}
				if string(dstContent) != string(srcContent) {
					t.Errorf("file content mismatch: got %q, want %q", string(dstContent), string(srcContent))
				}

				// Verify permissions
				srcInfo, _ := os.Stat(srcFile)
				dstInfo, _ := os.Stat(testDst)
				if srcInfo.Mode() != dstInfo.Mode() {
					t.Errorf("permissions mismatch: got %o, want %o", dstInfo.Mode(), srcInfo.Mode())
				}
			}
		})
	}
}

func TestSafeJoinPath(t *testing.T) {
	tmpDir := t.TempDir()

	tests := []struct {
		name      string
		base      string
		elem      string
		wantError bool
		errorMsg  string
	}{
		{
			name:      "valid subpath",
			base:      tmpDir,
			elem:      "subdir/file.txt",
			wantError: false,
		},
		{
			name:      "path traversal attempt",
			base:      tmpDir,
			elem:      "../../etc/passwd",
			wantError: true,
			errorMsg:  "path traversal",
		},
		{
			name:      "direct path traversal",
			base:      tmpDir,
			elem:      "..",
			wantError: true,
			errorMsg:  "path traversal",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := SafeJoinPath(tt.base, tt.elem)
			if (err != nil) != tt.wantError {
				t.Errorf("SafeJoinPath() error = %v, wantError %v", err, tt.wantError)
			}

			if err != nil && tt.errorMsg != "" {
				if !containsTestString(err.Error(), tt.errorMsg) {
					t.Errorf("error message %q does not contain %q", err.Error(), tt.errorMsg)
				}
			}

			if !tt.wantError && got == "" {
				t.Error("SafeJoinPath() returned empty path for valid input")
			}
		})
	}
}

func containsTestString(s, substr string) bool {
	for i := 0; i < len(s)-len(substr)+1; i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
