package filetree

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"go.yaml.in/yaml/v4"
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
	// Empty directories (with no YAML files) should not appear in output
	if _, ok := resultMap["empty_dir"]; ok {
		t.Error("MarshalYAML() result should not contain 'empty_dir' key (empty directories are ignored)")
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

	// Verify the error message indicates a YAML parsing issue and includes file path
	// The exact message may vary by YAML library version, but should contain "yaml" or "YAML"
	errStr := strings.ToLower(err.Error())
	if !strings.Contains(errStr, "yaml") {
		t.Errorf("yaml.Marshal() error = %v, expected YAML parsing error", err)
	}
	// Verify the error includes the file path for better debugging
	if !strings.Contains(err.Error(), anotherDirFile) {
		t.Errorf("yaml.Marshal() error = %v, expected error to include file path %s", err, anotherDirFile)
	}

	// Verify the error includes position information (line:column) if available
	// New error format should include position like "YAML syntax error in file:line:column"
	if strings.Contains(err.Error(), "YAML syntax error in") {
		// Should have line:column format (e.g., ":0:9:")
		if !strings.Contains(err.Error(), ":") {
			t.Error("yaml.Marshal() error should include position information (line:column)")
		}
	}
}

func TestFormatYAMLError_ParserError(t *testing.T) {
	// Test that formatYAMLError properly formats ParserError with position info
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.yml")

	// Create invalid YAML that will trigger a ParserError
	invalidYAML := "key: [unclosed"
	if err := os.WriteFile(testFile, []byte(invalidYAML), 0600); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	tree, err := NewTree(tmpDir)
	if err != nil {
		t.Fatalf("NewTree() error = %v", err)
	}

	_, err = yaml.Marshal(tree)
	if err == nil {
		t.Fatal("Expected error for invalid YAML")
	}

	// Verify error message format includes position
	errStr := err.Error()
	if !strings.Contains(errStr, "YAML syntax error in") {
		t.Errorf("Expected 'YAML syntax error in' in error message, got: %s", errStr)
	}
	if !strings.Contains(errStr, testFile) {
		t.Errorf("Expected file path in error message, got: %s", errStr)
	}
	// Should have line:column format
	if !strings.Contains(errStr, ":") {
		t.Errorf("Expected position information (line:column) in error message, got: %s", errStr)
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

func TestMarshalLeaf_WithIncludes(t *testing.T) {
	// Test marshalLeaf with EnableIncludes enabled
	tmpDir := t.TempDir()

	commandsDir := filepath.Join(tmpDir, "commands")
	scriptsDir := filepath.Join(commandsDir, "scripts")
	if err := os.MkdirAll(scriptsDir, 0700); err != nil {
		t.Fatalf("Failed to create directories: %v", err)
	}

	// Create a script to include
	scriptFile := filepath.Join(scriptsDir, "test.sh")
	scriptContent := "#!/bin/bash\necho 'test'"
	if err := os.WriteFile(scriptFile, []byte(scriptContent), 0600); err != nil {
		t.Fatalf("Failed to create script: %v", err)
	}

	// Create YAML with include
	yamlFile := filepath.Join(commandsDir, "test.yml")
	yamlContent := `command: <<include(scripts/test.sh)>>`
	if err := os.WriteFile(yamlFile, []byte(yamlContent), 0600); err != nil {
		t.Fatalf("Failed to create YAML file: %v", err)
	}

	absDir, err := filepath.Abs(tmpDir)
	if err != nil {
		t.Fatalf("Failed to get absolute path: %v", err)
	}

	opts := &Options{
		EnableIncludes: true,
		PackRoot:       absDir,
	}

	tree, err := NewTree(tmpDir)
	if err != nil {
		t.Fatalf("NewTree() error = %v", err)
	}

	// Find the test.yml node
	var testNode *Node
	for _, child := range tree.Children {
		if child.Info.Name() == "commands" {
			for _, cmdChild := range child.Children {
				if cmdChild.Info.Name() == "test.yml" {
					testNode = cmdChild
					break
				}
			}
		}
	}

	if testNode == nil {
		t.Fatal("Could not find test.yml node")
	}

	// Test with includes enabled
	result, err := testNode.marshalLeaf(opts)
	if err != nil {
		t.Fatalf("marshalLeaf() with includes error = %v", err)
	}

	resultMap, ok := result.(map[string]interface{})
	if !ok {
		t.Fatalf("marshalLeaf() returned %T, want map[string]interface{}", result)
	}

	// The include should have been processed
	commandVal, ok := resultMap["command"].(string)
	if !ok {
		t.Fatalf("marshalLeaf() command value is %T, want string", resultMap["command"])
	}

	if !strings.Contains(commandVal, "echo 'test'") {
		t.Errorf("marshalLeaf() should contain included content. Got: %q", commandVal)
	}
	if strings.Contains(commandVal, "<<include") {
		t.Error("marshalLeaf() should not contain include directive after processing")
	}
}

func TestMarshalLeaf_WithIncludes_ErrorCases(t *testing.T) {
	tmpDir := t.TempDir()

	// Test error when include file doesn't exist
	yamlFile := filepath.Join(tmpDir, "test.yml")
	yamlContent := `command: <<include(nonexistent.sh)>>`
	if err := os.WriteFile(yamlFile, []byte(yamlContent), 0600); err != nil {
		t.Fatalf("Failed to create YAML file: %v", err)
	}

	absDir, err := filepath.Abs(tmpDir)
	if err != nil {
		t.Fatalf("Failed to get absolute path: %v", err)
	}

	opts := &Options{
		EnableIncludes: true,
		PackRoot:       absDir,
	}

	tree, err := NewTree(tmpDir)
	if err != nil {
		t.Fatalf("NewTree() error = %v", err)
	}

	testNode := tree.Children[0]

	_, err = testNode.marshalLeaf(opts)
	if err == nil {
		t.Error("marshalLeaf() expected error for missing include file")
	}
	if !strings.Contains(err.Error(), "could not open") {
		t.Errorf("marshalLeaf() error = %v, want 'could not open'", err)
	}
}

func TestMarshalLeaf_FileReadError(t *testing.T) {
	// Test error handling when file cannot be read
	tmpDir := t.TempDir()
	yamlFile := filepath.Join(tmpDir, "test.yml")
	if err := os.WriteFile(yamlFile, []byte("key: value"), 0600); err != nil {
		t.Fatalf("Failed to create file: %v", err)
	}

	tree, err := NewTree(tmpDir)
	if err != nil {
		t.Fatalf("NewTree() error = %v", err)
	}

	testNode := tree.Children[0]

	// Remove read permission to trigger error
	if err := os.Chmod(yamlFile, 0000); err != nil {
		t.Fatalf("Failed to chmod file: %v", err)
	}
	defer func() {
		_ = os.Chmod(yamlFile, 0600) // Restore for cleanup
	}()

	_, err = testNode.marshalLeaf(nil)
	if err == nil {
		t.Error("marshalLeaf() expected error for unreadable file")
	}
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
	tmpDir := t.TempDir()

	// Create a subdirectory so the file becomes a nested key
	configDir := filepath.Join(tmpDir, "config")
	if err := os.MkdirAll(configDir, 0700); err != nil {
		t.Fatalf("Failed to create config directory: %v", err)
	}

	yamlFile := filepath.Join(configDir, "test.yml")
	yamlContent := `enabled: on
disabled: "off"
name: on_call_service`
	if err := os.WriteFile(yamlFile, []byte(yamlContent), 0600); err != nil {
		t.Fatalf("Failed to create YAML file: %v", err)
	}

	absDir, err := filepath.Abs(tmpDir)
	if err != nil {
		t.Fatalf("Failed to get absolute path: %v", err)
	}

	// Test with conversion enabled
	opts := &Options{
		EnableIncludes:  false,
		PackRoot:        absDir,
		ConvertBooleans: true,
	}

	tree, err := NewTree(tmpDir)
	if err != nil {
		t.Fatalf("NewTree() error = %v", err)
	}

	result, err := tree.Marshal(opts)
	if err != nil {
		t.Fatalf("Marshal() error = %v", err)
	}

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
	tmpDir := t.TempDir()

	// Create a subdirectory so the file becomes a nested key
	configDir := filepath.Join(tmpDir, "config")
	if err := os.MkdirAll(configDir, 0700); err != nil {
		t.Fatalf("Failed to create config directory: %v", err)
	}

	yamlFile := filepath.Join(configDir, "test.yml")
	yamlContent := `enabled: on
disabled: off`
	if err := os.WriteFile(yamlFile, []byte(yamlContent), 0600); err != nil {
		t.Fatalf("Failed to create YAML file: %v", err)
	}

	absDir, err := filepath.Abs(tmpDir)
	if err != nil {
		t.Fatalf("Failed to get absolute path: %v", err)
	}

	// Test WITHOUT conversion
	opts := &Options{
		EnableIncludes:  false,
		PackRoot:        absDir,
		ConvertBooleans: false,
	}

	tree, err := NewTree(tmpDir)
	if err != nil {
		t.Fatalf("NewTree() error = %v", err)
	}

	result, err := tree.Marshal(opts)
	if err != nil {
		t.Fatalf("Marshal() error = %v", err)
	}

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
	tmpDir := t.TempDir()

	// Create a directory with empty subdirectories
	empty1 := filepath.Join(tmpDir, "empty1")
	empty2 := filepath.Join(tmpDir, "empty2")
	hasContent := filepath.Join(tmpDir, "has_content")
	hasContentFile := filepath.Join(hasContent, "file.yml")

	if err := os.MkdirAll(empty1, 0700); err != nil {
		t.Fatalf("Failed to create empty1: %v", err)
	}
	if err := os.MkdirAll(empty2, 0700); err != nil {
		t.Fatalf("Failed to create empty2: %v", err)
	}
	if err := os.MkdirAll(hasContent, 0700); err != nil {
		t.Fatalf("Failed to create has_content: %v", err)
	}
	if err := os.WriteFile(hasContentFile, []byte("key: value"), 0600); err != nil {
		t.Fatalf("Failed to create file: %v", err)
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
