package resource

import (
	"testing"

	"github.com/google/uuid"
	v1 "github.com/kong/koko/internal/gen/grpc/kong/admin/model/v1"
	"github.com/kong/koko/internal/model/validation"
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
		assert.Equal(t, &v1.Service{
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
	s := v1.Service{
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
	err := svc.Validate()
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
		name      string
		Service   func() Service
		wantErr   bool
		errFields []string
	}{
		{
			name: "empty service throws an error",
			Service: func() Service {
				return NewService()
			},
			wantErr: true,
			errFields: []string{
				"host",
				"path",
				"connect_timeout",
				"read_timeout",
				"write_timeout",
				"id",
				"port",
			},
		},
		{
			name: "default service throws an error",
			Service: func() Service {
				s := NewService()
				_ = s.ProcessDefaults()
				return s
			},
			wantErr:   true,
			errFields: []string{"host", "path"},
		},
		{
			name: "invalid ID throws an error",
			Service: func() Service {
				s := goodService()
				s.Service.Id = "bad-uuid"
				return s
			},
			wantErr:   true,
			errFields: []string{"id"},
		},
		{
			name: "invalid name throws an error",
			Service: func() Service {
				s := goodService()
				s.Service.Name = "%foo"
				return s
			},
			wantErr:   true,
			errFields: []string{"name"},
		},
		{
			name: "invalid name throws an error",
			Service: func() Service {
				s := goodService()
				s.Service.Name = "%foo"
				return s
			},
			wantErr:   true,
			errFields: []string{"name"},
		},
		{
			name: "invalid protocol throws an error",
			Service: func() Service {
				s := goodService()
				s.Service.Protocol = "smtp"
				return s
			},
			wantErr: true,
			errFields: []string{
				"protocol",
				"path",
			},
		},
		{
			name: "invalid retries throws an error",
			Service: func() Service {
				s := goodService()
				s.Service.Retries = -1
				return s
			},
			wantErr:   true,
			errFields: []string{"retries"},
		},
		{
			name: "invalid port throws an error",
			Service: func() Service {
				s := goodService()
				s.Service.Port = 69420
				return s
			},
			wantErr:   true,
			errFields: []string{"port"},
		},
		{
			name: "path must begin with /",
			Service: func() Service {
				s := goodService()
				s.Service.Path = "foo"
				return s
			},
			wantErr:   true,
			errFields: []string{"path"},
		},
		{
			name: "path must not contain '//'",
			Service: func() Service {
				s := goodService()
				s.Service.Path = "/foo//bar"
				return s
			},
			wantErr:   true,
			errFields: []string{"path"},
		},
		{
			name: "ca_certificates must be ID",
			Service: func() Service {
				s := goodService()
				s.Service.CaCertificates = []string{"foo"}
				return s
			},
			wantErr: true,
			errFields: []string{
				"ca_certificates",
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
			errFields: []string{
				"tls_verify_depth",
				"tls_verify",
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
			if tt.errFields != nil {
				verr, _ := err.(validation.Error)
				gotFields := fieldsFromErr(verr)
				assert.ElementsMatchf(t, tt.errFields, gotFields,
					"mismatch in expected errors")
			}
		})
	}
}
