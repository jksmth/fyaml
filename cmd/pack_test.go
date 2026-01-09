package cmd

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/jksmth/fyaml/internal/logger"
)

// Helper to create PackOptions for tests.
// Mode defaults to "canonical" if not specified (for mode-agnostic tests).
// Mode-specific tests (golden tests) should explicitly specify the mode.
func testOpts(dir, format string, enableIncludes, convertBooleans bool, mode ...string) PackOptions {
	m := "canonical" // default for mode-agnostic tests
	if len(mode) > 0 && mode[0] != "" {
		m = mode[0]
	}
	return PackOptions{
		Dir:             dir,
		Format:          format,
		EnableIncludes:  enableIncludes,
		ConvertBooleans: convertBooleans,
		Indent:          2, // Default indent for tests
		Mode:            m,
		MergeStrategy:   "shallow", // Default merge strategy
	}
}

// Helper to create PackOptions with merge strategy
func testOptsWithMerge(dir, format string, enableIncludes, convertBooleans bool, mergeStrategy string, mode ...string) PackOptions {
	opts := testOpts(dir, format, enableIncludes, convertBooleans, mode...)
	opts.MergeStrategy = mergeStrategy
	return opts
}

// Helper to create PackOptions with custom indent
func testOptsWithIndent(dir, format string, enableIncludes, convertBooleans bool, indent int) PackOptions {
	return PackOptions{
		Dir:             dir,
		Format:          format,
		EnableIncludes:  enableIncludes,
		ConvertBooleans: convertBooleans,
		Indent:          indent,
	}
}

// assertErrorContains asserts that an error exists and contains the specified substring.
// If the error is nil or doesn't contain the substring, the test fails.
func assertErrorContains(t *testing.T, err error, substr string) {
	t.Helper()
	if err == nil {
		t.Fatalf("expected error containing %q, got nil", substr)
	}
	if !strings.Contains(err.Error(), substr) {
		t.Errorf("error = %v, want error containing %q", err, substr)
	}
}

// assertNoError asserts that an error is nil. If the error is not nil, the test fails.
func assertNoError(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

// createTestDir creates a temporary directory with the specified files and empty directories.
// The files map keys are relative paths (e.g., "dir/file.yml"), and values are file contents.
// The emptyDirs slice contains relative paths to empty directories to create.
// Returns the path to the created temporary directory.
func createTestDir(t *testing.T, files map[string]string, emptyDirs []string) string {
	t.Helper()
	tmpDir := t.TempDir()
	for path, content := range files {
		fullPath := filepath.Join(tmpDir, path)
		dir := filepath.Dir(fullPath)
		if err := os.MkdirAll(dir, 0700); err != nil {
			t.Fatalf("Failed to create directory %q: %v", dir, err)
		}
		if err := os.WriteFile(fullPath, []byte(content), 0600); err != nil {
			t.Fatalf("Failed to create file %q: %v", fullPath, err)
		}
	}
	for _, dir := range emptyDirs {
		fullPath := filepath.Join(tmpDir, dir)
		if err := os.MkdirAll(fullPath, 0700); err != nil {
			t.Fatalf("Failed to create empty directory %q: %v", fullPath, err)
		}
	}
	return tmpDir
}

// assertOutputEqual asserts that two byte slices are equal.
// Provides a clear error message showing both got and want values.
func assertOutputEqual(t *testing.T, got, want []byte) {
	t.Helper()
	if string(got) != string(want) {
		t.Errorf("output does not match expected\nGot:\n%s\nWant:\n%s", string(got), string(want))
	}
}

func TestPack_InvalidYAML(t *testing.T) {
	// Test error handling for invalid YAML
	// Other successful cases are covered by TestPack_Golden_Canonical
	_, err := pack(testOpts("../testdata/invalid-yaml", "yaml", false, false), nil)
	assertErrorContains(t, err, "yaml")

	// Verify error includes file path for better debugging
	// Error may be "failed to parse YAML/JSON in" or "YAML/JSON syntax error in" depending on error type
	errStr := err.Error()
	hasPathContext := strings.Contains(errStr, "failed to parse YAML/JSON in") ||
		strings.Contains(errStr, "YAML/JSON syntax error in") ||
		strings.Contains(errStr, "YAML/JSON type errors in")
	if !hasPathContext {
		t.Errorf("error should include file path context: %v", err)
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
			tmpDir := createTestDir(t, map[string]string{
				"value.yml": tt.content,
			}, nil)

			_, err := pack(testOpts(tmpDir, "yaml", false, false), nil)
			assertErrorContains(t, err, "expected a map")
		})
	}
}

func TestPack_ArrayFile(t *testing.T) {
	// Test that files containing only an array (not a map) return an error using fixture
	_, err := pack(testOpts("../testdata/array-file/input", "yaml", false, false), nil)
	assertErrorContains(t, err, "expected a map")
}

func TestPack_Golden_Canonical(t *testing.T) {
	tests := []struct {
		name     string
		dir      string
		expected string
	}{
		{
			name:     "simple",
			dir:      "../testdata/simple/input",
			expected: "../testdata/simple/expected-canonical.yml",
		},
		{
			name:     "nested",
			dir:      "../testdata/nested/input",
			expected: "../testdata/nested/expected-canonical.yml",
		},
		{
			name:     "at-root",
			dir:      "../testdata/at-root/input",
			expected: "../testdata/at-root/expected-canonical.yml",
		},
		{
			name:     "at-files",
			dir:      "../testdata/at-files/input",
			expected: "../testdata/at-files/expected-canonical.yml",
		},
		{
			name:     "ordering",
			dir:      "../testdata/ordering/input",
			expected: "../testdata/ordering/expected-canonical.yml",
		},
		{
			name:     "anchors",
			dir:      "../testdata/anchors/input",
			expected: "../testdata/anchors/expected-canonical.yml",
		},
		{
			name:     "includes",
			dir:      "../testdata/includes/input",
			expected: "../testdata/includes/expected-canonical.yml",
		},
		{
			name:     "at-directories",
			dir:      "../testdata/at-directories/input",
			expected: "../testdata/at-directories/expected-canonical.yml",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Includes test requires --enable-includes flag
			enableIncludes := tt.name == "includes"
			result, err := pack(testOpts(tt.dir, "yaml", enableIncludes, false, "canonical"), nil)
			assertNoError(t, err)

			expected, err := os.ReadFile(tt.expected)
			if err != nil {
				t.Fatalf("Failed to read expected file: %v", err)
			}

			assertOutputEqual(t, result, expected)
		})
	}
}

func TestPack_Golden_Preserve(t *testing.T) {
	tests := []struct {
		name     string
		dir      string
		expected string
	}{
		{
			name:     "simple",
			dir:      "../testdata/simple/input",
			expected: "../testdata/simple/expected-preserve.yml",
		},
		{
			name:     "nested",
			dir:      "../testdata/nested/input",
			expected: "../testdata/nested/expected-preserve.yml",
		},
		{
			name:     "at-root",
			dir:      "../testdata/at-root/input",
			expected: "../testdata/at-root/expected-preserve.yml",
		},
		{
			name:     "at-files",
			dir:      "../testdata/at-files/input",
			expected: "../testdata/at-files/expected-preserve.yml",
		},
		{
			name:     "ordering",
			dir:      "../testdata/ordering/input",
			expected: "../testdata/ordering/expected-preserve.yml",
		},
		{
			name:     "anchors",
			dir:      "../testdata/anchors/input",
			expected: "../testdata/anchors/expected-preserve.yml",
		},
		{
			name:     "includes",
			dir:      "../testdata/includes/input",
			expected: "../testdata/includes/expected-preserve.yml",
		},
		{
			name:     "at-directories",
			dir:      "../testdata/at-directories/input",
			expected: "../testdata/at-directories/expected-preserve.yml",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Includes test requires --enable-includes flag
			enableIncludes := tt.name == "includes"
			result, err := pack(testOpts(tt.dir, "yaml", enableIncludes, false, "preserve"), nil)
			assertNoError(t, err)

			expected, err := os.ReadFile(tt.expected)
			if err != nil {
				t.Fatalf("Failed to read expected file: %v", err)
			}

			assertOutputEqual(t, result, expected)
		})
	}
}

func TestPack_NonexistentDir(t *testing.T) {
	_, err := pack(testOpts("nonexistent/dir", "yaml", false, false), nil)
	// Error message may vary, but should contain something about the directory
	// Common error messages: "no such file", "not a directory", etc.
	if err == nil {
		t.Fatal("pack() expected error for nonexistent directory")
	}
	// Just verify an error was returned (error message format may vary by OS)
	if err.Error() == "" {
		t.Error("pack() error should have a message")
	}
}

func TestPack_EmptyDir(t *testing.T) {
	// Create a temporary empty directory (no files, no subdirectories)
	tmpDir := t.TempDir()

	// Test YAML format - should return empty bytes (aligns with yq)
	t.Run("yaml", func(t *testing.T) {
		result, err := pack(testOpts(tmpDir, "yaml", false, false), nil)
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
		result, err := pack(testOpts(tmpDir, "json", false, false), nil)
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
	// Test that directories with only empty subdirectories (no YAML files) are ignored
	// Empty directories don't appear in output - only directories with YAML content
	tmpDir := createTestDir(t, nil, []string{"services", "database"})

	result, err := pack(testOpts(tmpDir, "yaml", false, false, "canonical"), nil)
	if err != nil {
		t.Fatalf("pack() error = %v, expected no error for directories with empty subdirs", err)
	}

	// Empty directories should not appear in output
	resultStr := string(result)
	if strings.Contains(resultStr, "services") || strings.Contains(resultStr, "database") {
		t.Errorf("pack() result should not contain empty directory keys. Got: %s", resultStr)
	}
	// Result should be empty since there are no YAML files
	if len(result) != 0 {
		t.Errorf("pack() should return empty result for directories with no YAML files. Got: %s", resultStr)
	}
}

func TestPack_JSONInput_Golden(t *testing.T) {
	// Test that JSON files are processed correctly using fixtures
	result, err := pack(testOpts("../testdata/json-input/input", "yaml", false, false), nil)
	assertNoError(t, err)

	expected, err := os.ReadFile("../testdata/json-input/expected-canonical.yml")
	if err != nil {
		t.Fatalf("Failed to read expected file: %v", err)
	}

	assertOutputEqual(t, result, expected)
}

func TestPack_JSONOutput(t *testing.T) {
	// Test JSON output format with content
	// Note: Empty directory JSON output is tested in TestPack_EmptyDir
	result, err := pack(testOpts("../testdata/simple/input", "json", false, false), nil)
	assertNoError(t, err)

	resultStr := string(result)
	if len(resultStr) == 0 {
		t.Error("pack() returned empty JSON output")
	}
	// Should start with { or [ for valid JSON
	if resultStr[0] != '{' && resultStr[0] != '[' {
		t.Errorf("pack() JSON output doesn't start with { or [. Got: %s", resultStr[:min(50, len(resultStr))])
	}
}

func TestPack_InvalidFormat(t *testing.T) {
	// Test invalid format parameter
	_, err := pack(testOpts("../testdata/simple/input", "invalid", false, false), nil)
	assertErrorContains(t, err, "invalid format")
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
	assertOutputEqual(t, content, result)

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
	assertOutputEqual(t, content, newContent)
}

func TestWriteOutput_CreateTempError(t *testing.T) {
	// Test with invalid directory (should fail to create temp file)
	invalidDir := "/nonexistent/path/that/does/not/exist"
	outputFile := filepath.Join(invalidDir, "output.yml")

	err := writeOutput(outputFile, []byte("test"))
	assertErrorContains(t, err, "failed to create temp file")
}

func TestHandleCheck_Success(t *testing.T) {
	tmpDir := t.TempDir()
	outputFile := filepath.Join(tmpDir, "output.yml")
	content := []byte("test: value\n")

	// Create existing file with matching content
	if err := os.WriteFile(outputFile, content, 0644); err != nil {
		t.Fatalf("Failed to create file: %v", err)
	}

	err := handleCheck(outputFile, content, "yaml")
	if err != nil {
		t.Errorf("handleCheck() error = %v, want nil", err)
	}
}

// Note: Tests for handleCheck mismatch and nonexistent file scenarios are not included
// because handleCheck calls os.Exit(2) on mismatch, which terminates the test process.
// This behavior is tested via integration tests in the CI pipeline (verify-testdata step).

func TestHandleCheck_StdinTerminal(t *testing.T) {
	// Test error when stdin is a terminal without input
	// We can't easily simulate a terminal in tests, but we can test the logic
	// by checking that the function properly handles the terminal case
	// In practice, this will be a terminal, so we'll test the error message path
	// by using the actual stdin (which may be a terminal in test environment)
	// However, we can't reliably test this without mocking, so we'll skip
	// the terminal detection test and rely on integration testing
	// Instead, we test that pipe-based stdin works correctly
}

func TestHandleCheck_EmptyOutput_WithPipe(t *testing.T) {
	// Test that empty output with pipe stdin works (reads from stdin)
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("Failed to create pipe: %v", err)
	}
	defer func() { _ = r.Close() }()

	oldStdin := os.Stdin
	os.Stdin = r
	defer func() { os.Stdin = oldStdin }()

	// Write empty content to pipe and close
	_ = w.Close()

	// Test with empty output - should read from stdin (empty) and compare
	// YAML format: empty stdin should match empty result
	err = handleCheck("", []byte(""), "yaml")
	if err != nil {
		t.Errorf("handleCheck() with empty stdin and empty result should succeed, got error: %v", err)
	}
}

func TestHandleCheck_StdinSuccess(t *testing.T) {
	// Test successful comparison with stdin input
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("Failed to create pipe: %v", err)
	}
	defer func() { _ = r.Close() }()

	oldStdin := os.Stdin
	os.Stdin = r
	defer func() { os.Stdin = oldStdin }()

	content := []byte("test: value\n")
	expected := []byte("test: value\n")

	// Write content to pipe in goroutine
	go func() {
		defer func() { _ = w.Close() }()
		_, _ = w.Write(content)
	}()

	// Test with empty output (implicit stdin)
	err = handleCheck("", expected, "yaml")
	if err != nil {
		t.Errorf("handleCheck() with matching stdin should succeed, got error: %v", err)
	}
}

func TestHandleCheck_StdinSuccess_ExplicitDash(t *testing.T) {
	// Test successful comparison with explicit --output -
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("Failed to create pipe: %v", err)
	}
	defer func() { _ = r.Close() }()

	oldStdin := os.Stdin
	os.Stdin = r
	defer func() { os.Stdin = oldStdin }()

	content := []byte("test: value\n")
	expected := []byte("test: value\n")

	// Write content to pipe in goroutine
	go func() {
		defer func() { _ = w.Close() }()
		_, _ = w.Write(content)
	}()

	// Test with explicit dash
	err = handleCheck("-", expected, "yaml")
	if err != nil {
		t.Errorf("handleCheck() with matching stdin and explicit dash should succeed, got error: %v", err)
	}
}

func TestHandleCheck_StdinEmpty(t *testing.T) {
	// Test comparison with empty stdin (should compare against empty for YAML, null\n for JSON)
	t.Run("yaml", func(t *testing.T) {
		r, w, err := os.Pipe()
		if err != nil {
			t.Fatalf("Failed to create pipe: %v", err)
		}
		defer func() { _ = r.Close() }()

		oldStdin := os.Stdin
		os.Stdin = r
		defer func() { os.Stdin = oldStdin }()

		// Close write end immediately (empty stdin)
		_ = w.Close()

		// Compare against empty result - should match (YAML format)
		err = handleCheck("", []byte{}, "yaml")
		if err != nil {
			t.Errorf("handleCheck() with empty stdin and empty result should succeed, got error: %v", err)
		}
	})

	t.Run("json", func(t *testing.T) {
		r, w, err := os.Pipe()
		if err != nil {
			t.Fatalf("Failed to create pipe: %v", err)
		}
		defer func() { _ = r.Close() }()

		oldStdin := os.Stdin
		os.Stdin = r
		defer func() { os.Stdin = oldStdin }()

		// Close write end immediately (empty stdin)
		_ = w.Close()

		// Compare against null\n result - should match (JSON format normalizes empty to null\n)
		err = handleCheck("", []byte("null\n"), "json")
		if err != nil {
			t.Errorf("handleCheck() with empty stdin and null\\n result should succeed for JSON, got error: %v", err)
		}
	})

	// Compare against non-empty result - should mismatch (but we can't test os.Exit)
	// The mismatch/exit behavior is tested via integration tests in CI
}

func TestHandleCheck_ReadFileError(t *testing.T) {
	// Test error path when ReadFile fails (not IsNotExist)
	// Use a directory path instead of file to trigger read error
	tmpDir := t.TempDir()

	err := handleCheck(tmpDir, []byte("test"), "yaml")
	assertErrorContains(t, err, "failed to read output file")
}

func TestPack_EnableIncludes(t *testing.T) {
	// Test the --enable-includes extension feature
	tmpDir := createTestDir(t, map[string]string{
		"commands/scripts/hello.sh": "#!/bin/bash\necho 'Hello World'",
		"commands/hello.yml": `description: A test command
steps:
  - run:
      name: Hello
      command: <<include(scripts/hello.sh)>>`,
	}, nil)

	// Test WITHOUT includes enabled - directive should remain as-is
	result, err := pack(testOpts(tmpDir, "yaml", false, false), nil)
	assertNoError(t, err)
	resultStr := string(result)
	if !strings.Contains(resultStr, "<<include(scripts/hello.sh)>>") {
		t.Error("pack() without --enable-includes should preserve include directive")
	}

	// Test WITH includes enabled - directive should be replaced
	result, err = pack(testOpts(tmpDir, "yaml", true, false), nil)
	assertNoError(t, err)
	resultStr = string(result)
	if strings.Contains(resultStr, "<<include") {
		t.Error("pack() with --enable-includes should replace include directive")
	}
	if !strings.Contains(resultStr, "echo 'Hello World'") {
		t.Errorf("pack() with --enable-includes should contain included content. Got:\n%s", resultStr)
	}
}

func TestPack_EnableIncludes_ErrorFileNotFound(t *testing.T) {
	// Test error handling when included file doesn't exist
	tmpDir := createTestDir(t, map[string]string{
		"commands/hello.yml": `command: <<include(nonexistent.sh)>>`,
	}, nil)

	_, err := pack(testOpts(tmpDir, "yaml", true, false), nil)
	assertErrorContains(t, err, "could not open")
}

func TestPack_EnableIncludes_RelativePathWithParent(t *testing.T) {
	// Test that relative paths with ../ work in include directives
	tmpDir := createTestDir(t, map[string]string{
		"scripts/script.sh": "echo 'relative path with ..'",
		"commands/test.yml": `command: <<include(../scripts/script.sh)>>`,
	}, nil)

	result, err := pack(testOpts(tmpDir, "yaml", true, false, "canonical"), nil)
	assertNoError(t, err)

	resultStr := string(result)
	if !strings.Contains(resultStr, "echo 'relative path with ..'") {
		t.Errorf("pack() should contain included content from relative path. Got:\n%s", resultStr)
	}
}

func TestPack_EnableIncludes_InvalidYAMLWithIncludes(t *testing.T) {
	// Test that invalid YAML still fails even with includes enabled
	tmpDir := createTestDir(t, map[string]string{
		"invalid.yml": "key: [unclosed", // Invalid YAML that would fail Unmarshal
	}, nil)

	_, err := pack(testOpts(tmpDir, "yaml", true, false), nil)
	if err == nil {
		t.Error("pack() expected error for invalid YAML even with includes enabled")
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func TestPack_EmptyDir_Warning(t *testing.T) {
	// Test that empty directory warning uses the logger
	tmpDir := t.TempDir()

	var buf bytes.Buffer
	log := logger.New(&buf, false) // Even without verbose, warnings should show

	_, err := pack(testOpts(tmpDir, "yaml", false, false), log)
	if err != nil {
		t.Fatalf("pack() error = %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "[WARN]") {
		t.Errorf("pack() empty dir should produce [WARN], got: %s", output)
	}
	if !strings.Contains(output, "no YAML/JSON files found") {
		t.Errorf("pack() warning should mention no files, got: %s", output)
	}
}

func TestPack_ConvertBooleans(t *testing.T) {
	// Test the --convert-booleans flag using fixtures
	tests := []struct {
		name            string
		convertBooleans bool
		expectedFile    string
	}{
		{
			name:            "without conversion",
			convertBooleans: false,
			expectedFile:    "../testdata/convert-booleans/expected-canonical-without-conversion.yml",
		},
		{
			name:            "with conversion",
			convertBooleans: true,
			expectedFile:    "../testdata/convert-booleans/expected-canonical-with-conversion.yml",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := pack(testOpts("../testdata/convert-booleans/input", "yaml", false, tt.convertBooleans), nil)
			assertNoError(t, err)

			expected, err := os.ReadFile(tt.expectedFile)
			if err != nil {
				t.Fatalf("Failed to read expected file: %v", err)
			}

			assertOutputEqual(t, result, expected)
		})
	}
}

func TestPack_ConvertBooleans_WithIncludes(t *testing.T) {
	// Test that normalization works together with includes using fixtures
	result, err := pack(testOpts("../testdata/convert-booleans-with-includes/input", "yaml", true, true), nil)
	assertNoError(t, err)

	expected, err := os.ReadFile("../testdata/convert-booleans-with-includes/expected-canonical.yml")
	if err != nil {
		t.Fatalf("Failed to read expected file: %v", err)
	}

	assertOutputEqual(t, result, expected)
}

func TestPack_ConvertBooleans_AllVariants(t *testing.T) {
	// Test all boolean variants (y, Y, yes, Yes, YES, on, On, ON, n, N, no, No, NO, off, Off, OFF)
	// Use unique keys for each variant to avoid duplicate key errors after conversion
	yamlContent := `true_values:
  val_y: y
  val_Y: Y
  val_yes: yes
  val_Yes: Yes
  val_YES: YES
  val_on: on
  val_On: On
  val_ON: ON
false_values:
  val_n: n
  val_N: N
  val_no: no
  val_No: No
  val_NO: NO
  val_off: off
  val_Off: Off
  val_OFF: OFF`
	tmpDir := createTestDir(t, map[string]string{
		"config.yml": yamlContent,
	}, nil)

	result, err := pack(testOpts(tmpDir, "yaml", false, true), nil)
	if err != nil {
		t.Fatalf("pack() error = %v", err)
	}
	resultStr := string(result)

	// All true variants should become true
	trueVariants := []string{"val_y: true", "val_Y: true", "val_yes: true", "val_Yes: true", "val_YES: true", "val_on: true", "val_On: true", "val_ON: true"}
	for _, variant := range trueVariants {
		if !strings.Contains(resultStr, variant) {
			t.Errorf("pack() should convert %q to true. Got:\n%s", variant, resultStr)
		}
	}

	// All false variants should become false
	falseVariants := []string{"val_n: false", "val_N: false", "val_no: false", "val_No: false", "val_NO: false", "val_off: false", "val_Off: false", "val_OFF: false"}
	for _, variant := range falseVariants {
		if !strings.Contains(resultStr, variant) {
			t.Errorf("pack() should convert %q to false. Got:\n%s", variant, resultStr)
		}
	}
}

func TestPack_ConvertBooleans_AlreadyBoolean(t *testing.T) {
	// Test that already-boolean values (true/false) are not double-converted
	tmpDir := createTestDir(t, map[string]string{
		"config.yml": `enabled: true
disabled: false`,
	}, nil)

	result, err := pack(testOpts(tmpDir, "yaml", false, true), nil)
	assertNoError(t, err)
	resultStr := string(result)

	// Already boolean values should remain as booleans
	if !strings.Contains(resultStr, "enabled: true") {
		t.Errorf("pack() should preserve already-boolean true. Got:\n%s", resultStr)
	}
	if !strings.Contains(resultStr, "disabled: false") {
		t.Errorf("pack() should preserve already-boolean false. Got:\n%s", resultStr)
	}
}

func TestPack_ConvertBooleans_DeeplyNested(t *testing.T) {
	// Test convert-booleans in deeply nested structures using fixtures
	result, err := pack(testOpts("../testdata/convert-booleans-deeply-nested/input", "yaml", false, true), nil)
	assertNoError(t, err)

	expected, err := os.ReadFile("../testdata/convert-booleans-deeply-nested/expected-canonical.yml")
	if err != nil {
		t.Fatalf("Failed to read expected file: %v", err)
	}

	assertOutputEqual(t, result, expected)
}

func TestPack_WithLogger(t *testing.T) {
	// Test that pack() works correctly with different logger configurations
	tmpDir := createTestDir(t, map[string]string{
		"test.yml": "key: value",
	}, nil)

	tests := []struct {
		name           string
		verbose        bool
		validateOutput func(t *testing.T, result []byte, logOutput string)
	}{
		{
			name:    "nil logger",
			verbose: false, // Not used for nil case
			validateOutput: func(t *testing.T, result []byte, logOutput string) {
				if len(result) == 0 {
					t.Error("pack() should return result even with nil logger")
				}
			},
		},
		{
			name:    "verbose logger",
			verbose: true,
			validateOutput: func(t *testing.T, result []byte, logOutput string) {
				if len(result) == 0 {
					t.Error("pack() should return result")
				}
				if !strings.Contains(logOutput, "[DEBUG] Processing:") {
					t.Errorf("pack() should log processing messages, got: %s", logOutput)
				}
				if !strings.Contains(logOutput, "test.yml") {
					t.Errorf("pack() should log file paths, got: %s", logOutput)
				}
			},
		},
		{
			name:    "quiet logger",
			verbose: false,
			validateOutput: func(t *testing.T, result []byte, logOutput string) {
				if len(result) == 0 {
					t.Error("pack() should return result")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			var testLogger logger.Logger
			if tt.name == "nil logger" {
				testLogger = nil
			} else {
				testLogger = logger.New(&buf, tt.verbose)
			}

			result, err := pack(testOpts(tmpDir, "yaml", false, false), testLogger)
			if err != nil {
				t.Fatalf("pack() error = %v", err)
			}

			logOutput := buf.String()
			tt.validateOutput(t, result, logOutput)
		})
	}
}

func TestPack_Indent_YAML(t *testing.T) {
	// Test indent values for YAML (default and custom)
	tests := []struct {
		name   string
		indent int
	}{
		{"default (2 spaces)", 2},
		{"1 space", 1},
		{"4 spaces", 4},
		{"8 spaces", 8},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := pack(testOptsWithIndent("../testdata/simple/input", "yaml", false, false, tt.indent), nil)
			if err != nil {
				t.Fatalf("pack() error = %v", err)
			}

			resultStr := string(result)
			indentStr := strings.Repeat(" ", tt.indent)
			expectedPrefix := indentStr + "item1:"

			if !strings.Contains(resultStr, expectedPrefix) {
				t.Errorf("pack() YAML output should use %d-space indent. Expected line starting with %q. Got:\n%s", tt.indent, expectedPrefix, resultStr)
			}
		})
	}
}

func TestPack_Indent_JSON(t *testing.T) {
	// Test indent values for JSON (default and custom)
	tests := []struct {
		name   string
		indent int
	}{
		{"default (2 spaces)", 2},
		{"1 space", 1},
		{"4 spaces", 4},
		{"8 spaces", 8},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := pack(testOptsWithIndent("../testdata/simple/input", "json", false, false, tt.indent), nil)
			if err != nil {
				t.Fatalf("pack() error = %v", err)
			}

			resultStr := string(result)
			indentStr := strings.Repeat(" ", tt.indent)
			expectedPrefix := indentStr + `"entities": {`

			if !strings.Contains(resultStr, expectedPrefix) {
				t.Errorf("pack() JSON output should use %d-space indent. Expected line starting with %q. Got:\n%s", tt.indent, expectedPrefix, resultStr)
			}
		})
	}
}

func TestPack_Indent_Invalid(t *testing.T) {
	// Test that invalid indent values are rejected
	tests := []struct {
		name   string
		indent int
	}{
		{"zero", 0},
		{"negative", -1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := pack(testOptsWithIndent("../testdata/simple/input", "yaml", false, false, tt.indent), nil)
			assertErrorContains(t, err, "invalid indent")
		})
	}
}

func TestPack_NonStringKeys_Canonical(t *testing.T) {
	// Bug fix: canonical mode would fail with mapstructure.Decode error on non-string keys
	tmpDir := createTestDir(t, map[string]string{
		"test.yml": `123: "numeric key"
true: "boolean key"
nested:
  456: "nested numeric"`,
	}, nil)

	// Before the fix, this would error with: mapstructure decode error on non-string key
	_, err := pack(testOpts(tmpDir, "yaml", false, false, "canonical"), nil)
	assertNoError(t, err)
}

func TestPack_NonStringKeys_Preserve_JSON(t *testing.T) {
	// Bug fix: preserve mode + JSON would fail with "unsupported type: map[interface{}]interface{}"
	tmpDir := createTestDir(t, map[string]string{
		"test.yml": `123: "numeric key"
true: "boolean key"
nested:
  456: "nested numeric"`,
	}, nil)

	result, err := pack(testOpts(tmpDir, "json", false, false, "preserve"), nil)
	assertNoError(t, err)

	// Verify it's valid JSON (the bug was json.Marshal failing entirely)
	var jsonData interface{}
	if err := json.Unmarshal(result, &jsonData); err != nil {
		t.Errorf("Output is not valid JSON: %v", err)
	}
}

func TestPack_DeepMerge_Canonical(t *testing.T) {
	// Test deep merge in canonical mode with nested maps
	tmpDir := createTestDir(t, map[string]string{
		"@base.yml": `config:
  setting1: value1
  setting2: value2
  nested:
    a: 1
    b: 2`,
		"@override.yml": `config:
  setting3: value3
  nested:
    c: 3`,
	}, nil)

	result, err := pack(testOptsWithMerge(tmpDir, "yaml", false, false, "deep", "canonical"), nil)
	assertNoError(t, err)

	resultStr := string(result)

	// In deep merge, values from both files should exist
	if !strings.Contains(resultStr, "setting1: value1") {
		t.Error("Deep merge should preserve setting1 from base file")
	}
	if !strings.Contains(resultStr, "setting2: value2") {
		t.Error("Deep merge should preserve setting2 from base file")
	}
	if !strings.Contains(resultStr, "setting3: value3") {
		t.Error("Deep merge should include setting3 from override file")
	}

	// Nested map should be merged recursively
	if !strings.Contains(resultStr, "a: 1") {
		t.Error("Deep merge should preserve 'a' from base file")
	}
	if !strings.Contains(resultStr, "b: 2") {
		t.Error("Deep merge should preserve 'b' from base file")
	}
	if !strings.Contains(resultStr, "c: 3") {
		t.Error("Deep merge should include 'c' from override file")
	}
}

func TestPack_DeepMerge_Preserve(t *testing.T) {
	// Test deep merge in preserve mode with nested maps
	tmpDir := createTestDir(t, map[string]string{
		"@base.yml": `config:
  setting1: value1
  setting2: value2
  nested:
    a: 1
    b: 2`,
		"@override.yml": `config:
  setting3: value3
  nested:
    c: 3`,
	}, nil)

	result, err := pack(testOptsWithMerge(tmpDir, "yaml", false, false, "deep", "preserve"), nil)
	assertNoError(t, err)

	resultStr := string(result)

	// In deep merge, values from both files should exist
	if !strings.Contains(resultStr, "setting1: value1") {
		t.Error("Deep merge should preserve setting1 from base file")
	}
	if !strings.Contains(resultStr, "setting2: value2") {
		t.Error("Deep merge should preserve setting2 from base file")
	}
	if !strings.Contains(resultStr, "setting3: value3") {
		t.Error("Deep merge should include setting3 from override file")
	}

	// Nested map should be merged recursively
	if !strings.Contains(resultStr, "a: 1") {
		t.Error("Deep merge should preserve 'a' from base file")
	}
	if !strings.Contains(resultStr, "b: 2") {
		t.Error("Deep merge should preserve 'b' from base file")
	}
	if !strings.Contains(resultStr, "c: 3") {
		t.Error("Deep merge should include 'c' from override file")
	}
}

func TestPack_ShallowMerge_Default(t *testing.T) {
	// Test that default (shallow) merge replaces entire nested maps
	tmpDir := createTestDir(t, map[string]string{
		"@base.yml": `config:
  setting1: value1
  setting2: value2
  nested:
    a: 1
    b: 2`,
		"@override.yml": `config:
  setting3: value3
  nested:
    c: 3`,
	}, nil)

	result, err := pack(testOpts(tmpDir, "yaml", false, false, "canonical"), nil)
	assertNoError(t, err)

	resultStr := string(result)

	// In shallow merge (default), later file completely replaces earlier one
	if strings.Contains(resultStr, "setting1:") {
		t.Error("Shallow merge (default) should replace entire map - setting1 should not exist")
	}
	if strings.Contains(resultStr, "setting2:") {
		t.Error("Shallow merge (default) should replace entire map - setting2 should not exist")
	}
	if !strings.Contains(resultStr, "setting3: value3") {
		t.Error("Shallow merge should include values from later file")
	}

	// Nested map should also be completely replaced
	if strings.Contains(resultStr, "a: 1") {
		t.Error("Shallow merge should replace nested map - 'a' should not exist")
	}
	if strings.Contains(resultStr, "b: 2") {
		t.Error("Shallow merge should replace nested map - 'b' should not exist")
	}
	if !strings.Contains(resultStr, "c: 3") {
		t.Error("Shallow merge should include nested values from later file")
	}
}

func TestPack_NonStringKeys_Canonical_YAML_Preserved(t *testing.T) {
	// Test that non-string keys are preserved in canonical YAML output
	tmpDir := createTestDir(t, map[string]string{
		"config.yml": `123: numeric key
true: boolean key
string_key: string value`,
	}, nil)

	result, err := pack(testOpts(tmpDir, "yaml", false, false, "canonical"), nil)
	assertNoError(t, err)

	resultStr := string(result)

	// Numeric key should be preserved as number (not quoted as "123")
	if !strings.Contains(resultStr, "123:") {
		t.Errorf("Non-string numeric key should be preserved. Got:\n%s", resultStr)
	}
	// Should not be quoted
	if strings.Contains(resultStr, `"123":`) {
		t.Errorf("Numeric key should not be quoted as string. Got:\n%s", resultStr)
	}

	// Boolean key should be preserved
	if !strings.Contains(resultStr, "true:") {
		t.Errorf("Non-string boolean key should be preserved. Got:\n%s", resultStr)
	}

	// String key should work normally
	if !strings.Contains(resultStr, "string_key:") {
		t.Errorf("String key should be preserved. Got:\n%s", resultStr)
	}
}

func TestPack_NonStringKeys_Canonical_JSON_Normalized(t *testing.T) {
	// Test that non-string keys are normalized to strings for JSON output
	tmpDir := createTestDir(t, map[string]string{
		"config.yml": `123: numeric key
true: boolean key
string_key: string value`,
	}, nil)

	result, err := pack(testOpts(tmpDir, "json", false, false, "canonical"), nil)
	assertNoError(t, err)

	resultStr := string(result)

	// JSON requires string keys, so numeric key should be converted to string "123"
	if !strings.Contains(resultStr, `"123":`) {
		t.Errorf("Numeric key should be normalized to string for JSON. Got:\n%s", resultStr)
	}

	// Boolean key should be converted to string "true"
	if !strings.Contains(resultStr, `"true":`) {
		t.Errorf("Boolean key should be normalized to string for JSON. Got:\n%s", resultStr)
	}

	// String key should be quoted
	if !strings.Contains(resultStr, `"string_key":`) {
		t.Errorf("String key should be quoted in JSON. Got:\n%s", resultStr)
	}
}
