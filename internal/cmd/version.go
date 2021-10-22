package cmd

import (
	"fmt"

	"github.com/kong/koko/internal/info"
	"github.com/spf13/cobra"
)

// versionCmd represents the version command.
var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the koko version",
	Long: `The version command prints the version of koko along with a Git short
commit hash of the source tree.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("koko %s (%s)\n", info.VERSION, info.COMMIT)
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
