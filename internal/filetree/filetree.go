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
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"github.com/jksmth/fyaml/internal/include"
	"github.com/jksmth/fyaml/internal/logger"
	"github.com/mitchellh/mapstructure"
	"go.yaml.in/yaml/v4"
)

// Options controls how the filetree is processed during marshaling.
type Options struct {
	// Include processing
	EnableIncludes bool   // Process <<include(file)>> directives
	PackRoot       string // Absolute path to pack root (confinement boundary)

	// YAML processing
	ConvertBooleans bool // Convert unquoted YAML 1.1 booleans to true/false

	// Logging
	Logger logger.Logger // Logger for verbose output (nil-safe: defaults to Nop())
}

// log returns the logger, defaulting to Nop() if nil.
func (o *Options) log() logger.Logger {
	if o == nil || o.Logger == nil {
		return logger.Nop()
	}
	return o.Logger
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
func NewTree(rootPath string) (*Node, error) {
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

// isYaml checks if a file is a supported YAML/JSON file.
// Note: This project is YAML-first; JSON support is provided as a convenience
// since JSON is a subset of YAML.
func isYaml(info os.FileInfo) bool {
	re := regexp.MustCompile(`.+\.(yml|yaml|json)$`)
	return re.MatchString(info.Name())
}

// normalizeYAML11Booleans recursively converts unquoted YAML 1.1 boolean
// strings to canonical boolean values. Quoted strings are left unchanged.
// Per YAML 1.1 spec: https://yaml.org/type/bool.html
func normalizeYAML11Booleans(n *yaml.Node) {
	if n == nil {
		return
	}

	// Only normalize unquoted scalar nodes that aren't already booleans
	if n.Kind == yaml.ScalarNode && n.Style == 0 {
		// Skip if already tagged as boolean (YAML 1.2 true/false)
		if n.ShortTag() == "!!bool" {
			return
		}

		// Convert YAML 1.1 boolean values to canonical form
		switch n.Value {
		case "y", "Y", "yes", "Yes", "YES", "on", "On", "ON":
			n.Value = "true"
			n.Tag = "!!bool"
		case "n", "N", "no", "No", "NO", "off", "Off", "OFF":
			n.Value = "false"
			n.Tag = "!!bool"
		}
	}

	// Recurse into children
	for _, child := range n.Content {
		normalizeYAML11Booleans(child)
	}
}

// formatYAMLError formats a yaml error with position information if available.
// Returns a formatted error string with file path and optional line/column info.
func formatYAMLError(err error, filePath string) error {
	if err == nil {
		return nil
	}

	// Check for ParserError (syntax errors)
	var parserErr *yaml.ParserError
	if errors.As(err, &parserErr) {
		return fmt.Errorf("YAML/JSON syntax error in %s:%d:%d: %s",
			filePath, parserErr.Line, parserErr.Column, parserErr.Message)
	}

	// Check for TypeError (type conversion errors)
	var typeErr *yaml.TypeError
	if errors.As(err, &typeErr) {
		var errMsgs []string
		for _, e := range typeErr.Errors {
			if e.Line > 0 && e.Column > 0 {
				errMsgs = append(errMsgs, fmt.Sprintf("  line %d:%d: %v",
					e.Line, e.Column, e.Err))
			} else {
				errMsgs = append(errMsgs, fmt.Sprintf("  %v", e.Err))
			}
		}
		return fmt.Errorf("YAML/JSON type errors in %s:\n%s",
			filePath, strings.Join(errMsgs, "\n"))
	}

	// Fallback to generic error with file path
	return fmt.Errorf("failed to parse YAML/JSON in %s: %w", filePath, err)
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

func (n *Node) specialCaseDirectory() bool {
	return n.Info.IsDir() && strings.HasPrefix(n.basename(), "@")
}

// MarshalYAML serializes the tree into YAML.
// Implements yaml.Marshaler interface (called by yaml.Marshal).
func (n *Node) MarshalYAML() (interface{}, error) {
	return n.Marshal(nil)
}

// Marshal serializes the tree into YAML with processing options.
// If opts is nil, processing features are disabled.
func (n *Node) Marshal(opts *Options) (interface{}, error) {
	if len(n.Children) == 0 {
		return n.marshalLeaf(opts)
	}
	return n.marshalParent(opts)
}

func (n *Node) marshalLeaf(opts *Options) (interface{}, error) {
	var content interface{}

	if n.Info.IsDir() {
		return content, nil
	}
	if !isYaml(n.Info) {
		return content, nil
	}

	opts.log().Debugf("Processing: %s", n.FullPath)

	buf, err := os.ReadFile(n.FullPath)
	if err != nil {
		return content, fmt.Errorf("failed to read file %s: %w", n.FullPath, err)
	}

	// Always parse to yaml.Node first to support Style-aware processing
	var node yaml.Node
	if err := yaml.Unmarshal(buf, &node); err != nil {
		return content, formatYAMLError(err, n.FullPath)
	}

	// Process all include mechanisms if enabled (!include, !include-text, <<include()>>)
	if opts != nil && opts.EnableIncludes {
		baseDir := filepath.Dir(n.FullPath)
		if err := include.ProcessIncludes(&node, baseDir, opts.PackRoot); err != nil {
			return content, fmt.Errorf("failed to process includes in %s: %w", n.FullPath, err)
		}
	}

	// Convert YAML 1.1 booleans if enabled
	if opts != nil && opts.ConvertBooleans {
		normalizeYAML11Booleans(&node)
	}

	// Decode the processed node to interface{}
	if err := node.Decode(&content); err != nil {
		return content, formatYAMLError(err, n.FullPath)
	}
	return content, nil
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

func (n *Node) marshalParent(opts *Options) (interface{}, error) {
	subtree := map[string]interface{}{}

	for _, child := range n.Children {
		c, err := child.Marshal(opts)
		if err != nil {
			// Error already includes file path from marshalLeaf, just propagate it
			return nil, err
		}

		// Skip directories/files with no YAML content
		if IsEmptyContent(c) {
			continue
		}

		switch c.(type) {
		case map[string]interface{}, map[interface{}]interface{}:
			var err error
			if child.rootFile() {
				subtree, err = mergeTree(subtree, c)
			} else if child.specialCaseDirectory() {
				subtree, err = mergeTree(subtree, c)
			} else if child.specialCase() {
				subtree, err = mergeTree(subtree, subtree[child.Parent.name()], c)
			} else {
				subtree[child.name()], err = mergeTree(subtree[child.name()], c)
			}
			if err != nil {
				return nil, fmt.Errorf("failed to merge tree for %s: %w", child.FullPath, err)
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
