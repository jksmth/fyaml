// Package include provides file inclusion pre-processing for FYAML packing.
//
// This package implements unified include processing with three mechanisms:
//   - !include tag: Include parsed YAML structures
//   - !include-text tag: Include raw text content
//   - <<include()>> directive: Backward-compatible alias for !include-text
//
// The include feature is an extension to the FYAML specification and must be
// explicitly enabled via the --enable-includes flag.
//
// This file contains code adapted from:
//   - CircleCI CLI: https://github.com/CircleCI-Public/circleci-cli
//   - go-yamltools: https://github.com/jcwillox/go-yamltools
//
// Original copyright: Copyright (c) 2018 Circle Internet Services, Inc.
// Original license: MIT License
package include

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"go.yaml.in/yaml/v4"
)

// includeRegex matches <<include(file)>> syntax with optional whitespace.
// View: https://regexr.com/599mq
var includeRegex = regexp.MustCompile(`<<[\s]*include\(([-\w\/\.]+)\)[\s]*>>`)

// Fragment is used to parse YAML into a node instead of an interface.
// This allows us to preserve the YAML node structure for tag processing.
type Fragment struct {
	Content *yaml.Node
}

func (f *Fragment) UnmarshalYAML(n *yaml.Node) error {
	f.Content = n
	return nil
}

// TagProcessor is a function that processes a YAML node with a specific tag.
type TagProcessor = func(n *yaml.Node, baseDir string, packRoot string) error

// resolvePath resolves a path relative to baseDir and validates it's within packRoot.
// Returns the absolute pack root, the relative path within pack root, and any error.
func resolvePath(path string, baseDir string, packRoot string) (absPackRoot string, relPath string, err error) {
	// Resolve pack root to absolute path
	absPackRoot, err = filepath.Abs(packRoot)
	if err != nil {
		return "", "", fmt.Errorf("could not resolve pack root %s: %w", packRoot, err)
	}

	// Resolve the include path to absolute
	var absIncludePath string
	if filepath.IsAbs(path) {
		absIncludePath = filepath.Clean(path)
	} else {
		includePath := filepath.Join(baseDir, path)
		absIncludePath, err = filepath.Abs(includePath)
		if err != nil {
			return "", "", fmt.Errorf("could not resolve path %s for inclusion: %w", includePath, err)
		}
	}

	// Convert to relative path within pack root
	relPath, err = filepath.Rel(absPackRoot, absIncludePath)
	if err != nil {
		return "", "", fmt.Errorf("could not determine relative path for %s: %w", path, err)
	}

	// Check if the path escapes the pack root
	if strings.HasPrefix(relPath, "..") {
		return "", "", fmt.Errorf("include path %s escapes pack root %s", path, packRoot)
	}

	return absPackRoot, relPath, nil
}

// LoadFileText reads a file and returns its contents as a string.
// Paths are resolved relative to baseDir and must be within packRoot.
func LoadFileText(path string, baseDir string, packRoot string) (string, error) {
	absPackRoot, relPath, err := resolvePath(path, baseDir, packRoot)
	if err != nil {
		return "", err
	}

	// Use os.Root to read file - automatically prevents directory traversal
	root, err := os.OpenRoot(absPackRoot)
	if err != nil {
		return "", fmt.Errorf("could not open pack root %s: %w", packRoot, err)
	}
	defer func() {
		_ = root.Close() // Ignore error in defer - resource cleanup
	}()

	data, err := root.ReadFile(relPath)
	if err != nil {
		return "", fmt.Errorf("could not open %s for inclusion", path)
	}

	return string(data), nil
}

// LoadFileFragment reads in and parses a given file returning a YAML node.
// Paths are resolved relative to baseDir and must be within packRoot.
func LoadFileFragment(path string, baseDir string, packRoot string) (*yaml.Node, error) {
	absPackRoot, relPath, err := resolvePath(path, baseDir, packRoot)
	if err != nil {
		return nil, err
	}

	// Use os.Root to read file - automatically prevents directory traversal
	root, err := os.OpenRoot(absPackRoot)
	if err != nil {
		return nil, fmt.Errorf("could not open pack root %s: %w", packRoot, err)
	}
	defer func() {
		_ = root.Close() // Ignore error in defer - resource cleanup
	}()

	data, err := root.ReadFile(relPath)
	if err != nil {
		return nil, fmt.Errorf("could not open %s for inclusion: %w", path, err)
	}

	var f Fragment
	err = yaml.Unmarshal(data, &f)
	if err != nil {
		return nil, fmt.Errorf("failed to parse YAML/JSON in %s: %w", path, err)
	}

	return f.Content, nil
}

// HandleCustomTag recursively searches YAML nodes for the tag and calls the tag processor function.
func HandleCustomTag(n *yaml.Node, tag string, fn TagProcessor, baseDir string, packRoot string) error {
	if n == nil {
		return nil
	}

	if n.Tag == tag {
		err := fn(n, baseDir, packRoot)
		if err != nil {
			return err
		}
		// After processing, recursively check the replaced content for more tags
		if n.Kind == yaml.SequenceNode || n.Kind == yaml.MappingNode || n.Kind == yaml.DocumentNode {
			for _, child := range n.Content {
				err := HandleCustomTag(child, tag, fn, baseDir, packRoot)
				if err != nil {
					return err
				}
			}
		}
	} else {
		// Recursively search children (including DocumentNode which wraps the content)
		if n.Kind == yaml.SequenceNode || n.Kind == yaml.MappingNode || n.Kind == yaml.DocumentNode {
			for _, child := range n.Content {
				err := HandleCustomTag(child, tag, fn, baseDir, packRoot)
				if err != nil {
					return err
				}
			}
		}
	}
	return nil
}

// ProcessIncludeTag recursively searches for the !include tag from the given node
// and replaces the tag node with content of the included file (parsed as YAML).
func ProcessIncludeTag(n *yaml.Node, baseDir string, packRoot string) error {
	return HandleCustomTag(n, "!include", func(n *yaml.Node, baseDir string, packRoot string) error {
		if n.Kind != yaml.ScalarNode {
			return fmt.Errorf("!include tag must be used on a scalar value, got %v", n.Kind)
		}

		fragment, err := LoadFileFragment(n.Value, baseDir, packRoot)
		if err != nil {
			return err
		}

		// Replace the node with the fragment content
		*n = *fragment
		return nil
	}, baseDir, packRoot)
}

// ProcessIncludeTextTag recursively searches for the !include-text tag from the given node
// and replaces the tag node with the raw text content of the included file.
func ProcessIncludeTextTag(n *yaml.Node, baseDir string, packRoot string) error {
	return HandleCustomTag(n, "!include-text", func(n *yaml.Node, baseDir string, packRoot string) error {
		if n.Kind != yaml.ScalarNode {
			return fmt.Errorf("!include-text tag must be used on a scalar value, got %v", n.Kind)
		}

		text, err := LoadFileText(n.Value, baseDir, packRoot)
		if err != nil {
			return err
		}

		// Replace node with text content
		n.Tag = "!!str"
		n.Value = text
		return nil
	}, baseDir, packRoot)
}

// MaybeIncludeFile checks if the string s is an include directive and returns
// the file contents if so. Returns the original string if not an include.
//
// The entire string must be an include statement - partial includes are not allowed.
// Only one include per value is permitted.
//
// This function supports the <<include(file)>> directive syntax, which was
// inspired by CircleCI's orb pack implementation.
//
// Based on CircleCI CLI: https://github.com/CircleCI-Public/circleci-cli
func MaybeIncludeFile(s string, baseDir string, packRoot string) (string, error) {
	// Only find up to 2 matches, because we throw an error if we find >1
	includeMatches := includeRegex.FindAllStringSubmatch(s, 2)
	if len(includeMatches) > 1 {
		return "", fmt.Errorf("multiple include statements: '%s'", s)
	}

	if len(includeMatches) == 1 {
		match := includeMatches[0]
		fullMatch, subMatch := match[0], match[1]

		// Throw an error if the entire string wasn't matched
		if fullMatch != s {
			return "", fmt.Errorf("entire string must be include statement: '%s'", s)
		}

		// Use shared LoadFileText for actual file loading
		return LoadFileText(subMatch, baseDir, packRoot)
	}

	return s, nil
}

// InlineIncludes recursively walks a yaml.Node tree, replacing <<include(file)>>
// directives in scalar node values with the contents of the referenced files.
//
// This function supports the <<include(file)>> directive syntax, which was
// inspired by CircleCI's orb pack implementation.
//
// Based on CircleCI CLI: https://github.com/CircleCI-Public/circleci-cli
func InlineIncludes(node *yaml.Node, baseDir string, packRoot string) error {
	if node == nil {
		return nil
	}

	// If we're dealing with a ScalarNode, we can replace the contents.
	// Otherwise, we recurse into the children of the Node.
	if node.Kind == yaml.ScalarNode && node.Value != "" {
		v, err := MaybeIncludeFile(node.Value, baseDir, packRoot)
		if err != nil {
			return err
		}
		node.Value = v
	} else {
		for _, child := range node.Content {
			err := InlineIncludes(child, baseDir, packRoot)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

// ProcessIncludes is the main entry point for all include processing.
// It processes includes in the correct order:
//  1. !include tags (YAML structures)
//  2. !include-text tags (text content)
//  3. <<include()>> directives (backward-compatible alias for !include-text)
func ProcessIncludes(node *yaml.Node, baseDir string, packRoot string) error {
	if node == nil {
		return nil
	}

	// 1. Process !include tags (YAML structures)
	if err := ProcessIncludeTag(node, baseDir, packRoot); err != nil {
		return err
	}

	// 2. Process !include-text tags (text content)
	if err := ProcessIncludeTextTag(node, baseDir, packRoot); err != nil {
		return err
	}

	// 3. Process <<include()>> directives (backward compat)
	if err := InlineIncludes(node, baseDir, packRoot); err != nil {
		return err
	}

	return nil
}
