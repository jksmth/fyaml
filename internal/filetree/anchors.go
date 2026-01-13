// Package filetree provides filesystem traversal for FYAML packing.
package filetree

import (
	"bytes"
	"fmt"
	"sort"

	"go.yaml.in/yaml/v4"
)

// anchorRegistry stores collected anchors from all files.
// Anchors are keyed by anchor name. Implements "last wins" policy for duplicates.
type anchorRegistry struct {
	anchors map[string]*yaml.Node // anchor name -> anchor node (last wins)
}

// collectContext holds the context for anchor collection from a single file.
type collectContext struct {
	registry *anchorRegistry
	filePath string
	opts     *Options // For logging warnings
}

// collectAnchorsFromNode recursively walks a yaml.Node tree and collects all anchors.
// Nodes with the Anchor field set are stored in the registry.
// Implements "last wins" policy: since files are processed in deterministic order,
// if an anchor already exists, the current file is later and replaces it.
func collectAnchorsFromNode(node *yaml.Node, ctx *collectContext) {
	if node == nil {
		return
	}

	if node.Anchor != "" {
		// Log warning if anchor already exists (last wins)
		if _, exists := ctx.registry.anchors[node.Anchor]; exists {
			ctx.opts.log().Warnf("Anchor '%s' in %s replaces existing definition (last wins)",
				node.Anchor, ctx.filePath)
		}

		ctx.registry.anchors[node.Anchor] = deepCopyNode(node)
	}

	for _, child := range node.Content {
		collectAnchorsFromNode(child, ctx)
	}
}

// collectAnchorsFromTree walks the file tree and collects anchors from all files.
// Uses an incremental approach: parses files that succeed, collects their anchors,
// then retries failed files with collected anchors prepended. Continues until no
// progress is made, supporting any depth of cross-file anchor dependencies.
// Files are processed in deterministic order (from filetree's sorted structure),
// so "last wins" policy is automatic - later files replace earlier anchor definitions.
// Files that fail to parse after all passes are logged as warnings.
func collectAnchorsFromTree(root *Node, opts *Options) (*anchorRegistry, error) {
	registry := &anchorRegistry{
		anchors: make(map[string]*yaml.Node),
	}

	allYamlFiles := collectFileNodes(root)

	// Create parse options with cross-file anchors enabled
	// The registry starts empty and grows as anchors are collected
	parseOpts := &Options{
		Logger:         opts.Logger,
		EnableAnchors:  true,
		anchorRegistry: registry,
	}

	// Track which files have been successfully parsed
	successfullyParsed := make(map[string]bool)

	// Keep trying until no progress - handles any nesting depth
	for {
		progressMade := false

		// Iterate in deterministic order (maintains "last wins" policy)
		for _, fileNode := range allYamlFiles {
			if successfullyParsed[fileNode.FullPath] {
				continue
			}

			documents, err := fileNode.decodeYAMLDocuments(parseOpts)
			if err != nil {
				continue // Will retry in next pass
			}

			ctx := &collectContext{
				registry: registry,
				filePath: fileNode.FullPath,
				opts:     opts,
			}
			for _, doc := range documents {
				collectAnchorsFromNode(doc, ctx)
			}
			successfullyParsed[fileNode.FullPath] = true
			progressMade = true
		}

		if !progressMade {
			break
		}
	}

	// Log warnings for files that failed to parse (likely have unresolvable aliases)
	for _, fileNode := range allYamlFiles {
		if !successfullyParsed[fileNode.FullPath] {
			opts.log().Warnf("Failed to collect anchors from %s (may have unresolved aliases or invalid YAML)", fileNode.FullPath)
		}
	}

	return registry, nil
}

// buildAnchorMapping constructs the anchor definitions as a yaml.Node mapping.
// Format: anchorName: &anchorName <content>
// Anchor names are sorted for deterministic output.
func buildAnchorMapping(registry *anchorRegistry) *yaml.Node {
	// Collect and sort anchor names for deterministic output
	anchorNames := make([]string, 0, len(registry.anchors))
	for anchorName := range registry.anchors {
		anchorNames = append(anchorNames, anchorName)
	}
	sort.Strings(anchorNames)

	mapping := newMapping()

	for _, anchorName := range anchorNames {
		// Deep copy the anchor node and set the anchor name
		// yaml.v4 will resolve any aliases within anchors when parsing
		valueNode := deepCopyNode(registry.anchors[anchorName])
		valueNode.Anchor = anchorName

		mappingSet(mapping, newScalarKey(anchorName), valueNode)
	}

	return mapping
}

// prependAnchorsToYAML prepends anchor definitions in a separate document so yaml.v4 can resolve aliases.
// Uses document separator (---) to separate anchor definitions from user YAML.
// yaml.v4 automatically resolves aliases across documents, so anchors defined in the first document
// are available to aliases in subsequent documents.
func prependAnchorsToYAML(yamlBytes []byte, registry *anchorRegistry) ([]byte, error) {
	if registry == nil || len(registry.anchors) == 0 {
		return yamlBytes, nil
	}

	anchorMapping := buildAnchorMapping(registry)

	var buf bytes.Buffer
	enc := yaml.NewEncoder(&buf)
	if err := enc.Encode(anchorMapping); err != nil {
		return nil, fmt.Errorf("failed to encode anchor definitions: %w", err)
	}
	if err := enc.Close(); err != nil {
		return nil, fmt.Errorf("failed to close encoder: %w", err)
	}

	// Append document separator and original YAML
	// Only add separator if user's YAML doesn't already start with one
	if !bytes.HasPrefix(yamlBytes, []byte("---")) {
		buf.WriteString("---\n")
	}
	buf.Write(yamlBytes)

	return buf.Bytes(), nil
}

// deepCopyNode creates a deep copy of a yaml.Node.
func deepCopyNode(node *yaml.Node) *yaml.Node {
	if node == nil {
		return nil
	}

	// Shallow copy all fields (Alias pointer is copied as-is; will be resolved by yaml.v4)
	copy := *node

	// Deep copy content, preserving nil vs empty slice distinction
	if node.Content != nil {
		copy.Content = make([]*yaml.Node, len(node.Content))
		for i, child := range node.Content {
			copy.Content[i] = deepCopyNode(child)
		}
	}

	return &copy
}
