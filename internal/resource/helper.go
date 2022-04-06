package resource

import (
	"fmt"
	"net"
	"net/url"
	"reflect"
	"strconv"
	"strings"

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
	if s.Protocol == "http" {
		s.Port = 80
	} else if s.Protocol == "https" {
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
