package resource

import (
	"context"
	"testing"

	"github.com/google/uuid"
	model "github.com/kong/koko/internal/gen/grpc/kong/admin/model/v1"
	protoJSON "github.com/kong/koko/internal/json"
	"github.com/kong/koko/internal/model/json/validation"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

func TestNewUpstream(t *testing.T) {
	u := NewUpstream()
	require.NotNil(t, u)
	require.NotNil(t, u.Upstream)
}

func TestUpstream_ID(t *testing.T) {
	var u Upstream
	id := u.ID()
	require.Empty(t, id)
	u = NewUpstream()
	id = u.ID()
	require.Empty(t, id)
}

func TestUpstream_Type(t *testing.T) {
	require.Equal(t, TypeUpstream, NewUpstream().Type())
}

func TestUpstream_ProcessDefaults(t *testing.T) {
	t.Run("defaults are correctly injected", func(t *testing.T) {
		r := NewUpstream()
		err := r.ProcessDefaults(context.Background())
		require.Nil(t, err)
		require.True(t, validUUID(r.ID()))
		// empty out the id for equality comparison
		r.Upstream.Id = ""
		r.Upstream.CreatedAt = 0
		r.Upstream.UpdatedAt = 0
		require.Equal(t, r.Resource(), defaultUpstream)
	})
	t.Run("defaults do not override explicit values", func(t *testing.T) {
		r := NewUpstream()
		r.Upstream.Name = "foo"
		r.Upstream.HashOn = "cookie"
		r.Upstream.HashFallback = "ip"
		r.Upstream.Healthchecks = &model.Healthchecks{
			Active: &model.ActiveHealthcheck{
				Type:        "https",
				HttpsSni:    "*test.com",
				Concurrency: wrapperspb.Int32(32),
				Healthy: &model.ActiveHealthyCondition{
					Interval:  wrapperspb.Int32(1),
					Successes: wrapperspb.Int32(5),
				},
			},
		}
		err := r.ProcessDefaults(context.Background())
		require.Nil(t, err)
		require.True(t, validUUID(r.ID()))
		// empty out the id equality comparison
		r.Upstream.Id = ""
		expected := &model.Upstream{
			Name:             "foo",
			Algorithm:        "round-robin",
			Slots:            wrapperspb.Int32(defaultSlots),
			HashOn:           "cookie",
			HashFallback:     "ip",
			HashOnCookiePath: "/",
			Healthchecks: &model.Healthchecks{
				Threshold: wrapperspb.Float(0),
				Active: &model.ActiveHealthcheck{
					Concurrency: wrapperspb.Int32(32),
					Healthy: &model.ActiveHealthyCondition{
						HttpStatuses: []int32{200, 302},
						Interval:     wrapperspb.Int32(1),
						Successes:    wrapperspb.Int32(5),
					},
					HttpPath:               "/",
					HttpsVerifyCertificate: wrapperspb.Bool(true),
					HttpsSni:               "*test.com",
					Type:                   "https",
					Timeout:                wrapperspb.Int32(1),
					Unhealthy: &model.ActiveUnhealthyCondition{
						HttpFailures: wrapperspb.Int32(0),
						TcpFailures:  wrapperspb.Int32(0),
						HttpStatuses: []int32{429, 404, 500, 501, 502, 503, 504, 505},
						Timeouts:     wrapperspb.Int32(0),
						Interval:     wrapperspb.Int32(0),
					},
				},
				Passive: &model.PassiveHealthcheck{
					Healthy: &model.PassiveHealthyCondition{
						HttpStatuses: []int32{
							200, 201, 202, 203, 204, 205, 206, 207, 208, 226,
							300, 301, 302, 303, 304, 305, 306, 307, 308,
						},
						Successes: wrapperspb.Int32(0),
					},
					Type: "http",
					Unhealthy: &model.PassiveUnhealthyCondition{
						HttpFailures: wrapperspb.Int32(0),
						TcpFailures:  wrapperspb.Int32(0),
						HttpStatuses: []int32{429, 500, 503},
						Timeouts:     wrapperspb.Int32(0),
					},
				},
			},
		}
		actual := r.Resource()
		expectedJSON, err := protoJSON.ProtoJSONMarshal(expected)
		require.Nil(t, err)
		actualJSON, err := protoJSON.ProtoJSONMarshal(actual)
		require.Nil(t, err)
		require.JSONEq(t, string(expectedJSON), string(actualJSON))
	})
}

func TestUpstream_Validate(t *testing.T) {
	tests := []struct {
		name     string
		Upstream func() Upstream
		wantErr  bool
		Errs     []*model.ErrorDetail
	}{
		{
			name: "empty upstream throws an error",
			Upstream: func() Upstream {
				return NewUpstream()
			},
			wantErr: true,
			Errs: []*model.ErrorDetail{
				{
					Type: model.ErrorType_ERROR_TYPE_ENTITY,
					Messages: []string{
						"missing properties: 'id', 'name'",
					},
				},
			},
		},
		{
			name: "hash_on_header is required when hash_on is set to 'header'",
			Upstream: func() Upstream {
				u := NewUpstream()
				u.Upstream.Id = uuid.NewString()
				u.Upstream.Name = "foo"
				u.Upstream.HashOn = "header"
				return u
			},
			wantErr: true,
			Errs: []*model.ErrorDetail{
				{
					Type: model.ErrorType_ERROR_TYPE_ENTITY,
					Messages: []string{
						"when 'hash_on' is set to 'header'," +
							"'hash_on_header' must be set",
					},
				},
			},
		},
		{
			name: "hash_fallback_header is required when hash_fallback is set" +
				" to 'header'",
			Upstream: func() Upstream {
				u := NewUpstream()
				u.Upstream.Id = uuid.NewString()
				u.Upstream.Name = "foo"
				u.Upstream.HashFallback = "header"
				return u
			},
			wantErr: true,
			Errs: []*model.ErrorDetail{
				{
					Type: model.ErrorType_ERROR_TYPE_ENTITY,
					Messages: []string{
						"when 'hash_fallback' is set to 'header'," +
							"'hash_fallback_header' must be set",
					},
				},
			},
		},
		{
			name: "hash_on_cookie is required when hash_on is set" +
				" to 'cookie'",
			Upstream: func() Upstream {
				u := NewUpstream()
				u.Upstream.Id = uuid.NewString()
				u.Upstream.Name = "foo"
				u.Upstream.HashOn = "cookie"
				u.Upstream.HashFallback = "none"
				return u
			},
			wantErr: true,
			Errs: []*model.ErrorDetail{
				{
					Type: model.ErrorType_ERROR_TYPE_ENTITY,
					Messages: []string{
						"when 'hash_on' is set to 'cookie', " +
							"'hash_on_cookie' must be set",
					},
				},
			},
		},
		{
			name: "hash_fallback must be set to 'none' when 'hash_on' is" +
				" 'none'",
			Upstream: func() Upstream {
				u := NewUpstream()
				u.Upstream.Id = uuid.NewString()
				u.Upstream.Name = "foo"
				u.Upstream.HashOn = "none"
				u.Upstream.HashFallback = "cookie"
				u.Upstream.HashOnCookie = "foobar"
				return u
			},
			wantErr: true,
			Errs: []*model.ErrorDetail{
				{
					Type: model.ErrorType_ERROR_TYPE_ENTITY,
					Messages: []string{
						"when 'hash_on' is set to 'none', " +
							"'hash_fallback' must be set to 'none'",
					},
				},
			},
		},
		{
			name: "hash_fallback must be set to 'none' when 'hash_on' is" +
				" 'cookie'",
			Upstream: func() Upstream {
				u := NewUpstream()
				u.Upstream.Id = uuid.NewString()
				u.Upstream.Name = "foo"
				u.Upstream.HashOn = "cookie"
				u.Upstream.HashFallback = "cookie"
				u.Upstream.HashOnCookie = "foobar"
				return u
			},
			wantErr: true,
			Errs: []*model.ErrorDetail{
				{
					Type: model.ErrorType_ERROR_TYPE_ENTITY,
					Messages: []string{
						"when 'hash_on' is set to 'cookie', " +
							"'hash_fallback' must be set to 'none'",
					},
				},
			},
		},
		{
			name: "hash_fallback must not be set to 'consumer' when 'hash_on" +
				"' is set to 'consumer'",
			Upstream: func() Upstream {
				u := NewUpstream()
				u.Upstream.Id = uuid.NewString()
				u.Upstream.Name = "foo"
				u.Upstream.HashOn = "consumer"
				u.Upstream.HashFallback = "consumer"
				return u
			},
			wantErr: true,
			Errs: []*model.ErrorDetail{
				{
					Type: model.ErrorType_ERROR_TYPE_ENTITY,
					Messages: []string{
						"when 'hash_on' is set to 'consumer', " +
							"'hash_fallback' must be set to one of 'none', " +
							"'ip', 'header', 'cookie', 'path', " +
							"'query_arg', 'uri_capture'",
					},
				},
			},
		},
		{
			name: "hash_fallback must not be set to 'ip' when 'hash_on" +
				"' is set to 'ip'",
			Upstream: func() Upstream {
				u := NewUpstream()
				u.Upstream.Id = uuid.NewString()
				u.Upstream.Name = "foo"
				u.Upstream.HashOn = "ip"
				u.Upstream.HashFallback = "ip"
				return u
			},
			wantErr: true,
			Errs: []*model.ErrorDetail{
				{
					Type: model.ErrorType_ERROR_TYPE_ENTITY,
					Messages: []string{
						"when 'hash_on' is set to 'ip', " +
							"'hash_fallback' must be set to one of 'none', " +
							"'consumer', 'header', 'cookie', 'path', " +
							"'query_arg', 'uri_capture'",
					},
				},
			},
		},
		{
			name: "upstream with negative interval throws an error",
			Upstream: func() Upstream {
				u := NewUpstream()
				u.Upstream.Id = uuid.NewString()
				u.Upstream.Name = "foo"
				u.Upstream.Healthchecks = &model.Healthchecks{
					Threshold: wrapperspb.Float(0),
					Active: &model.ActiveHealthcheck{
						Concurrency: wrapperspb.Int32(32),
						Healthy: &model.ActiveHealthyCondition{
							HttpStatuses: []int32{200, 302},
							Interval:     wrapperspb.Int32(-1),
							Successes:    wrapperspb.Int32(5),
						},
					},
				}
				return u
			},
			wantErr: true,
			Errs: []*model.ErrorDetail{
				{
					Type:  model.ErrorType_ERROR_TYPE_FIELD,
					Field: "healthchecks.active.healthy.interval",
					Messages: []string{
						"must be >= 0 but found -1",
					},
				},
			},
		},
		{
			name: "hash_fallback_header must not be equal to hash_on_header",
			Upstream: func() Upstream {
				u := NewUpstream()
				u.Upstream.Id = uuid.NewString()
				u.Upstream.Name = "foo"
				u.Upstream.HashOn = "header"
				u.Upstream.HashFallback = "header"
				u.Upstream.HashOnHeader = "foobar"
				u.Upstream.HashFallbackHeader = "foobar"
				return u
			},
			wantErr: true,
			Errs: []*model.ErrorDetail{
				{
					Type: model.ErrorType_ERROR_TYPE_ENTITY,
					Messages: []string{
						"'hash_fallback_header' must not be equal to" +
							" 'hash_on_header'",
					},
				},
			},
		},
		{
			name: "healthchecks types can be set to http",
			Upstream: func() Upstream {
				u := NewUpstream()
				u.Upstream.Id = uuid.NewString()
				u.Upstream.Name = "foo"
				u.Upstream.HashFallback = "ip"
				u.Upstream.Healthchecks = &model.Healthchecks{
					Active: &model.ActiveHealthcheck{
						Type:        "http",
						Concurrency: wrapperspb.Int32(32),
						Healthy: &model.ActiveHealthyCondition{
							Interval:  wrapperspb.Int32(1),
							Successes: wrapperspb.Int32(5),
						},
					},
					Passive: &model.PassiveHealthcheck{
						Type: "http",
						Healthy: &model.PassiveHealthyCondition{
							Successes: wrapperspb.Int32(5),
						},
					},
				}
				return u
			},
		},
		{
			name: "healthchecks types can be set to https",
			Upstream: func() Upstream {
				u := NewUpstream()
				u.Upstream.Id = uuid.NewString()
				u.Upstream.Name = "foo"
				u.Upstream.HashFallback = "ip"
				u.Upstream.Healthchecks = &model.Healthchecks{
					Active: &model.ActiveHealthcheck{
						Type:        "https",
						Concurrency: wrapperspb.Int32(32),
						Healthy: &model.ActiveHealthyCondition{
							Interval:  wrapperspb.Int32(1),
							Successes: wrapperspb.Int32(5),
						},
					},
					Passive: &model.PassiveHealthcheck{
						Type: "https",
						Healthy: &model.PassiveHealthyCondition{
							Successes: wrapperspb.Int32(5),
						},
					},
				}
				return u
			},
		},
		{
			name: "healthchecks types can be set to tcp",
			Upstream: func() Upstream {
				u := NewUpstream()
				u.Upstream.Id = uuid.NewString()
				u.Upstream.Name = "foo"
				u.Upstream.HashFallback = "ip"
				u.Upstream.Healthchecks = &model.Healthchecks{
					Active: &model.ActiveHealthcheck{
						Type:        "tcp",
						Concurrency: wrapperspb.Int32(32),
						Healthy: &model.ActiveHealthyCondition{
							Interval:  wrapperspb.Int32(1),
							Successes: wrapperspb.Int32(5),
						},
					},
					Passive: &model.PassiveHealthcheck{
						Type: "tcp",
						Healthy: &model.PassiveHealthyCondition{
							Successes: wrapperspb.Int32(5),
						},
					},
				}
				return u
			},
		},
		{
			name: "healthchecks types can be set to grpc",
			Upstream: func() Upstream {
				u := NewUpstream()
				u.Upstream.Id = uuid.NewString()
				u.Upstream.Name = "foo"
				u.Upstream.HashFallback = "ip"
				u.Upstream.Healthchecks = &model.Healthchecks{
					Active: &model.ActiveHealthcheck{
						Type:        "grpc",
						Concurrency: wrapperspb.Int32(32),
						Healthy: &model.ActiveHealthyCondition{
							Interval:  wrapperspb.Int32(1),
							Successes: wrapperspb.Int32(5),
						},
					},
					Passive: &model.PassiveHealthcheck{
						Type: "grpc",
						Healthy: &model.PassiveHealthyCondition{
							Successes: wrapperspb.Int32(5),
						},
					},
				}
				return u
			},
		},
		{
			name: "healthchecks types can be set to grpcs",
			Upstream: func() Upstream {
				u := NewUpstream()
				u.Upstream.Id = uuid.NewString()
				u.Upstream.Name = "foo"
				u.Upstream.HashFallback = "ip"
				u.Upstream.Healthchecks = &model.Healthchecks{
					Active: &model.ActiveHealthcheck{
						Type:        "grpcs",
						Concurrency: wrapperspb.Int32(32),
						Healthy: &model.ActiveHealthyCondition{
							Interval:  wrapperspb.Int32(1),
							Successes: wrapperspb.Int32(5),
						},
					},
					Passive: &model.PassiveHealthcheck{
						Type: "grpcs",
						Healthy: &model.PassiveHealthyCondition{
							Successes: wrapperspb.Int32(5),
						},
					},
				}
				return u
			},
		},
		{
			name: "set wrong types for healthchecks",
			Upstream: func() Upstream {
				u := NewUpstream()
				u.Upstream.Id = uuid.NewString()
				u.Upstream.Name = "foo"
				u.Upstream.HashFallback = "ip"
				u.Upstream.Healthchecks = &model.Healthchecks{
					Active: &model.ActiveHealthcheck{
						Type:        "bar",
						Concurrency: wrapperspb.Int32(32),
						Healthy: &model.ActiveHealthyCondition{
							Interval:  wrapperspb.Int32(1),
							Successes: wrapperspb.Int32(5),
						},
					},
					Passive: &model.PassiveHealthcheck{
						Type: "baz",
						Healthy: &model.PassiveHealthyCondition{
							Successes: wrapperspb.Int32(5),
						},
					},
				}
				return u
			},
			wantErr: true,
			Errs: []*model.ErrorDetail{
				{
					Type:  model.ErrorType_ERROR_TYPE_FIELD,
					Field: "healthchecks.active.type",
					Messages: []string{
						`value must be one of "tcp", "http", "https", "grpc", "grpcs"`,
					},
				},
				{
					Type:  model.ErrorType_ERROR_TYPE_FIELD,
					Field: "healthchecks.passive.type",
					Messages: []string{
						`value must be one of "tcp", "http", "https", "grpc", "grpcs"`,
					},
				},
			},
		},
		{
			name: "hash_fallback must not be set to 'path' when 'hash_on" +
				"' is set to 'path'",
			Upstream: func() Upstream {
				u := NewUpstream()
				u.Upstream.Id = uuid.NewString()
				u.Upstream.Name = "foo"
				u.Upstream.HashOn = "path"
				u.Upstream.HashFallback = "path"
				return u
			},
			wantErr: true,
			Errs: []*model.ErrorDetail{
				{
					Type: model.ErrorType_ERROR_TYPE_ENTITY,
					Messages: []string{
						"when 'hash_on' is set to 'path', " +
							"'hash_fallback' must be set to one of 'none', " +
							"'consumer', 'ip', 'header', 'cookie', " +
							"'query_arg', 'uri_capture'",
					},
				},
			},
		},
		{
			name: "hash_on_query_arg is required when hash_on is set" +
				" to 'query_arg'",
			Upstream: func() Upstream {
				u := NewUpstream()
				u.Upstream.Id = uuid.NewString()
				u.Upstream.Name = "foo"
				u.Upstream.HashOn = "query_arg"
				return u
			},
			wantErr: true,
			Errs: []*model.ErrorDetail{
				{
					Type: model.ErrorType_ERROR_TYPE_ENTITY,
					Messages: []string{
						"when 'hash_on' is set to 'query_arg', " +
							"'hash_on_query_arg' must be set",
					},
				},
			},
		},
		{
			name: "hash_fallback_query_arg is required when hash_fallback is set" +
				" to 'query_arg'",
			Upstream: func() Upstream {
				u := NewUpstream()
				u.Upstream.Id = uuid.NewString()
				u.Upstream.Name = "foo"
				u.Upstream.HashFallback = "query_arg"
				return u
			},
			wantErr: true,
			Errs: []*model.ErrorDetail{
				{
					Type: model.ErrorType_ERROR_TYPE_ENTITY,
					Messages: []string{
						"when 'hash_fallback' is set to 'query_arg', " +
							"'hash_fallback_query_arg' must be set",
					},
				},
			},
		},
		{
			name: "hash_on_query_arg is required when hash_on is set" +
				" to 'uri_capture'",
			Upstream: func() Upstream {
				u := NewUpstream()
				u.Upstream.Id = uuid.NewString()
				u.Upstream.Name = "foo"
				u.Upstream.HashOn = "uri_capture"
				return u
			},
			wantErr: true,
			Errs: []*model.ErrorDetail{
				{
					Type: model.ErrorType_ERROR_TYPE_ENTITY,
					Messages: []string{
						"when 'hash_on' is set to 'uri_capture', " +
							"'hash_on_uri_capture' must be set",
					},
				},
			},
		},
		{
			name: "hash_fallback_uri_capture is required when hash_fallback is set" +
				" to 'uri_capture'",
			Upstream: func() Upstream {
				u := NewUpstream()
				u.Upstream.Id = uuid.NewString()
				u.Upstream.Name = "foo"
				u.Upstream.HashFallback = "uri_capture"
				return u
			},
			wantErr: true,
			Errs: []*model.ErrorDetail{
				{
					Type: model.ErrorType_ERROR_TYPE_ENTITY,
					Messages: []string{
						"when 'hash_fallback' is set to 'uri_capture', " +
							"'hash_fallback_uri_capture' must be set",
					},
				},
			},
		},
		{
			name: "hash_on_query_arg must not be equal to hash_fallback_query_arg",
			Upstream: func() Upstream {
				u := NewUpstream()
				u.Upstream.Id = uuid.NewString()
				u.Upstream.Name = "foo"
				u.Upstream.HashOn = "query_arg"
				u.Upstream.HashFallbackQueryArg = "query"
				u.Upstream.HashOnQueryArg = "query"
				return u
			},
			wantErr: true,
			Errs: []*model.ErrorDetail{
				{
					Type: model.ErrorType_ERROR_TYPE_ENTITY,
					Messages: []string{
						"'hash_on_query_arg' must not be equal to" +
							" 'hash_fallback_query_arg'",
					},
				},
			},
		},
		{
			name: "hash_on_uri_capture must not be equal to hash_fallback_uri_capture",
			Upstream: func() Upstream {
				u := NewUpstream()
				u.Upstream.Id = uuid.NewString()
				u.Upstream.Name = "foo"
				u.Upstream.HashOn = "uri_capture"
				u.Upstream.HashFallbackUriCapture = "foobar"
				u.Upstream.HashOnUriCapture = "foobar"
				return u
			},
			wantErr: true,
			Errs: []*model.ErrorDetail{
				{
					Type: model.ErrorType_ERROR_TYPE_ENTITY,
					Messages: []string{
						"'hash_on_uri_capture' must not be equal to" +
							" 'hash_fallback_uri_capture'",
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.Upstream().Validate(context.Background())
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
