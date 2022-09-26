package resource

import (
	"context"
	"fmt"

	"github.com/imdario/mergo"
	v1 "github.com/kong/koko/internal/gen/grpc/kong/admin/model/v1"
	"github.com/kong/koko/internal/model"
	"github.com/kong/koko/internal/model/json/extension"
	"github.com/kong/koko/internal/model/json/generator"
	"github.com/kong/koko/internal/model/json/validation"
	"github.com/kong/koko/internal/model/json/validation/typedefs"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

const (
	TypeUpstream model.Type = "upstream"

	hashFieldPattern = "^[a-zA-Z0-9-_]+$"
	maxSlots         = 1 << 16
	minSlots         = 10
	maxConcurrency   = 1 << 31
	maxSeconds       = 65535
	maxStatuses      = 32
	minStatus        = 100
	maxStatus        = 999
	maxOneByteInt    = 255
	maxThreshold     = 100

	defaultSlots       = 10000
	defaultConcurrency = 10
)

var (
	_               model.Object = Upstream{}
	defaultUpstream              = &v1.Upstream{
		Algorithm:        "round-robin",
		Slots:            wrapperspb.Int32(defaultSlots),
		HashOn:           "none",
		HashFallback:     "none",
		HashOnCookiePath: "/",
		Healthchecks: &v1.Healthchecks{
			Threshold: wrapperspb.Float(0),
			Active: &v1.ActiveHealthcheck{
				Concurrency: wrapperspb.Int32(defaultConcurrency),
				Healthy: &v1.ActiveHealthyCondition{
					HttpStatuses: []int32{200, 302},
					Interval:     wrapperspb.Int32(0),
					Successes:    wrapperspb.Int32(0),
				},
				HttpPath:               "/",
				HttpsVerifyCertificate: wrapperspb.Bool(true),
				Type:                   "http",
				Timeout:                wrapperspb.Int32(1),
				Unhealthy: &v1.ActiveUnhealthyCondition{
					HttpFailures: wrapperspb.Int32(0),
					TcpFailures:  wrapperspb.Int32(0),
					HttpStatuses: []int32{429, 404, 500, 501, 502, 503, 504, 505},
					Timeouts:     wrapperspb.Int32(0),
					Interval:     wrapperspb.Int32(0),
				},
			},
			Passive: &v1.PassiveHealthcheck{
				Healthy: &v1.PassiveHealthyCondition{
					HttpStatuses: []int32{
						200, 201, 202, 203, 204, 205, 206, 207, 208, 226,
						300, 301, 302, 303, 304, 305, 306, 307, 308,
					},
					Successes: wrapperspb.Int32(0),
				},
				Type: "http",
				Unhealthy: &v1.PassiveUnhealthyCondition{
					HttpFailures: wrapperspb.Int32(0),
					TcpFailures:  wrapperspb.Int32(0),
					HttpStatuses: []int32{429, 500, 503},
					Timeouts:     wrapperspb.Int32(0),
				},
			},
		},
	}

	typedefHashOn = &generator.Schema{
		Type: "string",
		Enum: []interface{}{
			"none",
			"consumer",
			"ip",
			"header",
			"cookie",
			// 3.0+
			"path",
			"query_arg",
			"uri_capture",
		},
	}
	typedefSeconds = &generator.Schema{
		Type:    "integer",
		Minimum: intP(0),
		Maximum: maxSeconds,
	}
	typedefOneByteInteger = &generator.Schema{
		Type:    "integer",
		Minimum: intP(0),
		Maximum: maxOneByteInt,
	}
	typedefHTTPStatuses = &generator.Schema{
		Type: "array",
		Items: &generator.Schema{
			Type:    "integer",
			Minimum: intP(minStatus),
			Maximum: maxStatus,
		},
		MaxItems: maxStatuses,
	}
	typedefHealthCheckTypes = &generator.Schema{
		Type: "string",
		Enum: []interface{}{
			"tcp",
			"http",
			"https",
			"grpc",
			"grpcs",
		},
	}
)

func NewUpstream() Upstream {
	return Upstream{
		Upstream: &v1.Upstream{},
	}
}

type Upstream struct {
	Upstream *v1.Upstream
}

func (r Upstream) ID() string {
	if r.Upstream == nil {
		return ""
	}
	return r.Upstream.Id
}

func (r Upstream) Type() model.Type {
	return TypeUpstream
}

func (r Upstream) Resource() model.Resource {
	return r.Upstream
}

// SetResource implements the Object.SetResource interface.
func (r Upstream) SetResource(pr model.Resource) error { return model.SetResource(r, pr) }

func (r Upstream) Validate(ctx context.Context) error {
	err := validation.Validate(string(TypeUpstream), r.Upstream)
	if err != nil {
		return err
	}
	// not possible to check via json-schema
	if r.Upstream.HashOnHeader != "" || r.Upstream.HashFallbackHeader != "" {
		if r.Upstream.HashOnHeader == r.Upstream.HashFallbackHeader {
			return validation.Error{
				Errs: []*v1.ErrorDetail{
					{
						Type: v1.ErrorType_ERROR_TYPE_ENTITY,
						Messages: []string{
							"'hash_fallback_header' must not be" +
								" equal to 'hash_on_header'",
						},
					},
				},
			}
		}
	}
	if r.Upstream.HashOnQueryArg != "" || r.Upstream.HashFallbackQueryArg != "" {
		if r.Upstream.HashOnQueryArg == r.Upstream.HashFallbackQueryArg {
			return validation.Error{
				Errs: []*v1.ErrorDetail{
					{
						Type: v1.ErrorType_ERROR_TYPE_ENTITY,
						Messages: []string{
							"'hash_on_query_arg' must not be" +
								" equal to 'hash_fallback_query_arg'",
						},
					},
				},
			}
		}
	}
	if r.Upstream.HashOnUriCapture != "" || r.Upstream.HashFallbackUriCapture != "" {
		if r.Upstream.HashOnUriCapture == r.Upstream.HashFallbackUriCapture {
			return validation.Error{
				Errs: []*v1.ErrorDetail{
					{
						Type: v1.ErrorType_ERROR_TYPE_ENTITY,
						Messages: []string{
							"'hash_on_uri_capture' must not be" +
								" equal to 'hash_fallback_uri_capture'",
						},
					},
				},
			}
		}
	}
	return nil
}

func (r Upstream) ProcessDefaults(ctx context.Context) error {
	if r.Upstream == nil {
		return fmt.Errorf("invalid nil resource")
	}
	err := mergo.Merge(r.Upstream, defaultUpstream)
	if err != nil {
		return err
	}
	defaultID(&r.Upstream.Id)
	return nil
}

func (r Upstream) Indexes() []model.Index {
	indexes := []model.Index{{
		Name:      "name",
		Type:      model.IndexUnique,
		Value:     r.Upstream.Name,
		FieldName: "name",
	}}
	if r.Upstream.ClientCertificate != nil {
		indexes = append(indexes, model.Index{
			Name:        "client_certificate_id",
			Type:        model.IndexForeign,
			ForeignType: TypeCertificate,
			FieldName:   "client_certificate.id",
			Value:       r.Upstream.ClientCertificate.Id,
		})
	}
	return indexes
}

func init() {
	err := model.RegisterType(TypeUpstream, &v1.Upstream{}, func() model.Object {
		return NewUpstream()
	})
	if err != nil {
		panic(err)
	}

	upstreamSchema := &generator.Schema{
		Type: "object",
		Properties: map[string]*generator.Schema{
			"id":         typedefs.ID,
			"name":       typedefs.Name,
			"tags":       typedefs.Tags,
			"created_at": typedefs.UnixEpoch,
			"updated_at": typedefs.UnixEpoch,
			"algorithm": {
				Type: "string",
				Enum: []interface{}{
					"round-robin",
					"consistent-hashing",
					"least-connections",
				},
			},
			"hash_on":              typedefHashOn,
			"hash_fallback":        typedefHashOn,
			"hash_on_header":       typedefs.Header,
			"hash_fallback_header": typedefs.Header,
			"hash_on_cookie": {
				Type:    "string",
				Pattern: hashFieldPattern,
			},
			"hash_on_cookie_path": typedefs.Path,
			"hash_on_query_arg": {
				Type:    "string",
				Pattern: hashFieldPattern,
			},
			"hash_fallback_query_arg": {
				Type:    "string",
				Pattern: hashFieldPattern,
			},
			"hash_on_uri_capture": {
				Type:    "string",
				Pattern: hashFieldPattern,
			},
			"hash_fallback_uri_capture": {
				Type:    "string",
				Pattern: hashFieldPattern,
			},
			"slots": {
				Type:    "integer",
				Minimum: intP(minSlots),
				Maximum: maxSlots,
			},
			"host_header": typedefs.Host,
			"healthchecks": {
				Type: "object",
				Properties: map[string]*generator.Schema{
					"threshold": {
						Type:    "number",
						Minimum: intP(0),
						Maximum: maxThreshold,
					},
					"active": {
						Type: "object",
						Properties: map[string]*generator.Schema{
							"concurrency": {
								Type:    "integer",
								Minimum: intP(1),
								Maximum: maxConcurrency,
							},
							"https_sni": typedefs.Host,
							"http_path": typedefs.Path,
							"https_verify_certificate": {
								Type: "boolean",
							},
							"type":    typedefHealthCheckTypes,
							"timeout": typedefSeconds,
							"healthy": {
								Type: "object",
								Properties: map[string]*generator.Schema{
									"http_statuses": typedefHTTPStatuses,
									"interval":      typedefSeconds,
									"success":       typedefOneByteInteger,
								},
							},
							"unhealthy": {
								Type: "object",
								Properties: map[string]*generator.Schema{
									"http_failures": typedefOneByteInteger,
									"tcp_failures":  typedefOneByteInteger,
									"timeouts":      typedefOneByteInteger,
									"interval":      typedefSeconds,
									"http_statuses": typedefHTTPStatuses,
								},
							},
						},
					},
					"passive": {
						Type: "object",
						Properties: map[string]*generator.Schema{
							"type":    typedefHealthCheckTypes,
							"timeout": typedefSeconds,
							"healthy": {
								Type: "object",
								Properties: map[string]*generator.Schema{
									"http_statuses": typedefHTTPStatuses,
									"success":       typedefOneByteInteger,
								},
							},
							"unhealthy": {
								Type: "object",
								Properties: map[string]*generator.Schema{
									"http_failures": typedefOneByteInteger,
									"tcp_failures":  typedefOneByteInteger,
									"timeouts":      typedefOneByteInteger,
									"http_statuses": typedefHTTPStatuses,
								},
							},
						},
					},
				},
			},
			"client_certificate": typedefs.ReferenceObject,
		},
		AdditionalProperties: &falsy,
		Required: []string{
			"id",
			"name",
		},
		AllOf: []*generator.Schema{
			{
				Description: "when 'hash_on' is set to 'header'," +
					"'hash_on_header' must be set",
				If: &generator.Schema{
					Required: []string{"hash_on"},
					Properties: map[string]*generator.Schema{
						"hash_on": {
							Const: "header",
						},
					},
				},
				Then: &generator.Schema{
					Required: []string{"hash_on_header"},
				},
			},
			{
				Description: "when 'hash_fallback' is set to 'header'," +
					"'hash_fallback_header' must be set",
				If: &generator.Schema{
					Required: []string{"hash_fallback"},
					Properties: map[string]*generator.Schema{
						"hash_fallback": {
							Const: "header",
						},
					},
				},
				Then: &generator.Schema{
					Required: []string{"hash_fallback_header"},
				},
			},
			{
				Description: "when 'hash_on' is set to 'cookie', " +
					"'hash_on_cookie' must be set",
				If: &generator.Schema{
					Required: []string{"hash_on"},
					Properties: map[string]*generator.Schema{
						"hash_on": {
							Const: "cookie",
						},
					},
				},
				Then: &generator.Schema{
					Required: []string{"hash_on_cookie"},
				},
			},
			{
				Description: "when 'hash_fallback' is set to 'cookie', " +
					"'hash_on_cookie' must be set",
				If: &generator.Schema{
					Required: []string{"hash_fallback"},
					Properties: map[string]*generator.Schema{
						"hash_fallback": {
							Const: "cookie",
						},
					},
				},
				Then: &generator.Schema{
					Required: []string{"hash_on_cookie"},
				},
			},
			{
				Description: "when 'hash_on' is set to 'none', " +
					"'hash_fallback' must be set to 'none'",
				If: &generator.Schema{
					Required: []string{"hash_on"},
					Properties: map[string]*generator.Schema{
						"hash_on": {
							Const: "none",
						},
					},
				},
				Then: &generator.Schema{
					Required: []string{"hash_fallback"},
					Properties: map[string]*generator.Schema{
						"hash_fallback": {
							Const: "none",
						},
					},
				},
			},
			{
				Description: "when 'hash_on' is set to 'cookie', " +
					"'hash_fallback' must be set to 'none'",
				If: &generator.Schema{
					Required: []string{"hash_on"},
					Properties: map[string]*generator.Schema{
						"hash_on": {
							Const: "cookie",
						},
					},
				},
				Then: &generator.Schema{
					Required: []string{"hash_fallback"},
					Properties: map[string]*generator.Schema{
						"hash_fallback": {
							Const: "none",
						},
					},
				},
			},
			{
				Description: "when 'hash_on' is set to 'consumer', " +
					"'hash_fallback' must be set to one of 'none', 'ip', " +
					"'header', 'cookie', 'path', 'query_arg', 'uri_capture'",
				If: &generator.Schema{
					Required: []string{"hash_on"},
					Properties: map[string]*generator.Schema{
						"hash_on": {
							Const: "consumer",
						},
					},
				},
				Then: &generator.Schema{
					Required: []string{"hash_fallback"},
					Properties: map[string]*generator.Schema{
						"hash_fallback": {
							AnyOf: []*generator.Schema{
								{
									Type:  "string",
									Const: "none",
								},
								{
									Type:  "string",
									Const: "ip",
								},
								{
									Type:  "string",
									Const: "header",
								},
								{
									Type:  "string",
									Const: "cookie",
								},
								{
									Type:  "string",
									Const: "path",
								},
								{
									Type:  "string",
									Const: "query_arg",
								},
								{
									Type:  "string",
									Const: "uri_capture",
								},
							},
						},
					},
				},
			},
			{
				Description: "when 'hash_on' is set to 'ip', " +
					"'hash_fallback' must be set to one of 'none', 'consumer', " +
					"'header', 'cookie', 'path', 'query_arg', 'uri_capture'",
				If: &generator.Schema{
					Required: []string{"hash_on"},
					Properties: map[string]*generator.Schema{
						"hash_on": {
							Const: "ip",
						},
					},
				},
				Then: &generator.Schema{
					Required: []string{"hash_fallback"},
					Properties: map[string]*generator.Schema{
						"hash_fallback": {
							AnyOf: []*generator.Schema{
								{
									Type:  "string",
									Const: "none",
								},
								{
									Type:  "string",
									Const: "consumer",
								},
								{
									Type:  "string",
									Const: "header",
								},
								{
									Type:  "string",
									Const: "cookie",
								},
								{
									Type:  "string",
									Const: "path",
								},
								{
									Type:  "string",
									Const: "query_arg",
								},
								{
									Type:  "string",
									Const: "uri_capture",
								},
							},
						},
					},
				},
			},
			{
				Description: "when 'hash_on' is set to 'path', " +
					"'hash_fallback' must be set to one of 'none', 'consumer', 'ip', " +
					"'header', 'cookie', 'query_arg', 'uri_capture'",
				If: &generator.Schema{
					Required: []string{"hash_on"},
					Properties: map[string]*generator.Schema{
						"hash_on": {
							Const: "path",
						},
					},
				},
				Then: &generator.Schema{
					Required: []string{"hash_fallback"},
					Properties: map[string]*generator.Schema{
						"hash_fallback": {
							AnyOf: []*generator.Schema{
								{
									Type:  "string",
									Const: "none",
								},
								{
									Type:  "string",
									Const: "consumer",
								},
								{
									Type:  "string",
									Const: "header",
								},
								{
									Type:  "string",
									Const: "cookie",
								},
								{
									Type:  "string",
									Const: "ip",
								},
								{
									Type:  "string",
									Const: "query_arg",
								},
								{
									Type:  "string",
									Const: "uri_capture",
								},
							},
						},
					},
				},
			},
			{
				Description: "when 'hash_on' is set to 'query_arg', " +
					"'hash_on_query_arg' must be set",
				If: &generator.Schema{
					Required: []string{"hash_on"},
					Properties: map[string]*generator.Schema{
						"hash_on": {
							Const: "query_arg",
						},
					},
				},
				Then: &generator.Schema{
					Required: []string{"hash_on_query_arg"},
				},
			},
			{
				Description: "when 'hash_fallback' is set to 'query_arg', " +
					"'hash_fallback_query_arg' must be set",
				If: &generator.Schema{
					Required: []string{"hash_fallback"},
					Properties: map[string]*generator.Schema{
						"hash_fallback": {
							Const: "query_arg",
						},
					},
				},
				Then: &generator.Schema{
					Required: []string{"hash_fallback_query_arg"},
				},
			},
			{
				Description: "when 'hash_on' is set to 'uri_capture', " +
					"'hash_on_uri_capture' must be set",
				If: &generator.Schema{
					Required: []string{"hash_on"},
					Properties: map[string]*generator.Schema{
						"hash_on": {
							Const: "uri_capture",
						},
					},
				},
				Then: &generator.Schema{
					Required: []string{"hash_on_uri_capture"},
				},
			},
			{
				Description: "when 'hash_fallback' is set to 'uri_capture', " +
					"'hash_fallback_uri_capture' must be set",
				If: &generator.Schema{
					Required: []string{"hash_fallback"},
					Properties: map[string]*generator.Schema{
						"hash_fallback": {
							Const: "uri_capture",
						},
					},
				},
				Then: &generator.Schema{
					Required: []string{"hash_fallback_uri_capture"},
				},
			},
		},
		XKokoConfig: &extension.Config{
			ResourceAPIPath: "upstreams",
		},
	}
	err = generator.DefaultRegistry.Register(string(TypeUpstream), upstreamSchema)
	if err != nil {
		panic(err)
	}
}
