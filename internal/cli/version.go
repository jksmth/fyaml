package cli

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/jksmth/fyaml/internal/version"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print version information",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println(version.Full())
	},
}
