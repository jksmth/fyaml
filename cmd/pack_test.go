package cmd

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestPack_InvalidYAML(t *testing.T) {
	// Test error handling for invalid YAML
	// Other successful cases are covered by TestPack_Golden
	_, err := pack("../testdata/invalid-yaml", "yaml")
	if err == nil {
		t.Error("pack() expected error for invalid YAML")
	}
	if err != nil && !strings.Contains(err.Error(), "yaml") {
		t.Errorf("pack() error = %v, want error containing 'yaml'", err)
	}
}

func TestPack_ScalarFile(t *testing.T) {
	// Test that files containing only a scalar (not a map) return an error
	tests := []struct {
		name    string
		content string
	}{
		{"string", "hello"},
		{"number", "42"},
		{"boolean", "true"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			scalarFile := filepath.Join(tmpDir, "value.yml")
			if err := os.WriteFile(scalarFile, []byte(tt.content), 0600); err != nil {
				t.Fatalf("Failed to create scalar file: %v", err)
			}

			_, err := pack(tmpDir, "yaml")
			if err == nil {
				t.Error("pack() expected error for scalar file")
			}
			if err != nil && !strings.Contains(err.Error(), "expected a map") {
				t.Errorf("pack() error = %v, want error containing 'expected a map'", err)
			}
		})
	}
}

func TestPack_ArrayFile(t *testing.T) {
	// Test that files containing only an array (not a map) return an error
	tmpDir := t.TempDir()
	arrayFile := filepath.Join(tmpDir, "items.yml")
	arrayContent := "- item1\n- item2\n- item3"
	if err := os.WriteFile(arrayFile, []byte(arrayContent), 0600); err != nil {
		t.Fatalf("Failed to create array file: %v", err)
	}

	_, err := pack(tmpDir, "yaml")
	if err == nil {
		t.Error("pack() expected error for array file")
	}
	if err != nil && !strings.Contains(err.Error(), "expected a map") {
		t.Errorf("pack() error = %v, want error containing 'expected a map'", err)
	}
}

func TestPack_Golden(t *testing.T) {
	tests := []struct {
		name     string
		dir      string
		expected string
	}{
		{
			name:     "simple",
			dir:      "../testdata/simple/input",
			expected: "../testdata/simple/expected.yml",
		},
		{
			name:     "nested",
			dir:      "../testdata/nested/input",
			expected: "../testdata/nested/expected.yml",
		},
		{
			name:     "at-root",
			dir:      "../testdata/at-root/input",
			expected: "../testdata/at-root/expected.yml",
		},
		{
			name:     "at-files",
			dir:      "../testdata/at-files/input",
			expected: "../testdata/at-files/expected.yml",
		},
		{
			name:     "ordering",
			dir:      "../testdata/ordering/input",
			expected: "../testdata/ordering/expected.yml",
		},
		{
			name:     "anchors",
			dir:      "../testdata/anchors/input",
			expected: "../testdata/anchors/expected.yml",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := pack(tt.dir, "yaml")
			if err != nil {
				t.Fatalf("pack() error = %v", err)
			}

			expected, err := os.ReadFile(tt.expected)
			if err != nil {
				t.Fatalf("Failed to read expected file: %v", err)
			}

			if string(result) != string(expected) {
				t.Errorf("pack() output does not match expected\nGot:\n%s\nWant:\n%s", string(result), string(expected))
			}
		})
	}
}

func TestPack_Deterministic(t *testing.T) {
	dir := "../testdata/ordering/input"

	result1, err := pack(dir, "yaml")
	if err != nil {
		t.Fatalf("pack() error = %v", err)
	}

	result2, err := pack(dir, "yaml")
	if err != nil {
		t.Fatalf("pack() error = %v", err)
	}

	if string(result1) != string(result2) {
		t.Errorf("pack() output is not deterministic\nFirst run:\n%s\nSecond run:\n%s", string(result1), string(result2))
	}
}

func TestPack_NonexistentDir(t *testing.T) {
	_, err := pack("nonexistent/dir", "yaml")
	if err == nil {
		t.Error("pack() expected error for nonexistent directory")
	}
}

func TestPack_EmptyDir(t *testing.T) {
	// Create a temporary empty directory (no files, no subdirectories)
	tmpDir := t.TempDir()

	// Test YAML format - should return empty bytes (aligns with yq)
	t.Run("yaml", func(t *testing.T) {
		result, err := pack(tmpDir, "yaml")
		if err != nil {
			t.Fatalf("pack() error = %v, expected no error for empty directory", err)
		}
		// Should produce empty bytes for completely empty directory
		if len(result) != 0 {
			t.Errorf("pack() result = %q, want empty bytes", string(result))
		}
	})

	// Test JSON format - should return "null\n" (aligns with jq/yq)
	t.Run("json", func(t *testing.T) {
		result, err := pack(tmpDir, "json")
		if err != nil {
			t.Fatalf("pack() error = %v, expected no error for empty directory", err)
		}
		// Should produce "null\n" for completely empty directory
		if string(result) != "null\n" {
			t.Errorf("pack() result = %q, want 'null\\n'", string(result))
		}
	})
}

func TestPack_EmptySubdirs(t *testing.T) {
	// Test that directories with only empty subdirectories are valid
	// This aligns with the spec: directory names become keys
	tmpDir := t.TempDir()

	servicesDir := filepath.Join(tmpDir, "services")
	databaseDir := filepath.Join(tmpDir, "database")

	if err := os.Mkdir(servicesDir, 0700); err != nil {
		t.Fatalf("Failed to create services directory: %v", err)
	}
	if err := os.Mkdir(databaseDir, 0700); err != nil {
		t.Fatalf("Failed to create database directory: %v", err)
	}

	result, err := pack(tmpDir, "yaml")
	if err != nil {
		t.Fatalf("pack() error = %v, expected no error for directories with empty subdirs", err)
	}

	// Should produce YAML with empty map values for the directories
	if len(result) == 0 {
		t.Error("pack() returned empty result")
	}
	// Result should contain the directory names as keys
	resultStr := string(result)
	if !strings.Contains(resultStr, "services") || !strings.Contains(resultStr, "database") {
		t.Errorf("pack() result missing expected keys. Got: %s", resultStr)
	}
}

func TestPack_JSONInput(t *testing.T) {
	// Create a temporary directory with JSON files
	tmpDir := t.TempDir()

	servicesDir := filepath.Join(tmpDir, "services")
	if err := os.Mkdir(servicesDir, 0700); err != nil {
		t.Fatalf("Failed to create services directory: %v", err)
	}

	// Create a JSON file
	jsonFile := filepath.Join(servicesDir, "api.json")
	jsonContent := `{"name": "api", "port": 8080}`
	if err := os.WriteFile(jsonFile, []byte(jsonContent), 0600); err != nil {
		t.Fatalf("Failed to create JSON file: %v", err)
	}

	// Pack should process JSON files the same as YAML
	result, err := pack(tmpDir, "yaml")
	if err != nil {
		t.Fatalf("pack() error = %v", err)
	}

	// Result should contain the JSON content
	resultStr := string(result)
	if !strings.Contains(resultStr, "api") || !strings.Contains(resultStr, "8080") {
		t.Errorf("pack() result missing expected content from JSON file. Got: %s", resultStr)
	}
}

func TestPack_JSONOutput(t *testing.T) {
	// Test JSON output format
	dir := "../testdata/simple/input"

	result, err := pack(dir, "json")
	if err != nil {
		t.Fatalf("pack() error = %v", err)
	}

	// Result should be valid JSON
	resultStr := string(result)
	if len(resultStr) == 0 {
		t.Error("pack() returned empty JSON output")
	}

	// Should start with { or [ for valid JSON
	if resultStr[0] != '{' && resultStr[0] != '[' {
		t.Errorf("pack() JSON output doesn't start with { or [. Got: %s", resultStr[:min(50, len(resultStr))])
	}
}

func TestPack_JSONOutput_EmptyDir(t *testing.T) {
	// Test JSON output for empty directory - should return "null\n"
	tmpDir := t.TempDir()

	result, err := pack(tmpDir, "json")
	if err != nil {
		t.Fatalf("pack() error = %v, expected no error for empty directory", err)
	}

	// Should produce "null\n" for completely empty directory
	if string(result) != "null\n" {
		t.Errorf("pack() result = %q, want 'null\\n'", string(result))
	}
}

func TestPack_InvalidFormat(t *testing.T) {
	// Test invalid format parameter
	dir := "../testdata/simple/input"

	_, err := pack(dir, "invalid")
	if err == nil {
		t.Error("pack() expected error for invalid format")
	}
	if err != nil && !strings.Contains(err.Error(), "invalid format") {
		t.Errorf("pack() error = %v, want error containing 'invalid format'", err)
	}
}

func TestPack_YAMLAnchors(t *testing.T) {
	// Test that YAML anchors and aliases are expanded within a single file
	dir := "../testdata/anchors/input"
	expectedFile := "../testdata/anchors/expected.yml"

	result, err := pack(dir, "yaml")
	if err != nil {
		t.Fatalf("pack() error = %v", err)
	}

	expected, err := os.ReadFile(expectedFile)
	if err != nil {
		t.Fatalf("Failed to read expected file: %v", err)
	}

	if string(result) != string(expected) {
		t.Errorf("pack() output does not match expected\nGot:\n%s\nWant:\n%s", string(result), string(expected))
	}

	// Verify that anchors are expanded (not present as references)
	resultStr := string(result)
	if strings.Contains(resultStr, "&defaults") || strings.Contains(resultStr, "*defaults") {
		t.Error("pack() output contains anchor/alias references, should be expanded")
	}

	// Verify that the expanded values are present
	if !strings.Contains(resultStr, "timeout: 30") || !strings.Contains(resultStr, "retries: 3") {
		t.Error("pack() output missing expanded anchor values")
	}
}

func TestWriteOutput_Stdout(t *testing.T) {
	// Test stdout path
	result := []byte("test output\n")
	err := writeOutput("", result)
	if err != nil {
		t.Errorf("writeOutput() to stdout error = %v", err)
	}
}

func TestWriteOutput_File(t *testing.T) {
	tmpDir := t.TempDir()
	outputFile := filepath.Join(tmpDir, "output.yml")
	result := []byte("test: value\n")

	err := writeOutput(outputFile, result)
	if err != nil {
		t.Fatalf("writeOutput() error = %v", err)
	}

	// Verify file exists and has correct content
	content, err := os.ReadFile(outputFile)
	if err != nil {
		t.Fatalf("Failed to read output file: %v", err)
	}
	if string(content) != string(result) {
		t.Errorf("writeOutput() content = %q, want %q", string(content), string(result))
	}

	// Verify permissions (should be 0644)
	info, err := os.Stat(outputFile)
	if err != nil {
		t.Fatalf("Failed to stat output file: %v", err)
	}
	expectedPerms := os.FileMode(0644)
	if info.Mode().Perm() != expectedPerms {
		t.Errorf("writeOutput() permissions = %o, want %o", info.Mode().Perm(), expectedPerms)
	}
}

func TestWriteOutput_AtomicWrite(t *testing.T) {
	// Test that atomic write replaces existing file correctly
	tmpDir := t.TempDir()
	outputFile := filepath.Join(tmpDir, "output.yml")
	originalContent := []byte("original content\n")

	// Create existing file
	if err := os.WriteFile(outputFile, originalContent, 0644); err != nil {
		t.Fatalf("Failed to create existing file: %v", err)
	}

	// Write new content atomically
	newContent := []byte("new content\n")
	err := writeOutput(outputFile, newContent)
	if err != nil {
		t.Fatalf("writeOutput() error = %v", err)
	}

	// Verify new content replaced old content
	content, err := os.ReadFile(outputFile)
	if err != nil {
		t.Fatalf("Failed to read output file: %v", err)
	}
	if string(content) != string(newContent) {
		t.Errorf("writeOutput() atomic write failed - content = %q, want %q", string(content), string(newContent))
	}
}

func TestWriteOutput_CreateTempError(t *testing.T) {
	// Test with invalid directory (should fail to create temp file)
	invalidDir := "/nonexistent/path/that/does/not/exist"
	outputFile := filepath.Join(invalidDir, "output.yml")

	err := writeOutput(outputFile, []byte("test"))
	if err == nil {
		t.Error("writeOutput() expected error for invalid directory")
	}
	if !strings.Contains(err.Error(), "failed to create temp file") {
		t.Errorf("writeOutput() error = %v, want error about temp file", err)
	}
}

func TestHandleCheck_Success(t *testing.T) {
	tmpDir := t.TempDir()
	outputFile := filepath.Join(tmpDir, "output.yml")
	content := []byte("test: value\n")

	// Create existing file with matching content
	if err := os.WriteFile(outputFile, content, 0644); err != nil {
		t.Fatalf("Failed to create file: %v", err)
	}

	err := handleCheck(outputFile, content)
	if err != nil {
		t.Errorf("handleCheck() error = %v, want nil", err)
	}
}

// Note: Tests for handleCheck mismatch and nonexistent file scenarios are not included
// because handleCheck calls os.Exit(2) on mismatch, which terminates the test process.
// This behavior is tested via integration tests in the CI pipeline (verify-testdata step).

func TestHandleCheck_EmptyOutput(t *testing.T) {
	err := handleCheck("", []byte("test"))
	if err == nil {
		t.Error("handleCheck() expected error for empty output path")
	}
	if !strings.Contains(err.Error(), "--check requires --output") {
		t.Errorf("handleCheck() error = %v, want error about --output", err)
	}
}

func TestHandleCheck_ReadFileError(t *testing.T) {
	// Test error path when ReadFile fails (not IsNotExist)
	// Use a directory path instead of file to trigger read error
	tmpDir := t.TempDir()

	err := handleCheck(tmpDir, []byte("test"))
	if err == nil {
		t.Error("handleCheck() expected error when reading directory")
	}
	if !strings.Contains(err.Error(), "failed to read output file") {
		t.Errorf("handleCheck() error = %v, want error about reading file", err)
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
