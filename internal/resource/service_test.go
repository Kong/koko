package resource

import (
	"testing"

	"github.com/google/uuid"
	model "github.com/kong/koko/internal/gen/grpc/kong/admin/model/v1"
	"github.com/kong/koko/internal/model/json/validation"
	"github.com/stretchr/testify/assert"
)

func TestNewService(t *testing.T) {
	s := NewService()
	assert.NotNil(t, s)
	assert.NotNil(t, s.Service)
}

func TestService_ID(t *testing.T) {
	var s Service
	id := s.ID()
	assert.Empty(t, id)
	s = NewService()
	id = s.ID()
	assert.Empty(t, id)
}

func TestService_Type(t *testing.T) {
	assert.Equal(t, TypeService, NewService().Type())
}

func TestService_ProcessDefaults(t *testing.T) {
	t.Run("defaults are correctly injected", func(t *testing.T) {
		r := NewService()
		err := r.ProcessDefaults()
		assert.Nil(t, err)
		assert.True(t, validUUID(r.ID()))
		// empty out the id for equality comparison
		r.Service.Id = ""
		r.Service.CreatedAt = 0
		r.Service.UpdatedAt = 0
		assert.Equal(t, r.Resource(), defaultService)
	})
	t.Run("defaults do not override explicit values", func(t *testing.T) {
		r := NewService()
		r.Service.ConnectTimeout = 42
		r.Service.Port = 4242
		r.Service.Retries = 1
		r.Service.Protocol = "grpc"
		err := r.ProcessDefaults()
		assert.Nil(t, err)
		assert.True(t, validUUID(r.ID()))
		assert.NotEmpty(t, r.Service.CreatedAt)
		assert.NotEmpty(t, r.Service.UpdatedAt)
		// empty out the id and ts for equality comparison
		r.Service.Id = ""
		r.Service.CreatedAt = 0
		r.Service.UpdatedAt = 0
		assert.Equal(t, &model.Service{
			Protocol:       "grpc",
			Port:           4242,
			Retries:        1,
			ConnectTimeout: 42,
			ReadTimeout:    defaultTimeout,
			WriteTimeout:   defaultTimeout,
		}, r.Resource())
	})
}

func TestValidate(t *testing.T) {
	s := model.Service{
		Id:             uuid.NewString(),
		Name:           "foo",
		Retries:        1<<15 - 1,
		Tags:           []string{"foo"},
		Protocol:       "http",
		Host:           "foo.com",
		Port:           80,
		Path:           "/sf/bar/foo",
		ConnectTimeout: 3,
		ReadTimeout:    3,
		WriteTimeout:   3,
	}

	svc := &Service{Service: &s}
	err := svc.ValidateCompat()
	assert.Nil(t, err)
}

func goodService() Service {
	s := NewService()
	_ = s.ProcessDefaults()
	s.Service.Host = "good.example.com"
	s.Service.Path = "/good-path"
	return s
}

func TestService_Validate(t *testing.T) {
	tests := []struct {
		name    string
		Service func() Service
		wantErr bool
		Errs    []*model.ErrorDetail
	}{
		{
			name: "empty service throws an error",
			Service: func() Service {
				return NewService()
			},
			wantErr: true,
			Errs: []*model.ErrorDetail{
				{
					Type: model.ErrorType_ERROR_TYPE_ENTITY,
					Messages: []string{
						"missing properties: 'id', 'protocol', 'host', " +
							"'port', 'connect_timeout', 'read_timeout', 'write_timeout'",
					},
				},
			},
		},
		{
			name: "default service throws an error",
			Service: func() Service {
				s := NewService()
				_ = s.ProcessDefaults()
				return s
			},
			wantErr: true,
			Errs: []*model.ErrorDetail{
				{
					Type: model.ErrorType_ERROR_TYPE_ENTITY,
					Messages: []string{
						"missing properties: 'host'",
						"path is required when protocol is http or https",
					},
				},
			},
		},
		{
			name: "invalid timeout throws an error",
			Service: func() Service {
				s := goodService()
				s.Service.ReadTimeout = -1
				return s
			},
			wantErr: true,
			Errs: []*model.ErrorDetail{
				{
					Type:  model.ErrorType_ERROR_TYPE_FIELD,
					Field: "read_timeout",
					Messages: []string{
						"must be >= 1 but found -1",
					},
				},
			},
		},
		{
			name: "invalid ID throws an error",
			Service: func() Service {
				s := goodService()
				s.Service.Id = "bad-uuid"
				return s
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
			Service: func() Service {
				s := goodService()
				s.Service.Name = "%foo"
				return s
			},
			wantErr: true,
			Errs: []*model.ErrorDetail{
				{
					Type:  model.ErrorType_ERROR_TYPE_FIELD,
					Field: "name",
					Messages: []string{
						"must match pattern '^[0-9a-zA-Z.-_~]*$'",
					},
				},
			},
		},
		{
			name: "invalid host throws an error",
			Service: func() Service {
				s := goodService()
				s.Service.Host = "%foo"
				return s
			},
			wantErr: true,
			Errs: []*model.ErrorDetail{
				{
					Type:  model.ErrorType_ERROR_TYPE_FIELD,
					Field: "host",
					Messages: []string{
						"must be a valid hostname",
					},
				},
			},
		},
		{
			name: "invalid tags throws an error",
			Service: func() Service {
				s := goodService()
				s.Service.Tags = []string{"$tag"}
				return s
			},
			wantErr: true,
			Errs: []*model.ErrorDetail{
				{
					Type:  model.ErrorType_ERROR_TYPE_FIELD,
					Field: "tags[0]",
					Messages: []string{
						"must match pattern '^[0-9a-zA-Z.-_~]*$'",
					},
				},
			},
		},
		{
			name: "more than 8 tags throw an error",
			Service: func() Service {
				s := goodService()
				s.Service.Tags = []string{
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
				return s
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
			Service: func() Service {
				s := goodService()
				s.Service.
					Name = "anyservicewithareallylongnameisnotveryhelpful" + "" +
					"toanyoneatallisitifyouthinkitisthistestisgoingtoprove" +
					"youwrongandifitdoesnottheniguesswearedoomed"
				return s
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
			Service: func() Service {
				s := goodService()
				s.Service.Protocol = "smtp"
				return s
			},
			wantErr: true,
			Errs: []*model.ErrorDetail{
				{
					Type:  model.ErrorType_ERROR_TYPE_FIELD,
					Field: "protocol",
					Messages: []string{
						`value must be one of "http", "https", "grpc", ` +
							`"grpcs", "tcp", "udp", "tls"`,
					},
				},
			},
		},
		{
			name: "invalid retries throws an error",
			Service: func() Service {
				s := goodService()
				s.Service.Retries = -1
				return s
			},
			wantErr: true,
			Errs: []*model.ErrorDetail{
				{
					Type:  model.ErrorType_ERROR_TYPE_FIELD,
					Field: "retries",
					Messages: []string{
						"must be >= 1 but found -1",
					},
				},
			},
		},
		{
			name: "invalid port throws an error",
			Service: func() Service {
				s := goodService()
				s.Service.Port = 69420
				return s
			},
			wantErr: true,
			Errs: []*model.ErrorDetail{
				{
					Type:  model.ErrorType_ERROR_TYPE_FIELD,
					Field: "port",
					Messages: []string{
						"must be <= 65535 but found 69420",
					},
				},
			},
		},
		{
			name: "path must begin with /",
			Service: func() Service {
				s := goodService()
				s.Service.Path = "foo"
				return s
			},
			wantErr: true,
			Errs: []*model.ErrorDetail{
				{
					Type:  model.ErrorType_ERROR_TYPE_FIELD,
					Field: "path",
					Messages: []string{
						"must begin with `/`",
					},
				},
			},
		},
		{
			name: "path must not contain '//'",
			Service: func() Service {
				s := goodService()
				s.Service.Path = "/foo//bar"
				return s
			},
			wantErr: true,
			Errs: []*model.ErrorDetail{
				{
					Type:  model.ErrorType_ERROR_TYPE_FIELD,
					Field: "path",
					Messages: []string{
						"must not contain `//`",
					},
				},
			},
		},
		{
			name: "ca_certificates must be ID",
			Service: func() Service {
				s := goodService()
				s.Service.CaCertificates = []string{"foo"}
				return s
			},
			wantErr: true,
			Errs: []*model.ErrorDetail{
				{
					Type:  model.ErrorType_ERROR_TYPE_FIELD,
					Field: "ca_certificates[0]",
					Messages: []string{
						"must be a valid UUID",
					},
				},
				{
					Type: model.ErrorType_ERROR_TYPE_ENTITY,
					Messages: []string{
						"ca_certificates are not yet implemented",
					},
				},
			},
		},
		{
			name: "tls properties cannot be set when protocol is not https",
			Service: func() Service {
				s := goodService()
				s.Service.TlsVerify = true
				s.Service.TlsVerifyDepth = 1
				s.Service.Protocol = "http"
				return s
			},
			wantErr: true,
			Errs: []*model.ErrorDetail{
				{
					Type: model.ErrorType_ERROR_TYPE_ENTITY,
					Messages: []string{
						"tls_verify can be set only when protocol is `https`",
						"tls_verify_depth can be set only when protocol is" +
							" `https`",
					},
				},
			},
		},
		{
			name: "tls properties can be set when protocol is  https",
			Service: func() Service {
				s := goodService()
				s.Service.TlsVerify = true
				s.Service.TlsVerifyDepth = 1
				s.Service.Protocol = "https"
				return s
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.Service().Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.Errs != nil {
				verr, _ := err.(validation.Error)
				assert.ElementsMatch(t, tt.Errs, verr.Errs)
			}
		})
	}
}
