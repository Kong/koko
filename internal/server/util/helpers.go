package util

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	pbModel "github.com/kong/koko/internal/gen/grpc/kong/admin/model/v1"
	"github.com/kong/koko/internal/model/json/validation"
	"github.com/kong/koko/internal/store"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	proto2 "google.golang.org/protobuf/proto"
)

const (
	StatusCodeKey = "koko-status-code"
)

type helperKey int

var routeKey helperKey

type RouteStatus struct {
	success bool
}

type ErrClient struct {
	Message string
}

func (e ErrClient) Error() string {
	return e.Message
}

func AddRouteStatus(ctx context.Context) context.Context {
	return context.WithValue(ctx, routeKey, &RouteStatus{success: true})
}

func RouteFailed(ctx context.Context) bool {
	route, ok := ctx.Value(routeKey).(*RouteStatus)
	if ok {
		return !route.success
	}
	return false
}

func RouteErrorHandler(ctx context.Context, mux *runtime.ServeMux, m runtime.Marshaler, w http.ResponseWriter, r *http.Request, status int) {
	route, ok := ctx.Value(routeKey).(*RouteStatus)
	if ok {
		route.success = false
	}
	runtime.DefaultRoutingErrorHandler(ctx, mux, m, w, r, status)
}

func SetHeader(ctx context.Context, code int) {
	err := grpc.SetHeader(ctx, metadata.Pairs(StatusCodeKey,
		fmt.Sprintf("%d", code)))
	if err != nil {
		panic(err)
	}
}

func SetHTTPStatus(ctx context.Context, w http.ResponseWriter,
	_ proto2.Message,
) error {
	md, ok := runtime.ServerMetadataFromContext(ctx)
	if !ok {
		return fmt.Errorf("no server metadata in context")
	}
	values := md.HeaderMD.Get(StatusCodeKey)
	if len(values) > 0 {
		code, err := strconv.Atoi(values[0])
		if err != nil {
			return err
		}
		defer w.WriteHeader(code)
	}
	w.Header().Del("grpc-metadata-" + StatusCodeKey)
	return nil
}

func HandleErr(logger *zap.Logger, err error) error {
	if errors.Is(err, store.ErrNotFound) {
		return status.Error(codes.NotFound, "")
	}

	switch e := err.(type) {
	case store.ErrConstraint:
		s := status.New(codes.InvalidArgument, "data constraint error")
		errDetail := &pbModel.ErrorDetail{
			Type:     pbModel.ErrorType_ERROR_TYPE_REFERENCE,
			Field:    e.Index.FieldName,
			Messages: []string{e.Error()},
		}
		s, err := s.WithDetails(errDetail)
		if err != nil {
			panic(err)
		}
		return s.Err()
	case validation.Error:
		s := status.New(codes.InvalidArgument, "validation error")
		var errs []Message
		for _, err := range e.Errs {
			errs = append(errs, err)
		}
		s, err := s.WithDetails(errs...)
		if err != nil {
			panic(err)
		}
		return s.Err()
	case ErrClient:
		s := status.New(codes.InvalidArgument, e.Message)
		return s.Err()
	default:
		logger.With(zap.Error(err)).Error("error in service")
		return status.Error(codes.Internal, "")
	}
}
