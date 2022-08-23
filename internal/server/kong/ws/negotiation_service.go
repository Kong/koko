package ws

import (
	"context"
	"fmt"

	"github.com/kong/go-wrpc/wrpc"
	"github.com/kong/koko/internal/gen/wrpc/kong/model"
	nego "github.com/kong/koko/internal/gen/wrpc/kong/services/negotiation/v1"
	"go.uber.org/zap"
)

const (
	nodeTypeKong = "KONG"

	invalidCPNodeType     = "Invalid CP Node Type"
	unknownServiceMessage = "Unknown service."
	noKnownVersionMessage = "No known version"
)

// Registerer is the object that handles registering a service to a wrpc.Peer
// A concrete implementation of this interface should hold any extra
// information the service will need to instantiate the service object,
// except from the Manager, which is provided on the Register method call.
type Registerer interface {
	// Register should be called when the negotiator chooses
	// a specific service.
	Register(peer *wrpc.Peer, m *Manager) error
}

type serviceVersion struct {
	version  string
	message  string
	register Registerer
}

// NegotiationRegisterer holds a map of services, each with a list of
// known versions and respective registerers.
type negotiationRegisterer struct {
	Logger        *zap.Logger
	knownVersions map[string][]serviceVersion
}

func NewNegotiationRegisterer(logger *zap.Logger) (*negotiationRegisterer, error) {
	if logger == nil {
		return nil, fmt.Errorf("NegotiationRegisterer requires a logger")
	}
	return &negotiationRegisterer{
		Logger: logger,
	}, nil
}

// AddService associates a service name and version
// with a registerer object and a descriptive message.
// To be used during startup to define which
// services are available on a server.
func (n *negotiationRegisterer) AddService(
	serviceName, version, message string,
	register Registerer,
) error {
	if n.knownVersions == nil {
		n.knownVersions = map[string][]serviceVersion{}
	}

	knownServ, ok := n.knownVersions[serviceName]
	if !ok {
		knownServ = []serviceVersion{}
	}

	for _, knownVersion := range knownServ {
		if knownVersion.version == version {
			return fmt.Errorf("%s.%s already registered", serviceName, version)
		}
	}

	n.knownVersions[serviceName] = append(knownServ, serviceVersion{
		version:  version,
		message:  message,
		register: register,
	})

	return nil
}

// Register adds the version negotiation service to the peer.
func (n *negotiationRegisterer) Register(peer *wrpc.Peer, m *Manager) error {
	return peer.Register(
		&nego.NegotiationServiceServer{
			NegotiationService: &negotiationService{
				manager:    m,
				registerer: n,
			},
		})
}

// Each negotiationService handles service negotiation for a given cluster.
// Keeps a link to the NegotiationRegisterer with the map of services.
type negotiationService struct {
	manager    *Manager
	registerer *negotiationRegisterer
}

// chooseVersion selects the best version for a requested service.
func (ns *negotiationService) chooseVersion(requestedServ *model.ServiceRequest) (choice serviceVersion, ok bool) {
	known, ok := ns.registerer.knownVersions[requestedServ.Name]
	if !ok {
		return serviceVersion{message: unknownServiceMessage}, false
	}

	for _, knownVers := range known {
		for _, reqVers := range requestedServ.Versions {
			if reqVers == knownVers.version {
				return knownVers, true
			}
		}
	}

	return serviceVersion{message: noKnownVersionMessage}, false
}

// NegotiateServices is the method handler for the only RPC in this service.
// The response to the client includes information about the node (only ID
// currently) and each requested service.
//
// For a service in the accepted list, respond with the version and a description and
// call the registerer object associated with that service/version to activate
// the right responses on this specific peer.
// For a service in the rejected list, respond with a message relevant to the reason
// (unknown or disabled service, bad versions).
func (ns *negotiationService) NegotiateServices(
	_ context.Context,
	peer *wrpc.Peer,
	req *model.NegotiateServicesRequest,
) (resp *model.NegotiateServicesResponse, err error) {
	cpNodeID := ns.manager.Cluster.Get()
	resp = &model.NegotiateServicesResponse{
		Node:             &model.CPNodeDescription{Id: cpNodeID},
		ServicesAccepted: []*model.AcceptedService{},
		ServicesRejected: []*model.RejectedService{},
	}

	logger := ns.registerer.Logger.With(zap.String("cluster-id", cpNodeID))

	if req.Node == nil {
		logger.Error("Missing Node information")
		return &model.NegotiateServicesResponse{
			ErrorMessage: invalidCPNodeType,
		}, nil
	}

	if req.Node.Type != nodeTypeKong {
		logger.Error("Invalid Node type", zap.String("type", req.Node.Type))
		return &model.NegotiateServicesResponse{
			ErrorMessage: invalidCPNodeType,
		}, nil
	}

	for _, requestedServ := range req.ServicesRequested {
		choice, ok := ns.chooseVersion(requestedServ)
		if ok {
			resp.ServicesAccepted = append(resp.ServicesAccepted, &model.AcceptedService{
				Name:    requestedServ.Name,
				Version: choice.version,
				Message: choice.message,
			})
			err := choice.register.Register(peer, ns.manager)
			if err != nil {
				return nil, fmt.Errorf("error registering service %s, version %s: %w",
					requestedServ.Name, choice.version, err)
			}
			logger.Info("Service accepted",
				zap.String("service", requestedServ.Name),
				zap.String("version", choice.version),
				zap.String("message", choice.message),
			)
		} else {
			resp.ServicesRejected = append(resp.ServicesRejected, &model.RejectedService{
				Name:    requestedServ.Name,
				Message: choice.message,
			})
		}
	}

	return resp, nil
}
