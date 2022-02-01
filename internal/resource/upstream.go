package resource

import (
	"fmt"

	"github.com/imdario/mergo"
	v1 "github.com/kong/koko/internal/gen/grpc/kong/admin/model/v1"
	"github.com/kong/koko/internal/model"
	"github.com/kong/koko/internal/model/json/generator"
	"github.com/kong/koko/internal/model/json/validation"
	"github.com/kong/koko/internal/model/json/validation/typedefs"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

const (
	TypeUpstream model.Type = "upstream"

	cookieNamePattern = "^[a-zA-Z0-9-_]+$"
	maxSlots          = 1 << 16
	minSlots          = 10
	maxConcurrency    = 1 << 31
	maxSeconds        = 65535
	maxStatuses       = 32
	minStatus         = 100
	maxStatus         = 999
	maxOneByteInt     = 255
	maxThreshold      = 100

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
		},
	}
	typedefSeconds = &generator.Schema{
		Type:    "integer",
		Minimum: 0,
		Maximum: maxSeconds,
	}
	typedefOneByteInteger = &generator.Schema{
		Type:    "integer",
		Minimum: 0,
		Maximum: maxOneByteInt,
	}
	typedefHTTPStatuses = &generator.Schema{
		Type: "array",
		Items: &generator.Schema{
			Type:    "integer",
			Minimum: minStatus,
			Maximum: maxStatus,
		},
		MaxItems: maxStatuses,
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

func (r Upstream) Validate() error {
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
	return nil
}

func (r Upstream) ProcessDefaults() error {
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
	return []model.Index{
		{
			Name:      "name",
			Type:      model.IndexUnique,
			Value:     r.Upstream.Name,
			FieldName: "name",
		},
	}
}

func init() {
	err := model.RegisterType(TypeUpstream, func() model.Object {
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
				Pattern: cookieNamePattern,
			},
			"hash_on_cookie_path": typedefs.Path,
			"slots": {
				Type:    "integer",
				Minimum: minSlots,
				Maximum: maxSlots,
			},
			"host_header": typedefs.Header,
			"healthchecks": {
				Type: "object",
				Properties: map[string]*generator.Schema{
					"threshold": {
						Type:    "number",
						Minimum: 0,
						Maximum: maxThreshold,
					},
					"active": {
						Type: "object",
						Properties: map[string]*generator.Schema{
							"concurrency": {
								Type:    "integer",
								Minimum: 1,
								Maximum: maxConcurrency,
							},
							"http_sni":  {},
							"http_path": typedefs.Path,
							"https_verify_certificate": {
								Type: "boolean",
							},
							"type": {
								Type: "string",
								Enum: []interface{}{
									"tcp",
									"http",
									"https",
								},
							},
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
							"type": {
								Type: "string",
								Enum: []interface{}{
									"tcp",
									"http",
								},
							},
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
					"'header', 'cookie'",
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
							},
						},
					},
				},
			},
			{
				Description: "when 'hash_on' is set to 'ip', " +
					"'hash_fallback' must be set to one of 'none', 'consumer', " +
					"'header', 'cookie'",
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
							},
						},
					},
				},
			},
		},
	}
	err = generator.Register(string(TypeUpstream), upstreamSchema)
	if err != nil {
		panic(err)
	}
}