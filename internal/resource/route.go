package resource

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/gofrs/uuid"
	"github.com/imdario/mergo"
	goatcrouter "github.com/kong/go-atc-router"
	v1 "github.com/kong/koko/internal/gen/grpc/kong/admin/model/v1"
	"github.com/kong/koko/internal/model"
	"github.com/kong/koko/internal/model/json/extension"
	"github.com/kong/koko/internal/model/json/generator"
	"github.com/kong/koko/internal/model/json/validation"
	"github.com/kong/koko/internal/model/json/validation/typedefs"
	"github.com/samber/lo"
	"github.com/santhosh-tekuri/jsonschema/v5"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

const (
	// TypeRoute denotes the Route type.
	TypeRoute model.Type = "route"

	maxMatchElements     = 16
	maxHeaderValueLength = 64

	// RouteSNIRuleTitle denotes the name of the schema rule to apply
	// when using SNIs.
	RouteSNIRuleTitle = "sni_rule"
	// WSProtocolsRuleTitle denotes the name of the schema rule to apply
	// to ws protocols.
	WSProtocolsRuleTitle = "ws_protocols_rule"
	// ExpressionRouteRuleTitle denotes the name of the schema rule to apply
	// to expression-based routes.
	ExpressionRouteRuleTitle = "atc_route_rule"
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

// SetResource implements the Object.SetResource interface.
func (r Route) SetResource(pr model.Resource) error { return model.SetResource(r, pr) }

func (r Route) Indexes() []model.Index {
	res := make([]model.Index, 0)
	if r.Route.Name != "" {
		res = append(res, model.Index{
			Name:      "name",
			Type:      model.IndexUnique,
			Value:     r.Route.Name,
			FieldName: "name",
		})
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

func (r Route) Validate(ctx context.Context) error {
	for _, sni := range r.Route.Snis {
		if hostnameCheck(sni) != typeName {
			errWrap := validation.Error{}
			errWrap.Errs = append(errWrap.Errs, &v1.ErrorDetail{
				Type:  v1.ErrorType_ERROR_TYPE_FIELD,
				Field: "snis",
				Messages: []string{
					fmt.Sprintf("must not be an IP: '%s'", sni),
				},
			})
			return errWrap
		}
		if len(strings.Split(sni, ":")) > 1 {
			errWrap := validation.Error{}
			errWrap.Errs = append(errWrap.Errs, &v1.ErrorDetail{
				Type:  v1.ErrorType_ERROR_TYPE_FIELD,
				Field: "snis",
				Messages: []string{
					fmt.Sprintf("must not contain a port: '%s'", sni),
				},
			})
			return errWrap
		}
	}
	return validation.Validate(string(TypeRoute), r.Route)
}

func (r Route) ProcessDefaults(ctx context.Context) error {
	if r.Route == nil {
		return fmt.Errorf("invalid nil resource")
	}
	// When protocols has 'grpc' or 'grpcs', 'strip_path' cannot be set.
	// If 'strip_path' is null, the default value (true) would be applied,
	// so we set it to 'false' in order to prevent that.
	if lo.Some(r.Route.Protocols, []string{typedefs.ProtocolGRPC, typedefs.ProtocolGRPCS}) {
		if r.Route.StripPath == nil {
			r.Route.StripPath = wrapperspb.Bool(false)
		}
	}
	err := mergo.Merge(r.Route, defaultRoute,
		mergo.WithTransformers(wrappersPBTransformer{}))
	if err != nil {
		return err
	}
	defaultID(&r.Route.Id)
	return nil
}

func buildSchema() *goatcrouter.Schema {
	schema := goatcrouter.NewSchema()

	for _, fieldname := range []string{
		"net.protocol", "tls.sni",
		"http.method", "http.host",
		"http.path", "http.raw_path",
		"http.headers.*",
	} {
		schema.AddField(fieldname, goatcrouter.String)
	}

	for _, fieldname := range []string{"net.port"} {
		schema.AddField(fieldname, goatcrouter.Int)
	}

	return schema
}

var cachedSchema = buildSchema()

func init() {
	jsonschema.Formats["route-expression"] = func(v interface{}) bool {
		expression, ok := v.(string)
		if !ok {
			return false
		}

		router := goatcrouter.NewRouter(cachedSchema)
		defer router.Free()
		err := router.AddMatcher(0, uuid.Must(uuid.NewV4()), expression)
		return err == nil
	}

	err := model.RegisterType(TypeRoute, &v1.Route{}, func() model.Object {
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
				Items: typedefs.AllProtocols,
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
					{
						Description: "must contain only one subset [ tls_passthrough ]",
						Items: &generator.Schema{
							Type: "string",
							Enum: []interface{}{
								typedefs.ProtocolTLSPassthrough,
							},
						},
					},
					{
						Description: "must contain only one subset [ ws wss ]",
						Items: &generator.Schema{
							Type: "string",
							Enum: []interface{}{
								typedefs.ProtocolWS,
								typedefs.ProtocolWSS,
							},
						},
					},
				},
				Default: []interface{}{
					typedefs.ProtocolHTTP,
					typedefs.ProtocolHTTPS,
				},
			},
			"methods": {
				Type: "array",
				Items: &generator.Schema{
					Type:     "string",
					Pattern:  "^[A-Z]+$",
					MaxItems: maxMatchElements,
				},
			},
			"hosts": {
				Type:     "array",
				Items:    typedefs.WilcardHost,
				MaxItems: maxMatchElements,
			},
			"paths": {
				Type:     "array",
				Items:    typedefs.RouterPath,
				MaxItems: maxMatchElements,
			},
			"headers": {
				Type:                 "object",
				AdditionalProperties: &falsy,
				MaxProperties:        maxMatchElements,
				PatternProperties: map[string]*generator.Schema{
					"^[Hh][Oo][Ss][Tt]$": {
						Not: &generator.Schema{
							Description: "must not contain 'host' header",
						},
					},
					typedefs.HTTPHeaderNamePattern: {
						Type: "object",
						Properties: map[string]*generator.Schema{
							"values": {
								Type:     "array",
								MaxItems: maxMatchElements,
								Items: &generator.Schema{
									Type:      "string",
									MaxLength: maxHeaderValueLength,
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
				Default: http.StatusUpgradeRequired,
			},
			"regex_priority": {
				Type:             "integer",
				ExclusiveMinimum: -1,
				Default:          0,
			},
			"strip_path": {
				Type:    "boolean",
				Default: true,
			},
			"path_handling": {
				Type: "string",
				Enum: []interface{}{
					"v0",
					"v1",
				},
				Default: "v0",
			},
			"preserve_host": {
				Type:    "boolean",
				Default: false,
			},
			"request_buffering": {
				Type:    "boolean",
				Default: true,
			},
			"response_buffering": {
				Type:    "boolean",
				Default: true,
			},
			"snis": {
				Type:     "array",
				Items:    typedefs.Host,
				MaxItems: maxMatchElements,
			},
			"sources": {
				Type:     "array",
				Items:    typedefs.CIDRPort,
				MaxItems: maxMatchElements,
			},
			"destinations": {
				Type:     "array",
				Items:    typedefs.CIDRPort,
				MaxItems: maxMatchElements,
			},
			// TODO "service": find a way to reference
			"tags":       typedefs.Tags,
			"created_at": typedefs.UnixEpoch,
			"updated_at": typedefs.UnixEpoch,
			"service":    typedefs.ReferenceObject,
			"expression": {
				Type:   "string",
				Format: "route-expression",
			},
			"priority": {
				Type: "integer",
			},
		},
		AdditionalProperties: &falsy,
		Required: []string{
			"id",
			"protocols",
		},
		AllOf: []*generator.Schema{
			{
				Title: RouteSNIRuleTitle,
				Description: "'snis' can be set only when protocols has one of" +
					" 'https', 'grpcs', 'tls' or 'tls_passthrough'",
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
									{
										Type:  "string",
										Const: typedefs.ProtocolTLSPassthrough,
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
				Description: "when protocol has 'tcp', 'tls', 'tls_passthrough' or 'udp', " +
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
									{
										Type:  "string",
										Const: typedefs.ProtocolTLSPassthrough,
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
			{
				Description: "when protocols has 'tls_passthrough', " +
					"'snis' must be set",
				If: &generator.Schema{
					Required: []string{"protocols"},
					Properties: map[string]*generator.Schema{
						"protocols": {
							Contains: &generator.Schema{
								Const: typedefs.ProtocolTLSPassthrough,
							},
						},
					},
				},
				Then: &generator.Schema{
					AnyOf: []*generator.Schema{
						{
							Required: []string{"snis"},
						},
					},
				},
			},
			{
				Title: WSProtocolsRuleTitle,
				Description: "'ws' and 'wss' protocols are Kong Enterprise-only features. " +
					"Please upgrade to Kong Enterprise to use this feature.",
				Not: &generator.Schema{
					Required: []string{"protocols"},
					Properties: map[string]*generator.Schema{
						"protocols": {
							Contains: &generator.Schema{
								AnyOf: []*generator.Schema{
									{
										Type:  "string",
										Const: typedefs.ProtocolWS,
									},
									{
										Type:  "string",
										Const: typedefs.ProtocolWSS,
									},
								},
							},
						},
					},
				},
			},
			{
				Title:       ExpressionRouteRuleTitle,
				Description: "When 'expression' is defined, 'priority' is required.",
				If: &generator.Schema{
					Required: []string{"expression"},
				},
				Then: &generator.Schema{
					Required: []string{"priority"},
				},
			},
			{
				Title: ExpressionRouteRuleTitle,
				Description: "When 'expression' is defined, 'snis', 'sources' " +
					"or 'destinations' cannot be set.",
				If: &generator.Schema{
					Required: []string{"expression"},
				},
				Then: &generator.Schema{
					Properties: map[string]*generator.Schema{
						"snis":         {Not: &generator.Schema{}},
						"sources":      {Not: &generator.Schema{}},
						"destinations": {Not: &generator.Schema{}},
					},
				},
			},
		},
		XKokoConfig: &extension.Config{
			ResourceAPIPath: "routes",
		},
	}
	err = generator.DefaultRegistry.Register(string(TypeRoute), routeSchema)
	if err != nil {
		panic(err)
	}
}
