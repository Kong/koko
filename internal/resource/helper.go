package resource

import (
	"errors"
	"fmt"
	"net"
	"net/url"
	"reflect"
	"strconv"
	"strings"

	"github.com/google/uuid"
	v1 "github.com/kong/koko/internal/gen/grpc/kong/admin/model/v1"
	"github.com/kong/koko/internal/model"
	"github.com/kong/koko/internal/model/json/validation/typedefs"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

// SetResource replaces the object's underlining resource with the provided resource.
func SetResource(o model.Object, r model.Resource) error {
	expected, actual := o.Resource().ProtoReflect().Descriptor(), r.ProtoReflect().Descriptor()
	if expected != actual {
		return fmt.Errorf("unable to set resource: expected %q but got %q", expected.FullName(), actual.FullName())
	}
	dst := o.Resource()
	if !dst.ProtoReflect().IsValid() {
		return errors.New("unable to set resource: got invalid destination resource")
	}
	proto.Reset(dst)
	proto.Merge(dst, r)
	return nil
}

func defaultID(id *string) {
	if id == nil || *id == "" {
		*id = uuid.NewString()
	}
}

func parseURL(s *v1.Service) error {
	svcURL := s.Url
	s.Url = ""
	u, err := url.Parse(svcURL)
	if err != nil {
		return fmt.Errorf("parse url field: %v", err)
	}
	if u.Host == "" {
		return nil
	}
	s.Protocol = u.Scheme
	if s.Protocol == typedefs.ProtocolHTTP {
		s.Port = 80
	} else if s.Protocol == typedefs.ProtocolHTTPS {
		s.Port = 443
	}
	host := u.Host
	if strings.Contains(host, ":") {
		var portStr string
		host, portStr, err = net.SplitHostPort(u.Host)
		if err != nil {
			return fmt.Errorf("unpack host and port: %v", err)
		}
		port, err := strconv.Atoi(portStr) //nolint:gosec
		if err != nil {
			return fmt.Errorf("convert port field to int: %v", err)
		}

		s.Port = int32(port)
	}
	s.Host = host
	s.Path = u.Path
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
