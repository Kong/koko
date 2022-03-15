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
	defaultTimeout = 60000
	defaultRetries = 5
	defaultPort    = 80
	maxRetries     = 32767
	maxVerifyDepth = 64
	// TypeService denotes the Service type.
	TypeService = model.Type("service")
)

var (
	defaultService = &v1.Service{
		Protocol:       "http",
		Port:           defaultPort,
		Retries:        defaultRetries,
		ConnectTimeout: defaultTimeout,
		ReadTimeout:    defaultTimeout,
		WriteTimeout:   defaultTimeout,
		Enabled:        wrapperspb.Bool(true),
	}
	_ model.Object = Service{}
)

func NewService() Service {
	return Service{
		Service: &v1.Service{},
	}
}

type Service struct {
	Service *v1.Service
}

func (r Service) ID() string {
	if r.Service == nil {
		return ""
	}
	return r.Service.Id
}

func (r Service) Type() model.Type {
	return TypeService
}

func (r Service) Resource() model.Resource {
	return r.Service
}

func (r Service) Validate() error {
	return validation.Validate(string(TypeService), r.Service)
}

func (r Service) ProcessDefaults() error {
	if r.Service == nil {
		return fmt.Errorf("invalid nil resource")
	}
	err := mergo.Merge(r.Service, defaultService,
		mergo.WithTransformers(wrappersPBTransformer{}))
	if err != nil {
		return err
	}
	defaultID(&r.Service.Id)
	return nil
}

func (r Service) Indexes() []model.Index {
	return []model.Index{
		{
			Name:      "name",
			Type:      model.IndexUnique,
			Value:     r.Service.Name,
			FieldName: "name",
		},
	}
}

func init() {
	err := model.RegisterType(TypeService, func() model.Object {
		return NewService()
	})
	if err != nil {
		panic(err)
	}

	serviceSchema := &generator.Schema{
		Type: "object",
		Properties: map[string]*generator.Schema{
			"id":   typedefs.ID,
			"name": typedefs.Name,
			"retries": {
				Type:    "integer",
				Minimum: intP(1),
				Maximum: maxRetries,
			},
			"protocol":        typedefs.Protocol,
			"host":            typedefs.Host,
			"port":            typedefs.Port,
			"path":            typedefs.Path,
			"connect_timeout": typedefs.Timeout,
			"read_timeout":    typedefs.Timeout,
			"write_timeout":   typedefs.Timeout,
			"tags":            typedefs.Tags,
			"tls_verify": {
				Type: "boolean",
			},
			"tls_verify_depth": {
				Type:    "integer",
				Minimum: intP(0),
				Maximum: maxVerifyDepth,
			},
			"ca_certificates": {
				Type:  "array",
				Items: typedefs.ID,
			},
			"created_at": typedefs.UnixEpoch,
			"updated_at": typedefs.UnixEpoch,
			"enabled": {
				Type: "boolean",
			},
		},
		AdditionalProperties: &falsy,
		Required: []string{
			"id",
			"protocol",
			"host",
			"port",
			"connect_timeout",
			"read_timeout",
			"write_timeout",
		},
		AllOf: []*generator.Schema{
			{
				Description: "tls_verify can be set only when protocol is" +
					" `https`",
				If: &generator.Schema{
					Required: []string{"tls_verify"},
					Properties: map[string]*generator.Schema{
						"tls_verify": {
							Const: true,
						},
					},
				},
				Then: &generator.Schema{
					Required: []string{"protocol"},
					Properties: map[string]*generator.Schema{
						"protocol": {
							Const: typedefs.ProtocolHTTPS,
						},
					},
				},
			},
			{
				Description: "tls_verify_depth can be set only when protocol" +
					" is `https`",
				If: &generator.Schema{
					Required: []string{"tls_verify_depth"},
				},
				Then: &generator.Schema{
					Required: []string{"protocol"},
					Properties: map[string]*generator.Schema{
						"protocol": {
							Const: typedefs.ProtocolHTTPS,
						},
					},
				},
			},
			{
				Description: "ca_certificates can be set only when protocol" +
					" is `https`",
				If: &generator.Schema{
					Required: []string{"ca_certificates"},
				},
				Then: &generator.Schema{
					Required: []string{"protocol"},
					Properties: map[string]*generator.Schema{
						"protocol": {
							Const: typedefs.ProtocolHTTPS,
						},
					},
				},
			},
			{
				Description: "path is required when protocol is http or https",
				If: &generator.Schema{
					Required: []string{"protocol"},
					Properties: map[string]*generator.Schema{
						"protocol": {
							OneOf: []*generator.Schema{
								{
									Const: typedefs.ProtocolHTTPS,
								},
								{
									Const: typedefs.ProtocolHTTP,
								},
							},
						},
					},
				},
				Then: &generator.Schema{
					Required: []string{"path"},
				},
			},
			{
				Description: "path can be set only when protocol is 'http' or" +
					" 'https'",
				If: &generator.Schema{
					Required: []string{"protocol"},
					Properties: map[string]*generator.Schema{
						"protocol": {
							OneOf: []*generator.Schema{
								{
									Const: typedefs.ProtocolGRPC,
								},
								{
									Const: typedefs.ProtocolGRPCS,
								},
								{
									Const: typedefs.ProtocolTCP,
								},
								{
									Const: typedefs.ProtocolTLS,
								},
								{
									Const: typedefs.ProtocolUDP,
								},
							},
						},
					},
				},
				Then: &generator.Schema{
					Properties: map[string]*generator.Schema{
						"path": {Not: &generator.Schema{}},
					},
				},
			},
		},
	}
	err = generator.Register(string(TypeService), serviceSchema)
	if err != nil {
		panic(err)
	}
}
