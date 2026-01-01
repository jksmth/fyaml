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
// File paths are resolved relative to baseDir.
//
// Based on CircleCI CLI: https://github.com/CircleCI-Public/circleci-cli
func MaybeIncludeFile(s string, baseDir string) (string, error) {
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

		includePath := filepath.Join(baseDir, subMatch)
		file, err := os.ReadFile(includePath)
		if err != nil {
			return "", fmt.Errorf("could not open %s for inclusion", includePath)
		}

		return string(file), nil
	}

	return s, nil
}

// InlineIncludes recursively walks a yaml.Node tree, replacing include directives
// in scalar node values with the contents of the referenced files.
//
// Based on CircleCI CLI: https://github.com/CircleCI-Public/circleci-cli
func InlineIncludes(node *yaml.Node, baseDir string) error {
	if node == nil {
		return nil
	}

	// If we're dealing with a ScalarNode, we can replace the contents.
	// Otherwise, we recurse into the children of the Node.
	if node.Kind == yaml.ScalarNode && node.Value != "" {
		v, err := MaybeIncludeFile(node.Value, baseDir)
		if err != nil {
			return err
		}
		node.Value = v
	} else {
		for _, child := range node.Content {
			err := InlineIncludes(child, baseDir)
			if err != nil {
				return err
			}
		}
	}

	return nil
}
