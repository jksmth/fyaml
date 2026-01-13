package filetree

import (
	"bytes"
	"reflect"
	"strings"
	"testing"

	"go.yaml.in/yaml/v4"
)

// Helper functions for test assertions

// getNestedMap extracts a nested map value, handling both map types
func getNestedMap(t *testing.T, m map[interface{}]interface{}, key string) map[interface{}]interface{} {
	t.Helper()
	val, ok := m[key]
	if !ok {
		keys := make([]interface{}, 0, len(m))
		for k := range m {
			keys = append(keys, k)
		}
		t.Fatalf("key %q not found in map. Available keys: %v", key, keys)
	}
	return asMapShared(t, val)
}

func TestCollectAnchorsFromNode(t *testing.T) {
	tests := []struct {
		name     string
		yaml     string
		expected map[string]bool // anchor names we expect to find
	}{
		{
			name: "single anchor",
			yaml: `defaults: &defaults
  timeout: 30
  retries: 3`,
			expected: map[string]bool{"defaults": true},
		},
		{
			name: "multiple anchors",
			yaml: `step1: &step1
  name: Step 1
step2: &step2
  name: Step 2`,
			expected: map[string]bool{"step1": true, "step2": true},
		},
		{
			name: "nested anchor",
			yaml: `config:
  defaults: &defaults
    timeout: 30`,
			expected: map[string]bool{"defaults": true},
		},
		{
			name: "no anchors",
			yaml: `key: value
other: data`,
			expected: map[string]bool{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var node yaml.Node
			if err := yaml.Unmarshal([]byte(tt.yaml), &node); err != nil {
				t.Fatalf("Failed to parse YAML: %v", err)
			}

			registry := &anchorRegistry{
				anchors: make(map[string]*yaml.Node),
			}

			opts := &Options{}
			ctx := &collectContext{
				registry: registry,
				filePath: "test.yml",
				opts:     opts,
			}
			collectAnchorsFromNode(&node, ctx)

			// Check that we found the expected anchors
			if len(registry.anchors) != len(tt.expected) {
				t.Errorf("Expected %d anchors, got %d", len(tt.expected), len(registry.anchors))
			}

			for anchorName := range tt.expected {
				if _, found := registry.anchors[anchorName]; !found {
					t.Errorf("Expected anchor %s not found", anchorName)
				}
			}

			for anchorName := range registry.anchors {
				if !tt.expected[anchorName] {
					t.Errorf("Unexpected anchor %s found", anchorName)
				}
			}
		})
	}
}

func TestDeepCopyNode(t *testing.T) {
	original := &yaml.Node{
		Kind:   yaml.MappingNode,
		Tag:    "!!map",
		Value:  "",
		Anchor: "test_anchor",
		Content: []*yaml.Node{
			{Kind: yaml.ScalarNode, Value: "key"},
			{Kind: yaml.ScalarNode, Value: "value"},
		},
	}

	copy := deepCopyNode(original)

	if copy == original {
		t.Error("deepCopyNode returned the same pointer")
	}

	if !reflect.DeepEqual(copy, original) {
		t.Error("deepCopyNode did not produce an equal copy")
	}

	// Verify it's a deep copy - modifying copy shouldn't affect original
	copy.Content[0].Value = "modified"
	if original.Content[0].Value == "modified" {
		t.Error("deepCopyNode did not create a deep copy - modifying copy affected original")
	}
}

// TestCrossFileAnchors_Integration tests the full cross-file anchor resolution flow.
func TestCrossFileAnchors_Integration(t *testing.T) {
	tmpDir := createTestDir(t, map[string]string{
		"definitions.yml": `template: &common_template
  timeout: 30
  retries: 3
  enabled: true
`,
		"uses.yml": `entity:
  id: example1
  config: *common_template
  attributes:
    name: sample name
`,
	}, nil)

	// Build tree
	tree, err := NewTree(tmpDir)
	assertNoError(t, err)

	// Test with cross-file anchors enabled
	opts := &Options{
		EnableAnchors: true,
		Mode:          ModeCanonical,
		Logger:        nil,
	}

	result, err := tree.Marshal(opts)
	assertNoError(t, err)

	// Verify result is a map
	resultMap := asMapShared(t, result)

	// With multiple files, they're merged at root level
	// Check that entity exists at root level (from uses.yml)
	entity := getNestedMap(t, resultMap, "entity")

	// In canonical mode, alias should be expanded
	config := getNestedMap(t, entity, "config")

	// Verify the resolved values
	if config["timeout"] != 30 {
		t.Errorf("Expected timeout=30, got %v", config["timeout"])
	}
	if config["retries"] != 3 {
		t.Errorf("Expected retries=3, got %v", config["retries"])
	}
	if config["enabled"] != true {
		t.Errorf("Expected enabled=true, got %v", config["enabled"])
	}
}

// TestCrossFileAnchors_PreserveMode tests cross-file anchors in preserve mode.
func TestCrossFileAnchors_PreserveMode(t *testing.T) {
	tmpDir := createTestDir(t, map[string]string{
		"definitions.yml": `template: &common_template
  timeout: 30
  retries: 3
`,
		"uses.yml": `entity:
  id: example1
  config: *common_template
`,
	}, nil)

	// Build tree
	tree, err := NewTree(tmpDir)
	assertNoError(t, err)

	// Test with cross-file anchors enabled in preserve mode
	opts := &Options{
		EnableAnchors: true,
		Mode:          ModePreserve,
		Logger:        nil,
	}

	result, err := tree.Marshal(opts)
	assertNoError(t, err)

	// In preserve mode, result is a *yaml.Node
	resultNode, ok := result.(*yaml.Node)
	if !ok {
		t.Fatalf("Expected *yaml.Node in preserve mode, got %T", result)
	}

	// Marshal to YAML string to check alias is preserved
	var buf bytes.Buffer
	enc := yaml.NewEncoder(&buf)
	if err := enc.Encode(resultNode); err != nil {
		t.Fatalf("Failed to encode result: %v", err)
	}
	if err := enc.Close(); err != nil {
		t.Fatalf("Failed to close encoder: %v", err)
	}

	resultYAML := buf.String()

	// In preserve mode, alias reference should be preserved
	if !strings.Contains(resultYAML, "*common_template") {
		t.Errorf("Expected alias reference *common_template to be preserved. Got:\n%s", resultYAML)
	}
}

// TestCrossFileAnchors_MultiDocumentFile tests that anchors defined in multiple
// documents within a single file are properly collected and available.
func TestCrossFileAnchors_MultiDocumentFile(t *testing.T) {
	tmpDir := createTestDir(t, map[string]string{
		"multi-doc.yml": `---
# First document
anchor1: &anchor1
  value: first
  source: doc1
---
# Second document
anchor2: &anchor2
  value: second
  source: doc2
---
# Third document uses anchor1
uses:
  ref1: *anchor1
  ref2: *anchor2
`,
	}, nil)

	// Build tree
	tree, err := NewTree(tmpDir)
	assertNoError(t, err)

	// Test with cross-file anchors enabled
	opts := &Options{
		EnableAnchors: true,
		Mode:          ModeCanonical,
		Logger:        nil,
	}

	result, err := tree.Marshal(opts)
	assertNoError(t, err)

	// Verify result is a map
	resultMap := asMapShared(t, result)

	// Single file in directory: output is merged content directly (no file-name key)
	// All three documents are merged into one map
	anchor1 := getNestedMap(t, resultMap, "anchor1")
	if anchor1["value"] != "first" {
		t.Errorf("Expected anchor1 value 'first', got %v", anchor1["value"])
	}

	anchor2 := getNestedMap(t, resultMap, "anchor2")
	if anchor2["value"] != "second" {
		t.Errorf("Expected anchor2 value 'second', got %v", anchor2["value"])
	}

	// Document 3: uses both anchors (should be resolved)
	uses := getNestedMap(t, resultMap, "uses")
	ref1 := getNestedMap(t, uses, "ref1")
	if ref1["value"] != "first" {
		t.Errorf("Expected ref1 to resolve to anchor1 (value='first'), got %v", ref1)
	}

	ref2 := getNestedMap(t, uses, "ref2")
	if ref2["value"] != "second" {
		t.Errorf("Expected ref2 to resolve to anchor2 (value='second'), got %v", ref2)
	}
}

// TestCrossFileAnchors_MultiDocumentFile_CrossFile tests anchors defined in
// multiple documents of one file, used in another file.
func TestCrossFileAnchors_MultiDocumentFile_CrossFile(t *testing.T) {
	tmpDir := createTestDir(t, map[string]string{
		"definitions.yml": `---
# First document
anchor1: &anchor1
  value: first
---
# Second document
anchor2: &anchor2
  value: second
`,
		"uses.yml": `uses:
  ref1: *anchor1
  ref2: *anchor2
`,
	}, nil)

	// Build tree
	tree, err := NewTree(tmpDir)
	assertNoError(t, err)

	// Test with cross-file anchors enabled
	opts := &Options{
		EnableAnchors: true,
		Mode:          ModeCanonical,
		Logger:        nil,
	}

	result, err := tree.Marshal(opts)
	assertNoError(t, err)

	// Verify result
	resultMap := asMapShared(t, result)

	// Multiple files are merged at root level (not wrapped in file-name keys)
	// Check that anchors from definitions.yml are at root level
	anchor1 := getNestedMap(t, resultMap, "anchor1")
	if anchor1["value"] != "first" {
		t.Errorf("Expected anchor1 value 'first', got %v", anchor1["value"])
	}

	anchor2 := getNestedMap(t, resultMap, "anchor2")
	if anchor2["value"] != "second" {
		t.Errorf("Expected anchor2 value 'second', got %v", anchor2["value"])
	}

	// Check uses file resolved both anchors
	uses := getNestedMap(t, resultMap, "uses")
	ref1 := getNestedMap(t, uses, "ref1")
	if ref1["value"] != "first" {
		t.Errorf("Expected ref1 to resolve to anchor1, got %v", ref1)
	}

	ref2 := getNestedMap(t, uses, "ref2")
	if ref2["value"] != "second" {
		t.Errorf("Expected ref2 to resolve to anchor2, got %v", ref2)
	}
}

// TestCrossFileAnchors_MultiDocumentFile_AnchorInLaterDoc tests that an anchor
// defined in a later document (not the first) is still collected and available.
func TestCrossFileAnchors_MultiDocumentFile_AnchorInLaterDoc(t *testing.T) {
	tmpDir := createTestDir(t, map[string]string{
		"test.yml": `---
# First document (no anchors)
data1: value1
---
# Second document (defines anchor)
anchor1: &anchor1
  value: from-doc2
---
# Third document (uses anchor from doc 2)
uses: *anchor1
`,
	}, nil)

	// Build tree
	tree, err := NewTree(tmpDir)
	assertNoError(t, err)

	// Test with cross-file anchors enabled
	opts := &Options{
		EnableAnchors: true,
		Mode:          ModeCanonical,
		Logger:        nil,
	}

	result, err := tree.Marshal(opts)
	assertNoError(t, err)

	// Verify result
	resultMap := asMapShared(t, result)

	// Single file: output is merged content directly (no file-name key)
	// Check first document
	if resultMap["data1"] != "value1" {
		t.Errorf("Expected data1='value1', got %v", resultMap["data1"])
	}

	// Check second document (anchor definition)
	anchor1 := getNestedMap(t, resultMap, "anchor1")
	if anchor1["value"] != "from-doc2" {
		t.Errorf("Expected anchor1 value 'from-doc2', got %v", anchor1["value"])
	}

	// Check third document (uses anchor - should be resolved)
	uses := getNestedMap(t, resultMap, "uses")
	if uses["value"] != "from-doc2" {
		t.Errorf("Expected uses to resolve to anchor1 (value='from-doc2'), got %v", uses)
	}
}

// TestCrossFileAnchors_EmptyDocumentHandling tests that empty documents don't
// interfere with skipping the prepended anchor definitions document.
func TestCrossFileAnchors_EmptyDocumentHandling(t *testing.T) {
	tmpDir := createTestDir(t, map[string]string{
		"anchor.yml": `template: &common_template
  timeout: 30
  retries: 3
`,
		"test.yml": `---
# Empty first document (will be skipped)
entity:
  id: example1
  config: *common_template
  attributes:
    name: sample name
`,
	}, nil)

	// Build tree
	tree, err := NewTree(tmpDir)
	assertNoError(t, err)

	// Test with cross-file anchors enabled
	opts := &Options{
		EnableAnchors: true,
		Mode:          ModeCanonical,
		Logger:        nil,
	}

	result, err := tree.Marshal(opts)
	assertNoError(t, err)

	// Verify result
	resultMap := asMapShared(t, result)

	// With multiple files, they're merged at root level
	// The test.yml file should have its content (entity) in the output
	// If the bug exists, the entity would be missing because we skipped it incorrectly
	// Check that entity exists at root level (from test.yml)
	// Note: anchor.yml contributes "template" at root, test.yml contributes "entity" at root
	entity := getNestedMap(t, resultMap, "entity")

	// Verify the entity has the expected structure
	if entity["id"] != "example1" {
		t.Errorf("Expected id='example1', got %v", entity["id"])
	}

	// Verify the anchor was resolved
	config := getNestedMap(t, entity, "config")
	if config["timeout"] != 30 {
		t.Errorf("Expected timeout=30, got %v", config["timeout"])
	}
}

// TestCrossFileAnchors_UserYAMLStartsWithSeparator tests that user YAML starting with ---
// doesn't create double separators when prepending anchors.
func TestCrossFileAnchors_UserYAMLStartsWithSeparator(t *testing.T) {
	tmpDir := createTestDir(t, map[string]string{
		"anchor.yml": `template: &common_template
  timeout: 30
`,
		"uses.yml": `---
# First document
entity:
  id: example1
  config: *common_template
---
# Second document
other:
  value: test
`,
	}, nil)

	// Build tree
	tree, err := NewTree(tmpDir)
	assertNoError(t, err)

	opts := &Options{
		EnableAnchors: true,
		Mode:          ModeCanonical,
		Logger:        nil,
	}

	result, err := tree.Marshal(opts)
	assertNoError(t, err)

	// Verify result - both documents should be present
	resultMap := asMapShared(t, result)

	// Check first document
	entity := getNestedMap(t, resultMap, "entity")

	// Verify anchor was resolved
	config := getNestedMap(t, entity, "config")
	if config["timeout"] != 30 {
		t.Errorf("Expected timeout=30, got %v", config["timeout"])
	}

	// Check second document
	if resultMap["other"] == nil {
		t.Errorf("Expected 'other' key from second document")
	}
}

// TestCrossFileAnchors_AllFilesFailToParse tests that the incremental collection
// terminates correctly when all files fail to parse.
func TestCrossFileAnchors_AllFilesFailToParse(t *testing.T) {
	tmpDir := createTestDir(t, map[string]string{
		"file1.yml": `entity:
  config: *nonexistent_anchor
`,
		"file2.yml": `other:
  ref: *another_missing_anchor
`,
	}, nil)

	// Build tree
	tree, err := NewTree(tmpDir)
	assertNoError(t, err)

	opts := &Options{
		EnableAnchors: true,
		Mode:          ModeCanonical,
		Logger:        nil,
	}

	// Should not hang or error - should return empty result or handle gracefully
	result, err := tree.Marshal(opts)
	if err != nil {
		// Error is acceptable - means it detected the issue
		return
	}

	// If no error, result should be empty or minimal
	if result != nil {
		resultMap := asMapShared(t, result)
		if len(resultMap) > 0 {
			t.Logf("Warning: Got result with %d keys despite unresolvable aliases", len(resultMap))
		}
	}
}

// TestCrossFileAnchors_FileFailsThenSucceeds tests incremental collection where
// a file fails initially but succeeds after anchors are collected.
func TestCrossFileAnchors_FileFailsThenSucceeds(t *testing.T) {
	tmpDir := createTestDir(t, map[string]string{
		"definitions.yml": `template: &common_template
  timeout: 30
  retries: 3
`,
		"uses.yml": `entity:
  id: example1
  config: *common_template
`,
	}, nil)

	// Build tree
	tree, err := NewTree(tmpDir)
	assertNoError(t, err)

	opts := &Options{
		EnableAnchors: true,
		Mode:          ModeCanonical,
		Logger:        nil,
	}

	result, err := tree.Marshal(opts)
	assertNoError(t, err)

	// Verify both files contributed to result
	resultMap := asMapShared(t, result)

	// Should have template from file1
	if resultMap["template"] == nil {
		t.Errorf("Expected 'template' key from definitions.yml")
	}

	// Should have entity from file2
	entity := getNestedMap(t, resultMap, "entity")

	// Verify anchor was resolved
	config := getNestedMap(t, entity, "config")
	if config["timeout"] != 30 {
		t.Errorf("Expected timeout=30, got %v", config["timeout"])
	}
}

// TestCrossFileAnchors_CircularDependencies tests that circular anchor dependencies
// are handled correctly by the incremental collection mechanism.
func TestCrossFileAnchors_CircularDependencies(t *testing.T) {
	tmpDir := createTestDir(t, map[string]string{
		"file1.yml": `anchor1: &anchor1
  name: first
  ref: *anchor2
`,
		"file2.yml": `anchor2: &anchor2
  name: second
  ref: *anchor1
`,
		"uses.yml": `uses:
  item1: *anchor1
  item2: *anchor2
`,
	}, nil)

	// Build tree
	tree, err := NewTree(tmpDir)
	assertNoError(t, err)

	opts := &Options{
		EnableAnchors: true,
		Mode:          ModeCanonical,
		Logger:        nil,
	}

	// Should handle circular dependencies gracefully
	// The incremental collection should eventually collect both anchors
	result, err := tree.Marshal(opts)
	if err != nil {
		// Circular dependencies might cause issues, but shouldn't crash
		// If it errors, that's acceptable - we're testing it doesn't hang
		t.Logf("Circular dependencies caused error (acceptable): %v", err)
		return
	}

	// If successful, verify structure
	resultMap := asMapShared(t, result)

	// Should have both anchors defined
	if resultMap["anchor1"] == nil {
		t.Errorf("Expected anchor1 to be defined")
	}
	if resultMap["anchor2"] == nil {
		t.Errorf("Expected anchor2 to be defined")
	}
}

// TestCrossFileAnchors_AnchorsButNoAliases tests that files with anchors but no aliases
// that use them work correctly.
func TestCrossFileAnchors_AnchorsButNoAliases(t *testing.T) {
	tmpDir := createTestDir(t, map[string]string{
		"anchor.yml": `template: &unused_template
  timeout: 30
  retries: 3
other:
  value: test
`,
		"other.yml": `entity:
  id: example1
`,
	}, nil)

	// Build tree
	tree, err := NewTree(tmpDir)
	assertNoError(t, err)

	opts := &Options{
		EnableAnchors: true,
		Mode:          ModeCanonical,
		Logger:        nil,
	}

	result, err := tree.Marshal(opts)
	assertNoError(t, err)

	// Should work fine - anchor is collected but not used
	resultMap := asMapShared(t, result)

	// Should have all content
	if resultMap["template"] == nil {
		t.Errorf("Expected 'template' key")
	}
	if resultMap["other"] == nil {
		t.Errorf("Expected 'other' key")
	}
	if resultMap["entity"] == nil {
		t.Errorf("Expected 'entity' key")
	}
}
