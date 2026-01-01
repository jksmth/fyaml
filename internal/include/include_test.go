package include

import (
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

	for _, s := range tests {
		result, err := MaybeIncludeFile(s, "/tmp")
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

	for _, tt := range tests {
		result, err := MaybeIncludeFile(tt.input, tmpDir)
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

	result, err := MaybeIncludeFile("<<include(scripts/hello.sh)>>", tmpDir)
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

	result, err := MaybeIncludeFile("<<include(test.sh)>>", tmpDir)
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

	_, err := MaybeIncludeFile(input, "/tmp")
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
		_, err := MaybeIncludeFile(input, "/tmp")
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

	_, err := MaybeIncludeFile("<<include(nonexistent.sh)>>", tmpDir)
	if err == nil {
		t.Error("MaybeIncludeFile() expected error for missing file")
	}
	if !strings.Contains(err.Error(), "could not open") {
		t.Errorf("MaybeIncludeFile() error = %v, want 'could not open'", err)
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

	err := InlineIncludes(node, tmpDir)
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

	err := InlineIncludes(&node, tmpDir)
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
	err := InlineIncludes(nil, "/tmp")
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

	err := InlineIncludes(node, "/tmp")
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

	err := InlineIncludes(node, "/tmp")
	if err != nil {
		t.Errorf("InlineIncludes() error = %v", err)
	}
	if node.Value != "echo 'Hello World'" {
		t.Errorf("InlineIncludes() node.Value = %q, want 'echo 'Hello World''", node.Value)
	}
}
