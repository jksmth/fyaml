// Package filetree provides filesystem traversal for FYAML packing.
package filetree

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/jksmth/fyaml/internal/include"
	"go.yaml.in/yaml/v4"
)

// parseYAMLFile reads and parses a YAML file, applying includes and boolean conversion.
// For multi-document files, all documents are merged into a single result.
// Returns the root yaml.Node, or nil if the file is empty/not YAML.
func (n *Node) parseYAMLFile(opts *Options) (*yaml.Node, error) {
	if n.Info.IsDir() || !isYaml(n.Info) {
		return nil, nil
	}

	opts.log().Debugf("Processing: %s", n.FullPath)

	documents, err := n.decodeYAMLDocuments(opts)
	if err != nil {
		return nil, err
	}
	if len(documents) == 0 {
		return nil, nil
	}

	documents, err = n.processIncludesForDocuments(documents, opts)
	if err != nil {
		return nil, err
	}
	if len(documents) == 0 {
		return nil, nil
	}

	root, err := n.mergeDocuments(documents, opts)
	if err != nil {
		return nil, err
	}
	if root == nil {
		return nil, nil
	}

	// Convert YAML 1.1 booleans if enabled
	if opts != nil && opts.ConvertBooleans {
		normalizeYAML11Booleans(root)
	}

	return root, nil
}

// prepareYAMLBuffer reads the file and optionally prepends anchor definitions.
// Returns the YAML buffer and whether anchor definitions were prepended.
func (n *Node) prepareYAMLBuffer(opts *Options) ([]byte, bool, error) {
	buf, err := os.ReadFile(n.FullPath)
	if err != nil {
		return nil, false, fmt.Errorf("failed to read file %s: %w", n.FullPath, err)
	}

	if !opts.hasAnchors() {
		return buf, false, nil
	}

	prepended, prependErr := prependAnchorsToYAML(buf, opts.anchorRegistry)
	if prependErr != nil {
		// If prepending fails, try parsing without it (might work if no cross-file aliases)
		opts.log().Debugf("Failed to prepend anchors to %s, trying without: %v", n.FullPath, prependErr)
		return buf, false, nil
	}

	// Only mark as prepended if the buffer actually changed
	// (prependAnchorsToYAML returns original buffer if no anchors)
	hasPrependedAnchors := !bytes.Equal(prepended, buf)
	return prepended, hasPrependedAnchors, nil
}

// decodeYAMLDocuments reads and decodes all YAML documents from the file.
// Returns a slice of document content nodes (one per document).
func (n *Node) decodeYAMLDocuments(opts *Options) ([]*yaml.Node, error) {
	buf, hasPrependedAnchors, err := n.prepareYAMLBuffer(opts)
	if err != nil {
		return nil, err
	}

	decoder := yaml.NewDecoder(bytes.NewReader(buf))
	var documents []*yaml.Node

	for {
		var doc yaml.Node
		if err := decoder.Decode(&doc); err != nil {
			if err == io.EOF {
				break
			}
			return nil, formatYAMLError(err, n.FullPath)
		}

		// Always skip empty documents
		if len(doc.Content) == 0 {
			continue
		}

		contentNode := doc.Content[0]

		// Skip the first document when anchors were prepended (anchor definitions)
		if hasPrependedAnchors {
			hasPrependedAnchors = false
			continue
		}

		// Keep all user documents (supports multi-document files)
		// Note: mergeDocuments will validate that all documents are mappings
		documents = append(documents, contentNode)
	}

	return documents, nil
}

// processIncludesForDocuments processes includes for each document.
// Returns a new slice with only non-empty documents after processing.
func (n *Node) processIncludesForDocuments(documents []*yaml.Node, opts *Options) ([]*yaml.Node, error) {
	if opts == nil || !opts.EnableIncludes {
		return documents, nil
	}

	baseDir := filepath.Dir(n.FullPath)
	var validDocuments []*yaml.Node

	for i, docNode := range documents {
		// Wrap in a document node for ProcessIncludes
		doc := &yaml.Node{
			Kind:    yaml.DocumentNode,
			Content: []*yaml.Node{docNode},
		}
		if err := include.ProcessIncludes(doc, baseDir, opts.PackRoot); err != nil {
			return nil, fmt.Errorf("failed to process includes in document %d of %s: %w", i+1, n.FullPath, err)
		}
		if len(doc.Content) == 0 {
			// Document became empty after includes, skip it
			continue
		}
		validDocuments = append(validDocuments, doc.Content[0])
	}

	return validDocuments, nil
}

// mergeDocuments merges multiple documents into a single node.
// Returns the merged node, or the single document if only one exists.
// Returns an error if any document is not a map.
func (n *Node) mergeDocuments(documents []*yaml.Node, opts *Options) (*yaml.Node, error) {
	if len(documents) == 0 {
		return nil, nil
	}

	// Get merge strategy from options (default to shallow)
	strategy := MergeShallow
	if opts != nil && opts.MergeStrategy == MergeDeep {
		strategy = MergeDeep
	}

	// Validate and merge all documents in a single loop
	merged := documents[0]
	for i, doc := range documents {
		if doc.Kind != yaml.MappingNode {
			return nil, fmt.Errorf("expected map for document %d, got %v in %s", i+1, doc.Kind, n.FullPath)
		}
		if i > 0 {
			mergeMapping(merged, doc, strategy)
		}
	}

	return merged, nil
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
