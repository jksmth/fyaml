// Package filetree provides filesystem traversal for FYAML packing.
//
// This file contains code vendored and adapted from CircleCI CLI:
// https://github.com/CircleCI-Public/circleci-cli/blob/main/filetree/filetree.go
//
// Original copyright: Copyright (c) 2018 Circle Internet Services, Inc.
// Original license: MIT License
//
// Modifications:
// - Removed SpecialCase global variable
// - Added deterministic key sorting for reproducible output
// - Refactored for internal use
// - Added optional include file processing
package filetree

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"github.com/jksmth/fyaml/internal/include"
	"github.com/mitchellh/mapstructure"
	"gopkg.in/yaml.v3"
)

// IncludeOptions controls include processing behavior.
type IncludeOptions struct {
	Enabled  bool
	PackRoot string // Absolute path to pack root (confinement boundary)
}

// Node represents a node in the filetree
type Node struct {
	FullPath string
	Info     os.FileInfo
	Children []*Node
	Parent   *Node
}

// PathNodes is a map of filepaths to tree nodes with ordered path keys.
type PathNodes struct {
	Map  map[string]*Node
	Keys []string
}

// NewTree creates a new filetree starting at the root.
// It collects all YAML files and directories, skipping dotfiles and dotfolders.
func NewTree(rootPath string, opts *IncludeOptions) (*Node, error) {
	absRootPath, err := filepath.Abs(rootPath)
	if err != nil {
		return nil, err
	}

	pathNodes, err := collectNodes(absRootPath)
	if err != nil {
		return nil, err
	}

	// Sort keys for deterministic ordering
	sort.Strings(pathNodes.Keys)

	rootNode := buildTree(absRootPath, pathNodes)

	return rootNode, err
}

func collectNodes(absRootPath string) (PathNodes, error) {
	pathNodes := PathNodes{}
	pathNodes.Map = make(map[string]*Node)
	pathNodes.Keys = []string{}

	err := filepath.Walk(absRootPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip any dotfolders
		if absRootPath != path && dotfolder(info) {
			return filepath.SkipDir
		}

		fp, err := filepath.Abs(path)
		if err != nil {
			return err
		}

		pathNodes.Keys = append(pathNodes.Keys, path)
		pathNodes.Map[path] = &Node{
			FullPath: fp,
			Info:     info,
			Children: make([]*Node, 0),
		}
		return nil
	})

	return pathNodes, err
}

func buildTree(absRootPath string, pathNodes PathNodes) *Node {
	var rootNode *Node

	for _, path := range pathNodes.Keys {
		node := pathNodes.Map[path]
		// skip dotfile nodes that aren't the root path
		if absRootPath != path && node.Info.Mode().IsRegular() {
			if dotfile(node.Info) || !isYaml(node.Info) {
				continue
			}
		}
		parentPath := filepath.Dir(path)
		parent, exists := pathNodes.Map[parentPath]
		if exists {
			node.Parent = parent
			parent.Children = append(parent.Children, node)
		} else {
			rootNode = node
		}
	}

	// Sort children for deterministic ordering
	if rootNode != nil {
		sortChildren(rootNode)
	}

	return rootNode
}

func sortChildren(node *Node) {
	// Sort children by FullPath for deterministic ordering
	sort.Slice(node.Children, func(i, j int) bool {
		return node.Children[i].FullPath < node.Children[j].FullPath
	})

	// Recursively sort all children
	for _, child := range node.Children {
		sortChildren(child)
	}
}

func dotfile(info os.FileInfo) bool {
	re := regexp.MustCompile(`^\..+`)
	return re.MatchString(info.Name())
}

func dotfolder(info os.FileInfo) bool {
	return info.IsDir() && dotfile(info)
}

func isYaml(info os.FileInfo) bool {
	re := regexp.MustCompile(`.+\.(yml|yaml|json)$`)
	return re.MatchString(info.Name())
}

func (n *Node) basename() string {
	return n.Info.Name()
}

func (n *Node) name() string {
	return strings.TrimSuffix(n.basename(), filepath.Ext(n.basename()))
}

func (n *Node) root() *Node {
	root := n.Parent
	if root == nil {
		return n
	}
	for root.Parent != nil {
		root = root.Parent
	}
	return root
}

func (n *Node) rootFile() bool {
	return n.Info.Mode().IsRegular() && n.root() == n.Parent
}

func (n *Node) specialCase() bool {
	re := regexp.MustCompile(`^@.*\.(yml|yaml|json)$`)
	return re.MatchString(n.basename())
}

// MarshalYAML serializes the tree into YAML.
// Implements yaml.Marshaler interface (called by yaml.Marshal).
func (n *Node) MarshalYAML() (interface{}, error) {
	return n.MarshalYAMLWithOptions(nil)
}

// MarshalYAMLWithOptions serializes the tree into YAML with include options.
// If opts is nil, includes are disabled.
func (n *Node) MarshalYAMLWithOptions(opts *IncludeOptions) (interface{}, error) {
	if len(n.Children) == 0 {
		return n.marshalLeaf(opts)
	}
	return n.marshalParent(opts)
}

func (n *Node) marshalLeaf(opts *IncludeOptions) (interface{}, error) {
	var content interface{}

	if n.Info.IsDir() {
		return content, nil
	}
	if !isYaml(n.Info) {
		return content, nil
	}

	buf, err := os.ReadFile(n.FullPath)
	if err != nil {
		return content, err
	}

	// If includes are enabled, parse to yaml.Node to allow walking and replacing
	if opts != nil && opts.Enabled {
		var node yaml.Node
		if err := yaml.Unmarshal(buf, &node); err != nil {
			return content, err
		}

		// Process include directives relative to the file's directory
		baseDir := filepath.Dir(n.FullPath)
		if err := include.InlineIncludes(&node, baseDir, opts.PackRoot); err != nil {
			return content, err
		}

		// Decode the processed node back to interface{}
		if err := node.Decode(&content); err != nil {
			return content, err
		}
		return content, nil
	}

	err = yaml.Unmarshal(buf, &content)
	return content, err
}

// IsEmptyContent checks if a value is nil or an empty map.
// Used to skip directories/files with no YAML content.
func IsEmptyContent(v interface{}) bool {
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
// This is adapted from the CircleCI CLI implementation.
// Per CircleCI behavior, later values overwrite earlier values (no collision errors).
func mergeTree(trees ...interface{}) map[string]interface{} {
	result := make(map[string]interface{})
	for _, tree := range trees {
		if tree == nil {
			continue
		}

		kvp := make(map[string]interface{})
		if err := mapstructure.Decode(tree, &kvp); err != nil {
			panic(err)
		}
		for k, v := range kvp {
			result[k] = v
		}
	}
	return result
}

func (n *Node) marshalParent(opts *IncludeOptions) (interface{}, error) {
	subtree := map[string]interface{}{}

	for _, child := range n.Children {
		c, err := child.MarshalYAMLWithOptions(opts)
		if err != nil {
			return nil, err
		}

		// Skip directories/files with no YAML content
		if IsEmptyContent(c) {
			continue
		}

		switch c.(type) {
		case map[string]interface{}, map[interface{}]interface{}:
			if child.rootFile() {
				subtree = mergeTree(subtree, c)
			} else if child.specialCase() {
				subtree = mergeTree(subtree, subtree[child.Parent.name()], c)
			} else {
				subtree[child.name()] = mergeTree(subtree[child.name()], c)
			}
		default:
			return nil, fmt.Errorf("expected a map, got a `%T` which is not supported at this time for \"%s\"", c, child.FullPath)
		}
	}

	// Return nil for directories with no YAML content
	if len(subtree) == 0 {
		return nil, nil
	}

	return sortMapKeys(subtree), nil
}

// sortMapKeys recursively sorts all map keys for deterministic output
func sortMapKeys(m map[string]interface{}) map[string]interface{} {
	result := make(map[string]interface{})
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, k := range keys {
		v := m[k]
		// Recursively sort nested maps
		if nestedMap, ok := v.(map[string]interface{}); ok {
			v = sortMapKeys(nestedMap)
		}
		result[k] = v
	}
	return result
}
