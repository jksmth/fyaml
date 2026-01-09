package filetree

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/jksmth/fyaml/internal/logger"
	"go.yaml.in/yaml/v4"
)

// marshal_preserve_test.go contains tests for preserve mode marshaling.

func TestMarshalPreserve_PreservesComments(t *testing.T) {
	tmpDir := createTestDir(t, map[string]string{
		"entities/item1.yml": `# Header comment for item1
key: value  # inline comment
# Footer comment`,
	}, nil)

	tree, err := NewTree(tmpDir)
	assertNoError(t, err)

	opts := &Options{
		PackRoot: tmpDir,
		Mode:     ModePreserve,
		Logger:   logger.Nop(),
	}

	result, err := tree.Marshal(opts)
	assertNoError(t, err)

	node, ok := result.(*yaml.Node)
	if !ok {
		t.Fatalf("Expected *yaml.Node, got %T", result)
	}

	out, err := yaml.Marshal(node)
	assertNoError(t, err)

	outStr := string(out)
	if !strings.Contains(outStr, "Header comment") {
		t.Error("Output should contain header comment")
	}
	if !strings.Contains(outStr, "inline comment") {
		t.Error("Output should contain inline comment")
	}
}

func TestMarshalPreserve_PreservesKeyOrder(t *testing.T) {
	tmpDir := createTestDir(t, map[string]string{
		"config.yml": `z: 1
a: 2
m: 3`,
	}, nil)

	tree, err := NewTree(tmpDir)
	assertNoError(t, err)

	opts := &Options{
		PackRoot: tmpDir,
		Mode:     ModePreserve,
		Logger:   logger.Nop(),
	}

	result, err := tree.Marshal(opts)
	assertNoError(t, err)

	node, ok := result.(*yaml.Node)
	if !ok {
		t.Fatalf("Expected *yaml.Node, got %T", result)
	}

	out, err := yaml.Marshal(node)
	assertNoError(t, err)
	outStr := string(out)

	zIdx := strings.Index(outStr, "z:")
	aIdx := strings.Index(outStr, "a:")
	mIdx := strings.Index(outStr, "m:")

	if zIdx == -1 || aIdx == -1 || mIdx == -1 {
		t.Fatalf("Could not find all keys in output:\n%s", outStr)
	}

	if zIdx >= aIdx || aIdx >= mIdx {
		t.Errorf("Expected key order: z, a, m (authored order). Got:\n%s", outStr)
	}
}

func TestMappingHelpers(t *testing.T) {
	m := newMapping()
	if m.Kind != yaml.MappingNode {
		t.Errorf("newMapping().Kind = %v, want MappingNode", m.Kind)
	}
	if m.Tag != "!!map" {
		t.Errorf("newMapping().Tag = %v, want !!map", m.Tag)
	}

	key := newScalarKey("testkey")
	val := &yaml.Node{Kind: yaml.ScalarNode, Value: "testvalue"}
	mappingSet(m, key, val)

	gotVal, ok := mappingGet(m, "testkey")
	if !ok {
		t.Error("mappingGet() returned ok=false for existing key")
	}
	if gotVal.Value != "testvalue" {
		t.Errorf("mappingGet() val = %v, want testvalue", gotVal.Value)
	}

	_, ok = mappingGet(m, "nonexistent")
	if ok {
		t.Error("mappingGet() returned ok=true for non-existent key")
	}
}

func TestMergeMapping(t *testing.T) {
	dst := newMapping()
	mappingSet(dst, newScalarKey("key1"), &yaml.Node{Kind: yaml.ScalarNode, Value: "value1"})

	src := newMapping()
	mappingSet(src, newScalarKey("key1"), &yaml.Node{Kind: yaml.ScalarNode, Value: "value1_updated"})
	mappingSet(src, newScalarKey("key2"), &yaml.Node{Kind: yaml.ScalarNode, Value: "value2"})

	mergeMapping(dst, src, MergeShallow)

	val1, ok := mappingGet(dst, "key1")
	if !ok {
		t.Fatal("key1 should exist after merge")
	}
	if val1.Value != "value1_updated" {
		t.Errorf("key1 value = %v, want value1_updated", val1.Value)
	}

	val2, ok := mappingGet(dst, "key2")
	if !ok {
		t.Fatal("key2 should exist after merge")
	}
	if val2.Value != "value2" {
		t.Errorf("key2 value = %v, want value2", val2.Value)
	}
}

func TestIsEmptyNode(t *testing.T) {
	if !isEmptyNode(nil) {
		t.Error("isEmptyNode(nil) should return true")
	}

	emptyMapping := &yaml.Node{Kind: yaml.MappingNode, Content: []*yaml.Node{}}
	if !isEmptyNode(emptyMapping) {
		t.Error("isEmptyNode(empty mapping) should return true")
	}

	nonEmpty := newMapping()
	mappingSet(nonEmpty, newScalarKey("key"), &yaml.Node{Kind: yaml.ScalarNode, Value: "value"})
	if isEmptyNode(nonEmpty) {
		t.Error("isEmptyNode(non-empty mapping) should return false")
	}

	scalar := &yaml.Node{Kind: yaml.ScalarNode, Value: "value"}
	if isEmptyNode(scalar) {
		t.Error("isEmptyNode(scalar) should return false")
	}
}

func TestMergeMapping_NilInputs(t *testing.T) {
	mergeMapping(nil, newMapping(), MergeShallow)

	dst := newMapping()
	mergeMapping(dst, nil, MergeShallow)

	scalar := &yaml.Node{Kind: yaml.ScalarNode, Value: "test"}
	mergeMapping(dst, scalar, MergeShallow)
	mergeMapping(scalar, dst, MergeShallow)
}

func TestMappingGet_EdgeCases(t *testing.T) {
	val, ok := mappingGet(nil, "key")
	if ok || val != nil {
		t.Error("mappingGet(nil) should return nil, false")
	}

	scalar := &yaml.Node{Kind: yaml.ScalarNode, Value: "test"}
	val, ok = mappingGet(scalar, "key")
	if ok || val != nil {
		t.Error("mappingGet(scalar) should return nil, false")
	}
}

func TestMarshalPreserve_WithIncludes(t *testing.T) {
	tmpDir := createTestDir(t, map[string]string{
		"entities/scripts/test.sh": "#!/bin/bash\necho 'test'",
		"entities/item1.yml":       "# Comment before entity\nentity:\n  id: example1\n  attributes:\n    command: <<include(scripts/test.sh)>>  # inline comment",
	}, nil)

	absDir, err := filepath.Abs(tmpDir)
	assertNoError(t, err)

	opts := &Options{
		EnableIncludes: true,
		PackRoot:       absDir,
		Mode:           ModePreserve,
		Logger:         logger.Nop(),
	}

	tree, err := NewTree(tmpDir)
	assertNoError(t, err)

	testNode := findNodeByName(t, tree, "item1.yml")
	if testNode == nil {
		t.Fatal("Could not find item1.yml node")
	}

	result, err := testNode.marshalLeafPreserve(opts)
	assertNoError(t, err)

	if result == nil {
		t.Fatal("marshalLeafPreserve() returned nil")
	}

	out, err := yaml.Marshal(result)
	assertNoError(t, err)
	outStr := string(out)

	if !strings.Contains(outStr, "echo 'test'") {
		t.Errorf("marshalLeafPreserve() should contain included content. Got: %q", outStr)
	}
	if strings.Contains(outStr, "<<include") {
		t.Error("marshalLeafPreserve() should not contain include directive after processing")
	}
	if !strings.Contains(outStr, "Comment before entity") {
		t.Error("marshalLeafPreserve() should preserve comments")
	}
	if !strings.Contains(outStr, "inline comment") {
		t.Error("marshalLeafPreserve() should preserve inline comments")
	}
}

func TestMarshalPreserve_WithConvertBooleans(t *testing.T) {
	tmpDir := createTestDir(t, map[string]string{
		"config/test.yml": `# Config file
enabled: on  # should become true
disabled: "off"  # quoted, should stay string
name: on_call_service  # not a boolean`,
	}, nil)

	absDir, err := filepath.Abs(tmpDir)
	assertNoError(t, err)

	opts := &Options{
		EnableIncludes:  false,
		PackRoot:        absDir,
		ConvertBooleans: true,
		Mode:            ModePreserve,
		Logger:          logger.Nop(),
	}

	tree, err := NewTree(tmpDir)
	assertNoError(t, err)

	result, err := tree.Marshal(opts)
	assertNoError(t, err)

	node, ok := result.(*yaml.Node)
	if !ok {
		t.Fatalf("Expected *yaml.Node, got %T", result)
	}

	out, err := yaml.Marshal(node)
	assertNoError(t, err)
	outStr := string(out)

	// Verify boolean conversion
	if !strings.Contains(outStr, "enabled: true") {
		t.Errorf("enabled should be converted to true. Got:\n%s", outStr)
	}
	// Verify quoted string is not converted
	if !strings.Contains(outStr, `disabled: "off"`) {
		t.Errorf("disabled should remain as quoted string. Got:\n%s", outStr)
	}
	// Verify non-boolean string is not converted
	if !strings.Contains(outStr, "name: on_call_service") {
		t.Errorf("name should remain as string. Got:\n%s", outStr)
	}
	// Verify comments are preserved
	if !strings.Contains(outStr, "Config file") {
		t.Error("Comments should be preserved")
	}
}

func TestMarshalPreserve_WithoutConvertBooleans(t *testing.T) {
	tmpDir := createTestDir(t, map[string]string{
		"config/test.yml": `# Config file
enabled: on  # should stay as string
disabled: off  # should stay as string`,
	}, nil)

	absDir, err := filepath.Abs(tmpDir)
	assertNoError(t, err)

	opts := &Options{
		EnableIncludes:  false,
		PackRoot:        absDir,
		ConvertBooleans: false,
		Mode:            ModePreserve,
		Logger:          logger.Nop(),
	}

	tree, err := NewTree(tmpDir)
	assertNoError(t, err)

	result, err := tree.Marshal(opts)
	assertNoError(t, err)

	node, ok := result.(*yaml.Node)
	if !ok {
		t.Fatalf("Expected *yaml.Node, got %T", result)
	}

	out, err := yaml.Marshal(node)
	assertNoError(t, err)
	outStr := string(out)

	// Verify booleans remain as strings
	if !strings.Contains(outStr, "enabled: on") {
		t.Errorf("enabled should remain as string 'on' without normalization. Got:\n%s", outStr)
	}
	if !strings.Contains(outStr, "disabled: off") {
		t.Errorf("disabled should remain as string 'off' without normalization. Got:\n%s", outStr)
	}
	// Verify comments are preserved
	if !strings.Contains(outStr, "Config file") {
		t.Error("Comments should be preserved")
	}
}

func TestMarshalPreserve_AtFiles(t *testing.T) {
	tmpDir := createTestDir(t, map[string]string{
		"entities/@common.yml": "# Common config\ntimeout: 30\nretries: 3",
		"entities/item1.yml":   "# Item 1\nentity:\n  id: example1\n  attributes:\n    name: sample name",
	}, nil)

	tree, err := NewTree(tmpDir)
	assertNoError(t, err)

	opts := &Options{
		PackRoot: tmpDir,
		Mode:     ModePreserve,
		Logger:   logger.Nop(),
	}

	result, err := tree.Marshal(opts)
	assertNoError(t, err)

	node, ok := result.(*yaml.Node)
	if !ok {
		t.Fatalf("Expected *yaml.Node, got %T", result)
	}

	out, err := yaml.Marshal(node)
	assertNoError(t, err)
	outStr := string(out)

	// Verify @common.yml content is merged
	if !strings.Contains(outStr, "timeout: 30") {
		t.Error("MarshalPreserve() @common.yml content not merged into entities map")
	}
	if !strings.Contains(outStr, "item1") {
		t.Error("MarshalPreserve() item1.yml not found in entities map")
	}
	// Verify comments are preserved
	if !strings.Contains(outStr, "Common config") {
		t.Error("Comments from @common.yml should be preserved")
	}
	if !strings.Contains(outStr, "Item 1") {
		t.Error("Comments from item1.yml should be preserved")
	}
}

func TestMarshalPreserve_AtDirectories(t *testing.T) {
	tmpDir := createTestDir(t, map[string]string{
		"entities/@group1/item1.yml": "# Item 1\nentity:\n  id: example1\n  attributes:\n    name: first item",
		"entities/@group1/item2.yml": "# Item 2\nentity:\n  id: example2\n  attributes:\n    name: second item",
		"entities/@group2/item3.yml": "# Item 3\nentity:\n  id: example3\n  attributes:\n    name: third item",
		"entities/item4.yml":         "# Item 4\nentity:\n  id: example4\n  attributes:\n    name: fourth item",
	}, nil)

	tree, err := NewTree(tmpDir)
	assertNoError(t, err)

	opts := &Options{
		PackRoot: tmpDir,
		Mode:     ModePreserve,
		Logger:   logger.Nop(),
	}

	result, err := tree.Marshal(opts)
	assertNoError(t, err)

	node, ok := result.(*yaml.Node)
	if !ok {
		t.Fatalf("Expected *yaml.Node, got %T", result)
	}

	out, err := yaml.Marshal(node)
	assertNoError(t, err)
	outStr := string(out)

	// Verify all items are present
	if !strings.Contains(outStr, "item1") {
		t.Error("MarshalPreserve() item1.yml from @group1 not found in entities map")
	}
	if !strings.Contains(outStr, "item2") {
		t.Error("MarshalPreserve() item2.yml from @group1 not found in entities map")
	}
	if !strings.Contains(outStr, "item3") {
		t.Error("MarshalPreserve() item3.yml from @group2 not found in entities map")
	}
	if !strings.Contains(outStr, "item4") {
		t.Error("MarshalPreserve() item4.yml not found in entities map")
	}
	// Verify @directories are not keys
	if strings.Contains(outStr, "@group1:") || strings.Contains(outStr, "@group2:") {
		t.Error("MarshalPreserve() @group directories should not be keys in entities map")
	}
	// Verify comments are preserved
	if !strings.Contains(outStr, "Item 1") || !strings.Contains(outStr, "Item 2") {
		t.Error("Comments should be preserved")
	}
}

func TestMarshalPreserve_RootFile(t *testing.T) {
	tmpDir := createTestDir(t, map[string]string{
		"root_file.yml":   "# Root file comment\nkey1: value1",
		"subdir/file.yml": "# Subdir file comment\nkey2: value2",
	}, nil)

	tree, err := NewTree(tmpDir)
	assertNoError(t, err)

	rootFileNode := findNodeByName(t, tree, "root_file.yml")
	if rootFileNode == nil {
		t.Fatal("Could not find root_file.yml node")
	}

	if !rootFileNode.rootFile() {
		t.Error("rootFile() returned false for root_file.yml")
	}

	opts := &Options{
		PackRoot: tmpDir,
		Mode:     ModePreserve,
		Logger:   logger.Nop(),
	}

	result, err := tree.Marshal(opts)
	assertNoError(t, err)

	node, ok := result.(*yaml.Node)
	if !ok {
		t.Fatalf("Expected *yaml.Node, got %T", result)
	}

	out, err := yaml.Marshal(node)
	assertNoError(t, err)
	outStr := string(out)

	// Verify root file content is merged at top level
	if !strings.Contains(outStr, "key1: value1") {
		t.Error("Root file content should be merged at top level, not nested")
	}
	// Verify subdirectory is nested
	if !strings.Contains(outStr, "subdir:") {
		t.Error("Subdirectory should be nested under 'subdir' key")
	}
	// Verify comments are preserved
	if !strings.Contains(outStr, "Root file comment") {
		t.Error("Root file comments should be preserved")
	}
	if !strings.Contains(outStr, "Subdir file comment") {
		t.Error("Subdir file comments should be preserved")
	}
}

func TestMarshalPreserve_Nested(t *testing.T) {
	tmpDir := createTestDir(t, map[string]string{
		"level1/level2/level3/deep.yml": `# Deep nested file
outer:
  middle:
    inner:
      key: value
      # Nested comment
      another: value2`,
	}, nil)

	tree, err := NewTree(tmpDir)
	assertNoError(t, err)

	opts := &Options{
		PackRoot: tmpDir,
		Mode:     ModePreserve,
		Logger:   logger.Nop(),
	}

	result, err := tree.Marshal(opts)
	assertNoError(t, err)

	node, ok := result.(*yaml.Node)
	if !ok {
		t.Fatalf("Expected *yaml.Node, got %T", result)
	}

	out, err := yaml.Marshal(node)
	assertNoError(t, err)
	outStr := string(out)

	// Verify nested structure is preserved
	if !strings.Contains(outStr, "level1:") {
		t.Error("Nested structure should be preserved")
	}
	if !strings.Contains(outStr, "key: value") {
		t.Error("Deep nested values should be preserved")
	}
	// Verify comments are preserved
	if !strings.Contains(outStr, "Deep nested file") {
		t.Error("Comments should be preserved in nested structures")
	}
	if !strings.Contains(outStr, "Nested comment") {
		t.Error("Nested comments should be preserved")
	}
}

func TestMarshalPreserve_JSONFiles(t *testing.T) {
	tmpDir := createTestDir(t, map[string]string{
		"entities/item1.json": `{"entity": {"id": "example1", "attributes": {"name": "sample name"}}}`,
	}, nil)

	tree, err := NewTree(tmpDir)
	assertNoError(t, err)

	opts := &Options{
		PackRoot: tmpDir,
		Mode:     ModePreserve,
		Logger:   logger.Nop(),
	}

	result, err := tree.Marshal(opts)
	assertNoError(t, err)

	node, ok := result.(*yaml.Node)
	if !ok {
		t.Fatalf("Expected *yaml.Node, got %T", result)
	}

	out, err := yaml.Marshal(node)
	assertNoError(t, err)
	outStr := string(out)

	// Verify JSON content is parsed correctly
	if !strings.Contains(outStr, "entities:") {
		t.Error("MarshalPreserve() result missing 'entities' key")
	}
	if !strings.Contains(outStr, "item1:") {
		t.Error("MarshalPreserve() result missing 'item1' key from JSON file")
	}
	if !strings.Contains(outStr, "example1") {
		t.Error("MarshalPreserve() JSON content id not parsed correctly")
	}
	if !strings.Contains(outStr, "sample name") {
		t.Error("MarshalPreserve() JSON content name not parsed correctly")
	}
}

func TestMarshalPreserve_CommentMerging(t *testing.T) {
	tmpDir := createTestDir(t, map[string]string{
		"file1.yml": `# First file comment
key: value1  # first inline`,
		"file2.yml": `# Second file comment
key: value2  # second inline`,
	}, nil)

	tree, err := NewTree(tmpDir)
	assertNoError(t, err)

	opts := &Options{
		PackRoot: tmpDir,
		Mode:     ModePreserve,
		Logger:   logger.Nop(),
	}

	result, err := tree.Marshal(opts)
	assertNoError(t, err)

	node, ok := result.(*yaml.Node)
	if !ok {
		t.Fatalf("Expected *yaml.Node, got %T", result)
	}

	out, err := yaml.Marshal(node)
	assertNoError(t, err)
	outStr := string(out)

	// Verify later file wins for value
	if !strings.Contains(outStr, "key: value2") {
		t.Errorf("Expected 'key: value2' (later value), got:\n%s", outStr)
	}
	// Verify later file's comment is used (later wins)
	if !strings.Contains(outStr, "second inline") {
		t.Errorf("Expected 'second inline' comment from later file, got:\n%s", outStr)
	}
	// Note: Header comments from both files may be present, but inline comment should be from later file
}

func TestMarshalPreserve_EmptyMaps(t *testing.T) {
	tmpDir := createTestDir(t, map[string]string{
		"has_content/file.yml": "# Has content\nkey: value",
	}, []string{"empty1", "empty2"})

	tree, err := NewTree(tmpDir)
	assertNoError(t, err)

	opts := &Options{
		PackRoot: tmpDir,
		Mode:     ModePreserve,
		Logger:   logger.Nop(),
	}

	result, err := tree.Marshal(opts)
	assertNoError(t, err)

	node, ok := result.(*yaml.Node)
	if !ok {
		t.Fatalf("Expected *yaml.Node, got %T", result)
	}

	out, err := yaml.Marshal(node)
	assertNoError(t, err)
	outStr := string(out)

	// Verify empty directories are not included
	if strings.Contains(outStr, "empty1:") || strings.Contains(outStr, "empty2:") {
		t.Error("MarshalPreserve() should not contain empty directory keys")
	}
	// Verify content directory is included
	if !strings.Contains(outStr, "has_content:") {
		t.Error("MarshalPreserve() should contain 'has_content' key")
	}
	// Verify comments are preserved
	if !strings.Contains(outStr, "Has content") {
		t.Error("Comments should be preserved")
	}
}

func TestMarshalPreserve_RendersToYAML(t *testing.T) {
	tmpDir := createTestDir(t, map[string]string{
		"sub_dir/sub_dir_file.yml": "# Sub directory file\nentity:\n  id: example1\n  attributes:\n    name: sample name",
	}, []string{"empty_dir"})

	tree, err := NewTree(tmpDir)
	assertNoError(t, err)

	opts := &Options{
		PackRoot: tmpDir,
		Mode:     ModePreserve,
		Logger:   logger.Nop(),
	}

	result, err := tree.Marshal(opts)
	assertNoError(t, err)

	node, ok := result.(*yaml.Node)
	if !ok {
		t.Fatalf("Expected *yaml.Node, got %T", result)
	}

	out, err := yaml.Marshal(node)
	assertNoError(t, err)
	if len(out) == 0 {
		t.Error("yaml.Marshal() returned empty output")
	}

	outStr := string(out)
	if !strings.Contains(outStr, "sub_dir:") {
		t.Error("MarshalPreserve() result missing 'sub_dir' key")
	}
	if strings.Contains(outStr, "empty_dir:") {
		t.Error("MarshalPreserve() result should not contain 'empty_dir' key (empty directories are ignored)")
	}
	// Verify comments are preserved
	if !strings.Contains(outStr, "Sub directory file") {
		t.Error("Comments should be preserved")
	}
}

func TestMarshalPreserve_EmptyAtDirectory(t *testing.T) {
	tmpDir := createTestDir(t, map[string]string{
		"entities/item1.yml": "# Item 1\nentity:\n  id: example1\n  attributes:\n    name: sample name",
	}, []string{"entities/@group1"})

	tree, err := NewTree(tmpDir)
	assertNoError(t, err)

	opts := &Options{
		PackRoot: tmpDir,
		Mode:     ModePreserve,
		Logger:   logger.Nop(),
	}

	result, err := tree.Marshal(opts)
	assertNoError(t, err)

	node, ok := result.(*yaml.Node)
	if !ok {
		t.Fatalf("Expected *yaml.Node, got %T", result)
	}

	entitiesVal, ok := mappingGet(node, "entities")
	if !ok {
		t.Fatal("MarshalPreserve() result missing 'entities' key")
	}

	out, err := yaml.Marshal(node)
	assertNoError(t, err)
	outStr := string(out)

	if strings.Contains(outStr, "@group1:") || strings.Contains(outStr, "group1:") {
		t.Error("MarshalPreserve() empty @group1 directory should not create a key")
	}
	if !strings.Contains(outStr, "item1:") {
		t.Error("MarshalPreserve() item1.yml not found in entities map")
	}
	// Verify comments are preserved
	if !strings.Contains(outStr, "Item 1") {
		t.Error("Comments should be preserved")
	}

	// Verify entitiesVal is not nil
	if entitiesVal == nil {
		t.Error("entities value should not be nil")
	}
}

func TestMarshalPreserve_NestedAtDirectories(t *testing.T) {
	tmpDir := createTestDir(t, map[string]string{
		"entities/@group1/@group2/item1.yml": "# Item 1\nentity:\n  id: example1\n  attributes:\n    name: first item",
		"entities/@group1/@group2/item2.yml": "# Item 2\nentity:\n  id: example2\n  attributes:\n    name: second item",
		"entities/@group1/item3.yml":         "# Item 3\nentity:\n  id: example3\n  attributes:\n    name: third item",
	}, nil)

	tree, err := NewTree(tmpDir)
	assertNoError(t, err)

	opts := &Options{
		PackRoot: tmpDir,
		Mode:     ModePreserve,
		Logger:   logger.Nop(),
	}

	result, err := tree.Marshal(opts)
	assertNoError(t, err)

	node, ok := result.(*yaml.Node)
	if !ok {
		t.Fatalf("Expected *yaml.Node, got %T", result)
	}

	out, err := yaml.Marshal(node)
	assertNoError(t, err)
	outStr := string(out)

	// Verify all items are present
	if !strings.Contains(outStr, "item1:") {
		t.Error("MarshalPreserve() item1.yml from @group1/@group2 not found in entities map")
	}
	if !strings.Contains(outStr, "item2:") {
		t.Error("MarshalPreserve() item2.yml from @group1/@group2 not found in entities map")
	}
	if !strings.Contains(outStr, "item3:") {
		t.Error("MarshalPreserve() item3.yml from @group1 not found in entities map")
	}
	// Verify @directories are not keys
	if strings.Contains(outStr, "@group1:") || strings.Contains(outStr, "@group2:") {
		t.Error("MarshalPreserve() @group directories should not be keys in entities map")
	}
	if strings.Contains(outStr, "group1:") || strings.Contains(outStr, "group2:") {
		t.Error("MarshalPreserve() group directories (without @) should not be keys in entities map")
	}
	// Verify comments are preserved
	if !strings.Contains(outStr, "Item 1") || !strings.Contains(outStr, "Item 2") || !strings.Contains(outStr, "Item 3") {
		t.Error("Comments should be preserved")
	}
}

func TestMarshalPreserve_ShallowMerge(t *testing.T) {
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
		Mode:          ModePreserve,
		MergeStrategy: MergeShallow,
		Logger:        logger.Nop(),
	}

	tree, err := NewTree(tmpDir)
	assertNoError(t, err)

	result, err := tree.Marshal(opts)
	assertNoError(t, err)

	node, ok := result.(*yaml.Node)
	if !ok {
		t.Fatalf("Expected *yaml.Node, got %T", result)
	}

	out, err := yaml.Marshal(node)
	assertNoError(t, err)
	outStr := string(out)

	// In shallow merge, later file completely replaces earlier one
	// So setting1 and setting2 should be gone, setting3 should exist
	if strings.Contains(outStr, "setting1:") {
		t.Error("Shallow merge should replace entire map - setting1 should not exist")
	}
	if strings.Contains(outStr, "setting2:") {
		t.Error("Shallow merge should replace entire map - setting2 should not exist")
	}
	if !strings.Contains(outStr, "setting3: value3") {
		t.Error("Shallow merge should include values from later file")
	}

	// Nested map should also be completely replaced
	if strings.Contains(outStr, "a: 1") {
		t.Error("Shallow merge should replace nested map - 'a' should not exist")
	}
	if strings.Contains(outStr, "b: 2") {
		t.Error("Shallow merge should replace nested map - 'b' should not exist")
	}
	if !strings.Contains(outStr, "c: 3") {
		t.Error("Shallow merge should include nested values from later file")
	}
}

func TestMarshalPreserve_DeepMerge(t *testing.T) {
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
		Mode:          ModePreserve,
		MergeStrategy: MergeDeep,
		Logger:        logger.Nop(),
	}

	tree, err := NewTree(tmpDir)
	assertNoError(t, err)

	result, err := tree.Marshal(opts)
	assertNoError(t, err)

	node, ok := result.(*yaml.Node)
	if !ok {
		t.Fatalf("Expected *yaml.Node, got %T", result)
	}

	out, err := yaml.Marshal(node)
	assertNoError(t, err)
	outStr := string(out)

	// In deep merge, values from both files should exist
	if !strings.Contains(outStr, "setting1: value1") {
		t.Error("Deep merge should preserve setting1 from base file")
	}
	if !strings.Contains(outStr, "setting2: value2") {
		t.Error("Deep merge should preserve setting2 from base file")
	}
	if !strings.Contains(outStr, "setting3: value3") {
		t.Error("Deep merge should include setting3 from override file")
	}

	// Nested map should be merged recursively
	if !strings.Contains(outStr, "a: 1") {
		t.Error("Deep merge should preserve 'a' from base file")
	}
	if !strings.Contains(outStr, "b: 2") {
		t.Error("Deep merge should preserve 'b' from base file")
	}
	if !strings.Contains(outStr, "c: 3") {
		t.Error("Deep merge should include 'c' from override file")
	}
}
