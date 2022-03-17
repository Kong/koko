package resource

import (
	"net"
	"net/url"
	"reflect"
	"strconv"

	"github.com/google/uuid"
	v1 "github.com/kong/koko/internal/gen/grpc/kong/admin/model/v1"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

func defaultID(id *string) {
	if id == nil || *id == "" {
		*id = uuid.NewString()
	}
}

func parseURL(s *v1.Service) error {
	u, err := url.Parse(s.Url)
	if err != nil {
		return err
	}
	s.Protocol = u.Scheme
	host, port, err := net.SplitHostPort(u.Host)
	if err != nil {
		return err
	}
	intPort, err := strconv.Atoi(port) //nolint:gosec
	if err != nil {
		return err
	}
	s.Port = int32(intPort)
	s.Host = host
	s.Path = u.Path
	s.Url = ""
	return nil
}

type wrappersPBTransformer struct{}

func (t wrappersPBTransformer) Transformer(typ reflect.Type) func(dst, src reflect.Value) error {
	var b *wrapperspb.BoolValue
	switch typ {
	case reflect.TypeOf(b):
		return func(dst, src reflect.Value) error {
			if !dst.IsNil() {
				return nil
			}
			return nil
		}
	default:

		return nil
	}
}

func intP(i int) *int {
	return &i
}
