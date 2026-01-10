package cli

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/jksmth/fyaml"
)

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

// handleCheck compares the generated output with an existing file or stdin.
// Returns an error if the file cannot be read (except if it doesn't exist).
// Returns ErrCheckMismatch if the contents don't match.
// format is used to normalize empty stdin/file content to match format-specific empty output.
func handleCheck(output string, result []byte, format string) error {
	var existing []byte
	var err error

	// Determine source: stdin or file
	if output == "" || output == "-" {
		// Check if stdin is a terminal (would block)
		fi, statErr := os.Stdin.Stat()
		if statErr == nil && (fi.Mode()&os.ModeCharDevice) != 0 {
			// Stdin is a terminal
			if output == "" {
				return fmt.Errorf("--check without --output requires stdin input (use pipe or redirect)")
			}
			// output == "-" explicitly requested, but terminal - still error
			return fmt.Errorf("--output - specified but stdin is a terminal (use pipe or redirect)")
		}

		// Read from stdin (pipe/file, safe even if empty)
		existing, err = io.ReadAll(os.Stdin)
		if err != nil {
			return fmt.Errorf("failed to read from stdin: %w", err)
		}
	} else {
		// Read from file (existing behavior)
		// #nosec G304 - user-controlled paths are expected for CLI tools
		existing, err = os.ReadFile(output)
		if err != nil && !os.IsNotExist(err) {
			return fmt.Errorf("failed to read output file: %w", err)
		}
	}

	// Parse format and use public API for comparison
	parsedFormat, err := fyaml.ParseFormat(format)
	if err != nil {
		return fmt.Errorf("invalid format: %w", err)
	}

	return fyaml.Check(result, existing, fyaml.CheckOptions{
		Format: parsedFormat,
	})
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
