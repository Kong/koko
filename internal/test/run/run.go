package run

import (
	"context"
	"crypto/tls"
	"sync"
	"testing"

	"github.com/kong/koko/internal/cmd"
	"github.com/kong/koko/internal/config"
	"github.com/kong/koko/internal/crypto"
	"github.com/kong/koko/internal/db"
	"github.com/kong/koko/internal/log"
	"github.com/kong/koko/internal/test/certs"
	"github.com/kong/koko/internal/test/kong"
	"github.com/kong/koko/internal/test/util"
	"github.com/stretchr/testify/require"
)

type ServerConfigOpt func(*cmd.ServerConfig) error

func WithSQlite3(inMemory bool, filename string) ServerConfigOpt {
	return func(serverConfig *cmd.ServerConfig) error {
		switch inMemory {
		case true:
			serverConfig.Database = config.Database{
				Dialect: db.DialectSQLite3,
				SQLite: config.SQLite{
					InMemory: true,
				},
			}
		default:
			serverConfig.Database = config.Database{
				Dialect: db.DialectSQLite3,
				SQLite: config.SQLite{
					Filename: filename,
				},
			}
		}
		return nil
	}
}

func WithPostgres(dbName, hostname, user, password string, port int) ServerConfigOpt {
	return func(serverConfig *cmd.ServerConfig) error {
		serverConfig.Database = config.Database{
			Dialect: db.DialectPostgres,
			Postgres: config.Postgres{
				DBName:   dbName,
				Hostname: hostname,
				Port:     port,
				User:     user,
				Password: password,
			},
		}
		return nil
	}
}

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
		}
		return nil
	}
}

func Koko(t *testing.T, options ...ServerConfigOpt) func() {
	// build default config
	cert, err := tls.X509KeyPair(certs.DefaultSharedCert, certs.DefaultSharedKey)
	require.Nil(t, err)
	config := &cmd.ServerConfig{
		Logger:     log.Logger,
		KongCPCert: cert,
		DPAuthCert: cert,
		DPAuthMode: cmd.DPAuthSharedMTLS,
		Database: config.Database{
			Dialect: db.DialectSQLite3,
			SQLite: config.SQLite{
				InMemory: true,
			},
		},
	}

	// inject user options
	for _, o := range options {
		err := o(config)
		require.Nil(t, err)
	}

	require.Nil(t, util.CleanDB())

	var wg sync.WaitGroup
	ctx, cancel := context.WithCancel(context.Background())
	wg.Add(1)
	go func() {
		defer wg.Done()
		err := cmd.Run(ctx, *config)
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
