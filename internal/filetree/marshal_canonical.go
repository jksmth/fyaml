// Package filetree provides filesystem traversal for FYAML packing.
package filetree

import (
	"fmt"
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
	subtree := map[interface{}]interface{}{}

	// Get merge strategy from options (default to shallow)
	strategy := MergeShallow
	if opts != nil && opts.MergeStrategy == MergeDeep {
		strategy = MergeDeep
	}

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
			subtree, err = mergeTree(subtree, c, strategy)
		} else {
			subtree[child.name()], err = mergeTree(subtree[child.name()], c, strategy)
		}
		if err != nil {
			return nil, fmt.Errorf("failed to merge tree for %s: %w", child.FullPath, err)
		}
	}

	if len(subtree) == 0 {
		return nil, nil
	}

	return subtree, nil
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

// mergeTree merges dst and src maps based on strategy, returning the merged result.
//
// Note: We don't need to sort keys here because yaml.v4's encoder automatically sorts
// map keys during encoding. For map[interface{}]interface{}, it sorts by type order
// (bool < int < string) then by value within each type. This provides deterministic
// canonical output without explicit sorting.
// Per CircleCI behavior, later values overwrite earlier values (no collision errors).
func mergeTree(dst, src interface{}, strategy MergeStrategy) (map[interface{}]interface{}, error) {
	// Handle nil dst - start with empty map
	result := toInterfaceMap(dst)
	if result == nil {
		result = make(map[interface{}]interface{})
	}

	// Handle nil src - just return dst
	if src == nil {
		return result, nil
	}

	srcMap := toInterfaceMap(src)
	if srcMap == nil {
		return nil, fmt.Errorf("expected map, got %T", src)
	}

	// Merge src into result
	for k, v := range srcMap {
		if strategy == MergeDeep {
			dstVal := toInterfaceMap(result[k])
			srcVal := toInterfaceMap(v)
			if dstVal != nil && srcVal != nil {
				// Recursive merge - error can't happen since types are already validated
				merged, _ := mergeTree(dstVal, srcVal, strategy)
				result[k] = merged
				continue
			}
		}
		result[k] = v
	}

	return result, nil
}

// toInterfaceMap converts either map type to map[interface{}]interface{}, or returns nil.
func toInterfaceMap(v interface{}) map[interface{}]interface{} {
	switch m := v.(type) {
	case map[interface{}]interface{}:
		return m
	case map[string]interface{}:
		result := make(map[interface{}]interface{}, len(m))
		for k, val := range m {
			result[k] = val
		}
		return result
	}
	return nil
}
