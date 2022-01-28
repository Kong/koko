package cmd

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"

	"github.com/hbagdi/gang"
	"github.com/kong/koko/internal/config"
	"github.com/kong/koko/internal/db"
	v1 "github.com/kong/koko/internal/gen/grpc/kong/admin/service/v1"
	relay "github.com/kong/koko/internal/gen/grpc/kong/relay/service/v1"
	grpcKongUtil "github.com/kong/koko/internal/gen/grpc/kong/util/v1"
	"github.com/kong/koko/internal/persistence"
	"github.com/kong/koko/internal/plugin"
	"github.com/kong/koko/internal/resource"
	"github.com/kong/koko/internal/server"
	"github.com/kong/koko/internal/server/admin"
	"github.com/kong/koko/internal/server/health"
	"github.com/kong/koko/internal/server/kong/ws"
	relayImpl "github.com/kong/koko/internal/server/relay"
	serverUtil "github.com/kong/koko/internal/server/util"
	"github.com/kong/koko/internal/store"
	"github.com/kong/koko/internal/util"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

type ServerConfig struct {
	DPAuthMode    DPAuthMode
	DPAuthCert    tls.Certificate
	DPAuthCACerts []*x509.Certificate

	KongCPCert tls.Certificate

	Logger   *zap.Logger
	Database config.Database
}

type DPAuthMode int

const (
	DPAuthSharedMTLS = iota
	DPAuthPKIMTLS
)

func Run(ctx context.Context, config ServerConfig) error {
	logger := config.Logger
	var g gang.Gang

	persister, err := setupDB(logger, config.Database)
	if err != nil {
		return fmt.Errorf("database: %v", err)
	}

	store := store.New(persister, logger.With(zap.String("component",
		"store"))).ForCluster("default")

	validator, err := plugin.NewLuaValidator(plugin.Opts{Logger: logger})
	if err != nil {
		return err
	}
	err = validator.LoadSchemasFromEmbed(plugin.Schemas, "schemas")
	if err != nil {
		return err
	}
	resource.SetValidator(validator)

	storeLoader := serverUtil.DefaultStoreLoader{Store: store}
	adminLogger := logger.With(zap.String("component", "admin-server"))
	h, err := admin.NewHandler(admin.HandlerOpts{
		Logger:      adminLogger,
		StoreLoader: storeLoader,
	})
	if err != nil {
		return err
	}

	// setup Admin API server
	s, err := server.NewHTTP(server.HTTPOpts{
		Address: ":3000",
		Logger:  adminLogger,
		Handler: h,
	})
	if err != nil {
		return err
	}
	g.AddWithCtxE(s.Run)

	// setup relay server
	rawGRPCServer := admin.NewGRPC(admin.HandlerOpts{
		Logger:      logger.With(zap.String("component", "admin-server")),
		StoreLoader: storeLoader,
	})
	if err != nil {
		return err
	}

	grpcServer, err := server.NewGRPC(server.GRPCOpts{
		Address:    ":3001",
		GRPCServer: rawGRPCServer,
		Logger:     logger,
	})
	if err != nil {
		return err
	}
	eventService := relayImpl.NewEventService(ctx,
		relayImpl.EventServiceOpts{
			Store:  store,
			Logger: logger.With(zap.String("component", "relay-server")),
		})
	relay.RegisterEventServiceServer(rawGRPCServer, eventService)
	statusService := relayImpl.NewStatusService(relayImpl.StatusServiceOpts{
		StoreLoader: storeLoader,
		Logger:      logger.With(zap.String("component", "relay-server")),
	})
	relay.RegisterStatusServiceServer(rawGRPCServer, statusService)
	g.AddWithCtxE(grpcServer.Run)

	// setup relay client
	configClient, err := setupRelayClient()
	if err != nil {
		return err
	}

	// setup control server

	controlLogger := logger.With(zap.String("component", "control-server"))
	m := ws.NewManager(ws.ManagerOpts{
		Logger:  controlLogger,
		Client:  configClient,
		Cluster: ws.DefaultCluster{},
		// TODO(hbagdi): make this configurable
		Config: ws.ManagerConfig{
			DataPlaneRequisites: []*grpcKongUtil.DataPlanePrerequisite{
				{
					Config: &grpcKongUtil.DataPlanePrerequisite_RequiredPlugins{
						RequiredPlugins: &grpcKongUtil.RequiredPluginsFilter{
							RequiredPlugins: []string{"rate-limiting"},
						},
					},
				},
			},
		},
	})
	var authFn ws.AuthFn
	switch config.DPAuthMode {
	case DPAuthSharedMTLS:
		authFn, err = ws.AuthFnSharedTLS(config.DPAuthCert)
		if err != nil {
			return err
		}
	case DPAuthPKIMTLS:
		authFn, err = ws.AuthFnPKITLS(config.DPAuthCACerts)
		if err != nil {
			return err
		}
	default:
		return fmt.Errorf("unknown auth mode: %v", config.DPAuthMode)
	}

	authenticator := &ws.DefaultAuthenticator{
		Manager: m,
		Context: ctx,
		AuthFn:  authFn,
	}
	handler, err := ws.NewHandler(ws.HandlerOpts{
		Logger:        controlLogger,
		Authenticator: authenticator,
	})
	if err != nil {
		return err
	}

	s, err = server.NewHTTP(server.HTTPOpts{
		Address: ":3100",
		Logger:  controlLogger,
		Handler: handler,
		TLS: &tls.Config{
			MinVersion:   tls.VersionTLS12,
			Certificates: []tls.Certificate{config.KongCPCert},
			ClientAuth:   tls.RequestClientCert,
		},
	})
	if err != nil {
		return err
	}
	g.AddWithCtxE(s.Run)

	// health endpoint
	handler, err = health.NewHandler(health.HandlerOpts{})
	if err != nil {
		return err
	}

	s, err = server.NewHTTP(server.HTTPOpts{
		Address: ":4200",
		Logger:  logger.With(zap.String("component", "health-server")),
		Handler: handler,
	})
	if err != nil {
		return err
	}
	g.AddWithCtxE(s.Run)

	// run rabbit run
	errCh := g.Run(ctx)
	var mErr util.MultiError
	for err := range errCh {
		mErr.Errors = append(mErr.Errors, err)
	}
	if len(mErr.Errors) > 0 {
		return mErr
	}
	return nil
}

func setupRelayClient() (ws.ConfigClient, error) {
	cc, err := grpc.Dial("localhost:3001", grpc.WithInsecure())
	if err != nil {
		return ws.ConfigClient{}, err
	}
	return ws.ConfigClient{
		Service: v1.NewServiceServiceClient(cc),
		Route:   v1.NewRouteServiceClient(cc),
		Plugin:  v1.NewPluginServiceClient(cc),

		Node: v1.NewNodeServiceClient(cc),

		Event:  relay.NewEventServiceClient(cc),
		Status: relay.NewStatusServiceClient(cc),
	}, nil
}

func setupDB(logger *zap.Logger, configDB config.Database) (persistence.Persister, error) {
	config := config.ToDBConfig(configDB)
	config.Logger = logger
	m, err := db.NewMigrator(config)
	if err != nil {
		return nil, err
	}
	c, l, err := m.Status()
	if err != nil {
		return nil, err
	}
	logger.Sugar().Debugf("migration status: current: %d, latest: %d", c, l)

	if c != l {
		if configDB.Dialect == db.DialectSQLite3 {
			logger.Sugar().Info("migration out of date")
			logger.Sugar().Info("running migration in-process as sqlite" +
				" database detected")
			err := runMigrations(m)
			if err != nil {
				return nil, err
			}
		} else {
			return nil, fmt.Errorf("database schema out of date, " +
				"please run 'koko db migrate-up' to migrate the schema to" +
				" latest version")
		}
	}

	// setup data store
	return db.NewPersister(config)
}

func runMigrations(m *db.Migrator) error {
	if err := m.Up(); err != nil {
		return fmt.Errorf("migrating database: %v", err)
	}
	return nil
}
