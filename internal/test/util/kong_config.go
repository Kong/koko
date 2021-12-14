package util

import (
	"context"
	"fmt"

	"github.com/kong/go-kong/kong"
	model "github.com/kong/koko/internal/gen/grpc/kong/admin/model/v1"
)

type KongConfig struct {
	Services []*kong.Service `json:"services,omitempty"`
	Routes   []*kong.Route   `json:"routes,omitempty"`
	Plugins  []*kong.Plugin  `json:"plugins,omitempty"`
}

func EnsureConfig(expectedConfig *model.TestingConfig) error {
	gotConfig, err := fetchKongConfig()
	if err != nil {
		return fmt.Errorf("fetching kong config: %v", err)
	}
	err = JSONSubset(expectedConfig, gotConfig)
	return err
}

var BasedKongAdminAPIAddr = kong.String("http://localhost:8001")

func fetchKongConfig() (KongConfig, error) {
	ctx := context.Background()
	client, err := kong.NewClient(BasedKongAdminAPIAddr, nil)
	if err != nil {
		return KongConfig{}, fmt.Errorf("create go client for kong: %v", err)
	}
	services, err := client.Services.ListAll(ctx)
	if err != nil {
		return KongConfig{}, fmt.Errorf("fetch services: %v", err)
	}
	routes, err := client.Routes.ListAll(ctx)
	if err != nil {
		return KongConfig{}, fmt.Errorf("fetch routes: %v", err)
	}
	plugins, err := client.Plugins.ListAll(ctx)
	if err != nil {
		return KongConfig{}, fmt.Errorf("fetch plugins: %v", err)
	}
	return KongConfig{
		Services: services,
		Routes:   routes,
		Plugins:  plugins,
	}, nil
}
