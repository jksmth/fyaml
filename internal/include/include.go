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
type TagProcessor = func(n *yaml.Node, baseDir string, chroot string) error

// resolvePath resolves a path relative to baseDir and validates it's within chroot.
// Returns the absolute chroot, the relative path within chroot, and any error.
func resolvePath(path string, baseDir string, chroot string) (absChroot string, relPath string, err error) {
	// Resolve chroot to absolute path
	absChroot, err = filepath.Abs(chroot)
	if err != nil {
		return "", "", fmt.Errorf("could not resolve chroot %s: %w", chroot, err)
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

	// Convert to relative path within chroot
	relPath, err = filepath.Rel(absChroot, absIncludePath)
	if err != nil {
		return "", "", fmt.Errorf("could not determine relative path for %s: %w", path, err)
	}

	// Check if the path escapes the chroot boundary
	if strings.HasPrefix(relPath, "..") {
		return "", "", fmt.Errorf("include path %s escapes chroot boundary %s", path, chroot)
	}

	return absChroot, relPath, nil
}

// LoadFileText reads a file and returns its contents as a string.
// Paths are resolved relative to baseDir and must be within chroot.
func LoadFileText(path string, baseDir string, chroot string) (string, error) {
	absChroot, relPath, err := resolvePath(path, baseDir, chroot)
	if err != nil {
		return "", err
	}

	// Use os.Root to read file - automatically prevents directory traversal
	root, err := os.OpenRoot(absChroot)
	if err != nil {
		return "", fmt.Errorf("could not open chroot %s: %w", chroot, err)
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
// Paths are resolved relative to baseDir and must be within chroot.
func LoadFileFragment(path string, baseDir string, chroot string) (*yaml.Node, error) {
	absChroot, relPath, err := resolvePath(path, baseDir, chroot)
	if err != nil {
		return nil, err
	}

	// Use os.Root to read file - automatically prevents directory traversal
	root, err := os.OpenRoot(absChroot)
	if err != nil {
		return nil, fmt.Errorf("could not open chroot %s: %w", chroot, err)
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
func HandleCustomTag(n *yaml.Node, tag string, fn TagProcessor, baseDir string, chroot string) error {
	if n == nil {
		return nil
	}

	if n.Tag == tag {
		err := fn(n, baseDir, chroot)
		if err != nil {
			return err
		}
		// After processing, recursively check the replaced content for more tags
		if n.Kind == yaml.SequenceNode || n.Kind == yaml.MappingNode || n.Kind == yaml.DocumentNode {
			for _, child := range n.Content {
				err := HandleCustomTag(child, tag, fn, baseDir, chroot)
				if err != nil {
					return err
				}
			}
		}
	} else {
		// Recursively search children (including DocumentNode which wraps the content)
		if n.Kind == yaml.SequenceNode || n.Kind == yaml.MappingNode || n.Kind == yaml.DocumentNode {
			for _, child := range n.Content {
				err := HandleCustomTag(child, tag, fn, baseDir, chroot)
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
func ProcessIncludeTag(n *yaml.Node, baseDir string, chroot string) error {
	return HandleCustomTag(n, "!include", func(n *yaml.Node, baseDir string, chroot string) error {
		if n.Kind != yaml.ScalarNode {
			return fmt.Errorf("!include tag must be used on a scalar value, got %v", n.Kind)
		}

		// Save the include path before we replace the node
		includePath := n.Value

		fragment, err := LoadFileFragment(includePath, baseDir, chroot)
		if err != nil {
			return err
		}

		// Calculate the new baseDir (directory of the included file) for processing
		// includes within the fragment. This ensures that relative paths in the
		// included file are resolved relative to that file's location, not the
		// original file's location.
		_, relPath, err := resolvePath(includePath, baseDir, chroot)
		if err != nil {
			// If we can't resolve the path, fall back to original baseDir
			// This shouldn't happen since LoadFileFragment already validated it
			*n = *fragment
			return ProcessIncludes(fragment, baseDir, chroot)
		}

		// Get the directory of the included file relative to chroot
		absChroot, err := filepath.Abs(chroot)
		if err != nil {
			*n = *fragment
			return ProcessIncludes(fragment, baseDir, chroot)
		}
		absIncludePath := filepath.Join(absChroot, relPath)
		// Clean the path to resolve any .. components
		absIncludePath = filepath.Clean(absIncludePath)
		newBaseDir := filepath.Dir(absIncludePath)

		// Process includes in the fragment using the included file's directory as baseDir
		// This must be done before replacing the node, so that ProcessIncludes can
		// process all includes (!include, !include-text, <<include()>>) in the fragment
		if err := ProcessIncludes(fragment, newBaseDir, chroot); err != nil {
			return err
		}

		// Replace the node with the (now fully processed) fragment content
		*n = *fragment
		return nil
	}, baseDir, chroot)
}

// ProcessIncludeTextTag recursively searches for the !include-text tag from the given node
// and replaces the tag node with the raw text content of the included file.
func ProcessIncludeTextTag(n *yaml.Node, baseDir string, chroot string) error {
	return HandleCustomTag(n, "!include-text", func(n *yaml.Node, baseDir string, chroot string) error {
		if n.Kind != yaml.ScalarNode {
			return fmt.Errorf("!include-text tag must be used on a scalar value, got %v", n.Kind)
		}

		text, err := LoadFileText(n.Value, baseDir, chroot)
		if err != nil {
			return err
		}

		// Replace node with text content
		// Clear Tag to remove the custom !include-text tag from output
		n.Tag = ""
		n.Value = text
		return nil
	}, baseDir, chroot)
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
func MaybeIncludeFile(s string, baseDir string, chroot string) (string, error) {
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
		return LoadFileText(subMatch, baseDir, chroot)
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
func InlineIncludes(node *yaml.Node, baseDir string, chroot string) error {
	if node == nil {
		return nil
	}

	// If we're dealing with a ScalarNode, we can replace the contents.
	// Otherwise, we recurse into the children of the Node.
	if node.Kind == yaml.ScalarNode && node.Value != "" {
		v, err := MaybeIncludeFile(node.Value, baseDir, chroot)
		if err != nil {
			return err
		}
		node.Value = v
	} else {
		for _, child := range node.Content {
			err := InlineIncludes(child, baseDir, chroot)
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
func ProcessIncludes(node *yaml.Node, baseDir string, chroot string) error {
	if node == nil {
		return nil
	}

	// 1. Process !include tags (YAML structures)
	if err := ProcessIncludeTag(node, baseDir, chroot); err != nil {
		return err
	}

	// 2. Process !include-text tags (text content)
	if err := ProcessIncludeTextTag(node, baseDir, chroot); err != nil {
		return err
	}

	// 3. Process <<include()>> directives (backward compat)
	if err := InlineIncludes(node, baseDir, chroot); err != nil {
		return err
	}

	return nil
}
