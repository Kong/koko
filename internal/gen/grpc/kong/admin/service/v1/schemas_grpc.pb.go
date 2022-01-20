// Code generated by protoc-gen-go-grpc. DO NOT EDIT.

package v1

import (
	context "context"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
	structpb "google.golang.org/protobuf/types/known/structpb"
)

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
// Requires gRPC-Go v1.32.0 or later.
const _ = grpc.SupportPackageIsVersion7

// SchemasServiceClient is the client API for SchemasService service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type SchemasServiceClient interface {
	// buf:lint:ignore RPC_RESPONSE_STANDARD_NAME
	GetSchemas(ctx context.Context, in *GetSchemasRequest, opts ...grpc.CallOption) (*structpb.Struct, error)
}

type schemasServiceClient struct {
	cc grpc.ClientConnInterface
}

func NewSchemasServiceClient(cc grpc.ClientConnInterface) SchemasServiceClient {
	return &schemasServiceClient{cc}
}

func (c *schemasServiceClient) GetSchemas(ctx context.Context, in *GetSchemasRequest, opts ...grpc.CallOption) (*structpb.Struct, error) {
	out := new(structpb.Struct)
	err := c.cc.Invoke(ctx, "/kong.admin.service.v1.SchemasService/GetSchemas", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// SchemasServiceServer is the server API for SchemasService service.
// All implementations must embed UnimplementedSchemasServiceServer
// for forward compatibility
type SchemasServiceServer interface {
	// buf:lint:ignore RPC_RESPONSE_STANDARD_NAME
	GetSchemas(context.Context, *GetSchemasRequest) (*structpb.Struct, error)
	mustEmbedUnimplementedSchemasServiceServer()
}

// UnimplementedSchemasServiceServer must be embedded to have forward compatible implementations.
type UnimplementedSchemasServiceServer struct {
}

func (UnimplementedSchemasServiceServer) GetSchemas(context.Context, *GetSchemasRequest) (*structpb.Struct, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetSchemas not implemented")
}
func (UnimplementedSchemasServiceServer) mustEmbedUnimplementedSchemasServiceServer() {}

// UnsafeSchemasServiceServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to SchemasServiceServer will
// result in compilation errors.
type UnsafeSchemasServiceServer interface {
	mustEmbedUnimplementedSchemasServiceServer()
}

func RegisterSchemasServiceServer(s grpc.ServiceRegistrar, srv SchemasServiceServer) {
	s.RegisterService(&SchemasService_ServiceDesc, srv)
}

func _SchemasService_GetSchemas_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetSchemasRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(SchemasServiceServer).GetSchemas(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/kong.admin.service.v1.SchemasService/GetSchemas",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(SchemasServiceServer).GetSchemas(ctx, req.(*GetSchemasRequest))
	}
	return interceptor(ctx, in, info, handler)
}

// SchemasService_ServiceDesc is the grpc.ServiceDesc for SchemasService service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var SchemasService_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "kong.admin.service.v1.SchemasService",
	HandlerType: (*SchemasServiceServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "GetSchemas",
			Handler:    _SchemasService_GetSchemas_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "kong/admin/service/v1/schemas.proto",
}
