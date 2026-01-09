package filetree

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// Test helpers - shared across test files

// asMapShared converts interface{} to map[interface{}]interface{} for test assertions.
// Handles both map[string]interface{} and map[interface{}]interface{}.
func asMapShared(t *testing.T, v interface{}) map[interface{}]interface{} {
	t.Helper()
	switch m := v.(type) {
	case map[interface{}]interface{}:
		return m
	case map[string]interface{}:
		result := make(map[interface{}]interface{}, len(m))
		for k, val := range m {
			result[k] = val
		}
		return result
	default:
		t.Fatalf("Expected map, got %T", v)
		return nil
	}
}

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
// and returns the result as a map[interface{}]interface{}.
func createTreeAndMarshal(t *testing.T, dir string) map[interface{}]interface{} {
	t.Helper()
	tree, err := NewTree(dir)
	assertNoError(t, err)
	result, err := tree.MarshalYAML()
	assertNoError(t, err)
	resultMap, ok := result.(map[interface{}]interface{})
	if !ok {
		t.Fatalf("MarshalYAML() returned %T, want map[interface{}]interface{}", result)
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

// Tree building tests

func TestNewTree(t *testing.T) {
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

func TestNewTree_NonexistentDirectory(t *testing.T) {
	_, err := NewTree("/nonexistent/path/that/does/not/exist")
	assertErrorContains(t, err, "no such file")
}

func TestNewTree_JSONFiles(t *testing.T) {
	tmpDir := createTestDir(t, map[string]string{
		"entities/item1.json": `{"entity": {"id": "example1", "attributes": {"name": "sample name"}}}`,
	}, nil)

	tree, err := NewTree(tmpDir)
	assertNoError(t, err)

	if tree == nil {
		t.Fatal("NewTree() returned nil tree")
	}

	resultMap := createTreeAndMarshal(t, tmpDir)

	entitiesMap := asMapShared(t, resultMap["entities"])
	item1Map := asMapShared(t, entitiesMap["item1"])
	entityMap := asMapShared(t, item1Map["entity"])

	if entityMap["id"] != "example1" {
		t.Errorf("MarshalYAML() JSON content id not parsed correctly. Got: %v", entityMap["id"])
	}

	attributesMap := asMapShared(t, entityMap["attributes"])

	if attributesMap["name"] != "sample name" {
		t.Errorf("MarshalYAML() JSON content name not parsed correctly. Got: %v", attributesMap["name"])
	}
}

func TestNewTree_JSONSpecialCase(t *testing.T) {
	tmpDir := createTestDir(t, map[string]string{
		"entities/@common.json": `{"timeout": 30, "retries": 3}`,
		"entities/item1.yml":    "entity:\n  id: example1\n  attributes:\n    name: sample name",
	}, nil)

	resultMap := createTreeAndMarshal(t, tmpDir)
	entitiesMap := asMapShared(t, resultMap["entities"])

	timeout := entitiesMap["timeout"]
	if timeout != 30 && timeout != float64(30) {
		t.Errorf("MarshalYAML() @common.json content not merged into entities map. Got timeout: %v (type: %T)", timeout, timeout)
	}
	if entitiesMap["item1"] == nil {
		t.Error("MarshalYAML() item1.yml not found in entities map")
	}
}

func TestSpecialCaseDirectory(t *testing.T) {
	tmpDir := createTestDir(t, nil, []string{"@group1", "entities"})

	tree, err := NewTree(tmpDir)
	assertNoError(t, err)

	atDirNode := findNodeByName(t, tree, "@group1")
	if atDirNode == nil {
		t.Fatal("Could not find @group1 node")
	}

	regularDirNode := findNodeByName(t, tree, "entities")
	if regularDirNode == nil {
		t.Fatal("Could not find entities node")
	}

	if !atDirNode.specialCaseDirectory() {
		t.Error("specialCaseDirectory() returned false for @group1 directory")
	}

	if regularDirNode.specialCaseDirectory() {
		t.Error("specialCaseDirectory() returned true for regular entities directory")
	}
}

func TestSpecialCaseFile(t *testing.T) {
	tmpDir := createTestDir(t, map[string]string{
		"entities/@common.yml": "timeout: 30",
		"entities/item1.yml":   "entity:\n  id: example1",
	}, nil)

	tree, err := NewTree(tmpDir)
	assertNoError(t, err)

	atFileNode := findNodeByName(t, tree, "@common.yml")
	if atFileNode == nil {
		t.Fatal("Could not find @common.yml node")
	}

	regularFileNode := findNodeByName(t, tree, "item1.yml")
	if regularFileNode == nil {
		t.Fatal("Could not find item1.yml node")
	}

	if !atFileNode.specialCase() {
		t.Error("specialCase() returned false for @common.yml file")
	}

	if regularFileNode.specialCase() {
		t.Error("specialCase() returned true for regular item1.yml file")
	}

	if atFileNode.specialCaseDirectory() {
		t.Error("specialCaseDirectory() returned true for @common.yml file (should only be true for directories)")
	}
}
