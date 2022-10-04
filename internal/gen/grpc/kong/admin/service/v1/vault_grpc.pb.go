// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.2.0
// - protoc             (unknown)
// source: kong/admin/service/v1/vault.proto

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

// VaultServiceClient is the client API for VaultService service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type VaultServiceClient interface {
	GetVault(ctx context.Context, in *GetVaultRequest, opts ...grpc.CallOption) (*GetVaultResponse, error)
	CreateVault(ctx context.Context, in *CreateVaultRequest, opts ...grpc.CallOption) (*CreateVaultResponse, error)
	UpsertVault(ctx context.Context, in *UpsertVaultRequest, opts ...grpc.CallOption) (*UpsertVaultResponse, error)
	DeleteVault(ctx context.Context, in *DeleteVaultRequest, opts ...grpc.CallOption) (*DeleteVaultResponse, error)
	ListVaults(ctx context.Context, in *ListVaultsRequest, opts ...grpc.CallOption) (*ListVaultsResponse, error)
}

type vaultServiceClient struct {
	cc grpc.ClientConnInterface
}

func NewVaultServiceClient(cc grpc.ClientConnInterface) VaultServiceClient {
	return &vaultServiceClient{cc}
}

func (c *vaultServiceClient) GetVault(ctx context.Context, in *GetVaultRequest, opts ...grpc.CallOption) (*GetVaultResponse, error) {
	out := new(GetVaultResponse)
	err := c.cc.Invoke(ctx, "/kong.admin.service.v1.VaultService/GetVault", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *vaultServiceClient) CreateVault(ctx context.Context, in *CreateVaultRequest, opts ...grpc.CallOption) (*CreateVaultResponse, error) {
	out := new(CreateVaultResponse)
	err := c.cc.Invoke(ctx, "/kong.admin.service.v1.VaultService/CreateVault", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *vaultServiceClient) UpsertVault(ctx context.Context, in *UpsertVaultRequest, opts ...grpc.CallOption) (*UpsertVaultResponse, error) {
	out := new(UpsertVaultResponse)
	err := c.cc.Invoke(ctx, "/kong.admin.service.v1.VaultService/UpsertVault", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *vaultServiceClient) DeleteVault(ctx context.Context, in *DeleteVaultRequest, opts ...grpc.CallOption) (*DeleteVaultResponse, error) {
	out := new(DeleteVaultResponse)
	err := c.cc.Invoke(ctx, "/kong.admin.service.v1.VaultService/DeleteVault", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *vaultServiceClient) ListVaults(ctx context.Context, in *ListVaultsRequest, opts ...grpc.CallOption) (*ListVaultsResponse, error) {
	out := new(ListVaultsResponse)
	err := c.cc.Invoke(ctx, "/kong.admin.service.v1.VaultService/ListVaults", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// VaultServiceServer is the server API for VaultService service.
// All implementations must embed UnimplementedVaultServiceServer
// for forward compatibility
type VaultServiceServer interface {
	GetVault(context.Context, *GetVaultRequest) (*GetVaultResponse, error)
	CreateVault(context.Context, *CreateVaultRequest) (*CreateVaultResponse, error)
	UpsertVault(context.Context, *UpsertVaultRequest) (*UpsertVaultResponse, error)
	DeleteVault(context.Context, *DeleteVaultRequest) (*DeleteVaultResponse, error)
	ListVaults(context.Context, *ListVaultsRequest) (*ListVaultsResponse, error)
	mustEmbedUnimplementedVaultServiceServer()
}

// UnimplementedVaultServiceServer must be embedded to have forward compatible implementations.
type UnimplementedVaultServiceServer struct {
}

func (UnimplementedVaultServiceServer) GetVault(context.Context, *GetVaultRequest) (*GetVaultResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetVault not implemented")
}
func (UnimplementedVaultServiceServer) CreateVault(context.Context, *CreateVaultRequest) (*CreateVaultResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method CreateVault not implemented")
}
func (UnimplementedVaultServiceServer) UpsertVault(context.Context, *UpsertVaultRequest) (*UpsertVaultResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method UpsertVault not implemented")
}
func (UnimplementedVaultServiceServer) DeleteVault(context.Context, *DeleteVaultRequest) (*DeleteVaultResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method DeleteVault not implemented")
}
func (UnimplementedVaultServiceServer) ListVaults(context.Context, *ListVaultsRequest) (*ListVaultsResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ListVaults not implemented")
}
func (UnimplementedVaultServiceServer) mustEmbedUnimplementedVaultServiceServer() {}

// UnsafeVaultServiceServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to VaultServiceServer will
// result in compilation errors.
type UnsafeVaultServiceServer interface {
	mustEmbedUnimplementedVaultServiceServer()
}

func RegisterVaultServiceServer(s grpc.ServiceRegistrar, srv VaultServiceServer) {
	s.RegisterService(&VaultService_ServiceDesc, srv)
}

func _VaultService_GetVault_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetVaultRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(VaultServiceServer).GetVault(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/kong.admin.service.v1.VaultService/GetVault",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(VaultServiceServer).GetVault(ctx, req.(*GetVaultRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _VaultService_CreateVault_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(CreateVaultRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(VaultServiceServer).CreateVault(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/kong.admin.service.v1.VaultService/CreateVault",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(VaultServiceServer).CreateVault(ctx, req.(*CreateVaultRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _VaultService_UpsertVault_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(UpsertVaultRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(VaultServiceServer).UpsertVault(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/kong.admin.service.v1.VaultService/UpsertVault",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(VaultServiceServer).UpsertVault(ctx, req.(*UpsertVaultRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _VaultService_DeleteVault_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(DeleteVaultRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(VaultServiceServer).DeleteVault(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/kong.admin.service.v1.VaultService/DeleteVault",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(VaultServiceServer).DeleteVault(ctx, req.(*DeleteVaultRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _VaultService_ListVaults_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ListVaultsRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(VaultServiceServer).ListVaults(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/kong.admin.service.v1.VaultService/ListVaults",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(VaultServiceServer).ListVaults(ctx, req.(*ListVaultsRequest))
	}
	return interceptor(ctx, in, info, handler)
}

// VaultService_ServiceDesc is the grpc.ServiceDesc for VaultService service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var VaultService_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "kong.admin.service.v1.VaultService",
	HandlerType: (*VaultServiceServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "GetVault",
			Handler:    _VaultService_GetVault_Handler,
		},
		{
			MethodName: "CreateVault",
			Handler:    _VaultService_CreateVault_Handler,
		},
		{
			MethodName: "UpsertVault",
			Handler:    _VaultService_UpsertVault_Handler,
		},
		{
			MethodName: "DeleteVault",
			Handler:    _VaultService_DeleteVault_Handler,
		},
		{
			MethodName: "ListVaults",
			Handler:    _VaultService_ListVaults_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "kong/admin/service/v1/vault.proto",
}
