package cmd

import (
	"fmt"

	"github.com/kong/koko/internal/config"
	"github.com/kong/koko/internal/db"
	"github.com/spf13/cobra"
)

// dbMigrateUpCmd is the 'koko db migrate-up' command.
var dbMigrateUpCmd = &cobra.Command{
	Use:   "migrate-up",
	Short: "migrates the database to the most recent version",
	RunE: func(cmd *cobra.Command, args []string) error {
		opts, err := setup()
		if err != nil {
			return err
		}
		logger := opts.Logger
		logger.Debug("setup successful")

		dbConfig, err := config.ToDBConfig(opts.Config.Database, logger)
		if err != nil {
			logger.Fatal(err.Error())
		}

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
		if c == l {
			return fmt.Errorf("database schema already up-to-date, " +
				"no action required")
		}
		if err := m.Up(); err != nil {
			return fmt.Errorf("failed to upgrade database schema: %v", err)
		}
		return nil
	},
}

func init() {
	dbCmd.AddCommand(dbMigrateUpCmd)
}
