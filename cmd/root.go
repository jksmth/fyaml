package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/jksmth/fyaml/internal/logger"
	"github.com/jksmth/fyaml/internal/version"
)

var (
	// Global flags
	verbose bool

	// Global logger, initialized in PersistentPreRun
	log logger.Logger

	// Pack flags (shared between root and pack subcommand)
	dir             string
	output          string
	check           bool
	format          string
	enableIncludes  bool
	convertBooleans bool
	indent          int
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "fyaml [DIR]",
	Short: "Compile directory-structured YAML/JSON into a single document",
	Long: `fyaml compiles a directory of YAML/JSON files into a single document.

Organize your YAML/JSON configuration across multiple files and directories, then
use fyaml to combine them into one file. Directory names become map keys,
file names (without extension) become nested keys, and files starting with
@ merge their contents into the parent directory.

Examples:
  fyaml                             # Pack current directory to stdout (YAML)
  fyaml -o out.yml                  # Pack current directory to file
  fyaml --format json               # Output as JSON
  fyaml -o out.yml --check          # Verify output matches file
  fyaml config/                     # Pack specific directory
  fyaml --dir pack                  # Pack directory named "pack" (avoids subcommand conflict)`,
	Args: cobra.MaximumNArgs(1),
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		// Initialize logger based on global verbose flag
		// Always writes to stderr to avoid interfering with stdout
		log = logger.New(os.Stderr, verbose)
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		// Check for --version flag first
		if showVersion, _ := cmd.Flags().GetBool("version"); showVersion {
			fmt.Println(version.Full())
			return nil
		}

		// Determine directory: --dir flag takes precedence, then positional arg, then default
		targetDir := dir
		if targetDir == "" {
			if len(args) > 0 {
				targetDir = args[0]
			} else {
				targetDir = "."
			}
		}

		opts := PackOptions{
			Dir:             targetDir,
			Format:          format,
			EnableIncludes:  enableIncludes,
			ConvertBooleans: convertBooleans,
			Indent:          indent,
		}

		result, err := pack(opts, log)
		if err != nil {
			return fmt.Errorf("pack error: %w", err)
		}

		if check {
			return handleCheck(output, result)
		}

		return writeOutput(output, result)
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	// Global flags (persistent = available to all subcommands)
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false,
		"Show debug output (applies to all commands)")

	// Pack flags (persistent = available to root and pack subcommand)
	rootCmd.PersistentFlags().StringVar(&dir, "dir", "",
		"Explicitly specify directory to pack (avoids subcommand conflicts)")
	rootCmd.PersistentFlags().StringVarP(&output, "output", "o", "",
		"Write output to file (default: stdout)")
	rootCmd.PersistentFlags().BoolVarP(&check, "check", "c", false,
		"Compare generated output to --output, exit non-zero if different")
	rootCmd.PersistentFlags().StringVarP(&format, "format", "f", "yaml",
		"Output format: yaml or json (default: yaml)")
	rootCmd.PersistentFlags().BoolVar(&enableIncludes, "enable-includes", false,
		"Process <<include(file)>> directives (extension)")
	rootCmd.PersistentFlags().BoolVar(&convertBooleans, "convert-booleans", false,
		"Convert unquoted YAML 1.1 boolean values (on/off, yes/no) to true/false")
	rootCmd.PersistentFlags().IntVar(&indent, "indent", 2,
		"Number of spaces for indentation")

	// Version flag
	rootCmd.Flags().BoolP("version", "V", false,
		"Print version information and exit")

	rootCmd.AddCommand(packCmd)
	rootCmd.AddCommand(versionCmd)
}
