package cmd

import (
	"os"

	"github.com/spf13/cobra"

	"github.com/jksmth/fyaml/internal/logger"
)

var (
	// Global flags
	verbose bool

	// Global logger, initialized in PersistentPreRun
	log logger.Logger
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
  fyaml pack config/              # Pack config directory to stdout (YAML)
  fyaml pack config/ -o out.yml  # Pack to output file
  fyaml pack config/ --format json  # Output as JSON
  fyaml pack config/ -o out.yml --check  # Verify output matches file`,
	Args: cobra.MaximumNArgs(1),
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		// Initialize logger based on global verbose flag
		// Always writes to stderr to avoid interfering with stdout
		log = logger.New(os.Stderr, verbose)
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		return packCmd.RunE(cmd, args)
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

	rootCmd.AddCommand(packCmd)
	rootCmd.AddCommand(versionCmd)
}
