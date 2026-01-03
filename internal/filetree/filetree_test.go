package filetree

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/jksmth/fyaml/internal/logger"
	"go.yaml.in/yaml/v4"
)

// assertNoError asserts that an error is nil. If the error is not nil, the test fails.
func assertNoError(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
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

// createTreeAndMarshal creates a tree from the given directory, marshals it to YAML,
// and returns the result as a map[string]interface{}.
func createTreeAndMarshal(t *testing.T, dir string) map[string]interface{} {
	t.Helper()
	tree, err := NewTree(dir)
	assertNoError(t, err)
	result, err := tree.MarshalYAML()
	assertNoError(t, err)
	resultMap, ok := result.(map[string]interface{})
	if !ok {
		t.Fatalf("MarshalYAML() returned %T, want map[string]interface{}", result)
	}
	return resultMap
}

// findNodeByName recursively searches the tree for a node with the matching name.
// Returns nil if not found.
func findNodeByName(t *testing.T, tree *Node, name string) *Node {
	t.Helper()
	if tree.Info.Name() == name {
		return tree
	}
	for _, child := range tree.Children {
		if found := findNodeByName(t, child, name); found != nil {
			return found
		}
	}
	return nil
}

func TestNewTree(t *testing.T) {
	// Test basic tree building with a temporary directory
	// Integration tests in cmd/fyaml/main_test.go cover full pack() behavior
	tmpDir := createTestDir(t, map[string]string{
		"sub_dir/sub_dir_file.yml": "entity:\n  id: example1\n  attributes:\n    name: sample name",
	}, []string{"empty_dir"})

	tree, err := NewTree(tmpDir)
	assertNoError(t, err)

	if tree == nil {
		t.Fatal("NewTree() returned nil tree")
	}
	if tree.FullPath != tmpDir {
		t.Errorf("NewTree() FullPath = %v, want %v", tree.FullPath, tmpDir)
	}
	if tree.Info.Name() != filepath.Base(tmpDir) {
		t.Errorf("NewTree() Info.Name() = %v, want %v", tree.Info.Name(), filepath.Base(tmpDir))
	}
	if len(tree.Children) != 2 {
		t.Errorf("NewTree() Children length = %v, want 2", len(tree.Children))
	}
}

func TestMarshalYAML_RendersToYAML(t *testing.T) {
	tmpDir := createTestDir(t, map[string]string{
		"sub_dir/sub_dir_file.yml": "entity:\n  id: example1\n  attributes:\n    name: sample name",
	}, []string{"empty_dir"})

	// Verify it contains expected keys
	resultMap := createTreeAndMarshal(t, tmpDir)

	// Verify yaml.Marshal works
	out, err := yaml.Marshal(resultMap)
	assertNoError(t, err)
	if len(out) == 0 {
		t.Error("yaml.Marshal() returned empty output")
	}

	if _, ok := resultMap["sub_dir"]; !ok {
		t.Error("MarshalYAML() result missing 'sub_dir' key")
	}
	// Empty directories (with no YAML files) should not appear in output
	if _, ok := resultMap["empty_dir"]; ok {
		t.Error("MarshalYAML() result should not contain 'empty_dir' key (empty directories are ignored)")
	}
}

func TestNewTree_NonexistentDirectory(t *testing.T) {
	// Test that NewTree returns an error for non-existent directory
	_, err := NewTree("/nonexistent/path/that/does/not/exist")
	assertErrorContains(t, err, "no such file")
}

func TestMarshalYAML_InvalidYAML(t *testing.T) {
	// Test that invalid YAML content causes an error when marshaling
	// This matches the original CircleCI test behavior
	tmpDir := createTestDir(t, map[string]string{
		"another_dir/another_dir_file.yml": "1some: in: valid: yaml",
	}, nil)
	anotherDirFile := filepath.Join(tmpDir, "another_dir", "another_dir_file.yml")

	// NewTree should succeed - it doesn't validate YAML content
	tree, err := NewTree(tmpDir)
	assertNoError(t, err)

	// yaml.Marshal should fail when trying to marshal the tree with invalid YAML
	_, err = yaml.Marshal(tree)
	assertErrorContains(t, err, anotherDirFile)
	// Verify the error is a YAML/JSON parsing error
	errStr := strings.ToLower(err.Error())
	if !strings.Contains(errStr, "yaml") && !strings.Contains(errStr, "json") {
		t.Errorf("yaml.Marshal() error = %v, expected YAML/JSON parsing error", err)
	}
	// Verify the error includes position information (line:column) if available
	if strings.Contains(err.Error(), "YAML/JSON syntax error in") {
		if !strings.Contains(err.Error(), ":") {
			t.Error("yaml.Marshal() error should include position information (line:column)")
		}
	}
}

func TestFormatYAMLError_ParserError(t *testing.T) {
	// Test that formatYAMLError properly formats ParserError with position info
	tmpDir := createTestDir(t, map[string]string{
		"test.yml": "key: [unclosed",
	}, nil)
	testFile := filepath.Join(tmpDir, "test.yml")

	tree, err := NewTree(tmpDir)
	assertNoError(t, err)

	_, err = yaml.Marshal(tree)
	assertErrorContains(t, err, "YAML/JSON syntax error in")
	// Verify error includes file path
	if !strings.Contains(err.Error(), testFile) {
		t.Errorf("Expected file path in error message, got: %s", err.Error())
	}
	// Should have line:column format
	if !strings.Contains(err.Error(), ":") {
		t.Errorf("Expected position information (line:column) in error message, got: %s", err.Error())
	}
}

func TestFormatYAMLError_TypeError(t *testing.T) {
	// Test that formatYAMLError properly formats TypeError with position info
	// Create YAML that will cause a type error (e.g., trying to use a string as a number)
	tmpDir := createTestDir(t, map[string]string{
		"test.yml": "key: not_a_number\nnumber: !!int not_a_number",
	}, nil)
	testFile := filepath.Join(tmpDir, "test.yml")

	tree, err := NewTree(tmpDir)
	assertNoError(t, err)

	_, err = yaml.Marshal(tree)
	// TypeError should be formatted with file path and error details
	assertErrorContains(t, err, testFile)
	// Should contain type error indication
	if !strings.Contains(err.Error(), "type") && !strings.Contains(err.Error(), "YAML/JSON") {
		t.Errorf("Expected type error indication in error message, got: %s", err.Error())
	}
}

func TestNewTree_JSONFiles(t *testing.T) {
	// Test that JSON files are recognized and processed
	tmpDir := createTestDir(t, map[string]string{
		"entities/item1.json": `{"entity": {"id": "example1", "attributes": {"name": "sample name"}}}`,
	}, nil)

	tree, err := NewTree(tmpDir)
	assertNoError(t, err)

	if tree == nil {
		t.Fatal("NewTree() returned nil tree")
	}

	// Verify JSON file is included in the tree
	resultMap := createTreeAndMarshal(t, tmpDir)

	entitiesMap, ok := resultMap["entities"].(map[string]interface{})
	if !ok {
		t.Fatal("MarshalYAML() result missing 'entities' key or not a map")
	}

	item1Map, ok := entitiesMap["item1"].(map[string]interface{})
	if !ok {
		t.Fatal("MarshalYAML() result missing 'item1' key from JSON file or not a map")
	}

	// Verify JSON content was parsed correctly
	entityMap, ok := item1Map["entity"].(map[string]interface{})
	if !ok {
		t.Fatal("MarshalYAML() result missing 'entity' key from JSON file or not a map")
	}

	if entityMap["id"] != "example1" {
		t.Errorf("MarshalYAML() JSON content id not parsed correctly. Got: %v", entityMap["id"])
	}

	attributesMap, ok := entityMap["attributes"].(map[string]interface{})
	if !ok {
		t.Fatal("MarshalYAML() result missing 'attributes' key from JSON file or not a map")
	}

	if attributesMap["name"] != "sample name" {
		t.Errorf("MarshalYAML() JSON content name not parsed correctly. Got: %v", attributesMap["name"])
	}
}

func TestNewTree_JSONSpecialCase(t *testing.T) {
	// Test that @*.json files are recognized as special case files
	tmpDir := createTestDir(t, map[string]string{
		"entities/@common.json": `{"timeout": 30, "retries": 3}`,
		"entities/item1.yml":    "entity:\n  id: example1\n  attributes:\n    name: sample name",
	}, nil)

	resultMap := createTreeAndMarshal(t, tmpDir)

	entitiesMap, ok := resultMap["entities"].(map[string]interface{})
	if !ok {
		t.Fatal("MarshalYAML() result missing 'entities' key or not a map")
	}

	// @common.json should merge into entities map
	// JSON numbers may be parsed as int or float64 depending on the YAML library
	timeout := entitiesMap["timeout"]
	if timeout != 30 && timeout != float64(30) {
		t.Errorf("MarshalYAML() @common.json content not merged into entities map. Got timeout: %v (type: %T)", timeout, timeout)
	}
	// item1.yml should also be present
	if entitiesMap["item1"] == nil {
		t.Error("MarshalYAML() item1.yml not found in entities map")
	}
}

func TestSpecialCaseDirectory(t *testing.T) {
	// Test that specialCaseDirectory() correctly identifies @ directories
	tmpDir := createTestDir(t, nil, []string{"@group1", "entities"})

	tree, err := NewTree(tmpDir)
	assertNoError(t, err)

	// Find the @group1 node
	atDirNode := findNodeByName(t, tree, "@group1")
	if atDirNode == nil {
		t.Fatal("Could not find @group1 node")
	}

	// Find the entities node
	regularDirNode := findNodeByName(t, tree, "entities")
	if regularDirNode == nil {
		t.Fatal("Could not find entities node")
	}

	// @group1 should be identified as special case directory
	if !atDirNode.specialCaseDirectory() {
		t.Error("specialCaseDirectory() returned false for @group1 directory")
	}

	// entities should not be identified as special case directory
	if regularDirNode.specialCaseDirectory() {
		t.Error("specialCaseDirectory() returned true for regular entities directory")
	}
}

func TestSpecialCaseFile(t *testing.T) {
	// Test that specialCase() correctly identifies @*.yml files (not directories)
	tmpDir := createTestDir(t, map[string]string{
		"entities/@common.yml": "timeout: 30",
		"entities/item1.yml":   "entity:\n  id: example1",
	}, nil)

	tree, err := NewTree(tmpDir)
	assertNoError(t, err)

	// Find the @common.yml node
	atFileNode := findNodeByName(t, tree, "@common.yml")
	if atFileNode == nil {
		t.Fatal("Could not find @common.yml node")
	}

	// Find a regular file node
	regularFileNode := findNodeByName(t, tree, "item1.yml")
	if regularFileNode == nil {
		t.Fatal("Could not find item1.yml node")
	}

	// @common.yml should be identified as special case file
	if !atFileNode.specialCase() {
		t.Error("specialCase() returned false for @common.yml file")
	}

	// item1.yml should not be identified as special case file
	if regularFileNode.specialCase() {
		t.Error("specialCase() returned true for regular item1.yml file")
	}

	// @common.yml should not be identified as special case directory
	if atFileNode.specialCaseDirectory() {
		t.Error("specialCaseDirectory() returned true for @common.yml file (should only be true for directories)")
	}
}

func TestMarshalYAML_AtDirectory(t *testing.T) {
	// Test that @ directories merge their contents into parent map
	tmpDir := createTestDir(t, map[string]string{
		"entities/@group1/item1.yml": "entity:\n  id: example1\n  attributes:\n    name: first item",
		"entities/@group1/item2.yml": "entity:\n  id: example2\n  attributes:\n    name: second item",
		"entities/@group2/item3.yml": "entity:\n  id: example3\n  attributes:\n    name: third item",
		"entities/item4.yml":         "entity:\n  id: example4\n  attributes:\n    name: fourth item",
	}, nil)

	resultMap := createTreeAndMarshal(t, tmpDir)

	entitiesMap, ok := resultMap["entities"].(map[string]interface{})
	if !ok {
		t.Fatal("MarshalYAML() result missing 'entities' key or not a map")
	}

	// Files from @group1 should be in entities map (not nested under @group1)
	if entitiesMap["item1"] == nil {
		t.Error("MarshalYAML() item1.yml from @group1 not found in entities map")
	}
	if entitiesMap["item2"] == nil {
		t.Error("MarshalYAML() item2.yml from @group1 not found in entities map")
	}

	// Files from @group2 should be in entities map (not nested under @group2)
	if entitiesMap["item3"] == nil {
		t.Error("MarshalYAML() item3.yml from @group2 not found in entities map")
	}

	// Regular file should also be present
	if entitiesMap["item4"] == nil {
		t.Error("MarshalYAML() item4.yml not found in entities map")
	}

	// @group1 and @group2 should NOT be keys in entities map
	if entitiesMap["@group1"] != nil {
		t.Error("MarshalYAML() @group1 should not be a key in entities map")
	}
	if entitiesMap["@group2"] != nil {
		t.Error("MarshalYAML() @group2 should not be a key in entities map")
	}
	if entitiesMap["group1"] != nil {
		t.Error("MarshalYAML() group1 (without @) should not be a key in entities map")
	}
	if entitiesMap["group2"] != nil {
		t.Error("MarshalYAML() group2 (without @) should not be a key in entities map")
	}
}

func TestMarshalYAML_EmptyAtDirectory(t *testing.T) {
	// Test that empty @ directories are ignored
	tmpDir := createTestDir(t, map[string]string{
		"entities/item1.yml": "entity:\n  id: example1\n  attributes:\n    name: sample name",
	}, []string{"entities/@group1"})

	resultMap := createTreeAndMarshal(t, tmpDir)

	entitiesMap, ok := resultMap["entities"].(map[string]interface{})
	if !ok {
		t.Fatal("MarshalYAML() result missing 'entities' key or not a map")
	}

	// Empty @ directory should not create any keys
	if entitiesMap["@group1"] != nil {
		t.Error("MarshalYAML() empty @group1 directory should not create a key")
	}
	if entitiesMap["group1"] != nil {
		t.Error("MarshalYAML() empty @group1 directory should not create a 'group1' key")
	}

	// Regular file should still be present
	if entitiesMap["item1"] == nil {
		t.Error("MarshalYAML() item1.yml not found in entities map")
	}
}

func TestMarshalYAML_NestedAtDirectories(t *testing.T) {
	// Test that nested @ directories work recursively
	// @group1/@group2/ should merge into parent of @group1/
	tmpDir := createTestDir(t, map[string]string{
		"entities/@group1/@group2/item1.yml": "entity:\n  id: example1\n  attributes:\n    name: first item",
		"entities/@group1/@group2/item2.yml": "entity:\n  id: example2\n  attributes:\n    name: second item",
		"entities/@group1/item3.yml":         "entity:\n  id: example3\n  attributes:\n    name: third item",
	}, nil)

	resultMap := createTreeAndMarshal(t, tmpDir)

	entitiesMap, ok := resultMap["entities"].(map[string]interface{})
	if !ok {
		t.Fatal("MarshalYAML() result missing 'entities' key or not a map")
	}

	// Files from nested @group2 should be in entities map (merged through @group1)
	if entitiesMap["item1"] == nil {
		t.Error("MarshalYAML() item1.yml from @group1/@group2 not found in entities map")
	}
	if entitiesMap["item2"] == nil {
		t.Error("MarshalYAML() item2.yml from @group1/@group2 not found in entities map")
	}

	// File from @group1 should also be in entities map
	if entitiesMap["item3"] == nil {
		t.Error("MarshalYAML() item3.yml from @group1 not found in entities map")
	}

	// @group1 and @group2 should NOT be keys in entities map
	if entitiesMap["@group1"] != nil {
		t.Error("MarshalYAML() @group1 should not be a key in entities map")
	}
	if entitiesMap["@group2"] != nil {
		t.Error("MarshalYAML() @group2 should not be a key in entities map")
	}
	if entitiesMap["group1"] != nil {
		t.Error("MarshalYAML() group1 (without @) should not be a key in entities map")
	}
	if entitiesMap["group2"] != nil {
		t.Error("MarshalYAML() group2 (without @) should not be a key in entities map")
	}
}

func TestMarshalLeaf_WithIncludes(t *testing.T) {
	// Test marshalLeaf with EnableIncludes enabled
	tmpDir := createTestDir(t, map[string]string{
		"entities/scripts/test.sh": "#!/bin/bash\necho 'test'",
		"entities/item1.yml":       "entity:\n  id: example1\n  attributes:\n    command: <<include(scripts/test.sh)>>",
	}, nil)

	absDir, err := filepath.Abs(tmpDir)
	assertNoError(t, err)

	opts := &Options{
		EnableIncludes: true,
		PackRoot:       absDir,
	}

	tree, err := NewTree(tmpDir)
	assertNoError(t, err)

	// Find the item1.yml node
	testNode := findNodeByName(t, tree, "item1.yml")
	if testNode == nil {
		t.Fatal("Could not find item1.yml node")
	}

	// Test with includes enabled
	result, err := testNode.marshalLeaf(opts)
	assertNoError(t, err)

	resultMap, ok := result.(map[string]interface{})
	if !ok {
		t.Fatalf("marshalLeaf() returned %T, want map[string]interface{}", result)
	}

	// The include should have been processed
	entityMap, ok := resultMap["entity"].(map[string]interface{})
	if !ok {
		t.Fatalf("marshalLeaf() entity value is %T, want map[string]interface{}", resultMap["entity"])
	}

	attributesMap, ok := entityMap["attributes"].(map[string]interface{})
	if !ok {
		t.Fatalf("marshalLeaf() attributes value is %T, want map[string]interface{}", entityMap["attributes"])
	}

	commandVal, ok := attributesMap["command"].(string)
	if !ok {
		t.Fatalf("marshalLeaf() command value is %T, want string", attributesMap["command"])
	}

	if !strings.Contains(commandVal, "echo 'test'") {
		t.Errorf("marshalLeaf() should contain included content. Got: %q", commandVal)
	}
	if strings.Contains(commandVal, "<<include") {
		t.Error("marshalLeaf() should not contain include directive after processing")
	}
}

func TestMarshalLeaf_WithIncludes_ErrorCases(t *testing.T) {
	tmpDir := createTestDir(t, map[string]string{
		"test.yml": `command: <<include(nonexistent.sh)>>`,
	}, nil)

	absDir, err := filepath.Abs(tmpDir)
	assertNoError(t, err)

	opts := &Options{
		EnableIncludes: true,
		PackRoot:       absDir,
	}

	tree, err := NewTree(tmpDir)
	assertNoError(t, err)

	testNode := findNodeByName(t, tree, "test.yml")
	if testNode == nil {
		t.Fatal("Could not find test.yml node")
	}

	_, err = testNode.marshalLeaf(opts)
	assertErrorContains(t, err, "could not open")
}

func TestMarshalLeaf_FileReadError(t *testing.T) {
	// Test error handling when file cannot be read
	tmpDir := createTestDir(t, map[string]string{
		"test.yml": "key: value",
	}, nil)
	yamlFile := filepath.Join(tmpDir, "test.yml")

	tree, err := NewTree(tmpDir)
	assertNoError(t, err)

	testNode := findNodeByName(t, tree, "test.yml")
	if testNode == nil {
		t.Fatal("Could not find test.yml node")
	}

	// Remove read permission to trigger error
	if err := os.Chmod(yamlFile, 0000); err != nil {
		t.Fatalf("Failed to chmod file: %v", err)
	}
	t.Cleanup(func() {
		_ = os.Chmod(yamlFile, 0600) // Restore for cleanup
	})

	_, err = testNode.marshalLeaf(nil)
	assertErrorContains(t, err, "failed to read file")
}

func TestIsEmptyContent(t *testing.T) {
	tests := []struct {
		name     string
		input    interface{}
		expected bool
	}{
		{"nil", nil, true},
		{"empty map[string]interface{}", map[string]interface{}{}, true},
		{"empty map[interface{}]interface{}", map[interface{}]interface{}{}, true},
		{"non-empty map[string]interface{}", map[string]interface{}{"key": "value"}, false},
		{"non-empty map[interface{}]interface{}", map[interface{}]interface{}{"key": "value"}, false},
		{"string", "not empty", false},
		{"int", 42, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsEmptyContent(tt.input)
			if result != tt.expected {
				t.Errorf("IsEmptyContent(%v) = %v, want %v", tt.input, result, tt.expected)
			}
		})
	}
}

func TestNormalizeYAML11Booleans(t *testing.T) {
	// Test that normalizeYAML11Booleans correctly converts YAML 1.1 booleans
	tests := []struct {
		name     string
		input    string
		expected string
		wantBool bool // true if the value should become a bool, false if string
	}{
		// Unquoted YAML 1.1 booleans should be normalized
		{"unquoted on", "value: on", "true", true},
		{"unquoted off", "value: off", "false", true},
		{"unquoted yes", "value: yes", "true", true},
		{"unquoted no", "value: no", "false", true},
		{"unquoted y", "value: y", "true", true},
		{"unquoted n", "value: n", "false", true},
		{"unquoted Y", "value: Y", "true", true},
		{"unquoted N", "value: N", "false", true},
		{"unquoted Yes", "value: Yes", "true", true},
		{"unquoted No", "value: No", "false", true},
		{"unquoted ON", "value: ON", "true", true},
		{"unquoted OFF", "value: OFF", "false", true},
		{"unquoted On", "value: On", "true", true},
		{"unquoted Off", "value: Off", "false", true},
		{"unquoted YES", "value: YES", "true", true},
		{"unquoted NO", "value: NO", "false", true},
		// Quoted strings should NOT be normalized
		{"double-quoted on", `value: "on"`, "on", false},
		{"double-quoted yes", `value: "yes"`, "yes", false},
		{"single-quoted on", "value: 'on'", "on", false},
		{"single-quoted yes", "value: 'yes'", "yes", false},
		// Already canonical booleans should remain unchanged
		{"unquoted true", "value: true", "true", true},
		{"unquoted false", "value: false", "false", true},
		// Non-boolean strings should not be affected
		{"unquoted norway", "value: norway", "norway", false},
		{"unquoted yesss", "value: yesss", "yesss", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var node yaml.Node
			if err := yaml.Unmarshal([]byte(tt.input), &node); err != nil {
				t.Fatalf("Failed to parse YAML: %v", err)
			}

			normalizeYAML11Booleans(&node)

			var result map[string]interface{}
			if err := node.Decode(&result); err != nil {
				t.Fatalf("Failed to decode node: %v", err)
			}

			val := result["value"]
			if tt.wantBool {
				boolVal, ok := val.(bool)
				if !ok {
					t.Errorf("Expected bool, got %T: %v", val, val)
					return
				}
				expectedBool := tt.expected == "true"
				if boolVal != expectedBool {
					t.Errorf("Got %v, want %v", boolVal, expectedBool)
				}
			} else {
				strVal, ok := val.(string)
				if !ok {
					t.Errorf("Expected string, got %T: %v", val, val)
					return
				}
				if strVal != tt.expected {
					t.Errorf("Got %q, want %q", strVal, tt.expected)
				}
			}
		})
	}
}

func TestNormalizeYAML11Booleans_Nested(t *testing.T) {
	// Test that normalization works recursively in nested structures
	input := `
config:
  enabled: on
  debug: off
  settings:
    verbose: yes
    quiet: "no"
  items:
    - active: on
    - active: "off"
`
	var node yaml.Node
	if err := yaml.Unmarshal([]byte(input), &node); err != nil {
		t.Fatalf("Failed to parse YAML: %v", err)
	}

	normalizeYAML11Booleans(&node)

	var result map[string]interface{}
	if err := node.Decode(&result); err != nil {
		t.Fatalf("Failed to decode node: %v", err)
	}

	config := result["config"].(map[string]interface{})

	// Unquoted should be normalized
	if enabled, ok := config["enabled"].(bool); !ok || !enabled {
		t.Errorf("enabled should be bool true, got %T: %v", config["enabled"], config["enabled"])
	}
	if debug, ok := config["debug"].(bool); !ok || debug {
		t.Errorf("debug should be bool false, got %T: %v", config["debug"], config["debug"])
	}

	settings := config["settings"].(map[string]interface{})
	if verbose, ok := settings["verbose"].(bool); !ok || !verbose {
		t.Errorf("verbose should be bool true, got %T: %v", settings["verbose"], settings["verbose"])
	}
	// Quoted should remain string
	if quiet, ok := settings["quiet"].(string); !ok || quiet != "no" {
		t.Errorf("quiet should be string 'no', got %T: %v", settings["quiet"], settings["quiet"])
	}

	items := config["items"].([]interface{})
	item0 := items[0].(map[string]interface{})
	if active, ok := item0["active"].(bool); !ok || !active {
		t.Errorf("items[0].active should be bool true, got %T: %v", item0["active"], item0["active"])
	}
	item1 := items[1].(map[string]interface{})
	// Quoted should remain string
	if active, ok := item1["active"].(string); !ok || active != "off" {
		t.Errorf("items[1].active should be string 'off', got %T: %v", item1["active"], item1["active"])
	}
}

func TestMarshalLeaf_WithConvertBooleans(t *testing.T) {
	// Test that the ConvertBooleans option works in marshalLeaf
	tmpDir := createTestDir(t, map[string]string{
		"config/test.yml": `enabled: on
disabled: "off"
name: on_call_service`,
	}, nil)

	absDir, err := filepath.Abs(tmpDir)
	assertNoError(t, err)

	// Test with conversion enabled
	opts := &Options{
		EnableIncludes:  false,
		PackRoot:        absDir,
		ConvertBooleans: true,
	}

	tree, err := NewTree(tmpDir)
	assertNoError(t, err)

	result, err := tree.Marshal(opts)
	assertNoError(t, err)

	resultMap, ok := result.(map[string]interface{})
	if !ok {
		t.Fatalf("Expected map[string]interface{}, got %T", result)
	}
	configMap, ok := resultMap["config"].(map[string]interface{})
	if !ok {
		t.Fatalf("Expected config to be map[string]interface{}, got %T", resultMap["config"])
	}
	testMap, ok := configMap["test"].(map[string]interface{})
	if !ok {
		t.Fatalf("Expected test to be map[string]interface{}, got %T", configMap["test"])
	}

	// Unquoted 'on' should be normalized to bool true
	if enabled, ok := testMap["enabled"].(bool); !ok || !enabled {
		t.Errorf("enabled should be bool true, got %T: %v", testMap["enabled"], testMap["enabled"])
	}

	// Quoted 'off' should remain string
	if disabled, ok := testMap["disabled"].(string); !ok || disabled != "off" {
		t.Errorf("disabled should be string 'off', got %T: %v", testMap["disabled"], testMap["disabled"])
	}

	// Non-boolean string should remain unchanged
	if name, ok := testMap["name"].(string); !ok || name != "on_call_service" {
		t.Errorf("name should be string 'on_call_service', got %T: %v", testMap["name"], testMap["name"])
	}
}

func TestMarshalLeaf_WithoutConvertBooleans(t *testing.T) {
	// Test that without the option, YAML 1.1 booleans remain as strings
	tmpDir := createTestDir(t, map[string]string{
		"config/test.yml": `enabled: on
disabled: off`,
	}, nil)

	absDir, err := filepath.Abs(tmpDir)
	assertNoError(t, err)

	// Test WITHOUT conversion
	opts := &Options{
		EnableIncludes:  false,
		PackRoot:        absDir,
		ConvertBooleans: false,
	}

	tree, err := NewTree(tmpDir)
	assertNoError(t, err)

	result, err := tree.Marshal(opts)
	assertNoError(t, err)

	resultMap, ok := result.(map[string]interface{})
	if !ok {
		t.Fatalf("Expected map[string]interface{}, got %T", result)
	}
	configMap, ok := resultMap["config"].(map[string]interface{})
	if !ok {
		t.Fatalf("Expected config to be map[string]interface{}, got %T", resultMap["config"])
	}
	testMap, ok := configMap["test"].(map[string]interface{})
	if !ok {
		t.Fatalf("Expected test to be map[string]interface{}, got %T", configMap["test"])
	}

	// Without normalization, 'on' should remain string (YAML 1.2 behavior with interface{})
	if enabled, ok := testMap["enabled"].(string); !ok || enabled != "on" {
		t.Errorf("enabled should be string 'on' without normalization, got %T: %v", testMap["enabled"], testMap["enabled"])
	}
}

func TestMarshalParent_WithEmptyMaps(t *testing.T) {
	// Test that empty maps are properly skipped
	tmpDir := createTestDir(t, map[string]string{
		"has_content/file.yml": "key: value",
	}, []string{"empty1", "empty2"})

	resultMap := createTreeAndMarshal(t, tmpDir)

	// Empty directories should not appear
	if _, ok := resultMap["empty1"]; ok {
		t.Error("MarshalYAML() should not contain 'empty1' key")
	}
	if _, ok := resultMap["empty2"]; ok {
		t.Error("MarshalYAML() should not contain 'empty2' key")
	}

	// Directory with content should appear
	if _, ok := resultMap["has_content"]; !ok {
		t.Error("MarshalYAML() should contain 'has_content' key")
	}
}

func TestMarshalParent_NonMapTypeError(t *testing.T) {
	// Test that marshalParent returns an error when a child returns a non-map type
	// This happens when a YAML file contains a scalar value instead of a map
	tmpDir := createTestDir(t, map[string]string{
		"dir/scalar.yml": "just a string",
		"dir/map.yml":    "key: value",
	}, nil)

	tree, err := NewTree(tmpDir)
	assertNoError(t, err)

	// Find the directory node
	dirNode := findNodeByName(t, tree, "dir")
	if dirNode == nil {
		t.Fatal("Could not find dir node")
	}

	// Marshal should fail because scalar.yml returns a string, not a map
	_, err = dirNode.Marshal(nil)
	assertErrorContains(t, err, "expected a map")
	assertErrorContains(t, err, "scalar.yml")
}

func TestRootFile(t *testing.T) {
	// Test that rootFile() correctly identifies files at the root level
	tmpDir := createTestDir(t, map[string]string{
		"root_file.yml":   "key1: value1",
		"subdir/file.yml": "key2: value2",
	}, nil)

	tree, err := NewTree(tmpDir)
	assertNoError(t, err)

	// Find the root file node
	rootFileNode := findNodeByName(t, tree, "root_file.yml")
	if rootFileNode == nil {
		t.Fatal("Could not find root_file.yml node")
	}

	// Find a file in a subdirectory
	subFileNode := findNodeByName(t, tree, "file.yml")
	if subFileNode == nil {
		t.Fatal("Could not find file.yml node")
	}

	// root_file.yml should be identified as a root file
	if !rootFileNode.rootFile() {
		t.Error("rootFile() returned false for root_file.yml")
	}

	// file.yml in subdirectory should not be identified as a root file
	if subFileNode.rootFile() {
		t.Error("rootFile() returned true for file.yml in subdirectory")
	}

	// Verify root files are merged correctly (not nested under a key)
	resultMap := createTreeAndMarshal(t, tmpDir)
	// Root file content should be merged at the top level
	if resultMap["key1"] == nil {
		t.Error("Root file content should be merged at top level, not nested")
	}
	if resultMap["key1"] != "value1" {
		t.Errorf("Root file content key1 = %v, want 'value1'", resultMap["key1"])
	}
	// Subdirectory should still be nested
	subdirMap, ok := resultMap["subdir"].(map[string]interface{})
	if !ok {
		t.Fatal("Subdirectory should be nested under 'subdir' key")
	}
	// File in subdirectory is nested under its name (without extension)
	fileMap, ok := subdirMap["file"].(map[string]interface{})
	if !ok {
		t.Fatal("File should be nested under 'file' key in subdirectory")
	}
	if fileMap["key2"] != "value2" {
		t.Errorf("Subdirectory file content key2 = %v, want 'value2'", fileMap["key2"])
	}
}

func TestOptions_Log(t *testing.T) {
	// Test that Options.log() returns appropriate logger in different scenarios
	t.Run("nil options", func(t *testing.T) {
		log := (*Options)(nil).log()
		if log == nil {
			t.Error("Options.log() should return a logger, not nil")
		}
		// Should be a NoOpLogger
		log.Debugf("test") // Should not panic
		log.Warnf("test")  // Should not panic
	})

	t.Run("nil logger", func(t *testing.T) {
		opts := &Options{
			EnableIncludes:  false,
			PackRoot:        "/tmp",
			ConvertBooleans: false,
			Logger:          nil,
		}
		log := opts.log()
		if log == nil {
			t.Error("Options.log() should return a logger, not nil")
		}
		// Should be a NoOpLogger
		log.Debugf("test") // Should not panic
		log.Warnf("test")  // Should not panic
	})

	t.Run("with logger", func(t *testing.T) {
		var buf strings.Builder
		testLogger := logger.New(&buf, true)
		opts := &Options{
			EnableIncludes:  false,
			PackRoot:        "/tmp",
			ConvertBooleans: false,
			Logger:          testLogger,
		}
		log := opts.log()
		if log != testLogger {
			t.Error("Options.log() should return the configured logger")
		}
		log.Debugf("test message")
		if !strings.Contains(buf.String(), "test message") {
			t.Error("Options.log() should return the configured logger that actually logs")
		}
	})
}

func TestMergeTree_InvalidInput(t *testing.T) {
	// Test that mergeTree returns an error for types that mapstructure.Decode can't handle
	tests := []struct {
		name        string
		input       interface{}
		expectError bool
		errorSubstr string
		validate    func(t *testing.T, result map[string]interface{})
	}{
		{
			name:        "channel",
			input:       make(chan int),
			expectError: true,
			errorSubstr: "failed to decode tree structure",
		},
		{
			name:        "function",
			input:       func() {},
			expectError: true,
			errorSubstr: "failed to decode tree structure",
		},
		{
			name:        "nil",
			input:       nil,
			expectError: false,
			validate: func(t *testing.T, result map[string]interface{}) {
				if result == nil {
					t.Error("mergeTree() returned nil result for nil input")
				}
				if len(result) != 0 {
					t.Errorf("mergeTree() result should be empty for nil input, got %v", result)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := mergeTree(tt.input)
			if tt.expectError {
				assertErrorContains(t, err, tt.errorSubstr)
			} else {
				assertNoError(t, err)
				if tt.validate != nil {
					tt.validate(t, result)
				}
			}
		})
	}
}
