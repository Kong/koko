package util

import (
	"context"
	"fmt"

	"github.com/kong/go-kong/kong"
	model "github.com/kong/koko/internal/gen/grpc/kong/admin/model/v1"
)

type KongConfig struct {
	Services       []*kong.Service       `json:"services,omitempty"`
	Routes         []*kong.Route         `json:"routes,omitempty"`
	Plugins        []*kong.Plugin        `json:"plugins,omitempty"`
	Upstreams      []*kong.Upstream      `json:"upstreams,omitempty"`
	Targets        []*kong.Target        `json:"targets,omitempty"`
	Consumers      []*kong.Consumer      `json:"consumers,omitempty"`
	Certificates   []*kong.Certificate   `json:"certificates,omitempty"`
	CACertificates []*kong.CACertificate `json:"ca_certificates,omitempty"`
	SNIs           []*kong.SNI           `json:"snis,omitempty"`
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
	upstreams, err := client.Upstreams.ListAll(ctx)
	if err != nil {
		return KongConfig{}, fmt.Errorf("fetch upstreams: %v", err)
	}
	consumers, err := client.Consumers.ListAll(ctx)
	if err != nil {
		return KongConfig{}, fmt.Errorf("fetch consumers: %v", err)
	}
	certificates, err := client.Certificates.ListAll(ctx)
	if err != nil {
		return KongConfig{}, fmt.Errorf("fetch certificates: %v", err)
	}
	caCertificates, err := client.CACertificates.ListAll(ctx)
	if err != nil {
		return KongConfig{}, fmt.Errorf("fetch CA certificates: %v", err)
	}
	snis, err := client.SNIs.ListAll(ctx)
	if err != nil {
		return KongConfig{}, fmt.Errorf("fetch SNIs: %v", err)
	}
	var allTargets []*kong.Target
	for _, u := range upstreams {
		targets, err := client.Targets.ListAll(ctx, u.ID)
		if err != nil {
			return KongConfig{},
				fmt.Errorf("fetch targets for upstream '%v': %v", *u.ID, err)
		}
		allTargets = append(allTargets, targets...)
	}
	return KongConfig{
		Services:       services,
		Routes:         routes,
		Plugins:        plugins,
		Upstreams:      upstreams,
		Consumers:      consumers,
		Certificates:   certificates,
		CACertificates: caCertificates,
		SNIs:           snis,
		Targets:        allTargets,
	}, nil
}
