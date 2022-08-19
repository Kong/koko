package ws_test

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/url"
	"sync"
	"testing"

	"github.com/kong/go-wrpc/wrpc"
	"github.com/kong/koko/internal/gen/wrpc/kong/model"
	config_service "github.com/kong/koko/internal/gen/wrpc/kong/services/config/v1"
	nego "github.com/kong/koko/internal/gen/wrpc/kong/services/negotiation/v1"
	"github.com/kong/koko/internal/test/certs"
	"github.com/kong/koko/internal/test/run"
	"github.com/kong/koko/internal/test/util"
	"github.com/stretchr/testify/require"
)

const (
	nodeID = "758435aa-bab5-4786-92d1-c509ac520c2d"
)

// wprcConn connects to koko's wrpc endpoint.
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
			"node_id":       {nodeID},
			"node_hostname": {"localhost"},
			"node_version":  {"3.0.0"},
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
	defer peer.Close()

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

// trivial config service mock to log calls.

type configMock struct {
	lock sync.Mutex
	log  []string
}

func (cm *configMock) reset() {
	cm.lock.Lock()
	defer cm.lock.Unlock()
	cm.log = []string{}
}

func (cm *configMock) requireCalls(t *testing.T, l []string) {
	err := fmt.Errorf("less than %d calls", len(l))

	util.WaitFunc(t, func() error {
		cm.lock.Lock()
		defer cm.lock.Unlock()

		if len(cm.log) < len(l) {
			return err
		}
		require.Equal(t, l, cm.log[:len(l)])
		return nil
	})
}

// implement all config service's RPCs, just log their names

func (cm *configMock) GetCapabilities(
	ctx context.Context,
	p *wrpc.Peer,
	req *config_service.GetCapabilitiesRequest,
) (resp *config_service.GetCapabilitiesResponse, err error) {
	cm.lock.Lock()
	defer cm.lock.Unlock()

	cm.log = append(cm.log, "GetCapabilities")
	return
}

func (cm *configMock) PingCP(
	ctx context.Context,
	p *wrpc.Peer,
	req *config_service.PingCPRequest,
) (resp *config_service.PingCPResponse, err error) {
	cm.lock.Lock()
	defer cm.lock.Unlock()

	cm.log = append(cm.log, "PingCP")
	return
}

func (cm *configMock) ReportMetadata(
	ctx context.Context,
	p *wrpc.Peer,
	req *config_service.ReportMetadataRequest,
) (resp *config_service.ReportMetadataResponse, err error) {
	cm.lock.Lock()
	defer cm.lock.Unlock()

	cm.log = append(cm.log, "ReportMetadata")
	return
}

func (cm *configMock) SyncConfig(
	ctx context.Context,
	p *wrpc.Peer,
	req *config_service.SyncConfigRequest,
) (resp *config_service.SyncConfigResponse, err error) {
	cm.lock.Lock()
	defer cm.lock.Unlock()

	cm.log = append(cm.log, "SyncConfig")
	return
}

func TestConfigService(t *testing.T) {
	cleanup := run.Koko(t)
	defer cleanup()

	conn := wprcConn(t)
	require.NotNil(t, conn)

	peer := &wrpc.Peer{}
	peer.AddConn(conn)
	defer peer.Close()

	configMock := configMock{}
	peer.Register(&nego.NegotiationServiceServer{})
	peer.Register(&config_service.ConfigServiceServer{
		ConfigService: &configMock,
	})

	negotiationClient := nego.NegotiationServiceClient{Peer: peer}
	resp, err := negotiationClient.NegotiateServices(context.Background(), &model.NegotiateServicesRequest{
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

	configClient := config_service.ConfigServiceClient{Peer: peer}

	t.Run("send empty initial report message, fail validation", func(t *testing.T) {
		resp, err := configClient.ReportMetadata(context.Background(), &config_service.ReportMetadataRequest{})
		require.ErrorContains(t, err, "node failed to meet pre-requisites")
		require.Nil(t, resp)
	})

	t.Run("send some acceptable plugins, get an initial config", func(t *testing.T) {
		configMock.reset()
		resp, err := configClient.ReportMetadata(context.Background(), &config_service.ReportMetadataRequest{
			Plugins: []*config_service.PluginVersion{
				{
					Name: "rate-limiting",
				},
			},
		})
		require.NoError(t, err)
		require.EqualValues(t, &config_service.ReportMetadataResponse_Ok{Ok: "valid"}, resp.Response)
		configMock.requireCalls(t, []string{"SyncConfig"})
	})
}
