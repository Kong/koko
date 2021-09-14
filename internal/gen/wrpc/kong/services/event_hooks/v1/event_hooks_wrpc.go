// Code generated by protoc-gen-go-wrpc. DO NOT EDIT
// protoc-gen-go-wrpc version: v0.0.0-20210914213024-d4348db6b815

package v1

import (
	context "context"
	wrpc "github.com/kong/go-wrpc/wrpc"
)

type EventHooksService interface {
	SyncHooks(context.Context, *SyncHooksRequest) (*SyncHooksResponse, error)
}

type EventHooksServiceClient struct {
	Peer *wrpc.Peer
}

func (c *EventHooksServiceClient) SyncHooks(ctx context.Context, in *SyncHooksRequest) (*SyncHooksResponse, error) {
	var out SyncHooksResponse
	err := c.Peer.Do(ctx, 3, 1, in, &out)
	if err != nil {
		return nil, err
	}

	return &out, nil
}

type EventHooksServiceServer struct {
	EventHooksService EventHooksService
}

func (s *EventHooksServiceServer) ID() wrpc.ID {
	return 3
}

func (s *EventHooksServiceServer) RPC(rpc wrpc.ID) wrpc.RPC {
	switch rpc {
	case 1:
		return wrpc.RPCImpl{
			HandlerFunc: func(ctx context.Context, decode func(interface{}) error) (interface{}, error) {
				var in SyncHooksRequest
				err := decode(&in)
				if err != nil {
					return nil, err
				}
				return s.EventHooksService.SyncHooks(ctx, &in)
			},
		}
	default:
		return nil
	}
}
