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

func getConfig(pkiMtls, sharedMtls bool) (*cmd.ServerConfig, error) {
	config := &cmd.ServerConfig{
		Logger: log.Logger,
		Database: config.Database{
			Dialect: db.DialectSQLite3,
			SQLite: config.SQLite{
				InMemory: true,
			},
		},
	}
	if pkiMtls {
		cpCert, err := tls.X509KeyPair(certs.CPCert, certs.CPKey)
		if err != nil {
			return nil, err
		}
		dpCACert, err := crypto.ParsePEMCerts(certs.DPTree1CACert)
		if err != nil {
			return nil, err
		}
		config.KongCPCert = cpCert
		config.DPAuthMode = cmd.DPAuthPKIMTLS
		config.DPAuthCACerts = dpCACert
	} else {
		cert, err := tls.X509KeyPair(certs.DefaultSharedCert, certs.DefaultSharedKey)
		if err != nil {
			return nil, err
		}
		config.KongCPCert = cert
		config.DPAuthCert = cert
		if sharedMtls {
			config.DPAuthMode = cmd.DPAuthSharedMTLS
		}
	}
	return config, nil
}

func Koko(t *testing.T, pkiMtls, sharedMtls bool) func() {
	config, err := getConfig(pkiMtls, sharedMtls)
	require.Nil(t, err)

	require.Nil(t, util.CleanDB())

	var wg sync.WaitGroup
	ctx, cancel := context.WithCancel(context.Background())
	wg.Add(1)
	go func() {
		defer wg.Done()
		err := cmd.Run(ctx, *config)
		require.Nil(t, err)
	}()
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
