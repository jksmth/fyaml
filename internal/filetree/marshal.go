// Package filetree provides filesystem traversal for FYAML packing.
package filetree

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/jksmth/fyaml/internal/include"
	"github.com/jksmth/fyaml/internal/logger"
	"go.yaml.in/yaml/v4"
)

// Mode controls the marshaling behavior.
type Mode string

const (
	// ModeCanonical produces canonical output with sorted keys and no comments (default).
	ModeCanonical Mode = "canonical"
	// ModePreserve preserves authored key order and comments.
	ModePreserve Mode = "preserve"
)

// MergeStrategy controls how maps are merged when multiple files contribute to the same key.
type MergeStrategy string

const (
	// MergeShallow uses "last wins" behavior - later values completely replace earlier ones (default).
	MergeShallow MergeStrategy = "shallow"
	// MergeDeep recursively merges nested maps, only replacing values at the leaf level.
	MergeDeep MergeStrategy = "deep"
)

// Options controls how the filetree is processed during marshaling.
type Options struct {
	// Include processing
	EnableIncludes bool   // Process <<include(file)>> directives
	PackRoot       string // Absolute path to pack root (confinement boundary)

	// YAML processing
	ConvertBooleans bool          // Convert unquoted YAML 1.1 booleans to true/false
	Mode            Mode          // Marshaling mode: canonical (default) or preserve
	MergeStrategy   MergeStrategy // Merge strategy: shallow (default) or deep

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

// MarshalYAML serializes the tree into YAML.
// Implements yaml.Marshaler interface (called by yaml.Marshal).
func (n *Node) MarshalYAML() (interface{}, error) {
	return n.Marshal(nil)
}

// Marshal serializes the tree into YAML with processing options.
// If opts is nil, processing features are disabled and canonical mode is used.
// Returns *yaml.Node for preserve mode, interface{} for canonical mode.
func (n *Node) Marshal(opts *Options) (interface{}, error) {
	mode := ModeCanonical
	if opts != nil && opts.Mode == ModePreserve {
		mode = ModePreserve
	}

	if len(n.Children) == 0 {
		// Leaf node
		if mode == ModePreserve {
			return n.marshalLeafPreserve(opts)
		}
		return n.marshalLeaf(opts)
	}

	// Parent node
	if mode == ModePreserve {
		return n.marshalParentPreserve(opts)
	}
	return n.marshalParent(opts)
}

// parseYAMLFile reads and parses a YAML file, applying includes and boolean conversion.
// Returns the root yaml.Node (doc.Content[0]), or nil if the file is empty/not YAML.
func (n *Node) parseYAMLFile(opts *Options) (*yaml.Node, error) {
	if n.Info.IsDir() || !isYaml(n.Info) {
		return nil, nil
	}

	opts.log().Debugf("Processing: %s", n.FullPath)

	buf, err := os.ReadFile(n.FullPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file %s: %w", n.FullPath, err)
	}

	var doc yaml.Node
	if err := yaml.Unmarshal(buf, &doc); err != nil {
		return nil, formatYAMLError(err, n.FullPath)
	}
	if len(doc.Content) == 0 {
		return nil, nil
	}

	// Process includes if enabled
	if opts != nil && opts.EnableIncludes {
		baseDir := filepath.Dir(n.FullPath)
		if err := include.ProcessIncludes(&doc, baseDir, opts.PackRoot); err != nil {
			return nil, fmt.Errorf("failed to process includes in %s: %w", n.FullPath, err)
		}
		if len(doc.Content) == 0 {
			return nil, nil
		}
	}

	root := doc.Content[0]

	// Convert YAML 1.1 booleans if enabled
	if opts != nil && opts.ConvertBooleans {
		normalizeYAML11Booleans(root)
	}

	return root, nil
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

// NormalizeKeys recursively converts all map keys to strings.
// This is required for JSON output because JSON only supports string keys.
// YAML allows non-string keys (numbers, booleans, etc.), so this function
// converts them using fmt.Sprintf("%v", key).
//
// Example: map[interface{}]interface{}{123: "value"} becomes map[string]interface{}{"123": "value"}
func NormalizeKeys(v interface{}) interface{} {
	switch val := v.(type) {
	case map[interface{}]interface{}:
		result := make(map[string]interface{})
		for k, v := range val {
			result[fmt.Sprintf("%v", k)] = NormalizeKeys(v)
		}
		return result
	case map[string]interface{}:
		result := make(map[string]interface{})
		for k, v := range val {
			result[k] = NormalizeKeys(v)
		}
		return result
	case []interface{}:
		result := make([]interface{}, len(val))
		for i, elem := range val {
			result[i] = NormalizeKeys(elem)
		}
		return result
	default:
		return v
	}
}
