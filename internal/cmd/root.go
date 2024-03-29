package cmd

import (
	"context"

	"github.com/spf13/cobra"
)

// rootCmd represents the base command when called without any subcommands.
var rootCmd = &cobra.Command{
	Use:           "koko",
	Short:         "Control plane software for Kong Gateway",
	SilenceUsage:  true,
	SilenceErrors: true,
}

func ExecuteContext(ctx context.Context) {
	cobra.CheckErr(rootCmd.ExecuteContext(ctx))
}

func init() {
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "koko.yaml",
		"path to configuration file")
}
