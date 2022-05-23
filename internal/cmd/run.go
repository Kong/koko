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
	"github.com/kong/koko/internal/metrics"
	"github.com/kong/koko/internal/persistence"
	"github.com/kong/koko/internal/plugin"
	"github.com/kong/koko/internal/plugin/validators"
	"github.com/kong/koko/internal/resource"
	"github.com/kong/koko/internal/server"
	"github.com/kong/koko/internal/server/admin"
	"github.com/kong/koko/internal/server/health"
	"github.com/kong/koko/internal/server/kong/ws"
	kongConfigWS "github.com/kong/koko/internal/server/kong/ws/config"
	"github.com/kong/koko/internal/server/kong/ws/config/compat"
	relayImpl "github.com/kong/koko/internal/server/relay"
	serverUtil "github.com/kong/koko/internal/server/util"
	"github.com/kong/koko/internal/store"
	"github.com/kong/koko/internal/util"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type ServerConfig struct {
	DPAuthMode    DPAuthMode
	DPAuthCert    tls.Certificate
	DPAuthCACerts []*x509.Certificate

	KongCPCert tls.Certificate

	Logger   *zap.Logger
	Metrics  config.Metrics
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

	err := metrics.InitMetricsClient(logger.With(zap.String("component", "metrics-collector")), config.Metrics.ClientType)
	if err != nil {
		return fmt.Errorf("init metrics client failure: %w", err)
	}

	defer metrics.Close()
	if config.Metrics.ClientType == metrics.Prometheus.String() {
		metricsHandler, err := metrics.CreateHandler(logger)
		if err != nil {
			return fmt.Errorf("create metrics handler failure: %w", err)
		}
		s, err := server.NewHTTP(server.HTTPOpts{
			Address: ":9090",
			Logger:  logger.With(zap.String("component", "metrics")),
			Handler: metricsHandler,
		})
		if err != nil {
			return err
		}
		g.AddWithCtxE(s.Run)
	}

	persister, err := setupDB(logger, config.Database)
	if err != nil {
		return fmt.Errorf("database: %v", err)
	}
	defer persister.Close()

	store := store.New(persister, logger.With(zap.String("component",
		"store"))).ForCluster("default")
	storeLoader := serverUtil.DefaultStoreLoader{Store: store}

	validator, err := validators.NewLuaValidator(validators.Opts{
		Logger:      logger,
		StoreLoader: storeLoader,
	})
	if err != nil {
		return err
	}
	err = validator.LoadSchemasFromEmbed(plugin.Schemas, "schemas")
	if err != nil {
		return err
	}
	resource.SetValidator(validator)

	adminOpts := admin.HandlerOpts{
		Logger:      logger.With(zap.String("component", "admin-server")),
		StoreLoader: storeLoader,
		Validator:   validator,
	}

	// Validate the handler options & set up the admin API handler.
	h, err := admin.NewHandler(adminOpts)
	if err != nil {
		return err
	}

	// setup Admin API server
	s, err := server.NewHTTP(server.HTTPOpts{
		Address: ":3000",
		Logger:  adminOpts.Logger,
		Handler: serverUtil.HandlerWithLogger(h, adminOpts.Logger),
	})
	if err != nil {
		return err
	}
	g.AddWithCtxE(s.Run)

	// Set up relay server using the same opts as the admin API server.
	rawGRPCServer := grpc.NewServer(
		grpc.UnaryInterceptor(serverUtil.LoggerInterceptor(adminOpts.Logger)),
	)
	admin.RegisterAdminService(rawGRPCServer, adminOpts)

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
	grpcClients, err := setupGRPCClients()
	if err != nil {
		return err
	}

	loader := &kongConfigWS.KongConfigurationLoader{}
	err = loader.Register(&kongConfigWS.KongServiceLoader{Client: grpcClients.
		Service})
	if err != nil {
		panic(err.Error())
	}
	err = loader.Register(&kongConfigWS.KongRouteLoader{Client: grpcClients.Route})
	if err != nil {
		panic(err.Error())
	}
	err = loader.Register(&kongConfigWS.KongPluginLoader{Client: grpcClients.Plugin})
	if err != nil {
		panic(err.Error())
	}

	err = loader.Register(&kongConfigWS.KongUpstreamLoader{Client: grpcClients.Upstream})
	if err != nil {
		panic(err.Error())
	}

	err = loader.Register(&kongConfigWS.KongTargetLoader{Client: grpcClients.Target})
	if err != nil {
		panic(err.Error())
	}

	err = loader.Register(&kongConfigWS.KongConsumerLoader{Client: grpcClients.Consumer})
	if err != nil {
		panic(err.Error())
	}

	err = loader.Register(&kongConfigWS.KongCertificateLoader{Client: grpcClients.Certificate})
	if err != nil {
		panic(err.Error())
	}

	err = loader.Register(&kongConfigWS.KongCACertificateLoader{Client: grpcClients.CACertificate})
	if err != nil {
		panic(err.Error())
	}

	err = loader.Register(&kongConfigWS.KongSNILoader{Client: grpcClients.SNI})
	if err != nil {
		panic(err.Error())
	}

	err = loader.Register(&kongConfigWS.VersionLoader{})
	if err != nil {
		panic(err.Error())
	}

	// setup version compatibility processor
	vcLogger := logger.With(zap.String("component", "version-compatibility"))
	vc, err := kongConfigWS.NewVersionCompatibilityProcessor(kongConfigWS.VersionCompatibilityOpts{
		Logger:        vcLogger,
		KongCPVersion: kongConfigWS.KongGatewayCompatibilityVersion,
	})
	if err != nil {
		panic(err.Error())
	}
	if err := vc.AddConfigTableUpdates(compat.PluginConfigTableUpdates); err != nil {
		panic(err.Error())
	}
	vcLogger.With(zap.String("control-plane", kongConfigWS.KongGatewayCompatibilityVersion)).
		Info("Lua control plane compatibility version")

	// setup control server
	controlLogger := logger.With(zap.String("component", "control-server"))
	m := ws.NewManager(ws.ManagerOpts{
		Logger:                 controlLogger,
		DPConfigLoader:         loader,
		DPVersionCompatibility: vc,
		Client: ws.ConfigClient{
			Node:   grpcClients.Node,
			Status: grpcClients.Status,
			Event:  grpcClients.Event,
		},
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
		Handler: serverUtil.HandlerWithLogger(handler, controlLogger),
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

type grpcClients struct {
	Service       v1.ServiceServiceClient
	Route         v1.RouteServiceClient
	Plugin        v1.PluginServiceClient
	PluginSchema  v1.PluginSchemaServiceClient
	Upstream      v1.UpstreamServiceClient
	Target        v1.TargetServiceClient
	Consumer      v1.ConsumerServiceClient
	Certificate   v1.CertificateServiceClient
	CACertificate v1.CACertificateServiceClient
	SNI           v1.SNIServiceClient

	Status relay.StatusServiceClient
	Node   v1.NodeServiceClient
	Event  relay.EventServiceClient
}

func setupGRPCClients() (grpcClients, error) {
	cc, err := grpc.Dial("localhost:3001",
		grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return grpcClients{}, err
	}
	return grpcClients{
		Service:       v1.NewServiceServiceClient(cc),
		Route:         v1.NewRouteServiceClient(cc),
		Plugin:        v1.NewPluginServiceClient(cc),
		PluginSchema:  v1.NewPluginSchemaServiceClient(cc),
		Upstream:      v1.NewUpstreamServiceClient(cc),
		Target:        v1.NewTargetServiceClient(cc),
		Consumer:      v1.NewConsumerServiceClient(cc),
		Certificate:   v1.NewCertificateServiceClient(cc),
		CACertificate: v1.NewCACertificateServiceClient(cc),
		SNI:           v1.NewSNIServiceClient(cc),

		Node:   v1.NewNodeServiceClient(cc),
		Event:  relay.NewEventServiceClient(cc),
		Status: relay.NewStatusServiceClient(cc),
	}, nil
}

func setupDB(logger *zap.Logger, configDB config.Database) (persistence.Persister, error) {
	config, err := config.ToDBConfig(configDB)
	if err != nil {
		logger.Fatal(err.Error())
	}

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
