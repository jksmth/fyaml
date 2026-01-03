package include

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"go.yaml.in/yaml/v4"
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

func TestMaybeIncludeFile_OpenRootError(t *testing.T) {
	// Test error when pack root doesn't exist
	// The path validation catches this before os.OpenRoot, resulting in "escapes pack root" error
	tmpDir := t.TempDir()
	nonexistentRoot := filepath.Join(tmpDir, "nonexistent")

	_, err := MaybeIncludeFile("<<include(test.sh)>>", tmpDir, nonexistentRoot)
	if err == nil {
		t.Error("MaybeIncludeFile() expected error for non-existent pack root")
	}
	// When pack root doesn't exist, the relative path calculation results in ".." prefix
	// which triggers the "escapes pack root" validation error
	if !strings.Contains(err.Error(), "escapes pack root") {
		t.Errorf("MaybeIncludeFile() error = %v, want 'escapes pack root'", err)
	}
}

func TestInlineIncludes_ErrorPropagation(t *testing.T) {
	// Test that errors from MaybeIncludeFile are propagated correctly
	tmpDir := t.TempDir()
	nonexistentRoot := filepath.Join(tmpDir, "nonexistent")

	node := &yaml.Node{
		Kind:  yaml.ScalarNode,
		Value: "<<include(nonexistent.sh)>>",
	}

	err := InlineIncludes(node, tmpDir, nonexistentRoot)
	if err == nil {
		t.Error("InlineIncludes() expected error for invalid include")
	}
	// Error should be propagated from MaybeIncludeFile
	// When pack root doesn't exist, this results in "escapes pack root" error
	if !strings.Contains(err.Error(), "escapes pack root") {
		t.Errorf("InlineIncludes() error = %v, want 'escapes pack root'", err)
	}
}

// =============================================================================
// Tag-based include tests (!include and !include-text)
// =============================================================================

func TestProcessIncludeTag_YAMLFile(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a YAML file to include
	includeFile := filepath.Join(tmpDir, "defaults.yml")
	includeContent := "timeout: 30\nretries: 3"
	if err := os.WriteFile(includeFile, []byte(includeContent), 0600); err != nil {
		t.Fatalf("Failed to write include file: %v", err)
	}

	// Create a YAML document with !include tag
	yamlContent := `name: api
config: !include defaults.yml`

	var node yaml.Node
	if err := yaml.Unmarshal([]byte(yamlContent), &node); err != nil {
		t.Fatalf("Failed to unmarshal YAML: %v", err)
	}

	absTmpDir, err := filepath.Abs(tmpDir)
	if err != nil {
		t.Fatalf("Failed to get absolute path: %v", err)
	}

	err = ProcessIncludeTag(&node, tmpDir, absTmpDir)
	if err != nil {
		t.Errorf("ProcessIncludeTag() error = %v", err)
	}

	// Marshal back and verify include was replaced
	out, err := yaml.Marshal(&node)
	if err != nil {
		t.Fatalf("Failed to marshal YAML: %v", err)
	}

	if !strings.Contains(string(out), "timeout: 30") {
		t.Errorf("ProcessIncludeTag() output does not contain included content:\n%s", out)
	}
	if !strings.Contains(string(out), "retries: 3") {
		t.Errorf("ProcessIncludeTag() output does not contain included content:\n%s", out)
	}
	if strings.Contains(string(out), "!include") {
		t.Errorf("ProcessIncludeTag() output still contains !include tag:\n%s", out)
	}
}

func TestProcessIncludeTag_NestedYAML(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a YAML file with nested structure
	includeFile := filepath.Join(tmpDir, "defaults.yml")
	includeContent := `settings:
  timeout: 30
  retries: 3`
	if err := os.WriteFile(includeFile, []byte(includeContent), 0600); err != nil {
		t.Fatalf("Failed to write include file: %v", err)
	}

	// Create a YAML document with !include tag
	yamlContent := `name: api
config: !include defaults.yml`

	var node yaml.Node
	if err := yaml.Unmarshal([]byte(yamlContent), &node); err != nil {
		t.Fatalf("Failed to unmarshal YAML: %v", err)
	}

	absTmpDir, err := filepath.Abs(tmpDir)
	if err != nil {
		t.Fatalf("Failed to get absolute path: %v", err)
	}

	err = ProcessIncludeTag(&node, tmpDir, absTmpDir)
	if err != nil {
		t.Errorf("ProcessIncludeTag() error = %v", err)
	}

	// Marshal back and verify structure
	out, err := yaml.Marshal(&node)
	if err != nil {
		t.Fatalf("Failed to marshal YAML: %v", err)
	}

	if !strings.Contains(string(out), "settings:") {
		t.Errorf("ProcessIncludeTag() output does not contain nested structure:\n%s", out)
	}
}

func TestProcessIncludeTextTag_TextFile(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a text file to include
	scriptFile := filepath.Join(tmpDir, "script.sh")
	scriptContent := "#!/bin/bash\necho 'Hello World'"
	if err := os.WriteFile(scriptFile, []byte(scriptContent), 0600); err != nil {
		t.Fatalf("Failed to write script file: %v", err)
	}

	// Create a YAML document with !include-text tag
	yamlContent := `name: deploy
command: !include-text script.sh`

	var node yaml.Node
	if err := yaml.Unmarshal([]byte(yamlContent), &node); err != nil {
		t.Fatalf("Failed to unmarshal YAML: %v", err)
	}

	absTmpDir, err := filepath.Abs(tmpDir)
	if err != nil {
		t.Fatalf("Failed to get absolute path: %v", err)
	}

	err = ProcessIncludeTextTag(&node, tmpDir, absTmpDir)
	if err != nil {
		t.Errorf("ProcessIncludeTextTag() error = %v", err)
	}

	// Marshal back and verify include was replaced with text
	out, err := yaml.Marshal(&node)
	if err != nil {
		t.Fatalf("Failed to marshal YAML: %v", err)
	}

	if !strings.Contains(string(out), "echo 'Hello World'") {
		t.Errorf("ProcessIncludeTextTag() output does not contain script content:\n%s", out)
	}
	if strings.Contains(string(out), "!include-text") {
		t.Errorf("ProcessIncludeTextTag() output still contains !include-text tag:\n%s", out)
	}
}

func TestProcessIncludeTextTag_EquivalentToDirective(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a text file
	scriptFile := filepath.Join(tmpDir, "script.sh")
	scriptContent := "echo 'test content'"
	if err := os.WriteFile(scriptFile, []byte(scriptContent), 0600); err != nil {
		t.Fatalf("Failed to write script file: %v", err)
	}

	absTmpDir, err := filepath.Abs(tmpDir)
	if err != nil {
		t.Fatalf("Failed to get absolute path: %v", err)
	}

	// Test with !include-text tag
	yamlContentTag := `command: !include-text script.sh`
	var nodeTag yaml.Node
	if err := yaml.Unmarshal([]byte(yamlContentTag), &nodeTag); err != nil {
		t.Fatalf("Failed to unmarshal YAML: %v", err)
	}
	err = ProcessIncludeTextTag(&nodeTag, tmpDir, absTmpDir)
	if err != nil {
		t.Errorf("ProcessIncludeTextTag() error = %v", err)
	}

	// Test with <<include()>> directive
	yamlContentDirective := `command: <<include(script.sh)>>`
	var nodeDirective yaml.Node
	if err := yaml.Unmarshal([]byte(yamlContentDirective), &nodeDirective); err != nil {
		t.Fatalf("Failed to unmarshal YAML: %v", err)
	}
	err = InlineIncludes(&nodeDirective, tmpDir, absTmpDir)
	if err != nil {
		t.Errorf("InlineIncludes() error = %v", err)
	}

	// Both should produce equivalent content when decoded
	var resultTag, resultDirective interface{}
	if err := nodeTag.Decode(&resultTag); err != nil {
		t.Fatalf("Failed to decode tag result: %v", err)
	}
	if err := nodeDirective.Decode(&resultDirective); err != nil {
		t.Fatalf("Failed to decode directive result: %v", err)
	}

	// Compare the decoded values
	tagMap := resultTag.(map[string]interface{})
	directiveMap := resultDirective.(map[string]interface{})

	if tagMap["command"] != directiveMap["command"] {
		t.Errorf("!include-text and <<include()>> produced different content:\nTag: %v\nDirective: %v",
			tagMap["command"], directiveMap["command"])
	}
}

func TestProcessIncludes_CombinedUsage(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a YAML file for !include
	yamlFile := filepath.Join(tmpDir, "metadata.yml")
	yamlContent := "version: 1.0\nauthor: test"
	if err := os.WriteFile(yamlFile, []byte(yamlContent), 0600); err != nil {
		t.Fatalf("Failed to write YAML file: %v", err)
	}

	// Create a text file for !include-text
	scriptFile := filepath.Join(tmpDir, "script.sh")
	scriptContent := "echo 'deploy'"
	if err := os.WriteFile(scriptFile, []byte(scriptContent), 0600); err != nil {
		t.Fatalf("Failed to write script file: %v", err)
	}

	// Create a text file for <<include()>>
	validateFile := filepath.Join(tmpDir, "validate.sh")
	validateContent := "echo 'validate'"
	if err := os.WriteFile(validateFile, []byte(validateContent), 0600); err != nil {
		t.Fatalf("Failed to write validate file: %v", err)
	}

	// Create a YAML document with all three include mechanisms
	mainContent := `name: deploy
metadata: !include metadata.yml
steps:
  - deploy: !include-text script.sh
  - validate: <<include(validate.sh)>>`

	var node yaml.Node
	if err := yaml.Unmarshal([]byte(mainContent), &node); err != nil {
		t.Fatalf("Failed to unmarshal YAML: %v", err)
	}

	absTmpDir, err := filepath.Abs(tmpDir)
	if err != nil {
		t.Fatalf("Failed to get absolute path: %v", err)
	}

	// Process all includes
	err = ProcessIncludes(&node, tmpDir, absTmpDir)
	if err != nil {
		t.Errorf("ProcessIncludes() error = %v", err)
	}

	// Marshal and verify all includes were processed
	out, err := yaml.Marshal(&node)
	if err != nil {
		t.Fatalf("Failed to marshal YAML: %v", err)
	}

	outStr := string(out)
	if !strings.Contains(outStr, "version: 1.0") {
		t.Errorf("ProcessIncludes() did not process !include:\n%s", outStr)
	}
	if !strings.Contains(outStr, "echo 'deploy'") {
		t.Errorf("ProcessIncludes() did not process !include-text:\n%s", outStr)
	}
	if !strings.Contains(outStr, "echo 'validate'") {
		t.Errorf("ProcessIncludes() did not process <<include()>>:\n%s", outStr)
	}
	if strings.Contains(outStr, "!include") || strings.Contains(outStr, "<<include") {
		t.Errorf("ProcessIncludes() output still contains include markers:\n%s", outStr)
	}
}

func TestProcessIncludes_NilNode(t *testing.T) {
	err := ProcessIncludes(nil, "/tmp", "/tmp")
	if err != nil {
		t.Errorf("ProcessIncludes(nil) error = %v", err)
	}
}

func TestProcessIncludeTag_ErrorInvalidNodeType(t *testing.T) {
	// !include tag on non-scalar should error
	yamlContent := `config: !include
  - item1
  - item2`

	var node yaml.Node
	if err := yaml.Unmarshal([]byte(yamlContent), &node); err != nil {
		t.Fatalf("Failed to unmarshal YAML: %v", err)
	}

	err := ProcessIncludeTag(&node, "/tmp", "/tmp")
	if err == nil {
		t.Error("ProcessIncludeTag() expected error for non-scalar node")
	}
	if !strings.Contains(err.Error(), "must be used on a scalar value") {
		t.Errorf("ProcessIncludeTag() error = %v, want 'must be used on a scalar value'", err)
	}
}

func TestProcessIncludeTextTag_ErrorInvalidNodeType(t *testing.T) {
	// !include-text tag on non-scalar should error
	yamlContent := `config: !include-text
  - item1
  - item2`

	var node yaml.Node
	if err := yaml.Unmarshal([]byte(yamlContent), &node); err != nil {
		t.Fatalf("Failed to unmarshal YAML: %v", err)
	}

	err := ProcessIncludeTextTag(&node, "/tmp", "/tmp")
	if err == nil {
		t.Error("ProcessIncludeTextTag() expected error for non-scalar node")
	}
	if !strings.Contains(err.Error(), "must be used on a scalar value") {
		t.Errorf("ProcessIncludeTextTag() error = %v, want 'must be used on a scalar value'", err)
	}
}

func TestProcessIncludeTag_PathEscaping(t *testing.T) {
	tmpDir := t.TempDir()
	absTmpDir, err := filepath.Abs(tmpDir)
	if err != nil {
		t.Fatalf("Failed to get absolute path: %v", err)
	}

	// Create a file outside pack root
	outsideDir := t.TempDir()
	outsideFile := filepath.Join(outsideDir, "secret.yml")
	if err := os.WriteFile(outsideFile, []byte("secret: data"), 0600); err != nil {
		t.Fatalf("Failed to write file: %v", err)
	}

	// Try to include file outside pack root
	relPath, _ := filepath.Rel(tmpDir, outsideFile)
	yamlContent := fmt.Sprintf(`config: !include %s`, relPath)

	var node yaml.Node
	if err := yaml.Unmarshal([]byte(yamlContent), &node); err != nil {
		t.Fatalf("Failed to unmarshal YAML: %v", err)
	}

	err = ProcessIncludeTag(&node, tmpDir, absTmpDir)
	if err == nil {
		t.Error("ProcessIncludeTag() expected error for path escaping")
	}
	if !strings.Contains(err.Error(), "escapes pack root") {
		t.Errorf("ProcessIncludeTag() error = %v, want 'escapes pack root'", err)
	}
}

func TestLoadFileText_Valid(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.txt")
	testContent := "hello world"
	if err := os.WriteFile(testFile, []byte(testContent), 0600); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	absTmpDir, err := filepath.Abs(tmpDir)
	if err != nil {
		t.Fatalf("Failed to get absolute path: %v", err)
	}

	result, err := LoadFileText("test.txt", tmpDir, absTmpDir)
	if err != nil {
		t.Errorf("LoadFileText() error = %v", err)
	}
	if result != testContent {
		t.Errorf("LoadFileText() = %q, want %q", result, testContent)
	}
}

func TestLoadFileFragment_Valid(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.yml")
	testContent := "key: value"
	if err := os.WriteFile(testFile, []byte(testContent), 0600); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	absTmpDir, err := filepath.Abs(tmpDir)
	if err != nil {
		t.Fatalf("Failed to get absolute path: %v", err)
	}

	fragment, err := LoadFileFragment("test.yml", tmpDir, absTmpDir)
	if err != nil {
		t.Errorf("LoadFileFragment() error = %v", err)
	}
	if fragment == nil {
		t.Error("LoadFileFragment() returned nil fragment")
	}
}

func TestLoadFileFragment_InvalidYAML(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "invalid.yml")
	// Invalid YAML - unclosed bracket
	testContent := "key: [unclosed"
	if err := os.WriteFile(testFile, []byte(testContent), 0600); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	absTmpDir, err := filepath.Abs(tmpDir)
	if err != nil {
		t.Fatalf("Failed to get absolute path: %v", err)
	}

	_, err = LoadFileFragment("invalid.yml", tmpDir, absTmpDir)
	if err == nil {
		t.Error("LoadFileFragment() expected error for invalid YAML")
	}
	if !strings.Contains(err.Error(), "failed to parse YAML/JSON") {
		t.Errorf("LoadFileFragment() error = %v, want 'failed to parse YAML/JSON'", err)
	}
}

func TestNestedIncludes(t *testing.T) {
	tmpDir := t.TempDir()

	// Create nested include files
	level2File := filepath.Join(tmpDir, "level2.yml")
	if err := os.WriteFile(level2File, []byte("deep: value"), 0600); err != nil {
		t.Fatalf("Failed to write level2 file: %v", err)
	}

	level1File := filepath.Join(tmpDir, "level1.yml")
	level1Content := `nested: !include level2.yml`
	if err := os.WriteFile(level1File, []byte(level1Content), 0600); err != nil {
		t.Fatalf("Failed to write level1 file: %v", err)
	}

	// Main file includes level1
	mainContent := `root: !include level1.yml`
	var node yaml.Node
	if err := yaml.Unmarshal([]byte(mainContent), &node); err != nil {
		t.Fatalf("Failed to unmarshal YAML: %v", err)
	}

	absTmpDir, err := filepath.Abs(tmpDir)
	if err != nil {
		t.Fatalf("Failed to get absolute path: %v", err)
	}

	err = ProcessIncludeTag(&node, tmpDir, absTmpDir)
	if err != nil {
		t.Errorf("ProcessIncludeTag() error = %v", err)
	}

	// Marshal and verify nested includes were processed
	out, err := yaml.Marshal(&node)
	if err != nil {
		t.Fatalf("Failed to marshal YAML: %v", err)
	}

	if !strings.Contains(string(out), "deep: value") {
		t.Errorf("Nested includes not fully processed:\n%s", out)
	}
}

func TestProcessIncludeTag_JSONFile(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a JSON file to include
	jsonFile := filepath.Join(tmpDir, "defaults.json")
	jsonContent := `{"timeout": 30, "retries": 3}`
	if err := os.WriteFile(jsonFile, []byte(jsonContent), 0600); err != nil {
		t.Fatalf("Failed to write JSON file: %v", err)
	}

	// Create a YAML file that includes the JSON file
	yamlContent := `name: api
config: !include defaults.json`

	var node yaml.Node
	if err := yaml.Unmarshal([]byte(yamlContent), &node); err != nil {
		t.Fatalf("Failed to unmarshal YAML: %v", err)
	}

	absTmpDir, err := filepath.Abs(tmpDir)
	if err != nil {
		t.Fatalf("Failed to get absolute path: %v", err)
	}

	err = ProcessIncludeTag(&node, tmpDir, absTmpDir)
	if err != nil {
		t.Errorf("ProcessIncludeTag() error = %v", err)
	}

	// Decode and verify JSON was parsed and included
	var result map[string]interface{}
	if err := node.Decode(&result); err != nil {
		t.Fatalf("Failed to decode result: %v", err)
	}

	config, ok := result["config"].(map[string]interface{})
	if !ok {
		t.Fatalf("config is not a map, got %T: %v", result["config"], result["config"])
	}

	// JSON numbers may be parsed as int or float64 depending on the library
	timeout := config["timeout"]
	if timeout != 30 && timeout != float64(30) {
		t.Errorf("ProcessIncludeTag() did not include JSON content. Got timeout: %v (type: %T)", timeout, timeout)
	}
	retries := config["retries"]
	if retries != 3 && retries != float64(3) {
		t.Errorf("ProcessIncludeTag() did not include JSON content. Got retries: %v (type: %T)", retries, retries)
	}
}

func TestProcessIncludes_JSONFileWithDirective(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a script file
	scriptFile := filepath.Join(tmpDir, "script.sh")
	scriptContent := "echo 'test'"
	if err := os.WriteFile(scriptFile, []byte(scriptContent), 0600); err != nil {
		t.Fatalf("Failed to write script file: %v", err)
	}

	// Create a JSON file with <<include()>> directive
	jsonContent := `{
  "name": "api",
  "command": "<<include(script.sh)>>"
}`

	var node yaml.Node
	if err := yaml.Unmarshal([]byte(jsonContent), &node); err != nil {
		t.Fatalf("Failed to unmarshal JSON: %v", err)
	}

	absTmpDir, err := filepath.Abs(tmpDir)
	if err != nil {
		t.Fatalf("Failed to get absolute path: %v", err)
	}

	// Process includes
	err = ProcessIncludes(&node, tmpDir, absTmpDir)
	if err != nil {
		t.Errorf("ProcessIncludes() error = %v", err)
	}

	// Decode and verify
	var result map[string]interface{}
	if err := node.Decode(&result); err != nil {
		t.Fatalf("Failed to decode result: %v", err)
	}

	if result["command"] != scriptContent {
		t.Errorf("ProcessIncludes() did not process <<include()>> in JSON. Got: %v", result["command"])
	}
}

func TestProcessIncludeTag_JSONFileWithTag(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a JSON file to include
	jsonFile := filepath.Join(tmpDir, "defaults.json")
	jsonContent := `{"timeout": 30, "retries": 3}`
	if err := os.WriteFile(jsonFile, []byte(jsonContent), 0600); err != nil {
		t.Fatalf("Failed to write JSON file: %v", err)
	}

	// Create a JSON file with !include tag (non-standard JSON, but supported)
	mainContent := `{
  "name": "api",
  "config": !include defaults.json
}`

	var node yaml.Node
	if err := yaml.Unmarshal([]byte(mainContent), &node); err != nil {
		t.Fatalf("Failed to unmarshal JSON with tag: %v", err)
	}

	absTmpDir, err := filepath.Abs(tmpDir)
	if err != nil {
		t.Fatalf("Failed to get absolute path: %v", err)
	}

	err = ProcessIncludeTag(&node, tmpDir, absTmpDir)
	if err != nil {
		t.Errorf("ProcessIncludeTag() error = %v", err)
	}

	// Decode and verify
	var result map[string]interface{}
	if err := node.Decode(&result); err != nil {
		t.Fatalf("Failed to decode result: %v", err)
	}

	config, ok := result["config"].(map[string]interface{})
	if !ok {
		t.Fatalf("config is not a map, got %T", result["config"])
	}

	if config["timeout"] != 30 {
		t.Errorf("ProcessIncludeTag() did not include JSON content. Got timeout: %v", config["timeout"])
	}
	if config["retries"] != 3 {
		t.Errorf("ProcessIncludeTag() did not include JSON content. Got retries: %v", config["retries"])
	}
}
