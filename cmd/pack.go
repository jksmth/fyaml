package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"

	"github.com/jksmth/fyaml/internal/filetree"
)

var packCmd = &cobra.Command{
	Use:   "pack [DIR]",
	Short: "Compile directory-structured YAML/JSON into a single file",
	Long: `Pack compiles a directory of YAML/JSON files into a single document.

DIR defaults to the current working directory if not specified.

The output is deterministic - identical directory structures always produce
identical output, with keys sorted alphabetically.

By default, output is YAML. Use --format json to output JSON instead.`,
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		dir := "."
		if len(args) > 0 {
			dir = args[0]
		}

		output, _ := cmd.Flags().GetString("output")
		check, _ := cmd.Flags().GetBool("check")
		format, _ := cmd.Flags().GetString("format")

		result, err := pack(dir, format)
		if err != nil {
			return fmt.Errorf("pack error: %w", err)
		}

		if check {
			return handleCheck(output, result)
		}

		return writeOutput(output, result)
	},
}

func init() {
	// Add pack command flags
	packCmd.Flags().StringP("output", "o", "", "Write output to file (default: stdout)")
	packCmd.Flags().Bool("check", false, "Compare generated output to --output, exit non-zero if different")
	packCmd.Flags().StringP("format", "f", "yaml", "Output format: yaml or json (default: yaml)")
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
func pack(dir string, format string) ([]byte, error) {
	// Validate format early
	if format != "yaml" && format != "json" {
		return nil, fmt.Errorf("invalid format: %s (must be 'yaml' or 'json')", format)
	}

	// Build the filetree
	tree, err := filetree.NewTree(dir)
	if err != nil {
		return nil, fmt.Errorf("failed to build filetree: %w", err)
	}

	// Handle empty directory
	if tree == nil {
		return handleEmptyOutput(dir, format)
	}

	// Get the marshaled data structure (avoids circular references)
	marshaledData, err := tree.MarshalYAML()
	if err != nil {
		return nil, fmt.Errorf("failed to marshal tree: %w", err)
	}

	// Marshal based on format
	result, err := marshalToFormat(marshaledData, format)
	if err != nil {
		return nil, err
	}

	// Check if result is effectively empty and handle accordingly
	if strings.TrimSpace(string(result)) == "null" {
		return handleEmptyOutput(dir, format)
	}

	return result, nil
}

// handleEmptyOutput returns the appropriate empty output for the given format.
func handleEmptyOutput(dir, format string) ([]byte, error) {
	fmt.Fprintf(os.Stderr, "warning: no YAML/JSON files found in directory: %s\n", dir)
	if format == "json" {
		return []byte("null\n"), nil
	}
	return []byte{}, nil
}

// marshalToFormat marshals data to the specified format.
func marshalToFormat(data interface{}, format string) ([]byte, error) {
	switch format {
	case "json":
		return json.MarshalIndent(data, "", "  ")
	case "yaml":
		return yaml.Marshal(data)
	default:
		// Should never happen due to early validation, but be safe
		return nil, fmt.Errorf("invalid format: %s", format)
	}
}
