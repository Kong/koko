package resource

import (
	"fmt"
	"net/http"
	"strings"
	"testing"

	model "github.com/kong/koko/internal/gen/grpc/kong/admin/model/v1"
	"github.com/kong/koko/internal/model/json/validation"
	"github.com/kong/koko/internal/model/json/validation/typedefs"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

func TestNewRoute(t *testing.T) {
	r := NewRoute()
	require.NotNil(t, r)
	require.NotNil(t, r.Route)
}

func TestRoute_ID(t *testing.T) {
	var r Route
	id := r.ID()
	require.Empty(t, id)
	r = NewRoute()
	id = r.ID()
	require.Empty(t, id)
}

func TestRoute_Type(t *testing.T) {
	require.Equal(t, TypeRoute, NewRoute().Type())
}

func TestRoute_ProcessDefaults(t *testing.T) {
	t.Run("defaults are correctly injected", func(t *testing.T) {
		r := NewRoute()
		err := r.ProcessDefaults()
		require.Nil(t, err)
		require.True(t, validUUID(r.ID()))
		// empty out the id for equality comparison
		r.Route.Id = ""
		r.Route.CreatedAt = 0
		r.Route.UpdatedAt = 0
		require.Equal(t, r.Resource(), defaultRoute)
	})
	t.Run("defaults do not override explicit values", func(t *testing.T) {
		r := NewRoute()
		r.Route.HttpsRedirectStatusCode = 302
		r.Route.Protocols = []string{"grpc"}
		r.Route.PathHandling = "v1"
		r.Route.RegexPriority = wrapperspb.Int32(1)
		r.Route.PreserveHost = wrapperspb.Bool(true)
		r.Route.StripPath = wrapperspb.Bool(false)
		r.Route.RequestBuffering = wrapperspb.Bool(false)
		r.Route.ResponseBuffering = wrapperspb.Bool(false)
		err := r.ProcessDefaults()
		require.Nil(t, err)
		require.True(t, validUUID(r.ID()))
		// empty out the id and ts for equality comparison
		r.Route.Id = ""
		require.Equal(t, &model.Route{
			Protocols:               []string{typedefs.ProtocolGRPC},
			RegexPriority:           wrapperspb.Int32(1),
			PreserveHost:            wrapperspb.Bool(true),
			StripPath:               wrapperspb.Bool(false),
			RequestBuffering:        wrapperspb.Bool(false),
			ResponseBuffering:       wrapperspb.Bool(false),
			PathHandling:            "v1",
			HttpsRedirectStatusCode: 302,
		}, r.Resource())
	})
}

func goodRoute() Route {
	r := NewRoute()
	_ = r.ProcessDefaults()
	r.Route.Hosts = []string{"good.example.com"}
	return r
}

func TestRoute_Validate(t *testing.T) {
	tests := []struct {
		name    string
		Route   func() Route
		wantErr bool
		Errs    []*model.ErrorDetail
		Skip    bool
	}{
		{
			name: "empty route throws an error",
			Route: func() Route {
				return NewRoute()
			},
			wantErr: true,
			Errs: []*model.ErrorDetail{
				{
					Type: model.ErrorType_ERROR_TYPE_ENTITY,
					Messages: []string{
						"missing properties: 'id', 'protocols'",
					},
				},
			},
		},
		{
			name: "good route doesn't throw any error",
			Route: func() Route {
				return goodRoute()
			},
			wantErr: false,
			Errs:    []*model.ErrorDetail{},
		},
		{
			name: "default route throws an error",
			Route: func() Route {
				r := NewRoute()
				_ = r.ProcessDefaults()
				return r
			},
			wantErr: true,
			Errs: []*model.ErrorDetail{
				{
					Type: model.ErrorType_ERROR_TYPE_ENTITY,
					Messages: []string{
						"when protocols has 'http', at least one of 'hosts', " +
							"'methods', 'paths' or 'headers' must be set",
						"when protocols has 'https', " +
							"at least one of 'snis', 'hosts', 'methods', 'paths' or 'headers' must be set",
					},
				},
			},
		},
		{
			name: "invalid ID throws an error",
			Route: func() Route {
				r := goodRoute()
				r.Route.Id = "bad-uuid"
				return r
			},
			wantErr: true,
			Errs: []*model.ErrorDetail{
				{
					Type:  model.ErrorType_ERROR_TYPE_FIELD,
					Field: "id",
					Messages: []string{
						"must be a valid UUID",
					},
				},
			},
		},
		{
			name: "invalid name throws an error",
			Route: func() Route {
				r := goodRoute()
				r.Route.Name = "%foo"
				return r
			},
			wantErr: true,
			Errs: []*model.ErrorDetail{
				{
					Type:  model.ErrorType_ERROR_TYPE_FIELD,
					Field: "name",
					Messages: []string{
						"must match pattern '^[0-9a-zA-Z.\\-_~]*$'",
					},
				},
			},
		},
		{
			name: "invalid host throws an error",
			Route: func() Route {
				r := goodRoute()
				r.Route.Hosts = []string{"%foo"}
				return r
			},
			wantErr: true,
			Errs: []*model.ErrorDetail{
				{
					Type:  model.ErrorType_ERROR_TYPE_FIELD,
					Field: "hosts[0]",
					Messages: []string{
						"must be a valid hostname",
					},
				},
			},
		},
		{
			name: "invalid tags throws an error",
			Route: func() Route {
				r := goodRoute()
				r.Route.Tags = []string{"$tag"}
				return r
			},
			wantErr: true,
			Errs: []*model.ErrorDetail{
				{
					Type:  model.ErrorType_ERROR_TYPE_FIELD,
					Field: "tags[0]",
					Messages: []string{
						"must match pattern '^[0-9a-zA-Z.\\-_~]*$'",
					},
				},
			},
		},
		{
			name: "more than 8 tags throw an error",
			Route: func() Route {
				r := goodRoute()
				r.Route.Tags = []string{
					"tag0",
					"tag1",
					"tag2",
					"tag3",
					"tag4",
					"tag5",
					"tag6",
					"tag7",
					"tag8",
				}
				return r
			},
			wantErr: true,
			Errs: []*model.ErrorDetail{
				{
					Type:  model.ErrorType_ERROR_TYPE_FIELD,
					Field: "tags",
					Messages: []string{
						"maximum 8 items required, but found 9 items",
					},
				},
			},
		},
		{
			name: "name longer than 128 character errors out",
			Route: func() Route {
				r := goodRoute()
				r.Route.
					Name = "anyservicewithareallylongnameisnotveryhelpful" + "" +
					"toanyoneatallisitifyouthinkitisthistestisgoingtoprove" +
					"youwrongandifitdoesnottheniguesswearedoomed"
				return r
			},
			wantErr: true,
			Errs: []*model.ErrorDetail{
				{
					Type:  model.ErrorType_ERROR_TYPE_FIELD,
					Field: "name",
					Messages: []string{
						"length must be <= 128, but got 141",
					},
				},
			},
		},
		{
			name: "invalid protocol throws an error",
			Route: func() Route {
				r := goodRoute()
				r.Route.Protocols = []string{"smtp"}
				return r
			},
			wantErr: true,
			Errs: []*model.ErrorDetail{
				{
					Type:  model.ErrorType_ERROR_TYPE_FIELD,
					Field: "protocols[0]",
					Messages: []string{
						`value must be one of "http", "https", "grpc", ` +
							`"grpcs", "tcp", "udp", "tls", "tls_passthrough"`,
						"must contain only one subset [ http https ]",
						"must contain only one subset [ tcp udp tls ]",
						"must contain only one subset [ grpc grpcs ]",
						"must contain only one subset [ tls_passthrough ]",
					},
				},
			},
		},
		{
			name: "invalid method throws an error",
			Route: func() Route {
				r := goodRoute()
				r.Route.Methods = []string{"lower-case-method"}
				return r
			},
			wantErr: true,
			Errs: []*model.ErrorDetail{
				{
					Type:  model.ErrorType_ERROR_TYPE_FIELD,
					Field: "methods[0]",
					Messages: []string{
						"must match pattern '^[A-Z]+$'",
					},
				},
			},
		},
		{
			name: "invalid paths throws an error",
			Route: func() Route {
				r := goodRoute()
				r.Route.Paths = []string{
					"/valid-path",
					"invalid-path",
					"/path/must/not/have/two//slashes",
				}
				return r
			},
			wantErr: true,
			Errs: []*model.ErrorDetail{
				{
					Type:  model.ErrorType_ERROR_TYPE_FIELD,
					Field: "paths[1]",
					Messages: []string{
						"must begin with `/`",
					},
				},
				{
					Type:  model.ErrorType_ERROR_TYPE_FIELD,
					Field: "paths[2]",
					Messages: []string{
						"must not contain `//`",
					},
				},
			},
		},
		{
			name: "long paths throws an error",
			Route: func() Route {
				r := goodRoute()
				path := strings.Repeat("/longpath", 114)
				r.Route.Paths = []string{path}
				return r
			},
			wantErr: true,
			Errs: []*model.ErrorDetail{
				{
					Type:  model.ErrorType_ERROR_TYPE_FIELD,
					Field: "paths[0]",
					Messages: []string{
						"length must not exceed 1024",
					},
				},
			},
		},
		{
			name: "route with empty service throws an error",
			Route: func() Route {
				r := goodRoute()
				r.Route.Service = &model.Service{}
				return r
			},
			wantErr: true,
			Errs: []*model.ErrorDetail{
				{
					Type:  model.ErrorType_ERROR_TYPE_FIELD,
					Field: "service",
					Messages: []string{
						"missing properties: 'id'",
					},
				},
			},
		},
		{
			name: "invalid https_redirect_status_code throws an error",
			Route: func() Route {
				r := goodRoute()
				r.Route.HttpsRedirectStatusCode = http.StatusBadRequest
				return r
			},
			wantErr: true,
			Errs: []*model.ErrorDetail{
				{
					Type:  model.ErrorType_ERROR_TYPE_FIELD,
					Field: "https_redirect_status_code",
					Messages: []string{
						`value must be one of "426", "301", "302", "307", "308"`,
					},
				},
			},
		},
		{
			name: "more than 16 hosts throw an error",
			Route: func() Route {
				r := goodRoute()
				for i := 0; i < 17; i++ {
					r.Route.Hosts = append(r.Route.Hosts,
						fmt.Sprintf("f%d.example.com", i))
				}
				return r
			},
			wantErr: true,
			Errs: []*model.ErrorDetail{
				{
					Type:  model.ErrorType_ERROR_TYPE_FIELD,
					Field: "hosts",
					Messages: []string{
						"maximum 16 items required, but found 18 items",
					},
				},
			},
		},
		{
			name: "host header cannot be part of headers",
			Route: func() Route {
				r := goodRoute()
				r.Route.Headers = map[string]*model.HeaderValues{
					"foo":  {Values: []string{"bar"}},
					"host": {Values: []string{"bad-key"}},
				}
				return r
			},
			wantErr: true,
			Errs: []*model.ErrorDetail{
				{
					Type:  model.ErrorType_ERROR_TYPE_FIELD,
					Field: "headers.host",
					Messages: []string{
						"must not contain 'host' header",
					},
				},
			},
		},
		{
			name: "header key more than 64 chars errors",
			Route: func() Route {
				r := goodRoute()
				longKey := strings.Repeat("buzz", 17)
				r.Route.Headers = map[string]*model.HeaderValues{
					longKey: {Values: []string{"bar"}},
				}
				return r
			},
			wantErr: true,
			Errs: []*model.ErrorDetail{
				{
					Type:  model.ErrorType_ERROR_TYPE_FIELD,
					Field: "headers",
					Messages: []string{
						"additionalProperties" +
							" 'buzzbuzzbuzzbuzzbuzzbuzzbuzzbuzzbuzzbuzzbuzzbuzzbuzzbuzzbuzzbuzzbuzz' not allowed",
					},
				},
			},
		},
		{
			name: "header value more than 64 chars errors",
			Route: func() Route {
				r := goodRoute()
				longValue := strings.Repeat("buzz", 17)
				r.Route.Headers = map[string]*model.HeaderValues{
					"foo": {Values: []string{longValue}},
				}
				return r
			},
			wantErr: true,
			Errs: []*model.ErrorDetail{
				{
					Type:  model.ErrorType_ERROR_TYPE_FIELD,
					Field: "headers.foo.values[0]",
					Messages: []string{
						"length must be <= 64, but got 68",
					},
				},
			},
		},
		{
			name: "invalid headers throws an error",
			Route: func() Route {
				r := goodRoute()
				r.Route.Headers = map[string]*model.HeaderValues{
					"f[]oo@bar.com": {Values: []string{"bad-key"}},
				}
				return r
			},
			wantErr: true,
			// TODO(hbagdi): cant get this to throw an error
			// requires a deeper understanding of patternProperties
			// in JSON-schema
			Skip: true,
		},
		{
			name: "negative regex priority throws an error",
			Route: func() Route {
				r := goodRoute()
				r.Route.RegexPriority = wrapperspb.Int32(-32)
				return r
			},
			wantErr: true,
			Errs: []*model.ErrorDetail{
				{
					Type:  model.ErrorType_ERROR_TYPE_FIELD,
					Field: "regex_priority",
					Messages: []string{
						"must be > -1 but found -32",
					},
				},
			},
		},
		{
			name: "invalid path_handling throws an error",
			Route: func() Route {
				r := goodRoute()
				r.Route.PathHandling = "foo-invalid"
				return r
			},
			wantErr: true,
			Errs: []*model.ErrorDetail{
				{
					Type:  model.ErrorType_ERROR_TYPE_FIELD,
					Field: "path_handling",
					Messages: []string{
						`value must be one of "v0", "v1"`,
					},
				},
			},
		},
		{
			name: "invalid CIDR range throws and error",
			Route: func() Route {
				r := NewRoute()
				_ = r.ProcessDefaults()
				r.Route.Protocols = []string{typedefs.ProtocolTCP}
				r.Route.Destinations = []*model.CIDRPort{
					{
						Ip: "foobar",
					},
					{
						Port: -32,
					},
				}
				r.Route.Sources = []*model.CIDRPort{
					{
						Ip: "10.0.0/8",
					},
				}
				return r
			},
			wantErr: true,
			Errs: []*model.ErrorDetail{
				{
					Type:  model.ErrorType_ERROR_TYPE_FIELD,
					Field: "destinations[0].ip",
					Messages: []string{
						"must be a valid IP or CIDR",
					},
				},
				{
					Type:  model.ErrorType_ERROR_TYPE_FIELD,
					Field: "destinations[1].port",
					Messages: []string{
						"must be >= 1 but found -32",
					},
				},
				{
					Type:  model.ErrorType_ERROR_TYPE_FIELD,
					Field: "sources[0].ip",
					Messages: []string{
						"must be a valid IP or CIDR",
					},
				},
			},
		},
		{
			name: "setting sni with http protocol errors",
			Route: func() Route {
				r := goodRoute()
				r.Route.Protocols = []string{typedefs.ProtocolHTTP}
				r.Route.Snis = []string{"foo.example.com"}
				return r
			},
			wantErr: true,
			Errs: []*model.ErrorDetail{
				{
					Type: model.ErrorType_ERROR_TYPE_ENTITY,
					Messages: []string{
						"'snis' can be set only when protocols has one of" +
							" 'https', 'grpcs', 'tls' or 'tls_passthrough'",
					},
				},
			},
		},
		{
			name: "setting sni with tcp protocol errors",
			Route: func() Route {
				r := NewRoute()
				_ = r.ProcessDefaults()
				r.Route.Sources = []*model.CIDRPort{
					{
						Ip: "10.0.0.0/8",
					},
				}
				r.Route.Protocols = []string{typedefs.ProtocolTCP}
				r.Route.Snis = []string{"foo.example.com"}
				return r
			},
			wantErr: true,
			Errs: []*model.ErrorDetail{
				{
					Type: model.ErrorType_ERROR_TYPE_ENTITY,
					Messages: []string{
						"'snis' can be set only when protocols has one of" +
							" 'https', 'grpcs', 'tls' or 'tls_passthrough'",
					},
				},
			},
		},
		{
			name: "setting sni with udp protocol errors",
			Route: func() Route {
				r := NewRoute()
				_ = r.ProcessDefaults()
				r.Route.Sources = []*model.CIDRPort{
					{
						Ip: "10.0.0.0/8",
					},
				}
				r.Route.Protocols = []string{typedefs.ProtocolUDP}
				r.Route.Snis = []string{"foo.example.com"}
				return r
			},
			wantErr: true,
			Errs: []*model.ErrorDetail{
				{
					Type: model.ErrorType_ERROR_TYPE_ENTITY,
					Messages: []string{
						"'snis' can be set only when protocols has one of" +
							" 'https', 'grpcs', 'tls' or 'tls_passthrough'",
					},
				},
			},
		},
		{
			name: "setting sources, or destination with http protocol errors",
			Route: func() Route {
				r := NewRoute()
				_ = r.ProcessDefaults()
				r.Route.Sources = []*model.CIDRPort{
					{
						Ip: "10.0.0.0/8",
					},
				}
				r.Route.Protocols = []string{typedefs.ProtocolHTTP}
				return r
			},
			wantErr: true,
			Errs: []*model.ErrorDetail{
				{
					Type: model.ErrorType_ERROR_TYPE_ENTITY,
					Messages: []string{
						"when protocols has 'http' or 'https', " +
							"'sources' or 'destinations' cannot be set",
						"when protocols has 'http', " +
							"at least one of 'hosts', 'methods', " +
							"'paths' or 'headers' must be set",
					},
				},
			},
		},
		{
			name: "setting sources, or destination with https protocol errors",
			Route: func() Route {
				r := NewRoute()
				_ = r.ProcessDefaults()
				r.Route.Sources = []*model.CIDRPort{
					{
						Ip: "10.0.0.0/8",
					},
				}
				r.Route.Protocols = []string{typedefs.ProtocolHTTPS}
				return r
			},
			wantErr: true,
			Errs: []*model.ErrorDetail{
				{
					Type: model.ErrorType_ERROR_TYPE_ENTITY,
					Messages: []string{
						"when protocols has 'http' or 'https', " +
							"'sources' or 'destinations' cannot be set",
						"when protocols has 'https', " +
							"at least one of 'snis', 'hosts', 'methods', " +
							"'paths' or 'headers' must be set",
					},
				},
			},
		},
		{
			name: "setting methods with UDP, TCP or TLS protocol errors",
			Route: func() Route {
				r := NewRoute()
				_ = r.ProcessDefaults()
				r.Route.Sources = []*model.CIDRPort{
					{
						Ip: "10.0.0.0/8",
					},
				}
				r.Route.Methods = []string{"GET"}
				r.Route.Protocols = []string{
					typedefs.ProtocolTCP,
					typedefs.ProtocolUDP,
					typedefs.ProtocolTLS,
				}
				return r
			},
			wantErr: true,
			Errs: []*model.ErrorDetail{
				{
					Type: model.ErrorType_ERROR_TYPE_ENTITY,
					Messages: []string{
						"when protocol has 'tcp', 'tls', 'tls_passthrough' or 'udp', " +
							"'methods', 'hosts', 'paths', 'headers' cannot be set",
					},
				},
			},
		},
		{
			name: "setting paths with UDP, TCP or TLS protocol errors",
			Route: func() Route {
				r := NewRoute()
				_ = r.ProcessDefaults()
				r.Route.Sources = []*model.CIDRPort{
					{
						Ip: "10.0.0.0/8",
					},
				}
				r.Route.Paths = []string{"/foo"}
				r.Route.Protocols = []string{
					typedefs.ProtocolTCP,
					typedefs.ProtocolUDP,
					typedefs.ProtocolTLS,
				}
				return r
			},
			wantErr: true,
			Errs: []*model.ErrorDetail{
				{
					Type: model.ErrorType_ERROR_TYPE_ENTITY,
					Messages: []string{
						"when protocol has 'tcp', 'tls', 'tls_passthrough' or 'udp', " +
							"'methods', 'hosts', 'paths', 'headers' cannot be set",
					},
				},
			},
		},
		{
			name: "setting headers with UDP, TCP or TLS protocol errors",
			Route: func() Route {
				r := NewRoute()
				_ = r.ProcessDefaults()
				r.Route.Sources = []*model.CIDRPort{
					{
						Ip: "10.0.0.0/8",
					},
				}
				r.Route.Headers = map[string]*model.HeaderValues{
					"foo": {Values: []string{"bar"}},
				}
				r.Route.Protocols = []string{
					typedefs.ProtocolTCP,
					typedefs.ProtocolUDP,
					typedefs.ProtocolTLS,
				}
				return r
			},
			wantErr: true,
			Errs: []*model.ErrorDetail{
				{
					Type: model.ErrorType_ERROR_TYPE_ENTITY,
					Messages: []string{
						"when protocol has 'tcp', 'tls', 'tls_passthrough' or 'udp', " +
							"'methods', 'hosts', 'paths', 'headers' cannot be set",
					},
				},
			},
		},
		{
			name: "setting hosts with UDP, TCP or TLS protocol errors",
			Route: func() Route {
				r := NewRoute()
				_ = r.ProcessDefaults()
				r.Route.Sources = []*model.CIDRPort{
					{
						Ip: "10.0.0.0/8",
					},
				}
				r.Route.Hosts = []string{"foo.example.com"}
				r.Route.Protocols = []string{
					typedefs.ProtocolTCP,
					typedefs.ProtocolUDP,
					typedefs.ProtocolTLS,
				}
				return r
			},
			wantErr: true,
			Errs: []*model.ErrorDetail{
				{
					Type: model.ErrorType_ERROR_TYPE_ENTITY,
					Messages: []string{
						"when protocol has 'tcp', 'tls', 'tls_passthrough' or 'udp', " +
							"'methods', 'hosts', 'paths', 'headers' cannot be set",
					},
				},
			},
		},
		{
			name: "when protocol has tcp, tls or udp, " +
				"at least one of sources, destinations or snis must be set",
			Route: func() Route {
				r := NewRoute()
				_ = r.ProcessDefaults()
				r.Route.Protocols = []string{
					typedefs.ProtocolTCP,
					typedefs.ProtocolUDP,
					typedefs.ProtocolTLS,
				}
				return r
			},
			wantErr: true,
			Errs: []*model.ErrorDetail{
				{
					Type: model.ErrorType_ERROR_TYPE_ENTITY,
					Messages: []string{
						"when protocols has 'tcp', 'tls' or 'udp', " +
							"then at least one of " +
							"'sources', 'destinations' or 'snis' must be set",
					},
				},
			},
		},
		{
			name: "when protocol has grpc, sources cannot be set",
			Route: func() Route {
				r := NewRoute()
				_ = r.ProcessDefaults()
				r.Route.Sources = []*model.CIDRPort{
					{
						Ip: "10.0.0.0/8",
					},
				}
				r.Route.Hosts = []string{"foo.example.com"}
				r.Route.Protocols = []string{
					typedefs.ProtocolGRPC,
					typedefs.ProtocolGRPCS,
				}
				return r
			},
			wantErr: true,
			Errs: []*model.ErrorDetail{
				{
					Type: model.ErrorType_ERROR_TYPE_ENTITY,
					Messages: []string{
						"when protocol has 'grpc' or 'grpcs', 'strip_path', " +
							"'methods', 'sources', 'destinations' cannot be set",
					},
				},
			},
		},
		{
			name: "when protocol has grpc, destination cannot be set",
			Route: func() Route {
				r := NewRoute()
				_ = r.ProcessDefaults()
				r.Route.Destinations = []*model.CIDRPort{
					{
						Ip: "10.0.0.0/8",
					},
				}
				r.Route.Hosts = []string{"foo.example.com"}
				r.Route.Protocols = []string{
					typedefs.ProtocolGRPC,
					typedefs.ProtocolGRPCS,
				}
				return r
			},
			wantErr: true,
			Errs: []*model.ErrorDetail{
				{
					Type: model.ErrorType_ERROR_TYPE_ENTITY,
					Messages: []string{
						"when protocol has 'grpc' or 'grpcs', 'strip_path', " +
							"'methods', 'sources', 'destinations' cannot be set",
					},
				},
			},
		},
		{
			name: "when protocol has grpc, strip_path cannot be set",
			Route: func() Route {
				r := NewRoute()
				_ = r.ProcessDefaults()
				r.Route.StripPath = wrapperspb.Bool(true)
				r.Route.Hosts = []string{"foo.example.com"}
				r.Route.Protocols = []string{
					typedefs.ProtocolGRPC,
					typedefs.ProtocolGRPCS,
				}
				return r
			},
			wantErr: true,
			Errs: []*model.ErrorDetail{
				{
					Type: model.ErrorType_ERROR_TYPE_ENTITY,
					Messages: []string{
						"when protocol has 'grpc' or 'grpcs', 'strip_path', " +
							"'methods', 'sources', 'destinations' cannot be set",
					},
				},
			},
		},
		{
			name: "when protocol has grpc, methods cannot be set",
			Route: func() Route {
				r := NewRoute()
				_ = r.ProcessDefaults()
				r.Route.Methods = []string{"POST"}
				r.Route.Hosts = []string{"foo.example.com"}
				r.Route.Protocols = []string{
					typedefs.ProtocolGRPC,
					typedefs.ProtocolGRPCS,
				}
				return r
			},
			wantErr: true,
			Errs: []*model.ErrorDetail{
				{
					Type: model.ErrorType_ERROR_TYPE_ENTITY,
					Messages: []string{
						"when protocol has 'grpc' or 'grpcs', 'strip_path', " +
							"'methods', 'sources', 'destinations' cannot be set",
					},
				},
			},
		},
		{
			name: "when protocol has grpcs, at least one of hosts, " +
				"headers, paths or snis must be set",
			Route: func() Route {
				r := NewRoute()
				r.Route.Protocols = []string{
					typedefs.ProtocolGRPCS,
				}
				_ = r.ProcessDefaults()
				return r
			},
			wantErr: true,
			Errs: []*model.ErrorDetail{
				{
					Type: model.ErrorType_ERROR_TYPE_ENTITY,
					Messages: []string{
						"when protocols has 'grpcs', " +
							"at least one of 'hosts', 'headers', 'paths' or 'snis' must be set",
					},
				},
			},
		},
		{
			name: "protocol has tls_passthrough, but no snis is set",
			Route: func() Route {
				r := NewRoute()
				r.Route.Protocols = []string{
					typedefs.ProtocolTLSPassthrough,
				}
				_ = r.ProcessDefaults()
				return r
			},
			wantErr: true,
			Errs: []*model.ErrorDetail{
				{
					Type: model.ErrorType_ERROR_TYPE_ENTITY,
					Messages: []string{
						"when protocols has 'tls_passthrough', 'snis' must be set",
					},
				},
			},
		},
		{
			name: "protocol has tls_passthrough, and snis are set",
			Route: func() Route {
				r := NewRoute()
				r.Route.Protocols = []string{
					typedefs.ProtocolTLSPassthrough,
				}
				r.Route.Snis = []string{"snis"}
				_ = r.ProcessDefaults()
				return r
			},
		},
		{
			name: "snis set but with unsupported protocol",
			Route: func() Route {
				r := NewRoute()
				r.Route.Protocols = []string{
					typedefs.ProtocolHTTP,
				}
				r.Route.Snis = []string{"snis"}
				r.Route.Methods = []string{"GET"}
				_ = r.ProcessDefaults()
				return r
			},
			wantErr: true,
			Errs: []*model.ErrorDetail{
				{
					Type: model.ErrorType_ERROR_TYPE_ENTITY,
					Messages: []string{
						"'snis' can be set only when protocols has one of" +
							" 'https', 'grpcs', 'tls' or 'tls_passthrough'",
					},
				},
			},
		},
		{
			name: "snis set with supported protocol",
			Route: func() Route {
				r := NewRoute()
				r.Route.Protocols = []string{
					typedefs.ProtocolTLSPassthrough,
				}
				r.Route.Snis = []string{"snis"}
				_ = r.ProcessDefaults()
				return r
			},
		},
		{
			name: "cannot set methods with tls_passthrough protocol",
			Route: func() Route {
				r := NewRoute()
				r.Route.Protocols = []string{
					typedefs.ProtocolTLSPassthrough,
				}
				r.Route.Methods = []string{"GET"}
				r.Route.Snis = []string{"snis"}
				_ = r.ProcessDefaults()
				return r
			},
			wantErr: true,
			Errs: []*model.ErrorDetail{
				{
					Type: model.ErrorType_ERROR_TYPE_ENTITY,
					Messages: []string{
						"when protocol has 'tcp', 'tls', 'tls_passthrough' or 'udp', " +
							"'methods', 'hosts', 'paths', 'headers' cannot be set",
					},
				},
			},
		},
		{
			name: "setting methods with TLSPassthrough protocol errors",
			Route: func() Route {
				r := NewRoute()
				_ = r.ProcessDefaults()
				r.Route.Methods = []string{"GET"}
				r.Route.Protocols = []string{
					typedefs.ProtocolTLSPassthrough,
				}
				r.Route.Snis = []string{"snis"}
				return r
			},
			wantErr: true,
			Errs: []*model.ErrorDetail{
				{
					Type: model.ErrorType_ERROR_TYPE_ENTITY,
					Messages: []string{
						"when protocol has 'tcp', 'tls', 'tls_passthrough' or 'udp', " +
							"'methods', 'hosts', 'paths', 'headers' cannot be set",
					},
				},
			},
		},
		{
			name: "setting paths with TLSPassthrough protocol errors",
			Route: func() Route {
				r := NewRoute()
				_ = r.ProcessDefaults()
				r.Route.Paths = []string{"/foo"}
				r.Route.Protocols = []string{
					typedefs.ProtocolTLSPassthrough,
				}
				r.Route.Snis = []string{"snis"}
				return r
			},
			wantErr: true,
			Errs: []*model.ErrorDetail{
				{
					Type: model.ErrorType_ERROR_TYPE_ENTITY,
					Messages: []string{
						"when protocol has 'tcp', 'tls', 'tls_passthrough' or 'udp', " +
							"'methods', 'hosts', 'paths', 'headers' cannot be set",
					},
				},
			},
		},
		{
			name: "setting headers with TLSPassthrough protocol errors",
			Route: func() Route {
				r := NewRoute()
				_ = r.ProcessDefaults()
				r.Route.Headers = map[string]*model.HeaderValues{
					"foo": {Values: []string{"bar"}},
				}
				r.Route.Protocols = []string{
					typedefs.ProtocolTLSPassthrough,
				}
				r.Route.Snis = []string{"snis"}
				return r
			},
			wantErr: true,
			Errs: []*model.ErrorDetail{
				{
					Type: model.ErrorType_ERROR_TYPE_ENTITY,
					Messages: []string{
						"when protocol has 'tcp', 'tls', 'tls_passthrough' or 'udp', " +
							"'methods', 'hosts', 'paths', 'headers' cannot be set",
					},
				},
			},
		},
		{
			name: "setting hosts with TLSPassthrough protocol errors",
			Route: func() Route {
				r := NewRoute()
				_ = r.ProcessDefaults()
				r.Route.Hosts = []string{"foo.example.com"}
				r.Route.Protocols = []string{
					typedefs.ProtocolTLSPassthrough,
				}
				r.Route.Snis = []string{"snis"}
				return r
			},
			wantErr: true,
			Errs: []*model.ErrorDetail{
				{
					Type: model.ErrorType_ERROR_TYPE_ENTITY,
					Messages: []string{
						"when protocol has 'tcp', 'tls', 'tls_passthrough' or 'udp', " +
							"'methods', 'hosts', 'paths', 'headers' cannot be set",
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.Skip {
				t.Skip()
			}
			route := tt.Route()
			err := route.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.Errs != nil {
				verr, _ := err.(validation.Error)
				require.ElementsMatch(t, tt.Errs, verr.Errs)
			}
		})
	}
}
