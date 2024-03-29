package server

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"time"

	"go.uber.org/zap"
	"google.golang.org/grpc"
)

type (
	GRPCServer struct{}
)

type GrpcInterceptorInjector interface {
	Handle(ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler) (resp interface{}, err error)
}

type GRPC struct {
	server  *grpc.Server
	logger  *zap.Logger
	address string
}

type GRPCOpts struct {
	Address    string
	Logger     *zap.Logger
	GRPCServer *grpc.Server
}

func NewGRPC(opts GRPCOpts) (*GRPC, error) {
	if opts.GRPCServer == nil {
		return nil, fmt.Errorf("GRPCServer is required")
	}
	return &GRPC{
		address: opts.Address,
		server:  opts.GRPCServer,
		logger:  opts.Logger.With(zap.String("address", opts.Address)),
	}, nil
}

func (g *GRPC) Run(ctx context.Context) error {
	errCh := make(chan error)
	s := g.server
	go func() {
		g.logger.Info("starting server")
		listener, err := net.Listen("tcp", g.address)
		if err != nil {
			errCh <- err
			return
		}
		// TODO(hbagdi): figure out TLS details
		err = s.Serve(listener)
		if err != nil {
			if err != http.ErrServerClosed {
				errCh <- err
			}
		}
	}()

	select {
	case err := <-errCh:
		return err
	case <-ctx.Done():
		// Guarantee the server stops. GracefulStop will never close
		// when a client is holding open a streaming connection.
		stopped := make(chan struct{})
		go func() {
			g.logger.Info("attempting to gracefully stop GRPC server")
			s.GracefulStop()
			close(stopped)
		}()
		t := time.NewTimer(defaultShutdownTimeout)
		select {
		case <-t.C:
			g.logger.Info("stopping GRPC server")
			s.Stop()
		case <-stopped:
			t.Stop()
		}
	}
	return nil
}
