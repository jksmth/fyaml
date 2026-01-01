// Package include provides file inclusion pre-processing for FYAML packing.
//
// This file contains code adapted from CircleCI CLI:
// https://github.com/CircleCI-Public/circleci-cli/blob/main/process/process.go
// https://github.com/CircleCI-Public/circleci-cli/blob/main/cmd/orb.go
//
// Original copyright: Copyright (c) 2018 Circle Internet Services, Inc.
// Original license: MIT License
//
// The include feature is an extension to the FYAML specification and must be
// explicitly enabled via the --enable-includes flag.
package include

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"gopkg.in/yaml.v3"
)

// includeRegex matches <<include(file)>> syntax with optional whitespace.
// View: https://regexr.com/599mq
var includeRegex = regexp.MustCompile(`<<[\s]*include\(([-\w\/\.]+)\)[\s]*>>`)

// MaybeIncludeFile checks if the string s is an include directive and returns
// the file contents if so. Returns the original string if not an include.
//
// The entire string must be an include statement - partial includes are not allowed.
// Only one include per value is permitted.
//
// File paths are resolved relative to baseDir. Both absolute and relative paths
// are supported, but must resolve to a path within packRoot (confinement boundary).
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

		// Resolve pack root to absolute path
		absPackRoot, err := filepath.Abs(packRoot)
		if err != nil {
			return "", fmt.Errorf("could not resolve pack root %s: %w", packRoot, err)
		}

		// Resolve the include path to absolute
		var absIncludePath string
		if filepath.IsAbs(subMatch) {
			absIncludePath = filepath.Clean(subMatch)
		} else {
			includePath := filepath.Join(baseDir, subMatch)
			absIncludePath, err = filepath.Abs(includePath)
			if err != nil {
				return "", fmt.Errorf("could not resolve path %s for inclusion: %w", includePath, err)
			}
		}

		// Convert to relative path within pack root
		relPath, err := filepath.Rel(absPackRoot, absIncludePath)
		if err != nil {
			return "", fmt.Errorf("could not determine relative path for %s: %w", subMatch, err)
		}

		// Check if the path escapes the pack root (os.Root will also reject this, but we provide a clearer error)
		if strings.HasPrefix(relPath, "..") {
			return "", fmt.Errorf("include path %s escapes pack root %s", subMatch, packRoot)
		}

		// Use os.Root to read file - automatically prevents directory traversal
		root, err := os.OpenRoot(absPackRoot)
		if err != nil {
			return "", fmt.Errorf("could not open pack root %s: %w", packRoot, err)
		}
		defer func() {
			_ = root.Close() // Ignore error in defer - resource cleanup
		}()

		file, err := root.ReadFile(relPath)
		if err != nil {
			return "", fmt.Errorf("could not open %s for inclusion", subMatch)
		}

		return string(file), nil
	}

	return s, nil
}

// InlineIncludes recursively walks a yaml.Node tree, replacing include directives
// in scalar node values with the contents of the referenced files.
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
