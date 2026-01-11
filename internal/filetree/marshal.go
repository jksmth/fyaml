// Package filetree provides filesystem traversal for FYAML packing.
package filetree

import (
	"github.com/jksmth/fyaml/internal/logger"
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
			return n.marshalLeafNode(opts)
		}
		return n.marshalLeafMap(opts)
	}

	// Parent node
	if mode == ModePreserve {
		return n.marshalParentNode(opts)
	}
	return n.marshalParentMap(opts)
}
