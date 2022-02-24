package resource

import (
	"fmt"

	v1 "github.com/kong/koko/internal/gen/grpc/kong/admin/model/v1"
	"github.com/kong/koko/internal/model"
	"github.com/kong/koko/internal/model/json/generator"
	"github.com/kong/koko/internal/model/json/validation"
	"github.com/kong/koko/internal/model/json/validation/typedefs"
)

const (
	TypeConsumer = model.Type("consumer")
)

func NewConsumer() Consumer {
	return Consumer{
		Consumer: &v1.Consumer{},
	}
}

type Consumer struct {
	Consumer *v1.Consumer
}

func (c Consumer) ID() string {
	if c.Consumer == nil {
		return ""
	}
	return c.Consumer.Id
}

func (c Consumer) Type() model.Type {
	return TypeConsumer
}

func (c Consumer) Validate() error {
	return validation.Validate(string(TypeConsumer), c.Consumer)
}

func (c Consumer) ProcessDefaults() error {
	if c.Consumer == nil {
		return fmt.Errorf("invalid nil resource")
	}
	defaultID(&c.Consumer.Id)
	return nil
}

func (c Consumer) Resource() model.Resource {
	return c.Consumer
}

func (c Consumer) Indexes() []model.Index {
	maxIdxSize := 2
	idx := make([]model.Index, 0, maxIdxSize)
	if c.Consumer.Username != "" {
		userIdx := model.Index{
			Name:      "username",
			Type:      model.IndexUnique,
			Value:     c.Consumer.Username,
			FieldName: "username",
		}
		idx = append(idx, userIdx)
	}
	if c.Consumer.CustomId != "" {
		custIdx := model.Index{
			Name:      "custom_id",
			Type:      model.IndexUnique,
			Value:     c.Consumer.CustomId,
			FieldName: "custom_id",
		}
		idx = append(idx, custIdx)
	}
	return idx
}

func init() {
	err := model.RegisterType(TypeConsumer, func() model.Object {
		return NewConsumer()
	})
	if err != nil {
		panic(err)
	}

	consumerSchema := &generator.Schema{
		Type: "object",
		Properties: map[string]*generator.Schema{
			"id":         typedefs.ID,
			"username":   typedefs.Name,
			"created_at": typedefs.UnixEpoch,
			"updated_at": typedefs.UnixEpoch,
			"custom_id":  typedefs.Name,
			"tags":       typedefs.Tags,
		},
		AdditionalProperties: &falsy,
		Required:             []string{"id"},
		AnyOf: []*generator.Schema{
			{
				Description: "at least one of custom_id or username must be set",
				Required:    []string{"custom_id"},
			},
			{
				Description: "at least one of custom_id or username must be set",
				Required:    []string{"username"},
			},
		},
	}
	err = generator.Register(string(TypeConsumer), consumerSchema)
	if err != nil {
		panic(err)
	}
}
