package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"go.yaml.in/yaml/v4"

	"github.com/jksmth/fyaml/internal/filetree"
	"github.com/jksmth/fyaml/internal/logger"
)

// PackOptions contains all options for the pack command.
type PackOptions struct {
	Dir             string // Directory to pack
	Format          string // Output format: yaml or json
	EnableIncludes  bool   // Process <<include(file)>> directives
	ConvertBooleans bool   // Convert unquoted YAML 1.1 booleans to true/false
	Indent          int    // Number of spaces for indentation
	Mode            string // Marshaling mode: canonical or preserve
	MergeStrategy   string // Merge strategy: shallow or deep
}

// Flag variables are now defined in root.go and shared between root and pack commands

var packCmd = &cobra.Command{
	Use:   "pack [DIR]",
	Short: "Compile directory-structured YAML/JSON into a single file",
	Long: `Pack compiles a directory of YAML/JSON files into a single document.

DIR defaults to the current working directory if not specified.

The output is deterministic - identical directory structures always produce
identical output, with keys sorted alphabetically.

By default, output is YAML. Use --format json to output JSON instead.

Use --enable-includes to process <<include(file)>> directives.

This is an alias for the root command. Both 'fyaml' and 'fyaml pack' work identically.`,
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		// Simply delegate to rootCmd - it has all the logic and flags
		return rootCmd.RunE(rootCmd, args)
	},
}

func init() {
	// Flags are defined as persistent flags on rootCmd, so they're automatically
	// available to packCmd. No need to define them here.
}

// handleCheck compares the generated output with an existing file.
// Returns an error if the file cannot be read (except if it doesn't exist).
// Exits with code 2 if the contents don't match.
func handleCheck(output string, result []byte) error {
	if output == "" {
		return fmt.Errorf("--check requires --output to be specified")
	}
	// #nosec G304 - user-controlled paths are expected for CLI tools
	existing, err := os.ReadFile(output)
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to read output file: %w", err)
	}
	if string(existing) != string(result) {
		os.Exit(2)
	}
	return nil
}

// writeOutput writes the result to a file (atomically) or stdout.
func writeOutput(output string, result []byte) error {
	if output == "" {
		_, err := os.Stdout.Write(result)
		return err
	}

	dir := filepath.Dir(output)
	base := filepath.Base(output)

	tmp, err := os.CreateTemp(dir, base+".tmp.*")
	if err != nil {
		return fmt.Errorf("failed to create temp file: %w", err)
	}
	tmpPath := tmp.Name()

	ok := false
	defer func() {
		_ = tmp.Close() // Ignore error in defer - file may already be closed
		if !ok {
			_ = os.Remove(tmpPath)
		}
	}()

	if _, err := tmp.Write(result); err != nil {
		return fmt.Errorf("failed to write temp file: %w", err)
	}
	if err := tmp.Sync(); err != nil {
		return fmt.Errorf("failed to sync temp file: %w", err)
	}
	if err := tmp.Close(); err != nil {
		return fmt.Errorf("failed to close temp file: %w", err)
	}

	// #nosec G302 - 0644 is standard for config files, umask applies
	if err := os.Chmod(tmpPath, 0o644); err != nil {
		return fmt.Errorf("failed to chmod temp file: %w", err)
	}

	if err := os.Rename(tmpPath, output); err != nil {
		return fmt.Errorf("failed to rename temp file: %w", err)
	}

	ok = true
	return nil
}

// pack compiles a directory-structured YAML/JSON tree into a single document.
// It follows the FYAML specification exactly for YAML, with JSON as an extension.
// If log is nil, a no-op logger is used.
func pack(opts PackOptions, log logger.Logger) ([]byte, error) {
	// Default to no-op logger if not provided
	if log == nil {
		log = logger.Nop()
	}

	// Validate format early
	if opts.Format != "yaml" && opts.Format != "json" {
		return nil, fmt.Errorf("invalid format: %s (must be 'yaml' or 'json')", opts.Format)
	}

	// Validate mode
	if opts.Mode != "" && opts.Mode != "canonical" && opts.Mode != "preserve" {
		return nil, fmt.Errorf("invalid mode: %s (must be 'canonical' or 'preserve')", opts.Mode)
	}

	// Validate merge strategy
	if opts.MergeStrategy != "" && opts.MergeStrategy != "shallow" && opts.MergeStrategy != "deep" {
		return nil, fmt.Errorf("invalid merge strategy: %s (must be 'shallow' or 'deep')", opts.MergeStrategy)
	}

	// Validate indent
	if opts.Indent < 1 {
		return nil, fmt.Errorf("invalid indent: %d (must be positive)", opts.Indent)
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

	// Parse mode (default to canonical)
	mode := filetree.ModeCanonical
	if opts.Mode == "preserve" {
		mode = filetree.ModePreserve
	}

	// Parse merge strategy (default to shallow)
	mergeStrategy := filetree.MergeShallow
	if opts.MergeStrategy == "deep" {
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
func handleEmptyOutput(dir, format string, log logger.Logger) ([]byte, error) {
	log.Warnf("no YAML/JSON files found in directory: %s", dir)
	if format == "json" {
		return []byte("null\n"), nil
	}
	return []byte{}, nil
}

// marshalToFormat marshals data to the specified format with the given indent.
// data can be *yaml.Node (preserve mode) or interface{} (canonical mode).
func marshalToFormat(data interface{}, format string, indent int) ([]byte, error) {
	switch format {
	case "json":
		// JSON doesn't support comments - if we got a yaml.Node, decode it first
		jsonData := data
		if node, ok := data.(*yaml.Node); ok {
			if err := node.Decode(&jsonData); err != nil {
				return nil, fmt.Errorf("failed to decode node for JSON: %w", err)
			}
		}
		// JSON only supports string keys, so normalize any non-string keys
		normalizedData := filetree.NormalizeKeys(jsonData)
		indentStr := strings.Repeat(" ", indent)
		return json.MarshalIndent(normalizedData, "", indentStr)
	case "yaml":
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
		return nil, fmt.Errorf("invalid format: %s", format)
	}
}
