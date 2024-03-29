// Code generated by protoc-gen-go-wrpc. DO NOT EDIT
// protoc-gen-go-wrpc version: v0.0.0-20220926162517-2374aa556d56

package v1

import (
	context "context"
	wrpc "github.com/kong/go-wrpc/wrpc"
	model "github.com/kong/koko/internal/gen/wrpc/kong/model"
)

type NegotiationService interface {
	NegotiateServices(context.Context, *wrpc.Peer, *model.NegotiateServicesRequest) (*model.NegotiateServicesResponse, error)
}

func PrepareNegotiationServiceNegotiateServicesRequest(in *model.NegotiateServicesRequest) (wrpc.Request, error) {
	return wrpc.CreateRequest(5, 1, in)
}

type NegotiationServiceClient struct {
	Peer *wrpc.Peer
}

func (c *NegotiationServiceClient) NegotiateServices(ctx context.Context, in *model.NegotiateServicesRequest) (*model.NegotiateServicesResponse, error) {
	err := c.Peer.VerifyRPC(5, 1)
	if err != nil {
		return nil, err
	}

	req, err := PrepareNegotiationServiceNegotiateServicesRequest(in)
	if err != nil {
		return nil, err
	}

	var out model.NegotiateServicesResponse
	err = c.Peer.DoRequest(ctx, req, &out)
	if err != nil {
		return nil, err
	}

	return &out, nil
}

type NegotiationServiceServer struct {
	NegotiationService NegotiationService
}

func (s *NegotiationServiceServer) ID() wrpc.ID {
	return 5
}

func (s *NegotiationServiceServer) RPC(rpc wrpc.ID) wrpc.RPC {
	switch rpc {
	case 1:
		return wrpc.RPCImpl{
			HandlerFunc: func(ctx context.Context, peer *wrpc.Peer, decode func(interface{}) error) (interface{}, error) {
				var in model.NegotiateServicesRequest
				err := decode(&in)
				if err != nil {
					return nil, err
				}
				return s.NegotiationService.NegotiateServices(ctx, peer, &in)
			},
		}
	default:
		return nil
	}
}
