package resource

import (
	"reflect"
	"testing"
	"time"

	model "github.com/kong/koko/internal/gen/grpc/kong/admin/model/v1"
	"github.com/kong/koko/internal/log"
	internalModel "github.com/kong/koko/internal/model"
	"github.com/kong/koko/internal/model/json/validation"
	"github.com/kong/koko/internal/plugin"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/types/known/structpb"
)

func TestNewPlugin(t *testing.T) {
	s := NewPlugin()
	require.NotNil(t, s)
	require.NotNil(t, s.Plugin)
}

func TestPlugin_ID(t *testing.T) {
	var s Plugin
	id := s.ID()
	require.Empty(t, id)
	s = NewPlugin()
	id = s.ID()
	require.Empty(t, id)
}

func TestPlugin_Type(t *testing.T) {
	require.Equal(t, TypePlugin, NewPlugin().Type())
}

func setupLuaValidator(t *testing.T) {
	validator, err := plugin.NewLuaValidator(plugin.Opts{Logger: log.Logger})
	require.Nil(t, err)
	err = validator.LoadSchemasFromEmbed(plugin.Schemas, "schemas")
	require.Nil(t, err)
	SetValidator(validator)
}

func TestPlugin_ProcessDefaults(t *testing.T) {
	setupLuaValidator(t)
	defer plugin.ClearLuaSchemas()
	t.Run("defaults are correctly injected", func(t *testing.T) {
		r := NewPlugin()
		r.Plugin.Name = "basic-auth"
		err := r.ProcessDefaults()
		require.Nil(t, err)
		require.True(t, validUUID(r.ID()))
		require.LessOrEqual(t, r.Plugin.CreatedAt, int32(time.Now().Unix()))
		require.LessOrEqual(t, r.Plugin.UpdatedAt, int32(time.Now().Unix()))
		require.True(t, r.Plugin.Enabled.Value)
		require.ElementsMatch(t, []string{"http", "https", "grpc", "grpcs"},
			r.Plugin.Protocols)
		require.Nil(t, r.Plugin.Config.AsMap()["anonymous"])
		require.False(t, r.Plugin.Config.AsMap()["hide_credentials"].(bool))
	})
	t.Run("defaults do not override explicit values", func(t *testing.T) {
		r := NewPlugin()
		r.Plugin.Name = "rate-limiting"
		config, err := structpb.NewStruct(map[string]interface{}{
			"redis_port": 4242,
			"redis_ssl":  true,
		})
		require.Nil(t, err)
		r.Plugin.Config = config
		err = r.ProcessDefaults()
		require.Nil(t, err)
		require.Equal(t, float64(4242), r.Plugin.Config.AsMap()["redis_port"].(float64))
		require.True(t, r.Plugin.Config.AsMap()["redis_ssl"].(bool))
	})
}

func TestPlugin_Validate(t *testing.T) {
	setupLuaValidator(t)
	plugin.ClearLuaSchemas()
	tests := []struct {
		name    string
		Plugin  func() Plugin
		wantErr bool
		Errs    []*model.ErrorDetail
	}{
		{
			name: "valid plugin returns no errors",
			Plugin: func() Plugin {
				res := NewPlugin()
				res.Plugin.Name = "http-log"
				config, err := structpb.NewStruct(map[string]interface{}{
					"http_endpoint": "https://log.example.com",
				})
				if err != nil {
					panic(err)
				}
				res.Plugin.Config = config
				err = res.ProcessDefaults()
				if err != nil {
					panic(err)
				}
				return res
			},
			wantErr: false,
		},
		{
			name: "throws error when plugin doesn't exist",
			Plugin: func() Plugin {
				res := NewPlugin()
				res.Plugin.Name = "no-log"
				return res
			},
			wantErr: true,
			Errs: []*model.ErrorDetail{
				{
					Type:  model.ErrorType_ERROR_TYPE_FIELD,
					Field: "name",
					Messages: []string{
						"plugin(no-log) does not exist",
					},
				},
			},
		},
		{
			name: "throws error with invalid plugin",
			Plugin: func() Plugin {
				res := NewPlugin()
				res.Plugin.Name = "proxy-cache"
				config, err := structpb.NewStruct(map[string]interface{}{
					"bad_field": "what if?",
				})
				if err != nil {
					panic(err)
				}
				res.Plugin.Config = config
				err = res.ProcessDefaults()
				if err != nil {
					panic(err)
				}

				return res
			},
			wantErr: true,
			Errs: []*model.ErrorDetail{
				{
					Type:  model.ErrorType_ERROR_TYPE_FIELD,
					Field: "config.bad_field",
					Messages: []string{
						"unknown field",
					},
				},
				{
					Type:  model.ErrorType_ERROR_TYPE_FIELD,
					Field: "config.strategy",
					Messages: []string{
						"required field missing",
					},
				},
			},
		},
		{
			name: "throws error with invalid protocols for plugins",
			Plugin: func() Plugin {
				res := NewPlugin()
				res.Plugin.Name = "jwt"
				res.Plugin.Protocols = []string{"tcp"}
				err := res.ProcessDefaults()
				if err != nil {
					panic(err)
				}
				return res
			},
			wantErr: true,
			Errs: []*model.ErrorDetail{
				{
					Type:  model.ErrorType_ERROR_TYPE_FIELD,
					Field: "protocols[0]",
					Messages: []string{
						"expected one of: grpc, grpcs, http, https",
					},
				},
			},
		},
		{
			name: "throws error with invalid protocols based on jsonschema",
			Plugin: func() Plugin {
				res := NewPlugin()
				res.Plugin.Name = "jwt"
				res.Plugin.Protocols = []string{"smtp"}
				err := res.ProcessDefaults()
				if err != nil {
					panic(err)
				}
				return res
			},
			wantErr: true,
			Errs: []*model.ErrorDetail{
				{
					Type:  model.ErrorType_ERROR_TYPE_FIELD,
					Field: "protocols[0]",
					Messages: []string{
						`value must be one of "http", "https", ` +
							`"grpc", "grpcs", "tcp", "udp", "tls"`,
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.Plugin().Validate()
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

func TestPlugin_Indexes(t *testing.T) {
	type fields struct {
		Plugin *model.Plugin
	}
	tests := []struct {
		name   string
		fields fields
		want   []internalModel.Index
	}{
		{
			name: "returns an index for global plugin",
			fields: fields{
				Plugin: &model.Plugin{
					Name: "key-auth",
				},
			},
			want: []internalModel.Index{
				{
					Name:  "unique-plugin-per-entity",
					Type:  internalModel.IndexUnique,
					Value: "key-auth..",
				},
			},
		},
		{
			name: "returns indexes for a service-level plugin",
			fields: fields{
				Plugin: &model.Plugin{
					Name: "key-auth",
					Service: &model.Service{
						Id: "a03e65a1-a2f8-4953-9fca-2995d6ff4f6aB",
					},
				},
			},
			want: []internalModel.Index{
				{
					Name:  "unique-plugin-per-entity",
					Type:  internalModel.IndexUnique,
					Value: "key-auth.a03e65a1-a2f8-4953-9fca-2995d6ff4f6aB.",
				},
				{
					Name:        "service_id",
					Type:        internalModel.IndexForeign,
					FieldName:   "service.id",
					ForeignType: TypeService,
					Value:       "a03e65a1-a2f8-4953-9fca-2995d6ff4f6aB",
				},
			},
		},
		{
			name: "returns indexes for a route-level plugin",
			fields: fields{
				Plugin: &model.Plugin{
					Name: "key-auth",
					Route: &model.Route{
						Id: "7ed5812a-1281-4af0-aaaa-0490c1144451",
					},
				},
			},
			want: []internalModel.Index{
				{
					Name:  "unique-plugin-per-entity",
					Type:  internalModel.IndexUnique,
					Value: "key-auth..7ed5812a-1281-4af0-aaaa-0490c1144451",
				},
				{
					Name:        "route_id",
					Type:        internalModel.IndexForeign,
					FieldName:   "route.id",
					ForeignType: TypeRoute,
					Value:       "7ed5812a-1281-4af0-aaaa-0490c1144451",
				},
			},
		},
		{
			name: "returns indexes for a route asd service-level plugin",
			fields: fields{
				Plugin: &model.Plugin{
					Name: "key-auth",
					Route: &model.Route{
						Id: "7ed5812a-1281-4af0-aaaa-0490c1144451",
					},
					Service: &model.Service{
						Id: "33c3e0cc-bd5f-44bb-b642-e8441eaa4c56",
					},
				},
			},
			want: []internalModel.Index{
				{
					Name:  "unique-plugin-per-entity",
					Type:  internalModel.IndexUnique,
					Value: "key-auth.33c3e0cc-bd5f-44bb-b642-e8441eaa4c56.7ed5812a-1281-4af0-aaaa-0490c1144451",
				},
				{
					Name:        "route_id",
					Type:        internalModel.IndexForeign,
					FieldName:   "route.id",
					ForeignType: TypeRoute,
					Value:       "7ed5812a-1281-4af0-aaaa-0490c1144451",
				},
				{
					Name:        "service_id",
					Type:        internalModel.IndexForeign,
					FieldName:   "service.id",
					ForeignType: TypeService,
					Value:       "33c3e0cc-bd5f-44bb-b642-e8441eaa4c56",
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := Plugin{
				Plugin: tt.fields.Plugin,
			}
			if got := r.Indexes(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Indexes() = %v, want %v", got, tt.want)
			}
		})
	}
}
