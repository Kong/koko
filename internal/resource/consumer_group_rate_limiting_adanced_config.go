package resource

import (
	"context"
	"fmt"

	v1 "github.com/kong/koko/internal/gen/grpc/kong/admin/model/v1"
	"github.com/kong/koko/internal/model"
	"github.com/kong/koko/internal/model/json/generator"
	"github.com/kong/koko/internal/model/json/validation"
	"github.com/kong/koko/internal/model/json/validation/typedefs"
)

const defaultWindowType = "sliding"

const TypeConsumerGroupRateLimitingAdvancedConfig = model.Type("consumer-group-rate-limiting-advanced-config")

func NewConsumerGroupRateLimitingAdvancedConfig() ConsumerGroupRateLimitingAdvancedConfig {
	return ConsumerGroupRateLimitingAdvancedConfig{Config: &v1.ConsumerGroupRateLimitingAdvancedConfig{}}
}

type ConsumerGroupRateLimitingAdvancedConfig struct {
	Config *v1.ConsumerGroupRateLimitingAdvancedConfig
}

func (c ConsumerGroupRateLimitingAdvancedConfig) ID() string {
	if c.Config == nil {
		return ""
	}
	return c.Config.Id
}

func (c ConsumerGroupRateLimitingAdvancedConfig) Type() model.Type {
	return TypeConsumerGroupRateLimitingAdvancedConfig
}

func (c ConsumerGroupRateLimitingAdvancedConfig) Validate(_ context.Context) error {
	if len(c.Config.Limit) != len(c.Config.WindowSize) {
		return validation.Error{
			Errs: []*v1.ErrorDetail{
				{
					Type:     v1.ErrorType_ERROR_TYPE_ENTITY,
					Messages: []string{"you must provide the same number of windows and limits"},
				},
			},
		}
	}
	return validation.Validate(string(TypeConsumerGroupRateLimitingAdvancedConfig), c.Config)
}

func (c ConsumerGroupRateLimitingAdvancedConfig) ProcessDefaults(_ context.Context) error {
	if c.Config == nil {
		return fmt.Errorf("invalid nil resource")
	}
	defaultID(&c.Config.Id)

	if c.Config.WindowType == "" {
		c.Config.WindowType = defaultWindowType
	}

	return nil
}

func (c ConsumerGroupRateLimitingAdvancedConfig) Resource() model.Resource {
	return c.Config
}

// SetResource implements the Object.SetResource interface.
func (c ConsumerGroupRateLimitingAdvancedConfig) SetResource(r model.Resource) error {
	return model.SetResource(c, r)
}

func (c ConsumerGroupRateLimitingAdvancedConfig) Indexes() []model.Index {
	indexes := make([]model.Index, 0)
	if c.Config.ConsumerGroupId != "" {
		indexes = append(indexes, []model.Index{
			{
				Name:      "consumer_group_id",
				Type:      model.IndexUnique,
				Value:     c.Config.ConsumerGroupId,
				FieldName: "consumer_group_id",
			},
			{
				Name:        "consumer_group_id",
				Type:        model.IndexForeign,
				ForeignType: TypeConsumerGroup,
				Value:       c.Config.ConsumerGroupId,
				FieldName:   "consumer_group_id",
			},
		}...)
	}
	return indexes
}

func init() {
	if err := model.RegisterType(
		TypeConsumerGroupRateLimitingAdvancedConfig,
		&v1.ConsumerGroupRateLimitingAdvancedConfig{},
		func() model.Object {
			return NewConsumerGroupRateLimitingAdvancedConfig()
		},
	); err != nil {
		panic(err)
	}

	schema := &generator.Schema{
		Type: "object",
		Properties: map[string]*generator.Schema{
			"id":                typedefs.ID,
			"created_at":        typedefs.UnixEpoch,
			"updated_at":        typedefs.UnixEpoch,
			"consumer_group_id": typedefs.ID,
			"limit": {
				Type:  "array",
				Items: &generator.Schema{Type: "integer"},
			},
			"retry_after_jitter_max": {
				Type:    "integer",
				Minimum: intP(0),
				Default: float64(0),
			},
			"window_size": {
				Type:  "array",
				Items: &generator.Schema{Type: "integer"},
			},
			"window_type": {
				Type: "string",
				Enum: []interface{}{
					"fixed",
					"sliding",
				},
				Default: defaultWindowType,
			},
		},
		AdditionalProperties: &falsy,
		Required:             []string{"id", "window_size", "limit", "consumer_group_id"},
	}
	if err := generator.DefaultRegistry.Register(
		string(TypeConsumerGroupRateLimitingAdvancedConfig),
		schema,
	); err != nil {
		panic(err)
	}
}
