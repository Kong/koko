package cmd

import (
	"github.com/spf13/cobra"
)

// dbCmd is the 'koko db' command.
var dbCmd = &cobra.Command{
	Use:   "db",
	Short: "db command manages database schema and connectivity",
}

func init() {
	rootCmd.AddCommand(dbCmd)
}
