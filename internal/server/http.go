package server

import (
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"strings"
	"time"

	"go.uber.org/zap"
)

const tlsHandshakeError = "http: TLS handshake error"

var defaultShutdownTimeout = 5 * time.Second

type HTTP struct {
	server *http.Server
	logger *zap.Logger
}

type HTTPOpts struct {
	Address string
	Logger  *zap.Logger
	Handler http.Handler
	// TLS protocol is used when this field is not nil.
	TLS *tls.Config
}

func NewHTTP(opts HTTPOpts) (*HTTP, error) {
	if opts.Handler == nil {
		return nil, fmt.Errorf("handler is required")
	}
	return &HTTP{
		server: &http.Server{
			Addr:      opts.Address,
			Handler:   opts.Handler,
			TLSConfig: opts.TLS,
		},
		logger: opts.Logger,
	}, nil
}

type tlsErrorWriter struct {
	io.Writer
}

func (w *tlsErrorWriter) Write(p []byte) (int, error) {
	if strings.HasPrefix(string(p), tlsHandshakeError) {
		return len(p), nil
	}
	// for non tls handshake error, log it as usual
	return w.Writer.Write(p)
}

func (h *HTTP) addTLSHandshakeErrorHandler() {
	tlsErrorWriter := &tlsErrorWriter{os.Stderr}
	tlsErrorLogger := log.New(tlsErrorWriter, "", 0)
	h.server.ErrorLog = tlsErrorLogger
}

func (h *HTTP) Run(ctx context.Context) error {
	errCh := make(chan error)
	s := h.server
	h.addTLSHandshakeErrorHandler()
	go func() {
		h.logger.Debug("starting server")
		listener, err := net.Listen("tcp", h.server.Addr)
		if err != nil {
			errCh <- err
			return
		}
		if h.server.TLSConfig != nil {
			listener = tls.NewListener(listener, h.server.TLSConfig)
		}
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
		ctx, cleanup := context.WithDeadline(context.Background(),
			time.Now().Add(defaultShutdownTimeout))
		defer cleanup()
		// ctx not inheritted since the parent ctx will already be Done()
		// at this point
		err := s.Shutdown(ctx) //nolint:contextcheck
		if err != nil {
			return err
		}
	}
	return nil
}
