package resource

import (
	"fmt"
	"net/http"

	"github.com/imdario/mergo"
	v1 "github.com/kong/koko/internal/gen/grpc/kong/admin/model/v1"
	"github.com/kong/koko/internal/model"
	"github.com/kong/koko/internal/model/json/generator"
	"github.com/kong/koko/internal/model/json/validation"
	"github.com/kong/koko/internal/model/json/validation/typedefs"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

const (
	// TypeRoute denotes the Route type.
	TypeRoute model.Type = "route"
)

var (
	defaultRoute = &v1.Route{
		Protocols:               []string{typedefs.ProtocolHTTP, typedefs.ProtocolHTTPS},
		RegexPriority:           wrapperspb.Int32(0),
		PreserveHost:            wrapperspb.Bool(false),
		StripPath:               wrapperspb.Bool(true),
		RequestBuffering:        wrapperspb.Bool(true),
		ResponseBuffering:       wrapperspb.Bool(true),
		PathHandling:            "v0",
		HttpsRedirectStatusCode: http.StatusUpgradeRequired,
	}
	_ model.Object = Route{}
)

func NewRoute() Route {
	return Route{
		Route: &v1.Route{},
	}
}

type Route struct {
	Route *v1.Route
}

func (r Route) ID() string {
	if r.Route == nil {
		return ""
	}
	return r.Route.Id
}

func (r Route) Type() model.Type {
	return TypeRoute
}

func (r Route) Resource() model.Resource {
	return r.Route
}

func (r Route) Indexes() []model.Index {
	res := []model.Index{
		{
			Name:      "name",
			Type:      model.IndexUnique,
			Value:     r.Route.Name,
			FieldName: "name",
		},
	}
	if r.Route.Service != nil {
		res = append(res, model.Index{
			Name:        "svc_id",
			Type:        model.IndexForeign,
			ForeignType: TypeService,
			FieldName:   "service.id",
			Value:       r.Route.Service.Id,
		})
	}
	return res
}

func (r Route) Validate() error {
	return validation.Validate(string(TypeRoute), r.Route)
}

func (r Route) ProcessDefaults() error {
	if r.Route == nil {
		return fmt.Errorf("invalid nil resource")
	}
	if len(r.Route.Protocols) == 0 {
		err := mergo.Merge(r.Route, defaultRoute,
			mergo.WithTransformers(wrappersPBTransformer{}))
		if err != nil {
			return err
		}
	}
	defaultID(&r.Route.Id)
	return nil
}

func init() {
	err := model.RegisterType(TypeRoute, func() model.Object {
		return NewRoute()
	})
	if err != nil {
		panic(err)
	}

	routeSchema := &generator.Schema{
		Type: "object",
		Properties: map[string]*generator.Schema{
			"id":   typedefs.ID,
			"name": typedefs.Name,
			"protocols": {
				Type:  "array",
				Items: typedefs.Protocol,
				AnyOf: []*generator.Schema{
					{
						Description: "must contain only one subset [ http" +
							" https ]",
						Items: &generator.Schema{
							Type: "string",
							Enum: []interface{}{
								typedefs.ProtocolHTTP,
								typedefs.ProtocolHTTPS,
							},
						},
					},
					{
						Description: "must contain only one subset [ tcp" +
							" udp tls ]",
						Items: &generator.Schema{
							Type: "string",
							Enum: []interface{}{
								typedefs.ProtocolTCP,
								typedefs.ProtocolUDP,
								typedefs.ProtocolTLS,
							},
						},
					},
					{
						Description: "must contain only one subset [ grpc" +
							" grpcs ]",
						Items: &generator.Schema{
							Type: "string",
							Enum: []interface{}{
								typedefs.ProtocolGRPC,
								typedefs.ProtocolGRPCS,
							},
						},
					},
				},
			},
			"methods": {
				Type: "array",
				Items: &generator.Schema{
					Type:    "string",
					Pattern: "^[A-Z]+$",
				},
			},
			"hosts": {
				Type:  "array",
				Items: typedefs.Host,
			},
			"paths": {
				Type:  "array",
				Items: typedefs.Path,
			},
			"headers": {
				Type:                 "object",
				AdditionalProperties: false,
				PatternProperties: map[string]*generator.Schema{
					"^[Hh][Oo][Ss][Tt]$": {
						Not: &generator.Schema{
							Description: "must not contain 'host' header",
						},
					},
					"^[A-Za-z0-9!#$%&'*+-.^_|~]+$": {
						Type: "object",
						Properties: map[string]*generator.Schema{
							"values": {
								Type: "array",
								Items: &generator.Schema{
									Type: "string",
								},
							},
						},
					},
				},
			},
			"https_redirect_status_code": {
				Type: "integer",
				Enum: []interface{}{
					http.StatusUpgradeRequired,
					http.StatusMovedPermanently,
					http.StatusFound,
					http.StatusTemporaryRedirect,
					http.StatusPermanentRedirect,
				},
			},
			"regex_priority": {
				Type:             "integer",
				ExclusiveMinimum: -1,
			},
			"strip_path": {
				Type: "boolean",
			},
			"path_handling": {
				Type: "string",
				Enum: []interface{}{
					"v0",
					"v1",
				},
			},
			"preserve_host": {
				Type: "boolean",
			},
			"request_buffering": {
				Type: "boolean",
			},
			"response_buffering": {
				Type: "boolean",
			},
			"snis": {
				Type:  "array",
				Items: typedefs.Host,
			},
			"sources": {
				Type:  "array",
				Items: typedefs.CIDRPort,
			},
			"destinations": {
				Type:  "array",
				Items: typedefs.CIDRPort,
			},
			// TODO "service": find a way to reference
			"tags": typedefs.Tags,
		},
		AdditionalProperties: false,
		Required: []string{
			"id",
			"protocols",
		},
		AllOf: []*generator.Schema{
			{
				Description: "'snis' can be set only when protocols has one of" +
					" 'https', 'grpcs' or 'tls'",
				If: &generator.Schema{
					Required: []string{"snis"},
				},
				Then: &generator.Schema{
					Properties: map[string]*generator.Schema{
						"protocols": {
							Contains: &generator.Schema{
								OneOf: []*generator.Schema{
									{
										Type:  "string",
										Const: typedefs.ProtocolHTTPS,
									},
									{
										Type:  "string",
										Const: typedefs.ProtocolGRPCS,
									},
									{
										Type:  "string",
										Const: typedefs.ProtocolTLS,
									},
								},
							},
						},
					},
				},
			},
			{
				Description: "when protocols has 'http' or 'https', " +
					"'sources' or 'destinations' cannot be set",
				If: &generator.Schema{
					Required: []string{"protocols"},
					Properties: map[string]*generator.Schema{
						"protocols": {
							Contains: &generator.Schema{
								AnyOf: []*generator.Schema{
									{
										Type:  "string",
										Const: typedefs.ProtocolHTTPS,
									},
									{
										Type:  "string",
										Const: typedefs.ProtocolHTTP,
									},
								},
							},
						},
					},
				},
				Then: &generator.Schema{
					Properties: map[string]*generator.Schema{
						"sources": {
							Not: &generator.Schema{
								Description: "when protocols has 'http' or" +
									" 'https', 'sources' or" +
									" 'destination' cannot be set",
							},
						},
						"destinations": {Not: &generator.Schema{}},
					},
				},
			},
			{
				Description: "when protocols has 'http', at least one of 'hosts'," +
					" 'methods', 'paths' or 'headers' must be set",
				If: &generator.Schema{
					Required: []string{"protocols"},
					Properties: map[string]*generator.Schema{
						"protocols": {
							Contains: &generator.Schema{
								Const: typedefs.ProtocolHTTP,
							},
						},
					},
				},
				Then: &generator.Schema{
					AnyOf: []*generator.Schema{
						{
							Required: []string{"methods"},
						},
						{
							Required: []string{"hosts"},
						},
						{
							Required: []string{"paths"},
						},
						{
							Required: []string{"paths"},
						},
						{
							Required: []string{"headers"},
						},
					},
				},
			},
			{
				Description: "when protocols has 'https', at least one of 'snis'," +
					" 'hosts', 'methods', 'paths' or 'headers' must be set",
				If: &generator.Schema{
					Required: []string{"protocols"},
					Properties: map[string]*generator.Schema{
						"protocols": {
							Contains: &generator.Schema{
								Const: typedefs.ProtocolHTTPS,
							},
						},
					},
				},
				Then: &generator.Schema{
					AnyOf: []*generator.Schema{
						{
							Required: []string{"methods"},
						},
						{
							Required: []string{"hosts"},
						},
						{
							Required: []string{"paths"},
						},
						{
							Required: []string{"paths"},
						},
						{
							Required: []string{"headers"},
						},
						{
							Required: []string{"snis"},
						},
					},
				},
			},
			{
				Description: "when protocol has 'tcp', 'tls' or 'udp', " +
					"'methods', 'hosts', 'paths', 'headers' cannot be set",
				If: &generator.Schema{
					Required: []string{"protocols"},
					Properties: map[string]*generator.Schema{
						"protocols": {
							Contains: &generator.Schema{
								AnyOf: []*generator.Schema{
									{
										Type:  "string",
										Const: typedefs.ProtocolTCP,
									},
									{
										Type:  "string",
										Const: typedefs.ProtocolUDP,
									},
									{
										Type:  "string",
										Const: typedefs.ProtocolTLS,
									},
								},
							},
						},
					},
				},
				Then: &generator.Schema{
					Properties: map[string]*generator.Schema{
						"methods": {Not: &generator.Schema{}},
						"hosts":   {Not: &generator.Schema{}},
						"paths":   {Not: &generator.Schema{}},
						"headers": {Not: &generator.Schema{}},
					},
				},
			},
			{
				Description: "when protocols has 'tcp', 'tls' or 'udp', " +
					"then at least one of " +
					"'sources', 'destinations' or 'snis' must be set",
				If: &generator.Schema{
					Required: []string{"protocols"},
					Properties: map[string]*generator.Schema{
						"protocols": {
							Contains: &generator.Schema{
								AnyOf: []*generator.Schema{
									{
										Type:  "string",
										Const: typedefs.ProtocolTCP,
									},
									{
										Type:  "string",
										Const: typedefs.ProtocolUDP,
									},
									{
										Type:  "string",
										Const: typedefs.ProtocolTLS,
									},
								},
							},
						},
					},
				},
				Then: &generator.Schema{
					AnyOf: []*generator.Schema{
						{
							Required: []string{"sources"},
						},
						{
							Required: []string{"destinations"},
						},
						{
							Required: []string{"snis"},
						},
					},
				},
			},
			{
				Description: "when protocol has 'grpc' or 'grpcs', 'strip_path', " +
					"'methods', 'sources', 'destinations' cannot be set",
				If: &generator.Schema{
					Required: []string{"protocols"},
					Properties: map[string]*generator.Schema{
						"protocols": {
							Contains: &generator.Schema{
								AnyOf: []*generator.Schema{
									{
										Type:  "string",
										Const: typedefs.ProtocolGRPC,
									},
									{
										Type:  "string",
										Const: typedefs.ProtocolGRPCS,
									},
								},
							},
						},
					},
				},
				Then: &generator.Schema{
					Properties: map[string]*generator.Schema{
						"strip_path":   {Not: &generator.Schema{Const: true}},
						"methods":      {Not: &generator.Schema{}},
						"sources":      {Not: &generator.Schema{}},
						"destinations": {Not: &generator.Schema{}},
					},
				},
			},
			{
				Description: "when protocols has 'grpc', at least one of 'hosts'," +
					" 'headers' or 'paths' must be set",
				If: &generator.Schema{
					Required: []string{"protocols"},
					Properties: map[string]*generator.Schema{
						"protocols": {
							Contains: &generator.Schema{
								Const: typedefs.ProtocolGRPC,
							},
						},
					},
				},
				Then: &generator.Schema{
					AnyOf: []*generator.Schema{
						{
							Required: []string{"hosts"},
						},
						{
							Required: []string{"headers"},
						},
						{
							Required: []string{"paths"},
						},
					},
				},
			},
			{
				Description: "when protocols has 'grpcs', " +
					"at least one of 'hosts', 'headers', 'paths' or 'snis' must be set",
				If: &generator.Schema{
					Required: []string{"protocols"},
					Properties: map[string]*generator.Schema{
						"protocols": {
							Contains: &generator.Schema{
								Const: typedefs.ProtocolGRPCS,
							},
						},
					},
				},
				Then: &generator.Schema{
					AnyOf: []*generator.Schema{
						{
							Required: []string{"hosts"},
						},
						{
							Required: []string{"headers"},
						},
						{
							Required: []string{"paths"},
						},
						{
							Required: []string{"snis"},
						},
					},
				},
			},
		},
	}
	err = generator.Register(string(TypeRoute), routeSchema)
	if err != nil {
		panic(err)
	}
}
