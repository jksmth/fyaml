// Package filetree provides filesystem traversal for FYAML packing.
package filetree

import (
	"fmt"
	"sort"

	"github.com/mitchellh/mapstructure"
)

// marshal_canonical.go contains canonical mode marshaling (sorted keys, no comments).

func (n *Node) marshalLeaf(opts *Options) (interface{}, error) {
	node, err := n.parseYAMLFile(opts)
	if err != nil {
		return nil, err
	}
	if node == nil {
		return nil, nil
	}

	// Decode to interface{} (loses comments and key order)
	var content interface{}
	if err := node.Decode(&content); err != nil {
		return nil, formatYAMLError(err, n.FullPath)
	}
	return content, nil
}

func (n *Node) marshalParent(opts *Options) (interface{}, error) {
	subtree := map[string]interface{}{}

	for _, child := range n.Children {
		c, err := child.Marshal(opts)
		if err != nil {
			return nil, err
		}

		if isEmptyContent(c) {
			continue
		}

		_, isStringMap := c.(map[string]interface{})
		_, isInterfaceMap := c.(map[interface{}]interface{})
		if !isStringMap && !isInterfaceMap {
			return nil, fmt.Errorf("expected a map, got a `%T` which is not supported at this time for \"%s\"", c, child.FullPath)
		}

		if child.rootFile() || child.specialCaseDirectory() || child.specialCase() {
			subtree, err = mergeTree(subtree, c)
		} else {
			subtree[child.name()], err = mergeTree(subtree[child.name()], c)
		}
		if err != nil {
			return nil, fmt.Errorf("failed to merge tree for %s: %w", child.FullPath, err)
		}
	}

	if len(subtree) == 0 {
		return nil, nil
	}

	return sortMapKeys(subtree), nil
}

// isEmptyContent checks if a value is nil or an empty map.
func isEmptyContent(v interface{}) bool {
	if v == nil {
		return true
	}
	switch m := v.(type) {
	case map[string]interface{}:
		return len(m) == 0
	case map[interface{}]interface{}:
		return len(m) == 0
	}
	return false
}

// mergeTree merges multiple interface{} values into a single map[string]interface{}.
// Per CircleCI behavior, later values overwrite earlier values (no collision errors).
func mergeTree(trees ...interface{}) (map[string]interface{}, error) {
	result := make(map[string]interface{})
	for _, tree := range trees {
		if tree == nil {
			continue
		}

		kvp := make(map[string]interface{})
		if err := mapstructure.Decode(tree, &kvp); err != nil {
			return nil, fmt.Errorf("failed to decode tree structure: %w", err)
		}
		for k, v := range kvp {
			result[k] = v
		}
	}
	return result, nil
}

// sortMapKeys recursively sorts all map keys for deterministic output.
func sortMapKeys(m map[string]interface{}) map[string]interface{} {
	result := make(map[string]interface{})
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, k := range keys {
		v := m[k]
		if nestedMap, ok := v.(map[string]interface{}); ok {
			v = sortMapKeys(nestedMap)
		}
		result[k] = v
	}
	return result
}
