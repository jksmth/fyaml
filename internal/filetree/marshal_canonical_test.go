package filetree

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/jksmth/fyaml/internal/logger"
	"go.yaml.in/yaml/v4"
)

// marshal_canonical_test.go contains tests for canonical mode marshaling.

// asMap converts interface{} to map[interface{}]interface{} for test assertions.
// Handles both map[string]interface{} and map[interface{}]interface{}.
func asMap(t *testing.T, v interface{}) map[interface{}]interface{} {
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

func TestMarshalCanonical_WithIncludes(t *testing.T) {
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

	testNode := findNodeByName(t, tree, "item1.yml")
	if testNode == nil {
		t.Fatal("Could not find item1.yml node")
	}

	result, err := testNode.marshalLeaf(opts)
	assertNoError(t, err)

	resultMap := asMap(t, result)
	entityMap := asMap(t, resultMap["entity"])
	attributesMap := asMap(t, entityMap["attributes"])

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

func TestMarshalCanonical_WithConvertBooleans(t *testing.T) {
	tmpDir := createTestDir(t, map[string]string{
		"config/test.yml": `enabled: on
disabled: "off"
name: on_call_service`,
	}, nil)

	absDir, err := filepath.Abs(tmpDir)
	assertNoError(t, err)

	opts := &Options{
		EnableIncludes:  false,
		PackRoot:        absDir,
		ConvertBooleans: true,
	}

	tree, err := NewTree(tmpDir)
	assertNoError(t, err)

	result, err := tree.Marshal(opts)
	assertNoError(t, err)

	resultMap := asMap(t, result)
	configMap := asMap(t, resultMap["config"])
	testMap := asMap(t, configMap["test"])

	if enabled, ok := testMap["enabled"].(bool); !ok || !enabled {
		t.Errorf("enabled should be bool true, got %T: %v", testMap["enabled"], testMap["enabled"])
	}
	if disabled, ok := testMap["disabled"].(string); !ok || disabled != "off" {
		t.Errorf("disabled should be string 'off', got %T: %v", testMap["disabled"], testMap["disabled"])
	}
	if name, ok := testMap["name"].(string); !ok || name != "on_call_service" {
		t.Errorf("name should be string 'on_call_service', got %T: %v", testMap["name"], testMap["name"])
	}
}

func TestMarshalCanonical_WithoutConvertBooleans(t *testing.T) {
	tmpDir := createTestDir(t, map[string]string{
		"config/test.yml": `enabled: on
disabled: off`,
	}, nil)

	absDir, err := filepath.Abs(tmpDir)
	assertNoError(t, err)

	opts := &Options{
		EnableIncludes:  false,
		PackRoot:        absDir,
		ConvertBooleans: false,
	}

	tree, err := NewTree(tmpDir)
	assertNoError(t, err)

	result, err := tree.Marshal(opts)
	assertNoError(t, err)

	resultMap := asMap(t, result)
	configMap := asMap(t, resultMap["config"])
	testMap := asMap(t, configMap["test"])

	if enabled, ok := testMap["enabled"].(string); !ok || enabled != "on" {
		t.Errorf("enabled should be string 'on' without normalization, got %T: %v", testMap["enabled"], testMap["enabled"])
	}
}

func TestMarshalCanonical_WithEmptyMaps(t *testing.T) {
	tmpDir := createTestDir(t, map[string]string{
		"has_content/file.yml": "key: value",
	}, []string{"empty1", "empty2"})

	resultMap := createTreeAndMarshal(t, tmpDir)

	if _, ok := resultMap["empty1"]; ok {
		t.Error("MarshalYAML() should not contain 'empty1' key")
	}
	if _, ok := resultMap["empty2"]; ok {
		t.Error("MarshalYAML() should not contain 'empty2' key")
	}
	if _, ok := resultMap["has_content"]; !ok {
		t.Error("MarshalYAML() should contain 'has_content' key")
	}
}

func TestMarshalCanonical_RootFile(t *testing.T) {
	tmpDir := createTestDir(t, map[string]string{
		"root_file.yml":   "key1: value1",
		"subdir/file.yml": "key2: value2",
	}, nil)

	tree, err := NewTree(tmpDir)
	assertNoError(t, err)

	rootFileNode := findNodeByName(t, tree, "root_file.yml")
	if rootFileNode == nil {
		t.Fatal("Could not find root_file.yml node")
	}

	subFileNode := findNodeByName(t, tree, "file.yml")
	if subFileNode == nil {
		t.Fatal("Could not find file.yml node")
	}

	if !rootFileNode.rootFile() {
		t.Error("rootFile() returned false for root_file.yml")
	}

	if subFileNode.rootFile() {
		t.Error("rootFile() returned true for file.yml in subdirectory")
	}

	resultMap := createTreeAndMarshal(t, tmpDir)
	if resultMap["key1"] == nil {
		t.Error("Root file content should be merged at top level, not nested")
	}
	if resultMap["key1"] != "value1" {
		t.Errorf("Root file content key1 = %v, want 'value1'", resultMap["key1"])
	}
	subdirMap := asMap(t, resultMap["subdir"])
	fileMap := asMap(t, subdirMap["file"])
	if fileMap["key2"] != "value2" {
		t.Errorf("Subdirectory file content key2 = %v, want 'value2'", fileMap["key2"])
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
			result := isEmptyContent(tt.input)
			if result != tt.expected {
				t.Errorf("isEmptyContent(%v) = %v, want %v", tt.input, result, tt.expected)
			}
		})
	}
}

func TestMergeTree_InvalidInput(t *testing.T) {
	tests := []struct {
		name        string
		input       interface{}
		expectError bool
		errorSubstr string
		validate    func(t *testing.T, result map[interface{}]interface{})
	}{
		{
			name:        "channel",
			input:       make(chan int),
			expectError: true,
			errorSubstr: "expected map",
		},
		{
			name:        "function",
			input:       func() {},
			expectError: true,
			errorSubstr: "expected map",
		},
		{
			name:        "nil",
			input:       nil,
			expectError: false,
			validate: func(t *testing.T, result map[interface{}]interface{}) {
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
			result, err := mergeTree(nil, tt.input, MergeShallow)
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

func TestMarshalCanonical_SortsKeys(t *testing.T) {
	tmpDir := createTestDir(t, map[string]string{
		"config.yml": `z: 1
a: 2
m: 3`,
	}, nil)

	tree, err := NewTree(tmpDir)
	assertNoError(t, err)

	opts := &Options{
		PackRoot: tmpDir,
		Mode:     ModeCanonical,
		Logger:   logger.Nop(),
	}

	result, err := tree.Marshal(opts)
	assertNoError(t, err)

	_, isNode := result.(*yaml.Node)
	if isNode {
		t.Fatal("Canonical mode should return interface{}, not *yaml.Node")
	}

	out, err := yaml.Marshal(result)
	assertNoError(t, err)
	outStr := string(out)

	aIdx := strings.Index(outStr, "a:")
	mIdx := strings.Index(outStr, "m:")
	zIdx := strings.Index(outStr, "z:")

	if aIdx == -1 || mIdx == -1 || zIdx == -1 {
		t.Fatalf("Could not find all keys in output:\n%s", outStr)
	}

	if aIdx >= mIdx || mIdx >= zIdx {
		t.Errorf("Expected key order: a, m, z (sorted). Got:\n%s", outStr)
	}
}

func TestMarshal_RendersToYAML(t *testing.T) {
	tmpDir := createTestDir(t, map[string]string{
		"sub_dir/sub_dir_file.yml": "entity:\n  id: example1\n  attributes:\n    name: sample name",
	}, []string{"empty_dir"})

	resultMap := createTreeAndMarshal(t, tmpDir)

	out, err := yaml.Marshal(resultMap)
	assertNoError(t, err)
	if len(out) == 0 {
		t.Error("yaml.Marshal() returned empty output")
	}

	if _, ok := resultMap["sub_dir"]; !ok {
		t.Error("MarshalYAML() result missing 'sub_dir' key")
	}
	if _, ok := resultMap["empty_dir"]; ok {
		t.Error("MarshalYAML() result should not contain 'empty_dir' key (empty directories are ignored)")
	}
}

func TestMarshal_AtDirectory(t *testing.T) {
	tmpDir := createTestDir(t, map[string]string{
		"entities/@group1/item1.yml": "entity:\n  id: example1\n  attributes:\n    name: first item",
		"entities/@group1/item2.yml": "entity:\n  id: example2\n  attributes:\n    name: second item",
		"entities/@group2/item3.yml": "entity:\n  id: example3\n  attributes:\n    name: third item",
		"entities/item4.yml":         "entity:\n  id: example4\n  attributes:\n    name: fourth item",
	}, nil)

	resultMap := createTreeAndMarshal(t, tmpDir)
	entitiesMap := asMap(t, resultMap["entities"])

	if entitiesMap["item1"] == nil {
		t.Error("MarshalYAML() item1.yml from @group1 not found in entities map")
	}
	if entitiesMap["item2"] == nil {
		t.Error("MarshalYAML() item2.yml from @group1 not found in entities map")
	}
	if entitiesMap["item3"] == nil {
		t.Error("MarshalYAML() item3.yml from @group2 not found in entities map")
	}
	if entitiesMap["item4"] == nil {
		t.Error("MarshalYAML() item4.yml not found in entities map")
	}
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

func TestMarshal_EmptyAtDirectory(t *testing.T) {
	tmpDir := createTestDir(t, map[string]string{
		"entities/item1.yml": "entity:\n  id: example1\n  attributes:\n    name: sample name",
	}, []string{"entities/@group1"})

	resultMap := createTreeAndMarshal(t, tmpDir)
	entitiesMap := asMap(t, resultMap["entities"])

	if entitiesMap["@group1"] != nil {
		t.Error("MarshalYAML() empty @group1 directory should not create a key")
	}
	if entitiesMap["group1"] != nil {
		t.Error("MarshalYAML() empty @group1 directory should not create a 'group1' key")
	}
	if entitiesMap["item1"] == nil {
		t.Error("MarshalYAML() item1.yml not found in entities map")
	}
}

func TestMarshal_NestedAtDirectories(t *testing.T) {
	tmpDir := createTestDir(t, map[string]string{
		"entities/@group1/@group2/item1.yml": "entity:\n  id: example1\n  attributes:\n    name: first item",
		"entities/@group1/@group2/item2.yml": "entity:\n  id: example2\n  attributes:\n    name: second item",
		"entities/@group1/item3.yml":         "entity:\n  id: example3\n  attributes:\n    name: third item",
	}, nil)

	resultMap := createTreeAndMarshal(t, tmpDir)
	entitiesMap := asMap(t, resultMap["entities"])

	if entitiesMap["item1"] == nil {
		t.Error("MarshalYAML() item1.yml from @group1/@group2 not found in entities map")
	}
	if entitiesMap["item2"] == nil {
		t.Error("MarshalYAML() item2.yml from @group1/@group2 not found in entities map")
	}
	if entitiesMap["item3"] == nil {
		t.Error("MarshalYAML() item3.yml from @group1 not found in entities map")
	}
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

func TestMarshalCanonical_Nested(t *testing.T) {
	tmpDir := createTestDir(t, map[string]string{
		"level1/level2/level3/deep.yml": `outer:
  middle:
    inner:
      key: value
      another: value2`,
	}, nil)

	resultMap := createTreeAndMarshal(t, tmpDir)

	// Verify nested structure is preserved
	level1Map := asMap(t, resultMap["level1"])
	level2Map := asMap(t, level1Map["level2"])
	level3Map := asMap(t, level2Map["level3"])
	deepMap := asMap(t, level3Map["deep"])
	// Verify file content is nested correctly
	outerMap := asMap(t, deepMap["outer"])
	middleMap := asMap(t, outerMap["middle"])
	innerMap := asMap(t, middleMap["inner"])
	if innerMap["key"] != "value" {
		t.Error("Deep nested values should be preserved")
	}
	if innerMap["another"] != "value2" {
		t.Error("Deep nested values should be preserved")
	}
}

func TestMarshalCanonical_AtFiles(t *testing.T) {
	tmpDir := createTestDir(t, map[string]string{
		"entities/@common.yml": "timeout: 30\nretries: 3",
		"entities/item1.yml":   "entity:\n  id: example1\n  attributes:\n    name: sample name",
	}, nil)

	resultMap := createTreeAndMarshal(t, tmpDir)
	entitiesMap := asMap(t, resultMap["entities"])

	// Verify @common.yml content is merged
	if entitiesMap["timeout"] == nil {
		t.Error("MarshalYAML() @common.yml content not merged into entities map")
	}
	if entitiesMap["item1"] == nil {
		t.Error("MarshalYAML() item1.yml not found in entities map")
	}
}

func TestMarshalCanonical_JSONFiles(t *testing.T) {
	tmpDir := createTestDir(t, map[string]string{
		"entities/item1.json": `{"entity": {"id": "example1", "attributes": {"name": "sample name"}}}`,
	}, nil)

	tree, err := NewTree(tmpDir)
	assertNoError(t, err)

	opts := &Options{
		PackRoot: tmpDir,
		Mode:     ModeCanonical,
		Logger:   logger.Nop(),
	}

	result, err := tree.Marshal(opts)
	assertNoError(t, err)

	resultMap := asMap(t, result)

	// Verify JSON content is parsed correctly
	entitiesMap := asMap(t, resultMap["entities"])
	item1Map := asMap(t, entitiesMap["item1"])
	entityMap := asMap(t, item1Map["entity"])
	if entityMap["id"] != "example1" {
		t.Error("MarshalCanonical() JSON content id not parsed correctly")
	}
	attributesMap := asMap(t, entityMap["attributes"])
	if attributesMap["name"] != "sample name" {
		t.Error("MarshalCanonical() JSON content name not parsed correctly")
	}
}

func TestMarshalCanonical_ShallowMerge(t *testing.T) {
	// Test that shallow merge (default) replaces entire nested maps
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

	absDir, err := filepath.Abs(tmpDir)
	assertNoError(t, err)

	opts := &Options{
		PackRoot:      absDir,
		MergeStrategy: MergeShallow,
		Logger:        logger.Nop(),
	}

	tree, err := NewTree(tmpDir)
	assertNoError(t, err)

	result, err := tree.Marshal(opts)
	assertNoError(t, err)

	resultMap := asMap(t, result)
	configMap := asMap(t, resultMap["config"])

	// In shallow merge, later file completely replaces earlier one
	// So setting1 and setting2 should be gone, setting3 should exist
	if configMap["setting1"] != nil {
		t.Error("Shallow merge should replace entire map - setting1 should not exist")
	}
	if configMap["setting2"] != nil {
		t.Error("Shallow merge should replace entire map - setting2 should not exist")
	}
	if configMap["setting3"] != "value3" {
		t.Error("Shallow merge should include values from later file")
	}

	// Nested map should also be completely replaced
	nestedMap := asMap(t, configMap["nested"])
	if nestedMap["a"] != nil {
		t.Error("Shallow merge should replace nested map - 'a' should not exist")
	}
	if nestedMap["b"] != nil {
		t.Error("Shallow merge should replace nested map - 'b' should not exist")
	}
	if nestedMap["c"] != 3 {
		t.Error("Shallow merge should include nested values from later file")
	}
}

func TestMarshalCanonical_DeepMerge(t *testing.T) {
	// Test that deep merge recursively merges nested maps
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

	absDir, err := filepath.Abs(tmpDir)
	assertNoError(t, err)

	opts := &Options{
		PackRoot:      absDir,
		MergeStrategy: MergeDeep,
		Logger:        logger.Nop(),
	}

	tree, err := NewTree(tmpDir)
	assertNoError(t, err)

	result, err := tree.Marshal(opts)
	assertNoError(t, err)

	resultMap := asMap(t, result)
	configMap := asMap(t, resultMap["config"])

	// In deep merge, values from both files should exist
	if configMap["setting1"] != "value1" {
		t.Error("Deep merge should preserve setting1 from base file")
	}
	if configMap["setting2"] != "value2" {
		t.Error("Deep merge should preserve setting2 from base file")
	}
	if configMap["setting3"] != "value3" {
		t.Error("Deep merge should include setting3 from override file")
	}

	// Nested map should be merged recursively
	nestedMap := asMap(t, configMap["nested"])
	if nestedMap["a"] != 1 {
		t.Error("Deep merge should preserve 'a' from base file")
	}
	if nestedMap["b"] != 2 {
		t.Error("Deep merge should preserve 'b' from base file")
	}
	if nestedMap["c"] != 3 {
		t.Error("Deep merge should include 'c' from override file")
	}
}

// TestYAMLEncoderSortsBoolIntString verifies that the yaml.v4 encoder sorts mixed-type keys
// in a deterministic order: bool < int < string. This test will catch any changes in yaml.v4
// encoder behavior that would break our assumption about key sorting.
func TestYAMLEncoderSortsBoolIntString(t *testing.T) {
	m := map[interface{}]interface{}{
		"zebra": 1,
		100:     2,
		true:    3,
		5:       4,
		"apple": 5,
	}

	out, err := yaml.Marshal(m)
	assertNoError(t, err)

	outStr := string(out)

	// Find positions of each key in output
	trueIdx := strings.Index(outStr, "true:")
	fiveIdx := strings.Index(outStr, "5:")
	hundredIdx := strings.Index(outStr, "100:")
	appleIdx := strings.Index(outStr, "apple:")
	zebraIdx := strings.Index(outStr, "zebra:")

	if trueIdx == -1 || fiveIdx == -1 || hundredIdx == -1 || appleIdx == -1 || zebraIdx == -1 {
		t.Fatalf("Could not find all keys in output:\n%s", outStr)
	}

	// Verify order: bool < int (sorted numerically) < string (sorted alphabetically)
	// Expected: true, 5, 100, apple, zebra
	if trueIdx >= fiveIdx {
		t.Errorf("Expected bool before int: true should come before 5\n%s", outStr)
	}
	if fiveIdx >= hundredIdx {
		t.Errorf("Expected ints sorted numerically: 5 should come before 100\n%s", outStr)
	}
	if hundredIdx >= appleIdx {
		t.Errorf("Expected int before string: 100 should come before apple\n%s", outStr)
	}
	if appleIdx >= zebraIdx {
		t.Errorf("Expected strings sorted alphabetically: apple should come before zebra\n%s", outStr)
	}
}

// TestYAMLEncoderDeterministic verifies that the yaml.v4 encoder produces identical output
// for the same input across multiple runs. This test will catch any changes in yaml.v4
// encoder behavior that would break deterministic output.
func TestYAMLEncoderDeterministic(t *testing.T) {
	m := map[interface{}]interface{}{
		"z": 1, "a": 2, 100: 3, 5: 4, true: 5, false: 6,
	}

	var outputs []string
	for i := 0; i < 10; i++ {
		out, err := yaml.Marshal(m)
		assertNoError(t, err)
		outputs = append(outputs, string(out))
	}

	// All outputs should be identical
	for i := 1; i < len(outputs); i++ {
		if outputs[i] != outputs[0] {
			t.Errorf("Output not deterministic:\nRun 0:\n%s\nRun %d:\n%s", outputs[0], i, outputs[i])
		}
	}
}
