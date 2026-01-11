// Package filetree provides filesystem traversal for FYAML packing.
package filetree

import (
	"fmt"
	"path/filepath"
	"strings"
	"testing"

	"go.yaml.in/yaml/v4"
)

// parse_test.go contains tests for YAML parsing functions (parse.go).

func TestFormatYAMLError_TypeError(t *testing.T) {
	testFile := "/test/path/file.yml"

	tests := []struct {
		name     string
		typeErr  *yaml.TypeError
		wantLine bool
	}{
		{
			name: "with line/column info",
			typeErr: &yaml.TypeError{
				Errors: []*yaml.UnmarshalError{
					{Line: 2, Column: 5, Err: fmt.Errorf("cannot decode !!str as !!int")},
				},
			},
			wantLine: true,
		},
		{
			name: "without line/column info",
			typeErr: &yaml.TypeError{
				Errors: []*yaml.UnmarshalError{
					{Line: 0, Column: 0, Err: fmt.Errorf("type conversion error")},
				},
			},
			wantLine: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			formattedErr := formatYAMLError(tt.typeErr, testFile)
			errStr := formattedErr.Error()

			assertErrorContains(t, formattedErr, "YAML/JSON type errors in")
			assertErrorContains(t, formattedErr, testFile)

			lines := strings.Split(errStr, "\n")
			if len(lines) < 2 || lines[1] == "" {
				t.Fatalf("Expected error detail line, got: %s", errStr)
			}
			detailLine := lines[1]
			hasLineFormat := strings.Contains(detailLine, "line ")
			if tt.wantLine != hasLineFormat {
				t.Errorf("wantLine=%v but hasLineFormat=%v, detail: %q", tt.wantLine, hasLineFormat, detailLine)
			}
			if !strings.HasPrefix(detailLine, "  ") {
				t.Errorf("Expected error line to start with '  ', got: %q", detailLine)
			}
		})
	}
}

func TestNormalizeYAML11Booleans(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
		wantBool bool
	}{
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
		{"double-quoted on", `value: "on"`, "on", false},
		{"double-quoted yes", `value: "yes"`, "yes", false},
		{"single-quoted on", "value: 'on'", "on", false},
		{"single-quoted yes", "value: 'yes'", "yes", false},
		{"unquoted true", "value: true", "true", true},
		{"unquoted false", "value: false", "false", true},
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
	if quiet, ok := settings["quiet"].(string); !ok || quiet != "no" {
		t.Errorf("quiet should be string 'no', got %T: %v", settings["quiet"], settings["quiet"])
	}

	items := config["items"].([]interface{})
	item0 := items[0].(map[string]interface{})
	if active, ok := item0["active"].(bool); !ok || !active {
		t.Errorf("items[0].active should be bool true, got %T: %v", item0["active"], item0["active"])
	}
	item1 := items[1].(map[string]interface{})
	if active, ok := item1["active"].(string); !ok || active != "off" {
		t.Errorf("items[1].active should be string 'off', got %T: %v", item1["active"], item1["active"])
	}
}

// --- Multi-document YAML file tests ---

func TestParseYAMLFile_MultiDocument_Canonical(t *testing.T) {
	tmpDir := createTestDir(t, map[string]string{
		"config.yml": `timeout: 30
retries: 3
---
timeout: 60
debug: true`,
	}, nil)

	absDir, err := filepath.Abs(tmpDir)
	assertNoError(t, err)

	opts := &Options{
		PackRoot: absDir,
		Mode:     ModeCanonical,
	}

	tree, err := NewTree(tmpDir)
	assertNoError(t, err)

	configNode := findNodeByName(t, tree, "config.yml")
	if configNode == nil {
		t.Fatal("Could not find config.yml node")
	}

	result, err := configNode.parseYAMLFile(opts)
	assertNoError(t, err)

	var content map[string]interface{}
	if err := result.Decode(&content); err != nil {
		t.Fatalf("Failed to decode result: %v", err)
	}

	// Verify merged result - later document overwrites timeout, adds debug
	if content["timeout"] != 60 {
		t.Errorf("Expected timeout=60 (from second document), got %v", content["timeout"])
	}
	if content["retries"] != 3 {
		t.Errorf("Expected retries=3 (from first document), got %v", content["retries"])
	}
	if content["debug"] != true {
		t.Errorf("Expected debug=true (from second document), got %v", content["debug"])
	}
}

func TestParseYAMLFile_MultiDocument_Preserve(t *testing.T) {
	tmpDir := createTestDir(t, map[string]string{
		"config.yml": `# First document
timeout: 30
retries: 3
---
# Second document
timeout: 60
debug: true`,
	}, nil)

	absDir, err := filepath.Abs(tmpDir)
	assertNoError(t, err)

	opts := &Options{
		PackRoot: absDir,
		Mode:     ModePreserve,
	}

	tree, err := NewTree(tmpDir)
	assertNoError(t, err)

	configNode := findNodeByName(t, tree, "config.yml")
	if configNode == nil {
		t.Fatal("Could not find config.yml node")
	}

	result, err := configNode.parseYAMLFile(opts)
	assertNoError(t, err)

	if result == nil {
		t.Fatal("Expected non-nil result")
	}

	// Verify it's a mapping node
	if result.Kind != yaml.MappingNode {
		t.Fatalf("Expected MappingNode, got %v", result.Kind)
	}

	// Decode to verify content
	var content map[string]interface{}
	if err := result.Decode(&content); err != nil {
		t.Fatalf("Failed to decode result: %v", err)
	}

	if content["timeout"] != 60 {
		t.Errorf("Expected timeout=60, got %v", content["timeout"])
	}
	if content["retries"] != 3 {
		t.Errorf("Expected retries=3, got %v", content["retries"])
	}
	if content["debug"] != true {
		t.Errorf("Expected debug=true, got %v", content["debug"])
	}
}

func TestParseYAMLFile_MultiDocument_AtFile(t *testing.T) {
	tmpDir := createTestDir(t, map[string]string{
		"entities/@shared.yml": `timeout: 30
retries: 3
---
timeout: 60
debug: true`,
		"entities/item1.yml": `entity:
  id: example1`,
	}, nil)

	absDir, err := filepath.Abs(tmpDir)
	assertNoError(t, err)

	opts := &Options{
		PackRoot: absDir,
		Mode:     ModeCanonical,
	}

	tree, err := NewTree(tmpDir)
	assertNoError(t, err)

	result, err := tree.Marshal(opts)
	assertNoError(t, err)

	resultMap := asMapShared(t, result)
	entitiesMap := asMapShared(t, resultMap["entities"])

	// @shared.yml content should be merged into entities (not under a key)
	if entitiesMap["timeout"] != 60 {
		t.Errorf("Expected timeout=60 from merged @shared.yml, got %v", entitiesMap["timeout"])
	}
	if entitiesMap["retries"] != 3 {
		t.Errorf("Expected retries=3 from merged @shared.yml, got %v", entitiesMap["retries"])
	}
	if entitiesMap["debug"] != true {
		t.Errorf("Expected debug=true from merged @shared.yml, got %v", entitiesMap["debug"])
	}

	// item1 should still be present
	if entitiesMap["item1"] == nil {
		t.Error("Expected item1 to be present in entities map")
	}
}

func TestParseYAMLFile_MultiDocument_RootFile(t *testing.T) {
	tmpDir := createTestDir(t, map[string]string{
		"shared.yml": `timeout: 30
retries: 3
---
timeout: 60
debug: true`,
		"subdir/item.yml": `key: value`,
	}, nil)

	absDir, err := filepath.Abs(tmpDir)
	assertNoError(t, err)

	opts := &Options{
		PackRoot: absDir,
		Mode:     ModeCanonical,
	}

	tree, err := NewTree(tmpDir)
	assertNoError(t, err)

	result, err := tree.Marshal(opts)
	assertNoError(t, err)

	resultMap := asMapShared(t, result)

	// Root file content should be merged at root level
	if resultMap["timeout"] != 60 {
		t.Errorf("Expected timeout=60 from merged root file, got %v", resultMap["timeout"])
	}
	if resultMap["retries"] != 3 {
		t.Errorf("Expected retries=3 from merged root file, got %v", resultMap["retries"])
	}
	if resultMap["debug"] != true {
		t.Errorf("Expected debug=true from merged root file, got %v", resultMap["debug"])
	}

	// Subdir should still be present
	if resultMap["subdir"] == nil {
		t.Error("Expected subdir to be present")
	}
}

func TestParseYAMLFile_MultiDocument_DeepMerge(t *testing.T) {
	tmpDir := createTestDir(t, map[string]string{
		"config.yml": `settings:
  timeout: 30
  retries: 3
  nested:
    key1: value1
---
settings:
  timeout: 60
  nested:
    key2: value2`,
	}, nil)

	absDir, err := filepath.Abs(tmpDir)
	assertNoError(t, err)

	opts := &Options{
		PackRoot:      absDir,
		Mode:          ModeCanonical,
		MergeStrategy: MergeDeep,
	}

	tree, err := NewTree(tmpDir)
	assertNoError(t, err)

	configNode := findNodeByName(t, tree, "config.yml")
	if configNode == nil {
		t.Fatal("Could not find config.yml node")
	}

	result, err := configNode.parseYAMLFile(opts)
	assertNoError(t, err)

	var content map[string]interface{}
	if err := result.Decode(&content); err != nil {
		t.Fatalf("Failed to decode result: %v", err)
	}

	settings := asMapShared(t, content["settings"])
	nested := asMapShared(t, settings["nested"])

	// Deep merge: timeout should be overwritten, nested should have both keys
	if settings["timeout"] != 60 {
		t.Errorf("Expected timeout=60 (overwritten), got %v", settings["timeout"])
	}
	if settings["retries"] != 3 {
		t.Errorf("Expected retries=3 (preserved), got %v", settings["retries"])
	}
	if nested["key1"] != "value1" {
		t.Errorf("Expected key1=value1 (from first doc), got %v", nested["key1"])
	}
	if nested["key2"] != "value2" {
		t.Errorf("Expected key2=value2 (from second doc), got %v", nested["key2"])
	}
}

func TestParseYAMLFile_MultiDocument_NonMapError(t *testing.T) {
	tmpDir := createTestDir(t, map[string]string{
		"config.yml": `timeout: 30
---
- item1
- item2`,
	}, nil)

	absDir, err := filepath.Abs(tmpDir)
	assertNoError(t, err)

	opts := &Options{
		PackRoot: absDir,
	}

	tree, err := NewTree(tmpDir)
	assertNoError(t, err)

	configNode := findNodeByName(t, tree, "config.yml")
	if configNode == nil {
		t.Fatal("Could not find config.yml node")
	}

	_, err = configNode.parseYAMLFile(opts)
	if err == nil {
		t.Fatal("Expected error for non-map document, got nil")
	}

	if !strings.Contains(err.Error(), "expected map for document 2") {
		t.Errorf("Expected error message about non-map document, got: %v", err)
	}
}

func TestParseYAMLFile_SingleDocument_Regression(t *testing.T) {
	tmpDir := createTestDir(t, map[string]string{
		"config.yml": `timeout: 30
retries: 3`,
	}, nil)

	absDir, err := filepath.Abs(tmpDir)
	assertNoError(t, err)

	opts := &Options{
		PackRoot: absDir,
		Mode:     ModeCanonical,
	}

	tree, err := NewTree(tmpDir)
	assertNoError(t, err)

	configNode := findNodeByName(t, tree, "config.yml")
	if configNode == nil {
		t.Fatal("Could not find config.yml node")
	}

	result, err := configNode.parseYAMLFile(opts)
	assertNoError(t, err)

	var content map[string]interface{}
	if err := result.Decode(&content); err != nil {
		t.Fatalf("Failed to decode result: %v", err)
	}

	// Single document should work as before
	if content["timeout"] != 30 {
		t.Errorf("Expected timeout=30, got %v", content["timeout"])
	}
	if content["retries"] != 3 {
		t.Errorf("Expected retries=3, got %v", content["retries"])
	}
}
