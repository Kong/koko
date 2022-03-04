package resource

import (
	"reflect"
	"testing"

	"github.com/google/uuid"
	v1 "github.com/kong/koko/internal/gen/grpc/kong/admin/model/v1"
	"github.com/kong/koko/internal/model"
	"github.com/kong/koko/internal/model/json/validation"
	"github.com/stretchr/testify/require"
)

func TestNewSNI(t *testing.T) {
	s := NewSNI()
	require.NotNil(t, s)
	require.NotNil(t, s.SNI)
}

func TestSNI_Type(t *testing.T) {
	require.Equal(t, TypeSNI, NewSNI().Type())
}

func TestSNI_ProcessDefaults(t *testing.T) {
	sni := NewSNI()
	require.Nil(t, sni.ProcessDefaults())
	require.NotPanics(t, func() {
		uuid.MustParse(sni.ID())
	})
}

func TestSNI_Validate(t *testing.T) {
	tests := []struct {
		name    string
		SNI     func() SNI
		wantErr bool
		Errs    []*v1.ErrorDetail
	}{
		{
			name: "SNI missing required fields returns an error",
			SNI: func() SNI {
				res := NewSNI()
				return res
			},
			wantErr: true,
			Errs: []*v1.ErrorDetail{
				{
					Type: v1.ErrorType_ERROR_TYPE_ENTITY,
					Messages: []string{
						"missing properties: 'id', 'name', 'certificate'",
					},
				},
			},
		},
		{
			name: "SNI without a name returns an error",
			SNI: func() SNI {
				res := NewSNI()
				_ = res.ProcessDefaults()
				res.SNI.Certificate = &v1.Certificate{
					Id: uuid.NewString(),
				}
				return res
			},
			wantErr: true,
			Errs: []*v1.ErrorDetail{
				{
					Type: v1.ErrorType_ERROR_TYPE_ENTITY,
					Messages: []string{
						"missing properties: 'name'",
					},
				},
			},
		},
		{
			name: "SNI without certificate returns an error",
			SNI: func() SNI {
				res := NewSNI()
				_ = res.ProcessDefaults()
				res.SNI.Name = "example.com"
				return res
			},
			wantErr: true,
			Errs: []*v1.ErrorDetail{
				{
					Type: v1.ErrorType_ERROR_TYPE_ENTITY,
					Messages: []string{
						"missing properties: 'certificate'",
					},
				},
			},
		},
		{
			name: "SNI without a certificate id returns an error",
			SNI: func() SNI {
				res := NewSNI()
				_ = res.ProcessDefaults()
				res.SNI.Name = "*.example.com"
				res.SNI.Certificate = &v1.Certificate{}
				return res
			},
			wantErr: true,
			Errs: []*v1.ErrorDetail{
				{
					Field: "certificate",
					Type:  v1.ErrorType_ERROR_TYPE_FIELD,
					Messages: []string{
						"missing properties: 'id'",
					},
				},
			},
		},
		{
			name: "SNI with invalid name returns an error",
			SNI: func() SNI {
				res := NewSNI()
				_ = res.ProcessDefaults()
				res.SNI.Name = "TeST"
				res.SNI.Certificate = &v1.Certificate{
					Id: uuid.NewString(),
				}
				return res
			},
			wantErr: true,
			Errs: []*v1.ErrorDetail{
				{
					Field: "name",
					Type:  v1.ErrorType_ERROR_TYPE_FIELD,
					Messages: []string{
						"must be a valid hostname with a wildcard prefix '*' or without",
					},
				},
			},
		},
		{
			name: "SNI with an invalid wildcard position returns an error",
			SNI: func() SNI {
				res := NewSNI()
				_ = res.ProcessDefaults()
				res.SNI.Name = "foo.example.*"
				res.SNI.Certificate = &v1.Certificate{
					Id: uuid.NewString(),
				}
				return res
			},
			wantErr: true,
			Errs: []*v1.ErrorDetail{
				{
					Field: "name",
					Type:  v1.ErrorType_ERROR_TYPE_FIELD,
					Messages: []string{
						"must be a valid hostname with a wildcard prefix '*' or without",
					},
				},
			},
		},
		{
			name: "valid SNI with wildcard hostname returns no errors",
			SNI: func() SNI {
				res := NewSNI()
				_ = res.ProcessDefaults()
				res.SNI.Name = "*.example.com"
				res.SNI.Certificate = &v1.Certificate{
					Id: uuid.NewString(),
				}
				return res
			},
			wantErr: false,
		},
		{
			name: "valid SNI without wildcard hostname returns no errors",
			SNI: func() SNI {
				res := NewSNI()
				_ = res.ProcessDefaults()
				res.SNI.Name = "one-two.example.com"
				res.SNI.Certificate = &v1.Certificate{
					Id: uuid.NewString(),
				}
				return res
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.SNI().Validate()
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

func TestSNI_Indexes(t *testing.T) {
	type fields struct {
		SNI *v1.SNI
	}
	certID := uuid.NewString()
	tests := []struct {
		name   string
		fields fields
		want   []model.Index
	}{
		{
			name: "returns an index for name and certificate id",
			fields: fields{
				SNI: &v1.SNI{
					Name: "hostname",
					Certificate: &v1.Certificate{
						Id: certID,
					},
				},
			},
			want: []model.Index{
				{
					Name:      "unique-name",
					Type:      model.IndexUnique,
					Value:     "hostname",
					FieldName: "name",
				},
				{
					Name:        "certificate_id",
					Type:        model.IndexForeign,
					ForeignType: TypeCertificate,
					FieldName:   "certificate.id",
					Value:       certID,
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := SNI{
				SNI: tt.fields.SNI,
			}
			if got := r.Indexes(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Indexes() = %v, want %v", got, tt.want)
			}
		})
	}
}
