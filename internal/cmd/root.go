package cmd

import (
	"context"

	"github.com/kong/koko/internal/db"
	"github.com/kong/koko/internal/persistence/postgres"
	"github.com/kong/koko/internal/persistence/sqlite"
	"github.com/spf13/cobra"
)

// TODO(hbagdi): replace this with a configuration file

var dbConfig = db.Config{
	Dialect: db.DialectSQLite3,
	// Dialect: db.DialectPostgres,
	SQLite: sqlite.Opts{
		Filename: "test.db",
		InMemory: true,
	},
	Postgres: postgres.Opts{
		Hostname: "localhost",
		Port:     postgres.DefaultPort,
		User:     "koko",
		Password: "koko",
		DBName:   "koko",
	},
}

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
