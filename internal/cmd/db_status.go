package cmd

import (
	"fmt"

	"github.com/kong/koko/internal/config"
	"github.com/kong/koko/internal/db"
	"github.com/spf13/cobra"
)

// dbStatusCmd is the 'koko db status' command.
var dbStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "shows the current status of the database",
	RunE: func(cmd *cobra.Command, args []string) error {
		opts, err := setup()
		if err != nil {
			return err
		}
		logger := opts.Logger
		logger.Debug("setup successful")

		dbConfig := config.ToDBConfig(opts.Config.Database)
		dbConfig.Logger = logger

		m, err := db.NewMigrator(dbConfig)
		if err != nil {
			return err
		}
		c, l, err := m.Status()
		if err != nil {
			return err
		}
		logger.Sugar().Infof("current schema version: %d, "+
			"latest schema version: %d", c, l)
		if c != l {
			return fmt.Errorf("database schema out of date")
		}
		logger.Info("database schema is up to date")
		return nil
	},
}

func init() {
	dbCmd.AddCommand(dbStatusCmd)
}
