// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.2.0
// - protoc             (unknown)
// source: kong/admin/service/v1/schemas.proto

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

// SchemasServiceClient is the client API for SchemasService service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type SchemasServiceClient interface {
	ValidateLuaPlugin(ctx context.Context, in *ValidateLuaPluginRequest, opts ...grpc.CallOption) (*ValidateLuaPluginResponse, error)
	ValidateCACertificateSchema(ctx context.Context, in *ValidateCACertificateSchemaRequest, opts ...grpc.CallOption) (*ValidateCACertificateSchemaResponse, error)
	ValidateCertificateSchema(ctx context.Context, in *ValidateCertificateSchemaRequest, opts ...grpc.CallOption) (*ValidateCertificateSchemaResponse, error)
	ValidateConsumerSchema(ctx context.Context, in *ValidateConsumerSchemaRequest, opts ...grpc.CallOption) (*ValidateConsumerSchemaResponse, error)
	ValidateConsumerGroupSchema(ctx context.Context, in *ValidateConsumerGroupSchemaRequest, opts ...grpc.CallOption) (*ValidateConsumerGroupSchemaResponse, error)
	ValidateConsumerGroupRateLimitingAdvancedConfigSchema(ctx context.Context, in *ValidateConsumerGroupRateLimitingAdvancedConfigSchemaRequest, opts ...grpc.CallOption) (*ValidateConsumerGroupRateLimitingAdvancedConfigSchemaResponse, error)
	ValidatePluginSchema(ctx context.Context, in *ValidatePluginSchemaRequest, opts ...grpc.CallOption) (*ValidatePluginSchemaResponse, error)
	ValidateRouteSchema(ctx context.Context, in *ValidateRouteSchemaRequest, opts ...grpc.CallOption) (*ValidateRouteSchemaResponse, error)
	ValidateServiceSchema(ctx context.Context, in *ValidateServiceSchemaRequest, opts ...grpc.CallOption) (*ValidateServiceSchemaResponse, error)
	ValidateSNISchema(ctx context.Context, in *ValidateSNISchemaRequest, opts ...grpc.CallOption) (*ValidateSNISchemaResponse, error)
	ValidateTargetSchema(ctx context.Context, in *ValidateTargetSchemaRequest, opts ...grpc.CallOption) (*ValidateTargetSchemaResponse, error)
	ValidateUpstreamSchema(ctx context.Context, in *ValidateUpstreamSchemaRequest, opts ...grpc.CallOption) (*ValidateUpstreamSchemaResponse, error)
	ValidateVaultSchema(ctx context.Context, in *ValidateVaultSchemaRequest, opts ...grpc.CallOption) (*ValidateVaultSchemaResponse, error)
	GetSchemas(ctx context.Context, in *GetSchemasRequest, opts ...grpc.CallOption) (*GetSchemasResponse, error)
	GetLuaSchemasPlugin(ctx context.Context, in *GetLuaSchemasPluginRequest, opts ...grpc.CallOption) (*GetLuaSchemasPluginResponse, error)
}

type schemasServiceClient struct {
	cc grpc.ClientConnInterface
}

func NewSchemasServiceClient(cc grpc.ClientConnInterface) SchemasServiceClient {
	return &schemasServiceClient{cc}
}

func (c *schemasServiceClient) ValidateLuaPlugin(ctx context.Context, in *ValidateLuaPluginRequest, opts ...grpc.CallOption) (*ValidateLuaPluginResponse, error) {
	out := new(ValidateLuaPluginResponse)
	err := c.cc.Invoke(ctx, "/kong.admin.service.v1.SchemasService/ValidateLuaPlugin", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *schemasServiceClient) ValidateCACertificateSchema(ctx context.Context, in *ValidateCACertificateSchemaRequest, opts ...grpc.CallOption) (*ValidateCACertificateSchemaResponse, error) {
	out := new(ValidateCACertificateSchemaResponse)
	err := c.cc.Invoke(ctx, "/kong.admin.service.v1.SchemasService/ValidateCACertificateSchema", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *schemasServiceClient) ValidateCertificateSchema(ctx context.Context, in *ValidateCertificateSchemaRequest, opts ...grpc.CallOption) (*ValidateCertificateSchemaResponse, error) {
	out := new(ValidateCertificateSchemaResponse)
	err := c.cc.Invoke(ctx, "/kong.admin.service.v1.SchemasService/ValidateCertificateSchema", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *schemasServiceClient) ValidateConsumerSchema(ctx context.Context, in *ValidateConsumerSchemaRequest, opts ...grpc.CallOption) (*ValidateConsumerSchemaResponse, error) {
	out := new(ValidateConsumerSchemaResponse)
	err := c.cc.Invoke(ctx, "/kong.admin.service.v1.SchemasService/ValidateConsumerSchema", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *schemasServiceClient) ValidateConsumerGroupSchema(ctx context.Context, in *ValidateConsumerGroupSchemaRequest, opts ...grpc.CallOption) (*ValidateConsumerGroupSchemaResponse, error) {
	out := new(ValidateConsumerGroupSchemaResponse)
	err := c.cc.Invoke(ctx, "/kong.admin.service.v1.SchemasService/ValidateConsumerGroupSchema", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *schemasServiceClient) ValidateConsumerGroupRateLimitingAdvancedConfigSchema(ctx context.Context, in *ValidateConsumerGroupRateLimitingAdvancedConfigSchemaRequest, opts ...grpc.CallOption) (*ValidateConsumerGroupRateLimitingAdvancedConfigSchemaResponse, error) {
	out := new(ValidateConsumerGroupRateLimitingAdvancedConfigSchemaResponse)
	err := c.cc.Invoke(ctx, "/kong.admin.service.v1.SchemasService/ValidateConsumerGroupRateLimitingAdvancedConfigSchema", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *schemasServiceClient) ValidatePluginSchema(ctx context.Context, in *ValidatePluginSchemaRequest, opts ...grpc.CallOption) (*ValidatePluginSchemaResponse, error) {
	out := new(ValidatePluginSchemaResponse)
	err := c.cc.Invoke(ctx, "/kong.admin.service.v1.SchemasService/ValidatePluginSchema", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *schemasServiceClient) ValidateRouteSchema(ctx context.Context, in *ValidateRouteSchemaRequest, opts ...grpc.CallOption) (*ValidateRouteSchemaResponse, error) {
	out := new(ValidateRouteSchemaResponse)
	err := c.cc.Invoke(ctx, "/kong.admin.service.v1.SchemasService/ValidateRouteSchema", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *schemasServiceClient) ValidateServiceSchema(ctx context.Context, in *ValidateServiceSchemaRequest, opts ...grpc.CallOption) (*ValidateServiceSchemaResponse, error) {
	out := new(ValidateServiceSchemaResponse)
	err := c.cc.Invoke(ctx, "/kong.admin.service.v1.SchemasService/ValidateServiceSchema", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *schemasServiceClient) ValidateSNISchema(ctx context.Context, in *ValidateSNISchemaRequest, opts ...grpc.CallOption) (*ValidateSNISchemaResponse, error) {
	out := new(ValidateSNISchemaResponse)
	err := c.cc.Invoke(ctx, "/kong.admin.service.v1.SchemasService/ValidateSNISchema", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *schemasServiceClient) ValidateTargetSchema(ctx context.Context, in *ValidateTargetSchemaRequest, opts ...grpc.CallOption) (*ValidateTargetSchemaResponse, error) {
	out := new(ValidateTargetSchemaResponse)
	err := c.cc.Invoke(ctx, "/kong.admin.service.v1.SchemasService/ValidateTargetSchema", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *schemasServiceClient) ValidateUpstreamSchema(ctx context.Context, in *ValidateUpstreamSchemaRequest, opts ...grpc.CallOption) (*ValidateUpstreamSchemaResponse, error) {
	out := new(ValidateUpstreamSchemaResponse)
	err := c.cc.Invoke(ctx, "/kong.admin.service.v1.SchemasService/ValidateUpstreamSchema", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *schemasServiceClient) ValidateVaultSchema(ctx context.Context, in *ValidateVaultSchemaRequest, opts ...grpc.CallOption) (*ValidateVaultSchemaResponse, error) {
	out := new(ValidateVaultSchemaResponse)
	err := c.cc.Invoke(ctx, "/kong.admin.service.v1.SchemasService/ValidateVaultSchema", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *schemasServiceClient) GetSchemas(ctx context.Context, in *GetSchemasRequest, opts ...grpc.CallOption) (*GetSchemasResponse, error) {
	out := new(GetSchemasResponse)
	err := c.cc.Invoke(ctx, "/kong.admin.service.v1.SchemasService/GetSchemas", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *schemasServiceClient) GetLuaSchemasPlugin(ctx context.Context, in *GetLuaSchemasPluginRequest, opts ...grpc.CallOption) (*GetLuaSchemasPluginResponse, error) {
	out := new(GetLuaSchemasPluginResponse)
	err := c.cc.Invoke(ctx, "/kong.admin.service.v1.SchemasService/GetLuaSchemasPlugin", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// SchemasServiceServer is the server API for SchemasService service.
// All implementations must embed UnimplementedSchemasServiceServer
// for forward compatibility
type SchemasServiceServer interface {
	ValidateLuaPlugin(context.Context, *ValidateLuaPluginRequest) (*ValidateLuaPluginResponse, error)
	ValidateCACertificateSchema(context.Context, *ValidateCACertificateSchemaRequest) (*ValidateCACertificateSchemaResponse, error)
	ValidateCertificateSchema(context.Context, *ValidateCertificateSchemaRequest) (*ValidateCertificateSchemaResponse, error)
	ValidateConsumerSchema(context.Context, *ValidateConsumerSchemaRequest) (*ValidateConsumerSchemaResponse, error)
	ValidateConsumerGroupSchema(context.Context, *ValidateConsumerGroupSchemaRequest) (*ValidateConsumerGroupSchemaResponse, error)
	ValidateConsumerGroupRateLimitingAdvancedConfigSchema(context.Context, *ValidateConsumerGroupRateLimitingAdvancedConfigSchemaRequest) (*ValidateConsumerGroupRateLimitingAdvancedConfigSchemaResponse, error)
	ValidatePluginSchema(context.Context, *ValidatePluginSchemaRequest) (*ValidatePluginSchemaResponse, error)
	ValidateRouteSchema(context.Context, *ValidateRouteSchemaRequest) (*ValidateRouteSchemaResponse, error)
	ValidateServiceSchema(context.Context, *ValidateServiceSchemaRequest) (*ValidateServiceSchemaResponse, error)
	ValidateSNISchema(context.Context, *ValidateSNISchemaRequest) (*ValidateSNISchemaResponse, error)
	ValidateTargetSchema(context.Context, *ValidateTargetSchemaRequest) (*ValidateTargetSchemaResponse, error)
	ValidateUpstreamSchema(context.Context, *ValidateUpstreamSchemaRequest) (*ValidateUpstreamSchemaResponse, error)
	ValidateVaultSchema(context.Context, *ValidateVaultSchemaRequest) (*ValidateVaultSchemaResponse, error)
	GetSchemas(context.Context, *GetSchemasRequest) (*GetSchemasResponse, error)
	GetLuaSchemasPlugin(context.Context, *GetLuaSchemasPluginRequest) (*GetLuaSchemasPluginResponse, error)
	mustEmbedUnimplementedSchemasServiceServer()
}

// UnimplementedSchemasServiceServer must be embedded to have forward compatible implementations.
type UnimplementedSchemasServiceServer struct {
}

func (UnimplementedSchemasServiceServer) ValidateLuaPlugin(context.Context, *ValidateLuaPluginRequest) (*ValidateLuaPluginResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ValidateLuaPlugin not implemented")
}
func (UnimplementedSchemasServiceServer) ValidateCACertificateSchema(context.Context, *ValidateCACertificateSchemaRequest) (*ValidateCACertificateSchemaResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ValidateCACertificateSchema not implemented")
}
func (UnimplementedSchemasServiceServer) ValidateCertificateSchema(context.Context, *ValidateCertificateSchemaRequest) (*ValidateCertificateSchemaResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ValidateCertificateSchema not implemented")
}
func (UnimplementedSchemasServiceServer) ValidateConsumerSchema(context.Context, *ValidateConsumerSchemaRequest) (*ValidateConsumerSchemaResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ValidateConsumerSchema not implemented")
}
func (UnimplementedSchemasServiceServer) ValidateConsumerGroupSchema(context.Context, *ValidateConsumerGroupSchemaRequest) (*ValidateConsumerGroupSchemaResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ValidateConsumerGroupSchema not implemented")
}
func (UnimplementedSchemasServiceServer) ValidateConsumerGroupRateLimitingAdvancedConfigSchema(context.Context, *ValidateConsumerGroupRateLimitingAdvancedConfigSchemaRequest) (*ValidateConsumerGroupRateLimitingAdvancedConfigSchemaResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ValidateConsumerGroupRateLimitingAdvancedConfigSchema not implemented")
}
func (UnimplementedSchemasServiceServer) ValidatePluginSchema(context.Context, *ValidatePluginSchemaRequest) (*ValidatePluginSchemaResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ValidatePluginSchema not implemented")
}
func (UnimplementedSchemasServiceServer) ValidateRouteSchema(context.Context, *ValidateRouteSchemaRequest) (*ValidateRouteSchemaResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ValidateRouteSchema not implemented")
}
func (UnimplementedSchemasServiceServer) ValidateServiceSchema(context.Context, *ValidateServiceSchemaRequest) (*ValidateServiceSchemaResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ValidateServiceSchema not implemented")
}
func (UnimplementedSchemasServiceServer) ValidateSNISchema(context.Context, *ValidateSNISchemaRequest) (*ValidateSNISchemaResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ValidateSNISchema not implemented")
}
func (UnimplementedSchemasServiceServer) ValidateTargetSchema(context.Context, *ValidateTargetSchemaRequest) (*ValidateTargetSchemaResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ValidateTargetSchema not implemented")
}
func (UnimplementedSchemasServiceServer) ValidateUpstreamSchema(context.Context, *ValidateUpstreamSchemaRequest) (*ValidateUpstreamSchemaResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ValidateUpstreamSchema not implemented")
}
func (UnimplementedSchemasServiceServer) ValidateVaultSchema(context.Context, *ValidateVaultSchemaRequest) (*ValidateVaultSchemaResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ValidateVaultSchema not implemented")
}
func (UnimplementedSchemasServiceServer) GetSchemas(context.Context, *GetSchemasRequest) (*GetSchemasResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetSchemas not implemented")
}
func (UnimplementedSchemasServiceServer) GetLuaSchemasPlugin(context.Context, *GetLuaSchemasPluginRequest) (*GetLuaSchemasPluginResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetLuaSchemasPlugin not implemented")
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

func _SchemasService_ValidateLuaPlugin_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ValidateLuaPluginRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(SchemasServiceServer).ValidateLuaPlugin(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/kong.admin.service.v1.SchemasService/ValidateLuaPlugin",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(SchemasServiceServer).ValidateLuaPlugin(ctx, req.(*ValidateLuaPluginRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _SchemasService_ValidateCACertificateSchema_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ValidateCACertificateSchemaRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(SchemasServiceServer).ValidateCACertificateSchema(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/kong.admin.service.v1.SchemasService/ValidateCACertificateSchema",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(SchemasServiceServer).ValidateCACertificateSchema(ctx, req.(*ValidateCACertificateSchemaRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _SchemasService_ValidateCertificateSchema_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ValidateCertificateSchemaRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(SchemasServiceServer).ValidateCertificateSchema(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/kong.admin.service.v1.SchemasService/ValidateCertificateSchema",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(SchemasServiceServer).ValidateCertificateSchema(ctx, req.(*ValidateCertificateSchemaRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _SchemasService_ValidateConsumerSchema_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ValidateConsumerSchemaRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(SchemasServiceServer).ValidateConsumerSchema(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/kong.admin.service.v1.SchemasService/ValidateConsumerSchema",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(SchemasServiceServer).ValidateConsumerSchema(ctx, req.(*ValidateConsumerSchemaRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _SchemasService_ValidateConsumerGroupSchema_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ValidateConsumerGroupSchemaRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(SchemasServiceServer).ValidateConsumerGroupSchema(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/kong.admin.service.v1.SchemasService/ValidateConsumerGroupSchema",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(SchemasServiceServer).ValidateConsumerGroupSchema(ctx, req.(*ValidateConsumerGroupSchemaRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _SchemasService_ValidateConsumerGroupRateLimitingAdvancedConfigSchema_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ValidateConsumerGroupRateLimitingAdvancedConfigSchemaRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(SchemasServiceServer).ValidateConsumerGroupRateLimitingAdvancedConfigSchema(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/kong.admin.service.v1.SchemasService/ValidateConsumerGroupRateLimitingAdvancedConfigSchema",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(SchemasServiceServer).ValidateConsumerGroupRateLimitingAdvancedConfigSchema(ctx, req.(*ValidateConsumerGroupRateLimitingAdvancedConfigSchemaRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _SchemasService_ValidatePluginSchema_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ValidatePluginSchemaRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(SchemasServiceServer).ValidatePluginSchema(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/kong.admin.service.v1.SchemasService/ValidatePluginSchema",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(SchemasServiceServer).ValidatePluginSchema(ctx, req.(*ValidatePluginSchemaRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _SchemasService_ValidateRouteSchema_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ValidateRouteSchemaRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(SchemasServiceServer).ValidateRouteSchema(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/kong.admin.service.v1.SchemasService/ValidateRouteSchema",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(SchemasServiceServer).ValidateRouteSchema(ctx, req.(*ValidateRouteSchemaRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _SchemasService_ValidateServiceSchema_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ValidateServiceSchemaRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(SchemasServiceServer).ValidateServiceSchema(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/kong.admin.service.v1.SchemasService/ValidateServiceSchema",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(SchemasServiceServer).ValidateServiceSchema(ctx, req.(*ValidateServiceSchemaRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _SchemasService_ValidateSNISchema_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ValidateSNISchemaRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(SchemasServiceServer).ValidateSNISchema(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/kong.admin.service.v1.SchemasService/ValidateSNISchema",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(SchemasServiceServer).ValidateSNISchema(ctx, req.(*ValidateSNISchemaRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _SchemasService_ValidateTargetSchema_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ValidateTargetSchemaRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(SchemasServiceServer).ValidateTargetSchema(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/kong.admin.service.v1.SchemasService/ValidateTargetSchema",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(SchemasServiceServer).ValidateTargetSchema(ctx, req.(*ValidateTargetSchemaRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _SchemasService_ValidateUpstreamSchema_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ValidateUpstreamSchemaRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(SchemasServiceServer).ValidateUpstreamSchema(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/kong.admin.service.v1.SchemasService/ValidateUpstreamSchema",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(SchemasServiceServer).ValidateUpstreamSchema(ctx, req.(*ValidateUpstreamSchemaRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _SchemasService_ValidateVaultSchema_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ValidateVaultSchemaRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(SchemasServiceServer).ValidateVaultSchema(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/kong.admin.service.v1.SchemasService/ValidateVaultSchema",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(SchemasServiceServer).ValidateVaultSchema(ctx, req.(*ValidateVaultSchemaRequest))
	}
	return interceptor(ctx, in, info, handler)
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

func _SchemasService_GetLuaSchemasPlugin_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetLuaSchemasPluginRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(SchemasServiceServer).GetLuaSchemasPlugin(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/kong.admin.service.v1.SchemasService/GetLuaSchemasPlugin",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(SchemasServiceServer).GetLuaSchemasPlugin(ctx, req.(*GetLuaSchemasPluginRequest))
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
			MethodName: "ValidateLuaPlugin",
			Handler:    _SchemasService_ValidateLuaPlugin_Handler,
		},
		{
			MethodName: "ValidateCACertificateSchema",
			Handler:    _SchemasService_ValidateCACertificateSchema_Handler,
		},
		{
			MethodName: "ValidateCertificateSchema",
			Handler:    _SchemasService_ValidateCertificateSchema_Handler,
		},
		{
			MethodName: "ValidateConsumerSchema",
			Handler:    _SchemasService_ValidateConsumerSchema_Handler,
		},
		{
			MethodName: "ValidateConsumerGroupSchema",
			Handler:    _SchemasService_ValidateConsumerGroupSchema_Handler,
		},
		{
			MethodName: "ValidateConsumerGroupRateLimitingAdvancedConfigSchema",
			Handler:    _SchemasService_ValidateConsumerGroupRateLimitingAdvancedConfigSchema_Handler,
		},
		{
			MethodName: "ValidatePluginSchema",
			Handler:    _SchemasService_ValidatePluginSchema_Handler,
		},
		{
			MethodName: "ValidateRouteSchema",
			Handler:    _SchemasService_ValidateRouteSchema_Handler,
		},
		{
			MethodName: "ValidateServiceSchema",
			Handler:    _SchemasService_ValidateServiceSchema_Handler,
		},
		{
			MethodName: "ValidateSNISchema",
			Handler:    _SchemasService_ValidateSNISchema_Handler,
		},
		{
			MethodName: "ValidateTargetSchema",
			Handler:    _SchemasService_ValidateTargetSchema_Handler,
		},
		{
			MethodName: "ValidateUpstreamSchema",
			Handler:    _SchemasService_ValidateUpstreamSchema_Handler,
		},
		{
			MethodName: "ValidateVaultSchema",
			Handler:    _SchemasService_ValidateVaultSchema_Handler,
		},
		{
			MethodName: "GetSchemas",
			Handler:    _SchemasService_GetSchemas_Handler,
		},
		{
			MethodName: "GetLuaSchemasPlugin",
			Handler:    _SchemasService_GetLuaSchemasPlugin_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "kong/admin/service/v1/schemas.proto",
}
