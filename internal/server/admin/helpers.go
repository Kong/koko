package admin

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"time"

	proto1 "github.com/golang/protobuf/proto"
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
	statusCodeKey = "koko-status-code"

	dbQueryTimeout = 5 * time.Second
)

type ErrClient struct {
	Message string
}

func (e ErrClient) Error() string {
	return e.Message
}

func setHeader(ctx context.Context, code int) {
	err := grpc.SetHeader(ctx, metadata.Pairs(statusCodeKey,
		fmt.Sprintf("%d", code)))
	if err != nil {
		panic(err)
	}
}

func setHTTPStatus(ctx context.Context, w http.ResponseWriter,
	_ proto2.Message) error {
	md, ok := runtime.ServerMetadataFromContext(ctx)
	if !ok {
		return fmt.Errorf("no server metadata in context")
	}
	values := md.HeaderMD.Get(statusCodeKey)
	if len(values) > 0 {
		code, err := strconv.Atoi(values[0])
		if err != nil {
			return err
		}
		defer w.WriteHeader(code)
	}
	w.Header().Del("grpc-metadata-" + statusCodeKey)
	return nil
}

func handleErr(logger *zap.Logger, err error) error {
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
		var errs []proto1.Message
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
