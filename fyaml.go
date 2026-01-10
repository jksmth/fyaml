package fyaml

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"path/filepath"
	"strings"

	"go.yaml.in/yaml/v4"

	"github.com/jksmth/fyaml/internal/filetree"
	"github.com/jksmth/fyaml/internal/logger"
)

// Pack compiles a directory of YAML/JSON files into a single document.
//
// Pack is safe for concurrent use by multiple goroutines, provided that
// each call uses a separate PackOptions instance. If multiple goroutines
// share the same Logger instance, log output may interleave (this does not
// affect correctness, only log formatting).
//
// The context can be used to cancel the operation. If the context is canceled,
// Pack will return an error wrapping context.Canceled or context.DeadlineExceeded.
//
// PackOptions.Dir is required. All other options have sensible defaults:
//   - Format defaults to FormatYAML
//   - Mode defaults to ModeCanonical
//   - MergeStrategy defaults to MergeShallow
//   - Indent defaults to 2
//   - Logger defaults to a no-op logger if nil
//
// Returns the packed document as bytes, or an error if packing fails.
func Pack(ctx context.Context, opts PackOptions) ([]byte, error) {
	// Check for context cancellation
	if err := ctx.Err(); err != nil {
		return nil, fmt.Errorf("context canceled: %w", err)
	}

	// Validate directory first
	if opts.Dir == "" {
		return nil, fmt.Errorf("%w", ErrDirectoryRequired)
	}

	// Apply defaults
	if opts.Format == "" {
		opts.Format = FormatYAML
	}
	if opts.Mode == "" {
		opts.Mode = ModeCanonical
	}
	if opts.MergeStrategy == "" {
		opts.MergeStrategy = MergeShallow
	}
	if opts.Indent == 0 {
		opts.Indent = 2
	}

	// Validate indent (after defaults applied, must be positive)
	if opts.Indent < 1 {
		return nil, fmt.Errorf("%w: %d (must be positive)", ErrInvalidIndent, opts.Indent)
	}

	// Use no-op logger if not provided
	log := opts.Logger
	if log == nil {
		log = logger.Nop()
	}

	// Validate format
	if opts.Format != FormatYAML && opts.Format != FormatJSON {
		return nil, fmt.Errorf("%w: %s (must be 'yaml' or 'json')", ErrInvalidFormat, opts.Format)
	}

	// Validate mode
	if opts.Mode != ModeCanonical && opts.Mode != ModePreserve {
		return nil, fmt.Errorf("%w: %s (must be 'canonical' or 'preserve')", ErrInvalidMode, opts.Mode)
	}

	// Validate merge strategy
	if opts.MergeStrategy != MergeShallow && opts.MergeStrategy != MergeDeep {
		return nil, fmt.Errorf("%w: %s (must be 'shallow' or 'deep')", ErrInvalidMergeStrategy, opts.MergeStrategy)
	}

	// Check for context cancellation before I/O operations
	if err := ctx.Err(); err != nil {
		return nil, fmt.Errorf("context canceled: %w", err)
	}

	// Resolve dir to absolute path to use as pack root
	absDir, err := filepath.Abs(opts.Dir)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve directory path: %w", err)
	}

	// Build the filetree
	tree, err := filetree.NewTree(opts.Dir)
	if err != nil {
		return nil, fmt.Errorf("failed to build filetree: %w", err)
	}

	// Handle empty directory
	if tree == nil {
		return handleEmptyOutput(opts.Dir, opts.Format, log)
	}

	// Convert public types to internal types
	mode := filetree.ModeCanonical
	if opts.Mode == ModePreserve {
		mode = filetree.ModePreserve
	}

	mergeStrategy := filetree.MergeShallow
	if opts.MergeStrategy == MergeDeep {
		mergeStrategy = filetree.MergeDeep
	}

	// Create processing options
	procOpts := &filetree.Options{
		EnableIncludes:  opts.EnableIncludes,
		PackRoot:        absDir,
		ConvertBooleans: opts.ConvertBooleans,
		Mode:            mode,
		MergeStrategy:   mergeStrategy,
		Logger:          log,
	}

	// Check for context cancellation before marshaling
	if err := ctx.Err(); err != nil {
		return nil, fmt.Errorf("context canceled: %w", err)
	}

	// Get the marshaled data structure (avoids circular references)
	marshaledData, err := tree.Marshal(procOpts)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal tree: %w", err)
	}

	// Marshal based on format
	result, err := marshalToFormat(marshaledData, opts.Format, opts.Indent)
	if err != nil {
		return nil, err
	}

	// Check if result is effectively empty and handle accordingly
	if strings.TrimSpace(string(result)) == "null" {
		return handleEmptyOutput(opts.Dir, opts.Format, log)
	}

	return result, nil
}

// handleEmptyOutput returns the appropriate empty output for the given format.
func handleEmptyOutput(dir string, format Format, log Logger) ([]byte, error) {
	log.Warnf("no YAML/JSON files found in directory: %s", dir)
	if format == FormatJSON {
		return []byte("null\n"), nil
	}
	return []byte{}, nil
}

// marshalToFormat marshals data to the specified format with the given indent.
// data can be *yaml.Node (preserve mode) or interface{} (canonical mode).
func marshalToFormat(data interface{}, format Format, indent int) ([]byte, error) {
	switch format {
	case FormatJSON:
		// JSON doesn't support comments - if we got a yaml.Node, decode it first
		jsonData := data
		if node, ok := data.(*yaml.Node); ok {
			// Handle nil node (can happen in preserve mode with empty trees)
			if node == nil {
				jsonData = nil
			} else if err := node.Decode(&jsonData); err != nil {
				return nil, fmt.Errorf("failed to decode node for JSON: %w", err)
			}
		}
		// JSON only supports string keys, so normalize any non-string keys
		normalizedData := filetree.NormalizeKeys(jsonData)
		indentStr := strings.Repeat(" ", indent)
		return json.MarshalIndent(normalizedData, "", indentStr)
	case FormatYAML:
		var buf bytes.Buffer
		enc := yaml.NewEncoder(&buf)
		enc.SetIndent(indent)
		if err := enc.Encode(data); err != nil {
			_ = enc.Close() // Close on error, ignore close error
			return nil, err
		}
		if err := enc.Close(); err != nil {
			return nil, err
		}
		return buf.Bytes(), nil
	default:
		// Should never happen due to early validation, but be safe
		return nil, fmt.Errorf("%w: %s", ErrInvalidFormat, format)
	}
}

// Check compares generated output with expected content using exact byte comparison.
// Returns ErrCheckMismatch if contents don't match.
// Whitespace differences will be detected as mismatches.
// opts.Format is used to normalize empty expected content to match format-specific empty output.
// opts defaults to FormatYAML if Format is empty.
func Check(generated []byte, expected []byte, opts CheckOptions) error {
	// Apply format default
	format := opts.Format
	if format == "" {
		format = FormatYAML
	}

	// Normalize empty input to match format-specific empty output
	// JSON format returns "null\n" for empty output, YAML returns empty bytes
	if len(expected) == 0 {
		if format == FormatJSON {
			expected = []byte("null\n")
		}
		// YAML format returns empty bytes, so no change needed
	}

	// Compare contents
	// TODO: When adding options like IgnoreWhitespace, implement them here
	if string(expected) != string(generated) {
		return ErrCheckMismatch
	}
	return nil
}
