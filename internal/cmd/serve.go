package cmd

import (
	"context"
	"fmt"

	"github.com/kong/koko/internal/config"
	"github.com/kong/koko/internal/log"
	"github.com/kong/koko/internal/util"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

var cfgFile string

// serveCmd is 'koko serve' command.
var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Control plane software for Kong Gateway",
	RunE: func(cmd *cobra.Command, args []string) error {
		err := serveMain(cmd.Context())
		if err != nil {
			return err
		}
		return nil
	},
}

func init() {
	serveCmd.Flags().StringVar(&cfgFile, "config", "koko.yaml",
		"path to configuration file")

	rootCmd.AddCommand(serveCmd)
}

type initOpts struct {
	Config config.Config
	Logger *zap.Logger
}

func serveMain(ctx context.Context) error {
	opts, err := setup()
	if err != nil {
		return err
	}
	logger := opts.Logger
	logger.Debug("setup successful")

	return Run(ctx, logger)
}

func setup() (initOpts, error) {
	cfg, err := config.Get(cfgFile)
	if err != nil {
		return initOpts{}, err
	}

	errs := config.Validate(cfg)
	if len(errs) > 0 {
		return initOpts{}, util.MultiError{Errors: errs}
	}

	logger, err := setupLogging(cfg.Log)
	if err != nil {
		return initOpts{}, err
	}
	return initOpts{Config: cfg, Logger: logger}, nil
}

func setupLogging(c config.Log) (*zap.Logger, error) {
	zapConfig := zap.NewProductionConfig()
	level := config.Levels[c.Level]
	zapConfig.Level.SetLevel(level)
	logger, err := zapConfig.Build()
	if err != nil {
		return nil, fmt.Errorf("create logger: %v", err)
	}
	log.Logger = logger
	return logger, nil
}
