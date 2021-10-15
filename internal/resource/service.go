package resource

import (
	"fmt"

	ozzo "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/imdario/mergo"
	v1 "github.com/kong/koko/internal/gen/grpc/kong/admin/model/v1"
	"github.com/kong/koko/internal/model"
	"github.com/kong/koko/internal/model/json/generator"
	"github.com/kong/koko/internal/model/validation/typedefs"
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
	}
	_ model.Object = Service{}
)

func init() {
	err := model.RegisterType(TypeService, func() model.Object {
		return NewService()
	})
	if err != nil {
		panic(err)
	}
}

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
	panic("implement me")
}

func (r Service) ValidateCompat() error {
	if r.Service == nil {
		return fmt.Errorf("invalid nil resource")
	}
	s := r.Service
	err := ozzo.ValidateStruct(r.Service,
		ozzo.Field(&s.Id, typedefs.IDRules()...),
		ozzo.Field(&s.Name, typedefs.NameRule()...),
		ozzo.Field(&s.Retries, ozzo.Min(1), ozzo.Max(maxRetries)),
		ozzo.Field(&s.Protocol, typedefs.ProtocolRule()...),
		ozzo.Field(&s.Host, mergeRules(
			ozzo.Required,
			typedefs.HostRule(),
		)...,
		),
		ozzo.Field(&s.Port, mergeRules(
			ozzo.Required,
			typedefs.PortRule(),
		)...,
		),
		ozzo.Field(&s.Path, mergeRules(
			ozzo.Required,
			typedefs.PathRule(),
		)...,
		),
		ozzo.Field(&s.ConnectTimeout, mergeRules(
			ozzo.Required,
			typedefs.TimeoutRule(),
		)...,
		),
		ozzo.Field(&s.ReadTimeout, mergeRules(
			ozzo.Required,
			typedefs.TimeoutRule(),
		)...,
		),
		ozzo.Field(&s.WriteTimeout, mergeRules(
			ozzo.Required,
			typedefs.TimeoutRule(),
		)...,
		),
		ozzo.Field(&s.Tags, typedefs.TagsRule()...),
		ozzo.Field(&s.TlsVerifyDepth, ozzo.Min(0),
			ozzo.Max(maxVerifyDepth)),
		ozzo.Field(&s.CaCertificates, ozzo.Each(typedefs.UUID())),
		ozzo.Field(&s.TlsVerifyDepth, ozzo.Empty.When(s.
			Protocol != typedefs.ProtocolHTTPS).Error(
			"tls_verify_depth must be empty when protocol is not 'http'")),
		ozzo.Field(&s.TlsVerify, ozzo.Empty.When(s.
			Protocol != typedefs.ProtocolHTTPS).Error(
			"tls_verify must not be 'true' when protocol is not 'http'")),
		ozzo.Field(&s.CaCertificates, ozzo.Empty.When(s.
			Protocol != typedefs.ProtocolHTTPS).Error(
			"ca_certificates must not be set when protocol is not"+
				" 'https'")),
		ozzo.Field(&s.Path, ozzo.Empty.When(
			notHTTPProtocol(s.Protocol)).Error(
			"path must be empty when protocol is not 'http' or 'https'")),
		ozzo.Field(&s.CaCertificates, ozzo.Empty.Error(
			"ca certificates are not yet supported")),
	)
	if err != nil {
		return validationErr(err)
	}
	return nil
}

func (r Service) ProcessDefaults() error {
	if r.Service == nil {
		return fmt.Errorf("invalid nil resource")
	}
	err := mergo.Merge(r.Service, defaultService)
	if err != nil {
		return err
	}
	defaultID(&r.Service.Id)
	addTZ(r.Service)
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
	serviceSchema := &generator.Schema{
		Type: "object",
		Properties: map[string]*generator.Schema{
			"id":   typedefs.ID,
			"name": typedefs.Name,
			"retries": {
				Type:    "integer",
				Minimum: 1,
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
				Minimum: 0,
				Maximum: maxVerifyDepth,
			},
			"ca_certificates": {
				Type:  "array",
				Items: typedefs.ID,
			},
		},
		AdditionalProperties: false,
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
					" https",
				If: &generator.Schema{
					Properties: map[string]*generator.Schema{
						"tls_verify": {
							Const: true,
						},
					},
				},
				Then: &generator.Schema{
					Properties: map[string]*generator.Schema{
						"protocol": {
							Const: typedefs.ProtocolHTTPS,
						},
					},
				},
			},
			{
				// tls_verify_depth can be set only when protocol is https
				If: &generator.Schema{
					Required: []string{"tls_verify_depth"},
				},
				Then: &generator.Schema{
					Properties: map[string]*generator.Schema{
						"protocol": {
							Const: typedefs.ProtocolHTTPS,
						},
					},
				},
			},
			{
				// ca certificates can be set only when protocol is https
				If: &generator.Schema{
					Required: []string{"ca_certificates"},
				},
				Then: &generator.Schema{
					Properties: map[string]*generator.Schema{
						"protocol": {
							Const: typedefs.ProtocolHTTPS,
						},
					},
				},
			},
			{
				// path is required when protocol is http or https
				If: &generator.Schema{
					Properties: map[string]*generator.Schema{
						"protocol": {
							Enum: []interface{}{
								typedefs.ProtocolHTTP,
								typedefs.ProtocolHTTPS,
							},
						},
					},
				},
				Then: &generator.Schema{
					Required: []string{"path"},
				},
			},
			{
				// NYI
				Not: &generator.Schema{
					Required: []string{"ca_certificates"},
				},
			},
		},
	}
	err := generator.Register(string(TypeService), serviceSchema)
	if err != nil {
		panic(err)
	}
}
