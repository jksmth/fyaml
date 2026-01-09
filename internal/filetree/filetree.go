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
// - Added cross-platform path normalization
package filetree

import (
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

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

// collectNodes walks absRootPath and returns nodes keyed by a canonical path.
// Canonical key = normalized absolute path (forward slashes for cross-platform consistency).
func collectNodes(absRootPath string) (PathNodes, error) {
	pathNodes := PathNodes{
		Map:  make(map[string]*Node),
		Keys: []string{},
	}

	err := filepath.Walk(absRootPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip dotfolders (but not the root itself)
		if absRootPath != path && dotfolder(info) {
			return filepath.SkipDir
		}

		fp, err := filepath.Abs(path)
		if err != nil {
			return err
		}

		key := filepath.ToSlash(fp) // canonical key for cross-platform determinism

		pathNodes.Keys = append(pathNodes.Keys, key)
		pathNodes.Map[key] = &Node{
			FullPath: fp, // keep native absolute path for I/O
			Info:     info,
			Children: make([]*Node, 0),
		}

		return nil
	})

	return pathNodes, err
}

func buildTree(absRootPath string, pathNodes PathNodes) *Node {
	var rootNode *Node

	rootKey := filepath.ToSlash(absRootPath) // canonical root key

	for _, key := range pathNodes.Keys {
		node := pathNodes.Map[key]

		// Skip dotfiles + non-YAML regular files, except at the root itself
		if rootKey != key && node.Info.Mode().IsRegular() {
			if dotfile(node.Info) || !isYaml(node.Info) {
				continue
			}
		}

		parentKey := filepath.ToSlash(filepath.Dir(node.FullPath))
		if parent, exists := pathNodes.Map[parentKey]; exists {
			node.Parent = parent
			parent.Children = append(parent.Children, node)
		} else {
			// Only true for the actual root directory node
			rootNode = node
		}
	}

	if rootNode != nil {
		sortChildren(rootNode)
	}

	return rootNode
}

func sortChildren(node *Node) {
	// Deterministic across OS: compare canonical absolute paths
	sort.Slice(node.Children, func(i, j int) bool {
		return filepath.ToSlash(node.Children[i].FullPath) < filepath.ToSlash(node.Children[j].FullPath)
	})
	for _, child := range node.Children {
		sortChildren(child)
	}
}

// --- Node helper methods ---

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

// --- File detection helpers ---

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
