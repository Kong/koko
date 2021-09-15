package server

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"go.uber.org/zap"
)

var defaultShutdownTimeout = 5 * time.Second

type (
	GRPCServer struct{}
)

type HTTP struct {
	server *http.Server
	logger *zap.Logger
}

type HTTPOpts struct {
	Address string
	Logger  *zap.Logger
	Handler http.Handler
}

func NewHTTP(opts HTTPOpts) (*HTTP, error) {
	if opts.Handler == nil {
		return nil, fmt.Errorf("handler is required")
	}
	return &HTTP{
		server: &http.Server{
			Addr:    opts.Address,
			Handler: opts.Handler,
		},
		logger: opts.Logger,
	}, nil
}

func (h *HTTP) Run(ctx context.Context) error {
	errCh := make(chan error)
	s := h.server
	go func() {
		err := s.ListenAndServe()
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
		ctx, cleanup := context.WithDeadline(context.Background(),
			time.Now().Add(defaultShutdownTimeout))
		defer cleanup()
		err := s.Shutdown(ctx)
		if err != nil {
			return err
		}
	}
	return nil
}
