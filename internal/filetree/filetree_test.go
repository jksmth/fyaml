package filetree

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"gopkg.in/yaml.v3"
)

func TestNewTree(t *testing.T) {
	// Test basic tree building with a temporary directory
	// Integration tests in cmd/fyaml/main_test.go cover full pack() behavior
	tmpDir := t.TempDir()

	subDir := filepath.Join(tmpDir, "sub_dir")
	subDirFile := filepath.Join(tmpDir, "sub_dir", "sub_dir_file.yml")
	emptyDir := filepath.Join(tmpDir, "empty_dir")

	if err := os.Mkdir(subDir, 0700); err != nil {
		t.Fatalf("Failed to create subdirectory: %v", err)
	}
	if err := os.WriteFile(subDirFile, []byte("foo:\n  bar:\n    baz"), 0600); err != nil {
		t.Fatalf("Failed to create file: %v", err)
	}
	if err := os.Mkdir(emptyDir, 0700); err != nil {
		t.Fatalf("Failed to create empty directory: %v", err)
	}

	tree, err := NewTree(tmpDir)
	if err != nil {
		t.Fatalf("NewTree() error = %v", err)
	}

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
	tmpDir := t.TempDir()

	subDir := filepath.Join(tmpDir, "sub_dir")
	subDirFile := filepath.Join(tmpDir, "sub_dir", "sub_dir_file.yml")
	emptyDir := filepath.Join(tmpDir, "empty_dir")

	if err := os.Mkdir(subDir, 0700); err != nil {
		t.Fatalf("Failed to create subdirectory: %v", err)
	}
	if err := os.WriteFile(subDirFile, []byte("foo:\n  bar:\n    baz"), 0600); err != nil {
		t.Fatalf("Failed to create file: %v", err)
	}
	if err := os.Mkdir(emptyDir, 0700); err != nil {
		t.Fatalf("Failed to create empty directory: %v", err)
	}

	tree, err := NewTree(tmpDir)
	if err != nil {
		t.Fatalf("NewTree() error = %v", err)
	}

	out, err := yaml.Marshal(tree)
	if err != nil {
		t.Fatalf("yaml.Marshal() error = %v", err)
	}

	if len(out) == 0 {
		t.Error("yaml.Marshal() returned empty output")
	}

	// Verify it contains expected keys (order may vary due to sorting)
	result, err := tree.MarshalYAML()
	if err != nil {
		t.Fatalf("MarshalYAML() error = %v", err)
	}

	resultMap, ok := result.(map[string]interface{})
	if !ok {
		t.Fatalf("MarshalYAML() returned %T, want map[string]interface{}", result)
	}

	if _, ok := resultMap["sub_dir"]; !ok {
		t.Error("MarshalYAML() result missing 'sub_dir' key")
	}
	if _, ok := resultMap["empty_dir"]; !ok {
		t.Error("MarshalYAML() result missing 'empty_dir' key")
	}
}

func TestMarshalYAML_InvalidYAML(t *testing.T) {
	// Test that invalid YAML content causes an error when marshaling
	// This matches the original CircleCI test behavior
	tmpDir := t.TempDir()

	anotherDir := filepath.Join(tmpDir, "another_dir")
	anotherDirFile := filepath.Join(tmpDir, "another_dir", "another_dir_file.yml")

	if err := os.Mkdir(anotherDir, 0700); err != nil {
		t.Fatalf("Failed to create directory: %v", err)
	}
	if err := os.WriteFile(anotherDirFile, []byte("1some: in: valid: yaml"), 0600); err != nil {
		t.Fatalf("Failed to create invalid YAML file: %v", err)
	}

	// NewTree should succeed - it doesn't validate YAML content
	tree, err := NewTree(tmpDir)
	if err != nil {
		t.Fatalf("NewTree() error = %v, expected no error", err)
	}

	// yaml.Marshal should fail when trying to marshal the tree with invalid YAML
	_, err = yaml.Marshal(tree)
	if err == nil {
		t.Error("yaml.Marshal() expected error for invalid YAML content, got nil")
		return
	}

	// Verify the error message indicates a YAML parsing issue
	// The exact message may vary by YAML library version, but should contain "yaml"
	if !strings.Contains(err.Error(), "yaml") {
		t.Errorf("yaml.Marshal() error = %v, expected YAML parsing error", err)
	}
}

func TestNewTree_JSONFiles(t *testing.T) {
	// Test that JSON files are recognized and processed
	tmpDir := t.TempDir()

	servicesDir := filepath.Join(tmpDir, "services")
	jsonFile := filepath.Join(servicesDir, "api.json")

	if err := os.Mkdir(servicesDir, 0700); err != nil {
		t.Fatalf("Failed to create services directory: %v", err)
	}
	if err := os.WriteFile(jsonFile, []byte(`{"name": "api", "port": 8080}`), 0600); err != nil {
		t.Fatalf("Failed to create JSON file: %v", err)
	}

	tree, err := NewTree(tmpDir)
	if err != nil {
		t.Fatalf("NewTree() error = %v", err)
	}

	if tree == nil {
		t.Fatal("NewTree() returned nil tree")
	}

	// Verify JSON file is included in the tree
	result, err := tree.MarshalYAML()
	if err != nil {
		t.Fatalf("MarshalYAML() error = %v", err)
	}

	resultMap, ok := result.(map[string]interface{})
	if !ok {
		t.Fatalf("MarshalYAML() returned %T, want map[string]interface{}", result)
	}

	servicesMap, ok := resultMap["services"].(map[string]interface{})
	if !ok {
		t.Fatal("MarshalYAML() result missing 'services' key or not a map")
	}

	apiMap, ok := servicesMap["api"].(map[string]interface{})
	if !ok {
		t.Fatal("MarshalYAML() result missing 'api' key from JSON file or not a map")
	}

	// Verify JSON content was parsed correctly
	// JSON numbers may be parsed as int or float64 depending on the YAML library
	if apiMap["name"] != "api" {
		t.Errorf("MarshalYAML() JSON content name not parsed correctly. Got: %v", apiMap["name"])
	}
	port := apiMap["port"]
	if port != 8080 && port != float64(8080) {
		t.Errorf("MarshalYAML() JSON content port not parsed correctly. Got: %v (type: %T)", port, port)
	}
}

func TestNewTree_JSONSpecialCase(t *testing.T) {
	// Test that @*.json files are recognized as special case files
	tmpDir := t.TempDir()

	servicesDir := filepath.Join(tmpDir, "services")
	atCommonFile := filepath.Join(servicesDir, "@common.json")
	apiFile := filepath.Join(servicesDir, "api.yml")

	if err := os.Mkdir(servicesDir, 0700); err != nil {
		t.Fatalf("Failed to create services directory: %v", err)
	}
	if err := os.WriteFile(atCommonFile, []byte(`{"timeout": 30, "retries": 3}`), 0600); err != nil {
		t.Fatalf("Failed to create @common.json file: %v", err)
	}
	if err := os.WriteFile(apiFile, []byte("name: api"), 0600); err != nil {
		t.Fatalf("Failed to create api.yml file: %v", err)
	}

	tree, err := NewTree(tmpDir)
	if err != nil {
		t.Fatalf("NewTree() error = %v", err)
	}

	result, err := tree.MarshalYAML()
	if err != nil {
		t.Fatalf("MarshalYAML() error = %v", err)
	}

	resultMap, ok := result.(map[string]interface{})
	if !ok {
		t.Fatalf("MarshalYAML() returned %T, want map[string]interface{}", result)
	}

	servicesMap, ok := resultMap["services"].(map[string]interface{})
	if !ok {
		t.Fatal("MarshalYAML() result missing 'services' key or not a map")
	}

	// @common.json should merge into services map
	// JSON numbers may be parsed as int or float64 depending on the YAML library
	timeout := servicesMap["timeout"]
	if timeout != 30 && timeout != float64(30) {
		t.Errorf("MarshalYAML() @common.json content not merged into services map. Got timeout: %v (type: %T)", timeout, timeout)
	}
	// api.yml should also be present
	if servicesMap["api"] == nil {
		t.Error("MarshalYAML() api.yml not found in services map")
	}
}
