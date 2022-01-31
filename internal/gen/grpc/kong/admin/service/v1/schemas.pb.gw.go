// Code generated by protoc-gen-grpc-gateway. DO NOT EDIT.
// source: kong/admin/service/v1/schemas.proto

/*
Package v1 is a reverse proxy.

It translates gRPC into RESTful JSON APIs.
*/
package v1

import (
	"context"
	"io"
	"net/http"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/grpc-ecosystem/grpc-gateway/v2/utilities"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/grpclog"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"
)

// Suppress "imported and not used" errors
var _ codes.Code
var _ io.Reader
var _ status.Status
var _ = runtime.String
var _ = utilities.NewDoubleArray
var _ = metadata.Join

func request_SchemasService_GetSchemas_0(ctx context.Context, marshaler runtime.Marshaler, client SchemasServiceClient, req *http.Request, pathParams map[string]string) (proto.Message, runtime.ServerMetadata, error) {
	var protoReq GetSchemasRequest
	var metadata runtime.ServerMetadata

	var (
		val string
		ok  bool
		err error
		_   = err
	)

	val, ok = pathParams["name"]
	if !ok {
		return nil, metadata, status.Errorf(codes.InvalidArgument, "missing parameter %s", "name")
	}

	protoReq.Name, err = runtime.String(val)
	if err != nil {
		return nil, metadata, status.Errorf(codes.InvalidArgument, "type mismatch, parameter: %s, error: %v", "name", err)
	}

	msg, err := client.GetSchemas(ctx, &protoReq, grpc.Header(&metadata.HeaderMD), grpc.Trailer(&metadata.TrailerMD))
	return msg, metadata, err

}

func local_request_SchemasService_GetSchemas_0(ctx context.Context, marshaler runtime.Marshaler, server SchemasServiceServer, req *http.Request, pathParams map[string]string) (proto.Message, runtime.ServerMetadata, error) {
	var protoReq GetSchemasRequest
	var metadata runtime.ServerMetadata

	var (
		val string
		ok  bool
		err error
		_   = err
	)

	val, ok = pathParams["name"]
	if !ok {
		return nil, metadata, status.Errorf(codes.InvalidArgument, "missing parameter %s", "name")
	}

	protoReq.Name, err = runtime.String(val)
	if err != nil {
		return nil, metadata, status.Errorf(codes.InvalidArgument, "type mismatch, parameter: %s, error: %v", "name", err)
	}

	msg, err := server.GetSchemas(ctx, &protoReq)
	return msg, metadata, err

}

func request_SchemasService_GetLuaSchemasPlugin_0(ctx context.Context, marshaler runtime.Marshaler, client SchemasServiceClient, req *http.Request, pathParams map[string]string) (proto.Message, runtime.ServerMetadata, error) {
	var protoReq GetLuaSchemasPluginRequest
	var metadata runtime.ServerMetadata

	var (
		val string
		ok  bool
		err error
		_   = err
	)

	val, ok = pathParams["name"]
	if !ok {
		return nil, metadata, status.Errorf(codes.InvalidArgument, "missing parameter %s", "name")
	}

	protoReq.Name, err = runtime.String(val)
	if err != nil {
		return nil, metadata, status.Errorf(codes.InvalidArgument, "type mismatch, parameter: %s, error: %v", "name", err)
	}

	msg, err := client.GetLuaSchemasPlugin(ctx, &protoReq, grpc.Header(&metadata.HeaderMD), grpc.Trailer(&metadata.TrailerMD))
	return msg, metadata, err

}

func local_request_SchemasService_GetLuaSchemasPlugin_0(ctx context.Context, marshaler runtime.Marshaler, server SchemasServiceServer, req *http.Request, pathParams map[string]string) (proto.Message, runtime.ServerMetadata, error) {
	var protoReq GetLuaSchemasPluginRequest
	var metadata runtime.ServerMetadata

	var (
		val string
		ok  bool
		err error
		_   = err
	)

	val, ok = pathParams["name"]
	if !ok {
		return nil, metadata, status.Errorf(codes.InvalidArgument, "missing parameter %s", "name")
	}

	protoReq.Name, err = runtime.String(val)
	if err != nil {
		return nil, metadata, status.Errorf(codes.InvalidArgument, "type mismatch, parameter: %s, error: %v", "name", err)
	}

	msg, err := server.GetLuaSchemasPlugin(ctx, &protoReq)
	return msg, metadata, err

}

// RegisterSchemasServiceHandlerServer registers the http handlers for service SchemasService to "mux".
// UnaryRPC     :call SchemasServiceServer directly.
// StreamingRPC :currently unsupported pending https://github.com/grpc/grpc-go/issues/906.
// Note that using this registration option will cause many gRPC library features to stop working. Consider using RegisterSchemasServiceHandlerFromEndpoint instead.
func RegisterSchemasServiceHandlerServer(ctx context.Context, mux *runtime.ServeMux, server SchemasServiceServer) error {

	mux.Handle("GET", pattern_SchemasService_GetSchemas_0, func(w http.ResponseWriter, req *http.Request, pathParams map[string]string) {
		ctx, cancel := context.WithCancel(req.Context())
		defer cancel()
		var stream runtime.ServerTransportStream
		ctx = grpc.NewContextWithServerTransportStream(ctx, &stream)
		inboundMarshaler, outboundMarshaler := runtime.MarshalerForRequest(mux, req)
		rctx, err := runtime.AnnotateIncomingContext(ctx, mux, req, "/kong.admin.service.v1.SchemasService/GetSchemas", runtime.WithHTTPPathPattern("/v1/schemas/json/{name}"))
		if err != nil {
			runtime.HTTPError(ctx, mux, outboundMarshaler, w, req, err)
			return
		}
		resp, md, err := local_request_SchemasService_GetSchemas_0(rctx, inboundMarshaler, server, req, pathParams)
		md.HeaderMD, md.TrailerMD = metadata.Join(md.HeaderMD, stream.Header()), metadata.Join(md.TrailerMD, stream.Trailer())
		ctx = runtime.NewServerMetadataContext(ctx, md)
		if err != nil {
			runtime.HTTPError(ctx, mux, outboundMarshaler, w, req, err)
			return
		}

		forward_SchemasService_GetSchemas_0(ctx, mux, outboundMarshaler, w, req, response_SchemasService_GetSchemas_0{resp}, mux.GetForwardResponseOptions()...)

	})

	mux.Handle("GET", pattern_SchemasService_GetLuaSchemasPlugin_0, func(w http.ResponseWriter, req *http.Request, pathParams map[string]string) {
		ctx, cancel := context.WithCancel(req.Context())
		defer cancel()
		var stream runtime.ServerTransportStream
		ctx = grpc.NewContextWithServerTransportStream(ctx, &stream)
		inboundMarshaler, outboundMarshaler := runtime.MarshalerForRequest(mux, req)
		rctx, err := runtime.AnnotateIncomingContext(ctx, mux, req, "/kong.admin.service.v1.SchemasService/GetLuaSchemasPlugin", runtime.WithHTTPPathPattern("/v1/schemas/plugins/lua/{name}"))
		if err != nil {
			runtime.HTTPError(ctx, mux, outboundMarshaler, w, req, err)
			return
		}
		resp, md, err := local_request_SchemasService_GetLuaSchemasPlugin_0(rctx, inboundMarshaler, server, req, pathParams)
		md.HeaderMD, md.TrailerMD = metadata.Join(md.HeaderMD, stream.Header()), metadata.Join(md.TrailerMD, stream.Trailer())
		ctx = runtime.NewServerMetadataContext(ctx, md)
		if err != nil {
			runtime.HTTPError(ctx, mux, outboundMarshaler, w, req, err)
			return
		}

		forward_SchemasService_GetLuaSchemasPlugin_0(ctx, mux, outboundMarshaler, w, req, response_SchemasService_GetLuaSchemasPlugin_0{resp}, mux.GetForwardResponseOptions()...)

	})

	return nil
}

// RegisterSchemasServiceHandlerFromEndpoint is same as RegisterSchemasServiceHandler but
// automatically dials to "endpoint" and closes the connection when "ctx" gets done.
func RegisterSchemasServiceHandlerFromEndpoint(ctx context.Context, mux *runtime.ServeMux, endpoint string, opts []grpc.DialOption) (err error) {
	conn, err := grpc.Dial(endpoint, opts...)
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			if cerr := conn.Close(); cerr != nil {
				grpclog.Infof("Failed to close conn to %s: %v", endpoint, cerr)
			}
			return
		}
		go func() {
			<-ctx.Done()
			if cerr := conn.Close(); cerr != nil {
				grpclog.Infof("Failed to close conn to %s: %v", endpoint, cerr)
			}
		}()
	}()

	return RegisterSchemasServiceHandler(ctx, mux, conn)
}

// RegisterSchemasServiceHandler registers the http handlers for service SchemasService to "mux".
// The handlers forward requests to the grpc endpoint over "conn".
func RegisterSchemasServiceHandler(ctx context.Context, mux *runtime.ServeMux, conn *grpc.ClientConn) error {
	return RegisterSchemasServiceHandlerClient(ctx, mux, NewSchemasServiceClient(conn))
}

// RegisterSchemasServiceHandlerClient registers the http handlers for service SchemasService
// to "mux". The handlers forward requests to the grpc endpoint over the given implementation of "SchemasServiceClient".
// Note: the gRPC framework executes interceptors within the gRPC handler. If the passed in "SchemasServiceClient"
// doesn't go through the normal gRPC flow (creating a gRPC client etc.) then it will be up to the passed in
// "SchemasServiceClient" to call the correct interceptors.
func RegisterSchemasServiceHandlerClient(ctx context.Context, mux *runtime.ServeMux, client SchemasServiceClient) error {

	mux.Handle("GET", pattern_SchemasService_GetSchemas_0, func(w http.ResponseWriter, req *http.Request, pathParams map[string]string) {
		ctx, cancel := context.WithCancel(req.Context())
		defer cancel()
		inboundMarshaler, outboundMarshaler := runtime.MarshalerForRequest(mux, req)
		rctx, err := runtime.AnnotateContext(ctx, mux, req, "/kong.admin.service.v1.SchemasService/GetSchemas", runtime.WithHTTPPathPattern("/v1/schemas/json/{name}"))
		if err != nil {
			runtime.HTTPError(ctx, mux, outboundMarshaler, w, req, err)
			return
		}
		resp, md, err := request_SchemasService_GetSchemas_0(rctx, inboundMarshaler, client, req, pathParams)
		ctx = runtime.NewServerMetadataContext(ctx, md)
		if err != nil {
			runtime.HTTPError(ctx, mux, outboundMarshaler, w, req, err)
			return
		}

		forward_SchemasService_GetSchemas_0(ctx, mux, outboundMarshaler, w, req, response_SchemasService_GetSchemas_0{resp}, mux.GetForwardResponseOptions()...)

	})

	mux.Handle("GET", pattern_SchemasService_GetLuaSchemasPlugin_0, func(w http.ResponseWriter, req *http.Request, pathParams map[string]string) {
		ctx, cancel := context.WithCancel(req.Context())
		defer cancel()
		inboundMarshaler, outboundMarshaler := runtime.MarshalerForRequest(mux, req)
		rctx, err := runtime.AnnotateContext(ctx, mux, req, "/kong.admin.service.v1.SchemasService/GetLuaSchemasPlugin", runtime.WithHTTPPathPattern("/v1/schemas/plugins/lua/{name}"))
		if err != nil {
			runtime.HTTPError(ctx, mux, outboundMarshaler, w, req, err)
			return
		}
		resp, md, err := request_SchemasService_GetLuaSchemasPlugin_0(rctx, inboundMarshaler, client, req, pathParams)
		ctx = runtime.NewServerMetadataContext(ctx, md)
		if err != nil {
			runtime.HTTPError(ctx, mux, outboundMarshaler, w, req, err)
			return
		}

		forward_SchemasService_GetLuaSchemasPlugin_0(ctx, mux, outboundMarshaler, w, req, response_SchemasService_GetLuaSchemasPlugin_0{resp}, mux.GetForwardResponseOptions()...)

	})

	return nil
}

type response_SchemasService_GetSchemas_0 struct {
	proto.Message
}

func (m response_SchemasService_GetSchemas_0) XXX_ResponseBody() interface{} {
	response := m.Message.(*GetSchemasResponse)
	return response.Schema
}

type response_SchemasService_GetLuaSchemasPlugin_0 struct {
	proto.Message
}

func (m response_SchemasService_GetLuaSchemasPlugin_0) XXX_ResponseBody() interface{} {
	response := m.Message.(*GetLuaSchemasPluginResponse)
	return response.Schema
}

var (
	pattern_SchemasService_GetSchemas_0 = runtime.MustPattern(runtime.NewPattern(1, []int{2, 0, 2, 1, 2, 2, 1, 0, 4, 1, 5, 3}, []string{"v1", "schemas", "json", "name"}, ""))

	pattern_SchemasService_GetLuaSchemasPlugin_0 = runtime.MustPattern(runtime.NewPattern(1, []int{2, 0, 2, 1, 2, 2, 2, 3, 1, 0, 4, 1, 5, 4}, []string{"v1", "schemas", "plugins", "lua", "name"}, ""))
)

var (
	forward_SchemasService_GetSchemas_0 = runtime.ForwardResponseMessage

	forward_SchemasService_GetLuaSchemasPlugin_0 = runtime.ForwardResponseMessage
)
