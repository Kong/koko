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

// Registerer is the object that handles registering
// a service to a wrpc.Peer
// A concrete implementation of this interface should hold
// any extra information the service will need.
type Registerer interface {
	Register(peer *wrpc.Peer) error
}

type serviceVersion struct {
	version  string
	message  string
	register Registerer
}

// Negotiator handles service negotiation.
// It holds a map of services, each with a list of
// known versions and respective registerers.
type Negotiator struct {
	Cluster       Cluster
	Logger        *zap.Logger
	knownVersions map[string][]serviceVersion
}

// Associates a service name and version with
// a registerer object and a descriptive message.
// To be used during startup to define which
// services are available on a server.
func (n *Negotiator) AddService(
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
func (n *Negotiator) Register(peer *wrpc.Peer) error {
	return peer.Register(
		&nego.NegotiationServiceServer{
			NegotiationService: n,
		})
}

// Choose the best version for a requested service.
func (n *Negotiator) chooseVersion(requestedServ *model.ServiceRequest) (choice serviceVersion, ok bool) {
	known, ok := n.knownVersions[requestedServ.Name]
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
// The response to the client includes:
//    - information about the node (just the node ID for now).
//    - for each requested service, it's either
//      - in the accepted list:
//        - respond with the version and a description.
//        - call the registerer object associated with that service/version
//          to activate the right responses on this specific peer.
//      - in the rejected list:
//        - with a message relevant to the reason (unknown or disabled
//          service, bad versions).
func (n *Negotiator) NegotiateServices(
	_ context.Context,
	peer *wrpc.Peer,
	req *model.NegotiateServicesRequest,
) (resp *model.NegotiateServicesResponse, err error) {
	resp = &model.NegotiateServicesResponse{
		Node:             &model.CPNodeDescription{Id: n.Cluster.Get()},
		ServicesAccepted: []*model.AcceptedService{},
		ServicesRejected: []*model.RejectedService{},
	}

	logger := n.Logger
	if logger == nil {
		logger = zap.L().With(zap.String("cluster-id", n.Cluster.Get()))
	}

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
		choice, ok := n.chooseVersion(requestedServ)
		if ok {
			resp.ServicesAccepted = append(resp.ServicesAccepted, &model.AcceptedService{
				Name:    requestedServ.Name,
				Version: choice.version,
				Message: choice.message,
			})
			err := choice.register.Register(peer)
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
