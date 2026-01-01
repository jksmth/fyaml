package include

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"gopkg.in/yaml.v3"
)

func TestMaybeIncludeFile_NoInclude(t *testing.T) {
	// Regular strings without include directives should pass through unchanged
	tests := []string{
		"hello world",
		"echo 'Hello World'",
		"some: yaml: value",
		"",
		"<<notaninclude>>",
		"include(file.txt)",
	}

	packRoot := "/tmp"
	for _, s := range tests {
		result, err := MaybeIncludeFile(s, "/tmp", packRoot)
		if err != nil {
			t.Errorf("MaybeIncludeFile(%q) error = %v", s, err)
		}
		if result != s {
			t.Errorf("MaybeIncludeFile(%q) = %q, want %q", s, result, s)
		}
	}
}

func TestMaybeIncludeFile_ValidInclude(t *testing.T) {
	// Create a temp directory with a test file
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.sh")
	testContent := "#!/bin/bash\necho 'Hello World'"

	if err := os.WriteFile(testFile, []byte(testContent), 0600); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	tests := []struct {
		input    string
		expected string
	}{
		{"<<include(test.sh)>>", testContent},
		{"<< include(test.sh) >>", testContent},
		{"<<  include(test.sh)  >>", testContent},
	}

	absTmpDir, err := filepath.Abs(tmpDir)
	if err != nil {
		t.Fatalf("Failed to get absolute path: %v", err)
	}

	for _, tt := range tests {
		result, err := MaybeIncludeFile(tt.input, tmpDir, absTmpDir)
		if err != nil {
			t.Errorf("MaybeIncludeFile(%q) error = %v", tt.input, err)
		}
		if result != tt.expected {
			t.Errorf("MaybeIncludeFile(%q) = %q, want %q", tt.input, result, tt.expected)
		}
	}
}

func TestMaybeIncludeFile_SubdirectoryInclude(t *testing.T) {
	// Create a temp directory with a nested structure
	tmpDir := t.TempDir()
	scriptsDir := filepath.Join(tmpDir, "scripts")
	if err := os.Mkdir(scriptsDir, 0700); err != nil {
		t.Fatalf("Failed to create scripts directory: %v", err)
	}

	testFile := filepath.Join(scriptsDir, "hello.sh")
	testContent := "echo 'Hello from subdirectory'"

	if err := os.WriteFile(testFile, []byte(testContent), 0600); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	absTmpDir, err := filepath.Abs(tmpDir)
	if err != nil {
		t.Fatalf("Failed to get absolute path: %v", err)
	}

	result, err := MaybeIncludeFile("<<include(scripts/hello.sh)>>", tmpDir, absTmpDir)
	if err != nil {
		t.Errorf("MaybeIncludeFile() error = %v", err)
	}
	if result != testContent {
		t.Errorf("MaybeIncludeFile() = %q, want %q", result, testContent)
	}
}

func TestMaybeIncludeFile_PreservesDoubleAngleBrackets(t *testing.T) {
	// Content with << should be preserved as-is (not escaped)
	// Since the regex requires exact matching, there's no risk of accidental processing
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.sh")
	testContent := "echo '<<something>>'"

	if err := os.WriteFile(testFile, []byte(testContent), 0600); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	absTmpDir, err := filepath.Abs(tmpDir)
	if err != nil {
		t.Fatalf("Failed to get absolute path: %v", err)
	}

	result, err := MaybeIncludeFile("<<include(test.sh)>>", tmpDir, absTmpDir)
	if err != nil {
		t.Errorf("MaybeIncludeFile() error = %v", err)
	}

	// Content should be returned exactly as-is, without escaping
	expected := "echo '<<something>>'"
	if result != expected {
		t.Errorf("MaybeIncludeFile() = %q, want %q", result, expected)
	}
}

func TestMaybeIncludeFile_ErrorMultipleIncludes(t *testing.T) {
	input := "<<include(a.sh)>> <<include(b.sh)>>"

	_, err := MaybeIncludeFile(input, "/tmp", "/tmp")
	if err == nil {
		t.Error("MaybeIncludeFile() expected error for multiple includes")
	}
	if !strings.Contains(err.Error(), "multiple include statements") {
		t.Errorf("MaybeIncludeFile() error = %v, want 'multiple include statements'", err)
	}
}

func TestMaybeIncludeFile_ErrorPartialMatch(t *testing.T) {
	tests := []string{
		"echo <<include(test.sh)>>",
		"<<include(test.sh)>> && exit",
		"prefix <<include(test.sh)>> suffix",
	}

	for _, input := range tests {
		_, err := MaybeIncludeFile(input, "/tmp", "/tmp")
		if err == nil {
			t.Errorf("MaybeIncludeFile(%q) expected error for partial match", input)
		}
		if !strings.Contains(err.Error(), "entire string must be include statement") {
			t.Errorf("MaybeIncludeFile(%q) error = %v, want 'entire string must be include statement'", input, err)
		}
	}
}

func TestMaybeIncludeFile_ErrorFileNotFound(t *testing.T) {
	tmpDir := t.TempDir()

	absTmpDir, err := filepath.Abs(tmpDir)
	if err != nil {
		t.Fatalf("Failed to get absolute path: %v", err)
	}

	_, err = MaybeIncludeFile("<<include(nonexistent.sh)>>", tmpDir, absTmpDir)
	if err == nil {
		t.Error("MaybeIncludeFile() expected error for missing file")
	}
	if !strings.Contains(err.Error(), "could not open") {
		t.Errorf("MaybeIncludeFile() error = %v, want 'could not open'", err)
	}
}

func TestMaybeIncludeFile_RelativePathsWithParent(t *testing.T) {
	// Test that relative paths with ../ work (sibling directory access)
	tmpDir := t.TempDir()
	baseDir := filepath.Join(tmpDir, "config", "commands")
	scriptsDir := filepath.Join(tmpDir, "config", "scripts")
	if err := os.MkdirAll(baseDir, 0700); err != nil {
		t.Fatalf("Failed to create base directory: %v", err)
	}
	if err := os.MkdirAll(scriptsDir, 0700); err != nil {
		t.Fatalf("Failed to create scripts directory: %v", err)
	}

	// Create a file in the scripts directory (sibling to commands)
	scriptFile := filepath.Join(scriptsDir, "script.sh")
	scriptContent := "echo 'relative path with ..'"
	if err := os.WriteFile(scriptFile, []byte(scriptContent), 0600); err != nil {
		t.Fatalf("Failed to create script: %v", err)
	}

	// Include from commands to scripts (../scripts/script.sh)
	absTmpDir, err := filepath.Abs(tmpDir)
	if err != nil {
		t.Fatalf("Failed to get absolute path: %v", err)
	}

	result, err := MaybeIncludeFile("<<include(../scripts/script.sh)>>", baseDir, absTmpDir)
	if err != nil {
		t.Errorf("MaybeIncludeFile() error = %v, expected relative path to work", err)
	}
	if result != scriptContent {
		t.Errorf("MaybeIncludeFile() = %q, want %q", result, scriptContent)
	}
}

func TestInlineIncludes_ScalarNode(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "script.sh")
	testContent := "echo 'Hello'"

	if err := os.WriteFile(testFile, []byte(testContent), 0600); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	node := &yaml.Node{
		Kind:  yaml.ScalarNode,
		Value: "<<include(script.sh)>>",
	}

	absTmpDir, err := filepath.Abs(tmpDir)
	if err != nil {
		t.Fatalf("Failed to get absolute path: %v", err)
	}

	err = InlineIncludes(node, tmpDir, absTmpDir)
	if err != nil {
		t.Errorf("InlineIncludes() error = %v", err)
	}
	if node.Value != testContent {
		t.Errorf("InlineIncludes() node.Value = %q, want %q", node.Value, testContent)
	}
}

func TestInlineIncludes_NestedDocument(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "script.sh")
	testContent := "#!/bin/bash\necho 'Hello'"

	if err := os.WriteFile(testFile, []byte(testContent), 0600); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	// Create a YAML document with nested structure
	yamlContent := `
description: A test command
steps:
  - run:
      name: Hello
      command: <<include(script.sh)>>
`

	var node yaml.Node
	if err := yaml.Unmarshal([]byte(yamlContent), &node); err != nil {
		t.Fatalf("Failed to unmarshal YAML: %v", err)
	}

	absTmpDir, err := filepath.Abs(tmpDir)
	if err != nil {
		t.Fatalf("Failed to get absolute path: %v", err)
	}

	err = InlineIncludes(&node, tmpDir, absTmpDir)
	if err != nil {
		t.Errorf("InlineIncludes() error = %v", err)
	}

	// Marshal back and verify include was replaced
	out, err := yaml.Marshal(&node)
	if err != nil {
		t.Fatalf("Failed to marshal YAML: %v", err)
	}

	if !strings.Contains(string(out), "echo 'Hello'") {
		t.Errorf("InlineIncludes() output does not contain included content:\n%s", out)
	}
	if strings.Contains(string(out), "<<include") {
		t.Errorf("InlineIncludes() output still contains include directive:\n%s", out)
	}
}

func TestInlineIncludes_NilNode(t *testing.T) {
	// Should not panic on nil node
	err := InlineIncludes(nil, "/tmp", "/tmp")
	if err != nil {
		t.Errorf("InlineIncludes(nil) error = %v", err)
	}
}

func TestInlineIncludes_EmptyScalar(t *testing.T) {
	// Empty scalar values should be skipped
	node := &yaml.Node{
		Kind:  yaml.ScalarNode,
		Value: "",
	}

	err := InlineIncludes(node, "/tmp", "/tmp")
	if err != nil {
		t.Errorf("InlineIncludes() error = %v", err)
	}
	if node.Value != "" {
		t.Errorf("InlineIncludes() node.Value = %q, want empty", node.Value)
	}
}

func TestInlineIncludes_PreservesNonIncludeValues(t *testing.T) {
	// Non-include scalar values should be preserved
	node := &yaml.Node{
		Kind:  yaml.ScalarNode,
		Value: "echo 'Hello World'",
	}

	err := InlineIncludes(node, "/tmp", "/tmp")
	if err != nil {
		t.Errorf("InlineIncludes() error = %v", err)
	}
	if node.Value != "echo 'Hello World'" {
		t.Errorf("InlineIncludes() node.Value = %q, want 'echo 'Hello World''", node.Value)
	}
}

func TestMaybeIncludeFile_AbsolutePathWithinPackRoot(t *testing.T) {
	// Test that absolute paths within pack root are allowed
	tmpDir := t.TempDir()
	absTmpDir, err := filepath.Abs(tmpDir)
	if err != nil {
		t.Fatalf("Failed to get absolute path: %v", err)
	}

	scriptFile := filepath.Join(tmpDir, "script.sh")
	scriptContent := "echo 'absolute path within root'"
	if err := os.WriteFile(scriptFile, []byte(scriptContent), 0600); err != nil {
		t.Fatalf("Failed to create script: %v", err)
	}

	absScriptFile, err := filepath.Abs(scriptFile)
	if err != nil {
		t.Fatalf("Failed to get absolute script path: %v", err)
	}

	// Use absolute path in include - should work since it's within pack root
	result, err := MaybeIncludeFile(fmt.Sprintf("<<include(%s)>>", absScriptFile), tmpDir, absTmpDir)
	if err != nil {
		t.Errorf("MaybeIncludeFile() with absolute path within root error = %v", err)
	}
	if result != scriptContent {
		t.Errorf("MaybeIncludeFile() = %q, want %q", result, scriptContent)
	}
}

func TestMaybeIncludeFile_AbsolutePathOutsidePackRoot(t *testing.T) {
	// Test that absolute paths outside pack root are rejected
	tmpDir := t.TempDir()
	absTmpDir, err := filepath.Abs(tmpDir)
	if err != nil {
		t.Fatalf("Failed to get absolute path: %v", err)
	}

	// Create a file outside the pack root
	outsideDir := t.TempDir()
	scriptFile := filepath.Join(outsideDir, "script.sh")
	scriptContent := "echo 'outside root'"
	if err := os.WriteFile(scriptFile, []byte(scriptContent), 0600); err != nil {
		t.Fatalf("Failed to create script: %v", err)
	}

	absScriptFile, err := filepath.Abs(scriptFile)
	if err != nil {
		t.Fatalf("Failed to get absolute script path: %v", err)
	}

	// Use absolute path outside pack root - should fail
	_, err = MaybeIncludeFile(fmt.Sprintf("<<include(%s)>>", absScriptFile), tmpDir, absTmpDir)
	if err == nil {
		t.Error("MaybeIncludeFile() expected error for absolute path outside pack root")
	}
	if !strings.Contains(err.Error(), "escapes pack root") {
		t.Errorf("MaybeIncludeFile() error = %v, want 'escapes pack root'", err)
	}
}

func TestMaybeIncludeFile_RelativePathEscapesPackRoot(t *testing.T) {
	// Test that relative paths that escape pack root are rejected
	tmpDir := t.TempDir()
	absTmpDir, err := filepath.Abs(tmpDir)
	if err != nil {
		t.Fatalf("Failed to get absolute path: %v", err)
	}

	// Create a file outside the pack root
	outsideDir := t.TempDir()
	scriptFile := filepath.Join(outsideDir, "script.sh")
	scriptContent := "echo 'outside root'"
	if err := os.WriteFile(scriptFile, []byte(scriptContent), 0600); err != nil {
		t.Fatalf("Failed to create script: %v", err)
	}

	// Create a subdirectory in pack root
	subDir := filepath.Join(tmpDir, "subdir")
	if err := os.MkdirAll(subDir, 0700); err != nil {
		t.Fatalf("Failed to create subdir: %v", err)
	}

	// Try to include file outside using ../ - should fail
	// Calculate relative path that would escape
	relPath, err := filepath.Rel(subDir, scriptFile)
	if err != nil {
		t.Fatalf("Failed to calculate relative path: %v", err)
	}

	_, err = MaybeIncludeFile(fmt.Sprintf("<<include(%s)>>", relPath), subDir, absTmpDir)
	if err == nil {
		t.Error("MaybeIncludeFile() expected error for path escaping pack root")
	}
	if !strings.Contains(err.Error(), "escapes pack root") {
		t.Errorf("MaybeIncludeFile() error = %v, want 'escapes pack root'", err)
	}
}
