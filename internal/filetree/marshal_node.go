// Package filetree provides filesystem traversal for FYAML packing.
package filetree

import (
	"fmt"

	"go.yaml.in/yaml/v4"
)

// marshal_node.go contains preserve mode marshaling implementations (works with yaml.Node)
// and yaml.Node utility functions.

// isEmptyNode checks if a yaml.Node is nil or an empty mapping.
func isEmptyNode(node *yaml.Node) bool {
	if node == nil {
		return true
	}
	if node.Kind == yaml.MappingNode {
		return len(node.Content) == 0
	}
	return false
}

// newMapping creates an empty yaml.Node of kind MappingNode.
func newMapping() *yaml.Node {
	return &yaml.Node{Kind: yaml.MappingNode, Tag: "!!map"}
}

// mappingGet finds a key in a mapping node and returns its value node.
func mappingGet(m *yaml.Node, key string) (val *yaml.Node, ok bool) {
	if m == nil || m.Kind != yaml.MappingNode {
		return nil, false
	}
	for i := 0; i+1 < len(m.Content); i += 2 {
		k := m.Content[i]
		if k.Kind == yaml.ScalarNode && k.Value == key {
			return m.Content[i+1], true
		}
	}
	return nil, false
}

// mappingSet adds a key-value pair to a mapping node.
func mappingSet(m *yaml.Node, keyNode, valNode *yaml.Node) {
	m.Content = append(m.Content, keyNode, valNode)
}

// newScalarKey creates a scalar key node from a string.
func newScalarKey(s string) *yaml.Node {
	return &yaml.Node{Kind: yaml.ScalarNode, Tag: "!!str", Value: s}
}

// mergeMapping merges src mapping node into dst mapping node.
// Later (src) values overwrite earlier (dst) values - "later wins" semantics.
func mergeMapping(dst, src *yaml.Node, strategy MergeStrategy) {
	if src == nil || dst == nil {
		return
	}
	if dst.Kind != yaml.MappingNode || src.Kind != yaml.MappingNode {
		return
	}

	// Build index of dst keys -> key node position
	dstIndex := make(map[string]int, len(dst.Content)/2)
	for i := 0; i+1 < len(dst.Content); i += 2 {
		dstIndex[dst.Content[i].Value] = i
	}

	// Walk src in authored order
	for i := 0; i+1 < len(src.Content); i += 2 {
		srcKey := src.Content[i]
		srcVal := src.Content[i+1]
		k := srcKey.Value

		dstKeyPos, exists := dstIndex[k]
		if !exists {
			// New key - add it
			mappingSet(dst, srcKey, srcVal)
			dstIndex[k] = len(dst.Content) - 2
			continue
		}

		// Existing key - deep merge if both are maps, otherwise replace
		if strategy == MergeDeep {
			dstVal := dst.Content[dstKeyPos+1]
			if dstVal.Kind == yaml.MappingNode && srcVal.Kind == yaml.MappingNode {
				mergeMapping(dstVal, srcVal, strategy)
				continue
			}
		}
		dst.Content[dstKeyPos+1] = srcVal
	}
}

// marshalLeafNode marshals a leaf node in preserve mode (returns *yaml.Node).
func (n *Node) marshalLeafNode(opts *Options) (*yaml.Node, error) {
	return n.parseYAMLFile(opts)
}

// marshalParentNode marshals a parent node in preserve mode (returns *yaml.Node).
func (n *Node) marshalParentNode(opts *Options) (*yaml.Node, error) {
	subtree := newMapping()

	// Get merge strategy from options (default to shallow)
	strategy := MergeShallow
	if opts != nil && opts.MergeStrategy == MergeDeep {
		strategy = MergeDeep
	}

	for _, child := range n.Children {
		// Use child.Marshal() to ensure anchor registry is available and processing is consistent
		// This matches the approach in canonical mode (marshalParentMap)
		marshaled, err := child.Marshal(opts)
		if err != nil {
			return nil, err
		}

		// In preserve mode, Marshal returns *yaml.Node
		c, ok := marshaled.(*yaml.Node)
		if !ok {
			return nil, fmt.Errorf("expected *yaml.Node in preserve mode, got %T for \"%s\"", marshaled, child.FullPath)
		}

		if isEmptyNode(c) {
			continue
		}
		if c.Kind != yaml.MappingNode {
			return nil, fmt.Errorf("expected a map, got a `%v` which is not supported at this time for \"%s\"", c.Kind, child.FullPath)
		}

		if child.rootFile() || child.specialCaseDirectory() || child.specialCase() {
			mergeMapping(subtree, c, strategy)
		} else {
			childName := child.name()
			dv, ok := mappingGet(subtree, childName)
			if !ok {
				dv = newMapping()
				mappingSet(subtree, newScalarKey(childName), dv)
			}
			mergeMapping(dv, c, strategy)
		}
	}

	if len(subtree.Content) == 0 {
		return nil, nil
	}

	return subtree, nil
}
