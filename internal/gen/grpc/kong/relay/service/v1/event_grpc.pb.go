// Code generated by protoc-gen-go-grpc. DO NOT EDIT.

package v1

import (
	context "context"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
)

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
// Requires gRPC-Go v1.32.0 or later.
const _ = grpc.SupportPackageIsVersion7

// EventServiceClient is the client API for EventService service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type EventServiceClient interface {
	FetchReconfigureEvents(ctx context.Context, in *FetchReconfigureEventsRequest, opts ...grpc.CallOption) (EventService_FetchReconfigureEventsClient, error)
}

type eventServiceClient struct {
	cc grpc.ClientConnInterface
}

func NewEventServiceClient(cc grpc.ClientConnInterface) EventServiceClient {
	return &eventServiceClient{cc}
}

func (c *eventServiceClient) FetchReconfigureEvents(ctx context.Context, in *FetchReconfigureEventsRequest, opts ...grpc.CallOption) (EventService_FetchReconfigureEventsClient, error) {
	stream, err := c.cc.NewStream(ctx, &EventService_ServiceDesc.Streams[0], "/kong.relay.service.v1.EventService/FetchReconfigureEvents", opts...)
	if err != nil {
		return nil, err
	}
	x := &eventServiceFetchReconfigureEventsClient{stream}
	if err := x.ClientStream.SendMsg(in); err != nil {
		return nil, err
	}
	if err := x.ClientStream.CloseSend(); err != nil {
		return nil, err
	}
	return x, nil
}

type EventService_FetchReconfigureEventsClient interface {
	Recv() (*FetchReconfigureEventsResponse, error)
	grpc.ClientStream
}

type eventServiceFetchReconfigureEventsClient struct {
	grpc.ClientStream
}

func (x *eventServiceFetchReconfigureEventsClient) Recv() (*FetchReconfigureEventsResponse, error) {
	m := new(FetchReconfigureEventsResponse)
	if err := x.ClientStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

// EventServiceServer is the server API for EventService service.
// All implementations must embed UnimplementedEventServiceServer
// for forward compatibility
type EventServiceServer interface {
	FetchReconfigureEvents(*FetchReconfigureEventsRequest, EventService_FetchReconfigureEventsServer) error
	mustEmbedUnimplementedEventServiceServer()
}

// UnimplementedEventServiceServer must be embedded to have forward compatible implementations.
type UnimplementedEventServiceServer struct {
}

func (UnimplementedEventServiceServer) FetchReconfigureEvents(*FetchReconfigureEventsRequest, EventService_FetchReconfigureEventsServer) error {
	return status.Errorf(codes.Unimplemented, "method FetchReconfigureEvents not implemented")
}
func (UnimplementedEventServiceServer) mustEmbedUnimplementedEventServiceServer() {}

// UnsafeEventServiceServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to EventServiceServer will
// result in compilation errors.
type UnsafeEventServiceServer interface {
	mustEmbedUnimplementedEventServiceServer()
}

func RegisterEventServiceServer(s grpc.ServiceRegistrar, srv EventServiceServer) {
	s.RegisterService(&EventService_ServiceDesc, srv)
}

func _EventService_FetchReconfigureEvents_Handler(srv interface{}, stream grpc.ServerStream) error {
	m := new(FetchReconfigureEventsRequest)
	if err := stream.RecvMsg(m); err != nil {
		return err
	}
	return srv.(EventServiceServer).FetchReconfigureEvents(m, &eventServiceFetchReconfigureEventsServer{stream})
}

type EventService_FetchReconfigureEventsServer interface {
	Send(*FetchReconfigureEventsResponse) error
	grpc.ServerStream
}

type eventServiceFetchReconfigureEventsServer struct {
	grpc.ServerStream
}

func (x *eventServiceFetchReconfigureEventsServer) Send(m *FetchReconfigureEventsResponse) error {
	return x.ServerStream.SendMsg(m)
}

// EventService_ServiceDesc is the grpc.ServiceDesc for EventService service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var EventService_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "kong.relay.service.v1.EventService",
	HandlerType: (*EventServiceServer)(nil),
	Methods:     []grpc.MethodDesc{},
	Streams: []grpc.StreamDesc{
		{
			StreamName:    "FetchReconfigureEvents",
			Handler:       _EventService_FetchReconfigureEvents_Handler,
			ServerStreams: true,
		},
	},
	Metadata: "kong/relay/service/v1/event.proto",
}
