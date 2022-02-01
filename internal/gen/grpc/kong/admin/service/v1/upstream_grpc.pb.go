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

// UpstreamServiceClient is the client API for UpstreamService service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type UpstreamServiceClient interface {
	GetUpstream(ctx context.Context, in *GetUpstreamRequest, opts ...grpc.CallOption) (*GetUpstreamResponse, error)
	CreateUpstream(ctx context.Context, in *CreateUpstreamRequest, opts ...grpc.CallOption) (*CreateUpstreamResponse, error)
	UpsertUpstream(ctx context.Context, in *UpsertUpstreamRequest, opts ...grpc.CallOption) (*UpsertUpstreamResponse, error)
	DeleteUpstream(ctx context.Context, in *DeleteUpstreamRequest, opts ...grpc.CallOption) (*DeleteUpstreamResponse, error)
	ListUpstreams(ctx context.Context, in *ListUpstreamsRequest, opts ...grpc.CallOption) (*ListUpstreamsResponse, error)
}

type upstreamServiceClient struct {
	cc grpc.ClientConnInterface
}

func NewUpstreamServiceClient(cc grpc.ClientConnInterface) UpstreamServiceClient {
	return &upstreamServiceClient{cc}
}

func (c *upstreamServiceClient) GetUpstream(ctx context.Context, in *GetUpstreamRequest, opts ...grpc.CallOption) (*GetUpstreamResponse, error) {
	out := new(GetUpstreamResponse)
	err := c.cc.Invoke(ctx, "/kong.admin.service.v1.UpstreamService/GetUpstream", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *upstreamServiceClient) CreateUpstream(ctx context.Context, in *CreateUpstreamRequest, opts ...grpc.CallOption) (*CreateUpstreamResponse, error) {
	out := new(CreateUpstreamResponse)
	err := c.cc.Invoke(ctx, "/kong.admin.service.v1.UpstreamService/CreateUpstream", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *upstreamServiceClient) UpsertUpstream(ctx context.Context, in *UpsertUpstreamRequest, opts ...grpc.CallOption) (*UpsertUpstreamResponse, error) {
	out := new(UpsertUpstreamResponse)
	err := c.cc.Invoke(ctx, "/kong.admin.service.v1.UpstreamService/UpsertUpstream", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *upstreamServiceClient) DeleteUpstream(ctx context.Context, in *DeleteUpstreamRequest, opts ...grpc.CallOption) (*DeleteUpstreamResponse, error) {
	out := new(DeleteUpstreamResponse)
	err := c.cc.Invoke(ctx, "/kong.admin.service.v1.UpstreamService/DeleteUpstream", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *upstreamServiceClient) ListUpstreams(ctx context.Context, in *ListUpstreamsRequest, opts ...grpc.CallOption) (*ListUpstreamsResponse, error) {
	out := new(ListUpstreamsResponse)
	err := c.cc.Invoke(ctx, "/kong.admin.service.v1.UpstreamService/ListUpstreams", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// UpstreamServiceServer is the server API for UpstreamService service.
// All implementations must embed UnimplementedUpstreamServiceServer
// for forward compatibility
type UpstreamServiceServer interface {
	GetUpstream(context.Context, *GetUpstreamRequest) (*GetUpstreamResponse, error)
	CreateUpstream(context.Context, *CreateUpstreamRequest) (*CreateUpstreamResponse, error)
	UpsertUpstream(context.Context, *UpsertUpstreamRequest) (*UpsertUpstreamResponse, error)
	DeleteUpstream(context.Context, *DeleteUpstreamRequest) (*DeleteUpstreamResponse, error)
	ListUpstreams(context.Context, *ListUpstreamsRequest) (*ListUpstreamsResponse, error)
	mustEmbedUnimplementedUpstreamServiceServer()
}

// UnimplementedUpstreamServiceServer must be embedded to have forward compatible implementations.
type UnimplementedUpstreamServiceServer struct {
}

func (UnimplementedUpstreamServiceServer) GetUpstream(context.Context, *GetUpstreamRequest) (*GetUpstreamResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetUpstream not implemented")
}
func (UnimplementedUpstreamServiceServer) CreateUpstream(context.Context, *CreateUpstreamRequest) (*CreateUpstreamResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method CreateUpstream not implemented")
}
func (UnimplementedUpstreamServiceServer) UpsertUpstream(context.Context, *UpsertUpstreamRequest) (*UpsertUpstreamResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method UpsertUpstream not implemented")
}
func (UnimplementedUpstreamServiceServer) DeleteUpstream(context.Context, *DeleteUpstreamRequest) (*DeleteUpstreamResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method DeleteUpstream not implemented")
}
func (UnimplementedUpstreamServiceServer) ListUpstreams(context.Context, *ListUpstreamsRequest) (*ListUpstreamsResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ListUpstreams not implemented")
}
func (UnimplementedUpstreamServiceServer) mustEmbedUnimplementedUpstreamServiceServer() {}

// UnsafeUpstreamServiceServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to UpstreamServiceServer will
// result in compilation errors.
type UnsafeUpstreamServiceServer interface {
	mustEmbedUnimplementedUpstreamServiceServer()
}

func RegisterUpstreamServiceServer(s grpc.ServiceRegistrar, srv UpstreamServiceServer) {
	s.RegisterService(&UpstreamService_ServiceDesc, srv)
}

func _UpstreamService_GetUpstream_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetUpstreamRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(UpstreamServiceServer).GetUpstream(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/kong.admin.service.v1.UpstreamService/GetUpstream",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(UpstreamServiceServer).GetUpstream(ctx, req.(*GetUpstreamRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _UpstreamService_CreateUpstream_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(CreateUpstreamRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(UpstreamServiceServer).CreateUpstream(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/kong.admin.service.v1.UpstreamService/CreateUpstream",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(UpstreamServiceServer).CreateUpstream(ctx, req.(*CreateUpstreamRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _UpstreamService_UpsertUpstream_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(UpsertUpstreamRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(UpstreamServiceServer).UpsertUpstream(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/kong.admin.service.v1.UpstreamService/UpsertUpstream",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(UpstreamServiceServer).UpsertUpstream(ctx, req.(*UpsertUpstreamRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _UpstreamService_DeleteUpstream_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(DeleteUpstreamRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(UpstreamServiceServer).DeleteUpstream(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/kong.admin.service.v1.UpstreamService/DeleteUpstream",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(UpstreamServiceServer).DeleteUpstream(ctx, req.(*DeleteUpstreamRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _UpstreamService_ListUpstreams_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ListUpstreamsRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(UpstreamServiceServer).ListUpstreams(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/kong.admin.service.v1.UpstreamService/ListUpstreams",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(UpstreamServiceServer).ListUpstreams(ctx, req.(*ListUpstreamsRequest))
	}
	return interceptor(ctx, in, info, handler)
}

// UpstreamService_ServiceDesc is the grpc.ServiceDesc for UpstreamService service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var UpstreamService_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "kong.admin.service.v1.UpstreamService",
	HandlerType: (*UpstreamServiceServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "GetUpstream",
			Handler:    _UpstreamService_GetUpstream_Handler,
		},
		{
			MethodName: "CreateUpstream",
			Handler:    _UpstreamService_CreateUpstream_Handler,
		},
		{
			MethodName: "UpsertUpstream",
			Handler:    _UpstreamService_UpsertUpstream_Handler,
		},
		{
			MethodName: "DeleteUpstream",
			Handler:    _UpstreamService_DeleteUpstream_Handler,
		},
		{
			MethodName: "ListUpstreams",
			Handler:    _UpstreamService_ListUpstreams_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "kong/admin/service/v1/upstream.proto",
}