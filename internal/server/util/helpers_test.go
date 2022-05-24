package util

import (
	"context"
	"errors"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gavv/httpexpect/v2"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/test/bufconn"
)

func TestHandlerWithRecovery(t *testing.T) {
	t.Run("gracefully handles panics", func(t *testing.T) {
		l, err := zap.NewProduction()
		require.NoError(t, err)
		s := httptest.NewServer(HandlerWithRecovery(thisWillPanic(), l))
		defer s.Close()
		c := httpexpect.New(t, s.URL)
		res := c.POST("/v1/resource/id").WithHeader("content-type", "application/json").Expect()
		res.Status(http.StatusInternalServerError)
		res.ContentType("application/json")
		require.Equal(t, `{"message":"internal server error"}`, res.Body().Raw())
	})
}

func thisWillPanic() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		panic(errors.New("something bad happened"))
	})
}

func TestPanicInterceptor(t *testing.T) {
	logger, err := zap.NewProduction()
	require.NoError(t, err)
	serverOpts := []grpc.ServerOption{
		grpc.ChainUnaryInterceptor(LoggerInterceptor(logger), PanicInterceptor(logger)),
		grpc.ChainStreamInterceptor(PanicStreamInterceptor(logger)),
	}
	grpcServer := grpc.NewServer(serverOpts...)

	desc := &grpc.ServiceDesc{
		ServiceName: "TestService",
		HandlerType: (*TestServiceServer)(nil),
		Methods: []grpc.MethodDesc{
			{
				MethodName: "PanicTest",
				Handler:    panicTestHandler,
			},
		},
		Streams: []grpc.StreamDesc{
			{
				StreamName:    "PanicStreamTest",
				Handler:       panicStreamTestHandler,
				ServerStreams: true,
				ClientStreams: true,
			},
		},
	}
	grpcServer.RegisterService(desc, &TestService{})
	l := setupBufConn()
	go func() {
		_ = grpcServer.Serve(l)
	}()
	defer grpcServer.Stop()
	ctx := context.Background()

	t.Run("intercepts panics", func(t *testing.T) {
		cc := clientConn(ctx, t, l)
		err := cc.Invoke(ctx, "TestService/PanicTest", nil, nil)
		require.ErrorContains(t, err, "rpc error: code = Internal desc = internal server error")
	})

	t.Run("intercepts panics for streams", func(t *testing.T) {
		cc := clientConn(ctx, t, l)
		cs, err := cc.NewStream(ctx, &desc.Streams[0], "TestService/PanicStreamTest")
		require.NoError(t, err)
		err = cs.RecvMsg(struct{}{})
		require.ErrorContains(t, err, "rpc error: code = Internal desc = internal server error")
	})
}

type TestServiceServer interface {
	PanicTest() (interface{}, error)
	PanicStreamTest() error
}

type TestService struct {
	grpc.ServerStream
}

func (ts *TestService) PanicTest() (interface{}, error) {
	panic(errors.New("something bad happened"))
}

// nolint:revive // ctx must be second to satisfy interface
func panicTestHandler(srv interface{}, ctx context.Context, _ func(interface{}) error,
	interceptor grpc.UnaryServerInterceptor,
) (interface{}, error) {
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(TestServiceServer).PanicTest()
	}
	return interceptor(ctx, struct{}{}, &grpc.UnaryServerInfo{Server: srv}, handler)
}

func (ts *TestService) PanicStreamTest() error {
	panic(errors.New("something bad happened"))
}

func panicStreamTestHandler(srv interface{}, _ grpc.ServerStream) error {
	return srv.(TestServiceServer).PanicStreamTest()
}

func setupBufConn() *bufconn.Listener {
	const bufSize = 1024 * 1024
	return bufconn.Listen(bufSize)
}

func clientConn(ctx context.Context, t *testing.T, l *bufconn.Listener) grpc.ClientConnInterface {
	conn, err := grpc.DialContext(ctx,
		"bufnet",
		grpc.WithContextDialer(func(ctx context.Context, s string) (net.Conn, error) {
			return l.Dial()
		}),
		grpc.WithTransportCredentials(insecure.NewCredentials()))
	require.Nil(t, err)
	return conn
}
