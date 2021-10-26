package cmd

import (
	"context"
	"crypto/tls"

	"github.com/hbagdi/gang"
	v1 "github.com/kong/koko/internal/gen/grpc/kong/admin/service/v1"
	relay "github.com/kong/koko/internal/gen/grpc/kong/relay/service/v1"
	"github.com/kong/koko/internal/persistence"
	"github.com/kong/koko/internal/server"
	"github.com/kong/koko/internal/server/admin"
	"github.com/kong/koko/internal/server/kong/ws"
	relayImpl "github.com/kong/koko/internal/server/relay"
	"github.com/kong/koko/internal/store"
	"github.com/kong/koko/internal/util"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

func Run(ctx context.Context, logger *zap.Logger) error {
	var g gang.Gang

	// setup data store
	memory, err := persistence.NewMemory()
	if err != nil {
		return err
	}
	store := store.New(memory, logger.With(zap.String("component",
		"store"))).ForCluster("default")

	storeLoader := admin.DefaultStoreLoader{Store: store}
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
	relayService := relayImpl.NewEventService(ctx,
		relayImpl.EventServiceOpts{
			Store:  store,
			Logger: logger.With(zap.String("component", "relay-server")),
		})
	relay.RegisterEventServiceServer(rawGRPCServer, relayService)
	if err != nil {
		return err
	}
	g.AddWithCtxE(grpcServer.Run)

	// setup relay client
	configClient, err := setupRelayClient()
	if err != nil {
		return err
	}

	// setup control server
	cert, err := tls.LoadX509KeyPair("cluster.crt", "cluster.key")
	if err != nil {
		return err
	}
	controlLogger := logger.With(zap.String("component", "control-server"))
	m := ws.NewManager(ws.ManagerOpts{
		Logger:  controlLogger,
		Client:  configClient,
		Cluster: ws.DefaultCluster{},
	})
	authFn, err := ws.AuthFnSharedTLS(cert)
	if err != nil {
		return err
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
			Certificates: []tls.Certificate{cert},
			ClientAuth:   tls.RequestClientCert,
		},
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

		Event: relay.NewEventServiceClient(cc),
	}, nil
}
