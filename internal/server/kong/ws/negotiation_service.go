package ws

import (
	"context"

	"github.com/kong/go-wrpc/wrpc"
	"github.com/kong/koko/internal/gen/wrpc/kong/model"
	negotiation_service "github.com/kong/koko/internal/gen/wrpc/kong/services/negotiation/v1"
	"go.uber.org/zap"
)

const (
	invalidCPNodeType     = "Invalid CP Node Type"
	unknownServiceMessage = "Unknown service."
	noKnownVersionMessage = "No known version"
)

// a registerer is any object which can receive (or register)
// a wrpc.Service.  Basically a wrpc.Peer, but in theory
// it could also be a shared registry if we ever make one.
type registerer interface {
	Register(s wrpc.Service) error
}

// HandleRegisterer is the object that handles registering
// a service to a registerer (a wrpc.Peer)
// A concrete implementation of this interface should hold
// any extra information the service will need.
type HandleRegisterer interface {
	Register(peer registerer) error
}

type knownVersion struct {
	version  string
	message  string
	register HandleRegisterer
}

// The Negotiation type handles service negotiation.
// It holds a map of services, each with a list of
// known versions and respective registerers.
type Negotiator struct {
	CpNodeID      string
	KnownVersions map[string][]knownVersion
	Logger        *zap.Logger
}

// Associates a service name and version with
// a registerer object and a descriptive message.
func (n *Negotiator) AddService(
	serviceName, version, message string,
	register HandleRegisterer,
) {
	if n.KnownVersions == nil {
		n.KnownVersions = map[string][]knownVersion{}
	}

	knownServ, ok := n.KnownVersions[serviceName]
	if !ok {
		knownServ = []knownVersion{}
	}
	knownServ = append(knownServ, knownVersion{
		version:  version,
		message:  message,
		register: register,
	})
	n.KnownVersions[serviceName] = knownServ
}

// Register adds the version negotiation service to the peer.
func (n *Negotiator) Register(peer registerer) error {
	return peer.Register(
		&negotiation_service.NegotiationServiceServer{
			NegotiationService: n,
		})
}

// Choose the best version for a requested service.
func (n *Negotiator) chooseVersion(requestedServ *model.ServiceRequest) (ok bool, choice knownVersion) {
	known, ok := n.KnownVersions[requestedServ.Name]
	if !ok {
		return false, knownVersion{message: unknownServiceMessage}
	}

	for _, knownVers := range known {
		for _, reqVers := range requestedServ.Versions {
			if reqVers == knownVers.version {
				return true, knownVers
			}
		}
	}

	return false, knownVersion{message: noKnownVersionMessage}
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
	ctx context.Context,
	peer *wrpc.Peer,
	req *model.NegotiateServicesRequest,
) (resp *model.NegotiateServicesResponse, err error) {
	resp = &model.NegotiateServicesResponse{
		Node:             &model.CPNodeDescription{Id: n.CpNodeID},
		ServicesAccepted: []*model.AcceptedService{},
		ServicesRejected: []*model.RejectedService{},
	}

	logger := n.Logger
	if logger == nil {
		logger = zap.L()
	}

	if req.Node == nil {
		logger.Error("Missing Node information")
		return &model.NegotiateServicesResponse{
			ErrorMessage: invalidCPNodeType,
		}, nil
	}

	if req.Node.Type != "KONG" {
		logger.Error("Invalid Node type", zap.String("type", req.Node.Type))
		return &model.NegotiateServicesResponse{
			ErrorMessage: invalidCPNodeType,
		}, nil
	}

	for _, requestedServ := range req.ServicesRequested {
		ok, choice := n.chooseVersion(requestedServ)
		if ok {
			resp.ServicesAccepted = append(resp.ServicesAccepted, &model.AcceptedService{
				Name:    requestedServ.Name,
				Version: choice.version,
				Message: choice.message,
			})
			err := choice.register.Register(peer)
			if err != nil {
				return nil, err // TODO: should we mask the error?
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
