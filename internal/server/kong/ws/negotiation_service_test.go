package ws

import (
	"context"
	"strings"
	"testing"

	"github.com/kong/go-wrpc/wrpc"
	"github.com/kong/koko/internal/gen/wrpc/kong/model"
	negotiation_service "github.com/kong/koko/internal/gen/wrpc/kong/services/negotiation/v1"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type MockPeer struct {
	mock.Mock
}

func (m *MockPeer) Register(s wrpc.Service) error {
	args := m.Called(s)
	return args.Error(0)
}

func TestRegisterNegotiationService(t *testing.T) {
	negotiator := &Negotiator{}

	testPeer := new(MockPeer)
	testPeer.On("Register", mock.MatchedBy(
		func(s *negotiation_service.NegotiationServiceServer) bool {
			return s.NegotiationService != nil
		})).Return(nil)

	negotiator.Register(testPeer)

	testPeer.AssertExpectations(t)
}

type MockRegisterer struct {
	mock.Mock
}

func (m *MockRegisterer) Register(peer registerer) error {
	args := m.Called(peer)
	return args.Error(0)
}

func TestChooseServiceVersionUnknown(t *testing.T) {
	r := require.New(t)
	testPeer := &wrpc.Peer{}

	testRegisterer := new(MockRegisterer)
	testRegisterer.On("Register", testPeer)

	negotiator := &Negotiator{}
	negotiator.AddService("infundibulum", "chrono-synclastic", "So it goes", testRegisterer)

	ok, choice := negotiator.chooseVersion(&model.ServiceRequest{Name: "gizmo"})
	r.False(ok, "should not find")
	r.Contains(strings.ToLower(choice.message), "unknown")
}

func TestChooseServiceVersionEmpty(t *testing.T) {
	r := require.New(t)
	testPeer := &wrpc.Peer{}

	testRegisterer := new(MockRegisterer)
	testRegisterer.On("Register", testPeer)

	negotiator := &Negotiator{}
	negotiator.AddService("infundibulum", "chrono-synclastic", "So it goes", testRegisterer)

	ok, choice := negotiator.chooseVersion(&model.ServiceRequest{Name: "infundibulum"})
	r.False(ok, "should not find")
	r.Contains(strings.ToLower(choice.message), "no known version")
}

func TestChooseServiceVersionMismatch(t *testing.T) {
	r := require.New(t)
	testPeer := &wrpc.Peer{}

	testRegisterer := new(MockRegisterer)
	testRegisterer.On("Register", testPeer)

	negotiator := &Negotiator{}
	negotiator.AddService("infundibulum", "chrono-synclastic", "So it goes", testRegisterer)

	ok, choice := negotiator.chooseVersion(&model.ServiceRequest{
		Name:     "infundibulum",
		Versions: []string{"hypothalamus"},
	})
	r.False(ok, "should not find")
	r.Contains(strings.ToLower(choice.message), "no known version")
}

func TestChooseServiceVersionFirst(t *testing.T) {
	r := require.New(t)
	testPeer := &wrpc.Peer{}

	testRegisterer := new(MockRegisterer)
	testRegisterer.On("Register", testPeer)

	negotiator := &Negotiator{}
	negotiator.AddService("infundibulum", "chrono-synclastic", "So it goes", testRegisterer)

	ok, choice := negotiator.chooseVersion(&model.ServiceRequest{
		Name:     "infundibulum",
		Versions: []string{"chrono-synclastic"},
	})
	r.True(ok, "should find")
	r.Equal(choice.version, "chrono-synclastic")
	r.Contains(choice.message, "So it goes")
	r.Same(choice.register, testRegisterer)
}

func TestNegotiationInvalid(t *testing.T) {
	r := require.New(t)
	testPeer := &wrpc.Peer{}

	testRegisterer := new(MockRegisterer)
	testRegisterer.On("Register", testPeer)

	negotiator := &Negotiator{CpNodeID: "00A"}
	negotiator.AddService("infundibulum", "chrono-synclastic", "So it goes", testRegisterer)

	req := &model.NegotiateServicesRequest{
		Node: &model.DPNodeDescription{
			Id:      "001",
			Type:    "notKONG",
			Version: "0.00t",
		},
	}
	resp, err := negotiator.NegotiateServices(context.Background(), testPeer, req)
	r.Equal(&model.NegotiateServicesResponse{
		ErrorMessage: "Invalid CP Node Type",
	}, resp)
	r.NoError(err)
}

func TestNegotiation(t *testing.T) {
	t.Run("Empty request", func(t *testing.T) {
		testPeer := &wrpc.Peer{}

		testRegisterer := new(MockRegisterer)
		testRegisterer.On("Register", testPeer)

		negotiator := &Negotiator{CpNodeID: "00A"}
		negotiator.AddService("infundibulum", "chrono-synclastic", "So it goes", testRegisterer)

		req := &model.NegotiateServicesRequest{
			Node: &model.DPNodeDescription{
				Id:      "001",
				Type:    "KONG",
				Version: "0.00t",
			},
		}
		resp, err := negotiator.NegotiateServices(context.Background(), testPeer, req)
		require.NoError(t, err)
		require.Equal(t, &model.CPNodeDescription{Id: "00A"}, resp.Node)
		require.Empty(t, resp.ServicesAccepted)
		require.Empty(t, resp.ServicesRejected)
	})

	t.Run("Unknown empty service requested", func(t *testing.T) {
		testPeer := &wrpc.Peer{}

		testRegisterer := new(MockRegisterer)
		testRegisterer.On("Register", testPeer)

		negotiator := &Negotiator{CpNodeID: "00A"}
		negotiator.AddService("infundibulum", "chrono-synclastic", "So it goes", testRegisterer)

		req := &model.NegotiateServicesRequest{
			Node: &model.DPNodeDescription{
				Id:      "001",
				Type:    "KONG",
				Version: "0.00t",
			},
			ServicesRequested: []*model.ServiceRequest{
				{Name: "gizmo"},
			},
		}
		resp, err := negotiator.NegotiateServices(context.Background(), testPeer, req)
		require.NoError(t, err)
		require.Equal(t, &model.CPNodeDescription{Id: "00A"}, resp.Node)
		require.Empty(t, resp.ServicesAccepted)
		require.Equal(t, 1, len(resp.ServicesRejected))
		require.Equal(t, &model.RejectedService{
			Name:    "gizmo",
			Message: "Unknown service.",
		}, resp.ServicesRejected[0])
	})

	t.Run("Known service, no version match", func(t *testing.T) {
		testPeer := &wrpc.Peer{}

		testRegisterer := new(MockRegisterer)
		testRegisterer.On("Register", testPeer)

		negotiator := &Negotiator{CpNodeID: "00A"}
		negotiator.AddService("infundibulum", "chrono-synclastic", "So it goes", testRegisterer)

		req := &model.NegotiateServicesRequest{
			Node: &model.DPNodeDescription{
				Id:      "001",
				Type:    "KONG",
				Version: "0.00t",
			},
			ServicesRequested: []*model.ServiceRequest{
				{
					Name:     "infundibulum",
					Versions: []string{"hypothalamus"},
				},
			},
		}
		resp, err := negotiator.NegotiateServices(context.Background(), testPeer, req)
		require.NoError(t, err)
		require.Equal(t, &model.CPNodeDescription{Id: "00A"}, resp.Node)
		require.Empty(t, resp.ServicesAccepted)
		require.Equal(t, 1, len(resp.ServicesRejected))
		require.Equal(t, &model.RejectedService{
			Name:    "infundibulum",
			Message: "No known version",
		}, resp.ServicesRejected[0])
	})

	t.Run("One version known, same as requested", func(t *testing.T) {
		testPeer := &wrpc.Peer{}

		testRegisterer := new(MockRegisterer)
		testRegisterer.On("Register", testPeer).Return(nil)

		negotiator := &Negotiator{CpNodeID: "00A"}
		negotiator.AddService("infundibulum", "chrono-synclastic", "So it goes", testRegisterer)

		req := &model.NegotiateServicesRequest{
			Node: &model.DPNodeDescription{
				Id:      "001",
				Type:    "KONG",
				Version: "0.00t",
			},
			ServicesRequested: []*model.ServiceRequest{
				{
					Name:     "infundibulum",
					Versions: []string{"chrono-synclastic"},
				},
			},
		}
		resp, err := negotiator.NegotiateServices(context.Background(), testPeer, req)
		require.NoError(t, err)
		require.Equal(t, &model.CPNodeDescription{Id: "00A"}, resp.Node)
		require.Equal(t, 1, len(resp.ServicesAccepted))
		require.Equal(t, &model.AcceptedService{
			Name:    "infundibulum",
			Version: "chrono-synclastic",
			Message: "So it goes",
		}, resp.ServicesAccepted[0])
		require.Empty(t, resp.ServicesRejected)
	})

	t.Run("Multiple versions requested, one known", func(t *testing.T) {
		testPeer := &wrpc.Peer{}

		testRegisterer := new(MockRegisterer)
		testRegisterer.On("Register", testPeer).Return(nil)

		negotiator := &Negotiator{CpNodeID: "00A"}
		negotiator.AddService("infundibulum", "chrono-synclastic", "So it goes", testRegisterer)

		req := &model.NegotiateServicesRequest{
			Node: &model.DPNodeDescription{
				Id:      "001",
				Type:    "KONG",
				Version: "0.00t",
			},
			ServicesRequested: []*model.ServiceRequest{
				{
					Name:     "infundibulum",
					Versions: []string{"chrono-synclastic", "coquina"},
				},
			},
		}
		resp, err := negotiator.NegotiateServices(context.Background(), testPeer, req)
		require.NoError(t, err)
		require.Equal(t, &model.CPNodeDescription{Id: "00A"}, resp.Node)
		require.Equal(t, 1, len(resp.ServicesAccepted))
		require.Equal(t, &model.AcceptedService{
			Name:    "infundibulum",
			Version: "chrono-synclastic",
			Message: "So it goes",
		}, resp.ServicesAccepted[0])
		require.Empty(t, resp.ServicesRejected)
	})

	t.Run("Multiple matchs, CP chooses which", func(t *testing.T) {
		t.Run("Same order, choose first", func(t *testing.T) {
			testPeer := &wrpc.Peer{}

			testRegisterer := new(MockRegisterer)
			testRegisterer.On("Register", testPeer).Return(nil)

			negotiator := &Negotiator{CpNodeID: "00A"}
			negotiator.AddService("infundibulum", "chrono-synclastic", "So it goes", testRegisterer)
			negotiator.AddService("infundibulum", "coquina", "arbitrii mihi jura mei", testRegisterer)

			req := &model.NegotiateServicesRequest{
				Node: &model.DPNodeDescription{
					Id:      "001",
					Type:    "KONG",
					Version: "0.00t",
				},
				ServicesRequested: []*model.ServiceRequest{
					{
						Name:     "infundibulum",
						Versions: []string{"chrono-synclastic", "coquina"},
					},
				},
			}
			resp, err := negotiator.NegotiateServices(context.Background(), testPeer, req)
			require.NoError(t, err)
			require.Equal(t, &model.CPNodeDescription{Id: "00A"}, resp.Node)
			require.Equal(t, 1, len(resp.ServicesAccepted))
			require.Equal(t, &model.AcceptedService{
				Name:    "infundibulum",
				Version: "chrono-synclastic",
				Message: "So it goes",
			}, resp.ServicesAccepted[0])
			require.Empty(t, resp.ServicesRejected)
		})

		t.Run("Change request, same response", func(t *testing.T) {
			testPeer := &wrpc.Peer{}

			testRegisterer := new(MockRegisterer)
			testRegisterer.On("Register", testPeer).Return(nil)

			negotiator := &Negotiator{CpNodeID: "00A"}
			negotiator.AddService("infundibulum", "chrono-synclastic", "So it goes", testRegisterer)
			negotiator.AddService("infundibulum", "coquina", "arbitrii mihi jura mei", testRegisterer)

			req := &model.NegotiateServicesRequest{
				Node: &model.DPNodeDescription{
					Id:      "001",
					Type:    "KONG",
					Version: "0.00t",
				},
				ServicesRequested: []*model.ServiceRequest{
					{
						Name:     "infundibulum",
						Versions: []string{"coquina", "chrono-synclastic"},
					},
				},
			}
			resp, err := negotiator.NegotiateServices(context.Background(), testPeer, req)
			require.NoError(t, err)
			require.Equal(t, &model.CPNodeDescription{Id: "00A"}, resp.Node)
			require.Equal(t, 1, len(resp.ServicesAccepted))
			require.Equal(t, &model.AcceptedService{
				Name:    "infundibulum",
				Version: "chrono-synclastic",
				Message: "So it goes",
			}, resp.ServicesAccepted[0])
			require.Empty(t, resp.ServicesRejected)
		})

		t.Run("Change priotity, change choice", func(t *testing.T) {
			testPeer := &wrpc.Peer{}

			testRegisterer := new(MockRegisterer)
			testRegisterer.On("Register", testPeer).Return(nil)

			negotiator := &Negotiator{CpNodeID: "00A"}
			negotiator.AddService("infundibulum", "coquina", "arbitrii mihi jura mei", testRegisterer)
			negotiator.AddService("infundibulum", "chrono-synclastic", "So it goes", testRegisterer)

			req := &model.NegotiateServicesRequest{
				Node: &model.DPNodeDescription{
					Id:      "001",
					Type:    "KONG",
					Version: "0.00t",
				},
				ServicesRequested: []*model.ServiceRequest{
					{
						Name:     "infundibulum",
						Versions: []string{"chrono-synclastic", "coquina"},
					},
				},
			}
			resp, err := negotiator.NegotiateServices(context.Background(), testPeer, req)
			require.NoError(t, err)
			require.Equal(t, &model.CPNodeDescription{Id: "00A"}, resp.Node)
			require.Equal(t, 1, len(resp.ServicesAccepted))
			require.Equal(t, &model.AcceptedService{
				Name:    "infundibulum",
				Version: "coquina",
				Message: "arbitrii mihi jura mei",
			}, resp.ServicesAccepted[0])
			require.Empty(t, resp.ServicesRejected)
		})
	})
}
