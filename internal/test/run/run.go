package run

import (
	"context"
	"sync"
	"testing"

	"github.com/kong/koko/internal/cmd"
	"github.com/kong/koko/internal/plugin"
	"github.com/kong/koko/internal/test/kong"
	"github.com/stretchr/testify/require"
)

func Koko(t *testing.T, serverConfig cmd.ServerConfig) func() {
	var wg sync.WaitGroup
	ctx, cancel := context.WithCancel(context.Background())
	wg.Add(1)
	go func() {
		defer wg.Done()
		err := cmd.Run(ctx, serverConfig)
		require.Nil(t, err)
	}()
	return func() {
		cancel()
		plugin.ClearLuaSchemas()
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
