// Code generated by protoc-gen-go-wrpc. DO NOT EDIT
// protoc-gen-go-wrpc version: v0.0.0-20210914213024-d4348db6b815

package v1

import (
	context "context"
	wrpc "github.com/kong/go-wrpc/wrpc"
)

type LegalService interface {
	GetLicenseFromDP(context.Context, *GetLicenseFromDPRequest) (*GetLicenseFromDPResponse, error)
	SetLicense(context.Context, *SetLicenseRequest) (*SetLicenseResponse, error)
	ReportRequestCount(context.Context, *ReportRequestCountRequest) (*ReportRequestCountResponse, error)
}

type LegalServiceClient struct {
	Peer *wrpc.Peer
}

func (c *LegalServiceClient) GetLicenseFromDP(ctx context.Context, in *GetLicenseFromDPRequest) (*GetLicenseFromDPResponse, error) {
	var out GetLicenseFromDPResponse
	err := c.Peer.Do(ctx, 2, 1, in, &out)
	if err != nil {
		return nil, err
	}

	return &out, nil
}

func (c *LegalServiceClient) SetLicense(ctx context.Context, in *SetLicenseRequest) (*SetLicenseResponse, error) {
	var out SetLicenseResponse
	err := c.Peer.Do(ctx, 2, 2, in, &out)
	if err != nil {
		return nil, err
	}

	return &out, nil
}

func (c *LegalServiceClient) ReportRequestCount(ctx context.Context, in *ReportRequestCountRequest) (*ReportRequestCountResponse, error) {
	var out ReportRequestCountResponse
	err := c.Peer.Do(ctx, 2, 3, in, &out)
	if err != nil {
		return nil, err
	}

	return &out, nil
}

type LegalServiceServer struct {
	LegalService LegalService
}

func (s *LegalServiceServer) ID() wrpc.ID {
	return 2
}

func (s *LegalServiceServer) RPC(rpc wrpc.ID) wrpc.RPC {
	switch rpc {
	case 1:
		return wrpc.RPCImpl{
			HandlerFunc: func(ctx context.Context, decode func(interface{}) error) (interface{}, error) {
				var in GetLicenseFromDPRequest
				err := decode(&in)
				if err != nil {
					return nil, err
				}
				return s.LegalService.GetLicenseFromDP(ctx, &in)
			},
		}
	case 2:
		return wrpc.RPCImpl{
			HandlerFunc: func(ctx context.Context, decode func(interface{}) error) (interface{}, error) {
				var in SetLicenseRequest
				err := decode(&in)
				if err != nil {
					return nil, err
				}
				return s.LegalService.SetLicense(ctx, &in)
			},
		}
	case 3:
		return wrpc.RPCImpl{
			HandlerFunc: func(ctx context.Context, decode func(interface{}) error) (interface{}, error) {
				var in ReportRequestCountRequest
				err := decode(&in)
				if err != nil {
					return nil, err
				}
				return s.LegalService.ReportRequestCount(ctx, &in)
			},
		}
	default:
		return nil
	}
}
