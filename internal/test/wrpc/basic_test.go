package wrpc_test

import (
	"context"
	"crypto/tls"
	"net/url"
	"testing"

	"github.com/kong/go-wrpc/wrpc"
	"github.com/kong/koko/internal/gen/wrpc/kong/model"
	config_service "github.com/kong/koko/internal/gen/wrpc/kong/services/config/v1"
	nego "github.com/kong/koko/internal/gen/wrpc/kong/services/negotiation/v1"
	"github.com/kong/koko/internal/test/certs"
	"github.com/kong/koko/internal/test/run"
	"github.com/stretchr/testify/require"
)

const (
	node_id = "758435aa-bab5-4786-92d1-c509ac520c2d"
)

func wprcConn(t *testing.T) *wrpc.Conn {
	dialer := *wrpc.DefaultDialer

	cert, err := tls.X509KeyPair(certs.DefaultSharedCert, certs.DefaultSharedKey)
	require.NoError(t, err)

	dialer.Dialer.TLSClientConfig = &tls.Config{
		InsecureSkipVerify: true,
		Certificates:       []tls.Certificate{cert},
		ServerName:         "kong_clustering",
	}

	u := url.URL{
		Scheme: "wss",
		Host:   "localhost:3100",
		Path:   "/v1/wrpc",
		RawQuery: url.Values{
			"node_id":       {node_id},
			"node_hostname": {"localhost"},
			"node_version":  {"3.0"},
		}.Encode(),
	}
	c, _, err := dialer.Dial(context.Background(), u.String(), nil)
	require.NoError(t, err)
	require.NotNil(t, c)

	return c
}

func TestConnectWRPC(t *testing.T) {
	cleanup := run.Koko(t)
	defer cleanup()

	conn := wprcConn(t)
	require.NotNil(t, conn)
}

func TestNegotiationService(t *testing.T) {
	cleanup := run.Koko(t)
	defer cleanup()

	conn := wprcConn(t)
	require.NotNil(t, conn)

	peer := &wrpc.Peer{}
	peer.AddConn(conn)
	peer.Register(&nego.NegotiationServiceServer{})

	cli := nego.NegotiationServiceClient{Peer: peer}

	t.Run("Empty Negotiation Request", func(t *testing.T) {
		resp, err := cli.NegotiateServices(context.Background(), &model.NegotiateServicesRequest{})
		require.NoError(t, err)
		require.NotEmpty(t, resp.ErrorMessage)
		require.Empty(t, resp.ServicesAccepted)
	})

	t.Run("minimal negotiation request", func(t *testing.T) {
		resp, err := cli.NegotiateServices(context.Background(), &model.NegotiateServicesRequest{
			Node: &model.DPNodeDescription{
				Type: "KONG",
			},
		})
		require.NoError(t, err)
		require.Empty(t, resp.ErrorMessage)
		require.Empty(t, resp.ServicesAccepted)
	})
}

func TestConfigService(t *testing.T) {
	cleanup := run.Koko(t)
	defer cleanup()

	conn := wprcConn(t)
	require.NotNil(t, conn)

	peer := &wrpc.Peer{}
	peer.AddConn(conn)
	peer.Register(&nego.NegotiationServiceServer{})
	peer.Register(&config_service.ConfigServiceServer{})

	cli_nego := nego.NegotiationServiceClient{Peer: peer}
	resp, err := cli_nego.NegotiateServices(context.Background(), &model.NegotiateServicesRequest{
		Node: &model.DPNodeDescription{
			Type: "KONG",
		},
		ServicesRequested: []*model.ServiceRequest{
			{
				Name:     "config",
				Versions: []string{"v1"},
			},
		},
	})
	require.NoError(t, err)
	require.Empty(t, resp.ErrorMessage)
	require.ElementsMatch(t, []*model.AcceptedService{
		{
			Name:    "config",
			Version: "v1",
			Message: "wRPC configuration",
		},
	}, resp.ServicesAccepted)

	cli_config := config_service.ConfigServiceClient{Peer: peer}

	t.Run("empty initial report message", func(t *testing.T) {
		resp, err := cli_config.ReportMetadata(context.Background(), &config_service.ReportMetadataRequest{})
		require.NoError(t, err)
		require.Empty(t, resp)
	})
}
