// Package filetree provides filesystem traversal for FYAML packing.
package filetree

import (
	"fmt"

	"go.yaml.in/yaml/v4"
)

// marshal_preserve.go contains preserve mode marshaling (authored order, with comments).

func (n *Node) marshalLeafPreserve(opts *Options) (*yaml.Node, error) {
	return n.parseYAMLFile(opts)
}

func (n *Node) marshalParentPreserve(opts *Options) (*yaml.Node, error) {
	subtree := newMapping()

	for _, child := range n.Children {
		var c *yaml.Node
		var err error
		if len(child.Children) == 0 {
			c, err = child.marshalLeafPreserve(opts)
		} else {
			c, err = child.marshalParentPreserve(opts)
		}
		if err != nil {
			return nil, err
		}
		if isEmptyNode(c) {
			continue
		}
		if c.Kind != yaml.MappingNode {
			return nil, fmt.Errorf("expected a map, got a `%v` which is not supported at this time for \"%s\"", c.Kind, child.FullPath)
		}

		if child.rootFile() || child.specialCaseDirectory() || child.specialCase() {
			mergeMapping(subtree, c)
		} else {
			childName := child.name()
			dv, ok := mappingGet(subtree, childName)
			if !ok {
				dv = newMapping()
				mappingSet(subtree, newScalarKey(childName), dv)
			}
			mergeMapping(dv, c)
		}
	}

	if len(subtree.Content) == 0 {
		return nil, nil
	}

	return subtree, nil
}

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
func mergeMapping(dst, src *yaml.Node) {
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

		// Existing key - later wins, just replace the value
		dst.Content[dstKeyPos+1] = srcVal
	}
}
