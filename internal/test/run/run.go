package run

import (
	"context"
	"crypto/tls"
	"fmt"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/kong/koko/internal/cmd"
	"github.com/kong/koko/internal/config"
	"github.com/kong/koko/internal/crypto"
	"github.com/kong/koko/internal/db"
	"github.com/kong/koko/internal/log"
	"github.com/kong/koko/internal/persistence/postgres"
	"github.com/kong/koko/internal/test/certs"
	"github.com/kong/koko/internal/test/kong"
	"github.com/kong/koko/internal/test/util"
	"github.com/stretchr/testify/require"
)

const queryTimeout = 3 * time.Second

type ServerConfigOpt func(*cmd.ServerConfig) error

func WithDPAuthMode(dpAuthMode cmd.DPAuthMode) ServerConfigOpt {
	return func(serverConfig *cmd.ServerConfig) error {
		switch dpAuthMode {
		case cmd.DPAuthPKIMTLS:
			cpCert, err := tls.X509KeyPair(certs.CPCert, certs.CPKey)
			if err != nil {
				return err
			}

			dpCACert, err := crypto.ParsePEMCerts(certs.DPTree1CACert)
			if err != nil {
				return err
			}

			serverConfig.KongCPCert = cpCert
			serverConfig.DPAuthMode = cmd.DPAuthPKIMTLS
			serverConfig.DPAuthCACerts = dpCACert
		case cmd.DPAuthSharedMTLS:
			cert, err := tls.X509KeyPair(certs.DefaultSharedCert, certs.DefaultSharedKey)
			if err != nil {
				return err
			}

			serverConfig.KongCPCert = cert
			serverConfig.DPAuthCert = cert
			serverConfig.DPAuthMode = cmd.DPAuthSharedMTLS
		default:
			panic(fmt.Sprintf("unknown DPAuthMode: %v", dpAuthMode))
		}
		return nil
	}
}

func Koko(t *testing.T, options ...ServerConfigOpt) func() {
	// build default config
	cert, err := tls.X509KeyPair(certs.DefaultSharedCert, certs.DefaultSharedKey)
	require.Nil(t, err)
	serverConfig := &cmd.ServerConfig{
		Logger:     log.Logger,
		KongCPCert: cert,
		DPAuthCert: cert,
		DPAuthMode: cmd.DPAuthSharedMTLS,
	}

	dialect := os.Getenv("KOKO_TEST_DB")
	if dialect == "" {
		dialect = "sqlite3"
	}
	switch dialect {
	case "sqlite3":
		serverConfig.Database = config.Database{
			Dialect: db.DialectSQLite3,
			SQLite: config.SQLite{
				InMemory: true,
			},
			QueryTimeout: queryTimeout.String(),
		}
	case "postgres":
		serverConfig.Database = config.Database{
			Dialect: db.DialectPostgres,
			Postgres: config.Postgres{
				Hostname: "localhost",
				Port:     postgres.DefaultPort,
				User:     "koko",
				Password: "koko",
				DBName:   "koko",
			},
			QueryTimeout: queryTimeout.String(),
		}
	default:
		panic(fmt.Sprintf("unknown dialect: %s", dialect))
	}

	// inject user options
	for _, o := range options {
		err := o(serverConfig)
		require.Nil(t, err)
	}

	require.Nil(t, util.CleanDB(t))

	var wg sync.WaitGroup
	ctx, cancel := context.WithCancel(context.Background())
	wg.Add(1)
	go func() {
		defer wg.Done()
		err := cmd.Run(ctx, *serverConfig)
		require.Nil(t, err)
	}()
	require.Nil(t, util.WaitForAdminAPI(t))
	return func() {
		cancel()
		wg.Wait()
	}
}

func KongDP(input kong.DockerInput) func() {
	var wg sync.WaitGroup
	ctx, cancel := context.WithCancel(context.Background())
	wg.Add(1)
	go func() {
		defer wg.Done()
		_ = kong.RunDP(ctx, input)
	}()
	return func() {
		cancel()
		wg.Wait()
	}
}
